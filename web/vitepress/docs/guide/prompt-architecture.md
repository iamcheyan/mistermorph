---
title: Prompt Architecture (Top-Down)
description: How system prompt content is assembled from identity to runtime blocks.
---

# Prompt Architecture (Top-Down)

This is the runtime order used in current code.

## Top-Down Flow Diagram

```text
PromptSpecWithSkills(...)
        |
        v
ApplyPersonaIdentity(...)
        |
        v
AppendLocalToolNotesBlock(...)
        |
        v
AppendPlanCreateGuidanceBlock(...)
        |
        v
AppendTodoWorkflowBlock(...)
        |
        v
Channel PromptAugment(...)
        |
        v
AppendMemorySummariesBlock(...)
        |
        v
BuildSystemPrompt(...) -> system prompt text
```

## 1) Identity Layer

Base identity comes from `agent.DefaultPromptSpec()`.

Then runtime may replace it via local persona files:

- `file_state_dir/IDENTITY.md`
- `file_state_dir/SOUL.md`

Applied by `promptprofile.ApplyPersonaIdentity(...)`.

## 2) Skills Layer

`skillsutil.PromptSpecWithSkills(...)` discovers selected skills and injects only metadata into the system prompt:

- skill name
- `SKILL.md` path
- short description
- required `auth_profiles` (if provided)

The model reads the actual skill file through `read_file` when needed.

## 3) Core Policy Blocks

Common blocks are appended in task runtime order:

1. local tool notes (`SCRIPTS.md`) via `AppendLocalToolNotesBlock`
2. `plan_create` guidance via `AppendPlanCreateGuidanceBlock`
3. TODO workflow policy via `AppendTodoWorkflowBlock`

## 4) Channel Blocks

Channel runtime then appends channel-specific blocks.

Examples:

- Telegram: `AppendTelegramRuntimeBlocks(...)`
- Slack: `AppendSlackRuntimeBlocks(...)`
- LINE: `AppendLineRuntimeBlocks(...)`
- Lark: `AppendLarkRuntimeBlocks(...)`

## 5) Memory Injection Block

If enabled, memory snapshot text is appended as another block through `AppendMemorySummariesBlock(...)`.

## 6) System Prompt Rendering

`agent.BuildSystemPrompt(...)` renders `agent/prompts/system.md` with:

- identity
- skills metadata
- appended blocks
- tool summary from registry
- rules list

## 7) Message Stack Sent to LLM

In `engine.Run(...)`, message order is:

```text
[system] rendered system prompt
   ->
[user] mister_morph_meta (optional)
   ->
[history] non-system messages
   ->
[user] current message or raw task
```

1. system prompt
2. injected runtime metadata message (`mister_morph_meta`, if present)
3. history messages (non-system)
4. current message (if provided) or raw task text

## Practical Rule

If you need to customize prompt behavior, prefer these extension points in order:

1. `IDENTITY.md` / `SOUL.md`
2. `SKILL.md` + `skills.load`
3. `SCRIPTS.md`
4. runtime-specific prompt augment
5. low-level `agent.WithPromptBuilder(...)` (only for full custom wiring)
