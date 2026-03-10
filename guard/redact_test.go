package guard

import (
	"strings"
	"testing"
)

func TestRedactor_RedactsMisterMorphEnvAssignments(t *testing.T) {
	r := NewRedactor(RedactionConfig{})
	raw := "MISTER_MORPH_API_KEY=abc123\nMISTER_MORPH_LARK_SECRET: short\n"

	got, changed := r.RedactString(raw)
	if !changed {
		t.Fatal("RedactString() changed = false, want true")
	}
	for _, needle := range []string{"MISTER_MORPH_API_KEY", "MISTER_MORPH_LARK_SECRET", "abc123", "short"} {
		if strings.Contains(got, needle) {
			t.Fatalf("redacted output leaked %q: %q", needle, got)
		}
	}
	if !strings.Contains(got, "[redacted_env]=[redacted]") {
		t.Fatalf("redacted output missing env assignment placeholder: %q", got)
	}
	if !strings.Contains(got, "[redacted_env]: [redacted]") {
		t.Fatalf("redacted output missing env kv placeholder: %q", got)
	}
}

func TestRedactor_RedactsStandaloneMisterMorphEnvNames(t *testing.T) {
	r := NewRedactor(RedactionConfig{})
	raw := "echo $MISTER_MORPH_API_KEY && printenv MISTER_MORPH_SLACK_APP_TOKEN"

	got, changed := r.RedactString(raw)
	if !changed {
		t.Fatal("RedactString() changed = false, want true")
	}
	if strings.Contains(got, "MISTER_MORPH_") {
		t.Fatalf("redacted output still contains env name: %q", got)
	}
	if !strings.Contains(got, "[redacted_env]") {
		t.Fatalf("redacted output missing placeholder: %q", got)
	}
}
