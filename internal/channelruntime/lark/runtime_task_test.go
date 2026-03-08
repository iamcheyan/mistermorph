package lark

import (
	"strings"
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/internal/chathistory"
)

func TestBuildLarkPromptMessagesSeparatesHistoryAndCurrent(t *testing.T) {
	t.Parallel()

	historyMsg, currentMsg, err := buildLarkPromptMessages([]chathistory.ChatHistoryItem{{
		Channel:   chathistory.ChannelLark,
		Kind:      chathistory.KindInboundUser,
		MessageID: "101",
		SentAt:    time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC),
		Text:      "earlier",
	}}, larkJob{
		ChatID:      "oc_123",
		ChatType:    "group",
		MessageID:   "102",
		FromUserID:  "ou_123",
		DisplayName: "Alice",
		Text:        "latest",
		SentAt:      time.Date(2026, 3, 8, 9, 2, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("buildLarkPromptMessages() error = %v", err)
	}
	if historyMsg == nil {
		t.Fatalf("historyMsg = nil")
	}
	if strings.Contains(historyMsg.Content, "\"text\": \"latest\"") {
		t.Fatalf("history should not contain latest message: %s", historyMsg.Content)
	}
	if !strings.Contains(historyMsg.Content, "\"text\": \"earlier\"") {
		t.Fatalf("history should contain prior message: %s", historyMsg.Content)
	}
	if currentMsg == nil {
		t.Fatalf("currentMsg = nil")
	}
	if !strings.Contains(currentMsg.Content, "\"text\": \"latest\"") {
		t.Fatalf("current message should contain latest text: %s", currentMsg.Content)
	}
}

func TestBuildLarkPromptMessagesOmitsEmptyHistory(t *testing.T) {
	t.Parallel()

	historyMsg, currentMsg, err := buildLarkPromptMessages(nil, larkJob{
		ChatID:      "oc_123",
		ChatType:    "group",
		MessageID:   "102",
		FromUserID:  "ou_123",
		DisplayName: "Alice",
		Text:        "latest",
		SentAt:      time.Date(2026, 3, 8, 9, 2, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("buildLarkPromptMessages() error = %v", err)
	}
	if historyMsg != nil {
		t.Fatalf("historyMsg should be nil when history is empty")
	}
	if currentMsg == nil || !strings.Contains(currentMsg.Content, "\"text\": \"latest\"") {
		t.Fatalf("current message should still be present: %#v", currentMsg)
	}
}
