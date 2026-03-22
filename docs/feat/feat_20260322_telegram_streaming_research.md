---
date: 2026-03-22
title: Telegram Streaming 现状预研（基于 2026-03-22 最新代码）
status: superseded
---

# Telegram Streaming 现状预研（基于 2026-03-22 最新代码）

本文回答一个非常具体的问题：

> 最新 commit 已经把 streaming 接到了 `console serve` 的 web UI；如果要把同样能力扩展到 Telegram 私聊，当前代码已经具备什么，还缺什么，最小可行方案是什么？

实现结论更新：

- 本文早期版本假设 Telegram Bot API 上可以接受 `draft preview + final sendMessage` 这一路线。
- 实际验证后发现，Bot API 层没有 same-bubble handoff 能力；要么 reply 最终消失，要么出现不可接受的双影。
- 因此当前仓库结论不是“Telegram streaming 先用 preview-only 顶着”，而是：**Telegram runtime 暂不启用 draft streaming**。
- 本文后续关于 Telegram 私聊 draft streaming 的设计内容保留为历史研究，不再代表当前 runtime 行为。

结论先行：

- 共享 streaming 基建已经基本具备。
- Console Local 已经完成端到端闭环，可以作为 Telegram 的参考实现。
- Telegram 当前并没有 streaming；它离“私聊可用”还差一个很有限但很明确的接线层。
- 最小正确做法不是重新设计一整套 streaming 平台，而是复用现有 shared extractor，加一个 Telegram 私聊专用 `ReplySink`。

---

## 1) 时间线结论

近期与本问题直接相关的提交有 3 个：

- `2026-03-03` `1587a93 feat(telegram): stream draft delivery via sendMessageDraft`
  - 仓库里曾经实现过 Telegram draft streaming。
- `2026-03-16` `9114f47 feat: manage console runtimes and remove telegram streaming`
  - 在 managed runtime 重构时，Telegram streaming 被移除。
- `2026-03-22` `1a1ab0f Feat/streaming (#26)`
  - 这次把 shared streaming 基建接回来了，但只落到了 Console Local + console web UI。

所以当前真实状态不是“从零开始给 Telegram 做 streaming”，而是：

```text
旧 Telegram streaming 被拿掉
        +
shared streaming 基建重新变得可复用
        =
现在只差 Telegram 这一层重新接回去
```

---

## 2) 当前真实代码状态

### 2.1 Shared LLM / Agent / Runtime 层

共享 streaming 主干已经通了：

- `llm.Request.OnStream`
- `providers/uniai/client.go`
- `agent.RunOptions.OnStream`
- `internal/channelruntime/taskruntime.RunRequest.OnStream`

说明：

- provider 侧不是阻塞点。
- `agent` 和 `taskruntime` 不是阻塞点。
- 真正的差异只在各 channel task path 有没有把 `OnStream` 往下传。

### 2.2 共享 extractor 已经存在

这次 console streaming 不是直接把 provider raw delta 推给 UI，而是先过一个 shared extractor：

- `internal/streaming/streaming.go`

它已经提供：

- `ReplySink` interface
- `FinalOutputStreamer`
- 从 JSON stream 中提取 `final.output`
- 节流与 done flush

这点非常关键，因为当前 agent loop 的输出协议不是 plain text，而是强制 JSON：

```json
{
  "type": "final",
  "reasoning": "...",
  "output": "..."
}
```

所以真正可复用的核心不是“拿到 token delta”，而是：

> 能稳定地从 JSON 流里只提取用户可见的 `final.output`。

### 2.3 Console Local 已经闭环

Console 现在已经是完整参考实现：

- `cmd/mistermorph/consolecmd/local_runtime.go`
  - 在 `handleTaskJob()` 里创建 `ReplySink`
  - 用 `streaming.NewFinalOutputStreamer(...)`
  - 把 `OnStream` 传进 `runTask()`
- `cmd/mistermorph/consolecmd/streaming.go`
  - 提供 in-memory `consoleStreamHub`
  - 提供 `POST /api/stream/ticket`
  - 提供 `GET /api/stream/ws`
- `web/console/src/views/ChatView.js`
  - 每个 `task_id` 绑定一个 websocket
  - 走 same-bubble update
  - websocket 失败时保留 polling fallback

