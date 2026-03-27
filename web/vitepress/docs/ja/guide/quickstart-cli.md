---
title: クイックスタート（CLI）
description: 数分で CLI を実行可能にする最短手順。
---

# クイックスタート（CLI）

## 1. インストール

```bash
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

## 2. ワークスペース初期化

```bash
mistermorph install
```

## 3. モデル認証情報を設定

```bash
export MISTER_MORPH_LLM_PROVIDER="openai"
export MISTER_MORPH_LLM_MODEL="gpt-5.4"
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
```

## 4. 最初のタスク実行

```bash
mistermorph run --task "Summarize this repository"
```

## 5. デバッグスイッチ

```bash
mistermorph run --inspect-prompt --inspect-request --task "hello"
```

設定の全体像は [設定パターン](/ja/guide/config-patterns) と `assets/config/config.example.yaml` を参照。
