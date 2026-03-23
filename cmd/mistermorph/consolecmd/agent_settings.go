package consolecmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/integration"
	"github.com/quailyquaily/mistermorph/internal/configutil"
	"github.com/quailyquaily/mistermorph/internal/jsonutil"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/internal/pathutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	llmSettingsKey        = "llm"
	multimodalSettingsKey = "multimodal"
	toolsSettingsKey      = "tools"
)

var supportedMultimodalSources = []string{"telegram", "slack", "line", "remote_download"}

type llmSettingsPayload struct {
	Provider            string `json:"provider"`
	Endpoint            string `json:"endpoint"`
	Model               string `json:"model"`
	APIKey              string `json:"api_key"`
	CloudflareAccountID string `json:"cloudflare_account_id"`
	ReasoningEffort     string `json:"reasoning_effort"`
	ToolsEmulationMode  string `json:"tools_emulation_mode"`
}

type multimodalSettingsPayload struct {
	ImageSources []string `json:"image_sources"`
}

type toolsSettingsPayload struct {
	WriteFileEnabled    bool `json:"write_file_enabled"`
	ContactsSendEnabled bool `json:"contacts_send_enabled"`
	TodoUpdateEnabled   bool `json:"todo_update_enabled"`
	PlanCreateEnabled   bool `json:"plan_create_enabled"`
	URLFetchEnabled     bool `json:"url_fetch_enabled"`
	WebSearchEnabled    bool `json:"web_search_enabled"`
	BashEnabled         bool `json:"bash_enabled"`
}

type agentSettingsPayload struct {
	LLM        llmSettingsPayload        `json:"llm"`
	Multimodal multimodalSettingsPayload `json:"multimodal"`
	Tools      toolsSettingsPayload      `json:"tools"`
}

type agentSettingsModelsRequest struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
}

type agentSettingsTestRequest struct {
	LLM llmSettingsPayload `json:"llm"`
}

type agentSettingsBenchmarkResult struct {
	ID         string `json:"id"`
	OK         bool   `json:"ok"`
	DurationMS int64  `json:"duration_ms"`
	Detail     string `json:"detail,omitempty"`
	Error      string `json:"error,omitempty"`
}

type agentSettingsTestResult struct {
	Provider   string
	Model      string
	Benchmarks []agentSettingsBenchmarkResult
}

var runAgentSettingsConnectionTest = defaultAgentSettingsConnectionTest

func (s *server) handleAgentSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleAgentSettingsGet(w, r)
	case http.MethodPut:
		s.handleAgentSettingsPut(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleAgentSettingsGet(w http.ResponseWriter, _ *http.Request) {
	configPath, err := resolveConsoleConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	configExists, configSource, err := inspectAgentSettingsConfigSource(configPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	configValid := true
	settings, err := readAgentSettings(configPath)
	if err != nil {
		if !isInvalidConfigYAMLError(err) {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		settings = readAgentSettingsFromReader(viper.GetViper())
		configSource = "defaults"
		configValid = false
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"llm":           settings.LLM,
		"multimodal":    settings.Multimodal,
		"tools":         settings.Tools,
		"config_path":   configPath,
		"config_exists": configExists,
		"config_valid":  configValid,
		"config_source": configSource,
	})
}

func (s *server) handleAgentSettingsPut(w http.ResponseWriter, r *http.Request) {
	var req agentSettingsPayload
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	configPath, err := resolveConsoleConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	serialized, err := writeAgentSettings(configPath, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tmp, err := validateAgentConfigDocument(serialized)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := os.WriteFile(configPath, serialized, 0o600); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	viper.Set(llmSettingsKey, tmp.Get(llmSettingsKey))
	viper.Set(multimodalSettingsKey, tmp.Get(multimodalSettingsKey))
	viper.Set(toolsSettingsKey, tmp.Get(toolsSettingsKey))
	if s != nil && s.localRuntime != nil {
		if err := s.localRuntime.ReloadAgentConfig(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if s != nil && s.managed != nil {
		if err := s.managed.Restart(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	next := readAgentSettingsFromReader(tmp)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"llm":           next.LLM,
		"multimodal":    next.Multimodal,
		"tools":         next.Tools,
		"config_path":   configPath,
		"config_exists": true,
		"config_valid":  true,
		"config_source": "config",
	})
}

func (s *server) handleAgentSettingsModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req agentSettingsModelsRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(req.APIKey) == "" {
		writeError(w, http.StatusBadRequest, "api key is required")
		return
	}
	models, err := fetchOpenAICompatibleModels(r.Context(), req.Endpoint, req.APIKey)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": models,
	})
}

func (s *server) handleAgentSettingsTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req agentSettingsTestRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := runAgentSettingsConnectionTest(r.Context(), req.LLM)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"provider":   result.Provider,
		"model":      result.Model,
		"benchmarks": result.Benchmarks,
	})
}

