package memory

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestJournalAppendAndRotateBySize(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 300,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")
	defer j.Close()

	first, err := j.Append(baseJournalEvent("evt_1", "run_1"))
	if err != nil {
		t.Fatalf("Append(first) error = %v", err)
	}
	second, err := j.Append(baseJournalEvent("evt_2", "run_2"))
	if err != nil {
		t.Fatalf("Append(second) error = %v", err)
	}
	third, err := j.Append(baseJournalEvent("evt_3", "run_3"))
	if err != nil {
		t.Fatalf("Append(third) error = %v", err)
	}

	if first.File != "since-2026-02-28-0001.jsonl" {
		t.Fatalf("first file = %q, want %q", first.File, "since-2026-02-28-0001.jsonl")
	}
	if second.File != "since-2026-02-28-0002.jsonl" {
		t.Fatalf("second file = %q, want %q", second.File, "since-2026-02-28-0002.jsonl")
	}
	if third.File != "since-2026-02-28-0003.jsonl" {
		t.Fatalf("third file = %q, want %q", third.File, "since-2026-02-28-0003.jsonl")
	}
	if first.Line != 1 || second.Line != 1 || third.Line != 1 {
		t.Fatalf("line offsets mismatch: first=%d second=%d third=%d, want all 1", first.Line, second.Line, third.Line)
	}
}

func TestJournalNoRotateByDateBoundary(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	defer j.Close()

	j.now = fixedNow("2026-02-28T23:59:59Z")
	a, err := j.Append(baseJournalEvent("evt_a", "run_a"))
	if err != nil {
		t.Fatalf("Append(day1) error = %v", err)
	}

	j.now = fixedNow("2026-03-01T00:00:00Z")
	b, err := j.Append(baseJournalEvent("evt_b", "run_b"))
	if err != nil {
		t.Fatalf("Append(day2) error = %v", err)
	}

	if a.File != "since-2026-02-28-0001.jsonl" {
		t.Fatalf("day1 file = %q, want %q", a.File, "since-2026-02-28-0001.jsonl")
	}
	if b.File != a.File {
		t.Fatalf("expected same file across day boundary when size not exceeded: got %q and %q", a.File, b.File)
	}
	if b.Line != 2 {
		t.Fatalf("second append line = %d, want 2", b.Line)
	}
}

func TestJournalRotateBySizeUsesCurrentDayAsSegmentStartDate(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 320,
	})
	defer j.Close()

	j.now = fixedNow("2026-02-28T23:59:59Z")
	a, err := j.Append(baseJournalEvent("evt_a", "run_a"))
	if err != nil {
		t.Fatalf("Append(day1) error = %v", err)
	}
	if a.File != "since-2026-02-28-0001.jsonl" {
		t.Fatalf("day1 file = %q, want %q", a.File, "since-2026-02-28-0001.jsonl")
	}

	j.now = fixedNow("2026-03-01T00:00:00Z")
	b, err := j.Append(baseJournalEvent("evt_b", "run_b"))
	if err != nil {
		t.Fatalf("Append(day2) error = %v", err)
	}

	if b.File != "since-2026-03-01-0001.jsonl" {
		t.Fatalf("day2 rotated file = %q, want %q", b.File, "since-2026-03-01-0001.jsonl")
	}
	if b.Line != 1 {
		t.Fatalf("day2 rotated line = %d, want 1", b.Line)
	}
}

func TestJournalRotateCompressesClosedSegments(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes:           320,
		CompressClosedSegments: true,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")
	defer j.Close()

	a, err := j.Append(baseJournalEvent("evt_a", "run_a"))
	if err != nil {
		t.Fatalf("Append(first) error = %v", err)
	}
	b, err := j.Append(baseJournalEvent("evt_b", "run_b"))
	if err != nil {
		t.Fatalf("Append(second) error = %v", err)
	}

	aPath := filepath.Join(root, "log", a.File)
	aGzPath := aPath + ".gz"
	if _, err := os.Stat(aPath); !os.IsNotExist(err) {
		t.Fatalf("expected closed segment to be removed after compression, stat(%s) err=%v", aPath, err)
	}
	if _, err := os.Stat(aGzPath); err != nil {
		t.Fatalf("expected compressed segment %s to exist, err=%v", aGzPath, err)
	}

	bPath := filepath.Join(root, "log", b.File)
	if _, err := os.Stat(bPath); err != nil {
		t.Fatalf("expected active segment %s to exist, err=%v", bPath, err)
	}
}

