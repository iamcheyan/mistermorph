# VitePress Theme Design Implementation Notes

## Scope

本文档只记录 `mistermorph/web/vitepress/` 自己独有的主题规则。

凡是已经在
`mistermorph-website/docs/design_impl.md`
里定义过的 canonical 站点规范，这里都不重复。

一句话理解：

- `mistermorph-website/docs/design_impl.md`
  - 定义站点级 canonical system
- `mistermorph/web/vitepress/design_impl.md`
  - 只定义这套系统如何映射到 VitePress runtime

## 1. Theme Boundary

VitePress 端保留这些原生文档能力：

- 本地搜索
- 多语言
- sidebar
- outline
- 上一页 / 下一页
- edit link
- `vitepress-plugin-llms`

VitePress 主题的工作不是重新发明 docs product，而是：

- 保留 VitePress 的文档 runtime
- 用 Mister Morph 的站点系统重写它的外壳、布局和交互表面

## 2. Canonical Page Types

VitePress 端当前只维护两种页面模板。

### 2.1 Docs Homepage

适用页面：

- `docs/index.md`
- `docs/zh/index.md`
- `docs/ja/index.md`

首页不是官网 marketing 首页的复制版。

它是一个更克制的 docs entry page，职责只有三件事：

- 说明这套文档是什么
- 给出最快的进入路径
- 提供 overview 级入口分组

首页允许的骨架是：

- 自定义 header
- docs hero
- start routes
- docs overview

首页不再使用 VitePress 默认 hero image / feature cards 叙事。

### 2.2 Content Page

适用页面：

- 各语言下的 `guide/*.md`

所有内容页统一视为一种模板：

- `rail-aligned prose` in VitePress

内容页不使用默认 VitePress docs layout 观感，而是做以下映射：

- `VPSidebar`
  - 第一栏
- `VPDoc` prose
  - 第二栏
- `VPDoc aside`
  - 第三栏

## 3. VitePress-Specific Components

### 3.1 Header

VitePress 禁用默认 `VPNav`，统一使用自定义 `MorphSiteHeader`。

header 的 VitePress-specific 规则：

- docs search 入口放在 header 右侧
- search trigger 是图标按钮，不是文字按钮
- language switch 继续作为 utility menu
- 不引入默认 mobile menu 体系

### 3.2 Local Search

VitePress 继续使用 `VPLocalSearchBox`，但不能直接吃默认视觉。

原因：

- 站点主题把 `--vp-c-bg` 变成了透明语义
- 如果不单独覆写，local search modal 会失去自己的 shell surface

因此 local search 必须单独定义：

- backdrop
- shell background
- result card surface
- selected state
- search bar
- keyboard footer

### 3.3 Sidebar

`VPSidebar` 不是默认 docs tree，而是文件列表语法。

当前 sidebar 的独有规则：

- desktop 下贴合 rail 第一栏
- active item 高度 `44px`
- active item 使用 panel surface + 右侧 dot badge
- group title 和 item 文本左起点对齐
- mobile 下：
  - `background: var(--bg-1) !important`
  - `padding: 2rem 1rem`

### 3.4 Doc Hero

内容页在 `VPDoc` 之前插入 `MorphDocHero`。

它的职责是把 VitePress frontmatter 的：

- `title`
- `description`

转换成站点语法下的页首入口区。

当前 doc hero 规则：

- 不使用 kicker
- bottom divider 使用 dashed
- `morph-doc-title` 不再限制 `max-width`

### 3.5 Aside / Outline / Tools

宽屏下 aside 是第三栏，不是浮在正文里的小组件。

当前 VitePress-specific 约束：

- `aside-container` 不作为内部滚动容器
- `VPDocAsideOutline > .content` 的 `padding-left` 为 `0`
- page tools 顶部只保留 copy / download controls

## 4. VitePress-Specific Responsive Rules

这些规则是 VitePress runtime 自己需要补的，不属于 website canonical 文档本身：

- mobile `<960px`：
  - `VPContent` 使用 `padding: 2rem 1rem`
  - `VPDoc` 额外 top padding 归零
- desktop `>=960px`：
  - `VPSidebar` 锚定到第一栏
- wide desktop `>=1280px`：
  - `aside` 进入第三栏

## 5. VitePress-Specific Typographic / Interaction Overrides

以下是为了覆盖 VitePress 默认行为而存在的主题规则：

- `.vp-doc h2 .header-anchor`
  - `top: 0`
- code block language / copy button
  - 必须重新摆位，不能沿用默认 docs chrome
- 外链、inline code、Shiki block
  - 继续继承 website canonical 语法，但实现层面要通过 VitePress selector 覆盖完成

## 6. Implementation Surface

当前 VitePress 主题实现主要落在：

- `docs/.vitepress/theme/Layout.vue`
- `docs/.vitepress/theme/components/MorphSiteHeader.vue`
- `docs/.vitepress/theme/components/MorphHomeHero.vue`
- `docs/.vitepress/theme/components/MorphDocHero.vue`
- `docs/.vitepress/theme/components/MorphDocAsideTop.vue`
- `docs/.vitepress/theme/styles/tokens.css`
- `docs/.vitepress/theme/styles/base.css`
- `docs/.vitepress/theme/styles/header.css`
- `docs/.vitepress/theme/styles/home.css`
- `docs/.vitepress/theme/styles/doc.css`

## 7. Acceptance For VitePress Theme Work

当改动只发生在 `web/vitepress/` 时，最低验收应包括：

- `pnpm build`
- desktop 首页检查
- desktop 内容页检查
- mobile 内容页检查
- local search 打开态检查
- sidebar / aside / outline 没有被 VitePress 默认样式打回去
