import http, { unwrap } from '@/api/http'

export interface IntegrationCenterApiEnvelope<T = unknown> {
  data: T
}

export function fetchIntegrationOverview<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/overview'))
}

export function fetchIntegrationPlatforms<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/platforms'))
}

export interface IntegrationPlatformPayload {
  name: string
  short_name?: string
  code?: string
  platform_type: string
  access_mode: string
  status: string
  tenant_visible: boolean
  owner_name?: string
  official_url?: string
  sort_order?: number
  description?: string
}

export function createIntegrationPlatform<T = unknown>(payload: IntegrationPlatformPayload) {
  return unwrap<T>(http.post('/api/integration-center/platforms', payload))
}

export function updateIntegrationPlatform<T = unknown>(code: string, payload: IntegrationPlatformPayload) {
  return unwrap<T>(http.put(`/api/integration-center/platforms/${encodeURIComponent(code)}`, payload))
}

export function fetchIntegrationWorkspace<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/workspace'))
}

export function fetchIntegrationTenantConnections<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/tenant-connections'))
}

export function fetchIntegrationSyncMonitor<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/sync-monitor'))
}

export function fetchIntegrationQuota<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/quota'))
}

export function fetchIntegrationAlerts<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/alerts'))
}

export function fetchIntegrationLogs<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/logs'))
}
