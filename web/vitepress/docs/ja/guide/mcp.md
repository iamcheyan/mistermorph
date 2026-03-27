---
title: MCP
description: MCP サーバー設定と、リモートツールのローカル統合。
---

# MCP

Mister Morph は MCP サーバーへ接続し、リモートツールを同じ tool-calling ループに統合できます。

## ツール名マッピング

MCP ツールは次の名前で登録されます。

- `mcp_<server_name>__<tool_name>`

例: `mcp_filesystem__read_file`

## 対応トランスポート

- `stdio`（デフォルト）
- `http`

## 設定形式

```yaml
mcp:
  servers:
    - name: filesystem
      type: stdio
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
      allowed_tools: []

    - name: remote
      type: http
      url: "https://mcp.example.com/mcp"
      headers:
        Authorization: "Bearer ${MCP_REMOTE_TOKEN}"
      allowed_tools: ["search"]
```

フィールド挙動:

- `enable: false` でそのサーバーを無効化
- `allowed_tools: []` はそのサーバーの全ツールを許可
- 不正な設定は warning を出してスキップ

## ライフサイクル

1. `mcp.servers` を読む
2. 有効かつ正しいサーバーへ接続
3. ツール一覧を取得
4. ローカル registry へアダプト登録
5. 終了時に MCP session を close

## 障害モデル

- サーバー単位で分離（1台失敗しても他は継続）
- MCP が0件でも通常ツールで runtime は動作

## セキュリティ

- `allowed_tools` で最小権限化
- 認証ヘッダーは `${ENV_VAR}` で管理
- guard/network 制御と併用する
