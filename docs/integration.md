# Integration

## Embedding to other projects

Two common integration options:

- As a Go library: see `demo/embed-go/`.
- As a subprocess CLI: see `demo/embed-cli/`.

For Go-library embedding with built-in wiring, use `integration`:

```go
cfg := integration.DefaultConfig()
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"} // optional; empty = all built-ins
cfg.AddPromptBlock(`[[ Project Policy ]]
- Keep answers under 3 sentences unless asked for detail.`) // optional; appended under "Additional Policies"
cfg.Inspect.Prompt = true   // optional
cfg.Inspect.Request = true  // optional
cfg.Set("llm.routes", map[string]any{
	"main_loop": map[string]any{
		"candidates": []map[string]any{
			{"profile": "default", "weight": 1},
			{"profile": "cheap", "weight": 1},
		},
		"fallback_profiles": []string{"reasoning"},
	},
})
cfg.Set("llm.api_key", os.Getenv("OPENAI_API_KEY"))

rt := integration.New(cfg)

reg := rt.NewRegistry() // built-in tools wiring
prepared, err := rt.NewRunEngineWithRegistry(ctx, task, reg)
if err != nil { /* ... */ }
defer prepared.Cleanup()

final, runCtx, err := prepared.Engine.Run(ctx, task, agent.RunOptions{Model: prepared.Model})
_ = final
_ = runCtx
```

## Prompt blocks

`integration.Config.AddPromptBlock(...)` appends static prompt content into the system prompt's `Additional Policies` section.

Use it when you want integration-level customization without dropping down to `agent.New(...)`:

- project-wide safety or style rules
- tenant- or deployment-specific policies
- static instructions that should apply to one-shot runs and channel runtimes

The same configured blocks are applied to:

- `rt.NewRunEngine(...)`
- `rt.NewRunEngineWithRegistry(...)`
- `rt.RunTask(...)`
- `rt.NewTelegramBot(...)`
- `rt.NewSlackBot(...)`

If you need per-task dynamic prompt shaping or full prompt replacement, use the lower-level `agent` APIs instead.

## Route Policies

`integration.Config.Set(...)` can also configure the same route policies used by first-party runtimes.

Use `llm.routes.<purpose>` to choose one of:

- a fixed profile, for example `plan_create: reasoning`
- a weighted candidate list for per-run traffic split
- a route-local `fallback_profiles` chain

Example:

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.profiles", map[string]any{
	"cheap": map[string]any{"model": "gpt-4.1-mini"},
	"reasoning": map[string]any{
		"provider": "xai",
		"model":    "grok-4.1-fast-reasoning",
		"api_key":  os.Getenv("XAI_API_KEY"),
	},
})
cfg.Set("llm.routes", map[string]any{
	"main_loop": map[string]any{
		"candidates": []map[string]any{
			{"profile": "default", "weight": 1},
			{"profile": "cheap", "weight": 1},
		},
		"fallback_profiles": []string{"reasoning"},
	},
	"plan_create": "reasoning",
})
```

When a route uses `candidates`, one primary profile is selected once for the current run and reused for that run's LLM calls. If the selected primary fails with a fallback-eligible error, the route's `fallback_profiles` are tried in order.
