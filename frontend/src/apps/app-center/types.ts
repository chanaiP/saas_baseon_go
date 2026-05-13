export interface AppCenterApp {
  id: number
  app_code: string
  app_name: string
  icon?: string | null
  app_type: string
  source: string
  status: string
  charge_mode: string
  visibility_scope: string
  owner?: string | null
  owner_user_ids?: string | null
  version?: string | null
  description?: string | null
  detail_description?: string | null
  deployment_mode: string
  communication_modes?: string | null
  visibility_mode?: string | null
  visible_tenants?: string | null
  open_method?: string | null
  trial_policy?: string | null
  trial_start_rule?: string | null
  asset_config?: string | null
  doc_config?: string | null
  release_channel?: string | null
  release_note?: string | null
  health_check_url?: string | null
  api_base_url?: string | null
  webhook_url?: string | null
  manifest_hash?: string | null
  manifest_version?: string | null
  last_manifest_synced_at?: string | null
  is_builtin: boolean
  is_platform_only: boolean
  sort_order: number
  created_at: string
  updated_at: string
  clients?: AppCenterClient[]
  assets?: AppCenterAssets
}

export interface AppCenterClient {
  id?: number
  app_id?: number
  client_code: string
  client_name: string
  enabled: boolean
  sort_order: number
  config_note?: string | null
}

export interface AppCenterAssets {
  entries: AppCenterEntry[]
  apis: AppCenterAPI[]
  permissions: AppCenterPermission[]
  package_features: AppCenterPackageFeature[]
  quotas: AppCenterQuota[]
  manifest_loads: AppCenterManifestLoad[]
}

export interface AppCenterEntry {
  resource_code: string
  name: string
  path: string
  parent_code?: string | null
  sort_order: number
  platform_only: boolean
  tenant_visible: boolean
  tenant_editable: boolean
  include_in_package: boolean
  feature_code?: string | null
  data_perm_mode: string
  status: string
  last_synced_at?: string | null
  protection_source?: string | null
  protection_reason?: string | null
  protected_at?: string | null
}

export interface AppCenterAPI {
  method: string
  path: string
  permission_code?: string | null
  public: boolean
  audit: boolean
  status: string
  last_synced_at?: string | null
  protection_source?: string | null
  protection_reason?: string | null
  protected_at?: string | null
}

export interface AppCenterPermission {
  permission_code: string
  name: string
  permission_type: string
  menu_code?: string | null
  platform_only: boolean
  include_in_package: boolean
  data_perm_mode: string
  status: string
  last_synced_at?: string | null
  protection_source?: string | null
  protection_reason?: string | null
  protected_at?: string | null
}

export interface AppCenterPackageFeature {
  feature_code: string
  feature_name: string
  feature_type: string
  parent_code?: string | null
  source_code?: string | null
  package_policy: string
  include_in_package: boolean
  status: string
  last_synced_at?: string | null
  protection_source?: string | null
  protection_reason?: string | null
  protected_at?: string | null
}

export interface AppCenterQuota {
  quota_code: string
  quota_name: string
  quota_type: string
  unit?: string | null
  period_type?: string | null
  include_in_package: boolean
  status: string
  last_synced_at?: string | null
  protection_source?: string | null
  protection_reason?: string | null
  protected_at?: string | null
}

export interface AppCenterManifestLoad {
  id: number
  action: string
  source_type: string
  source_name?: string | null
  manifest_version: string
  manifest_hash: string
  fragment_role: string
  status: string
  summary?: string | null
  error_summary?: string | null
  operator_user_id: number
  created_at: string
  updated_at: string
}

export interface AppCenterListResponse {
  items: AppCenterApp[]
  total: number
  skip: number
  limit: number
}

export interface AppCenterStats {
  total: number
  online: number
  beta: number
  developing: number
  builtin: number
  disabled: number
  categories: number
  client_apps: number
  tenant_openings: number
  trial_invites: number
  manifest_loads: number
  audit_logs: number
}

export interface AppManifestParseResult {
  file_name: string
  file_path?: string
  manifest_hash: string
  manifest_version: string
  fragment_role: string
  app_code: string
  app_name: string
  app_type: string
  source: string
  status: string
  deployment_mode: string
  communication_modes: string[]
  visibility_scope: string
  charge_policy: string
  billing_mode: string
  package_policy: string
  client_codes: string[]
  exists: boolean
  importable: boolean
  valid: boolean
  blockers: string[]
  warnings: string[]
  counts: {
    clients: number
    menus: number
    operations: number
    permissions: number
    apis: number
    package_features: number
    quotas: number
    documents: number
  }
}

export interface AppManifestScanGroup {
  app_code: string
  app_name: string
  mode: 'CREATE' | 'SYNC'
  loadable: boolean
  fragment_count: number
  main_count: number
  files: AppManifestParseResult[]
  merged: AppManifestParseResult
  blockers: string[]
  warnings: string[]
}

export interface AppManifestScanResult {
  items: AppManifestParseResult[]
  groups: AppManifestScanGroup[]
  total: number
  importable_count: number
  blocked_count: number
  scan_root: string
  elapsed_ms: number
}

export interface AppManifestDiffSummary {
  create: number
  update: number
  no_change: number
  disable: number
  conflict: number
}

export interface AppManifestDiffChange {
  resource_type: string
  resource_code: string
  name: string
  action: 'CREATE' | 'UPDATE' | 'NO_CHANGE' | 'DISABLE' | 'CONFLICT'
  severity: 'INFO' | 'WARN' | 'ERROR'
  message: string
}

export interface AppManifestDiffResult {
  parse: AppManifestParseResult
  mode: 'CREATE' | 'SYNC'
  loadable: boolean
  summary: AppManifestDiffSummary
  changes: AppManifestDiffChange[]
  blockers: string[]
  warnings: string[]
}

export interface AppManifestLoadResult {
  load_id: number
  app_code: string
  status: string
  summary: AppManifestDiffSummary
  diff: AppManifestDiffResult
}

export interface AppCenterListQuery {
  skip?: number
  limit?: number
  keyword?: string
  type?: string
  status?: string
  source?: string
}

export interface AppCenterCreatePayload {
  app_code: string
  app_name: string
  icon?: string | null
  app_type: string
  status: string
  charge_mode: string
  visibility_scope: string
  owner?: string | null
  owner_user_ids?: string | null
  version?: string | null
  description?: string | null
  detail_description?: string | null
  deployment_mode?: string
  communication_modes?: string | null
  visibility_mode?: string | null
  visible_tenants?: string | null
  open_method?: string | null
  trial_policy?: string | null
  trial_start_rule?: string | null
  asset_config?: string | null
  doc_config?: string | null
  release_channel?: string | null
  release_note?: string | null
  health_check_url?: string | null
  api_base_url?: string | null
  webhook_url?: string | null
  sort_order: number
  clients?: AppCenterClient[]
}

export interface AppCenterUpdatePayload {
  app_name: string
  icon?: string | null
  app_type: string
  charge_mode: string
  visibility_scope: string
  owner?: string | null
  owner_user_ids?: string | null
  version?: string | null
  description?: string | null
  detail_description?: string | null
  deployment_mode?: string
  communication_modes?: string | null
  visibility_mode?: string | null
  visible_tenants?: string | null
  open_method?: string | null
  trial_policy?: string | null
  trial_start_rule?: string | null
  asset_config?: string | null
  doc_config?: string | null
  release_channel?: string | null
  release_note?: string | null
  health_check_url?: string | null
  api_base_url?: string | null
  webhook_url?: string | null
  sort_order: number
  clients?: AppCenterClient[]
}
