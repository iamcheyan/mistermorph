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
