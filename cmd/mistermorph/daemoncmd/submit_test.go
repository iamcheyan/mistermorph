package daemoncmd

import (
	"strings"
	"testing"
)

func TestSubmitRequiresServerURL(t *testing.T) {
	cmd := NewSubmitCmd()
	cmd.SetArgs([]string{"--task", "hello", "--auth-token", "token"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --server-url is missing")
	}
	if !strings.Contains(err.Error(), "missing --server-url") {
		t.Fatalf("error = %q, want missing --server-url", err)
	}
}

func TestSubmitHelpMentionsRequiredServerURL(t *testing.T) {
	cmd := NewSubmitCmd()
	flag := cmd.Flags().Lookup("server-url")
	if flag == nil {
		t.Fatal("server-url flag missing")
	}
	if !strings.Contains(flag.Usage, "required") {
		t.Fatalf("server-url usage = %q, want required wording", flag.Usage)
	}
}
