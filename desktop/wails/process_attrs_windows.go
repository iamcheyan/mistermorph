//go:build wailsdesktop && windows

package main

import (
	"os/exec"
	"syscall"
)

const desktopCreateNoWindow = 0x08000000

func applyDesktopChildProcessAttrs(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: desktopCreateNoWindow,
	}
}
