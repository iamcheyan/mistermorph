---
title: 配置模式
description: 常用的 profiles、routes 与工具策略配置方法。
---

# 配置模式

## LLM Profiles 与 Routes

```yaml
llm:
  model: gpt-5.4
  profiles:
    cheap:
      model: gpt-4.1-mini
    backup:
      provider: xai
      model: grok-4.1-fast-reasoning
  fallback_profiles: [cheap, backup]
  routes:
    main_loop: default
    addressing: cheap
    heartbeat: cheap
```

## 工具开关

```yaml
tools:
  bash:
    enabled: false
  url_fetch:
    enabled: true
    timeout: "30s"
```

## 运行上限

```yaml
max_steps: 20
tool_repeat_limit: 4
```

完整键定义以 `assets/config/config.example.yaml` 为准。
