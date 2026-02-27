package telegram

import (
	"testing"
	"time"

	"github.com/google/uuid"
	busruntime "github.com/quailyquaily/mistermorph/internal/bus"
)

func TestTelegramOutboundEventFromBusMessage(t *testing.T) {
	sessionID, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("uuid.NewV7() error = %v", err)
	}
	payload, err := busruntime.EncodeMessageEnvelope(busruntime.TopicChatMessage, busruntime.MessageEnvelope{
		MessageID: "msg_1",
		Text:      "hello",
		SentAt:    time.Now().UTC().Format(time.RFC3339),
		SessionID: sessionID.String(),
		ReplyTo:   "123",
	})
	if err != nil {
		t.Fatalf("EncodeMessageEnvelope() error = %v", err)
	}
	event, err := telegramOutboundEventFromBusMessage(busruntime.BusMessage{
		ConversationKey: "tg:42",
		Topic:           busruntime.TopicChatMessage,
		PayloadBase64:   payload,
		CorrelationID:   "telegram:plan:42:1",
		Extensions: busruntime.MessageExtensions{
			ReplyTo: "123",
		},
	})
	if err != nil {
		t.Fatalf("telegramOutboundEventFromBusMessage() error = %v", err)
	}
	if event.ChatID != 42 {
		t.Fatalf("chat id = %d, want 42", event.ChatID)
	}
	if event.ReplyToMessageID != 123 {
		t.Fatalf("reply to message id = %d, want 123", event.ReplyToMessageID)
	}
	if event.Text != "hello" {
		t.Fatalf("text = %q, want hello", event.Text)
	}
	if event.Kind != "plan_progress" {
		t.Fatalf("kind = %q, want plan_progress", event.Kind)
	}
}

func TestTelegramOutboundKind(t *testing.T) {
	if got := telegramOutboundKind("telegram:error:1:2"); got != "error" {
		t.Fatalf("kind(error) = %q, want error", got)
	}
	if got := telegramOutboundKind("telegram:file_download_error:1:2"); got != "error" {
		t.Fatalf("kind(file_download_error) = %q, want error", got)
	}
	if got := telegramOutboundKind("telegram:message:1:2"); got != "message" {
		t.Fatalf("kind(message) = %q, want message", got)
	}
}

func TestNextTelegramPlanProgressStateAppendsLines(t *testing.T) {
	state, rendered := nextTelegramPlanProgressState(telegramPlanProgressEditState{}, "telegram:plan:1:1", "step 1")
	if state.CorrelationID != "telegram:plan:1:1" {
		t.Fatalf("correlation_id = %q", state.CorrelationID)
	}
	if len(state.Lines) != 1 || state.Lines[0] != "step 1" {
		t.Fatalf("lines = %#v, want single step 1", state.Lines)
	}
	if rendered != "<blockquote expandable>🤔 1. step 1</blockquote>" {
		t.Fatalf("rendered = %q", rendered)
	}

	state.MessageID = 123
	next, rendered := nextTelegramPlanProgressState(state, "telegram:plan:1:1", "step 2")
	if next.MessageID != 123 {
		t.Fatalf("message_id = %d, want 123", next.MessageID)
	}
	if len(next.Lines) != 2 || next.Lines[0] != "step 1" || next.Lines[1] != "step 2" {
		t.Fatalf("lines = %#v, want step1+step2", next.Lines)
	}
	if rendered != "<blockquote expandable>🤔 2. step 2<br>🤔 1. step 1</blockquote>" {
		t.Fatalf("rendered = %q, want numbered reverse lines", rendered)
	}
}

func TestNextTelegramPlanProgressStateResetsOnCorrelationChange(t *testing.T) {
	prev := telegramPlanProgressEditState{
		CorrelationID: "telegram:plan:1:1",
		MessageID:     99,
		Lines:         []string{"step 1"},
	}
	next, rendered := nextTelegramPlanProgressState(prev, "telegram:plan:1:2", "step 2")
	if next.MessageID != 0 {
		t.Fatalf("message_id = %d, want 0", next.MessageID)
	}
	if len(next.Lines) != 1 || next.Lines[0] != "step 2" {
		t.Fatalf("lines = %#v, want only step 2", next.Lines)
	}
	if rendered != "<blockquote expandable>🤔 1. step 2</blockquote>" {
		t.Fatalf("rendered = %q", rendered)
	}
}

func TestRenderTelegramPlanProgressExpandableEscapesHTML(t *testing.T) {
	got := renderTelegramPlanProgressExpandable([]string{"<b>x</b>"})
	if got != "<blockquote expandable>🤔 1. &lt;b&gt;x&lt;/b&gt;</blockquote>" {
		t.Fatalf("rendered = %q", got)
	}
}
