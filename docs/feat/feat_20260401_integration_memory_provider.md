---
date: 2026-04-01
title: Integration Memory Provider External Contract (V1)
status: draft
---

# Integration Memory Provider External Contract (V1)

## 1) 目标

定义一套尽可能小、稳定、可对外说明的 memory contract，供第三方通过 `integration` 接入自定义 memory 系统。

这份文档只关注外部接口与协议，不讨论内部实现细节。

## 2) 第一性原理

从 runtime 的角度，memory 只需要完成两件事：

1. 在本轮 LLM 调用前，提供一段可注入 prompt 的 memory context
2. 在本轮结束后，接收一条关于本轮交互的 memory record

因此，V1 public contract 只定义两个动作：

- `Prepare`
- `Record`

其他能力都不是 V1 必需项：

- 不单独公开 `ShouldRecord`
- 不公开 `NotifyRecorded`
- 不要求 capabilities negotiation
- 不要求 event export / WAL 订阅
- 不要求定义跨进程固定传输协议

## 3) 适用范围

本 contract 面向两类接入方式：

- 使用 `integration` 嵌入 mistermorph 的第三方 Go 程序
- 通过独立二进制、子进程或外部服务接入的第三方系统

本 contract 不强制 provider 的实现方式：

- 可以是本进程内实现
- 可以是 RPC / HTTP / gRPC / subprocess 包装的远程实现
- 可以是数据库前的适配层

## 4) 非目标

V1 不解决以下问题：

- 不定义数据库 schema
- 不定义 WAL 文件格式
- 不定义 markdown projection 结构
- 不定义 memory 的后台投影、通知、索引任务
- 不定义复杂策略引擎
- 不定义 provider 能力发现协议

## 5) 生命周期语义

### 5.1 Prepare

`Prepare` 的语义是：

“针对当前这一轮请求，生成一段可以直接注入 prompt 的 memory text。”

它不是：

- 通用 memory 查询接口
- 原始事件导出接口
- 写入前校验接口

它必须满足：

- 只读
- 无副作用
- 同步
- 输出可直接注入 prompt

provider 可以自行决定如何生成这段内容：

- 只按 `subject_id` 简单召回
- 按 `subject_id + task_text` 做相关检索
- 结合更多上下文字段做复杂检索

这些都属于 provider 内部实现自由。

### 5.2 Record

`Record` 的语义是：

“将本轮交互提交给 memory 系统。”

它不是：

- projector 通知接口
- 事件总线广播接口
- 异步任务编排接口

provider 可以自行决定如何处理这条 record：

- 立即持久化
- 异步入队
- 策略跳过

runtime 只关心返回结果。

## 6) Go 接口

V1 建议的最小 Go 接口如下：

```go
type MemoryProvider interface {
    Prepare(ctx context.Context, req MemoryPrepareRequest) (MemoryPrepareResult, error)
    Record(ctx context.Context, req MemoryRecordRequest) (MemoryRecordResult, error)
}
```

这就是 V1 的全部核心接口。

## 7) 二进制第三方接入

除了 Go in-process 接口，V1 也允许通过独立二进制接入。

这里的“二进制接入”指：

- runtime 启动一个外部 provider 进程
- runtime 将 `Prepare` / `Record` 请求发送给该进程
- provider 进程返回对应响应

V1 不要求唯一传输方式，但推荐一个最小 profile：

- transport: `stdin/stdout`
- encoding: JSON
- interaction model: request/response

推荐原因：

- 最简单
- 与语言无关
- 易于调试
- 不要求额外网络监听

### 7.1 二进制接入的最小语义

外部 provider binary 只需要支持两个操作：

- `prepare`
- `record`

每次请求包含：

- `protocol_version`
- `op`
- `payload`

每次响应包含：

- `protocol_version`
- `ok`
- `result` 或 `error`

### 7.2 stdin/stdout profile 建议

建议约定：

- runtime 向 provider 的 `stdin` 写入一条完整 JSON
- provider 从 `stdout` 返回一条完整 JSON
- provider 的日志输出写到 `stderr`

如果需要多次请求，可以使用 line-delimited JSON：

- 一行一个 request
- 一行一个 response

V1 不要求 provider 支持并发多路复用。最简单实现可以串行处理。

## 8) 请求与响应结构

### 8.1 Prepare Request

V1 核心字段：

```go
type MemoryPrepareRequest struct {
    SubjectID      string
    SessionID      string
    RequestContext string
    TaskText       string
    MaxItems       int

    Extensions map[string]any
}
```

字段语义：

