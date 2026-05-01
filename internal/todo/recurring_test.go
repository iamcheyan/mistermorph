package todo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseAndRenderRECUR(t *testing.T) {
	raw := `---
created_at: "1970-01-01T00:00:00Z"
updated_at: "1970-01-01T00:00:00Z"
recurring_count: 1
---

# TODO Recurring

- [ ] [Next](2026-05-02 09:00), [Repeat](daily), [TZ](Asia/Tokyo), [ChatID](tg:-100123) | Remind [John](tg:@john) to submit report.
`
	file, err := ParseRECUR(raw)
	if err != nil {
		t.Fatalf("ParseRECUR() error = %v", err)
	}
	if len(file.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(file.Entries))
	}
	entry := file.Entries[0]
	if entry.NextAt != "2026-05-02 09:00" || entry.Repeat != "daily" || entry.TZ != "Asia/Tokyo" || entry.ChatID != "tg:-100123" {
		t.Fatalf("entry metadata mismatch: %#v", entry)
	}
	rendered := RenderRECUR(file)
	if !strings.Contains(rendered, "[Next](2026-05-02 09:00), [Repeat](daily), [TZ](Asia/Tokyo), [ChatID](tg:-100123)") {
		t.Fatalf("rendered recurring entry missing metadata:\n%s", rendered)
	}
}

func TestParseRECURIgnoresHTMLCommentExamples(t *testing.T) {
	raw := `---
created_at: "1970-01-01T00:00:00Z"
updated_at: "1970-01-01T00:00:00Z"
recurring_count: 0
---

# TODO Recurring

<!--
- [ ] [Next](2026-05-02 09:00), [Repeat](daily) | Example only.
-->
`
	file, err := ParseRECUR(raw)
	if err != nil {
		t.Fatalf("ParseRECUR() error = %v", err)
	}
	if len(file.Entries) != 0 {
		t.Fatalf("entries = %d, want 0", len(file.Entries))
	}
}

func TestMaterializeDueRecurring(t *testing.T) {
	root := t.TempDir()
	store := NewStore(filepath.Join(root, "TODO.md"), filepath.Join(root, "TODO.DONE.md"))
	store.Now = func() time.Time {
		return time.Date(2026, 5, 2, 9, 30, 0, 0, time.UTC)
	}
	recur := RECURFile{
		CreatedAt: "1970-01-01T00:00:00Z",
		UpdatedAt: "1970-01-01T00:00:00Z",
		Entries: []RecurringEntry{
			{
				NextAt:  "2026-05-02 09:00",
				Repeat:  "daily",
				ChatID:  "tg:-100123",
				Content: "Remind [John](tg:@john) to submit report.",
			},
			{
				NextAt:  "2026-05-03 09:00",
				Repeat:  "weekly",
				Content: "Future task.",
			},
		},
	}
	if err := store.writeRECUR(recur); err != nil {
		t.Fatalf("writeRECUR() error = %v", err)
	}

	result, err := store.MaterializeDueRecurring()
	if err != nil {
		t.Fatalf("MaterializeDueRecurring() error = %v", err)
	}
	if result.Generated != 1 || result.Advanced != 1 {
		t.Fatalf("result = %#v, want one generated and advanced", result)
	}
	if _, err := os.Stat(filepath.Join(root, "TODO.DONE.md")); !os.IsNotExist(err) {
		t.Fatalf("TODO.DONE.md should not be created by recurring materialize, stat err=%v", err)
	}

	wip, _, err := store.readFiles()
	if err != nil {
		t.Fatalf("readFiles() error = %v", err)
	}
	if len(wip.Entries) != 1 {
		t.Fatalf("wip entries = %d, want 1", len(wip.Entries))
	}
	if got := wip.Entries[0].Content; got != "2026-05-02 09:00 Remind [John](tg:@john) to submit report." {
		t.Fatalf("materialized content = %q", got)
	}
	if wip.Entries[0].ChatID != "tg:-100123" {
		t.Fatalf("chat id = %q, want tg:-100123", wip.Entries[0].ChatID)
	}

	updated, _, err := store.readRECUR(store.nowUTC())
	if err != nil {
		t.Fatalf("readRECUR() error = %v", err)
	}
	if got := updated.Entries[0].NextAt; got != "2026-05-03 09:00" {
		t.Fatalf("advanced next = %q, want 2026-05-03 09:00", got)
	}
	if got := updated.Entries[1].NextAt; got != "2026-05-03 09:00" {
		t.Fatalf("future next = %q, want unchanged", got)
	}
}

func TestMaterializeDueRecurringAdvancesPastNow(t *testing.T) {
	next, err := nextRecurringTimeAfter(
		time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC),
		"every 3 days",
		time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("nextRecurringTimeAfter() error = %v", err)
	}
	if got := next.Format(TimestampLayout); got != "2026-05-10 09:00" {
		t.Fatalf("next = %q, want 2026-05-10 09:00", got)
	}
}

