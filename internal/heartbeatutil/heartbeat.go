package heartbeatutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/quailyquaily/mistermorph/internal/entryutil"
	"github.com/quailyquaily/mistermorph/internal/todo"
)

const (
	heartbeatFailureThreshold = 3
)

var heartbeatHTMLComment = regexp.MustCompile(`(?s)<!--.*?-->`)

func BuildHeartbeatTask(checklistPath string) (string, bool, error) {
	if err := materializeDueRecurringTodos(checklistPath); err != nil {
		return "", true, err
	}
	todoBlock, err := readOpenTodosBlock(checklistPath)
	if err != nil {
		return "", true, err
	}
	checklist, checklistEmpty, err := readHeartbeatChecklist(checklistPath)
	if err != nil {
		return "", true, err
	}
	var sections []string
	if todoBlock != "" {
		sections = append(sections, todoBlock)
	}
	if checklist != "" {
		sections = append(sections, checklist)
	}
	return strings.Join(sections, "\n\n"), checklistEmpty, nil
}

func readOpenTodosBlock(checklistPath string) (string, error) {
	todoPath := todoWIPPathForChecklist(checklistPath)
	if todoPath == "" {
		return "", nil
	}
	raw, err := os.ReadFile(todoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	wip, err := todo.ParseWIP(string(raw))
	if err != nil {
		return "", err
	}
	if len(wip.Entries) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("## Current TODO.md Open Items\n\n")
	for _, item := range wip.Entries {
		line := heartbeatTODOEntryLine(item)
		if line == "" {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String()), nil
}

func heartbeatTODOEntryLine(item todo.Entry) string {
	content := strings.TrimSpace(item.Content)
	createdAt := strings.TrimSpace(item.CreatedAt)
	if content == "" || !entryutil.IsValidTimestamp(createdAt) {
		return ""
	}
	meta := []string{entryutil.FormatMetadataTuple("Created", createdAt)}
	if chatID := strings.TrimSpace(item.ChatID); chatID != "" {
		meta = append(meta, entryutil.FormatMetadataTuple("ChatID", chatID))
	}
	return "- [ ] " + strings.Join(meta, ", ") + " | " + content
}

func materializeDueRecurringTodos(checklistPath string) error {
	checklistPath = strings.TrimSpace(checklistPath)
	todoPath := todoWIPPathForChecklist(checklistPath)
	if todoPath == "" {
		return nil
	}
	store := todo.NewStore(
		todoPath,
		filepath.Join(filepath.Dir(todoPath), todo.DefaultDONEFilename),
	)
	store.RecurringPath = filepath.Join(filepath.Dir(todoPath), todo.DefaultRECURFilename)
	_, err := store.MaterializeDueRecurring()
	return err
}

func todoWIPPathForChecklist(checklistPath string) string {
	checklistPath = strings.TrimSpace(checklistPath)
	if checklistPath == "" {
		return ""
	}
	stateDir := filepath.Dir(checklistPath)
	if strings.TrimSpace(stateDir) == "" || stateDir == "." {
		return ""
	}
	return filepath.Join(stateDir, todo.DefaultWIPFilename)
}

func BuildHeartbeatMeta(source string, interval time.Duration, checklistPath string, checklistEmpty bool, state *State, extra map[string]any) map[string]any {
	hb := map[string]any{
		"source":           source,
		"scheduled_at_utc": time.Now().UTC().Format(time.RFC3339),
		"interval":         interval.String(),
	}
	if strings.TrimSpace(checklistPath) != "" {
		hb["checklist_path"] = checklistPath
	}
	if checklistEmpty {
		hb["checklist_empty"] = true
	}
	if state != nil {
		failures, lastSuccess, lastError, _ := state.Snapshot()
		if failures > 0 {
			hb["failures"] = failures
		}
		if !lastSuccess.IsZero() {
			hb["last_success_utc"] = lastSuccess.UTC().Format(time.RFC3339)
		}
		if strings.TrimSpace(lastError) != "" {
			hb["last_error"] = lastError
		}
	}
	for k, v := range extra {
		if strings.TrimSpace(k) == "" {
			continue
		}
		hb[k] = v
	}
	return map[string]any{
		"trigger":   "heartbeat",
		"heartbeat": hb,
	}
}

type State struct {
	mu          sync.Mutex
	running     bool
	failures    int
	lastSuccess time.Time
	lastError   string
}

func (s *State) Start() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	return true
}

func (s *State) EndSkipped() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

func (s *State) EndSuccess(now time.Time) {
	s.mu.Lock()
	s.running = false
	s.failures = 0
	s.lastError = ""
	s.lastSuccess = now
	s.mu.Unlock()
}

func (s *State) EndFailure(err error) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.failures++
	if err != nil {
		s.lastError = strings.TrimSpace(err.Error())
	}
	if s.failures >= heartbeatFailureThreshold {
		msg := "heartbeat_failed"
		if s.lastError != "" {
			msg = fmt.Sprintf("heartbeat_failed (%s)", s.lastError)
		}
		s.failures = 0
		return true, "ALERT: " + msg
	}
	return false, ""
}

func (s *State) Snapshot() (failures int, lastSuccess time.Time, lastError string, running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failures, s.lastSuccess, s.lastError, s.running
}

func readHeartbeatChecklist(path string) (string, bool, error) {
	path = strings.TrimSpace(path)
	if strings.TrimSpace(path) == "" {
		return "", true, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", true, nil
		}
		return "", true, err
	}
	content := string(raw)
	if isChecklistEmptyContent(content) {
		return "", true, nil
	}
	return strings.TrimSpace(content), false, nil
}

func isChecklistEmptyContent(content string) bool {
	stripped := heartbeatHTMLComment.ReplaceAllString(content, "")
	lines := strings.Split(stripped, "\n")
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "#") {
			continue
		}
		return false
	}
	return true
}
