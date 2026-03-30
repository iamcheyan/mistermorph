---
title: Build an Agent with Core
description: Embed Mister Morph integration runtime in your Go project.
---

# Build an Agent with Core

Use `integration` as the embedding entrypoint.

## Minimal Example

```go
package main

import (
  "context"
  "fmt"

  "github.com/quailyquaily/mistermorph/agent"
  "github.com/quailyquaily/mistermorph/integration"
)

func main() {
  cfg := integration.DefaultConfig()
  cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
  cfg.AddPromptBlock(`[[ Project Policy ]]
- Answer in one short paragraph by default.`)
  cfg.Set("llm.provider", "openai")
  cfg.Set("llm.model", "gpt-5.4")
  cfg.Set("llm.api_key", "YOUR_API_KEY")

  rt := integration.New(cfg)

  task := "Read README and output a short summary"
  final, _, err := rt.RunTask(context.Background(), task, agent.RunOptions{})
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## Common Tuning

- `cfg.BuiltinToolNames`
- `cfg.AddPromptBlock(...)`
- `cfg.Set("max_steps", N)`
- `cfg.Set("tool_repeat_limit", N)`
- `cfg.Inspect.Prompt`, `cfg.Inspect.Request`

Prepared mode (`NewRunEngine*`) is covered in [Advanced Core Embedding](/guide/core-advanced-embedding).
