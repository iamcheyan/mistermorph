---
title: MCP
description: 配置 MCP，并接入本地 Agent 工具循环。
---

# MCP

Mister Morph 可连接 MCP，并把 MCP 提供的工具注册到同一个 tool-calling 循环中。

## 工具名映射

MCP 工具会被注册为：`mcp_<server_name>__<tool_name>`

例如：`mcp_example__read_file`

## 支持的传输类型

- `stdio`（默认）
- `http`

## 配置

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

其中：

- `enable`: 可禁用或者启用某个 server
- `allowed_tools`：表示该 server 可以使用的其他工具，留空为不限制

## 生命周期

1. 读取 `mcp.servers`
2. 连接每个启用且合法的 server
3. 拉取工具列表
4. 适配并注册进本地 registry
5. 关闭时清理 MCP session
