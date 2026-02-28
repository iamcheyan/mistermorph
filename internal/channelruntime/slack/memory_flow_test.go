package slack

import (
	"strings"
	"testing"
)

func TestSlackMemorySubjectID(t *testing.T) {
	got := slackMemorySubjectID(slackJob{
		TeamID:    "T123",
		ChannelID: "C456",
	})
	if got != "slack--t123--c456" {
		t.Fatalf("subject_id = %q, want slack--t123--c456", got)
	}
}

func TestSlackMemorySessionID(t *testing.T) {
	job := slackJob{TeamID: "T1", ChannelID: "C1"}
	if got := slackMemorySessionID(job); got != "slack:T1:C1" {
		t.Fatalf("session_id = %q, want slack:T1:C1", got)
	}
	job.ThreadTS = "1739667000.000050"
	if got := slackMemorySessionID(job); got != "slack:T1:C1" {
		t.Fatalf("session_id(thread) = %q, want slack:T1:C1", got)
	}
}

func TestSlackMemoryTaskRunID(t *testing.T) {
	if got := slackMemoryTaskRunID(slackJob{TaskID: " task_123 "}); got != "task_123" {
		t.Fatalf("task_run_id = %q, want task_123", got)
	}
	if got := slackMemoryTaskRunID(slackJob{}); got != "" {
		t.Fatalf("empty task_run_id = %q, want empty", got)
	}
}

func TestSlackMemoryParticipants(t *testing.T) {
	got := slackMemoryParticipants(slackJob{
		UserID:       "U111",
		MentionUsers: []string{"U222", "U111", " U333 "},
	})
	if len(got) != 3 {
		t.Fatalf("participants len = %d, want 3 (%#v)", len(got), got)
	}
	if got[0].ID != "U111" || got[1].ID != "U222" || got[2].ID != "U333" {
		t.Fatalf("participants order/id mismatch: %#v", got)
	}
	for i, p := range got {
		if p.Protocol != "slack" {
			t.Fatalf("participants[%d].protocol = %q, want slack", i, p.Protocol)
		}
	}
}

func TestBuildSlackMemoryDraft(t *testing.T) {
	if got := buildSlackMemoryDraft(""); len(got.SummaryItems) != 0 {
		t.Fatalf("empty output should produce empty draft: %#v", got)
	}
	got := buildSlackMemoryDraft("  hello   world  ")
	if len(got.SummaryItems) != 1 || got.SummaryItems[0] != "hello world" {
		t.Fatalf("draft summary mismatch: %#v", got)
	}

	long := strings.Repeat("a", slackMemorySummaryMaxRunes+50)
	got = buildSlackMemoryDraft(long)
	if len(got.SummaryItems) != 1 {
		t.Fatalf("long draft summary count = %d, want 1", len(got.SummaryItems))
	}
	if len([]rune(got.SummaryItems[0])) != slackMemorySummaryMaxRunes {
		t.Fatalf("long draft summary rune len = %d, want %d", len([]rune(got.SummaryItems[0])), slackMemorySummaryMaxRunes)
	}
}

func TestSlackMemoryRequestContext(t *testing.T) {
	if got := slackMemoryRequestContext("im"); got != "private" {
		t.Fatalf("im request context = %q, want private", got)
	}
	if got := slackMemoryRequestContext("channel"); got != "public" {
		t.Fatalf("channel request context = %q, want public", got)
	}
}
