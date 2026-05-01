package heartbeatutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildHeartbeatTaskMaterializesRecurringTodos(t *testing.T) {
	root := t.TempDir()
	checklistPath := filepath.Join(root, "HEARTBEAT.md")
	if err := os.WriteFile(checklistPath, []byte("## Check TODO.md\n\n- Check TODO.md\n"), 0o600); err != nil {
		t.Fatalf("write heartbeat checklist: %v", err)
	}
	dueAt := time.Now().UTC().Add(-1 * time.Hour).Format("2006-01-02 15:04")
	recurPath := filepath.Join(root, "TODO.RECUR.md")
	recur := `---
created_at: "1970-01-01T00:00:00Z"
updated_at: "1970-01-01T00:00:00Z"
recurring_count: 1
---

# TODO Recurring

- [ ] [Next](` + dueAt + `), [Repeat](daily) | Review open invoices.
`
	if err := os.WriteFile(recurPath, []byte(recur), 0o600); err != nil {
		t.Fatalf("write recurring todo: %v", err)
	}

	task, empty, err := BuildHeartbeatTask(checklistPath)
	if err != nil {
		t.Fatalf("BuildHeartbeatTask() error = %v", err)
	}
	if empty || !strings.Contains(task, "Check TODO.md") {
		t.Fatalf("heartbeat task mismatch: task=%q empty=%v", task, empty)
	}
	if !strings.Contains(task, "## Current TODO.md Open Items") || !strings.Contains(task, dueAt+" Review open invoices.") {
		t.Fatalf("heartbeat task missing current TODO snapshot:\n%s", task)
	}

	todoRaw, err := os.ReadFile(filepath.Join(root, "TODO.md"))
	if err != nil {
		t.Fatalf("read materialized TODO.md: %v", err)
	}
	if !strings.Contains(string(todoRaw), dueAt+" Review open invoices.") {
		t.Fatalf("TODO.md missing materialized recurring task:\n%s", string(todoRaw))
	}

	recurRaw, err := os.ReadFile(recurPath)
	if err != nil {
		t.Fatalf("read updated TODO.RECUR.md: %v", err)
	}
	if strings.Contains(string(recurRaw), "[Next]("+dueAt+")") {
		t.Fatalf("TODO.RECUR.md did not advance Next:\n%s", string(recurRaw))
	}
}

func TestBuildHeartbeatTaskRunsForOpenTodosWhenChecklistEmpty(t *testing.T) {
	root := t.TempDir()
	checklistPath := filepath.Join(root, "HEARTBEAT.md")
	if err := os.WriteFile(checklistPath, []byte("# Heartbeat\n\n"), 0o600); err != nil {
		t.Fatalf("write heartbeat checklist: %v", err)
	}
	todoText := `---
created_at: "1970-01-01T00:00:00Z"
updated_at: "1970-01-01T00:00:00Z"
open_count: 1
---

# TODO

- [ ] [Created](2026-05-01 12:41) | Review open invoices.
`
	if err := os.WriteFile(filepath.Join(root, "TODO.md"), []byte(todoText), 0o600); err != nil {
		t.Fatalf("write TODO.md: %v", err)
	}

	task, checklistEmpty, err := BuildHeartbeatTask(checklistPath)
	if err != nil {
		t.Fatalf("BuildHeartbeatTask() error = %v", err)
	}
	if !checklistEmpty {
		t.Fatalf("checklistEmpty = false, want true")
	}
	if !strings.Contains(task, "## Current TODO.md Open Items") || !strings.Contains(task, "Review open invoices.") {
		t.Fatalf("heartbeat task missing open TODO snapshot:\n%s", task)
	}
}

func TestBuildHeartbeatTaskMaterializesRecurringTodosWhenChecklistEmpty(t *testing.T) {
	root := t.TempDir()
	checklistPath := filepath.Join(root, "HEARTBEAT.md")
	if err := os.WriteFile(checklistPath, []byte("# Heartbeat\n\n"), 0o600); err != nil {
		t.Fatalf("write heartbeat checklist: %v", err)
	}
	dueAt := time.Now().UTC().Add(-1 * time.Hour).Format("2006-01-02 15:04")
	recurPath := filepath.Join(root, "TODO.RECUR.md")
	recur := `---
created_at: "1970-01-01T00:00:00Z"
updated_at: "1970-01-01T00:00:00Z"
recurring_count: 1
---

# TODO Recurring

- [ ] [Next](` + dueAt + `), [Repeat](daily) | Review open invoices.
`
	if err := os.WriteFile(recurPath, []byte(recur), 0o600); err != nil {
		t.Fatalf("write recurring todo: %v", err)
	}

	task, checklistEmpty, err := BuildHeartbeatTask(checklistPath)
	if err != nil {
		t.Fatalf("BuildHeartbeatTask() error = %v", err)
	}
	if !checklistEmpty {
		t.Fatalf("checklistEmpty = false, want true")
	}
	if !strings.Contains(task, "## Current TODO.md Open Items") || !strings.Contains(task, dueAt+" Review open invoices.") {
		t.Fatalf("heartbeat task missing materialized TODO snapshot:\n%s", task)
	}
}
