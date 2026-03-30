<script setup lang="ts">
import { computed } from 'vue'
import { useData, withBase } from 'vitepress'

type HeroAction = {
  link: string
  rel?: string
  target?: string
  text: string
  theme?: 'brand' | 'alt'
}

const { frontmatter, lang } = useData()

const hero = computed(
  () =>
    (frontmatter.value.hero ?? {}) as {
      actions?: HeroAction[]
      tagline?: string
      text?: string
    }
)

const localeKey = computed(() => {
  if (lang.value.startsWith('zh')) return 'zh'
  if (lang.value.startsWith('ja')) return 'ja'
  return 'en'
})

const copy = computed(() => {
  const key = localeKey.value
  if (key === 'zh') {
    return {
      helperCopy:
        '先走最短路径，再按运行模式、配置或嵌入层级继续展开。',
      helperTitle: 'Start'
    }
  }

  if (key === 'ja') {
    return {
      helperCopy:
        '最短ルートで入り、その後に runtime、config、embedding へ進む。',
      helperTitle: 'Start'
    }
  }

  return {
    helperCopy:
      'Take the shortest path first, then widen into runtime, config, or embedding.',
    helperTitle: 'Start'
  }
})

const primaryAction = computed(() => hero.value.actions?.[0] ?? null)

function normalizeHref(link: string) {
  if (/^(?:[a-z]+:)?\/\//i.test(link)) {
    return link
  }

  return withBase(link)
}
</script>

<template>
  <div class="morph-home-hero-stage">
    <div class="morph-home-hero-copy">
      <h1 class="morph-home-hero-title">
        {{ hero.text }}
      </h1>
      <p class="morph-home-hero-lead">
        {{ hero.tagline }}
      </p>
    </div>

    <div class="morph-home-hero-support">
      <div class="morph-home-hero-support-head">
        <p class="morph-home-helper-label">
          {{ copy.helperTitle }}
        </p>
        <p class="morph-home-hero-support-copy">
          {{ copy.helperCopy }}
        </p>
      </div>

      <div class="morph-home-hero-actions">
        <a
          v-if="primaryAction"
          class="install-cta-button morph-install-cta-button"
          :href="normalizeHref(primaryAction.link)"
          :target="primaryAction.target"
          :rel="primaryAction.rel"
        >
          {{ primaryAction.text }}
        </a>
      </div>
    </div>
  </div>
</template>
