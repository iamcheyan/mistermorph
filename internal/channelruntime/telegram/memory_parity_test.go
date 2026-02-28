package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/memory"
)

func TestTelegramMemoryE2Parity_WritebackGateFixtures(t *testing.T) {
	fixtures := loadTelegramMemoryE2Fixtures(t)
	for _, tc := range fixtures.WritebackGate {
		t.Run(tc.Name, func(t *testing.T) {
			var orchestrator *memoryruntime.Orchestrator
			if tc.HasOrchestrator {
				orchestrator = &memoryruntime.Orchestrator{}
			}
			got := shouldWriteMemory(tc.PublishText, orchestrator, tc.SubjectID)

			var mgr *memory.Manager
			if tc.HasOrchestrator {
				mgr = &memory.Manager{}
			}
			wantLegacy := legacyShouldWriteMemory(tc.PublishText, mgr, tc.SubjectID)
			if got != wantLegacy {
				t.Fatalf("parity mismatch: new=%v legacy=%v", got, wantLegacy)
			}
			if got != tc.Expected {
				t.Fatalf("shouldWriteMemory() = %v, want %v", got, tc.Expected)
			}
		})
	}
}

func TestTelegramMemoryE2Parity_RecordAndInjectionFixtures(t *testing.T) {
	fixtures := loadTelegramMemoryE2Fixtures(t)
	for _, fc := range fixtures.RecordCases {
		t.Run(fc.Name, func(t *testing.T) {
			root := t.TempDir()
			mgr := memory.NewManager(root, 7)
			seedFixtureMemory(t, mgr, fc)

			journal := mgr.NewJournal(memory.JournalOptions{MaxFileBytes: 1 << 20})
			projector := memory.NewProjector(mgr, journal, memory.ProjectorOptions{})
			now := mustParseRFC3339(t, "2026-03-01T09:10:00Z")
			orch, err := memoryruntime.New(mgr, journal, projector, memoryruntime.OrchestratorOptions{
				Now:        func() time.Time { return now },
				NewEventID: func() string { return "evt_" + uuid.NewString() },
			})
			if err != nil {
				t.Fatalf("memoryruntime.New() error = %v", err)
			}
			defer func() { _ = journal.Close() }()

			job := fc.Job.toTelegramJob(fc.TaskText)
			final := &agent.Final{Output: fc.FinalOutput}
			history := []chathistory.ChatHistoryItem{
				{
					Channel:  chathistory.ChannelTelegram,
					Kind:     chathistory.KindInboundUser,
					ChatID:   fmt.Sprintf("%d", fc.Job.ChatID),
					Text:     "previous message",
					SentAt:   now.Add(-5 * time.Minute),
					ChatType: strings.ToLower(strings.TrimSpace(fc.Job.ChatType)),
				},
			}
			client := stubMemoryDraftLLMClient{Response: fc.LLMDraftJSON}
			adapter := telegramMemoryRecordAdapter{
				ctx:        context.Background(),
				client:     client,
				model:      "test-model",
				manager:    mgr,
				job:        job,
				history:    history,
				historyCap: 8,
				final:      final,
			}

			gotReq, err := adapter.BuildRecordRequest()
			if err != nil {
				t.Fatalf("BuildRecordRequest() error = %v", err)
			}
			legacyReq, err := buildLegacyRecordRequestForParity(context.Background(), client, "test-model", mgr, job, history, 8, final)
			if err != nil {
				t.Fatalf("legacy record build error = %v", err)
			}
			if !reflect.DeepEqual(gotReq, legacyReq) {
				t.Fatalf("record request parity mismatch:\nnew=%#v\nlegacy=%#v", gotReq, legacyReq)
			}

			assertRecordRequestMatchesFixture(t, gotReq, fc.Expected)

			reqCtx := telegramMemoryRequestContext(job.ChatType)
			if reqCtx != memory.RequestContext(fc.Expected.RequestContext) {
				t.Fatalf("request context = %q, want %q", reqCtx, fc.Expected.RequestContext)
			}

			legacySnap, err := mgr.BuildInjection(fc.Expected.SubjectID, reqCtx, fc.MaxItems)
			if err != nil {
				t.Fatalf("legacy BuildInjection() error = %v", err)
			}
			newSnap, err := orch.PrepareInjectionWithAdapter(telegramMemoryInjectionAdapter{job: job}, fc.MaxItems)
			if err != nil {
				t.Fatalf("PrepareInjectionWithAdapter() error = %v", err)
			}
			if newSnap != legacySnap {
				t.Fatalf("injection parity mismatch\nnew:\n%s\nlegacy:\n%s", newSnap, legacySnap)
			}

			_, err = orch.RecordWithAdapter(adapter)
			if err != nil {
				t.Fatalf("RecordWithAdapter() error = %v", err)
			}
			var rec memory.JournalRecord
			_, _, err = journal.ReplayFrom(memory.JournalOffset{}, 1, func(r memory.JournalRecord) error {
				rec = r
				return nil
			})
			if err != nil {
				t.Fatalf("ReplayFrom() error = %v", err)
			}
			assertJournalEventMatchesFixture(t, rec.Event, fc.Expected)
		})
	}
}

