package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/guard"
	busruntime "github.com/quailyquaily/mistermorph/internal/bus"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/idempotency"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/internal/promptprofile"
	"github.com/quailyquaily/mistermorph/internal/todo"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type runtimeTaskOptions struct {
	SecretsRequireSkillProfiles bool
	MemoryEnabled               bool
	MemoryInjectionEnabled      bool
	MemoryInjectionMaxItems     int
	MemoryOrchestrator          *memoryruntime.Orchestrator
	MemoryProjectionWorker      *memoryruntime.ProjectionWorker
}

func runSlackTask(
	ctx context.Context,
	d Dependencies,
	logger *slog.Logger,
	logOpts agent.LogOptions,
	client llm.Client,
	baseReg *tools.Registry,
	sharedGuard *guard.Guard,
	cfg agent.Config,
	model string,
	job slackJob,
	history []chathistory.ChatHistoryItem,
	stickySkills []string,
	runtimeOpts runtimeTaskOptions,
) (*agent.Final, *agent.Context, []string, error) {
	task := strings.TrimSpace(job.Text)
	if task == "" {
		return nil, nil, nil, fmt.Errorf("empty slack task")
	}
	historyWithCurrent := append([]chathistory.ChatHistoryItem(nil), history...)
	historyWithCurrent = append(historyWithCurrent, newSlackInboundHistoryItem(job))
	historyRaw, err := json.MarshalIndent(map[string]any{
		"chat_history_messages": chathistory.BuildMessages(chathistory.ChannelSlack, historyWithCurrent),
	}, "", "  ")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("render slack history context: %w", err)
	}
	llmHistory := []llm.Message{{Role: "user", Content: string(historyRaw)}}

	if baseReg == nil {
		return nil, nil, nil, fmt.Errorf("base registry is nil")
	}
	reg := buildSlackRegistry(baseReg, job.ChatType)
	toolsutil.RegisterRuntimeTools(reg, d.RuntimeToolsConfig, client, model)
	toolsutil.SetTodoUpdateToolAddContext(reg, todoResolveContextForSlack(job))

	promptSpec, loadedSkills, skillAuthProfiles, err := depsutil.PromptSpecFromCommon(d, ctx, logger, logOpts, task, client, model, stickySkills)
	if err != nil {
		return nil, nil, nil, err
	}
	promptprofile.ApplyPersonaIdentity(&promptSpec, logger)
	promptprofile.AppendLocalToolNotesBlock(&promptSpec, logger)
	promptprofile.AppendPlanCreateGuidanceBlock(&promptSpec, reg)
	promptprofile.AppendSlackRuntimeBlocks(&promptSpec, isSlackGroupChat(job.ChatType), job.MentionUsers)

	memSubjectID := slackMemorySubjectID(job)
	if runtimeOpts.MemoryEnabled && runtimeOpts.MemoryOrchestrator != nil && memSubjectID != "" && runtimeOpts.MemoryInjectionEnabled {
		snap, memErr := runtimeOpts.MemoryOrchestrator.PrepareInjection(memoryruntime.PrepareInjectionRequest{
			SubjectID:      memSubjectID,
			RequestContext: slackMemoryRequestContext(job.ChatType),
			MaxItems:       runtimeOpts.MemoryInjectionMaxItems,
		})
		if memErr != nil {
			if logger != nil {
				logger.Warn("memory_injection_error", "source", "slack", "subject_id", memSubjectID, "error", memErr.Error())
			}
		} else if strings.TrimSpace(snap) != "" {
			promptprofile.AppendMemorySummariesBlock(&promptSpec, snap)
			if logger != nil {
				logger.Info("memory_injection_applied", "source", "slack", "subject_id", memSubjectID, "channel_id", job.ChannelID, "snapshot_len", len(snap))
			}
		}
	}

	engine := agent.New(
		client,
		reg,
		cfg,
		promptSpec,
		agent.WithLogger(logger),
		agent.WithLogOptions(logOpts),
		agent.WithSkillAuthProfiles(skillAuthProfiles, runtimeOpts.SecretsRequireSkillProfiles),
		agent.WithGuard(sharedGuard),
	)

	meta := map[string]any{
		"trigger":            "slack",
		"slack_team_id":      job.TeamID,
		"slack_channel_id":   job.ChannelID,
		"slack_chat_type":    job.ChatType,
		"slack_message_ts":   job.MessageTS,
		"slack_thread_ts":    job.ThreadTS,
		"slack_from_user_id": job.UserID,
	}
	final, runCtx, err := engine.Run(ctx, task, agent.RunOptions{
		Model:           model,
		History:         llmHistory,
		Meta:            meta,
		SkipTaskMessage: true,
	})
	if err != nil {
		return final, runCtx, loadedSkills, err
	}

	if runtimeOpts.MemoryEnabled && runtimeOpts.MemoryOrchestrator != nil && memSubjectID != "" {
		finalOutput := strings.TrimSpace(depsutil.FormatFinalOutput(final))
		recordOffset, memErr := runtimeOpts.MemoryOrchestrator.Record(memoryruntime.RecordRequest{
			TaskRunID:    slackMemoryTaskRunID(job),
			SessionID:    slackMemorySessionID(job),
			SubjectID:    memSubjectID,
			Channel:      "slack",
			Participants: slackMemoryParticipants(job),
			TaskText:     task,
			FinalOutput:  finalOutput,
			Draft:        buildSlackMemoryDraft(finalOutput),
		})
		if memErr != nil {
			if logger != nil {
				logger.Warn("memory_record_error", "source", "slack", "subject_id", memSubjectID, "error", memErr.Error())
			}
		} else {
			if logger != nil {
				logger.Debug("memory_record_ok", "source", "slack", "subject_id", memSubjectID, "offset_file", recordOffset.File, "offset_line", recordOffset.Line)
			}
			if runtimeOpts.MemoryProjectionWorker != nil {
				runtimeOpts.MemoryProjectionWorker.NotifyRecordAppended()
			}
		}
	}

	return final, runCtx, loadedSkills, nil
}

