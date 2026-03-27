---
title: Runtime モード
description: CLI、Bot、Console の実行モードを選ぶ。
---

# Runtime モード

## 単発タスク

```bash
mistermorph run --task "..."
```

## Telegram Bot

```bash
mistermorph telegram --log-level info
```

## Slack Bot

```bash
mistermorph slack --log-level info
```

## Console バックエンド

```bash
mistermorph console serve
```

## 旧 Daemon（必要時のみ）

```bash
mistermorph serve --server-listen 127.0.0.1:8787
```

詳細は次を参照:

- `docs/console.md`
- `docs/slack.md`
- `docs/line.md`
- `docs/lark.md`
