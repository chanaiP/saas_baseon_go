WITH business_unit_menu AS (
  SELECT p.tenant_id, p.id AS parent_id
  FROM permission p
  WHERE p.path = '/business-units'
    AND p.perm_type = 3
    AND p.deleted_at IS NULL
),
seed(name, path, perm_type, sort_order, feature_code) AS (
  VALUES
    ('绑定资源', 'business_unit:resource_bind', 2, 40, 'button_business_unit_resource_bind'),
    ('业务资源', '/business-resources', 3, 45, 'business_resource_manage'),
    ('业务资源-新增', 'business_resource:create', 2, 50, 'button_business_resource_create'),
    ('业务资源-编辑', 'business_resource:edit', 2, 60, 'button_business_resource_edit'),
    ('业务资源-删除', 'business_resource:delete', 2, 70, 'button_business_resource_delete'),
    ('资源关系管理', 'business_resource:relation_manage', 2, 80, 'button_business_resource_relation_manage')
)
INSERT INTO permission (
  tenant_id, parent_id, name, path, perm_type, sort_order, enabled, visible, show_in_admin,
  is_platform_only, is_package_feature, tenant_editable, app_code, feature_code, feature_type,
  data_perm_mode, created_at, updated_at, deleted_at
)
SELECT
  bum.tenant_id,
  bum.parent_id,
  seed.name,
  seed.path,
  seed.perm_type,
  seed.sort_order,
  true,
  seed.path <> '/business-resources',
  seed.path <> '/business-resources',
  false,
  true,
  false,
  'system-management',
  seed.feature_code,
  CASE WHEN seed.perm_type = 3 THEN 'MENU' ELSE 'BUTTON' END,
  'ORG',
  now(),
  now(),
  NULL
FROM business_unit_menu bum
CROSS JOIN seed
WHERE NOT EXISTS (
  SELECT 1
  FROM permission p
  WHERE p.tenant_id = bum.tenant_id
    AND p.path = seed.path
    AND p.perm_type = seed.perm_type
    AND p.deleted_at IS NULL
);

UPDATE permission
SET visible = false,
    show_in_admin = false,
    updated_at = now()
WHERE path = '/business-resources'
  AND perm_type = 3
  AND deleted_at IS NULL;

WITH feature_seed(feature_code, feature_name, feature_type, parent_code, description) AS (
  VALUES
    ('business_resource_manage', '业务资源', 'MENU', 'business_unit_manage', '业务资源主数据维护'),
    ('button_business_unit_resource_bind', '业务单元-绑定资源', 'OPERATION', 'business_unit_manage', '业务单元资源绑定能力'),
    ('button_business_resource_create', '业务资源-新增', 'OPERATION', 'business_resource_manage', '业务资源新增能力'),
    ('button_business_resource_edit', '业务资源-编辑', 'OPERATION', 'business_resource_manage', '业务资源编辑能力'),
    ('button_business_resource_delete', '业务资源-删除', 'OPERATION', 'business_resource_manage', '业务资源归档能力'),
    ('button_business_resource_relation_manage', '业务资源-关系管理', 'OPERATION', 'business_resource_manage', '业务资源关系维护能力')
),
feature_rows AS (
  SELECT
    fs.feature_code,
    fs.feature_name,
    fs.feature_type,
    COALESCE(parent.id, 0) AS parent_id,
    fs.description,
    p.id AS menu_id
  FROM feature_seed fs
  LEFT JOIN saas_feature parent ON parent.feature_code = fs.parent_code
  LEFT JOIN permission p ON p.feature_code = fs.feature_code
    AND p.deleted_at IS NULL
  WHERE NOT EXISTS (
    SELECT 1
    FROM saas_feature removed
    WHERE removed.feature_code = fs.feature_code
      AND removed.status = 0
  )
)
INSERT INTO saas_feature (
  feature_code, feature_name, feature_type, parent_id, menu_id, status, description, app_code, created_at, updated_at
)
SELECT
  feature_code,
  feature_name,
  feature_type,
  parent_id,
  menu_id,
  1,
  description,
  'system-management',
  now(),
  now()
FROM feature_rows
ON CONFLICT (feature_code) DO UPDATE
SET
  feature_name = EXCLUDED.feature_name,
  feature_type = EXCLUDED.feature_type,
  parent_id = EXCLUDED.parent_id,
  menu_id = EXCLUDED.menu_id,
  description = EXCLUDED.description,
  app_code = EXCLUDED.app_code,
  updated_at = now()
WHERE saas_feature.status <> 0;

INSERT INTO saas_plan_feature (plan_id, feature_id, enabled, created_at, updated_at)
SELECT plan.id, feature.id, false, now(), now()
FROM saas_plan plan
JOIN saas_feature feature ON feature.feature_code IN (
  'business_resource_manage',
  'button_business_unit_resource_bind',
  'button_business_resource_create',
  'button_business_resource_edit',
  'button_business_resource_delete',
  'button_business_resource_relation_manage'
)
WHERE feature.status = 1
  AND NOT EXISTS (
    SELECT 1
    FROM saas_plan_feature existing
    WHERE existing.plan_id = plan.id
      AND existing.feature_id = feature.id
  );

INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM role r
JOIN permission p ON p.tenant_id = r.tenant_id
WHERE p.path IN (
    'business_unit:resource_bind',
    '/business-resources',
    'business_resource:create',
    'business_resource:edit',
    'business_resource:delete',
    'business_resource:relation_manage'
  )
  AND p.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND (r.code IN ('platform_admin', 'tenant_admin', 'admin') OR r.name LIKE '%管理员%')
  AND NOT EXISTS (
    SELECT 1
    FROM role_permission rp
    WHERE rp.role_id = r.id
      AND rp.permission_id = p.id
  );