func todoResolveContextForSlack(job slackJob) todo.AddResolveContext {
	user := strings.TrimSpace(job.UserID)
	if user != "" {
		user = "slack:" + user
	}
	mentions := normalizeMentionUsersForTodo(job.MentionUsers)
	return todo.AddResolveContext{
		Channel:          "slack",
		ChatType:         strings.ToLower(strings.TrimSpace(job.ChatType)),
		SpeakerUsername:  user,
		MentionUsernames: mentions,
		UserInputRaw:     job.Text,
	}
}

func normalizeMentionUsersForTodo(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		out = append(out, "slack:"+item)
	}
	return out
}

func newSlackInboundHistoryItem(job slackJob) chathistory.ChatHistoryItem {
	return chathistory.ChatHistoryItem{
		Channel:          chathistory.ChannelSlack,
		Kind:             chathistory.KindInboundUser,
		ChatID:           "slack:" + job.TeamID + ":" + job.ChannelID,
		ChatType:         strings.ToLower(strings.TrimSpace(job.ChatType)),
		MessageID:        strings.TrimSpace(job.MessageTS),
		ReplyToMessageID: strings.TrimSpace(job.ThreadTS),
		SentAt:           job.SentAt.UTC(),
		Sender:           slackSenderFromJob(job, false, ""),
		Text:             strings.TrimSpace(job.Text),
	}
}

func newSlackOutboundAgentHistoryItem(job slackJob, output string, sentAt time.Time, botUserID string) chathistory.ChatHistoryItem {
	return chathistory.ChatHistoryItem{
		Channel:          chathistory.ChannelSlack,
		Kind:             chathistory.KindOutboundAgent,
		ChatID:           "slack:" + job.TeamID + ":" + job.ChannelID,
		ChatType:         strings.ToLower(strings.TrimSpace(job.ChatType)),
		ReplyToMessageID: strings.TrimSpace(job.ThreadTS),
		SentAt:           sentAt.UTC(),
		Sender:           slackSenderFromJob(job, true, botUserID),
		Text:             strings.TrimSpace(output),
	}
}

