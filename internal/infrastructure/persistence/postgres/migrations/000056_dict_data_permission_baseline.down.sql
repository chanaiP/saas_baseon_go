WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
),
dict_data_permissions AS (
  SELECT p.id
  FROM permission p
  JOIN platform_tenant pt ON pt.tenant_id = p.tenant_id
  WHERE p.path IN ('data:dict_type', 'data:dict_item')
    AND p.deleted_at IS NULL
)
DELETE FROM role_permission
WHERE permission_id IN (SELECT id FROM dict_data_permissions);

WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
)
DELETE FROM permission
WHERE tenant_id IN (SELECT tenant_id FROM platform_tenant)
  AND path IN ('data:dict_type', 'data:dict_item')
  AND deleted_at IS NULL;
