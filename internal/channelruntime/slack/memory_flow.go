package slack

import (
	"fmt"
	"strings"

	"github.com/quailyquaily/mistermorph/memory"
)

const (
	slackMemorySummaryMaxRunes = 1024
)

func slackMemorySubjectID(job slackJob) string {
	teamID := strings.ToLower(strings.TrimSpace(job.TeamID))
	channelID := strings.ToLower(strings.TrimSpace(job.ChannelID))
	if teamID == "" || channelID == "" {
		return ""
	}
	return memory.SanitizeSubjectID("slack--" + teamID + "--" + channelID)
}

func slackMemorySessionID(job slackJob) string {
	teamID := strings.TrimSpace(job.TeamID)
	channelID := strings.TrimSpace(job.ChannelID)
	return fmt.Sprintf("slack:%s:%s", teamID, channelID)
}

func slackMemoryTaskRunID(job slackJob) string {
	return strings.TrimSpace(job.TaskID)
}

func slackMemoryRequestContext(chatType string) memory.RequestContext {
	switch strings.ToLower(strings.TrimSpace(chatType)) {
	case "im":
		return memory.ContextPrivate
	default:
		return memory.ContextPublic
	}
}

func slackMemoryParticipants(job slackJob) []memory.MemoryParticipant {
	seen := map[string]bool{}
	out := make([]memory.MemoryParticipant, 0, 1+len(job.MentionUsers))
	appendParticipant := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, memory.MemoryParticipant{
			ID:       id,
			Nickname: id,
			Protocol: "slack",
		})
	}

	appendParticipant(job.UserID)
	for _, user := range job.MentionUsers {
		appendParticipant(user)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildSlackMemoryDraft(finalOutput string) memory.SessionDraft {
	item := strings.TrimSpace(finalOutput)
	if item == "" {
		return memory.SessionDraft{}
	}
	item = strings.Join(strings.Fields(item), " ")
	if item == "" {
		return memory.SessionDraft{}
	}
	runes := []rune(item)
	if len(runes) > slackMemorySummaryMaxRunes {
		item = strings.TrimSpace(string(runes[:slackMemorySummaryMaxRunes]))
	}
	if item == "" {
		return memory.SessionDraft{}
	}
	return memory.SessionDraft{
		SummaryItems: []string{item},
	}
}
