<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useData, useRoute, withBase } from 'vitepress'
import VPLocalSearchBox from 'vitepress/dist/client/theme-default/components/VPLocalSearchBox.vue'

type NavItem = {
  activeMatch?: string
  link?: string
  text: string
}

type SidebarItem = {
  items?: SidebarItem[]
  link?: string
  text: string
}

const { hash, localeIndex, page, site, theme } = useData()
const route = useRoute()

const langDetails = ref<HTMLDetailsElement | null>(null)
const menuCloseTimer = ref<number | null>(null)
const menuOpen = ref(false)
const menuVisible = ref(false)
const showSearch = ref(false)

const navItems = computed(() => ((theme.value.nav ?? []) as NavItem[]).filter((item) => item.link))
const sidebarGroups = computed(() => (theme.value.sidebar ?? []) as SidebarItem[])
const referenceGroup = computed(() => {
  return sidebarGroups.value.find((group) => {
    return (group.items ?? []).some((item) => {
      return item.link?.includes('integration-references') || item.link?.includes('config-reference')
    })
  })
})

const searchButtonLabel = computed(() => {
  const options = theme.value.search?.options ?? theme.value.algolia
  return (
    options?.locales?.[localeIndex.value]?.translations?.button?.buttonText ||
    options?.translations?.button?.buttonText ||
    'Search'
  )
})

const brandLink = 'https://mistermorph.com'

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
    closeMenu()
  }
)

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('pointerdown', onPointerDown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('pointerdown', onPointerDown)
  clearMenuCloseTimer()
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
  if (event.key === 'Escape') {
    closeMenu()
    if (langDetails.value) langDetails.value.open = false
    return
  }

  if (event.key.toLowerCase() === 'k' && (event.ctrlKey || event.metaKey)) {
    event.preventDefault()
    openSearch()
    return
  }

  if (event.key === '/' && !isEditing(event)) {
    event.preventDefault()
    openSearch()
  }
}

function onPointerDown(event: PointerEvent) {
  const target = event.target as HTMLElement | null
  if (!target || !menuOpen.value) return
  if (target.closest('.morph-nav-menu') || target.closest('.nav-menu-panel')) return

  closeMenu()
}

function openSearch() {
  closeMenu()
  showSearch.value = true
}

function clearMenuCloseTimer() {
  if (menuCloseTimer.value === null) return

  window.clearTimeout(menuCloseTimer.value)
  menuCloseTimer.value = null
}

function toggleMenu() {
  if (menuOpen.value) {
    closeMenu()
    return
  }

  openMenu()
}

function openMenu() {
  clearMenuCloseTimer()
  menuVisible.value = true
  window.requestAnimationFrame(() => {
    menuOpen.value = true
  })
}

