# VitePress Theme Design Implementation Notes

## Scope

本文档记录 `mistermorph/web/vitepress/` 这一轮主题重实现的实现思路。

目标不是继续在 VitePress 默认主题上做零散视觉覆写，而是把
`mistermorph-website/docs/design_impl.md`
里已经确认的 Mister Morph 站点设计系统，压到 VitePress 这套 docs 运行时里。

当前 VitePress 端实际只有两类页面：

- 首页：
  - `docs/index.md`
  - `docs/zh/index.md`
  - `docs/ja/index.md`
- 内容页：
  - 各语言下的 `guide/*.md`

因此这次主题重实现的范围是有限的，但不应误判成“只是换颜色和字体”。

真正需要重做的是：

- 全局 header / utility / language / menu 的结构语言
- 首页骨架如何映射到 rail system
- 内容页如何成为 `rail-aligned prose`
- sidebar / outline / code block 如何在不破坏 VitePress 功能的前提下接入站点规范

## 1. Goal

这次 VitePress 主题应满足以下目标：

- 视觉上属于同一个 Mister Morph 站点系统，而不是“另一个 docs 子站”
- 首页和内容页共享同一套 canvas、rail、divider、heading、kicker 语法
- 保留 VitePress 已经提供的文档能力：
  - 本地搜索
  - 多语言
  - sidebar
  - outline
  - 上一页 / 下一页
  - edit link
  - Shiki 代码高亮
  - `vitepress-plugin-llms`
- 尽量不增加 markdown 作者负担：
  - 现有 `title`
  - 现有 `description`
  - 现有首页 section HTML
  - 应该尽可能继续可用

一句话概括：

- 保留 VitePress 的文档能力
- 替换 VitePress 的站点外壳

## 2. Mapping from Website Canonical System

VitePress 主题必须直接继承以下站点规范，而不是自行发明一套 docs-only 体系。

### 2.1 Global Canvas

VitePress 全局容器继续使用站点基线：

- `--site-max-width: 2256px`
- `--site-gutter: clamp(0.75rem, 4vw, 4.5rem)`

也就是说：

- 不再使用当前 VitePress 主题里的 `1440px` 小画布思路
- docs 不是“窄内容站”
- 即使 prose 列受控，站点级外壳仍然应该先建立大画布

### 2.2 Rail Tokens

VitePress 主题也应直接采用站点 rail token：

- `--rail-section-sidebar`
- `--rail-section-gap`
- `--rail-section-main`
- `--rail-section-support`
- `--rail-section-divider-left`
- `--rail-section-sticky-top`

说明：

- VitePress 不需要照搬首页 marketing section 的全部 HTML
- 但它必须共享 rail 的命名与对齐逻辑

### 2.3 Responsive Rules

VitePress 主题的响应式行为应继续遵循站点 rail 断点：

- `1280px+`
  - 默认三栏
- `860px–1279px`
  - 默认两栏
- `<860px`
  - 单栏

映射到 VitePress：

- 首页：
  - 宽屏可保留三栏 hero / helper 关系
  - 中等宽度下退成两栏
- 内容页：
  - 宽屏为 `sidebar rail / prose / support`
  - 中等宽度下为 `sidebar rail / prose`
  - mobile 用 local nav + 单栏 prose

### 2.4 Divider Language

VitePress 主题也要继续使用站点结构线语法：

- 顶部 section 横线：
  - 轻量
  - 渐隐
- 页面级长竖线：
  - 与 `--rail-section-divider-left` 对齐
- 标题级短竖线：
  - 与 rail 长竖线共享同一锚点

内容页这里尤其关键：

- `h1` 不再只是 VitePress 默认标题
- 它应像 install / deploy page heading 一样，被当成页面入口标记

### 2.5 Kicker Rule

VitePress 主题应引入与站点相同的 kicker tokenization 规则：

- `//`
  - `hero-install-kicker-sep`
- `[` / `]`
  - `hero-install-kicker-bracket`

并继续遵循内容规则：

