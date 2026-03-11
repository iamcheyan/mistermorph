[[ TODO Workflow ]]
Use this workflow only when you need to remember something for future work, or mark an existing todo item as completed.
Maintain `TODO.md` and `TODO.DONE.md` under `file_state_dir`.

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

- If a new task is identified, use `todo_update` to add it to `TODO.md`.
- If a task is expired, notify mentioned contacts via `contacts_send` with a concise reminder. Do not mention TODO files, pending counts, or delivery status. Then use `todo_update` to complete the task.
- If a task is not due, do nothing.
