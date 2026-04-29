---
date: 2026-04-24
title: Log Observability（日志观测）
status: draft
---

# Log Observability（日志观测）

## 1) 背景

当前日志主要通过 `slog` 输出到 stderr。桌面端和 Console Web 可以看到任务状态、审计日志、统计数据，但不能直接看到当前进程的运行日志。

这导致几个实际问题：

1. 桌面端出错时，用户很难在界面里看到最近发生了什么。
2. Console API、runtime、provider、tool 调用相关的错误分散在进程输出里，不便排查。
3. 现有 `logging.*` 已经定义了级别、格式、脱敏和截断策略，但缺少文件输出和保留策略。

本需求要补齐一个最小可用的日志观测能力：

- 后端继续把日志输出到 stderr。
- 同一份 `slog` 记录再写一份到本地日志文件。
- Console Web 能读取并显示最新日志。

## 2) 目标

1. 在配置里支持日志输出目录。
2. 当日志输出目录为空时，使用 `<file_state_dir>/logs/`。
3. 日志按天写入，每天一个文件。
4. 日志文件有最大保存时间，默认 7 天，可配置。
5. 桌面端通过 Console API 读取最新日志。
6. Console Web 增加日志视图，显示最新日志内容。
7. 不改变现有 stderr 输出行为。

默认行为：

- 文件日志默认启用，不增加 `logging.file.enabled` 开关。
- 只要进程使用 `internal/logutil` 构建 logger，就会写入文件日志。

适用范围：

- 使用 `internal/logutil` 构建 logger 的 CLI、Console 后端和桌面端后端进程。
- 嵌入方如果显式传入自己的 logger，不强制接管。

## 3) 非目标

第一版不做这些事：

1. 不做全文搜索和索引。
2. 不做日志级别筛选。
3. 不做多进程集中采集。
4. 不做跨 endpoint 的日志聚合视图。
5. 不做 WebSocket/SSE 实时日志流。
6. 不提供任意文件读取能力。
7. 不改变 `logging.include_*`、`logging.redact_keys` 等现有日志内容策略。

## 4) 配置需求

在现有 `logging` 段下增加文件输出配置：

```yaml
logging:
  level: "info"
  format: "text"
  add_source: false

  file:
    # Empty means <file_state_dir>/logs.
    dir: ""
    # Go duration string. Default is 7 days.
    max_age: "168h"
```

字段语义：

- `logging.file.dir`
  - 类型：string
  - 为空时使用 `<file_state_dir>/logs/`
  - 非空时使用用户指定目录
  - 相对路径按现有配置路径规则处理，具体规则应与仓库内其他本地路径配置保持一致

- `logging.file.max_age`
  - 类型：duration string
  - 默认值：`168h`
  - 必须是正数
  - 早于 `now - max_age` 的日志文件会被清理

需要同步更新：

1. `internal/configdefaults/defaults.go`
2. `assets/config/config.example.yaml`
3. 相关配置读取结构

文件日志固定使用 JSONL，不新增 `logging.file.format`。`logging.format` 继续只控制 stderr 输出。

## 5) 日志文件需求

### 5.1 输出行为

日志输出应同时写入：

1. 原有 stderr handler。
2. 当前日期对应的日志文件 handler。

文件输出和 stderr 输出使用同一套：

- `logging.level`
- `logging.add_source`
- 已有日志脱敏与截断逻辑

格式规则：

- stderr 继续使用 `logging.format: text|json`。
- 文件日志固定使用 JSONL，也就是一行一条 `slog.JSONHandler` 记录。

这样做的原因是：stderr 主要给人看，文件日志主要给系统读取。JSONL 适合 tail、分页和后续字段筛选，也不需要正则解析 text log。

### 5.2 文件目录

默认目录：

```text
<file_state_dir>/logs/
```

目录不存在时自动创建。

权限要求：

- 日志目录建议使用仅当前用户可读写的权限。
- 日志文件建议使用仅当前用户可读写的权限。

原因是日志可能包含任务内容、错误信息、工具调用摘要和运行环境信息。

### 5.3 文件命名

每天一个文件，文件名建议：

```text
mistermorph-YYYY-MM-DD.jsonl
```

日期使用进程本地时区。

示例：

```text
mistermorph-2026-04-24.jsonl
```

文件内固定为 JSONL。Console Web 第一版可以先按原始行显示，不强制做结构化表格。

