WITH split_permissions(name, path, is_platform_only, is_package_feature) AS (
  VALUES
    ('字典类型-新增', 'dict_type:create', true, false),
    ('字典类型-编辑', 'dict_type:edit', true, false),
    ('字典类型-删除', 'dict_type:delete', true, false),
    ('字典项-新增', 'dict_item:create', false, true),
    ('字典项-编辑', 'dict_item:edit', false, true),
    ('字典项-删除', 'dict_item:delete', false, true)
)
INSERT INTO permission (
  tenant_id, name, path, perm_type, sort_order, enabled, visible,
  is_platform_only, is_package_feature, tenant_editable, feature_type,
  data_perm_mode, created_at, updated_at
)
SELECT
  1, sp.name, sp.path, 2, 0, true, true,
  sp.is_platform_only, sp.is_package_feature, false, 'OPERATION',
  'ORG', now(), now()
FROM split_permissions sp
WHERE NOT EXISTS (
  SELECT 1 FROM permission p
  WHERE p.tenant_id = 1 AND p.path = sp.path AND p.deleted_at IS NULL
);

WITH split_permissions(name, path, is_platform_only, is_package_feature) AS (
  VALUES
    ('字典类型-新增', 'dict_type:create', true, false),
    ('字典类型-编辑', 'dict_type:edit', true, false),
    ('字典类型-删除', 'dict_type:delete', true, false),
    ('字典项-新增', 'dict_item:create', false, true),
    ('字典项-编辑', 'dict_item:edit', false, true),
    ('字典项-删除', 'dict_item:delete', false, true)
)
UPDATE permission p
SET name = sp.name,
    perm_type = 2,
    enabled = true,
    visible = true,
    is_platform_only = sp.is_platform_only,
    is_package_feature = sp.is_package_feature,
    feature_type = 'OPERATION',
    data_perm_mode = 'ORG',
    updated_at = now()
FROM split_permissions sp
WHERE p.tenant_id = 1
  AND p.path = sp.path
  AND p.deleted_at IS NULL;

UPDATE permission
SET is_package_feature = false,
    visible = false,
    updated_at = now()
WHERE tenant_id = 1
  AND path IN ('dict:create', 'dict:edit', 'dict:delete')
  AND deleted_at IS NULL;

INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM role r
JOIN permission p ON p.tenant_id = 1
WHERE p.path IN (
  'dict_type:create', 'dict_type:edit', 'dict_type:delete',
  'dict_item:create', 'dict_item:edit', 'dict_item:delete'
)
  AND p.deleted_at IS NULL
  AND r.deleted_at IS NULL
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO saas_feature (feature_code, feature_name, feature_type, parent_id, status, description, created_at, updated_at)
VALUES ('dict_manage', '数据字典', 'MENU', 0, 1, '字典类型、字典项与租户覆盖', now(), now())
ON CONFLICT (feature_code) DO UPDATE
SET feature_name = EXCLUDED.feature_name,
    feature_type = EXCLUDED.feature_type,
    status = 1,
    description = EXCLUDED.description,
    updated_at = now();

WITH dict_item_features(feature_code, feature_name, old_feature_code) AS (
  VALUES
    ('button_dict_item_create', '字典项-新增', 'button_dict_create'),
    ('button_dict_item_edit', '字典项-编辑', 'button_dict_edit'),
    ('button_dict_item_delete', '字典项-删除', 'button_dict_delete')
)
INSERT INTO saas_feature (feature_code, feature_name, feature_type, parent_id, status, description, created_at, updated_at)
SELECT
  f.feature_code,
  f.feature_name,
  'OPERATION',
  COALESCE((SELECT id FROM saas_feature WHERE feature_code = 'dict_manage' LIMIT 1), 0),
  1,
  '数据字典字典项操作能力',
  now(),
  now()
FROM dict_item_features f
ON CONFLICT (feature_code) DO UPDATE
SET feature_name = EXCLUDED.feature_name,
    feature_type = EXCLUDED.feature_type,
    parent_id = EXCLUDED.parent_id,
    status = 1,
    description = EXCLUDED.description,
    updated_at = now();

WITH feature_map(new_code, old_code) AS (
  VALUES
    ('button_dict_item_create', 'button_dict_create'),
    ('button_dict_item_edit', 'button_dict_edit'),
    ('button_dict_item_delete', 'button_dict_delete')
)
INSERT INTO saas_plan_feature (plan_id, feature_id, enabled, created_at, updated_at)
SELECT
  p.id,
  nf.id,
  COALESCE(opf.enabled, false),
  now(),
  now()
FROM saas_plan p
CROSS JOIN feature_map fm
JOIN saas_feature nf ON nf.feature_code = fm.new_code
LEFT JOIN saas_feature ofe ON ofe.feature_code = fm.old_code
LEFT JOIN saas_plan_feature opf ON opf.plan_id = p.id AND opf.feature_id = ofe.id
WHERE p.deleted_at IS NULL
ON CONFLICT (plan_id, feature_id) DO UPDATE
SET enabled = EXCLUDED.enabled,
    updated_at = now();

UPDATE saas_feature
SET status = 0,
    updated_at = now()
WHERE feature_code IN ('button_dict_create', 'button_dict_edit', 'button_dict_delete');
