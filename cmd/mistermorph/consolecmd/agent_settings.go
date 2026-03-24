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
	"regexp"
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

var benchmarkErrorStatusPattern = regexp.MustCompile(`(?is)\bstatus\s+\d{3}\s*:\s*(.+)$`)

type llmSettingsPayload struct {
	Provider            string `json:"provider"`
	Endpoint            string `json:"endpoint"`
	Model               string `json:"model"`
	APIKey              string `json:"api_key"`
	CloudflareAPIToken  string `json:"cloudflare_api_token"`
	CloudflareAccountID string `json:"cloudflare_account_id"`
	ReasoningEffort     string `json:"reasoning_effort"`
	ToolsEmulationMode  string `json:"tools_emulation_mode"`
}

type llmSettingsUpdatePayload struct {
	Provider            *string `json:"provider,omitempty"`
	Endpoint            *string `json:"endpoint,omitempty"`
	Model               *string `json:"model,omitempty"`
	APIKey              *string `json:"api_key,omitempty"`
	CloudflareAPIToken  *string `json:"cloudflare_api_token,omitempty"`
	CloudflareAccountID *string `json:"cloudflare_account_id,omitempty"`
	ReasoningEffort     *string `json:"reasoning_effort,omitempty"`
	ToolsEmulationMode  *string `json:"tools_emulation_mode,omitempty"`
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

type agentSettingsUpdatePayload struct {
	LLM        llmSettingsUpdatePayload  `json:"llm"`
	Multimodal multimodalSettingsPayload `json:"multimodal"`
	Tools      toolsSettingsPayload      `json:"tools"`
}

type agentSettingsEnvManagedField struct {
	EnvName string `json:"env_name"`
	Value   string `json:"value,omitempty"`
}

type agentSettingsEnvManagedPayload struct {
	LLM map[string]agentSettingsEnvManagedField `json:"llm,omitempty"`
}

type agentSettingsModelsRequest struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
}

type agentSettingsTestRequest struct {
	LLM llmSettingsPayload `json:"llm"`
}

type agentSettingsBenchmarkResult struct {
	ID          string `json:"id"`
	OK          bool   `json:"ok"`
	DurationMS  int64  `json:"duration_ms"`
	Detail      string `json:"detail,omitempty"`
	Error       string `json:"error,omitempty"`
	RawResponse string `json:"raw_response,omitempty"`
}

type agentSettingsTestResult struct {
	Provider   string
	Model      string
	Benchmarks []agentSettingsBenchmarkResult
}

type agentSettingsConnectionTestOptions struct {
	InspectPrompt  bool
	InspectRequest bool
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
		settings = defaultAgentSettingsPayload()
		configSource = "defaults"
		configValid = false
	}
	effectiveLLM := settingsFromCurrentRuntime()
	writeJSON(w, http.StatusOK, map[string]any{
		"llm":           settings.LLM,
		"env_managed":   currentAgentSettingsEnvManaged(effectiveLLM.Provider),
		"multimodal":    settings.Multimodal,
		"tools":         settings.Tools,
		"config_path":   configPath,
		"config_exists": configExists,
		"config_valid":  configValid,
		"config_source": configSource,
	})
}

