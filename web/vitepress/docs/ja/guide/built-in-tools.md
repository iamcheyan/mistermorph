---
title: 組み込みツール
description: 静的ツール、ランタイム注入ツール、チャネル専用ツール。
---

# 組み込みツール

ツール登録は「静的登録 + ランタイム注入」の2段階です。

## 静的ツール（設定駆動）

runtime 文脈なしで作られるツール:

- `read_file`
- `write_file`
- `bash`
- `url_fetch`
- `web_search`
- `contacts_send`

## ランタイムツール（LLM 文脈依存）

runtime で注入されるツール:

- `plan_create`
- `todo_update`

## チャネル専用ツール

- Telegram では次が追加される場合があります:
  - `telegram_send_voice`
  - `telegram_send_photo`
  - `telegram_send_file`
  - `message_react`
- Slack では次が追加される場合があります:
  - `message_react`

## Core 組み込み時のホワイトリスト

`integration.Config.BuiltinToolNames` で組み込みツールを制限できます。

```go
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
```

空配列は全組み込みツールを意味します。

## 設定セクション

```yaml
tools:
  read_file: ...
  write_file: ...
  bash: ...
  url_fetch: ...
  web_search: ...
  contacts_send: ...
  todo_update: ...
  plan_create: ...
```

パラメータ詳細は `docs/tools.md` を参照してください。
