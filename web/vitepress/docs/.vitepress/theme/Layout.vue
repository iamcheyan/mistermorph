<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import MorphDocAsideTop from './components/MorphDocAsideTop.vue'
import MorphDocHero from './components/MorphDocHero.vue'
import MorphHomeHero from './components/MorphHomeHero.vue'
import MorphSiteHeader from './components/MorphSiteHeader.vue'

const { Layout } = DefaultTheme
</script>

<template>
  <Layout>
    <template #layout-top>
      <MorphSiteHeader />
    </template>

    <template #home-hero-info>
      <MorphHomeHero />
    </template>

    <template #doc-before>
      <MorphDocHero />
    </template>

    <template #aside-top>
      <MorphDocAsideTop />
    </template>
  </Layout>
</template>
