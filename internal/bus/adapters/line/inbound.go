package line

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	busruntime "github.com/quailyquaily/mistermorph/internal/bus"
	baseadapters "github.com/quailyquaily/mistermorph/internal/bus/adapters"
	"github.com/quailyquaily/mistermorph/internal/idempotency"
)

type InboundAdapterOptions struct {
	Bus   *busruntime.Inproc
	Store baseadapters.InboundStore
	Now   func() time.Time
}

// InboundMessage is the normalized line group-message event for bus ingress.
// V1 is intentionally group-only.
type InboundMessage struct {
	GroupID      string
	MessageID    string
	ReplyToken   string
	SentAt       time.Time
	ChatType     string
	FromUserID   string
	FromUsername string
	DisplayName  string
	Text         string
	MentionUsers []string
	ImagePaths   []string
	EventID      string
}

type InboundAdapter struct {
	flow  *baseadapters.InboundFlow
	nowFn func() time.Time
}

func NewInboundAdapter(opts InboundAdapterOptions) (*InboundAdapter, error) {
	flow, err := baseadapters.NewInboundFlow(baseadapters.InboundFlowOptions{
		Bus:     opts.Bus,
		Store:   opts.Store,
		Channel: string(busruntime.ChannelLine),
		Now:     opts.Now,
	})
	if err != nil {
		return nil, err
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	return &InboundAdapter{flow: flow, nowFn: nowFn}, nil
}

func (a *InboundAdapter) HandleInboundMessage(ctx context.Context, msg InboundMessage) (bool, error) {
	if a == nil || a.flow == nil {
		return false, fmt.Errorf("line inbound adapter is not initialized")
	}
	if ctx == nil {
		return false, fmt.Errorf("context is required")
	}
	groupID := strings.TrimSpace(msg.GroupID)
	if groupID == "" {
		return false, fmt.Errorf("group_id is required")
	}
	messageID := strings.TrimSpace(msg.MessageID)
	if messageID == "" {
		return false, fmt.Errorf("message_id is required")
	}
	chatType := strings.ToLower(strings.TrimSpace(msg.ChatType))
	if chatType == "" {
		return false, fmt.Errorf("chat_type is required")
	}
	if chatType != "group" {
		return false, fmt.Errorf("line chat_type %q is not supported in v1", chatType)
	}
	fromUserID := strings.TrimSpace(msg.FromUserID)
	if fromUserID == "" {
		return false, fmt.Errorf("from_user_id is required")
	}
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return false, fmt.Errorf("text is required")
	}
	mentionUsers, err := normalizeMentionUsers(msg.MentionUsers)
	if err != nil {
		return false, err
	}
	imagePaths, err := normalizeImagePaths(msg.ImagePaths)
	if err != nil {
		return false, err
	}

	now := a.nowFn().UTC()
	sentAt := msg.SentAt.UTC()
	if sentAt.IsZero() {
		sentAt = now
	}

	sessionUUID, err := uuid.NewV7()
	if err != nil {
		return false, err
	}
	sessionID := sessionUUID.String()
	envelopeMessageID := lineEnvelopeMessageID(groupID, messageID)
	payloadBase64, err := busruntime.EncodeMessageEnvelope(busruntime.TopicChatMessage, busruntime.MessageEnvelope{
		MessageID: envelopeMessageID,
		Text:      text,
		SentAt:    sentAt.Format(time.RFC3339),
		SessionID: sessionID,
		ReplyTo:   strings.TrimSpace(msg.ReplyToken),
	})
	if err != nil {
		return false, err
	}

	conversationKey, err := busruntime.BuildLineGroupConversationKey(groupID)
	if err != nil {
		return false, err
	}
	platformMessageID := linePlatformMessageID(groupID, messageID)

	busMsg := busruntime.BusMessage{
		ID:              "bus_" + uuid.NewString(),
		Direction:       busruntime.DirectionInbound,
		Channel:         busruntime.ChannelLine,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: conversationKey,
		ParticipantKey:  lineParticipantKey(fromUserID),
		IdempotencyKey:  idempotency.MessageEnvelopeKey(envelopeMessageID),
		CorrelationID:   "line:" + platformMessageID,
		PayloadBase64:   payloadBase64,
		CreatedAt:       sentAt,
		Extensions: busruntime.MessageExtensions{
			PlatformMessageID: platformMessageID,
			ReplyTo:           strings.TrimSpace(msg.ReplyToken),
			SessionID:         sessionID,
			ChatType:          chatType,
			FromUsername:      strings.TrimSpace(msg.FromUsername),
			FromDisplayName:   strings.TrimSpace(msg.DisplayName),
			ChannelID:         groupID,
			FromUserRef:       fromUserID,
			EventID:           strings.TrimSpace(msg.EventID),
			MentionUsers:      mentionUsers,
			ImagePaths:        imagePaths,
		},
	}
	return a.flow.PublishValidatedInbound(ctx, platformMessageID, busMsg)
}

