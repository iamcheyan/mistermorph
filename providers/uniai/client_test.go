package uniai

import (
	"testing"

	"github.com/quailyquaily/mistermorph/llm"
	uniaiapi "github.com/quailyquaily/uniai"
	uniaichat "github.com/quailyquaily/uniai/chat"
)

func TestBuildChatOptionsReplaceMessages(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "new"},
		},
	}

	opts := append(
		[]uniaiapi.ChatOption{uniaiapi.WithMessages(uniaiapi.User("old"))},
		buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, nil, "", nil)...,
	)

	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if len(built.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(built.Messages))
	}
	if built.Messages[0].Content != "new" {
		t.Fatalf("expected replaced message content 'new', got %q", built.Messages[0].Content)
	}
}

func TestBuildChatOptionsPreserveToolCallIDAsIs(t *testing.T) {
	rawID := "  call_1|ts:abc  "
	req := llm.Request{
		Messages: []llm.Message{
			{
				Role:       "tool",
				Content:    `{"content":"ok"}`,
				ToolCallID: rawID,
			},
		},
	}

	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, nil, "", nil)
	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if len(built.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(built.Messages))
	}
	if built.Messages[0].ToolCallID != rawID {
		t.Fatalf("expected tool_call_id preserved as-is %q, got %q", rawID, built.Messages[0].ToolCallID)
	}
}

func TestBuildChatOptionsMapsMessageParts(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Type: llm.PartTypeText, Text: "describe this"},
					{Type: llm.PartTypeImageBase64, MIMEType: "image/png", DataBase64: "QUJD"},
				},
			},
		},
	}

	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, nil, "", nil)
	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if len(built.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(built.Messages))
	}
	if got := len(built.Messages[0].Parts); got != 2 {
		t.Fatalf("message parts length = %d, want 2", got)
	}
	if built.Messages[0].Parts[0].Type != uniaichat.PartTypeText || built.Messages[0].Parts[0].Text != "describe this" {
		t.Fatalf("text part mismatch: %+v", built.Messages[0].Parts[0])
	}
	if built.Messages[0].Parts[1].Type != uniaichat.PartTypeImageBase64 || built.Messages[0].Parts[1].DataBase64 != "QUJD" {
		t.Fatalf("image part mismatch: %+v", built.Messages[0].Parts[1])
	}
}

func TestBuildChatOptionsMapsOnStream(t *testing.T) {
	called := false
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
		OnStream: func(ev llm.StreamEvent) error {
			called = true
			if ev.Delta != "abc" {
				t.Fatalf("delta = %q, want abc", ev.Delta)
			}
			if ev.ToolCallDelta == nil || ev.ToolCallDelta.Name != "message_react" {
				t.Fatalf("tool_call_delta = %#v", ev.ToolCallDelta)
			}
			if !ev.Done {
				t.Fatalf("done = false, want true")
			}
			if ev.Usage == nil || ev.Usage.TotalTokens != 9 {
				t.Fatalf("usage = %#v", ev.Usage)
			}
			return nil
		},
	}
	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, nil, "", nil)

	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.OnStream == nil {
		t.Fatalf("expected on_stream callback")
	}
	if err := built.Options.OnStream(uniaiapi.StreamEvent{
		Delta: "abc",
		ToolCallDelta: &uniaiapi.ToolCallDelta{
			Index:     1,
			ID:        "call_1",
			Name:      "message_react",
			ArgsChunk: "{\"emoji\":\"ok_hand\"}",
		},
		Usage: &uniaichat.Usage{
			InputTokens:  4,
			OutputTokens: 5,
			TotalTokens:  9,
		},
		Done: true,
	}); err != nil {
		t.Fatalf("on_stream callback returned error: %v", err)
	}
	if !called {
		t.Fatalf("expected callback to be called")
	}
}

func TestBuildChatOptionsDisablesOnStreamForGeminiProvider(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
		OnStream: func(llm.StreamEvent) error { return nil },
	}
	opts := buildChatOptions(req, "gemini", false, uniaiapi.ToolsEmulationOff, nil, "", nil)

	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.OnStream != nil {
		t.Fatalf("expected on_stream callback to be disabled for gemini provider")
	}
}

func TestBuildChatOptionsDisablesOnStreamForCloudflareProvider(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
		OnStream: func(llm.StreamEvent) error { return nil },
	}
	opts := buildChatOptions(req, "cloudflare", false, uniaiapi.ToolsEmulationOff, nil, "", nil)

	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.OnStream != nil {
		t.Fatalf("expected on_stream callback to be disabled for cloudflare provider")
	}
}

