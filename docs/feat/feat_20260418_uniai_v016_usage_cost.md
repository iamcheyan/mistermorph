---
date: 2026-04-18
title: uniai v0.1.16 升级与 Usage Cost 接入
status: draft
---

# uniai v0.1.16 升级与 Usage Cost 接入

## 1) 目标

这次只解决一个很具体的问题：

1. 将 `github.com/quailyquaily/uniai` 依赖声明更新到 `v0.1.16`
2. 把 `uniai` 的价格配置注入能力接到 `mistermorph`
3. 让 `mistermorph` 在最终 `usage` 里直接读到 cache token 与 cost 信息

不在这次范围里的事：

- 不做账单对账
- 不做 Console 新价格视图
- 不改 LLM usage journal 的存储格式
- 不引入第二套 pricing schema

## 2) 直接回答问题

问题是：

> `uniai v0.1.16` 支持注入价格配置，这样可以在 usage 里直接读到 cost 相关字段，对吧？

结论分两层。

### 2.1 在 `uniai` 里，答案是对

当且仅当下面两个条件同时成立时，`uniai` 会填充 `Usage.Cost`：

1. `Config.Pricing != nil`
2. pricing catalog 里存在匹配当前 `provider + model` 的规则

此时：

- `Usage.Cache` 会带缓存 token 拆分
- 阻塞 `Chat()` 返回值会带 `resp.Usage.Cost`
- 流式场景下，最终一次 `ev.Usage` 也会带 `Cost`

但这个 `Cost` 是本地推导值，不是上游厂商账单原文。

### 2.2 在当前 `mistermorph` 里，答案还不是

当前代码还没有把这条链路接通：

- `go.mod` 里还是 `uniai v0.1.11`
- `providers/uniai/client.go` 只复制了顶层 token 数，没有复制 cache 和 cost
- `llm.Usage` 目前只有一个扁平的 `Cost float64`
- 配置层还没有把 pricing catalog 传给 `uniai.Config`

所以现在即使底层 `uniai` 能算出 `Usage.Cost`，`mistermorph` 这一层也会把它丢掉。

## 3) 当前事实

基于当前仓库代码，已经确认这些事实：

- `go.mod` 当前声明的是 `github.com/quailyquaily/uniai v0.1.11`
- 开发工作区里的 `go.work` 已经把 sibling `uniai` 模块纳入进来
- 当前本地 `uniai` 工作区版本已经是 `v0.1.16`
- `uniai v0.1.16` 已有：
  - `Config.Pricing *PricingCatalog`
  - `ParsePricingYAML([]byte)`
  - `resp.Usage.Cache`
  - `resp.Usage.Cost`
  - 最终流式 `ev.Usage.Cost`
- `providers/uniai/client.go` 当前在两个位置只映射了：
  - `InputTokens`
  - `OutputTokens`
  - `TotalTokens`
- 当前没有映射：
  - `Cache.CachedInputTokens`
  - `Cache.CacheCreationInputTokens`
  - `Cache.Details`
  - `Cost`
- `agent/context.go` 目前按 `usage.Cost` 累加 `Metrics.TotalCost`
- `internal/llmstats` 目前只记录 token，不记录价格

这里还有一个很容易误判的点：

> 本地开发时，因为 `go.work` 的存在，测试和编译可能已经实际使用 `v0.1.16` 源码；但发布和非 workspace 环境仍然以 `go.mod` 为准。

所以这次需要同时处理“依赖声明”和“运行时映射”，不能只改其中一边。

## 4) 第一性原理约束

1. cost 的计算源头只能有一个。  
   应该复用 `uniai` 的 `PricingCatalog` 计算结果，不在 `mistermorph` 再算一次。

2. 不要重新定义 pricing schema。  
   直接复用 `uniai.ParsePricingYAML(...)` 和它已有的 YAML 格式。

3. 共享 `llm.Usage` 应该贴近上游真实语义。  
   如果上游 `Usage.Cost` 是结构化对象，这里就直接接结构化对象，不再额外发明一个扁平总价字段。

4. 聚合逻辑应该依赖结构化 cost 的 `Total`。  
   `Metrics.TotalCost` 只是一个聚合结果，不应该反过来主导共享数据结构设计。

5. cache token 与 cache cost 是同一个问题的一部分。  
   如果已经接 `Usage.Cost`，就不能把 `Usage.Cache` 丢掉，否则 cache cost 的来源不完整。

6. 老数据不迁移。  
   缺失的新字段一律按 0 或 `nil` 处理，不单独做回填脚本。

## 5) 建议方案

### 5.1 依赖升级

把 `go.mod` 中的 `github.com/quailyquaily/uniai` 更新到 `v0.1.16`。

这是发布正确性的要求，不是样式问题。

### 5.2 共享 `llm.Usage` 结构最小扩展

建议直接让 `llm.Usage` 对齐 `uniai` 的 cost 语义：

```go
type Usage struct {
    InputTokens  int
    OutputTokens int
    TotalTokens  int
    Cache        UsageCache
    Cost         *UsageCost
}

type UsageCache struct {
    CachedInputTokens        int
    CacheCreationInputTokens int
    Details                  map[string]int
}

type UsageCost struct {
    Currency           string
    Estimated          bool
    Input              float64
    CachedInput        float64
    CacheCreationInput float64
    Output             float64
    Total              float64
}
```

为什么这比“保留 `float64 Cost` 再补一个 `CostBreakdown`”更简单：

