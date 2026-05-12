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
  is_builtin: boolean
  is_platform_only: boolean
  sort_order: number
  created_at: string
  updated_at: string
  clients?: AppCenterClient[]
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
  sort_order: number
  clients?: AppCenterClient[]
}