function closeMenu() {
  clearMenuCloseTimer()
  menuOpen.value = false
  menuCloseTimer.value = window.setTimeout(() => {
    menuVisible.value = false
    menuCloseTimer.value = null
  }, 260)
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
        <a class="morph-nav-brand-stack" :href="brandLink" aria-label="Mister Morph">
          <span class="brand-mark" aria-hidden="true">
            <svg class="brand-logo" viewBox="0 0 24 24" role="presentation">
              <path class="logo-spy-hat" d="M3 11h18"></path>
              <path class="logo-spy-hat" d="M5 11v-4a3 3 0 0 1 3 -3h8a3 3 0 0 1 3 3v4"></path>
              <path class="logo-spy-eye" d="M7 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path>
              <path class="logo-spy-eye" d="M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path>
              <path class="logo-spy-bridge" d="M10 17h4"></path>
            </svg>
          </span>
          <span class="nav-brand-name">Mister Morph</span>
        </a>
      </div>

      <div class="morph-nav-actions">
        <button
          class="nav-primary-trigger nav-action-trigger morph-search-trigger"
          type="button"
          :aria-label="searchButtonLabel"
          :title="searchButtonLabel"
          @click="openSearch"
        >
          <svg class="morph-search-icon" viewBox="0 0 24 24" role="presentation" aria-hidden="true">
            <circle cx="11" cy="11" r="6.5"></circle>
            <path d="M16 16l4 4"></path>
          </svg>
        </button>

        <details ref="langDetails" class="nav-primary-item lang-menu" aria-label="Language">
          <summary class="nav-primary-trigger nav-action-trigger lang-menu-trigger" title="Language">
            <span class="lang-menu-kicker">LANG</span>
            <span class="lang-menu-current">{{ currentLocaleCode }}</span>
            <span class="nav-primary-caret" aria-hidden="true">
              <svg class="nav-primary-caret-icon" viewBox="0 0 16 16" role="presentation">
                <path d="M8 3.5v9"></path>
                <path d="M3.5 8h9"></path>
              </svg>
            </span>
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

        <div class="morph-nav-menu" :class="{ 'is-open': menuOpen }">
          <button
            class="nav-primary-trigger nav-action-trigger nav-menu-trigger"
            type="button"
            :aria-expanded="menuOpen"
            aria-controls="site-menu-panel"
            @click="toggleMenu"
          >
            <span class="nav-menu-kicker">MENU</span>
            <span class="nav-primary-caret" aria-hidden="true">
              <svg class="nav-primary-caret-icon" viewBox="0 0 16 16" role="presentation">
                <path d="M8 3.5v9"></path>
                <path d="M3.5 8h9"></path>
              </svg>
            </span>
          </button>
        </div>
      </div>

      <div
        id="site-menu-panel"
        class="nav-menu-panel"
        :class="{ 'is-open': menuOpen }"
        :hidden="!menuVisible"
        :aria-hidden="!menuOpen"
      >
        <div class="nav-menu-panel-inner">
          <nav class="nav-menu-groups" aria-label="Menu">
            <section class="nav-menu-group" aria-labelledby="nav-menu-docs">
              <p id="nav-menu-docs" class="nav-menu-panel-kicker">Navigation</p>
              <div class="nav-menu-links">
                <a
                  v-for="item in navItems"
                  :key="item.text"
                  class="nav-menu-link"
                  :class="{ 'is-active': isActive(item) }"
                  :href="normalizeHref(item.link)"
                  @click="closeMenu"
                >
                  <span class="nav-menu-link-label">{{ item.text }}</span>
                  <span class="nav-menu-link-description">Mister Morph docs</span>
                </a>
              </div>
            </section>

            <section v-if="referenceGroup" class="nav-menu-group">
              <p class="nav-menu-panel-kicker">References</p>
              <div class="nav-menu-links">
                <a
                  v-for="item in referenceGroup.items ?? []"
                  :key="item.text"
                  class="nav-menu-link"
                  :class="{ 'is-active': item.link ? isActive(item) : false }"
                  :href="normalizeHref(item.link)"
                  @click="closeMenu"
                >
                  <span class="nav-menu-link-label">{{ item.text }}</span>
                </a>
              </div>
            </section>

            <section class="nav-menu-group" aria-labelledby="nav-menu-resources">
              <p id="nav-menu-resources" class="nav-menu-panel-kicker">Resources</p>
              <div class="nav-menu-links">
                <a class="nav-menu-link" :href="brandLink" @click="closeMenu">
                  <span class="nav-menu-link-label">Website</span>
                  <span class="nav-menu-link-description">Mister Morph product site</span>
                </a>
                <a
                  class="nav-menu-link"
                  href="https://github.com/quailyquaily/mistermorph"
                  rel="noopener noreferrer"
                  target="_blank"
                  @click="closeMenu"
                >
                  <span class="nav-menu-link-label">GitHub</span>
                  <span class="nav-menu-link-description">Source, issues, and releases</span>
                </a>
              </div>
            </section>
          </nav>
        </div>
      </div>
    </div>

    <VPLocalSearchBox v-if="showSearch" @close="showSearch = false" />
  </header>
</template>
