const QUOTA_CODE_ORDER: Record<string, number> = {
  max_users: 10,
  max_companies: 20,
  max_departments: 30,
  max_stores: 40,
  max_business_units: 50,
  max_roles: 60,
  daily_api_calls: 70,
  max_storage_gb: 80,
  max_file_size_mb: 90,
  max_api_keys: 100,
  max_webhooks: 110,
  daily_import_times: 120,
  daily_export_times: 130,
  monthly_sms_count: 140,
  monthly_email_count: 150,
}

export interface QuotaDisplayOrderItem {
  quota_code?: string
  quota_id?: number
  id?: number | string
}

export function compareQuotaDisplayOrder(a: QuotaDisplayOrderItem, b: QuotaDisplayOrderItem) {
  const rankA = QUOTA_CODE_ORDER[a.quota_code || ''] ?? Number.MAX_SAFE_INTEGER
  const rankB = QUOTA_CODE_ORDER[b.quota_code || ''] ?? Number.MAX_SAFE_INTEGER
  if (rankA !== rankB) return rankA - rankB
  return numericFallback(a) - numericFallback(b)
}

export function sortQuotasByDisplayOrder<T extends QuotaDisplayOrderItem>(items: T[]): T[] {
  return [...items].sort(compareQuotaDisplayOrder)
}

function numericFallback(item: QuotaDisplayOrderItem) {
  if (typeof item.quota_id === 'number') return item.quota_id
  if (typeof item.id === 'number') return item.id
  return 0
}