当前 console 路径可以概括为：

```text
LLM OnStream
  -> FinalOutputStreamer
  -> ConsoleReplySink
  -> StreamHub
  -> WebSocket
  -> ChatView same bubble update
```

### 2.4 Telegram 当前仍是非 streaming

Telegram 当前主链路是：

- `internal/channelruntime/telegram/runtime.go`
- `internal/channelruntime/telegram/runtime_task.go`

但它现在还缺 3 个关键点：

1. `runTelegramTask(...)` 没把 `OnStream` 传给 `taskruntime.Run(...)`
2. `telegram_api.go` 没有 `sendMessageDraft` wrapper
3. runtime 完成后仍然只走 canonical final send path

也就是说，Telegram 现在仍然是：

```text
runTelegramTask
  -> taskruntime.Run (no OnStream)
  -> final output
  -> publish outbound bus text
  -> telegram delivery adapter -> sendMessage
```

而不是：

```text
runTelegramTask
  -> taskruntime.Run (with OnStream)
  -> ReplySink.Update() during stream
  -> sendMessageDraft(private only)
  -> final path finalize/fallback
```

---

## 3) What Already Exists

这部分很重要，因为本问题不是新建一套系统，而是尽量复用现有能力。

### 3.1 已存在的共享能力

- `OnStream` 通路已经打通
- shared `ReplySink` interface 已存在
- shared `FinalOutputStreamer` 已存在
- JSON partial extractor 已存在
- tool-call delta 已经在 shared llm event 结构里建模

### 3.2 已存在的 Console 参考实现

- same-bubble 更新语义
- in-process ephemeral stream hub
- ticket-based websocket auth
- polling fallback
- done / failed / pending 等 lifecycle 与 stream 共存

### 3.3 历史上做过的 Telegram 方案

旧实现已经证明过几件事：

- Telegram streaming 不需要改 provider 层
- Telegram runtime 本地直发 draft 是合理边界
- final canonical send path 可以与 draft path 共存

但旧实现也有一个缺点：

- 它是 Telegram 自己维护一套 extractor / publisher 逻辑，没有复用现在更干净的 shared `internal/streaming`

因此现在最合理的做法不是把旧实现整段抄回来，而是：

- 复用新的 shared extractor
- 只补 Telegram transport + Telegram sink

---

## 4) 当前缺口

### 4.1 Telegram API wrapper 缺失

当前 `telegram_api.go` 只有：

- `sendMessage...`
- `editMessage...`
- `sendMessageChunkedReply...`

缺：

- `sendMessageDraft`

这一点不是理论缺口，而是实打实的代码缺口。

### 4.2 Telegram task path 没有接 `OnStream`

`runTelegramTask(...)` 当前调用 `taskruntime.Run(...)` 时没有传 `OnStream`。

所以即使 shared layer 能 stream，Telegram 也完全收不到。

### 4.3 缺一个 Telegram 私聊专用 ReplySink

Console 的 `ReplySink` 负责“同一个 task 只对应一个气泡”。

Telegram 也需要相同收口：

- stream 期间：`Update()`
- 终态：`Finalize()`
- 错误：`Abort()`

否则很容易出现：

- stream path 创建一条可见消息
- final path 再发一条最终消息
- 最终重复

### 4.4 私聊 / 群聊范围没有重新明确

当前最合理的 V1 范围应当是：

- `private` chat only

原因：

- 私聊是最明确的目标场景
- draft API 语义更匹配私聊
- 群聊会额外引入噪音控制、线程/回复语义和触发策略风险

### 4.5 文档已经发生漂移

当前仓库里有文档漂移：

- `docs/telegram.md` 把 `sendMessageDraft` 写成“当前运行中的行为”
- 但代码里实际上没有 `sendMessageDraft` wrapper，也没有 `OnStream` wiring

所以如果不先澄清真实状态，后续实现很容易被误导。

---

## 5) 官方 API 结论

本次预研补充确认了 Telegram 官方 Bot API：

- `sendMessageDraft`
- `editMessageText`

官方文档：

- `https://core.telegram.org/bots/api#sendmessagedraft`
- `https://core.telegram.org/bots/api#editmessagetext`

对本项目的直接意义：

