---
title: 内置工具
description: 静态工具、运行时注入工具与通道专属工具。
---

# 内置工具

工具分两阶段注册：静态注册 + 运行时注入。

## 静态工具（配置驱动）

无需 runtime 上下文即可创建：

- `read_file`
- `write_file`
- `bash`
- `url_fetch`
- `web_search`
- `contacts_send`

## 运行时工具（依赖 LLM 上下文）

运行时会注入：

- `plan_create`
- `todo_update`

## 通道专属工具

- Telegram 可能注入：
  - `telegram_send_voice`
  - `telegram_send_photo`
  - `telegram_send_file`
  - `message_react`
- Slack 可能注入：
  - `message_react`

## Core 嵌入里的白名单

可通过 `integration.Config.BuiltinToolNames` 限制内置工具。

```go
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "todo_update"}
```

为空表示启用全部内置工具。

## 配置入口

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

参数级细节见 `docs/tools.md`。
