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

## Legacy daemon (if needed)

```bash
mistermorph serve --server-listen 127.0.0.1:8787
```

Detailed setup docs:

- `docs/console.md`
- `docs/slack.md`
- `docs/line.md`
- `docs/lark.md`
