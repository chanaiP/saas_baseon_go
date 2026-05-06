<script setup lang="ts">
import { computed } from 'vue'

import { useUiPreferencesStore } from '@/stores/uiPreferences'

const ui = useUiPreferencesStore()

function escapeXml(s: string) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

const bgImage = computed(() => {
  const raw = ui.watermarkText.trim()
  if (!raw) return 'none'
  const text = escapeXml(raw.slice(0, 48))
  const fill = ui.theme === 'dark' ? 'rgba(255,255,255,0.07)' : 'rgba(0,0,0,0.06)'
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="200" viewBox="0 0 320 200">
    <text x="160" y="100" dominant-baseline="middle" text-anchor="middle"
      fill="${fill}" font-size="15" font-family="system-ui,-apple-system,sans-serif"
      transform="rotate(-22 160 100)">${text}</text>
  </svg>`
  const encoded = encodeURIComponent(svg).replace(/'/g, '%27')
  return `url("data:image/svg+xml;charset=utf-8,${encoded}")`
})
</script>

<template>
  <div
    v-if="ui.watermarkText.trim()"
    class="watermark-layer"
    aria-hidden="true"
    :style="{ backgroundImage: bgImage }"
  />
</template>

<style scoped>
.watermark-layer {
  position: fixed;
  inset: 0;
  z-index: 1500;
  pointer-events: none;
  background-repeat: repeat;
  background-size: 320px 200px;
}
</style>
