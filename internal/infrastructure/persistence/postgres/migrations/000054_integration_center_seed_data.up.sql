-- Safe, idempotent starter data for the integration center domain.
-- No real third-party credentials are stored here; credential_ref values point to external secret storage placeholders.

INSERT INTO integration_platforms (
  platform_code, platform_name, platform_short_name, platform_type, access_mode,
  official_url, status, tenant_visible, owner_name, sort_order, description, created_at, updated_at
)
SELECT *
FROM (VALUES
  ('wecom', '企业微信', '企微', '协同办公', '第三方服务商', 'https://open.work.weixin.qq.com', 'online', TRUE, '平台集成组', 10, '企业微信通讯录、审批与消息事件接入。', now(), now()),
  ('douyin-shop', '抖音电商', '抖店', '电商平台', 'OAuth2', 'https://op.jinritemai.com', 'online', TRUE, '电商集成组', 20, '抖店订单、商品与售后数据接入。', now(), now()),
  ('jd-shop', '京东电商', '京东', '电商平台', 'OAuth2', 'https://jos.jd.com', 'beta', TRUE, '电商集成组', 30, '京东店铺订单、商品和库存同步。', now(), now()),
  ('generic-openapi', '通用 OpenAPI', 'OpenAPI', '开放接口', 'API Key', NULL, 'draft', FALSE, '平台集成组', 90, '面向非标准第三方系统的 API Key 与 HMAC 接入模板。', now(), now())
) AS seed(platform_code, platform_name, platform_short_name, platform_type, access_mode, official_url, status, tenant_visible, owner_name, sort_order, description, created_at, updated_at)
WHERE NOT EXISTS (
  SELECT 1 FROM integration_platforms p
  WHERE p.platform_code = seed.platform_code AND p.deleted_at IS NULL
);

WITH platform AS (
  SELECT id, platform_code FROM integration_platforms WHERE deleted_at IS NULL
),
seed AS (
  SELECT platform.id AS platform_id, v.capability_code, v.capability_name, v.capability_type, v.auth_scope_code, v.data_direction, v.status, v.description
  FROM platform
  JOIN (VALUES
    ('wecom', 'org_sync', '组织同步', 'organization', 'contact.read', 'pull', 'enabled', '同步部门、成员和组织关系。'),
    ('wecom', 'approval_event', '审批事件', 'workflow', 'approval.read', 'pull', 'enabled', '接收审批完成和变更事件。'),
    ('douyin-shop', 'order_sync', '订单同步', 'order', 'order.search', 'pull', 'enabled', '同步抖店订单和售后状态。'),
    ('douyin-shop', 'product_sync', '商品同步', 'product', 'product.read', 'pull', 'enabled', '同步商品档案和上下架状态。'),
    ('jd-shop', 'order_sync', '订单同步', 'order', 'order.search', 'pull', 'enabled', '同步京东订单和发货状态。'),
    ('generic-openapi', 'api_proxy', 'API 代理', 'api', 'proxy.invoke', 'proxy', 'enabled', '通过底座网关代理第三方 API。')
  ) AS v(platform_code, capability_code, capability_name, capability_type, auth_scope_code, data_direction, status, description)
    ON v.platform_code = platform.platform_code
)
INSERT INTO integration_platform_capabilities (
  platform_id, capability_code, capability_name, capability_type, auth_scope_code,
  data_direction, status, description, created_at, updated_at
)
SELECT seed.platform_id, seed.capability_code, seed.capability_name, seed.capability_type,
  seed.auth_scope_code, seed.data_direction, seed.status, seed.description, now(), now()
FROM seed
WHERE NOT EXISTS (
  SELECT 1 FROM integration_platform_capabilities c
  WHERE c.platform_id = seed.platform_id AND c.capability_code = seed.capability_code AND c.deleted_at IS NULL
);

