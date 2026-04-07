---
date: 2026-04-04
title: Memory Plugin API (Phase 1)
status: draft
---

# Memory Plugin API (Phase 1)

这份文档定义 Phase 1 的外部 plugin 协议。

前置概念映射见：

- `docs/feat/feat_20260403_nmem_mapping.md`

## 1. Scope

Phase 1 只冻结 5 个 hook：

1. `memory.prepare`
2. `memory.stm.upsert`
3. `topic.upsert`
4. `topic.append`
5. `topic.delete`

Phase 1 不做：

- 长期 memory
- Telegram topic 映射
- 通用 event export
- 原始 WAL event 直出给 plugin

## 2. First Principles

### 2.1 协议只暴露已经成形的对象

plugin 不应该消费内部 journal event。

Phase 1 对外只暴露两个对象：

- `memory`
  - 当前只指短期 markdown 文件
- `topic`
  - 当前统一表达 Console topic 和 Slack thread

### 2.2 hook 名跟我们的对象走

所以协议层用：

- `topic.*`

而不是：

- `thread.*`

`nmem` 侧如果需要，再把 `topic` 映射成它的 `thread`。

### 2.3 Phase 1 只保留最少字段

如果一个字段可以稳定推导，就先不冻结进 schema。

例如：

- `memory.prepare` 不要求 `session_id`
- `memory.prepare` 不要求 `runtime`
- `memory.prepare` 不要求 `request_context`
- `memory.stm.upsert` 不要求 `date`
- `memory.stm.upsert` 不要求 `summary`
- `topic.append` 不要求 `runtime`

## 3. Shared Conventions

### 3.1 IDs

短期 `memory_id`：

- `stm:<YYYY-MM-DD>:<subject_id>`

例子：

- `stm:2026-04-04:console:topic_123`
- `stm:2026-04-04:slack--t1--c1`
- `stm:2026-04-04:tg:-1001234567890`

`topic_id`：

- Console: `console:<topic_id>`
- Slack: `slack:<team_id>:<channel_id>:thread:<thread_ts>`

### 3.2 caller-supplied ID 是前提

Phase 1 的标准方案要求 backend 直接接受我们给出的 ID。

也就是说：

- `memory_id`
- `topic_id`

应直接成为 backend 里的 canonical ID。

Phase 1 不接受“adapter 自己维护外部 ID 到 backend 内部 ID 的映射表”作为标准方案。

### 3.3 Wire Format

Phase 1 不再使用 `input/output` 包裹层。

请求统一是扁平 JSON：

```json
{
  "protocol_version": "v1",
  "hook": "memory.prepare",
  "...": "hook 自己的字段"
}
```

成功响应也统一是扁平 JSON：

- 读操作：

```json
{
  "ok": true,
  "...": "hook 自己的返回字段"
}
```

- 写操作：

```json
{
  "ok": true
}
```

错误响应：

```json
{
  "ok": false,
  "code": "invalid_input",
  "message": "subject_id is required"
}
```

共享错误响应 schema：

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["ok", "code", "message"],
  "properties": {
    "ok": { "const": false },
    "code": { "type": "string", "minLength": 1 },
    "message": { "type": "string", "minLength": 1 }
  }
}
```

## 4. `memory.prepare`

### 4.1 语义

- 用当前用户输入做召回
- 返回一段可直接塞进现有 memory prompt slot 的文本
- 只读，无副作用

### 4.2 Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["protocol_version", "hook", "subject_id", "task_text"],
  "properties": {
    "protocol_version": { "const": "v1" },
    "hook": { "const": "memory.prepare" },
    "subject_id": { "type": "string", "minLength": 1 },
    "task_text": { "type": "string" },
    "max_items": { "type": "integer", "minimum": 1 }
  }
}
```

### 4.3 Response Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["ok", "prompt_text"],
  "properties": {
    "ok": { "const": true },
    "prompt_text": { "type": "string" }
  }
}
```

### 4.4 `nmem` 连线

最小推荐调用：

```bash
nmem --json m search "<task_text>" -n <max_items> \
  -l "mm:subject:<subject_id>"
