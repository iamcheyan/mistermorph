[[ TODO Workflow ]]
Use this workflow only when you need to remember something for future work, or mark an existing todo item as completed.
Maintain `TODO.md` and `TODO.DONE.md` under `file_state_dir`.
Recurring tasks live in `TODO.RECUR.md` under `file_state_dir`.

`TODO.md` entry format examples:
```text
- [ ] [Created](2026-02-11 09:30) | at 2026-02-11 10:00 Remind [John](tg:@johnwick) to submit the report.
- [ ] [Created](2026-02-11 09:30), [ChatID](tg:-1001981343441) | 2026-02-11 10:00 Have lunch with [John](tg:@johnwick), Miss Louis and [Sarah](tg:29930) at the Italian restaurant.
```

`TODO.DONE.md` entry format examples:
```text
- [x] [Created](2026-02-11 09:30), [Done](2026-02-11 10:00) | at 2026-02-11 10:00 Remind [John](tg:@johnwick) to submit the report.
- [x] [Created](2026-02-11 09:30), [Done](2026-02-11 10:00), [ChatID](tg:-1001981343441) | 2026-02-11 10:00 Had lunch with [John](tg:@johnwick), Miss Louis and [Sarah](tg:29930) at the Italian restaurant.
```

`TODO.RECUR.md` entry format examples:
```text
- [ ] [Next](2026-02-12 09:00), [Repeat](daily), [TZ](Asia/Tokyo), [ChatID](tg:-1001981343441) | Remind [John](tg:@johnwick) to submit the report.
- [ ] [Next](2026-02-16 10:00), [Repeat](weekly) | Review open invoices.
- [ ] [Next](2026-02-14 18:00), [Repeat](every 3 days) | Back up notes.
- [ ] [Next](2026-02-14 18:00), [Repeat](every 6 hours) | Check the feeder.
```

- If a new task is identified, use `todo_update` to add it to `TODO.md`.
- If a new recurring task is identified, use `todo_update` with action `add_recurring`. Pass `content`, `next` (`YYYY-MM-DD HH:mm`), `repeat`, optional `tz`, and optional `chat_id`; supported repeat values are `daily`, `weekly`, `every N days`, and `every N hours`.
- If the user states a timezone, write it as an IANA timezone in `TZ` (for example `Asia/Tokyo`). If no timezone is stated, omit `TZ`; the runtime local timezone is used.
- If a task is expired, notify mentioned contacts via `contacts_send` with a concise reminder. Do not mention TODO files, pending counts, or delivery status. Then use `todo_update` to complete the task.
- If a task is not due, do nothing.
