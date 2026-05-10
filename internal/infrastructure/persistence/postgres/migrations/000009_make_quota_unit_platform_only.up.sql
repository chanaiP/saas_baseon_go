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
  WHERE dt.code = 'quota_unit'
  ORDER BY dt.id ASC
  LIMIT 1
),
inserted_type AS (
  INSERT INTO dict_type (
    tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
  )
  SELECT
    pt.id,
    'quota_unit',
    '配额单位',
    '套餐配额值的展示单位',
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
  name = '配额单位',
  remark = '套餐配额值的展示单位',
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
  WHERE dt.code = 'quota_unit'
  ORDER BY dt.id ASC
  LIMIT 1
)
INSERT INTO dict_item (
  tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT tt.tenant_id, tt.id, v.label, v.value, v.sort_order, true, now(), now(), NULL
FROM target_type tt
CROSS JOIN (
  VALUES
    ('数量', 'COUNT', 1),
    ('MB', 'MB', 2),
    ('GB', 'GB', 3),
    ('次', 'TIMES', 4),
    ('个', 'ITEM', 5)
) AS v(label, value, sort_order)
WHERE NOT EXISTS (
  SELECT 1
  FROM dict_item di
  WHERE di.dict_type_id = tt.id
    AND di.value = v.value
    AND di.deleted_at IS NULL
);

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
)
UPDATE dict_type dt
SET
  deleted_at = now(),
  updated_at = now()
FROM platform_tenant pt
WHERE dt.code = 'quota_unit'
  AND dt.tenant_id <> pt.id
  AND dt.deleted_at IS NULL;
