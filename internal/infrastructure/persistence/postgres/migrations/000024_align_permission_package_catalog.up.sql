-- 统一菜单归属与套餐中心收录口径：
-- 1. 平台专属、首页、监控、兼容入口不进入租户套餐矩阵。
-- 2. 租户侧菜单与真实可控操作显式进入套餐矩阵。
-- 3. 套餐中心只保留可被租户套餐控制的能力。

UPDATE permission
SET is_package_feature = FALSE
WHERE deleted_at IS NULL
  AND (
    path IN (
      '/home',
      '/tenants',
      '/plans',
      '/permissions',
      'home:view',
      'menu:delete',
      'menu:package_feature',
      'dict:create',
      'dict:edit',
      'dict:delete'
    )
    OR path LIKE '/monitor/%'
    OR path LIKE 'tenant:%'
    OR path LIKE 'plan:%'
    OR path LIKE 'dict_type:%'
    OR path LIKE 'perm:%'
    OR path LIKE 'data:%'
    OR perm_type NOT IN (2, 3)
  );

UPDATE permission
SET is_platform_only = TRUE,
    is_package_feature = FALSE,
    data_perm_mode = 'NONE'
WHERE deleted_at IS NULL
  AND (
    path LIKE 'tenant:%'
    OR path LIKE 'plan:%'
    OR path LIKE 'dict_type:%'
    OR path LIKE 'perm:%'
    OR path IN ('menu:delete', 'menu:package_feature')
    OR path IN ('/tenants', '/plans')
    OR path LIKE '/monitor/%'
  );

UPDATE permission
SET is_platform_only = FALSE,
    is_package_feature = TRUE
WHERE deleted_at IS NULL
  AND path IN (
    '/organization',
    '/positions',
    '/business-units',
    '/users',
    '/roles',
    '/menus',
    '/dict',
    '/params',
    '/audit-logs',
    '/login-logs'
  );

UPDATE permission
SET is_platform_only = FALSE,
    is_package_feature = TRUE
WHERE deleted_at IS NULL
  AND path IN (
    'org:create',
    'org:edit',
    'org:delete',
    'pos:create',
    'pos:edit',
    'pos:delete',
    'business_unit:create',
    'business_unit:edit',
    'business_unit:delete',
    'user:create',
    'user:edit',
    'user:reset_password',
    'user:delete',
    'role:create',
    'role:edit',
    'role:delete',
    'role:permission',
    'menu:create',
    'menu:edit',
    'dict_item:create',
    'dict_item:edit',
    'dict_item:delete',
    'param:create',
    'param:edit',
    'param:delete',
    'brand:edit'
  );

UPDATE permission
SET feature_code = 'brand_config',
    feature_type = 'CONFIG',
    tenant_editable = TRUE,
    tenant_edit_scope = 'NAME_ICON',
    is_platform_only = FALSE,
    is_package_feature = TRUE
WHERE deleted_at IS NULL
  AND path = 'brand:edit';

UPDATE saas_feature
SET status = 0
WHERE feature_code IN ('home', 'tenant_manage', 'plan_manage', 'system_monitor', 'button_menu_delete', 'button_menu_package_feature');

DELETE FROM saas_plan_feature
WHERE feature_id IN (
  SELECT id
  FROM saas_feature
  WHERE feature_code IN ('home', 'tenant_manage', 'plan_manage', 'system_monitor', 'button_menu_delete', 'button_menu_package_feature')
);
