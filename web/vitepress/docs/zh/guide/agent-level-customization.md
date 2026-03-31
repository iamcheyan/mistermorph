---
title: Agent 底层扩展
description: 超出 integration Runtime API 的低层定制能力。
---

# Agent 底层扩展

当你需要的能力不在 `integration.Runtime` 暴露范围内时，使用本页方案。

## 最小可运行示例（agent 底层 API）

这个例子直接使用 `agent.New(...)`，不经过 `integration.Runtime`。

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

  reg := tools.NewRegistry() // 最小示例：不注册任何工具

  spec := agent.DefaultPromptSpec()
  spec.Blocks = append(spec.Blocks, agent.PromptBlock{
    Content: "[[ Runtime Rule ]]\\n请用一段简短中文回答。",
  })

  eng := agent.New(client, reg, agent.Config{
    MaxSteps:        8,
    ParseRetries:    2,
    ToolRepeatLimit: 3,
  }, spec)

  final, _, err := eng.Run(context.Background(), "介绍一下你自己", agent.RunOptions{
    Model: "gpt-5.4",
  })
  if err != nil {
    panic(err)
  }

  fmt.Println(final.Output)
}
```

## 典型场景

- 按任务动态注入 prompt block
- 完全替换 system prompt 构造逻辑
- 注入额外 LLM 请求参数
- 自定义工具成功/失败后的处理行为

## 用 `PromptSpec` 注入 Prompt Block

```go
spec := agent.DefaultPromptSpec()
spec.Blocks = append(spec.Blocks, agent.PromptBlock{
  Content: "[[ Project Policy ]]\n外部 API 调用必须带 trace_id。",
})

engine := agent.New(client, reg, agentCfg, spec)
```

## 用 `WithPromptBuilder` 完全替换 Prompt 构造

```go
engine := agent.New(
  client,
  reg,
  agentCfg,
  spec,
  agent.WithPromptBuilder(func(reg *tools.Registry, task string) string {
    return "你的完整自定义 system prompt"
  }),
)
```

## 其他低层 Hook

- `agent.WithParamsBuilder(...)`
- `agent.WithOnToolSuccess(...)`
- `agent.WithFallbackFinal(...)`
- `agent.WithPlanStepUpdate(...)`

## 范围边界

- 自定义工具注册：放在 [创建自己的 AI Agent：进阶](/zh/guide/build-your-own-agent-advanced)
- Telegram channel 接入：放在 [创建自己的 AI Agent：进阶](/zh/guide/build-your-own-agent-advanced)

## 建议

默认优先使用 `integration`。现在静态 prompt block 已可通过 `integration.Config.AddPromptBlock(...)` 完成。只有当你需要按任务动态改 prompt，或完全替换 system prompt 构造逻辑时，再下沉到这一层。