func InboundMessageFromBusMessage(msg busruntime.BusMessage) (InboundMessage, error) {
	if msg.Direction != busruntime.DirectionInbound {
		return InboundMessage{}, fmt.Errorf("direction must be inbound")
	}
	if msg.Channel != busruntime.ChannelLine {
		return InboundMessage{}, fmt.Errorf("channel must be line")
	}
	groupID, err := groupIDFromConversationKey(msg.ConversationKey)
	if err != nil {
		return InboundMessage{}, err
	}
	pmGroupID, messageID, err := parseLinePlatformMessageID(msg.Extensions.PlatformMessageID)
	if err != nil {
		return InboundMessage{}, err
	}
	if pmGroupID != groupID {
		return InboundMessage{}, fmt.Errorf("platform_message_id does not match conversation_key")
	}
	env, err := msg.Envelope()
	if err != nil {
		return InboundMessage{}, err
	}
	sentAt, err := time.Parse(time.RFC3339, strings.TrimSpace(env.SentAt))
	if err != nil {
		return InboundMessage{}, fmt.Errorf("sent_at is invalid")
	}
	chatType := strings.ToLower(strings.TrimSpace(msg.Extensions.ChatType))
	if chatType == "" {
		return InboundMessage{}, fmt.Errorf("chat_type is required")
	}
	if chatType != "group" {
		return InboundMessage{}, fmt.Errorf("line chat_type %q is not supported in v1", chatType)
	}
	fromUserID := strings.TrimSpace(msg.Extensions.FromUserRef)
	if fromUserID == "" {
		fromUserID = lineUserIDFromParticipantKey(msg.ParticipantKey)
	}
	if fromUserID == "" {
		return InboundMessage{}, fmt.Errorf("from_user_id is required")
	}
	mentionUsers, err := normalizeMentionUsers(msg.Extensions.MentionUsers)
	if err != nil {
		return InboundMessage{}, err
	}
	imagePaths, err := normalizeImagePaths(msg.Extensions.ImagePaths)
	if err != nil {
		return InboundMessage{}, err
	}

	replyToken := strings.TrimSpace(msg.Extensions.ReplyTo)
	if replyToken == "" {
		replyToken = strings.TrimSpace(env.ReplyTo)
	}

	return InboundMessage{
		GroupID:      groupID,
		MessageID:    messageID,
		ReplyToken:   replyToken,
		SentAt:       sentAt.UTC(),
		ChatType:     chatType,
		FromUserID:   fromUserID,
		FromUsername: strings.TrimSpace(msg.Extensions.FromUsername),
		DisplayName:  strings.TrimSpace(msg.Extensions.FromDisplayName),
		Text:         strings.TrimSpace(env.Text),
		MentionUsers: mentionUsers,
		ImagePaths:   imagePaths,
		EventID:      strings.TrimSpace(msg.Extensions.EventID),
	}, nil
}

func lineEnvelopeMessageID(groupID, messageID string) string {
	return "line:" + linePlatformMessageID(groupID, messageID)
}

func linePlatformMessageID(groupID, messageID string) string {
	return strings.TrimSpace(groupID) + ":" + strings.TrimSpace(messageID)
}

func lineParticipantKey(userID string) string {
	return "user:" + strings.TrimSpace(userID)
}

func lineUserIDFromParticipantKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	const prefix = "user:"
	if !strings.HasPrefix(raw, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(raw, prefix))
}

func parseLinePlatformMessageID(platformMessageID string) (string, string, error) {
	platformMessageID = strings.TrimSpace(platformMessageID)
	if platformMessageID == "" {
		return "", "", fmt.Errorf("platform_message_id is required")
	}
	parts := strings.Split(platformMessageID, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("platform_message_id is invalid")
	}
	groupID := strings.TrimSpace(parts[0])
	messageID := strings.TrimSpace(parts[1])
	if groupID == "" || messageID == "" {
		return "", "", fmt.Errorf("platform_message_id is invalid")
	}
	return groupID, messageID, nil
}

func normalizeMentionUsers(items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		item := strings.TrimSpace(raw)
		if item == "" {
			return nil, fmt.Errorf("mention user is required")
		}
		out = append(out, item)
	}
	return out, nil
}

func normalizeImagePaths(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(paths))
	seen := make(map[string]bool, len(paths))
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			return nil, fmt.Errorf("image path is required")
		}
		if seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	return out, nil
}

func groupIDFromConversationKey(conversationKey string) (string, error) {
	const prefix = "line:"
	if !strings.HasPrefix(conversationKey, prefix) {
		return "", fmt.Errorf("line conversation key is invalid")
	}
	groupID := strings.TrimSpace(strings.TrimPrefix(conversationKey, prefix))
	if groupID == "" {
		return "", fmt.Errorf("line group id is required")
	}
	return groupID, nil
}
