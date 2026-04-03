---
title: Skills
description: Skill discovery, loading strategy, and runtime behavior.
---

# Skills

Skills are local instruction packs centered on `SKILL.md`.

## Discovery Path

Mistermorph recursively scans `SKILL.md` files under `file_state_dir/skills` (default: `~/.morph/skills`) to discover skills.

## Loading Controls

You can control skills in config like this:

```yaml
skills:
  enabled: true
  dir_name: "skills"
  load: []
```

Where:

- `enabled`: whether skills are loaded
- `load`: the skill ids/names to load. For example, `["apple", "banana"]` means only those two skills are loaded. If empty, all discovered skills are loaded.

Task text can also trigger loading via `$skill-name` or `$skill-id`.

## Skill Injection

The system prompt only injects skill metadata:

- `name`
- `file_path`
- `description`
- `auth_profiles` (optional)

The actual `SKILL.md` content must be loaded by the model via `read_file`.

## Common Commands

```bash
# Show which skills can currently be discovered. Useful for confirming installs
# and checking whether your skill directories are being picked up.
mistermorph skills list
# Install or update built-in skills into the local skills directory. The default
# destination is ~/.morph/skills.
mistermorph skills install
# Install a single skill from a remote SKILL.md. Useful for bringing in an
# external skill.
mistermorph skills install "https://example.com/SKILL.md"
```

Common companion flags:

- `--skills-dir`: add an extra scan root for `skills list`
- `--dry-run`: preview what `skills install` would write without actually writing files
- `--dest`: install the skill into a specific directory for testing or isolation
- `--clean`: delete the existing skills directory before install for a full overwrite update

## Safety Mechanisms

Mistermorph handles skill safety in two stages:

1. The installation phase prevents "download a remote file and execute it immediately".
2. The runtime phase prevents the skill or the model from directly obtaining secrets.

The core principle is that skills can provide process and context, but they cannot escalate privileges on their own.

### Installation: review first, then write

Remote skill installation does not simply download files straight into the local machine. It follows a confirmation-based flow:

```text
+--------------------------------------+
  Remote SKILL.md
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  Show content and confirm
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  Review as untrusted input
  Extract declared extra files 
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  Show write plan and potential risks 
  Confirm again 
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  Write to ~/.morph/skills/<name>/
+--------------------------------------+
```

> If you only want to see what the installer would do, use `--dry-run` first.

### Runtime: a skill declares a profile, but does not get the secret

When a skill needs to access a protected HTTP API, prefer config-based injection via `auth_profile` instead of writing the secret into the skill itself.

For example, a skill can declare that it expects to use a profile:

```yaml
auth_profiles: ["jsonbill"]
```

But that does not mean it is already authorized. The real authorization boundary is defined by config:

```yaml
secrets:
  allow_profiles: ["jsonbill"]

auth_profiles:
  jsonbill:
    credential:
      kind: api_key
      secret: "${JSONBILL_API_KEY}"
    allow:
      url_prefixes: ["https://api.jsonbill.com/tasks"]
      methods: ["GET", "POST"]
      follow_redirects: false
      deny_private_ips: true
```

In this setup, the skill and the LLM only see the profile id `jsonbill`. They never directly see the value of `JSONBILL_API_KEY`.

Mistermorph resolves the real secret from the environment when loading config, then injects it into `url_fetch`. This avoids exposing API keys in prompts, `SKILL.md`, tool parameters, or logs.

At the same time, `auth_profiles` can define request boundaries. For example, `url_prefixes` limits which URL prefixes that profile can access, while `methods`, `follow_redirects`, and `deny_private_ips` further constrain its behavior.
