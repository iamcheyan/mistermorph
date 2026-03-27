---
title: Prompt 設計（トップダウン）
description: Identity 層から Runtime 層まで、システム Prompt の組み立て順を説明。
---

# Prompt 設計（トップダウン）

以下は現行コードの実際の順序です。

## トップダウン図

```text
PromptSpecWithSkills(...)
        |
        v
ApplyPersonaIdentity(...)
        |
        v
AppendLocalToolNotesBlock(...)
        |
        v
AppendPlanCreateGuidanceBlock(...)
        |
        v
AppendTodoWorkflowBlock(...)
        |
        v
Channel PromptAugment(...)
        |
        v
AppendMemorySummariesBlock(...)
        |
        v
BuildSystemPrompt(...) -> system prompt テキスト
```

## 1) Identity 層

ベースは `agent.DefaultPromptSpec()`。

次にローカル persona ファイルで上書きされる場合があります。

- `file_state_dir/IDENTITY.md`
- `file_state_dir/SOUL.md`

適用関数は `promptprofile.ApplyPersonaIdentity(...)`。

## 2) Skills 層

`skillsutil.PromptSpecWithSkills(...)` は skill のメタ情報だけを注入します。

- 名前
- `SKILL.md` パス
- 説明
- `auth_profiles`（ある場合）

実際の `SKILL.md` 本文は必要時に `read_file` で読みます。

## 3) Core ポリシーブロック

task runtime では次の順で追加されます。

1. `SCRIPTS.md` のローカルツールノート
2. `plan_create` ガイダンス
3. TODO ワークフローポリシー

## 4) チャネルブロック

次にチャネルごとのブロックを追加します。

- Telegram
- Slack
- LINE
- Lark

## 5) Memory 注入ブロック

memory が有効な場合、memory summary ブロックを追加します。

## 6) システム Prompt のレンダリング

`agent.BuildSystemPrompt(...)` が `agent/prompts/system.md` を使って以下を統合します。

- identity
- skills メタ情報
- 追加ブロック
- ツール要約
- 追加ルール

## 7) LLM に送るメッセージ順

`engine.Run(...)` の順序:

```text
[system] レンダリング済み system prompt
   ->
[user] mister_morph_meta（任意）
   ->
[history] system 以外の履歴
   ->
[user] current message または raw task
```

1. system prompt
2. runtime metadata（`mister_morph_meta`）
3. history（system 以外）
4. current message または raw task

## 実務上の優先順位

Prompt を調整する時は次の順で拡張するのが安全です。

1. `IDENTITY.md` / `SOUL.md`
2. `SKILL.md` + `skills.load`
3. `SCRIPTS.md`
4. runtime prompt augment
5. `agent.WithPromptBuilder(...)`（完全カスタム時のみ）
