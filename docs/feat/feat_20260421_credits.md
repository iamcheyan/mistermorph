---
date: 2026-04-21
title: Credits in CLI and Console
status: in_progress
---

# Credits in CLI and Console

## 1) 目标

这次需求只解决一个直接问题：

- 产品里需要一份统一的 `credits` 清单，CLI 和 Console 都能展示同一份内容

这里的 `credits` 是上层概念，不等于 `licenses`。

当前这份清单首版就包含两类内容：

- 主要开源项目
- 主要贡献者

这不是为了预埋一个很远的扩展点。当前仓库已经有多个贡献者，所以 contributors section 不是以后再说，而是现在就该纳入。

这份清单要同时服务两处入口：

- 新增一个 CLI 命令，展示 credits
- 在 Console 的 Settings 下新增一个子页面，展示同一份内容

重点不是自动扫描一切依赖，也不是做成一个复杂的档案系统。重点只是把产品需要公开说明的事实，放进一份简单、稳定、可复用的数据里。

## 2) 非目标

这版不做下面这些事：

- 不做全量 SBOM
- 不做所有传递依赖的枚举
- 不在 CLI 或 Web 里展示完整 LICENSE 正文
- 不做运行时动态扫描 `go.mod`、`pnpm-lock.yaml` 或 `node_modules`
- 不把文档站、构建链、部署链工具都塞进首版清单
- 不在这期里解释各协议的法律义务；这里只展示事实信息
- 不做多语言 credits 描述；首版只保留一套固定展示文案
- 不自动从 GitHub contributors graph 或 Git 历史实时生成 contributors
- 不做贡献者打分、排序算法、社交资料聚合

## 3) 顶层概念

### 3.1 为什么不能继续叫 `licenses`

如果这份内容里还要包含主要贡献者，下面这些命名都会马上失真：

- `mistermorph licenses`
- `/settings/open-source`
- `/api/settings/open-source`
- `assets/open_source/projects.json`

这些名字都在表达“这里只放第三方开源项目和协议”。贡献者不是这个语义。

所以顶层概念应该统一成：

- `credits`

开源项目和主要贡献者只是 `credits` 下面的两个 section。

### 3.2 为什么不需要更复杂的模型

这个需求的本质很简单：

- 有两组公开信息
- 两端共用一份数据
- 用固定顺序展示出来

所以不需要：

- 把所有条目压成一维数组再靠 `type` 区分
- 设计一套复杂的继承或多态模型
- 额外引入排序规则、筛选规则、评分规则

顶层保持两个数组就够了：

- `open_source`
- `contributors`

## 4) 首版纳入范围

### 4.1 开源项目

首版只纳入和产品能力面直接相关的主要开源项目：

- `go.mod` 里的主要直接依赖
- `web/console` 的主要直接依赖
- `web/markdown-renderer` 的主要直接依赖，且产物会进入 Console

首版明确不纳入：

- `vite`、`vitepress`、`wrangler`、`esbuild` 这类构建、文档、发布工具
- 只在测试里出现、但不进入产品能力面的库
- 价值很低的通用小依赖

### 4.2 主要贡献者

首版 contributors 名单以当前仓库 Git 历史为候选集合，再做最小人工筛选。

这里保持简单：

- 以当前仓库 Git author 为候选集合
- 明显属于同一人的 author name / email 允许手工合并
- 只有很小的文档修正或零碎改动的人，不必进入 contributors
- GitHub profile URL 允许手工映射
- 不在运行时实时扫描 Git 历史
- GitHub 头像直接由 profile URL 派生，不做额外抓取逻辑

每条贡献者记录只保留最小信息：

- 名字
- 链接
- 一句话说明

## 5) 当前首版开源项目清单

### 5.1 当前 Go / Desktop 清单

| 项目 | 链接 | 协议 | 当前用途 |
| --- | --- | --- | --- |
| Cobra | https://github.com/spf13/cobra | Apache-2.0 | 构建 CLI 命令树、参数解析和帮助输出。 |
| Viper | https://github.com/spf13/viper | MIT | 读取配置文件、环境变量和运行时设置。 |
| Gorilla WebSocket | https://github.com/gorilla/websocket | BSD-2-Clause | 处理 Slack 等通道里的 WebSocket 连接。 |
| UniAI | https://github.com/quailyquaily/uniai | Apache-2.0 | 统一对接多个 LLM 提供方，并承接请求与计费模型。 |
| Goldmark | https://github.com/yuin/goldmark | MIT | 把 Telegram 等场景里的 Markdown 转成受控 HTML。 |
| Wails | https://github.com/wailsapp/wails | MIT | 支撑可选的桌面壳和桌面端应用入口。 |

