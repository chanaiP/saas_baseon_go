import type { Profile } from '@/api/auth'

const MENU_FEATURE_BY_PATH: Record<string, string> = {
  '/organization': 'org_manage',
  '/positions': 'position_manage',
  '/users': 'user_manage',
  '/roles': 'role_manage',
  '/permissions': 'role_manage',
  '/menus': 'menu_manage',
  '/dict': 'dict_manage',
  '/params': 'param_manage',
  '/business-units': 'business_unit_manage',
  '/audit-logs': 'audit_log',
  '/login-logs': 'login_log',
  '/monitor/health': 'system_monitor',
  '/monitor/server': 'system_monitor',
  '/monitor/jobs': 'system_monitor',
  '/monitor/services': 'system_monitor',
  '/monitor/cache': 'system_monitor',
  '/monitor/cache-keys': 'system_monitor',
}

const ACTION_FEATURE_BY_PERMISSION: Record<string, string> = {
  'org:create': 'org_manage',
  'org:edit': 'org_manage',
  'org:delete': 'org_manage',
  'pos:create': 'position_manage',
  'pos:edit': 'position_manage',
  'pos:delete': 'position_manage',
  'user:create': 'user_manage',
  'user:edit': 'user_manage',
  'user:delete': 'user_manage',
  'user:reset_password': 'user_manage',
  'role:create': 'role_manage',
  'role:edit': 'role_manage',
  'role:delete': 'role_manage',
  'role:permission': 'role_manage',
  'menu:create': 'menu_manage',
  'menu:edit': 'menu_manage',
  'menu:package_feature': 'menu_manage',
  'dict_item:create': 'dict_manage',
  'dict_item:edit': 'dict_manage',
  'dict_item:delete': 'dict_manage',
  'param:create': 'param_manage',
  'param:edit': 'param_manage',
  'param:delete': 'param_manage',
  'business_unit:create': 'business_unit_manage',
  'business_unit:edit': 'business_unit_manage',
  'business_unit:delete': 'business_unit_manage',
  'brand:edit': 'brand_config',
}

const QUOTA_CODES_BY_FEATURE: Record<string, string[]> = {
  user_manage: ['max_users'],
  org_manage: ['max_companies', 'max_departments'],
  business_unit_manage: ['max_departments'],
  role_manage: ['max_roles'],
  import_data: ['daily_import_times'],
  export_data: ['daily_export_times'],
  file_manage: ['max_storage_gb', 'max_file_size_mb'],
}

export function featureForMenuPath(path: string): string | null {
  return MENU_FEATURE_BY_PATH[path] || null
}

export function featureForPermission(code: string): string | null {
  return ACTION_FEATURE_BY_PERMISSION[code] || null
}

/** profile 权限码已隐含某 MENU 级套餐能力时（后端已对订阅裁剪），侧栏应与之一致，避免 SKU 列表滞后导致整棵菜单被隐藏 */
function codesGrantMenuFeature(profile: Profile, featureCode: string): boolean {
  const codes = profile.permission_codes ?? []
  for (const [path, fc] of Object.entries(MENU_FEATURE_BY_PATH)) {
    if (fc === featureCode && codes.includes(path)) return true
  }
  for (const [op, fc] of Object.entries(ACTION_FEATURE_BY_PERMISSION)) {
    if (fc === featureCode && codes.includes(op)) return true
  }
  return false
}

/** 与后端 Permission / SaasFeature 对齐的按钮级套餐编码，如 org:delete → button_org_delete */
export function packageFeatureCodeForOperation(operationCode: string): string | null {
  if (operationCode === 'brand:edit') return 'brand_config'
  if (!operationCode || !operationCode.includes(':')) return null
  const normalized = operationCode.replace(/[^a-zA-Z0-9]+/g, '_').replace(/^_|_$/g, '').toLowerCase()
  return normalized ? `button_${normalized}` : null
}

/**
 * 操作是否在套餐内可用：必须命中对应的 ``button_*``（由后端 profile.features 下发）。
 * 菜单只决定是否可进入页面，不能隐式开启新增、编辑、删除等操作。
 */
function bypassPlanSkuGate(profile: Profile | null | undefined): boolean {
  return !!(profile?.is_platform_admin || profile?.tenant_is_platform)
}

export function hasPlanFeatureForOperation(profile: Profile | null | undefined, operationCode: string): boolean {
  if (bypassPlanSkuGate(profile)) return true
  const granular = packageFeatureCodeForOperation(operationCode)
  if (!granular) return true
  if (!profile?.features || !Array.isArray(profile.features)) return true
  const feats = profile.features as string[]
  if (feats.length === 0) return true
  if (!ACTION_FEATURE_BY_PERMISSION[operationCode]) return true
  return feats.includes(granular)
}

/** 与同路径菜单绑定的操作码任一命中即视为可进入该菜单（对齐后端 _require_menu_path_or_any_operation）。 */
export function hasMenuAccessViaBundledOperations(profile: Profile | null | undefined, path: string): boolean {
  const fc = featureForMenuPath(path)
  if (!fc || !profile?.permission_codes?.length) return false
  const codes = profile.permission_codes
  for (const [op, opFc] of Object.entries(ACTION_FEATURE_BY_PERMISSION)) {
    if (opFc === fc && codes.includes(op)) return true
  }
  return false
}

export function hasPlanFeature(profile: Profile | null | undefined, featureCode: string | null): boolean {
  if (!featureCode) return true
  if (bypassPlanSkuGate(profile)) return true
  if (!profile) return false
  if (!Array.isArray(profile.features)) return true
  const feats = profile.features as string[]
  if (feats.includes(featureCode)) return true
  if (codesGrantMenuFeature(profile, featureCode)) return true
  /** 能力列表未下发或为空：由 permission_codes + 路由守卫、接口依赖兜底 */
  if (feats.length === 0) return true
  return false
}

export function quotaCodesForFeatures(featureCodes: string[]): Set<string> {
  const out = new Set<string>()
  for (const featureCode of featureCodes) {
    for (const quotaCode of QUOTA_CODES_BY_FEATURE[featureCode] || []) {
      out.add(quotaCode)
    }
  }
  return out
}
