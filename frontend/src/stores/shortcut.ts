import { defineStore } from 'pinia'
import { ref } from 'vue'

import { fetchShortcuts, saveShortcuts } from '@/api/auth'

const LOCAL_KEY = 'neuro-shortcut-menu-ids'

function readLocal(): string[] {
  try {
    return JSON.parse(localStorage.getItem(LOCAL_KEY) || '[]')
  } catch {
    return []
  }
}

export const useShortcutStore = defineStore('shortcut', () => {
  const shortcutIds = ref<string[]>([])

  /** 初始化：优先从 profile（后端）加载，降级到 localStorage */
  async function init(fromProfile?: string[]) {
    if (fromProfile?.length) {
      shortcutIds.value = [...fromProfile]
      return
    }
    // 从后端加载
    try {
      const payload = await fetchShortcuts()
      shortcutIds.value = payload.shortcut_ids || []
    } catch {
      // 降级到 localStorage
      shortcutIds.value = readLocal()
    }
  }

  /** 持久化：同步到后端，失败则写入 localStorage */
  async function persist() {
    try {
      await saveShortcuts(shortcutIds.value)
    } catch {
      localStorage.setItem(LOCAL_KEY, JSON.stringify(shortcutIds.value))
    }
  }

  return { shortcutIds, init, persist }
})
