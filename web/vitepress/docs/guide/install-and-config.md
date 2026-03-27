---
title: Install and Configure
description: Installation options and baseline configuration model.
---

# Install and Configure

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

Use `assets/config/config.example.yaml` as the full key reference.
