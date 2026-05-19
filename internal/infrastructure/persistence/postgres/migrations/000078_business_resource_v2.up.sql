ALTER TABLE business_resource
  ADD COLUMN IF NOT EXISTS unit_type_code varchar(64),
  ADD COLUMN IF NOT EXISTS unit_type_name varchar(64),
  ADD COLUMN IF NOT EXISTS business_unit_code varchar(64),
  ADD COLUMN IF NOT EXISTS business_unit_name varchar(64);

UPDATE business_resource
SET
  unit_type_code = COALESCE(NULLIF(unit_type_code, ''), resource_category),
  unit_type_name = COALESCE(NULLIF(unit_type_name, ''), resource_category),
  business_unit_code = COALESCE(NULLIF(business_unit_code, ''), resource_type),
  business_unit_name = COALESCE(NULLIF(business_unit_name, ''), resource_type)
WHERE unit_type_code IS NULL
   OR unit_type_name IS NULL
   OR business_unit_code IS NULL
   OR business_unit_name IS NULL;

ALTER TABLE business_resource
  ALTER COLUMN unit_type_code SET NOT NULL,
  ALTER COLUMN unit_type_name SET NOT NULL,
  ALTER COLUMN business_unit_code SET NOT NULL,
  ALTER COLUMN business_unit_name SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_business_resource_unit_tree
  ON business_resource (tenant_id, unit_type_code, business_unit_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS business_resource_actor (
  id bigserial PRIMARY KEY,
  tenant_id bigint NOT NULL,
  resource_id bigint NOT NULL,
  actor_type varchar(32) NOT NULL,
  actor_id bigint NOT NULL,
  role_type varchar(32) NOT NULL,
  include_children boolean NOT NULL DEFAULT false,
  start_date date,
  end_date date,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_business_resource_actor_resource
  ON business_resource_actor (tenant_id, resource_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_resource_actor_actor
  ON business_resource_actor (tenant_id, actor_type, actor_id)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_resource_actor_unique
  ON business_resource_actor (tenant_id, resource_id, actor_type, actor_id, role_type)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS business_resource_field_config (
  id bigserial PRIMARY KEY,
  tenant_id bigint,
  unit_type_code varchar(64) NOT NULL,
  business_unit_code varchar(64) NOT NULL,
  field_key varchar(128) NOT NULL,
  field_label varchar(128) NOT NULL,
  field_type varchar(32) NOT NULL,
  dict_code varchar(128),
  relation_unit_type_code varchar(64),
  relation_business_unit_code varchar(64),
  required boolean NOT NULL DEFAULT false,
  default_value text,
  placeholder varchar(256),
  help_text varchar(256),
  validation_rule jsonb,
  show_in_list boolean NOT NULL DEFAULT false,
  show_in_detail boolean NOT NULL DEFAULT true,
  show_in_import boolean NOT NULL DEFAULT true,
  import_required boolean NOT NULL DEFAULT false,
  sort_order int NOT NULL DEFAULT 0,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_resource_field_config_unique
  ON business_resource_field_config (
    COALESCE(tenant_id, 0),
    unit_type_code,
    business_unit_code,
    field_key
  )
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_resource_field_config_scope
  ON business_resource_field_config (tenant_id, unit_type_code, business_unit_code)
  WHERE deleted_at IS NULL;

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
tree_type AS (
  INSERT INTO dict_type (
    tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
  )
  SELECT
    id,
    'business_unit_tree',
    '业务单元类型',
    '业务单元类型与业务单元三级树',
    'platform',
    false,
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
    deleted_at = NULL
  RETURNING id, tenant_id
),
seed_item(label, value, sort_order, parent_value) AS (
  VALUES
    ('店铺门店', 'store', 10, NULL),
    ('线上店铺', 'online_store', 101, 'store'),
    ('平台店铺', 'platform_store', 102, 'store'),
    ('线下门店', 'offline_store', 103, 'store'),
    ('投放广告', 'ad', 20, NULL),
    ('巨量千川', 'qianchuan', 201, 'ad'),
    ('阿里妈妈', 'alimama', 202, 'ad'),
    ('京准通', 'jingzhuntong', 203, 'ad'),
    ('小红书聚光', 'xiaohongshu_juguang', 204, 'ad'),
    ('快手磁力金牛', 'kuaishou_ad', 205, 'ad'),
    ('腾讯广告', 'tencent_ad', 206, 'ad'),
    ('内容直播', 'content', 30, NULL),
    ('抖音内容', 'douyin_content', 301, 'content'),
    ('小红书内容', 'xiaohongshu_content', 302, 'content'),
    ('视频号内容', 'wechat_video', 303, 'content'),
    ('快手内容', 'kuaishou_content', 304, 'content'),
    ('直播账号', 'live_account', 305, 'content'),
    ('达人账号', 'kol_account', 306, 'content'),
    ('供应履约', 'supply', 40, NULL),
    ('仓库', 'warehouse', 401, 'supply'),
    ('供应商', 'supplier', 402, 'supply'),
    ('物流商', 'logistics', 403, 'supply'),
    ('采购组织', 'purchase_org', 404, 'supply'),
    ('库存组织', 'inventory_org', 405, 'supply'),
    ('财务结算', 'finance', 50, NULL),
    ('经营主体', 'operating_entity', 501, 'finance'),
    ('结算主体', 'settlement_entity', 502, 'finance'),
    ('结算账户', 'settlement_account', 503, 'finance'),
    ('成本中心', 'cost_center', 504, 'finance'),
    ('利润中心', 'profit_center', 505, 'finance')
),
target_item AS (
  SELECT tt.tenant_id, tt.id AS dict_type_id, si.label, si.value, si.sort_order, si.parent_value
  FROM tree_type tt
  CROSS JOIN seed_item si
)
INSERT INTO dict_item (
  tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT ti.tenant_id, ti.dict_type_id, ti.label, ti.value, ti.sort_order, true, now(), now(), NULL
FROM target_item ti
WHERE NOT EXISTS (
  SELECT 1
  FROM dict_item di
  WHERE di.tenant_id = ti.tenant_id
    AND di.dict_type_id = ti.dict_type_id
    AND di.value = ti.value
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
dt AS (
  SELECT dict_type.id, dict_type.tenant_id
  FROM dict_type
  JOIN platform_tenant ON platform_tenant.id = dict_type.tenant_id
  WHERE dict_type.code = 'business_unit_tree'
    AND dict_type.deleted_at IS NULL
),
parents AS (
  SELECT di.id, di.value
  FROM dict_item di
  JOIN dt ON dt.id = di.dict_type_id AND dt.tenant_id = di.tenant_id
  WHERE di.parent_id IS NULL
    AND di.deleted_at IS NULL
),
child_parent(value, parent_value) AS (
  VALUES
    ('online_store', 'store'),
    ('platform_store', 'store'),
    ('offline_store', 'store'),
    ('qianchuan', 'ad'),
    ('alimama', 'ad'),
    ('jingzhuntong', 'ad'),
    ('xiaohongshu_juguang', 'ad'),
    ('kuaishou_ad', 'ad'),
    ('tencent_ad', 'ad'),
    ('douyin_content', 'content'),
    ('xiaohongshu_content', 'content'),
    ('wechat_video', 'content'),
    ('kuaishou_content', 'content'),
    ('live_account', 'content'),
    ('kol_account', 'content'),
    ('warehouse', 'supply'),
    ('supplier', 'supply'),
    ('logistics', 'supply'),
    ('purchase_org', 'supply'),
    ('inventory_org', 'supply'),
    ('operating_entity', 'finance'),
    ('settlement_entity', 'finance'),
    ('settlement_account', 'finance'),
    ('cost_center', 'finance'),
    ('profit_center', 'finance')
)
UPDATE dict_item di
SET parent_id = p.id, updated_at = now()
FROM dt, child_parent cp, parents p
WHERE di.dict_type_id = dt.id
  AND di.tenant_id = dt.tenant_id
  AND cp.value = di.value
  AND p.value = cp.parent_value
  AND di.deleted_at IS NULL;

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
seed_type(code, name, remark, sort_order) AS (
  VALUES
    ('business_resource.actor_type', '资源责任方类型', '业务资源负责人类型', 100),
    ('business_resource.actor_role', '资源责任角色', '业务资源负责人角色', 110),
    ('business_resource.field_type', '资源字段类型', '业务资源个性化字段类型', 120)
)
INSERT INTO dict_type (
  tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
)
SELECT pt.id, st.code, st.name, st.remark, 'platform', false, false, now(), now(), NULL
FROM platform_tenant pt
CROSS JOIN seed_type st
ON CONFLICT (tenant_id, code) DO UPDATE
SET name = EXCLUDED.name, remark = EXCLUDED.remark, updated_at = now(), deleted_at = NULL;

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
seed_item(type_code, label, value, sort_order) AS (
  VALUES
    ('business_resource.actor_type', '组织节点', 'org_node', 10),
    ('business_resource.actor_type', '用户', 'user', 20),
    ('business_resource.actor_role', '负责人', 'owner', 10),
    ('business_resource.actor_role', '管理员', 'manager', 20),
    ('business_resource.actor_role', '运营人员', 'operator', 30),
    ('business_resource.actor_role', '数据分析', 'analyst', 40),
    ('business_resource.actor_role', '只读人员', 'viewer', 50),
    ('business_resource.actor_role', '协作人员', 'support', 60),
    ('business_resource.field_type', '单行文本', 'input', 10),
    ('business_resource.field_type', '多行文本', 'textarea', 20),
    ('business_resource.field_type', '数字', 'number', 30),
    ('business_resource.field_type', '日期', 'date', 40),
    ('business_resource.field_type', '下拉选择', 'select', 50),
    ('business_resource.field_type', '字典选择', 'dict', 60),
    ('business_resource.field_type', '关联资源', 'relation', 70),
    ('business_resource.field_type', '开关', 'switch', 80)
),
target_item AS (
  SELECT dt.tenant_id, dt.id AS dict_type_id, si.label, si.value, si.sort_order
  FROM dict_type dt
  JOIN platform_tenant pt ON pt.id = dt.tenant_id
  JOIN seed_item si ON si.type_code = dt.code
  WHERE dt.deleted_at IS NULL
)
INSERT INTO dict_item (
  tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at
)
SELECT ti.tenant_id, ti.dict_type_id, ti.label, ti.value, ti.sort_order, true, now(), now(), NULL
FROM target_item ti
WHERE NOT EXISTS (
  SELECT 1
  FROM dict_item di
  WHERE di.tenant_id = ti.tenant_id
    AND di.dict_type_id = ti.dict_type_id
    AND di.value = ti.value
    AND di.deleted_at IS NULL
);

WITH defaults(unit_type_code, business_unit_code, field_key, field_label, field_type, sort_order, required, show_in_list) AS (
  VALUES
    ('store', 'platform_store', 'brand_name', '品牌名称', 'input', 10, true, true),
    ('store', 'platform_store', 'platform_store_type', '店铺类型', 'input', 20, true, true),
    ('store', 'platform_store', 'region', '经营区域', 'input', 30, false, true),
    ('store', 'platform_store', 'auth_status', '授权状态', 'dict', 40, false, true),
    ('ad', 'qianchuan', 'ad_account_id', '广告账号ID', 'input', 10, true, true),
    ('ad', 'qianchuan', 'account_balance', '账户余额', 'number', 20, false, true),
    ('ad', 'qianchuan', 'daily_budget', '日预算', 'number', 30, false, true),
    ('ad', 'qianchuan', 'auth_status', '授权状态', 'dict', 40, false, true),
    ('content', 'douyin_content', 'account_id', '抖音号ID', 'input', 10, true, true),
    ('content', 'douyin_content', 'fans_count', '粉丝数', 'number', 20, false, true),
    ('content', 'douyin_content', 'is_official', '是否官方号', 'switch', 30, false, true),
    ('supply', 'warehouse', 'warehouse_address', '仓库地址', 'input', 10, false, true),
    ('finance', 'settlement_account', 'bank_name', '开户行', 'input', 10, false, true)
)
INSERT INTO business_resource_field_config (
  tenant_id, unit_type_code, business_unit_code, field_key, field_label, field_type,
  dict_code, required, show_in_list, show_in_detail, show_in_import, import_required,
  sort_order, status, created_at, updated_at
)
SELECT
  NULL,
  d.unit_type_code,
  d.business_unit_code,
  d.field_key,
  d.field_label,
  d.field_type,
  CASE WHEN d.field_key = 'auth_status' THEN 'business_resource.resource_status' ELSE NULL END,
  d.required,
  d.show_in_list,
  true,
  true,
  d.required,
  d.sort_order,
  'active',
  now(),
  now()
FROM defaults d
ON CONFLICT DO NOTHING;