### 5.2 当前 Console 前端清单

#### `web/console`

| 项目 | 链接 | 协议 | 当前用途 |
| --- | --- | --- | --- |
| Vue | https://github.com/vuejs/core/tree/main/packages/vue | MIT | 搭建 Console SPA 的组件、状态和渲染模型。 |
| Quail UI | https://github.com/quailyquaily/quail-ui/ | Apache-2.0 | 提供 Console 当前使用的组件、样式和图标体系。 |
| OverType | https://github.com/panphora/overtype | MIT | 提供 Markdown 编辑器能力。 |

#### `web/markdown-renderer`

| 项目 | 链接 | 协议 | 当前用途 |
| --- | --- | --- | --- |
| unified | https://github.com/unifiedjs/unified | MIT | 组织整个 Markdown 到 HTML 的处理流水线。 |
| remark-parse | https://github.com/remarkjs/remark/tree/main/packages/remark-parse | MIT | 解析 Markdown 源文本。 |
| remark-gfm | https://github.com/remarkjs/remark-gfm | MIT | 支持 GFM 表格、任务列表等扩展语法。 |
| remark-math | https://github.com/remarkjs/remark-math/tree/main/packages/remark-math | MIT | 解析数学公式块和行内公式。 |
| remark-rehype | https://github.com/remarkjs/remark-rehype | MIT | 把 remark AST 转成 rehype AST。 |
| rehype-raw | https://github.com/rehypejs/rehype-raw | MIT | 把 Markdown 里的原始 HTML 纳入统一处理链。 |
| rehype-stringify | https://github.com/rehypejs/rehype/tree/main/packages/rehype-stringify | MIT | 把处理后的 rehype AST 输出成 HTML。 |
| rehype-katex | https://github.com/remarkjs/remark-math/tree/main/packages/rehype-katex | MIT | 把数学公式节点转成 KaTeX HTML。 |
| DOMPurify | https://github.com/cure53/DOMPurify | MPL-2.0 OR Apache-2.0 | 对最终 HTML 做净化，降低注入风险。 |
| Mermaid | https://github.com/mermaid-js/mermaid | MIT | 渲染 Mermaid 图表块。 |
| KaTeX | https://github.com/KaTeX/KaTeX | MIT | 渲染数学公式。 |
| Shiki | https://github.com/shikijs/shiki | MIT | 提供代码高亮。 |
| Shiki Stream | https://github.com/antfu/shiki-stream | MIT | 处理流式代码高亮分词。 |
| Viz.js | https://github.com/mdaines/viz-js | MIT | 渲染 Graphviz / DOT 图。 |
| AntV Infographic | https://github.com/antvis/Infographic | MIT | 渲染 infographic 自定义块。 |

### 5.3 当前 contributors section 的要求

当前仓库已经不是单作者状态，所以 contributors section 不能留到以后。

但首版也不需要把名单问题搞复杂。实现上只需要做到：

- 以当前仓库 Git 历史收集 contributors 候选名单
- 对明显别名做最小合并
- 排除只有很小文档修正或零碎提交的人
- 给每个人补一个 GitHub profile URL
- contributors section 首版不应为空
- 记录名字、链接、说明
- 和开源项目一起放进同一份 `credits` 数据里

## 6) 单一数据源

这类信息不能在 CLI、Console backend、Console 页面里各写一份。

首版应该只有一份结构化 credits 数据，建议放在：

- `assets/credits/data.json`

### 6.1 数据结构

顶层结构保持最小，不做抽象过度：

```json
{
  "open_source": [
    {
      "id": "cobra",
      "name": "Cobra",
      "link": "https://github.com/spf13/cobra",
      "license": "Apache-2.0",
      "summary": "Builds the CLI command tree and help output."
    }
  ],
  "contributors": [
    {
      "id": "contributor_1",
      "name": "Contributor Name",
      "link": "https://example.com/contributor",
      "summary": "Made a notable product contribution."
    }
  ]
}
```

### 6.2 字段规则

`open_source` 条目字段：

- `id`：稳定主键
- `name`
- `link`
- `license`
- `summary`

`contributors` 条目字段：

- `id`：稳定主键
- `name`
- `link`
- `summary`

这里不要：

- 给开源项目加 `surface`
- 给贡献者加 `role`
- 把两类条目硬塞成一维数组

### 6.3 读取原则

实现上只需要做到：

- Go 端从这份 JSON 读取成 `CreditsData`
- CLI 命令直接用它
- Console backend 接口直接用它
- Console 页面通过接口读出来展示

数组顺序就是最终展示顺序，不再额外引入排序规则。

