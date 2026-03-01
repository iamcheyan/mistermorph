package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestSlackAPIUserIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users.info" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/users.info")
		}
		if got := strings.TrimSpace(r.Header.Get("Authorization")); got != "Bearer xoxb-test" {
			t.Fatalf("authorization = %q", got)
		}
		if got := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))); !strings.Contains(got, "application/x-www-form-urlencoded") {
			t.Fatalf("content-type = %q", got)
		}
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read payload: %v", err)
		}
		payload, err := url.ParseQuery(string(rawBody))
		if err != nil {
			t.Fatalf("parse payload: %v", err)
		}
		if got := strings.TrimSpace(payload.Get("user")); got != "U123" {
			t.Fatalf("user = %q, want %q", got, "U123")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"user": map[string]any{
				"id":   "U123",
				"name": "alice",
				"profile": map[string]any{
					"display_name": "Alice",
					"real_name":    "Alice Real",
				},
			},
		})
	}))
	defer server.Close()

	api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
	identity, err := api.userIdentity(context.Background(), "U123")
	if err != nil {
		t.Fatalf("userIdentity() error = %v", err)
	}
	if identity.UserID != "U123" {
		t.Fatalf("user id = %q, want %q", identity.UserID, "U123")
	}
	if identity.Username != "alice" {
		t.Fatalf("username = %q, want %q", identity.Username, "alice")
	}
	if identity.DisplayName != "Alice" {
		t.Fatalf("display name = %q, want %q", identity.DisplayName, "Alice")
	}
}

func TestSlackAPIUserIdentityFallbackAndError(t *testing.T) {
	t.Run("fallback to username for display name", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"user": map[string]any{
					"id":   "U222",
					"name": "bob",
					"profile": map[string]any{
						"display_name": "",
						"real_name":    "",
					},
				},
			})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		identity, err := api.userIdentity(context.Background(), "U222")
		if err != nil {
			t.Fatalf("userIdentity() error = %v", err)
		}
		if identity.DisplayName != "bob" {
			t.Fatalf("display name = %q, want %q", identity.DisplayName, "bob")
		}
	})

	t.Run("fallback to user id when username is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"user": map[string]any{
					"id":   "",
					"name": "",
					"profile": map[string]any{
						"display_name": "",
						"real_name":    "",
					},
				},
			})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		identity, err := api.userIdentity(context.Background(), "U333")
		if err != nil {
			t.Fatalf("userIdentity() error = %v", err)
		}
		if identity.UserID != "U333" {
			t.Fatalf("user id = %q, want %q", identity.UserID, "U333")
		}
		if identity.Username != "U333" {
			t.Fatalf("username = %q, want %q", identity.Username, "U333")
		}
		if identity.DisplayName != "U333" {
			t.Fatalf("display name = %q, want %q", identity.DisplayName, "U333")
		}
	})

	t.Run("fallback to user id on user_not_found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":    false,
				"error": "user_not_found",
			})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		identity, err := api.userIdentity(context.Background(), "U404")
		if err != nil {
			t.Fatalf("userIdentity() error = %v", err)
		}
		if identity.UserID != "U404" {
			t.Fatalf("user id = %q, want %q", identity.UserID, "U404")
		}
		if identity.Username != "U404" {
			t.Fatalf("username = %q, want %q", identity.Username, "U404")
		}
		if identity.DisplayName != "U404" {
			t.Fatalf("display name = %q, want %q", identity.DisplayName, "U404")
		}
	})

	t.Run("slack api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":    false,
				"error": "invalid_auth",
			})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		_, err := api.userIdentity(context.Background(), "U404")
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "invalid_auth") {
			t.Fatalf("error = %v, want invalid_auth", err)
		}
	})
}

func TestSlackAPIAddReaction(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/reactions.add" {
				t.Fatalf("path = %q, want %q", r.URL.Path, "/reactions.add")
			}
			if got := strings.TrimSpace(r.Header.Get("Authorization")); got != "Bearer xoxb-test" {
				t.Fatalf("authorization = %q", got)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if got := strings.TrimSpace(payload["channel"].(string)); got != "C123" {
				t.Fatalf("channel = %q, want %q", got, "C123")
			}
			if got := strings.TrimSpace(payload["timestamp"].(string)); got != "1739667600.000100" {
				t.Fatalf("timestamp = %q, want %q", got, "1739667600.000100")
			}
			if got := strings.TrimSpace(payload["name"].(string)); got != "thumbsup" {
				t.Fatalf("name = %q, want %q", got, "thumbsup")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		if err := api.addReaction(context.Background(), "C123", "1739667600.000100", "thumbsup"); err != nil {
			t.Fatalf("addReaction() error = %v", err)
		}
	})

	t.Run("already_reacted treated as success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":    false,
				"error": "already_reacted",
			})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		if err := api.addReaction(context.Background(), "C123", "1739667600.000100", "thumbsup"); err != nil {
			t.Fatalf("addReaction() error = %v", err)
		}
	})

	t.Run("slack error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":    false,
				"error": "invalid_name",
			})
		}))
		defer server.Close()

		api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
		err := api.addReaction(context.Background(), "C123", "1739667600.000100", "not-valid")
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "invalid_name") {
			t.Fatalf("error = %v, want invalid_name", err)
		}
	})
}

func TestSlackAPIListEmojiNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emoji.list" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/emoji.list")
		}
		if got := strings.TrimSpace(r.Header.Get("Authorization")); got != "Bearer xoxb-test" {
			t.Fatalf("authorization = %q", got)
		}
		if got := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))); !strings.Contains(got, "application/x-www-form-urlencoded") {
			t.Fatalf("content-type = %q", got)
		}
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read payload: %v", err)
		}
		payload, err := url.ParseQuery(string(rawBody))
		if err != nil {
			t.Fatalf("parse payload: %v", err)
		}
		if got := strings.TrimSpace(payload.Get("include_categories")); got != "true" {
			t.Fatalf("include_categories = %q, want %q", got, "true")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"emoji": map[string]any{
				"party_parrot": "https://example.com/parrot.png",
				"shipit":       "alias:party_parrot",
			},
			"categories": []map[string]any{
				{
					"name":        "Smileys & Emotion",
					"emoji_names": []string{"thumbsup", "older_woman"},
				},
			},
		})
	}))
	defer server.Close()

	api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
	names, err := api.listEmojiNames(context.Background())
	if err != nil {
		t.Fatalf("listEmojiNames() error = %v", err)
	}
	want := []string{"older_woman", "party_parrot", "shipit", "thumbsup"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %#v, want %#v", names, want)
	}
}
