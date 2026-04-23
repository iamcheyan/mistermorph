# Mister Morph（个人维护版）

这是 [quailyquaily/mistermorph](https://github.com/quailyquaily/mistermorph) 的个人 fork，基于上游最新代码，保留了个别上游尚未合并的功能。

> **2025-04-19 更新**：已完成 upstream/master 全量合并（`34e887e`）。所有冲突一律以上游版本为准，上游已实现的功能（交互式 chat、workspace、plan/activity panels、agent loop limits 提升等）均已同步。本 fork 目前仅保留 Bedrock provider 等少量上游未有的功能。

---

## 与上游的差异

| 项目 | 上游原版 | 本 fork |
|------|---------|---------|
| **Bedrock 支持** | ❌ | ✅ AWS Bedrock 独立 provider（AWS CLI 认证） |
| ~~交互式对话~~ | ✅ | ✅ 已合并到上游 |
| ~~模型切换~~ | ✅ | ✅ 已合并到上游 |
| ~~Slash 命令~~ | ✅ | ✅ 已合并到上游 |
| ~~Workspace 机制~~ | ✅ | ✅ 已合并到上游 |
| ~~Plan/Activity panels~~ | ✅ | ✅ 已合并到上游 |
| ~~Agent loop limits~~ | ✅ | ✅ 已合并到上游 |

> 历史功能如 `clifmt`（语法高亮）、`gemini_oauth`（Gemini CLI ACP）、`compact_mode`、`BuiltinAskUser` 等已在本次合并中移除，以上游实现为准。

---

## 分支说明

| 分支 | 用途 | 状态 |
|------|------|------|
| `main` | 个人维护版，GitHub 默认显示 | ✅ 活跃 |
| `backup-main-20260419` | 合并前完整备份 | 🗄️ 保留 |
| `pr/bedrock-aws-cli` | 给上游提交 Bedrock PR 的干净分支 | ✅ 已推送到 origin |

> 给上游提交 PR 时，请从 `upstream/master` 新建干净分支（如 `pr/bedrock-aws-cli`），不要从个人 `main` 分支发 PR。

---

## 快速开始

### 安装

```bash
# 从源码安装
go install github.com/iamcheyan/mistermorph/cmd/mistermorph@latest

# 或克隆后编译
git clone https://github.com/iamcheyan/mistermorph.git
cd mistermorph
go build -o ./bin/mistermorph ./cmd/mistermorph
```

### 运行

```bash
# 初始化配置
mistermorph install

# 设置 API Key
export MISTER_MORPH_LLM_API_KEY="your-api-key"

# 单次任务
mistermorph run --task "Hello!"

# 交互式对话（upstream 已支持）
mistermorph chat
```

---

## 交互式对话命令

在 `chat` 模式下可用（upstream 已实现）：

| 命令 | 说明 |
|------|------|
| `/exit` 或 `/quit` | 退出对话 |
| `/reset` | 重置对话历史（不清除记忆） |
| `/model <模型名>` | 切换当前使用的模型 |
| `/memory` | 查看记忆状态 |
| `/remember <内容>` | 添加长期记忆 |
| `/init` | 读取 AGENTS.md 作为项目上下文 |
| `/update` | 通过 AI 重新生成 AGENTS.md |
| `/help` | 显示帮助 |

---

## 开发相关

### 本地工作区

`.local/` 目录存放个人开发笔记、日志和脚本（不会向上游提交）：

```
.local/
├── notes/          # 开发笔记、功能清单
├── logs/           # 运行日志
├── scripts/        # 辅助脚本
└── backups/        # 本地备份
```

### 常用命令

```bash
# 编译
go build -o ./bin/mistermorph ./cmd/mistermorph

# 测试
go test ./...

# 同步上游更新
git fetch upstream
git log upstream/master --oneline -10  # 查看上游新提交
```

### 合并上游流程

```bash
# 1. 创建备份
git branch backup-main-$(date +%Y%m%d)

# 2. 合并 upstream/master
git merge upstream/master

# 3. 冲突处理原则：一律以上游版本为准
#    - 保留本地新增文件（如 providers/bedrock/）
#    - 冲突文件使用 upstream 版本

# 4. 验证
go build ./... && go test ./...

# 5. 推送
git push origin main
```

---

## 注意事项

1. **上游同步**：`main` 分支目前仅比 upstream 多出 Bedrock provider 和少量本地文件，合并冲突时以上游为准即可。

2. **保留的本地文件**（无冲突，未清理）：
   - `providers/bedrock/` — Bedrock provider（功能完整，已整理为 `pr/bedrock-aws-cli` 分支）
   - `providers/gemini/` — Gemini OAuth provider 代码（flag 已移除，未启用）
   - `internal/clifmt/` — clifmt 包（已无人调用，死代码）
   - `internal/llmutil/acp_llm_client.go` — ACP LLM 适配器
   - `internal/acpclient/client.go` — oauth-personal 认证方法
   - `internal/channelruntime/slack/runtime_task.go` — ACP 事件路由
   - `internal/channelruntime/telegram/runtime_task.go` — ACP 事件路由

3. **PR 开发**：给上游提交 PR 时，从 `upstream/master` 新建干净分支，不要从个人 `main` 分支发 PR。

4. **私有内容**：`.local/` 和 `README.md` 等个人修改只存在于本 fork，不会向上游提交。

---

## 上游仓库

- 原版地址：https://github.com/quailyquaily/mistermorph
- 原版文档：[docs/README.md](docs/README.md)

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=quailyquaily/mistermorph&type=date&legend=top-left)](https://www.star-history.com/#quailyquaily/mistermorph&type=date&legend=top-left)
