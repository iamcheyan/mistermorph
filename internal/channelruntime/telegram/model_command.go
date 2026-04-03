package telegram

import (
	"context"
	htmlstd "html"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/channelruntime/taskruntime"
	"github.com/quailyquaily/mistermorph/llm"
)

func executeTelegramProfileCommand(d Dependencies, api *telegramAPI, chatID int64, text string) bool {
	if d.HandleModelCommand == nil {
		return false
	}
	output, handled, err := d.HandleModelCommand(text)
	if !handled {
		return false
	}
	if err != nil {
		output = "error: " + strings.TrimSpace(err.Error())
	}
	_ = api.sendMessageHTML(context.Background(), chatID, htmlstd.EscapeString(output), true)
	return true
}

func resolveTelegramMainForUse(rt *taskruntime.Runtime) (llm.Client, string, func(), error) {
	route, err := rt.ResolveMainRouteForRun()
	if err != nil {
		return nil, "", func() {}, err
	}
	client, err := rt.CreateClientForRoute(route)
	if err != nil {
		return nil, "", func() {}, err
	}
	model := strings.TrimSpace(route.ClientConfig.Model)
	cleanup := func() {
		if closer, ok := client.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
	return client, model, cleanup, nil
}
