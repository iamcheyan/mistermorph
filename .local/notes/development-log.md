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

*记录格式: [日期] - [事项] - [状态]*
