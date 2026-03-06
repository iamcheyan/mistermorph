package lark

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
	var gotReplyTo string
	calls := 0

	adapter, err := NewDeliveryAdapter(DeliveryAdapterOptions{
		SendText: func(ctx context.Context, target any, text string, opts SendTextOptions) error {
			typed, ok := target.(DeliveryTarget)
			if !ok {
				t.Fatalf("target type mismatch: got %T want DeliveryTarget", target)
			}
			gotTarget = typed
			gotText = text
			gotReplyTo = opts.ReplyToMessageID
			calls++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewDeliveryAdapter() error = %v", err)
	}

	payloadBase64, err := busruntime.EncodeMessageEnvelope(busruntime.TopicChatMessage, busruntime.MessageEnvelope{
		MessageID: "msg_9001",
		Text:      "hello lark",
		SentAt:    "2026-03-06T00:00:00Z",
		SessionID: "0194e9d5-2f8f-7000-8000-000000000001",
		ReplyTo:   "om_1001",
	})
	if err != nil {
		t.Fatalf("EncodeMessageEnvelope() error = %v", err)
	}
	msg := busruntime.BusMessage{
		Direction:       busruntime.DirectionOutbound,
		Channel:         busruntime.ChannelLark,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: "lark:oc_group123",
		IdempotencyKey:  "msg:msg_9001",
		CorrelationID:   "corr_9001",
		PayloadBase64:   payloadBase64,
		CreatedAt:       time.Now().UTC(),
		Extensions: busruntime.MessageExtensions{
			ReplyTo: "om_1001",
		},
	}

	accepted, deduped, err := adapter.Deliver(context.Background(), msg)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if !accepted || deduped {
		t.Fatalf("accepted=%v deduped=%v, want true false", accepted, deduped)
	}
	if calls != 1 {
		t.Fatalf("send calls mismatch: got %d want 1", calls)
	}
	if gotTarget.ChatID != "oc_group123" {
		t.Fatalf("chat_id mismatch: got %q want %q", gotTarget.ChatID, "oc_group123")
	}
	if gotText != "hello lark" {
		t.Fatalf("text mismatch: got %q want %q", gotText, "hello lark")
	}
	if gotReplyTo != "om_1001" {
		t.Fatalf("reply_to mismatch: got %q want %q", gotReplyTo, "om_1001")
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
		SentAt:    "2026-03-06T00:00:00Z",
		SessionID: "0194e9d5-2f8f-7000-8000-000000000002",
	})
	if err != nil {
		t.Fatalf("EncodeMessageEnvelope() error = %v", err)
	}
	msg := busruntime.BusMessage{
		Direction:       busruntime.DirectionOutbound,
		Channel:         busruntime.ChannelLark,
		Topic:           busruntime.TopicChatMessage,
		ConversationKey: "lark:",
		IdempotencyKey:  "msg:msg_9002",
		CorrelationID:   "corr_9002",
		PayloadBase64:   payloadBase64,
		CreatedAt:       time.Now().UTC(),
	}
	accepted, deduped, err := adapter.Deliver(context.Background(), msg)
	if err == nil {
		t.Fatalf("Deliver() expected error for invalid conversation key")
	}
	if !strings.Contains(err.Error(), "lark chat id is required") {
		t.Fatalf("Deliver() error mismatch: got %q", err.Error())
	}
	if accepted || deduped {
		t.Fatalf("accepted=%v deduped=%v, want false false", accepted, deduped)
	}
	if calls != 0 {
		t.Fatalf("send calls mismatch: got %d want 0", calls)
	}
}
