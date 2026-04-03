---
title: Skills
description: Skill 发现、加载策略与运行时行为。
---

# Skills

Skill 是以 `SKILL.md` 为核心的本地指令包。

## 路径

Mistermorph 会尝试在 `file_state_dir/skills`（默认是 `~/.morph/skills`）下，递归扫描 `SKILL.md` 来发现 Skills。

## 加载控制

在配置文件中，可以通过下面的选项来控制 Skills：

```yaml
skills:
  enabled: true
  dir_name: "skills"
  load: []
```

其中，

- `enabled`：控制是否加载 skills
- `load`：用于指定加载 skill 的 id/name，例如 `["apple", "banana"]`表示只指定 skills `apple` 和 `banana`。如果留空则表示全部加载。

另外，如果在 agent 的任务文本中写 `$skill-name` / `$skill-id` 也会触发加载。

## Skills 的注入

系统 Prompt 只注入 skill 元数据：

- `name`
- `file_path`
- `description`
- `auth_profiles`（可选）

真正的 `SKILL.md` 内容需要 LLM 自行 `read_file`。

## 常用命令

```bash
# 查看当前能发现到哪些 skill，适合用来确认安装结果、排查目录是否生效。
mistermorph skills list
# 把内置 skill 安装或更新到本地 skills 目录，默认写入 `~/.morph/skills`。
mistermorph skills install
# 从远程 `SKILL.md` 安装单个 skill，适合引入外部 skill。
mistermorph skills install "https://example.com/SKILL.md"
```

常见配套参数：

- `--skills-dir`：给 `skills list` 额外添加扫描根目录。
- `--dry-run`：预览 `skills install` 会写入什么内容，但不真正写入。
- `--dest`：把 skill 安装到指定目录，方便测试或隔离环境。
- `--clean`：安装前删除已有 skills 目录，适合做彻底覆盖更新。

## 安全机制

Mister Morph 对 skill 的安全处理分成两个阶段：

1. 安装阶段防止拿到远程文件就执行
2. 运行阶段防止 Skill 或 LLM 直接拿到密钥

基本原则是：skills 可以提供流程和上下文，但不能越权。

### 安装：先审查，再写入

远程 skill 安装不会直接下载到本地，而是一个带确认的流程：

```text
+--------------------------------------+
  远程 SKILL.md
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  展示内容并确认
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  按不可信输入做审查
  提取声明的附加文件 
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  展示写入计划和潜在风险 
  再次确认 
+--------------------------------------+
                  |
                  v
+--------------------------------------+
  写入 ~/.morph/skills/<name>/
+--------------------------------------+
```

> 如果你只是想看安装器会做什么，可以先用 `--dry-run` 预览。

### 运行阶段：skill 只声明 profile，不直接拿密钥

当 skill 需要访问受保护的 HTTP API 时，推荐通过 `auth_profile` 走配置注入，而不是把密钥写进 skill。

比如 skill 可以声明自己会用到某个认证档案：

```yaml
auth_profiles: ["jsonbill"]
```

但这并不等于它已经被授权。真正的授权边界由配置决定：

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

在这个配置下，skill 和 LLM 只会看到 `jsonbill` 这个 profile id，不会直接看到 `JSONBILL_API_KEY` 的值。

真实密钥由 Mister Morph 在加载配置时，从环境变量解析，再注入到 `url_fetch` 工具，这样可以避免把 API Key 暴露到 prompt、`SKILL.md`、工具参数或日志里。

同时，在 auth_profiles 中我们还可以设定访问边界，例如 `url_prefixes` 限制该  profile 可以发请求的 URL 前缀，用 `methods` 限定方法、用 `follow_redirects`、`deny_private_ips` 来限定更多行为边界。
