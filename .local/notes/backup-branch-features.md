# backup-main-20260419 / backup-pre-upstream-merge-20260419 分支功能清单

> 记录备份分支中尚未合并到 PR #35 的功能，供后续挑选使用。
> 备份分支基于: a79b3be (initial commit) - 这是独立的历史线
> 共 36 个提交，vs upstream/master

---

## 已合并到 PR #35 的功能（不需要再处理）

以下功能已经在 `pr-chat` 分支中，已发 PR #35：

- [x] 交互式 chat 命令基础框架
- [x] `/exit`, `/quit`, `/forget`, `/memory`, `/remember`, `/init`, `/update`, `/model`, `/help` 命令
- [x] `internal/chatcommands` 命令分发器
- [x] `compact_mode` 紧凑模式
- [x] 记忆系统集成（`memoryruntime`）
- [x] AGENTS.md 生成（`/init`, `/update`）
- [x] `openai_custom` → `openai` 映射修复
- [x] `session.go` + `repl.go` 拆分

---

## 备份分支中独有的功能（待挑选）

按功能模块分类：

### 1. CLI 输出美化

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `5431e7d` | bordered code blocks + tool notifications | chatcmd | 中 |
| `f2143aa` | syntax highlighting for plain code blocks + Chinese text filtering | chatcmd | 中 |
| `3821cf1` | terminal syntax highlighting with chroma | chatcmd | 中 |

### 2. Plan Progress 显示

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `1b51740` | format plan progress as human-readable colored text | agent/chatcmd | 高 |
| `b80f172` | thinking animation to /init command | chatcmd | 低 |
| `05da6b6` | plan progress display with icons and progress bar | chatcmd | 中 |
| `f63788e` | simplify plan progress display | chatcmd | 中 |
| `fc5aed8` | simplify plan progress display format | chatcmd | 中 |
| `dbb917b` | integrate plan progress into spinner and truncate long messages | chatcmd | 中 |

### 3. ACP/Provider 相关

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `ee3e08d` | runtime model switching, --profile flag, OpenCode integration, auto-discover models | llmutil/chatcmd | 高 |
| `c898e94` | one-click deploy script for opencode2api | scripts | 低 |
| `794b86a` | real-time progress feedback for gemini_oauth ACP provider | integration | 中 |
| `baa3266` | route gemini_oauth ACP progress events to Telegram and Slack | integration | 中 |
| `9f1779e` | structured logging for ACP tool events | acpclient | 低 |

### 4. Prompt/UI 改进

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `01c7787` | use system username for user prompt | chatcmd | 高 |
| `6550c8f` | display persona name in assistant prompt | chatcmd | 中 |
| `30aadfc` | restore you> prompt for user input, use persona name for assistant output | chatcmd | 中 |
| `6b1e3e0` | inject AGENTS.md content into conversation history after /init | chatcmd | 中 |

### 5. 输入/交互改进

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `f638f7f` | replace bufio with readline for CLI input | chatcmd | 高 |
| `d435dd4` | Ctrl+C interrupt handling + disable Glamour | chatcmd | 中 |
| `bd84ad0` | use os/signal.Notify for Ctrl+C interrupt | chatcmd | 中 |
| `032d607` | TAB auto-completion for built-in CLI commands | chatcmd | 中 |

### 6. 文档整理

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `fc98225` | merge help/ docs into docs/ and update navigation | docs | 低 |
| `2c7f1af` | add /init command to CLI system prompt | chatcmd | 低 |

### 7. 其他修复

| 提交 | 功能 | 文件 | 优先级 |
|------|------|------|--------|
| `6e01e01` | use rl.Stdout() as unified writer to fix newline rendering | chatcmd | 高 |
| `4e4acd3` | clear multi-line spinner output correctly | chatcmd | 中 |
| `316de16` | inject working directory context into system prompt | chatcmd | 中 |
| `421f5d3` | tell AI not to use write_file in /init, return AGENTS.md content directly | chatcmd | 中 |
| `3668905` | remove accidentally committed npm tarball | chore | 低 |

### 8. 早期实验（可能已废弃）

| 提交 | 功能 | 状态 |
|------|------|------|
| `53ae334` | add CLI project-level memory (.morph/memory.md) | ❌ 已被 memoryruntime 替代 |
| `6bc717a` | rename chat command to cli | ❌ 已改回 chat |
| `15c901a` | fix: chat mod | ❌ 早期修复，可能已过时 |

---

## 挑选建议

### 高优先级（建议优先合并）
1. `f638f7f` - readline 替换 bufio（已在 PR #35 中）
2. `ee3e08d` - runtime model switching + OpenCode（核心功能）
3. `01c7787` - system username for prompt（用户体验）
4. `1b51740` - plan progress 彩色格式化（可读性）
5. `6e01e01` - rl.Stdout() 统一 writer（bug 修复）

### 中优先级
6. `05da6b6` ~ `dbb917b` - plan progress 显示改进
7. `6550c8f` ~ `30aadfc` - persona name 显示
8. `032d607` - TAB 自动补全
9. `794b86a` ~ `baa3266` - gemini_oauth 进度反馈

### 低优先级 / 可选
10. `5431e7d` ~ `3821cf1` - 语法高亮（可能和现有渲染冲突）
11. `c898e94` - opencode2api 部署脚本
12. `fc98225` - 文档整理

---

## 如何挑选提交

```bash
# 查看某个提交的详细改动
git show <commit-hash>

# 将某个提交 cherry-pick 到当前分支
git cherry-pick <commit-hash>

# 查看提交修改了哪些文件
git show --stat <commit-hash>
```

---

*最后更新: 2026-04-19*
