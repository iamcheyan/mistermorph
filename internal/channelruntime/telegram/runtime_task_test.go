package telegram

import (
	"context"
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