func TestBuildChatOptionsMapsDebugFn(t *testing.T) {
	var gotLabel, gotPayload string
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
		DebugFn: func(label, payload string) {
			gotLabel = label
			gotPayload = payload
		},
	}
	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, nil, "", nil)

	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.DebugFn == nil {
		t.Fatalf("expected debug callback")
	}
	built.Options.DebugFn("openai.chat.request", `{"messages":[]}`)
	if gotLabel != "openai.chat.request" || gotPayload != `{"messages":[]}` {
		t.Fatalf("debug callback mismatch: label=%q payload=%q", gotLabel, gotPayload)
	}
}

func TestBuildChatOptionsDoesNotInjectTemperatureWhenUnset(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
	}
	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, nil, "", nil)
	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.Temperature != nil {
		t.Fatalf("expected temperature to remain unset, got %v", *built.Options.Temperature)
	}
}

func TestBuildChatOptionsAppliesConfiguredDefaults(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
	}
	temperature := 0.4
	reasoningBudget := 8192
	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, &temperature, "high", &reasoningBudget)
	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.Temperature == nil || *built.Options.Temperature != 0.4 {
		t.Fatalf("temperature = %#v, want 0.4", built.Options.Temperature)
	}
	if built.Options.ReasoningEffort == nil || *built.Options.ReasoningEffort != uniaichat.ReasoningEffortHigh {
		t.Fatalf("reasoning effort = %#v, want high", built.Options.ReasoningEffort)
	}
	if built.Options.ReasoningBudget == nil || *built.Options.ReasoningBudget != 8192 {
		t.Fatalf("reasoning budget = %#v, want 8192", built.Options.ReasoningBudget)
	}
}

func TestBuildChatOptionsSkipsReasoningBudgetForOpenAIResp(t *testing.T) {
	req := llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "hello"}},
	}
	reasoningBudget := 8192
	opts := buildChatOptions(req, "openai_resp", false, uniaiapi.ToolsEmulationOff, nil, "high", &reasoningBudget)
	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.ReasoningEffort == nil || *built.Options.ReasoningEffort != uniaichat.ReasoningEffortHigh {
		t.Fatalf("reasoning effort = %#v, want high", built.Options.ReasoningEffort)
	}
	if built.Options.ReasoningBudget != nil {
		t.Fatalf("reasoning budget = %#v, want nil", built.Options.ReasoningBudget)
	}
}

func TestBuildChatOptionsRequestTemperatureOverridesConfiguredDefault(t *testing.T) {
	req := llm.Request{
		Messages:   []llm.Message{{Role: "user", Content: "hello"}},
		Parameters: map[string]any{"temperature": 0.1},
	}
	temperature := 0.4
	opts := buildChatOptions(req, "", false, uniaiapi.ToolsEmulationOff, &temperature, "", nil)
	built, err := uniaichat.BuildRequest(opts...)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if built.Options.Temperature == nil || *built.Options.Temperature != 0.1 {
		t.Fatalf("temperature = %#v, want 0.1", built.Options.Temperature)
	}
}

func TestPartRoundTripBetweenLLMAndUniai(t *testing.T) {
	src := []llm.Part{
		{Type: llm.PartTypeText, Text: "hello"},
		{Type: llm.PartTypeImageURL, URL: "https://example.com/a.png"},
		{Type: llm.PartTypeImageBase64, MIMEType: "image/jpeg", DataBase64: "QUJD"},
	}
	toUniai := toUniaiPartsFromLLM(src)
	back := toLLMParts(toUniai)

	if len(back) != len(src) {
		t.Fatalf("parts length mismatch: got %d want %d", len(back), len(src))
	}
	for i := range src {
		if back[i] != src[i] {
			t.Fatalf("part[%d] mismatch: got %+v want %+v", i, back[i], src[i])
		}
	}
}

