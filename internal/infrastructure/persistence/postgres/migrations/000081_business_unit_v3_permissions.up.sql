WITH business_unit_menu AS (
  SELECT p.tenant_id, p.id AS parent_id
  FROM permission p
  WHERE p.path = '/business-units'
    AND p.perm_type = 3
    AND p.deleted_at IS NULL
),
seed(name, path, sort_order, feature_code, data_perm_mode) AS (
  VALUES
    ('业务单元-新增', 'business_unit:create', 50, 'button_business_unit_create', 'BU'),
    ('业务单元-编辑', 'business_unit:edit', 60, 'button_business_unit_edit', 'BU'),
    ('业务单元-归档', 'business_unit:delete', 70, 'button_business_unit_delete', 'BU'),
    ('业务单元-关系维护', 'business_unit:relation_manage', 80, 'button_business_unit_relation_manage', 'BU'),
    ('业务单元-责任方维护', 'business_unit:actor_manage', 90, 'button_business_unit_actor_manage', 'BU'),
    ('业务单元-属性模板', 'business_unit:attr_template_manage', 100, 'button_business_unit_attr_template_manage', 'NONE')
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
  2,
  seed.sort_order,
  true,
  true,
  true,
  false,
  true,
  false,
  'system-management',
  seed.feature_code,
  'OPERATION',
  seed.data_perm_mode,
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
    AND p.perm_type = 2
    AND p.deleted_at IS NULL
);

WITH business_unit_menu AS (
  SELECT p.tenant_id, p.id AS parent_id
  FROM permission p
  WHERE p.path = '/business-units'
    AND p.perm_type = 3
    AND p.deleted_at IS NULL
),
seed(name, path, sort_order, feature_code, data_perm_mode) AS (
  VALUES
    ('业务单元-新增', 'business_unit:create', 50, 'button_business_unit_create', 'BU'),
    ('业务单元-编辑', 'business_unit:edit', 60, 'button_business_unit_edit', 'BU'),
    ('业务单元-归档', 'business_unit:delete', 70, 'button_business_unit_delete', 'BU'),
    ('业务单元-关系维护', 'business_unit:relation_manage', 80, 'button_business_unit_relation_manage', 'BU'),
    ('业务单元-责任方维护', 'business_unit:actor_manage', 90, 'button_business_unit_actor_manage', 'BU'),
    ('业务单元-属性模板', 'business_unit:attr_template_manage', 100, 'button_business_unit_attr_template_manage', 'NONE')
)
UPDATE permission p
SET name = seed.name,
    parent_id = bum.parent_id,
    sort_order = seed.sort_order,
    enabled = true,
    visible = true,
    show_in_admin = true,
    is_platform_only = false,
    is_package_feature = true,
    app_code = 'system-management',
    feature_code = seed.feature_code,
    feature_type = 'OPERATION',
    data_perm_mode = seed.data_perm_mode,
    updated_at = now()
FROM business_unit_menu bum, seed
WHERE p.tenant_id = bum.tenant_id
  AND p.path = seed.path
  AND p.perm_type = 2
  AND p.deleted_at IS NULL;

WITH feature_seed(feature_code, feature_name, parent_code, source_code, description) AS (
  VALUES
    ('button_business_unit_create', '业务单元-新增', 'business_unit_manage', 'business_unit:create', '业务单元新增能力'),
    ('button_business_unit_edit', '业务单元-编辑', 'business_unit_manage', 'business_unit:edit', '业务单元编辑能力'),
    ('button_business_unit_delete', '业务单元-归档', 'business_unit_manage', 'business_unit:delete', '业务单元归档能力'),
    ('button_business_unit_relation_manage', '业务单元-关系维护', 'business_unit_manage', 'business_unit:relation_manage', '业务单元横向关联维护能力'),
    ('button_business_unit_actor_manage', '业务单元-责任方维护', 'business_unit_manage', 'business_unit:actor_manage', '业务单元负责组织和负责人维护能力'),
    ('button_business_unit_attr_template_manage', '业务单元-属性模板', 'business_unit_manage', 'business_unit:attr_template_manage', '业务单元自定义属性模板维护能力')
),
feature_rows AS (
  SELECT
    fs.feature_code,
    fs.feature_name,
    COALESCE(parent.id, 0) AS parent_id,
    fs.source_code,
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
  'OPERATION',
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
  'button_business_unit_create',
  'button_business_unit_edit',
  'button_business_unit_delete',
  'button_business_unit_relation_manage',
  'button_business_unit_actor_manage',
  'button_business_unit_attr_template_manage'
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
    'business_unit:create',
    'business_unit:edit',
    'business_unit:delete',
    'business_unit:relation_manage',
    'business_unit:actor_manage',
    'business_unit:attr_template_manage'
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
