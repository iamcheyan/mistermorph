package consolecmd

import (
	"strings"
	"testing"

	heartbeatloop "github.com/quailyquaily/mistermorph/internal/channelruntime/heartbeat"
	"github.com/quailyquaily/mistermorph/internal/daemonruntime"
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
