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
