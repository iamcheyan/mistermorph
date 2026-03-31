import { defineConfig } from 'vitepress'
import llmstxt, { copyOrDownloadAsMarkdownButtons } from 'vitepress-plugin-llms'
import { createLocalesConfig } from './i18n'

const editLinkPattern = 'https://github.com/quailyquaily/mistermorph/edit/master/web/vitepress/docs/:path'

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
    }
  },
  locales: createLocalesConfig(editLinkPattern),
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
