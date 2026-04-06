---
title: 概要
description: Mister Morph の使い方と推奨読書順を把握する。
---

# 概要

Mister Morph には主に 3 つの使い方があります。

- デスクトップ App。個人向け AI アシスタントとして使ったり、複数の Mister Morph インスタンスを管理したりできます。
- CLI ワークフロー
  - Telegram や Slack など、異なる Channels 向けの Agent プロセスを個別に動かせます。
  - Console の Web UI を使えます。これも複数の Mister Morph インスタンス管理に使えます。
  - 単発の Agent タスクを実行できます。
- Go core を自分のプログラムに組み込み、Agent ランタイムを注入できます。

## 目的別の読み方

- まず素早く動かしたい: [クイックスタート（CLI）](/ja/guide/quickstart-cli)
- Go プロジェクトに組み込みたい: [24行のコードで自分の AI Agent を作る](/ja/guide/build-your-own-agent)
- 長時間動かす入口を理解したい: [Runtime モード](/ja/guide/runtime-modes)
- Subagent への委譲を理解したい: [Subagents](/ja/guide/subagents)
- 本番前のガバナンスを確認したい: [セキュリティと Guard](/ja/guide/security-and-guard)
