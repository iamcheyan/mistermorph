//go:build wailsdesktop

package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildConsoleServeArgs(t *testing.T) {
	cfg := DesktopHostConfig{
		ConsoleBasePath: "console",
		ConfigPath:      "/tmp/morph.yaml",
	}
	args := buildConsoleServeArgs([]string{"console", "serve"}, cfg, "127.0.0.1:12345")
	want := []string{
		"console",
		"serve",
		"--console-listen", "127.0.0.1:12345",
		"--console-base-path", "/console",
		"--allow-empty-password",
		"--config", "/tmp/morph.yaml",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("buildConsoleServeArgs() mismatch\nwant: %#v\ngot : %#v", want, args)
	}
}

func TestExtractConfigPathFromArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "split form",
			args: []string{"--config", "/tmp/a.yaml"},
			want: filepath.Clean("/tmp/a.yaml"),
		},
		{
			name: "equals form",
			args: []string{"--config=/tmp/b.yaml"},
			want: filepath.Clean("/tmp/b.yaml"),
		},
		{
			name: "no config",
			args: []string{"--foo", "bar"},
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractConfigPathFromArgs(tc.args)
			if got != tc.want {
				t.Fatalf("extractConfigPathFromArgs() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeConsoleBasePath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"console", "/console"},
		{"/console", "/console"},
		{"/console/", "/console"},
	}
	for _, tc := range cases {
		if got := normalizeConsoleBasePath(tc.in); got != tc.want {
			t.Fatalf("normalizeConsoleBasePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestProxyHandlerRootPathPassesThroughWhenBasePathIsRoot(t *testing.T) {
	host := &DesktopHost{cfg: DesktopHostConfig{ConsoleBasePath: "/"}}
	host.proxy = nil

	req := httptest.NewRequest("GET", "http://desktop/", nil)
	rec := httptest.NewRecorder()
	host.ProxyHandler().ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("ProxyHandler root status = %d, want %d", rec.Code, 503)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Fatalf("ProxyHandler root should not redirect, got Location=%q", loc)
	}
}

func TestSameExecutablePath(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "mistermorph")
	if err := os.WriteFile(a, []byte("x"), 0o755); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if !sameExecutablePath(a, a) {
		t.Fatalf("same path should be true")
	}
	if sameExecutablePath(a, filepath.Join(dir, "other")) {
		t.Fatalf("different path should be false")
	}
}
