---
title: 总览
description: 了解 Mister Morph 的能力边界和推荐阅读路径。
---

# 总览

Mister Morph 主要有两种使用方式：

- CLI 工作流（`mistermorph run`、`telegram`、`slack`、`console serve`）
- Go 内嵌 core（`integration` 包）

## 选择你的路径

- 先快速跑通：看 [快速开始（CLI）](/zh/guide/quickstart-cli)
- 嵌入到 Go 项目：看 [用 Core 快速搭建 Agent](/zh/guide/build-agent-with-core)
- 了解长期运行入口：看 [Runtime 模式](/zh/guide/runtime-modes)
- 上线前治理：看 [安全与 Guard](/zh/guide/security-and-guard)

## 仓库结构快照

- CLI 入口：`cmd/mistermorph/`
- Agent 引擎：`agent/`
- 嵌入 Core：`integration/`
- 内置工具：`tools/`
- Provider 后端：`providers/`
- 详细文档：`docs/`
