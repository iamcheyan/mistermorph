# MisterMorph 的 nmem 映射结论

这份文档只写当前已经达成的约定。

## 1. 分阶段范围

### Phase 1

只做下面两件事：

1. 短期记忆文件 <-> `nmem memory`
2. `console topic` / 真实 `slack thread` <-> `nmem thread`

### Phase 2

再研究下面这些问题：

- 长期记忆如何映射成 `nmem memory`
- 长期记忆如何稳定地产生 `memory_id`
- `sourceThreadId` 如何回连到长期记忆

## 2. 先记住四条总规则

1. `subject_id` 不是 `memory_id`
2. `task_run_id` / `event_id` 不是 `thread_id`
3. `memory/log/*.jsonl` 是 WAL，不是 thread store
4. `nmem` 不需要 `bucket_id`

说明：

- `bucket_id` 如果被提到，只是内部思考“短期记忆按谁分桶”的术语
- 对外协议里只需要 `memory_id` 和 `thread_id`

## 3. Phase 1 的 `nmem memory`

Phase 1 里，`nmem memory` 只对应短期记忆文件。

来源：

- `memory/YYYY-MM-DD/*.md`

映射规则：

- 每个短期 markdown 文件，对应一条 synthetic `nmem memory`

结论：

- 短期 memory 的最小单位是“整个文件”
- 不是单条 summary line

## 4. 当前实现 vs 目标实现

### 4.1 当前实现

现在短期文件名的底层构造函数是：

- `memory/YYYY-MM-DD/{sanitize(sessionID)}.md`

但 projector 实际上传进去的是 `event.subject_id`，不是 `event.session_id`。

所以当前真实效果是：

- `memory/YYYY-MM-DD/{sanitize(subject_id)}.md`

### 4.2 Phase 1 目标实现

Phase 1 不改变短期 memory 的聚合边界。

继续沿用当前规则：

- 短期 memory 文件按 `subject_id` 分桶
- 不因为 runtime 里存在 thread，就把短期 memory 再拆成 thread-scoped 文件

这意味着：

- 短期 memory 和 `nmem thread` 可以不是一一对应关系
- `nmem thread` 是原始对话容器
- 短期 `nmem memory` 是当前系统已有的摘要文件

## 5. 短期 `memory_id` 规则

Phase 1 里，短期 `memory_id` 用确定性规则生成：

- `stm:<YYYY-MM-DD>:<subject_id>`

例子：

- Console
  - `stm:2026-04-04:console:topic_123`
- Slack thread
  - 当前短期 memory 仍然按 channel 聚合，因此是 `stm:2026-04-04:slack--t1--c1`
- Telegram
  - `stm:2026-04-04:tg:-1001234567890`

## 6. Phase 1 的 `nmem thread`

`nmem thread` 只对应真实存在的 runtime 对话容器。

### 6.1 Console

映射：

- `console topic -> nmem thread`

`thread_id`：

- `console:<topic_id>`

raw 数据来源：

- `tasks/console/topic.json`
- `tasks/console/log/YYYY-MM-DD_<topic_key>.jsonl`

### 6.2 Slack

映射：

- 真实 `slack thread -> nmem thread`

`thread_id`：

- `slack:<team_id>:<channel_id>:thread:<thread_ts>`

raw 数据来源：

- `tasks/slack/log/tasks.jsonl`

补充说明：

- Slack 的 runtime history 已经是 thread-aware
- 但当前 memory persistence 仍然是 channel-scoped
- Phase 1 不要求让短期 memory 与 Slack thread 对齐
- 也就是说：
  - 短期 memory 继续按 channel 聚合
  - `nmem thread` 单独按真实 Slack thread 建立

### 6.3 Telegram

Phase 1 不要求支持 Telegram -> `nmem thread`。

原因：

- Telegram 当前没有真正独立的 thread 对应物
- 现在只有 chat-scoped 语义：
  - `subject_id = tg:<chat_id>`
  - `session_id = tg:<chat_id>`

## 7. `slack:<team>:<channel>:thread:<thread_ts>` 格式约定

这个格式作为通用 internal id / `thread_id` 是合理的。

原因：

- 我们通用的 reference id 规则本质上是 `protocol:id`
- `id` 部分允许继续带 `:`
- 所以 `slack:T1:C1:thread:1739667600.000100` 是合法的 generic refid

但要注意：

- 它今天还不是现成可用的 `contacts_send chat_id` hint
- 因为当前 Slack chat hint parser 只支持 `slack:<team_id>:<channel_id>`

所以：

- 它可以直接作为 `thread_id`
- 但不要默认它已经是当前 contacts/chat hint 体系里的现成格式

## 8. 各 runtime 对照表

| Runtime | Phase 1 `nmem memory` | Phase 1 `nmem thread` |
| --- | --- | --- |
| Telegram | 按 `subject_id` 聚合的短期文件 | 无 |
| Console | 按 `subject_id` 聚合的短期文件 | `console topic` |
| Slack | 按 `subject_id` 聚合的短期文件 | 真实 Slack thread |

## 9. 对实现的直接要求

Phase 1 至少要补下面几件事：

1. 短期 `memory_id` 按 `日期 + subject_id` 生成
2. `console topic` / 真实 `slack thread` 分别建立 `thread_id`
3. 如果某个 runtime 没开 task persistence：
   - `thread` 同步不能依赖本地文件 replay
   - 只能依赖 live runtime hook

## 10. 一句话总结

Phase 1 的原则就是：

- `memory` 只处理短期文件
- `thread` 只处理真实会话容器
- `memory_id` 用确定性规则生成
- 短期 memory 继续按当前 `subject_id` 聚合，不按 thread 再拆
- `thread_id` 直接复用 runtime 的原始对话容器 identity
