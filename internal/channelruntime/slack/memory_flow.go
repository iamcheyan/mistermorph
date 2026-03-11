package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/memory"
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

func buildSlackMemoryHistory(history []chathistory.ChatHistoryItem, job slackJob, output string, sentAt time.Time, maxItems int) []chathistory.ChatHistoryItem {
	out := append([]chathistory.ChatHistoryItem{}, history...)
	out = append(out, newSlackInboundHistoryItem(job))
	if strings.TrimSpace(output) != "" {
		out = append(out, newSlackOutboundAgentHistoryItem(job, output, sentAt, ""))
	}
	if maxItems > 0 && len(out) > maxItems {
		out = out[len(out)-maxItems:]
	}
	return out
}

func slackMemorySessionContext(job slackJob) memory.SessionContext {
	ctx := memory.SessionContext{
		ConversationID:   strings.TrimSpace(job.ChannelID),
		ConversationType: strings.ToLower(strings.TrimSpace(job.ChatType)),
		CounterpartyID:   strings.TrimSpace(job.UserID),
		CounterpartyName: strings.TrimSpace(job.DisplayName),
	}
	if username := strings.TrimSpace(job.Username); username != "" {
		ctx.CounterpartyHandle = username
	}
	ctx.CounterpartyLabel = slackMemoryCounterpartyLabel(job)
	return ctx
}

func slackMemoryCounterpartyLabel(job slackJob) string {
	id := strings.TrimSpace(job.UserID)
	name := strings.TrimSpace(job.DisplayName)
	if name == "" {
		name = strings.TrimSpace(job.Username)
	}
	if id != "" && name != "" {
		return "[" + name + "](slack:" + id + ")"
	}
	if name != "" {
		return name
	}
	if id != "" {
		return "slack:" + id
	}
	return ""
}
