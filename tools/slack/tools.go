package slack

import "context"

// API is the minimal Slack transport surface needed by Slack tools.
type API interface {
	AddReaction(ctx context.Context, channelID, messageTS, emoji string) error
}

type Reaction struct {
	ChannelID string
	MessageTS string
	Emoji     string
	Source    string
}
