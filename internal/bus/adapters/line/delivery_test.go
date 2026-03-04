package line

import (
	"context"
	"strings"
	"testing"
	"time"

	busruntime "github.com/quailyquaily/mistermorph/internal/bus"
)

func TestDeliveryAdapterDeliver(t *testing.T) {
	t.Parallel()

	var gotTarget DeliveryTarget
	var gotText string
	var gotReplyToken string
	calls := 0

	adapter, err := NewDeliveryAdapter(DeliveryAdapterOptions{
		SendText: func(ctx context.Context, target any, text string, opts SendTextOptions) error {
			typed, ok := target.(DeliveryTarget)
			if !ok {
				t.Fatalf("target type mismatch: got %T want DeliveryTarget", target)
			}
			gotTarget = typed
			gotText = text
			gotReplyToken = opts.ReplyToken
			calls++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewDeliveryAdapter() error = %v", err)
	}

	payloadBase64, err := busruntime.EncodeMessageEnvelope(busruntime.TopicChatMessage, busruntime.MessageEnvelope{
		MessageID: "msg_9001",
		Text:      "hello line",
		SentAt:    "2026-03-04T00:00:00Z",
		SessionID: "0194e9d5-2f8f-7000-8000-000000000001",
		ReplyTo:   "rtok_abc",
	})
	if err != nil {
		t.Fatalf("EncodeMessageEnvelope() error = %v", err)
	}
	conversationKey, err := busruntime.BuildLineGroupConversationKey("Cgroup123")
	if err != nil {
		t.Fatalf("BuildLineGroupConversationKey() error = %v", err)
	}
	msg := busruntime.BusMessage{
		Direction:       busruntime.DirectionOutbound,
		Channel:         busruntime.ChannelLine,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: conversationKey,
		IdempotencyKey:  "msg:msg_9001",
		CorrelationID:   "corr_9001",
		PayloadBase64:   payloadBase64,
		CreatedAt:       time.Now().UTC(),
		Extensions: busruntime.MessageExtensions{
			ReplyTo: "rtok_abc",
		},
	}
	accepted, deduped, err := adapter.Deliver(context.Background(), msg)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if !accepted {
		t.Fatalf("accepted mismatch: got %v want true", accepted)
	}
	if deduped {
		t.Fatalf("deduped mismatch: got %v want false", deduped)
	}
	if calls != 1 {
		t.Fatalf("send calls mismatch: got %d want 1", calls)
	}
	if gotTarget.GroupID != "Cgroup123" {
		t.Fatalf("group_id mismatch: got %q want %q", gotTarget.GroupID, "Cgroup123")
	}
	if gotText != "hello line" {
		t.Fatalf("text mismatch: got %q want %q", gotText, "hello line")
	}
	if gotReplyToken != "rtok_abc" {
		t.Fatalf("reply_token mismatch: got %q want %q", gotReplyToken, "rtok_abc")
	}
}

func TestDeliveryAdapterRejectsInvalidConversationKey(t *testing.T) {
	t.Parallel()

	calls := 0
	adapter, err := NewDeliveryAdapter(DeliveryAdapterOptions{
		SendText: func(ctx context.Context, target any, text string, opts SendTextOptions) error {
			calls++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewDeliveryAdapter() error = %v", err)
	}

	payloadBase64, err := busruntime.EncodeMessageEnvelope(busruntime.TopicChatMessage, busruntime.MessageEnvelope{
		MessageID: "msg_9002",
		Text:      "bad target",
		SentAt:    "2026-03-04T00:00:00Z",
		SessionID: "0194e9d5-2f8f-7000-8000-000000000002",
	})
	if err != nil {
		t.Fatalf("EncodeMessageEnvelope() error = %v", err)
	}
	msg := busruntime.BusMessage{
		Direction:       busruntime.DirectionOutbound,
		Channel:         busruntime.ChannelLine,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: "line:",
		IdempotencyKey:  "msg:msg_9002",
		CorrelationID:   "corr_9002",
		PayloadBase64:   payloadBase64,
		CreatedAt:       time.Now().UTC(),
	}
	accepted, deduped, err := adapter.Deliver(context.Background(), msg)
	if err == nil {
		t.Fatalf("Deliver() expected error for invalid conversation key")
	}
	if !strings.Contains(err.Error(), "line group id is required") {
		t.Fatalf("Deliver() error mismatch: got %q", err.Error())
	}
	if accepted {
		t.Fatalf("accepted mismatch: got %v want false", accepted)
	}
	if deduped {
		t.Fatalf("deduped mismatch: got %v want false", deduped)
	}
	if calls != 0 {
		t.Fatalf("send calls mismatch: got %d want 0", calls)
	}
}
