---
title: 快速开始（CLI）
description: 几分钟内完成可运行的 CLI 配置。
---

# 快速开始（CLI）

## 1. 安装

```bash
curl -fsSL -o /tmp/install.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install.sh
```

## 2. 初始化

```bash
mistermorph install
```

Mister Morph 会初始化所需的文件，默认情况下，会安装到 `~/.morph/`，并以 `~/.morph/config.yaml` 为配置文件。

初始化过程中，Mister Morph 会向用户询问最小化配置，包括 LLM 配置信息，Agent 名字和个性等。


### 2.1 可选：使用环境变量进行配置

在需要加强安全性的环境中，Mister Murph 支持将敏感信息放到环境变量，并在配置文件中引用。

例如，可以把 LLM 的 API key 写到环境变量：

```bash
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
```

然后在配置文件中引用它：

```yaml
llm:
  api_key: "${MISTER_MORPH_LLM_API_KEY}"
```

## 3. 跑第一个任务

```bash
mistermorph run --task "Hello"
```

它可能输出：

```json
{
  "reasoning": "Greet the user briefly.",
  "output": "Hello 👀",
  "reaction": "👀"
}
```

## 4. 调试开关

```bash
mistermorph run --inspect-prompt --inspect-request --task "Hello"
```

此时会在当前目录下生成 `dump`，包含 prompt 和请求的详细信息。

完整配置项见 [配置模式](/zh/guide/config-patterns) 和 `assets/config/config.example.yaml`。
