package telegram

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFileCachePath(t *testing.T) {
	cacheDir := t.TempDir()
	filePath := filepath.Join(cacheDir, "nested", "photo.png")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filePath, []byte("png"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := resolveFileCachePath(cacheDir, "file_cache_dir/nested/photo.png", 1024)
	if err != nil {
		t.Fatalf("resolveFileCachePath() error = %v", err)
	}
	if got != filePath {
		t.Fatalf("path = %q, want %q", got, filePath)
	}
}

func TestResolveFileCachePathRejectsDirectory(t *testing.T) {
	cacheDir := t.TempDir()
	dirPath := filepath.Join(cacheDir, "nested")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	_, err := resolveFileCachePath(cacheDir, "nested", 1024)
	if err == nil {
		t.Fatalf("resolveFileCachePath() error = nil, want directory error")
	}
	if !strings.Contains(err.Error(), "path is a directory") {
		t.Fatalf("error = %v, want directory error", err)
	}
}

func TestResolveFileCachePathRejectsOutsideCacheDir(t *testing.T) {
	cacheDir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "photo.png")
	if err := os.WriteFile(outsidePath, []byte("png"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := resolveFileCachePath(cacheDir, outsidePath, 1024)
	if err == nil {
		t.Fatalf("resolveFileCachePath() error = nil, want outside-file error")
	}
	if !strings.Contains(err.Error(), "outside file_cache_dir") {
		t.Fatalf("error = %v, want outside file_cache_dir", err)
	}
}

func TestResolveFileCachePathRejectsTooLargeFile(t *testing.T) {
	cacheDir := t.TempDir()
	filePath := filepath.Join(cacheDir, "big.bin")
	if err := os.WriteFile(filePath, []byte("12345"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := resolveFileCachePath(cacheDir, "big.bin", 4)
	if err == nil {
		t.Fatalf("resolveFileCachePath() error = nil, want size error")
	}
	if !strings.Contains(err.Error(), "file too large to send") {
		t.Fatalf("error = %v, want size error", err)
	}
}
