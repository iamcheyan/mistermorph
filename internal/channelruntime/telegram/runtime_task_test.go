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
			longTermSubjectID: heartbeatMemorySessionID,
			want:              false,
		},
		{
			name:              "skip when memory manager is missing",
			publishText:       true,
			memManager:        nil,
			longTermSubjectID: heartbeatMemorySessionID,
			want:              false,
		},
		{
			name:              "write when long-term subject is resolved",
			publishText:       true,
			memManager:        mgr,
			longTermSubjectID: heartbeatMemorySessionID,
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
	if got := resolveLongTermSubjectID(telegramJob{IsHeartbeat: true}, memory.Identity{}); got != heartbeatMemorySessionID {
		t.Fatalf("heartbeat subject = %q, want %q", got, heartbeatMemorySessionID)
	}
	if got := resolveLongTermSubjectID(telegramJob{}, memory.Identity{Enabled: true, SubjectID: "ext:telegram:1"}); got != "ext:telegram:1" {
		t.Fatalf("normal subject = %q, want %q", got, "ext:telegram:1")
	}
	if got := resolveLongTermSubjectID(telegramJob{}, memory.Identity{Enabled: false, SubjectID: "ext:telegram:1"}); got != "" {
		t.Fatalf("disabled identity subject = %q, want empty", got)
	}
}

func TestShouldSkipTaskMessage(t *testing.T) {
	if got := shouldSkipTaskMessage(telegramJob{IsHeartbeat: true}); got {
		t.Fatalf("heartbeat should not skip task message")
	}
	if got := shouldSkipTaskMessage(telegramJob{IsHeartbeat: false}); !got {
		t.Fatalf("non-heartbeat should skip task message")
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
	if msg != "patch bug" {
		t.Fatalf("message = %q, want %q", msg, "patch bug")
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
	if msg != "修复问题" {
		t.Fatalf("message = %q, want %q", msg, "修复问题")
	}
}
