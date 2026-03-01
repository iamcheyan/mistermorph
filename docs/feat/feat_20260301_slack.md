---
date: 2026-03-01
title: Slack vs Telegram Feature Gap Checklist
status: draft
---

# Slack vs Telegram Feature Gap Checklist

## 1) Scope

- Baseline for comparison: capabilities currently implemented in `mistermorph telegram`.
- Target for alignment: current `mistermorph slack` implementation.
- This document lists only *remaining or tracked alignment items* and does not repeat already-finished foundations (for example bus pipeline, contacts routing, baseline group-trigger classification, baseline memory integration).

## 2) Latest Priority (Reconfirmed on 2026-03-01)

### Mainline (Execute in Order)

- [x] 0. Bus identity enrichment: fill Slack inbound `username` / `display_name`
- Goal:
  - Align with Telegram-level sender identity richness, improve readability for contacts/history/memory.
- Current anchors:
  - Slack inbound events are user-id centric: `internal/channelruntime/slack/socket_events.go`
  - Bus extensions already include `FromUsername` / `FromDisplayName`: `internal/bus/adapters/slack/inbound.go`

- [x] 1. Emoji reaction parity (`slack_react`)
- Goal:
  - Provide channel-tool reaction capability aligned with Telegram `telegram_react`.
- Current anchors:
  - Telegram reaction + lightweight path: `internal/channelruntime/telegram/trigger.go`, `tools/telegram/react_tool.go`
  - Slack reaction tool + runtime wiring: `tools/slack/react_tool.go`, `internal/channelruntime/slack/runtime_task.go`

- [x] 2. Slack presentation for `plan_create`
- Goal:
  - Align with Telegram plan-progress visibility while using Slack-appropriate rendering.
- Current anchors:
  - Telegram has `WithPlanStepUpdate(...)` + progress publishing: `internal/channelruntime/telegram/runtime_task.go`
  - Slack now has plan progress hook wiring: `internal/channelruntime/slack/runtime_task.go`

- [x] 3. Channel-specific prompt block (Slack)
- Goal:
  - Add Slack-specific policy block (thread etiquette, mention semantics, output constraints).
- Current anchors:
  - Prompt block assembly: `internal/promptprofile/prompt_blocks.go`
  - Slack block template: `internal/promptprofile/prompts/block_slack.md`

- [x] 4. Heartbeat support for Slack
- Goal:
  - Support combined channel runtime + heartbeat operation, same as Telegram mode.
- Current anchors:
  - Telegram command includes optional heartbeat combined run: `cmd/mistermorph/telegramcmd/command.go`
  - Slack command includes optional heartbeat combined run: `cmd/mistermorph/slackcmd/command.go`

### P2 (Deferred)

- [ ] Inbound attachment processing (file-only messages, download, task injection).
- [ ] Slack command parity (`/reset` / `/id` / help) and init-flow parity.
- [ ] Further convergence of group-trigger boundaries (for example, thread-message edge behavior).
- [ ] Outbound text delivery improvements (long-message splitting, formatting strategy, rich-content degradation).
- [ ] Edited-message event handling.

## 3) Mainline Task Breakdown (Executable)

### 0. Bus identity enrichment: `username` / `display_name`

- Task breakdown:
  - [x] Add Slack API user profile lookup (`users.info`) with lightweight cache (`team_id:user_id`).
  - [x] Fill `username` / `display_name` on Slack inbound pipeline (fail-close on enrichment failure; do not publish to bus).
  - [x] Write enriched values to inbound extensions (`FromUsername` / `FromDisplayName`).
  - [x] Confirm contacts observation path can consume these fields (prefer `display_name` for nickname).
- Acceptance criteria:
  - [ ] Subsequent messages from same user in same team should not trigger high-frequency repeated `users.info` calls.
  - [ ] Bus inbound payload should stably contain `from_username` or `from_display_name` (when available).
  - [ ] On Slack API failure, inbound message should not enter bus and should remain observable via logs/hooks.
