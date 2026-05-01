---
title: 待办事项与 Heartbeat
description: TODO 文件和 HEARTBEAT.md 如何让 agent 在当前对话之外继续跟踪工作。
---

# 待办事项与 Heartbeat

Heartbeat 是用于重复检查的 runtime 触发。它可以按间隔运行，也可以由外部 poke 启动。每次 heartbeat 都会创建一条新的 runtime task，不包含聊天历史。

## Heartbeat 流程

每次 heartbeat tick 或 poke 时：

1. Runtime 读取 `TODO.RECUR.md`。
2. 到期的循环待办会被复制到 `TODO.md`。
3. 对应记录的 `Next` 时间会向后推进。
4. Runtime 读取 `TODO.md`。
5. 如果有未完成待办，就加入 heartbeat task。
6. Runtime 读取 `HEARTBEAT.md`。
7. 如果 `HEARTBEAT.md` 不是空的，就加入 heartbeat task。
8. 如果 heartbeat task 有内容，agent 用普通工具处理这次任务，包括 `todo_update`。

如果 `TODO.RECUR.md`、`TODO.md` 和 `HEARTBEAT.md` 都没有产生任务内容，就不会启动 agent task。

## HEARTBEAT.md

`HEARTBEAT.md` 是每次 heartbeat 的固定说明。它应该描述 agent 要检查什么，而不是保存某个一次性用户请求。

适合放在这里的内容：

- 检查打开的待办事项。
- 查找到期的后续跟进。
- 检查例行文件。
- 当待办事项要求提醒时，发送提醒。

不要把一次性任务直接写进 `HEARTBEAT.md`。一次性任务放进 `TODO.md`；重复任务放进 `TODO.RECUR.md`。

## TODO 流程

TODO 文件保存具体待办。`todo_update` 工具负责写入和完成 TODO 记录。Heartbeat 运行时，当前 `TODO.md` 里的未完成待办会被加入 heartbeat task，让 agent 有机会处理它们。

TODO 文件有三个。

### TODO.md

`TODO.md` 保存只需要执行一次的工作：

```text
- [ ] [Created](2026-05-01 12:41), [ChatID](tg:-100123) | Remind [John](tg:@john) to submit report.
```

一次性提醒和一次性任务放在 `TODO.md`。

### TODO.DONE.md

`TODO.DONE.md` 保存已经完成的一次性待办。`todo_update` 完成 `TODO.md` 里的项目时，会把记录移动到这里。

循环待办不会移动到 `TODO.DONE.md`。

### TODO.RECUR.md

`TODO.RECUR.md` 保存重复规则：

```text
- [ ] [Next](2026-05-07 15:00), [Repeat](weekly), [TZ](Asia/Tokyo) | Play tennis.
- [ ] [Next](2026-05-02 09:00), [Repeat](every 6 hours) | Check the report queue.
```

当前支持的 `Repeat` 值：

- `daily`
- `weekly`
- `every N days`
- `every N hours`

`TZ` 可选。省略时使用 runtime 的本地时区。

循环记录会留在 `TODO.RECUR.md`。Heartbeat 运行时，到期记录会被复制到 `TODO.md`，只有 `Next` 时间向后推进。

## 该写到哪里

| 需求 | 文件 |
|---|---|
| 告诉 agent 每次 heartbeat 要检查什么 | `HEARTBEAT.md` |
| 只做一次 | `TODO.md` |
| 保存已完成的一次性待办 | `TODO.DONE.md` |
| 重复执行 | `TODO.RECUR.md` |

更新 TODO 文件的工具见 [`todo_update`](/zh/guide/built-in-tools#todo_update)。状态目录的位置见 [文件系统根目录](/zh/guide/filesystem-roots)。
