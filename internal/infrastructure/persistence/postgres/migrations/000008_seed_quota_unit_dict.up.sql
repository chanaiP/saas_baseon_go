INSERT INTO dict_type (
  tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
)
SELECT
  t.id,
  'quota_unit',
  '配额单位',
  '套餐配额值的展示单位',
  'HYBRID',
  true,
  false,
  now(),
  now(),
  NULL
FROM tenant t
WHERE t.deleted_at IS NULL
ON CONFLICT (tenant_id, code) DO UPDATE
SET
  name = EXCLUDED.name,
  remark = EXCLUDED.remark,
  scope = EXCLUDED.scope,
  tenant_editable = EXCLUDED.tenant_editable,
  is_platform_only = EXCLUDED.is_platform_only,
  updated_at = now(),
  deleted_at = NULL;

INSERT INTO dict_item (
  tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT dt.tenant_id, dt.id, v.label, v.value, v.sort_order, true, now(), now(), NULL
FROM dict_type dt
CROSS JOIN (
  VALUES
    ('数量', 'COUNT', 1),
    ('MB', 'MB', 2),
    ('GB', 'GB', 3),
    ('次', 'TIMES', 4),
    ('个', 'ITEM', 5)
) AS v(label, value, sort_order)
WHERE dt.code = 'quota_unit'
  AND dt.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM dict_item di
    WHERE di.tenant_id = dt.tenant_id
      AND di.dict_type_id = dt.id
      AND di.value = v.value
      AND di.deleted_at IS NULL
  );
