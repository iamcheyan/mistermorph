package heartbeat

import (
	"strings"
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/memory"
)

func TestHeartbeatTaskRunID(t *testing.T) {
	now := time.Date(2026, 2, 28, 1, 2, 3, 456000000, time.UTC)
	id := heartbeatTaskRunID(now)
	if !strings.HasPrefix(id, "heartbeat:20260228T010203") {
		t.Fatalf("unexpected task run id: %q", id)
	}
}

func TestHeartbeatMemoryParticipants(t *testing.T) {
	ev := memory.MemoryEvent{
		SchemaVersion: memory.CurrentMemoryEventSchemaVersion,
		EventID:       "evt_01",
		TaskRunID:     "heartbeat:20260312T120651.547000000Z",
		TSUTC:         "2026-03-12T12:06:51Z",
		SessionID:     heartbeatMemorySessionID,
		SubjectID:     heartbeatMemorySubjectID,
		Channel:       "heartbeat",
		Participants:  heartbeatMemoryParticipants(),
		TaskText:      "heartbeat task",
		FinalOutput:   "summary",
	}
	if err := ev.ValidateForAppend(); err != nil {
		t.Fatalf("ValidateForAppend() error = %v", err)
	}
}
