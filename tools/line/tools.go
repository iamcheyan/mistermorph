package line

import "context"

// API is the minimal LINE transport surface needed by LINE tools.
type API interface {
	AddReaction(ctx context.Context, chatID, messageID, emoji string) error
}

type Reaction struct {
	ChatID    string
	MessageID string
	Emoji     string
	Source    string
}
