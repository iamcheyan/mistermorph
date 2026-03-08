package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/jsonutil"
	"github.com/quailyquaily/mistermorph/internal/llminspect"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/memory"
)

func recordMemoryFromJob(ctx context.Context, logger *slog.Logger, client llm.Client, model string, orchestrator *memoryruntime.Orchestrator, mgr *memory.Manager, job telegramJob, history []chathistory.ChatHistoryItem, historyCap int, final *agent.Final, requestTimeout time.Duration) error {
	recordOffset, err := orchestrator.RecordWithAdapter(telegramMemoryRecordAdapter{
		ctx:            ctx,
		client:         client,
		model:          model,
		manager:        mgr,
		job:            job,
		history:        history,
		historyCap:     historyCap,
		final:          final,
		requestTimeout: requestTimeout,
	})
	if err != nil {
		return err
	}
	if logger != nil {
		logger.Debug("memory_record_ok",
			"source", "telegram",
			"subject_id", telegramMemorySubjectID(job),
			"offset_file", recordOffset.File,
			"offset_line", recordOffset.Line,
		)
	}
	return nil
}

type telegramMemoryInjectionAdapter struct {
	job telegramJob
}

func (a telegramMemoryInjectionAdapter) ResolveSubjectID() (string, error) {
	return telegramMemorySubjectID(a.job), nil
}

func (a telegramMemoryInjectionAdapter) ResolveRequestContext() (memory.RequestContext, error) {
	return telegramMemoryRequestContext(a.job.ChatType), nil
}

type telegramMemoryRecordAdapter struct {
	ctx            context.Context
	client         llm.Client
	model          string
	manager        *memory.Manager
	job            telegramJob
	history        []chathistory.ChatHistoryItem
	historyCap     int
	final          *agent.Final
	requestTimeout time.Duration
}

func (a telegramMemoryRecordAdapter) BuildRecordRequest() (memoryruntime.RecordRequest, error) {
	output := depsutil.FormatFinalOutput(a.final)
	now := time.Now().UTC()
	meta := buildMemoryWriteMeta(a.job)
	taskRunID := strings.TrimSpace(a.job.TaskID)
	if taskRunID == "" {
		return memoryruntime.RecordRequest{}, fmt.Errorf("telegram task run id is required")
	}

	ctxInfo := MemoryDraftContext{
		SessionID:          meta.SessionID,
		ChatID:             a.job.ChatID,
		ChatType:           a.job.ChatType,
		CounterpartyID:     a.job.FromUserID,
		CounterpartyName:   strings.TrimSpace(a.job.FromDisplayName),
		CounterpartyHandle: strings.TrimSpace(a.job.FromUsername),
		TimestampUTC:       now.Format(time.RFC3339),
	}
	if ctxInfo.CounterpartyName == "" {
		ctxInfo.CounterpartyName = strings.TrimSpace(strings.Join([]string{a.job.FromFirstName, a.job.FromLastName}, " "))
	}

	_, existingContent, _, err := a.manager.LoadShortTerm(now, meta.SessionID)
	if err != nil {
		return memoryruntime.RecordRequest{}, err
	}

	memCtx := a.ctx
	if memCtx == nil {
		memCtx = context.Background()
	}
	cancel := func() {}
	if a.requestTimeout > 0 {
		memCtx, cancel = context.WithTimeout(memCtx, a.requestTimeout)
	}
	defer cancel()
	ctxInfo.CounterpartyLabel = buildMemoryCounterpartyLabel(meta, ctxInfo)

	draftHistory := buildMemoryDraftHistory(a.history, a.job, output, now, a.historyCap)
	draft, err := BuildMemoryDraft(memCtx, a.client, a.model, draftHistory, a.job.Text, output, existingContent, ctxInfo)
	if err != nil {
		return memoryruntime.RecordRequest{}, err
	}
	draft.Promote = EnforceLongTermPromotionRules(draft.Promote, nil, a.job.Text)

	return memoryruntime.RecordRequest{
		TaskRunID:    taskRunID,
		SessionID:    telegramMemorySessionID(a.job),
		SubjectID:    telegramMemorySubjectID(a.job),
		Channel:      "telegram",
		Participants: telegramMemoryParticipants(a.job),
		TaskText:     strings.TrimSpace(a.job.Text),
		FinalOutput:  output,
		Draft:        draft,
	}, nil
}

