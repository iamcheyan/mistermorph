package llmstats

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/llm"
)

const testCostEpsilon = 1e-9

func costAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < testCostEpsilon
}

type stubUsageClient struct{}

func (stubUsageClient) Chat(ctx context.Context, req llm.Request) (llm.Result, error) {
	return llm.Result{
		Text: "ok",
		Usage: llm.Usage{
			InputTokens:  11,
			OutputTokens: 7,
			TotalTokens:  18,
			Cache: llm.UsageCache{
				CachedInputTokens:        5,
				CacheCreationInputTokens: 3,
				Details: map[string]int{
					"ephemeral_5m_input_tokens": 3,
				},
			},
			Cost: &llm.UsageCost{
				Currency:           "USD",
				Estimated:          true,
				Input:              0.01,
				CachedInput:        0.002,
				CacheCreationInput: 0.003,
				Output:             0.02,
				Total:              0.035,
			},
		},
		Duration: 250 * time.Millisecond,
	}, nil
}

func TestUsageClientRecordsRequestMetadata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	client := WrapClient(stubUsageClient{}, ClientOptions{
		Provider:     "openai",
		APIBase:      "https://api.openai.com",
		DefaultModel: "gpt-5.2",
		JournalDir:   root,
	}).(*UsageClient)
	client.now = func() time.Time {
		return time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	}
	defer func() { _ = client.Close() }()

	ctx := WithMetadata(context.Background(), "run_test_1", "evt_test_1")
	_, err := client.Chat(ctx, llm.Request{Model: "gpt-5.2", Scene: "agent.step"})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(root, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var rec RequestRecord
	if err := json.Unmarshal(data[:len(data)-1], &rec); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if rec.RunID != "run_test_1" || rec.OriginEventID != "evt_test_1" {
		t.Fatalf("record metadata = %+v", rec)
	}
	if rec.Scene != "agent.step" || rec.APIHost != "api.openai.com" || rec.TotalTokens != 18 {
		t.Fatalf("record content = %+v", rec)
	}
	if rec.CachedInputTokens != 5 || rec.CacheCreationInputTokens != 3 {
		t.Fatalf("record cache tokens = %+v", rec)
	}
	if got := rec.CacheDetails["ephemeral_5m_input_tokens"]; got != 3 {
		t.Fatalf("record cache details = %+v", rec.CacheDetails)
	}
	if rec.CostCurrency != "USD" || !rec.CostEstimated || !costAlmostEqual(rec.TotalCost, 0.035) {
		t.Fatalf("record cost = %+v", rec)
	}
}
