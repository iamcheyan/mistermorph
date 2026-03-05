package line

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
	linebus "github.com/quailyquaily/mistermorph/internal/bus/adapters/line"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/idempotency"
	"github.com/quailyquaily/mistermorph/internal/promptprofile"
	"github.com/quailyquaily/mistermorph/internal/todo"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
	linetools "github.com/quailyquaily/mistermorph/tools/line"
)

type runtimeTaskOptions struct {
	SecretsRequireSkillProfiles bool
}

const lineStickySkillsCap = 16

func runLineTask(
	ctx context.Context,
	d Dependencies,
	logger *slog.Logger,
	logOpts agent.LogOptions,
	client llm.Client,
	baseReg *tools.Registry,
	api *lineAPI,
	sharedGuard *guard.Guard,
	cfg agent.Config,
	model string,
	job lineJob,
	history []chathistory.ChatHistoryItem,
	stickySkills []string,
	runtimeOpts runtimeTaskOptions,
) (*agent.Final, *agent.Context, []string, *linetools.Reaction, error) {
	task := strings.TrimSpace(job.Text)
	if task == "" {
		return nil, nil, nil, nil, fmt.Errorf("empty line task")
	}
	historyWithCurrent := append([]chathistory.ChatHistoryItem(nil), history...)
	historyWithCurrent = append(historyWithCurrent, newLineInboundHistoryItem(job))
	historyRaw, err := json.MarshalIndent(map[string]any{
		"chat_history_messages": chathistory.BuildMessages(chathistory.ChannelLine, historyWithCurrent),
	}, "", "  ")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("render line history context: %w", err)
	}
	llmHistory := []llm.Message{{Role: "user", Content: string(historyRaw)}}

	if baseReg == nil {
		return nil, nil, nil, nil, fmt.Errorf("base registry is nil")
	}
	reg := buildLineRegistry(baseReg, job.ChatType)
	toolsutil.RegisterRuntimeTools(reg, d.RuntimeToolsConfig, client, model)
	toolsutil.SetTodoUpdateToolAddContext(reg, todoResolveContextForLine(job))
	var reactTool *linetools.ReactTool
	if api != nil &&
		strings.TrimSpace(job.ChatID) != "" &&
		strings.TrimSpace(job.MessageID) != "" {
		reactTool = linetools.NewReactTool(newLineToolAPI(api), job.ChatID, job.MessageID, map[string]bool{
			job.ChatID: true,
		})
		reg.Register(reactTool)
	}

	promptSpec, loadedSkills, skillAuthProfiles, err := depsutil.PromptSpecFromCommon(d, ctx, logger, logOpts, task, client, model, stickySkills)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	promptprofile.ApplyPersonaIdentity(&promptSpec, logger)
	promptprofile.AppendLocalToolNotesBlock(&promptSpec, logger)
	promptprofile.AppendPlanCreateGuidanceBlock(&promptSpec, reg)
	promptprofile.AppendLineRuntimeBlocks(&promptSpec, isLineGroupChat(job.ChatType))

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
		"trigger":         "line",
		"line_chat_id":    job.ChatID,
		"line_chat_type":  job.ChatType,
		"line_user_id":    job.FromUserID,
		"line_message_id": job.MessageID,
	}
	final, runCtx, err := engine.Run(ctx, task, agent.RunOptions{
		Model:           model,
		History:         llmHistory,
		Meta:            meta,
		SkipTaskMessage: true,
	})
	if err != nil {
		return final, runCtx, loadedSkills, nil, err
	}

	var reaction *linetools.Reaction
	if reactTool != nil {
		reaction = reactTool.LastReaction()
		if reaction != nil && logger != nil {
			logger.Info("message_reaction_applied",
				"chat_id", reaction.ChatID,
				"message_id", reaction.MessageID,
				"emoji", reaction.Emoji,
				"source", reaction.Source,
			)
		}
	}
	return final, runCtx, loadedSkills, reaction, nil
}

func todoResolveContextForLine(job lineJob) todo.AddResolveContext {
	speaker := strings.TrimSpace(job.FromUserID)
	if speaker != "" {
		speaker = "line:" + speaker
	}
	mentions := normalizeLineMentionUsersForTodo(job.MentionUsers)
	return todo.AddResolveContext{
		Channel:          "line",
		ChatType:         strings.ToLower(strings.TrimSpace(job.ChatType)),
		SpeakerUsername:  speaker,
		MentionUsernames: mentions,
		UserInputRaw:     job.Text,
	}
}

func normalizeLineMentionUsersForTodo(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		out = append(out, "line:"+item)
	}
	return out
}

func newLineInboundHistoryItem(job lineJob) chathistory.ChatHistoryItem {
	return chathistory.ChatHistoryItem{
		Channel:          chathistory.ChannelLine,
		Kind:             chathistory.KindInboundUser,
		ChatID:           "line:" + strings.TrimSpace(job.ChatID),
		ChatType:         strings.ToLower(strings.TrimSpace(job.ChatType)),
		MessageID:        strings.TrimSpace(job.MessageID),
		ReplyToMessageID: strings.TrimSpace(job.ReplyToken),
		SentAt:           job.SentAt.UTC(),
		Sender:           lineSenderFromJob(job, false),
		Text:             strings.TrimSpace(job.Text),
	}
}

