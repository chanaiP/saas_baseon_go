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
    ('base_business.unit_scenario', '业务单元场景', '业务单元服务的业务目的', 10),
    ('base_business.unit_form', '业务单元形态', '业务单元按什么对象组织', 20),
    ('business_resource.resource_category', '业务资源大类', '业务资源一级分组', 30),
    ('business_resource.resource_type', '业务资源类型', '业务资源具体业务对象，使用 parent_id 关联一级分组', 40),
    ('business_resource.source_mode', '业务资源来源模式', '业务资源原生创建或引用已有主数据', 50),
    ('business_resource.unit_resource_relation_type', '业务单元资源关系类型', '业务单元与资源之间的归属、权限和分析关系', 60),
    ('business_resource.resource_relation_type', '业务资源关系类型', '业务资源之间的业务连接关系', 70),
    ('business_resource.resource_status', '业务资源状态', '业务资源生命周期状态', 80),
    ('base.status', '基础状态', '通用启停状态', 90)
)
INSERT INTO dict_type (
  tenant_id, code, name, remark, scope, tenant_editable, is_platform_only, created_at, updated_at, deleted_at
)
SELECT
  pt.id,
  st.code,
  st.name,
  st.remark,
  'platform',
  false,
  false,
  now(),
  now(),
  NULL