WITH platform AS (
  SELECT id, platform_code FROM integration_platforms WHERE deleted_at IS NULL
),
seed AS (
  SELECT platform.id AS platform_id, v.app_code, v.app_name, v.auth_mode, v.environment, v.status, v.tenant_visible, v.callback_url, v.webhook_url, v.credential_ref, v.owner_name, v.description
  FROM platform
  JOIN (VALUES
    ('wecom', 'wecom-suite-standard', '企业微信第三方标准应用', 'OAuth2', 'prod', 'online', TRUE, '/api/integration-center/callbacks/wecom/oauth', '/api/integration-center/webhooks/wecom', 'vault://integration-center/wecom-suite-standard', '平台集成组', '平台统一创建的企业微信第三方应用身份。'),
    ('douyin-shop', 'douyin-shop-isv-prod', '抖音电商正式服务商应用', 'OAuth2', 'prod', 'online', TRUE, '/api/integration-center/callbacks/douyin/oauth', '/api/integration-center/webhooks/douyin', 'vault://integration-center/douyin-shop-isv-prod', '电商集成组', '抖音电商服务商应用，用于店铺授权和订单同步。'),
    ('jd-shop', 'jd-shop-isv-beta', '京东电商测试服务商应用', 'OAuth2', 'sandbox', 'beta', TRUE, '/api/integration-center/callbacks/jd/oauth', '/api/integration-center/webhooks/jd', 'vault://integration-center/jd-shop-isv-beta', '电商集成组', '京东电商联调应用。'),
    ('generic-openapi', 'generic-hmac-template', '通用 HMAC 接入模板', 'API_KEY_HMAC', 'prod', 'draft', FALSE, NULL, NULL, 'vault://integration-center/generic-hmac-template', '平台集成组', '用于未标准化第三方系统的通用模板。')
  ) AS v(platform_code, app_code, app_name, auth_mode, environment, status, tenant_visible, callback_url, webhook_url, credential_ref, owner_name, description)
    ON v.platform_code = platform.platform_code
)
INSERT INTO integration_provider_apps (
  platform_id, app_code, app_name, auth_mode, environment, status, tenant_visible,
  callback_url, webhook_url, credential_ref, owner_name, description, created_at, updated_at
)
SELECT seed.platform_id, seed.app_code, seed.app_name, seed.auth_mode, seed.environment,
  seed.status, seed.tenant_visible, seed.callback_url, seed.webhook_url, seed.credential_ref,
  seed.owner_name, seed.description, now(), now()
FROM seed
WHERE NOT EXISTS (
  SELECT 1 FROM integration_provider_apps app
  WHERE app.platform_id = seed.platform_id AND app.app_code = seed.app_code AND app.deleted_at IS NULL
);

WITH app AS (
  SELECT pa.id, pa.app_code, pa.platform_id FROM integration_provider_apps pa WHERE pa.deleted_at IS NULL
),
capability AS (
  SELECT id, platform_id, capability_code FROM integration_platform_capabilities WHERE deleted_at IS NULL
),
seed AS (
  SELECT app.id AS provider_app_id, capability.id AS platform_capability_id, v.connection_status, v.review_status, v.enabled
  FROM app
  JOIN (VALUES
    ('wecom-suite-standard', 'org_sync', 'connected', 'approved', TRUE),
    ('wecom-suite-standard', 'approval_event', 'connected', 'approved', TRUE),
    ('douyin-shop-isv-prod', 'order_sync', 'connected', 'approved', TRUE),
    ('douyin-shop-isv-prod', 'product_sync', 'connected', 'approved', TRUE),
    ('jd-shop-isv-beta', 'order_sync', 'testing', 'pending', FALSE),
    ('generic-hmac-template', 'api_proxy', 'pending', 'pending', FALSE)
  ) AS v(app_code, capability_code, connection_status, review_status, enabled)
    ON v.app_code = app.app_code
  JOIN capability ON capability.platform_id = app.platform_id AND capability.capability_code = v.capability_code
)
INSERT INTO integration_provider_app_capabilities (
  provider_app_id, platform_capability_id, connection_status, review_status, enabled, config, created_at, updated_at
)
SELECT seed.provider_app_id, seed.platform_capability_id, seed.connection_status, seed.review_status,
  seed.enabled, '{}'::jsonb, now(), now()
FROM seed
WHERE NOT EXISTS (
  SELECT 1 FROM integration_provider_app_capabilities ac
  WHERE ac.provider_app_id = seed.provider_app_id
    AND ac.platform_capability_id = seed.platform_capability_id
    AND ac.deleted_at IS NULL
);

INSERT INTO integration_quota_policies (
  policy_code, policy_name, quota_code, quota_unit, period_type, default_limit,
  over_limit_action, status, description, created_at, updated_at
)
SELECT *
FROM (VALUES
  ('integration_default_connections', '租户默认连接实例上限', 'integration_connection_count', 'COUNT', 'STATIC', 20, 'reject', 'enabled', '控制单个租户可建立的第三方连接实例数。', now(), now()),
  ('integration_default_api_daily', '第三方接口日调用默认策略', 'integration_api_calls_daily', 'CALL', 'DAY', 100000, 'reject', 'enabled', '控制租户第三方接口每日调用量。', now(), now()),
  ('integration_default_sync_daily', '第三方同步日记录默认策略', 'integration_sync_records_daily', 'RECORD', 'DAY', 500000, 'queue', 'enabled', '控制租户第三方同步每日记录量。', now(), now())
) AS seed(policy_code, policy_name, quota_code, quota_unit, period_type, default_limit, over_limit_action, status, description, created_at, updated_at)
WHERE NOT EXISTS (
  SELECT 1 FROM integration_quota_policies p
  WHERE p.policy_code = seed.policy_code AND p.deleted_at IS NULL
);

