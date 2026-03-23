package consolecmd

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateConsoleStaticFS(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		err := validateConsoleStaticFS(fstest.MapFS{
			"index.html": {Data: []byte("<html></html>")},
		})
		if err != nil {
			t.Fatalf("validateConsoleStaticFS() error = %v", err)
		}
	})

	t.Run("missing index", func(t *testing.T) {
		err := validateConsoleStaticFS(fstest.MapFS{})
		if err == nil || !strings.Contains(err.Error(), "missing index.html") {
			t.Fatalf("validateConsoleStaticFS() error = %v, want missing index.html", err)
		}
	})
}
