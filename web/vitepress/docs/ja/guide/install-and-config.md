---
title: インストールと設定
description: 導入方法と基本設定モデル。
---

# インストールと設定

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

全キーは `assets/config/config.example.yaml` を基準にしてください。
