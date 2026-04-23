# Upstream Merge Report — 2026-04-19

## 概述

将 `upstream/master` (commit `34e887e`) 合并到本地 `main` 分支。
合并前创建了备份分支 `backup-main-20260419`。

**原则：所有冲突一律使用上游版本。**

---

## 上游带来的新功能（已合并）

| 功能 | Commit | 说明 |
|------|--------|------|
| Workspace 目录支持 | `ec4b72a` | `--workspace`, `--no-workspace` flag, `/workspace` attach/detach/status 命令 |
| Console plan + activity panels | `c583f9f` | 持久化计划面板和活动进度面板 |
| Cache stats 列 | `324778a` | StatsView 新增缓存率、缓存费用变动展示 |
| 提高默认 agent loop limits | `d4337e4` | 默认循环限制提高 |
| Chat title heights 修复 | `34e887e` | UI 对齐修复 |
| 依赖更新 | `bb603f9` | uniai 升级到 v0.1.19 等 |

---

## 冲突文件及处理方式

以下文件在合并时产生冲突，**全部使用上游版本解决**：

### 1. `cmd/mistermorph/chatcmd/agents.go`
- **冲突**：`pathroots` import, `projectDir` vs `chatFileCacheDir`
- **处理**：保留上游的 `pathroots` + `projectDir` 方式
- **去掉本地**：`chatFileCacheDir` 简化逻辑

### 2. `cmd/mistermorph/chatcmd/chat.go`
- **冲突**：`resolveChatFileCacheDir()` 函数 vs upstream workspace flags
- **处理**：保留上游的 `--workspace`/`--no-workspace` flags
- **去掉本地**：`resolveChatFileCacheDir()` 函数

### 3. `cmd/mistermorph/chatcmd/commands.go`
- **冲突**：`/workspace` 命令, `chatBuiltinCommandsBlock()`, `/init` `/update` 路径
- **处理**：保留上游的 `/workspace` 命令和 `chatBuiltinCommandsBlock()`
- **去掉本地**：使用 `chatFileCacheDir` 的路径（改为 `projectDir()`）

### 4. `cmd/mistermorph/chatcmd/format.go`
- **冲突**：注释差异（纯格式）
- **处理**：保留上游版本（无注释）

### 5. `cmd/mistermorph/chatcmd/memory.go`
- **冲突**：末尾空行
- **处理**：保留上游版本

### 6. `cmd/mistermorph/chatcmd/repl.go`
- **冲突**：`clifmt` import + `clifmt.RenderMarkdown()` vs `pathroots`
- **处理**：保留上游的 `pathroots`，**去掉** `clifmt.RenderMarkdown()`
- **去掉本地**：Markdown 语法高亮渲染

### 7. `cmd/mistermorph/chatcmd/session.go`
- **冲突**：workspace 字段, `projectDir()`, `rebuildPromptSpec()`, `WithOnToolStart` 回调
- **处理**：保留上游的 workspace 机制和 2-param `WithOnToolStart`
- **去掉本地**：3-param `WithOnToolStart`（带工具参数显示）

### 8. `cmd/mistermorph/chatcmd/ui.go`
- **冲突**：`printChatSessionHeader` 签名（2 param vs 4 param）
- **处理**：保留上游的 4-param 版本（含 `workspaceDir`）
- **去掉本地**：2-param 简化版本

### 9. `cmd/mistermorph/runcmd/run.go`
- **冲突**：`--workspace`/`--no-workspace` flags + provider 列表
- **处理**：保留上游新增 flags，但保留本地 `gemini_oauth` 在 provider 列表中
- **注意**：后续严格化处理中，已将 `gemini_oauth` 从 provider 列表移除（完全上游化）

### 10. `go.sum`
- **冲突**：`charmbracelet`/`clipperhouse` 依赖 vs `uniai` 版本
- **处理**：保留上游的 `uniai v0.1.19`，但保留 `charmbracelet` 依赖（供 `internal/clifmt` 使用）
- **注意**：`internal/clifmt` 是本地无冲突新增包，go.sum 中保留其依赖