- Telegram Bot API 侧没有挡路。
- 私聊 draft streaming 在 API 能力上是成立的。
- 当前阻塞点在项目内接线，不在外部平台能力。

---

## 6) 推荐的最小实现

### 6.1 目标

最小目标不是“所有 channel 统一流式”，而是：

> 让 Telegram 私聊复用现有 shared streaming 核心，获得和 Console Local 类似的渐进式回复体验。

### 6.2 推荐方案

```text
LLM OnStream
  -> shared FinalOutputStreamer
  -> TelegramReplySink
       -> sendMessageDraft(private)
       -> finalize existing draft if possible
       -> fallback to normal final send if draft path fails
```

### 6.3 为什么这是最小 diff

因为它只需要补这几层：

1. Telegram API wrapper
2. TelegramReplySink
3. `runTelegramTask()` 的 `OnStream` wiring
4. runtime 终态时的 finalize/fallback 配合

它不需要：

- 新建通用 stream broker
- 新建 daemon stream 协议
- 改 bus 模型
- 改 task store 语义

---

## 7) 推荐实施边界

### 7.1 In scope

- Telegram private chat draft streaming
- 复用 `internal/streaming`
- 保留现有 canonical final send path
- draft 失败时回退到普通最终发送
- 补测试

### 7.2 NOT in scope

- Telegram group / supergroup draft streaming
- 把 stream snapshot 写进 bus
- 把 stream snapshot 写进 `TaskInfo.Result`
- 远端 `console.endpoints[]` 的统一 streaming
- 新建 cross-channel 通用 stream 平台

---

## 8) 推荐实现步骤

### Phase 1: 补 Telegram transport

文件：

- `internal/channelruntime/telegram/telegram_api.go`
- `internal/channelruntime/telegram/telegram_api_*test.go`

内容：

- 补 `sendMessageDraft` wrapper
- 复用现有 HTML render / entity parse fallback 风格
- 补 request/response/error mapping 测试

### Phase 2: 接 shared streaming core

文件：

- `internal/channelruntime/telegram/runtime_task.go`

内容：

- 创建 Telegram 私聊专用 sink
- 用 `streaming.NewFinalOutputStreamer(...)`
- 把 `OnStream` 传给 `taskruntime.Run(...)`
- 只在 `private` chat 启用

### Phase 3: 终态收口

文件：

- `internal/channelruntime/telegram/runtime.go`
- `internal/channelruntime/telegram/runtime_task.go`

内容：

- final 时优先 finalize 已存在的 draft
- draft 失败时继续 canonical final send
- 不破坏现有 bus outbound path

### Phase 4: 文档与测试

文件：

- `internal/channelruntime/telegram/runtime_task_test.go`
- `docs/telegram.md`

内容：

- 补 stream update / finalize / fallback 测试
- 修正文档漂移

---

## 9) 测试建议

至少需要覆盖：

### 9.1 Shared 行为

- `final.output` 可被持续提取
- `plan` / tool-call 不会误推给用户
- done 时 flush

### 9.2 Telegram transport

- `sendMessageDraft` 成功
- entity parse 失败时回退 plain text
- API error 时正确返回错误

### 9.3 Telegram runtime

- 私聊开启 draft streaming
- 群聊不启用 draft streaming
- draft path 失败时 final 仍正常发送
- 不会重复发两条可见最终回复

---

## 10) 实施前决策表

下表不是讨论“要不要做 Telegram streaming”，而是实现前必须先锁定的边界条件。