func resolveConsoleConfigPath() (string, error) {
	explicit := strings.TrimSpace(viper.GetString("config"))
	if explicit != "" {
		return pathutil.ExpandHomePath(explicit), nil
	}
	for _, candidate := range []string{"config.yaml", "~/.morph/config.yaml"} {
		resolved := pathutil.ExpandHomePath(candidate)
		if _, err := os.Stat(resolved); err == nil {
			return resolved, nil
		}
	}
	return filepath.Clean(pathutil.ExpandHomePath("config.yaml")), nil
}

func readAgentSettings(configPath string) (agentSettingsPayload, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return readAgentSettingsFromReader(viper.GetViper()), nil
		}
		return agentSettingsPayload{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return readAgentSettingsFromReader(viper.GetViper()), nil
	}
	tmp := viper.New()
	integration.ApplyViperDefaults(tmp)
	if err := readExpandedConsoleConfig(tmp, configPath); err != nil {
		return agentSettingsPayload{}, fmt.Errorf("invalid config yaml: %w", err)
	}
	return readAgentSettingsFromReader(tmp), nil
}

func inspectAgentSettingsConfigSource(configPath string) (bool, string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "defaults", nil
		}
		return false, "", err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return true, "defaults", nil
	}
	return true, "config", nil
}

func writeAgentSettings(configPath string, values agentSettingsPayload) ([]byte, error) {
	doc, err := loadYAMLDocument(configPath)
	if err != nil {
		if !isInvalidConfigYAMLError(err) {
			return nil, err
		}
		doc = newEmptyYAMLDocument()
	}
	root, err := documentMapping(doc)
	if err != nil {
		return nil, err
	}

	llmNode := ensureMappingValue(root, llmSettingsKey)
	setOrDeleteMappingScalar(llmNode, "provider", values.LLM.Provider)
	setOrDeleteMappingScalar(llmNode, "endpoint", values.LLM.Endpoint)
	setOrDeleteMappingScalar(llmNode, "model", values.LLM.Model)
	setOrDeleteMappingScalar(llmNode, "api_key", values.LLM.APIKey)
	if strings.EqualFold(strings.TrimSpace(values.LLM.Provider), "cloudflare") {
		cloudflareNode := ensureMappingValue(llmNode, "cloudflare")
		setOrDeleteMappingScalar(cloudflareNode, "account_id", values.LLM.CloudflareAccountID)
	} else {
		deleteMappingKey(llmNode, "cloudflare")
	}
	setOrDeleteMappingScalar(llmNode, "reasoning_effort", values.LLM.ReasoningEffort)
	setOrDeleteMappingScalar(llmNode, "tools_emulation_mode", values.LLM.ToolsEmulationMode)

	multimodalNode := ensureMappingValue(root, multimodalSettingsKey)
	imageNode := ensureMappingValue(multimodalNode, "image")
	setMappingStringList(imageNode, "sources", values.Multimodal.ImageSources)

	toolsNode := ensureMappingValue(root, toolsSettingsKey)
	setMappingBoolPath(toolsNode, "write_file", "enabled", values.Tools.WriteFileEnabled)
	setMappingBoolPath(toolsNode, "contacts_send", "enabled", values.Tools.ContactsSendEnabled)
	setMappingBoolPath(toolsNode, "todo_update", "enabled", values.Tools.TodoUpdateEnabled)
	setMappingBoolPath(toolsNode, "plan_create", "enabled", values.Tools.PlanCreateEnabled)
	setMappingBoolPath(toolsNode, "url_fetch", "enabled", values.Tools.URLFetchEnabled)
	setMappingBoolPath(toolsNode, "web_search", "enabled", values.Tools.WebSearchEnabled)
	setMappingBoolPath(toolsNode, "bash", "enabled", values.Tools.BashEnabled)

	return marshalYAMLDocument(doc)
}