func TestMaterializeDueRecurringWithTimezone(t *testing.T) {
	root := t.TempDir()
	store := NewStore(filepath.Join(root, "TODO.md"), filepath.Join(root, "TODO.DONE.md"))
	store.Now = func() time.Time {
		return time.Date(2026, 5, 7, 6, 30, 0, 0, time.UTC)
	}
	recur := RECURFile{
		CreatedAt: "1970-01-01T00:00:00Z",
		UpdatedAt: "1970-01-01T00:00:00Z",
		Entries: []RecurringEntry{
			{
				NextAt:  "2026-05-07 15:00",
				Repeat:  "weekly",
				TZ:      "Asia/Tokyo",
				Content: "去打网球。",
			},
		},
	}
	if err := store.writeRECUR(recur); err != nil {
		t.Fatalf("writeRECUR() error = %v", err)
	}

	result, err := store.MaterializeDueRecurring()
	if err != nil {
		t.Fatalf("MaterializeDueRecurring() error = %v", err)
	}
	if result.Generated != 1 {
		t.Fatalf("generated = %d, want 1", result.Generated)
	}

	wip, _, err := store.readFiles()
	if err != nil {
		t.Fatalf("readFiles() error = %v", err)
	}
	if len(wip.Entries) != 1 {
		t.Fatalf("wip entries = %d, want 1", len(wip.Entries))
	}
	if got := wip.Entries[0].Content; got != "2026-05-07 15:00 (Asia/Tokyo) 去打网球。" {
		t.Fatalf("materialized content = %q", got)
	}

	updated, _, err := store.readRECUR(store.nowUTC())
	if err != nil {
		t.Fatalf("readRECUR() error = %v", err)
	}
	if got := updated.Entries[0].NextAt; got != "2026-05-14 15:00" {
		t.Fatalf("advanced next = %q, want 2026-05-14 15:00", got)
	}
	if got := updated.Entries[0].TZ; got != "Asia/Tokyo" {
		t.Fatalf("advanced tz = %q, want Asia/Tokyo", got)
	}
}

func TestMaterializeDueRecurringWithTimezoneNotDueYet(t *testing.T) {
	root := t.TempDir()
	store := NewStore(filepath.Join(root, "TODO.md"), filepath.Join(root, "TODO.DONE.md"))
	store.Now = func() time.Time {
		return time.Date(2026, 5, 7, 5, 59, 0, 0, time.UTC)
	}
	if err := store.writeRECUR(RECURFile{
		CreatedAt: "1970-01-01T00:00:00Z",
		UpdatedAt: "1970-01-01T00:00:00Z",
		Entries: []RecurringEntry{
			{
				NextAt:  "2026-05-07 15:00",
				Repeat:  "weekly",
				TZ:      "Asia/Tokyo",
				Content: "去打网球。",
			},
		},
	}); err != nil {
		t.Fatalf("writeRECUR() error = %v", err)
	}

	result, err := store.MaterializeDueRecurring()
	if err != nil {
		t.Fatalf("MaterializeDueRecurring() error = %v", err)
	}
	if result.Generated != 0 {
		t.Fatalf("generated = %d, want 0", result.Generated)
	}
}

func TestMaterializeDueRecurringDefaultsToLocalTimezone(t *testing.T) {
	oldLocal := time.Local
	time.Local = time.FixedZone("JST", 9*60*60)
	t.Cleanup(func() { time.Local = oldLocal })

	root := t.TempDir()
	store := NewStore(filepath.Join(root, "TODO.md"), filepath.Join(root, "TODO.DONE.md"))
	store.Now = func() time.Time {
		return time.Date(2026, 5, 7, 6, 0, 0, 0, time.UTC)
	}
	if err := store.writeRECUR(RECURFile{
		CreatedAt: "1970-01-01T00:00:00Z",
		UpdatedAt: "1970-01-01T00:00:00Z",
		Entries: []RecurringEntry{
			{
				NextAt:  "2026-05-07 15:00",
				Repeat:  "weekly",
				Content: "去打网球。",
			},
		},
	}); err != nil {
		t.Fatalf("writeRECUR() error = %v", err)
	}

	result, err := store.MaterializeDueRecurring()
	if err != nil {
		t.Fatalf("MaterializeDueRecurring() error = %v", err)
	}
	if result.Generated != 1 {
		t.Fatalf("generated = %d, want 1", result.Generated)
	}

	wip, _, err := store.readFiles()
	if err != nil {
		t.Fatalf("readFiles() error = %v", err)
	}
	if got := wip.Entries[0].Content; got != "2026-05-07 15:00 去打网球。" {
		t.Fatalf("materialized content = %q", got)
	}
}
