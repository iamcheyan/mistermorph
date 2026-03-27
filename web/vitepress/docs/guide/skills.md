---
title: Skills
description: Skill discovery, loading strategy, and runtime behavior.
---

# Skills

Skills are local instruction packs centered on `SKILL.md`.

## What a Skill Is

A skill is prompt context, not a tool.

- Skills tell the model how to solve a type of task.
- Tools execute actions (`read_file`, `url_fetch`, `bash`, ...).

## Discovery

Default discovery root:

- `file_state_dir/skills` (usually `~/.morph/skills`)

Runtime scans for `SKILL.md` recursively.

## Loading Controls

```yaml
skills:
  enabled: true
  dir_name: "skills"
  load: []
```

- `enabled: false` => no skill loading
- `load: []` => load all discovered skills
- `load: ["a", "b"]` => load selected skills
- unknown entries are ignored

Task text can also trigger by `$skill-name` or `$skill-id`.

## Runtime Injection Model

Loaded skill metadata is injected into system prompt:

- `name`
- `file_path`
- `description`
- `auth_profiles` (optional)

Actual `SKILL.md` content is loaded only when the model calls `read_file`.

## CLI Commands

```bash
mistermorph skills list
mistermorph skills install
mistermorph skills install "https://example.com/SKILL.md"
```

## Safety Notes

- Remote skill install reviews content before writing.
- Files are downloaded and written only; not executed during install.
