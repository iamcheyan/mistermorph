# 开发日志

## 分支管理策略（方案 A：main = 全能版）

### 分支说明

| 分支 | 用途 | 给谁用 |
|------|------|--------|
| `main` | **全能版**，包含所有个人开发功能 | 别人直接 clone 使用 |
| `feat/xxx` | **给上游的 PR 分支**，只含特定功能 | 发 PR 给 upstream |
| `backup-main-20260419` | **原始开发备份**，36 个提交 | 从中挑代码 |
| `pr-chat` | PR #35 chat 命令 | 等待上游合并 |

### 核心规则

> **`main` 是给别人用的全能版，`feat/xxx` 是给上游的干净 PR。**

**给别人用**：
```bash
git clone https://github.com/iamcheyan/mistermorph.git
# 默认就是 main 分支，包含所有功能
```

**发 PR 给上游**：
```bash
# 必须从 upstream/master 新建干净分支！
git fetch upstream
git checkout -b feat/xxx upstream/master
# 开发...
git push origin feat/xxx
# 发 PR（对比 upstream/master）
```

**⚠️ 绝对不要从 main 发 PR 给上游！main 包含太多个人修改。**

### main 分支内容

当前 `main` 包含：
- ✅ 上游最新代码
- ✅ 中文 README（个人展示）
- ✅ `.local/` 私有工作区
- ✅ Bedrock provider（AWS CLI 免 API Key）
- ✅ chat 交互命令（已 cherry-pick 从 pr-chat）
- ✅ Plan Progress 彩色格式化 + spinner 整合
- ✅ 语法高亮（chroma）
- ✅ 运行时模型切换
- ⏳ 其他 backup 功能（逐个挑）

---

## 2026-04-19

### 已完成

1. ✅ 重建 `main` 分支（基于 upstream/master + 个人修改）
2. ✅ 推送 `backup-main-20260419` 到远程备份
3. ✅ 重写中文 README
4. ✅ 创建 `feat/bedrock-aws-cli` PR 分支并推送
5. ✅ Bedrock provider 已合并到 `main`
6. ✅ chat 命令已 cherry-pick 到 `main`

### 待办

- [ ] 从 backup 挑功能做新 PR：
  - [ ] plan progress 彩色格式化 (`1b51740`)
  - [ ] 系统用户名 (`01c7787`)
  - [ ] rl.Stdout() 统一 writer (`6e01e01`)
- [ ] 上游更新时同步 main：`git merge upstream/master`

---

## 2026-04-19 (后续) - Backup 分支功能移植

### 背景

`backup-main-20260419` 分支有 36 个提交不在 `main` 中。这些提交原本是在旧 main（单文件 `cli.go` 结构）上开发的，而当前 `main` 已将 CLI 代码重构为 `cmd/mistermorph/chatcmd/` 多文件结构，导致直接 cherry-pick 会产生大量冲突。

**解决策略**：将 36 个提交按功能分组，用子代理并行分析每组提交的累积 diff，然后手动将逻辑移植到 `main` 分支的对应文件中。

### 提交分组（共 8 组）

| 组 | 提交范围 | 功能描述 | 提交数 |
|---|---------|---------|-------|
| PR #1 | `1b51740` ~ `4e4acd3` | Plan Progress 彩色格式化 + spinner 整合 | 6 |
| PR #2 | `6550c8f` ~ `032d607` | Persona 名称显示 + 用户名 + TAB 补全 | 5 |
| PR #3 | `3821cf1` ~ `f2143aa` | 语法高亮（chroma）+ 边框代码块 | 3 |
| PR #4 | `f638f7f` ~ `bd84ad0` | Readline 替换 bufio + Ctrl+C 中断 | 3 |
| PR #5 | `0ee5a2e` ~ `6b1e3e0` | /init 命令增强（AGENTS.md 注入等） | 4 |
| PR #6 | `ee3e08d` | 运行时模型切换 + --profile + OpenCode | 1 |
| PR #7 | `794b86a` ~ `9f1779e` | ACP gemini_oauth 实时进度 + Telegram/Slack | 3 |
| PR #8 | `b4ed69f` ~ `696e0b4` + 杂项 | Compact Mode + 文档合并 + 部署脚本 | 6 |

### 移植结果

