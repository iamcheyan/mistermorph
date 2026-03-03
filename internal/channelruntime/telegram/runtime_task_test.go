package telegram

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/llm"
)

func TestShouldWriteMemory(t *testing.T) {
	orchestrator := &memoryruntime.Orchestrator{}

	tests := []struct {
		name         string
		publishText  bool
		orchestrator *memoryruntime.Orchestrator
		subjectID    string
		want         bool
	}{
		{
			name:         "skip when output is not published",
			publishText:  false,
			orchestrator: orchestrator,
			subjectID:    "tg:1",
			want:         false,
		},
		{
			name:         "skip when orchestrator is missing",
			publishText:  true,
			orchestrator: nil,
			subjectID:    "tg:1",
			want:         false,
		},
		{
			name:         "write when subject is resolved",
			publishText:  true,
			orchestrator: orchestrator,
			subjectID:    "tg:1",
			want:         true,
		},
		{
			name:         "skip when subject is empty",
			publishText:  true,
			orchestrator: orchestrator,
			subjectID:    "",
			want:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldWriteMemory(tc.publishText, tc.orchestrator, tc.subjectID)
			if got != tc.want {
				t.Fatalf("shouldWriteMemory() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateTelegramPlanProgressMessageProgrammaticFormat(t *testing.T) {
	plan := &agent.Plan{
		Steps: []agent.PlanStep{
			{Step: "scan repo", Status: agent.PlanStatusCompleted},
			{Step: "patch bug", Status: agent.PlanStatusInProgress},
		},
	}
	msg, err := generateTelegramPlanProgressMessage(
		context.Background(),
		nil,
		"",
		"fix this flow",
		plan,
		agent.PlanStepUpdate{
			CompletedIndex: 0,
			CompletedStep:  "scan repo",
			StartedIndex:   1,
			StartedStep:    "patch bug",
		},
		0,
	)
	if err != nil {
		t.Fatalf("generateTelegramPlanProgressMessage() error = %v", err)
	}
	if msg != "scan repo" {
		t.Fatalf("message = %q, want %q", msg, "scan repo")
	}
}

func TestGenerateTelegramPlanProgressMessageChineseFallbackByPlanStep(t *testing.T) {
	plan := &agent.Plan{
		Steps: []agent.PlanStep{
			{Step: "检查日志", Status: agent.PlanStatusCompleted},
			{Step: "修复问题", Status: agent.PlanStatusPending},
		},
	}
	msg, err := generateTelegramPlanProgressMessage(
		context.Background(),
		nil,
		"",
		"请处理这个问题",
		plan,
		agent.PlanStepUpdate{
			CompletedIndex: 0,
			CompletedStep:  "",
			StartedIndex:   1,
			StartedStep:    "",
		},
		0,
	)
	if err != nil {
		t.Fatalf("generateTelegramPlanProgressMessage() error = %v", err)
	}
	if msg != "检查日志" {
		t.Fatalf("message = %q, want %q", msg, "检查日志")
	}
}

func TestTelegramOutputStreamExtractorPartialOutput(t *testing.T) {
	var ex telegramOutputStreamExtractor
	if changed := ex.Append(`{"type":"final","final":{"output":"hello`); !changed {
		t.Fatalf("expected changed on first append")
	}
	if got := ex.Output(); got != "hello" {
		t.Fatalf("output = %q, want hello", got)
	}
	if changed := ex.Append(` world"}}`); !changed {
		t.Fatalf("expected changed on second append")
	}
	if got := ex.Output(); got != "hello world" {
		t.Fatalf("output = %q, want hello world", got)
	}
}

func TestExtractTelegramFinalOutputFromJSONStreamEscapes(t *testing.T) {
	got, complete := extractTelegramFinalOutputFromJSONStream(`{"type":"final","final":{"output":"line1\nline2 \u4f60\u597d"}}`)
	if !complete {
		t.Fatalf("complete = false, want true")
	}
	if got != "line1\nline2 你好" {
		t.Fatalf("output = %q, want %q", got, "line1\nline2 你好")
	}
}

func TestTelegramDraftStreamPublisherPublishesDraftUpdates(t *testing.T) {
	var calls []telegramSendMessageDraftRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottoken/sendMessageDraft" {
			http.NotFound(w, r)
			return
		}
		var req telegramSendMessageDraftRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		calls = append(calls, req)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	api := newTelegramAPI(srv.Client(), srv.URL, "token")
	p := newTelegramDraftStreamPublisher(nil, api, 42, 77)
	if err := p.OnStream(llm.StreamEvent{Delta: `{"type":"final","final":{"output":"he`}); err != nil {
		t.Fatalf("OnStream() error = %v", err)
	}
	if err := p.OnStream(llm.StreamEvent{Delta: `llo"}}`}); err != nil {
		t.Fatalf("OnStream() error = %v", err)
	}
	if err := p.OnStream(llm.StreamEvent{Done: true}); err != nil {
		t.Fatalf("OnStream(done) error = %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("publish count = %d, want 2", len(calls))
	}
	if calls[0].Text != "he" || calls[1].Text != "hello" {
		t.Fatalf("texts = %#v, want [he hello]", []string{calls[0].Text, calls[1].Text})
	}
	if calls[0].DraftID != 77 || calls[1].DraftID != 77 {
		t.Fatalf("draft ids = %#v, want both 77", []int64{calls[0].DraftID, calls[1].DraftID})
	}
}

func TestTelegramDraftStreamPublisherFinalize(t *testing.T) {
	var calls []telegramSendMessageDraftRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottoken/sendMessageDraft" {
			http.NotFound(w, r)
			return
		}
		var req telegramSendMessageDraftRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		calls = append(calls, req)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	api := newTelegramAPI(srv.Client(), srv.URL, "token")
	p := newTelegramDraftStreamPublisher(nil, api, 42, 88)
	if ok := p.Finalize("hello"); !ok {
		t.Fatalf("Finalize() = false, want true")
	}
	if ok := p.Finalize("hello"); !ok {
		t.Fatalf("Finalize(same) = false, want true")
	}
	if len(calls) != 1 {
		t.Fatalf("finalize calls = %d, want 1", len(calls))
	}
}

func TestBuildTelegramHistoryMessageWithImageParts(t *testing.T) {
	orig := encodeImageToWebP
	encodeImageToWebP = func(raw []byte) ([]byte, error) { return []byte("webp-bytes"), nil }
	t.Cleanup(func() { encodeImageToWebP = orig })

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "x.jpg")
	if err := os.WriteFile(imgPath, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg, err := buildTelegramHistoryMessage("history", "gpt-5.2", []string{imgPath}, nil)
	if err != nil {
		t.Fatalf("buildTelegramHistoryMessage() error = %v", err)
	}
	if msg.Role != "user" {
		t.Fatalf("role = %q, want user", msg.Role)
	}
	if msg.Content != "history" {
		t.Fatalf("content = %q, want history", msg.Content)
	}
	if len(msg.Parts) != 2 {
		t.Fatalf("parts len = %d, want 2", len(msg.Parts))
	}
	if msg.Parts[0].Type != "text" || msg.Parts[0].Text != "history" {
		t.Fatalf("text part mismatch: %+v", msg.Parts[0])
	}
	if msg.Parts[1].Type != "image_base64" {
		t.Fatalf("image part type = %q, want image_base64", msg.Parts[1].Type)
	}
	if msg.Parts[1].MIMEType != "image/webp" {
		t.Fatalf("image part mime = %q, want image/webp", msg.Parts[1].MIMEType)
	}
	if msg.Parts[1].DataBase64 != base64.StdEncoding.EncodeToString([]byte("webp-bytes")) {
		t.Fatalf("image part data mismatch")
	}
}

func TestBuildTelegramHistoryMessageSkipsMissingAndCapsCount(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		path := filepath.Join(dir, "img_"+string(rune('a'+i))+".png")
		if err := os.WriteFile(path, []byte("ok"), 0o600); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", path, err)
		}
		paths = append(paths, path)
	}

	msg, err := buildTelegramHistoryMessage("history", "grok-4", append([]string{"/missing.png"}, paths...), nil)
	if err != nil {
		t.Fatalf("buildTelegramHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 4 {
		t.Fatalf("parts len = %d, want 4 (1 text + 3 images)", len(msg.Parts))
	}
	if msg.Parts[0].Type != "text" {
		t.Fatalf("parts[0] type = %q, want text", msg.Parts[0].Type)
	}
	for i := 1; i < len(msg.Parts); i++ {
		if msg.Parts[i].MIMEType != "image/png" {
			t.Fatalf("parts[%d] mime = %q, want image/png", i, msg.Parts[i].MIMEType)
		}
	}
}

func TestBuildTelegramHistoryMessageUnsupportedModelSkipsImageParts(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "x.jpg")
	if err := os.WriteFile(imgPath, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg, err := buildTelegramHistoryMessage("history", "qwen-max", []string{imgPath}, nil)
	if err != nil {
		t.Fatalf("buildTelegramHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 0 {
		t.Fatalf("parts len = %d, want 0", len(msg.Parts))
	}
	if msg.Content != "history" {
		t.Fatalf("content = %q, want history", msg.Content)
	}
}

func TestBuildTelegramHistoryMessageReturnsErrorWhenImageTooLarge(t *testing.T) {
	orig := encodeImageToWebP
	encodeImageToWebP = func(raw []byte) ([]byte, error) { return raw, nil }
	t.Cleanup(func() { encodeImageToWebP = orig })

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "too_large.jpg")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := f.Truncate(telegramLLMMaxImageBytes + 1); err != nil {
		_ = f.Close()
		t.Fatalf("Truncate() error = %v", err)
	}
	_ = f.Close()

	_, err = buildTelegramHistoryMessage("history", "gpt-5.2", []string{imgPath}, nil)
	if err == nil {
		t.Fatalf("buildTelegramHistoryMessage() expected error")
	}
	if !strings.Contains(err.Error(), "图片太大") {
		t.Fatalf("error = %q, want contains 图片太大", err.Error())
	}
}

func TestBuildTelegramHistoryMessageUsesWebPForSupportedModel(t *testing.T) {
	orig := encodeImageToWebP
	encodeImageToWebP = func(raw []byte) ([]byte, error) { return []byte("webp-bytes"), nil }
	t.Cleanup(func() { encodeImageToWebP = orig })

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "x.jpg")
	if err := os.WriteFile(imgPath, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg, err := buildTelegramHistoryMessage("history", "gpt-5.2", []string{imgPath}, nil)
	if err != nil {
		t.Fatalf("buildTelegramHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 2 {
		t.Fatalf("parts len = %d, want 2", len(msg.Parts))
	}
	if msg.Parts[1].MIMEType != "image/webp" {
		t.Fatalf("mime = %q, want image/webp", msg.Parts[1].MIMEType)
	}
	if msg.Parts[1].DataBase64 != base64.StdEncoding.EncodeToString([]byte("webp-bytes")) {
		t.Fatalf("data mismatch")
	}
}

func TestBuildTelegramHistoryMessageDoesNotForceWebPForUnsupportedModel(t *testing.T) {
	orig := encodeImageToWebP
	encodeImageToWebP = func(raw []byte) ([]byte, error) { return []byte("unexpected"), nil }
	t.Cleanup(func() { encodeImageToWebP = orig })

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "x.jpg")
	if err := os.WriteFile(imgPath, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg, err := buildTelegramHistoryMessage("history", "grok-4", []string{imgPath}, nil)
	if err != nil {
		t.Fatalf("buildTelegramHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 2 {
		t.Fatalf("parts len = %d, want 2", len(msg.Parts))
	}
	if msg.Parts[1].MIMEType != "image/jpeg" {
		t.Fatalf("mime = %q, want image/jpeg", msg.Parts[1].MIMEType)
	}
	if msg.Parts[1].DataBase64 != base64.StdEncoding.EncodeToString([]byte("abc")) {
		t.Fatalf("data mismatch")
	}
}

func TestBuildTelegramHistoryMessageSkipsUnsupportedImageFormats(t *testing.T) {
	orig := encodeImageToWebP
	encodeImageToWebP = func(raw []byte) ([]byte, error) { return []byte("unexpected"), nil }
	t.Cleanup(func() { encodeImageToWebP = orig })

	dir := t.TempDir()
	gifPath := filepath.Join(dir, "x.gif")
	if err := os.WriteFile(gifPath, []byte("gif-bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg, err := buildTelegramHistoryMessage("history", "gpt-5.2", []string{gifPath}, nil)
	if err != nil {
		t.Fatalf("buildTelegramHistoryMessage() error = %v", err)
	}
	if len(msg.Parts) != 0 {
		t.Fatalf("parts len = %d, want 0", len(msg.Parts))
	}
	if msg.Content != "history" {
		t.Fatalf("content = %q, want history", msg.Content)
	}
}
