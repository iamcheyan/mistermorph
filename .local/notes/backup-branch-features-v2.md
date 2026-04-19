# backup-main-20260419 功能清单（按影响范围分类）

> 将 36 个提交按**是否影响 CLI/chat 命令**分为两大类。
> 便于决策：CLI 相关的可以直接合到 main，非 CLI 相关的需要评估是否影响其他运行时。
> 最后更新: 2026-04-19

---

## 一、CLI/chat 命令相关（27 个提交）

这些提交只修改 `cmd/mistermorph/clicmd/` 或 `internal/clifmt/`，**不影响 agent 核心、provider、或其他运行时**。

### 1.1 输入交互改进（5 个）

| 提交 | 说明 | 优先级 | 状态 |
|------|------|--------|------|
| `f638f7f` | 用 readline 替换 bufio（历史、行编辑） | 🔴 高 | 已在 PR #35 |
| `d435dd4` | Ctrl+C 中断处理 + 禁用 Glamour | 🟡 中 | |
| `bd84ad0` | 用 os/signal.Notify 处理 Ctrl+C | 🟡 中 | |
| `032d607` | TAB 自动补全（/exit, /init 等） | 🟡 中 | |
| `6e01e01` | rl.Stdout() 统一 writer 修复换行 bug | 🔴 高 | |

### 1.2 输出美化（6 个）

| 提交 | 说明 | 优先级 | 状态 |
|------|------|--------|------|
| `5431e7d` | 带边框的代码块 + 工具通知 | 🟡 中 | |
| `f2143aa` | 纯代码块语法高亮 + 中文过滤 | 🟡 中 | |
| `3821cf1` | 终端语法高亮（chroma） | 🟡 中 | |
| `1b51740` | plan progress 彩色格式化（替代 raw JSON） | 🔴 高 | |
| `05da6b6` | plan progress 图标 + 进度条 | 🟡 中 | |
| `f63788e` | 简化 plan progress 显示 | 🟡 中 | |
| `fc5aed8` | 简化 plan progress 格式 | 🟡 中 | |
| `dbb917b` | plan progress 整合到 spinner + 截断长消息 | 🟡 中 | |
| `4e4acd3` | 清除多行 spinner 输出 | 🟡 中 | |

### 1.3 Prompt/UI 改进（5 个）

| 提交 | 说明 | 优先级 | 状态 |
|------|------|--------|------|
| `01c7787` | 用系统用户名替代硬编码 'you' | 🔴 高 | |
| `6550c8f` | assistant prompt 显示 persona 名 | 🟡 中 | |
| `30aadfc` | 恢复 you> 提示，用 persona 名输出 | 🟡 中 | |
| `316de16` | 注入工作目录到 system prompt | 🟡 中 | |
| `b4ed69f` | compact mode 最小显示模式 | 🟡 中 | |
| `696e0b4` | compact mode 绿色 bullet | 🟡 中 | |

### 1.4 /init 命令相关（5 个）

| 提交 | 说明 | 优先级 | 状态 |
|------|------|--------|------|
| `0ee5a2e` | /init 命令生成 AGENTS.md | 🟡 中 | 已在 PR #35 |
| `2c7f1af` | 在 system prompt 中添加 /init 说明 | 🟡 中 | 已在 PR #35 |
| `b80f172` | /init 的 thinking 动画 | 🟢 低 | |
| `6b1e3e0` | /init 后注入 AGENTS.md 到对话历史 | 🟡 中 | |
| `421f5d3` | 禁止 AI 用 write_file，直接返回内容 | 🟡 中 | |

### 1.5 其他 CLI 修复/改进（4 个）

| 提交 | 说明 | 优先级 | 状态 |
|------|------|--------|------|
| `15c901a` | chat mod 修复（早期） | ⚪ 低 | 可能已过时 |
| `6bc717a` | chat → cli 重命名（已改回） | ❌ 废弃 | |
| `53ae334` | CLI 项目级记忆（.morph/memory.md） | ❌ 废弃 | 已被 memoryruntime 替代 |

---

## 二、非 CLI/影响全局（9 个提交）

这些提交影响 agent 核心、provider、配置系统、或其他运行时（Telegram/Slack）。

### 2.1 ACP/Provider 相关（4 个）

| 提交 | 说明 | 影响范围 | 优先级 |
|------|------|----------|--------|
| `ee3e08d` | runtime model switching + --profile + OpenCode 集成 + auto-discover models | `cmd/mistermorph/`, `internal/llmselect/`, `docs/` | 🔴 高 |
| `794b86a` | gemini_oauth ACP 实时进度反馈 | `internal/acpclient/`, `internal/llmutil/`, `assets/config/` | 🟡 中 |
| `baa3266` | gemini_oauth 进度事件路由到 Telegram/Slack | `internal/channelruntime/`, `providers/gemini/` | 🟡 中 |
| `9f1779e` | ACP tool 事件结构化日志 | `internal/llmutil/` | 🟢 低 |

### 2.2 文档/脚本（3 个）

| 提交 | 说明 | 影响范围 | 优先级 |
|------|------|----------|--------|
| `fc98225` | 合并 help/ 到 docs/ 并更新导航 | `docs/`, `help/` | 🟢 低 |
| `c898e94` | opencode2api 一键部署脚本 | `scripts/`, `docs/` | 🟢 低 |
| `3668905` | 删除误提交的 npm tarball | `nsalerni-gemini-acp-0.1.18.tgz` | ✅ 已完成 |

---

## 三、快速决策表

| 你想做什么 | 推荐提交 | 风险 |
|------------|----------|------|
| **只改进 chat 命令体验** | `01c7787`, `6e01e01`, `1b51740` + 1.2/1.3 组 | 低，只改 cli.go |
| **改进所有运行时（含 Telegram/Slack）** | `ee3e08d` | 中，影响 model 选择逻辑 |
| **发 PR 给上游** | 从 `upstream/master` 新建分支，cherry-pick 对应提交 | |
| **清理废弃代码** | 跳过 `53ae334`, `6bc717a`, `15c901a` | |

---

## 四、与 PR #35 的关系

PR #35 已包含以下提交（无需重复处理）：
- `f638f7f` — readline
- `0ee5a2e` — /init 命令
- `2c7f1af` — /init system prompt
- 以及 PR #35 中其他相关改动

---

*生成方式: `git show --stat` 分析每个提交修改的文件*
