---
title: Memory
description: Introduces the WAL-based memory architecture, projection, and injection rules.
---

# Memory

Mister Morph memory uses an append-first WAL (Write-ahead logging) model plus asynchronous projection.

## WAL

In plain terms, anything that happens is first written as raw data to a jsonl file in order. That is why WAL is the real source of truth.

The path is `memory/log/`, and file names follow the format `since-YYYY-MM-DD-0001.jsonl`.

When a WAL file reaches a certain size, it is rotated into a file ending in `.jsonl.gz`.

## Projection

Mister Morph builds simple projections from WAL into two targets:

- `memory/index.md` (long-term memory)
- `memory/YYYY-MM-DD/*.md` (short-term memory)
  - Short-term memory files are isolated by channel. For example, memories from different Telegram group chats do not mix.

Projection means reading WAL for a period of time, summarizing it with an LLM, and writing the result into those target files.

The projection checkpoint file is `memory/log/checkpoint.json`, and it looks roughly like this:

```json
{
  "file": "since-2026-02-28-0001.jsonl",
  "line": 18,
  "updated_at": "2026-02-28T06:30:12Z"
}
```

## Injection

When the following config is enabled, part of the projected memory is injected into the prompt:

- `memory.enabled = true`
- `memory.injection.enabled = true`

`memory.injection.max_items` controls the maximum number of injected items.

## Notes

1. If memory projections are damaged, rebuild them from WAL by deleting `memory/log/checkpoint.json` and letting the Agent continue running.
2. In production, keep `file_state_dir` on persistent storage so runtime state, including memory, survives.
