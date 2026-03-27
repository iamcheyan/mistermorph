---
title: Skills
description: Skill 发现、加载策略与运行时行为。
---

# Skills

Skill 是以 `SKILL.md` 为核心的本地指令包。

## Skill 是什么

Skill 是提示上下文，不是工具。

- Skill 负责“怎么做”
- Tool 负责“执行动作”（`read_file`、`url_fetch`、`bash` 等）

## 发现路径

默认根目录：

- `file_state_dir/skills`（通常是 `~/.morph/skills`）

运行时会递归扫描 `SKILL.md`。

## 加载控制

```yaml
skills:
  enabled: true
  dir_name: "skills"
  load: []
```

- `enabled: false`：不加载 skill
- `load: []`：加载全部已发现 skill
- `load: ["a", "b"]`：只加载指定 skill
- 未知条目会被忽略

任务文本中写 `$skill-name` / `$skill-id` 也会触发加载。

## 注入方式

系统 Prompt 只注入 skill 元数据：

- `name`
- `file_path`
- `description`
- `auth_profiles`（可选）

真正的 `SKILL.md` 内容需要模型自行 `read_file`。

## 常用命令

```bash
mistermorph skills list
mistermorph skills install
mistermorph skills install "https://example.com/SKILL.md"
```

## 安全说明

- 远程 skill 安装前会审查与确认。
- 安装过程只下载写盘，不会执行脚本。
