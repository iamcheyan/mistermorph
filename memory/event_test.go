package memory

import (
	"strings"
	"testing"
)

func TestMemoryEventValidateForAppend(t *testing.T) {
	t.Run("valid no participants", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = nil
		if err := ev.ValidateForAppend(); err != nil {
			t.Fatalf("ValidateForAppend() error = %v", err)
		}
	})

	t.Run("valid participant with protocol", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = []MemoryParticipant{
			{ID: "@johnwick", Nickname: "John Wick", Protocol: "tg"},
		}
		if err := ev.ValidateForAppend(); err != nil {
			t.Fatalf("ValidateForAppend() error = %v", err)
		}
	})

	t.Run("valid agent self marker", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = []MemoryParticipant{
			{ID: 0, Nickname: "阿嬷", Protocol: ""},
		}
		if err := ev.ValidateForAppend(); err != nil {
			t.Fatalf("ValidateForAppend() error = %v", err)
		}
	})

	t.Run("invalid missing event_id", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.EventID = ""
		err := ev.ValidateForAppend()
		if err == nil || !strings.Contains(err.Error(), "event_id is required") {
			t.Fatalf("ValidateForAppend() error = %v, want event_id is required", err)
		}
	})

	t.Run("invalid bad ts_utc", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.TSUTC = "2026/02/28"
		err := ev.ValidateForAppend()
		if err == nil || !strings.Contains(err.Error(), "ts_utc must be RFC3339") {
			t.Fatalf("ValidateForAppend() error = %v, want ts_utc must be RFC3339", err)
		}
	})

	t.Run("invalid participant protocol empty but id not zero", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = []MemoryParticipant{
			{ID: "@johnwick", Nickname: "John Wick", Protocol: ""},
		}
		err := ev.ValidateForAppend()
		if err == nil || !strings.Contains(err.Error(), "agent-self marker requires id=0") {
			t.Fatalf("ValidateForAppend() error = %v, want agent-self marker error", err)
		}
	})

	t.Run("invalid participant id zero with non-empty protocol", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = []MemoryParticipant{
			{ID: 0, Nickname: "Zero", Protocol: "tg"},
		}
		err := ev.ValidateForAppend()
		if err == nil || !strings.Contains(err.Error(), "id=0 is reserved for agent-self marker") {
			t.Fatalf("ValidateForAppend() error = %v, want reserved id=0 error", err)
		}
	})

	t.Run("invalid participant empty nickname", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = []MemoryParticipant{
			{ID: "@johnwick", Nickname: "", Protocol: "tg"},
		}
		err := ev.ValidateForAppend()
		if err == nil || !strings.Contains(err.Error(), "participants[0].nickname is required") {
			t.Fatalf("ValidateForAppend() error = %v, want nickname required", err)
		}
	})

	t.Run("invalid participant id space padded", func(t *testing.T) {
		ev := baseMemoryEvent()
		ev.Participants = []MemoryParticipant{
			{ID: " @johnwick ", Nickname: "John Wick", Protocol: "tg"},
		}
		err := ev.ValidateForAppend()
		if err == nil || !strings.Contains(err.Error(), "participants[0].id must not contain leading/trailing spaces") {
			t.Fatalf("ValidateForAppend() error = %v, want id trim error", err)
		}
	})

}

func baseMemoryEvent() MemoryEvent {
	return MemoryEvent{
		SchemaVersion: CurrentMemoryEventSchemaVersion,
		EventID:       "evt_01",
		TaskRunID:     "run_01",
		TSUTC:         "2026-02-28T06:15:12Z",
		SessionID:     "tg:-1003824466118",
		SubjectID:     "ext:telegram:28036192",
		Channel:       "telegram",
		Participants:  nil,
		TaskText:      "啧啧啧",
		FinalOutput:   "",
	}
}
