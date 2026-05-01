package builtin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	refid "github.com/quailyquaily/mistermorph/internal/entryutil/refid"
	"github.com/quailyquaily/mistermorph/internal/pathutil"
	"github.com/quailyquaily/mistermorph/internal/todo"
	"github.com/quailyquaily/mistermorph/llm"
)

type TodoUpdateTool struct {
	Enabled    bool
	WIPPath    string
	DONEPath   string
	Contacts   string
	Client     llm.Client
	Model      string
	AddContext todo.AddResolveContext
}

func NewTodoUpdateTool(enabled bool, wipPath string, donePath string, contactsDir string) *TodoUpdateTool {
	return NewTodoUpdateToolWithLLM(enabled, wipPath, donePath, contactsDir, nil, "")
}

func NewTodoUpdateToolWithLLM(enabled bool, wipPath string, donePath string, contactsDir string, client llm.Client, model string) *TodoUpdateTool {
	return &TodoUpdateTool{
		Enabled:  enabled,
		WIPPath:  strings.TrimSpace(wipPath),
		DONEPath: strings.TrimSpace(donePath),
		Contacts: strings.TrimSpace(contactsDir),
		Client:   client,
		Model:    strings.TrimSpace(model),
	}
}

func (t *TodoUpdateTool) BindLLM(client llm.Client, model string) {
	if t == nil {
		return
	}
	t.Client = client
	t.Model = strings.TrimSpace(model)
}

func (t *TodoUpdateTool) Clone() *TodoUpdateTool {
	if t == nil {
		return nil
	}
	cp := *t
	cp.AddContext.MentionUsernames = append([]string(nil), t.AddContext.MentionUsernames...)
	return &cp
}

func (t *TodoUpdateTool) SetAddContext(ctx todo.AddResolveContext) {
	if t == nil {
		return
	}
	ctx.Channel = strings.ToLower(strings.TrimSpace(ctx.Channel))
	ctx.ChatType = strings.ToLower(strings.TrimSpace(ctx.ChatType))
	ctx.SpeakerUsername = strings.TrimPrefix(strings.TrimSpace(ctx.SpeakerUsername), "@")
	ctx.MentionUsernames = normalizeTodoUpdateUsernames(ctx.MentionUsernames)
	ctx.UserInputRaw = strings.TrimSpace(ctx.UserInputRaw)
	t.AddContext = ctx
}

func (t *TodoUpdateTool) Name() string { return "todo_update" }

func (t *TodoUpdateTool) Description() string {
	return "Updates TODO files under file_state_dir. Supports add, complete, and add_recurring actions, keeps counts in TODO.md/TODO.DONE.md/TODO.RECUR.md consistent."
}

func (t *TodoUpdateTool) ParameterSchema() string {
	s := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action: add|complete|add_recurring.",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Todo content. Required for add, complete, and add_recurring.",
			},
			"next": map[string]any{
				"type":        "string",
				"description": "Next scheduled time for add_recurring, in YYYY-MM-DD HH:mm.",
			},
			"repeat": map[string]any{
				"type":        "string",
				"description": "Repeat rule for add_recurring: daily|weekly|every N days|every N hours.",
			},
			"tz": map[string]any{
				"type":        "string",
				"description": "Optional IANA timezone for add_recurring, for example Asia/Tokyo. Omit to use runtime local timezone.",
			},
			"people": map[string]any{
				"type": "array",
				"description": "List of people mentioned in the content (required for add). " +
					"If the speaker mentions theirselve (said `I` or `me`), resolve as '$SPEAKER' in the array." +
					"If the speaker mentions `you`, resolve as '$AGENT' in the array. " +
					"For others, put their nickname or an ID in the arrary.",
				"items": map[string]any{
					"type": "string",
				},
			},
			"chat_id": map[string]any{
				"type":        "string",
				"description": "Optional task context chat id (for example tg:-1001234567890).",
			},
		},
		"required": []string{"action", "content"},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

