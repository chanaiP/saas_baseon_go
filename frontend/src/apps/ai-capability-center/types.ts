export interface AiPage<T = unknown> {
  items: T[]
  total: number
  skip: number
  limit: number
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
