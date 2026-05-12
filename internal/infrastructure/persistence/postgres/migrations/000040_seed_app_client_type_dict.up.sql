WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
canonical_type AS (
  SELECT dt.id, dt.tenant_id
  FROM dict_type dt
  JOIN platform_tenant pt ON pt.id = dt.tenant_id
  WHERE dt.code = 'app_client_type'
  ORDER BY dt.id ASC
  LIMIT 1
),
inserted_type AS (
  INSERT INTO dict_type (
    tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
  )
  SELECT
    pt.id,
    'app_client_type',
    '应用客户端类型',
    '应用中心可选客户端形态',
    'platform',
    false,
    true,
    now(),
    now(),
    NULL
  FROM platform_tenant pt
  WHERE NOT EXISTS (SELECT 1 FROM canonical_type)
  RETURNING id, tenant_id
),
target_type AS (
  SELECT id, tenant_id FROM canonical_type
  UNION ALL
  SELECT id, tenant_id FROM inserted_type
)
UPDATE dict_type dt
SET
  name = '应用客户端类型',
  remark = '应用中心可选客户端形态',
  scope = 'platform',
  tenant_editable = false,
  is_platform_only = true,
  deleted_at = NULL,
  updated_at = now()
FROM target_type tt
WHERE dt.id = tt.id;

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
target_type AS (
  SELECT dt.id, dt.tenant_id
  FROM dict_type dt
  JOIN platform_tenant pt ON pt.id = dt.tenant_id
  WHERE dt.code = 'app_client_type'
  ORDER BY dt.id ASC
  LIMIT 1
),
seed(label, value, sort_order) AS (
  VALUES
    ('PC Web', 'PC_WEB', 10),
    ('API Only', 'API_ONLY', 20),
    ('H5', 'H5', 30),
    ('iOS', 'IOS', 40),
    ('Android', 'ANDROID', 50),
    ('鸿蒙', 'HARMONYOS', 60),
    ('Windows', 'WINDOWS', 70),
    ('macOS', 'MACOS', 80),
    ('小程序', 'MINIAPP', 90),
    ('企业微信', 'WECHAT', 100),
    ('钉钉', 'DINGTALK', 110),
    ('飞书', 'FEISHU', 120)
)
INSERT INTO dict_item (
  tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT tt.tenant_id, tt.id, seed.label, seed.value, seed.sort_order, true, now(), now(), NULL
FROM target_type tt
CROSS JOIN seed
WHERE NOT EXISTS (
  SELECT 1
  FROM dict_item di
  WHERE di.dict_type_id = tt.id
    AND di.value = seed.value
    AND di.deleted_at IS NULL
);

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
target_type AS (
  SELECT dt.id
  FROM dict_type dt
  JOIN platform_tenant pt ON pt.id = dt.tenant_id
  WHERE dt.code = 'app_client_type'
  ORDER BY dt.id ASC
  LIMIT 1
),
seed(label, value, sort_order) AS (
  VALUES
    ('PC Web', 'PC_WEB', 10),
    ('API Only', 'API_ONLY', 20),
    ('H5', 'H5', 30),
    ('iOS', 'IOS', 40),
    ('Android', 'ANDROID', 50),
    ('鸿蒙', 'HARMONYOS', 60),
    ('Windows', 'WINDOWS', 70),
    ('macOS', 'MACOS', 80),
    ('小程序', 'MINIAPP', 90),
    ('企业微信', 'WECHAT', 100),
    ('钉钉', 'DINGTALK', 110),
    ('飞书', 'FEISHU', 120)
)
UPDATE dict_item di
SET
  label = seed.label,
  sort_order = seed.sort_order,
  enabled = true,
  deleted_at = NULL,
  updated_at = now()
FROM target_type tt
CROSS JOIN seed
WHERE di.dict_type_id = tt.id
  AND di.value = seed.value;
