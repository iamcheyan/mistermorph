package daemonruntime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/internal/llmstats"
	"github.com/quailyquaily/mistermorph/internal/statepaths"
	"github.com/spf13/viper"
)

func TestLLMUsageStatsRoute(t *testing.T) {
	stateDir := t.TempDir()
	oldStateDir := viper.GetString("file_state_dir")
	t.Cleanup(func() {
		viper.Set("file_state_dir", oldStateDir)
	})
	viper.Set("file_state_dir", stateDir)

	journal := llmstats.NewJournal(statepaths.LLMUsageJournalDir(), llmstats.JournalOptions{})
	defer func() { _ = journal.Close() }()
	if _, err := journal.Append(llmstats.RequestRecord{
		TS:                       time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Provider:                 "openai",
		APIBase:                  "https://api.openai.com",
		Model:                    "gpt-5.2",
		InputTokens:              8,
		OutputTokens:             4,
		TotalTokens:              12,
		CachedInputTokens:        2,
		CacheCreationInputTokens: 1,
		CostCurrency:             "USD",
		CostEstimated:            true,
		CachedInputCost:          0.002,
		CacheCreationInputCost:   0.001,
		TotalCost:                0.015,
	}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, RoutesOptions{Mode: "serve", AuthToken: "token"})

	req := httptest.NewRequest(http.MethodGet, "/stats/llm/usage", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}

	var payload struct {
		Summary struct {
			Requests                 int64   `json:"requests"`
			TotalTokens              int64   `json:"total_tokens"`
			CachedInputTokens        int64   `json:"cached_input_tokens"`
			CacheCreationInputTokens int64   `json:"cache_creation_input_tokens"`
			CostCurrency             string  `json:"cost_currency"`
			TotalCost                float64 `json:"total_cost"`
		} `json:"summary"`
		APIHosts []struct {
			APIHost string `json:"api_host"`
		} `json:"api_hosts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload.Summary.Requests != 1 || payload.Summary.TotalTokens != 12 {
		t.Fatalf("summary = %+v", payload.Summary)
	}
	if payload.Summary.CachedInputTokens != 2 || payload.Summary.CacheCreationInputTokens != 1 {
		t.Fatalf("summary cache = %+v", payload.Summary)
	}
	if payload.Summary.CostCurrency != "USD" || payload.Summary.TotalCost < 0.014999 || payload.Summary.TotalCost > 0.015001 {
		t.Fatalf("summary cost = %+v", payload.Summary)
	}
	if len(payload.APIHosts) != 1 || payload.APIHosts[0].APIHost != "api.openai.com" {
		t.Fatalf("api_hosts = %+v", payload.APIHosts)
	}
}