func legacyShouldWriteMemory(publishText bool, mgr *memory.Manager, subjectID string) bool {
	if !publishText || mgr == nil {
		return false
	}
	return strings.TrimSpace(subjectID) != ""
}

func buildLegacyRecordRequestForParity(ctx context.Context, client llm.Client, model string, mgr *memory.Manager, job telegramJob, history []chathistory.ChatHistoryItem, historyCap int, final *agent.Final) (memoryruntime.RecordRequest, error) {
	output := depsutil.FormatFinalOutput(final)
	date := time.Now().UTC()
	meta := buildMemoryWriteMeta(job)

	ctxInfo := MemoryDraftContext{
		SessionID:          meta.SessionID,
		ChatID:             job.ChatID,
		ChatType:           job.ChatType,
		CounterpartyID:     job.FromUserID,
		CounterpartyName:   strings.TrimSpace(job.FromDisplayName),
		CounterpartyHandle: strings.TrimSpace(job.FromUsername),
		TimestampUTC:       date.Format(time.RFC3339),
	}
	if ctxInfo.CounterpartyName == "" {
		ctxInfo.CounterpartyName = strings.TrimSpace(strings.Join([]string{job.FromFirstName, job.FromLastName}, " "))
	}
	_, existingContent, _, err := mgr.LoadShortTerm(date, meta.SessionID)
	if err != nil {
		return memoryruntime.RecordRequest{}, err
	}
	ctxInfo.CounterpartyLabel = buildMemoryCounterpartyLabel(meta, ctxInfo)

	draftHistory := buildMemoryDraftHistory(history, job, output, date, historyCap)
	draft, err := BuildMemoryDraft(ctx, client, model, draftHistory, job.Text, output, existingContent, ctxInfo)
	if err != nil {
		return memoryruntime.RecordRequest{}, err
	}
	draft.Promote = EnforceLongTermPromotionRules(draft.Promote, nil, job.Text)

	taskRunID := strings.TrimSpace(job.TaskID)
	if taskRunID == "" {
		return memoryruntime.RecordRequest{}, fmt.Errorf("telegram task run id is required")
	}
	return memoryruntime.RecordRequest{
		TaskRunID:    taskRunID,
		SessionID:    telegramMemorySessionID(job),
		SubjectID:    telegramMemorySubjectID(job),
		Channel:      "telegram",
		Participants: telegramMemoryParticipants(job),
		TaskText:     strings.TrimSpace(job.Text),
		FinalOutput:  output,
		Draft:        draft,
	}, nil
}

func seedFixtureMemory(t *testing.T, mgr *memory.Manager, fixture telegramRecordFixture) {
	t.Helper()
	sessionID := telegramMemorySessionID(fixture.Job.toTelegramJob(""))
	day := time.Now().UTC()
	if len(fixture.ExistingShortTerm) > 0 {
		items := make([]memory.SummaryItem, 0, len(fixture.ExistingShortTerm))
		for i, content := range fixture.ExistingShortTerm {
			items = append(items, memory.SummaryItem{
				Created: fmt.Sprintf("2026-03-01 08:%02d", i),
				Content: strings.TrimSpace(content),
			})
		}
		_, err := mgr.WriteShortTerm(day, memory.ShortTermContent{SummaryItems: items}, memory.WriteMeta{SessionID: sessionID})
		if err != nil {
			t.Fatalf("WriteShortTerm() error = %v", err)
		}
	}
	if len(fixture.ExistingLongTermGoals) == 0 && len(fixture.ExistingLongTermFacts) == 0 {
		return
	}
	promote := memory.PromoteDraft{
		GoalsProjects: append([]string(nil), fixture.ExistingLongTermGoals...),
		KeyFacts:      toMemoryKVItems(fixture.ExistingLongTermFacts),
	}
	if _, err := mgr.UpdateLongTerm("ignored", promote); err != nil {
		t.Fatalf("UpdateLongTerm() error = %v", err)
	}
}

