---
title: Config Patterns
description: Common patterns for routes, profiles, and tool policy.
---

# Config Patterns

## LLM Profiles and Routes

```yaml
llm:
  model: gpt-5.4
  profiles:
    cheap:
      model: gpt-4.1-mini
  routes:
    main_loop: default
    addressing: cheap
    heartbeat: cheap
```

## Tool Toggles

```yaml
tools:
  bash:
    enabled: false
  url_fetch:
    enabled: true
    timeout: "30s"
```

## Runtime Limits

```yaml
max_steps: 20
tool_repeat_limit: 4
```

Use `assets/config/config.example.yaml` as canonical key list.
