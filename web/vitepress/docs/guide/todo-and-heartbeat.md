---
title: TODO and Heartbeat
description: How TODO files and HEARTBEAT.md let the agent track work outside the current chat.
---

# TODO and Heartbeat

Heartbeat is the runtime trigger for recurring checks. It can run on a timer or be started by an external poke. Each heartbeat run creates a fresh runtime task and does not include chat history.

## Heartbeat Flow

On each heartbeat tick or poke:

1. The runtime reads `TODO.RECUR.md`.
2. Due recurring todos are copied into `TODO.md`.
3. Their `Next` timestamp is advanced.
4. The runtime reads `TODO.md`.
5. If there are open todos, they are added to the heartbeat task.
6. The runtime reads `HEARTBEAT.md`.
7. If `HEARTBEAT.md` is not empty, it is added to the heartbeat task.
8. If the task has any content, the agent handles it with normal tools, including `todo_update`.

If `TODO.RECUR.md`, `TODO.md`, and `HEARTBEAT.md` produce no task content, no agent task is started.

## HEARTBEAT.md

`HEARTBEAT.md` is the standing instruction for each heartbeat. It should describe what the agent should check, not a one-off user request.

Good uses:

- Check open todos.
- Look for due follow-ups.
- Inspect routine files.
- Send reminders when a todo asks for it.

Avoid putting one-off tasks directly into `HEARTBEAT.md`. Put those in `TODO.md`, or use `TODO.RECUR.md` if they repeat.

## TODO Flow

TODO files hold concrete work. The `todo_update` tool writes and completes TODO records. During heartbeat, current open `TODO.md` items are added to the heartbeat task so the agent can act on them.

There are three TODO files.

### TODO.md

`TODO.md` contains work that should happen once:

```text
- [ ] [Created](2026-05-01 12:41), [ChatID](tg:-100123) | Remind [John](tg:@john) to submit report.
```

Use `TODO.md` for reminders and tasks that should be handled once.

### TODO.DONE.md

`TODO.DONE.md` contains completed one-off todos. When `todo_update` completes a `TODO.md` item, it moves the record here.

Recurring todos do not move to `TODO.DONE.md`.

### TODO.RECUR.md

`TODO.RECUR.md` contains repeat rules:

```text
- [ ] [Next](2026-05-07 15:00), [Repeat](weekly), [TZ](Asia/Tokyo) | Play tennis.
- [ ] [Next](2026-05-02 09:00), [Repeat](every 6 hours) | Check the report queue.
```

Supported repeat values:

- `daily`
- `weekly`
- `every N days`
- `every N hours`

`TZ` is optional. If it is omitted, the runtime local timezone is used.

Recurring records stay in `TODO.RECUR.md`. On heartbeat, due records are copied into `TODO.md`, and only their `Next` timestamp moves forward.

## Choosing the File

| Need | File |
|---|---|
| Tell the agent what to check on each heartbeat | `HEARTBEAT.md` |
| Do this once | `TODO.md` |
| Keep a record of completed one-off todos | `TODO.DONE.md` |
| Do this repeatedly | `TODO.RECUR.md` |

For the tool that updates todo files, see [`todo_update`](/guide/built-in-tools#todo_update). For state directory placement, see [Filesystem Roots](/guide/filesystem-roots).
