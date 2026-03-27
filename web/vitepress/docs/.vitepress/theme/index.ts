import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import CopyOrDownloadAsMarkdownButtons from 'vitepress-plugin-llms/vitepress-components/CopyOrDownloadAsMarkdownButtons.vue'
import 'quail-ui/dist/index.css'
import './custom.css'

const applyMorphTheme = () => {
  if (typeof document === 'undefined') {
    return
  }
  document.body.dataset.theme = 'morph'
  document.body.classList.remove('dark')
  document.body.classList.add('light')
}

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('CopyOrDownloadAsMarkdownButtons', CopyOrDownloadAsMarkdownButtons)
  },
  setup() {
    applyMorphTheme()
  }
} satisfies Theme
