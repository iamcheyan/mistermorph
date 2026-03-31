---
title: LLM ルーティングポリシー
description: runtime purpose ごとに profile、候補分流、fallback チェーンを選ぶ。
---

# LLM ルーティングポリシー

`llm.routes.*` を使うと、runtime purpose ごとに異なるモデル設定を割り当てられます。

この設定は第一方 runtime と `integration.Config` の両方で共通なので、CLI / Console / Channel / Go 組み込みで同じルーティング意味論を使います。

## routes が必要になる場面

典型的なケース:

- メインループはデフォルトモデルのままにして移行コストを下げたい。
- `addressing` は安くて速いモデルにしたい。
- `plan_create` は強い推論モデルに固定したい。
- `main_loop` で分流しつつ、失敗時の fallback チェーンも持たせたい。

## 対応している purpose

- `main_loop`: Agent のメイン step loop。
- `addressing`: グループやチャンネルでの addressing 判定。
- `heartbeat`: 定期 heartbeat タスク。
- `plan_create`: `plan_create` ツール内部の planning リクエスト。
- `memory_draft`: memory 草稿の整理。

## 最小構成

```yaml
llm:
  provider: openai
  model: gpt-5.4
  api_key: ${OPENAI_API_KEY}

  profiles:
    cheap:
      model: gpt-4.1-mini
    reasoning:
      provider: xai
      model: grok-4.1-fast-reasoning
      api_key: ${XAI_API_KEY}

  routes:
    plan_create: reasoning
    addressing: cheap
```

この意味は次の通りです:

- デフォルトのメインループは引き続きトップレベルの `llm.*` を使う。
- `plan_create` は `reasoning` に固定される。
- `addressing` は `cheap` に固定される。

## 3つの書き方

### 1. profile 名を直接書く

最短形です:

```yaml
llm:
  routes:
    heartbeat: cheap
```

これは、その purpose を 1 つの profile に直接バインドする形です。

### 2. 明示オブジェクト

ローカルな fallback チェーンも持たせたい場合はオブジェクトで書けます:

```yaml
llm:
  routes:
    plan_create:
      profile: reasoning
      fallback_profiles: [default]
```

ルール:

- `profile` はメインルートの profile。
- `fallback_profiles` はその route 専用の fallback チェーン。

### 3. 候補分流

同じ purpose に対してトラフィック分流したい場合は `candidates` を使います:

```yaml
llm:
  routes:
    main_loop:
      candidates:
        - profile: default
          weight: 1
        - profile: cheap
          weight: 1
      fallback_profiles: [reasoning]
```

ルール:

- `weight` は候補選択の重みを決めます。
- 1 回の run ごとに 1 つの主候補を選び、その run 内で再利用します。
- 主候補が fallback 可能なエラーに当たった場合は、同じ route 内の残り candidate を先に試し、その後 `fallback_profiles` を順に試します。

## integration ではどう書くか

Go から組み込む場合も設定の考え方は同じで、YAML の代わりに `cfg.Set(...)` で書きます:

```go
cfg := integration.DefaultConfig()
cfg.Set("llm.routes.plan_create", "reasoning")
cfg.Set("llm.routes.addressing", map[string]any{
  "profile": "cheap",
  "fallback_profiles": []string{"default"},
})
```

全フィールド名を見たい場合は [設定フィールド](/ja/guide/config-reference)、よくある YAML パターンを見たい場合は [設定パターン](/ja/guide/config-patterns) を参照してください。
