package heartbeat

import (
	"fmt"
	"time"
)

const (
	heartbeatMemorySubjectID = "heartbeat"
	heartbeatMemorySessionID = "heartbeat"
)

func heartbeatTaskRunID(now time.Time) string {
	now = now.UTC()
	return fmt.Sprintf("heartbeat:%s", now.Format("20060102T150405.000000000Z07:00"))
}
