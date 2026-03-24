//go:build !noembedconsole

package consolecmd

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var embeddedConsoleAssets embed.FS

var consoleStaticFS = mustConsoleStaticFS()

func embeddedConsoleAssetsEnabled() bool {
	return true
}

func mustConsoleStaticFS() fs.FS {
	staticFS, err := fs.Sub(embeddedConsoleAssets, "static")
	if err != nil {
		panic("console embedded assets unavailable: " + err.Error())
	}
	if err := validateConsoleStaticFS(staticFS); err != nil {
		panic(err.Error())
	}
	return staticFS
}
