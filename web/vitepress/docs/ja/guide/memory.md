---
title: Memory
description: WAL ベースの memory アーキテクチャ、投影、注入ルールを説明します。
---

# Memory

Mister Morph の memory システムは、先に WAL（Write-ahead logging）へ追記し、その後で非同期投影する仕組みです。

## WAL

平たく言えば、起きたことは大小を問わず jsonl 形式で発生順に生データとして記録されます。だから WAL が本当のデータ源になります。

パスは `memory/log/` で、ファイル名は `since-YYYY-MM-DD-0001.jsonl` の形式です。

WAL ファイルが一定サイズに達すると、`.jsonl.gz` で終わるファイルへローテーションされます。

## 投影

Mister Morph は WAL から次の 2 つへ単純な投影を行います。

- `memory/index.md`（長期記憶）
- `memory/YYYY-MM-DD/*.md`（短期記憶）
  - 短期記憶ファイルは Channel ごとに分離されます。例えば Telegram では別々のグループチャット間で記憶は共有されません。

投影とは、一定期間の WAL を読み、LLM で要約し、対応する対象ファイルへ書き出すことです。

投影の記録点ファイルは `memory/log/checkpoint.json` で、中身はおおよそ次のようになります。

```json
{
  "file": "since-2026-02-28-0001.jsonl",
  "line": 18,
  "updated_at": "2026-02-28T06:30:12Z"
}
```

## 注入

次の設定が有効なとき、一部の記憶投影が prompt に注入されます。

- `memory.enabled = true`
- `memory.injection.enabled = true`

`memory.injection.max_items` は注入する項目数の上限を制御します。

## 備考

1. Memory 投影が壊れても WAL から再構築できます。`memory/log/checkpoint.json` を削除して Agent を継続実行すればよいです。
2. 本番環境では、memory を含む実行状態を維持するために `file_state_dir` を永続ストレージへ置いてください。
