---
title: 設定パターン
description: profiles、routes、ツールポリシーの代表例。
---

# 設定パターン

## LLM Profiles と Routes

```yaml
llm:
  model: gpt-5.4
  profiles:
    cheap:
      model: gpt-4.1-mini
    backup:
      provider: xai
      model: grok-4.1-fast-reasoning
  routes:
    main_loop:
      candidates:
        - profile: default
          weight: 1
        - profile: cheap
          weight: 1
      fallback_profiles: [backup]
    addressing:
      profile: cheap
    heartbeat: cheap
```

## ツールの有効/無効

```yaml
tools:
  bash:
    enabled: false
  url_fetch:
    enabled: true
    timeout: "30s"
```

## 実行上限

```yaml
max_steps: 20
tool_repeat_limit: 4
```

キーの完全一覧は `assets/config/config.example.yaml` を参照。
