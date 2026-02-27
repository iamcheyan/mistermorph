package telegram

import (
	"context"
	"fmt"
	"testing"

	"github.com/quailyquaily/mistermorph/llm"
)

type stubAddressingLLMClient struct {
	results []llm.Result
	err     error
	calls   []llm.Request
}

func (s *stubAddressingLLMClient) Chat(_ context.Context, req llm.Request) (llm.Result, error) {
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

type stubAddressingTool struct {
	name        string
	execCount   int
	lastEmoji   string
	failOnEmoji string
}

func (s *stubAddressingTool) Name() string { return s.name }

func (s *stubAddressingTool) Description() string { return "stub tool" }

func (s *stubAddressingTool) ParameterSchema() string {
	return `{"type":"object","properties":{"emoji":{"type":"string"}},"required":["emoji"]}`
}

func (s *stubAddressingTool) Execute(_ context.Context, params map[string]any) (string, error) {
	s.execCount++
	emoji, _ := params["emoji"].(string)
	s.lastEmoji = emoji
	if emoji == s.failOnEmoji {
		return "", fmt.Errorf("emoji not allowed: %s", emoji)
	}
	return "ok", nil
}

func TestAddressingDecisionViaLLM_EnforceLightweightReaction(t *testing.T) {
	client := &stubAddressingLLMClient{
		results: []llm.Result{
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":"🤨","reason":"x"}`},
		},
	}
	tool := &stubAddressingTool{name: "telegram_react"}

	got, ok, err := addressingDecisionViaLLM(context.Background(), client, "gpt-5.2", nil, "啧", nil, tool)
	if err != nil {
		t.Fatalf("addressingDecisionViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("addressingDecisionViaLLM() ok = false, want true")
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

func TestAddressingDecisionViaLLM_EnforceLightweightReactionReturnsErrorOnToolFailure(t *testing.T) {
	client := &stubAddressingLLMClient{
		results: []llm.Result{
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":"🚫","reason":"x"}`},
		},
	}
	tool := &stubAddressingTool{name: "telegram_react", failOnEmoji: "🚫"}

	_, _, err := addressingDecisionViaLLM(context.Background(), client, "gpt-5.2", nil, "啧", nil, tool)
	if err == nil {
		t.Fatalf("addressingDecisionViaLLM() error = nil, want non-nil")
	}
	if tool.execCount != 1 {
		t.Fatalf("tool exec count = %d, want 1", tool.execCount)
	}
}

func TestAddressingDecisionViaLLM_EmptyReactionKeepsLightweight(t *testing.T) {
	client := &stubAddressingLLMClient{
		results: []llm.Result{
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":"","reason":"x"}`},
		},
	}
	tool := &stubAddressingTool{name: "telegram_react"}

	got, ok, err := addressingDecisionViaLLM(context.Background(), client, "gpt-5.2", nil, "啧", nil, tool)
	if err != nil {
		t.Fatalf("addressingDecisionViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("addressingDecisionViaLLM() ok = false, want true")
	}
	if !got.IsLightweight {
		t.Fatalf("IsLightweight = false, want true")
	}
	if tool.execCount != 0 {
		t.Fatalf("tool exec count = %d, want 0", tool.execCount)
	}
}

func TestAddressingDecisionViaLLM_NoReactionWhenNotLightweight(t *testing.T) {
	client := &stubAddressingLLMClient{
		results: []llm.Result{
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":false,"interject":0.05,"impulse":0.1,"is_lightweight":false,"reason":"x"}`},
		},
	}
	tool := &stubAddressingTool{name: "telegram_react"}

	got, ok, err := addressingDecisionViaLLM(context.Background(), client, "gpt-5.2", nil, "啧", nil, tool)
	if err != nil {
		t.Fatalf("addressingDecisionViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("addressingDecisionViaLLM() ok = false, want true")
	}
	if got.IsLightweight {
		t.Fatalf("IsLightweight = true, want false")
	}
	if tool.execCount != 0 {
		t.Fatalf("tool exec count = %d, want 0", tool.execCount)
	}
}

func TestAddressingDecisionViaLLM_NoDuplicateReactionWhenModelAlreadyCalledTool(t *testing.T) {
	client := &stubAddressingLLMClient{
		results: []llm.Result{
			{
				ToolCalls: []llm.ToolCall{
					{
						ID:        "tc_1",
						Name:      "telegram_react",
						Arguments: map[string]any{"emoji": "🤨"},
					},
				},
			},
			{Text: `{"addressed":false,"confidence":0.2,"wanna_interject":true,"interject":0.1,"impulse":0.3,"is_lightweight":true,"reaction":"🤨","reason":"x"}`},
		},
	}
	tool := &stubAddressingTool{name: "telegram_react"}

	got, ok, err := addressingDecisionViaLLM(context.Background(), client, "gpt-5.2", nil, "啧", nil, tool)
	if err != nil {
		t.Fatalf("addressingDecisionViaLLM() error = %v", err)
	}
	if !ok {
		t.Fatalf("addressingDecisionViaLLM() ok = false, want true")
	}
	if !got.IsLightweight {
		t.Fatalf("IsLightweight = false, want true")
	}
	if tool.execCount != 1 {
		t.Fatalf("tool exec count = %d, want 1", tool.execCount)
	}
}