func TestJournalReplayReadsCompressedSegments(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes:           320,
		CompressClosedSegments: true,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")
	defer j.Close()

	off1, err := j.Append(baseJournalEvent("evt_1", "run_1"))
	if err != nil {
		t.Fatalf("Append(evt_1) error = %v", err)
	}
	_, err = j.Append(baseJournalEvent("evt_2", "run_2"))
	if err != nil {
		t.Fatalf("Append(evt_2) error = %v", err)
	}
	_, err = j.Append(baseJournalEvent("evt_3", "run_3"))
	if err != nil {
		t.Fatalf("Append(evt_3) error = %v", err)
	}

	var all []string
	nextAll, exhaustedAll, err := j.ReplayFrom(JournalOffset{}, 100, func(rec JournalRecord) error {
		all = append(all, rec.Event.EventID)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplayFrom(all) error = %v", err)
	}
	if !exhaustedAll {
		t.Fatalf("ReplayFrom(all) exhausted=false, want true")
	}
	if nextAll.File != "since-2026-02-28-0003.jsonl" || nextAll.Line != 1 {
		t.Fatalf("ReplayFrom(all) next = %#v, want file=since-2026-02-28-0003.jsonl line=1", nextAll)
	}
	if !reflect.DeepEqual(all, []string{"evt_1", "evt_2", "evt_3"}) {
		t.Fatalf("ReplayFrom(all) ids = %#v, want %#v", all, []string{"evt_1", "evt_2", "evt_3"})
	}

	var tail []string
	nextTail, exhaustedTail, err := j.ReplayFrom(off1, 100, func(rec JournalRecord) error {
		tail = append(tail, rec.Event.EventID)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplayFrom(off1) error = %v", err)
	}
	if !exhaustedTail {
		t.Fatalf("ReplayFrom(off1) exhausted=false, want true")
	}
	if nextTail.File != "since-2026-02-28-0003.jsonl" || nextTail.Line != 1 {
		t.Fatalf("ReplayFrom(off1) next = %#v, want file=since-2026-02-28-0003.jsonl line=1", nextTail)
	}
	if !reflect.DeepEqual(tail, []string{"evt_2", "evt_3"}) {
		t.Fatalf("ReplayFrom(off1) ids = %#v, want %#v", tail, []string{"evt_2", "evt_3"})
	}
}

func TestJournalRestartReusesLatestSegment(t *testing.T) {
	root := t.TempDir()

	j1 := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j1.now = fixedNow("2026-02-28T06:00:00Z")
	a, err := j1.Append(baseJournalEvent("evt_a", "run_a"))
	if err != nil {
		t.Fatalf("Append(first process) error = %v", err)
	}
	if err := j1.Close(); err != nil {
		t.Fatalf("Close(first process) error = %v", err)
	}

	j2 := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j2.now = fixedNow("2026-02-28T06:01:00Z")
	b, err := j2.Append(baseJournalEvent("evt_b", "run_b"))
	if err != nil {
		t.Fatalf("Append(second process) error = %v", err)
	}
	defer j2.Close()

	if a.File != "since-2026-02-28-0001.jsonl" {
		t.Fatalf("first segment file = %q, want %q", a.File, "since-2026-02-28-0001.jsonl")
	}
	if b.File != a.File {
		t.Fatalf("second append should reuse latest segment: got %q want %q", b.File, a.File)
	}
	if b.Line != 2 {
		t.Fatalf("second append line = %d, want 2", b.Line)
	}
}

func TestJournalReplayFrom_OrderedAndSkipsCheckpoint(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")
	defer j.Close()

	off1, err := j.Append(baseJournalEvent("evt_1", "run_1"))
	if err != nil {
		t.Fatalf("Append(evt_1) error = %v", err)
	}
	_, err = j.Append(baseJournalEvent("evt_2", "run_2"))
	if err != nil {
		t.Fatalf("Append(evt_2) error = %v", err)
	}
	j.now = fixedNow("2026-03-01T06:00:00Z")
	_, err = j.Append(baseJournalEvent("evt_3", "run_3"))
	if err != nil {
		t.Fatalf("Append(evt_3) error = %v", err)
	}

	var all []string
	_, exhaustedAll, err := j.ReplayFrom(JournalOffset{}, 100, func(rec JournalRecord) error {
		all = append(all, rec.Event.EventID)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplayFrom(all) error = %v", err)
	}
	if !exhaustedAll {
		t.Fatalf("ReplayFrom(all) exhausted=false, want true")
	}
	if !reflect.DeepEqual(all, []string{"evt_1", "evt_2", "evt_3"}) {
		t.Fatalf("ReplayFrom(all) ids = %#v, want %#v", all, []string{"evt_1", "evt_2", "evt_3"})
	}

	var tail []string
	_, exhaustedTail, err := j.ReplayFrom(off1, 100, func(rec JournalRecord) error {
		tail = append(tail, rec.Event.EventID)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplayFrom(off1) error = %v", err)
	}
	if !exhaustedTail {
		t.Fatalf("ReplayFrom(off1) exhausted=false, want true")
	}
	if !reflect.DeepEqual(tail, []string{"evt_2", "evt_3"}) {
		t.Fatalf("ReplayFrom(off1) ids = %#v, want %#v", tail, []string{"evt_2", "evt_3"})
	}
}

func TestJournalReplayFromLimit(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")
	defer j.Close()

	_, err := j.Append(baseJournalEvent("evt_1", "run_1"))
	if err != nil {
		t.Fatalf("Append(evt_1) error = %v", err)
	}
	_, err = j.Append(baseJournalEvent("evt_2", "run_2"))
	if err != nil {
		t.Fatalf("Append(evt_2) error = %v", err)
	}
	_, err = j.Append(baseJournalEvent("evt_3", "run_3"))
	if err != nil {
		t.Fatalf("Append(evt_3) error = %v", err)
	}

	var got []string
	next, exhausted, err := j.ReplayFrom(JournalOffset{}, 2, func(rec JournalRecord) error {
		got = append(got, rec.Event.EventID)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplayFrom(limit=2) error = %v", err)
	}
	if exhausted {
		t.Fatalf("ReplayFrom(limit=2) exhausted=true, want false")
	}
	if !reflect.DeepEqual(got, []string{"evt_1", "evt_2"}) {
		t.Fatalf("ReplayFrom(limit=2) ids = %#v, want %#v", got, []string{"evt_1", "evt_2"})
	}
	if next.File == "" || next.Line != 2 {
		t.Fatalf("ReplayFrom(limit=2) next = %#v, want line=2", next)
	}

	got = nil
	next2, exhausted2, err := j.ReplayFrom(next, 2, func(rec JournalRecord) error {
		got = append(got, rec.Event.EventID)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplayFrom(from-next) error = %v", err)
	}
	if !exhausted2 {
		t.Fatalf("ReplayFrom(from-next) exhausted=false, want true")
	}
	if !reflect.DeepEqual(got, []string{"evt_3"}) {
		t.Fatalf("ReplayFrom(from-next) ids = %#v, want %#v", got, []string{"evt_3"})
	}
	if next2.Line != 3 {
		t.Fatalf("ReplayFrom(from-next) next = %#v, want line=3", next2)
	}
}

func TestJournalReplayFromInvalidArgs(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")
	defer j.Close()

	_, _, err := j.ReplayFrom(JournalOffset{}, 0, func(rec JournalRecord) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "limit must be > 0") {
		t.Fatalf("ReplayFrom(limit=0) error = %v, want limit error", err)
	}
	_, _, err = j.ReplayFrom(JournalOffset{File: "since-2026-02-28-0001.jsonl.gz"}, 1, func(rec JournalRecord) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "offset.file is invalid") {
		t.Fatalf("ReplayFrom(file=.gz) error = %v, want offset.file is invalid", err)
	}
}

func TestJournalCheckpointRoundTrip(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")

	cp := JournalCheckpoint{
		File: "since-2026-02-28-0002.jsonl",
		Line: 18,
	}
	if err := j.SaveCheckpoint(cp); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}
	got, ok, err := j.LoadCheckpoint()
	if err != nil {
		t.Fatalf("LoadCheckpoint() error = %v", err)
	}
	if !ok {
		t.Fatalf("LoadCheckpoint() ok = false, want true")
	}
	if got.File != cp.File || got.Line != cp.Line {
		t.Fatalf("checkpoint mismatch: got %#v want file=%q line=%d", got, cp.File, cp.Line)
	}
	if strings.TrimSpace(got.UpdatedAt) == "" {
		t.Fatalf("checkpoint.updated_at is empty")
	}

	path := filepath.Join(root, "log", "checkpoint.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat checkpoint file error: %v", err)
	}
}

func TestJournalCheckpointRejectsCompressedFileName(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")

	err := j.SaveCheckpoint(JournalCheckpoint{
		File: "since-2026-02-28-0002.jsonl.gz",
		Line: 3,
	})
	if err == nil || !strings.Contains(err.Error(), "checkpoint.file is invalid") {
		t.Fatalf("SaveCheckpoint(file=.gz) error = %v, want checkpoint.file is invalid", err)
	}
}

func TestJournalAppendRejectsInvalidEvent(t *testing.T) {
	root := t.TempDir()
	j := NewJournal(root, JournalOptions{
		MaxFileBytes: 1024 * 1024,
	})
	j.now = fixedNow("2026-02-28T06:00:00Z")

	invalid := baseJournalEvent("", "run_1")
	_, err := j.Append(invalid)
	if err == nil || !strings.Contains(err.Error(), "event_id is required") {
		t.Fatalf("Append(invalid) error = %v, want event_id is required", err)
	}
}

func fixedNow(iso string) func() time.Time {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		panic(err)
	}
	return func() time.Time { return t }
}

func baseJournalEvent(eventID, runID string) MemoryEvent {
	return MemoryEvent{
		SchemaVersion: CurrentMemoryEventSchemaVersion,
		EventID:       eventID,
		TaskRunID:     runID,
		TSUTC:         "2026-02-28T06:15:12Z",
		SessionID:     "tg:-1003824466118",
		SubjectID:     "ext:telegram:28036192",
		Channel:       "telegram",
		Participants: []MemoryParticipant{
			{
				ID:       "@johnwick",
				Nickname: "John Wick",
				Protocol: "tg",
			},
		},
		TaskText:    "hello",
		FinalOutput: "world",
	}
}