- 不重复保存同一个事实
- provider 映射只需要维护一种 cost 形状
- 调用方不会遇到 `Cost` 和 `CostBreakdown.Total` 不一致的问题
- `mistermorph` 不再维护一套自定义 cost 语义
- cache token 与 cache cost 能在同一份 `Usage` 里闭合

这会带来一个明确代价：

- 现有直接访问 `usage.Cost` 这个 `float64` 的代码要一起改

但当前仓库里这类使用点很少，主要就是 `agent.Context` 的总价累加和少量测试。  
为了换取更干净的数据模型，这个代价是可接受的。

### 5.3 provider 映射

`providers/uniai/client.go` 需要在两条路径上都补全映射：

1. `Client.Chat(...)` 的最终 `resp.Usage`
2. `WithOnStream(...)` 的最终 `ev.Usage`

映射规则：

- `llm.Usage.Cache = mapped(resp.Usage.Cache)`
- `llm.Usage.Cost = mapped(resp.Usage.Cost)`
- 若上游 `Cache` 缺字段，则对应字段保留零值
- 若上游 `Cost == nil`，则 `Cost = nil`

这样最终 `usage` 上的 cache 和 cost 语义就和 `uniai` 一致，不需要额外转换出第二个总价字段。

### 5.4 pricing 配置注入

建议只加一个最小入口：

```yaml
llm:
  pricing_file: "./pricing.yaml"
```

实现规则：

1. 读取 `llm.pricing_file`
2. 文件非空时读取 YAML 内容
3. 调用 `uniai.ParsePricingYAML(...)`
4. 将结果传入 `uniaiProvider.Config.Pricing`

一旦 pricing 命中，`uniai` 会基于：

- `Usage.InputTokens`
- `Usage.OutputTokens`
- `Usage.Cache.CachedInputTokens`
- `Usage.Cache.CacheCreationInputTokens`
- `Usage.Cache.Details`

推导出：

- `Cost.Input`
- `Cost.CachedInput`
- `Cost.CacheCreationInput`
- `Cost.Output`
- `Cost.Total`

这次不建议做：

- `llm.pricing` 内联大对象
- profile 级单独 pricing file
- 自定义 mistermorph pricing schema

原因很简单：

- 一个 pricing catalog 已经能覆盖多个 provider/model
- 直接复用 `uniai` 现有 YAML，认知成本最低
- profile 级拆分并不是当前问题的必要条件

### 5.5 指标与统计边界

`agent.Context.Metrics.TotalCost` 改为累加 `usage.Cost.Total`。  
这意味着：

- 主 agent 运行时总价可以直接继续工作
- 不需要在这次把 `agent.Context` 改成复杂结构
- 聚合逻辑和共享 `Usage` 模型的职责边界更清楚

但 `internal/llmstats` 这次先不扩。

原因：

- 现有 usage journal 设计文档明确以 token 为主
- 把价格落盘会引入新的存储兼容问题
- 这与“先把 usage 读到 cost”不是一个最小问题

## 6) 预期行为

完成后，行为应该是这样的：

1. 未配置 `llm.pricing_file`
   - `usage.Cache` 仍可有值
   - `usage.Cost == nil`

2. 配置了 `llm.pricing_file`，但没有匹配规则
   - `usage.Cache` 仍可有值
   - `usage.Cost == nil`

3. 配置了 `llm.pricing_file`，且规则命中
   - `usage.Cache` 带缓存 token 拆分
   - `usage.Cost != nil`
   - `usage.Cost.Total > 0`
   - 若存在缓存命中或缓存写入：
     - `usage.Cost.CachedInput >= 0`
     - `usage.Cost.CacheCreationInput >= 0`

4. 流式场景
   - 只有最终那次 `ev.Usage` 才应该带 cost
   - 中间 delta 不应伪造不完整的价格

5. 老数据或旧响应
   - 没有 `Cache` 字段时，缓存 token 视为 0
   - 没有 `Cost` 字段时，cost 视为未知，即 `nil`
   - 读取方如果只关心数值聚合，缺失项按 0 处理

## 7) 测试点

至少补这些测试：

1. `providers/uniai/client_test.go`
   - 阻塞结果映射 `Usage.Cache`
   - 阻塞结果映射 `Usage.Cost`
   - 流式最终事件映射 `Usage.Cache`
   - 流式最终事件映射 `Usage.Cost`

2. `llm/llm_test.go`
   - `Usage.Cache` 缺字段时零值正确
   - `Usage.Cost` 为 `nil` 时默认行为正确
   - `Usage.Cost.Total` 可正常读取

3. `agent/context_test.go`
   - `usage.Cost != nil` 时，`TotalCost` 按 `usage.Cost.Total` 累加
   - `usage.Cost == nil` 时，`TotalCost` 保持不变

4. `internal/llmutil` 相关测试
   - `llm.pricing_file` 未设置时不报错
   - pricing file 非法 YAML 时返回清晰错误
   - pricing file 合法时能成功注入到 provider config

## 8) 实施顺序

建议按下面顺序做：

1. 更新 `go.mod` 到 `uniai v0.1.16`
2. 扩展 `llm.Usage` 结构
3. 补 `providers/uniai` 的 cost 映射
4. 补 `providers/uniai` 的 cache 映射
5. 接 `llm.pricing_file` 到 `uniai.Config.Pricing`
6. 补测试

这个顺序能保证每一步都小，而且每一步都能独立验证。