func TestToolCallThoughtSignatureRoundTrip(t *testing.T) {
	sig := "sig_abc"
	origArgs := `{"path":"/tmp/a.txt","dup":1,"dup":2}`
	orig := []uniaiapi.ToolCall{
		{
			ID:               "call_1",
			Type:             "function",
			ThoughtSignature: sig,
			Function: uniaiapi.ToolCallFunction{
				Name:      "read_file",
				Arguments: origArgs,
			},
		},
	}

	out := toLLMToolCalls(orig)
	if len(out) != 1 {
		t.Fatalf("expected 1 llm tool call, got %d", len(out))
	}
	if out[0].ThoughtSignature != sig {
		t.Fatalf("expected round-tripped thought signature %q, got %q", sig, out[0].ThoughtSignature)
	}
	if out[0].RawArguments != origArgs {
		t.Fatalf("expected raw arguments %q, got %q", origArgs, out[0].RawArguments)
	}

	uniaiCalls := toUniaiToolCallsFromLLM(out)
	if len(uniaiCalls) != 1 {
		t.Fatalf("expected 1 uniai tool call, got %d", len(uniaiCalls))
	}
	if uniaiCalls[0].ThoughtSignature != sig {
		t.Fatalf("expected thought signature %q, got %q", sig, uniaiCalls[0].ThoughtSignature)
	}
	if uniaiCalls[0].Function.Arguments != origArgs {
		t.Fatalf("expected exact raw arguments %q, got %q", origArgs, uniaiCalls[0].Function.Arguments)
	}
}

func TestEnsureGeminiToolCallThoughtSignaturesDecodeFromID(t *testing.T) {
	rawID := "call_1|ts:c2lnX2Zyb21faWQ"
	calls := []llm.ToolCall{{
		ID:           rawID,
		Name:         "read_file",
		RawArguments: `{"path":"/tmp/a.txt"}`,
	}}

	out := ensureGeminiToolCallThoughtSignatures(calls)
	if len(out) != 1 {
		t.Fatalf("expected 1 call, got %d", len(out))
	}
	if out[0].ThoughtSignature != "sig_from_id" {
		t.Fatalf("expected signature decoded from id, got %q", out[0].ThoughtSignature)
	}
}

func TestEnsureGeminiToolCallThoughtSignaturesCarryForward(t *testing.T) {
	calls := []llm.ToolCall{
		{
			ID:               "call_1",
			Name:             "read_file",
			RawArguments:     `{"path":"a"}`,
			ThoughtSignature: "sig_1",
		},
		{
			ID:           "call_2",
			Name:         "read_file",
			RawArguments: `{"path":"b"}`,
		},
	}

	out := ensureGeminiToolCallThoughtSignatures(calls)
	if len(out) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(out))
	}
	if out[1].ThoughtSignature != "sig_1" {
		t.Fatalf("expected second call to inherit prior signature, got %q", out[1].ThoughtSignature)
	}
}

func TestEnsureGeminiToolCallThoughtSignaturesSynthesize(t *testing.T) {
	calls := []llm.ToolCall{{
		ID:           "call_42",
		Name:         "read_file",
		RawArguments: `{"path":"a"}`,
	}}

	out1 := ensureGeminiToolCallThoughtSignatures(calls)
	out2 := ensureGeminiToolCallThoughtSignatures(calls)
	if len(out1) != 1 || len(out2) != 1 {
		t.Fatalf("unexpected output lengths: %d, %d", len(out1), len(out2))
	}
	if out1[0].ThoughtSignature == "" {
		t.Fatalf("expected synthesized signature")
	}
	if out1[0].ThoughtSignature != out2[0].ThoughtSignature {
		t.Fatalf("expected deterministic synthesized signature, got %q vs %q", out1[0].ThoughtSignature, out2[0].ThoughtSignature)
	}
}

func TestShouldEnsureGeminiThoughtSignature(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     bool
	}{
		{
			name:     "provider gemini with non gemini model",
			provider: "gemini",
			model:    "gpt-5.2",
			want:     true,
		},
		{
			name:     "openai provider with gemini model",
			provider: "openai",
			model:    "carrot/gemini-3-pro",
			want:     false,
		},
		{
			name:     "empty provider with gemini model",
			provider: "",
			model:    "carrot/gemini-3-pro",
			want:     false,
		},
		{
			name:     "other provider with gemini model",
			provider: "anthropic",
			model:    "carrot/gemini-3-pro",
			want:     false,
		},
		{
			name:     "non gemini provider and model",
			provider: "openai",
			model:    "gpt-5.2",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEnsureGeminiThoughtSignature(tt.provider, tt.model); got != tt.want {
				t.Fatalf("shouldEnsureGeminiThoughtSignature(%q, %q) = %v, want %v", tt.provider, tt.model, got, tt.want)
			}
		})
	}
}
