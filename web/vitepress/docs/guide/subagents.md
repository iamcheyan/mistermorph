---
title: Subagents and Subtasks
description: "When to use `spawn`, when to use `bash.run_in_subtask`, and what the runtime guarantees."
---

# Subagents and Subtasks

Mistermorph currently has two explicit ways to put work behind a child-task boundary:

- `spawn`: starts a child agent with its own LLM loop and an explicit tool whitelist.
- `bash.run_in_subtask=true`: runs one shell command inside a direct subtask boundary without starting a second LLM loop.

Both paths return the same `SubtaskResult` envelope shape and share the same depth limit.

## Which entry to use

- Use `spawn` when the child work still needs agent-style tool reasoning, such as `read_file` plus `url_fetch`, or a fetch-then-extract flow.
- Use `bash.run_in_subtask=true` when the work is already one concrete shell command or script.
- Do not use a child task for trivial one-step work that the parent can complete directly.

Current status: both paths are synchronous. The parent waits until the child finishes. This is isolation, not background execution.

## `spawn`

`spawn` is an engine-scoped tool. It appears when an agent engine is assembled for a run.

Parameters:

- `task`: required child prompt.
- `tools`: required non-empty array of tool names the child may use.
- `model`: optional child model override. Defaults to the parent model.
- `output_schema`: optional contract label for structured output.
- `observe_profile`: optional observer hint. Supported values are `default`, `long_shell`, and `web_extract`.

Runtime behavior:

- The child registry starts from the `tools` names you pass. Unknown or unavailable names are ignored.
- If none of the requested tool names are available in the parent registry, the call fails.
- `spawn` is never re-exposed inside the child, even if you include it in `tools`.
- The current depth limit is `1`, so a child task cannot start another child task.
- Raw child transcript is not pushed back into the parent loop by default.

### About `output_schema`

`output_schema` is a schema identifier, not a built-in JSON Schema registry.

If you set it:

- the child is told to produce JSON final output;
- the runtime requires the final output to be JSON or JSON-parsable text;
- the same identifier is echoed back in `output_schema`.

Mistermorph does not validate the returned object against a real schema definition for you.

## Result Envelope

Both `spawn` and direct subtasks return JSON in this shape:

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

Field meanings:

- `status`: currently `done` or `failed`.
- `summary`: short status text suitable for parent-side progress or preview.
- `output_kind`: `text` or `json`.
- `output_schema`: empty for plain text output, or the identifier you passed in.
- `output`: child result payload.
- `error`: non-empty only when the child fails.

## `bash.run_in_subtask=true`

This is the lighter child-task path.

- It uses the subtask runner directly and does not start a second LLM loop.
- Its `output_schema` is fixed to `subtask.bash.result.v1`.
- Its observer profile is fixed to `long_shell`.
- The `output` payload contains `exit_code`, truncation flags, `stdout`, and `stderr`.

Example payload:

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

Use this when you want a separate child-task envelope around one shell step, but do not need the child to make more tool calls.

## Config and Embedding

- `tools.spawn.enabled` controls only the explicit `spawn` tool entry.
- Direct subtasks such as `bash.run_in_subtask=true` still use the subtask runtime even if `tools.spawn.enabled=false`.
- `integration.Config.BuiltinToolNames` can include or omit `spawn`. It is not limited to static tools.
- If you build an engine directly with `agent.New(...)`, `spawn` is enabled by default. Disable it with `agent.WithSpawnToolEnabled(false)`.

Example:

```go
cfg := integration.DefaultConfig()
cfg.BuiltinToolNames = []string{"read_file", "url_fetch", "spawn"}
cfg.Set("tools.spawn.enabled", true)
```

If you omit `spawn` from `BuiltinToolNames`, the agent loses the explicit child-agent tool, but the underlying subtask runtime can still be used by internal callers such as `bash.run_in_subtask=true`.

See also:

- [Built-in Tools](/guide/built-in-tools)
- [Create Your Own AI Agent: Advanced](/guide/build-your-own-agent-advanced)
- [Config Fields](/guide/config-reference)
