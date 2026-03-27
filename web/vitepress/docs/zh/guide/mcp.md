---
title: MCP
description: 配置 MCP server，并把远端工具接入本地 Agent 工具循环。
---

# MCP

Mister Morph 可连接 MCP server，并把远端工具注册到同一个 tool-calling 循环中。

## 工具名映射

MCP 工具会被注册为：

- `mcp_<server_name>__<tool_name>`

例如：`mcp_filesystem__read_file`

## 支持的传输类型

- `stdio`（默认）
- `http`

## 配置结构

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

字段行为：

- `enable: false` 可禁用某个 server
- `allowed_tools: []` 表示该 server 的全部工具都可用
- 配置非法的 server 会被跳过并记 warning

## 生命周期

1. 读取 `mcp.servers`
2. 连接每个启用且合法的 server
3. 拉取工具列表
4. 适配并注册进本地 registry
5. 关闭时清理 MCP session

## 故障模型

- 单个 server 失败不会拖垮整体
- 其他 server 和内置工具可继续工作
- 全部连接失败时仅是不加载 MCP 工具

## 安全建议

- 用 `allowed_tools` 做最小权限
- header token 放 `${ENV_VAR}`，不要明文
- 配合 guard/network 策略做外联控制
