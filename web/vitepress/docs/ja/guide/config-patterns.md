---
title: 設定パターン
description: インストール方法、基本設定、profiles、routes、ツールポリシーの代表例。
---

# 設定パターン

## インストール方法

```bash
# リリース版インストーラ
curl -fsSL -o /tmp/install-mistermorph.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install-mistermorph.sh
```

```bash
# Go からインストール
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

## 初期ファイル作成

```bash
mistermorph install
```

標準ワークスペースは `~/.morph/` です。

## 設定の優先順位

- CLI フラグ
- 環境変数
- `config.yaml`

## 最小 `config.yaml`

```yaml
llm:
  provider: openai
  model: gpt-5.4
  endpoint: https://api.openai.com
  api_key: ${OPENAI_API_KEY}
```

## LLM Profiles と Routes

```yaml
llm:
  model: gpt-5.4
  profiles:
    cheap:
      model: gpt-4.1-mini
    backup:
      provider: xai
      model: grok-4.1-fast-reasoning
  routes:
    main_loop:
      candidates:
        - profile: default
          weight: 1
        - profile: cheap
          weight: 1
      fallback_profiles: [backup]
    addressing:
      profile: cheap
    heartbeat: cheap
```

## ツールの有効/無効

```yaml
tools:
  bash:
    enabled: false
  url_fetch:
    enabled: true
    timeout: "30s"
```

## 実行上限

```yaml
max_steps: 20
tool_repeat_limit: 4
```

キーの完全一覧は `assets/config/config.example.yaml` を参照。