WITH tenant_seed AS (
  SELECT id AS tenant_id FROM tenant WHERE is_platform_tenant = FALSE AND deleted_at IS NULL ORDER BY id LIMIT 1
),
app_seed AS (
  SELECT pa.id AS provider_app_id, pa.platform_id
  FROM integration_provider_apps pa
  WHERE pa.app_code = 'wecom-suite-standard' AND pa.deleted_at IS NULL
  LIMIT 1
),
inserted_connection AS (
  INSERT INTO integration_tenant_connections (
    tenant_id, platform_id, provider_app_id, connection_name, auth_subject_type,
    auth_subject_id, auth_subject_name, auth_scope, auth_status, connection_status,
    token_status, authorized_at, last_sync_at, created_at, updated_at
  )
  SELECT tenant_seed.tenant_id, app_seed.platform_id, app_seed.provider_app_id,
    '默认租户企业微信连接', 'corp', 'demo-corp', '默认租户企业微信',
    '["contact.read","approval.read"]'::jsonb, 'authorized', 'connected', 'valid',
    now() - interval '2 days', now() - interval '15 minutes', now(), now()
  FROM tenant_seed, app_seed
  WHERE NOT EXISTS (
    SELECT 1 FROM integration_tenant_connections c
    WHERE c.tenant_id = tenant_seed.tenant_id
      AND c.platform_id = app_seed.platform_id
      AND c.provider_app_id = app_seed.provider_app_id
      AND c.auth_subject_type = 'corp'
      AND c.auth_subject_id = 'demo-corp'
      AND c.deleted_at IS NULL
  )
  RETURNING id, tenant_id, platform_id, provider_app_id
),
existing_connection AS (
  SELECT c.id, c.tenant_id, c.platform_id, c.provider_app_id
  FROM integration_tenant_connections c
  JOIN tenant_seed ON tenant_seed.tenant_id = c.tenant_id
  JOIN app_seed ON app_seed.platform_id = c.platform_id AND app_seed.provider_app_id = c.provider_app_id
  WHERE c.auth_subject_type = 'corp' AND c.auth_subject_id = 'demo-corp' AND c.deleted_at IS NULL
),
connection AS (
  SELECT * FROM inserted_connection
  UNION ALL
  SELECT * FROM existing_connection
  LIMIT 1
)
INSERT INTO integration_sync_jobs (
  tenant_id, tenant_connection_id, capability_code, job_type, trigger_mode, status,
  total_count, success_count, failed_count, started_at, finished_at, created_at, updated_at
)
SELECT connection.tenant_id, connection.id, 'org_sync', 'incremental', 'schedule', 'success',
  1280, 1276, 4, now() - interval '20 minutes', now() - interval '15 minutes', now(), now()
FROM connection
WHERE NOT EXISTS (
  SELECT 1 FROM integration_sync_jobs j
  WHERE j.tenant_connection_id = connection.id AND j.capability_code = 'org_sync' AND j.deleted_at IS NULL
);

WITH app_seed AS (
  SELECT pa.id AS provider_app_id, pa.platform_id FROM integration_provider_apps pa WHERE pa.app_code = 'jd-shop-isv-beta' AND pa.deleted_at IS NULL LIMIT 1
)
INSERT INTO integration_alerts (
  platform_id, provider_app_id, alert_type, severity, status, title, message,
  first_seen_at, last_seen_at, created_at, updated_at
)
SELECT app_seed.platform_id, app_seed.provider_app_id, 'oauth_review', 'warning', 'open',
  '京东服务商应用仍在联调审核中', '该应用能力尚未全部审核通过，租户侧授权入口保持受控开放。',
  now() - interval '1 day', now() - interval '30 minutes', now(), now()
FROM app_seed
WHERE NOT EXISTS (
  SELECT 1 FROM integration_alerts a
  WHERE a.provider_app_id = app_seed.provider_app_id
    AND a.alert_type = 'oauth_review'
    AND a.status <> 'resolved'
    AND a.deleted_at IS NULL
);

WITH app_seed AS (
  SELECT pa.id AS provider_app_id, pa.platform_id FROM integration_provider_apps pa WHERE pa.app_code = 'wecom-suite-standard' AND pa.deleted_at IS NULL LIMIT 1
)
INSERT INTO integration_api_call_logs (
  platform_id, provider_app_id, request_id, call_type, method, endpoint, status,
  http_status, duration_ms, request_digest, response_digest, called_at, created_at
)
SELECT app_seed.platform_id, app_seed.provider_app_id, 'seed-wecom-contact-list', 'third_party_api',
  'GET', '/cgi-bin/department/list', 'success', 200, 96,
  'sha256:seed-request', 'sha256:seed-response', now() - interval '10 minutes', now()
FROM app_seed
WHERE NOT EXISTS (
  SELECT 1 FROM integration_api_call_logs l WHERE l.request_id = 'seed-wecom-contact-list'
);
