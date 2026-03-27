---
title: Skills
description: Skill の発見、読み込み戦略、ランタイム挙動。
---

# Skills

Skill は `SKILL.md` を中心としたローカル指示パックです。

## Skill とは

Skill はツールではなく prompt コンテキストです。

- Skill: どう進めるかを定義
- Tool: 実際の操作を実行（`read_file`、`url_fetch`、`bash` など）

## 発見パス

デフォルトのルート:

- `file_state_dir/skills`（通常 `~/.morph/skills`）

runtime は `SKILL.md` を再帰的に探索します。

## 読み込み制御

```yaml
skills:
  enabled: true
  dir_name: "skills"
  load: []
```

- `enabled: false`: 読み込まない
- `load: []`: 発見した skill をすべて読み込む
- `load: ["a", "b"]`: 指定 skill のみ読み込む
- 不明な項目は無視される

タスク本文の `$skill-name` / `$skill-id` でもトリガーできます。

## 注入モデル

system prompt に入るのは skill メタ情報のみ:

- `name`
- `file_path`
- `description`
- `auth_profiles`（任意）

実体の `SKILL.md` は必要時に `read_file` で読みます。

## 主要コマンド

```bash
mistermorph skills list
mistermorph skills install
mistermorph skills install "https://example.com/SKILL.md"
```

## セーフティ

- リモート skill は保存前にレビュー/確認を行う
- インストール中にスクリプトは実行されない
