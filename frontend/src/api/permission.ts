import http, { unwrap } from './http'
import type { ApiResponse } from './types'

export interface MenuBundleOp {
  id: number
  path: string
  name: string
  is_platform_only?: boolean
  is_package_feature?: boolean
  feature_code?: string | null
  feature_type?: string | null
  app_code?: string | null
  tenant_visible?: boolean
  tenant_editable?: boolean
  tenant_edit_scope?: string | null
}

export interface MenuBundle {
  path: string
  title: string
  menu_permission_id: number
  data_permission_id: number
  operations: MenuBundleOp[]
  is_platform_only?: boolean
  is_package_feature?: boolean
  show_in_admin?: boolean
  feature_code?: string | null
  feature_type?: string | null
  app_code?: string | null
  tenant_visible?: boolean
  tenant_editable?: boolean
  tenant_edit_scope?: string | null
  data_perm_mode?: 'NONE' | 'ORG' | 'BU' | 'ORG_BU'
}

export interface TenantMenuOverride {
  id: number
  tenant_id: number
  permission_id: number
  custom_name?: string | null
  custom_icon?: string | null
  enabled?: boolean | null
  visible?: boolean | null
  sort_order?: number | null
}

export interface TenantMenuOverrideValue {
  permission_id: number
  custom_name?: string | null
  custom_icon?: string | null
  enabled?: boolean | null
  visible?: boolean | null
  sort_order?: number | null
}

/** 按优先级依次尝试，兼容未重启的后端或旧路由顺序导致的 404/422 */
const MENU_BUNDLE_URLS = [
  '/api/roles/permission-menu-bundles',
  '/api/permission-menu-bundles',
  '/api/permissions/menu-bundles',
] as const

export async function fetchMenuBundles(): Promise<MenuBundle[]> {
  for (const url of MENU_BUNDLE_URLS) {
    const res = await http.get<ApiResponse<MenuBundle[]>>(url, {
      validateStatus: (status) =>
        (status >= 200 && status < 300) || status === 404 || status === 422,
    })
    if (res.status === 404 || res.status === 422) continue
    if (res.data.code !== 0) throw new Error(res.data.message || '请求失败')
    return res.data.data as MenuBundle[]
  }
  throw new Error(
    '无法加载菜单权限数据：请在后端目录使用当前代码重启 uvicorn（需包含菜单 bundle 接口）。',
  )
}

export async function fetchPermissionMenuBundles(): Promise<MenuBundle[]> {
  return unwrap(
    http.get<ApiResponse<MenuBundle[]>>('/api/permissions/menu-bundles'),
  )
}

export async function fetchTenantMenuOverrides() {
  return unwrap(
    http.get<ApiResponse<{ tenant_id: number; overrides: TenantMenuOverride[] }>>('/api/permissions/menu-overrides'),
  )
}

export async function saveTenantMenuOverrides(overrides: TenantMenuOverrideValue[]) {
  return unwrap(
    http.put<ApiResponse<{ tenant_id: number; overrides: TenantMenuOverride[] }>>(
      '/api/permissions/menu-overrides',
      { overrides },
    ),
  )
}

export async function updatePermission(
  permissionId: number,
  payload: {
    data_perm_mode?: 'NONE' | 'ORG' | 'BU' | 'ORG_BU'
    is_platform_only?: boolean
    show_in_admin?: boolean
  },
) {
  return unwrap(
    http.put<ApiResponse<unknown>>(`/api/permissions/menu-data-perm-mode/${permissionId}`, payload),
  )
}

export async function updatePermissionPackageFeature(
  permissionId: number,
  isPackageFeature: boolean,
) {
  return unwrap(
    http.put<ApiResponse<unknown>>(`/api/permissions/menu-package-feature/${permissionId}`, {
      is_package_feature: isPackageFeature,
    }),
  )
}
