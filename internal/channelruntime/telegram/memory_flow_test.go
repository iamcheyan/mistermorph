package telegram

import (
	"testing"
)

func TestBuildMemoryWriteMeta(t *testing.T) {
	t.Run("normal chat keeps tg session and contact meta", func(t *testing.T) {
		meta := buildMemoryWriteMeta(telegramJob{
			ChatID:          777,
			FromUserID:      1001,
			FromUsername:    "@alice",
			FromDisplayName: "Alice",
		})
		if meta.SessionID != "tg:777" {
			t.Fatalf("session_id = %q, want %q", meta.SessionID, "tg:777")
		}
		if len(meta.ContactIDs) != 1 || meta.ContactIDs[0] != "tg:@alice" {
			t.Fatalf("contact_ids = %#v, want [\"tg:@alice\"]", meta.ContactIDs)
		}
		if len(meta.ContactNicknames) != 1 || meta.ContactNicknames[0] != "Alice" {
			t.Fatalf("contact_nicknames = %#v, want [\"Alice\"]", meta.ContactNicknames)
		}
	})
}
