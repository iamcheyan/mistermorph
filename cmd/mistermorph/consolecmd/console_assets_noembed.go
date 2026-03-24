//go:build noembedconsole

package consolecmd

import "io/fs"

var consoleStaticFS fs.FS

func embeddedConsoleAssetsEnabled() bool {
	return false
}
