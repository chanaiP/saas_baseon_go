INSERT INTO permission (
  tenant_id, name, path, perm_type, sort_order, enabled, visible,
  is_platform_only, is_package_feature, tenant_editable, feature_type,
  data_perm_mode, created_at, updated_at
)
SELECT
  1, '角色-权限设置', 'role:permission', 2, 0, true, true,
  false, true, false, 'OPERATION',
  'ORG', now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM permission
  WHERE tenant_id = 1 AND path = 'role:permission' AND deleted_at IS NULL
);

UPDATE permission
SET name = '角色-权限设置',
    perm_type = 2,
    enabled = true,
    visible = true,
    is_platform_only = false,
    is_package_feature = true,
    feature_type = 'OPERATION',
    data_perm_mode = 'ORG',
    updated_at = now()
WHERE tenant_id = 1
  AND path = 'role:permission'
  AND deleted_at IS NULL;

UPDATE permission p
SET name = v.name,
    updated_at = now()
FROM (
  VALUES
    ('role:create', '角色-新增'),
    ('role:edit', '角色-编辑'),
    ('role:delete', '角色-删除'),
    ('role:permission', '角色-权限设置')
) AS v(path, name)
WHERE p.tenant_id = 1
  AND p.path = v.path
  AND p.deleted_at IS NULL;

INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM role r
JOIN permission p ON p.tenant_id = 1 AND p.path = 'role:permission' AND p.deleted_at IS NULL
WHERE r.deleted_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM role_permission rp
    JOIN permission ep ON ep.id = rp.permission_id
    WHERE rp.role_id = r.id
      AND ep.path = 'role:edit'
      AND ep.deleted_at IS NULL
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO saas_feature (feature_code, feature_name, feature_type, parent_id, status, description, created_at, updated_at)
VALUES (
  'button_role_permission',
  '角色-权限设置',
  'OPERATION',
  COALESCE((SELECT id FROM saas_feature WHERE feature_code = 'role_manage' LIMIT 1), 0),
  1,
  '允许进入角色权限配置页面并保存菜单、操作与数据权限范围',
  now(),
  now()
)
ON CONFLICT (feature_code) DO UPDATE
SET feature_name = EXCLUDED.feature_name,
    feature_type = EXCLUDED.feature_type,
    parent_id = EXCLUDED.parent_id,
    status = 1,
    description = EXCLUDED.description,
    updated_at = now();

INSERT INTO saas_plan_feature (plan_id, feature_id, enabled, created_at, updated_at)
SELECT p.id, f.id, COALESCE(rmf.enabled, true), now(), now()
FROM saas_plan p
JOIN saas_feature f ON f.feature_code = 'button_role_permission'
LEFT JOIN saas_feature rf ON rf.feature_code = 'role_manage'
LEFT JOIN saas_plan_feature rmf ON rmf.plan_id = p.id AND rmf.feature_id = rf.id
WHERE p.deleted_at IS NULL
ON CONFLICT (plan_id, feature_id) DO UPDATE
SET enabled = EXCLUDED.enabled,
    updated_at = now();

UPDATE saas_feature f
SET feature_name = v.name,
    updated_at = now()
FROM (
  VALUES
    ('button_role_create', '角色-新增'),
    ('button_role_edit', '角色-编辑'),
    ('button_role_delete', '角色-删除'),
    ('button_role_permission', '角色-权限设置')
) AS v(feature_code, name)
WHERE f.feature_code = v.feature_code;
