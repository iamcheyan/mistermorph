package memoryruntime

import (
	"context"
	"fmt"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/jsonutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/memory"
)

type DraftRequest struct {
	Client         llm.Client
	Model          string
	History        []chathistory.ChatHistoryItem
	Task           string
	Output         string
	Existing       memory.ShortTermContent
	SessionContext memory.SessionContext
}

func BuildLLMDraft(ctx context.Context, req DraftRequest) (memory.SessionDraft, error) {
	if req.Client == nil {
		return memory.SessionDraft{}, fmt.Errorf("nil llm client")
	}

	sys, user, err := renderMemoryDraftPrompts(req.SessionContext, req.History, req.Task, req.Output, req.Existing)
	if err != nil {
		return memory.SessionDraft{}, fmt.Errorf("render memory draft prompts: %w", err)
	}

	res, err := req.Client.Chat(ctx, llm.Request{
		Model:     strings.TrimSpace(req.Model),
		Scene:     "memory.draft",
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
	out.SummaryItems = normalizeSummaryItems(out.SummaryItems)
	return out, nil
}

func normalizeSummaryItems(items []string) []string {
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
