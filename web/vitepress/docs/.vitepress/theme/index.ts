import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import CopyOrDownloadAsMarkdownButtons from 'vitepress-plugin-llms/vitepress-components/CopyOrDownloadAsMarkdownButtons.vue'
import Layout from './Layout.vue'
import MorphKicker from './components/MorphKicker.vue'
import './styles/tokens.css'
import './styles/base.css'
import './styles/header.css'
import './styles/home.css'
import './styles/doc.css'

const applyMorphTheme = () => {
  if (typeof document === 'undefined') {
    return
  }
  document.documentElement.dataset.theme = 'light'
  document.body.dataset.theme = 'light'
  document.body.classList.remove('dark')
  document.body.classList.add('light')
}

export default {
  extends: DefaultTheme,
  Layout,
  enhanceApp({ app }) {
    app.component('CopyOrDownloadAsMarkdownButtons', CopyOrDownloadAsMarkdownButtons)
    app.component('MorphKicker', MorphKicker)
  },
  setup() {
    applyMorphTheme()
  }
} satisfies Theme
