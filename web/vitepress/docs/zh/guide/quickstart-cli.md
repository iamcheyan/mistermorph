---
title: 快速开始（CLI）
description: 几分钟内完成可运行的 CLI 配置。
---

# 快速开始（CLI）

## 1. 安装

```bash
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

## 2. 初始化工作目录

```bash
mistermorph install
```

## 3. 设置模型凭据

```bash
export MISTER_MORPH_LLM_PROVIDER="openai"
export MISTER_MORPH_LLM_MODEL="gpt-5.4"
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
```

## 4. 跑第一个任务

```bash
mistermorph run --task "Summarize this repository"
```

## 5. 调试开关

```bash
mistermorph run --inspect-prompt --inspect-request --task "hello"
```

完整配置项见 [配置模式](/zh/guide/config-patterns) 和 `assets/config/config.example.yaml`。
