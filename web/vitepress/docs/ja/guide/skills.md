---
title: Skills
description: Skill の発見、読み込み戦略、ランタイム挙動。
---

# Skills

Skill は `SKILL.md` を中心としたローカル指示パックです。

## 発見パス

Mistermorph は `file_state_dir/skills`（デフォルトは `~/.morph/skills`）配下を再帰的に走査し、`SKILL.md` を見つけて skill を発見します。

## 読み込み制御

設定ファイルでは次のように skill の読み込みを制御できます。

```yaml
skills:
  enabled: true
  dir_name: "skills"
  load: []
```

ここでの意味は次の通りです。

- `enabled`：skill を読み込むかどうか
- `load`：読み込む skill の id/name を指定します。たとえば `["apple", "banana"]` ならその 2 つだけを読み込みます。空なら発見した skill をすべて読み込みます。

タスク本文の `$skill-name` / `$skill-id` でもトリガーできます。

## Skill の注入

system prompt に入るのは skill のメタ情報だけです。

- `name`
- `file_path`
- `description`
- `auth_profiles`（任意）

実際の `SKILL.md` 本文は、必要になったときにモデルが `read_file` で読み込みます。

## よく使うコマンド

```bash
# 現在発見できる skill を表示します。インストール結果の確認や、
# skill ディレクトリが正しく認識されているかの確認に便利です。
mistermorph skills list
# 組み込み skill をローカルの skills ディレクトリへインストールまたは
# 更新します。デフォルトの出力先は ~/.morph/skills です。
mistermorph skills install
# リモートの SKILL.md から単一の skill をインストールします。
# 外部 skill を取り込むときに使います。
mistermorph skills install "https://example.com/SKILL.md"
```

よく使う付随フラグ:

- `--skills-dir`：`skills list` に追加の走査ルートを渡す
- `--dry-run`：`skills install` が何を書き込むかだけを確認し、実際には書き込まない
- `--dest`：テストや分離環境向けに、指定ディレクトリへ skill をインストールする
- `--clean`：完全な上書き更新のため、インストール前に既存の skills ディレクトリを削除する

## セキュリティ機構

Mistermorph の skill 安全設計は 2 段階です。

1. インストール段階では、リモートファイルを取ってすぐ実行してしまうことを防ぎます。
2. 実行段階では、skill やモデルが秘密情報を直接取得することを防ぎます。

基本方針は、skill は手順や文脈を提供できても、自力で権限を広げることはできない、ということです。

### インストール: 先にレビューし、その後に書き込む

リモート skill のインストールは、単純にローカルへダウンロードするだけではありません。確認付きのフローになっています。

```text
+--------------------------------------+
  リモート SKILL.md
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  内容を表示して確認
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  信頼しない入力としてレビュー
  宣言された追加ファイルを抽出 
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  書き込み計画と潜在リスクを表示 
  再確認 
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  ~/.morph/skills/<name>/ に書き込む
+--------------------------------------+
```

> インストーラが何をするかだけ見たい場合は、先に `--dry-run` を使ってください。

### 実行段階: skill は profile を宣言するだけで、秘密そのものは持たない

skill が保護された HTTP API にアクセスする必要がある場合、秘密を skill に書くのではなく、`auth_profile` による設定注入を使うのが推奨です。

たとえば skill は、使いたい認証プロファイルを次のように宣言できます。

```yaml
auth_profiles: ["jsonbill"]
```

ただし、これだけで権限が与えられるわけではありません。実際の認可境界は設定で決まります。

```yaml
secrets:
  allow_profiles: ["jsonbill"]

auth_profiles:
  jsonbill:
    credential:
      kind: api_key
      secret: "${JSONBILL_API_KEY}"
    allow:
      url_prefixes: ["https://api.jsonbill.com/tasks"]
      methods: ["GET", "POST"]
      follow_redirects: false
      deny_private_ips: true
```

この構成では、skill と LLM が見えるのは `jsonbill` という profile id だけで、`JSONBILL_API_KEY` の実値は直接見えません。

Mistermorph は設定読み込み時に環境変数から実際の秘密を解決し、それを `url_fetch` に注入します。これにより、API キーが prompt、`SKILL.md`、ツール引数、ログに露出するのを避けられます。

また、`auth_profiles` ではアクセス境界も定義できます。たとえば `url_prefixes` で到達可能な URL プレフィックスを制限し、`methods`、`follow_redirects`、`deny_private_ips` で振る舞いをさらに絞り込めます。
