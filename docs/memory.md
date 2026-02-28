# Memory System (Core + Current Wiring)

This document describes how memory currently works in `mistermorph` as implemented today.

## 1. Scope and Status

- Core memory subsystem is channel-agnostic and lives in `memory/*`.
- Memory storage is markdown-file based under `memory/`.
- Runtime-level memory wiring is currently implemented in Telegram runtime.
- Slack runtime does not currently run this memory injection/writeback pipeline.

## 2. Core Components

- Memory core (channel-agnostic):
  - `memory/manager.go`, `memory/update.go`, `memory/merge.go`, `memory/inject.go`, `memory/identity.go`
- LLM semantic helpers used by memory merge:
  - `internal/entryutil/semantic_llm.go`
- Current Telegram runtime adapter:
  - `internal/channelruntime/telegram/runtime_task.go`
  - `internal/channelruntime/telegram/memory_flow.go`
  - `internal/channelruntime/telegram/memory_prompts.go`
  - `internal/channelruntime/telegram/prompts/memory_draft_system.md`
  - `internal/channelruntime/telegram/prompts/memory_draft_user.md`

## 3. ASCII Architecture (Memory Path)

```text
                 +------------------------------------+
                 | Runtime Memory Adapter             |
                 | (channel-specific wiring)          |
                 +----------------+-------------------+
                                  |
         +------------------------+------------------------+
         |                                                 |
+--------v---------+                             +---------v----------+
| Injection Path   |                             | Writeback Path     |
| BuildInjection   |                             | updateMemoryFromJob|
+--------+---------+                             +---------+----------+
         |                                                 |
         |                                      +----------v----------+
         |                                      | BuildMemoryDraft    |
         |                                      | (LLM: memory.draft) |
         |                                      +----------+----------+
         |                                                 |
         |                            +--------------------v-------------------+
         |                            | existing short-term && draft non-empty |
         |                            +----------+------------------------------+
         |                                       |
         |                    +------------------+------------------+
         |                    |                                     |
         |          +---------v---------+                 +---------v----------+
         |          | MergeShortTerm    |                 | SemanticMergeShort |
         |          |                   |                 | (LLM semantic dedup)|
         |          +---------+---------+                 +---------+----------+
         |                    |                                     |
         +--------------------+--------------------+----------------+
                                              |
                                  +-----------v------------+
                                  | memory.Manager         |
                                  | WriteShortTerm         |
                                  | UpdateLongTerm         |
                                  +-----------+------------+
                                              |
                                  +-----------v------------+
                                  | Markdown Store         |
                                  | memory/index.md        |
                                  | memory/YYYY-MM-DD/*.md |
                                  +------------------------+
```

## 4. ASCII Runtime Flows (Current Telegram Wiring)

### 4.1 Main Skeleton

```text
runtime task
  -> injection phase
  -> agent.Engine.Run
  -> writeback gate (publishText + identity)
  -> writeback phase (if gate passed)
```

### 4.2 Injection Flow

```text
Runtime(T)              MemoryMgr(MM)                      FS(memory/*.md)
    |                        |                                     |
    | NewManager(...)        |                                     |
    |----------------------->|                                     |
    | BuildInjection(...)    |                                     |
    |----------------------->|                                     |
    |                        | read index.md + recent short-term   |
    |                        |------------------------------------->|
    | snapshot text          |                                     |
    |<-----------------------|                                     |
    | append memory block into prompt                              |
```

### 4.3 Writeback Flow

```text
Runtime(T)              MemoryMgr(MM)            LLM                   FS(memory/*.md)
    |                        |                    |                           |
    | [gate] publishText && identity available ?                            |
    | LoadShortTerm(...)     |                    |                           |
    |----------------------->|                    |                           |
    | BuildMemoryDraft(...)  |                    |                           |
    |-------------------------------------------->|                           |
    | draft(summary/promote) |                    |                           |
    |<--------------------------------------------|                           |
    | merge stage (see 4.4)  |                    |                           |
    | WriteShortTerm(...)    |                    |                           |
    |----------------------->|                    |                           |
    |                        | write YYYY-MM-DD/session.md                    |
    |                        |-----------------------------------------------> |
    | UpdateLongTerm(...)    |                    |                           |
    |----------------------->|                    |                           |
    |                        | write index.md                                |
    |                        |-----------------------------------------------> |
```

