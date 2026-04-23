package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCreditsCommand(t *testing.T) {
	cmd := newCreditsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Open Source",
		"Cobra",
		"Contributors",
		"Lyric Wai",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}
