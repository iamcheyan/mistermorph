---
title: Runtime 模式
description: 选择 CLI、机器人或控制台运行方式。
---

# Runtime 模式

## 一次性任务

```bash
mistermorph run --task "..."
```

## Telegram Bot

```bash
mistermorph telegram --log-level info
```

## Slack Bot

```bash
mistermorph slack --log-level info
```

## Console 后端

```bash
mistermorph console serve
```

## 旧版守护进程（按需）

```bash
mistermorph serve --server-listen 127.0.0.1:8787
```

详细接入见：

- `docs/console.md`
- `docs/slack.md`
- `docs/line.md`
- `docs/lark.md`
