---
date: 2026-03-06
title: Lark Runtime Implementation Plan
status: completed
---

# Lark Runtime Implementation Plan

## 1) Scope

- Add a new long-running channel runtime: `mistermorph lark`.
- Keep the same execution pipeline as Telegram, Slack, and LINE:
  - inbound event -> bus -> per-conversation worker -> `run*Task` -> outbound bus -> delivery adapter.
- Support `private + group` text messaging in V1.
- Reuse shared group-trigger logic: `strict | smart | talkative`.
- Keep contacts, todo references, sender routing, and prompt profile integration aligned with existing channel architecture.

## 2) Decisions Locked In

### Identity and ref rules

- chat ref: `lark:<chat_id>`
- user ref: `lark_user:<open_id>`

Examples:

- `[ChatID](lark:oc_84983ff6516d731e5b5f68d4ea2e1da5)`
- `[John](lark_user:ou_7d8a6e6df7621556ce0d21922b676706ccs)`

Why this is the minimum useful model:

- `chat_id` and `open_id` represent different objects.
- `lark:` stays reserved for chat routing.
- `lark_user:` stays reserved for speaker/contact identity.
- This matches current `line` / `line_user` design and avoids inventing a new special case.

### V1 canonical user ID

- Canonical user identity in refs and contacts: `open_id`
- `user_id` / `union_id` may be accepted later for API interoperability, but they are not part of the primary internal ref model in V1.

### Delivery semantics

- send message:
  - API: `POST /open-apis/im/v1/messages`
  - uses `receive_id_type + receive_id`
- reply message:
  - API: `POST /open-apis/im/v1/messages/:message_id/reply`
  - uses inbound `message_id`, not `receive_id_type`

### Bus mapping direction

- route by chat
- identify people by `open_id`
- do not overload a single `lark:*` namespace to represent both chat and user

## 3) Platform Facts Verified

Verified against official Feishu docs on 2026-03-06:

- self-built apps use `app_id + app_secret` to obtain `tenant_access_token`
- `tenant_access_token` expires in 2 hours
- send message supports `receive_id_type` values:
  - `open_id`
  - `union_id`
  - `user_id`
  - `email`
  - `chat_id`
- for user delivery, the doc explicitly recommends `open_id`
- reply message uses `message_id`
- receive message event provides message payloads for inbound bot processing

## 4) First-Principles Constraints

1. One routing key per conversation.
   Use `conversation_key = lark:<chat_id>`.

2. One canonical identity per speaker in V1.
   Use `participant_key = <open_id>` and `contact_id = lark_user:<open_id>`.

3. Keep channel specifics at the edge.
   Token exchange, webhook verification, and send/reply selection stay inside Lark runtime and adapter layers.

4. Avoid speculative identity expansion.
   Do not introduce `lark_union:*`, `lark_tenant:*`, or multi-ID merge logic in V1.

5. Re-check after every compression step.
   After each PR or scope reduction, explicitly ask whether the current shape is still the minimum model needed to route chats and identify users.

Current compression reviews:

- 2026-03-06 / chunk 1:
  channel foundations + config + command scaffold only.
  Kept exactly two ref namespaces in V1: `lark:` and `lark_user:`.
  Deferred webhook, token exchange, delivery, contacts, and prompt changes to later chunks.
- 2026-03-06 / chunk 2:
  Added webhook ingress, inbound/delivery adapters, runtime worker, prompt block, and group trigger wiring.
  Kept token exchange as one small shared primitive (`internal/larkapi`) instead of creating a broad shared Lark SDK layer.
- 2026-03-06 / chunk 3:
  Added contacts observation, direct sender routing, and todo/ref verification.
  Still kept exactly two user-facing namespaces in V1 and did not introduce `union_id`, `user_id`, tenant merge, cards, or media flows.

## 5) Proposed Bus Mapping

Inbound Lark event -> normalized runtime fields:

- `event.message.chat_id` -> `ChatID`
- `event.message.message_id` -> `MessageID`
- `event.message.chat_type` -> `ChatType`
- `event.sender.sender_id.open_id` -> `FromUserID`
- `event.header.event_id` -> `EventID`
- parsed text content -> `Text`
- mention list `open_id`s -> `MentionUsers`
- sender display name -> `DisplayName`

Normalized runtime fields -> bus:

- `Channel` -> `lark`
- `Topic` -> `chat.message`
- `ConversationKey` -> `lark:<chat_id>`
- `ParticipantKey` -> `<open_id>`
- `Extensions.PlatformMessageID` -> `<chat_id>:<message_id>`
- `Extensions.ReplyTo` -> `<message_id>`
- `Extensions.ChatType` -> `private|group`
- `Extensions.FromUserRef` -> `<open_id>`
- `Extensions.FromDisplayName` -> sender display name
- `Extensions.ChannelID` -> `<chat_id>`
- `Extensions.EventID` -> `<event_id>`
- `Extensions.MentionUsers` -> `[]open_id`

Envelope shape:

- `MessageID` -> `lark:<chat_id>:<message_id>`
- `Text` -> normalized plain text
- `SentAt` -> RFC3339 UTC
- `ReplyTo` -> `<message_id>`

## 6) Runtime Shape

```text
Lark Event Subscription webhook
  -> verify / challenge / decrypt if needed
  -> normalize event
  -> publish bus inbound message
  -> per-conversation worker
  -> runLarkTask
  -> publish bus outbound message
  -> lark delivery adapter
     |- reply by message_id
     `- send by receive_id_type + receive_id
```

## 7) Mainline Task Breakdown

### 0. Docs and decision freeze

- [x] Add `docs/lark.md`
- [x] Freeze V1 ref model:
  - [x] `lark:<chat_id>`
  - [x] `lark_user:<open_id>`
- [x] Freeze V1 delivery policy:
  - [x] reply by `message_id`
  - [x] send by `receive_id_type + receive_id`
- [x] Write first-principles / anti-overdesign constraints into this document
- [x] Keep this document updated as implementation progresses
- [x] After each implementation chunk, perform one explicit "can this be simpler?" review

Acceptance criteria:

- [x] No conflicting Lark ID scheme appears in later code or docs

### 1. Channel foundations

- [x] Add `lark` channel constant to shared channel enums
- [x] Add bus validation support for `ChannelLark`
- [x] Add `BuildLarkConversationKey(...)`
- [x] Add ref parsing helpers for:
  - [x] `lark:<chat_id>`
  - [x] `lark_user:<open_id>`

Acceptance criteria:

- [x] Bus messages with `channel=lark` validate
- [x] Conversation key helpers match `lark:<chat_id>`
- [x] Ref helpers reject malformed values

### 2. Config and command scaffold

- [x] Add `lark:` section to `assets/config/config.example.yaml`
- [x] Add `LarkConfig` + `BuildLarkRunOptions(...)` to `internal/channelopts`
- [x] Add `cmd/mistermorph/larkcmd/*`
- [x] Register `mistermorph lark` in root command wiring

Planned config:

```yaml
lark:
  base_url: "https://open.feishu.cn/open-apis"
  app_id: ""
  app_secret: ""
  webhook_listen: "127.0.0.1:18081"
  webhook_path: "/lark/webhook"
  verification_token: ""
  encrypt_key: ""
  allowed_chat_ids: []
  group_trigger_mode: "smart"
  addressing_confidence_threshold: 0.6
  addressing_interject_threshold: 0.6
  task_timeout: "0s"
  max_concurrency: 3
```

Acceptance criteria:

- [x] `go run ./cmd/mistermorph lark --help` shows expected flags
- [x] missing required credentials fail with actionable errors

### 3. Token client

- [x] Implement self-built app token exchange:
  - [x] `app_id + app_secret -> tenant_access_token`
- [x] Add in-memory token cache with expiry
- [x] Refresh before expiry with conservative margin
- [x] Add tests for:
  - [x] cache hit
  - [x] refresh before expiry window
  - [x] auth failure

Acceptance criteria:

- [x] runtime does not fetch a new token on every outbound request
- [x] expired token path refreshes cleanly

### 4. Webhook ingress

- [x] Add HTTP webhook handler in `internal/channelruntime/lark`
- [x] Handle URL verification / challenge path
- [x] Validate `verification_token`
- [x] Support decrypt path when `encrypt_key` is configured
- [x] Parse message receive events
- [x] Ignore unsupported event types in V1 with explicit logs

Acceptance criteria:

- [x] webhook verification succeeds with official callback flow
- [x] non-message events do not crash runtime

### 5. Inbound adapter

- [x] Add `internal/bus/adapters/lark/inbound.go`
- [x] Normalize Lark inbound messages into bus messages
- [x] Use `conversation_key = lark:<chat_id>`
- [x] Use `participant_key = <open_id>`
- [x] Write `FromUserRef = <open_id>`
- [x] Write `ReplyTo = <message_id>`
- [x] Add dedupe/idempotency tests

Acceptance criteria:

- [x] one inbound Lark message becomes one validated bus inbound message
- [x] conversation sharding is stable by chat

### 6. Runtime worker + task

- [x] Add `internal/channelruntime/lark/runtime.go`
- [x] Add worker serialization by `conversation_key`
- [x] Add `runLarkTask(...)`
- [x] Build prompt/request dump names:
  - [x] `prompt_lark_*`
  - [x] `request_lark_*`
- [x] Add base history/message rendering

Acceptance criteria:

- [x] concurrent chats run concurrently
- [x] same chat remains serialized

### 7. Delivery adapter

- [x] Add `internal/bus/adapters/lark/delivery.go`
- [x] Implement reply path:
  - [x] `POST /im/v1/messages/:message_id/reply`
- [x] Implement send path:
  - [x] `POST /im/v1/messages?receive_id_type=chat_id`
  - [x] body `receive_id=<chat_id>`
- [x] Choose reply when outbound carries usable source `message_id`
- [x] Fallback to send when reply is unavailable or not appropriate

Acceptance criteria:

- [x] in-reply-context outbound uses reply API
- [x] proactive outbound uses send API

### 8. Group trigger parity

- [x] Map Lark chat type to `private|group`
- [x] Private chat always triggers
- [x] Group explicit triggers:
  - [x] mention present in event payload
  - [x] leading `/`
- [x] Reuse shared `strict|smart|talkative` flow for non-explicit group messages
- [x] Add group trigger tests

Acceptance criteria:

- [x] Lark group behavior matches Telegram/Slack/LINE semantics within the V1 mention heuristic

### 9. Contacts and refs

- [x] Extend contacts model/store for:
  - [x] `LarkOpenID`
  - [x] `LarkChatIDs`
- [x] Observe inbound bus messages into contacts
- [x] Derive default contact IDs:
  - [x] `lark_user:<open_id>`
  - [x] `lark:<chat_id>`
- [x] Add sender routing support for:
  - [x] contact -> Lark user
  - [x] chat hint `lark:<chat_id>`

Acceptance criteria:

- [x] inbound Lark speaker becomes a resolvable contact
- [x] manual send paths can route to Lark

### 10. Prompt profile integration

- [x] Add `internal/promptprofile/prompts/block_lark.md`
- [x] Inject Lark runtime block in task path
- [x] Mention:
  - [x] reply vs send behavior
  - [x] group/private expectations
  - [x] Lark-specific mention semantics

Acceptance criteria:

- [x] Lark prompt block is present in inspect dumps

### 11. Todo/reference compatibility

- [x] Ensure todo/reference rendering accepts:
  - [x] `[ChatID](lark:<chat_id>)`
  - [x] `[John](lark_user:<open_id>)`
- [x] Add tests for parser and rendering

Acceptance criteria:

- [x] todo and contact references remain stable round-trip

## 8) Suggested PR Split

- [x] PR-1: channel foundations + config + command scaffold
- [x] PR-2: token client + webhook ingress
- [x] PR-3: inbound adapter + runtime worker
- [x] PR-4: delivery adapter + send/reply policy
- [x] PR-5: group trigger + prompt block
- [x] PR-6: contacts + sender + refs/todo support

## 9) Risks and Controls

- Risk: token refresh bugs cause intermittent send failures
  - Control: centralize token client and cover expiry behavior with tests

- Risk: webhook verification/decrypt path becomes channel-specific spaghetti
  - Control: keep verification and event normalization in a small isolated package boundary

- Risk: ID ambiguity leaks into refs
  - Control: keep `lark:` for chats and `lark_user:` for users only

- Risk: reply/send selection drifts into agent logic
  - Control: keep selection strictly inside delivery adapter

## 10) Deliverable Checklist

- [x] `mistermorph lark` command runnable
- [x] token exchange implemented
- [x] webhook challenge + message ingress implemented
- [x] inbound bus mapping implemented
- [x] runtime worker + `runLarkTask` implemented
- [x] delivery adapter implemented
- [x] group trigger parity implemented
- [x] contacts + sender parity implemented
- [x] todo/ref support for `lark:` and `lark_user:` implemented
- [x] docs updated and synced

## 11) References

- [Send message](https://open.feishu.cn/document/server-docs/im-v1/message/create)
- [Reply message](https://open.feishu.cn/document/server-docs/im-v1/message/reply)
- [Receive message event](https://open.feishu.cn/document/server-docs/im-v1/message/events/receive)
- [Self-built app tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)
