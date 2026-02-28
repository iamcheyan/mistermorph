package heartbeat

import (
	"fmt"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/memory"
)

const (
	heartbeatMemorySubjectID    = "heartbeat"
	heartbeatMemorySessionID    = "heartbeat"
	heartbeatMemorySummaryRunes = 1024
)

func heartbeatTaskRunID(now time.Time) string {
	now = now.UTC()
	return fmt.Sprintf("heartbeat:%s", now.Format("20060102T150405.000000000Z07:00"))
}

func buildHeartbeatDraft(summary string) memory.SessionDraft {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return memory.SessionDraft{}
	}
	summary = strings.Join(strings.Fields(summary), " ")
	if summary == "" {
		return memory.SessionDraft{}
	}
	runes := []rune(summary)
	if len(runes) > heartbeatMemorySummaryRunes {
		summary = strings.TrimSpace(string(runes[:heartbeatMemorySummaryRunes]))
	}
	if summary == "" {
		return memory.SessionDraft{}
	}
	return memory.SessionDraft{SummaryItems: []string{summary}}
}
