---
title: Quickstart (CLI)
description: Get a runnable CLI setup in a few minutes.
---

# Quickstart (CLI)

## 1. Install

```bash
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

## 2. Initialize workspace

```bash
mistermorph install
```

## 3. Set model credentials

```bash
export MISTER_MORPH_LLM_PROVIDER="openai"
export MISTER_MORPH_LLM_MODEL="gpt-5.4"
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
```

## 4. Run first task

```bash
mistermorph run --task "Summarize this repository"
```

## 5. Useful debug switches

```bash
mistermorph run --inspect-prompt --inspect-request --task "hello"
```

For full config keys, see [Config Patterns](/guide/config-patterns) and `assets/config/config.example.yaml`.
