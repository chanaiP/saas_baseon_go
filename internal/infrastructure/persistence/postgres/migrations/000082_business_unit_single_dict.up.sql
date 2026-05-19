WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
)
INSERT INTO dict_type (
  tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
)
SELECT
  id,
  'business_unit',
  '业务单元',
  '一级为业务单元类型，二级为业务单元分组；业务单元功能唯一使用此字典',
  'platform',
  true,
  false,
  now(),
  now(),
  NULL
FROM platform_tenant
ON CONFLICT (tenant_id, code) DO UPDATE
SET
  name = EXCLUDED.name,
  remark = EXCLUDED.remark,
  scope = EXCLUDED.scope,
  tenant_editable = EXCLUDED.tenant_editable,
  is_platform_only = EXCLUDED.is_platform_only,
  updated_at = now(),
  deleted_at = NULL;

WITH target_type AS (
  SELECT id AS dict_type_id, tenant_id
  FROM dict_type
  WHERE code = 'business_unit'
    AND deleted_at IS NULL
),
legacy_roots AS (
  SELECT DISTINCT tt.tenant_id, tt.dict_type_id, di.label, di.value, di.sort_order
  FROM target_type tt
  JOIN dict_type old_dt
    ON old_dt.tenant_id = tt.tenant_id
   AND old_dt.code = 'business_unit_tree'
   AND old_dt.deleted_at IS NULL
  JOIN dict_item di
    ON di.dict_type_id = old_dt.id
   AND di.parent_id IS NULL
   AND di.deleted_at IS NULL
   AND di.enabled = true
),
existing_roots AS (
  SELECT DISTINCT
    tt.tenant_id,
    tt.dict_type_id,
    COALESCE(NULLIF(bu.unit_type_name, ''), bu.unit_type_code) AS label,
    bu.unit_type_code AS value,
    900 AS sort_order
  FROM target_type tt
  JOIN business_unit bu
    ON bu.unit_type_code IS NOT NULL
   AND bu.unit_type_code <> ''
   AND bu.deleted_at IS NULL
),
seed_roots AS (
  SELECT tt.tenant_id, tt.dict_type_id, seed.label, seed.value, seed.sort_order
  FROM target_type tt
  CROSS JOIN (
    VALUES
      ('门店', 'store', 10),
      ('仓库', 'warehouse', 20),
      ('项目', 'project', 30),
      ('业务线', 'business_line', 40),
      ('品牌', 'brand', 50),
      ('投放账号', 'ad_account', 60),
      ('内容账号', 'content_account', 70)
  ) AS seed(label, value, sort_order)
),
merged_roots AS (
  SELECT * FROM legacy_roots
  UNION ALL
  SELECT * FROM existing_roots
  UNION ALL
  SELECT * FROM seed_roots
),
ranked_roots AS (
  SELECT
    mr.*,
    row_number() OVER (
      PARTITION BY mr.tenant_id, mr.dict_type_id, mr.value
      ORDER BY mr.sort_order ASC, mr.label ASC
    ) AS rn
  FROM merged_roots mr
)
INSERT INTO dict_item (
  tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT mr.tenant_id, mr.dict_type_id, mr.label, mr.value, mr.sort_order, true, now(), now(), NULL
FROM ranked_roots mr
WHERE mr.value IS NOT NULL
  AND mr.value <> ''
  AND mr.rn = 1
  AND NOT EXISTS (
    SELECT 1
    FROM dict_item di
    WHERE di.tenant_id = mr.tenant_id
      AND di.dict_type_id = mr.dict_type_id
      AND di.value = mr.value
      AND di.parent_id IS NULL
      AND di.deleted_at IS NULL
  );

WITH target_type AS (
  SELECT id AS dict_type_id, tenant_id
  FROM dict_type
  WHERE code = 'business_unit'
    AND deleted_at IS NULL
),
legacy_children AS (
  SELECT DISTINCT
    tt.tenant_id,
    tt.dict_type_id,
    parent_new.id AS parent_id,
    child.label,
    child.value,
    child.sort_order
  FROM target_type tt
  JOIN dict_type old_dt
    ON old_dt.tenant_id = tt.tenant_id
   AND old_dt.code = 'business_unit_tree'
   AND old_dt.deleted_at IS NULL
  JOIN dict_item old_parent
    ON old_parent.dict_type_id = old_dt.id
   AND old_parent.parent_id IS NULL
   AND old_parent.deleted_at IS NULL
  JOIN dict_item child
    ON child.dict_type_id = old_dt.id
   AND child.parent_id = old_parent.id
   AND child.deleted_at IS NULL
   AND child.enabled = true
  JOIN dict_item parent_new
    ON parent_new.dict_type_id = tt.dict_type_id
   AND parent_new.value = old_parent.value
   AND parent_new.parent_id IS NULL
   AND parent_new.deleted_at IS NULL
),
existing_children AS (
  SELECT DISTINCT
    tt.tenant_id,
    tt.dict_type_id,
    parent_new.id AS parent_id,
    COALESCE(NULLIF(bu.unit_group_name, ''), bu.unit_group_code) AS label,
    bu.unit_group_code AS value,
    900 AS sort_order
  FROM target_type tt
  JOIN business_unit bu
    ON bu.unit_type_code IS NOT NULL
   AND bu.unit_type_code <> ''
   AND bu.unit_group_code IS NOT NULL
   AND bu.unit_group_code <> ''
   AND bu.deleted_at IS NULL
  JOIN dict_item parent_new
    ON parent_new.dict_type_id = tt.dict_type_id
   AND parent_new.value = bu.unit_type_code
   AND parent_new.parent_id IS NULL
   AND parent_new.deleted_at IS NULL
),
seed_children AS (
  SELECT tt.tenant_id, tt.dict_type_id, parent_new.id AS parent_id, seed.label, seed.value, seed.sort_order
  FROM target_type tt
  CROSS JOIN (
    VALUES
      ('store', '抖音', 'douyin', 11),
      ('store', '京东', 'jd', 12),
      ('store', '小红书', 'redbook', 13),
      ('store', '天猫', 'tmall', 14),
      ('store', '线下门店', 'offline', 15),
      ('warehouse', '华东', 'east_china', 21),
      ('warehouse', '华南', 'south_china', 22),
      ('project', '运营', 'operation', 31),
      ('project', '交付', 'delivery', 32),
      ('business_line', '电商', 'ecommerce', 41),
      ('business_line', '零售', 'retail', 42),
      ('brand', '运营', 'operation', 51),
      ('ad_account', '巨量千川', 'qianchuan', 61),
      ('content_account', '抖音', 'douyin', 71),
      ('content_account', '小红书', 'redbook', 72)
  ) AS seed(parent_value, label, value, sort_order)
  JOIN dict_item parent_new
    ON parent_new.dict_type_id = tt.dict_type_id
   AND parent_new.value = seed.parent_value
   AND parent_new.parent_id IS NULL
   AND parent_new.deleted_at IS NULL
),
merged_children AS (
  SELECT * FROM legacy_children
  UNION ALL
  SELECT * FROM existing_children
  UNION ALL
  SELECT * FROM seed_children
),
ranked_children AS (
  SELECT
    mc.*,
    row_number() OVER (
      PARTITION BY mc.tenant_id, mc.dict_type_id, mc.parent_id, mc.value
      ORDER BY mc.sort_order ASC, mc.label ASC
    ) AS rn
  FROM merged_children mc
)
INSERT INTO dict_item (
  tenant_id, dict_type_id, parent_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT mc.tenant_id, mc.dict_type_id, mc.parent_id, mc.label, mc.value, mc.sort_order, true, now(), now(), NULL
FROM ranked_children mc
WHERE mc.value IS NOT NULL
  AND mc.value <> ''
  AND mc.rn = 1
  AND NOT EXISTS (
    SELECT 1
    FROM dict_item di
    WHERE di.tenant_id = mc.tenant_id
      AND di.dict_type_id = mc.dict_type_id
      AND di.parent_id = mc.parent_id
      AND di.value = mc.value
      AND di.deleted_at IS NULL
  );

WITH old_types AS (
  SELECT id
  FROM dict_type
  WHERE code IN (
    'business_unit_tree',
    'business_unit_type',
    'base_business.unit_scenario',
    'base_business.unit_form'
  )
    AND deleted_at IS NULL
)
UPDATE dict_item
SET deleted_at = now(), updated_at = now()
WHERE dict_type_id IN (SELECT id FROM old_types)
  AND deleted_at IS NULL;

UPDATE dict_type
SET
  code = code || '__deprecated__' || id::text,
  name = name || '（已废弃）',
  deleted_at = now(),
  updated_at = now()
WHERE code IN (
  'business_unit_tree',
  'business_unit_type',
  'base_business.unit_scenario',
  'base_business.unit_form'
)
  AND deleted_at IS NULL;
