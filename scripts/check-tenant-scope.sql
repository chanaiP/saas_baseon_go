-- Scan menu/permission records whose tenant_scope is missing or outside the supported seven-value set.
-- Usage example:
--   psql "$DATABASE_URL" -f scripts/check-tenant-scope.sql

WITH allowed(scope) AS (
  VALUES
    ('platform_only'),
    ('enterprise_only'),
    ('personal_only'),
    ('all'),
    ('platform_enterprise'),
    ('enterprise_personal'),
    ('platform_personal')
)
SELECT 'permission' AS source, id::text AS id, path AS code, tenant_scope
FROM permission
WHERE deleted_at IS NULL
  AND (tenant_scope IS NULL OR tenant_scope = '' OR tenant_scope NOT IN (SELECT scope FROM allowed))
UNION ALL
SELECT 'sys_app_entry' AS source, id::text AS id, resource_code AS code, tenant_scope
FROM sys_app_entry
WHERE deleted_at IS NULL
  AND (tenant_scope IS NULL OR tenant_scope = '' OR tenant_scope NOT IN (SELECT scope FROM allowed))
UNION ALL
SELECT 'sys_app_permission' AS source, id::text AS id, permission_code AS code, tenant_scope
FROM sys_app_permission
WHERE deleted_at IS NULL
  AND (tenant_scope IS NULL OR tenant_scope = '' OR tenant_scope NOT IN (SELECT scope FROM allowed))
ORDER BY source, id;
