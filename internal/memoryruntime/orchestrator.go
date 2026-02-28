package memoryruntime

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/quailyquaily/mistermorph/memory"
)

type OrchestratorOptions struct {
	Now        func() time.Time
	NewEventID func() string
}

type Orchestrator struct {
	manager   *memory.Manager
	journal   *memory.Journal
	projector *memory.Projector
	now       func() time.Time
	newEvent  func() string
}

type PrepareInjectionRequest struct {
	SubjectID      string
	RequestContext memory.RequestContext
	MaxItems       int
}

type RecordRequest struct {
	TaskRunID    string
	TSUTC        string
	SessionID    string
	SubjectID    string
	Channel      string
	Participants []memory.MemoryParticipant
	TaskText     string
	FinalOutput  string
	Draft        memory.SessionDraft
}

func New(manager *memory.Manager, journal *memory.Journal, projector *memory.Projector, opts OrchestratorOptions) (*Orchestrator, error) {
	if manager == nil {
		return nil, fmt.Errorf("memory manager is required")
	}
	if journal == nil {
		return nil, fmt.Errorf("memory journal is required")
	}
	if projector == nil {
		return nil, fmt.Errorf("memory projector is required")
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.NewEventID == nil {
		opts.NewEventID = func() string { return "evt_" + uuid.NewString() }
	}
	return &Orchestrator{
		manager:   manager,
		journal:   journal,
		projector: projector,
		now:       opts.Now,
		newEvent:  opts.NewEventID,
	}, nil
}

func (o *Orchestrator) PrepareInjection(req PrepareInjectionRequest) (string, error) {
	return o.manager.BuildInjection(req.SubjectID, req.RequestContext, req.MaxItems)
}

func (o *Orchestrator) Record(req RecordRequest) (memory.JournalOffset, error) {
	tsUTC := strings.TrimSpace(req.TSUTC)
	if tsUTC == "" {
		tsUTC = o.now().UTC().Format(time.RFC3339)
	}
	event := memory.MemoryEvent{
		SchemaVersion:     memory.CurrentMemoryEventSchemaVersion,
		EventID:           strings.TrimSpace(o.newEvent()),
		TaskRunID:         strings.TrimSpace(req.TaskRunID),
		TSUTC:             tsUTC,
		SessionID:         strings.TrimSpace(req.SessionID),
		SubjectID:         strings.TrimSpace(req.SubjectID),
		Channel:           strings.TrimSpace(req.Channel),
		Participants:      append([]memory.MemoryParticipant(nil), req.Participants...),
		TaskText:          strings.TrimSpace(req.TaskText),
		FinalOutput:       strings.TrimSpace(req.FinalOutput),
		DraftSummaryItems: normalizeDraftSummaryItems(req.Draft.SummaryItems),
		DraftPromote:      normalizePromoteDraft(req.Draft.Promote),
	}
	return o.journal.Append(event)
}

func normalizeDraftSummaryItems(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, raw := range items {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizePromoteDraft(promote memory.PromoteDraft) memory.PromoteDraft {
	goals := make([]string, 0, len(promote.GoalsProjects))
	goalSeen := make(map[string]bool, len(promote.GoalsProjects))
	for _, raw := range promote.GoalsProjects {
		goal := strings.TrimSpace(raw)
		if goal == "" {
			continue
		}
		key := strings.ToLower(goal)
		if goalSeen[key] {
			continue
		}
		goalSeen[key] = true
		goals = append(goals, goal)
	}

	facts := make([]memory.KVItem, 0, len(promote.KeyFacts))
	factSeen := make(map[string]bool, len(promote.KeyFacts))
	for _, kv := range promote.KeyFacts {
		title := strings.TrimSpace(kv.Title)
		value := strings.TrimSpace(kv.Value)
		if title == "" && value == "" {
			continue
		}
		key := strings.ToLower(title + "|" + value)
		if factSeen[key] {
			continue
		}
		factSeen[key] = true
		facts = append(facts, memory.KVItem{Title: title, Value: value})
	}

	if len(goals) == 0 && len(facts) == 0 {
		return memory.PromoteDraft{}
	}
	return memory.PromoteDraft{
		GoalsProjects: goals,
		KeyFacts:      facts,
	}
}