func (s *server) handleAgentSettingsPut(w http.ResponseWriter, r *http.Request) {
	var req agentSettingsUpdatePayload
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	configPath, err := resolveConsoleConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	serialized, err := writeAgentSettingsUpdate(configPath, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	effectiveLLM := resolveAgentSettingsLLM(req.LLM)
	tmp, err := validateAgentConfigDocument(serialized, effectiveLLM)
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

	viper.Set(llmSettingsKey, llmSettingsPayloadToMap(effectiveLLM))
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
	if doc, docErr := loadYAMLDocumentBytes(serialized); docErr == nil {
		next = normalizeAgentSettingsConfigView(next, doc)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"llm":           next.LLM,
		"env_managed":   currentAgentSettingsEnvManaged(effectiveLLM.Provider),
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
	current := settingsFromCurrentRuntime()
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(current.Endpoint)
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(current.APIKey)
	}
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "api key is required")
		return
	}
	models, err := fetchOpenAICompatibleModels(r.Context(), endpoint, apiKey)
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

	result, err := runAgentSettingsConnectionTest(
		r.Context(),
		resolveAgentSettingsLLM(llmSettingsPayloadAsNonEmptyUpdate(req.LLM)),
		agentSettingsConnectionTestOptions{
			InspectPrompt:  s != nil && s.cfg.inspectPrompt,
			InspectRequest: s != nil && s.cfg.inspectRequest,
		},
	)
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
			return defaultAgentSettingsPayload(), nil
		}
		return agentSettingsPayload{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return defaultAgentSettingsPayload(), nil
	}
	tmp := viper.New()
	integration.ApplyViperDefaults(tmp)
	if err := readExpandedConsoleConfig(tmp, configPath); err != nil {
		return agentSettingsPayload{}, fmt.Errorf("invalid config yaml: %w", err)
	}
	settings := readAgentSettingsFromReader(tmp)
	doc, err := loadYAMLDocument(configPath)
	if err != nil {
		return agentSettingsPayload{}, err
	}
	return normalizeAgentSettingsConfigView(settings, doc), nil
}

func defaultAgentSettingsPayload() agentSettingsPayload {
	tmp := viper.New()
	integration.ApplyViperDefaults(tmp)
	settings := readAgentSettingsFromReader(tmp)
	settings.LLM.Endpoint = ""
	settings.LLM.Model = ""
	return settings
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
	return writeAgentSettingsUpdate(configPath, agentSettingsUpdatePayload{
		LLM:        llmSettingsPayloadAsUpdate(values.LLM),
		Multimodal: values.Multimodal,
		Tools:      values.Tools,
	})
}

func writeAgentSettingsUpdate(configPath string, values agentSettingsUpdatePayload) ([]byte, error) {
	doc, err := loadYAMLDocument(configPath)
	if err != nil {
		if !isInvalidConfigYAMLError(err) {
			return nil, err
		}
		doc = newEmptyYAMLDocument()
	}
	current := defaultAgentSettingsPayload()
	if existing, readErr := readAgentSettings(configPath); readErr == nil {
		current = existing
	} else if !isInvalidConfigYAMLError(readErr) && !os.IsNotExist(readErr) {
		return nil, readErr
	}
	nextLLM := applyLLMSettingsUpdate(current.LLM, values.LLM)
	root, err := documentMapping(doc)
	if err != nil {
		return nil, err
	}

	llmNode := ensureMappingValue(root, llmSettingsKey)
	if values.LLM.Provider != nil {
		setOrDeleteMappingScalar(llmNode, "provider", *values.LLM.Provider)
	}
	if values.LLM.Endpoint != nil {
		setOrDeleteMappingScalar(llmNode, "endpoint", *values.LLM.Endpoint)
	}
	if values.LLM.Model != nil {
		setOrDeleteMappingScalar(llmNode, "model", *values.LLM.Model)
	}
	if strings.EqualFold(strings.TrimSpace(nextLLM.Provider), "cloudflare") {
		setOrDeleteMappingScalar(llmNode, "api_key", "")
		cloudflareNode := ensureMappingValue(llmNode, "cloudflare")
		if values.LLM.CloudflareAccountID != nil {
			setOrDeleteMappingScalar(cloudflareNode, "account_id", *values.LLM.CloudflareAccountID)
		}
		if values.LLM.CloudflareAPIToken != nil {
			setOrDeleteMappingScalar(cloudflareNode, "api_token", *values.LLM.CloudflareAPIToken)
		}
	} else {
		if values.LLM.APIKey != nil {
			setOrDeleteMappingScalar(llmNode, "api_key", *values.LLM.APIKey)
		}
		deleteMappingKey(llmNode, "cloudflare")
	}
	if values.LLM.ReasoningEffort != nil {
		setOrDeleteMappingScalar(llmNode, "reasoning_effort", *values.LLM.ReasoningEffort)
	}
	if values.LLM.ToolsEmulationMode != nil {
		setOrDeleteMappingScalar(llmNode, "tools_emulation_mode", *values.LLM.ToolsEmulationMode)
	}

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

func validateAgentConfigDocument(data []byte, effectiveLLM llmSettingsPayload) (*viper.Viper, error) {
	tmp := viper.New()
	integration.ApplyViperDefaults(tmp)
	tmp.SetConfigType("yaml")
	if err := tmp.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("invalid config yaml: %w", err)
	}
	values := llmutil.RuntimeValuesFromReader(tmp)
	values.Provider = firstNonEmpty(strings.TrimSpace(effectiveLLM.Provider), values.Provider)
	values.Endpoint = firstNonEmpty(strings.TrimSpace(effectiveLLM.Endpoint), values.Endpoint)
	values.APIKey = firstNonEmpty(strings.TrimSpace(effectiveLLM.APIKey), values.APIKey)
	values.Model = firstNonEmpty(strings.TrimSpace(effectiveLLM.Model), values.Model)
	values.CloudflareAPIToken = firstNonEmpty(strings.TrimSpace(effectiveLLM.CloudflareAPIToken), values.CloudflareAPIToken)
	values.CloudflareAccountID = firstNonEmpty(strings.TrimSpace(effectiveLLM.CloudflareAccountID), values.CloudflareAccountID)
	values.ReasoningEffortRaw = firstNonEmpty(strings.TrimSpace(effectiveLLM.ReasoningEffort), values.ReasoningEffortRaw)
	values.ToolsEmulationMode = firstNonEmpty(strings.TrimSpace(effectiveLLM.ToolsEmulationMode), values.ToolsEmulationMode)
	route, err := llmutil.ResolveRoute(values, llmutil.RoutePurposeMainLoop)
	if err != nil {
		return nil, err
	}
	if _, err := llmutil.ClientFromConfigWithValues(route.ClientConfig, route.Values); err != nil {
		return nil, err
	}
	return tmp, nil
}