func lineJobFromInbound(inbound linebus.InboundMessage) lineJob {
	return lineJob{
		ChatID:       strings.TrimSpace(inbound.ChatID),
		ChatType:     strings.TrimSpace(inbound.ChatType),
		MessageID:    strings.TrimSpace(inbound.MessageID),
		ReplyToken:   strings.TrimSpace(inbound.ReplyToken),
		FromUserID:   strings.TrimSpace(inbound.FromUserID),
		FromUsername: strings.TrimSpace(inbound.FromUsername),
		DisplayName:  strings.TrimSpace(inbound.DisplayName),
		Text:         strings.TrimSpace(inbound.Text),
		SentAt:       inbound.SentAt.UTC(),
	}
}

func newLineInboundHistoryItemFromInbound(inbound linebus.InboundMessage) chathistory.ChatHistoryItem {
	return newLineInboundHistoryItem(lineJobFromInbound(inbound))
}

func newLineOutboundAgentHistoryItem(job lineJob, output string, sentAt time.Time) chathistory.ChatHistoryItem {
	return chathistory.ChatHistoryItem{
		Channel:          chathistory.ChannelLine,
		Kind:             chathistory.KindOutboundAgent,
		ChatID:           "line:" + strings.TrimSpace(job.ChatID),
		ChatType:         strings.ToLower(strings.TrimSpace(job.ChatType)),
		ReplyToMessageID: strings.TrimSpace(job.ReplyToken),
		SentAt:           sentAt.UTC(),
		Sender:           lineSenderFromJob(job, true),
		Text:             strings.TrimSpace(output),
	}
}

func newLineOutboundReactionHistoryItem(job lineJob, note, emoji string, sentAt time.Time) chathistory.ChatHistoryItem {
	item := newLineOutboundAgentHistoryItem(job, note, sentAt)
	item.Kind = chathistory.KindOutboundReaction
	if strings.TrimSpace(emoji) != "" {
		item.Text = strings.TrimSpace(note)
	}
	return item
}

func lineSenderFromJob(job lineJob, isBot bool) chathistory.ChatHistorySender {
	if isBot {
		return chathistory.ChatHistorySender{
			Username:   "line-bot",
			Nickname:   "line-bot",
			IsBot:      true,
			DisplayRef: "line-bot",
		}
	}
	nickname := strings.TrimSpace(job.DisplayName)
	if nickname == "" {
		nickname = strings.TrimSpace(job.FromUsername)
	}
	if nickname == "" {
		nickname = strings.TrimSpace(job.FromUserID)
	}
	username := strings.TrimSpace(job.FromUsername)
	if username == "" {
		username = strings.TrimSpace(job.FromUserID)
	}
	displayRef := nickname
	if displayRef == "" {
		displayRef = username
	}
	return chathistory.ChatHistorySender{
		UserID:     strings.TrimSpace(job.FromUserID),
		Username:   username,
		Nickname:   nickname,
		IsBot:      false,
		DisplayRef: displayRef,
	}
}

func lineHistoryCapForMode(mode string) int {
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

func publishLineBusOutbound(ctx context.Context, inprocBus *busruntime.Inproc, chatID, text, replyToken, correlationID string) (string, error) {
	if inprocBus == nil {
		return "", fmt.Errorf("bus is required")
	}
	if ctx == nil {
		return "", fmt.Errorf("context is required")
	}
	chatID = strings.TrimSpace(chatID)
	text = strings.TrimSpace(text)
	replyToken = strings.TrimSpace(replyToken)
	if chatID == "" {
		return "", fmt.Errorf("chat_id is required")
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
		ReplyTo:   replyToken,
	})
	if err != nil {
		return "", err
	}
	conversationKey, err := busruntime.BuildLineConversationKey(chatID)
	if err != nil {
		return "", err
	}
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		correlationID = "line:" + messageID
	}
	outbound := busruntime.BusMessage{
		ID:              "bus_" + uuid.NewString(),
		Direction:       busruntime.DirectionOutbound,
		Channel:         busruntime.ChannelLine,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: conversationKey,
		IdempotencyKey:  idempotency.MessageEnvelopeKey(messageID),
		CorrelationID:   correlationID,
		PayloadBase64:   payloadBase64,
		CreatedAt:       now,
		Extensions: busruntime.MessageExtensions{
			SessionID: sessionID,
			ReplyTo:   replyToken,
			ChannelID: chatID,
		},
	}
	if err := inprocBus.PublishValidated(ctx, outbound); err != nil {
		return "", err
	}
	return messageID, nil
}

func isLineGroupChat(chatType string) bool {
	return strings.EqualFold(strings.TrimSpace(chatType), "group")
}

func buildLineRegistry(baseReg *tools.Registry, chatType string) *tools.Registry {
	reg := tools.NewRegistry()
	if baseReg == nil {
		return reg
	}
	groupChat := isLineGroupChat(chatType)
	for _, t := range baseReg.All() {
		name := strings.TrimSpace(t.Name())
		if groupChat && strings.EqualFold(name, "contacts_send") {
			continue
		}
		reg.Register(t)
	}
	return reg
}

func shouldPublishLineText(final *agent.Final) bool {
	if final == nil {
		return true
	}
	return !final.IsLightweight
}
