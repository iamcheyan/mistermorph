import { defineConfig } from 'vitepress'
import llmstxt, { copyOrDownloadAsMarkdownButtons } from 'vitepress-plugin-llms'

const enSidebar = [
  {
    text: 'Getting Started',
    items: [
      { text: 'Overview', link: '/guide/overview' },
      { text: 'Quickstart (CLI)', link: '/guide/quickstart-cli' },
      { text: 'Install and Configure', link: '/guide/install-and-config' }
    ]
  },
  {
    text: 'Developer (Embedding)',
    items: [
      { text: 'Build an Agent with Core', link: '/guide/build-agent-with-core' },
      { text: 'Advanced Core Embedding', link: '/guide/core-advanced-embedding' },
      { text: 'Agent-Level Customization', link: '/guide/agent-level-customization' }
    ]
  },
  {
    text: 'Runtime',
    items: [
      { text: 'Runtime Modes', link: '/guide/runtime-modes' },
      { text: 'Prompt Architecture (Top-Down)', link: '/guide/prompt-architecture' },
      { text: 'Memory', link: '/guide/memory' },
      { text: 'Skills', link: '/guide/skills' },
      { text: 'Built-in Tools', link: '/guide/built-in-tools' },
      { text: 'MCP', link: '/guide/mcp' }
    ]
  },
  {
    text: 'Operations',
    items: [
      { text: 'Security and Guard', link: '/guide/security-and-guard' },
      { text: 'Config Patterns', link: '/guide/config-patterns' },
      { text: 'Config Fields Reference', link: '/guide/config-reference' },
      { text: 'Environment Variables Reference', link: '/guide/env-vars-reference' },
      { text: 'Repository Docs Map', link: '/guide/docs-map' }
    ]
  }
]

const zhSidebar = [
  {
    text: '开始使用',
    items: [
      { text: '总览', link: '/zh/guide/overview' },
      { text: '快速开始（CLI）', link: '/zh/guide/quickstart-cli' },
      { text: '安装与配置', link: '/zh/guide/install-and-config' }
    ]
  },
  {
    text: '开发者（嵌入）',
    items: [
      { text: '用 Core 快速搭建 Agent', link: '/zh/guide/build-agent-with-core' },
      { text: 'Core 嵌入进阶', link: '/zh/guide/core-advanced-embedding' },
      { text: 'Agent 底层扩展', link: '/zh/guide/agent-level-customization' }
    ]
  },
  {
    text: '运行模式',
    items: [
      { text: 'Runtime 模式', link: '/zh/guide/runtime-modes' },
      { text: 'Prompt 组织（自顶向下）', link: '/zh/guide/prompt-architecture' },
      { text: 'Memory', link: '/zh/guide/memory' },
      { text: 'Skills', link: '/zh/guide/skills' },
      { text: '内置工具', link: '/zh/guide/built-in-tools' },
      { text: 'MCP', link: '/zh/guide/mcp' }
    ]
  },
  {
    text: '运维与治理',
    items: [
      { text: '安全与 Guard', link: '/zh/guide/security-and-guard' },
      { text: '配置模式', link: '/zh/guide/config-patterns' },
      { text: '配置字段总览', link: '/zh/guide/config-reference' },
      { text: '环境变量总览', link: '/zh/guide/env-vars-reference' },
      { text: '仓库文档地图', link: '/zh/guide/docs-map' }
    ]
  }
]