FROM platform_tenant pt
CROSS JOIN seed_type st
ON CONFLICT (tenant_id, code) DO UPDATE
SET
  name = EXCLUDED.name,
  remark = EXCLUDED.remark,
  scope = EXCLUDED.scope,
  tenant_editable = EXCLUDED.tenant_editable,
  is_platform_only = EXCLUDED.is_platform_only,
  updated_at = now(),
  deleted_at = NULL;

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
seed_item(type_code, label, value, sort_order, parent_value) AS (
  VALUES
    ('base_business.unit_scenario', '经营场景', 'operation', 10, NULL),
    ('base_business.unit_scenario', '结算场景', 'settlement', 20, NULL),
    ('base_business.unit_scenario', '库存场景', 'inventory', 30, NULL),
    ('base_business.unit_scenario', '采购场景', 'procurement', 40, NULL),
    ('base_business.unit_scenario', '销售渠道场景', 'sales_channel', 50, NULL),
    ('base_business.unit_scenario', '营销投放场景', 'marketing', 60, NULL),
    ('base_business.unit_scenario', '内容运营场景', 'content', 70, NULL),
    ('base_business.unit_scenario', '客服售后场景', 'customer_service', 80, NULL),
    ('base_business.unit_scenario', '财务场景', 'finance', 90, NULL),
    ('base_business.unit_scenario', '数据权限场景', 'data_scope', 100, NULL),
    ('base_business.unit_scenario', '项目协同场景', 'project', 110, NULL),

    ('base_business.unit_form', '品牌型', 'brand', 10, NULL),
    ('base_business.unit_form', '区域型', 'region', 20, NULL),
    ('base_business.unit_form', '门店组', 'store_group', 30, NULL),
    ('base_business.unit_form', '线上店铺组', 'online_store', 40, NULL),
    ('base_business.unit_form', '平台型', 'platform', 50, NULL),
    ('base_business.unit_form', '仓库组', 'warehouse_group', 60, NULL),
    ('base_business.unit_form', '供应商组', 'supplier_group', 70, NULL),
    ('base_business.unit_form', '商品线', 'product_line', 80, NULL),
    ('base_business.unit_form', '团队型', 'team', 90, NULL),
    ('base_business.unit_form', '自定义', 'custom', 100, NULL),

    ('business_resource.resource_category', '门店类', 'store', 10, NULL),
    ('business_resource.resource_category', '账号类', 'account', 20, NULL),
    ('business_resource.resource_category', '投放类', 'ad', 30, NULL),
    ('business_resource.resource_category', '内容类', 'content', 40, NULL),
    ('business_resource.resource_category', '供应链类', 'supply', 50, NULL),
    ('business_resource.resource_category', '商品类', 'product', 60, NULL),
    ('business_resource.resource_category', '财务类', 'finance', 70, NULL),
    ('business_resource.resource_category', '组织类', 'org', 80, NULL),
    ('business_resource.resource_category', '集成类', 'integration', 90, NULL),

    ('business_resource.resource_type', '门店类', 'store', 10, NULL),
    ('business_resource.resource_type', '线下门店', 'offline_store', 11, 'store'),
    ('business_resource.resource_type', '内部线上店铺', 'online_store', 12, 'store'),
    ('business_resource.resource_type', '抖音店铺', 'douyin_shop', 13, 'store'),
    ('business_resource.resource_type', '淘宝店铺', 'taobao_shop', 14, 'store'),
    ('business_resource.resource_type', '天猫店铺', 'tmall_shop', 15, 'store'),
    ('business_resource.resource_type', '京东店铺', 'jd_shop', 16, 'store'),
    ('business_resource.resource_type', '拼多多店铺', 'pdd_shop', 17, 'store'),
    ('business_resource.resource_type', '快手店铺', 'kuaishou_shop', 18, 'store'),
    ('business_resource.resource_type', '微信小店', 'wechat_shop', 19, 'store'),

    ('business_resource.resource_type', '账号类', 'account', 20, NULL),
    ('business_resource.resource_type', '店铺授权账号', 'shop_account', 21, 'account'),
    ('business_resource.resource_type', 'API账号', 'api_account', 22, 'account'),
    ('business_resource.resource_type', '服务账号', 'service_account', 23, 'account'),

    ('business_resource.resource_type', '投放类', 'ad', 30, NULL),
    ('business_resource.resource_type', '抖音投放账号', 'douyin_ad_account', 31, 'ad'),
    ('business_resource.resource_type', '腾讯广告账号', 'tencent_ad_account', 32, 'ad'),
    ('business_resource.resource_type', '快手广告账号', 'kuaishou_ad_account', 33, 'ad'),
    ('business_resource.resource_type', '小红书广告账号', 'xiaohongshu_ad_account', 34, 'ad'),
    ('business_resource.resource_type', '广告计划', 'ad_campaign', 35, 'ad'),

    ('business_resource.resource_type', '内容类', 'content', 40, NULL),
    ('business_resource.resource_type', '抖音官方号', 'douyin_official_account', 41, 'content'),
    ('business_resource.resource_type', '抖音内容号', 'douyin_media_account', 42, 'content'),
    ('business_resource.resource_type', '小红书账号', 'xiaohongshu_account', 43, 'content'),
    ('business_resource.resource_type', '快手内容号', 'kuaishou_media_account', 44, 'content'),
    ('business_resource.resource_type', '视频号账号', 'wechat_video_account', 45, 'content'),
    ('business_resource.resource_type', '直播账号', 'live_account', 46, 'content'),
    ('business_resource.resource_type', '达人账号', 'kol_account', 47, 'content'),

    ('business_resource.resource_type', '供应链类', 'supply', 50, NULL),
    ('business_resource.resource_type', '仓库', 'warehouse', 51, 'supply'),
    ('business_resource.resource_type', '供应商', 'supplier', 52, 'supply'),
    ('business_resource.resource_type', '物流商', 'logistics', 53, 'supply'),
    ('business_resource.resource_type', '采购组织', 'purchase_org', 54, 'supply'),
    ('business_resource.resource_type', '库存组织', 'inventory_org', 55, 'supply'),

    ('business_resource.resource_type', '商品类', 'product', 60, NULL),
    ('business_resource.resource_type', '品牌', 'brand', 61, 'product'),
    ('business_resource.resource_type', '类目', 'category', 62, 'product'),
    ('business_resource.resource_type', '商品线', 'product_line', 63, 'product'),
    ('business_resource.resource_type', 'SPU', 'spu', 64, 'product'),
    ('business_resource.resource_type', 'SKU', 'sku', 65, 'product'),

    ('business_resource.resource_type', '财务类', 'finance', 70, NULL),
    ('business_resource.resource_type', '结算主体', 'settlement_entity', 71, 'finance'),
    ('business_resource.resource_type', '结算账户', 'settlement_account', 72, 'finance'),
    ('business_resource.resource_type', '成本中心', 'cost_center', 73, 'finance'),
    ('business_resource.resource_type', '利润中心', 'profit_center', 74, 'finance'),

    ('business_resource.resource_type', '组织类', 'org', 80, NULL),
    ('business_resource.resource_type', '组织节点', 'org_node', 81, 'org'),
    ('business_resource.resource_type', '用户', 'user', 82, 'org'),
    ('business_resource.resource_type', '角色', 'role', 83, 'org'),
    ('business_resource.resource_type', '岗位', 'position', 84, 'org'),

    ('business_resource.resource_type', '集成类', 'integration', 90, NULL),
    ('business_resource.resource_type', '连接实例', 'connection_instance', 91, 'integration'),
    ('business_resource.resource_type', '外部应用', 'external_app', 92, 'integration'),
    ('business_resource.resource_type', '外部API账号', 'external_api_account', 93, 'integration'),

    ('business_resource.source_mode', '原生资源', 'native', 10, NULL),
    ('business_resource.source_mode', '引用资源', 'reference', 20, NULL),

    ('business_resource.unit_resource_relation_type', '主负责', 'owner', 10, NULL),
    ('business_resource.unit_resource_relation_type', '运营负责', 'operator', 20, NULL),
    ('business_resource.unit_resource_relation_type', '可查看', 'viewer', 30, NULL),
    ('business_resource.unit_resource_relation_type', '结算归属', 'settlement', 40, NULL),
    ('business_resource.unit_resource_relation_type', '成本归属', 'cost', 50, NULL),
    ('business_resource.unit_resource_relation_type', '库存归属', 'inventory', 60, NULL),
    ('business_resource.unit_resource_relation_type', '采购归属', 'procurement', 70, NULL),
    ('business_resource.unit_resource_relation_type', '支持协作', 'support', 80, NULL),

    ('business_resource.resource_relation_type', '包含', 'contains', 10, NULL),
    ('business_resource.resource_relation_type', '归属', 'belongs_to', 20, NULL),
    ('business_resource.resource_relation_type', '官方账号绑定', 'official_account', 30, NULL),
    ('business_resource.resource_relation_type', '投放账号绑定', 'ad_account', 40, NULL),
    ('business_resource.resource_relation_type', '投放推广', 'promotes', 50, NULL),
    ('business_resource.resource_relation_type', '内容引流', 'content_drives', 60, NULL),
    ('business_resource.resource_relation_type', '直播引流', 'live_drives', 70, NULL),
    ('business_resource.resource_relation_type', '授权关系', 'authorized_by', 80, NULL),
    ('business_resource.resource_relation_type', '结算关系', 'settles', 90, NULL),
    ('business_resource.resource_relation_type', '履约关系', 'fulfills', 100, NULL),

    ('business_resource.resource_status', '启用', 'active', 10, NULL),
    ('business_resource.resource_status', '停用', 'disabled', 20, NULL),
    ('business_resource.resource_status', '授权过期', 'expired', 30, NULL),
    ('business_resource.resource_status', '待配置', 'pending', 40, NULL),
    ('business_resource.resource_status', '已归档', 'archived', 50, NULL),

    ('base.status', '启用', 'active', 10, NULL),
    ('base.status', '停用', 'disabled', 20, NULL)
),
target_item AS (
  SELECT dt.tenant_id, dt.id AS dict_type_id, si.label, si.value, si.sort_order, si.parent_value
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

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
  ORDER BY id ASC
  LIMIT 1
),
seed_item(type_code, label, value, sort_order, parent_value) AS (
  VALUES
    ('base_business.unit_scenario', '经营场景', 'operation', 10, NULL),
    ('base_business.unit_scenario', '结算场景', 'settlement', 20, NULL),
    ('base_business.unit_scenario', '库存场景', 'inventory', 30, NULL),
    ('base_business.unit_scenario', '采购场景', 'procurement', 40, NULL),
    ('base_business.unit_scenario', '销售渠道场景', 'sales_channel', 50, NULL),
    ('base_business.unit_scenario', '营销投放场景', 'marketing', 60, NULL),
    ('base_business.unit_scenario', '内容运营场景', 'content', 70, NULL),
    ('base_business.unit_scenario', '客服售后场景', 'customer_service', 80, NULL),
    ('base_business.unit_scenario', '财务场景', 'finance', 90, NULL),
    ('base_business.unit_scenario', '数据权限场景', 'data_scope', 100, NULL),
    ('base_business.unit_scenario', '项目协同场景', 'project', 110, NULL),
    ('base_business.unit_form', '品牌型', 'brand', 10, NULL),
    ('base_business.unit_form', '区域型', 'region', 20, NULL),
    ('base_business.unit_form', '门店组', 'store_group', 30, NULL),
    ('base_business.unit_form', '线上店铺组', 'online_store', 40, NULL),
    ('base_business.unit_form', '平台型', 'platform', 50, NULL),
    ('base_business.unit_form', '仓库组', 'warehouse_group', 60, NULL),
    ('base_business.unit_form', '供应商组', 'supplier_group', 70, NULL),
    ('base_business.unit_form', '商品线', 'product_line', 80, NULL),
    ('base_business.unit_form', '团队型', 'team', 90, NULL),
    ('base_business.unit_form', '自定义', 'custom', 100, NULL),
    ('business_resource.resource_category', '门店类', 'store', 10, NULL),
    ('business_resource.resource_category', '账号类', 'account', 20, NULL),
    ('business_resource.resource_category', '投放类', 'ad', 30, NULL),
    ('business_resource.resource_category', '内容类', 'content', 40, NULL),
    ('business_resource.resource_category', '供应链类', 'supply', 50, NULL),
    ('business_resource.resource_category', '商品类', 'product', 60, NULL),
    ('business_resource.resource_category', '财务类', 'finance', 70, NULL),
    ('business_resource.resource_category', '组织类', 'org', 80, NULL),
    ('business_resource.resource_category', '集成类', 'integration', 90, NULL),
    ('business_resource.source_mode', '原生资源', 'native', 10, NULL),
    ('business_resource.source_mode', '引用资源', 'reference', 20, NULL),
    ('business_resource.unit_resource_relation_type', '主负责', 'owner', 10, NULL),
    ('business_resource.unit_resource_relation_type', '运营负责', 'operator', 20, NULL),
    ('business_resource.unit_resource_relation_type', '可查看', 'viewer', 30, NULL),
    ('business_resource.unit_resource_relation_type', '结算归属', 'settlement', 40, NULL),
    ('business_resource.unit_resource_relation_type', '成本归属', 'cost', 50, NULL),
    ('business_resource.unit_resource_relation_type', '库存归属', 'inventory', 60, NULL),
    ('business_resource.unit_resource_relation_type', '采购归属', 'procurement', 70, NULL),
    ('business_resource.unit_resource_relation_type', '支持协作', 'support', 80, NULL),
    ('business_resource.resource_relation_type', '包含', 'contains', 10, NULL),
    ('business_resource.resource_relation_type', '归属', 'belongs_to', 20, NULL),
    ('business_resource.resource_relation_type', '官方账号绑定', 'official_account', 30, NULL),
    ('business_resource.resource_relation_type', '投放账号绑定', 'ad_account', 40, NULL),
    ('business_resource.resource_relation_type', '投放推广', 'promotes', 50, NULL),
    ('business_resource.resource_relation_type', '内容引流', 'content_drives', 60, NULL),
    ('business_resource.resource_relation_type', '直播引流', 'live_drives', 70, NULL),
    ('business_resource.resource_relation_type', '授权关系', 'authorized_by', 80, NULL),
    ('business_resource.resource_relation_type', '结算关系', 'settles', 90, NULL),
    ('business_resource.resource_relation_type', '履约关系', 'fulfills', 100, NULL),
    ('business_resource.resource_status', '启用', 'active', 10, NULL),
    ('business_resource.resource_status', '停用', 'disabled', 20, NULL),
    ('business_resource.resource_status', '授权过期', 'expired', 30, NULL),
    ('business_resource.resource_status', '待配置', 'pending', 40, NULL),
    ('business_resource.resource_status', '已归档', 'archived', 50, NULL),
    ('base.status', '启用', 'active', 10, NULL),
    ('base.status', '停用', 'disabled', 20, NULL),
    ('business_resource.resource_type', '门店类', 'store', 10, NULL),
    ('business_resource.resource_type', '线下门店', 'offline_store', 11, 'store'),
    ('business_resource.resource_type', '内部线上店铺', 'online_store', 12, 'store'),
    ('business_resource.resource_type', '抖音店铺', 'douyin_shop', 13, 'store'),
    ('business_resource.resource_type', '淘宝店铺', 'taobao_shop', 14, 'store'),
    ('business_resource.resource_type', '天猫店铺', 'tmall_shop', 15, 'store'),
    ('business_resource.resource_type', '京东店铺', 'jd_shop', 16, 'store'),
    ('business_resource.resource_type', '拼多多店铺', 'pdd_shop', 17, 'store'),
    ('business_resource.resource_type', '快手店铺', 'kuaishou_shop', 18, 'store'),
    ('business_resource.resource_type', '微信小店', 'wechat_shop', 19, 'store'),
    ('business_resource.resource_type', '账号类', 'account', 20, NULL),
    ('business_resource.resource_type', '店铺授权账号', 'shop_account', 21, 'account'),
    ('business_resource.resource_type', 'API账号', 'api_account', 22, 'account'),
    ('business_resource.resource_type', '服务账号', 'service_account', 23, 'account'),
    ('business_resource.resource_type', '投放类', 'ad', 30, NULL),
    ('business_resource.resource_type', '抖音投放账号', 'douyin_ad_account', 31, 'ad'),
    ('business_resource.resource_type', '腾讯广告账号', 'tencent_ad_account', 32, 'ad'),
    ('business_resource.resource_type', '快手广告账号', 'kuaishou_ad_account', 33, 'ad'),
    ('business_resource.resource_type', '小红书广告账号', 'xiaohongshu_ad_account', 34, 'ad'),
    ('business_resource.resource_type', '广告计划', 'ad_campaign', 35, 'ad'),
    ('business_resource.resource_type', '内容类', 'content', 40, NULL),
    ('business_resource.resource_type', '抖音官方号', 'douyin_official_account', 41, 'content'),
    ('business_resource.resource_type', '抖音内容号', 'douyin_media_account', 42, 'content'),
    ('business_resource.resource_type', '小红书账号', 'xiaohongshu_account', 43, 'content'),
    ('business_resource.resource_type', '快手内容号', 'kuaishou_media_account', 44, 'content'),
    ('business_resource.resource_type', '视频号账号', 'wechat_video_account', 45, 'content'),
    ('business_resource.resource_type', '直播账号', 'live_account', 46, 'content'),
    ('business_resource.resource_type', '达人账号', 'kol_account', 47, 'content'),
    ('business_resource.resource_type', '供应链类', 'supply', 50, NULL),
    ('business_resource.resource_type', '仓库', 'warehouse', 51, 'supply'),
    ('business_resource.resource_type', '供应商', 'supplier', 52, 'supply'),
    ('business_resource.resource_type', '物流商', 'logistics', 53, 'supply'),
    ('business_resource.resource_type', '采购组织', 'purchase_org', 54, 'supply'),
    ('business_resource.resource_type', '库存组织', 'inventory_org', 55, 'supply'),
    ('business_resource.resource_type', '商品类', 'product', 60, NULL),
    ('business_resource.resource_type', '品牌', 'brand', 61, 'product'),
    ('business_resource.resource_type', '类目', 'category', 62, 'product'),
    ('business_resource.resource_type', '商品线', 'product_line', 63, 'product'),
    ('business_resource.resource_type', 'SPU', 'spu', 64, 'product'),
    ('business_resource.resource_type', 'SKU', 'sku', 65, 'product'),
    ('business_resource.resource_type', '财务类', 'finance', 70, NULL),
    ('business_resource.resource_type', '结算主体', 'settlement_entity', 71, 'finance'),
    ('business_resource.resource_type', '结算账户', 'settlement_account', 72, 'finance'),
    ('business_resource.resource_type', '成本中心', 'cost_center', 73, 'finance'),
    ('business_resource.resource_type', '利润中心', 'profit_center', 74, 'finance'),
    ('business_resource.resource_type', '组织类', 'org', 80, NULL),
    ('business_resource.resource_type', '组织节点', 'org_node', 81, 'org'),
    ('business_resource.resource_type', '用户', 'user', 82, 'org'),
    ('business_resource.resource_type', '角色', 'role', 83, 'org'),
    ('business_resource.resource_type', '岗位', 'position', 84, 'org'),
    ('business_resource.resource_type', '集成类', 'integration', 90, NULL),
    ('business_resource.resource_type', '连接实例', 'connection_instance', 91, 'integration'),
    ('business_resource.resource_type', '外部应用', 'external_app', 92, 'integration'),
    ('business_resource.resource_type', '外部API账号', 'external_api_account', 93, 'integration')
),
target_type AS (
  SELECT dt.id AS dict_type_id, dt.tenant_id, dt.code
  FROM dict_type dt
  JOIN platform_tenant pt ON pt.id = dt.tenant_id
  WHERE dt.deleted_at IS NULL
),
target_item AS (
  SELECT tt.dict_type_id, tt.tenant_id, si.label, si.value, si.sort_order, si.parent_value
  FROM target_type tt
  JOIN seed_item si ON si.type_code = tt.code
),
parent_item AS (
  SELECT di.dict_type_id, di.value, di.id
  FROM dict_item di
  JOIN target_type tt ON tt.dict_type_id = di.dict_type_id
  WHERE di.deleted_at IS NULL
)
UPDATE dict_item di
SET
  label = ti.label,
  sort_order = ti.sort_order,
  enabled = true,
  parent_id = parent_item.id,
  deleted_at = NULL,
  updated_at = now()
FROM target_item ti
LEFT JOIN parent_item ON parent_item.dict_type_id = ti.dict_type_id
  AND parent_item.value = ti.parent_value
WHERE di.tenant_id = ti.tenant_id
  AND di.dict_type_id = ti.dict_type_id
  AND di.value = ti.value;
