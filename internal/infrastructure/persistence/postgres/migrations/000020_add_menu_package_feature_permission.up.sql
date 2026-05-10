WITH platform_tenants AS (
  SELECT id
  FROM tenant
  WHERE deleted_at IS NULL
    AND (is_platform = true OR is_platform_tenant = true OR code = 'platform')
),
new_permission AS (
  INSERT INTO permission (
    tenant_id, name, path, perm_type, sort_order, enabled, visible,
    is_platform_only, is_package_feature, tenant_editable, feature_type,
    data_perm_mode, created_at, updated_at
  )
  SELECT
    pt.id, '菜单-套餐中心收录', 'menu:package_feature', 2, 0, true, true,
    true, false, false, 'OPERATION',
    'NONE', now(), now()
  FROM platform_tenants pt
  WHERE NOT EXISTS (
    SELECT 1
    FROM permission p
    WHERE p.tenant_id = pt.id
      AND p.path = 'menu:package_feature'
      AND p.deleted_at IS NULL
  )
  RETURNING id
)
UPDATE permission p
SET name = '菜单-套餐中心收录',
    perm_type = 2,
    enabled = true,
    visible = true,
    is_platform_only = true,
    is_package_feature = false,
    tenant_editable = false,
    feature_code = NULL,
    feature_type = 'OPERATION',
    data_perm_mode = 'NONE',
    updated_at = now()
FROM platform_tenants pt
WHERE p.tenant_id = pt.id
  AND p.path = 'menu:package_feature'
  AND p.deleted_at IS NULL;

INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT rp.role_id, package_perm.id, now()
FROM role_permission rp
JOIN permission edit_perm
  ON edit_perm.id = rp.permission_id
 AND edit_perm.path = 'menu:edit'
 AND edit_perm.deleted_at IS NULL
JOIN permission package_perm
  ON package_perm.tenant_id = edit_perm.tenant_id
 AND package_perm.path = 'menu:package_feature'
 AND package_perm.deleted_at IS NULL
WHERE NOT EXISTS (
  SELECT 1
  FROM role_permission existing
  WHERE existing.role_id = rp.role_id
    AND existing.permission_id = package_perm.id
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

UPDATE permission
SET is_package_feature = false,
    updated_at = now()
WHERE path = 'menu:package_feature'
  AND deleted_at IS NULL;
