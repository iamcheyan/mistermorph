<script setup lang="ts">
import { computed } from 'vue'
import { useData } from 'vitepress'

const { frontmatter } = useData()

const title = computed(() => (frontmatter.value.title as string) ?? '')
const description = computed(
  () => (frontmatter.value.description as string | undefined) ?? ''
)
</script>

<template>
  <header class="morph-doc-hero">
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