func buildMemoryWriteMeta(job telegramJob) memory.WriteMeta {
	meta := memory.WriteMeta{SessionID: telegramMemorySessionID(job)}
	if contactID := telegramMemoryContactID(job.FromUsername, job.FromUserID); contactID != "" {
		meta.ContactIDs = []string{contactID}
	}
	contactNickname := strings.TrimSpace(job.FromDisplayName)
	if contactNickname == "" {
		contactNickname = strings.TrimSpace(strings.Join([]string{job.FromFirstName, job.FromLastName}, " "))
	}
	if contactNickname != "" {
		meta.ContactNicknames = []string{contactNickname}
	}
	return meta
}

func telegramMemorySessionID(job telegramJob) string {
	return fmt.Sprintf("tg:%d", job.ChatID)
}

func telegramMemorySubjectID(job telegramJob) string {
	return telegramMemorySessionID(job)
}

func telegramMemoryRequestContext(chatType string) memory.RequestContext {
	if strings.EqualFold(strings.TrimSpace(chatType), "private") {
		return memory.ContextPrivate
	}
	return memory.ContextPublic
}

func telegramMemoryParticipants(job telegramJob) []memory.MemoryParticipant {
	seen := make(map[string]bool, 1+len(job.MentionUsers))
	out := make([]memory.MemoryParticipant, 0, 1+len(job.MentionUsers))
	appendParticipant := func(id string, nickname string) {
		id = strings.TrimSpace(id)
		nickname = strings.TrimSpace(nickname)
		if id == "" || nickname == "" {
			return
		}
		if seen[id] {
			return
		}
		seen[id] = true
		out = append(out, memory.MemoryParticipant{
			ID:       id,
			Nickname: nickname,
			Protocol: "tg",
		})
	}

	if senderID := telegramSenderParticipantID(job.FromUsername, job.FromUserID); senderID != "" {
		senderName := strings.TrimSpace(job.FromDisplayName)
		if senderName == "" {
			senderName = strings.TrimSpace(strings.Join([]string{job.FromFirstName, job.FromLastName}, " "))
		}
		if senderName == "" {
			senderName = senderID
		}
		appendParticipant(senderID, senderName)
	}
	for _, raw := range job.MentionUsers {
		mentionID := normalizeTelegramMentionID(raw)
		if mentionID == "" {
			continue
		}
		appendParticipant(mentionID, mentionID)
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func telegramSenderParticipantID(username string, userID int64) string {
	username = normalizeTelegramMentionID(username)
	if username != "" {
		return username
	}
	if userID > 0 {
		return strconv.FormatInt(userID, 10)
	}
	return ""
}

func normalizeTelegramMentionID(raw string) string {
	id := strings.TrimSpace(raw)
	id = strings.TrimPrefix(id, "@")
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return "@" + id
}

func buildMemoryDraftHistory(history []chathistory.ChatHistoryItem, job telegramJob, output string, now time.Time, maxItems int) []chathistory.ChatHistoryItem {
	out := append([]chathistory.ChatHistoryItem{}, history...)
	out = append(out, newTelegramInboundHistoryItem(job))
	if strings.TrimSpace(output) != "" {
		out = append(out, newTelegramOutboundAgentHistoryItem(job.ChatID, job.ChatType, output, now, ""))
	}
	if maxItems > 0 && len(out) > maxItems {
		out = out[len(out)-maxItems:]
	}
	return out
}

type MemoryDraftContext struct {
	SessionID          string `json:"session_id,omitempty"`
	ChatID             int64  `json:"chat_id,omitempty"`
	ChatType           string `json:"chat_type,omitempty"`
	CounterpartyID     int64  `json:"counterparty_id,omitempty"`
	CounterpartyName   string `json:"counterparty_name,omitempty"`
	CounterpartyHandle string `json:"counterparty_handle,omitempty"`
	CounterpartyLabel  string `json:"counterparty_label,omitempty"`
	TimestampUTC       string `json:"timestamp_utc,omitempty"`
}

func buildMemoryCounterpartyLabel(meta memory.WriteMeta, ctxInfo MemoryDraftContext) string {
	contactID := firstNonEmptyString(meta.ContactIDs...)
	if contactID == "" {
		handle := strings.TrimPrefix(strings.TrimSpace(ctxInfo.CounterpartyHandle), "@")
		if handle != "" {
			contactID = "tg:@" + handle
		}
	}
	nickname := firstNonEmptyString(meta.ContactNicknames...)
	if nickname == "" {
		nickname = strings.TrimSpace(ctxInfo.CounterpartyName)
	}
	if nickname == "" {
		nickname = strings.TrimPrefix(strings.TrimPrefix(contactID, "tg:@"), "tg:")
	}
	nickname = sanitizeTelegramReferenceLabel(nickname)
	if nickname != "" && contactID != "" {
		return "[" + nickname + "](" + contactID + ")"
	}
	if nickname != "" {
		return nickname
	}
	return strings.TrimSpace(contactID)
}

func firstNonEmptyString(values ...string) string {
	for _, raw := range values {
		v := strings.TrimSpace(raw)
		if v != "" {
			return v
		}
	}
	return ""
}

func BuildMemoryDraft(ctx context.Context, client llm.Client, model string, history []chathistory.ChatHistoryItem, task string, output string, existing memory.ShortTermContent, ctxInfo MemoryDraftContext) (memory.SessionDraft, error) {
	if client == nil {
		return memory.SessionDraft{}, fmt.Errorf("nil llm client")
	}

	sys, user, err := renderMemoryDraftPrompts(ctxInfo, history, task, output, existing)
	if err != nil {
		return memory.SessionDraft{}, fmt.Errorf("render memory draft prompts: %w", err)
	}

	res, err := client.Chat(llminspect.WithModelScene(ctx, "memory.draft"), llm.Request{
		Model:     model,
		ForceJSON: true,
		Messages: []llm.Message{
			{Role: "system", Content: sys},
			{Role: "user", Content: user},
		},
		Parameters: map[string]any{
			"max_tokens": 10240,
		},
	})
	if err != nil {
		return memory.SessionDraft{}, err
	}

	raw := strings.TrimSpace(res.Text)
	if raw == "" {
		return memory.SessionDraft{}, fmt.Errorf("empty memory draft response")
	}

	var out memory.SessionDraft
	if err := jsonutil.DecodeWithFallback(raw, &out); err != nil {
		return memory.SessionDraft{}, fmt.Errorf("invalid memory draft json")
	}
	out.SummaryItems = normalizeMemorySummaryItems(out.SummaryItems)
	return out, nil
}

func EnforceLongTermPromotionRules(promote memory.PromoteDraft, history []llm.Message, task string) memory.PromoteDraft {
	if !hasExplicitMemoryRequest(history, task) {
		return memory.PromoteDraft{}
	}
	return limitPromoteToOne(promote)
}

func hasExplicitMemoryRequest(history []llm.Message, task string) bool {
	texts := make([]string, 0, len(history)+1)
	for _, m := range history {
		if strings.ToLower(strings.TrimSpace(m.Role)) != "user" {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		texts = append(texts, content)
	}
	if strings.TrimSpace(task) != "" {
		texts = append(texts, task)
	}
	if len(texts) == 0 {
		return false
	}
	combined := strings.ToLower(strings.Join(texts, "\n"))
	return containsExplicitMemoryRequest(combined)
}

func containsExplicitMemoryRequest(lowerText string) bool {
	if strings.TrimSpace(lowerText) == "" {
		return false
	}
	keywords := []string{
		"记住",
		"记下来",
		"别忘",
		"记得",
		"长期记忆",
		"写入长期记忆",
		"加入长期记忆",
		"记到长期",
		"remember",
		"don't forget",
		"dont forget",
		"long-term memory",
		"add to memory",
		"save this",
		"keep this",
		"store this",
		"memorize",
	}
	for _, k := range keywords {
		if strings.Contains(lowerText, k) {
			return true
		}
	}
	return false
}

func limitPromoteToOne(promote memory.PromoteDraft) memory.PromoteDraft {
	if item, ok := firstNonEmptyText(promote.GoalsProjects); ok {
		return memory.PromoteDraft{GoalsProjects: []string{item}}
	}
	if item, ok := firstKVItem(promote.KeyFacts); ok {
		return memory.PromoteDraft{KeyFacts: []memory.KVItem{item}}
	}
	return memory.PromoteDraft{}
}

func firstNonEmptyText(items []string) (string, bool) {
	for _, raw := range items {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		return v, true
	}
	return "", false
}

func firstKVItem(items []memory.KVItem) (memory.KVItem, bool) {
	for _, it := range items {
		title := strings.TrimSpace(it.Title)
		value := strings.TrimSpace(it.Value)
		if title == "" && value == "" {
			continue
		}
		it.Title = title
		it.Value = value
		return it, true
	}
	return memory.KVItem{}, false
}

func normalizeMemorySummaryItems(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	out := make([]string, 0, len(input))
	seen := make(map[string]bool, len(input))
	for _, raw := range input {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
