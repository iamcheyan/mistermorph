package telegramcmd

import (
	heartbeatruntime "github.com/quailyquaily/mistermorph/internal/channelruntime/heartbeat"
	telegramruntime "github.com/quailyquaily/mistermorph/internal/channelruntime/telegram"
)

// Dependencies defines runtime wiring hooks for telegram + heartbeat mode.
type Dependencies struct {
	heartbeatruntime.Dependencies
	HandleModelCommand telegramruntime.HandleModelCommandFunc
}
