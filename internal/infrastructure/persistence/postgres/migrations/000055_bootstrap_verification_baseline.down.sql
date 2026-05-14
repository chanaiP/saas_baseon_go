WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
),
home_view_permission AS (
  SELECT p.id
  FROM permission p
  JOIN platform_tenant pt ON pt.tenant_id = p.tenant_id
  WHERE p.path = 'home:view' AND p.deleted_at IS NULL
)
DELETE FROM role_permission
WHERE permission_id IN (SELECT id FROM home_view_permission);

WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
)
DELETE FROM permission
WHERE tenant_id IN (SELECT tenant_id FROM platform_tenant)
  AND path = 'home:view'
  AND deleted_at IS NULL;

WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
)
DELETE FROM sys_param
WHERE tenant_id IN (SELECT tenant_id FROM platform_tenant)
  AND param_key IN (
    'security.password_min_length',
    'security.password_require_complexity',
    'audit.log_retention_days'
  );
