package telegram

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
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

func TestBuildTelegramHistoryMessageWithImageParts(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "x.jpg")
	if err := os.WriteFile(imgPath, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg := buildTelegramHistoryMessage("history", "gpt-5.2", []string{imgPath}, nil)
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
	if msg.Parts[1].MIMEType != "image/jpeg" {
		t.Fatalf("image part mime = %q, want image/jpeg", msg.Parts[1].MIMEType)
	}
	if msg.Parts[1].DataBase64 != base64.StdEncoding.EncodeToString([]byte("abc")) {
		t.Fatalf("image part data mismatch")
	}
}

func TestLoadTelegramImagePartsSkipsMissingAndCapsCount(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		path := filepath.Join(dir, "img_"+string(rune('a'+i))+".png")
		if err := os.WriteFile(path, []byte("ok"), 0o600); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", path, err)
		}
		paths = append(paths, path)
	}

	parts := loadTelegramImageParts(append([]string{"/missing.png"}, paths...), nil)
	if len(parts) != 3 {
		t.Fatalf("parts len = %d, want 3", len(parts))
	}
	for i := range parts {
		if parts[i].MIMEType != "image/png" {
			t.Fatalf("parts[%d] mime = %q, want image/png", i, parts[i].MIMEType)
		}
	}
}

func TestBuildTelegramHistoryMessageUnsupportedModelSkipsImageParts(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "x.jpg")
	if err := os.WriteFile(imgPath, []byte("abc"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	msg := buildTelegramHistoryMessage("history", "qwen-max", []string{imgPath}, nil)
	if len(msg.Parts) != 0 {
		t.Fatalf("parts len = %d, want 0", len(msg.Parts))
	}
	if msg.Content != "history" {
		t.Fatalf("content = %q, want history", msg.Content)
	}
}