```

说明：

- Phase 1 只要求一个最小 scope label：
  - `mm:subject:<subject_id>`
- 不回退到 `t search`

如果 `m search` 返回信息不足，再补：

```bash
nmem --json m show <memory_id> --content-limit <chars>
```

`prompt_text` 推荐格式：

```text
<Memory:ShortTerm:Recent>
- 2026-04-04: discussed release plan
- 2026-04-03: reviewed deploy rollback
```

日期优先从 `memory_id` 解析。

### 4.5 示例

请求：

```json
{
  "hook": "memory.prepare",
  "protocol_version": "v1",
  "subject_id": "console:topic_123",
  "task_text": "summarize the release risks",
  "max_items": 8
}
```

响应：

```json
{
  "ok": true,
  "prompt_text": "<Memory:ShortTerm:Recent>\n- 2026-04-04: discussed release plan"
}
```

## 5. `memory.stm.upsert`

### 5.1 语义

- 把一整份短期 markdown 文件同步给 plugin
- 最小单位是整个文件，不是单条 summary item

### 5.2 Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["protocol_version", "hook", "memory_id", "subject_id", "markdown"],
  "properties": {
    "protocol_version": { "const": "v1" },
    "hook": { "const": "memory.stm.upsert" },
    "memory_id": { "type": "string", "minLength": 1 },
    "subject_id": { "type": "string", "minLength": 1 },
    "markdown": { "type": "string" },
    "updated_at": { "type": "string", "format": "date-time" },
    "source_relpath": { "type": "string" }
  }
}
```

### 5.3 Response Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["ok"],
  "properties": {
    "ok": { "const": true }
  }
}
```

### 5.4 `nmem` 连线

先从 markdown 里解析 frontmatter。

`nmem m add` / `m update` 里的 `-t` 应优先使用短期 memory 文件 frontmatter 里的 `summary`。

如果 frontmatter 里没有 `summary`：

- 可以省略 `-t`
- 不建议回退成 `memory_id`

前提条件：

- `nmem m add --id <memory_id>` 已可用

创建：

```bash
nmem --json m add "<markdown>" \
  --id "<memory_id>" \
  -t "<frontmatter_summary>" \
  --unit-type context \
  -l "mm:subject:<subject_id>"
```

更新：

```bash
nmem --json m update <memory_id> -c "<markdown>" -t "<frontmatter_summary>"
```

### 5.5 示例

```json
{
  "hook": "memory.stm.upsert",
  "protocol_version": "v1",
  "memory_id": "stm:2026-04-04:console:topic_123",
  "subject_id": "console:topic_123",
  "markdown": "---\nsummary: discussed release plan\n---\n\n- ...",
  "updated_at": "2026-04-04T09:20:00Z"
}
```

## 6. `topic.upsert`

### 6.1 语义

- 保证一个 topic 在 plugin/backend 里存在
- 用于 create-if-missing
- 也允许顺手更新轻量 metadata，例如 title

### 6.2 Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["protocol_version", "hook", "topic_id"],
  "properties": {
    "protocol_version": { "const": "v1" },
    "hook": { "const": "topic.upsert" },
    "topic_id": { "type": "string", "minLength": 1 },
    "title": { "type": "string" },
    "source_ref": { "type": "string" }
  }
}
```

### 6.3 Response Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["ok"],
  "properties": {
    "ok": { "const": true }
  }
}
```

### 6.4 `nmem` 连线

`topic.upsert` 的职责是 ensure-exists。

注意：

- `nmem t` 没有 `update`
- topic 侧的后续变更统一走 `t append`

推荐流程：

1. 先看 topic 是否已存在
2. 不存在再 create
3. 已存在则 no-op

查询：

```bash
nmem --json t show <topic_id> -n 1
```

创建：

```bash
nmem --json t create \
  --id "<topic_id>" \
  -t "<title>"
```

如果需要记录来源，可以在 `t create` 里额外传：

```bash
-s "<source_ref>"
```

### 6.5 示例

```json
{
  "hook": "topic.upsert",
  "protocol_version": "v1",
  "topic_id": "console:topic_123",
  "title": "Release Plan",
  "source_ref": "tasks/console/log/2026-04-04_topic_123.jsonl"
}
```

## 7. `topic.append`

### 7.1 语义

- 向一个已有 topic 追加新消息
- payload 是增量消息，不是完整快照

### 7.2 Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["protocol_version", "hook", "topic_id", "messages"],
  "properties": {
    "protocol_version": { "const": "v1" },
    "hook": { "const": "topic.append" },
    "topic_id": { "type": "string", "minLength": 1 },
    "messages": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["kind", "sent_at", "text"],
        "properties": {
          "kind": {
            "enum": [
              "inbound_user",
              "inbound_reaction",
              "outbound_agent",
              "outbound_reaction",
              "system"
            ]
          },
          "sent_at": { "type": "string", "format": "date-time" },
          "text": { "type": "string" },
          "message_id": { "type": "string" },
          "reply_to_message_id": { "type": "string" },
          "sender": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
              "user_id": { "type": "string" },
              "username": { "type": "string" },
              "nickname": { "type": "string" },
              "is_bot": { "type": "boolean" },
              "display_ref": { "type": "string" }
            }
          }
        }
      }
    }
  }
}
```

