<template>
  <span class="channel-logo" :class="logoClass" :title="displayName" aria-hidden="true">
    <img v-if="logo.src" :src="logo.src" :alt="displayName" />
    <svg v-else viewBox="0 0 24 24" role="img">
      <path fill="none" stroke="currentColor" stroke-width="1.8" d="M4.5 12a7.5 7.5 0 0 1 15 0a7.5 7.5 0 0 1-15 0Z" />
      <path fill="none" stroke="currentColor" stroke-width="1.8" d="M12 4.5c2.1 2 3.2 4.5 3.2 7.5s-1.1 5.5-3.2 7.5c-2.1-2-3.2-4.5-3.2-7.5S9.9 6.5 12 4.5Z" />
      <path fill="none" stroke="currentColor" stroke-width="1.8" d="M5 12h14" />
    </svg>
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  name: { type: String, default: '' },
  code: { type: String, default: '' },
})

const logoMap = {
  xiaohongshu: { key: 'xiaohongshu', color: '#ff2442', src: 'https://api.iconify.design/simple-icons:xiaohongshu.svg?color=%23ff2442' },
  douyin: { key: 'douyin', color: '#111827', src: 'https://api.iconify.design/simple-icons:tiktok.svg?color=%23111827' },
  wechat: { key: 'wechat', color: '#07c160', src: 'https://api.iconify.design/simple-icons:wechat.svg?color=%2307c160' },
  zhihu: { key: 'zhihu', color: '#0066ff', src: 'https://api.iconify.design/simple-icons:zhihu.svg?color=%230066ff' },
  weibo: { key: 'weibo', color: '#e6162d', src: 'https://api.iconify.design/simple-icons:sinaweibo.svg?color=%23e6162d' },
  baijiahao: { key: 'baijiahao', color: '#2938e6', src: 'https://api.iconify.design/simple-icons:baidu.svg?color=%232938e6' },
  website: { key: 'website', color: '#0f766e', src: '' },
}

const normalizedKey = computed(() => {
  const raw = `${props.code} ${props.name}`.toLowerCase()
  if (raw.includes('xiaohongshu') || raw.includes('小红书')) return 'xiaohongshu'
  if (raw.includes('douyin') || raw.includes('tiktok') || raw.includes('抖音')) return 'douyin'
  if (raw.includes('wechat') || raw.includes('weixin') || raw.includes('微信')) return 'wechat'
  if (raw.includes('zhihu') || raw.includes('知乎')) return 'zhihu'
  if (raw.includes('weibo') || raw.includes('微博')) return 'weibo'
  if (raw.includes('baijia') || raw.includes('baidu') || raw.includes('百家') || raw.includes('百度')) return 'baijiahao'
  return 'website'
})

const logo = computed(() => logoMap[normalizedKey.value] || logoMap.website)
const logoClass = computed(() => `channel-logo--${logo.value.key}`)
const displayName = computed(() => props.name || props.code || '渠道')
</script>
