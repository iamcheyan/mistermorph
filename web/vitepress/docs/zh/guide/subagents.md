---
title: Subagent 与子任务
description: "说明什么时候该用 `spawn`，什么时候该用 `bash.run_in_subtask`，以及运行时到底保证什么。"
---

# Subagent 与子任务

Mistermorph 现在有两种显式的子任务入口：

- `spawn`：启动一个带独立 LLM loop 的子 agent，并显式限制它能用的工具。
- `bash.run_in_subtask=true`：把一条 shell 命令放进 direct subtask 边界里执行，不会再起第二轮 LLM。

两条路径最后都会返回同一种 `SubtaskResult` envelope，也共用同一个深度限制。

## 什么时候用哪一种

- 子任务还需要自己做工具推理时，用 `spawn`。例如先 `url_fetch` 再抽取，或者组合 `read_file`、`url_fetch`、`bash` 这类步骤。
- 事情本身已经是一条明确 shell 命令时，用 `bash.run_in_subtask=true`。
- 很小的一步工作，父任务自己直接做完就行，不要硬拆子任务。

当前状态也要说清楚：这两条路径现在都是同步阻塞的。父任务会等子任务结束。它解决的是隔离和收敛，不是后台并发执行。

## `spawn`

`spawn` 是 engine 级工具。只有某次 agent engine 真正装配出来以后，它才会出现。

参数：

- `task`：必填，子任务提示词。
- `tools`：必填，非空字符串数组，表示子任务允许使用的工具名。
- `model`：可选，子任务模型覆盖；默认继承父任务模型。
- `output_schema`：可选，结构化输出的约定名。
- `observe_profile`：可选，观察提示。目前支持 `default`、`long_shell`、`web_extract`。

运行时行为：

- 子任务 registry 只会从你传入的 `tools` 里挑工具。未知工具名或父 registry 里没有的工具名会被忽略。
- 如果最后一个可用工具都没剩下，调用会失败。
- 即使你把 `spawn` 自己放进 `tools`，它也不会再暴露给子任务。
- 当前深度限制是 `1`，也就是子任务里不能继续起下一层子任务。
- 默认不会把子任务的原始 transcript 回灌给父 loop。

### `output_schema` 到底是什么

`output_schema` 只是一个结构化输出的标识，不是内建的 JSON Schema 注册中心。

如果你传了它：

- 子任务会被提示必须产出 JSON 最终输出；
- 运行时会要求最终输出是 JSON，或者至少是可解析的 JSON 字符串；
- 返回 envelope 时会把同一个标识原样放回 `output_schema`。

Mistermorph 现在不会替你拿真实 schema 去校验对象字段。

## 返回 Envelope

`spawn` 和 direct subtask 最后都会返回这种 JSON：

```json
{
  "task_id": "sub_123",
  "status": "done",
  "summary": "subtask completed",
  "output_kind": "text",
  "output_schema": "",
  "output": "child result",
  "error": ""
}
```

字段含义：

- `status`：现在主要是 `done` 或 `failed`。
- `summary`：给父任务侧做进度预览或简短状态展示用的短文本。
- `output_kind`：`text` 或 `json`。
- `output_schema`：纯文本输出时为空；结构化输出时回显你传入的标识。
- `output`：子任务真正的结果。
- `error`：子任务失败时才会有内容。

## `bash.run_in_subtask=true`

这是更轻的一条子任务路径。

- 它直接走 subtask runner，不会再起第二轮 LLM。
- 它的 `output_schema` 固定是 `subtask.bash.result.v1`。
- 它的观察 profile 固定是 `long_shell`。
- 它的 `output` 里会带 `exit_code`、截断标记、`stdout` 和 `stderr`。

返回 payload 例子：

```json
{
  "task_id": "sub_456",
  "status": "done",
  "summary": "bash exited with code 0",
  "output_kind": "json",
  "output_schema": "subtask.bash.result.v1",
  "output": {
    "exit_code": 0,
    "stdout_truncated": false,
    "stderr_truncated": false,
    "stdout": "hello\n",
    "stderr": ""
  },
  "error": ""
}
```

如果你只是想把一条长命令包进独立子任务边界里，但不需要子任务自己继续调用别的工具，就用这条路径。

## 配置与嵌入

- `tools.spawn.enabled` 只控制显式 `spawn` 工具入口。
- 即使 `tools.spawn.enabled=false`，像 `bash.run_in_subtask=true` 这种 direct subtask 仍然会走 subtask runtime。
- `integration.Config.BuiltinToolNames` 可以包含 `spawn`，也可以去掉它；它不只控制静态工具。
- 如果你直接用 `agent.New(...)` 组 engine，`spawn` 默认是开启的；需要关闭时用 `agent.WithSpawnToolEnabled(false)`。

示例：

```go
cfg := integration.DefaultConfig()
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "spawn"}
cfg.Set("tools.spawn.enabled", true)
```

如果你没把 `spawn` 放进 `BuiltinToolNames`，Agent 表面上就没有显式 child-agent 工具了；但底层 subtask runtime 仍然可以被 `bash.run_in_subtask=true` 这类内部入口使用。

延伸阅读：

- [内置工具](/zh/guide/built-in-tools)
- [创建自己的 AI Agent：进阶](/zh/guide/build-your-own-agent-advanced)
- [配置字段](/zh/guide/config-reference)
