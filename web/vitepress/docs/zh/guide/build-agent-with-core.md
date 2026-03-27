---
title: 用 Core 快速搭建 Agent
description: 在 Go 项目中嵌入 Mister Morph integration runtime。
---

# 用 Core 快速搭建 Agent

`integration` 是推荐的嵌入入口。

## 最小示例

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
  cfg.Set("llm.provider", "openai")
  cfg.Set("llm.model", "gpt-5.4")
  cfg.Set("llm.api_key", "YOUR_API_KEY")

  rt := integration.New(cfg)

  task := "读取 README 并输出简短摘要"
  final, _, err := rt.RunTask(context.Background(), task, agent.RunOptions{})
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## 常用调节点

- `cfg.BuiltinToolNames`
- `cfg.Set("max_steps", N)`
- `cfg.Set("tool_repeat_limit", N)`
- `cfg.Inspect.Prompt`、`cfg.Inspect.Request`

Prepared 方式（`NewRunEngine*`）放在 [Core 嵌入进阶](/zh/guide/core-advanced-embedding)。