## 7) CLI 命令

### 7.1 命令名

建议新增：

- `mistermorph credits`

不要再叫 `mistermorph licenses`。

原因很直接：

- 展示内容不只会有 license
- 现在已经要纳入 contributors
- 这个名字和上层概念一致

### 7.2 输出要求

CLI 只需要按两个 section 输出：

- `Open Source`
- `Contributors`

开源项目记录展示：

- 项目名
- 协议名
- 链接
- 一句话用途

贡献者记录展示：

- 名字
- 链接
- 一句话说明

这里不要求复杂表格，也不要求二级分组。核心要求只有一个：直接可读。

## 8) Console Settings 子页面

### 8.1 路由形态

这块不应该继续塞回现在的大型 `SettingsView` 里当一个 section。

建议新增一个明确的子路由：

- `/settings/credits`

这样做有两个直接好处：

- URL 语义清楚，能直接分享和刷新
- 不需要以后再为 contributors 改路由

### 8.2 页面位置

页面挂在 Settings 之下，但不是顶层导航的新入口。

首版可以这样处理：

- Settings 主页面保留现状
- 在 Settings 内部增加一个二级入口，进入 `Credits` 子页

### 8.3 页面内容

页面展示和 CLI 保持同一份内容，但表达方式更适合 Web：

- 顶部先说明：这是产品维护的一份精选 credits，不是全量依赖清单
- 页面内分成 `Open Source` 和 `Contributors`
- 每条记录展示名字、协议或链接、说明

首版不需要：

- 搜索
- 排序
- 筛选
- 下载 CSV
- 展示完整 LICENSE 正文

### 8.4 现有路由约束

当前 router 只把 `/settings` 本身当作 setup-free 路径。

如果新增 `/settings/credits`，实现时要一起更新这条规则，否则在 setup 未完成时，子页行为会和主设置页不一致。

## 9) Console API

这份数据是 Console 自己的静态产品信息，不属于 runtime proxy 的动态数据面。

所以首版直接在 Console backend 增一个只读接口即可：

- `GET /api/settings/credits`

返回体保持最小：

```json
{
  "open_source": [
    {
      "id": "cobra",
      "name": "Cobra",
      "link": "https://github.com/spf13/cobra",
      "license": "Apache-2.0",
      "summary": "Builds the CLI command tree and help output."
    }
  ],
  "contributors": [
    {
      "id": "contributor_1",
      "name": "Contributor Name",
      "link": "https://example.com/contributor",
      "summary": "Made a notable product contribution."
    }
  ]
}
```

这里只读，不做增删改。

这样有两个直接收益：

- Console 页面不需要自己维护第二份静态数据
- CLI 和 Web 的事实来源保持一致

## 10) 建议落点

首版实现落点建议如下：

- `assets/credits/data.json`：唯一数据源
- `cmd/mistermorph/credits.go`：CLI 命令
- `cmd/mistermorph/consolecmd/credits.go`：Console backend 只读接口
- `web/console/src/views/SettingsCreditsView.js`：Settings 子页
- `web/console/src/router/index.js`：新增 `/settings/credits` 路由，并修正 setup-free 规则

这里不建议：

- 把清单复制到 `web/console/src` 再读一遍
- 让前端直接读仓库文件
- 为这点静态数据引入新的存储或后台任务

## 11) 验收标准

首版完成后，至少满足下面这些条件：

- `mistermorph credits` 可直接输出首版清单
- CLI 输出里每条开源项目记录都有项目名、链接、协议名、用途说明
- CLI 输出里每条贡献者记录都有名字、链接、说明
- Console 可从 Settings 进入 `Credits` 子页
- `/settings/credits` 可直接访问和刷新
- Console 页面展示的开源项目和贡献者集合、顺序、字段值与 CLI 一致
- Console 页面数据来自 `GET /api/settings/credits`
- 新增或修改 credit 时，只需要改一份 `assets/credits/data.json`

## 12) Checklist

- [x] 把顶层概念从 `licenses` / `open source notices` 调整为 `credits`
- [x] 明确首版同时包含 open source 和 contributors
- [x] 扫描当前 Go 主要直接依赖
- [x] 扫描当前 Console 直接依赖
- [x] 扫描当前 shipped markdown renderer 依赖
- [x] 明确 contributors section 由维护者手工维护，不做运行时自动生成
- [x] 记录每个纳入项目的链接、协议、用途
- [x] 明确首版不做全量 SBOM
- [x] 明确 CLI 和 Console 共享同一份数据源
- [x] 明确 Console 走 Settings 子页面，不新开顶层导航
- [x] 明确 Console backend 只增加一个只读接口
