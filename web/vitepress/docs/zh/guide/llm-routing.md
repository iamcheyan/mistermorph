---
title: LLM 路由策略
description: 为不同 runtime purpose 选择 profile、分流候选与 fallback 链。
---

# LLM 路由策略

`llm.routes.*` 用来给不同的 runtime purpose 指定不同的模型配置。

这套配置同时适用于第一方 runtime 和 `integration.Config`，所以 CLI / Console / Channel / Go 嵌入用的是同一套路由语义。

## 什么时候需要 routes

典型场景：

- 主循环用默认模型，降低迁移成本。
- `addressing` 用更便宜更快的模型。
- `plan_create` 固定走更强的推理模型。
- `main_loop` 做分流，并给失败请求准备 fallback 链。

## 支持的 purpose

- `main_loop`：主 Agent step loop。
- `addressing`：群聊或频道里的 addressing 判定。
- `heartbeat`：定时 heartbeat 任务。
- `plan_create`：`plan_create` 工具内部的规划请求。
- `memory_draft`：memory 草稿整理。

## 最小配置

```yaml
llm:
  provider: openai
  model: gpt-5.4
  api_key: ${OPENAI_API_KEY}

  profiles:
    cheap:
      model: gpt-4.1-mini
    reasoning:
      provider: xai
      model: grok-4.1-fast-reasoning
      api_key: ${XAI_API_KEY}

  routes:
    plan_create: reasoning
    addressing: cheap
```

这里的含义是：

- 默认主循环继续走顶层 `llm.*`
- `plan_create` 固定走 `reasoning`
- `addressing` 固定走 `cheap`

## 三种写法

### 1. 直接写 profile 名

最短写法：

```yaml
llm:
  routes:
    heartbeat: cheap
```

等价于“这个 purpose 固定绑定到一个 profile”。

### 2. 显式对象

当你还想加本地 fallback 链时，可以写成对象：

```yaml
llm:
  routes:
    plan_create:
      profile: reasoning
      fallback_profiles: [default]
```

规则：

- `profile` 是主路由 profile。
- `fallback_profiles` 是这个 route 自己的回退链。

### 3. 候选分流

如果要对同一个 purpose 做流量分流，用 `candidates`：

```yaml
llm:
  routes:
    main_loop:
      candidates:
        - profile: default
          weight: 1
        - profile: cheap
          weight: 1
      fallback_profiles: [reasoning]
```

规则：

- `candidates` 里的 `weight` 用来决定主候选的选择权重。
- 同一个 run 会先选出一个主候选，并在这个 run 内复用。
- 如果主候选遇到可回退错误，运行时会先尝试同 route 下剩余 candidate，再按顺序尝试 `fallback_profiles`。

## 在 integration 里怎么写

如果你是 Go 嵌入，配置方式还是一样，只是把 YAML 改成 `cfg.Set(...)`：

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.routes.plan_create", "reasoning")
cfg.Set("llm.routes.addressing", map[string]any{
  "profile": "cheap",
  "fallback_profiles": []string{"default"},
})
```

如果要查所有字段名，见 [配置字段](/zh/guide/config-reference)；如果要看常见 YAML 模式，见 [配置模式](/zh/guide/config-patterns)。
