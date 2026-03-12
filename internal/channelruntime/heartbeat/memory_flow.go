package heartbeat

import (
	"fmt"
	"time"

	"github.com/quailyquaily/mistermorph/memory"
)

const (
	heartbeatMemorySubjectID = "heartbeat"
	heartbeatMemorySessionID = "heartbeat"
)

func heartbeatTaskRunID(now time.Time) string {
	now = now.UTC()
	return fmt.Sprintf("heartbeat:%s", now.Format("20060102T150405.000000000Z07:00"))
}

func heartbeatMemoryParticipants() []memory.MemoryParticipant {
	return []memory.MemoryParticipant{{
		ID:       0,
		Nickname: "agent",
		Protocol: "",
	}}
}
