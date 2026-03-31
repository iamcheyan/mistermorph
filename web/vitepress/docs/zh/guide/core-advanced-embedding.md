---
title: Core 嵌入进阶
description: 仅覆盖 integration 包能力：配置、注册表、运行引擎与通道运行器。
---

# Core 嵌入进阶

本页只讲 `integration` 包直接提供的能力。

## `integration` 提供了什么

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

## 配置层（`integration.Config`）

- `Overrides` + `Set(key, value)`：覆盖任意 Viper 配置键。
- `Features`：控制运行时能力注入（`PlanTool`、`Guard`、`Skills`）。
- `BuiltinToolNames`：内置工具白名单（空表示全部）。
- `AddPromptBlock(...)`：向 system prompt 的 `Additional Policies` 追加静态 prompt block。
- `Inspect`：Prompt/Request 落盘调试。

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.provider", "openai")
cfg.Set("llm.model", "gpt-5.4")
cfg.Set("llm.api_key", os.Getenv("OPENAI_API_KEY"))
cfg.Features.Skills = true
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
cfg.AddPromptBlock(`[[ Project Policy ]]
- 除非用户要求展开，否则回复控制在 3 句以内。`)
```

## Prompt Block

如果你想在 `integration` 层做 prompt 定制，而不是直接下沉到 `agent.New(...)`，用 `cfg.AddPromptBlock(...)`。

```go
cfg := integration.DefaultConfig()
cfg.AddPromptBlock(`[[ Tenant Policy ]]
- 讨论外部任务时总是带上 tenant_id。`)

rt := integration.New(cfg)
```

这些 block 会自动应用到：

- `NewRunEngine(...)`、`NewRunEngineWithRegistry(...)`、`RunTask(...)`
- `NewTelegramBot(...)` 与 `NewSlackBot(...)`

这类 block 是 `integration.Runtime` 级别的静态配置。若你需要按任务动态修改 prompt，还是应使用更底层的 agent API。

## 注册表与自定义工具

`integration` 支持在创建引擎前扩展工具注册表。

### 可运行示例（自定义工具 + integration）

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

## 运行 API

### Prepared Engine 方式

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

#### 为什么选 Prepared Engine 方式

- 生命周期可控：你可以明确在何时 `Cleanup()`，适合接入你自己的进程管理。
- 可复用：同一个 `prepared.Engine` 可以多次 `Run(...)`，避免重复准备。
- 运行参数可变：每次 `Run` 都可传不同 `RunOptions`（如 `History`、`Meta`、`OnStream`）。
- 便于编排：你能直接拿到 `prepared.Model` 与 `Engine`，更适合做上层会话/调度封装。

### 便捷方式

```go
final, runCtx, err := rt.RunTask(context.Background(), task, agent.RunOptions{})
_ = final
_ = runCtx
_ = err
```

## 调试与诊断

```go
cfg.Inspect.Prompt = true
cfg.Inspect.Request = true
cfg.Inspect.DumpDir = "./dump"
```

## LLM 路由策略

`integration.Config.Set(...)` 可以配置和第一方 runtime 完全一致的 LLM 路由策略。

路由配置统一放在 `llm.routes.<purpose>` 下，`purpose` 支持：

- `main_loop`
- `addressing`
- `heartbeat`
- `plan_create`
- `memory_draft`

每个 route 可以用这几种形态之一：

- 固定 profile：`plan_create: "reasoning"`
- 显式对象：`profile` + 可选 `fallback_profiles`
- 分流对象：`candidates` + 可选 `fallback_profiles`

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

行为规则：

- `profile`：该 route 固定走一个 profile。
- `candidates`：当前 run 会先按权重选出一个主候选，并在这个 run 内复用。
- 如果主候选遇到可回退错误，运行时会先尝试同 route 下其他 candidate，再按顺序尝试 `fallback_profiles`。

这套配置同时适用于 `integration` 和第一方 runtime，因此嵌入场景与内建运行模式使用的是同一模型。

## 接入 Telegram Channel（进阶）

```go
tg, _ := rt.NewTelegramBot(integration.TelegramOptions{BotToken: os.Getenv("MISTER_MORPH_TELEGRAM_BOT_TOKEN")})
_ = tg
```

## 接入 Slack Channel（可选）

```go
sl, _ := rt.NewSlackBot(integration.SlackOptions{
  BotToken: os.Getenv("MISTER_MORPH_SLACK_BOT_TOKEN"),
  AppToken: os.Getenv("MISTER_MORPH_SLACK_APP_TOKEN"),
})
_ = sl
```

## 本页不覆盖内容

更底层的能力请看 [Agent 底层扩展](/zh/guide/agent-level-customization)。