- kicker 是结构标签
- 不做本地化翻译
- 不同语言页面复用同一条英文 kicker 文案

这条规则会直接影响 docs 首页和内容页的页首设计。

## 3. VitePress Page Types

VitePress 端本轮只定义两种 canonical page template。

### 3.1 Homepage

首页不是产品官网首页的直接复制，但应复用它的语法。

首页应理解为：

- rail-aligned header
- `rail-section rail-section-hero`
- `rail-section rail-section-start`
- 一个 docs overview rail section

不需要把官网的这些 section 生搬过来：

- `section-capability`
- `section-catalog`

原因：

- docs 首页的任务更直接
- 它不是产品营销概览
- 它只需要把用户送到正确的文档入口

所以 docs 首页更适合做成一个“更安静的入口页”：

- hero：
  - 说明这套文档是什么
  - 给出一个主 CTA
- start band：
  - 3 条最快路径
- overview section：
  - 3 组参考入口

### 3.2 Content Page

所有 `guide/*.md` 内容页统一视为一种页面：

- `rail-aligned prose`

这类页面应接近站点里的 install / deploy 页面，而不是 VitePress 默认 docs page。

也就是说：

- header 与正文左边界共用同一套 rail token
- `h1` 是页面入口，不是普通文档标题
- prose 列钉在 rail 的正文列上，而不是简单居中
- outline / helper 信息只在宽屏时进入第三栏

这也是这轮重实现里最重要的一条判断：

- 内容页不是“marketing page”
- 但也不是“默认 VitePress 文档页”
- 它是 `rail-aligned prose`

## 4. Homepage Plan

### 4.1 Homepage Header

首页 header 继续沿用站点 header 语法，而不是 VitePress 默认 navbar。

保留的站点原则：

- 左上只保留 logo mark 作为 home anchor
- utility 和 submenu 使用同一套 helper surface 语法
- `LANG` 与 `MENU` 继续是 mono utility

VitePress 端允许的 docs-specific 扩展：

- 增加 `SEARCH` 这一 docs-native utility

但规则是：

- `SEARCH` 不应回退成默认大号 pill search bar
- 它应被处理成与 `LANG`、`MENU` 同级的 utility trigger
- search dialog 的入口语法必须和站点 header 一致

### 4.2 Homepage Hero

首页 hero 应用官网 `rail-section-hero` 的思路，但语义更偏 docs entry：

- 第 1 栏：
  - kicker
  - `h1`
  - 一句 lead
- 第 2 栏：
  - 1 条主 CTA
  - 简短路径判断
- 第 3 栏：
  - 可选 supporting facts
  - 或保持留白

首页 hero 不需要恢复 VitePress 默认的：

- oversized hero image
- gradient image halo
- generic feature-card continuation

也不需要把 docs 首页做成“技术产品 landing page”。

### 4.3 Homepage Start Band

现有首页里的三条路径卡：

- Configure providers
- Understand runtime modes
- Embed the Go core

应重写成一个 rail-aligned start band，而不是三张均匀卡片。

目标读法：

- 左 rail：
  - 一个统一 heading
- 中 / 右两栏：
  - 三条起步路径

实现上可以保留 3 条链接，但不应继续使用：

- 同尺寸卡片
- 同层级卡面网格

更合适的方向是：

- 更像 entry band / route list
- 重点是“选路径”
- 不是“展示 feature”

### 4.4 Homepage Overview Section

当前首页的三组 reference：

- Use
- Tools
- Reference

应做成一个更接近 rail section 的 docs map。

读法：

- 左 rail：
  - section heading
- 中栏：
  - 三组主要入口
- 右栏：
  - 对这三组内容的 usage hint
  - 或最近更新 / 阅读建议

第一轮可以先简化：

- 先只做左 heading + 中栏三组入口
- 右栏允许留空

重点是不再使用：

- 平铺三列对称块
- 每列都长得一样的轻卡片

## 5. Content Page Plan

### 5.1 Global Structure

内容页宽屏结构应明确为：

- 第 1 栏：
  - sidebar navigation
