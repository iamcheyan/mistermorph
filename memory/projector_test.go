package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/internal/entryutil"
)

func TestProjectorProjectOnce_ProjectsGroupedTargetsAndLongTerm(t *testing.T) {
	root := t.TempDir()
	mgr := NewManager(root, 7)
	j := mgr.NewJournal(JournalOptions{MaxFileBytes: 1 << 20})
	j.now = func() time.Time { return mustTimeRFC3339(t, "2026-02-28T06:00:00Z") }

	day := mustTimeRFC3339(t, "2026-02-28T06:00:00Z")
	e1 := baseProjectorEvent("evt_1", "run_1", "2026-02-28T06:01:00Z", "tg-1001", []string{"first item"})
	e2 := baseProjectorEvent("evt_2", "run_1", "2026-02-28T06:02:00Z", "tg-1001", []string{"second item"})
	e2.DraftPromote = PromoteDraft{GoalsProjects: []string{"Keep archive synced"}}
	e3 := baseProjectorEvent("evt_3", "run_2", "2026-02-28T06:03:00Z", "tg-1002", []string{"another room item"})

	if _, err := j.Append(e1); err != nil {
		t.Fatalf("Append(evt_1) error = %v", err)
	}
	if _, err := j.Append(e2); err != nil {
		t.Fatalf("Append(evt_2) error = %v", err)
	}
	if _, err := j.Append(e3); err != nil {
		t.Fatalf("Append(evt_3) error = %v", err)
	}

	p := NewProjector(mgr, j, ProjectorOptions{
		CheckpointBatch: 2,
	})

	got, err := p.ProjectOnce(context.Background(), 10)
	if err != nil {
		t.Fatalf("ProjectOnce() error = %v", err)
	}
	if got.Processed != 3 {
		t.Fatalf("ProjectOnce().Processed = %d, want 3", got.Processed)
	}
	if !got.Exhausted {
		t.Fatalf("ProjectOnce().Exhausted = false, want true")
	}

	_, content1001, ok1001, err := mgr.LoadShortTerm(day, "tg-1001")
	if err != nil {
		t.Fatalf("LoadShortTerm(tg-1001) error = %v", err)
	}
	if !ok1001 {
		t.Fatalf("LoadShortTerm(tg-1001) ok = false, want true")
	}
	if len(content1001.SummaryItems) != 2 {
		t.Fatalf("LoadShortTerm(tg-1001) summary count = %d, want 2", len(content1001.SummaryItems))
	}
	if content1001.SummaryItems[0].Content != "second item" || content1001.SummaryItems[1].Content != "first item" {
		t.Fatalf("LoadShortTerm(tg-1001) summary order = %#v, want [second item, first item]", content1001.SummaryItems)
	}

	_, content1002, ok1002, err := mgr.LoadShortTerm(day, "tg-1002")
	if err != nil {
		t.Fatalf("LoadShortTerm(tg-1002) error = %v", err)
	}
	if !ok1002 {
		t.Fatalf("LoadShortTerm(tg-1002) ok = false, want true")
	}
	if len(content1002.SummaryItems) != 1 || content1002.SummaryItems[0].Content != "another room item" {
		t.Fatalf("LoadShortTerm(tg-1002) content = %#v, want one item", content1002.SummaryItems)
	}

	longPath, _ := mgr.LongTermPath("ignored")
	data, err := os.ReadFile(longPath)
	if err != nil {
		t.Fatalf("ReadFile(long-term) error = %v", err)
	}
	_, body, _ := ParseFrontmatter(string(data))
	longContent := ParseLongTermContent(body)
	if len(longContent.Goals) != 1 || longContent.Goals[0].Content != "Keep archive synced" {
		t.Fatalf("long-term goals = %#v, want one promoted goal", longContent.Goals)
	}

	cp, ok, err := j.LoadCheckpoint()
	if err != nil {
		t.Fatalf("LoadCheckpoint() error = %v", err)
	}
	if !ok {
		t.Fatalf("LoadCheckpoint() ok = false, want true")
	}
	if cp.Line != 3 {
		t.Fatalf("checkpoint line = %d, want 3", cp.Line)
	}
}

