WITH platform_tenants AS (
  SELECT id
  FROM tenant
  WHERE deleted_at IS NULL
    AND (is_platform = true OR is_platform_tenant = true OR code = 'platform')
)
UPDATE permission p
SET deleted_at = COALESCE(deleted_at, now()),
    updated_at = now()
FROM platform_tenants pt
WHERE p.tenant_id = pt.id
  AND p.path IN (
    '/apps/clients',
    '/apps/tenant-openings',
    '/apps/trial-invites',
    '/apps/manifests',
    '/apps/audit-logs'
  )
  AND p.deleted_at IS NULL;

UPDATE permission p
SET name = '应用中心',
    feature_code = 'app_center',
    parent_id = NULL,
    updated_at = now()
FROM platform_tenants pt
WHERE p.tenant_id = pt.id
  AND p.path = '/apps'
  AND p.deleted_at IS NULL;