func settingsFromCurrentRuntime() llmSettingsPayload {
	values := currentConsoleLLMRuntimeValues()
	provider := strings.TrimSpace(values.Provider)
	cloudflareAPIToken := strings.TrimSpace(values.CloudflareAPIToken)
	if cloudflareAPIToken == "" && strings.EqualFold(provider, "cloudflare") {
		cloudflareAPIToken = llmutil.APIKeyForProviderWithValues(provider, values)
	}
	return llmSettingsPayload{
		Provider:            provider,
		Endpoint:            llmutil.EndpointForProviderWithValues(provider, values),
		Model:               llmutil.ModelForProviderWithValues(provider, values),
		APIKey:              strings.TrimSpace(values.APIKey),
		CloudflareAPIToken:  cloudflareAPIToken,
		CloudflareAccountID: strings.TrimSpace(values.CloudflareAccountID),
		ReasoningEffort:     strings.TrimSpace(values.ReasoningEffortRaw),
		ToolsEmulationMode:  strings.TrimSpace(values.ToolsEmulationMode),
	}
}

func resolveAgentSettingsLLM(overrides llmSettingsUpdatePayload) llmSettingsPayload {
	return applyLLMSettingsUpdate(settingsFromCurrentRuntime(), overrides)
}

func currentConsoleLLMRuntimeValues() llmutil.RuntimeValues {
	values := llmutil.RuntimeValuesFromViper()

	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_PROVIDER"); ok {
		values.Provider = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_ENDPOINT"); ok {
		values.Endpoint = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_API_KEY"); ok {
		values.APIKey = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_MODEL"); ok {
		values.Model = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_AZURE_DEPLOYMENT"); ok {
		values.AzureDeployment = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_REASONING_EFFORT"); ok {
		values.ReasoningEffortRaw = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_TOOLS_EMULATION_MODE"); ok {
		values.ToolsEmulationMode = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_CLOUDFLARE_ACCOUNT_ID"); ok {
		values.CloudflareAccountID = strings.TrimSpace(value)
	}
	if _, value, ok := firstManagedEnv("MISTER_MORPH_LLM_CLOUDFLARE_API_TOKEN"); ok {
		values.CloudflareAPIToken = strings.TrimSpace(value)
	}

	return values
}

