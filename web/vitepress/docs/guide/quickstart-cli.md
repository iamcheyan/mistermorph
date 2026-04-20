---
title: Quickstart (CLI)
description: Get a runnable CLI setup in a few minutes.
---

# Quickstart (CLI)

## 1. Install

```bash
curl -fsSL -o /tmp/install.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install.sh
```

## 2. Initialize

```bash
mistermorph install
```

Mister Morph initializes the required files. By default it keeps state under `~/.morph/`, cache under `~/.cache/morph`, and uses `~/.morph/config.yaml` as the config file.

During initialization, Mister Morph asks for the minimum required configuration, including the LLM setup, agent name, and persona.

### 2.1 Optional: configure with environment variables

In environments that need stronger security, Mister Morph supports placing sensitive values in environment variables and referencing them from the config file.

For example, you can put the LLM API key in an environment variable:

```bash
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
```

Then reference it in the config file:

```yaml
llm:
  api_key: "${MISTER_MORPH_LLM_API_KEY}"
```

## 3. Run your first task

```bash
mistermorph run --task "Hello"
```

It may output:

```json
{
  "reasoning": "Greet the user briefly.",
  "output": "Hello 👀",
  "reaction": "👀"
}
```

## 4. Start an interactive chat

```bash
mistermorph chat --workspace .
```

`mistermorph chat` also attaches the current working directory by default. Use `--no-workspace` when you want a chat session without a project tree.

For the difference between `workspace_dir`, `file_cache_dir`, and `file_state_dir`, see [Filesystem Roots](/guide/filesystem-roots).

## 5. Debug switches

```bash
mistermorph run --inspect-prompt --inspect-request --task "Hello"
```

This creates a `dump` directory in the current working directory with detailed prompt and request data.

For the full configuration surface, see [Config Patterns](/guide/config-patterns) and `assets/config/config.example.yaml`.
