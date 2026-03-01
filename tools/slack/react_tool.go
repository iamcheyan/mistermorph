package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ReactTool struct {
	api               API
	defaultChannelID  string
	defaultMessageTS  string
	allowedChannelIDs map[string]bool
	availableEmojis   []string
	availableEmojiSet map[string]bool
	lastReaction      *Reaction
}

var slackEmojiNamePattern = regexp.MustCompile(`^[A-Za-z0-9_+\-]+$`)

const slackEmojiPreviewLimit = 120

func NewReactTool(api API, defaultChannelID, defaultMessageTS string, allowedChannelIDs map[string]bool, availableEmojiNames []string) *ReactTool {
	allowed := make(map[string]bool, len(allowedChannelIDs))
	for raw := range allowedChannelIDs {
		channelID := strings.TrimSpace(raw)
		if channelID == "" {
			continue
		}
		allowed[channelID] = true
	}
	emojiSet := make(map[string]bool, len(availableEmojiNames))
	emojis := make([]string, 0, len(availableEmojiNames))
	for _, raw := range availableEmojiNames {
		name, err := normalizeSlackEmojiName(raw)
		if err != nil || name == "" || emojiSet[name] {
			continue
		}
		emojiSet[name] = true
		emojis = append(emojis, name)
	}
	sort.Strings(emojis)
	return &ReactTool{
		api:               api,
		defaultChannelID:  strings.TrimSpace(defaultChannelID),
		defaultMessageTS:  strings.TrimSpace(defaultMessageTS),
		allowedChannelIDs: allowed,
		availableEmojis:   emojis,
		availableEmojiSet: emojiSet,
	}
}

func (t *ReactTool) Name() string { return "slack_react" }

func (t *ReactTool) Description() string {
	return "Adds an emoji reaction to a Slack message. Prefer this for lightweight acknowledgements. " + t.emojiNameGuidance()
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
				"description": t.emojiNameGuidance(),
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
	if len(t.availableEmojiSet) > 0 && !t.availableEmojiSet[emoji] {
		return "", fmt.Errorf("emoji name is not available in this Slack workspace: %s", rawEmoji)
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

func (t *ReactTool) emojiNameGuidance() string {
	base := "Use Slack-style emoji name (not Unicode). Example: thumbsup or :thumbsup:."
	if t == nil || len(t.availableEmojis) == 0 {
		return base
	}
	limit := slackEmojiPreviewLimit
	if limit <= 0 || limit > len(t.availableEmojis) {
		limit = len(t.availableEmojis)
	}
	preview := strings.Join(t.availableEmojis[:limit], ", ")
	if len(t.availableEmojis) > limit {
		preview += ", ..."
	}
	return base + " Available emoji names (" + strconv.Itoa(len(t.availableEmojis)) + "): " + preview + "."
}