- `SubjectID`: memory 主体标识
- `SessionID`: 当前会话标识
- `RequestContext`: 请求上下文，例如 `private` / `public`
- `TaskText`: 当前任务文本
- `MaxItems`: 建议 provider 控制注入规模的上限
- `Extensions`: 可选扩展字段

为什么保留 `Extensions`：

- 简单 provider 可能只需要 `SubjectID`
- 复杂 provider 可能需要更多上下文
- 不应在 V1 一次性冻结全部字段

可能的扩展字段例子：

- `history`
- `current_message`
- `participants`
- `channel`
- `meta`

### 8.2 Prepare Result

```go
type MemoryPrepareResult struct {
    PromptText string
}
```

字段语义：

- `PromptText`: 可直接注入 prompt 的 memory 内容

如果 provider 没有命中任何内容，返回空字符串即可。

### 8.3 Record Request

V1 核心字段：

```go
type MemoryRecordRequest struct {
    SubjectID      string
    SessionID      string
    TaskRunID      string
    RequestContext string
    TaskText       string
    FinalOutput    string
    RecordedAt     time.Time

    Extensions map[string]any
}
```

字段语义：

- `SubjectID`: memory 主体标识
- `SessionID`: 当前会话标识
- `TaskRunID`: 当前运行标识
- `RequestContext`: 请求上下文
- `TaskText`: 当前任务文本
- `FinalOutput`: 本轮最终输出文本
- `RecordedAt`: 记录时间
- `Extensions`: 可选扩展字段

可能的扩展字段例子：

- `history`
- `participants`
- `channel`
- `meta`
- `final`

### 8.4 Record Result

```go
type MemoryRecordResult struct {
    Status   string
    RecordID string
}
```

`Status` 在 V1 建议只允许以下值：

- `persisted`
- `accepted_async`
- `skipped`

字段语义：

- `persisted`: 已完成持久化
- `accepted_async`: 已接收，后续异步处理
- `skipped`: provider 主动选择不记录
- `RecordID`: 可选的 provider 内部记录标识

V1 不要求更复杂的状态机。

## 9) 最小错误模型

V1 只区分两类错误：

- `Prepare` 返回错误
- `Record` 返回错误

不把以下概念提升为 public contract：

- policy failed
- notify failed
- post-record projection failed

这些都属于 provider 或 runtime 的内部实现细节。

## 10) Runtime 责任

runtime 对 `MemoryProvider` 的调用责任如下：

1. 构造 `MemoryPrepareRequest`
2. 调用 `Prepare`
3. 将 `PrepareResult.PromptText` 注入固定 memory prompt block
4. 执行本轮 LLM 调用
5. 构造 `MemoryRecordRequest`
6. 调用 `Record`

关键点：

- 第三方 provider 负责“memory 内容是什么”
- runtime 负责“何时调用”和“怎样注入 prompt”

这两者不应混在一起。

## 11) Integration 边界要求

对第三方公开的 memory contract，应该挂在“memory-aware runtime boundary”上，而不是单纯挂在裸 `agent.Engine` 上。

原因：

- `agent.Engine` 本身不知道 memory 生命周期
- memory 需要在 LLM 调用前后各有一次稳定调用点
- 如果只返回裸 engine，则第三方仍需自行保证这两个调用时机

因此，V1 对外要求的是：

- `integration` 必须提供一个会自动调用 `Prepare` / `Record` 的运行边界

至于具体是：

- 新的 `RunTaskWithMemory(...)`
- 新的 runtime option
- 还是新的 prepared runner 类型

这属于后续 API 设计问题，不在本文强行定死。

## 12) 二进制边界要求

如果第三方通过独立二进制接入，runtime 也必须提供一个对应的 provider adapter，使其能把二进制请求映射回同一套 `MemoryProvider` 语义。

也就是说：

- Go embedding 看到的是 `MemoryProvider`
- binary provider 看到的是 `prepare` / `record` 请求
- runtime 负责在两者之间适配

关键点：

- 外部 binary 不需要理解内部 hook 名称
- 外部 binary 不需要直接参与 prompt 拼接
- 外部 binary 只负责“生成 memory context”与“接收 memory record”

## 13) 协议映射原则

如果 provider 在进程外实现，任何 wire protocol 都应能无损表达同样的语义：

- `Prepare(request) -> result`
- `Record(request) -> result`

V1 不强制规定协议形态，但建议遵守以下原则：

- 显式版本号
- 请求/响应一一对应
- 核心字段与 Go struct 语义一致
- 可选字段放扩展区，避免频繁破坏协议

