---
title: Runtime Modes Overview
description: See which runtime modes Mister Morph supports.
---

# Runtime Modes Overview

## One-shot task

If you only need to call Mister Morph from the command line for a single task, use this mode.

```bash
mistermorph run --task "..."
```

## Console

This provides a full Web UI. Besides interacting with the agent directly, you can also use it to monitor other Mister Morph instances.

```bash
mistermorph console serve
```

## Telegram Bot

Run a standalone runtime connected to a Telegram channel and interact with Mister Morph inside Telegram.

```bash
mistermorph telegram --log-level info
```

## Slack Bot

Similar to the Telegram mode, but provides the interaction inside Slack.

```bash
mistermorph slack --log-level info
```
