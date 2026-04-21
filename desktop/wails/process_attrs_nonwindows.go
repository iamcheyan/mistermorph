//go:build wailsdesktop && !windows

package main

import "os/exec"

func applyDesktopChildProcessAttrs(cmd *exec.Cmd) {
	_ = cmd
}