func (t *TodoUpdateTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	if !t.Enabled {
		return "", fmt.Errorf("todo_update tool is disabled")
	}
	action, _ := params["action"].(string)
	action = strings.ToLower(strings.TrimSpace(action))
	content, _ := params["content"].(string)
	content = strings.TrimSpace(content)
	if action == "" {
		return "", fmt.Errorf("action is required")
	}
	if content == "" {
		return "", fmt.Errorf("content is required")
	}
	wipPath := pathutil.ExpandHomePath(strings.TrimSpace(t.WIPPath))
	donePath := pathutil.ExpandHomePath(strings.TrimSpace(t.DONEPath))
	contactsDir := pathutil.ExpandHomePath(strings.TrimSpace(t.Contacts))
	if wipPath == "" || donePath == "" {
		return "", fmt.Errorf("todo paths are not configured")
	}
	if contactsDir == "" {
		return "", fmt.Errorf("contacts dir is not configured")
	}
	if t.Client == nil {
		return "", fmt.Errorf("todo_update unavailable (missing llm client)")
	}
	if strings.TrimSpace(t.Model) == "" {
		return "", fmt.Errorf("todo_update unavailable (missing llm model)")
	}

	store := todo.NewStore(wipPath, donePath)
	store.Semantics = todo.NewLLMSemanticResolver(t.Client, t.Model)
	var (
		result any
		err    error
	)
	switch action {
	case "add":
		chatID, chatIDErr := parseTodoUpdateChatID(params)
		if chatIDErr != nil {
			return "", chatIDErr
		}
		people, peopleErr := parseTodoUpdatePeople(params)
		if peopleErr != nil {
			return "", peopleErr
		}
		slog.Default().Debug("todo_update_add_start",
			"content_len", len(content),
			"chat_id", chatID,
			"people_count", len(people),
			"context_channel", t.AddContext.Channel,
			"context_chat_type", t.AddContext.ChatType,
			"context_chat_id", t.AddContext.ChatID,
			"context_speaker_user_id", t.AddContext.SpeakerUserID,
			"context_speaker_username", t.AddContext.SpeakerUsername,
			"context_mentions_count", len(t.AddContext.MentionUsernames),
			"context_user_input_raw_len", len(t.AddContext.UserInputRaw),
		)
		rewritten, warnings, resolveErr := t.resolveAddContent(ctx, content, people, contactsDir, true)
		if resolveErr != nil {
			return "", resolveErr
		}
		result, err = store.AddWithChatID(ctx, rewritten, chatID)
		if err == nil && len(warnings) > 0 {
			addResult := result.(todo.UpdateResult)
			addResult.Warnings = append(addResult.Warnings, warnings...)
			result = addResult
		}
	case "add_recurring":
		chatID, chatIDErr := parseTodoUpdateChatID(params)
		if chatIDErr != nil {
			return "", chatIDErr
		}
		nextAt, nextErr := parseTodoUpdateString(params, "next", "next_at")
		if nextErr != nil {
			return "", nextErr
		}
		repeat, repeatErr := parseTodoUpdateString(params, "repeat")
		if repeatErr != nil {
			return "", repeatErr
		}
		tz, tzErr := parseTodoUpdateOptionalString(params, "tz")
		if tzErr != nil {
			return "", tzErr
		}
		people, peopleErr := parseTodoUpdatePeopleOptional(params)
		if peopleErr != nil {
			return "", peopleErr
		}
		rewritten, warnings, resolveErr := t.resolveAddContent(ctx, content, people, contactsDir, false)
		if resolveErr != nil {
			return "", resolveErr
		}
		result, err = store.AddRecurringWithChatID(rewritten, nextAt, repeat, tz, chatID)
		if err == nil && len(warnings) > 0 {
			recurringResult := result.(todo.RecurringUpdateResult)
			recurringResult.Warnings = append(recurringResult.Warnings, warnings...)
			result = recurringResult
		}
	case "complete":
		result, err = store.Complete(ctx, content)
	default:
		return "", fmt.Errorf("invalid action: %s", action)
	}
	if err != nil {
		return "", err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *TodoUpdateTool) resolveAddContent(ctx context.Context, content string, people []string, contactsDir string, enforceRequiredReferences bool) (string, []string, error) {
	if _, preErr := todo.ExtractReferenceIDs(content); preErr != nil {
		return "", nil, preErr
	}
	if len(people) == 0 {
		return t.resolveAddPlaceholders(content, content, nil)
	}
	snapshot, snapErr := todo.LoadContactSnapshot(ctx, contactsDir)
	if snapErr != nil {
		return "", nil, snapErr
	}
	resolver := todo.NewLLMReferenceResolver(t.Client, t.Model)
	rewritten, warnings, resolveErr := resolver.ResolveAddContent(ctx, content, people, snapshot, t.AddContext)
	fallbackRawWrite := false
	if resolveErr != nil {
		var missingErr *todo.MissingReferenceIDError
		if errors.As(resolveErr, &missingErr) {
			fallbackRawWrite = true
			rewritten = content
			warnings = appendIfMissingWarning(warnings, "reference_unresolved_write_raw")
			slog.Default().Debug("todo_update_add_reference_unresolved_fallback",
				"missing_count", len(missingErr.Items),
			)
		} else {
			return "", nil, resolveErr
		}
	}
	slog.Default().Debug("todo_update_add_resolved",
		"rewritten", rewritten,
		"warnings_count", len(warnings),
		"fallback_raw_write", fallbackRawWrite,
	)
	if enforceRequiredReferences && !fallbackRawWrite {
		if requiredErr := todo.ValidateRequiredReferenceMentions(rewritten, snapshot); requiredErr != nil {
			var missingErr *todo.MissingReferenceIDError
			if errors.As(requiredErr, &missingErr) {
				firstMention := ""
				firstSuggestion := ""
				firstReason := ""
				if len(missingErr.Items) > 0 {
					firstMention = strings.TrimSpace(missingErr.Items[0].Mention)
					firstSuggestion = strings.TrimSpace(missingErr.Items[0].Suggestion)
					firstReason = strings.TrimSpace(missingErr.Items[0].Reason)
				}
				slog.Default().Debug("todo_update_add_required_reference_fallback_detail",
					"rewritten_before_fallback", rewritten,
					"fallback_target_content", content,
					"first_missing_mention", firstMention,
					"first_missing_suggestion", firstSuggestion,
					"first_missing_reason", firstReason,
				)
				rewritten = content
				warnings = appendIfMissingWarning(warnings, "reference_unresolved_write_raw")
				slog.Default().Debug("todo_update_add_required_reference_fallback",
					"missing_count", len(missingErr.Items),
				)
			} else {
				return "", nil, requiredErr
			}
		}
	}
	rewritten, warnings, placeholderErr := t.resolveAddPlaceholders(content, rewritten, warnings)
	if placeholderErr != nil {
		return "", nil, placeholderErr
	}
	return rewritten, warnings, nil
}

func (t *TodoUpdateTool) resolveAddPlaceholders(original string, rewritten string, warnings []string) (string, []string, error) {
	rewritten = strings.TrimSpace(rewritten)
	if rewritten == "" {
		return "", nil, fmt.Errorf("content is required")
	}
	if strings.Contains(rewritten, "$SPEAKER") {
		speakerRef := todoUpdateSpeakerReferenceID(t.AddContext)
		if speakerRef != "" {
			label := todoUpdateSpeakerReferenceLabel(rewritten)
			formatted, err := refid.FormatMarkdownReference(label, speakerRef)
			if err != nil {
				return "", nil, err
			}
			rewritten = strings.ReplaceAll(rewritten, "$SPEAKER", formatted)
		}
	}
	if containsTodoUpdateReferencePlaceholder(rewritten) {
		original = strings.TrimSpace(original)
		if containsTodoUpdateReferencePlaceholder(original) {
			return "", nil, fmt.Errorf("unresolved reference placeholder in content")
		}
		warnings = appendIfMissingWarning(warnings, "reference_placeholder_unresolved_write_raw")
		return original, warnings, nil
	}
	return rewritten, warnings, nil
}

func todoUpdateSpeakerReferenceID(ctx todo.AddResolveContext) string {
	if ref, ok := refid.Normalize(ctx.SpeakerUsername); ok {
		return ref
	}
	switch strings.ToLower(strings.TrimSpace(ctx.Channel)) {
	case "console":
		return "console:user"
	case "telegram":
		if ctx.SpeakerUserID != 0 {
			return fmt.Sprintf("tg:%d", ctx.SpeakerUserID)
		}
		username := strings.TrimPrefix(strings.TrimSpace(ctx.SpeakerUsername), "@")
		if username != "" {
			return "tg:@" + username
		}
	}
	return ""
}

func todoUpdateSpeakerReferenceLabel(content string) string {
	for _, r := range content {
		if unicode.In(r, unicode.Han) {
			return "我"
		}
	}
	return "me"
}

func containsTodoUpdateReferencePlaceholder(content string) bool {
	return strings.Contains(content, "$SPEAKER") || strings.Contains(content, "$AGENT")
}

func normalizeTodoUpdateUsernames(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	out := make([]string, 0, len(input))
	seen := make(map[string]bool, len(input))
	for _, raw := range input {
		v := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "@"))
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
	return out
}