func assertRecordRequestMatchesFixture(t *testing.T, got memoryruntime.RecordRequest, want telegramRecordExpected) {
	t.Helper()
	if got.TaskRunID != want.TaskRunID {
		t.Fatalf("task_run_id = %q, want %q", got.TaskRunID, want.TaskRunID)
	}
	if got.SessionID != want.SessionID {
		t.Fatalf("session_id = %q, want %q", got.SessionID, want.SessionID)
	}
	if got.SubjectID != want.SubjectID {
		t.Fatalf("subject_id = %q, want %q", got.SubjectID, want.SubjectID)
	}
	if got.TaskText != want.TaskText {
		t.Fatalf("task_text = %q, want %q", got.TaskText, want.TaskText)
	}
	if got.FinalOutput != want.FinalOutput {
		t.Fatalf("final_output = %q, want %q", got.FinalOutput, want.FinalOutput)
	}
	if !reflect.DeepEqual(got.Draft.SummaryItems, want.SummaryItems) {
		t.Fatalf("summary_items = %#v, want %#v", got.Draft.SummaryItems, want.SummaryItems)
	}
	if !reflect.DeepEqual(normalizeStringSlice(got.Draft.Promote.GoalsProjects), normalizeStringSlice(want.PromoteGoals)) {
		t.Fatalf("promote_goals = %#v, want %#v", got.Draft.Promote.GoalsProjects, want.PromoteGoals)
	}
	if !reflect.DeepEqual(normalizeKVItems(got.Draft.Promote.KeyFacts), normalizeKVItems(toMemoryKVItems(want.PromoteKeyFacts))) {
		t.Fatalf("promote_key_facts = %#v, want %#v", got.Draft.Promote.KeyFacts, toMemoryKVItems(want.PromoteKeyFacts))
	}
	gotParticipants := canonicalParticipants(got.Participants)
	wantParticipants := canonicalFixtureParticipants(want.Participants)
	if !reflect.DeepEqual(gotParticipants, wantParticipants) {
		t.Fatalf("participants = %#v, want %#v", gotParticipants, wantParticipants)
	}
}

func assertJournalEventMatchesFixture(t *testing.T, got memory.MemoryEvent, want telegramRecordExpected) {
	t.Helper()
	if got.TaskRunID != want.TaskRunID {
		t.Fatalf("event.task_run_id = %q, want %q", got.TaskRunID, want.TaskRunID)
	}
	if got.SessionID != want.SessionID {
		t.Fatalf("event.session_id = %q, want %q", got.SessionID, want.SessionID)
	}
	if got.SubjectID != want.SubjectID {
		t.Fatalf("event.subject_id = %q, want %q", got.SubjectID, want.SubjectID)
	}
	if got.Channel != "telegram" {
		t.Fatalf("event.channel = %q, want telegram", got.Channel)
	}
	if got.TaskText != want.TaskText {
		t.Fatalf("event.task_text = %q, want %q", got.TaskText, want.TaskText)
	}
	if got.FinalOutput != want.FinalOutput {
		t.Fatalf("event.final_output = %q, want %q", got.FinalOutput, want.FinalOutput)
	}
	if !reflect.DeepEqual(got.DraftSummaryItems, want.SummaryItems) {
		t.Fatalf("event.draft_summary_items = %#v, want %#v", got.DraftSummaryItems, want.SummaryItems)
	}
	if !reflect.DeepEqual(normalizeStringSlice(got.DraftPromote.GoalsProjects), normalizeStringSlice(want.PromoteGoals)) {
		t.Fatalf("event.draft_promote.goals = %#v, want %#v", got.DraftPromote.GoalsProjects, want.PromoteGoals)
	}
	if !reflect.DeepEqual(normalizeKVItems(got.DraftPromote.KeyFacts), normalizeKVItems(toMemoryKVItems(want.PromoteKeyFacts))) {
		t.Fatalf("event.draft_promote.facts = %#v, want %#v", got.DraftPromote.KeyFacts, toMemoryKVItems(want.PromoteKeyFacts))
	}
	gotParticipants := canonicalParticipants(got.Participants)
	wantParticipants := canonicalFixtureParticipants(want.Participants)
	if !reflect.DeepEqual(gotParticipants, wantParticipants) {
		t.Fatalf("event.participants = %#v, want %#v", gotParticipants, wantParticipants)
	}
}

func canonicalParticipants(items []memory.MemoryParticipant) []telegramFixtureParticipant {
	if len(items) == 0 {
		return nil
	}
	out := make([]telegramFixtureParticipant, 0, len(items))
	for _, item := range items {
		out = append(out, telegramFixtureParticipant{
			ID:       strings.TrimSpace(fmt.Sprintf("%v", item.ID)),
			Nickname: strings.TrimSpace(item.Nickname),
			Protocol: strings.TrimSpace(item.Protocol),
		})
	}
	return out
}

func canonicalFixtureParticipants(items []telegramFixtureParticipant) []telegramFixtureParticipant {
	if len(items) == 0 {
		return nil
	}
	out := make([]telegramFixtureParticipant, 0, len(items))
	for _, item := range items {
		out = append(out, telegramFixtureParticipant{
			ID:       strings.TrimSpace(item.ID),
			Nickname: strings.TrimSpace(item.Nickname),
			Protocol: strings.TrimSpace(item.Protocol),
		})
	}
	return out
}

