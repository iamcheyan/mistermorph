---
title: Config Patterns
description: Install options, baseline configuration, and common patterns for routes, profiles, and tool policy.
---

# Config Patterns

## Install Options

```bash
# Release installer
curl -fsSL -o /tmp/install-mistermorph.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install-mistermorph.sh
```

```bash
# Go install
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

## Initialize Files

```bash
mistermorph install
```

Default workspace is `~/.morph/`.

## Config Sources (precedence)

- CLI flags
- Environment variables
- `config.yaml`

## Minimal `config.yaml`

```yaml
llm:
  provider: openai
  model: gpt-5.4
  endpoint: https://api.openai.com
  api_key: ${OPENAI_API_KEY}
```

## LLM Profiles and Routes

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
