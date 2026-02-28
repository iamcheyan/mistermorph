package idempotency

import (
	"strings"
	"testing"
)

func TestManualContactKey_Format(t *testing.T) {
	key := ManualContactKey("tg:@Alice")
	if !strings.HasPrefix(key, "manual:tg__alice:") {
		t.Fatalf("ManualContactKey() = %q, want prefix manual:tg__alice:", key)
	}
}

func TestMessageEnvelopeKey_UsesSharedAlgorithm(t *testing.T) {
	got := MessageEnvelopeKey("MSG-001")
	want := "msg:msg_001"
	if got != want {
		t.Fatalf("MessageEnvelopeKey() = %q, want %q", got, want)
	}
}
