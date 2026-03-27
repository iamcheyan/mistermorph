---
title: Memory
description: 基于 WAL 的 memory 架构、注入和回写规则。
---

# Memory

Mister Morph 的 memory 采用“先 WAL 追加，再异步投影”。

## 架构

- 真正的数据源：`memory/log/*.jsonl`（WAL）
- 可读投影：
  - `memory/index.md`（长期）
  - `memory/YYYY-MM-DD/*.md`（短期）
- 热路径只写 WAL，Markdown 投影异步更新。

## 运行流程

1. LLM 调用前：准备 memory snapshot 并注入 prompt block
2. 最终回复后：写入原始 memory event 到 WAL
3. 投影 worker 回放 WAL 并更新 markdown

## 注入条件

只有全部满足才注入：

- `memory.enabled = true`
- `memory.injection.enabled = true`
- runtime 提供有效 `subject_id`
- snapshot 非空

`memory.injection.max_items` 控制注入条目数上限。

## 回写条件

只有全部满足才回写：

- 最终回复需要实际发布
- memory orchestrator 存在
- `subject_id` 存在

轻量回复或空输出会跳过回写。

## 关键配置

```yaml
memory:
  enabled: true
  dir_name: "memory"
  short_term_days: 7
  injection:
    enabled: true
    max_items: 50
```

## 运维建议

- 投影损坏可由 WAL 重建。
- 生产环境把 `file_state_dir` 放到持久化存储。
