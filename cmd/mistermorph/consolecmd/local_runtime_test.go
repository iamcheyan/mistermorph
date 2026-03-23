package consolecmd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	heartbeatloop "github.com/quailyquaily/mistermorph/internal/channelruntime/heartbeat"
	"github.com/quailyquaily/mistermorph/internal/daemonruntime"
	"github.com/quailyquaily/mistermorph/internal/heartbeatutil"
)

func TestConsoleLocalRoutesOptionsPoke(t *testing.T) {
	rt := &consoleLocalRuntime{}
	if got := rt.routesOptions("token").Poke; got != nil {
		t.Fatalf("Poke = %#v, want nil when heartbeat loop is unavailable", got)
	}

	rt.heartbeatPokeRequests = make(chan heartbeatloop.PokeRequest)
	if got := rt.routesOptions("token").Poke; got == nil {
		t.Fatal("Poke = nil, want non-nil when heartbeat loop is available")
	}
}

func TestConsoleLocalRoutesOptionsOverviewHeartbeatRunning(t *testing.T) {
	rt := &consoleLocalRuntime{
		heartbeatState:        &heartbeatutil.State{},
		heartbeatPokeRequests: make(chan heartbeatloop.PokeRequest),
	}
	if ok := rt.heartbeatState.Start(); !ok {
		t.Fatal("Start() = false, want true")
	}

	payload, err := rt.routesOptions("token").Overview(context.Background())
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if got, _ := payload["heartbeat_running"].(bool); !got {
		t.Fatalf("heartbeat_running = %v, want true", payload["heartbeat_running"])
	}
}

func TestConsoleLocalRuntimeCompleteHeartbeatTask(t *testing.T) {
	t.Run("success clears running and records timestamp", func(t *testing.T) {
		now := time.Date(2026, time.March, 23, 10, 0, 0, 0, time.UTC)
		rt := &consoleLocalRuntime{heartbeatState: &heartbeatutil.State{}}
		if ok := rt.heartbeatState.Start(); !ok {
			t.Fatal("Start() = false, want true")
		}

		rt.completeHeartbeatTask(consoleLocalTaskJob{
			Trigger: daemonruntime.TaskTrigger{Source: "heartbeat"},
		}, heartbeatTaskResultSuccess, nil, now)

		failures, lastSuccess, lastError, running := rt.heartbeatState.Snapshot()
		if running {
			t.Fatal("running = true, want false")
		}
		if failures != 0 {
			t.Fatalf("failures = %d, want 0", failures)
		}
		if lastError != "" {
			t.Fatalf("lastError = %q, want empty", lastError)
		}
		if !lastSuccess.Equal(now) {
			t.Fatalf("lastSuccess = %v, want %v", lastSuccess, now)
		}
	})

	t.Run("failure clears running and records error", func(t *testing.T) {
		rt := &consoleLocalRuntime{heartbeatState: &heartbeatutil.State{}}
		if ok := rt.heartbeatState.Start(); !ok {
			t.Fatal("Start() = false, want true")
		}

		rt.completeHeartbeatTask(consoleLocalTaskJob{
			Trigger: daemonruntime.TaskTrigger{Source: "heartbeat"},
		}, heartbeatTaskResultFailure, errors.New("boom"), time.Time{})

		failures, _, lastError, running := rt.heartbeatState.Snapshot()
		if running {
			t.Fatal("running = true, want false")
		}
		if failures != 1 {
			t.Fatalf("failures = %d, want 1", failures)
		}
		if lastError != "boom" {
			t.Fatalf("lastError = %q, want %q", lastError, "boom")
		}
	})

	t.Run("skipped clears running without failure", func(t *testing.T) {
		rt := &consoleLocalRuntime{heartbeatState: &heartbeatutil.State{}}
		if ok := rt.heartbeatState.Start(); !ok {
			t.Fatal("Start() = false, want true")
		}

		rt.completeHeartbeatTask(consoleLocalTaskJob{
			Trigger: daemonruntime.TaskTrigger{Source: "heartbeat"},
		}, heartbeatTaskResultSkipped, nil, time.Time{})

		failures, lastSuccess, lastError, running := rt.heartbeatState.Snapshot()
		if running {
			t.Fatal("running = true, want false")
		}
		if failures != 0 {
			t.Fatalf("failures = %d, want 0", failures)
		}
		if !lastSuccess.IsZero() {
			t.Fatalf("lastSuccess = %v, want zero", lastSuccess)
		}
		if lastError != "" {
			t.Fatalf("lastError = %q, want empty", lastError)
		}
	})
}

func TestConsoleTopicTitleFromOutput(t *testing.T) {
	t.Run("short output becomes title", func(t *testing.T) {
		got := consoleTopicTitleFromOutput("  Short answer.  ")
		if got != "Short answer" {
			t.Fatalf("consoleTopicTitleFromOutput() = %q, want %q", got, "Short answer")
		}
	})

	t.Run("long output requires llm", func(t *testing.T) {
		got := consoleTopicTitleFromOutput(strings.Repeat("a", consoleTopicTitleDirectOutputMaxRunes+1))
		if got != "" {
			t.Fatalf("consoleTopicTitleFromOutput() = %q, want empty", got)
		}
	})
}

func TestConsoleLocalRuntimeMaybeRefreshTopicTitleUsesShortOutput(t *testing.T) {
	store, err := daemonruntime.NewConsoleFileStore(daemonruntime.ConsoleFileStoreOptions{
		HeartbeatTopicID: "_heartbeat",
		Persist:          false,
	})
	if err != nil {
		t.Fatalf("NewConsoleFileStore() error = %v", err)
	}
	topic, err := store.CreateTopic("seed title")
	if err != nil {
		t.Fatalf("CreateTopic() error = %v", err)
	}

	rt := &consoleLocalRuntime{store: store}
	rt.maybeRefreshTopicTitle(consoleLocalTaskJob{
		TopicID:         topic.ID,
		Task:            "first task",
		AutoRenameTopic: true,
	}, "Direct title")

	updated, ok := store.GetTopic(topic.ID)
	if !ok || updated == nil {
		t.Fatalf("GetTopic(%q) missing", topic.ID)
	}
	if updated.Title != "Direct title" {
		t.Fatalf("updated.Title = %q, want %q", updated.Title, "Direct title")
	}
	if updated.LLMTitleGeneratedAt != nil {
		t.Fatal("updated.LLMTitleGeneratedAt != nil, want nil for direct title path")
	}
}
