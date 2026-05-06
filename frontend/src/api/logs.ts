import http, { unwrap } from './http'
import type { ApiResponse } from './types'

export interface LoginLogRow {
  id: number
  tenant_id?: number | null
  tenant_name?: string | null
  user_id?: number | null
  account: string
  success: boolean
  message?: string | null
  ip?: string | null
  created_at?: string | null
}

export interface AuditLogRow {
  id: number
  tenant_id?: number | null
  tenant_name?: string | null
  user_id?: number | null
  module: string
  action: string
  summary: string
  detail?: string | null
  ip?: string | null
  created_at?: string | null
}

export interface LogListResult<T> {
  items: T[]
  total: number
  skip: number
  limit: number
}

export async function fetchLoginLogs(
  skip = 0,
  limit = 20,
  opts?: {
    success?: boolean
    account?: string
    ip?: string
    tenant_name_hint?: string
    date_from?: string
    date_to?: string
  },
) {
  const params: Record<string, string | number | boolean> = { skip, limit }
  if (opts?.success === true) params.success = true
  if (opts?.success === false) params.success = false
  const acc = (opts?.account ?? '').trim()
  if (acc) params.account = acc
  const ip = (opts?.ip ?? '').trim()
  if (ip) params.ip = ip
  const tnh = (opts?.tenant_name_hint ?? '').trim()
  if (tnh) params.tenant_name_hint = tnh
  const df = (opts?.date_from ?? '').trim()
  const dt = (opts?.date_to ?? '').trim()
  if (df) params.date_from = df
  if (dt) params.date_to = dt
  return unwrap(
    http.get<ApiResponse<LogListResult<LoginLogRow>>>('/api/logs/login', {
      params,
    }),
  )
}

export async function fetchAuditLogs(
  skip = 0,
  limit = 20,
  opts?: {
    module?: string
    keyword?: string
    tenant_name_hint?: string
    account?: string
    ip?: string
    date_from?: string
    date_to?: string
  },
) {
  const params: Record<string, string | number> = { skip, limit }
  if (opts?.module) params.module = opts.module
  const kw = (opts?.keyword ?? '').trim()
  if (kw) params.keyword = kw
  const tnh = (opts?.tenant_name_hint ?? '').trim()
  if (tnh) params.tenant_name_hint = tnh
  const acc = (opts?.account ?? '').trim()
  if (acc) params.account = acc
  const ip = (opts?.ip ?? '').trim()
  if (ip) params.ip = ip
  const df = (opts?.date_from ?? '').trim()
  const dt = (opts?.date_to ?? '').trim()
  if (df) params.date_from = df
  if (dt) params.date_to = dt
  return unwrap(
    http.get<ApiResponse<LogListResult<AuditLogRow>>>('/api/logs/audit', {
      params,
    }),
  )
}