### 4.4 Semantic Dedupe Subflow (Merge Stage)

```text
Condition:
  has existing short-term content && draft has summary_items

existing + incoming
  -> SemanticMergeShortTerm
  -> LLM semantic keep_indices
  -> deduped newest-first summary list

Else:
  -> MergeShortTerm (direct newest-first merge, no semantic dedupe)
```

```text
Runtime(T)              MemoryMgr(MM)            LLM                   FS(memory/*.md)
    |                        |                    |                           |
    | SemanticMergeShortTerm |                    |                           |
    |-------------------------------------------->|                           |
    | keep_indices           |                    |                           |
    |<--------------------------------------------|                           |
    | apply deduped list     |                    |                           |
    | WriteShortTerm(...)    |                    |                           |
    |----------------------->|                    |                           |
    |                        | write YYYY-MM-DD/session.md                    |
    |                        |-----------------------------------------------> |
```

Notes:

- Flows above show current Telegram wiring; `memory.Manager` and markdown formats remain channel-agnostic.
- New runtimes can reuse the same memory core by providing their own identity mapping and adapter calls.

## 5. Injection Behavior

- Identity is resolved from Telegram user id:
  - `ExternalKey = telegram:<user_id>`
  - `SubjectID = ext:telegram:<user_id>`
- Request context:
  - `private` chat -> `ContextPrivate`
  - `group/supergroup` -> `ContextPublic`
- Injection content:
  - `ContextPrivate`: long-term summary + recent short-term summaries
  - `ContextPublic`: recent short-term summaries only
- Injection is bounded by `memory.injection.max_items` (default fallback to 50 inside `BuildInjection` when value <= 0).

## 6. Writeback Conditions and Rules

- Writeback runs only when:
  - `publishText == true`
  - memory manager exists
  - long-term subject id is non-empty
- If final reply is lightweight (no text publish), memory write is skipped.
- On timeout (`context.DeadlineExceeded`), runtime schedules async retry for memory update.

## 7. Draft and Merge Rules

- Draft generation (`memory.draft`) outputs:
  - `summary_items`
  - `promote.goals_projects`
  - `promote.key_facts`
- Promote strictness:
  - Promotion is kept only when explicit “remember/store in memory” intent is detected.
  - At most one promote item is kept.
- Short-term merge strategy:
  - If existing day-session content exists and new draft is non-empty, do semantic dedupe by LLM keep-indexes.
  - Otherwise do direct newest-first merge.

## 8. Storage Layout and Data Model

Current file layout:

```text
memory/
  index.md
  YYYY-MM-DD/
    <sanitized-session-id>.md
```

Notes:

- Short-term path:
  - `memory/YYYY-MM-DD/<session_id>.md`
  - Telegram normal session id: `tg:<chat_id>`
  - Heartbeat session id: `heartbeat`
- Frontmatter stores:
  - `created_at`, `updated_at`, `summary`, `session_id`
  - `contact_id`, `contact_nickname`
- Long-term file path is currently global `memory/index.md` (current `LongTermPath(...)` ignores subject id).

## 9. `/mem` Command

- Telegram `/mem` (private chat only) loads the same injection snapshot via `BuildInjection(...)` and sends it back for inspection.
- This is a read/debug surface, not a separate storage mechanism.

## 10. Planned: Journal/WAL (First-Principles, Minimal Design)

This section captures the intended next step for memory reliability:

- `memory/log/*.jsonl` stores raw append-only memory events (source of truth).
- `memory/*.md` remains a projection/read model (rebuildable from logs).

Database analogy:

- `memory/log` = database WAL / raw change log
- `memory/index.md` + `memory/YYYY-MM-DD/*.md` = derived tables/views

### 10.1 First-Principles Goals

- Never lose accepted memory updates.
- Keep write path deterministic and auditable.
- Allow projection rebuild when markdown view is corrupted or outdated.
- Keep implementation minimal and local-first.

### 10.2 Minimal Scope (No Overdesign)

- Single-process append writer (current runtime model).
- JSONL only; one event per line.
- Local rotation only (size based).
- Replay from local logs only.

Explicit non-goals for first iteration:

- no distributed consensus / replication protocol
- no event bus dependency
- no background compaction pipeline
- no schema-registry service
- no multi-node writer arbitration

### 10.3 Proposed Layout

