export interface AiPage<T = unknown> {
  items: T[]
  total: number
  skip: number
  limit: number
  summary?: unknown
}

export interface AiOverview {
  metrics: Array<{ label: string; value: string; trend: string; tone: string }>
  tenant_metrics: Record<string, unknown>
  usage_trend: Array<Record<string, unknown>>
  model_cost_share: Array<Record<string, unknown>>
  tenant_ranking: Array<Record<string, unknown>>
  health_checks: Array<{ name: string; status: string; message: string }>
  core_base_routes: Array<Record<string, unknown>>
}

export type AiResource =
  | 'providers'
  | 'accounts'
  | 'apis'
  | 'capabilities'
  | 'models'
  | 'price-policies'
  | 'price-tiers'
  | 'base-routes'
  | 'route-models'
  | 'scenarios'
  | 'tenant-strategies'
  | 'quota-rules'
  | 'rate-limit-rules'
  | 'usage-records'
  | 'settings'

export interface AiColumn {
  key: string
  label: string
  width?: number
}

export interface AiSectionConfig {
  key: string
  title: string
  route: string
  resource?: AiResource
  description: string
  columns: AiColumn[]
  writable?: boolean
}

export interface AiProviderImportPayload {
  providers?: Array<Record<string, unknown>>
  accounts?: Array<Record<string, unknown>>
  apis?: Array<Record<string, unknown>>
}

export interface AiProviderImportResult {
  providers: number
  accounts: number
  apis: number
}

export interface AiModelImportPayload {
  models?: Array<Record<string, unknown>>
  price_policies?: Array<Record<string, unknown>>
  price_tiers?: Array<Record<string, unknown>>
}

export interface AiModelImportResult {
  models: number
  price_policies: number
  price_tiers: number
}

export interface AiScenarioImportPayload {
  scenarios?: Array<Record<string, unknown>>
}

export interface AiScenarioImportResult {
  scenarios: number
}

export interface AiRouteImportPayload {
  base_routes?: Array<Record<string, unknown>>
  route_models?: Array<Record<string, unknown>>
}

export interface AiRouteImportResult {
  base_routes: number
  route_models: number
}

export interface AiTenantStrategyImportPayload {
  policies?: Array<Record<string, unknown>>
  quota_rules?: Array<Record<string, unknown>>
  rate_limit_rules?: Array<Record<string, unknown>>
}

export interface AiTenantStrategyImportResult {
  policies: number
  quota_rules: number
  rate_limit_rules: number
}
