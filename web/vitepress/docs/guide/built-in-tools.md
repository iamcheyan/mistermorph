---
title: Built-in Tools
description: Static tools, runtime-injected tools, and channel-specific tools.
---

# Built-in Tools

Tools are registered in two phases: static + runtime.

## Static Tools (config-driven)

These are created from config without runtime context.

- `read_file`: read local files
- `write_file`: write local files under allowed dirs
- `bash`: execute local shell commands
- `url_fetch`: HTTP(S) fetch/download
- `web_search`: web search
- `contacts_send`: send outbound message by contact

## Runtime Tools (LLM runtime dependent)

These are injected when runtime has active LLM client/model.

- `plan_create`
- `todo_update`

## Channel-Specific Runtime Tools

- Telegram runtime can add:
  - `telegram_send_voice`
  - `telegram_send_photo`
  - `telegram_send_file`
  - `message_react`
- Slack runtime can add:
  - `message_react`

## Tool Selection in Core Embedding

You can whitelist built-ins via `integration.Config.BuiltinToolNames`.

```go
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
```

Empty list means all known built-ins.

## Key Config Sections

```yaml
tools:
  read_file: ...
  write_file: ...
  bash: ...
  url_fetch: ...
  web_search: ...
  contacts_send: ...
  todo_update: ...
  plan_create: ...
```

See `docs/tools.md` for deeper parameter-level behavior.
