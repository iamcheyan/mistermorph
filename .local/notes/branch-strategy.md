# 分支管理策略与开发工作流程

> 本文档定义本仓库的分支结构、职责划分和日常开发工作流程。
> 最后更新: 2026-04-19

---

## 分支总览

| 分支 | 用途 | 目标受众 | 备注 |
|------|------|----------|------|
| `main` | **全能版**，包含所有个人修改 + 上游代码 | 直接 clone 使用的用户 | GitHub 默认分支 |
| `feat/xxx` | **给上游的 PR 分支**，只含特定功能 | upstream (quailyquaily) | 必须从 `upstream/master` 新建 |
| `backup-main-20260419` | **原始开发备份**，36 个提交完整历史 | 从中挑选代码 | 长期保留，不删除 |
| `pr-chat` | PR #35 chat 命令 | upstream (等待 review) | 已发 PR，等 lyricat review |

---

## 核心原则

### main = 全能版（给别人用）

`main` 分支是**个人维护的全能版**，包含：
- ✅ 上游最新代码
- ✅ 中文 README（展示 fork 差异）
- ✅ `.local/` 私有工作区
- ✅ Bedrock provider（AWS CLI 免 API Key）
- ✅ chat 交互命令（cherry-pick 自 pr-chat）
- ⏳ 其他 backup 功能（逐个挑选合并）

别人直接 clone 就能用：
```bash
git clone https://github.com/iamcheyan/mistermorph.git
# 默认就是 main，包含所有功能
```

### feat/xxx = 干净 PR（给上游）

**⚠️ 绝对不要从 `main` 发 PR 给上游！** `main` 包含太多个人修改（中文 README、.local/ 等）。

正确流程：
```bash
# 1. 从 upstream/master 新建干净分支
git fetch upstream
git checkout -b feat/xxx upstream/master

# 2. 开发 / cherry-pick 需要的提交
# ...

# 3. 推送并创建 PR
git push origin feat/xxx
# 在 GitHub 上对比 upstream/master 创建 PR
```

### PR 合并后的同步

上游合并 PR 后，`main` 执行：
```bash
git fetch upstream
git checkout main
git merge upstream/master
```

这样 `main` 自动获得该功能，无需重复 cherry-pick。

---

## 当前分支状态

### `main`（当前所在分支）
- 基于 `upstream/master` + 个人修改
- 已合并: Bedrock provider, chat 命令
- 待合并: backup 分支中的其他功能

### `feat/bedrock-aws-cli`
- 已推送，可创建 PR 给上游
- 只包含 Bedrock provider 相关改动

### `pr-chat`
- PR #35，等待 `lyricat` review
- 包含 chat 命令基础框架

### `backup-main-20260419`
- 36 个提交的完整备份
- 从中挑选功能合并到 `main`

---

## 日常开发流程

### 场景 1：从 backup 挑选功能合并到 main

```bash
# 1. 查看 backup 分支中的提交
git log backup-main-20260419 --oneline

# 2. 查看某个提交的详细改动
git show <commit-hash>

# 3. cherry-pick 到 main
git checkout main
git cherry-pick <commit-hash>
# 如有冲突，解决后 git cherry-pick --continue

# 4. 编译测试
go build -o ./bin/mistermorph ./cmd/mistermorph

# 5. 提交
git push origin main
```

### 场景 2：发 PR 给上游

```bash
# 1. 从 upstream/master 新建分支
git fetch upstream
git checkout -b feat/xxx upstream/master

# 2. 添加功能（可以 cherry-pick 已验证的提交）
git cherry-pick <commit-hash>

# 3. 确保只包含目标功能的改动
git diff upstream/master --stat

# 4. 推送
git push origin feat/xxx

# 5. 在 GitHub 创建 PR，base 选择 upstream/master
```

### 场景 3：同步上游更新到 main

```bash
# 上游有更新时
git fetch upstream
git checkout main
git merge upstream/master

# 解决冲突（如果有）
# 编译测试
go build -o ./bin/mistermorph ./cmd/mistermorph

git push origin main
```

---

## 高优先级待办

从 backup 分支挑选以下功能到 `main`：

1. `ee3e08d` — runtime model switching + OpenCode 集成（部分已提取）
2. `01c7787` — system username for prompt（用户体验）
3. `1b51740` — plan progress 彩色格式化（可读性）
4. `6e01e01` — rl.Stdout() 统一 writer（bug 修复）

---

## 注意事项

- `main` 永远保持可编译、可运行
- 发 PR 前务必确认分支基于 `upstream/master`，不是 `main`
- 备份分支 `backup-main-20260419` 不要删除，长期保留
- `.local/` 目录已纳入版本管理，用于存放开发笔记和日志
