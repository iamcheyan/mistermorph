---
title: LLM Routing Policies
description: Choose profiles, traffic candidates, and fallback chains for different runtime purposes.
---

# LLM Routing Policies

`llm.routes.*` lets you assign different model configs to different runtime purposes.

These routes work the same way in first-party runtimes and in `integration.Config`, so CLI, Console, Channels, and Go embedding all share the same routing semantics.

## When you need routes

Typical cases:

- Keep the main loop on the default model to reduce migration cost.
- Send `addressing` to a cheaper and faster model.
- Pin `plan_create` to a stronger reasoning model.
- Split `main_loop` traffic and prepare a fallback chain for failures.

## Supported purposes

- `main_loop`: main agent step loop.
- `addressing`: group or channel addressing detection.
- `heartbeat`: scheduled heartbeat tasks.
- `plan_create`: planning requests inside the `plan_create` tool.
- `memory_draft`: memory draft consolidation.

## Minimal config

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

This means:

- The default main loop still uses the top-level `llm.*` settings.
- `plan_create` is pinned to `reasoning`.
- `addressing` is pinned to `cheap`.

## Three forms

### 1. Direct profile name

The shortest form:

```yaml
llm:
  routes:
    heartbeat: cheap
```

This means that the purpose is bound directly to one profile.

### 2. Explicit object

If you also want a local fallback chain, use an object:

```yaml
llm:
  routes:
    plan_create:
      profile: reasoning
      fallback_profiles: [default]
```

Rules:

- `profile` is the primary route profile.
- `fallback_profiles` is the route-local fallback chain.

### 3. Candidate-based routing

If you want traffic splitting for the same purpose, use `candidates`:

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

Rules:

- `weight` controls candidate selection weight.
- Each run picks one primary candidate and reuses it for the whole run.
- If the primary candidate hits a retryable fallback error, the runtime first tries the remaining candidates in the same route, then tries `fallback_profiles` in order.

## How to write this in integration

If you are embedding from Go, the config style is the same, but written with `cfg.Set(...)` instead of YAML:

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.routes.plan_create", "reasoning")
cfg.Set("llm.routes.addressing", map[string]any{
  "profile": "cheap",
  "fallback_profiles": []string{"default"},
})
```

If you want the full field list, see [Config Fields](/guide/config-reference). If you want common YAML patterns, see [Config Patterns](/guide/config-patterns).
