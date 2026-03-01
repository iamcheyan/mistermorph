package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type ReactTool struct {
	api               API
	defaultChannelID  string
	defaultMessageTS  string
	allowedChannelIDs map[string]bool
	lastReaction      *Reaction
}

var slackEmojiNamePattern = regexp.MustCompile(`^[A-Za-z0-9_+\-]+$`)

func NewReactTool(api API, defaultChannelID, defaultMessageTS string, allowedChannelIDs map[string]bool) *ReactTool {
	allowed := make(map[string]bool, len(allowedChannelIDs))
	for raw := range allowedChannelIDs {
		channelID := strings.TrimSpace(raw)
		if channelID == "" {
			continue
		}
		allowed[channelID] = true
	}
	return &ReactTool{
		api:               api,
		defaultChannelID:  strings.TrimSpace(defaultChannelID),
		defaultMessageTS:  strings.TrimSpace(defaultMessageTS),
		allowedChannelIDs: allowed,
	}
}

func (t *ReactTool) Name() string { return "slack_react" }

func (t *ReactTool) Description() string {
	return "Adds an emoji reaction to a Slack message. Prefer this for lightweight acknowledgements."
}

func (t *ReactTool) ParameterSchema() string {
	s := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"channel_id": map[string]any{
				"type":        "string",
				"description": "Target Slack channel id. Optional in active channel context.",
			},
			"message_ts": map[string]any{
				"type":        "string",
				"description": "Target Slack message ts. Optional in active channel context; defaults to the triggering message.",
			},
			"emoji": map[string]any{
				"type":        "string",
				"description": "Emoji name, for example thumbsup or :thumbsup:.",
			},
		},
		"required": []string{"emoji"},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

func (t *ReactTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	if t == nil || t.api == nil {
		return "", fmt.Errorf("slack_react is disabled")
	}

	channelID := strings.TrimSpace(t.defaultChannelID)
	if v, ok := params["channel_id"].(string); ok {
		channelID = strings.TrimSpace(v)
	}
	if channelID == "" {
		return "", fmt.Errorf("missing required param: channel_id")
	}
	if len(t.allowedChannelIDs) > 0 && !t.allowedChannelIDs[channelID] {
		return "", fmt.Errorf("unauthorized channel_id: %s", channelID)
	}

	messageTS := strings.TrimSpace(t.defaultMessageTS)
	if v, ok := params["message_ts"].(string); ok {
		messageTS = strings.TrimSpace(v)
	}
	if messageTS == "" {
		return "", fmt.Errorf("missing required param: message_ts")
	}

	rawEmoji, _ := params["emoji"].(string)
	emoji, err := normalizeSlackEmojiName(rawEmoji)
	if err != nil {
		return "", err
	}

	if err := t.api.AddReaction(ctx, channelID, messageTS, emoji); err != nil {
		return "", err
	}
	t.lastReaction = &Reaction{
		ChannelID: channelID,
		MessageTS: messageTS,
		Emoji:     emoji,
		Source:    "tool",
	}
	return fmt.Sprintf("reacted with :%s:", emoji), nil
}

func (t *ReactTool) LastReaction() *Reaction {
	if t == nil {
		return nil
	}
	return t.lastReaction
}

func normalizeSlackEmojiName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if strings.HasPrefix(name, ":") && strings.HasSuffix(name, ":") && len(name) >= 2 {
		name = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(name, ":"), ":"))
	}
	if name == "" {
		return "", fmt.Errorf("missing required param: emoji")
	}
	if strings.ContainsAny(name, " \t\r\n") {
		return "", fmt.Errorf("emoji name is invalid: %s", raw)
	}
	if !slackEmojiNamePattern.MatchString(name) {
		return "", fmt.Errorf("emoji name is invalid: %s", raw)
	}
	return strings.ToLower(name), nil
}
