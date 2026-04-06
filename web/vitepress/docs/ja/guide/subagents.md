---
title: Subagent と子タスク
description: "`spawn` を使う場面、`bash.run_in_subtask` を使う場面、そして runtime が何を保証するかをまとめる。"
---

# Subagent と子タスク

Mistermorph には現在、子タスク境界を明示的に作る方法が二つあります。

- `spawn`: 独立した LLM loop を持つ子 agent を起動し、使えるツールを明示的に制限する。
- `bash.run_in_subtask=true`: 1 本の shell コマンドを direct subtask 境界で実行し、2 回目の LLM loop は起動しない。

どちらも最後は同じ `SubtaskResult` envelope を返し、深さ制限も共有します。

## どちらを使うか

- 子タスク側でもツール推論が必要なら `spawn` を使います。たとえば `url_fetch` のあとに抽出する、`read_file` と `bash` を組み合わせる、といった流れです。
- 仕事の中身がすでに 1 本の shell コマンドなら `bash.run_in_subtask=true` を使います。
- 親がそのまま実行できる小さな 1 ステップ作業なら、わざわざ子タスクに分けない方がよいです。

現状も重要です。どちらも同期実行で、親は子タスク終了まで待ちます。解決しているのは隔離と結果の収束であり、バックグラウンド実行ではありません。

## `spawn`

`spawn` は engine スコープのツールです。1 回の agent engine が組み上がったときにだけ現れます。

引数:

- `task`: 必須。子タスクのプロンプト。
- `tools`: 必須。子タスクが使えるツール名の非空配列。
- `model`: 任意。子タスク用モデル上書き。省略時は親モデルを使います。
- `output_schema`: 任意。構造化出力の契約名。
- `observe_profile`: 任意。観測ヒント。現在は `default`、`long_shell`、`web_extract` をサポートします。

実行時の挙動:

- 子タスク registry は渡した `tools` だけから作られます。未知の名前や親 registry に存在しない名前は無視されます。
- 最終的に使えるツールが 1 つも残らなければ呼び出しは失敗します。
- `tools` に `spawn` 自身を入れても、子タスクには再公開されません。
- 現在の深さ制限は `1` です。つまり子タスクの中からさらに子タスクは起動できません。
- 子タスクの生 transcript はデフォルトで親 loop に戻しません。

### `output_schema` とは何か

`output_schema` は構造化出力の識別子であって、組み込み JSON Schema レジストリではありません。

これを指定すると:

- 子タスクには JSON の最終出力を求めます。
- runtime は最終出力が JSON、または JSON として解釈できる文字列であることを要求します。
- 返り値の envelope では同じ識別子が `output_schema` にそのまま入ります。

ただし Mistermorph 自体が実在する schema 定義でオブジェクト検証までは行いません。

## 返り値 Envelope

`spawn` と direct subtask はどちらも次の形の JSON を返します。

```json
{
  "task_id": "sub_123",
  "status": "done",
  "summary": "subtask completed",
  "output_kind": "text",
  "output_schema": "",
  "output": "child result",
  "error": ""
}
```

各フィールド:

- `status`: 現状は主に `done` または `failed`。
- `summary`: 親側の進捗表示や短い要約に使う短文。
- `output_kind`: `text` または `json`。
- `output_schema`: テキスト出力なら空、構造化出力なら渡した識別子。
- `output`: 子タスクの本体結果。
- `error`: 失敗時だけ非空になります。

## `bash.run_in_subtask=true`

こちらはより軽い子タスク経路です。

- subtask runner を直接使い、2 回目の LLM loop は起動しません。
- `output_schema` は `subtask.bash.result.v1` に固定です。
- 観測 profile は `long_shell` に固定です。
- `output` には `exit_code`、切り捨てフラグ、`stdout`、`stderr` が入ります。

payload 例:

```json
{
  "task_id": "sub_456",
  "status": "done",
  "summary": "bash exited with code 0",
  "output_kind": "json",
  "output_schema": "subtask.bash.result.v1",
  "output": {
    "exit_code": 0,
    "stdout_truncated": false,
    "stderr_truncated": false,
    "stdout": "hello\n",
    "stderr": ""
  },
  "error": ""
}
```

長い shell 実行を独立 envelope で包みたいが、子タスク側で追加のツール判断までは不要、という場面で使います。

## 設定と組み込み

- `tools.spawn.enabled` が制御するのは明示的な `spawn` ツール入口だけです。
- `tools.spawn.enabled=false` でも、`bash.run_in_subtask=true` のような direct subtask は subtask runtime を使います。
- `integration.Config.BuiltinToolNames` には `spawn` を含めることも外すこともできます。対象は静的ツールだけではありません。
- `agent.New(...)` で直接 engine を作る場合、`spawn` はデフォルトで有効です。無効化したいなら `agent.WithSpawnToolEnabled(false)` を使います。

例:

```go
cfg := integration.DefaultConfig()
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "spawn"}
cfg.Set("tools.spawn.enabled", true)
```

`BuiltinToolNames` から `spawn` を外すと、agent 表面からは明示的な child-agent ツールが消えます。ただし基盤の subtask runtime 自体は残るため、`bash.run_in_subtask=true` のような内部入口は引き続き利用できます。

あわせて読む:

- [組み込みツール](/ja/guide/built-in-tools)
- [自分の AI Agent を作る：上級編](/ja/guide/build-your-own-agent-advanced)
- [設定フィールド](/ja/guide/config-reference)
