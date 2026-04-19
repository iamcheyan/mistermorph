# backup-main-20260419 功能清单（按影响范围分类）

> 将 36 个提交按**是否影响 CLI/chat 命令**分为两大类。
> 便于决策：CLI 相关的可以直接合到 main，非 CLI 相关的需要评估是否影响其他运行时。
> 最后更新: 2026-04-19

---

## PR 分组总览

| PR 组 | 功能 | 提交数 | 优先级 | 状态 |
|-------|------|--------|--------|------|
| PR #1 | Plan Progress 彩色格式化 | 6 个 | 🔴 高 | ⬜ 待处理 |
| PR #2 | Prompt/UI 改进 | 6 个 | 🔴 高 | ⬜ 待处理 |
| PR #3 | 语法高亮 | 3 个 | 🟡 中 | ⬜ 待处理 |
| PR #4 | 输入交互改进 | 3 个 | 🔴 高 | ⬜ 待处理 |
| PR #5 | /init 命令增强 | 3 个 | 🟡 中 | ⬜ 待处理 |
| PR #6 | 运行时模型切换 | 1 个 | 🔴 高 | ⬜ 待处理 |
| - | gemini_oauth 进度 | 3 个 | 🟡 中 | ⬜ 评估中 |
| - | 文档整理 | 1 个 | 🟢 低 | ⬜ 可选 |

---

## PR #1：Plan Progress 彩色格式化（6 个提交）

AI 做计划时的进度显示，从 raw JSON 改成人类可读的彩色文字。

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `1b51740` | 核心：plan progress 彩色格式化 | ⬜ |
| `05da6b6` | 加图标（✓ ✗）和进度条 | ⬜ |
| `f63788e` | 简化显示，去掉冗余 | ⬜ |
| `fc5aed8` | 进一步简化格式 | ⬜ |
| `dbb917b` | 整合到 spinner + 截断长消息 | ⬜ |
| `4e4acd3` | 清除多行 spinner 残留字符 | ⬜ |

**影响文件**：`cmd/mistermorph/clicmd/cli.go`, `agent/engine_loop.go`
**上游接受度**：高 ✅

---

## PR #2：Prompt/UI 改进（6 个提交）

改善 chat 命令的提示符和显示方式。

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `01c7787` | 系统用户名替代 "you>" | ⬜ |
| `6550c8f` | AI 回复显示 persona 名 | ⬜ |
| `30aadfc` | 恢复 you> + persona 名区分 | ⬜ |
| `316de16` | 注入工作目录到 system prompt | ⬜ |
| `b4ed69f` | compact mode 紧凑模式 | ⬜ |
| `696e0b4` | compact mode 绿色圆点 | ⬜ |

**影响文件**：`cmd/mistermorph/clicmd/cli.go`
**上游接受度**：中 ⚠️（涉及个人偏好）

---

## PR #3：语法高亮（3 个提交）

让 AI 输出的代码块带语法高亮和边框。

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `3821cf1` | 引入 chroma 库做终端语法高亮 | ⬜ |
| `5431e7d` | 带边框的代码块 + 工具通知 | ⬜ |
| `f2143aa` | 纯代码块语法高亮 + 中文过滤 | ⬜ |

**影响文件**：`internal/clifmt/syntax.go`, `internal/clifmt/markdown.go`
**上游接受度**：中 ⚠️（增加新依赖 chroma）

---

## PR #4：输入交互改进（3 个提交）

改善命令行输入体验。

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `d435dd4` | Ctrl+C 中断 + 禁用 Glamour | ⬜ |
| `bd84ad0` | 改进 Ctrl+C 信号处理 | ⬜ |
| `032d607` | TAB 自动补全 slash 命令 | ⬜ |

**影响文件**：`cmd/mistermorph/clicmd/cli.go`
**上游接受度**：高 ✅

---

## PR #5：/init 命令增强（3 个提交）

增强 `/init` 命令的体验（依赖 PR #35 合并）。

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `b80f172` | /init thinking 动画 | ⬜ |
| `6b1e3e0` | /init 后注入 AGENTS.md | ⬜ |
| `421f5d3` | 禁止 AI 用 write_file | ⬜ |

**影响文件**：`cmd/mistermorph/clicmd/cli.go`
**上游接受度**：中 ⚠️（PR #35 未合并）

---

## PR #6：运行时模型切换（1 个提交）

在 chat 中切换模型、加载配置、集成 OpenCode。

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `ee3e08d` | 运行时模型切换 + --profile + OpenCode + auto-discover | ⬜ |

**影响文件**：`cmd/mistermorph/`, `internal/llmselect/`, `docs/`
**上游接受度**：高 ✅（大功能，可能需拆分成多个 PR）

---

## 其他（不上组）

| 提交 | 功能说明 | 处理方式 |
|------|----------|----------|
| `794b86a` | gemini_oauth 实时进度 | 影响全局，评估后决定 |
| `baa3266` | gemini_oauth 推送到 Telegram/Slack | 影响全局，评估后决定 |
| `9f1779e` | ACP 结构化日志 | 影响全局，评估后决定 |
| `fc98225` | 合并 help/ 到 docs/ | 文档整理，可选 |

---

## 已处理（无需再动）

| 提交 | 功能说明 | 状态 |
|------|----------|------|
| `f638f7f` | readline 输入 | ✅ 已在 PR #35 |
| `0ee5a2e` | /init 命令 | ✅ 已在 PR #35 |
| `2c7f1af` | /init system prompt | ✅ 已在 PR #35 |
| `15c901a` | 早期 chat bug 修复 | ❌ 跳过（过时） |
| `6bc717a` | chat → cli 重命名 | ❌ 跳过（废弃） |
| `53ae334` | .morph/memory.md | ❌ 跳过（已被替代） |
| `3668905` | 删除误提交的 npm tarball | ✅ 已完成 |
| `c898e94` | opencode2api 部署脚本 | ✅ 已放 .local/scripts |

---

## 执行顺序建议

1. **PR #1** Plan Progress（影响小、价值高、上游可能接受）
2. **PR #4** Ctrl+C + TAB（基础功能）
3. **PR #2** Prompt/UI（用户体验）
4. **PR #3** 语法高亮（输出美化）
5. **PR #5** /init 增强（等 PR #35 合并）
6. **PR #6** 模型切换（大功能，需更多测试）

---

## 每个组的 cherry-pick 流程

```bash
# 1. 在 main 上合并（自用）
git checkout main
git cherry-pick <commit-1>
git cherry-pick <commit-2>
# ...
go build -o ./bin/mistermorph ./cmd/mistermorph
# 测试...
git push origin main

# 2. 发上游 PR 时，从 upstream/master 新建分支
git fetch upstream
git checkout -b feat/plan-progress upstream/master
git cherry-pick <commit-1>
git cherry-pick <commit-2>
# ...
git push origin feat/plan-progress
# 创建 PR
```

---

## 进度统计

- **已完成/跳过**: 8 个
- **待处理 PR 组**: 6 组（22 个提交）
- **评估中**: 3 个提交
- **可选**: 1 个提交

---

*生成方式: `git show --stat` 分析每个提交修改的文件*
