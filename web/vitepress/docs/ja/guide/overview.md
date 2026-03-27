---
title: 概要
description: Mister Morph の全体像と推奨読書順。
---

# 概要

Mister Morph には主に 2 つの使い方があります。

- CLI ワークフロー（`mistermorph run`、`telegram`、`slack`、`console serve`）
- Go への組み込み（`integration` パッケージ）

## 目的別の読み方

- まず動かす: [クイックスタート（CLI）](/ja/guide/quickstart-cli)
- Go に組み込む: [Core で Agent を素早く構築](/ja/guide/build-agent-with-core)
- 常駐実行を理解: [Runtime モード](/ja/guide/runtime-modes)
- 本番運用の安全化: [セキュリティと Guard](/ja/guide/security-and-guard)

## リポジトリ構成（要点）

- CLI エントリ: `cmd/mistermorph/`
- Agent エンジン: `agent/`
- 組み込み Core: `integration/`
- 組み込みツール: `tools/`
- Provider 実装: `providers/`
- 詳細文書: `docs/`
