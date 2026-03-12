package daemoncmd

import "testing"

func TestNewServeCmdIncludesInspectFlags(t *testing.T) {
	cmd := NewServeCmd(ServeDependencies{})
	if cmd.Flags().Lookup("inspect-prompt") == nil {
		t.Fatal("inspect-prompt flag missing")
	}
	if cmd.Flags().Lookup("inspect-request") == nil {
		t.Fatal("inspect-request flag missing")
	}
}