- 第 2 栏：
  - page heading
  - lead
  - doc prose
- 第 3 栏：
  - outline
  - llms buttons
  - page helper surfaces

这里 sidebar 不是“另一个 panel”。

它在设计上就是 docs 的左 rail。

### 5.2 Page Heading Treatment

内容页页首应直接继承 install / deploy heading 语法：

- kicker 在上
- `h1` 使用 display serif
- lead 使用 `description`
- `h1::before` 短竖线与 rail divider 对齐

为了不增加 markdown 负担：

- `title` 继续来自 frontmatter
- `description` 继续作为 lead
- `page_kicker` 可选

默认 kicker 策略：

- guide 内容页统一 fallback 为英文 kicker
- 推荐默认值：
  - `[ DOCS // GUIDE ]`

如果后续需要区分路径，也可以按 section 派生：

- `guide/overview`:
  - `[ DOCS // OVERVIEW ]`
- `guide/config-reference`:
  - `[ DOCS // REFERENCE ]`

但第一轮无需把每页都手工补齐。

### 5.3 Prose Column

内容页 prose 列遵循 `rail-aligned prose` 原则：

- 宽度受控，但不自由居中
- 与 rail main 列对齐
- 上方 heading 与正文是一体，而不是两个互不相关的盒子

应避免继续使用当前做法：

- 给整块 `.VPDoc .content` 套一个大面板
- 再在上面补很多边框和背景

更合理的方向是：

- prose 自身承担阅读
- 结构线由 rail 和局部 rule 完成
- 少用“把整篇文章装进卡片”的方法

### 5.4 Sidebar

sidebar 的角色是 rail navigation，不是浮在页面上的半独立工具箱。

当前 config.ts 里的 sidebar 数据结构可以继续保留。

视觉规则应改成：

- level-0 group title：
  - mono
  - uppercase / label 感
- item：
  - sans
  - 更接近正文辅助导航
- 当前页高亮：
  - 使用站点 divider / underline / tone 语法
  - 不使用泛蓝胶囊感

### 5.5 Outline and Support Rail

右侧 outline 在 VitePress 里天然可以承担 support rail。

宽屏时：

- 第 3 栏显示 outline
- 同栏放 `CopyOrDownloadAsMarkdownButtons`
- 允许加入很轻的 page meta

中等宽度时：

- outline 不再占独立第三栏
- 可以折叠、下移，或直接隐藏

原则：

- helper 信息只在三栏时进入第三栏
- 不应在两栏和 mobile 上强行保留一个弱意义的侧栏

### 5.6 Code Block

VitePress 已经自带 Shiki，这与站点当前方向并不冲突。

第一轮不需要重写代码块渲染链路。

应保留：

- VitePress / Shiki 默认高亮
- copy button 功能
- `vitepress-plugin-llms` 的按钮组件

但视觉上应靠拢站点规则：

- inline code 与 block code 继续分开处理
- 语言标签和复制按钮属于外层工具条，而不是正文滚动内容的一部分
- 代码块不应继续像默认 VitePress 那样成为一个重量太高的黑盒

如果 VitePress 默认 DOM 足以实现：

- 优先通过 CSS 调整

只有在默认 DOM 不足时，才考虑二次包装。

## 6. Header Plan

VitePress 主题里的 header 应被视为“站点级共享外壳”，不是 docs 局部组件。

### 6.1 What to Preserve

应保留站点 header 当前已确认的规则：

- 左上只保留图标作为 home anchor
- 中间是一级结构，而不是品牌触发器
- utility 与 submenu 共用 `nav-primary-item / nav-helper-surface` 语法
- helper surface 是轻量 blur sheet + 渐隐边线

### 6.2 What Can Change for Docs

VitePress docs 不是官网产品导航，所以中间一级结构可以变成 docs 语义。

第一轮建议：

- `Guide`
- `Developer`

或保持 config 现有结构：

- `Docs Overview`
- `Developer`

但无论选择哪组文案，都要满足：

