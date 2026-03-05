package line

import (
	"context"
	"fmt"
	"testing"
)

type stubReactAPI struct {
	chatID    string
	messageID string
	emoji     string
	err       error
}

func (s *stubReactAPI) AddReaction(ctx context.Context, chatID, messageID, emoji string) error {
	_ = ctx
	s.chatID = chatID
	s.messageID = messageID
	s.emoji = emoji
	return s.err
}

func TestLineReactToolExecute_DefaultTarget(t *testing.T) {
	api := &stubReactAPI{}
	tool := NewReactTool(api, "C123", "m_1001", nil)
	out, err := tool.Execute(context.Background(), map[string]any{
		"emoji": "👍",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out != "reacted with 👍" {
		t.Fatalf("output = %q, want %q", out, "reacted with 👍")
	}
	if api.chatID != "C123" || api.messageID != "m_1001" || api.emoji != "👍" {
		t.Fatalf("api payload mismatch: chat=%q message=%q emoji=%q", api.chatID, api.messageID, api.emoji)
	}
	last := tool.LastReaction()
	if last == nil {
		t.Fatalf("LastReaction() = nil")
	}
	if last.ChatID != "C123" || last.MessageID != "m_1001" || last.Emoji != "👍" {
		t.Fatalf("last reaction mismatch: %+v", *last)
	}
}

func TestLineReactToolExecute_OverrideTarget(t *testing.T) {
	api := &stubReactAPI{}
	tool := NewReactTool(api, "C123", "m_1001", map[string]bool{
		"C123": true,
		"C456": true,
	})
	out, err := tool.Execute(context.Background(), map[string]any{
		"chat_id":    "C456",
		"message_id": "m_2001",
		"emoji":      "🎉",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out != "reacted with 🎉" {
		t.Fatalf("output = %q, want %q", out, "reacted with 🎉")
	}
	if api.chatID != "C456" || api.messageID != "m_2001" || api.emoji != "🎉" {
		t.Fatalf("api payload mismatch: chat=%q message=%q emoji=%q", api.chatID, api.messageID, api.emoji)
	}
}

func TestLineReactToolExecute_ValidationAndAPIError(t *testing.T) {
	t.Run("missing emoji", func(t *testing.T) {
		api := &stubReactAPI{}
		tool := NewReactTool(api, "C123", "m_1001", nil)
		if _, err := tool.Execute(context.Background(), map[string]any{}); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("invalid emoji", func(t *testing.T) {
		api := &stubReactAPI{}
		tool := NewReactTool(api, "C123", "m_1001", nil)
		if _, err := tool.Execute(context.Background(), map[string]any{
			"emoji": "not emoji",
		}); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("missing runtime chat context", func(t *testing.T) {
		api := &stubReactAPI{}
		tool := NewReactTool(api, "", "m_1001", nil)
		if _, err := tool.Execute(context.Background(), map[string]any{
			"emoji": "👍",
		}); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("unauthorized chat", func(t *testing.T) {
		api := &stubReactAPI{}
		tool := NewReactTool(api, "C123", "m_1001", map[string]bool{
			"C123": true,
		})
		if _, err := tool.Execute(context.Background(), map[string]any{
			"chat_id": "C999",
			"emoji":   "👍",
		}); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("api error", func(t *testing.T) {
		api := &stubReactAPI{err: fmt.Errorf("reaction is not allowed")}
		tool := NewReactTool(api, "C123", "m_1001", nil)
		if _, err := tool.Execute(context.Background(), map[string]any{
			"emoji": "👍",
		}); err == nil {
			t.Fatalf("expected error")
		}
	})
}
