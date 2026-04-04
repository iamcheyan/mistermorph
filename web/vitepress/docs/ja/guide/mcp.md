---
title: MCP
description: MCP を設定し、ローカル Agent のツールループへ接続する。
---

# MCP

Mister Morph は MCP へ接続し、MCP が提供するツールを同じ tool-calling ループへ登録できます。

## ツール名マッピング

MCP ツールは次の名前で登録されます: `mcp_<server_name>__<tool_name>`

例: `mcp_example__read_file`

## 対応トランスポート

- `stdio`（デフォルト）
- `http`

## 設定

```yaml
mcp:
  servers:
    - name: example_cmd
      type: stdio
      command: npx
      args: ["-y", "@modelcontextprotocol/example_cmd", "/tmp"]
      allowed_tools: []

    - name: remote
      type: http
      url: "https://mcp.example.com/mcp"
      headers:
        Authorization: "Bearer ${MCP_REMOTE_TOKEN}"
      allowed_tools: ["search"]
```

内容:

- `enable`: その server を有効または無効にする
- `allowed_tools`: その server で利用可能なツールを制限する。空なら制限しない

## ライフサイクル

1. `mcp.servers` を読む
2. 有効かつ正しいサーバーへ接続
3. ツール一覧を取得
4. ローカル registry へアダプト登録
5. 終了時に MCP session を close
