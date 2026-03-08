package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/llm"
	"github.com/spf13/viper"
)

type stubPlanCreateLLMClient struct {
	reply   string
	err     error
	lastReq llm.Request
}

func (s *stubPlanCreateLLMClient) Chat(_ context.Context, req llm.Request) (llm.Result, error) {
	s.lastReq = req
	if s.err != nil {
		return llm.Result{}, s.err
	}
	return llm.Result{Text: s.reply}, nil
}

func TestPlanCreateExecuteRejectsEmptyStepsAfterNormalization(t *testing.T) {
	client := &stubPlanCreateLLMClient{
		reply: `{"plan":{"steps":[{"step":"   "},{"step":""}]}}`,
	}
	tool := NewPlanCreateTool(client, "gpt-5.2", []string{"bash"}, 6)
	_, err := tool.Execute(context.Background(), map[string]any{"task": "t"})
	if err == nil {
		t.Fatal("expected error for empty normalized steps")
	}
	if !strings.Contains(err.Error(), "empty steps") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlanCreateExecuteNormalizesAndDropsEmptySteps(t *testing.T) {
	client := &stubPlanCreateLLMClient{
		reply: `{"plan":{"steps":[{"step":"   ","status":"pending"},{"step":" collect data ","status":""},{"step":"summarize","status":"unknown"}]}}`,
	}
	tool := NewPlanCreateTool(client, "gpt-5.2", []string{"bash"}, 6)
	out, err := tool.Execute(context.Background(), map[string]any{"task": "t"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var parsed struct {
		Plan struct {
			Steps []struct {
				Step   string `json:"step"`
				Status string `json:"status"`
			} `json:"steps"`
		} `json:"plan"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(parsed.Plan.Steps) != 2 {
		t.Fatalf("len(steps) = %d, want 2", len(parsed.Plan.Steps))
	}
	if parsed.Plan.Steps[0].Step != "collect data" || parsed.Plan.Steps[0].Status != "in_progress" {
		t.Fatalf("steps[0] = %+v, want step=%q status=%q", parsed.Plan.Steps[0], "collect data", "in_progress")
	}
	if parsed.Plan.Steps[1].Step != "summarize" || parsed.Plan.Steps[1].Status != "pending" {
		t.Fatalf("steps[1] = %+v, want step=%q status=%q", parsed.Plan.Steps[1], "summarize", "pending")
	}
}

func TestPlanCreateExecuteInjectsPersonaIdentity(t *testing.T) {
	stateDir := t.TempDir()
	identity := "# IDENTITY.md\n\nName: Persona Bot\n"
	soul := "# SOUL.md\n\n## Vibe\n\nDirect and calm.\n"
	if err := os.WriteFile(filepath.Join(stateDir, "IDENTITY.md"), []byte(identity), 0o644); err != nil {
		t.Fatalf("write IDENTITY.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "SOUL.md"), []byte(soul), 0o644); err != nil {
		t.Fatalf("write SOUL.md: %v", err)
	}

	prevStateDir := viper.GetString("file_state_dir")
	t.Cleanup(func() { viper.Set("file_state_dir", prevStateDir) })
	viper.Set("file_state_dir", stateDir)

	client := &stubPlanCreateLLMClient{
		reply: `{"plan":{"steps":[{"step":"collect data"}]}}`,
	}
	tool := NewPlanCreateTool(client, "gpt-5.2", []string{"bash"}, 6)
	if _, err := tool.Execute(context.Background(), map[string]any{"task": "t"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(client.lastReq.Messages) == 0 {
		t.Fatalf("expected LLM messages")
	}
	systemPrompt := client.lastReq.Messages[0].Content
	if !strings.Contains(systemPrompt, ">>> BEGIN OF IDENTITY.md <<<") {
		t.Fatalf("system prompt missing identity block: %q", systemPrompt)
	}
	if !strings.Contains(systemPrompt, "Persona Bot") {
		t.Fatalf("system prompt missing identity content: %q", systemPrompt)
	}
	if !strings.Contains(systemPrompt, ">>> BEGIN OF SOUL.md <<<") {
		t.Fatalf("system prompt missing soul block: %q", systemPrompt)
	}
	if !strings.Contains(systemPrompt, "Direct and calm.") {
		t.Fatalf("system prompt missing soul content: %q", systemPrompt)
	}
}

func TestPlanCreateExecuteOmitsStyleFromSchemaAndPayload(t *testing.T) {
	client := &stubPlanCreateLLMClient{
		reply: `{"plan":{"steps":[{"step":"collect data"}]}}`,
	}
	tool := NewPlanCreateTool(client, "gpt-5.2", []string{"bash"}, 6)

	schema := tool.ParameterSchema()
	if strings.Contains(schema, `"style"`) {
		t.Fatalf("schema should not contain style field: %s", schema)
	}

	if _, err := tool.Execute(context.Background(), map[string]any{
		"task":  "t",
		"style": "terse",
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(client.lastReq.Messages) < 2 {
		t.Fatalf("expected user payload message")
	}
	if strings.Contains(client.lastReq.Messages[1].Content, `"style"`) {
		t.Fatalf("payload should not contain style field: %s", client.lastReq.Messages[1].Content)
	}
}