const jaSidebar = [
  {
    text: 'はじめに',
    items: [
      { text: '概要', link: '/ja/guide/overview' },
      { text: 'クイックスタート（CLI）', link: '/ja/guide/quickstart-cli' },
      { text: 'インストールと設定', link: '/ja/guide/install-and-config' }
    ]
  },
  {
    text: '開発者（組み込み）',
    items: [
      { text: 'Core で Agent を素早く構築', link: '/ja/guide/build-agent-with-core' },
      { text: 'Core 高度な組み込み', link: '/ja/guide/core-advanced-embedding' },
      { text: 'Agent レイヤ拡張', link: '/ja/guide/agent-level-customization' }
    ]
  },
  {
    text: 'ランタイム',
    items: [
      { text: 'Runtime モード', link: '/ja/guide/runtime-modes' },
      { text: 'Prompt 設計（トップダウン）', link: '/ja/guide/prompt-architecture' },
      { text: 'Memory', link: '/ja/guide/memory' },
      { text: 'Skills', link: '/ja/guide/skills' },
      { text: '組み込みツール', link: '/ja/guide/built-in-tools' },
      { text: 'MCP', link: '/ja/guide/mcp' }
    ]
  },
  {
    text: '運用',
    items: [
      { text: 'セキュリティと Guard', link: '/ja/guide/security-and-guard' },
      { text: '設定パターン', link: '/ja/guide/config-patterns' },
      { text: '設定フィールド一覧', link: '/ja/guide/config-reference' },
      { text: '環境変数一覧', link: '/ja/guide/env-vars-reference' },
      { text: 'リポジトリ文書マップ', link: '/ja/guide/docs-map' }
    ]
  }
]

export default defineConfig({
  title: 'Mister Morph',
  description: 'Multilingual docs for usage, runtime, operations, and core embedding.',
  cleanUrls: true,
  lastUpdated: true,
  appearance: false,
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico', sizes: 'any' }],
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '16x16', href: '/favicon-16x16.png' }],
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' }],
    ['link', { rel: 'mask-icon', href: '/safari-pinned-tab.svg', color: '#141414' }],
    ['link', { rel: 'manifest', href: '/site.webmanifest' }],
    ['meta', { name: 'theme-color', content: '#070707' }]
  ],
  markdown: {
    config(md) {
      md.use(copyOrDownloadAsMarkdownButtons)
    }
  },
  themeConfig: {
    logo: {
      src: '/mister-morph-logo.svg',
      alt: 'Mister Morph'
    },
    siteTitle: '',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/quailyquaily/mistermorph' }
    ],
    search: {
      provider: 'local'
    },
    footer: {
      copyright: '© 2026 ARCH inc.'
    },
    editLink: {
      pattern: 'https://github.com/quailyquaily/mistermorph/edit/master/web/vitepress/docs/:path',
      text: 'Edit this page on GitHub'
    },
    outline: {
      level: [2, 3],
      label: 'On this page'
    }
  },
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Overview', link: '/guide/overview' },
          { text: 'Developer', link: '/guide/build-agent-with-core' }
        ],
        sidebar: enSidebar,
        docFooter: { prev: 'Previous', next: 'Next' },
        outline: { level: [2, 3], label: 'On this page' }
      }
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '文档总览', link: '/zh/guide/overview' },
          { text: '开发者', link: '/zh/guide/build-agent-with-core' }
        ],
        sidebar: zhSidebar,
        docFooter: { prev: '上一页', next: '下一页' },
        editLink: {
          pattern: 'https://github.com/quailyquaily/mistermorph/edit/master/web/vitepress/docs/:path',
          text: '在 GitHub 编辑此页'
        },
        outline: { level: [2, 3], label: '页面目录' }
      }
    },
    ja: {
      label: '日本語',
      lang: 'ja-JP',
      link: '/ja/',
      themeConfig: {
        nav: [
          { text: 'ホーム', link: '/ja/' },
          { text: 'ドキュメント総覧', link: '/ja/guide/overview' },
          { text: '開発者', link: '/ja/guide/build-agent-with-core' }
        ],
        sidebar: jaSidebar,
        docFooter: { prev: '前へ', next: '次へ' },
        editLink: {
          pattern: 'https://github.com/quailyquaily/mistermorph/edit/master/web/vitepress/docs/:path',
          text: 'GitHub でこのページを編集'
        },
        outline: { level: [2, 3], label: '目次' }
      }
    }
  },
  vite: {
    plugins: [
      llmstxt({
        excludeUnnecessaryFiles: false,
        customTemplateVariables: {
          title: 'Mister Morph Docs',
          description: 'Structured multilingual docs for usage and developer embedding.'
        }
      })
    ]
  }
})
