# Mister Morph（个人维护版）

这是 [quailyquaily/mistermorph](https://github.com/quailyquaily/mistermorph) 的个人 fork，基于上游最新代码，并加入了一些个人需要的功能和调整。

---

## 与上游的差异

| 项目 | 上游原版 | 本 fork |
|------|---------|---------|
| 交互式对话 | ❌ | ✅ `chat` 命令，支持 REPL 交互 |
| 模型切换 | ❌ | ✅ 运行时通过 `/model` 切换模型 |
| Slash 命令 | ❌ | ✅ `/exit`, `/model`, `/memory`, `/remember`, `/help` 等 |
| Bedrock 支持 | ❌ | ✅ AWS Bedrock 独立 provider（AWS CLI） |
| Gemini OAuth | ❌ | ✅ Gemini CLI ACP 集成 |
| OpenCode 集成 | ❌ | ✅ 支持 OpenCode 模型 |
| 自动发现模型 | ❌ | ✅ 自动获取可用模型列表 |
| 用户名显示 | ❌ | ✅ Prompt 中显示系统用户名 |
| Plan 进度彩色输出 | ❌ | ✅ 彩色格式化 |
| Markdown 语法高亮 | ❌ | ✅ 代码块语法高亮 + 行号 |
| Tool 参数显示 | ❌ | ✅ 显示 tool 调用的 path/url/cmd 等参数 |

> 详细变更记录见 [PR #35](https://github.com/quailyquaily/mistermorph/pull/35) 及后续提交。

---

## 当前状态

- **默认分支 `main`**：同步上游最新代码 + 个人修改
- **PR #35**：已合并到上游（squash 为 PR #36），lyricat 做了额外改进
- **上游同步**：所有 upstream 改进已 cherry-pick 到本地 main

---

## 分支说明

| 分支 | 用途 | 状态 |
|------|------|------|
| `main` | 个人维护版，GitHub 默认显示 | ✅ 活跃 |

> 历史分支 `pr-chat`、`backup-main-20260419`、`feat/bedrock-aws-cli` 已清理删除。

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

# 交互式对话（本 fork 新增）
mistermorph chat
```

---

## 交互式对话命令

在 `chat` 模式下可用：

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

> 注：`/forget` 命令已在 upstream 合并时移除，改为 `/reset` 仅清除历史，`/remember` 直接写入长期记忆。

---

## 开发相关

### 本地工作区

`.local/` 目录存放个人开发笔记、日志和脚本：

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
# 然后手动 cherry-pick 需要的改进
```

---

## 注意事项

1. **上游同步**：我们的 `main` 有大量 upstream 没有的功能（bedrock、gemini、clifmt 等），不能直接 `git merge upstream/master`，需要手动 cherry-pick。

2. **故意保留的分歧**：
   - `WithOnToolStart` 保持 3-param（显示 tool 参数），upstream 是 2-param
   - `internal/clifmt` 语法高亮包，upstream 没有
   - `providers/bedrock` 和 `providers/gemini`，upstream 没有

3. **PR 开发**：给上游提交 PR 时，从 `upstream/master` 新建干净分支，不要从个人 `main` 分支发 PR。

4. **私有内容**：`.local/` 和 `README.md` 等个人修改只存在于本 fork，不会向上游提交。

---

## 上游仓库

- 原版地址：https://github.com/quailyquaily/mistermorph
- 原版文档：[docs/README.md](docs/README.md)

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=quailyquaily/mistermorph&type=date&legend=top-left)](https://www.star-history.com/#quailyquaily/mistermorph&type=date&legend=top-left)
