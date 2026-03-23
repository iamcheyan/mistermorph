package consolecmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestReadAgentSettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(
		"llm:\n  provider: openai\n  model: gpt-5.2\n  reasoning_effort: high\n  cloudflare:\n    account_id: acc-123\n"+
			"multimodal:\n  image:\n    sources: [telegram, line]\n"+
			"tools:\n  bash:\n    enabled: false\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := readAgentSettings(configPath)
	if err != nil {
		t.Fatalf("readAgentSettings() error = %v", err)
	}
	if got.LLM.Provider != "openai" || got.LLM.Model != "gpt-5.2" || got.LLM.ReasoningEffort != "high" {
		t.Fatalf("got.LLM = %+v", got.LLM)
	}
	if got.LLM.CloudflareAccountID != "acc-123" {
		t.Fatalf("got.LLM.CloudflareAccountID = %q, want acc-123", got.LLM.CloudflareAccountID)
	}
	if len(got.Multimodal.ImageSources) != 2 || got.Multimodal.ImageSources[0] != "telegram" || got.Multimodal.ImageSources[1] != "line" {
		t.Fatalf("got.Multimodal = %+v", got.Multimodal)
	}
	if !got.Tools.WriteFileEnabled || !got.Tools.ContactsSendEnabled || !got.Tools.TodoUpdateEnabled ||
		!got.Tools.PlanCreateEnabled || !got.Tools.URLFetchEnabled || !got.Tools.WebSearchEnabled {
		t.Fatalf("got.Tools defaults not applied: %+v", got.Tools)
	}
	if got.Tools.BashEnabled {
		t.Fatalf("got.Tools.BashEnabled = true, want false")
	}
}

func TestWriteAgentSettingsPreservesOtherConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(
		"console:\n  listen: 127.0.0.1:9080\n"+
			"llm:\n  provider: openai\n  model: gpt-5.2\n"+
			"tools:\n  url_fetch:\n    timeout: 30s\n    enabled: true\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	serialized, err := writeAgentSettings(configPath, agentSettingsPayload{
		LLM: llmSettingsPayload{
			Provider:            "anthropic",
			Model:               "claude-3-7-sonnet",
			CloudflareAccountID: "acc-next",
			ToolsEmulationMode:  "fallback",
		},
		Multimodal: multimodalSettingsPayload{
			ImageSources: []string{"telegram", "remote_download"},
		},
		Tools: toolsSettingsPayload{
			WriteFileEnabled:    true,
			ContactsSendEnabled: true,
			TodoUpdateEnabled:   true,
			PlanCreateEnabled:   false,
			URLFetchEnabled:     true,
			WebSearchEnabled:    false,
			BashEnabled:         true,
		},
	})
	if err != nil {
		t.Fatalf("writeAgentSettings() error = %v", err)
	}
	out := string(serialized)
	if !strings.Contains(out, "console:") || !strings.Contains(out, "listen: 127.0.0.1:9080") {
		t.Fatalf("serialized config lost console block: %s", out)
	}
	if !strings.Contains(out, "provider: anthropic") || !strings.Contains(out, "tools_emulation_mode: fallback") {
		t.Fatalf("serialized config missing updated llm block: %s", out)
	}
	if strings.Contains(out, "\n  cloudflare:\n") || strings.Contains(out, "account_id: acc-next") {
		t.Fatalf("serialized config should prune cloudflare block for non-cloudflare provider: %s", out)
	}
	if !strings.Contains(out, "- remote_download") {
		t.Fatalf("serialized config missing multimodal sources: %s", out)
	}
	if !strings.Contains(out, "timeout: 30s") {
		t.Fatalf("serialized config lost existing tool config: %s", out)
	}
}

