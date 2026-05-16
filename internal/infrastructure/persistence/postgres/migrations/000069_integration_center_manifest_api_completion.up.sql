-- Keep integration-center runtime app state and API permission matrix aligned with
-- internal/apps/integration_center/app.manifest.yaml for already-migrated databases.

UPDATE sys_app
SET
  status = 'ONLINE',
  updated_at = now()
WHERE app_code = 'integration-center'
  AND deleted_at IS NULL
  AND status IS DISTINCT FROM 'ONLINE';

WITH api(method, path, permission_code, audit) AS (
  VALUES
    ('GET', '/api/integration-center/overview', '/integration-center', FALSE),
    ('POST', '/api/integration-center/connectivity-check', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/gateway/invoke', 'integration_center:connection_manage', TRUE),
    ('GET', '/api/integration-center/platforms', '/integration-center/platforms', FALSE),
    ('GET', '/api/integration-center/platforms/{code}', '/integration-center/platforms', FALSE),
    ('POST', '/api/integration-center/platforms', 'integration_center:platform_manage', TRUE),
    ('PUT', '/api/integration-center/platforms/{code}', 'integration_center:platform_manage', TRUE),
    ('GET', '/api/integration-center/workspace', '/integration-center/workspace', FALSE),
    ('GET', '/api/integration-center/platform-capabilities', '/integration-center/workspace', FALSE),
    ('POST', '/api/integration-center/platform-capabilities', 'integration_center:platform_manage', TRUE),
    ('PUT', '/api/integration-center/platform-capabilities/{id}', 'integration_center:platform_manage', TRUE),
    ('PATCH', '/api/integration-center/platform-capabilities/{id}/disable', 'integration_center:platform_manage', TRUE),
    ('GET', '/api/integration-center/app-capabilities', '/integration-center/workspace', FALSE),
    ('POST', '/api/integration-center/provider-apps', 'integration_center:app_manage', TRUE),
    ('GET', '/api/integration-center/provider-apps/{code}', '/integration-center/workspace', FALSE),
    ('PUT', '/api/integration-center/provider-apps/{code}', 'integration_center:app_manage', TRUE),
    ('PATCH', '/api/integration-center/provider-apps/{code}/credential', 'integration_center:app_manage', TRUE),
    ('PATCH', '/api/integration-center/app-capabilities/{id}', 'integration_center:app_manage', TRUE),
    ('GET', '/api/integration-center/tenant-connections', '/integration-center/tenant-connections', FALSE),
    ('GET', '/api/integration-center/tenant-connections/{id}', '/integration-center/tenant-connections', FALSE),
    ('POST', '/api/integration-center/tenant-connections', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/oauth/start', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/refresh', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/pause', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/resume', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/retry', 'integration_center:connection_manage', TRUE),
    ('GET', '/api/integration-center/sync-monitor', '/integration-center/sync-monitor', FALSE),
    ('GET', '/api/integration-center/sync-jobs/{id}', '/integration-center/sync-monitor', FALSE),
    ('POST', '/api/integration-center/sync-jobs/{id}/retry', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/sync-jobs/{id}/pause', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/sync-jobs/{id}/resume', 'integration_center:connection_manage', TRUE),
    ('GET', '/api/integration-center/quota', '/integration-center/quota', FALSE),
    ('GET', '/api/integration-center/quota-usages', '/integration-center/quota', FALSE),
    ('POST', '/api/integration-center/quota-policies', 'integration_center:quota_manage', TRUE),
    ('PUT', '/api/integration-center/quota-policies/{code}', 'integration_center:quota_manage', TRUE),
    ('PATCH', '/api/integration-center/quota-policies/{code}/status', 'integration_center:quota_manage', TRUE),
    ('GET', '/api/integration-center/alerts', '/integration-center/alerts', FALSE),
    ('POST', '/api/integration-center/alerts/{id}/process', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/alerts/{id}/resolve', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/alerts/{id}/ignore', 'integration_center:connection_manage', TRUE),
    ('GET', '/api/integration-center/logs', '/integration-center/logs', FALSE),
    ('GET', '/api/integration-center/logs/{id}', '/integration-center/logs', FALSE),
    ('POST', '/api/integration-center/logs/export', 'integration_center:connection_manage', TRUE)
)
INSERT INTO sys_app_api (
  app_code, method, path, permission_code, public, audit, manifest_hash,
  managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', method, path, permission_code, FALSE, audit,
  'INTEGRATION_CENTER_API_COMPLETION', TRUE, 'ACTIVE', now(), now(), now()
FROM api
ON CONFLICT (app_code, method, path) WHERE deleted_at IS NULL
DO UPDATE SET
  permission_code = EXCLUDED.permission_code,
  public = EXCLUDED.public,
  audit = EXCLUDED.audit,
  managed_by_manifest = TRUE,
  status = 'ACTIVE',
  last_synced_at = now(),
  updated_at = now();