- Test suggestions:
  - [ ] Add profile lookup + cache-hit tests under `internal/channelruntime/slack/*_test.go`.
  - [ ] Add nickname write/update cases under `contacts/*_test.go`.

### 1. Emoji reaction parity (`slack_react`)

- Task breakdown:
  - [x] Add Slack channel tool: `slack_react` (at minimum `emoji` parameter).
  - [x] Add `reactions.add` call in Slack API layer.
  - [x] Inject tool in `runSlackTask(...)` with current `channel_id` + `message_ts` context.
  - [x] Record reaction outbound history (aligned with Telegram outbound-reaction semantics).
- Acceptance criteria:
  - [ ] When model calls `slack_react`, reaction is applied to target message.
  - [x] Invalid emoji / permission errors are safely returned without runtime crash.
  - [x] Reaction trail is visible in task history (for memory/audit).
- Test suggestions:
  - [x] Cover schema and execute path in `tools/slack/*_test.go`.
  - [ ] Add runtime registration/invocation tests in `internal/channelruntime/slack/runtime*_test.go`.

### 2. Slack presentation for `plan_create`

- Recommended V1:
  - Use thread incremental progress messages instead of message-editing first, to keep complexity low and behavior stable.
- Task breakdown:
  - [x] Wire `agent.WithPlanStepUpdate(...)` in Slack task runtime.
  - [x] Publish concise step-progress message in same thread (`correlation_id` marked as `slack:plan:*`).
  - [x] Add `plan_progress` outbound kind classification for UI/log distinction.
- Acceptance criteria:
  - [ ] For planned tasks, each completed step yields one progress message in thread.
  - [ ] No noisy progress messages when plan is missing/empty.
  - [ ] Main response and progress updates remain readable and non-overwriting.
- Test suggestions:
  - [x] Add plan-hook behavior tests in `internal/channelruntime/slack/runtime_task_test.go`.
  - [x] Add outbound kind classification tests in `internal/channelruntime/slack/runtime_outbound_test.go`.

### 3. Channel-specific prompt block (Slack)

- Task breakdown:
  - [x] Add Slack policy template block.
  - [x] Add `AppendSlackRuntimeBlocks(...)` in `promptprofile`.
  - [x] Inject Slack runtime blocks in Slack task runtime (different behavior by `im/group`).
  - [x] Add Slack mention-users block (parallel to Telegram group-username block).
- Acceptance criteria:
  - [x] Slack runtime request contains Slack-specific policy blocks.
  - [ ] Does not affect Telegram/other runtime prompt assembly.
- Test suggestions:
  - [x] Add block-injection tests under `internal/promptprofile/*_test.go`.
  - [ ] Add Slack runtime prompt-spec inclusion checks in `internal/channelruntime/slack/runtime_task_test.go`.

### 4. Heartbeat support for Slack

- Task breakdown:
  - [x] Reuse Telegram pattern and add `runSlackWithOptionalHeartbeat(...)` combined startup logic.
  - [x] Implement Slack heartbeat notifier (based on `chat.postMessage`).
  - [x] Define target behavior: V1 uses `slack.allowed_channel_ids` as heartbeat delivery targets.
  - [x] Update README/config docs (enablement conditions and caveats).
- Acceptance criteria:
  - [ ] With heartbeat enabled, configured Slack channels receive periodic heartbeat messages.
  - [x] Correct linked shutdown behavior when either Slack runtime or heartbeat exits.
  - [x] No panic when no target channels are configured (warn and degrade to no-op delivery).
- Test suggestions:
  - [x] Add combined-run tests under `cmd/mistermorph/slackcmd/*_test.go`.
  - [x] Add Slack notifier adaptation tests under `internal/channelruntime/heartbeat/*_test.go`.

## 4) Suggested PR Split

- [x] PR-1: Mainline 0 (bus identity enrichment)
- [x] PR-2: Mainline 1 (emoji / `slack_react`)
- [x] PR-3: Mainline 2 (plan progress rendering)
- [x] PR-4: Mainline 3 (Slack prompt block)
- [x] PR-5: Mainline 4 (heartbeat integration)