func parseTodoUpdatePeople(params map[string]any) ([]string, error) {
	raw, exists := params["people"]
	if !exists {
		return nil, fmt.Errorf("people is required for add action")
	}
	return parseTodoUpdatePeopleValue(raw)
}

func parseTodoUpdatePeopleOptional(params map[string]any) ([]string, error) {
	raw, exists := params["people"]
	if !exists || raw == nil {
		return nil, nil
	}
	return parseTodoUpdatePeopleValue(raw)
}

func parseTodoUpdatePeopleValue(raw any) ([]string, error) {
	switch v := raw.(type) {
	case []string:
		return normalizeTodoUpdateUsernames(v), nil
	case []any:
		items := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("people must be an array of strings")
			}
			items = append(items, s)
		}
		return normalizeTodoUpdateUsernames(items), nil
	default:
		return nil, fmt.Errorf("people must be an array of strings")
	}
}

func parseTodoUpdateString(params map[string]any, names ...string) (string, error) {
	value, err := parseTodoUpdateOptionalString(params, names...)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", strings.TrimSpace(names[0]))
	}
	return value, nil
}

func parseTodoUpdateOptionalString(params map[string]any, names ...string) (string, error) {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		raw, exists := params[name]
		if !exists || raw == nil {
			continue
		}
		value, ok := raw.(string)
		if !ok {
			return "", fmt.Errorf("%s must be a string", name)
		}
		return strings.TrimSpace(value), nil
	}
	return "", nil
}

func parseTodoUpdateChatID(params map[string]any) (string, error) {
	if raw, exists := params["chat_id"]; exists && raw != nil {
		value, ok := raw.(string)
		if !ok {
			return "", fmt.Errorf("chat_id must be a string")
		}
		return strings.TrimSpace(value), nil
	}
	// Backward compatibility for older callers.
	if raw, exists := params["channel"]; exists && raw != nil {
		value, ok := raw.(string)
		if !ok {
			return "", fmt.Errorf("channel must be a string")
		}
		return strings.TrimSpace(value), nil
	}
	return "", nil
}

func appendIfMissingWarning(warnings []string, v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return warnings
	}
	for _, item := range warnings {
		if strings.TrimSpace(item) == v {
			return warnings
		}
	}
	return append(warnings, v)
}
