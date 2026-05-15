import http, { unwrap } from '@/api/http'

export interface IntegrationCenterApiEnvelope<T = unknown> {
  data: T
}

export function fetchIntegrationOverview<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/overview'))
}

export function checkIntegrationConnectivity<T = unknown>(target: string) {
  return unwrap<T>(http.post('/api/integration-center/connectivity-check', { target }))
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

export interface IntegrationProviderAppPayload {
  platform_code: string
  code?: string
  name: string
  app_type?: string
  auth_mode?: string
  environment?: string
  status?: string
  tenant_visible?: boolean
  callback_url?: string
  webhook_url?: string
  credential_ref?: string
  owner_name?: string
  description?: string
}

export function createIntegrationProviderApp<T = unknown>(payload: IntegrationProviderAppPayload) {
  return unwrap<T>(http.post('/api/integration-center/provider-apps', payload))
}

export function updateIntegrationProviderApp<T = unknown>(code: string, payload: IntegrationProviderAppPayload) {
  return unwrap<T>(http.put(`/api/integration-center/provider-apps/${encodeURIComponent(code)}`, payload))
}

export function updateIntegrationAppCapability<T = unknown>(id: string | number, payload: Record<string, unknown>) {
  return unwrap<T>(http.patch(`/api/integration-center/app-capabilities/${encodeURIComponent(String(id))}`, payload))
}

export function fetchIntegrationWorkspace<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/workspace'))
}

export function fetchIntegrationTenantConnections<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/tenant-connections'))
}

export function refreshIntegrationTenantConnection<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/tenant-connections/${encodeURIComponent(String(id))}/refresh`))
}

export function pauseIntegrationTenantConnection<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/tenant-connections/${encodeURIComponent(String(id))}/pause`))
}

export function resumeIntegrationTenantConnection<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/tenant-connections/${encodeURIComponent(String(id))}/resume`))
}

export function retryIntegrationTenantConnection<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/tenant-connections/${encodeURIComponent(String(id))}/retry`))
}

export function fetchIntegrationSyncMonitor<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/sync-monitor'))
}

export function retryIntegrationSyncJob<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/sync-jobs/${encodeURIComponent(String(id))}/retry`))
}

export function pauseIntegrationSyncJob<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/sync-jobs/${encodeURIComponent(String(id))}/pause`))
}

export function resumeIntegrationSyncJob<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/sync-jobs/${encodeURIComponent(String(id))}/resume`))
}

export function fetchIntegrationQuota<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/quota'))
}

export interface IntegrationQuotaPolicyPayload {
  code?: string
  name: string
  quota_code?: string
  quota_unit?: string
  period_type?: string
  default_limit?: number
  over_limit_action?: string
  status?: string
  description?: string
}

export function createIntegrationQuotaPolicy<T = unknown>(payload: IntegrationQuotaPolicyPayload) {
  return unwrap<T>(http.post('/api/integration-center/quota-policies', payload))
}

export function updateIntegrationQuotaPolicy<T = unknown>(code: string, payload: IntegrationQuotaPolicyPayload) {
  return unwrap<T>(http.put(`/api/integration-center/quota-policies/${encodeURIComponent(code)}`, payload))
}

export function updateIntegrationQuotaPolicyStatus<T = unknown>(code: string, enabled: boolean) {
  return unwrap<T>(http.patch(`/api/integration-center/quota-policies/${encodeURIComponent(code)}/status`, { enabled }))
}

export function fetchIntegrationAlerts<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/alerts'))
}

export function processIntegrationAlert<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/alerts/${encodeURIComponent(String(id))}/process`))
}

export function resolveIntegrationAlert<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/alerts/${encodeURIComponent(String(id))}/resolve`))
}

export function ignoreIntegrationAlert<T = unknown>(id: string | number) {
  return unwrap<T>(http.post(`/api/integration-center/alerts/${encodeURIComponent(String(id))}/ignore`))
}

export function fetchIntegrationLogs<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/logs'))
}

export function exportIntegrationLogs<T = unknown>() {
  return unwrap<T>(http.post('/api/integration-center/logs/export'))
}
