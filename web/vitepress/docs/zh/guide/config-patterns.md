---
title: 配置模式
description: 安装方式、基础配置，以及常用的 profiles、routes 与工具策略配置方法。
---

# 配置模式

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

## LLM Profiles 与 Routes

```yaml
llm:
  model: gpt-5.4
  profiles:
    cheap:
      model: gpt-4.1-mini
    backup:
      provider: xai
      model: grok-4.1-fast-reasoning
  routes:
    main_loop:
      candidates:
        - profile: default
          weight: 1
        - profile: cheap
          weight: 1
      fallback_profiles: [backup]
    addressing:
      profile: cheap
    heartbeat: cheap
```

## 工具开关

```yaml
tools:
  bash:
    enabled: false
  url_fetch:
    enabled: true
    timeout: "30s"
```

## 运行上限

```yaml
max_steps: 20
tool_repeat_limit: 4
```

完整键定义以 `assets/config/config.example.yaml` 为准。
