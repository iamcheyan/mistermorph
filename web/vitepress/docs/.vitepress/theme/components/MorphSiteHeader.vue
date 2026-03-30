<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useData, useRoute, withBase } from 'vitepress'
import VPLocalSearchBox from 'vitepress/dist/client/theme-default/components/VPLocalSearchBox.vue'

type NavItem = {
  activeMatch?: string
  link?: string
  text: string
}

const { hash, localeIndex, page, site, theme } = useData()
const route = useRoute()

const langDetails = ref<HTMLDetailsElement | null>(null)
const showSearch = ref(false)

const navItems = computed(() => ((theme.value.nav ?? []) as NavItem[]).filter((item) => item.link))

const homeLink = computed(() => {
  const locale = site.value.locales[localeIndex.value]
  return locale?.link || '/'
})

const currentLocaleCode = computed(() => {
  return localeIndex.value === 'root' ? 'EN' : localeIndex.value.toUpperCase()
})

const localeLinks = computed(() => {
  const current = site.value.locales[localeIndex.value]
  const currentPrefix = current?.link || (localeIndex.value === 'root' ? '/' : `/${localeIndex.value}/`)
  const relativePath = page.value.relativePath.slice(currentPrefix.length - 1)

  return Object.entries(site.value.locales)
    .filter(([, value]) => value.label !== current?.label)
    .map(([key, value]) => {
      const baseLink = value.link || (key === 'root' ? '/' : `/${key}/`)
      const useCorresponding = theme.value.i18nRouting !== false
      const normalizedPath = useCorresponding
        ? ensureStartingSlash(
            relativePath
              .replace(/(^|\/)index\.md$/, '$1')
              .replace(/\.md$/, site.value.cleanUrls ? '' : '.html')
          )
        : ''

      return {
        code: key === 'root' ? 'EN' : key.toUpperCase(),
        current: false,
        href: baseLink.replace(/\/$/, '') + normalizedPath + hash.value
      }
    })
})

watch(
  () => route.path,
  () => {
    if (langDetails.value) langDetails.value.open = false
  }
)

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})

function ensureStartingSlash(value: string) {
  return value.startsWith('/') ? value : `/${value}`
}

function isEditing(event: KeyboardEvent) {
  const element = event.target as HTMLElement | null
  if (!element) return false

  const tagName = element.tagName
  return (
    element.isContentEditable ||
    tagName === 'INPUT' ||
    tagName === 'SELECT' ||
    tagName === 'TEXTAREA'
  )
}

function onKeydown(event: KeyboardEvent) {
  if (event.key.toLowerCase() === 'k' && (event.ctrlKey || event.metaKey)) {
    event.preventDefault()
    showSearch.value = true
    return
  }

  if (event.key === '/' && !isEditing(event)) {
    event.preventDefault()
    showSearch.value = true
  }
}

function normalizeHref(link = '') {
  if (/^(?:[a-z]+:)?\/\//i.test(link)) {
    return link
  }

  return withBase(link)
}

function isActive(item: NavItem) {
  const itemLink = item.link || ''
  const activeMatch = item.activeMatch || itemLink
  const current = normalizePath(route.path)
  const target = normalizePath(normalizeHref(activeMatch))

  return current === target || current.startsWith(`${target}/`)
}

function normalizePath(value: string) {
  return value
    .replace(/\/index$/, '/')
    .replace(/\.html$/, '')
    .replace(/\/$/, '') || '/'
}
</script>

<template>
  <header class="morph-site-header">
    <div class="container morph-nav-shell">
      <div class="morph-nav-brand-shell">
        <a class="morph-nav-brand-stack" :href="homeLink" aria-label="Mister Morph Docs">
          <span class="brand-mark" aria-hidden="true">
            <svg class="brand-logo" viewBox="0 0 24 24" role="presentation">
              <path class="logo-spy-hat" d="M3 11h18"></path>
              <path class="logo-spy-hat" d="M5 11v-4a3 3 0 0 1 3 -3h8a3 3 0 0 1 3 3v4"></path>
              <path class="logo-spy-eye" d="M7 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path>
              <path class="logo-spy-eye" d="M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path>
              <path class="logo-spy-bridge" d="M10 17h4"></path>
            </svg>
          </span>
        </a>
      </div>

      <nav class="morph-nav-primary" aria-label="Primary">
        <a
          v-for="item in navItems"
          :key="item.text"
          class="nav-primary-link"
          :class="{ 'is-active': isActive(item) }"
          :href="normalizeHref(item.link)"
        >
          {{ item.text }}
        </a>
      </nav>

      <div class="morph-nav-actions">
        <button class="nav-primary-trigger morph-search-trigger" type="button" @click="showSearch = true">
          <span class="nav-menu-kicker">SEARCH</span>
        </button>

        <details ref="langDetails" class="nav-primary-item lang-menu" aria-label="Language">
          <summary class="nav-primary-trigger lang-menu-trigger" title="Language">
            <span class="lang-menu-kicker">LANG</span>
            <span class="lang-menu-current">{{ currentLocaleCode }}</span>
          </summary>

          <div class="nav-primary-submenu nav-helper-surface nav-helper-surface-end">
            <a
              class="lang-menu-item nav-primary-submenu-link nav-helper-item nav-helper-item-end is-active"
              :href="route.path"
            >
              {{ currentLocaleCode }}
            </a>
            <a
              v-for="locale in localeLinks"
              :key="locale.code"
              class="lang-menu-item nav-primary-submenu-link nav-helper-item nav-helper-item-end"
              :href="locale.href"
            >
              {{ locale.code }}
            </a>
          </div>
        </details>
      </div>
    </div>

    <VPLocalSearchBox v-if="showSearch" @close="showSearch = false" />
  </header>
</template>
