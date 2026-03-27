---
title: Memory
description: WAL ベースの memory アーキテクチャ、注入、書き戻しルール。
---

# Memory

Mister Morph の memory は「WAL 先行書き込み + 非同期投影」です。

## アーキテクチャ

- 真のデータ源: `memory/log/*.jsonl`（WAL）
- 読み取りモデル:
  - `memory/index.md`（長期）
  - `memory/YYYY-MM-DD/*.md`（短期）
- ホットパスは WAL のみ書き込み。Markdown 更新は非同期。

## 実行フロー

1. LLM 呼び出し前に snapshot を作成し prompt block へ注入
2. 最終返信後に raw memory event を WAL へ追記
3. projection worker が WAL を再生して markdown を更新

## 注入条件

すべて満たすときのみ注入:

- `memory.enabled = true`
- `memory.injection.enabled = true`
- 有効な `subject_id` がある
- snapshot が空でない

`memory.injection.max_items` で注入件数を制限します。

## 書き戻し条件

すべて満たすときのみ書き戻し:

- 最終返信が実際に publish される
- memory orchestrator が存在
- `subject_id` が存在

軽量返信や空出力では書き戻しをスキップします。

## 主要設定

```yaml
memory:
  enabled: true
  dir_name: "memory"
  short_term_days: 7
  injection:
    enabled: true
    max_items: 50
```

## 運用メモ

- 投影 markdown が壊れても WAL から再構築可能
- 本番では `file_state_dir` を永続ストレージに置く
