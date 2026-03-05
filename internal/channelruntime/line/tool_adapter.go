package line

import (
	"context"
	"fmt"

	linetools "github.com/quailyquaily/mistermorph/tools/line"
)

type lineToolAPI struct {
	api *lineAPI
}

func newLineToolAPI(api *lineAPI) linetools.API {
	if api == nil {
		return nil
	}
	return &lineToolAPI{api: api}
}

func (a *lineToolAPI) AddReaction(ctx context.Context, chatID, messageID, emoji string) error {
	if a == nil || a.api == nil {
		return fmt.Errorf("line api not available")
	}
	return a.api.addReaction(ctx, chatID, messageID, emoji)
}
