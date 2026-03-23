//go:build wailsdesktop

package main

import "testing"

func TestDesktopAppIconEmbedded(t *testing.T) {
	if len(desktopAppIconPNG) == 0 {
		t.Fatal("desktop app icon is not embedded")
	}
}
