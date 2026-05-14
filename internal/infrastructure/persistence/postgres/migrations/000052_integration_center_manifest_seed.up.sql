-- Bootstrap integration-center from its Manifest baseline.
-- This migration mirrors internal/apps/integration_center/app.manifest.yaml so a fresh
-- production database has the backend menu, permission, package feature, quota, and
-- app-center asset records before the first interactive Manifest sync.

UPDATE sys_app
SET
  app_name = '第三方集成中心',
  icon = 'Connection',
  app_type = 'CONNECTOR_APP',
  source = 'BUILTIN',
  status = 'ONLINE',
  charge_mode = 'PAID',
  visibility_scope = 'TENANT',
  deployment_mode = 'MERGED',
  communication_modes = 'PLATFORM_API,WEBHOOK,DATA_SYNC,GATEWAY_PROXY',
  owner = '平台集成组',
  version = '0.1.0',
  description = '外部平台、服务商应用、授权连接实例、同步任务、配额限流、异常和调用日志的统一治理入口',
  is_builtin = TRUE,
  is_platform_only = FALSE,
  sort_order = 40,
  deleted_at = NULL,
  updated_at = now()
WHERE app_code = 'integration-center';

INSERT INTO sys_app (
  app_code, app_name, icon, app_type, source, status, charge_mode, visibility_scope,
  owner, version, description, is_builtin, is_platform_only, sort_order, created_at, updated_at
)
SELECT
  'integration-center', '第三方集成中心', 'Connection', 'CONNECTOR_APP', 'BUILTIN', 'ONLINE', 'PAID', 'TENANT',
  '平台集成组', '0.1.0', '外部平台、服务商应用、授权连接实例、同步任务、配额限流、异常和调用日志的统一治理入口',
  TRUE, FALSE, 40, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_app WHERE app_code = 'integration-center' AND deleted_at IS NULL
);

WITH app AS (
  SELECT id FROM sys_app WHERE app_code = 'integration-center' AND deleted_at IS NULL
),
clients(client_code, client_name, sort_order) AS (
  VALUES ('PC_WEB', 'PC Web', 1), ('API_ONLY', 'API Only', 2)
)
INSERT INTO sys_app_client (app_id, client_code, client_name, enabled, sort_order, config_note, created_at, updated_at)
SELECT app.id, clients.client_code, clients.client_name, TRUE, clients.sort_order, 'manifest:integration-center:0.1.0', now(), now()
FROM app
CROSS JOIN clients
WHERE NOT EXISTS (
  SELECT 1
  FROM sys_app_client existing
  WHERE existing.app_id = app.id
    AND existing.client_code = clients.client_code
    AND existing.deleted_at IS NULL
);

