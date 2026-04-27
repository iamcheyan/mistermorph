package slack

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const (
	slackWorkingMessageText          = "working..."
	slackDoneMessageText             = "done."
	slackWorkingMessageDelay         = 1200 * time.Millisecond
	slackWorkingMessagePostTimeout   = 5 * time.Second
	slackWorkingMessageUpdateTimeout = 5 * time.Second
)

type slackWorkingMessage struct {
	api       *slackAPI
	logger    *slog.Logger
	channelID string
	threadTS  string
	messageTS string

	stopOnce sync.Once
	stop     chan struct{}
	result   chan slackWorkingMessagePostResult
}

type slackWorkingMessagePostResult struct {
	ref slackMessageRef
	err error
}

func startSlackWorkingMessage(ctx context.Context, logger *slog.Logger, api *slackAPI, job slackJob) *slackWorkingMessage {
	return startSlackWorkingMessageWithDelay(ctx, logger, api, job, slackWorkingMessageDelay)
}

func startSlackWorkingMessageWithDelay(ctx context.Context, logger *slog.Logger, api *slackAPI, job slackJob, delay time.Duration) *slackWorkingMessage {
	if ctx == nil {
		ctx = context.Background()
	}
	channelID := strings.TrimSpace(job.ChannelID)
	if api == nil || channelID == "" {
		return nil
	}
	w := &slackWorkingMessage{
		api:       api,
		logger:    logger,
		channelID: channelID,
		threadTS:  strings.TrimSpace(job.ThreadTS),
		messageTS: strings.TrimSpace(job.MessageTS),
		stop:      make(chan struct{}),
		result:    make(chan slackWorkingMessagePostResult, 1),
	}
	go w.run(ctx, delay)
	return w
}

func (w *slackWorkingMessage) run(ctx context.Context, delay time.Duration) {
	defer close(w.result)
	if delay > 0 {
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-w.stop:
			timer.Stop()
			return
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}

	postCtx, cancel := context.WithTimeout(ctx, slackWorkingMessagePostTimeout)
	defer cancel()
	ref, err := w.api.postMessageWithResult(postCtx, w.channelID, slackWorkingMessageText, w.threadTS)
	w.result <- slackWorkingMessagePostResult{ref: ref, err: err}
	if err != nil && w.logger != nil {
		w.logger.Warn("slack_working_message_post_error",
			"channel_id", w.channelID,
			"message_ts", w.messageTS,
			"error", err.Error(),
		)
	}
}

func (w *slackWorkingMessage) Update(ctx context.Context, text string) (bool, error) {
	if w == nil {
		return false, nil
	}
	w.stopOnce.Do(func() {
		close(w.stop)
	})

	text = strings.TrimSpace(text)
	if text == "" {
		text = slackDoneMessageText
	}

	result, ok := <-w.result
	if !ok || result.err != nil || strings.TrimSpace(result.ref.MessageTS) == "" {
		return false, nil
	}
	ref := result.ref
	if strings.TrimSpace(ref.ChannelID) == "" {
		ref.ChannelID = w.channelID
	}
	if ctx == nil {
		ctx = context.Background()
	}
	updateCtx, cancel := context.WithTimeout(ctx, slackWorkingMessageUpdateTimeout)
	defer cancel()
	return true, w.api.updateMessage(updateCtx, ref.ChannelID, ref.MessageTS, text)
}

func callSlackDirectOutboundHook(ctx context.Context, logger *slog.Logger, hooks Hooks, job slackJob, text, correlationID string) {
	if hooks.OnOutbound == nil {
		return
	}
	conversationKey := strings.TrimSpace(job.ConversationKey)
	if conversationKey == "" {
		if key, err := buildSlackConversationKey(job.TeamID, job.ChannelID); err == nil {
			conversationKey = key
		}
	}
	callOutboundHook(ctx, logger, hooks, OutboundEvent{
		ConversationKey: conversationKey,
		TeamID:          strings.TrimSpace(job.TeamID),
		ChannelID:       strings.TrimSpace(job.ChannelID),
		ThreadTS:        strings.TrimSpace(job.ThreadTS),
		Text:            strings.TrimSpace(text),
		CorrelationID:   strings.TrimSpace(correlationID),
		Kind:            slackOutboundKind(correlationID),
	})
}
