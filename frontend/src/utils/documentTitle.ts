import type { RouteLocationNormalizedLoaded } from 'vue-router'

import { DEFAULT_LOGIN_BRAND_NAME } from '@/stores/tenantBranding'

/** 浏览器标签标题：有页面 meta.title 时为「页面 - 平台名」，否则为平台名 */
export function buildDocumentTitle(
  to: RouteLocationNormalizedLoaded,
  platformName: string,
): string {
  const brand = platformName.trim() || DEFAULT_LOGIN_BRAND_NAME
  for (let i = to.matched.length - 1; i >= 0; i--) {
    const t = to.matched[i].meta.title
    if (typeof t === 'string' && t.trim()) {
      return `${t.trim()} - ${brand}`
    }
  }
  return brand
}
