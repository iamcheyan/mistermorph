---
title: Prompt 组织（自顶向下）
description: 从身份层到运行时层，说明系统 Prompt 的拼装顺序。
---

# Prompt 组织（自顶向下）

下面是当前代码里的真实顺序。

## 顶层流程图

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
BuildSystemPrompt(...) -> system prompt 文本
```

## 1) Identity 层

基础身份来自 `agent.DefaultPromptSpec()`。

随后 runtime 可能用本地 persona 覆盖：

- `file_state_dir/IDENTITY.md`
- `file_state_dir/SOUL.md`

通过 `promptprofile.ApplyPersonaIdentity(...)` 应用。

## 2) Skills 层

`skillsutil.PromptSpecWithSkills(...)` 只注入 skill 元数据：

- 名称
- `SKILL.md` 路径
- 描述
- `auth_profiles`（如果有）

需要具体技能内容时，模型再调用 `read_file` 读取 `SKILL.md`。

## 3) Core 策略块

在 task runtime 里依次追加：

1. 本地工具说明（`SCRIPTS.md`）
2. `plan_create` 指引
3. TODO 工作流策略

## 4) 通道策略块

再由各通道追加自己的策略块：

- Telegram
- Slack
- LINE
- Lark

## 5) Memory 注入块

若启用 memory，会再追加 memory summary 块。

## 6) 渲染系统 Prompt

`agent.BuildSystemPrompt(...)` 用 `agent/prompts/system.md` 渲染：

- identity
- skills 元数据
- 追加 block
- 工具摘要
- 附加规则

## 7) 发送给 LLM 的消息顺序

`engine.Run(...)` 中的顺序：

```text
[system] 渲染后的 system prompt
   ->
[user] mister_morph_meta（可选）
   ->
[history] 非 system 历史消息
   ->
[user] current message 或原始 task
```

1. system prompt
2. runtime meta（`mister_morph_meta`）
3. history（非 system）
4. current message 或原始 task

## 实践建议

自定义 prompt 时，优先按这个层级做：

1. `IDENTITY.md` / `SOUL.md`
2. `SKILL.md` + `skills.load`
3. `SCRIPTS.md`
4. runtime prompt augment
5. `agent.WithPromptBuilder(...)`（仅在必须完全自定义时）
