package consolecmd

import (
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
)

func TestFormatConsoleToolStartSuppressesPlanCreate(t *testing.T) {
	if got := formatConsoleToolStart(agent.Event{ToolName: "plan_create"}); got != "" {
		t.Fatalf("formatConsoleToolStart(plan_create) = %q, want empty", got)
	}
	if got := formatConsoleToolStart(agent.Event{ToolName: "web_search"}); got != "[web_search] running" {
		t.Fatalf("formatConsoleToolStart(web_search) = %q, want %q", got, "[web_search] running")
	}
}

func TestFormatConsoleToolDoneSuppressesPlanCreateSuccess(t *testing.T) {
	if got := formatConsoleToolDone(agent.Event{ToolName: "plan_create", Status: "done"}); got != "" {
		t.Fatalf("formatConsoleToolDone(plan_create success) = %q, want empty", got)
	}
	if got := formatConsoleToolDone(agent.Event{ToolName: "plan_create", Error: "boom"}); got == "" {
		t.Fatal("formatConsoleToolDone(plan_create failure) = empty, want failure preview")
	}
}
