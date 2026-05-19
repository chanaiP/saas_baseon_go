export interface PageResponse<T> {
  items: T[]
  total: number
  skip: number
  limit: number
}

export interface DataCenterQuery {
  skip?: number
  limit?: number
  time_range?: string
  start_date?: string
  end_date?: string
  keyword?: string
  brand_code?: string
  channel_code?: string
  platform_code?: string
  store_code?: string
  product_code?: string
  status?: string
  level?: string
  domain?: string
  data_type?: string
}

export interface DashboardSummary {
  health_score: number
  updated_at: string
  kpis: Array<{ code: string; label: string; value: number; unit: string; trend: string }>
}

export interface RawBatch {
  id: number
  batch_code: string
  data_type: string
  platform_code?: string
  connection_code?: string
  record_count: number
  success_count: number
  failed_count: number
  sync_time?: string
  status: string
  error_message?: string
}

export interface MetricDefinition {
  id: number
  metric_code: string
  metric_name: string
  metric_category: string
  formula?: string
  statistic_period?: string
  data_source?: string
  enabled: boolean
  anomaly_enabled: boolean
}

export interface AnomalyRule {
  id: number
  rule_code: string
  rule_name: string
  business_domain: string
  target_object_type: string
  enabled: boolean
  ai_enabled: boolean
  task_enabled: boolean
  priority: number
}

export interface AnomalyRecord {
  id: number
  anomaly_code: string
  title: string
  business_domain: string
  object_type: string
  object_code: string
  object_name?: string
  anomaly_level: string
  confidence_score: number
  impact_amount: number
  evidence_json: string
  occurred_at: string
  ai_status: string
  task_status: string
  review_status: string
  status: string
}

export interface RectificationTask {
  id: number
  task_code: string
  anomaly_code?: string
  title: string
  owner_role?: string
  priority: string
  deadline?: string
  target_desc?: string
  progress: number
  status: string
  review_status: string
}

export interface RectificationReview {
  id: number
  review_code: string
  task_code: string
  anomaly_code?: string
  improvement_result?: string
  review_conclusion: string
  reviewed_at?: string
}
