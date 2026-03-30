<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    className?: string
    text?: string
    tag?: string
  }>(),
  {
    className: 'ui-kicker',
    text: '',
    tag: 'p'
  }
)

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

const html = computed(() => {
  const raw = props.text.trim()

  if (!raw) {
    return ''
  }

  return escapeHtml(raw)
    .replace(
      /\[/g,
      '<span class="ui-kicker-bracket ui-kicker-bracket-open">[</span>'
    )
    .replace(/\]/g, '<span class="ui-kicker-bracket ui-kicker-bracket-close">]</span>')
    .replace(/\/\//g, '<span class="ui-kicker-sep">//</span>')
})
</script>

<template>
  <component :is="tag" :class="className" v-html="html" />
</template>
