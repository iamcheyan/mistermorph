//go:build wailsdesktop && windows

package main

import (
	"os/exec"
	"testing"
)

func TestApplyDesktopChildProcessAttrs(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "echo", "ok")

	applyDesktopChildProcessAttrs(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatalf("SysProcAttr = nil")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatalf("HideWindow = false, want true")
	}
	if cmd.SysProcAttr.CreationFlags&desktopCreateNoWindow == 0 {
		t.Fatalf("CreationFlags = %#x, want %#x bit set", cmd.SysProcAttr.CreationFlags, desktopCreateNoWindow)
	}
}
