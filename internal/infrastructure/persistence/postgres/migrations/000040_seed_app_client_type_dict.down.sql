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
)
DELETE FROM dict_item di
USING target_type tt
WHERE di.dict_type_id = tt.id
  AND di.value IN (
    'PC_WEB',
    'API_ONLY',
    'H5',
    'IOS',
    'ANDROID',
    'HARMONYOS',
    'WINDOWS',
    'MACOS',
    'MINIAPP',
    'WECHAT',
    'DINGTALK',
    'FEISHU'
  );

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
)
DELETE FROM dict_type dt
USING platform_tenant pt
WHERE dt.tenant_id = pt.id
  AND dt.code = 'app_client_type'
  AND NOT EXISTS (
    SELECT 1 FROM dict_item di WHERE di.dict_type_id = dt.id
  );