| ID | 决策项 | 推荐决策 | 为什么 | 如果不这样做，会发生什么 |
| --- | --- | --- | --- | --- |
| D1 | draft 成功后，final 是否还走当前 `sendMessage` | 保留最终 `sendMessage`；draft 仅做 preview，不承担 terminal delivery | 真实运行里，Telegram draft 更接近临时预览而不是真正消息实体；如果把 draft 当 final，会出现“内容最后消失”的问题 | 最终消息可能消失，或者为了避免消失而只能重新发送，造成策略摇摆 |
| D2 | draft 失败时怎么处理 | 立即停用 draft path，回退到当前 canonical final send path | 保持可逆性，确保 streaming 失败不影响任务完成 | 可能出现既没有流式更新，也没有最终消息，或者卡在半成品 draft |
| D3 | `final.is_lightweight=true` 时是否允许流式文本 | 默认不允许；私聊流式只面向最终需要发布文本的回复 | 当前 publish 决策发生在 final 阶段；如果先流后判 lightweight，容易提前泄露本不该发送的文本 | 会出现“最终规则说不发，但中途已经发出来了”的语义冲突 |
| D4 | Telegram streaming 的首个启用范围 | 仅 `private chat` | 私聊是目标场景最明确、噪音最小、验证成本最低的边界 | 如果一开始带上 group/supergroup，会同时引入触发策略、回复链路和噪音控制问题 |
| D5 | 是否复用 shared extractor | 必须复用 `internal/streaming` | 共享 extractor 已经在 console 验证过；继续复用能保持行为一致，也避免双实现漂移 | Telegram 会重新长出一套单独 extractor，后续修 bug 和协议变更都会分叉 |
| D6 | stream snapshot 是否进入 bus / task store | 不进入；保持 runtime-local ephemeral 输出 | 当前 snapshot 是 UI/渠道临时态，不是 durable task fact | 会污染持久化语义，增加写放大，也让 `/tasks/{id}` 变得不再稳定 |
| D7 | `draft_id` 生命周期怎么建模 | 每个 run 一个 `TelegramReplySink`，由 sink 独占 draft 状态 | 把“更新 / finalize / abort”收口到一个对象里，边界最清晰 | draft 状态散落在 runtime 和 transport 层，最终很难保证去重与回退一致 |
| D8 | 群聊是否沿用同一方案 | 不纳入这一轮 | 群聊不是“顺手打开”的小增量，它会显著扩大 blast radius | 需求会从“接回 Telegram 私聊 streaming”滑向“重做 Telegram 回复策略” |

### 10.1 推荐冻结版本

如果按本文建议推进，实施前可以先把下面这组决策冻结：

```text
D1: 保留最终 sendMessage；draft 只做 preview
D2: draft 任一步失败 => 立即降级到现有 final send path
D3: lightweight reply => 不启用流式文本
D4: 只做 private chat
D5: 必须复用 shared FinalOutputStreamer
D6: snapshot 不进 bus / task store
D7: per-run TelegramReplySink 独占 draft 生命周期
D8: group / supergroup 暂不纳入
```

### 10.2 决策冻结后的目标形态

```text
LLM OnStream
  -> shared FinalOutputStreamer
  -> TelegramReplySink(private only)
       -> Update() => sendMessageDraft preview
       -> Finalize() => no-op / keep preview semantics
       -> Abort() => disable draft path

if draft path failed
  -> fallback to current final sendMessage path

run finished
  -> normal final sendMessage
```

---

## 11) 风险判断

### 11.1 低风险

- shared extractor 已存在并有测试
- taskruntime `OnStream` 通路已存在
- Console 已证明整体模型成立

### 11.2 中风险

- Telegram draft API 的细节兼容性
- draft preview 与 final `sendMessage` 之间的短暂双影窗口

### 11.3 当前最需要避免的错误

最容易犯的错误不是“实现不出来”，而是“实现方向跑偏”：

1. 把 raw JSON delta 直接流给用户
2. 让 stream path 和 final path 各自产生一条可见消息
3. 为 Telegram 单独再造一套 extractor，而不是复用 shared core
4. 一上来把群聊也纳入范围

---

## 12) 最终建议

当前代码基础下，最优路线是：

```text
先不做大而全
只做 Telegram private chat
只补 transport + sink + wiring
完全复用 shared extractor
保留 canonical final send
```

这条路线的优点是：

- diff 小
- 风险清晰
- 可测试
- 与最新 console streaming 架构一致
- 后续要扩到 Slack/群聊时也不会推翻

---

## 13) 本次预研附带验证

本次本地验证已通过：

- `go test ./internal/streaming`
- `go test ./cmd/mistermorph/consolecmd`
- `go test ./internal/channelruntime/telegram`
- `go test ./internal/channelruntime/taskruntime`

说明：

- 最新 `2026-03-22` 的 console streaming 基建目前处于可工作的状态。
- 当前问题确实集中在 Telegram 没有重新接回 streaming，而不是 shared core 本身不稳定。
