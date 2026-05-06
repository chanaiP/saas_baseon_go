import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { TenantBrandingPut } from '@/api/tenantBranding'
import { fetchPublicTenantFooter, fetchTenantBranding, putTenantBranding } from '@/api/tenantBranding'

/** 未登录且本地无缓存时使用 */
export const DEFAULT_LOGIN_BRAND_NAME = 'Ai DevOS'

/** 未配置租户 Logo 时不显示图片，由组件展示 CSS 绘制的默认神经元 logo */

export const useTenantBrandingStore = defineStore('tenantBranding', () => {
  const displayName = ref('管理台')
  const brandDisplayNameOverride = ref<string | null>(null)
  const logoData = ref('')
  const footerText = ref('')
  const canEditFooter = ref(false)

  const displayLogoSrc = computed(() => {
    const v = logoData.value.trim()
    return v || ''
  })

  const footerDisplay = computed(() => footerText.value.trim())
  const tenantName = ref('')
  const canEdit = ref(false)
  const loaded = ref(false)
  const loadError = ref('')

  function applyPayload(data: {
    display_name: string
    brand_display_name?: string | null
    logo_data: string | null
    footer_text?: string | null
    tenant_name: string
    can_edit?: boolean
    canEdit?: boolean
    can_edit_footer?: boolean
  }) {
    displayName.value = data.display_name || DEFAULT_LOGIN_BRAND_NAME
    const raw = data.brand_display_name
    brandDisplayNameOverride.value =
      raw != null && String(raw) ? String(raw) : null
    logoData.value = data.logo_data || ''
    footerText.value = data.footer_text != null ? String(data.footer_text) : ''
    tenantName.value = data.tenant_name || ''
    canEdit.value = data.can_edit === true || data.canEdit === true
    canEditFooter.value = data.can_edit_footer === true
  }

  async function load() {
    loadError.value = ''
    try {
      const data = await fetchTenantBranding()
      applyPayload(data)
      loaded.value = true
    } catch (e) {
      loadError.value = e instanceof Error ? e.message : '加载品牌失败'
      // 加载失败时使用默认值，不依赖本地缓存
      displayName.value = DEFAULT_LOGIN_BRAND_NAME
      logoData.value = ''
      footerText.value = ''
      loaded.value = true
    }
  }

  async function save(name: string, logo: string, footerForPlatformAdmin?: string) {
    const body: TenantBrandingPut = {
      brand_display_name: name,
      brand_logo_data: logo || null,
    }
    if (footerForPlatformAdmin !== undefined) {
      body.brand_footer_text = footerForPlatformAdmin.trim()
    }
    const data = await putTenantBranding(body)
    applyPayload(data)
  }

  async function loadPublicFooterIfEmpty() {
    if (footerText.value.trim()) return
    try {
      const { footer_text } = await fetchPublicTenantFooter()
      if (footer_text?.trim()) {
        footerText.value = footer_text.trim()
      }
    } catch {
      /* 离线或接口不可达 */
    }
  }

  function clearSession() {
    loaded.value = false
    loadError.value = ''
  }

  return {
    displayName,
    brandDisplayNameOverride,
    logoData,
    displayLogoSrc,
    footerText,
    footerDisplay,
    canEditFooter,
    tenantName,
    canEdit,
    loaded,
    loadError,
    load,
    save,
    loadPublicFooterIfEmpty,
    clearSession,
  }
})