- header 信息架构直接、少量
- 不恢复 `Home` 这种冗余一级入口
- `LANG` 和 `MENU` 继续在右侧 utility

### 6.3 Search

搜索必须存在，因为它是 docs 基础能力。

但搜索入口应该接入当前 header 语言，而不是沿用默认 search bar。

推荐做法：

- 宽屏：
  - 用 `SEARCH` utility trigger 打开本地搜索
- 窄屏：
  - 仍可保留 icon-only fallback

## 7. Implementation Architecture

这次不建议继续把所有改动堆进一个 `custom.css`。

更合理的实现方式是：

- 继续以 `DefaultTheme` 为基础
- 但自定义 `Layout`
- 用少量 Vue 组件重建站点外壳

### 7.1 Keep

应保留 DefaultTheme 的这些能力：

- router / page resolution
- locale handling
- sidebar state
- outline state
- search integration
- markdown / code rendering
- doc footer / edit link

### 7.2 Replace

应替换或重组这些部分：

- navbar 外壳
- home hero 外壳
- doc page heading 外壳
- doc aside 组合方式
- page-level wrapper

### 7.3 Suggested File Shape

第一轮建议拆成以下结构：

- `docs/.vitepress/theme/index.ts`
- `docs/.vitepress/theme/Layout.vue`
- `docs/.vitepress/theme/components/MorphSiteHeader.vue`
- `docs/.vitepress/theme/components/MorphKicker.vue`
- `docs/.vitepress/theme/components/MorphHomeShell.vue`
- `docs/.vitepress/theme/components/MorphDocShell.vue`
- `docs/.vitepress/theme/styles/tokens.css`
- `docs/.vitepress/theme/styles/base.css`
- `docs/.vitepress/theme/styles/header.css`
- `docs/.vitepress/theme/styles/home.css`
- `docs/.vitepress/theme/styles/doc.css`

重点：

- 用组件决定结构
- 用分层 CSS 决定 token 与页面语法
- 不再靠一个超长 `custom.css` 同时处理所有状态

### 7.4 Authoring Constraints

为了避免把所有内容文件都改一遍，第一轮主题应遵循：

- 首页仍然允许继续使用 `layout: home`
- 现有首页 section HTML 可以保留，但类名和结构允许收敛
- 普通内容页只依赖：
  - `title`
  - `description`
- 额外 frontmatter 只作为可选增强：
  - `page_kicker`
  - `page_helper`

## 8. First-pass Decisions

第一轮实现建议刻意保守，不要一开始就扩 scope。

### 8.1 Do

- 先打通统一 canvas 和 rail token
- 先重建 header
- 先让首页进入正确骨架
- 先让内容页成为 `rail-aligned prose`
- 先统一 heading / kicker / divider / code block 外观

### 8.2 Do Not Yet

- 不先做复杂首页插画
- 不先做 docs 首页额外互动模块
- 不先重写 markdown renderer
- 不先重写 search 内核
- 不先引入一堆页面特例 frontmatter

## 9. Verification Baseline

这轮 VitePress 主题重实现至少应通过以下检查：

- `pnpm build` 通过
- homepage desktop：
  - header 与 hero 左边界对齐
  - hero 与后续 docs entry section 关系成立
- content page desktop：
  - sidebar / prose / outline 三栏关系成立
  - `h1` 的短竖线与 rail divider 对齐
- tablet：
  - 首页正确退化为两栏
  - 内容页 outline 正确退出第三栏
- mobile：
  - local nav 可用
  - prose 阅读无卡片化堆叠问题
- code block：
  - Shiki 正常
  - copy 按钮正常
  - inline code 不污染 block code
- i18n：
  - `LANG` 切换正常
  - kicker 不被本地化

## 10. Current Judgment

这件事的范围确实不大，因为页面类型只有两类。

但它也不只是“把官网 CSS 拷过来”。

真正的难点只有两个：

- 用 VitePress 默认能力承接站点级外壳，而不是和默认主题打补丁战
- 把内容页明确收成 `rail-aligned prose`

只要这两个判断锁定，后面的实现反而会比较直接。
