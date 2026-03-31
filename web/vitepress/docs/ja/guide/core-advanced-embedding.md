---
title: Core 高度な組み込み
description: integration パッケージの提供範囲のみを扱う（設定、レジストリ、実行、チャネルランナー）。
---

# Core 高度な組み込み

このページは `integration` パッケージが直接提供する機能のみを扱います。

## `integration` が提供する機能

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

## 設定レイヤ（`integration.Config`）

- `Overrides` + `Set(key, value)`: Viper キーを上書き。
- `Features`: ランタイム機能注入の切り替え（`PlanTool` / `Guard` / `Skills`）。
- `BuiltinToolNames`: 組み込みツールのホワイトリスト（空で全有効）。
- `AddPromptBlock(...)`: system prompt の `Additional Policies` に静的 block を追加。
- `Inspect`: Prompt/Request のダンプ制御。

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.provider", "openai")
cfg.Set("llm.model", "gpt-5.4")
cfg.Set("llm.api_key", os.Getenv("OPENAI_API_KEY"))
cfg.Features.Skills = true
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
cfg.AddPromptBlock(`[[ Project Policy ]]
- 詳細指定がない限り、回答は 3 文以内に保つ。`)
```

## Prompt Block

`agent.New(...)` まで下りずに integration 層で prompt を足したい場合は、`cfg.AddPromptBlock(...)` を使います。

```go
cfg := integration.DefaultConfig()
cfg.AddPromptBlock(`[[ Tenant Policy ]]
- 外部ジョブを話題にするときは tenant_id を必ず含める。`)

rt := integration.New(cfg)
```

設定した block は次に適用されます。

- `NewRunEngine(...)`、`NewRunEngineWithRegistry(...)`、`RunTask(...)`
- `NewTelegramBot(...)`、`NewSlackBot(...)`

これは `integration.Runtime` 単位の静的設定です。タスクごとに prompt を変えたい場合は、より低レベルの agent API を使ってください。

## レジストリとカスタムツール

エンジン作成前にツールレジストリを拡張できます。

### 実行可能な例（カスタムツール + integration）

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

## 実行 API

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

#### Prepared Engine API を選ぶ理由

- ライフサイクル制御: `Cleanup()` のタイミングを明示的に管理できる。
- 再利用性: 同じ `prepared.Engine` を複数回 `Run(...)` できる。
- 実行ごとの柔軟性: 各 `Run` で異なる `RunOptions` を渡せる。
- 編成しやすさ: `prepared.Model` と `Engine` を直接扱えるため、上位のセッション層に統合しやすい。

### 省略 API

```go
final, runCtx, err := rt.RunTask(context.Background(), task, agent.RunOptions{})
_ = final
_ = runCtx
_ = err
```

## デバッグと診断

```go
cfg.Inspect.Prompt = true
cfg.Inspect.Request = true
cfg.Inspect.DumpDir = "./dump"
```

## LLM ルートポリシー

`integration.Config.Set(...)` では、ファーストパーティ runtime と同じ LLM ルートポリシーを設定できます。

ルート設定は `llm.routes.<purpose>` に置きます。`purpose` は次を使えます。

- `main_loop`
- `addressing`
- `heartbeat`
- `plan_create`
- `memory_draft`

各 route は次のいずれかの形を取れます。

- 固定 profile: `plan_create: "reasoning"`
- 明示オブジェクト: `profile` + 任意の `fallback_profiles`
- 分流オブジェクト: `candidates` + 任意の `fallback_profiles`

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.profiles", map[string]any{
  "cheap": map[string]any{
    "model": "gpt-4.1-mini",
  },
  "reasoning": map[string]any{
    "provider": "xai",
    "model": "grok-4.1-fast-reasoning",
    "api_key": os.Getenv("XAI_API_KEY"),
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
  "addressing": map[string]any{
    "profile": "cheap",
    "fallback_profiles": []string{"default"},
  },
})
```

挙動は次の通りです。

- `profile`: その route では常にその profile を使います。
- `candidates`: 現在の run ごとに重み付きで primary を 1 つ選び、その run 中の LLM 呼び出しで再利用します。
- 選ばれた primary がフォールバック対象エラーで失敗した場合、同じ route の他 candidate を先に試し、その後 `fallback_profiles` を順に試します。

この設定モデルは `integration` とファーストパーティ runtime の両方で共通です。

## Telegram チャネル接続（上級）

```go
tg, _ := rt.NewTelegramBot(integration.TelegramOptions{BotToken: os.Getenv("MISTER_MORPH_TELEGRAM_BOT_TOKEN")})
_ = tg
```

## Slack チャネル接続（任意）

```go
sl, _ := rt.NewSlackBot(integration.SlackOptions{
  BotToken: os.Getenv("MISTER_MORPH_SLACK_BOT_TOKEN"),
  AppToken: os.Getenv("MISTER_MORPH_SLACK_APP_TOKEN"),
})
_ = sl
```

## このページの範囲外

より低レベルな内容は [Agent レイヤ拡張](/ja/guide/agent-level-customization) を参照してください。
