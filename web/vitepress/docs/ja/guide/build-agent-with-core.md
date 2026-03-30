---
title: Core で Agent を素早く構築
description: Go プロジェクトへ Mister Morph integration runtime を組み込む。
---

# Core で Agent を素早く構築

`integration` を組み込み入口として使います。

## 最小サンプル

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
- 既定では短い1段落で答えること。`)
  cfg.Set("llm.provider", "openai")
  cfg.Set("llm.model", "gpt-5.4")
  cfg.Set("llm.api_key", "YOUR_API_KEY")

  rt := integration.New(cfg)

  task := "README を読んで短く要約"
  final, _, err := rt.RunTask(context.Background(), task, agent.RunOptions{})
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## よく調整する項目

- `cfg.BuiltinToolNames`
- `cfg.AddPromptBlock(...)`
- `cfg.Set("max_steps", N)`
- `cfg.Set("tool_repeat_limit", N)`
- `cfg.Inspect.Prompt`、`cfg.Inspect.Request`

Prepared 方式（`NewRunEngine*`）は [Core 高度な組み込み](/ja/guide/core-advanced-embedding) にまとめています。
