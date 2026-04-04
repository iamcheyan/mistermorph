---
title: 记忆（Memory）
description: 介绍基于 WAL 的 memory 架构、注入和回写规则。
---

# Memory

Mister Morph 的 memory 系统采用先 WAL（Write-ahead logging）追加，再异步投影的机制。

## WAL

说白了，就是发生的大小事项都会以一个 jsonl 格式，按照发生顺序，把原始数据记录到文件。于是 WAL 才能提供真正的数据源。

路径为 `memory/log/`，文件名字格式为 `since-YYYY-MM-DD-0001.jsonl`。

WAL 的文件大小达到一定程度的时候，会转储成 `.jsonl.gz` 结尾的文件。

## 投影

Mister Morph 会根据 WAL 进行简单投影，投影的对象有两个：

- `memory/index.md`（长期记忆）
- `memory/YYYY-MM-DD/*.md`（短期记忆）
  - 短期记忆的文件会根据所在的 Channel 进行隔离。例如，Telegram 下不同群聊的记忆是不互通的。

所谓的投影就是读取一段时间内的 WAL，使用 LLM 进行总结归纳，写入到对应的对象文件中。

投影的记录点文件在 `memory/log/checkpoint.json`，内容大概是：

```json
{
  "file": "since-2026-02-28-0001.jsonl", // 当前处理的 WAL 文件
  "line": 18,                            // 当前处理行
  "updated_at": "2026-02-28T06:30:12Z"   // 更新时间
}
```

## 注入

当下面的配置开启时，一部分记忆的投影会被注入到 prompt 中：

- `memory.enabled = true`
- `memory.injection.enabled = true`

配置 `memory.injection.max_items` 控制注入条目数上限。

## 备注

1. Memory 投影损坏可由 WAL 重建，删除 `memory/log/checkpoint.json` 然后让 Agent 持续运行即可。
2. 生产环境把 `file_state_dir` 放到持久化存储来维持运行状态（包括记忆）。
