---
date: 2026-03-03
title: Telegram sendMessageDraft Streaming Rollout Plan
status: draft
---

# Telegram sendMessageDraft Streaming Rollout Plan

## 1) Scope

- Add Telegram `sendMessageDraft` support in our Telegram API/runtime path.
- Switch LLM response generation from `Chat` (non-streaming) to `Chat` with `WithOnStream(...)` callback.
- Deliver a private-chat first implementation.
- Deliver a group-chat implementation after private-chat behavior is stable.

## 2) Goals

- Improve perceived latency with incremental output.
- Keep final message quality and tool-calling behavior unchanged.
- Reuse one streaming core flow, with channel-context-specific publishing policy.

## 3) Non-Goals (V1)

- No cross-channel rollout in this phase (Slack/Discord not included yet).
- No full rewrite of task orchestration.
- No additional multimodal changes in this phase.

## 4) Design Overview

- LLM layer:
  - Use `uniai.WithOnStream(func(ev uniai.StreamEvent) error { ... })`.
  - Accumulate `ev.Delta` into full final text buffer.
  - Preserve tool-call handling semantics; stream callback must not break current tool loop expectations.
- Telegram delivery layer:
  - Add API wrapper for `sendMessageDraft`.
  - Introduce draft lifecycle per task: create/update/finalize behavior (implementation detail finalized in coding phase).
  - Keep fallback to current `sendMessage` when draft path is unavailable.
- Runtime policy:
  - Private chat: enable draft streaming by default (behind config flag during rollout).
  - Group chat: enable with stricter throttling and noise control.

## 5) Execution Plan

### Phase 1: Support `sendMessageDraft`

- Add Telegram API client method(s) and request/response structs.
- Add minimal validation and error mapping.
- Add unit tests with mocked Telegram API responses.
- Add logs for draft send/update/finalize and fallback path.

Acceptance:
- Runtime can call Telegram draft API successfully in isolation.
- On draft API failure, task still finishes with normal `sendMessage`.

### Phase 2: Use uniai stream instead of plain Chat

- Refactor Telegram task LLM call path to use stream callback.
- Keep one assembled final text result for downstream compatibility.
- Preserve current tool-call behavior and cancellation semantics.

Acceptance:
- Final output text is equivalent to non-streaming path for same prompt/model.
- Stream cancellation/error does not leave task in broken state.

### Phase 3: Private chat rollout

- Enable draft streaming in Telegram private chats.
- Add throttling policy for draft updates (avoid per-token API spam).
- Finalize response into normal visible completion state.

Acceptance:
- Private chat sees incremental output with stable completion.
- No duplicate final message and no stuck draft.

### Phase 4: Group chat rollout

- Extend the same streaming core to group/supergroup.
- Apply tighter update interval and mention/thread safety policy.
- Keep existing trigger/addressing decision unchanged.

Acceptance:
- Group chats get incremental output without excessive noise.
- Existing group-trigger behavior and reaction behavior remain compatible.

## 6) Config and Rollout Controls

- Add rollout toggle in config for Telegram draft streaming (default conservative).
- Allow separate enablement for private/group if needed.
- Keep kill-switch for immediate rollback to non-streaming send path.

## 7) Observability

- Add metrics/log fields:
  - stream started / completed / cancelled / failed
  - draft updates count
  - draft fallback count
  - stream duration and token usage (if available)
- Ensure logs are keyed by chat/task identifiers for debugging.

## 8) Test Plan

- Unit tests:
  - Telegram draft API wrapper and error mapping.
  - Stream delta accumulation and final text assembly.
  - Fallback behavior.
- Integration tests:
  - Private chat draft stream end-to-end (mock Telegram + mock LLM stream).
  - Group chat draft stream end-to-end.
- Regression tests:
  - Tool calling path unchanged.
  - Non-stream path still works when streaming disabled.

## 9) Reference Stream Pattern

```go
resp, err := client.Chat(ctx,
    uniai.WithModel("gpt-5.2"),
    uniai.WithMessages(uniai.User("Tell me a story.")),
    uniai.WithOnStream(func(ev uniai.StreamEvent) error {
        if ev.Done {
            return nil
        }
        if ev.Delta != "" {
            // append delta and publish throttled draft update
        }
        if ev.ToolCallDelta != nil {
            // preserve existing tool-call incremental handling semantics
        }
        return nil
    }),
)
// resp.Text is final accumulated text
```

## 10) Risks

- Telegram draft API behavior may differ across chat types and bot permissions.
- Too frequent draft updates can trigger rate limits.
- Streaming + tool-call interleaving can introduce ordering edge cases.

Mitigation:
- Throttled updates + fallback path.
- Strict logging and phased rollout (private first, group second).
