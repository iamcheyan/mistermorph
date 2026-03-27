---
title: 安装与配置
description: 安装方式与基础配置模型。
---

# 安装与配置

## 安装方式

```bash
# 发布版安装脚本
curl -fsSL -o /tmp/install-mistermorph.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install-mistermorph.sh
```

```bash
# Go 安装
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

## 初始化文件

```bash
mistermorph install
```

默认工作目录是 `~/.morph/`。

## 配置来源优先级

- CLI flags
- 环境变量
- `config.yaml`

## 最小 `config.yaml`

```yaml
llm:
  provider: openai
  model: gpt-5.4
  endpoint: https://api.openai.com
  api_key: ${OPENAI_API_KEY}
```

完整键列表以 `assets/config/config.example.yaml` 为准。