func TestHandleAgentSettingsPut(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(
		"llm:\n  provider: openai\n  model: gpt-5.2\n"+
			"multimodal:\n  image:\n    sources: [telegram]\n"+
			"tools:\n  bash:\n    enabled: true\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	prevConfig, hadConfig := viper.Get("config"), viper.IsSet("config")
	prevLLM, hadLLM := viper.Get("llm"), viper.IsSet("llm")
	prevMM, hadMM := viper.Get("multimodal"), viper.IsSet("multimodal")
	prevTools, hadTools := viper.Get("tools"), viper.IsSet("tools")
	viper.Set("config", configPath)
	t.Cleanup(func() {
		if hadConfig {
			viper.Set("config", prevConfig)
		} else {
			viper.Set("config", nil)
		}
		if hadLLM {
			viper.Set("llm", prevLLM)
		} else {
			viper.Set("llm", nil)
		}
		if hadMM {
			viper.Set("multimodal", prevMM)
		} else {
			viper.Set("multimodal", nil)
		}
		if hadTools {
			viper.Set("tools", prevTools)
		} else {
			viper.Set("tools", nil)
		}
	})

	body := bytes.NewBufferString(`{
		"llm":{"provider":"anthropic","model":"claude-3-7-sonnet","api_key":"${ANTHROPIC_API_KEY}","cloudflare_account_id":"acc-live","reasoning_effort":"high","tools_emulation_mode":"fallback"},
		"multimodal":{"image_sources":["telegram","remote_download"]},
		"tools":{"write_file_enabled":true,"contacts_send_enabled":false,"todo_update_enabled":true,"plan_create_enabled":false,"url_fetch_enabled":true,"web_search_enabled":false,"bash_enabled":false}
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/settings/agent", body)
	rec := httptest.NewRecorder()

	(&server{}).handleAgentSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(raw), "provider: anthropic") || !strings.Contains(string(raw), "reasoning_effort: high") {
		t.Fatalf("config missing updated llm settings: %s", string(raw))
	}
	if strings.Contains(string(raw), "\n  cloudflare:\n") || strings.Contains(string(raw), "account_id: acc-live") {
		t.Fatalf("config should prune cloudflare block for non-cloudflare provider: %s", string(raw))
	}
	if !strings.Contains(string(raw), "- remote_download") {
		t.Fatalf("config missing multimodal update: %s", string(raw))
	}
	if got := viper.GetString("llm.provider"); got != "anthropic" {
		t.Fatalf("viper llm.provider = %q, want anthropic", got)
	}
	if !viper.GetBool("tools.write_file.enabled") || viper.GetBool("tools.bash.enabled") {
		t.Fatalf("viper tools not updated: write_file=%v bash=%v", viper.GetBool("tools.write_file.enabled"), viper.GetBool("tools.bash.enabled"))
	}

	var payload struct {
		OK         bool                      `json:"ok"`
		LLM        llmSettingsPayload        `json:"llm"`
		Multimodal multimodalSettingsPayload `json:"multimodal"`
		Tools      toolsSettingsPayload      `json:"tools"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !payload.OK {
		t.Fatalf("payload.OK = false, want true")
	}
	if payload.LLM.Provider != "anthropic" || payload.LLM.ReasoningEffort != "high" {
		t.Fatalf("payload.LLM = %+v", payload.LLM)
	}
	if payload.LLM.CloudflareAccountID != "" {
		t.Fatalf("payload.LLM.CloudflareAccountID = %q, want empty", payload.LLM.CloudflareAccountID)
	}
	if len(payload.Multimodal.ImageSources) != 2 {
		t.Fatalf("payload.Multimodal = %+v", payload.Multimodal)
	}
	if payload.Tools.BashEnabled {
		t.Fatalf("payload.Tools.BashEnabled = true, want false")
	}
}

func TestWriteAgentSettingsKeepsCloudflareBlockForCloudflareProvider(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(
		"llm:\n  provider: openai\n  model: gpt-5.2\n  cloudflare:\n    api_token: ${CLOUDFLARE_API_TOKEN}\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	serialized, err := writeAgentSettings(configPath, agentSettingsPayload{
		LLM: llmSettingsPayload{
			Provider:            "cloudflare",
			Model:               "@cf/meta/llama-3.1-8b-instruct",
			CloudflareAccountID: "acc-live",
		},
	})
	if err != nil {
		t.Fatalf("writeAgentSettings() error = %v", err)
	}
	out := string(serialized)
	if !strings.Contains(out, "provider: cloudflare") || !strings.Contains(out, "account_id: acc-live") {
		t.Fatalf("serialized config missing cloudflare settings: %s", out)
	}
	if !strings.Contains(out, "api_token: ${CLOUDFLARE_API_TOKEN}") {
		t.Fatalf("serialized config should preserve cloudflare api_token: %s", out)
	}
}

func TestHandleAgentSettingsPutRejectsInvalidConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	prevConfig, hadConfig := viper.Get("config"), viper.IsSet("config")
	viper.Set("config", configPath)
	t.Cleanup(func() {
		if hadConfig {
			viper.Set("config", prevConfig)
		} else {
			viper.Set("config", nil)
		}
	})

	req := httptest.NewRequest(http.MethodPut, "/api/settings/agent", bytes.NewBufferString(`{"llm":{"provider":"not_a_provider"}}`))
	rec := httptest.NewRecorder()

	(&server{}).handleAgentSettings(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestHandleAgentSettingsGetFallsBackWhenConfigIsMalformed(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("llm: [\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	prevConfig, hadConfig := viper.Get("config"), viper.IsSet("config")
	prevLLM, hadLLM := viper.Get("llm"), viper.IsSet("llm")
	viper.Set("config", configPath)
	viper.Set("llm", map[string]any{
		"provider": "openai",
		"endpoint": "https://api.openai.com",
		"model":    "gpt-5.2",
	})
	t.Cleanup(func() {
		if hadConfig {
			viper.Set("config", prevConfig)
		} else {
			viper.Set("config", nil)
		}
		if hadLLM {
			viper.Set("llm", prevLLM)
		} else {
			viper.Set("llm", nil)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/settings/agent", nil)
	rec := httptest.NewRecorder()

	(&server{}).handleAgentSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "\"provider\":\"openai\"") {
		t.Fatalf("response should fall back to current viper defaults: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"config_source":"defaults"`) || !strings.Contains(rec.Body.String(), `"config_valid":false`) {
		t.Fatalf("response should expose config source metadata: %s", rec.Body.String())
	}
}

func TestWriteAgentSettingsRepairsMalformedConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("llm: [\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	serialized, err := writeAgentSettings(configPath, agentSettingsPayload{
		LLM: llmSettingsPayload{
			Provider: "openai",
			Model:    "gpt-5.2",
			APIKey:   "sk-test",
		},
	})
	if err != nil {
		t.Fatalf("writeAgentSettings() error = %v", err)
	}
	out := string(serialized)
	if strings.Contains(out, "llm: [") {
		t.Fatalf("serialized config should replace malformed yaml: %s", out)
	}
	if !strings.Contains(out, "provider: openai") || !strings.Contains(out, "model: gpt-5.2") {
		t.Fatalf("serialized config missing repaired llm block: %s", out)
	}
}

func TestHandleAgentSettingsModels(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v1/models" {
			t.Fatalf("path = %q, want /v1/models", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Fatalf("authorization = %q, want Bearer sk-test", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"gpt-5-mini"},{"id":"gpt-5"},{"id":"gpt-5-mini"}]}`))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/agent/models", bytes.NewBufferString(
		`{"endpoint":"`+upstream.URL+`","api_key":"sk-test"}`,
	))
	rec := httptest.NewRecorder()

	(&server{}).handleAgentSettingsModels(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); !strings.Contains(got, `"items":["gpt-5","gpt-5-mini"]`) {
		t.Fatalf("response missing sorted model ids: %s", got)
	}
}

func TestHandleAgentSettingsTest(t *testing.T) {
	prev := runAgentSettingsConnectionTest
	runAgentSettingsConnectionTest = func(_ context.Context, settings llmSettingsPayload) (agentSettingsTestResult, error) {
		if settings.Provider != "openai" {
			t.Fatalf("provider = %q, want openai", settings.Provider)
		}
		if settings.Model != "gpt-5" {
			t.Fatalf("model = %q, want gpt-5", settings.Model)
		}
		return agentSettingsTestResult{
			Provider: "openai",
			Model:    "gpt-5",
			Benchmarks: []agentSettingsBenchmarkResult{
				{ID: "text_reply", OK: true, DurationMS: 912, Detail: "OK"},
				{ID: "json_response", OK: true, DurationMS: 1044, Detail: "json ok"},
				{ID: "tool_calling", OK: false, DurationMS: 1350, Error: "model replied without calling the tool"},
			},
		}, nil
	}
	t.Cleanup(func() {
		runAgentSettingsConnectionTest = prev
	})

	req := httptest.NewRequest(http.MethodPost, "/api/settings/agent/test", bytes.NewBufferString(
		`{"llm":{"provider":"openai","model":"gpt-5","api_key":"sk-test"}}`,
	))
	rec := httptest.NewRecorder()

	(&server{}).handleAgentSettingsTest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Body.String(); !strings.Contains(got, `"benchmarks":[`) || !strings.Contains(got, `"id":"text_reply"`) || !strings.Contains(got, `"id":"json_response"`) || !strings.Contains(got, `"id":"tool_calling"`) {
		t.Fatalf("response missing test result fields: %s", got)
	}
}

func TestNormalizeAgentSettingsProvider(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "openai"},
		{name: "openai compatible", in: "openai_compatible", want: "openai"},
		{name: "cloudflare", in: "cloudflare", want: "cloudflare"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeAgentSettingsProvider(tc.in); got != tc.want {
				t.Fatalf("normalizeAgentSettingsProvider(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
