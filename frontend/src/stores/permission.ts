import { defineStore } from 'pinia'
import { ref } from 'vue'

import { fetchProfile } from '@/api/auth'
import type { Profile } from '@/api/auth'
import { featureForMenuPath, hasMenuAccessViaBundledOperations, hasPlanFeature, hasPlanFeatureForOperation } from '@/utils/planFeatures'

export const usePermissionStore = defineStore('permission', () => {
  const codes = ref<Set<string>>(new Set())
  const loaded = ref(false)
  const profile = ref<Profile | null>(null)
  const loadedAt = ref(0)

  async function load(options: { force?: boolean } = {}) {
    if (loaded.value && !options.force && !isStale()) return profile.value
    const p = await fetchProfile()
    profile.value = p
    codes.value = new Set(p.permission_codes || [])
    loaded.value = true
    loadedAt.value = Date.now()
    return p
  }

  /**
   * profile 体积较大且每次都触发后端多表查询；切换路由时 30s 过期太激进，
   * 实际权限/订阅变更很少。提到 5 分钟，必要时（登录、切租户、保存权限后）
   * 显式 force 即可。
   */
  function isStale(maxAgeMs = 5 * 60_000) {
    return !loaded.value || Date.now() - loadedAt.value > maxAgeMs
  }

  function clear() {
    codes.value = new Set()
    loaded.value = false
    profile.value = null
    loadedAt.value = 0
  }

  function can(code: string) {
    return codes.value.has(code)
  }

  function canFeature(featureCode: string) {
    return hasPlanFeature(profile.value, featureCode)
  }

  function canUseMenuPath(path: string) {
    const fc = featureForMenuPath(path)
    if (!hasPlanFeature(profile.value, fc)) return false
    return can(path) || hasMenuAccessViaBundledOperations(profile.value, path)
  }

  function canUseAction(code: string) {
    return can(code) && hasPlanFeatureForOperation(profile.value, code)
  }

  return { codes, loaded, profile, loadedAt, load, isStale, clear, can, canFeature, canUseMenuPath, canUseAction }
})