WITH menu(code, name, path, parent_code, sort_order, platform_only, tenant_visible, include_in_package, feature_code, data_perm_mode) AS (
  VALUES
    ('integration_overview', '第三方集成中心', '/integration-center', NULL, 70, TRUE, FALSE, FALSE, 'integration_overview', 'NONE'),
    ('integration_platforms', '接入平台', '/integration-center/platforms', 'integration_overview', 71, TRUE, FALSE, FALSE, 'integration_platforms', 'NONE'),
    ('integration_workspace', '集成工作台', '/integration-center/workspace', 'integration_overview', 72, TRUE, FALSE, FALSE, 'integration_workspace', 'NONE'),
    ('integration_tenant_connections', '租户连接', '/integration-center/tenant-connections', 'integration_overview', 73, TRUE, FALSE, FALSE, 'integration_tenant_connections', 'NONE'),
    ('integration_sync_monitor', '同步监控', '/integration-center/sync-monitor', 'integration_overview', 74, TRUE, FALSE, FALSE, 'integration_sync_monitor', 'NONE'),
    ('integration_quota', '配额与限流', '/integration-center/quota', 'integration_overview', 75, TRUE, FALSE, FALSE, 'integration_quota', 'NONE'),
    ('integration_alerts', '异常监控', '/integration-center/alerts', 'integration_overview', 76, TRUE, FALSE, FALSE, 'integration_alerts', 'NONE'),
    ('integration_logs', '调用日志', '/integration-center/logs', 'integration_overview', 77, TRUE, FALSE, FALSE, 'integration_logs', 'NONE')
)
INSERT INTO sys_app_entry (
  app_code, resource_code, name, path, parent_code, sort_order, platform_only,
  tenant_visible, show_in_admin, tenant_editable, include_in_package, feature_code,
  data_perm_mode, manifest_hash, managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', code, name, path, parent_code, sort_order, platform_only,
  tenant_visible, TRUE, FALSE, include_in_package, feature_code, data_perm_mode,
  'INTEGRATION_CENTER_BASELINE', TRUE, 'ACTIVE', now(), now(), now()
FROM menu
ON CONFLICT DO NOTHING;

WITH api(method, path, permission_code, audit) AS (
  VALUES
    ('GET', '/api/integration-center/overview', '/integration-center', FALSE),
    ('GET', '/api/integration-center/platforms', '/integration-center/platforms', FALSE),
    ('POST', '/api/integration-center/platforms', 'integration_center:platform_manage', TRUE),
    ('PUT', '/api/integration-center/platforms/{code}', 'integration_center:platform_manage', TRUE),
    ('GET', '/api/integration-center/workspace', '/integration-center/workspace', FALSE),
    ('GET', '/api/integration-center/tenant-connections', '/integration-center/tenant-connections', FALSE),
    ('GET', '/api/integration-center/sync-monitor', '/integration-center/sync-monitor', FALSE),
    ('GET', '/api/integration-center/quota', '/integration-center/quota', FALSE),
    ('GET', '/api/integration-center/alerts', '/integration-center/alerts', FALSE),
    ('GET', '/api/integration-center/logs', '/integration-center/logs', FALSE)
)
INSERT INTO sys_app_api (
  app_code, method, path, permission_code, public, audit, manifest_hash,
  managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', method, path, permission_code, FALSE, audit,
  'INTEGRATION_CENTER_BASELINE', TRUE, 'ACTIVE', now(), now(), now()
FROM api
ON CONFLICT DO NOTHING;

WITH perms(permission_code, name, permission_type, menu_code, platform_only, include_in_package, data_perm_mode) AS (
  VALUES
    ('/integration-center', '第三方集成中心', 'MENU', 'integration_overview', TRUE, FALSE, 'NONE'),
    ('/integration-center/platforms', '接入平台', 'MENU', 'integration_platforms', TRUE, FALSE, 'NONE'),
    ('/integration-center/workspace', '集成工作台', 'MENU', 'integration_workspace', TRUE, FALSE, 'NONE'),
    ('/integration-center/tenant-connections', '租户连接', 'MENU', 'integration_tenant_connections', TRUE, FALSE, 'NONE'),
    ('/integration-center/sync-monitor', '同步监控', 'MENU', 'integration_sync_monitor', TRUE, FALSE, 'NONE'),
    ('/integration-center/quota', '配额与限流', 'MENU', 'integration_quota', TRUE, FALSE, 'NONE'),
    ('/integration-center/alerts', '异常监控', 'MENU', 'integration_alerts', TRUE, FALSE, 'NONE'),
    ('/integration-center/logs', '调用日志', 'MENU', 'integration_logs', TRUE, FALSE, 'NONE'),
    ('integration_center:platform_manage', '第三方集成中心-平台配置', 'OPERATION', 'integration_platforms', TRUE, FALSE, 'NONE'),
    ('integration_center:app_manage', '第三方集成中心-服务商应用配置', 'OPERATION', 'integration_workspace', TRUE, FALSE, 'NONE'),
    ('integration_center:connection_manage', '第三方集成中心-连接治理', 'OPERATION', 'integration_tenant_connections', TRUE, FALSE, 'NONE'),
    ('integration_center:quota_manage', '第三方集成中心-配额限流配置', 'OPERATION', 'integration_quota', TRUE, FALSE, 'NONE')
)
INSERT INTO sys_app_permission (
  app_code, permission_code, name, permission_type, menu_code, platform_only,
  include_in_package, data_perm_mode, manifest_hash, managed_by_manifest, status,
  last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', permission_code, name, permission_type, menu_code, platform_only,
  include_in_package, data_perm_mode, 'INTEGRATION_CENTER_BASELINE', TRUE, 'ACTIVE',
  now(), now(), now()
FROM perms
ON CONFLICT DO NOTHING;

WITH platform_tenant AS (
  SELECT id AS tenant_id FROM tenant WHERE is_platform = TRUE AND deleted_at IS NULL ORDER BY id LIMIT 1
),
menus(name, path, sort_order, is_platform_only, is_package_feature, feature_code, data_perm_mode) AS (
  VALUES
    ('第三方集成中心', '/integration-center', 70, TRUE, FALSE, 'integration_overview', 'NONE'),
    ('接入平台', '/integration-center/platforms', 71, TRUE, FALSE, 'integration_platforms', 'NONE'),
    ('集成工作台', '/integration-center/workspace', 72, TRUE, FALSE, 'integration_workspace', 'NONE'),
    ('租户连接', '/integration-center/tenant-connections', 73, TRUE, FALSE, 'integration_tenant_connections', 'NONE'),
    ('同步监控', '/integration-center/sync-monitor', 74, TRUE, FALSE, 'integration_sync_monitor', 'NONE'),
    ('配额与限流', '/integration-center/quota', 75, TRUE, FALSE, 'integration_quota', 'NONE'),
    ('异常监控', '/integration-center/alerts', 76, TRUE, FALSE, 'integration_alerts', 'NONE'),
    ('调用日志', '/integration-center/logs', 77, TRUE, FALSE, 'integration_logs', 'NONE')
)
INSERT INTO permission (
  tenant_id, name, path, perm_type, sort_order, enabled, visible, show_in_admin,
  is_platform_only, is_package_feature, tenant_editable, app_code, feature_code,
  feature_type, data_perm_mode, created_at, updated_at
)
SELECT
  platform_tenant.tenant_id, menus.name, menus.path, 3, menus.sort_order, TRUE,
  FALSE, TRUE, menus.is_platform_only, menus.is_package_feature, FALSE,
  'integration-center', menus.feature_code, 'MENU', menus.data_perm_mode, now(), now()
FROM platform_tenant
CROSS JOIN menus
WHERE NOT EXISTS (
  SELECT 1 FROM permission p
  WHERE p.tenant_id = platform_tenant.tenant_id
    AND p.path = menus.path
    AND p.deleted_at IS NULL
);

WITH platform_tenant AS (
  SELECT id AS tenant_id FROM tenant WHERE is_platform = TRUE AND deleted_at IS NULL ORDER BY id LIMIT 1
),
ops(name, path, parent_path) AS (
  VALUES
    ('第三方集成中心-平台配置', 'integration_center:platform_manage', '/integration-center/platforms'),
    ('第三方集成中心-服务商应用配置', 'integration_center:app_manage', '/integration-center/workspace'),
    ('第三方集成中心-连接治理', 'integration_center:connection_manage', '/integration-center/tenant-connections'),
    ('第三方集成中心-配额限流配置', 'integration_center:quota_manage', '/integration-center/quota')
)
INSERT INTO permission (
  tenant_id, parent_id, name, path, perm_type, sort_order, enabled, visible,
  show_in_admin, is_platform_only, is_package_feature, tenant_editable, app_code,
  feature_type, data_perm_mode, created_at, updated_at
)
SELECT
  platform_tenant.tenant_id, parent.id, ops.name, ops.path, 2, 0, TRUE, FALSE,
  TRUE, TRUE, FALSE, FALSE, 'integration-center', 'OPERATION', 'NONE', now(), now()
FROM platform_tenant
CROSS JOIN ops
LEFT JOIN permission parent
  ON parent.tenant_id = platform_tenant.tenant_id
 AND parent.path = ops.parent_path
 AND parent.deleted_at IS NULL
WHERE NOT EXISTS (
  SELECT 1 FROM permission p
  WHERE p.tenant_id = platform_tenant.tenant_id
    AND p.path = ops.path
    AND p.deleted_at IS NULL
);

WITH platform_tenant AS (
  SELECT id AS tenant_id FROM tenant WHERE is_platform = TRUE AND deleted_at IS NULL ORDER BY id LIMIT 1
),
root AS (
  SELECT id FROM permission p, platform_tenant
  WHERE p.tenant_id = platform_tenant.tenant_id AND p.path = '/integration-center' AND p.deleted_at IS NULL
),
children AS (
  SELECT p.id
  FROM permission p, platform_tenant
  WHERE p.tenant_id = platform_tenant.tenant_id
    AND p.path LIKE '/integration-center/%'
    AND p.perm_type = 3
    AND p.deleted_at IS NULL
)
UPDATE permission
SET parent_id = (SELECT id FROM root), updated_at = now()
WHERE id IN (SELECT id FROM children)
  AND parent_id IS DISTINCT FROM (SELECT id FROM root);

WITH features(feature_code, feature_name, feature_type, parent_code, description) AS (
  VALUES
    ('integration_tenant_authorization', '第三方集成授权连接', 'FEATURE', NULL, '租户侧通过平台服务商应用创建第三方授权连接实例'),
    ('integration_data_sync', '第三方数据同步', 'FEATURE', 'integration_tenant_authorization', '第三方平台订单、组织、成员、商品等数据进入数据中心和 Agent 数据源')
)
INSERT INTO sys_app_package_feature (
  app_code, feature_code, feature_name, feature_type, parent_code, source_code,
  package_policy, include_in_package, description, manifest_hash,
  managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', feature_code, feature_name, feature_type, parent_code,
  feature_code, 'IN_PACKAGE', TRUE, description, 'INTEGRATION_CENTER_BASELINE',
  TRUE, 'ACTIVE', now(), now(), now()
FROM features
ON CONFLICT DO NOTHING;

WITH features(feature_code, feature_name, feature_type, parent_code, description) AS (
  VALUES
    ('integration_tenant_authorization', '第三方集成授权连接', 'FEATURE', NULL, '租户侧通过平台服务商应用创建第三方授权连接实例'),
    ('integration_data_sync', '第三方数据同步', 'FEATURE', 'integration_tenant_authorization', '第三方平台订单、组织、成员、商品等数据进入数据中心和 Agent 数据源')
)
INSERT INTO saas_feature (feature_code, feature_name, feature_type, app_code, parent_id, status, description, created_at, updated_at)
SELECT
  f.feature_code,
  f.feature_name,
  f.feature_type,
  'integration-center',
  COALESCE(parent.id, 0),
  1,
  f.description,
  now(),
  now()
FROM features f
LEFT JOIN saas_feature parent ON parent.feature_code = f.parent_code
WHERE NOT EXISTS (
  SELECT 1 FROM saas_feature existing WHERE existing.feature_code = f.feature_code
);

WITH quotas(quota_code, quota_name, quota_type, unit, period_type, description) AS (
  VALUES
    ('integration_connection_count', '第三方连接实例数', 'STATIC', 'COUNT', 'NONE', '单个租户可创建的第三方连接实例数量'),
    ('integration_api_calls_daily', '第三方接口日调用量', 'CONSUMABLE', 'TIMES', 'DAY', '单个租户第三方接口每日调用总量'),
    ('integration_sync_records_daily', '第三方同步日记录数', 'CONSUMABLE', 'RECORDS', 'DAY', '单个租户第三方同步任务每日写入记录总量')
)
INSERT INTO sys_app_quota (
  app_code, quota_code, quota_name, quota_type, unit, period_type, include_in_package,
  description, manifest_hash, managed_by_manifest, status, last_synced_at, created_at, updated_at
)
SELECT
  'integration-center', quota_code, quota_name, quota_type, unit, period_type, TRUE,
  description, 'INTEGRATION_CENTER_BASELINE', TRUE, 'ACTIVE', now(), now(), now()
FROM quotas
ON CONFLICT DO NOTHING;

WITH quotas(quota_code, quota_name, quota_type, unit, period_type, description) AS (
  VALUES
    ('integration_connection_count', '第三方连接实例数', 'STATIC', 'COUNT', 'NONE', '单个租户可创建的第三方连接实例数量'),
    ('integration_api_calls_daily', '第三方接口日调用量', 'CONSUMABLE', 'TIMES', 'DAY', '单个租户第三方接口每日调用总量'),
    ('integration_sync_records_daily', '第三方同步日记录数', 'CONSUMABLE', 'RECORDS', 'DAY', '单个租户第三方同步任务每日写入记录总量')
)
INSERT INTO saas_quota (quota_code, quota_name, quota_type, period_type, unit, status, description, created_at, updated_at)
SELECT quota_code, quota_name, quota_type, period_type, unit, 1, description, now(), now()
FROM quotas
WHERE NOT EXISTS (
  SELECT 1 FROM saas_quota existing WHERE existing.quota_code = quotas.quota_code
);

WITH integration_perms AS (
  SELECT id FROM permission
  WHERE app_code = 'integration-center' AND deleted_at IS NULL
),
platform_roles AS (
  SELECT id FROM role WHERE tenant_id = 1 AND deleted_at IS NULL
)
INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT platform_roles.id, integration_perms.id, now()
FROM platform_roles
CROSS JOIN integration_perms
ON CONFLICT (role_id, permission_id) DO NOTHING;