func applyLLMSettingsUpdate(current llmSettingsPayload, incoming llmSettingsUpdatePayload) llmSettingsPayload {
	merged := current
	if incoming.Provider != nil {
		merged.Provider = strings.TrimSpace(*incoming.Provider)
	}
	if incoming.Endpoint != nil {
		merged.Endpoint = strings.TrimSpace(*incoming.Endpoint)
	}
	if incoming.Model != nil {
		merged.Model = strings.TrimSpace(*incoming.Model)
	}
	if incoming.APIKey != nil {
		merged.APIKey = strings.TrimSpace(*incoming.APIKey)
	}
	if incoming.CloudflareAPIToken != nil {
		merged.CloudflareAPIToken = strings.TrimSpace(*incoming.CloudflareAPIToken)
	}
	if incoming.CloudflareAccountID != nil {
		merged.CloudflareAccountID = strings.TrimSpace(*incoming.CloudflareAccountID)
	}
	if incoming.ReasoningEffort != nil {
		merged.ReasoningEffort = strings.TrimSpace(*incoming.ReasoningEffort)
	}
	if incoming.ToolsEmulationMode != nil {
		merged.ToolsEmulationMode = strings.TrimSpace(*incoming.ToolsEmulationMode)
	}
	if strings.EqualFold(strings.TrimSpace(merged.Provider), "cloudflare") {
		merged.APIKey = ""
	} else {
		merged.CloudflareAPIToken = ""
		merged.CloudflareAccountID = ""
	}
	return merged
}

func llmSettingsPayloadAsUpdate(values llmSettingsPayload) llmSettingsUpdatePayload {
	return llmSettingsUpdatePayload{
		Provider:            stringPointer(values.Provider),
		Endpoint:            stringPointer(values.Endpoint),
		Model:               stringPointer(values.Model),
		APIKey:              stringPointer(values.APIKey),
		CloudflareAPIToken:  stringPointer(values.CloudflareAPIToken),
		CloudflareAccountID: stringPointer(values.CloudflareAccountID),
		ReasoningEffort:     stringPointer(values.ReasoningEffort),
		ToolsEmulationMode:  stringPointer(values.ToolsEmulationMode),
	}
}

func llmSettingsPayloadAsNonEmptyUpdate(values llmSettingsPayload) llmSettingsUpdatePayload {
	update := llmSettingsUpdatePayload{}
	if value := strings.TrimSpace(values.Provider); value != "" {
		update.Provider = stringPointer(value)
	}
	if value := strings.TrimSpace(values.Endpoint); value != "" {
		update.Endpoint = stringPointer(value)
	}
	if value := strings.TrimSpace(values.Model); value != "" {
		update.Model = stringPointer(value)
	}
	if value := strings.TrimSpace(values.APIKey); value != "" {
		update.APIKey = stringPointer(value)
	}
	if value := strings.TrimSpace(values.CloudflareAPIToken); value != "" {
		update.CloudflareAPIToken = stringPointer(value)
	}
	if value := strings.TrimSpace(values.CloudflareAccountID); value != "" {
		update.CloudflareAccountID = stringPointer(value)
	}
	if value := strings.TrimSpace(values.ReasoningEffort); value != "" {
		update.ReasoningEffort = stringPointer(value)
	}
	if value := strings.TrimSpace(values.ToolsEmulationMode); value != "" {
		update.ToolsEmulationMode = stringPointer(value)
	}
	return update
}

func stringPointer(value string) *string {
	next := value
	return &next
}

func llmSettingsPayloadToMap(values llmSettingsPayload) map[string]any {
	out := map[string]any{
		"provider":             strings.TrimSpace(values.Provider),
		"endpoint":             strings.TrimSpace(values.Endpoint),
		"model":                strings.TrimSpace(values.Model),
		"reasoning_effort":     strings.TrimSpace(values.ReasoningEffort),
		"tools_emulation_mode": strings.TrimSpace(values.ToolsEmulationMode),
	}
	if !strings.EqualFold(strings.TrimSpace(values.Provider), "cloudflare") {
		out["api_key"] = strings.TrimSpace(values.APIKey)
	}
	if strings.EqualFold(strings.TrimSpace(values.Provider), "cloudflare") &&
		(strings.TrimSpace(values.CloudflareAccountID) != "" || strings.TrimSpace(values.CloudflareAPIToken) != "") {
		out["cloudflare"] = map[string]any{}
		if accountID := strings.TrimSpace(values.CloudflareAccountID); accountID != "" {
			out["cloudflare"].(map[string]any)["account_id"] = accountID
		}
		if apiToken := strings.TrimSpace(values.CloudflareAPIToken); apiToken != "" {
			out["cloudflare"].(map[string]any)["api_token"] = apiToken
		}
	}
	return out
}

