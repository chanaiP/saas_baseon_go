import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useAvatarStore = defineStore('avatar', () => {
  // 头像URL状态
  const avatarUrl = ref<string>('')

  // 从本地存储加载头像
  const loadAvatarFromStorage = (userId: string) => {
    if (userId) {
      const storedAvatar = localStorage.getItem(`neuro_avatar_${userId}`)
      if (storedAvatar) {
        avatarUrl.value = storedAvatar
      }
    }
  }

  // 保存头像到本地存储
  const saveAvatarToStorage = (userId: string, url: string) => {
    if (userId) {
      localStorage.setItem(`neuro_avatar_${userId}`, url)
      avatarUrl.value = url
    }
  }

  // 更新头像URL
  const updateAvatarUrl = (url: string) => {
    avatarUrl.value = url
  }

  // 清除头像
  const clearAvatar = () => {
    avatarUrl.value = ''
  }

  return {
    avatarUrl,
    loadAvatarFromStorage,
    saveAvatarToStorage,
    updateAvatarUrl,
    clearAvatar
  }
})