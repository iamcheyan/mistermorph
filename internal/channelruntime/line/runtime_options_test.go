package line

import (
	"testing"
	"time"
)

func TestNormalizeRuntimeLoopOptionsDefaults(t *testing.T) {
	t.Parallel()

	opts := normalizeRuntimeLoopOptions(runtimeLoopOptions{})
	if opts.TaskTimeout != 10*time.Minute {
		t.Fatalf("task timeout = %s, want %s", opts.TaskTimeout, 10*time.Minute)
	}
	if opts.MaxConcurrency != 3 {
		t.Fatalf("max concurrency = %d, want 3", opts.MaxConcurrency)
	}
	if opts.WebhookListen != "127.0.0.1:18080" {
		t.Fatalf("webhook listen = %q, want %q", opts.WebhookListen, "127.0.0.1:18080")
	}
	if opts.WebhookPath != "/line/webhook" {
		t.Fatalf("webhook path = %q, want %q", opts.WebhookPath, "/line/webhook")
	}
}

func TestNormalizeRuntimeLoopOptionsWebhookPath(t *testing.T) {
	t.Parallel()

	opts := normalizeRuntimeLoopOptions(runtimeLoopOptions{
		WebhookPath: "line/hook",
	})
	if opts.WebhookPath != "/line/hook" {
		t.Fatalf("webhook path = %q, want %q", opts.WebhookPath, "/line/hook")
	}
}

func TestNormalizeRunStringSlice(t *testing.T) {
	t.Parallel()

	got := normalizeRunStringSlice([]string{" G1 ", "", "G2", "G1"})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0] != "G1" || got[1] != "G2" {
		t.Fatalf("values = %#v, want [G1 G2]", got)
	}
}
