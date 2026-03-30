<script setup lang="ts">
import { computed } from 'vue'
import { useData, useRoute } from 'vitepress'
import MorphKicker from './MorphKicker.vue'

const { frontmatter, lang } = useData()
const route = useRoute()

const localeKey = computed(() => {
  if (lang.value.startsWith('zh')) return 'zh'
  if (lang.value.startsWith('ja')) return 'ja'
  return 'en'
})

const title = computed(() => (frontmatter.value.title as string) ?? '')
const description = computed(
  () => (frontmatter.value.description as string | undefined) ?? ''
)

const inferredKicker = computed(() => {
  const path = route.path

  if (path.includes('overview')) return '[ DOCS // OVERVIEW ]'
  if (
    path.includes('config-reference') ||
    path.includes('env-vars-reference') ||
    path.includes('docs-map')
  ) {
    return '[ DOCS // REFERENCE ]'
  }
  if (path.includes('quickstart') || path.includes('install-and-config')) {
    return '[ DOCS // START ]'
  }

  return '[ DOCS // GUIDE ]'
})

const kicker = computed(
  () => (frontmatter.value.page_kicker as string | undefined) ?? inferredKicker.value
)

</script>

<template>
  <header class="morph-doc-hero">
    <MorphKicker :text="kicker" />
    <div class="morph-doc-heading">
      <h1 class="morph-doc-title">
        {{ title }}
      </h1>
      <p v-if="description" class="morph-doc-lead">
        {{ description }}
      </p>
    </div>
  </header>
</template>
