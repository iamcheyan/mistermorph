---
title: Advanced Core Embedding
description: "Integration runtime capabilities: config, registry, run engine, and channel runners."
---

# Advanced Core Embedding

This page only covers capabilities provided by the `integration` package.

## What `integration` Provides

- `integration.DefaultConfig()` / `integration.Config.Set(...)`
- `integration.Config.AddPromptBlock(...)`
- `integration.New(cfg)`
- `rt.NewRegistry()`
- `rt.NewRunEngine(...)`
- `rt.NewRunEngineWithRegistry(...)`
- `rt.RunTask(...)`
- `rt.RequestTimeout()`
- `rt.NewTelegramBot(...)`
- `rt.NewSlackBot(...)`

## Config Layer (Inside `integration.Config`)

- `Overrides` + `Set(key, value)`: override Viper keys.
- `Features`: toggle built-in runtime wiring (`PlanTool`, `Guard`, `Skills`).
- `BuiltinToolNames`: built-in tool whitelist (empty = all built-ins).
- `AddPromptBlock(...)`: append static prompt blocks under system prompt `Additional Policies`.
- `Inspect`: prompt/request dump behavior.

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.provider", "openai")
cfg.Set("llm.model", "gpt-5.4")
cfg.Set("llm.api_key", os.Getenv("OPENAI_API_KEY"))
cfg.Set("llm.routes", map[string]any{
  "main_loop": map[string]any{
    "candidates": []map[string]any{
      {"profile": "default", "weight": 1},
      {"profile": "cheap", "weight": 1},
    },
    "fallback_profiles": []string{"reasoning"},
  },
})
cfg.Features.Skills = true
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
cfg.AddPromptBlock(`[[ Project Policy ]]
- Keep answers under 3 sentences unless detail is requested.`)
```

## Prompt Blocks

Use `cfg.AddPromptBlock(...)` when you want integration-level prompt customization without moving to `agent.New(...)`.

```go
cfg := integration.DefaultConfig()
cfg.AddPromptBlock(`[[ Tenant Policy ]]
- Always include tenant_id when talking about external jobs.`)

rt := integration.New(cfg)
```

Configured blocks are applied to:

- one-shot runs via `NewRunEngine(...)`, `NewRunEngineWithRegistry(...)`, and `RunTask(...)`
- channel runtimes created by `NewTelegramBot(...)` and `NewSlackBot(...)`

This is intentionally static per `integration.Runtime`. If you need task-by-task prompt changes, use the lower-level agent APIs.

## Registry and Custom Tools

`integration` lets you extend runtime registry before engine creation.

### Runnable Example (Custom Tool + integration Runtime)

```go
package main

import (
  "context"
  "encoding/json"
  "fmt"
  "os"
  "strings"

  "github.com/quailyquaily/mistermorph/agent"
  "github.com/quailyquaily/mistermorph/integration"
)

type EchoTool struct{}

func (t *EchoTool) Name() string { return "echo_text" }

func (t *EchoTool) Description() string {
  return "Echoes input text as JSON."
}

func (t *EchoTool) ParameterSchema() string {
  return `{
  "type": "object",
  "properties": {
    "text": {"type": "string", "description": "Text to echo."}
  },
  "required": ["text"]
}`
}

func (t *EchoTool) Execute(_ context.Context, params map[string]any) (string, error) {
  text, _ := params["text"].(string)
  text = strings.TrimSpace(text)
  if text == "" {
    return "", fmt.Errorf("text is required")
  }
  b, _ := json.Marshal(map[string]any{"text": text})
  return string(b), nil
}

func main() {
  cfg := integration.DefaultConfig()
  cfg.Set("llm.provider", "openai")
  cfg.Set("llm.model", "gpt-5.4")
  cfg.Set("llm.api_key", os.Getenv("OPENAI_API_KEY"))

  rt := integration.New(cfg)
  reg := rt.NewRegistry()
  reg.Register(&EchoTool{})

  task := "Call tool echo_text with text 'hello from tool', then answer with that text."

  prepared, err := rt.NewRunEngineWithRegistry(context.Background(), task, reg)
  if err != nil {
    panic(err)
  }
  defer prepared.Cleanup()

  final, _, err := prepared.Engine.Run(context.Background(), task, agent.RunOptions{Model: prepared.Model})
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## Run APIs

### Prepared Engine API

```go
prepared, err := rt.NewRunEngine(context.Background(), task)
if err != nil {
  panic(err)
}
defer prepared.Cleanup()

final, _, err := prepared.Engine.Run(context.Background(), task, agent.RunOptions{
  Model: prepared.Model,
})
```

#### Why Use Prepared Engine API

- Controlled lifecycle: you decide exactly when to call `Cleanup()`.
- Reusability: reuse the same `prepared.Engine` for multiple runs.
- Per-run flexibility: pass different `RunOptions` on each run.
- Better orchestration: direct access to `prepared.Model` and `Engine` for your session/scheduler layer.

### Convenience API

```go
final, runCtx, err := rt.RunTask(context.Background(), task, agent.RunOptions{})
_ = final
_ = runCtx
_ = err
```

## Inspect and Runtime Diagnostics

```go
cfg.Inspect.Prompt = true
cfg.Inspect.Request = true
cfg.Inspect.DumpDir = "./dump"
```

## Route Policies

`integration.Config.Set(...)` can configure the same route policies used by first-party runtimes.

- fixed route: `plan_create: "reasoning"`
- weighted split: `main_loop.candidates`
- route-local fallback: `main_loop.fallback_profiles`

One candidate is selected once for the current run and reused for that run's LLM calls.

## Telegram Channel Integration (Advanced)

```go
tg, _ := rt.NewTelegramBot(integration.TelegramOptions{BotToken: os.Getenv("MISTER_MORPH_TELEGRAM_BOT_TOKEN")})
_ = tg
```

## Slack Channel Integration (Optional)

```go
sl, _ := rt.NewSlackBot(integration.SlackOptions{
  BotToken: os.Getenv("MISTER_MORPH_SLACK_BOT_TOKEN"),
  AppToken: os.Getenv("MISTER_MORPH_SLACK_APP_TOKEN"),
})
_ = sl
```

## Out of Scope for This Page

Low-level engine customization is documented in [Agent-Level Customization](/guide/agent-level-customization).