func TestProjectorProjectOnce_RespectsReplayLimitAndAdvancesCheckpoint(t *testing.T) {
	root := t.TempDir()
	mgr := NewManager(root, 7)
	j := mgr.NewJournal(JournalOptions{MaxFileBytes: 1 << 20})
	j.now = func() time.Time { return mustTimeRFC3339(t, "2026-02-28T08:00:00Z") }

	events := []MemoryEvent{
		baseProjectorEvent("evt_1", "run_1", "2026-02-28T08:01:00Z", "tg-2001", []string{"one"}),
		baseProjectorEvent("evt_2", "run_1", "2026-02-28T08:02:00Z", "tg-2001", []string{"two"}),
		baseProjectorEvent("evt_3", "run_1", "2026-02-28T08:03:00Z", "tg-2001", []string{"three"}),
	}
	for i, ev := range events {
		if _, err := j.Append(ev); err != nil {
			t.Fatalf("Append(events[%d]) error = %v", i, err)
		}
	}

	p := NewProjector(mgr, j, ProjectorOptions{
		CheckpointBatch: 2,
	})

	first, err := p.ProjectOnce(context.Background(), 2)
	if err != nil {
		t.Fatalf("ProjectOnce(limit=2 first) error = %v", err)
	}
	if first.Processed != 2 {
		t.Fatalf("first.Processed = %d, want 2", first.Processed)
	}
	if first.Exhausted {
		t.Fatalf("first.Exhausted = true, want false")
	}
	cp1, ok, err := j.LoadCheckpoint()
	if err != nil {
		t.Fatalf("LoadCheckpoint(first) error = %v", err)
	}
	if !ok || cp1.Line != 2 {
		t.Fatalf("checkpoint after first = %#v, want line=2", cp1)
	}

	second, err := p.ProjectOnce(context.Background(), 2)
	if err != nil {
		t.Fatalf("ProjectOnce(limit=2 second) error = %v", err)
	}
	if second.Processed != 1 {
		t.Fatalf("second.Processed = %d, want 1", second.Processed)
	}
	if !second.Exhausted {
		t.Fatalf("second.Exhausted = false, want true")
	}
	cp2, ok, err := j.LoadCheckpoint()
	if err != nil {
		t.Fatalf("LoadCheckpoint(second) error = %v", err)
	}
	if !ok || cp2.Line != 3 {
		t.Fatalf("checkpoint after second = %#v, want line=3", cp2)
	}
}

func TestProjectorProjectOnce_ProjectionErrorStillAdvancesCheckpoint(t *testing.T) {
	root := t.TempDir()
	mgr := NewManager(root, 7)
	j := mgr.NewJournal(JournalOptions{MaxFileBytes: 1 << 20})
	j.now = func() time.Time { return mustTimeRFC3339(t, "2026-02-28T09:00:00Z") }

	day := mustTimeRFC3339(t, "2026-02-28T09:00:00Z")
	_, err := mgr.WriteShortTerm(day, ShortTermContent{
		SummaryItems: []SummaryItem{
			{Created: "2026-02-28 08:59", Content: "existing item"},
		},
	}, WriteMeta{SessionID: "tg-3001"})
	if err != nil {
		t.Fatalf("WriteShortTerm(existing) error = %v", err)
	}

	e1 := baseProjectorEvent("evt_1", "run_1", "2026-02-28T09:01:00Z", "tg-3001", []string{"alpha"})
	e2 := baseProjectorEvent("evt_2", "run_1", "2026-02-28T09:02:00Z", "tg-3001", []string{"beta"})
	if _, err := j.Append(e1); err != nil {
		t.Fatalf("Append(evt_1) error = %v", err)
	}
	if _, err := j.Append(e2); err != nil {
		t.Fatalf("Append(evt_2) error = %v", err)
	}

	p := NewProjector(mgr, j, ProjectorOptions{
		CheckpointBatch:  10,
		SemanticResolver: alwaysFailResolver{},
	})

	_, err = p.ProjectOnce(context.Background(), 10)
	if err == nil {
		t.Fatalf("ProjectOnce() error = nil, want semantic dedupe error")
	}
	if !strings.Contains(err.Error(), "semantic dedupe failed") {
		t.Fatalf("ProjectOnce() error = %v, want semantic dedupe failed", err)
	}

	cp, ok, cpErr := j.LoadCheckpoint()
	if cpErr != nil {
		t.Fatalf("LoadCheckpoint() error = %v", cpErr)
	}
	if !ok {
		t.Fatalf("LoadCheckpoint() ok = false, want true")
	}
	if cp.Line != 2 {
		t.Fatalf("checkpoint line = %d, want 2", cp.Line)
	}

	errorLog := filepath.Join(root, "log", "projector-error.jsonl")
	if _, statErr := os.Stat(errorLog); !os.IsNotExist(statErr) {
		t.Fatalf("projector error log should not exist, stat error = %v", statErr)
	}
}

func TestProjectorProjectOnce_RejectsEmptySubjectID(t *testing.T) {
	ev := baseProjectorEvent("evt_1", "run_1", "2026-02-28T10:01:00Z", "", []string{"alpha"})
	_, err := shortTermTargetFromEvent(ev)
	if err == nil {
		t.Fatalf("shortTermTargetFromEvent() error = nil, want subject_id is required")
	}
	if !strings.Contains(err.Error(), "subject_id is required") {
		t.Fatalf("shortTermTargetFromEvent() error = %v, want subject_id is required", err)
	}
}

type alwaysFailResolver struct{}

func (alwaysFailResolver) SelectDedupKeepIndices(ctx context.Context, items []entryutil.SemanticItem) ([]int, error) {
	return nil, fmt.Errorf("resolver failed")
}

func baseProjectorEvent(eventID, runID, tsUTC, subjectID string, summaryItems []string) MemoryEvent {
	return MemoryEvent{
		SchemaVersion:     CurrentMemoryEventSchemaVersion,
		EventID:           eventID,
		TaskRunID:         runID,
		TSUTC:             tsUTC,
		SessionID:         "tg:-1000",
		SubjectID:         subjectID,
		Channel:           "telegram",
		Participants:      nil,
		TaskText:          "task",
		FinalOutput:       "output",
		DraftSummaryItems: summaryItems,
		DraftPromote:      PromoteDraft{},
	}
}

func mustTimeRFC3339(t *testing.T, value string) time.Time {
	t.Helper()
	out, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", value, err)
	}
	return out
}
