-- Add the tenant-facing Integration Center portal while keeping the platform
-- governance routes platform-only.

WITH menu(code, name, path, sort_order, platform_only, tenant_visible, show_in_admin, tenant_editable, include_in_package, feature_code, data_perm_mode) AS (
  VALUES
    ('integration_my_connections', '我的第三方连接', '/integration-center/my-connections', 73, FALSE, TRUE, TRUE, TRUE, TRUE, 'integration_tenant_authorization', 'TENANT')
)
INSERT INTO sys_app_entry (
  app_code, resource_code, name, path, parent_code, sort_order, platform_only,
  tenant_visible, show_in_admin, tenant_editable, include_in_package, feature_code,
  data_perm_mode, manifest_hash, managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', code, name, path, NULL, sort_order, platform_only,
  tenant_visible, show_in_admin, tenant_editable, include_in_package, feature_code,
  data_perm_mode, 'INTEGRATION_CENTER_TENANT_PORTAL', TRUE, 'ACTIVE', now(), now(), now()
FROM menu
ON CONFLICT (app_code, resource_code) WHERE deleted_at IS NULL
DO UPDATE SET
  name = EXCLUDED.name,
  path = EXCLUDED.path,
  sort_order = EXCLUDED.sort_order,
  platform_only = EXCLUDED.platform_only,
  tenant_visible = EXCLUDED.tenant_visible,
  show_in_admin = EXCLUDED.show_in_admin,
  tenant_editable = EXCLUDED.tenant_editable,
  include_in_package = EXCLUDED.include_in_package,
  feature_code = EXCLUDED.feature_code,
  data_perm_mode = EXCLUDED.data_perm_mode,
  managed_by_manifest = TRUE,
  status = 'ACTIVE',
  last_synced_at = now(),
  updated_at = now();

WITH perms(permission_code, name, permission_type, menu_code, platform_only, include_in_package, data_perm_mode) AS (
  VALUES
    ('/integration-center/my-connections', '我的第三方连接', 'MENU', 'integration_my_connections', FALSE, TRUE, 'TENANT')
)
INSERT INTO sys_app_permission (
  app_code, permission_code, name, permission_type, menu_code, platform_only,
  include_in_package, data_perm_mode, manifest_hash, managed_by_manifest, status,
  last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', permission_code, name, permission_type, menu_code, platform_only,
  include_in_package, data_perm_mode, 'INTEGRATION_CENTER_TENANT_PORTAL', TRUE, 'ACTIVE',
  now(), now(), now()
FROM perms
ON CONFLICT (app_code, permission_code) WHERE deleted_at IS NULL
DO UPDATE SET
  name = EXCLUDED.name,
  permission_type = EXCLUDED.permission_type,
  menu_code = EXCLUDED.menu_code,
  platform_only = EXCLUDED.platform_only,
  include_in_package = EXCLUDED.include_in_package,
  data_perm_mode = EXCLUDED.data_perm_mode,
  managed_by_manifest = TRUE,
  status = 'ACTIVE',
  last_synced_at = now(),
  updated_at = now();

WITH api(method, path, permission_code, audit) AS (
  VALUES
    ('GET', '/api/integration-center/my-connections', '/integration-center/my-connections', FALSE),
    ('GET', '/api/integration-center/my-connections/{id}', '/integration-center/my-connections', FALSE),
    ('POST', '/api/integration-center/my-connections', '/integration-center/my-connections', TRUE),
    ('POST', '/api/integration-center/my-oauth/start', '/integration-center/my-connections', TRUE),
    ('GET', '/api/integration-center/my-sync-jobs', '/integration-center/my-connections', FALSE),
    ('GET', '/api/integration-center/my-sync-jobs/{id}', '/integration-center/my-connections', FALSE)
)
INSERT INTO sys_app_api (
  app_code, method, path, permission_code, public, audit, manifest_hash,
  managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', method, path, permission_code, FALSE, audit,
  'INTEGRATION_CENTER_TENANT_PORTAL', TRUE, 'ACTIVE', now(), now(), now()
FROM api
ON CONFLICT (app_code, method, path) WHERE deleted_at IS NULL
DO UPDATE SET
  permission_code = EXCLUDED.permission_code,
  public = EXCLUDED.public,
  audit = EXCLUDED.audit,
  managed_by_manifest = TRUE,
  status = 'ACTIVE',
  last_synced_at = now(),
  updated_at = now();

WITH tenant_menu AS (
  SELECT
    t.id AS tenant_id,
    p.id AS existing_permission_id
  FROM tenant t
  LEFT JOIN permission p
    ON p.tenant_id = t.id
   AND p.path = '/integration-center/my-connections'
   AND p.deleted_at IS NULL
  WHERE t.deleted_at IS NULL
    AND t.is_platform_tenant = FALSE
),
inserted AS (
  INSERT INTO permission (
    tenant_id, name, path, perm_type, sort_order, enabled, visible, show_in_admin,
    is_platform_only, is_package_feature, tenant_editable, app_code, feature_code,
    feature_type, data_perm_mode, created_at, updated_at
  )
  SELECT
    tenant_id, '我的第三方连接', '/integration-center/my-connections', 3, 73,
    TRUE, TRUE, TRUE, FALSE, TRUE, TRUE, 'integration-center',
    'integration_tenant_authorization', 'MENU', 'TENANT', now(), now()
  FROM tenant_menu
  WHERE existing_permission_id IS NULL
  RETURNING id, tenant_id
),
all_portal_permissions AS (
  SELECT id, tenant_id FROM inserted
  UNION ALL
  SELECT existing_permission_id AS id, tenant_id
  FROM tenant_menu
  WHERE existing_permission_id IS NOT NULL
),
tenant_roles AS (
  SELECT r.id AS role_id, p.id AS permission_id
  FROM role r
  JOIN all_portal_permissions p ON p.tenant_id = r.tenant_id
  WHERE r.deleted_at IS NULL
    AND r.status = 1
)
INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT role_id, permission_id, now()
FROM tenant_roles
ON CONFLICT (role_id, permission_id) DO NOTHING;

UPDATE permission
SET
  name = '我的第三方连接',
  enabled = TRUE,
  visible = TRUE,
  show_in_admin = TRUE,
  is_platform_only = FALSE,
  is_package_feature = TRUE,
  tenant_editable = TRUE,
  app_code = 'integration-center',
  feature_code = 'integration_tenant_authorization',
  feature_type = 'MENU',
  data_perm_mode = 'TENANT',
  updated_at = now()
WHERE path = '/integration-center/my-connections'
  AND deleted_at IS NULL;
