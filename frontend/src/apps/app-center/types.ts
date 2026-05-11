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
  version?: string | null
  description?: string | null
  is_builtin: boolean
  is_platform_only: boolean
  sort_order: number
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
  version?: string | null
  description?: string | null
  sort_order: number
}

export interface AppCenterUpdatePayload {
  app_name: string
  icon?: string | null
  app_type: string
  charge_mode: string
  visibility_scope: string
  owner?: string | null
  version?: string | null
  description?: string | null
  sort_order: number
}
