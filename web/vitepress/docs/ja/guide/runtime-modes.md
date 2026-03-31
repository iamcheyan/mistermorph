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

## LLM ルーティングポリシー

`main_loop`、`addressing`、`plan_create`、`heartbeat`、`memory_draft` ごとに別のモデルを割り当てたい場合は、[LLM ルーティングポリシー](/ja/guide/llm-routing) を参照してください。

詳細は次を参照:

- `docs/console.md`
- `docs/slack.md`
- `docs/line.md`
- `docs/lark.md`

旧 standalone daemon モード `mistermorph serve` は削除されています。
