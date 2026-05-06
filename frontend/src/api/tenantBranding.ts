import http, { unwrap } from './http'
import type { ApiResponse } from './types'

export interface TenantBranding {
  display_name: string
  /** 自定义展示名，null/空 表示使用主体名称 */
  brand_display_name: string | null
  logo_data: string | null
  /** 底部版权/说明，仅系统管理员可维护 */
  footer_text: string | null
  tenant_name: string
  can_edit: boolean
  can_edit_footer: boolean
}

export interface TenantBrandingPut {
  brand_display_name?: string | null
  brand_logo_data?: string | null
  brand_footer_text?: string | null
}

export async function fetchTenantBranding() {
  return unwrap(http.get<ApiResponse<TenantBranding>>('/api/tenant/branding'))
}

export async function putTenantBranding(body: TenantBrandingPut) {
  const payload: Record<string, string | null | undefined> = {}
  if (body.brand_display_name != null) {
    payload.brand_display_name = body.brand_display_name
  }
  if (body.brand_logo_data !== undefined) {
    payload.brand_logo_data = body.brand_logo_data
  }
  if (body.brand_footer_text !== undefined) {
    payload.brand_footer_text = body.brand_footer_text
  }
  return unwrap(http.put<ApiResponse<TenantBranding>>('/api/tenant/branding', payload))
}

/** 未登录可读：取最小 id 主体页脚（首次打开登录页补全缓存） */
export async function fetchPublicTenantFooter() {
  return unwrap(http.get<ApiResponse<{ footer_text: string | null }>>('/api/public/tenant-footer'))
}
