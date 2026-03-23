//go:build wailsdesktop

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/pathutil"
)

func resolveDesktopConfigPath(args []string) (string, bool) {
	if explicit := strings.TrimSpace(extractConfigPathFromArgs(args)); explicit != "" {
		return filepath.Clean(pathutil.ExpandHomePath(explicit)), true
	}

	for _, candidate := range []string{"config.yaml", "~/.morph/config.yaml"} {
		resolved := filepath.Clean(pathutil.ExpandHomePath(candidate))
		if _, err := os.Stat(resolved); err == nil {
			return resolved, false
		}
	}
	return "", false
}

func printDesktopConfigPath(scope string, cfgPath string, explicit bool) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "desktop"
	}
	source := "auto"
	if explicit {
		source = "explicit"
	}
	cfgPath = strings.TrimSpace(cfgPath)
	if cfgPath == "" {
		cfgPath = "(none)"
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s config path [%s]: %s\n", scope, source, cfgPath)
}

func extractConfigPathFromArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	for i := 0; i < len(args); i++ {
		item := strings.TrimSpace(args[i])
		if item == "" {
			continue
		}
		if item == "--config" && i+1 < len(args) {
			return strings.TrimSpace(pathutil.ExpandHomePath(args[i+1]))
		}
		if strings.HasPrefix(item, "--config=") {
			return strings.TrimSpace(pathutil.ExpandHomePath(strings.TrimPrefix(item, "--config=")))
		}
	}
	return ""
}
