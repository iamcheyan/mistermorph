---
title: Agent レイヤ拡張
description: integration Runtime API を超える低レベルカスタマイズ。
---

# Agent レイヤ拡張

`integration.Runtime` が公開していない挙動が必要なときに使います。

## 最小実行例（Agent 低レベル API）

この例は `integration.Runtime` を使わず、`agent.New(...)` を直接使います。

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

  reg := tools.NewRegistry() // 最小例: ツール未登録

  spec := agent.DefaultPromptSpec()
  spec.Blocks = append(spec.Blocks, agent.PromptBlock{
    Content: "[[ Runtime Rule ]]\\n短い1段落で回答すること。",
  })

  eng := agent.New(client, reg, agent.Config{
    MaxSteps:        8,
    ParseRetries:    2,
    ToolRepeatLimit: 3,
  }, spec)

  final, _, err := eng.Run(context.Background(), "自己紹介してください", agent.RunOptions{
    Model: "gpt-5.4",
  })
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## 典型ケース

- タスク単位で動的に prompt block を注入したい
- system prompt 構築を丸ごと差し替えたい
- 追加の LLM リクエストパラメータを入れたい
- ツール成功/失敗後の処理を独自化したい

## `PromptSpec` で Prompt Block を注入

```go
spec := agent.DefaultPromptSpec()
spec.Blocks = append(spec.Blocks, agent.PromptBlock{
  Content: "[[ Project Policy ]]\n外部 API 呼び出しには trace_id を必ず付与する。",
})

engine := agent.New(client, reg, agentCfg, spec)
```

## `WithPromptBuilder` で Prompt 構築を差し替え

```go
engine := agent.New(
  client,
  reg,
  agentCfg,
  spec,
  agent.WithPromptBuilder(func(reg *tools.Registry, task string) string {
    return "完全にカスタムな system prompt"
  }),
)
```

## その他の低レベル Hook

- `agent.WithParamsBuilder(...)`
- `agent.WithOnToolSuccess(...)`
- `agent.WithFallbackFinal(...)`
- `agent.WithPlanStepUpdate(...)`

## 範囲の切り分け

- カスタムツール登録: [Core 高度な組み込み](/ja/guide/core-advanced-embedding)
- Telegram チャネル接続: [Core 高度な組み込み](/ja/guide/core-advanced-embedding)

## 推奨方針

通常は `integration` を優先し、静的な prompt block は `integration.Config.AddPromptBlock(...)` で足してください。タスクごとの動的な prompt 変更や、system prompt 全体の差し替えが必要なときだけこのレイヤへ下りるのが安全です。
