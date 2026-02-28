# Memory Implementation Plan (WAL-first, Minimal)

## 1. Principles

- `memory/log/*.jsonl` is the single source of truth.
- `memory/*.md` is a projection (rebuildable from logs).
- Write path rule: `append + fsync` first, projection second.
- Channel runtime should only provide an adapter, not duplicate memory flow logic.
- No overdesign in v1.

## 2. Non-goals (v1)

- No distributed consensus.
- No multi-node writer arbitration.
- No remote replication.
- No background compaction service.
- No event bus dependency for memory pipeline.

## 3. Deliverables

- `internal/memoryruntime` package: shared orchestration.
- `memory/log/*.jsonl` journal with rotate + replay + checkpoint.
- Async markdown projector.
- Telegram adapter migration (behavior parity).
- Follow-up adapters for heartbeat and slack.

## 4. Work Breakdown

### Phase A: Data Contract

- [ ] Define `MemoryEvent` schema with minimal fields:
  - `schema_version`
  - `event_id`
  - `task_run_id`
  - `ts_utc`
  - `session_id`
  - `subject_id`
  - `channel`
  - `participants[]` with:
    - `id`
    - `nickname`
    - `protocol`
  - `task_text`
  - `final_output`
  - `draft_summary_items`
  - `draft_promote`
- [ ] Define id rules:
  - `event_id` unique per event.
  - `task_run_id` links events to one run.
- [ ] Add schema validation helpers.
  - `participants` may be empty.
  - if `participants` contains items, each entry must satisfy one of:
    - normal participant: non-empty `id`, non-empty `nickname`, non-empty `protocol`
    - agent self marker: `id == 0`, non-empty `nickname`, `protocol == ""`

Acceptance:

- [ ] Invalid events are rejected before journal append.

### Phase B: Journal (WAL)

- [ ] Implement journal append API:
  - `Append(event) -> offset`
  - `Append` includes flush/sync.
- [ ] Implement file rotation:
  - Rotate on size threshold.
  - Rotate on date boundary.
- [ ] Implement monotonic naming:
  - `YYYY-MM-DD-0001.jsonl`, `YYYY-MM-DD-0002.jsonl`.
- [ ] Implement replay iterator (ordered by file + line).
- [ ] Implement checkpoint store:
  - last applied file
  - last applied offset/line

Acceptance:

- [ ] Crash after append but before projection keeps event durable.
- [ ] Replay from checkpoint re-processes missing projections only.

### Phase C: Projector (Async)

- [ ] Implement async projection worker:
  - consume events in order
  - update markdown projection
- [ ] Implement trigger policy (not per-event immediate):
  - dirty-key tracking after WAL append
  - count threshold trigger
  - age threshold trigger
  - periodic tick trigger
- [ ] Define projection windows:
  - journal read window (max events per pass)
  - semantic merge window (max items per LLM dedupe call)
- [ ] Projection writes:
  - short-term file
  - long-term file
- [ ] Projection idempotency:
  - duplicate replay must not double-apply.
- [ ] Reuse existing Semantic Dedupe Subflow for short-term merge (no new dedupe pipeline).
- [ ] Lag visibility:
  - expose checkpoint and pending lag.

Acceptance:

- [ ] Append path does not block on markdown projection.
- [ ] Markdown files can be rebuilt from journal + checkpoint.
- [ ] Backlog larger than one window is drained incrementally across passes.

### Phase D: Shared Runtime Orchestrator

- [ ] Create `internal/memoryruntime` with shared flow:
  - `PrepareInjection(...)`
  - `RecordAndProject(...)`
- [ ] Define adapter interface for runtime-specific mapping:
  - identity mapping
  - write meta mapping
  - draft context/history mapping
  - request context (public/private)
- [ ] Remove duplicated inline memory flow from channel runtime code paths.

Acceptance:

- [ ] Core flow logic exists in one place only.

### Phase E: Telegram Migration (Parity First)

- [ ] Implement Telegram adapter on top of shared orchestrator.
- [ ] Keep current behavior parity:
  - injection behavior
  - writeback gating
  - heartbeat session mapping
- [ ] Keep `/mem` command behavior (read projection snapshot).

Acceptance:

- [ ] Existing telegram memory tests pass with adapter path.

### Phase F: Heartbeat and Slack Wiring

- [ ] Heartbeat adapter wiring (reuse same orchestrator).
- [ ] Slack adapter wiring (no copy-paste of flow).
- [ ] Configure runtime flags consistently across channels.

Acceptance:

- [ ] Slack and heartbeat do not add a separate memory pipeline.

### Phase G: Verification and Operations

- [ ] Add tests for:
  - append durability
  - rotate
  - replay
  - projector idempotency
- [ ] Add kill/restart scenario test:
  - crash between append and projection
  - restart replay fixes projection
- [ ] Add minimal operational docs:
  - how to inspect logs
  - how to replay/rebuild projection

Acceptance:

- [ ] End-to-end recovery scenario is reproducible with documented steps.

## 5. Suggested Sequence

1. Phase A + Phase B
2. Phase C
3. Phase D
4. Phase E
5. Phase F
6. Phase G

## 6. Stop Conditions (Avoid Overdesign)

- If a task does not improve durability, replayability, or de-duplication of runtime wiring, defer it.
- Keep first implementation single-process and local-only.
- Prefer simple append/replay correctness over optimization.
