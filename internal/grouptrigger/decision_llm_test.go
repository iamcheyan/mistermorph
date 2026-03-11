package grouptrigger

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/llm"
)

type stubDecisionLLMClient struct {
	results []llm.Result
	err     error
	calls   []llm.Request
}

func (s *stubDecisionLLMClient) Chat(_ context.Context, req llm.Request) (llm.Result, error) {
	s.calls = append(s.calls, req)
	if s.err != nil {
		return llm.Result{}, s.err
	}
	if len(s.results) == 0 {
		return llm.Result{}, fmt.Errorf("no stub result")
	}
	res := s.results[0]
	s.results = s.results[1:]
	return res, nil
}

type stubDecisionTool struct {
	name      string
	execCount int
	lastEmoji string
}

func (s *stubDecisionTool) Name() string { return s.name }

func (s *stubDecisionTool) Description() string { return "stub tool" }

func (s *stubDecisionTool) ParameterSchema() string {
	return `{"type":"object","properties":{"emoji":{"type":"string"}}}`
}

func (s *stubDecisionTool) Execute(_ context.Context, params map[string]any) (string, error) {
	s.execCount++
	emoji, _ := params["emoji"].(string)
	s.lastEmoji = emoji
	return "ok", nil
}

func TestDecideViaLLM_EnforcesLightweightReaction(t *testing.T) {
	t.Parallel()

	client := &stubDecisionLLMClient{
		results: []llm.Result{
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":"🤨","reason":"x"}`},
		},
	}
	tool := &stubDecisionTool{name: "message_react"}

	got, ok, err := DecideViaLLM(context.Background(), LLMDecisionOptions{
		Client:         client,
		Model:          "gpt-5.2",
		SystemPrompt:   "system",
		UserPrompt:     "user",
		AddressingTool: tool,
	})
	if err != nil {
		t.Fatalf("DecideViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("DecideViaLLM() ok = false, want true")
	}
	if !got.IsLightweight {
		t.Fatalf("IsLightweight = false, want true")
	}
	if tool.execCount != 1 {
		t.Fatalf("tool exec count = %d, want 1", tool.execCount)
	}
	if tool.lastEmoji != "🤨" {
		t.Fatalf("last emoji = %q, want %q", tool.lastEmoji, "🤨")
	}
}

func TestDecideViaLLM_AllowsLightweightWithoutReactionTool(t *testing.T) {
	t.Parallel()

	client := &stubDecisionLLMClient{
		results: []llm.Result{
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":"🤨","reason":"x"}`},
		},
	}

	got, ok, err := DecideViaLLM(context.Background(), LLMDecisionOptions{
		Client:       client,
		Model:        "gpt-5.2",
		Scene:        "slack.addressing_decision",
		SystemPrompt: "system",
		UserPrompt:   "user",
	})
	if err != nil {
		t.Fatalf("DecideViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("DecideViaLLM() ok = false, want true")
	}
	if !got.IsLightweight {
		t.Fatalf("IsLightweight = false, want true")
	}
	if len(client.calls) != 1 {
		t.Fatalf("chat calls = %d, want 1", len(client.calls))
	}
	if len(client.calls[0].Tools) != 0 {
		t.Fatalf("tools len = %d, want 0", len(client.calls[0].Tools))
	}
	if client.calls[0].Scene != "slack.addressing_decision" {
		t.Fatalf("scene = %q, want %q", client.calls[0].Scene, "slack.addressing_decision")
	}
}

func TestDecideViaLLM_DoesNotDuplicateReactionAfterToolCall(t *testing.T) {
	t.Parallel()

	client := &stubDecisionLLMClient{
		results: []llm.Result{
			{
				ToolCalls: []llm.ToolCall{
					{
						ID:        "tc_1",
						Name:      "message_react",
						Arguments: map[string]any{"emoji": ":+1:"},
					},
				},
			},
			{Text: `{"addressed":false,"confidence":0.4,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":":+1:","reason":"x"}`},
		},
	}
	tool := &stubDecisionTool{name: "message_react"}

	_, ok, err := DecideViaLLM(context.Background(), LLMDecisionOptions{
		Client:         client,
		Model:          "gpt-5.2",
		SystemPrompt:   "system",
		UserPrompt:     "user",
		AddressingTool: tool,
	})
	if err != nil {
		t.Fatalf("DecideViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("DecideViaLLM() ok = false, want true")
	}
	if tool.execCount != 1 {
		t.Fatalf("tool exec count = %d, want 1", tool.execCount)
	}
	if tool.lastEmoji != ":+1:" {
		t.Fatalf("last emoji = %q, want %q", tool.lastEmoji, ":+1:")
	}
}

func TestDecideViaLLM_InvalidJSON(t *testing.T) {
	t.Parallel()

	client := &stubDecisionLLMClient{
		results: []llm.Result{
			{Text: `not-json`},
		},
	}

	_, _, err := DecideViaLLM(context.Background(), LLMDecisionOptions{
		Client:       client,
		Model:        "gpt-5.2",
		SystemPrompt: "system",
		UserPrompt:   "user",
	})
	if err == nil {
		t.Fatalf("DecideViaLLM() error = nil, want non-nil")
	}
}

func TestDecideViaLLM_NoToolRejectsToolCalls(t *testing.T) {
	t.Parallel()

	client := &stubDecisionLLMClient{
		results: []llm.Result{
			{
				ToolCalls: []llm.ToolCall{
					{
						ID:        "tc_1",
						Name:      "message_react",
						Arguments: map[string]any{"emoji": ":+1:"},
					},
				},
			},
		},
	}

	_, _, err := DecideViaLLM(context.Background(), LLMDecisionOptions{
		Client:       client,
		Model:        "gpt-5.2",
		SystemPrompt: "system",
		UserPrompt:   "user",
	})
	if err == nil {
		t.Fatalf("DecideViaLLM() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "no addressing tool is configured") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "no addressing tool is configured")
	}
}