#### 已移植并提交（commit `2c6af15`）

**PR #1: Plan Progress 彩色格式化** ✅
- `cmd/mistermorph/chatcmd/session.go` - Plan Progress 与 spinner 整合，修复 writer 引用
- `cmd/mistermorph/chatcmd/repl.go` - spinner 动画与 session 正确关联
- `cmd/mistermorph/chatcmd/format.go` - 格式化函数注释和逻辑优化
- 关键修复：将局部变量 `writer/stopAnim/setAnimMessage` 改为 `chatSession` 结构体字段，避免闭包捕获 nil 值

**PR #3: 语法高亮** ✅
- 新增 `internal/clifmt/syntax.go` - chroma 语法高亮核心（HighlightCodeBlocks, highlightCode, wrapInBox, 中文过滤）
- 新增 `internal/clifmt/markdown.go` - Glamour markdown 渲染器
- 新增 `internal/clifmt/syntax_test.go` - 单元测试
- `agent/engine.go` / `engine_loop.go` / `engine_concurrent_test.go` - `WithOnToolStart` 回调扩展为 `func(*Context, string, map[string]any)` 以传递工具参数
- `go.mod` - 新增 `github.com/alecthomas/chroma/v2` 依赖

**PR #6: 运行时模型切换** ✅
- `internal/llmselect/command.go` - 扩展 `/model` 命令支持 `next`/`prev`/`use <model>` 子命令
- `internal/llmselect/selection.go` - `MainSelection` 新增 `ManualModel` 字段
- `internal/llmselect/view.go` - 视图渲染支持手动模型覆盖
- `internal/llmutil/routes.go` - `ProfileConfig` 新增 `Models` 字段
- `providers/uniai/client.go` - 修复 `openai_custom` provider 处理
- `cmd/mistermorph/runcmd/run.go` - `run` 命令支持 `--profile` flag
- `cmd/mistermorph/registry.go` - 移除重复的 `configdefaults.Apply` 调用

#### main 分支已原生支持（无需移植）

**PR #2: Persona/Prompt 改进** ✅（已存在）
- `buildUserName()` 优先 `user.Current()`，回退 `$USER` - 更健壮
- `buildUserPrompt()` 绿色背景用户提示 - 已实现
- Persona 名称加载和显示 - 已实现
- TAB 自动补全（`/exit`, `/init`, `/model`, `/help` 等）- 已实现（更完善）
- /init 不用 write_file 提示 - 已实现

**PR #4: Readline + Ctrl+C** ✅（已存在）
- `github.com/chzyer/readline` 替换 bufio - 已实现
- 历史记录文件 `~/.mistermorph_chat_history` - 已实现
- `os/signal.Notify` 处理 SIGINT - 已实现（最终版本）
- turnCancel() 取消当前 turn - 已实现

**PR #5: /init 命令增强** ✅（已存在）
- `/init` 生成 AGENTS.md - 已实现
- 系统 prompt 包含 /init 说明 - 已实现
- /init thinking animation - 已实现
- /init 后注入 AGENTS.md 到对话历史 - 已实现

### 编译验证

```bash
go build ./cmd/mistermorph        # ✅ 通过
go test ./...                     # ✅ 全部通过（含新增 internal/clifmt 测试）
```

### 剩余未移植

| 组 | 提交 | 说明 |
|---|------|------|
| PR #7 | `794b86a`, `baa3266`, `9f1779e` | ACP gemini_oauth 实时进度反馈 → Telegram/Slack 路由 → ACP LLM 结构化日志 |
| PR #8 | `b4ed69f`, `696e0b4` | Compact Mode（最小化提示/输出显示） |
| PR #8 | `fc98225` | docs: merge help/ docs into docs/ |
| PR #8 | `c898e94` | feat: add one-click deploy script for opencode2api |
| PR #8 | `3668905` | chore: remove accidentally committed npm tarball |

### 待办

- [ ] 移植 PR #7: ACP gemini_oauth 实时进度 → Telegram/Slack
- [ ] 移植 PR #8: Compact Mode
- [ ] 评估 PR #8 杂项（文档合并、部署脚本）是否需要移植
- [ ] 上游更新时同步 main：`git merge upstream/master`

---

*记录格式: [日期] - [事项] - [状态]*