### 7.3 Response Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["ok"],
  "properties": {
    "ok": { "const": true }
  }
}
```

### 7.4 `nmem` 连线

直接追加增量消息：

```bash
nmem --json t append <topic_id> -m '<messages_json>'
```

`t append` 不支持 `-s`，也不需要 `-s`。

### 7.5 示例

```json
{
  "hook": "topic.append",
  "protocol_version": "v1",
  "topic_id": "console:topic_123",
  "messages": [
    {
      "kind": "inbound_user",
      "sent_at": "2026-04-04T09:00:00Z",
      "text": "summarize the release risks"
    },
    {
      "kind": "outbound_agent",
      "sent_at": "2026-04-04T09:00:10Z",
      "text": "here are the release risks"
    }
  ]
}
```

## 8. `topic.delete`

### 8.1 语义

- 删除一个 topic

Phase 1 主要对应：

- Console topic 删除

### 8.2 Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["protocol_version", "hook", "topic_id"],
  "properties": {
    "protocol_version": { "const": "v1" },
    "hook": { "const": "topic.delete" },
    "topic_id": { "type": "string", "minLength": 1 }
  }
}
```

### 8.3 Response Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["ok"],
  "properties": {
    "ok": { "const": true }
  }
}
```

### 8.4 `nmem` 连线

```bash
nmem --json t delete <topic_id> -f
```

### 8.5 示例

```json
{
  "hook": "topic.delete",
  "protocol_version": "v1",
  "topic_id": "console:topic_123"
}
```

## 9. Runtime Matrix

| Runtime | `memory.prepare` | `memory.stm.upsert` | `topic.upsert` | `topic.append` | `topic.delete` |
| --- | --- | --- | --- | --- | --- |
| Telegram | 支持 | 支持 | 不支持 | 不支持 | 不支持 |
| Console | 支持 | 支持 | 支持 | 支持 | 支持 |
| Slack | 支持 | 支持 | 支持 | 支持 | 暂不要求 |

## 10. Implementation Tasks

### 10.1 协议与类型

建议落点：

- `integration/memoryplugin/`
- `schema/integration/memory-plugin/v1/`

任务：

1. 定义 5 个 hook 常量
2. 定义扁平请求/响应格式
3. 定义最小 input/output types
4. 落实际 `.schema.json` 文件
5. 确认 `nmem m add --id` 已可用

### 10.2 `memory.prepare`

代码落点：

- `internal/channelruntime/taskruntime/runtime.go`

任务：

1. 在现有 memory 注入点前调用 plugin
2. 将 `prompt_text` 直接接进现有 memory prompt block
3. plugin 失败只记日志，不阻断主流程

### 10.3 `memory.stm.upsert`

代码落点：

- `memory/projector.go`

任务：

1. `WriteShortTerm(...)` 成功后拿到最新 markdown
2. 计算 `memory_id`
3. 发 `memory.stm.upsert`
4. plugin 失败不影响本地投影成功

### 10.4 Console topic

代码落点：

- `cmd/mistermorph/consolecmd/local_runtime_history.go`
- `internal/daemonruntime/console_store.go`
- `internal/daemonruntime/server.go`

任务：

1. topic 首次出现时发 `topic.upsert`
2. 每轮新增 turn 后发 `topic.append`
3. 删除 topic 时发 `topic.delete`

### 10.5 Slack thread

代码落点：

- `internal/channelruntime/slack/runtime.go`
- `internal/channelruntime/slack/runtime_task.go`

任务：

1. 只在存在真实 `thread_ts` 时生成 `topic_id`
2. 首次出现 thread 时发 `topic.upsert`
3. 每轮新增消息后发 `topic.append`
4. 若未开启 task persistence，则回退到 live runtime hook

### 10.6 测试

至少补这些：

1. protocol types roundtrip
2. `memory_id` 规则
3. `topic_id` 规则
4. projector 成功写文件后触发 `memory.stm.upsert`
5. console topic create/append/delete
6. slack thread create/append
7. plugin 失败不阻断主流程

## 11. 推荐顺序

1. 先定 5 个 hook 和 schema
2. 接 `memory.prepare`
3. 接 `memory.stm.upsert`
4. 先做 Console `topic.upsert/append/delete`
5. 再做 Slack `topic.upsert/append`

## 12. 一句话总结

Phase 1 的协议只做两类对象：

- `memory`
  - 只处理短期文件
- `topic`
  - 统一表达 Console topic 和 Slack thread

对应 5 个 hook：

- `memory.prepare`
- `memory.stm.upsert`
- `topic.upsert`
- `topic.append`
- `topic.delete`
