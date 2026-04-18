package llmutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/quailyquaily/mistermorph/internal/acpclient"
	"github.com/quailyquaily/mistermorph/llm"
)

// acpObserverContextKey is used to store an acpclient.Observer in context.
type acpObserverContextKey struct{}

// WithACPObserver attaches an acpclient.Observer to the context.
// When an acpLLMClient.Chat call uses this context, the observer will receive
// real-time ACP events (tool start / update / done) instead of the default
// stderr printer.  This allows channels (console, telegram, slack) to surface
// gemini_oauth progress in their own UI.
func WithACPObserver(ctx context.Context, observer acpclient.Observer) context.Context {
	return context.WithValue(ctx, acpObserverContextKey{}, observer)
}

// ACPObserverFromContext retrieves the acpclient.Observer stored in context.
func ACPObserverFromContext(ctx context.Context) (acpclient.Observer, bool) {
	observer, ok := ctx.Value(acpObserverContextKey{}).(acpclient.Observer)
	return observer, ok
}

// acpLLMClient adapts acpclient.RunPrompt to the llm.Client interface.
type acpLLMClient struct {
	cfg acpclient.PreparedAgentConfig
}

// newACPLLMClient creates a new LLM client that uses ACP protocol via stdio.
func newACPLLMClient(cfg acpclient.PreparedAgentConfig) *acpLLMClient {
	return &acpLLMClient{cfg: cfg}
}

// acpProgressObserver prints ACP tool-call events to stderr so the user can
// see what the external agent is doing in real time.  It intentionally skips
// agent-message chunks (the final answer is rendered by the caller) and only
// surfaces tool start / update / done events.
type acpProgressObserver struct {
	mu         sync.Mutex
	lastToolID string
}

func (o *acpProgressObserver) HandleACPEvent(_ context.Context, event acpclient.Event) {
	switch event.Kind {
	case acpclient.EventKindToolCallStart:
		o.mu.Lock()
		o.lastToolID = event.ToolCallID
		o.mu.Unlock()
		title := event.Title
		if title == "" {
			title = event.ToolKind
		}
		fmt.Fprintf(os.Stderr, "\r\033[K\033[90m  [gemini] 🔧 %s\033[0m\n", title)
	case acpclient.EventKindToolCallUpdate:
		if text := strings.TrimSpace(event.Text); text != "" {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					fmt.Fprintf(os.Stderr, "\r\033[K\033[90m    → %s\033[0m\n", truncate(line, 120))
				}
			}
		}
	case acpclient.EventKindToolCallDone:
		status := event.Status
		if status == "" {
			status = "done"
		}
		icon := "✓"
		if status == "failed" {
			icon = "✗"
		}
		fmt.Fprintf(os.Stderr, "\r\033[K\033[90m  [gemini] %s %s\033[0m\n", icon, status)
	}
}

func defaultACPObserver() acpclient.Observer {
	return &acpProgressObserver{}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Chat implements llm.Client by running an ACP prompt session.
func (c *acpLLMClient) Chat(ctx context.Context, req llm.Request) (llm.Result, error) {
	start := time.Now()

	prompt, err := buildACPPrompt(req)
	if err != nil {
		return llm.Result{}, err
	}

	observer := defaultACPObserver()
	if obs, ok := ACPObserverFromContext(ctx); ok && obs != nil {
		observer = obs
	}

	result, err := acpclient.RunPrompt(ctx, c.cfg, acpclient.RunRequest{
		Prompt:   prompt,
		Observer: observer,
	})
	if err != nil {
		return llm.Result{}, fmt.Errorf("acp prompt failed: %w", err)
	}

	output := strings.TrimSpace(result.Output)

	// Try to parse as JSON final answer; fallback to raw text
	return parseACPOutput(output, time.Since(start)), nil
}

func buildACPPrompt(req llm.Request) (string, error) {
	var parts []string

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			if strings.TrimSpace(msg.Content) != "" {
				parts = append(parts, fmt.Sprintf("[System]\n%s", msg.Content))
			}
		case "user":
			if strings.TrimSpace(msg.Content) != "" {
				parts = append(parts, fmt.Sprintf("[User]\n%s", msg.Content))
			}
		case "assistant":
			if strings.TrimSpace(msg.Content) != "" {
				parts = append(parts, fmt.Sprintf("[Assistant]\n%s", msg.Content))
			}
			for _, call := range msg.ToolCalls {
				parts = append(parts, fmt.Sprintf("[Tool Call] %s: %s", call.Name, call.RawArguments))
			}
		case "tool":
			parts = append(parts, fmt.Sprintf("[Tool Result]\n%s", msg.Content))
		default:
			return "", fmt.Errorf("acp_llm: unsupported message role %q", msg.Role)
		}
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("acp_llm: no messages to send")
	}

	return strings.Join(parts, "\n\n"), nil
}

func parseACPOutput(output string, duration time.Duration) llm.Result {
	// Try to parse as a JSON final answer
	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err == nil {
		if typ, ok := payload["type"].(string); ok && typ == "final" {
			if out, ok := payload["output"].(string); ok {
				return llm.Result{
					Text:     output,
					JSON:     payload,
					Usage:    llm.Usage{},
					Duration: duration,
					Parts:    []llm.Part{{Type: "text", Text: out}},
				}
			}
		}
	}

	// Fallback: wrap raw output as final answer
	payload = map[string]any{
		"type":   "final",
		"output": output,
	}
	jsonBytes, _ := json.Marshal(payload)

	return llm.Result{
		Text:     string(jsonBytes),
		JSON:     payload,
		Usage:    llm.Usage{},
		Duration: duration,
		Parts:    []llm.Part{{Type: "text", Text: output}},
	}
}

// newGeminiOAuthACPLLMClient creates an ACP LLM client for gemini_oauth provider.
func newGeminiOAuthACPLLMClient(model string) (llm.Client, error) {
	if _, err := exec.LookPath("gemini"); err != nil {
		return nil, fmt.Errorf("gemini CLI not found in PATH: %w", err)
	}

	cfg := acpclient.PreparedAgentConfig{
		Name:    "gemini_oauth",
		Command: "gemini",
		Args:    []string{"--acp"},
	}

	model = strings.TrimSpace(model)
	if model != "" {
		cfg.SessionOptionsMeta = map[string]any{
			"model": model,
		}
	}

	return newACPLLMClient(cfg), nil
}