func defaultAgentSettingsConnectionTest(ctx context.Context, settings llmSettingsPayload, opts agentSettingsConnectionTestOptions) (agentSettingsTestResult, error) {
	values := llmutil.RuntimeValues{
		Provider:            normalizeAgentSettingsProvider(settings.Provider),
		Endpoint:            strings.TrimSpace(settings.Endpoint),
		APIKey:              strings.TrimSpace(settings.APIKey),
		Model:               strings.TrimSpace(settings.Model),
		RequestTimeoutRaw:   "20s",
		ReasoningEffortRaw:  strings.TrimSpace(settings.ReasoningEffort),
		ToolsEmulationMode:  strings.TrimSpace(settings.ToolsEmulationMode),
		CloudflareAPIToken:  strings.TrimSpace(settings.CloudflareAPIToken),
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
	inspectors, err := newConsoleInspectors(opts.InspectPrompt, opts.InspectRequest, "console_settings_test", "settings_test", "20060102_150405.000000000")
	if err != nil {
		return agentSettingsTestResult{}, err
	}
	defer func() {
		if inspectors != nil {
			_ = inspectors.Close()
		}
	}()
	client = inspectors.Wrap(client, route)

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
		Scene: "console.settings_test.text_reply",
		Messages: []llm.Message{
			{Role: "system", Content: "You're acting the linux cmd `echo`, will echo back the text."},
			{Role: "user", Content: "OK"},
		},
		Parameters: map[string]any{
			"max_tokens": 1024,
		},
	})
	durationMS := time.Since(start).Milliseconds()
	if err != nil {
		return agentSettingsBenchmarkResult{
			ID:          "text_reply",
			OK:          false,
			DurationMS:  durationMS,
			Error:       strings.TrimSpace(err.Error()),
			RawResponse: benchmarkRawResponseFromError(err),
		}
	}

	text := strings.TrimSpace(result.Text)
	if text == "" {
		return agentSettingsBenchmarkResult{
			ID:          "text_reply",
			OK:          false,
			DurationMS:  durationMS,
			Error:       "received an empty text reply",
			RawResponse: benchmarkRawResponse(result),
		}
	}

	return agentSettingsBenchmarkResult{
		ID:          "text_reply",
		OK:          true,
		DurationMS:  durationMS,
		Detail:      summarizeBenchmarkDetail(text),
		RawResponse: benchmarkRawResponse(result),
	}
}

func runAgentSettingsJSONBenchmark(ctx context.Context, client llm.Client, model string) agentSettingsBenchmarkResult {
	start := time.Now()
	result, err := client.Chat(ctx, llm.Request{
		Model:     model,
		Scene:     "console.settings_test.json_response",
		ForceJSON: true,
		Messages: []llm.Message{
			{Role: "system", Content: "You wrap the input by a JSON object and echo back the JSON object only. for example, IF input is `Hello` THEN return {\"message\": \"Hello\"}."},
			{Role: "user", Content: `Hello`},
		},
		Parameters: map[string]any{
			"max_tokens": 1024,
		},
	})
	durationMS := time.Since(start).Milliseconds()
	if err != nil {
		return agentSettingsBenchmarkResult{
			ID:          "json_response",
			OK:          false,
			DurationMS:  durationMS,
			Error:       strings.TrimSpace(err.Error()),
			RawResponse: benchmarkRawResponseFromError(err),
		}
	}

	var payload struct {
		Message string `json:"message"`
	}
	if err := jsonutil.DecodeWithFallback(result.Text, &payload); err != nil {
		return agentSettingsBenchmarkResult{
			ID:          "json_response",
			OK:          false,
			DurationMS:  durationMS,
			Error:       "response was not valid json",
			RawResponse: benchmarkRawResponse(result),
		}
	}
	if strings.TrimSpace(payload.Message) != "Hello" {
		return agentSettingsBenchmarkResult{
			ID:          "json_response",
			OK:          false,
			DurationMS:  durationMS,
			Error:       "json response is not so correct",
			RawResponse: benchmarkRawResponse(result),
		}
	}

	detail := summarizeBenchmarkDetail(strings.TrimSpace(payload.Message))
	if detail == "" {
		detail = "status=ok"
	}
	return agentSettingsBenchmarkResult{
		ID:          "json_response",
		OK:          true,
		DurationMS:  durationMS,
		Detail:      detail,
		RawResponse: benchmarkRawResponse(result),
	}
}

