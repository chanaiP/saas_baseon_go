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

  function setShortcutIds(ids: string[]) {
    shortcutIds.value = [...ids]
    localStorage.setItem(LOCAL_KEY, JSON.stringify(shortcutIds.value))
  }

  /** 初始化：优先从 profile（后端）加载，降级到 localStorage */
  async function init(fromProfile?: string[]) {
    if (fromProfile?.length) {
      setShortcutIds(fromProfile)
      return
    }
    // 从后端加载
    try {
      const payload = await fetchShortcuts()
      const serverIds = payload.shortcut_ids || []
      if (serverIds.length) {
        setShortcutIds(serverIds)
        return
      }
      const localIds = readLocal()
      setShortcutIds(localIds)
      if (localIds.length) {
        await saveShortcuts(localIds)
      }
    } catch {
      // 降级到 localStorage
      shortcutIds.value = readLocal()
    }
  }

  /** 持久化：先保留本地副本，再同步到后端；失败时抛出给页面提示 */
  async function persist() {
    localStorage.setItem(LOCAL_KEY, JSON.stringify(shortcutIds.value))
    const payload = await saveShortcuts(shortcutIds.value)
    setShortcutIds(payload.shortcut_ids || shortcutIds.value)
  }

  return { shortcutIds, init, persist }
})