func validateAgentConfigDocument(data []byte) (*viper.Viper, error) {
	tmp := viper.New()
	integration.ApplyViperDefaults(tmp)
	tmp.SetConfigType("yaml")
	if err := tmp.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("invalid config yaml: %w", err)
	}
	values := llmutil.RuntimeValuesFromReader(tmp)
	route, err := llmutil.ResolveRoute(values, llmutil.RoutePurposeMainLoop)
	if err != nil {
		return nil, err
	}
	if _, err := llmutil.ClientFromConfigWithValues(route.ClientConfig, route.Values); err != nil {
		return nil, err
	}
	return tmp, nil
}

func defaultAgentSettingsConnectionTest(ctx context.Context, settings llmSettingsPayload) (agentSettingsTestResult, error) {
	values := llmutil.RuntimeValues{
		Provider:            normalizeAgentSettingsProvider(settings.Provider),
		Endpoint:            strings.TrimSpace(settings.Endpoint),
		APIKey:              strings.TrimSpace(settings.APIKey),
		Model:               strings.TrimSpace(settings.Model),
		RequestTimeoutRaw:   "20s",
		ReasoningEffortRaw:  strings.TrimSpace(settings.ReasoningEffort),
		ToolsEmulationMode:  strings.TrimSpace(settings.ToolsEmulationMode),
		CloudflareAccountID: strings.TrimSpace(settings.CloudflareAccountID),
	}

	route, err := llmutil.ResolveRoute(values, llmutil.RoutePurposeMainLoop)
	if err != nil {
		return agentSettingsTestResult{}, err
	}
	client, err := llmutil.ClientFromConfigWithValues(route.ClientConfig, route.Values)
	if err != nil {
		return agentSettingsTestResult{}, err
	}

	return agentSettingsTestResult{
		Provider: route.ClientConfig.Provider,
		Model:    route.ClientConfig.Model,
		Benchmarks: []agentSettingsBenchmarkResult{
			runAgentSettingsTextBenchmark(ctx, client, route.ClientConfig.Model),
			runAgentSettingsJSONBenchmark(ctx, client, route.ClientConfig.Model),
			runAgentSettingsToolCallingBenchmark(ctx, client, route.ClientConfig.Model),
		},
	}, nil
}

func runAgentSettingsTextBenchmark(ctx context.Context, client llm.Client, model string) agentSettingsBenchmarkResult {
	start := time.Now()
	result, err := client.Chat(ctx, llm.Request{
		Model: model,
		Messages: []llm.Message{
			{Role: "system", Content: "You are a connection test. Reply briefly."},
			{Role: "user", Content: "Reply with exactly: OK"},
		},
		Parameters: map[string]any{
			"max_tokens": 24,
		},
	})
	durationMS := time.Since(start).Milliseconds()
	if err != nil {
		return agentSettingsBenchmarkResult{
			ID:         "text_reply",
			OK:         false,
			DurationMS: durationMS,
			Error:      strings.TrimSpace(err.Error()),
		}
	}

	text := strings.TrimSpace(result.Text)
	if text == "" {
		return agentSettingsBenchmarkResult{
			ID:         "text_reply",
			OK:         false,
			DurationMS: durationMS,
			Error:      "received an empty text reply",
		}
	}

	return agentSettingsBenchmarkResult{
		ID:         "text_reply",
		OK:         true,
		DurationMS: durationMS,
		Detail:     summarizeBenchmarkDetail(text),
	}
}

func runAgentSettingsJSONBenchmark(ctx context.Context, client llm.Client, model string) agentSettingsBenchmarkResult {
	start := time.Now()
	result, err := client.Chat(ctx, llm.Request{
		Model:     model,
		ForceJSON: true,
		Messages: []llm.Message{
			{Role: "system", Content: "You are a JSON response test. Reply with a JSON object only."},
			{Role: "user", Content: `Return exactly this JSON object: {"status":"ok","message":"json ok"}`},
		},
		Parameters: map[string]any{
			"max_tokens": 48,
		},
	})
	durationMS := time.Since(start).Milliseconds()
	if err != nil {
		return agentSettingsBenchmarkResult{
			ID:         "json_response",
			OK:         false,
			DurationMS: durationMS,
			Error:      strings.TrimSpace(err.Error()),
		}
	}

	var payload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := jsonutil.DecodeWithFallback(result.Text, &payload); err != nil {
		return agentSettingsBenchmarkResult{
			ID:         "json_response",
			OK:         false,
			DurationMS: durationMS,
			Error:      "response was not valid json",
		}
	}
	if strings.TrimSpace(payload.Status) != "ok" {
		return agentSettingsBenchmarkResult{
			ID:         "json_response",
			OK:         false,
			DurationMS: durationMS,
			Error:      "json response missing status=ok",
		}
	}

	detail := summarizeBenchmarkDetail(strings.TrimSpace(payload.Message))
	if detail == "" {
		detail = "status=ok"
	}
	return agentSettingsBenchmarkResult{
		ID:         "json_response",
		OK:         true,
		DurationMS: durationMS,
		Detail:     detail,
	}
}

