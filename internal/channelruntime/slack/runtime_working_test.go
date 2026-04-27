package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSlackWorkingMessageSkipsPostBeforeDelay(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"channel": "C123",
			"ts":      "1739667601.000200",
		})
	}))
	defer server.Close()

	api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
	working := startSlackWorkingMessageWithDelay(context.Background(), nil, api, slackJob{
		ChannelID: "C123",
		ThreadTS:  "1739667600.000100",
		MessageTS: "1739667600.000100",
	}, time.Hour)
	updated, err := working.Update(context.Background(), "done")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated {
		t.Fatalf("updated = true, want false")
	}
	if callCount != 0 {
		t.Fatalf("call count = %d, want 0", callCount)
	}
}

func TestSlackWorkingMessageUpdatesPostedMessage(t *testing.T) {
	var (
		mu       sync.Mutex
		paths    []string
		finalMsg string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		switch r.URL.Path {
		case "/chat.postMessage":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode post payload: %v", err)
			}
			if got := strings.TrimSpace(payload["text"].(string)); got != slackWorkingMessageText {
				t.Fatalf("post text = %q, want %q", got, slackWorkingMessageText)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":      true,
				"channel": "C123",
				"ts":      "1739667601.000200",
			})
		case "/chat.update":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode update payload: %v", err)
			}
			finalMsg = strings.TrimSpace(payload["text"].(string))
			if got := strings.TrimSpace(payload["channel"].(string)); got != "C123" {
				t.Fatalf("update channel = %q, want C123", got)
			}
			if got := strings.TrimSpace(payload["ts"].(string)); got != "1739667601.000200" {
				t.Fatalf("update ts = %q, want 1739667601.000200", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	api := newSlackAPI(server.Client(), server.URL, "xoxb-test", "xapp-test")
	working := startSlackWorkingMessageWithDelay(context.Background(), nil, api, slackJob{
		ChannelID: "C123",
		ThreadTS:  "1739667600.000100",
		MessageTS: "1739667600.000100",
	}, 0)
	updated, err := working.Update(context.Background(), "final result")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !updated {
		t.Fatalf("updated = false, want true")
	}
	if finalMsg != "final result" {
		t.Fatalf("final message = %q, want final result", finalMsg)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 || paths[0] != "/chat.postMessage" || paths[1] != "/chat.update" {
		t.Fatalf("paths = %#v, want post then update", paths)
	}
}
