---
title: MCP
description: Configure MCP and expose MCP tools in the local agent loop.
---

# MCP

Mister Morph can connect to MCP and register MCP-provided tools into the same tool-calling loop.

## Tool Name Mapping

MCP tools are registered as: `mcp_<server_name>__<tool_name>`

Example: `mcp_example__read_file`

## Supported Transports

- `stdio` (default)
- `http`

## Configuration

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

Where:

- `enable`: enables or disables a server entry
- `allowed_tools`: limits which tools from that server are usable; leave it empty for no restriction

## Lifecycle

1. Runtime reads `mcp.servers`.
2. Connect to each enabled/valid server.
3. List server tools.
4. Adapt and register tools into local registry.
5. On shutdown, close MCP sessions.
