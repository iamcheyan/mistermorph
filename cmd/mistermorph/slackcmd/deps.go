package slackcmd

import (
	heartbeatruntime "github.com/quailyquaily/mistermorph/internal/channelruntime/heartbeat"
	slackruntime "github.com/quailyquaily/mistermorph/internal/channelruntime/slack"
)

// Dependencies defines runtime wiring hooks for slack + heartbeat mode.
type Dependencies struct {
	heartbeatruntime.Dependencies
	HandleModelCommand slackruntime.HandleModelCommandFunc
}
