
## 2026-04-19 22:45 - PR #35 合并完成，upstream 改进同步完毕

### 背景

PR #35（chat 交互命令）被 lyricat squash 合并为 PR #36，但 lyricat 在合并时做了额外修改：
- `735bb6f` — review feedback 修复（memory worker lifecycle、/reset 语义）
- `0d593d2` — 移除 /forget 命令、稳定 REPL UI state、/remember 改直接写长期记忆

### 已同步的 upstream 改进

| upstream commit | 内容 | 本地 commit |
|----------------|------|------------|
| `735bb6f` | review feedback: memory worker lifecycle, /reset 语义 | `2f128a0` |
| `0d593d2` | 移除 /forget, 稳定 REPL UI, /remember 改进 | `6d8c2f2` |
| `ea7c15a` | console stats view 优化 | `50c8627` |
| `b616f5d` | console vendor icons (Anthropic, OpenAI 等) | `50c8627` |
| `70f024a` | agent_settings_test 环境变量清理 helper | `65354e3` |

### 同步策略说明

我们的 `main` 不是 upstream/master 的干净下游，有 30 个独有 commit（.local/、bedrock、gemini、clifmt 语法高亮、ACP 进度事件等）。因此不能直接用 `git pull upstream/master`，而是手动 cherry-pick upstream 的改进，同时保留我们自己的功能。

### 故意保留的分歧（divergence）

| 项目 | upstream | 我们的 main | 理由 |
|------|---------|------------|------|
| `WithOnToolStart` | 2-param (ctx, toolName) | 3-param (ctx, toolName, params) | CLI 显示 tool 参数（path, url, cmd 等） |
| `internal/clifmt` | 不存在 | 存在（glamour/chroma） | Markdown 语法高亮 |
| `providers/bedrock` | 不存在 | 存在 | AWS Bedrock 独立 provider |
| `providers/gemini` | 不存在 | 存在 | Gemini OAuth ACP provider |

### 已删除的分支

- ✅ `pr-chat` — PR 已合并，不再需要
- ✅ `backup-main-20260419` — 备份分支，内容已移植
- ✅ `backup-pre-upstream-merge-20260419` — 备份分支
- ✅ `feat/bedrock-aws-cli` — 内容已在 main 中，且落后于 main

### 当前分支状态

```
本地分支: 只有 main
远程分支: 只有 origin/main
upstream: quailyquaily/mistermorph master
```

### 这周提交 Bedrock PR 的准备

Bedrock 代码已经在 main 中：
- `providers/bedrock/client.go` — 核心实现（519 行）
- `internal/llmutil/llmutil.go` — provider 路由

提交步骤：
```bash
git checkout -b feat/bedrock-improvements main
# 修改 providers/bedrock/client.go
git push origin feat/bedrock-improvements
# 在 GitHub 创建 PR
```

### 验证

- `go build ./cmd/mistermorph` ✅
- `go test ./...` ✅ 全部通过
- `main` 已推送 origin ✅
