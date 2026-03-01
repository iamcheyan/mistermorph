package slack

import (
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
