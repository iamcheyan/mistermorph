package consolecmd

import (
	"fmt"
	"io/fs"
)

func validateConsoleStaticFS(staticFS fs.FS) error {
	if staticFS == nil {
		return fmt.Errorf("console embedded assets missing index.html; run scripts/stage-console-assets.sh before building: static fs is nil")
	}
	if _, err := fs.Stat(staticFS, "index.html"); err != nil {
		return fmt.Errorf("console embedded assets missing index.html; run scripts/stage-console-assets.sh before building: %w", err)
	}
	return nil
}
