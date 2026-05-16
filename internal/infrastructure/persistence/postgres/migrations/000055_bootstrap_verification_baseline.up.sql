-- Align SQL bootstrap data with the bootstrap verifier baseline.

WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
)
INSERT INTO permission (
  tenant_id, name, path, perm_type, sort_order, enabled, visible, show_in_admin,
  is_platform_only, is_package_feature, tenant_editable, app_code, feature_code,
  feature_type, data_perm_mode, created_at, updated_at
)
SELECT
  platform_tenant.tenant_id, '首页-查看', 'home:view', 2, 0, TRUE, FALSE, TRUE,
  FALSE, FALSE, FALSE, 'system-management', 'home', 'VIEW', 'NONE', now(), now()
FROM platform_tenant
WHERE NOT EXISTS (
  SELECT 1
  FROM permission p
  WHERE p.tenant_id = platform_tenant.tenant_id
    AND p.path = 'home:view'
    AND p.deleted_at IS NULL
);

WITH platform_tenant AS (
  SELECT id AS tenant_id
  FROM tenant
  WHERE code = 'platform' AND is_platform_tenant = TRUE AND deleted_at IS NULL
  ORDER BY id
  LIMIT 1
),
params(param_key, param_value, remark, value_type, tenant_editable) AS (
  VALUES
    ('org.default_company_type', 'SUBSIDIARY', '新建公司默认类型（字典 company_type 的 value，须一致）', 'string', TRUE),
    ('user.list_default_page_size', '10', '用户列表默认每页条数', 'number', TRUE),
    ('security.password_min_length', '8', '用户密码最小长度', 'number', FALSE),
    ('security.password_require_complexity', 'true', '用户密码是否要求复杂度校验', 'boolean', FALSE),
    ('audit.log_retention_days', '180', '审计日志默认保留天数', 'number', FALSE)
)
INSERT INTO sys_param (
  tenant_id, param_key, param_value, remark, value_type, tenant_editable,
  is_platform_only, created_at, updated_at
)
SELECT
  platform_tenant.tenant_id, params.param_key, params.param_value, params.remark,
  params.value_type, params.tenant_editable, FALSE, now(), now()
FROM platform_tenant
CROSS JOIN params
ON CONFLICT (tenant_id, param_key) DO NOTHING;

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
),
platform_roles AS (
  SELECT role.id
  FROM role
  JOIN platform_tenant ON platform_tenant.tenant_id = role.tenant_id
  WHERE role.deleted_at IS NULL
)
INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT platform_roles.id, home_view_permission.id, now()
FROM platform_roles
CROSS JOIN home_view_permission
WHERE NOT EXISTS (
  SELECT 1
  FROM role_permission rp
  WHERE rp.role_id = platform_roles.id
    AND rp.permission_id = home_view_permission.id
);
