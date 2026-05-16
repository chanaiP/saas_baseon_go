-- Complete dictionary data permission baseline used by menu bundle assembly.

WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
),
data_permissions(name, path, sort_order) AS (
  VALUES
    ('字典类型-数据范围', 'data:dict_type', 0),
    ('字典项-数据范围', 'data:dict_item', 0)
)
INSERT INTO permission (
  tenant_id, name, path, perm_type, sort_order, enabled, visible, show_in_admin,
  is_platform_only, is_package_feature, tenant_editable, app_code, feature_type,
  data_perm_mode, created_at, updated_at
)
SELECT
  platform_tenant.tenant_id, data_permissions.name, data_permissions.path, 4,
  data_permissions.sort_order, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE,
  'system-management', 'DATA', 'ORG', now(), now()
FROM platform_tenant
CROSS JOIN data_permissions
WHERE NOT EXISTS (
  SELECT 1
  FROM permission p
  WHERE p.tenant_id = platform_tenant.tenant_id
    AND p.path = data_permissions.path
    AND p.deleted_at IS NULL
);

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
),
platform_roles AS (
  SELECT role.id
  FROM role
  JOIN platform_tenant ON platform_tenant.tenant_id = role.tenant_id
  WHERE role.deleted_at IS NULL
)
INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT platform_roles.id, dict_data_permissions.id, now()
FROM platform_roles
CROSS JOIN dict_data_permissions
WHERE NOT EXISTS (
  SELECT 1
  FROM role_permission rp
  WHERE rp.role_id = platform_roles.id
    AND rp.permission_id = dict_data_permissions.id
);
