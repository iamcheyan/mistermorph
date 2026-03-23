package consolecmd

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:static
var embeddedConsoleAssets embed.FS

var consoleStaticFS = mustConsoleStaticFS()

func mustConsoleStaticFS() fs.FS {
	staticFS, err := fs.Sub(embeddedConsoleAssets, "static")
	if err != nil {
		panic(fmt.Sprintf("console embedded assets unavailable: %v", err))
	}
	if err := validateConsoleStaticFS(staticFS); err != nil {
		panic(err.Error())
	}
	return staticFS
}

func validateConsoleStaticFS(staticFS fs.FS) error {
	if staticFS == nil {
		return fmt.Errorf("console embedded assets missing index.html; run scripts/stage-console-assets.sh before building: static fs is nil")
	}
	if _, err := fs.Stat(staticFS, "index.html"); err != nil {
		return fmt.Errorf("console embedded assets missing index.html; run scripts/stage-console-assets.sh before building: %w", err)
	}
	return nil
}