func runAgentSettingsToolCallingBenchmark(ctx context.Context, client llm.Client, model string) agentSettingsBenchmarkResult {
	start := time.Now()
	result, err := client.Chat(ctx, llm.Request{
		Model: model,
		Messages: []llm.Message{
			{Role: "system", Content: "You are a tool calling test. Always call the ping tool exactly once."},
			{Role: "user", Content: "Call the ping tool now."},
		},
		Tools: []llm.Tool{
			{
				Name:           "ping",
				Description:    "Connectivity check tool.",
				ParametersJSON: `{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`,
			},
		},
		Parameters: map[string]any{
			"max_tokens": 96,
		},
	})
	durationMS := time.Since(start).Milliseconds()
	if err != nil {
		return agentSettingsBenchmarkResult{
			ID:         "tool_calling",
			OK:         false,
			DurationMS: durationMS,
			Error:      strings.TrimSpace(err.Error()),
		}
	}

	for _, call := range result.ToolCalls {
		if strings.EqualFold(strings.TrimSpace(call.Name), "ping") {
			return agentSettingsBenchmarkResult{
				ID:         "tool_calling",
				OK:         true,
				DurationMS: durationMS,
				Detail:     "called ping",
			}
		}
	}

	detail := summarizeBenchmarkDetail(strings.TrimSpace(result.Text))
	if detail == "" {
		detail = "model replied without calling the tool"
	} else {
		detail = "model replied without calling the tool: " + detail
	}
	return agentSettingsBenchmarkResult{
		ID:         "tool_calling",
		OK:         false,
		DurationMS: durationMS,
		Error:      detail,
	}
}

func summarizeBenchmarkDetail(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	const maxLen = 140
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen-1]) + "…"
}

func normalizeAgentSettingsProvider(provider string) string {
	value := strings.ToLower(strings.TrimSpace(provider))
	switch value {
	case "", "openai_compatible":
		return "openai"
	default:
		return value
	}
}

func fetchOpenAICompatibleModels(ctx context.Context, endpoint string, apiKey string) ([]string, error) {
	modelsURL, err := normalizeOpenAICompatibleModelsURL(endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("model lookup failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("model lookup failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return nil, fmt.Errorf("model lookup failed: %s", msg)
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid models response")
	}

	seen := make(map[string]struct{}, len(payload.Data))
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		models = append(models, id)
	}
	sort.Strings(models)
	return models, nil
}

func normalizeOpenAICompatibleModelsURL(endpoint string) (string, error) {
	base := strings.TrimSpace(endpoint)
	if base == "" {
		base = "https://api.openai.com"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid api base")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid api base")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return "", fmt.Errorf("invalid api base")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	switch {
	case strings.HasSuffix(parsed.Path, "/models"):
	case strings.HasSuffix(parsed.Path, "/v1"):
		parsed.Path += "/models"
	default:
		parsed.Path += "/v1/models"
	}
	return parsed.String(), nil
}

func loadYAMLDocument(configPath string) (*yaml.Node, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return newEmptyYAMLDocument(), nil
		}
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return newEmptyYAMLDocument(), nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid config yaml: %w", err)
	}
	if len(doc.Content) == 0 {
		doc.Content = []*yaml.Node{{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}}
	}
	return &doc, nil
}

func newEmptyYAMLDocument() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}},
	}
}

func readExpandedConsoleConfig(v *viper.Viper, configPath string) error {
	if v == nil {
		return fmt.Errorf("config reader is nil")
	}
	return configutil.ReadExpandedConfig(v, configPath, nil)
}

func isInvalidConfigYAMLError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "invalid config yaml")
}

