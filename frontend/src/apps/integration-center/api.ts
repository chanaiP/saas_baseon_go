import http, { unwrap } from '@/api/http'

export interface IntegrationCenterApiEnvelope<T = unknown> {
  data: T
}

export interface IntegrationListParams {
  skip?: number
  limit?: number
  keyword?: string
  status?: string
  platform_code?: string
  provider_app_code?: string
  tenant_id?: string | number
  start_time?: string
  end_time?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export function fetchIntegrationOverview<T = unknown>() {
  return unwrap<T>(http.get('/api/integration-center/overview'))
}

export function checkIntegrationConnectivity<T = unknown>(target: string) {
  return unwrap<T>(http.post('/api/integration-center/connectivity-check', { target }))
}

export interface IntegrationGatewayInvokePayload {
  tenant_connection_id: string | number
  method: string
  path: string
  headers?: Record<string, string>
  body?: string
}

export function invokeIntegrationGateway<T = unknown>(payload: IntegrationGatewayInvokePayload) {
  return unwrap<T>(http.post('/api/integration-center/gateway/invoke', payload))
}

export function fetchIntegrationPlatforms<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/platforms', { params }))
}

export function fetchIntegrationPlatformDetail<T = unknown>(code: string) {
  return unwrap<T>(http.get(`/api/integration-center/platforms/${encodeURIComponent(code)}`))
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

export function rotateIntegrationProviderAppCredential<T = unknown>(code: string, credentialRef: string) {
  return unwrap<T>(http.patch(`/api/integration-center/provider-apps/${encodeURIComponent(code)}/credential`, { credential_ref: credentialRef }))
}

export function fetchIntegrationProviderAppDetail<T = unknown>(code: string) {
  return unwrap<T>(http.get(`/api/integration-center/provider-apps/${encodeURIComponent(code)}`))
}

export function updateIntegrationAppCapability<T = unknown>(id: string | number, payload: Record<string, unknown>) {
  return unwrap<T>(http.patch(`/api/integration-center/app-capabilities/${encodeURIComponent(String(id))}`, payload))
}

export function fetchIntegrationWorkspace<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/workspace', { params }))
}

export function fetchIntegrationPlatformCapabilities<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/platform-capabilities', { params }))
}

export interface IntegrationPlatformCapabilityPayload {
  platform_code?: string
  code?: string
  name: string
  capability_type?: string
  auth_scope_code?: string
  data_direction?: string
  status?: string
  description?: string
}

export function createIntegrationPlatformCapability<T = unknown>(payload: IntegrationPlatformCapabilityPayload) {
  return unwrap<T>(http.post('/api/integration-center/platform-capabilities', payload))
}

export function updateIntegrationPlatformCapability<T = unknown>(id: string | number, payload: IntegrationPlatformCapabilityPayload) {
  return unwrap<T>(http.put(`/api/integration-center/platform-capabilities/${encodeURIComponent(String(id))}`, payload))
}

export function disableIntegrationPlatformCapability<T = unknown>(id: string | number) {
  return unwrap<T>(http.patch(`/api/integration-center/platform-capabilities/${encodeURIComponent(String(id))}/disable`))
}

export function fetchIntegrationAppCapabilities<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/app-capabilities', { params }))
}

export function fetchIntegrationTenantConnections<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/tenant-connections', { params }))
}

export function fetchIntegrationMyConnections<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/my-connections', { params }))
}

export function fetchIntegrationTenantConnectionDetail<T = unknown>(id: string | number) {
  return unwrap<T>(http.get(`/api/integration-center/tenant-connections/${encodeURIComponent(String(id))}`))
}

export function fetchIntegrationMyConnectionDetail<T = unknown>(id: string | number) {
  return unwrap<T>(http.get(`/api/integration-center/my-connections/${encodeURIComponent(String(id))}`))
}

export interface IntegrationTenantConnectionPayload {
  tenant_id?: string | number
  provider_app_code: string
  connection_name: string
  auth_subject_type?: string
  auth_subject_id: string
  auth_subject_name: string
  auth_scope?: string[]
}

export function createIntegrationTenantConnection<T = unknown>(payload: IntegrationTenantConnectionPayload) {
  return unwrap<T>(http.post('/api/integration-center/tenant-connections', payload))
}

export function createIntegrationMyConnection<T = unknown>(payload: IntegrationTenantConnectionPayload) {
  return unwrap<T>(http.post('/api/integration-center/my-connections', payload))
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

export function fetchIntegrationSyncMonitor<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/sync-monitor', { params }))
}

export function fetchIntegrationMySyncJobs<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/my-sync-jobs', { params }))
}

export function fetchIntegrationSyncJobDetail<T = unknown>(id: string | number) {
  return unwrap<T>(http.get(`/api/integration-center/sync-jobs/${encodeURIComponent(String(id))}`))
}

export function fetchIntegrationMySyncJobDetail<T = unknown>(id: string | number) {
  return unwrap<T>(http.get(`/api/integration-center/my-sync-jobs/${encodeURIComponent(String(id))}`))
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

export function fetchIntegrationQuota<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/quota', { params }))
}

export function fetchIntegrationQuotaUsages<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/quota-usages', { params }))
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
  scope_type?: 'global' | 'tenant' | 'platform' | 'provider_app' | 'tenant_connection'
  tenant_id?: number
  platform_code?: string
  provider_app_code?: string
  tenant_connection_id?: number
  override_limit?: number
  priority?: number
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

export function fetchIntegrationAlerts<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/alerts', { params }))
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

export function fetchIntegrationLogs<T = unknown>(params?: IntegrationListParams) {
  return unwrap<T>(http.get('/api/integration-center/logs', { params }))
}

export function fetchIntegrationLogDetail<T = unknown>(id: string | number) {
  return unwrap<T>(http.get(`/api/integration-center/logs/${encodeURIComponent(String(id))}`))
}

export async function exportIntegrationLogs() {
  const response = await http.post('/api/integration-center/logs/export', undefined, { responseType: 'blob' })
  const disposition = response.headers?.['content-disposition'] || ''
  const matched = /filename="?([^"]+)"?/i.exec(disposition)
  return {
    blob: response.data as Blob,
    filename: matched?.[1] || `integration-call-logs-${Date.now()}.csv`,
    rowCount: Number(response.headers?.['x-integration-export-rows'] || 0),
  }
}
