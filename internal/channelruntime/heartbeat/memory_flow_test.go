package heartbeat

import (
	"strings"
	"testing"
	"time"
)

func TestHeartbeatTaskRunID(t *testing.T) {
	now := time.Date(2026, 2, 28, 1, 2, 3, 456000000, time.UTC)
	id := heartbeatTaskRunID(now)
	if !strings.HasPrefix(id, "heartbeat:20260228T010203") {
		t.Fatalf("unexpected task run id: %q", id)
	}
}

func TestBuildHeartbeatDraft(t *testing.T) {
	if got := buildHeartbeatDraft(""); len(got.SummaryItems) != 0 {
		t.Fatalf("empty summary should produce empty draft: %#v", got)
	}
	got := buildHeartbeatDraft("  hello   world  ")
	if len(got.SummaryItems) != 1 || got.SummaryItems[0] != "hello world" {
		t.Fatalf("draft summary mismatch: %#v", got)
	}

	long := strings.Repeat("a", heartbeatMemorySummaryRunes+100)
	got = buildHeartbeatDraft(long)
	if len(got.SummaryItems) != 1 {
		t.Fatalf("draft summary count = %d, want 1", len(got.SummaryItems))
	}
	if len([]rune(got.SummaryItems[0])) != heartbeatMemorySummaryRunes {
		t.Fatalf("draft summary rune len = %d, want %d", len([]rune(got.SummaryItems[0])), heartbeatMemorySummaryRunes)
	}
}
