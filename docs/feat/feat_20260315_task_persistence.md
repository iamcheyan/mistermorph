---
date: 2026-03-15
title: 任务持久化（按 Channel 选择开启，JSONL 文件存储）
status: draft
---

# 任务持久化（按 Channel 选择开启，JSONL 文件存储）

## 1) 背景

当前任务状态主要是内存态：

- `console local runtime` 使用 `daemonruntime.MemoryStore`，重启后任务丢失。
- Telegram/Slack/LINE/Lark 的 daemon 任务视图也是内存态。
- `mistermorph serve` 的 `TaskStore` 也是内存态。

目标是在文件存储前提下，补齐任务持久化，并保持实现简单可落地。

## 2) 第一性原理

任务持久化在当前阶段只需要满足 4 个不可约约束：

1. 重启后任务不可丢。  
2. 文件存储下写入必须低成本且可靠。  
3. 当前读取核心是“按 topic 看会话 + 按时间倒序列出”，不是复杂聚合检索。  
4. 当前删除主诉求是“删 topic 可见性”，不是高频物理删单条任务。  

由这 4 条约束可直接推导出最小方案：

- 事实源使用 append-only JSONL（避免大文件全量覆写）。
- console 按天+topic 分文件；其他 channel 走统一日志并 rotation。
- topic 元数据与 task 分离（`topic.json` + `task.topic_id`）。
- 不做 projection，不做本地搜索索引；启动回放即可。
- `GET /tasks` 固定按 `updated_at desc` 返回。
- topic 删除走 tombstone，物理清理延后到 compact。
- 每条任务事件带 `trigger`，标识来源（ui/heartbeat/webhook/api/system）。

## 3) 本轮共识

1. 开关采用 `tasks.persistence_channels: ["console", "telegram"]` 这种白名单形式。  
2. 这期不做 projection。  
3. `TaskInfo` 结构只新增 `topic_id`，不引入 `topic_title/deleted_at/search_text` 等字段。  
4. 删除优先满足“删 topic”场景，并使用逻辑删除（tombstone）更稳妥。  
5. Console 任务日志文件名不需要 `seq`，同一 `topic_id` 不做 rotation。  

## 4) 非目标

- 这期不做内建全文搜索索引。
- 这期不做 embedding 检索库。
- 这期不做运行中任务断点续跑。

## 4.1 TaskStore 接管范围

新 `TaskStore` 接管所有任务元数据的内存态存储（`TaskInfo` 读写）：

- `internal/daemonruntime.MemoryStore`
- `cmd/mistermorph/daemoncmd.TaskStore` 中任务信息 map（执行队列除外）
- 各 channel runtime 对任务的 `Upsert/Update/Get/List`

说明：

- worker 并发控制和执行队列继续留在各 runtime/daemon 执行层。
- `TaskStore` 统一负责：持久化、回放、topic 过滤、排序输出。

## 5) 存储设计

### 5.1 目录布局

```text
<file_state_dir>/
  tasks/
    console/
      topic.json
      log/
        <YYYY-MM-DD>_<topic_key>.jsonl
    serve/
      log/
        since-YYYY-MM-DD-0001.jsonl
        since-YYYY-MM-DD-0002.jsonl
    telegram/
      log/
        since-YYYY-MM-DD-0001.jsonl
    slack/
      log/
        since-YYYY-MM-DD-0001.jsonl
    line/
      log/
        since-YYYY-MM-DD-0001.jsonl
    lark/
      log/
        since-YYYY-MM-DD-0001.jsonl
```

说明：

- `console`：按“日期 + topic”分文件，不做 rotation，不带 `seq`。
- 其他 channel：单流 `tasks.jsonl` 风格，按 memory/log 方式 rotation（`-0001/-0002`）。
- `topic_key` 由 `topic_id` 运行时计算（建议 URL-escape 或安全字符映射），避免文件名非法字符。

### 5.2 `topic.json`

`tasks/console/topic.json` 作为 topic 元数据文件：

```json
{
  "version": 1,
  "updated_at": "2026-03-15T12:00:00Z",
  "items": [
    {
      "id": "default",
      "title": "Default",
      "created_at": "2026-03-15T11:00:00Z",
      "updated_at": "2026-03-15T12:00:00Z",
      "deleted_at": ""
    }
  ]
}
```

语义：

- `deleted_at` 非空表示 topic 逻辑删除。
- 默认列表不返回已删除 topic。
- 不持久化 `topic_key`，由 `topic_id` 即时计算文件名映射。

### 5.3 JSONL 行格式（任务事件）

建议每行写“任务快照事件”，示例：

