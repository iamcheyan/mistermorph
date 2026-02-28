package telegram

import (
	"context"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/memory"
)

func TestShouldWriteMemory(t *testing.T) {
	mgr := &memory.Manager{}

	tests := []struct {
		name              string
		publishText       bool
		memManager        *memory.Manager
		longTermSubjectID string
		want              bool
	}{
		{
			name:              "skip when output is not published",
			publishText:       false,
			memManager:        mgr,
			longTermSubjectID: "ext:telegram:1",
			want:              false,
		},
		{
			name:              "skip when memory manager is missing",
			publishText:       true,
			memManager:        nil,
			longTermSubjectID: "ext:telegram:1",
			want:              false,
		},
		{
			name:              "write when long-term subject is resolved",
			publishText:       true,
			memManager:        mgr,
			longTermSubjectID: "ext:telegram:1",
			want:              true,
		},
		{
			name:              "skip when long-term subject is empty",
			publishText:       true,
			memManager:        mgr,
			longTermSubjectID: "",
			want:              false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldWriteMemory(tc.publishText, tc.memManager, tc.longTermSubjectID)
			if got != tc.want {
				t.Fatalf("shouldWriteMemory() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolveLongTermSubjectID(t *testing.T) {
	if got := resolveLongTermSubjectID(memory.Identity{Enabled: true, SubjectID: "ext:telegram:1"}); got != "ext:telegram:1" {
		t.Fatalf("normal subject = %q, want %q", got, "ext:telegram:1")
	}
	if got := resolveLongTermSubjectID(memory.Identity{Enabled: false, SubjectID: "ext:telegram:1"}); got != "" {
		t.Fatalf("disabled identity subject = %q, want empty", got)
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
