package memoryruntime

import (
	"strings"
	"testing"
)

func TestBuildFallbackDraft(t *testing.T) {
	t.Parallel()

	if got := buildFallbackDraft(""); len(got.SummaryItems) != 0 {
		t.Fatalf("empty fallback draft = %#v, want empty", got)
	}

	got := buildFallbackDraft("  hello   world  ")
	if len(got.SummaryItems) != 1 || got.SummaryItems[0] != "hello world" {
		t.Fatalf("summary_items = %#v, want [\"hello world\"]", got.SummaryItems)
	}

	long := strings.Repeat("a", defaultFallbackSummaryMaxRunes+8)
	got = buildFallbackDraft(long)
	if len(got.SummaryItems) != 1 {
		t.Fatalf("len(summary_items) = %d, want 1", len(got.SummaryItems))
	}
	if len([]rune(got.SummaryItems[0])) != defaultFallbackSummaryMaxRunes {
		t.Fatalf("rune length = %d, want %d", len([]rune(got.SummaryItems[0])), defaultFallbackSummaryMaxRunes)
	}
}
