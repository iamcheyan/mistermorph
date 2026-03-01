package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/grouptrigger"
	"github.com/quailyquaily/mistermorph/internal/llminspect"
	"github.com/quailyquaily/mistermorph/internal/promptprofile"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type slackGroupTriggerDecision = grouptrigger.Decision

func quoteReplyThreadTSForGroupTrigger(event slackInboundEvent, dec slackGroupTriggerDecision) string {
	threadTS := strings.TrimSpace(event.ThreadTS)
	if threadTS != "" {
		return threadTS
	}
	if dec.Addressing.Impulse > 0.8 {
		return strings.TrimSpace(event.MessageTS)
	}
	return ""
}

func decideSlackGroupTrigger(
	ctx context.Context,
	client llm.Client,
	model string,
	event slackInboundEvent,
	botUserID string,
	mode string,
	addressingLLMTimeout time.Duration,
	addressingConfidenceThreshold float64,
	addressingInterjectThreshold float64,
	history []chathistory.ChatHistoryItem,
	addressingReactionTool tools.Tool,
) (slackGroupTriggerDecision, bool, error) {
	explicitReason, explicitMentioned := slackExplicitMentionReason(event, botUserID)
	return grouptrigger.Decide(ctx, grouptrigger.DecideOptions{
		Mode:                     mode,
		ConfidenceThreshold:      addressingConfidenceThreshold,
		InterjectThreshold:       addressingInterjectThreshold,
		ExplicitReason:           explicitReason,
		ExplicitMatched:          explicitMentioned,
		AddressingFallbackReason: mode,
		AddressingTimeout:        addressingLLMTimeout,
		Addressing: func(addrCtx context.Context) (grouptrigger.Addressing, bool, error) {
			return slackAddressingDecisionViaLLM(addrCtx, client, model, event, history, addressingReactionTool)
		},
	})
}

func slackExplicitMentionReason(event slackInboundEvent, botUserID string) (string, bool) {
	if event.IsAppMention {
		return "app_mention", true
	}
	if strings.TrimSpace(botUserID) != "" && strings.Contains(event.Text, "<@"+strings.TrimSpace(botUserID)+">") {
		return "mention", true
	}
	return "", false
}

func slackAddressingDecisionViaLLM(ctx context.Context, client llm.Client, model string, event slackInboundEvent, history []chathistory.ChatHistoryItem, addressingTool tools.Tool) (grouptrigger.Addressing, bool, error) {
	if ctx == nil || client == nil {
		return grouptrigger.Addressing{}, false, nil
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return grouptrigger.Addressing{}, false, fmt.Errorf("missing model for addressing_llm")
	}
	personaIdentity := loadAddressingPersonaIdentity()
	historyMessages := chathistory.BuildMessages(chathistory.ChannelSlack, history)
	currentMessage := map[string]any{
		"team_id":       event.TeamID,
		"channel_id":    event.ChannelID,
		"chat_type":     event.ChatType,
		"message_ts":    event.MessageTS,
		"thread_ts":     event.ThreadTS,
		"user_id":       event.UserID,
		"text":          event.Text,
		"mention_users": append([]string(nil), event.MentionUsers...),
	}
	systemPrompt, userPrompt, err := grouptrigger.RenderAddressingPrompts(personaIdentity, currentMessage, historyMessages)
	if err != nil {
		return grouptrigger.Addressing{}, false, fmt.Errorf("render addressing prompts: %w", err)
	}
	return grouptrigger.DecideViaLLM(llminspect.WithModelScene(ctx, "slack.addressing_decision"), grouptrigger.LLMDecisionOptions{
		Client:         client,
		Model:          model,
		SystemPrompt:   systemPrompt,
		UserPrompt:     userPrompt,
		AddressingTool: addressingTool,
		MaxToolRounds:  3,
	})
}

func loadAddressingPersonaIdentity() string {
	spec := agent.PromptSpec{}
	promptprofile.ApplyPersonaIdentity(&spec, slog.Default())
	return strings.TrimSpace(spec.Identity)
}
