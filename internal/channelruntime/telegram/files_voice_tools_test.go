package telegram

import "testing"

func TestCollectDownloadedImagePaths(t *testing.T) {
	files := []telegramDownloadedFile{
		{Kind: "document", MimeType: "application/pdf", Path: "/tmp/a.pdf"},
		{Kind: "photo", Path: "/tmp/p1.jpg"},
		{Kind: "document", MimeType: "image/png", Path: "/tmp/p2.png"},
		{Kind: "document", OriginalName: "x.webp", Path: "/tmp/p3.webp"},
		{Kind: "photo", Path: "/tmp/p1.jpg"},
	}
	got := collectDownloadedImagePaths(files, 3)
	if len(got) != 3 {
		t.Fatalf("collectDownloadedImagePaths() len = %d, want 3", len(got))
	}
	if got[0] != "/tmp/p1.jpg" || got[1] != "/tmp/p2.png" || got[2] != "/tmp/p3.webp" {
		t.Fatalf("collectDownloadedImagePaths() = %#v", got)
	}
}

func TestCollectDownloadedImagePathsMaxZero(t *testing.T) {
	files := []telegramDownloadedFile{{Kind: "photo", Path: "/tmp/p1.jpg"}}
	got := collectDownloadedImagePaths(files, 0)
	if len(got) != 0 {
		t.Fatalf("collectDownloadedImagePaths(max=0) = %#v, want nil/empty", got)
	}
}
