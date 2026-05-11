CREATE TABLE IF NOT EXISTS sys_app (
  id BIGSERIAL PRIMARY KEY,
  app_code VARCHAR(100) NOT NULL UNIQUE,
  app_name VARCHAR(100) NOT NULL,
  icon VARCHAR(80),
  app_type VARCHAR(32) NOT NULL,
  source VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  charge_mode VARCHAR(32) NOT NULL DEFAULT 'NON_SELLABLE',
  visibility_scope VARCHAR(32) NOT NULL DEFAULT 'PLATFORM_ONLY',
  owner VARCHAR(100),
  version VARCHAR(64),
  description VARCHAR(500),
  is_builtin BOOLEAN NOT NULL DEFAULT FALSE,
  is_platform_only BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sys_app_type_deleted ON sys_app (app_type, deleted_at);
CREATE INDEX IF NOT EXISTS idx_sys_app_status_deleted ON sys_app (status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_sys_app_source_deleted ON sys_app (source, deleted_at);
CREATE INDEX IF NOT EXISTS idx_sys_app_builtin_deleted ON sys_app (is_builtin, deleted_at);

INSERT INTO sys_app (
  app_code, app_name, icon, app_type, source, status, charge_mode, visibility_scope,
  owner, version, description, is_builtin, is_platform_only, sort_order, created_at, updated_at
) VALUES
  ('app-center', '应用中心', 'Boxes', 'SYSTEM_APP', 'BUILTIN', 'ONLINE', 'NON_SELLABLE', 'PLATFORM_ONLY',
   '平台架构组', '0.1.0', '应用注册、装载规范与平台内置应用治理入口', TRUE, TRUE, 1, now(), now()),
  ('system-management', '系统管理', 'Settings', 'SYSTEM_APP', 'BUILTIN', 'ONLINE', 'NON_SELLABLE', 'PLATFORM_ONLY',
   '平台架构组', '0.1.0', '用户、角色、菜单、字典、参数等基础治理能力', TRUE, TRUE, 10, now(), now()),
  ('system-monitor', '系统监控', 'MonitorCog', 'SYSTEM_APP', 'BUILTIN', 'ONLINE', 'NON_SELLABLE', 'PLATFORM_ONLY',
   '平台运维组', '0.1.0', '健康检查、服务状态、缓存和定时任务等运维监控能力', TRUE, TRUE, 20, now(), now()),
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
