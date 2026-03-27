---
title: Memory
description: WAL-based memory architecture, injection, and writeback behavior.
---

# Memory

Mister Morph memory is append-first and rebuildable.

## Architecture

- Source of truth: `memory/log/*.jsonl` (WAL)
- Read model: markdown projections
  - `memory/index.md` (long-term)
  - `memory/YYYY-MM-DD/*.md` (short-term)
- Projection runs asynchronously; hot path writes only WAL.

## Runtime Flow

1. Before LLM call: runtime prepares memory snapshot and injects it into prompt blocks.
2. After final reply: runtime records raw memory event to WAL.
3. Projection worker replays WAL and updates markdown files.

## Injection Rules

Memory injection runs only when all conditions match:

- `memory.enabled = true`
- `memory.injection.enabled = true`
- runtime provides a valid `subject_id`
- snapshot content is not empty

`memory.injection.max_items` limits injected summary items.

## Writeback Rules

Writeback runs only when:

- final reply is actually published
- memory orchestrator exists
- `subject_id` exists

If final output is empty/lightweight, writeback may be skipped.

## Config Keys

```yaml
memory:
  enabled: true
  dir_name: "memory"
  short_term_days: 7
  injection:
    enabled: true
    max_items: 50
```

## Operational Notes

- If markdown projections are damaged, rebuild from WAL.
- Treat `memory/log/*.jsonl` as durable data.
- Keep `file_state_dir` on persistent storage in production.