### 5.4 日切换

当日期变化时，新日志写入新日期文件。

实现上不要求后台定时任务。可以在写日志时检查当前日期：

1. 当前日期未变：继续写当前文件。
2. 当前日期已变：关闭旧文件，打开新文件。
3. 打开新文件后执行一次清理。

### 5.5 保留策略

保留策略只清理日志系统自己创建的文件。

清理范围：

- 只处理 `logging.file.dir` 下匹配 `mistermorph-YYYY-MM-DD.jsonl` 的文件。
- 不删除同目录下的其他文件。

清理时机：

1. 进程启动时执行一次。
2. 日切换时执行一次。

清理规则：

- 按文件名日期清理，不按文件 mtime 清理。
- 超过 `logging.file.max_age` 的文件应删除。
- 删除失败时记录 warn，不阻断主进程。

## 6) Runtime API + Console Proxy 需求

新增一个 runtime API，用于读取当前 runtime 的最新日志。

建议路径：

```text
GET /logs/latest
```

Console Web 通过现有 `/api/proxy` 调用：

```text
GET /api/proxy?endpoint=<endpoint_ref>&uri=/logs/latest
```

原因：

- 日志查看是 runtime 能力，和 `/audit/logs`、`/stats/llm/usage` 一样。
- 本地桌面端可以看本地 runtime 日志。
- 远端 endpoint 如果升级支持该 API，也可以通过同一套 Console UI 查看。
- 老版本 endpoint 不支持时返回 `404`，前端显示“不支持日志查看”。

允许查看已鉴权远端 endpoint 的日志。访问控制依赖现有 Console session、endpoint auth 和部署侧访问边界。

### 6.1 鉴权

要求：

- Console 侧仍使用现有 session 鉴权保护 `/api/proxy`。
- runtime 侧使用现有 endpoint auth。
- 未登录或 endpoint auth 失败返回 `401`。
- 不挂到 `/health` 这类公开接口下。

### 6.2 查询参数

```text
GET /logs/latest?limit=300
GET /logs/latest?limit=300&cursor=<opaque_cursor>
```

参数：

- `limit`
  - 默认：`300`
  - 最小：`1`
  - 最大：`1000`
  - 表示最多返回多少行

- `cursor`
  - 默认：空
  - 空表示读取当前最新日志尾部
  - 非空表示继续读取更早的日志
  - cursor 是不透明字符串，前端不解析

cursor 内部可以编码 `file` 和 `before` 行号。这样跨日切换时，前端仍然能继续读取上一份文件的旧日志。

分页允许跨天：

- 当前文件没有更早日志时，可以继续读取上一天的 `mistermorph-YYYY-MM-DD.jsonl`。
- 继续向上加载时，可以按日期倒序读取更早文件，直到达到保留范围内最早日志。

### 6.3 响应结构

建议响应：

```json
{
  "file": "mistermorph-2026-04-24.jsonl",
  "exists": true,
  "size_bytes": 123456,
  "mod_time": "2026-04-24T10:30:00Z",
  "limit": 300,
  "total_lines": 1200,
  "from": 901,
  "to": 1200,
  "has_older": true,
  "older_cursor": "opaque",
  "lines": [
    "{\"time\":\"2026-04-24T10:29:59Z\",\"level\":\"INFO\",\"msg\":\"run_start\"}",
    "{\"time\":\"2026-04-24T10:30:00Z\",\"level\":\"INFO\",\"msg\":\"final\"}"
  ]
}
```

注意：

- 响应不返回日志目录绝对路径。
- `file` 只返回 basename。
- 当没有日志文件时，返回 `200`，`exists: false`，`lines: []`。
- 当配置无法解析或日志目录无法访问时，返回 `500`，并给出简短错误。
- `older_cursor` 只在 `has_older: true` 时返回。

### 6.4 最新文件选择

“最新日志”定义为：

1. 优先读取当前日期对应的日志文件。
2. 如果当前日期文件不存在，则读取日志目录下最新的 `mistermorph-YYYY-MM-DD.jsonl`。
3. 如果没有匹配文件，则返回空结果。

## 7) Console Web 需求

新增 Logs 视图，用来显示最新日志。

建议：

- 路由：`/logs`
- 入口：`Settings -> Console -> Logs`，点击后进入独立日志视图
- 数据源：`runtimeApiFetch("/logs/latest")`
- 通过当前 endpoint 读取日志