func runAgentSettingsToolCallingBenchmark(ctx context.Context, client llm.Client, model string) agentSettingsBenchmarkResult {
	start := time.Now()
	result, err := client.Chat(ctx, llm.Request{
		Model: model,
		Scene: "console.settings_test.tool_calling",
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
			"max_tokens": 1024,
		},
	})
	durationMS := time.Since(start).Milliseconds()
	if err != nil {
		return agentSettingsBenchmarkResult{
			ID:          "tool_calling",
			OK:          false,
			DurationMS:  durationMS,
			Error:       strings.TrimSpace(err.Error()),
			RawResponse: benchmarkRawResponseFromError(err),
		}
	}

	for _, call := range result.ToolCalls {
		if strings.EqualFold(strings.TrimSpace(call.Name), "ping") {
			return agentSettingsBenchmarkResult{
				ID:          "tool_calling",
				OK:          true,
				DurationMS:  durationMS,
				Detail:      "called ping",
				RawResponse: benchmarkRawResponse(result),
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
		ID:          "tool_calling",
		OK:          false,
		DurationMS:  durationMS,
		Error:       detail,
		RawResponse: benchmarkRawResponse(result),
	}
}

func benchmarkRawResponse(result llm.Result) string {
	text := strings.TrimSpace(result.Text)
	if len(result.ToolCalls) == 0 && result.JSON == nil {
		return text
	}

	payload := map[string]any{}
	if text != "" {
		payload["text"] = text
	}
	if result.JSON != nil {
		payload["json"] = result.JSON
	}
	if len(result.ToolCalls) > 0 {
		payload["tool_calls"] = result.ToolCalls
	}
	if len(payload) == 0 {
		return ""
	}

	serialized, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return text
	}
	return string(serialized)
}

func benchmarkRawResponseFromError(err error) string {
	if err == nil {
		return ""
	}

	text := strings.TrimSpace(err.Error())
	if text == "" {
		return ""
	}

	matches := benchmarkErrorStatusPattern.FindStringSubmatch(text)
	if len(matches) == 2 {
		if raw := strings.TrimSpace(matches[1]); raw != "" {
			return raw
		}
	}

	return text
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
	return loadYAMLDocumentBytes(data)
}

