package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quailyquaily/mistermorph/tools"
)

const spawnToolName = "spawn"

type spawnTool struct {
	engine *Engine
}

func (t *spawnTool) Name() string { return spawnToolName }

func (t *spawnTool) Description() string {
	return "Spawn a sub-agent to handle a self-contained sub-task. " +
		"The sub-agent runs with its own context and a restricted set of tools you specify. " +
		"This call blocks until the sub-agent completes and returns its final output. " +
		"Use this to parallelise independent work items."
}

func (t *spawnTool) ParameterSchema() string {
	s := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task": map[string]any{
				"type":        "string",
				"description": "Task prompt for the sub-agent.",
			},
			"tools": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Whitelist of tool names the sub-agent can use. Cannot include 'spawn'.",
			},
			"model": map[string]any{
				"type":        "string",
				"description": "Optional model override for the sub-agent. Defaults to the parent's model.",
			},
		},
		"required": []string{"task", "tools"},
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

func (t *spawnTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	task, _ := params["task"].(string)
	task = strings.TrimSpace(task)
	if task == "" {
		return "", fmt.Errorf("missing required param: task")
	}

	rawTools, _ := params["tools"].([]any)
	if len(rawTools) == 0 {
		return "", fmt.Errorf("missing required param: tools (must be a non-empty array of tool names)")
	}

	subRegistry := tools.NewRegistry()
	for _, raw := range rawTools {
		name, ok := raw.(string)
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" || strings.EqualFold(name, spawnToolName) {
			continue
		}
		if tool, found := t.engine.registry.Get(name); found {
			subRegistry.Register(tool)
		}
	}
	if len(subRegistry.All()) == 0 {
		return "", fmt.Errorf("none of the requested tools are available in the parent registry")
	}

	model, _ := params["model"].(string)
	model = strings.TrimSpace(model)
	if model == "" {
		model = strings.TrimSpace(t.engine.config.DefaultModel)
	}

	client := t.engine.client
	var cleanup func()
	if t.engine.subClientFactory != nil {
		client, cleanup = t.engine.subClientFactory("spawn")
	}
	if cleanup != nil {
		defer cleanup()
	}

	subEngine := New(client, subRegistry, Config{
		MaxSteps:        t.engine.config.MaxSteps,
		MaxTokenBudget:  t.engine.config.MaxTokenBudget,
		ParseRetries:    t.engine.config.ParseRetries,
		ToolRepeatLimit: t.engine.config.ToolRepeatLimit,
		DefaultModel:    model,
		ToolCallTimeout: t.engine.config.ToolCallTimeout,
	}, t.engine.spec, WithLogger(t.engine.log))

	final, _, err := subEngine.Run(ctx, task, RunOptions{Model: model})
	if err != nil {
		return "", fmt.Errorf("sub-agent failed: %w", err)
	}
	if final == nil {
		return "{}", nil
	}

	b, err := json.Marshal(final)
	if err != nil {
		return fmt.Sprintf("%v", final.Output), nil
	}
	return string(b), nil
}
