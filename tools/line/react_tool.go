package line

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

type ReactTool struct {
	api            API
	defaultChatID  string
	defaultMessage string
	allowedChatIDs map[string]bool
	lastReaction   *Reaction
}

func NewReactTool(api API, defaultChatID, defaultMessageID string, allowedChatIDs map[string]bool) *ReactTool {
	allowed := make(map[string]bool, len(allowedChatIDs))
	for raw := range allowedChatIDs {
		chatID := strings.TrimSpace(raw)
		if chatID == "" {
			continue
		}
		allowed[chatID] = true
	}
	return &ReactTool{
		api:            api,
		defaultChatID:  strings.TrimSpace(defaultChatID),
		defaultMessage: strings.TrimSpace(defaultMessageID),
		allowedChatIDs: allowed,
	}
}

func (t *ReactTool) Name() string { return "message_react" }

func (t *ReactTool) Description() string {
	return "Adds an emoji reaction to a LINE message. Use this for lightweight acknowledgement when text reply is unnecessary."
}

func (t *ReactTool) ParameterSchema() string {
	s := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"chat_id": map[string]any{
				"type":        "string",
				"description": "Target LINE chat id. Optional in active chat context.",
			},
			"message_id": map[string]any{
				"type":        "string",
				"description": "Target LINE message id. Optional in active chat context; defaults to the triggering message.",
			},
			"emoji": map[string]any{
				"type":        "string",
				"description": "Emoji character to react with, for example 👍 or 🎉.",
			},
		},
		"required": []string{"emoji"},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

func (t *ReactTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	if t == nil || t.api == nil {
		return "", fmt.Errorf("message_react is disabled")
	}

	chatID := strings.TrimSpace(t.defaultChatID)
	if v, ok := params["chat_id"].(string); ok {
		chatID = strings.TrimSpace(v)
	}
	if chatID == "" {
		return "", fmt.Errorf("missing required param: chat_id")
	}
	if len(t.allowedChatIDs) > 0 && !t.allowedChatIDs[chatID] {
		return "", fmt.Errorf("unauthorized chat_id: %s", chatID)
	}

	messageID := strings.TrimSpace(t.defaultMessage)
	if v, ok := params["message_id"].(string); ok {
		messageID = strings.TrimSpace(v)
	}
	if messageID == "" {
		return "", fmt.Errorf("missing required param: message_id")
	}

	rawEmoji, _ := params["emoji"].(string)
	emoji, err := normalizeLineReactionEmoji(rawEmoji)
	if err != nil {
		return "", err
	}

	if err := t.api.AddReaction(ctx, chatID, messageID, emoji); err != nil {
		return "", err
	}
	t.lastReaction = &Reaction{
		ChatID:    chatID,
		MessageID: messageID,
		Emoji:     emoji,
		Source:    "tool",
	}
	return fmt.Sprintf("reacted with %s", emoji), nil
}

func (t *ReactTool) LastReaction() *Reaction {
	if t == nil {
		return nil
	}
	return t.lastReaction
}

func normalizeLineReactionEmoji(raw string) (string, error) {
	emoji := strings.TrimSpace(raw)
	if emoji == "" {
		return "", fmt.Errorf("missing required param: emoji")
	}
	if strings.ContainsAny(emoji, " \t\r\n") {
		return "", fmt.Errorf("emoji is invalid: %s", raw)
	}
	runeCount := utf8.RuneCountInString(emoji)
	if runeCount <= 0 || runeCount > 8 {
		return "", fmt.Errorf("emoji is invalid: %s", raw)
	}
	return emoji, nil
}
