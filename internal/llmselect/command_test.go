package llmselect

import (
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/internal/llmutil"
)

func TestExecuteCommandText_SetAndReset(t *testing.T) {
	values := llmutil.RuntimeValues{
		Provider: "openai",
		Model:    "gpt-5.2",
		Profiles: map[string]llmutil.ProfileConfig{
			"cheap": {Model: "gpt-4.1-mini"},
		},
		Routes: llmutil.RoutesConfig{
			PurposeRoutes: llmutil.PurposeRoutes{
				MainLoop: llmutil.RoutePolicyConfig{
					Candidates: []llmutil.RouteCandidateConfig{
						{Profile: llmutil.RouteProfileDefault, Weight: 80},
						{Profile: "cheap", Weight: 20},
					},
				},
			},
		},
	}
	store := NewStore()

	currentText, handled, err := ExecuteCommandText(values, store, "/model")
	if err != nil {
		t.Fatalf("ExecuteCommandText(/model) error = %v", err)
	}
	if !handled {
		t.Fatal("expected /model to be handled")
	}
	if !strings.Contains(currentText, "Current LLM selection: auto") {
		t.Fatalf("current text = %q, want auto selection", currentText)
	}
	if !strings.Contains(currentText, "weighted candidates") {
		t.Fatalf("current text = %q, want weighted candidates", currentText)
	}

	setText, handled, err := ExecuteCommandText(values, store, "/model set cheap")
	if err != nil {
		t.Fatalf("ExecuteCommandText(/model set) error = %v", err)
	}
	if !handled {
		t.Fatal("expected /model set to be handled")
	}
	if !strings.Contains(setText, "cheap") {
		t.Fatalf("set text = %q, want cheap profile", setText)
	}
	if got := store.Get(); got.Mode != ModeManual || got.ManualProfile != "cheap" {
		t.Fatalf("store.Get() = %#v, want manual cheap", got)
	}

	resetText, handled, err := ExecuteCommandText(values, store, "/model reset")
	if err != nil {
		t.Fatalf("ExecuteCommandText(/model reset) error = %v", err)
	}
	if !handled {
		t.Fatal("expected /model reset to be handled")
	}
	if !strings.Contains(resetText, "Current LLM selection: auto") {
		t.Fatalf("reset text = %q, want auto selection", resetText)
	}
	if got := store.Get(); got.Mode != ModeAuto {
		t.Fatalf("store.Get().Mode = %q, want auto", got.Mode)
	}
}

func TestExecuteCommandText_InvalidUsageHandled(t *testing.T) {
	values := llmutil.RuntimeValues{Provider: "openai", Model: "gpt-5.2"}
	_, handled, err := ExecuteCommandText(values, NewStore(), "/model set")
	if !handled {
		t.Fatal("expected invalid /model command to be handled")
	}
	if err == nil {
		t.Fatal("expected usage error")
	}
	if !strings.Contains(err.Error(), "/model set <profile_name>") {
		t.Fatalf("error = %q, want usage text", err.Error())
	}
}