func slackSenderFromJob(job slackJob, isBot bool, botUserID string) chathistory.ChatHistorySender {
	if isBot {
		return chathistory.ChatHistorySender{
			UserID:     strings.TrimSpace(botUserID),
			Username:   "slack-bot",
			Nickname:   "slack-bot",
			IsBot:      true,
			DisplayRef: "slack-bot",
		}
	}
	mentionRef := strings.TrimSpace(job.UserID)
	if mentionRef != "" {
		mentionRef = "<@" + mentionRef + ">"
	}
	nickname := strings.TrimSpace(job.DisplayName)
	if nickname == "" {
		nickname = mentionRef
	}
	displayRef := mentionRef
	if nickname != "" && mentionRef != "" && nickname != mentionRef {
		displayRef = nickname + " (" + mentionRef + ")"
	} else if nickname != "" {
		displayRef = nickname
	}
	username := strings.TrimSpace(job.Username)
	if username == "" {
		username = strings.TrimSpace(job.UserID)
	}
	return chathistory.ChatHistorySender{
		UserID:     strings.TrimSpace(job.UserID),
		Username:   username,
		Nickname:   nickname,
		IsBot:      false,
		DisplayRef: displayRef,
	}
}

func slackHistoryCapForMode(mode string) int {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "talkative":
		return 16
	default:
		return 8
	}
}

func trimChatHistoryItems(items []chathistory.ChatHistoryItem, limit int) []chathistory.ChatHistoryItem {
	if limit <= 0 || len(items) <= limit {
		return append([]chathistory.ChatHistoryItem(nil), items...)
	}
	return append([]chathistory.ChatHistoryItem(nil), items[len(items)-limit:]...)
}

func capUniqueStrings(items []string, limit int) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(items))
	out := make([]string, 0, len(items))
	for _, raw := range items {
		item := strings.TrimSpace(raw)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func buildSlackConversationKey(teamID, channelID string) (string, error) {
	return busruntime.BuildSlackChannelConversationKey(strings.TrimSpace(teamID) + ":" + strings.TrimSpace(channelID))
}

func busErrorCodeString(err error) string {
	if err == nil {
		return ""
	}
	return string(busruntime.ErrorCodeOf(err))
}

func publishSlackBusOutbound(ctx context.Context, inprocBus *busruntime.Inproc, teamID, channelID, text, threadTS, correlationID string) (string, error) {
	if inprocBus == nil {
		return "", fmt.Errorf("bus is required")
	}
	if ctx == nil {
		return "", fmt.Errorf("context is required")
	}
	teamID = strings.TrimSpace(teamID)
	channelID = strings.TrimSpace(channelID)
	text = strings.TrimSpace(text)
	threadTS = strings.TrimSpace(threadTS)
	if teamID == "" {
		return "", fmt.Errorf("team_id is required")
	}
	if channelID == "" {
		return "", fmt.Errorf("channel_id is required")
	}
	if text == "" {
		return "", fmt.Errorf("text is required")
	}

	now := time.Now().UTC()
	messageID := "msg_" + uuid.NewString()
	sessionUUID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	sessionID := sessionUUID.String()
	payloadBase64, err := busruntime.EncodeMessageEnvelope(busruntime.TopicChatMessage, busruntime.MessageEnvelope{
		MessageID: messageID,
		Text:      text,
		SentAt:    now.Format(time.RFC3339),
		SessionID: sessionID,
		ReplyTo:   threadTS,
	})
	if err != nil {
		return "", err
	}
	conversationKey, err := buildSlackConversationKey(teamID, channelID)
	if err != nil {
		return "", err
	}
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		correlationID = "slack:" + messageID
	}
	outbound := busruntime.BusMessage{
		ID:              "bus_" + uuid.NewString(),
		Direction:       busruntime.DirectionOutbound,
		Channel:         busruntime.ChannelSlack,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: conversationKey,
		IdempotencyKey:  idempotency.MessageEnvelopeKey(messageID),
		CorrelationID:   correlationID,
		PayloadBase64:   payloadBase64,
		CreatedAt:       now,
		Extensions: busruntime.MessageExtensions{
			SessionID: sessionID,
			ReplyTo:   threadTS,
			ThreadTS:  threadTS,
			TeamID:    teamID,
			ChannelID: channelID,
		},
	}
	if err := inprocBus.PublishValidated(ctx, outbound); err != nil {
		return "", err
	}
	return messageID, nil
}

func buildSlackRegistry(baseReg *tools.Registry, chatType string) *tools.Registry {
	reg := tools.NewRegistry()
	if baseReg == nil {
		return reg
	}
	groupChat := isSlackGroupChat(chatType)
	for _, t := range baseReg.All() {
		name := strings.TrimSpace(t.Name())
		if groupChat && strings.EqualFold(name, "contacts_send") {
			continue
		}
		reg.Register(t)
	}
	return reg
}