func documentMapping(doc *yaml.Node) (*yaml.Node, error) {
	if doc == nil {
		return nil, fmt.Errorf("config document is nil")
	}
	if doc.Kind == yaml.DocumentNode {
		if len(doc.Content) == 0 {
			doc.Content = []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}
		}
		doc = doc.Content[0]
	}
	if doc.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("config root must be a yaml mapping")
	}
	return doc, nil
}

func marshalYAMLDocument(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		_ = enc.Close()
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if strings.EqualFold(strings.TrimSpace(node.Content[i].Value), key) {
			return node.Content[i+1]
		}
	}
	return nil
}

func ensureMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	if value := findMappingValue(node, key); value != nil {
		if value.Kind == yaml.MappingNode {
			return value
		}
		*value = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		return value
	}
	child := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		child,
	)
	return child
}

func setOrDeleteMappingScalar(node *yaml.Node, key, value string) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	value = strings.TrimSpace(value)
	for i := 0; i+1 < len(node.Content); i += 2 {
		if !strings.EqualFold(strings.TrimSpace(node.Content[i].Value), key) {
			continue
		}
		if value == "" {
			node.Content = append(node.Content[:i], node.Content[i+2:]...)
			return
		}
		node.Content[i+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
		return
	}
	if value == "" {
		return
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
	)
}

func deleteMappingKey(node *yaml.Node, key string) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if !strings.EqualFold(strings.TrimSpace(node.Content[i].Value), key) {
			continue
		}
		node.Content = append(node.Content[:i], node.Content[i+2:]...)
		return
	}
}

func setMappingBoolPath(node *yaml.Node, section, key string, value bool) {
	sectionNode := ensureMappingValue(node, section)
	if sectionNode == nil {
		return
	}
	for i := 0; i+1 < len(sectionNode.Content); i += 2 {
		if !strings.EqualFold(strings.TrimSpace(sectionNode.Content[i].Value), key) {
			continue
		}
		sectionNode.Content[i+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: boolString(value)}
		return
	}
	sectionNode.Content = append(sectionNode.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: boolString(value)},
	)
}

func setMappingStringList(node *yaml.Node, key string, values []string) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	list := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(strings.ToLower(raw))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		list.Content = append(list.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if !strings.EqualFold(strings.TrimSpace(node.Content[i].Value), key) {
			continue
		}
		node.Content[i+1] = list
		return
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		list,
	)
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func readAgentSettingsFromReader(r interface {
	GetString(string) string
	GetStringSlice(string) []string
	GetBool(string) bool
}) agentSettingsPayload {
	if r == nil {
		return agentSettingsPayload{}
	}
	return agentSettingsPayload{
		LLM: llmSettingsPayload{
			Provider:            strings.TrimSpace(r.GetString("llm.provider")),
			Endpoint:            strings.TrimSpace(r.GetString("llm.endpoint")),
			Model:               strings.TrimSpace(r.GetString("llm.model")),
			APIKey:              strings.TrimSpace(r.GetString("llm.api_key")),
			CloudflareAccountID: strings.TrimSpace(r.GetString("llm.cloudflare.account_id")),
			ReasoningEffort:     strings.TrimSpace(r.GetString("llm.reasoning_effort")),
			ToolsEmulationMode:  strings.TrimSpace(r.GetString("llm.tools_emulation_mode")),
		},
		Multimodal: multimodalSettingsPayload{
			ImageSources: sanitizeMultimodalSources(r.GetStringSlice("multimodal.image.sources")),
		},
		Tools: toolsSettingsPayload{
			WriteFileEnabled:    r.GetBool("tools.write_file.enabled"),
			ContactsSendEnabled: r.GetBool("tools.contacts_send.enabled"),
			TodoUpdateEnabled:   r.GetBool("tools.todo_update.enabled"),
			PlanCreateEnabled:   r.GetBool("tools.plan_create.enabled"),
			URLFetchEnabled:     r.GetBool("tools.url_fetch.enabled"),
			WebSearchEnabled:    r.GetBool("tools.web_search.enabled"),
			BashEnabled:         r.GetBool("tools.bash.enabled"),
		},
	}
}

func sanitizeMultimodalSources(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(supportedMultimodalSources))
	for _, value := range supportedMultimodalSources {
		allowed[value] = struct{}{}
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(strings.ToLower(raw))
		if value == "" {
			continue
		}
		if _, ok := allowed[value]; !ok {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
