---
title: Agent-Level Customization
description: Low-level customization beyond integration runtime APIs.
---

# Agent-Level Customization

Use this page when you need behavior not exposed by `integration.Runtime`.

## Minimal Runnable Example (Agent Low-Level API)

This example uses `agent.New(...)` directly, without `integration.Runtime`.

```go
package main

import (
  "context"
  "fmt"
  "os"

  "github.com/quailyquaily/mistermorph/agent"
  "github.com/quailyquaily/mistermorph/providers/uniai"
  "github.com/quailyquaily/mistermorph/tools"
)

func main() {
  client := uniai.New(uniai.Config{
    Provider: "openai",
    Endpoint: "https://api.openai.com",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "gpt-5.4",
  })

  reg := tools.NewRegistry() // minimal example: no tools

  spec := agent.DefaultPromptSpec()
  spec.Blocks = append(spec.Blocks, agent.PromptBlock{
    Content: "[[ Runtime Rule ]]\\nAnswer in one short paragraph.",
  })

  eng := agent.New(client, reg, agent.Config{
    MaxSteps:        8,
    ParseRetries:    2,
    ToolRepeatLimit: 3,
  }, spec)

  final, _, err := eng.Run(context.Background(), "Introduce yourself", agent.RunOptions{
    Model: "gpt-5.4",
  })
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## Typical Cases

- inject dynamic prompt blocks from code per task
- replace system prompt rendering logic
- inject extra LLM request params
- customize tool success/fallback handling

## Inject Prompt Blocks with `PromptSpec`

```go
spec := agent.DefaultPromptSpec()
spec.Blocks = append(spec.Blocks, agent.PromptBlock{
  Content: "[[ Project Policy ]]\nAlways include trace_id in external API calls.",
})

engine := agent.New(client, reg, agentCfg, spec)
```

## Replace Prompt Builder

```go
engine := agent.New(
  client,
  reg,
  agentCfg,
  spec,
  agent.WithPromptBuilder(func(reg *tools.Registry, task string) string {
    return "your fully custom system prompt"
  }),
)
```

## Other Low-Level Hooks

- `agent.WithParamsBuilder(...)`
- `agent.WithOnToolSuccess(...)`
- `agent.WithFallbackFinal(...)`
- `agent.WithPlanStepUpdate(...)`

## Scope Boundary

- custom tool registration: [Create Your Own AI Agent: Advanced](/guide/core-advanced-embedding)
- Telegram channel integration: [Create Your Own AI Agent: Advanced](/guide/core-advanced-embedding)

## Recommendation

Prefer `integration` by default. Static prompt blocks are now supported via `integration.Config.AddPromptBlock(...)`. Move to this level only when you need capabilities that `integration` still does not expose, such as per-task dynamic prompt shaping or full system prompt replacement.
