---
title: Runtime Modes
description: Which runtime to use for CLI, bots, or web console.
---

# Runtime Modes

## One-shot task

```bash
mistermorph run --task "..."
```

## Telegram bot

```bash
mistermorph telegram --log-level info
```

## Slack bot

```bash
mistermorph slack --log-level info
```

## Console backend

```bash
mistermorph console serve
```

## LLM Routing Policies

If you want to assign different models to `main_loop`, `addressing`, `plan_create`, `heartbeat`, or `memory_draft`, see [LLM Routing Policies](/guide/llm-routing).

Detailed setup docs:

- `docs/console.md`
- `docs/slack.md`
- `docs/line.md`
- `docs/lark.md`

Legacy standalone daemon mode (`mistermorph serve`) has been removed.
