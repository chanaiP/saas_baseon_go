WITH api(method, path, permission_code, audit) AS (
  VALUES
    ('POST', '/api/integration-center/provider-apps', 'integration_center:app_manage', TRUE),
    ('POST', '/api/integration-center/connectivity-check', 'integration_center:connection_manage', TRUE),
    ('PUT', '/api/integration-center/provider-apps/{code}', 'integration_center:app_manage', TRUE),
    ('PATCH', '/api/integration-center/provider-apps/{code}/credential', 'integration_center:app_manage', TRUE),
    ('PATCH', '/api/integration-center/app-capabilities/{id}', 'integration_center:app_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/refresh', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/pause', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/resume', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/tenant-connections/{id}/retry', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/sync-jobs/{id}/retry', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/sync-jobs/{id}/pause', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/sync-jobs/{id}/resume', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/quota-policies', 'integration_center:quota_manage', TRUE),
    ('PUT', '/api/integration-center/quota-policies/{code}', 'integration_center:quota_manage', TRUE),
    ('PATCH', '/api/integration-center/quota-policies/{code}/status', 'integration_center:quota_manage', TRUE),
    ('POST', '/api/integration-center/alerts/{id}/process', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/alerts/{id}/resolve', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/alerts/{id}/ignore', 'integration_center:connection_manage', TRUE),
    ('POST', '/api/integration-center/logs/export', 'integration_center:connection_manage', TRUE)
)
INSERT INTO sys_app_api (
  app_code, method, path, permission_code, public, audit, manifest_hash,
  managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', method, path, permission_code, FALSE, audit,
  'INTEGRATION_CENTER_ACTIONS', TRUE, 'ACTIVE', now(), now(), now()
FROM api
ON CONFLICT DO NOTHING;

WITH platform_tenant AS (
  SELECT id AS tenant_id FROM tenant WHERE is_platform_tenant = TRUE AND deleted_at IS NULL ORDER BY id LIMIT 1
),
integration_perms AS (
  SELECT id FROM permission
  WHERE app_code = 'integration-center' AND deleted_at IS NULL
),
platform_roles AS (
  SELECT role.id
  FROM role
  JOIN platform_tenant ON platform_tenant.tenant_id = role.tenant_id
  WHERE role.deleted_at IS NULL
)
INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT platform_roles.id, integration_perms.id, now()
FROM platform_roles
CROSS JOIN integration_perms
ON CONFLICT (role_id, permission_id) DO NOTHING;