### 11. `web/console/src/i18n/index.js`
- **冲突**：新增 `stats_cache_rate`, `stats_cache_cost_delta` 翻译
- **处理**：保留上游新增翻译

### 12. `web/console/src/views/StatsView.js`
- **冲突**：多处新增缓存统计功能
- **处理**：完全保留上游版本

---

## 后续严格化修正（合并后执行）

合并后发现仍保留了一些本地改动，进行了二次修正：

### 已彻底去掉的本地功能

| 本地功能 | 所在文件 | 处理方式 |
|---------|---------|---------|
| `clifmt.RenderMarkdown()` Markdown 渲染 | `cmd/mistermorph/chatcmd/repl.go` | 删除调用和 import |
| 3-param `WithOnToolStart`（工具参数显示） | `agent/engine.go`, `agent/engine_loop.go`, `agent/engine_concurrent_test.go`, `cmd/mistermorph/chatcmd/session.go` | 全部改回 2-param 上游版本 |
| `chatFileCacheDir` 简化逻辑 | `cmd/mistermorph/chatcmd/ui.go`, `cmd/mistermorph/chatcmd/commands.go` | 改回上游的 `projectDir()` + `workspaceDir` |
| `gemini_oauth` provider flag | `cmd/mistermorph/runcmd/run.go` | 从 provider 列表移除 |
| `compact_mode` 配置 | `assets/config/config.example.yaml`, `internal/configdefaults/defaults.go` | 恢复上游版本 |
| `BuiltinAskUser` 工具 | `internal/toolsutil/static_register.go` | 恢复上游版本 |

---

## 仍保留的本地文件（无冲突新增）

以下文件是本地历史提交中**新增**的，合并时**无冲突**，目前仍保留在 main 中：

| 文件/目录 | 说明 | 是否应保留 |
|---------|------|-----------|
| `providers/bedrock/` | AWS Bedrock provider | ✅ PR 分支已单独提交 |
| `providers/gemini/` | Gemini OAuth provider | ❌ 上游无此功能 |
| `internal/clifmt/` | Markdown 语法高亮 | ❌ 已去掉调用，但文件仍存在 |
| `internal/llmutil/acp_llm_client.go` | ACP LLM 客户端适配器 | ❌ 上游无此功能 |
| `internal/acpclient/client.go` (修改) | oauth-personal 认证方法 | ❌ 上游无此功能 |
| `internal/channelruntime/slack/runtime_task.go` (修改) | ACP 事件路由到 Slack | ❌ 上游无此功能 |
| `internal/channelruntime/telegram/runtime_task.go` (修改) | ACP 事件路由到 Telegram | ❌ 上游无此功能 |
| `internal/llmutil/llmutil.go` (修改) | gemini_oauth 分支 | ❌ 上游无此功能 |
| `.local/` | 本地开发笔记、脚本 | ✅ 个人工作区 |
| `README.md` (中文) | 中文 README | ✅ 个人 fork 说明 |
| `docs/ja/git-workflow.md` | 日文 git 工作流文档 | ✅ 个人文档 |

---

## 验证结果

- ✅ `go build ./cmd/mistermorph` — 编译成功
- ✅ `go test ./...` — 全部通过
- ✅ 无合并冲突标记残留

---

## 分支状态

| 分支 | Commit | 说明 |
|------|--------|------|
| `main` | `21993a1` | 合并后，以上游为准 |
| `backup-main-20260419` | `3f03016` | 合并前备份 |
| `upstream-master` | `34e887e` | 上游最新 |
| `pr/bedrock-aws-cli` | `cc72664` | 干净 Bedrock PR 分支 |

---

## 备注

本次合并严格遵循"以上游为准"原则。所有冲突文件均使用上游版本，本地增强功能（clifmt、3-param WithOnToolStart、chatFileCacheDir 简化）已全部移除。

仍有部分本地新增文件（gemini provider、acp_llm_client、channelruntime 修改等）因无冲突而保留，如需完全同步上游，需进一步手动清理。