```json
{
  "type": "task_upsert",
  "at": "2026-03-15T12:00:00Z",
  "channel": "console",
  "trigger": {
    "source": "ui",
    "event": "chat_submit",
    "ref": "web/console"
  },
  "task": {
    "id": "console_xxx",
    "status": "done",
    "task": "hello",
    "model": "gpt-5.2",
    "timeout": "10m0s",
    "created_at": "2026-03-15T11:59:00Z",
    "started_at": "2026-03-15T11:59:01Z",
    "finished_at": "2026-03-15T11:59:10Z",
    "updated_at": "2026-03-15T11:59:10Z",
    "error": "",
    "result": {"output": "..."},
    "topic_id": "default"
  }
}
```

可预留（非必须）：

- `type: "task_deleted"`（未来若要支持单任务删除）。

`trigger` 字段约定：

- `source`: `ui | heartbeat | webhook | api | system`
- `event`: 触发动作名（例如 `chat_submit`、`heartbeat_tick`、`webhook_inbound`）
- `ref`: 可选来源标识（例如路由名、webhook path、job id）

默认值建议：

- console UI 提交：`source=ui`
- heartbeat 触发：`source=heartbeat`
- webhook 入站触发：`source=webhook`
- 外部 API 提交：`source=api`

与 console chat 呈现关系：

- 当前 chat 是“每个 task 对应一轮 user+assistant 展示”。
- 只要 `task.task/status/result/error/topic_id/created_at/updated_at` 完整，重启后即可重建现有 chat 视图。
- `trigger` 可用于后续在 UI 标记“该任务是否来自 UI/heartbeat/webhook”。

## 6) 配置设计

```yaml
tasks:
  dir_name: "tasks"
  persistence_channels: ["console"]
  rotate_max_bytes: 67108864
```

默认行为：

- 默认只有 `console` 开启持久化。
- 其他 channel 默认关闭，需要时加入白名单。

## 7) API 设计（本期）

### 6.1 提交任务

`POST /tasks`

新增可选字段：

- `topic_id`
- `topic_title`
- `trigger`（可选；未传则由 runtime 自动填充）

规则：

- 若 `topic_id` 为空：自动创建新 topic，并回填 `topic_id`。
- console 中每个 task 都必须带 `topic_id`。
- `topic_title` 只用于更新 `topic.json`，不写入 `TaskInfo`。
- 新 topic 首轮任务完成后，额外调用一次 LLM 生成 title，并异步更新 `topic.json`（不阻塞主任务返回）。
- `trigger` 未传时按运行上下文自动写入（例如 console chat 默认 `source=ui`）。

### 6.2 列表任务

`GET /tasks`

保留：

- `status`
- `limit`
- `topic_id`（console 场景）

说明：

- 这期不加内建全文搜索参数。
- 搜索需求先通过文件工具（`rg/jq`）解决。
- 返回顺序固定为 `updated_at desc`。

### 6.3 Topic 列表

`GET /tasks/topics`

返回 `topic.json` 中未删除的 topic。

### 6.4 删除 Topic

`DELETE /tasks/topics/{topic_id}`

语义：

- 逻辑删除 topic（写 `topic.json.deleted_at`）。
- 默认任务列表过滤该 topic 下任务。
- 本期不做物理清理；后续可加 `compact` 命令做离线清理。

## 8) 启动恢复语义

1. 读取本 channel 对应日志文件并回放到内存视图。  
2. 对非终态任务（`queued/running/pending`）统一收敛为 `canceled`，错误为 `runtime restarted`。  
3. 将收敛后的快照追加写回对应 JSONL。  

`updated_at` 规则：

- 每次状态变化或结果更新都刷新 `updated_at`。
- `GET /tasks` 按 `updated_at` 降序返回。

## 9) 为什么本期不做 Projection

本期范围下，projection 不是必需：

- 任务量阶段性可控；
- 暂无强搜索需求；
- 使用文件工具即可满足排查；
- 不做 projection 可显著降低实现复杂度和维护成本。

后续如果出现以下信号，再引入 projection：

- 任务量持续增长导致启动回放明显变慢；
- Console 内建搜索/聚合需求变重；
- 需要更稳定的分页统计能力。

## 10) 风险与控制

- 风险：console 单 topic 文件长期增长。  
控制：按天分文件（`YYYY-MM-DD_topic`），天然切分。

- 风险：topic 逻辑删除后文件仍在。  
控制：先逻辑删除保证安全，后续提供离线 compact。

- 风险：topic_id 含非法文件名字符。  
控制：统一 `topic_key` 映射规则；`topic.json` 只保留原始 `topic_id`。

## 11) DoD

- `tasks.persistence_channels` 生效，默认仅 `console`。  
- console 可按 topic 写入并在重启后可见。  
- console topic 删除为逻辑删除，并在列表层生效。  
- 其他 channel 可选开启，写入统一 `tasks.jsonl` 轮转日志。  
- 文档与配置模板同步（`assets/config/config.example.yaml`、`docs/console.md`）。  
