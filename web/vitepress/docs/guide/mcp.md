---
title: MCP
description: Configure MCP servers and expose remote tools as local agent tools.
---

# MCP

Mister Morph can connect to MCP servers and expose their tools in the same tool-calling loop.

## Tool Name Mapping

MCP tools are registered as:

- `mcp_<server_name>__<tool_name>`

Example:

- `mcp_filesystem__read_file`

## Supported Transports

- `stdio` (default)
- `http`

## Config Shape

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

Field behavior:

- `enable: false` disables a server entry
- `allowed_tools: []` means all tools from that server
- invalid server config is skipped with warning

## Lifecycle

1. Runtime reads `mcp.servers`.
2. Connect to each enabled/valid server.
3. List server tools.
4. Adapt and register tools into local registry.
5. On shutdown, close MCP sessions.

## Failure Model

- Per-server failures are isolated.
- Other servers and built-in tools keep working.
- If no server connects, runtime still runs without MCP tools.

## Security Notes

- Prefer `allowed_tools` to reduce blast radius.
- Keep auth headers in env vars (`${ENV_VAR}`), not plain text.
- Combine with guard/network policy for outbound controls.
