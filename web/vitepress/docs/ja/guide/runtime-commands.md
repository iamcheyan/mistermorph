---
title: コマンド
description: chat、Console、channel runtime で使えるコマンド。
---

# コマンド

コマンドは、interactive chat、Console task、channel runtime の中で送る、`/` で始まるメッセージです。

> Slack では `/` が Slack 自身の command system を起動するため、`/` の前に空白を入れます。例: ` /model`
>
> Slack の group chat では bot を明示してコマンドを送る必要があります。Telegram の group chat では `/model@BotName` のような通常の bot command を使えます。

## 共通コマンド

次のコマンドは CLI chat、Console Web、Telegram、Slack、LINE、Lark で使えます。

| コマンド | 内容 |
|---|---|
| `/help` | 現在使えるコマンドを表示します。 |
| `/model` | 現在の model を表示します。 |
| `/skill` | 現在の skills を表示します。 |
| `/workspace` | 現在の workspace directory を表示します。 |

`/workspace` は次の形に対応しています。

| コマンド | 内容 |
|---|---|
| `/workspace` | 現在の workspace directory を表示します。 |
| `/workspace attach <dir>` | workspace directory を設定または置き換えます。 |
| `/workspace detach` | 現在の workspace を外します。 |

`/model` は次の形に対応しています。

| コマンド | 内容 |
|---|---|
| `/model` | 現在の model を表示します。 |
| `/model set <profile_name>` | 現在の model を切り替えます。 |

## CLI Chat 専用コマンド

次のコマンドは `mistermorph chat` でのみ使えます。

| コマンド | 内容 |
|---|---|
| `/exit` | chat session を終了します。 |
| `/quit` | chat session を終了します。 |
| `/reset` | 現在の conversation history を消します。 |
| `/memory` | 現在の project memory を表示します。 |
| `/remember <content>` | 現在の project に long-term memory を追加します。 |
| `/init` | 現在の project に `AGENTS.md` を生成します。 |
| `/update` | `AGENTS.md` を再生成し、既存ファイルを上書きします。 |

## Telegram 専用コマンド

次のコマンドは Telegram でのみ使えます。

| コマンド | 内容 |
|---|---|
| `/id` | 現在の Telegram chat id と chat type を表示します。 |
| `/reset` | その chat の履歴、sticky skills、known mentions、init state を消します。 |
