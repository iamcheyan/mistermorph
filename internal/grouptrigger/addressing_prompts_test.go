package grouptrigger

import (
	"encoding/json"
	"testing"

	"github.com/quailyquaily/mistermorph/internal/chathistory"
)

func TestRenderAddressingPrompts_TrimHistoryToLatestThree(t *testing.T) {
	history := []chathistory.ChatHistoryItem{
		{MessageID: "1", Text: "m1"},
		{MessageID: "2", Text: "m2"},
		{MessageID: "3", Text: "m3"},
		{MessageID: "4", Text: "m4"},
		{MessageID: "5", Text: "m5"},
	}

	_, userPrompt, err := RenderAddressingPrompts("persona", "", map[string]any{"text": "current"}, history)
	if err != nil {
		t.Fatalf("RenderAddressingPrompts() error = %v", err)
	}

	var payload struct {
		ChatHistoryMessages []chathistory.ChatHistoryItem `json:"chat_history_messages"`
	}
	if err := json.Unmarshal([]byte(userPrompt), &payload); err != nil {
		t.Fatalf("json.Unmarshal(userPrompt) error = %v", err)
	}
	if len(payload.ChatHistoryMessages) != 3 {
		t.Fatalf("chat_history_messages len = %d, want 3", len(payload.ChatHistoryMessages))
	}

	wantIDs := []string{"3", "4", "5"}
	for i, want := range wantIDs {
		if got := payload.ChatHistoryMessages[i].MessageID; got != want {
			t.Fatalf("chat_history_messages[%d].message_id = %q, want %q", i, got, want)
		}
	}
}