```text
memory/
  log/
    since-2026-02-28-0001.jsonl
    since-2026-02-28-0002.jsonl
  index.md
  YYYY-MM-DD/
    <session>.md
```

### 10.4 Write Ordering (WAL Rule)

On each accepted memory update:

1. append event to `memory/log/*.jsonl`
2. flush/sync append result
3. mark projection as dirty (enqueue async projection work)

Hot path does not block on markdown projection. If projection fails or is delayed, step 1 is still durable and can be replay-recovered.

### 10.5 Rotation and Replay

- Rotate only when file exceeds size threshold.
- Keep monotonic file naming for deterministic replay order.
- File naming carries segment start date for indexing:
  - `since-YYYY-MM-DD-0001.jsonl`
  - `since-YYYY-MM-DD-0002.jsonl`
- Store a small checkpoint (last applied log file + offset/line) for fast restart.
- On startup, replay from checkpoint to rebuild/repair markdown projections.

Compression rule (minimal):

- Active segment stays plain `.jsonl`.
- Optional compression applies only to closed old segments as `.jsonl.gz`.
- Do not bundle WAL segments into `tar.gz` in v1.

Checkpoint structure (`memory/log/checkpoint.json`):

```json
{
  "file": "since-2026-02-28-0002.jsonl",
  "line": 18,
  "updated_at": "2026-02-28T06:30:12Z"
}
```

- `file`: last applied segment logical name (always `.jsonl` key).
- `line`: last applied line number in that segment.
- `updated_at`: checkpoint write timestamp (RFC3339, UTC).

### 10.6 Event Shape (Minimal)

First version keeps only fields needed for replay and audit, for example:

- `schema_version`
- `event_id`
- `task_run_id`
- `ts_utc`
- `session_id`
- `subject_id`
- `channel`
- `participants` (array, multi-party aware, may be empty)
- `task_text`
- `final_output`
- `draft_summary_items`
- `draft_promote`

The schema should evolve conservatively.

`participants` item shape (minimal required):

- `id` (for example `@johnwick`)
- `nickname` (for example `John Wick`)
- `protocol` (for example `tg`)

`participants` may be empty (`[]`) when participant identity is unavailable at write time.

Special case for agent self:

- if `id` is `0` and `protocol` is empty (`""`), this participant means the agent itself.
- recommended `nickname`: agent display name (for example `阿嬷`).

Example event:

```json
{
  "schema_version": 1,
  "event_id": "evt_01JY7K9M7T3H2QZ6A9D5V4N8P1",
  "task_run_id": "run_01JY7K9B3W8F6M2N4C1R0T9X5Q",
  "ts_utc": "2026-02-28T06:15:12Z",
  "session_id": "tg:-1003824466118",
  "subject_id": "ext:telegram:28036192",
  "channel": "telegram",
  "participants": [
    {
      "id": "@johnwick",
      "nickname": "John Wick",
      "protocol": "tg"
    }
  ],
  "task_text": "啧啧啧",
  "final_output": "",
  "draft_summary_items": [],
  "draft_promote": {}
}
```

Empty participants is also valid:

```json
{
  "participants": []
}
```

Agent-self participant example:

```json
{
  "participants": [
    {
      "id": 0,
      "nickname": "阿嬷",
      "protocol": ""
    }
  ]
}
```

### 10.7 Async Projection Trigger (When to Run)

Projection is asynchronous by design (not per-event immediate), because merge can involve LLM semantic dedupe.

Minimal trigger policy:

- Maintain dirty projection keys after WAL append.
- Run projector when at least one condition is met:
  - pending dirty count reaches threshold
  - oldest pending event age reaches threshold
  - periodic tick finds dirty keys
- On startup/restart, run replay+projection from checkpoint.

### 10.8 Projection Window (How Much Log per Run)

`window` means how much unapplied WAL is consumed in one projection pass.

Use two bounded windows:

- Journal window: max event lines read from checkpoint in one pass.
- Merge window: max summary candidates sent to one semantic dedupe merge call.

If backlog exceeds window, projector continues in later passes from updated checkpoint.

### 10.9 Semantic Dedupe Reuse

Projection merge reuses the existing Semantic Dedupe Subflow:

- existing short-term + incoming draft -> semantic dedupe merge
- otherwise -> direct merge

No new dedupe algorithm/prompt is introduced in WAL projection.
