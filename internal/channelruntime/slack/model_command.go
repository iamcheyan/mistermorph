package slack

import (
	"context"
	"fmt"
	"strings"

	busruntime "github.com/quailyquaily/mistermorph/internal/bus"
)

func maybeHandleSlackProfileCommand(ctx context.Context, d Dependencies, inprocBus *busruntime.Inproc, event slackInboundEvent, botUserID string) (bool, error) {
	if isSlackGroupChat(event.ChatType) && !slackCommandExplicitlyAddressed(event.Text, botUserID) {
		return false, nil
	}
	if d.HandleModelCommand == nil {
		return false, nil
	}
	output, handled, err := d.HandleModelCommand(normalizeSlackCommandText(event.Text, botUserID))
	if !handled {
		return false, nil
	}
	if err != nil {
		output = "error: " + strings.TrimSpace(err.Error())
	}
	if ctx == nil {
		ctx = context.Background()
	}
	_, publishErr := publishSlackBusOutbound(
		ctx,
		inprocBus,
		event.TeamID,
		event.ChannelID,
		output,
		event.ThreadTS,
		fmt.Sprintf("slack:model:%s:%s", event.ChannelID, event.MessageTS),
	)
	return true, publishErr
}

func normalizeSlackCommandText(text string, botUserID string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return ""
	}
	mention := "<@" + strings.TrimSpace(botUserID) + ">"
	if strings.TrimSpace(botUserID) != "" && fields[0] == mention {
		fields = fields[1:]
	}
	return strings.Join(fields, " ")
}

func slackCommandExplicitlyAddressed(text string, botUserID string) bool {
	if strings.TrimSpace(botUserID) == "" {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(text), "<@"+strings.TrimSpace(botUserID)+">")
}