## 14) JSON 协议示例

这不是强制协议，只是说明语义。

### 14.1 Prepare Request

```json
{
  "protocol_version": "v1",
  "op": "prepare",
  "subject_id": "user:42",
  "session_id": "slack:T1:C1",
  "request_context": "private",
  "task_text": "reply to this message",
  "max_items": 20,
  "extensions": {}
}
```

### 14.2 Prepare Result

```json
{
  "protocol_version": "v1",
  "ok": true,
  "prompt_text": "[Memory]\\n- user prefers concise replies"
}
```

### 14.3 Record Request

```json
{
  "protocol_version": "v1",
  "op": "record",
  "subject_id": "user:42",
  "session_id": "slack:T1:C1",
  "task_run_id": "run_abc",
  "request_context": "private",
  "task_text": "reply to this message",
  "final_output": "Here is the final reply.",
  "recorded_at": "2026-04-01T10:00:00Z",
  "extensions": {}
}
```

### 14.4 Record Result

```json
{
  "protocol_version": "v1",
  "ok": true,
  "status": "persisted",
  "record_id": "mem_123"
}
```

### 14.5 Error Result

```json
{
  "protocol_version": "v1",
  "ok": false,
  "error": {
    "code": "provider_unavailable",
    "message": "database timeout"
  }
}
```

## 15) 第三方系统映射示例

### 15.1 MemoryOS 类系统

像 MemoryOS 这样的外部 memory 系统，接入时不需要让 runtime 直接理解它的内部模块或工具名。

只需要在它前面包一层 adapter，使其对 runtime 暴露同一套 V1 contract：

- `Prepare`
- `Record`

推荐映射方式：

- `Prepare`
  - 输入：`subject_id`、`session_id`、`task_text`、`request_context`
  - adapter 调用外部系统的检索能力
  - adapter 将检索结果整理成 `PromptText`
- `Record`
  - 输入：`subject_id`、`session_id`、`task_run_id`、`task_text`、`final_output`
  - adapter 调用外部系统的写入能力
  - adapter 返回 `persisted` / `accepted_async` / `skipped`

关键点：

- runtime 不直接调用外部系统自己的 MCP/tool 名称
- runtime 不把 prompt 拼接责任交给外部系统
- runtime 只依赖 `PrepareResult.PromptText` 和 `RecordResult.Status`

如果外部系统本身已经提供：

- 读取相关记忆
- 写入对话记忆

那么它就可以很自然地被包成一个 binary provider。

### 15.2 为什么示例只做“映射”而不做“直连”

因为不同第三方系统的内部模型不同：

- 有的系统是 query/retrieve 风格
- 有的系统是 profile/facts 风格
- 有的系统把 response generation 也包进 memory 框架里

runtime 不应该把这些内部差异暴露进自己的 public contract。

V1 的正确边界是：

- 第三方系统内部做自己的事情
- adapter 负责把它翻译成 `Prepare` / `Record`

## 16) 为什么不在 V1 引入更多概念

### 13.1 不引入 `MemoryPolicy`

因为 runtime 真正需要的不是一个“策略对象”，而是：

- 调用 `Prepare`
- 调用 `Record`

是否记录可以先由 runtime 自己决定，或者由 provider 在 `Record` 中返回 `skipped`。

### 13.2 不引入 `NotifyRecorded`

因为第三方真正关心的是：

- 这次 record 是否被接受
- 是否已经持久化

这已经可以通过 `MemoryRecordResult.Status` 表达。

### 13.3 不引入 capabilities

因为 V1 只有两个动作，必要信息已经在 request / result 中。

只有当未来真的出现协商需求时，再加 capabilities 更合理。

### 13.4 不引入 event export

因为“给 provider 提供 prepare/record”与“导出完整 memory 事件流”是两件不同的事。

后者未来可以单独定义，不应混入 V1 核心 contract。

## 17) 兼容性要求

V1 需要满足：

1. 第一方本地 memory 实现可以适配为 `MemoryProvider`
2. 第三方 hosted memory 实现也可以适配为 `MemoryProvider`
3. `Extensions` 允许未来平滑扩展
4. V1 不依赖当前内部 hook 名称
5. 第三方 binary provider 也可以映射到同一套语义

重点是：

- 对外暴露的是稳定语义
- 不是内部 wiring 名称

## 18) 后续工作

在本文基础上，下一步只需要补三件事：

1. `integration` 对外 API 落点
2. 第一方 local provider 适配层
3. 第三方 hosted / binary provider 示例

在这三件事完成前，不继续扩展更多概念。
