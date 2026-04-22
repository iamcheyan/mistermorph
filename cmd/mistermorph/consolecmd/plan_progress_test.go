package consolecmd

import (
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
)

func TestBuildConsolePlanProgressSkipsBlankSteps(t *testing.T) {
	progress := buildConsolePlanProgress(&agent.Plan{
		Steps: []agent.PlanStep{
			{Step: "  scan repo  ", Status: agent.PlanStatusCompleted},
			{Step: "   ", Status: agent.PlanStatusPending},
			{Step: "patch bug", Status: agent.PlanStatusInProgress},
		},
	})
	if progress == nil {
		t.Fatal("progress = nil")
	}
	if len(progress.Steps) != 2 {
		t.Fatalf("len(progress.Steps) = %d, want 2", len(progress.Steps))
	}
	if progress.Steps[0].Step != "scan repo" {
		t.Fatalf("progress.Steps[0].Step = %q, want %q", progress.Steps[0].Step, "scan repo")
	}
	if progress.Steps[0].Status != agent.PlanStatusCompleted {
		t.Fatalf("progress.Steps[0].Status = %q, want %q", progress.Steps[0].Status, agent.PlanStatusCompleted)
	}
	if progress.Steps[1].Step != "patch bug" {
		t.Fatalf("progress.Steps[1].Step = %q, want %q", progress.Steps[1].Step, "patch bug")
	}
	if progress.Steps[1].Status != agent.PlanStatusInProgress {
		t.Fatalf("progress.Steps[1].Status = %q, want %q", progress.Steps[1].Status, agent.PlanStatusInProgress)
	}
}

func TestBuildConsoleTaskProgressResultIncludesPlan(t *testing.T) {
	result := buildConsoleTaskProgressResult(&agent.Plan{
		Steps: []agent.PlanStep{
			{Step: "collect logs", Status: agent.PlanStatusInProgress},
			{Step: "ship fix", Status: agent.PlanStatusPending},
		},
	}, nil)
	if result == nil {
		t.Fatal("result = nil")
	}
	progress, ok := result["plan"].(*consolePlanProgress)
	if !ok || progress == nil {
		t.Fatalf("result.plan = %#v, want *consolePlanProgress", result["plan"])
	}
	if len(progress.Steps) != 2 {
		t.Fatalf("len(progress.Steps) = %d, want 2", len(progress.Steps))
	}
}
