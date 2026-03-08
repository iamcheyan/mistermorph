package slack

import (
	"strings"
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
)

func TestGenerateSlackPlanProgressMessage(t *testing.T) {
	plan := &agent.Plan{
		Steps: []agent.PlanStep{
			{Step: "scan repo", Status: agent.PlanStatusCompleted},
			{Step: "patch bug", Status: agent.PlanStatusInProgress},
		},
	}
	msg := generateSlackPlanProgressMessage(plan, agent.PlanStepUpdate{
		CompletedIndex: 0,
		CompletedStep:  "scan repo",
		StartedIndex:   1,
		StartedStep:    "patch bug",
	})
	if msg != "scan repo" {
		t.Fatalf("message = %q, want %q", msg, "scan repo")
	}
}

func TestNewSlackOutboundReactionHistoryItem(t *testing.T) {
	job := slackJob{
		TeamID:      "T1",
		ChannelID:   "C1",
		ChatType:    "channel",
		ThreadTS:    "1739667600.000100",
		UserID:      "U1",
		Username:    "alice",
		DisplayName: "Alice",
		Text:        "hello",
	}
	item := newSlackOutboundReactionHistoryItem(job, "[reacted: :thumbsup:]", "thumbsup", time.Now().UTC(), "UBOT")
	if item.Kind != chathistory.KindOutboundReaction {
		t.Fatalf("kind = %q, want %q", item.Kind, chathistory.KindOutboundReaction)
	}
	if item.Text != "[reacted: :thumbsup:]" {
		t.Fatalf("text = %q, want %q", item.Text, "[reacted: :thumbsup:]")
	}
}

func TestBuildSlackPromptMessagesSeparatesHistoryAndCurrent(t *testing.T) {
	t.Parallel()

	historyMsg, currentMsg, err := buildSlackPromptMessages([]chathistory.ChatHistoryItem{{
		Channel:   chathistory.ChannelSlack,
		Kind:      chathistory.KindInboundUser,
		MessageID: "101",
		SentAt:    time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC),
		Text:      "earlier",
	}}, slackJob{
		TeamID:      "T1",
		ChannelID:   "C1",
		ChatType:    "channel",
		MessageTS:   "102.0001",
		ThreadTS:    "102.0001",
		UserID:      "U1",
		Username:    "alice",
		DisplayName: "Alice",
		Text:        "latest",
		SentAt:      time.Date(2026, 3, 8, 9, 2, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("buildSlackPromptMessages() error = %v", err)
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

func TestBuildSlackPromptMessagesOmitsEmptyHistory(t *testing.T) {
	t.Parallel()

	historyMsg, currentMsg, err := buildSlackPromptMessages(nil, slackJob{
		TeamID:      "T1",
		ChannelID:   "C1",
		ChatType:    "channel",
		MessageTS:   "102.0001",
		ThreadTS:    "102.0001",
		UserID:      "U1",
		Username:    "alice",
		DisplayName: "Alice",
		Text:        "latest",
		SentAt:      time.Date(2026, 3, 8, 9, 2, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("buildSlackPromptMessages() error = %v", err)
	}
	if historyMsg != nil {
		t.Fatalf("historyMsg should be nil when history is empty")
	}
	if currentMsg == nil || !strings.Contains(currentMsg.Content, "\"text\": \"latest\"") {
		t.Fatalf("current message should still be present: %#v", currentMsg)
	}
}
