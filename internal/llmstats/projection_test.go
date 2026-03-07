package llmstats

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectionRefreshAggregatesAndReplaysTail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	journalDir := filepath.Join(root, "journal")
	projectionPath := filepath.Join(root, "projection.json")
	journal := NewJournal(journalDir, JournalOptions{MaxFileBytes: 1024 * 1024})
	journal.now = func() time.Time {
		return time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	}
	defer func() { _ = journal.Close() }()

	appendRecord := func(host, model string, input, output int64) {
		t.Helper()
		_, err := journal.Append(RequestRecord{
			TS:           time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Provider:     "openai",
			APIBase:      "https://" + host,
			Model:        model,
			InputTokens:  input,
			OutputTokens: output,
			TotalTokens:  input + output,
		})
		if err != nil {
			t.Fatalf("Append(%s,%s) error = %v", host, model, err)
		}
	}

	appendRecord("api.openai.com", "gpt-5.2", 10, 5)
	appendRecord("api.openai.com", "gpt-5-mini", 20, 10)

	store := NewProjectionStore(journalDir, projectionPath)
	store.now = func() time.Time {
		return time.Date(2026, 3, 7, 12, 30, 0, 0, time.UTC)
	}
	proj, err := store.Refresh()
	if err != nil {
		t.Fatalf("Refresh(1) error = %v", err)
	}
	if proj.Summary.Requests != 2 || proj.Summary.TotalTokens != 45 {
		t.Fatalf("projection1 summary = %+v, want requests=2 total_tokens=45", proj.Summary)
	}
	if len(proj.APIHosts) != 1 || proj.APIHosts[0].APIHost != "api.openai.com" {
		t.Fatalf("projection1 hosts = %+v", proj.APIHosts)
	}
	if len(proj.Models) != 2 {
		t.Fatalf("len(projection1 models) = %d, want 2", len(proj.Models))
	}

	appendRecord("api.openai.com", "gpt-5.2", 3, 2)
	proj, err = store.Refresh()
	if err != nil {
		t.Fatalf("Refresh(2) error = %v", err)
	}
	if proj.Summary.Requests != 3 || proj.Summary.TotalTokens != 50 {
		t.Fatalf("projection2 summary = %+v, want requests=3 total_tokens=50", proj.Summary)
	}
	if proj.ProjectedOffset.File == "" || proj.ProjectedOffset.Line != 3 {
		t.Fatalf("projection2 offset = %+v, want line 3", proj.ProjectedOffset)
	}
}

func TestProjectionRefreshIgnoresIncompleteTail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	journalDir := filepath.Join(root, "journal")
	if err := os.MkdirAll(journalDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	segmentPath := filepath.Join(journalDir, "since-2026-03-07-0001.jsonl")
	content := "{\"ts\":\"2026-03-07T12:00:00Z\",\"provider\":\"openai\",\"api_host\":\"api.openai.com\",\"model\":\"gpt-5.2\",\"input_tokens\":10,\"output_tokens\":5,\"total_tokens\":15}\n" +
		"{\"ts\":\"2026-03-07T12:01:00Z\",\"provider\":\"openai\",\"api_host\":\"api.openai.com\",\"model\":\"gpt-5-mini\""
	if err := os.WriteFile(segmentPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewProjectionStore(journalDir, filepath.Join(root, "projection.json"))
	proj, err := store.Refresh()
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if proj.Summary.Requests != 1 || proj.Summary.TotalTokens != 15 {
		t.Fatalf("projection summary = %+v, want one committed record", proj.Summary)
	}
	if proj.ProjectedOffset.File != "since-2026-03-07-0001.jsonl" || proj.ProjectedOffset.Line != 1 {
		t.Fatalf("projection offset = %+v, want first line only", proj.ProjectedOffset)
	}
}
