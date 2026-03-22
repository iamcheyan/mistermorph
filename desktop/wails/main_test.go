//go:build wailsdesktop

package main

import "testing"

func TestBuildDesktopWindowOptions_UsesConsoleURL(t *testing.T) {
	consoleURL := "http://127.0.0.1:19080/console/"
	opts := buildDesktopWindowOptions(consoleURL)
	if opts.URL != consoleURL {
		t.Fatalf("buildDesktopWindowOptions() URL = %q, want %q", opts.URL, consoleURL)
	}
	if opts.Title != "MisterMorph" {
		t.Fatalf("buildDesktopWindowOptions() title = %q, want MisterMorph", opts.Title)
	}
}