func loadYAMLDocumentBytes(data []byte) (*yaml.Node, error) {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeAgentSettingsConfigView(settings agentSettingsPayload, doc *yaml.Node) agentSettingsPayload {
	if !agentSettingsYAMLHasLLMKey(doc, "endpoint") {
		settings.LLM.Endpoint = ""
	}
	if !agentSettingsYAMLHasLLMKey(doc, "model") {
		settings.LLM.Model = ""
	}
	return settings
}

func agentSettingsYAMLHasLLMKey(doc *yaml.Node, key string) bool {
	root, err := documentMapping(doc)
	if err != nil {
		return false
	}
	llmNode := findMappingValue(root, llmSettingsKey)
	if llmNode == nil || llmNode.Kind != yaml.MappingNode {
		return false
	}
	return findMappingValue(llmNode, key) != nil
}

func readAgentSettingsFromReader(r interface {
	GetString(string) string
	GetStringSlice(string) []string
	GetBool(string) bool
}) agentSettingsPayload {
	if r == nil {
		return agentSettingsPayload{}
	}
	provider := strings.TrimSpace(r.GetString("llm.provider"))
	cloudflareAPIToken := strings.TrimSpace(r.GetString("llm.cloudflare.api_token"))
	if strings.EqualFold(provider, "cloudflare") {
		cloudflareAPIToken = firstNonEmpty(cloudflareAPIToken, strings.TrimSpace(r.GetString("llm.api_key")))
	}
	return agentSettingsPayload{
		LLM: llmSettingsPayload{
			Provider:            provider,
			Endpoint:            strings.TrimSpace(r.GetString("llm.endpoint")),
			Model:               strings.TrimSpace(r.GetString("llm.model")),
			APIKey:              strings.TrimSpace(r.GetString("llm.api_key")),
			CloudflareAPIToken:  cloudflareAPIToken,
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

func currentAgentSettingsEnvManaged(provider string) agentSettingsEnvManagedPayload {
	return agentSettingsEnvManagedPayload{
		LLM: currentAgentSettingsLLMEnvManaged(provider),
	}
}

func currentAgentSettingsLLMEnvManaged(provider string) map[string]agentSettingsEnvManagedField {
	fields := map[string]agentSettingsEnvManagedField{}
	normalizedProvider := strings.TrimSpace(strings.ToLower(provider))

	if field, ok := currentAgentSettingsManagedEnvField(false, "MISTER_MORPH_LLM_PROVIDER"); ok {
		fields["provider"] = field
	}
	if field, ok := currentAgentSettingsManagedEnvField(false, "MISTER_MORPH_LLM_ENDPOINT"); ok {
		fields["endpoint"] = field
	}
	if field, ok := currentAgentSettingsModelEnvField(provider); ok {
		fields["model"] = field
	}
	if normalizedProvider == "cloudflare" {
		if field, ok := currentAgentSettingsManagedEnvField(
			true,
			"MISTER_MORPH_LLM_CLOUDFLARE_API_TOKEN",
			"MISTER_MORPH_LLM_API_KEY",
		); ok {
			fields["cloudflare_api_token"] = field
		}
	} else {
		if field, ok := currentAgentSettingsManagedEnvField(true, "MISTER_MORPH_LLM_API_KEY"); ok {
			fields["api_key"] = field
		}
		if field, ok := currentAgentSettingsManagedEnvField(true, "MISTER_MORPH_LLM_CLOUDFLARE_API_TOKEN"); ok {
			fields["cloudflare_api_token"] = field
		}
	}
	if field, ok := currentAgentSettingsManagedEnvField(false, "MISTER_MORPH_LLM_CLOUDFLARE_ACCOUNT_ID"); ok {
		fields["cloudflare_account_id"] = field
	}
	if field, ok := currentAgentSettingsManagedEnvField(false, "MISTER_MORPH_LLM_REASONING_EFFORT"); ok {
		fields["reasoning_effort"] = field
	}
	if field, ok := currentAgentSettingsManagedEnvField(false, "MISTER_MORPH_LLM_TOOLS_EMULATION_MODE"); ok {
		fields["tools_emulation_mode"] = field
	}

	if len(fields) == 0 {
		return nil
	}
	return fields
}

func currentAgentSettingsModelEnvField(provider string) (agentSettingsEnvManagedField, bool) {
	if strings.EqualFold(strings.TrimSpace(provider), "azure") {
		return currentAgentSettingsManagedEnvField(
			false,
			"MISTER_MORPH_LLM_AZURE_DEPLOYMENT",
			"MISTER_MORPH_LLM_MODEL",
		)
	}
	return currentAgentSettingsManagedEnvField(false, "MISTER_MORPH_LLM_MODEL")
}

func currentAgentSettingsManagedEnvField(sensitive bool, names ...string) (agentSettingsEnvManagedField, bool) {
	name, value, ok := firstManagedEnv(names...)
	if !ok {
		return agentSettingsEnvManagedField{}, false
	}
	field := agentSettingsEnvManagedField{EnvName: name}
	if !sensitive {
		field.Value = strings.TrimSpace(value)
	}
	return field, true
}

func firstManagedEnvName(names ...string) (string, bool) {
	name, _, ok := firstManagedEnv(names...)
	return name, ok
}

func firstManagedEnv(names ...string) (string, string, bool) {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if value, ok := os.LookupEnv(name); ok {
			return name, value, true
		}
	}
	return "", "", false
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
