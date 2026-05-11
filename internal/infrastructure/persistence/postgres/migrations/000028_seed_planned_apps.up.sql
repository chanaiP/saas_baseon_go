INSERT INTO sys_app (
  app_code, app_name, icon, app_type, source, status, charge_mode, visibility_scope,
  owner, version, description, is_builtin, is_platform_only, sort_order, created_at, updated_at
) VALUES
  ('workbench', '工作台', 'House', 'SYSTEM_APP', 'BUILTIN', 'PLANNED', 'NON_SELLABLE', 'PLATFORM_ONLY',
   '平台产品组', '0.1.0', '平台与租户用户的统一工作入口、待办与概览能力', TRUE, TRUE, 30, now(), now()),
  ('integration-center', '第三方集成中心', 'Connection', 'CONNECTOR_APP', 'BUILTIN', 'PLANNED', 'SUBSCRIPTION', 'TENANT',
   '平台集成组', '0.1.0', '第三方系统、开放 API、Webhook、OAuth 与外部连接器的统一接入中心', TRUE, FALSE, 40, now(), now()),
  ('data-center', '数据中心', 'DataAnalysis', 'ABILITY_APP', 'BUILTIN', 'PLANNED', 'SUBSCRIPTION', 'TENANT',
   '数据产品组', '0.1.0', '跨应用数据资产、指标、报表、经营预警与数据看板能力中心', TRUE, FALSE, 50, now(), now())
ON CONFLICT (app_code) DO UPDATE SET
  app_name = EXCLUDED.app_name,
  icon = EXCLUDED.icon,
  app_type = EXCLUDED.app_type,
  source = EXCLUDED.source,
  status = EXCLUDED.status,
  charge_mode = EXCLUDED.charge_mode,
  visibility_scope = EXCLUDED.visibility_scope,
  owner = EXCLUDED.owner,
  version = EXCLUDED.version,
  description = EXCLUDED.description,
  is_builtin = EXCLUDED.is_builtin,
  is_platform_only = EXCLUDED.is_platform_only,
  sort_order = EXCLUDED.sort_order,
  deleted_at = NULL,
  updated_at = now();
