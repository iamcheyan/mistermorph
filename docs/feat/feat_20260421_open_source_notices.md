---
date: 2026-04-21
title: Open Source Notices in CLI and Console
status: in_progress
---

# Open Source Notices in CLI and Console

## 1) 目标

这次需求只做一件简单的事：

- 把当前产品里实际用到的一组主要开源项目，整理成一份可读、可查、可复用的清单

这份清单要同时服务两处入口：

- 新增一个 CLI 命令，展示项目名、链接、协议名、以及一句话用途
- 在 Console 的 Settings 下新增一个子页面，用合适的页面结构展示同一份内容

这里的重点不是“自动扫描一切依赖”，而是先把对用户和开发者最有信息量的一批项目说明白。

## 2) 非目标

这版不做下面这些事：

- 不做全量 SBOM
- 不做所有传递依赖的枚举
- 不在 CLI 或 Web 里展示完整 LICENSE 正文
- 不做运行时动态扫描 `go.mod`、`pnpm-lock.yaml` 或 `node_modules`
- 不把文档站、构建链、部署链工具都塞进首版清单
- 不在这期里解释各协议的法律义务；这里只展示事实信息
- 不做多语言开源项目描述；首版只保留一套固定展示文案

## 3) 首版纳入范围

### 3.1 纳入原则

首版只纳入三类项目：

- `go.mod` 里直接依赖，且确实参与 CLI、runtime、channel、desktop 等主路径
- `web/console` 直接依赖
- `web/markdown-renderer` 直接依赖，且产物会被复制进 Console 使用

首版明确不纳入：

- `vite`、`vitepress`、`wrangler`、`esbuild` 这类构建、文档、发布工具
- 只在测试里出现、但不进入产品能力面的库
- 价值很低的通用小依赖

### 3.2 当前 Go / Desktop 清单

| 项目 | 链接 | 协议 | 当前用途 |
| --- | --- | --- | --- |
| Cobra | https://github.com/spf13/cobra | Apache-2.0 | 构建 CLI 命令树、参数解析和帮助输出。 |
| Viper | https://github.com/spf13/viper | MIT | 读取配置文件、环境变量和运行时设置。 |
| Gorilla WebSocket | https://github.com/gorilla/websocket | BSD-2-Clause | 处理 Slack 等通道里的 WebSocket 连接。 |
| Google UUID | https://github.com/google/uuid | BSD-3-Clause | 生成 task、session、message 等稳定 ID。 |
| MCP Go SDK | https://github.com/modelcontextprotocol/go-sdk | Apache-2.0 | 提供 MCP host 接入和 tool adapter 能力。 |
| UniAI | https://github.com/quailyquaily/uniai | Apache-2.0 | 统一对接多个 LLM 提供方，并承接请求与计费模型。 |
| Goldmark | https://github.com/yuin/goldmark | MIT | 把 Telegram 等场景里的 Markdown 转成受控 HTML。 |
| Wails | https://github.com/wailsapp/wails | MIT | 支撑可选的桌面壳和桌面端应用入口。 |

### 3.3 当前 Console 前端清单

#### `web/console`

| 项目 | 链接 | 协议 | 当前用途 |
| --- | --- | --- | --- |
| Vue | https://github.com/vuejs/core/tree/main/packages/vue | MIT | 搭建 Console SPA 的组件、状态和渲染模型。 |
| Vue Router | https://github.com/vuejs/router | MIT | 负责 Console 页面路由和 Settings 子页面切换。 |
| Quail UI | https://quailyquaily.github.io/quail-ui/ | AGPL-3.0 | 提供 Console 当前使用的组件、样式和图标体系。 |
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

### 3.4 说明事项

有三点需要在文档里提前说清楚：

- `modelcontextprotocol/go-sdk` 的 LICENSE 里写了项目正从 MIT 迁移到 Apache-2.0；首版展示值按当前 LICENSE 事实写 `Apache-2.0`，不额外展开法律解释
- `DOMPurify` 是双协议，首版按包元数据原样展示 `MPL-2.0 OR Apache-2.0`
- `quail-ui` 的本地包元数据没有公开仓库字段；首版链接先使用 README 里的项目主页

## 4) 单一数据源

这类信息不能在 CLI、Console API、Console 页面里各写一份。

首版应该只有一份结构化清单，建议放在：

- `assets/open_source/projects.json`

数据模型保持最小：

```json
[
  {
    "id": "cobra",
    "surface": "go",
    "name": "Cobra",
    "link": "https://github.com/spf13/cobra",
    "license": "Apache-2.0",
    "usage": "Builds the CLI command tree and help output."
  }
]
```

字段只保留当前展示真正需要的内容：

- `id`：稳定主键，用于测试和前端渲染 key
- `surface`：分组字段，首版可取 `go`、`desktop`、`console`、`console_markdown`
- `name`
- `link`
- `license`
- `usage`

这里不要再包一层没有价值的 wrapper，也不要再做一份前端专用 JSON。