func toMemoryKVItems(items []telegramFixtureKV) []memory.KVItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]memory.KVItem, 0, len(items))
	for _, item := range items {
		out = append(out, memory.KVItem{
			Title: strings.TrimSpace(item.Title),
			Value: strings.TrimSpace(item.Value),
		})
	}
	return out
}

func normalizeStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeKVItems(items []memory.KVItem) []memory.KVItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]memory.KVItem, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		value := strings.TrimSpace(item.Value)
		if title == "" && value == "" {
			continue
		}
		out = append(out, memory.KVItem{Title: title, Value: value})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func loadTelegramMemoryE2Fixtures(t *testing.T) telegramMemoryE2Fixtures {
	t.Helper()
	path := filepath.Join("testdata", "memory_e2_fixtures.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s error = %v", path, err)
	}
	var out telegramMemoryE2Fixtures
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal fixture %s error = %v", path, err)
	}
	return out
}

type stubMemoryDraftLLMClient struct {
	Response string
	Err      error
}

func (s stubMemoryDraftLLMClient) Chat(_ context.Context, _ llm.Request) (llm.Result, error) {
	if s.Err != nil {
		return llm.Result{}, s.Err
	}
	return llm.Result{Text: s.Response}, nil
}

type telegramMemoryE2Fixtures struct {
	WritebackGate []telegramWritebackGateFixture `json:"writeback_gate"`
	RecordCases   []telegramRecordFixture        `json:"record_cases"`
}

type telegramWritebackGateFixture struct {
	Name            string `json:"name"`
	PublishText     bool   `json:"publish_text"`
	HasOrchestrator bool   `json:"has_orchestrator"`
	SubjectID       string `json:"subject_id"`
	Expected        bool   `json:"expected"`
}

type telegramRecordFixture struct {
	Name                  string                 `json:"name"`
	MaxItems              int                    `json:"max_items"`
	Job                   telegramFixtureJob     `json:"job"`
	TaskText              string                 `json:"task_text"`
	FinalOutput           string                 `json:"final_output"`
	LLMDraftJSON          string                 `json:"llm_draft_json"`
	ExistingShortTerm     []string               `json:"existing_short_term"`
	ExistingLongTermGoals []string               `json:"existing_long_term_goals"`
	ExistingLongTermFacts []telegramFixtureKV    `json:"existing_long_term_facts"`
	Expected              telegramRecordExpected `json:"expected"`
}

type telegramFixtureJob struct {
	TaskID          string   `json:"task_id"`
	ChatID          int64    `json:"chat_id"`
	MessageID       int64    `json:"message_id"`
	ChatType        string   `json:"chat_type"`
	FromUserID      int64    `json:"from_user_id"`
	FromUsername    string   `json:"from_username"`
	FromFirstName   string   `json:"from_first_name"`
	FromLastName    string   `json:"from_last_name"`
	FromDisplayName string   `json:"from_display_name"`
	MentionUsers    []string `json:"mention_users"`
}

func (j telegramFixtureJob) toTelegramJob(taskText string) telegramJob {
	return telegramJob{
		TaskID:          strings.TrimSpace(j.TaskID),
		ChatID:          j.ChatID,
		MessageID:       j.MessageID,
		ChatType:        strings.TrimSpace(j.ChatType),
		FromUserID:      j.FromUserID,
		FromUsername:    strings.TrimSpace(j.FromUsername),
		FromFirstName:   strings.TrimSpace(j.FromFirstName),
		FromLastName:    strings.TrimSpace(j.FromLastName),
		FromDisplayName: strings.TrimSpace(j.FromDisplayName),
		MentionUsers:    append([]string(nil), j.MentionUsers...),
		Text:            strings.TrimSpace(taskText),
	}
}

type telegramRecordExpected struct {
	TaskRunID       string                       `json:"task_run_id"`
	SessionID       string                       `json:"session_id"`
	SubjectID       string                       `json:"subject_id"`
	RequestContext  string                       `json:"request_context"`
	Participants    []telegramFixtureParticipant `json:"participants"`
	SummaryItems    []string                     `json:"summary_items"`
	PromoteGoals    []string                     `json:"promote_goals"`
	PromoteKeyFacts []telegramFixtureKV          `json:"promote_key_facts"`
	TaskText        string                       `json:"task_text"`
	FinalOutput     string                       `json:"final_output"`
}

type telegramFixtureParticipant struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Protocol string `json:"protocol"`
}

type telegramFixtureKV struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func mustParseRFC3339(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", raw, err)
	}
	return parsed
}
