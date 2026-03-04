package bus

import (
	"fmt"
	"strings"
)

func BuildConversationKey(channel Channel, id string) (string, error) {
	if !isValidChannel(channel) {
		return "", fmt.Errorf("channel is invalid")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("conversation id is required")
	}
	if strings.Contains(id, " ") {
		return "", fmt.Errorf("conversation id must not contain spaces")
	}
	return fmt.Sprintf("%s:%s", conversationKeyPrefix(channel), id), nil
}

func BuildTelegramChatConversationKey(chatID string) (string, error) {
	return BuildConversationKey(ChannelTelegram, chatID)
}

func BuildSlackChannelConversationKey(channelID string) (string, error) {
	return BuildConversationKey(ChannelSlack, channelID)
}

func BuildLineConversationKey(chatID string) (string, error) {
	return BuildConversationKey(ChannelLine, chatID)
}

func BuildLineGroupConversationKey(groupID string) (string, error) {
	return BuildLineConversationKey(groupID)
}

func isValidChannel(channel Channel) bool {
	switch channel {
	case ChannelTelegram, ChannelSlack, ChannelLine, ChannelDiscord:
		return true
	default:
		return false
	}
}

func conversationKeyPrefix(channel Channel) string {
	switch channel {
	case ChannelTelegram:
		return "tg"
	case ChannelSlack:
		return "slack"
	case ChannelLine:
		return "line"
	case ChannelDiscord:
		return "discord"
	default:
		return ""
	}
}
