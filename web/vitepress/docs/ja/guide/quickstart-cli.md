---
title: クイックスタート（CLI）
description: 数分で CLI を実行可能にする最短手順。
---

# クイックスタート（CLI）

## 1. インストール

```bash
curl -fsSL -o /tmp/install.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install.sh
```

## 2. 初期化

```bash
mistermorph install
```

Mister Morph は必要なファイルを初期化します。デフォルトでは `~/.morph/` に配置され、設定ファイルは `~/.morph/config.yaml` です。

初期化中には、LLM の設定、Agent 名、persona など、最小限の必須項目を対話形式で確認します。

### 2.1 任意: 環境変数で設定する

より強いセキュリティが必要な環境では、Mister Morph は機密情報を環境変数に置き、設定ファイルから参照できます。

たとえば、LLM の API key を環境変数に入れられます。

```bash
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
```

その後、設定ファイルから参照します。

```yaml
llm:
  api_key: "${MISTER_MORPH_LLM_API_KEY}"
```

## 3. 最初のタスクを実行

```bash
mistermorph run --task "Hello"
```

出力例:

```json
{
  "reasoning": "Greet the user briefly.",
  "output": "Hello 👀",
  "reaction": "👀"
}
```

## 4. デバッグスイッチ

```bash
mistermorph run --inspect-prompt --inspect-request --task "Hello"
```

このとき、現在のディレクトリに `dump` が生成され、prompt と request の詳細を確認できます。

設定全体は [設定パターン](/ja/guide/config-patterns) と `assets/config/config.example.yaml` を参照してください。
