import http, { unwrap } from './http'
import type { ApiResponse } from './types'

export interface MonitorHealthDetail {
  mysql: boolean
  redis: boolean
}

export interface MonitorServerInfo {
  python_version: string
  pid: number
  cpu_percent?: number | null
  memory_mb?: number | null
  note?: string | null
}

export interface MonitorJobItem {
  id: string
  name: string
  schedule: string
  status: string
}

export interface MonitorJobsData {
  items: MonitorJobItem[]
  note?: string | null
}

export interface MonitorServicesOverview {
  mysql: boolean
  redis: boolean
  python_version: string
  pid: number
  cpu_percent?: number | null
  memory_mb?: number | null
  note?: string | null
}

export interface MonitorCacheStats {
  ok: boolean
  used_memory_human?: string | null
  keys: number
  connected_clients: number
  message?: string | null
}

export interface MonitorCacheKeyRow {
  key: string
  ttl: number
}

export interface MonitorCacheKeysData {
  items: MonitorCacheKeyRow[]
  cursor: number
}

export async function fetchMonitorHealth() {
  return unwrap(http.get<ApiResponse<MonitorHealthDetail>>('/api/monitor/health-detail'))
}

export async function fetchMonitorServer() {
  return unwrap(http.get<ApiResponse<MonitorServerInfo>>('/api/monitor/server-info'))
}

export async function fetchMonitorScheduledJobs() {
  return unwrap(http.get<ApiResponse<MonitorJobsData>>('/api/monitor/scheduled-jobs'))
}

export async function fetchMonitorServicesOverview() {
  return unwrap(http.get<ApiResponse<MonitorServicesOverview>>('/api/monitor/services-overview'))
}

export async function fetchMonitorCacheStats() {
  return unwrap(http.get<ApiResponse<MonitorCacheStats>>('/api/monitor/cache-stats'))
}

export async function fetchMonitorCacheKeys(cursor: number, limit: number, pattern: string) {
  return unwrap(
    http.get<ApiResponse<MonitorCacheKeysData>>('/api/monitor/cache-keys', {
      params: { cursor, limit, pattern },
    }),
  )
}