### 7.1 页面内容

页面至少显示：

1. 当前日志文件名。
2. 文件更新时间。
3. 文件大小。
4. 最新日志行。
5. 行数选择，例如 `100 / 300 / 1000`。
6. 默认自动刷新。
8. 加载更早日志的按钮。

### 7.2 展示要求

日志内容按原始行显示：

- 使用等宽字体。
- 保留空格。
- 长行允许横向滚动或软换行，但不能撑坏页面。
- 新日志默认显示在底部。
- 用户手动向上查看旧日志时，自动刷新不应强制滚回底部。

分页和无限加载：

- 初次打开读取最新 `limit` 行，并滚到底部。
- 用户滚到顶部或点击“加载更早”时，使用 `older_cursor` 读取更早日志。
- 新旧日志拼接时保持当前滚动位置，不让页面跳动。
- 自动刷新只在用户接近底部时滚到底部。
- 用户正在看旧日志时，自动刷新只提示有新日志，不强制跳转。

### 7.3 空状态和错误状态

空状态：

- 当 `exists: false` 时，展示“暂无日志”。
- 当 endpoint 返回 `404` 时，展示“当前 endpoint 不支持日志查看”。

错误状态：

- 鉴权失败走现有登录流程。
- 其他错误展示错误消息和重试按钮。

### 7.4 安全要求

前端不展示本地绝对路径。

日志内容本身可能包含敏感信息，因此：

- Logs 视图只能给已登录 Console 用户使用。
- 不提供公开分享链接。
- 不把日志内容写入浏览器本地长期存储。

## 8) 桌面端行为

桌面端当前通过本地 Console 后端提供 API 和 SPA。第一版不需要新增 Wails binding。

桌面端只需要满足：

1. 启动本地 Console 后端时，后端按配置写日志文件。
2. Wails WebView 通过已有代理访问 `/api/proxy?endpoint=...&uri=/logs/latest`。
3. Logs 视图展示当前本地后端的最新日志。

如果后续出现“桌面壳自身日志”和“后端日志”需要同时展示的问题，再单独设计。第一版只处理后端日志。

## 9) 实施建议

### Phase A：文件日志

- [x] 扩展 `internal/logutil.LoggerConfig`。
- [x] 读取 `logging.file.dir` 和 `logging.file.max_age`。
- [x] 实现一个 fan-out `slog.Handler`，同时写 stderr 和文件。
- [x] 实现按日打开文件与旧文件清理。
- [x] 增加单元测试：
  - 默认目录解析
  - 自定义目录解析
  - 日切换
  - 超期清理
  - 不删除非日志文件

### Phase B：Console API

- [x] 在 runtime API 中增加 `GET /logs/latest`。
- [x] 实现最新日志文件选择。
- [x] 实现尾部读取与 cursor 分页。
- [x] 响应中只返回 basename，不返回绝对路径。
- [x] 增加 handler 测试：
  - 未登录返回 `401`
  - 无日志文件返回空结果
  - 默认读取最新 300 行
  - `limit` 边界
  - cursor 翻页
  - 不返回绝对路径

### Phase C：Console Web

- [x] 新增 `LogsView`。
- [x] 新增路由 `/logs`。
- [x] 在 `Settings -> Console` 中增加 Logs 入口。
- [x] 通过 `runtimeApiFetch("/logs/latest")` 接入 runtime API。
- [x] 增加默认自动刷新、行数选择、加载更早或向上无限加载。
- [x] 构建验证：`pnpm build`。

## 10) 验收标准

1. 未配置 `logging.file.dir` 时，日志写入 `<file_state_dir>/logs/`。
2. 配置 `logging.file.dir` 后，日志写入指定目录。
3. 每天只写当天对应的日志文件。
4. 超过 `logging.file.max_age` 的旧日志会被清理。
5. stderr 日志仍然存在。
6. `GET /logs/latest` 只能通过已鉴权的 Console proxy 和 endpoint auth 访问。
7. API 响应不包含本地绝对路径。
8. Console Web 的 Logs 视图能显示最新日志。
9. Logs 视图刷新后能看到新写入的日志行。
10. 没有日志文件时，UI 显示空状态而不是报错。
11. Logs 视图可以通过 cursor 加载更早日志。
