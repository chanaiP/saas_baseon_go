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
