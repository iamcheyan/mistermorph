package line

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func TestBuildLineHistoryMessageWithImageParts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	if err := os.WriteFile(path, tinyPNG, 0o600); err != nil {
		t.Fatalf("write image: %v", err)
	}

	msg, err := buildLineHistoryMessage("hello", "gpt-5.2", []string{path}, nil)
	if err != nil {
		t.Fatalf("buildLineHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 2 {
		t.Fatalf("parts len = %d, want 2", len(msg.Parts))
	}
	if msg.Parts[0].Type != "text" {
		t.Fatalf("first part type = %q, want text", msg.Parts[0].Type)
	}
	if msg.Parts[1].Type != "image_base64" {
		t.Fatalf("second part type = %q, want image_base64", msg.Parts[1].Type)
	}
	if msg.Parts[1].MIMEType != "image/png" {
		t.Fatalf("second part mime = %q, want image/png", msg.Parts[1].MIMEType)
	}
	raw, err := base64.StdEncoding.DecodeString(msg.Parts[1].DataBase64)
	if err != nil {
		t.Fatalf("decode image base64: %v", err)
	}
	if string(raw) != string(tinyPNG) {
		t.Fatalf("image payload mismatch")
	}
}

func TestBuildLineHistoryMessageUnsupportedModel(t *testing.T) {
	t.Parallel()

	msg, err := buildLineHistoryMessage("hello", "text-only-model", []string{"/tmp/x.png"}, nil)
	if err != nil {
		t.Fatalf("buildLineHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 0 {
		t.Fatalf("parts len = %d, want 0", len(msg.Parts))
	}
}

func TestBuildLineHistoryMessageImageTooLarge(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "large.png")
	large := make([]byte, lineLLMMaxImageBytes+1)
	if err := os.WriteFile(path, large, 0o600); err != nil {
		t.Fatalf("write image: %v", err)
	}

	_, err := buildLineHistoryMessage("hello", "gpt-5.2", []string{path}, nil)
	if err == nil {
		t.Fatalf("buildLineHistoryMessage() expected error")
	}
	if !strings.Contains(err.Error(), "图片太大") {
		t.Fatalf("error = %v, want 图片太大", err)
	}
}

func TestDownloadLineImageToCache(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/bot/message/m_1001/content" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/v2/bot/message/m_1001/content")
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(tinyPNG)
	}))
	defer srv.Close()

	api := newLineAPI(srv.Client(), srv.URL, "line-token")
	dir := t.TempDir()
	path, err := downloadLineImageToCache(context.Background(), api, dir, "m_1001", 1024*1024)
	if err != nil {
		t.Fatalf("downloadLineImageToCache() error = %v", err)
	}
	if filepath.Ext(path) != ".png" {
		t.Fatalf("downloaded extension = %q, want .png", filepath.Ext(path))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(raw) != string(tinyPNG) {
		t.Fatalf("downloaded content mismatch")
	}
}

func TestDownloadLineImageToCacheUnsupportedMime(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		_, _ = io.WriteString(w, "gif-data")
	}))
	defer srv.Close()

	api := newLineAPI(srv.Client(), srv.URL, "line-token")
	_, err := downloadLineImageToCache(context.Background(), api, t.TempDir(), "m_1002", 1024*1024)
	if err == nil {
		t.Fatalf("downloadLineImageToCache() expected error")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("error = %v, want unsupported error", err)
	}
}
