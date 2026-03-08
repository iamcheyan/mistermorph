package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/llm"
)

func TestRun_MetaPrecedesHistoryAndCurrentMessageIsLast(t *testing.T) {
	t.Parallel()

	client := newMockClient(finalResponse("ok"))
	e := New(client, baseRegistry(), baseCfg(), DefaultPromptSpec())

	current := &llm.Message{Role: "user", Content: "CURRENT_TURN"}
	_, _, err := e.Run(context.Background(), "RAW_TASK_SHOULD_NOT_APPEAR", RunOptions{
		History: []llm.Message{{Role: "user", Content: "HISTORY_CONTEXT"}},
		Meta: map[string]any{
			"trigger": "telegram",
		},
		CurrentMessage: current,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	calls := client.allCalls()
	if len(calls) != 1 {
		t.Fatalf("chat calls = %d, want 1", len(calls))
	}
	msgs := calls[0].Messages
	if len(msgs) != 4 {
		t.Fatalf("messages len = %d, want 4", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Fatalf("messages[0].role = %q, want system", msgs[0].Role)
	}
	if !strings.Contains(msgs[1].Content, "mister_morph_meta") {
		t.Fatalf("messages[1] = %q, want injected meta", msgs[1].Content)
	}
	if msgs[2].Content != "HISTORY_CONTEXT" {
		t.Fatalf("messages[2] = %q, want history", msgs[2].Content)
	}
	if msgs[3].Content != "CURRENT_TURN" {
		t.Fatalf("messages[3] = %q, want current turn", msgs[3].Content)
	}
	if requestContains(calls, 0, "RAW_TASK_SHOULD_NOT_APPEAR") {
		t.Fatalf("raw task should not be appended when CurrentMessage is set")
	}
}

func TestRun_AppendsRawTaskWhenCurrentMessageIsNil(t *testing.T) {
	t.Parallel()

	client := newMockClient(finalResponse("ok"))
	e := New(client, baseRegistry(), baseCfg(), DefaultPromptSpec())

	_, _, err := e.Run(context.Background(), "RAW_TASK", RunOptions{
		History: []llm.Message{{Role: "user", Content: "HISTORY_CONTEXT"}},
		Meta: map[string]any{
			"trigger": "telegram",
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	calls := client.allCalls()
	if len(calls) != 1 {
		t.Fatalf("chat calls = %d, want 1", len(calls))
	}
	msgs := calls[0].Messages
	if len(msgs) != 4 {
		t.Fatalf("messages len = %d, want 4", len(msgs))
	}
	if msgs[3].Content != "RAW_TASK" {
		t.Fatalf("messages[3] = %q, want raw task", msgs[3].Content)
	}
}
