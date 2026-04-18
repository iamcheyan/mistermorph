package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/llm"
)

// Config holds configuration for the Gemini CLI wrapper provider.
type Config struct {
	DefaultModel string
	Debug        bool
}

// Client implements the LLM provider interface by wrapping the gemini CLI.
type Client struct {
	cfg Config
}

// New creates a new Gemini CLI wrapper client.
func New(cfg Config) (*Client, error) {
	// Verify gemini CLI is available
	if _, err := exec.LookPath("gemini"); err != nil {
		return nil, fmt.Errorf("gemini CLI not found in PATH: %w", err)
	}
	return &Client{cfg: cfg}, nil
}

// Chat sends a chat request via the gemini CLI.
func (c *Client) Chat(ctx context.Context, req llm.Request) (llm.Result, error) {
	start := time.Now()

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = strings.TrimSpace(c.cfg.DefaultModel)
	}
	if model == "" {
		return llm.Result{}, fmt.Errorf("gemini_cli: model is required")
	}

	// Build the prompt from messages
	prompt, err := c.buildPrompt(req)
	if err != nil {
		return llm.Result{}, err
	}

	// Build gemini CLI arguments
	// --yolo: auto-approve all tool calls, letting Gemini CLI handle everything
	args := []string{
		"-p", prompt,
		"--model", model,
		"--yolo",
		"--output-format", "json",
	}

	if c.cfg.Debug && req.DebugFn != nil {
		req.DebugFn("gemini_cli.exec", fmt.Sprintf("gemini %s", strings.Join(args, " ")))
	}

	cmd := exec.CommandContext(ctx, "gemini", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return llm.Result{}, ctx.Err()
		}
		return llm.Result{}, fmt.Errorf("gemini CLI failed: %w\nstderr: %s", err, stderr.String())
	}

	if c.cfg.Debug && req.DebugFn != nil {
		req.DebugFn("gemini_cli.stdout", stdout.String())
		if stderr.Len() > 0 {
			req.DebugFn("gemini_cli.stderr", stderr.String())
		}
	}

	result, err := c.parseOutput(stdout.Bytes(), model)
	if err != nil {
		return llm.Result{}, err
	}
	result.Duration = time.Since(start)
	return result, nil
}

// buildPrompt concatenates all messages into a single prompt string.
func (c *Client) buildPrompt(req llm.Request) (string, error) {
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
			return "", fmt.Errorf("gemini_cli: unsupported message role %q", msg.Role)
		}
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("gemini_cli: no messages to send")
	}

	return strings.Join(parts, "\n\n"), nil
}

// geminiCLIOutput represents the JSON output from `gemini --output-format json`.
type geminiCLIOutput struct {
	Response string `json:"response"`
	SessionID string `json:"session_id"`
	Stats    struct {
		Models map[string]struct {
			Tokens struct {
				Input      int `json:"input"`
				Prompt     int `json:"prompt"`
				Candidates int `json:"candidates"`
				Total      int `json:"total"`
				Thoughts   int `json:"thoughts"`
			} `json:"tokens"`
		} `json:"models"`
	} `json:"stats"`
}

func (c *Client) parseOutput(data []byte, model string) (llm.Result, error) {
	var out geminiCLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		// If not valid JSON, wrap the raw output as a final answer
		return c.wrapAsFinal(strings.TrimSpace(string(data)), llm.Usage{}), nil
	}

	// Extract token usage for the requested model
	var usage llm.Usage
	if modelStats, ok := out.Stats.Models[model]; ok {
		usage = llm.Usage{
			InputTokens:  modelStats.Tokens.Input,
			OutputTokens: modelStats.Tokens.Candidates,
			TotalTokens:  modelStats.Tokens.Total,
		}
	} else {
		// Try to find any model's stats
		for _, modelStats := range out.Stats.Models {
			usage = llm.Usage{
				InputTokens:  modelStats.Tokens.Input,
				OutputTokens: modelStats.Tokens.Candidates,
				TotalTokens:  modelStats.Tokens.Total,
			}
			break
		}
	}

	// Wrap Gemini CLI's response as a final answer in the format expected by the agent loop.
	// This allows the agent to treat the CLI's fully-executed result as a single-step final output.
	return c.wrapAsFinal(strings.TrimSpace(out.Response), usage), nil
}

// wrapAsFinal formats the given text as a JSON final answer that the agent loop recognizes.
func (c *Client) wrapAsFinal(output string, usage llm.Usage) llm.Result {
	payload := map[string]any{
		"type":   "final",
		"output": output,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		// Fallback: return raw text if JSON marshaling fails
		return llm.Result{Text: output, Usage: usage}
	}

	return llm.Result{
		Text:  string(jsonBytes),
		JSON:  payload,
		Usage: usage,
		Parts: []llm.Part{
			{Type: "text", Text: output},
		},
	}
}