实现上只需要做到：

- Go 端从这份 JSON 读取成 `[]OpenSourceProject`
- CLI 命令直接用它
- Console 后端接口直接用它
- Console 页面通过接口读出来展示

数组顺序就是最终展示顺序，不再额外引入排序规则。

## 5) CLI 命令

### 5.1 命令名

建议新增：

- `mistermorph licenses`

原因很直接：

- 这是一个静态信息命令
- 用户心智简单，和展示内容一致
- 不依赖当前 runtime、endpoint、API key 或 topic 状态

### 5.2 输出要求

输出内容必须至少包含四列：

- 项目名
- 协议名
- 链接
- 一句话用途

展示要求如下：

- 按 `surface` 分组输出，不要把 Go、Desktop、Console 全混在一起
- 分组内按清单文件顺序输出
- 宽度不够时允许换行，不做截断省略
- 命令在未配置任何 provider 时也能正常运行

示意：

```text
Go / Runtime
NAME              LICENSE          LINK                                   USED FOR
Cobra             Apache-2.0       https://github.com/spf13/cobra        Builds the CLI command tree and help output.
Viper             MIT              https://github.com/spf13/viper        Loads config files and environment overrides.
...
```

如果现有 `NameDetailTable` 不适合四列展示，就直接为这个命令写最小可读的表格输出，不要硬套两列表头。

## 6) Console Settings 子页面

### 6.1 路由形态

这块不应该继续塞回现在的大型 `SettingsView` 里当一个 section。

建议新增一个明确的子路由：

- `/settings/open-source`

这样做有两个好处：

- URL 语义清楚，能直接分享和刷新
- 后续如果 Settings 再拆页，不会继续把一个大文件堆得更重

### 6.2 页面位置

页面挂在 Settings 之下，但不是顶层导航的新入口。

首版可以这样处理：

- Settings 主页面保留现状
- 在 Settings 内部增加一个二级入口，进入 `Open Source` 子页

### 6.3 页面内容

页面展示和 CLI 保持同一份内容，但表达方式更适合 Web：

- 按 `surface` 分组
- 每个项目一行，至少展示名字、协议、链接、用途
- 链接可直接打开上游项目页
- 页面顶部明确说明“这是首版纳入的主要开源项目，不是全量依赖清单”

首版不需要：

- 搜索
- 排序
- 筛选
- 下载 CSV
- 展示完整 LICENSE 正文

### 6.4 现有路由约束

当前 router 只把 `/settings` 本身当作 setup-free 路径。

如果新增 `/settings/open-source`，实现时要一起更新这条规则，否则在 setup 未完成时，子页行为会和主设置页不一致。

## 7) Console API

这份数据是 Console 自己的静态产品信息，不属于 runtime proxy 的动态数据面。

所以首版直接在 Console backend 增一个只读接口即可：

- `GET /api/settings/open-source`

返回体保持最小：

```json
{
  "items": [
    {
      "id": "cobra",
      "surface": "go",
      "name": "Cobra",
      "link": "https://github.com/spf13/cobra",
      "license": "Apache-2.0",
      "usage": "Builds the CLI command tree and help output."
    }
  ]
}
```

这里只读，不做增删改。

这样有两个直接收益：

- Console 页面不需要自己维护第二份静态数据
- CLI 和 Web 的事实来源保持一致

## 8) 建议落点

首版实现落点建议如下：

- `assets/open_source/projects.json`：唯一数据源
- `cmd/mistermorph/licenses.go`：CLI 命令
- `cmd/mistermorph/consolecmd/open_source.go`：Console backend 只读接口
- `web/console/src/views/SettingsOpenSourceView.js`：Settings 子页
- `web/console/src/router/index.js`：新增 `/settings/open-source` 路由，并修正 setup-free 规则

这里不建议：

- 把清单复制到 `web/console/src` 再读一遍
- 让前端直接读仓库文件
- 为这点静态数据引入新的存储或后台任务

## 9) 验收标准

首版完成后，至少满足下面这些条件：

- `mistermorph licenses` 可直接输出首版清单
- CLI 输出里每条记录都有项目名、链接、协议名、用途说明
- Console 可从 Settings 进入 `Open Source` 子页
- `/settings/open-source` 可直接访问和刷新
- Console 页面展示的记录集合、顺序、字段值与 CLI 一致
- Console 页面数据来自 `GET /api/settings/open-source`
- 新增或修改项目时，只需要改一份 `assets/open_source/projects.json`

## 10) Checklist

- [x] 扫描当前 Go 主要直接依赖
- [x] 扫描当前 Console 直接依赖
- [x] 扫描当前 shipped markdown renderer 依赖
- [x] 记录每个纳入项目的链接、协议、用途
- [x] 明确首版不做全量 SBOM
- [x] 明确 CLI 和 Console 共享同一份数据源
- [x] 明确 Console 走 Settings 子页面，不新开顶层导航
- [x] 明确 Console backend 只增加一个只读接口
