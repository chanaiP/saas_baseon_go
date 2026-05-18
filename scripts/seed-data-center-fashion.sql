-- Fashion retail operating dataset for Ai经营决策中心.
-- Local/dev seed only. It replaces data-center rows for the target tenants.

BEGIN;

CREATE TEMP TABLE seed_tenants(tenant_id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO seed_tenants VALUES (1), (2), (3);

DELETE FROM data_center_raw_data_errors WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_rectification_reviews WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_task_logs WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_rectification_tasks WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_ai_diagnosis_records WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_anomaly_records WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_metric_results WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_anomaly_rules WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_metric_definitions WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_std_refund_orders WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_std_inventory_daily WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_std_store_sales_daily WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_std_ad_daily WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_std_sales_orders WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_std_products WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);
DELETE FROM data_center_raw_data_batches WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants);

CREATE TEMP TABLE fashion_brands(
  brand_code text PRIMARY KEY,
  brand_name text NOT NULL,
  price_band numeric NOT NULL,
  base_weight numeric NOT NULL
) ON COMMIT DROP;

INSERT INTO fashion_brands VALUES
  ('LEE', 'Lee', 0.92, 1.08),
  ('MARDI', 'Mardi Mercredi', 1.12, 0.92),
  ('UA', '安德玛', 1.28, 1.18),
  ('HAPPY_SOCKS', 'Happy Socks', 0.42, 0.76),
  ('CARHARTT_WIP', 'Carhartt WIP', 1.06, 0.84),
  ('NEW_BALANCE_APPAREL', 'New Balance Apparel', 0.98, 0.88);

CREATE TEMP TABLE fashion_products(
  brand_code text,
  product_code text PRIMARY KEY,
  product_name text,
  sku_code text,
  category_code text,
  category_name text,
  list_price numeric,
  cost_price numeric
) ON COMMIT DROP;

INSERT INTO fashion_products VALUES
  ('LEE', 'LEE-DENIM-501', 'Lee 101+ 男士直筒牛仔裤 原色蓝', 'LEE-DENIM-501-32', 'denim', '牛仔裤', 699, 286),
  ('LEE', 'LEE-JACKET-RIDER', 'Lee Rider 复古牛仔夹克 中蓝', 'LEE-JACKET-RIDER-M', 'outerwear', '外套', 899, 382),
  ('LEE', 'LEE-SHIRT-WORK', 'Lee Workwear 条纹长袖衬衫', 'LEE-SHIRT-WORK-L', 'shirt', '衬衫', 499, 198),
  ('LEE', 'LEE-TEE-LOGO', 'Lee 小标短袖 T 恤 米白', 'LEE-TEE-LOGO-M', 'tshirt', 'T 恤', 229, 82),
  ('MARDI', 'MARDI-SWEAT-FLOWER', 'Mardi Mercredi Flower 卫衣 灰色', 'MARDI-SWEAT-FLOWER-M', 'sweatshirt', '卫衣', 688, 286),
  ('MARDI', 'MARDI-CARDIGAN-CLASSIC', 'Mardi Mercredi 经典针织开衫 奶油色', 'MARDI-CARDIGAN-CLASSIC-F', 'knitwear', '针织衫', 798, 336),
  ('MARDI', 'MARDI-TEE-FLOWER', 'Mardi Mercredi Flower T 恤 白色', 'MARDI-TEE-FLOWER-S', 'tshirt', 'T 恤', 328, 118),
  ('MARDI', 'MARDI-BAG-ECO', 'Mardi Mercredi 帆布托特包', 'MARDI-BAG-ECO-OS', 'accessory', '配饰', 268, 92),
  ('UA', 'UA-HOODIE-RIVAL', '安德玛 Rival Fleece 连帽卫衣 黑色', 'UA-HOODIE-RIVAL-L', 'sweatshirt', '卫衣', 599, 246),
  ('UA', 'UA-LEGGING-HEATGEAR', '安德玛 HeatGear 女子训练紧身裤', 'UA-LEGGING-HEATGEAR-M', 'training', '训练服', 499, 196),
  ('UA', 'UA-SHORTS-TECH', '安德玛 Tech 训练短裤 深灰', 'UA-SHORTS-TECH-L', 'training', '训练服', 299, 106),
  ('UA', 'UA-TEE-RUSH', '安德玛 RUSH 速干训练 T 恤', 'UA-TEE-RUSH-M', 'tshirt', 'T 恤', 349, 128),
  ('HAPPY_SOCKS', 'HS-SOCK-BIGDOT', 'Happy Socks Big Dot 中筒袜', 'HS-SOCK-BIGDOT-41', 'socks', '袜品', 89, 24),
  ('HAPPY_SOCKS', 'HS-SOCK-STRIPE', 'Happy Socks 彩条中筒袜', 'HS-SOCK-STRIPE-39', 'socks', '袜品', 89, 24),
  ('HAPPY_SOCKS', 'HS-GIFT-BOX-4PK', 'Happy Socks 4 双礼盒装', 'HS-GIFT-BOX-4PK-OS', 'gift', '礼盒', 299, 96),
  ('HAPPY_SOCKS', 'HS-SOCK-DRESS', 'Happy Socks 商务趣味袜', 'HS-SOCK-DRESS-43', 'socks', '袜品', 99, 28),
  ('CARHARTT_WIP', 'CHW-JACKET-DETROIT', 'Carhartt WIP Detroit Jacket 棕色', 'CHW-JACKET-DETROIT-M', 'outerwear', '外套', 1599, 688),
  ('CARHARTT_WIP', 'CHW-PANT-SINGLEKNEE', 'Carhartt WIP Single Knee 工装裤', 'CHW-PANT-SINGLEKNEE-31', 'pants', '裤装', 999, 426),
  ('CARHARTT_WIP', 'CHW-TEE-POCKET', 'Carhartt WIP Pocket T 恤 黑色', 'CHW-TEE-POCKET-L', 'tshirt', 'T 恤', 399, 142),
  ('CARHARTT_WIP', 'CHW-BEANIE-ACRYLIC', 'Carhartt WIP Acrylic Watch Hat', 'CHW-BEANIE-ACRYLIC-OS', 'accessory', '配饰', 299, 92),
  ('NEW_BALANCE_APPAREL', 'NBA-TEE-ATHLETICS', 'New Balance Athletics 短袖 T 恤', 'NBA-TEE-ATHLETICS-M', 'tshirt', 'T 恤', 259, 94),
  ('NEW_BALANCE_APPAREL', 'NBA-PANT-ESSENTIAL', 'New Balance Essentials 运动长裤', 'NBA-PANT-ESSENTIAL-L', 'pants', '裤装', 499, 190),
  ('NEW_BALANCE_APPAREL', 'NBA-HOODIE-CLASSIC', 'New Balance 经典套头卫衣', 'NBA-HOODIE-CLASSIC-M', 'sweatshirt', '卫衣', 599, 238),
  ('NEW_BALANCE_APPAREL', 'NBA-JACKET-RUN', 'New Balance 轻量跑步夹克', 'NBA-JACKET-RUN-L', 'outerwear', '外套', 699, 268);

INSERT INTO data_center_raw_data_batches (
  tenant_id, batch_code, data_type, platform_code, app_code, connection_code,
  record_count, success_count, failed_count, sync_time, status, source_params, sample_payload,
  created_at, updated_at
)
SELECT t.tenant_id,
       'FASHION-' || t.tenant_id || '-' || b.data_type || '-' || b.platform_code || '-' || to_char(current_date - b.day_offset, 'YYYYMMDD'),
       b.data_type,
       b.platform_code,
       'data-center',
       b.connection_code,
       b.record_count,
       b.record_count - b.failed_count,
       b.failed_count,
       (current_date - b.day_offset) + b.sync_time,
       CASE WHEN b.failed_count = 0 THEN 'success' ELSE 'warning' END,
       jsonb_build_object('industry', 'fashion-retail', 'source', b.platform_code, 'brands', jsonb_build_array('Lee','Mardi Mercredi','安德玛','Happy Socks')),
       jsonb_build_object('sample_brand', b.sample_brand, 'sample_sku', b.sample_sku),
       now(), now()
FROM seed_tenants t
CROSS JOIN (
  VALUES
    ('sales', 'tmall', 'tmall-fashion-prod', 11840, 0, 1, time '02:10', 'Lee', 'LEE-DENIM-501-32'),
    ('sales', 'douyin', 'douyin-fashion-live', 9360, 18, 1, time '02:25', 'Mardi Mercredi', 'MARDI-SWEAT-FLOWER-M'),
    ('ad', 'douyin', 'douyin-ad-fashion', 1280, 6, 1, time '02:40', '安德玛', 'UA-HOODIE-RIVAL-L'),
    ('inventory', 'wms', 'wms-east-china', 4220, 9, 1, time '03:05', 'Happy Socks', 'HS-GIFT-BOX-4PK-OS'),
    ('store-sales', 'pos', 'pos-omni-channel', 3680, 0, 1, time '03:20', 'Carhartt WIP', 'CHW-JACKET-DETROIT-M'),
    ('refund', 'tmall', 'tmall-refund-center', 760, 4, 1, time '03:35', 'New Balance Apparel', 'NBA-HOODIE-CLASSIC-M'),
    ('sales', 'jd', 'jd-fashion-flagship', 6420, 0, 2, time '02:15', 'Lee', 'LEE-JACKET-RIDER-M'),
    ('ad', 'xiaohongshu', 'xhs-fashion-kol', 880, 0, 2, time '02:55', 'Mardi Mercredi', 'MARDI-CARDIGAN-CLASSIC-F')
) AS b(data_type, platform_code, connection_code, record_count, failed_count, day_offset, sync_time, sample_brand, sample_sku);

INSERT INTO data_center_raw_data_errors (tenant_id, batch_id, batch_code, row_number, error_code, error_reason, raw_payload, created_at)
SELECT b.tenant_id, b.id, b.batch_code, gs.n, 'FIELD_MISSING',
       CASE WHEN b.data_type = 'inventory' THEN '尺码库存缺少 warehouse_code，已进入待补齐队列'
            WHEN b.data_type = 'refund' THEN '退货原因字段为空，已按平台售后备注补齐'
            ELSE '平台原始字段缺失，已使用数据字典兜底' END,
       jsonb_build_object('batch', b.batch_code, 'row', gs.n, 'industry', 'fashion-retail'),
       now()
FROM data_center_raw_data_batches b
JOIN LATERAL generate_series(1, LEAST(b.failed_count, 6)) AS gs(n) ON true
WHERE b.tenant_id IN (SELECT tenant_id FROM seed_tenants);

INSERT INTO data_center_std_products (
  tenant_id, product_code, product_name, sku_code, brand_code, brand_name,
  category_code, category_name, list_price, cost_price, product_status, source_batch_code,
  created_at, updated_at
)
SELECT t.tenant_id, p.product_code, p.product_name, p.sku_code, p.brand_code, b.brand_name,
       p.category_code, p.category_name, p.list_price, p.cost_price, 'active',
       'FASHION-' || t.tenant_id || '-inventory-' || to_char(current_date - 1, 'YYYYMMDD'),
       now(), now()
FROM seed_tenants t
JOIN fashion_products p ON true
JOIN fashion_brands b ON b.brand_code = p.brand_code;

INSERT INTO data_center_std_sales_orders (
  tenant_id, order_code, brand_code, brand_name, channel_code, platform_code,
  resource_code, resource_name, product_code, product_name, sku_code,
  sales_amount, paid_amount, refund_amount, order_status, order_time, source_batch_code,
  created_at, updated_at
)
SELECT t.tenant_id,
       'ORD-' || t.tenant_id || '-' || to_char(d.day, 'YYYYMMDD') || '-' || p.product_code || '-' || ch.platform_code || '-' || gs.n,
       p.brand_code, b.brand_name, ch.channel_code, ch.platform_code,
       ch.resource_code, ch.resource_name, p.product_code, p.product_name, p.sku_code,
       round((p.list_price * b.price_band * ch.multiplier * d.day_factor * (0.88 + (gs.n % 5) * 0.035))::numeric, 2),
       round((p.list_price * b.price_band * ch.multiplier * d.day_factor * (0.83 + (gs.n % 4) * 0.032))::numeric, 2),
       CASE
         WHEN p.brand_code = 'HAPPY_SOCKS' AND d.day >= current_date - 2 AND gs.n IN (2,5) THEN round((p.list_price * 0.32)::numeric, 2)
         WHEN p.brand_code = 'MARDI' AND ch.channel_code = 'live' AND d.day = current_date - 1 AND gs.n = 3 THEN round((p.list_price * 0.22)::numeric, 2)
         WHEN gs.n = 7 THEN round((p.list_price * 0.08)::numeric, 2)
         ELSE 0
       END,
       'paid',
       d.day + make_interval(hours => 9 + ((gs.n * 2) % 11), mins => (gs.n * 7) % 60),
       'FASHION-' || t.tenant_id || '-sales-' || to_char(d.day, 'YYYYMMDD'),
       now(), now()
FROM seed_tenants t
JOIN generate_series(current_date - 13, current_date, interval '1 day') AS gs_day(day) ON true
JOIN LATERAL (
  SELECT gs_day.day,
         CASE
           WHEN gs_day.day = current_date THEN 1.13
           WHEN extract(isodow from gs_day.day) IN (5,6,7) THEN 1.18
           ELSE 1.00
         END AS day_factor
) d ON true
JOIN fashion_products p ON true
JOIN fashion_brands b ON b.brand_code = p.brand_code
JOIN (
  VALUES
    ('online', 'tmall', 'tmall-flagship', '天猫官方旗舰店', 1.02),
    ('online', 'jd', 'jd-flagship', '京东官方旗舰店', 0.82),
    ('live', 'douyin', 'dy-live-room', '抖音品牌直播间', 1.18),
    ('offline', 'pos', 'shanghai-store', '上海核心商圈门店', 0.68)
) AS ch(channel_code, platform_code, resource_code, resource_name, multiplier) ON true
JOIN generate_series(1, 2 + CASE WHEN p.brand_code IN ('LEE','UA') THEN 2 WHEN p.brand_code = 'HAPPY_SOCKS' THEN 1 ELSE 0 END) AS gs(n) ON true;

INSERT INTO data_center_std_store_sales_daily (
  tenant_id, brand_code, brand_name, channel_code, platform_code, store_code, store_name,
  stat_date, gmv, net_sales, order_count, customer_count, refund_amount, target_amount,
  source_batch_code, created_at, updated_at
)
SELECT t.tenant_id, b.brand_code, b.brand_name, s.channel_code, s.platform_code, b.brand_code || '-' || s.store_code, b.brand_name || ' ' || s.store_name,
       d.day::date,
       round((s.base_gmv * b.base_weight * d.factor)::numeric, 2),
       round((s.base_gmv * b.base_weight * d.factor * 0.91)::numeric, 2),
       round((s.base_orders * b.base_weight * d.factor)::numeric, 0)::bigint,
       round((s.base_orders * b.base_weight * d.factor * 0.78)::numeric, 0)::bigint,
       round((s.base_gmv * b.base_weight * d.factor * CASE WHEN b.brand_code = 'HAPPY_SOCKS' THEN 0.055 ELSE 0.028 END)::numeric, 2),
       round((s.base_gmv * b.base_weight * 1.08)::numeric, 2),
       'FASHION-' || t.tenant_id || '-store-sales-' || to_char(d.day, 'YYYYMMDD'),
       now(), now()
FROM seed_tenants t
JOIN fashion_brands b ON true
JOIN (
  VALUES
    ('SH-XTD', '上海新天地旗舰店', 'offline', 'pos', 38500, 92),
    ('BJ-SKP', '北京 SKP 店', 'offline', 'pos', 34200, 78),
    ('HZ-IN77', '杭州湖滨 in77 店', 'offline', 'pos', 31800, 74),
    ('SZ-VIENTIANE', '深圳万象天地店', 'offline', 'pos', 29600, 70)
) AS s(store_code, store_name, channel_code, platform_code, base_gmv, base_orders) ON true
JOIN (
  SELECT gs::date AS day,
         CASE WHEN gs::date = current_date THEN 1.12
              WHEN extract(isodow from gs::date) IN (5,6,7) THEN 1.18
              ELSE 0.96 END AS factor
  FROM generate_series(current_date - 13, current_date, interval '1 day') gs
) d ON true;

INSERT INTO data_center_std_ad_daily (
  tenant_id, brand_code, brand_name, platform_code, account_code, account_name,
  campaign_code, campaign_name, creative_code, product_code, stat_date,
  cost_amount, impression_count, click_count, order_amount, order_count, roi, click_rate, conversion_rate,
  source_batch_code, created_at, updated_at
)
SELECT t.tenant_id, b.brand_code, b.brand_name, ad.platform_code,
       ad.platform_code || '-' || lower(b.brand_code) || '-acct',
       b.brand_name || ' ' || ad.platform_name || '投放账户',
       ad.platform_code || '-' || lower(b.brand_code) || '-' || to_char(d.day, 'MMDD'),
       b.brand_name || ' ' || ad.campaign_name,
       'creative-' || lower(b.brand_code) || '-' || ad.platform_code,
       p.product_code,
       d.day::date,
       round((ad.base_cost * b.base_weight * d.factor)::numeric, 2),
       round((ad.base_impr * b.base_weight * d.factor)::numeric, 0)::bigint,
       round((ad.base_click * b.base_weight * d.factor)::numeric, 0)::bigint,
       round((ad.base_cost * b.base_weight * d.factor * ad.roi_factor * CASE WHEN b.brand_code = 'MARDI' AND ad.platform_code = 'douyin' AND d.day >= current_date - 2 THEN 1.42 ELSE 1 END)::numeric, 2),
       round((ad.base_orders * b.base_weight * d.factor)::numeric, 0)::bigint,
       round((ad.roi_factor * CASE WHEN b.brand_code = 'MARDI' AND ad.platform_code = 'douyin' AND d.day >= current_date - 2 THEN 1.42 ELSE 1 END)::numeric, 4),
       round(((ad.base_click / ad.base_impr) * 100)::numeric, 4),
       round(((ad.base_orders / ad.base_click) * 100)::numeric, 4),
       'FASHION-' || t.tenant_id || '-ad-' || to_char(d.day, 'YYYYMMDD'),
       now(), now()
FROM seed_tenants t
JOIN fashion_brands b ON true
JOIN LATERAL (
  SELECT product_code FROM fashion_products fp WHERE fp.brand_code = b.brand_code ORDER BY product_code LIMIT 1
) p ON true
JOIN (
  VALUES
    ('douyin', '抖音', '春夏新品直播加热', 9800, 280000, 15200, 238, 3.45),
    ('xiaohongshu', '小红书', '穿搭种草内容投放', 6200, 160000, 10400, 162, 3.05)
) AS ad(platform_code, platform_name, campaign_name, base_cost, base_impr, base_click, base_orders, roi_factor) ON true
JOIN (
  SELECT gs::date AS day,
         CASE WHEN extract(isodow from gs::date) IN (5,6,7) THEN 1.14 ELSE 0.98 END AS factor
  FROM generate_series(current_date - 13, current_date, interval '1 day') gs
) d ON true;

INSERT INTO data_center_std_inventory_daily (
  tenant_id, brand_code, product_code, product_name, sku_code, warehouse_code, store_code,
  stat_date, available_stock, in_transit_stock, sales_7d, available_days, inventory_status,
  source_batch_code, created_at, updated_at
)
SELECT t.tenant_id, p.brand_code, p.product_code, p.product_name, p.sku_code,
       'WH-EAST', s.store_code, d.day::date,
       CASE
         WHEN p.brand_code = 'HAPPY_SOCKS' AND p.category_code = 'gift' THEN 68 + (extract(day from d.day)::int % 9)
         WHEN p.brand_code = 'UA' AND p.category_code = 'training' THEN 145 + (extract(day from d.day)::int % 13)
         ELSE 220 + (extract(day from d.day)::int % 17) * 4
       END,
       20 + (extract(day from d.day)::int % 5) * 3,
       CASE
         WHEN p.brand_code = 'HAPPY_SOCKS' AND p.category_code = 'gift' THEN 52
         WHEN p.brand_code = 'LEE' AND p.category_code = 'denim' THEN 34
         ELSE 22 + (extract(day from d.day)::int % 8)
       END,
       CASE
         WHEN p.brand_code = 'HAPPY_SOCKS' AND p.category_code = 'gift' THEN 1.4
         WHEN p.brand_code = 'LEE' AND p.category_code = 'denim' THEN 7.2
         ELSE 14.6 + (extract(day from d.day)::int % 6)
       END,
       CASE
         WHEN p.brand_code = 'HAPPY_SOCKS' AND p.category_code = 'gift' THEN 'shortage'
         WHEN p.brand_code = 'LEE' AND p.category_code = 'denim' THEN 'watch'
         ELSE 'normal'
       END,
       'FASHION-' || t.tenant_id || '-inventory-' || to_char(d.day, 'YYYYMMDD'),
       now(), now()
FROM seed_tenants t
JOIN fashion_products p ON true
JOIN (
  VALUES ('SH-XTD'), ('BJ-SKP'), ('HZ-IN77'), ('SZ-VIENTIANE')
) AS s(store_code) ON true
JOIN generate_series(current_date - 13, current_date, interval '1 day') d(day) ON true;

INSERT INTO data_center_std_refund_orders (
  tenant_id, refund_code, order_code, brand_code, brand_name, channel_code, platform_code,
  product_code, product_name, sku_code, refund_amount, refund_reason, refund_status, refund_time,
  source_batch_code, created_at, updated_at
)
SELECT tenant_id,
       'RF-' || order_code,
       order_code, brand_code, brand_name, channel_code, platform_code, product_code, product_name, sku_code,
       refund_amount,
       CASE
         WHEN brand_code = 'HAPPY_SOCKS' THEN '礼盒包装轻微压痕'
         WHEN brand_code = 'MARDI' THEN '尺码偏大/换码'
         WHEN brand_code = 'LEE' THEN '牛仔裤版型不合适'
         ELSE '七天无理由'
       END,
       'completed',
       order_time + interval '1 day',
       source_batch_code,
       now(), now()
FROM data_center_std_sales_orders
WHERE tenant_id IN (SELECT tenant_id FROM seed_tenants)
  AND refund_amount > 0;

INSERT INTO data_center_metric_definitions (
  tenant_id, metric_code, metric_name, metric_category, formula, statistic_period, dimensions, data_source,
  enabled, anomaly_enabled, created_at, updated_at
)
SELECT t.tenant_id, m.metric_code, m.metric_name, m.metric_category, m.formula, 'DAY',
       '["brand","channel","platform","store","product"]'::jsonb,
       m.data_source, true, m.anomaly_enabled, now(), now()
FROM seed_tenants t
CROSS JOIN (
  VALUES
    ('gmv', 'GMV', 'sales', 'sum(sales_amount)', 'data_center_std_sales_orders', true),
    ('net_sales', '净销售额', 'sales', 'sum(paid_amount - refund_amount)', 'data_center_std_sales_orders', true),
    ('order_count', '订单数', 'sales', 'count(order_code)', 'data_center_std_sales_orders', true),
    ('avg_order_value', '客单价', 'sales', 'net_sales / order_count', 'data_center_std_sales_orders', true),
    ('refund_rate', '退款率', 'refund', 'sum(refund_amount) / sum(sales_amount)', 'data_center_std_sales_orders', true),
    ('ad_cost', '广告消耗', 'ad', 'sum(cost_amount)', 'data_center_std_ad_daily', true),
    ('ad_roi', '投放 ROI', 'ad', 'sum(order_amount) / sum(cost_amount)', 'data_center_std_ad_daily', true),
    ('click_rate', '点击率', 'ad', 'sum(click_count) / sum(impression_count)', 'data_center_std_ad_daily', false),
    ('conversion_rate', '转化率', 'ad', 'sum(order_count) / sum(click_count)', 'data_center_std_ad_daily', true),
    ('sell_through_7d', '7日动销率', 'inventory', 'sales_7d / available_stock', 'data_center_std_inventory_daily', true),
    ('available_days', '可售天数', 'inventory', 'available_stock / sales_7d', 'data_center_std_inventory_daily', true),
    ('store_target_rate', '门店目标达成率', 'store', 'gmv / target_amount', 'data_center_std_store_sales_daily', true)
) AS m(metric_code, metric_name, metric_category, formula, data_source, anomaly_enabled)
ON CONFLICT (tenant_id, metric_code) WHERE deleted_at IS NULL DO UPDATE
SET metric_name = EXCLUDED.metric_name,
    metric_category = EXCLUDED.metric_category,
    formula = EXCLUDED.formula,
    statistic_period = EXCLUDED.statistic_period,
    dimensions = EXCLUDED.dimensions,
    data_source = EXCLUDED.data_source,
    enabled = EXCLUDED.enabled,
    anomaly_enabled = EXCLUDED.anomaly_enabled,
    updated_at = now();

INSERT INTO data_center_metric_results (
  tenant_id, metric_code, resource_type, resource_code, resource_name, stat_date, period_type,
  metric_value, compare_value, compare_rate, target_value, dimension_json, created_at, updated_at
)
SELECT tenant_id, metric_code, 'brand', brand_code, brand_name, stat_date, 'DAY',
       metric_value, compare_value,
       CASE WHEN compare_value > 0 THEN round(((metric_value - compare_value) / compare_value * 100)::numeric, 4) END,
       target_value,
       jsonb_build_object('industry', 'fashion-retail', 'brand_code', brand_code),
       now(), now()
FROM (
  SELECT s.tenant_id, 'gmv' AS metric_code, s.brand_code, max(s.brand_name) brand_name, date(s.order_time) stat_date,
         sum(s.sales_amount)::numeric AS metric_value,
         lag(sum(s.sales_amount)::numeric) over(partition by s.tenant_id, s.brand_code order by date(s.order_time)) AS compare_value,
         (sum(s.sales_amount)::numeric * 1.08) AS target_value
  FROM data_center_std_sales_orders s
  WHERE s.tenant_id IN (SELECT tenant_id FROM seed_tenants) AND s.deleted_at IS NULL
  GROUP BY s.tenant_id, s.brand_code, date(s.order_time)
  UNION ALL
  SELECT s.tenant_id, 'refund_rate', s.brand_code, max(s.brand_name), date(s.order_time),
         CASE WHEN sum(s.sales_amount) > 0 THEN round((sum(s.refund_amount) / sum(s.sales_amount) * 100)::numeric, 4) ELSE 0 END,
         NULL, 5
  FROM data_center_std_sales_orders s
  WHERE s.tenant_id IN (SELECT tenant_id FROM seed_tenants) AND s.deleted_at IS NULL
  GROUP BY s.tenant_id, s.brand_code, date(s.order_time)
  UNION ALL
  SELECT a.tenant_id, 'ad_roi', a.brand_code, max(a.brand_name), a.stat_date,
         CASE WHEN sum(a.cost_amount) > 0 THEN round((sum(a.order_amount) / sum(a.cost_amount))::numeric, 4) ELSE 0 END,
         NULL, 3
  FROM data_center_std_ad_daily a
  WHERE a.tenant_id IN (SELECT tenant_id FROM seed_tenants) AND a.deleted_at IS NULL
  GROUP BY a.tenant_id, a.brand_code, a.stat_date
) x
ON CONFLICT (tenant_id, metric_code, resource_type, resource_code, stat_date, period_type) DO UPDATE
SET metric_value = EXCLUDED.metric_value,
    compare_value = EXCLUDED.compare_value,
    compare_rate = EXCLUDED.compare_rate,
    target_value = EXCLUDED.target_value,
    dimension_json = EXCLUDED.dimension_json,
    updated_at = now();

INSERT INTO data_center_anomaly_rules (
  tenant_id, rule_code, rule_name, business_domain, target_object_type, scope_json,
  metric_conditions_json, level_config_json, confidence_config_json, ai_enabled, task_enabled,
  enabled, priority, created_at, updated_at
)
SELECT t.tenant_id, r.rule_code, r.rule_name, r.business_domain, r.target_object_type,
       r.scope::jsonb, r.metric_conditions::jsonb, r.level_config::jsonb, r.confidence_config::jsonb,
       true, true, true, r.priority, now(), now()
FROM seed_tenants t
CROSS JOIN (
  VALUES
    ('fashion_gmv_drop', '品牌 GMV 连续下滑', '销售异常', 'brand', '{"brands":["Lee","Mardi Mercredi","安德玛","Happy Socks"]}', '{"gmv":{"wow_drop_pct":18,"min_impact":20000}}', '{"high":25,"critical":40}', '{"base":80,"evidence_weight":12}', 5),
    ('fashion_ad_roi_low', '投放 ROI 低于目标', '投流异常', 'brand', '{"platforms":["douyin","xiaohongshu"]}', '{"ad_roi":{"lt":2.2},"ad_cost":{"gt":5000}}', '{"high":2.0,"critical":1.6}', '{"base":82,"evidence_weight":10}', 4),
    ('fashion_inventory_shortage', '畅销 SKU 可售天数不足', '库存异常', 'product', '{"category":["gift","denim","training"]}', '{"available_days":{"lt":3},"sales_7d":{"gt":30}}', '{"high":2,"critical":1}', '{"base":84,"evidence_weight":11}', 4),
    ('fashion_refund_rate_high', '退款率高于行业阈值', '退款异常', 'brand', '{"channels":["online","live"]}', '{"refund_rate":{"gt":6}}', '{"high":8,"critical":12}', '{"base":78,"evidence_weight":12}', 3)
) AS r(rule_code, rule_name, business_domain, target_object_type, scope, metric_conditions, level_config, confidence_config, priority);

INSERT INTO data_center_anomaly_records (
  tenant_id, anomaly_code, rule_code, title, business_domain, object_type, object_code, object_name,
  stat_date, anomaly_level, confidence_score, impact_amount, evidence_json, occurred_at,
  ai_status, task_status, review_status, status, created_at, updated_at
)
SELECT t.tenant_id, a.anomaly_code || '-' || t.tenant_id, a.rule_code, a.title, a.business_domain,
       a.object_type, a.object_code, a.object_name, current_date - a.day_offset,
       a.level, a.confidence, a.impact,
       a.evidence::jsonb, (current_date - a.day_offset) + time '10:30',
       a.ai_status, a.task_status, a.review_status, a.status, now(), now()
FROM seed_tenants t
CROSS JOIN (
  VALUES
    ('ANOM-FASH-LEE-GMV', 'fashion_gmv_drop', 'Lee 上海新天地牛仔裤系列 GMV 环比下滑', '销售异常', 'brand', 'LEE', 'Lee', 1, 'high', 91, 46820, '{"指标":"GMV","本期":"¥146,300","环比":"-23.8%","影响":"牛仔裤核心尺码 31/32 断码"}', 'completed', 'generated', 'confirmed', 'confirmed'),
    ('ANOM-FASH-MARDI-LIVE', 'fashion_ad_roi_low', 'Mardi Mercredi 抖音直播投放 ROI 波动', '投流异常', 'brand', 'MARDI', 'Mardi Mercredi', 1, 'medium', 84, 28760, '{"指标":"ad_roi","本期":"2.14","目标":"3.00","证据":"直播间花朵卫衣素材点击率下降"}', 'completed', 'generated', 'none', 'pending'),
    ('ANOM-FASH-UA-ROI', 'fashion_ad_roi_low', '安德玛训练服小红书投放成本偏高', '投流异常', 'brand', 'UA', '安德玛', 2, 'high', 88, 35280, '{"指标":"ad_roi","本期":"1.92","消耗":"¥48,600","证据":"训练紧身裤关键词竞争加剧"}', 'completed', 'generated', 'confirmed', 'confirmed'),
    ('ANOM-FASH-HS-STOCK', 'fashion_inventory_shortage', 'Happy Socks 礼盒装可售天数不足', '库存异常', 'product', 'HS-GIFT-BOX-4PK', 'Happy Socks 4 双礼盒装', 0, 'critical', 94, 52600, '{"指标":"available_days","本期":"1.4天","销量":"近7天 52 件/店","证据":"520 礼赠场景拉动"}', 'completed', 'generated', 'none', 'processing'),
    ('ANOM-FASH-CARHARTT-STORE', 'fashion_gmv_drop', 'Carhartt WIP 北京 SKP 门店外套销售低于目标', '销售异常', 'brand', 'CARHARTT_WIP', 'Carhartt WIP', 2, 'medium', 82, 21840, '{"指标":"store_target_rate","本期":"78%","目标":"100%","证据":"Detroit Jacket 到店试穿转化低"}', 'pending', 'none', 'none', 'pending'),
    ('ANOM-FASH-NB-REFUND', 'fashion_refund_rate_high', 'New Balance Apparel 卫衣尺码退款率偏高', '退款异常', 'brand', 'NEW_BALANCE_APPAREL', 'New Balance Apparel', 3, 'medium', 80, 16320, '{"指标":"refund_rate","本期":"6.8%","阈值":"5%","证据":"套头卫衣 L/XL 尺码偏大反馈集中"}', 'completed', 'generated', 'confirmed', 'confirmed')
) AS a(anomaly_code, rule_code, title, business_domain, object_type, object_code, object_name, day_offset, level, confidence, impact, evidence, ai_status, task_status, review_status, status);

INSERT INTO data_center_ai_diagnosis_records (
  tenant_id, anomaly_id, anomaly_code, problem_summary, impact_summary,
  reason_analysis_json, evidence_summary_json, suggestion_json, confidence_explanation,
  task_suggestion_json, model_code, status, generated_at, created_at, updated_at
)
SELECT a.tenant_id, a.id, a.anomaly_code,
       a.title,
       '预计影响销售额 ' || to_char(a.impact_amount, 'FM999,999,990.00') || ' 元，需结合库存、投放和门店执行动作闭环。',
       jsonb_build_array(
         jsonb_build_object('reason', '核心尺码/畅销 SKU 供给与当前投放节奏不匹配', 'weight', 0.42),
         jsonb_build_object('reason', '直播或内容素材与当季穿搭场景疲劳', 'weight', 0.31),
         jsonb_build_object('reason', '门店陈列与库存调拨节奏未跟上周末峰值', 'weight', 0.27)
       ),
       jsonb_build_array(a.evidence_json, jsonb_build_object('source', '订单/广告/库存/门店日报联动校验')),
       jsonb_build_array(
         jsonb_build_object('action', '优先补齐核心尺码并做跨店调拨', 'owner', '商品运营'),
         jsonb_build_object('action', '替换投放素材，拆分品牌词与场景词预算', 'owner', '投放运营'),
         jsonb_build_object('action', '门店执行重点款陈列和导购话术复训', 'owner', '零售运营')
       ),
       '规则证据覆盖销售、广告、库存和售后指标，置信度来自多指标一致性。',
       jsonb_build_object('deadline_days', 5, 'review_metric', 'gmv, ad_roi, available_days, refund_rate'),
       'data-center-fashion-local-analyzer',
       'success',
       now(), now(), now()
FROM data_center_anomaly_records a
WHERE a.tenant_id IN (SELECT tenant_id FROM seed_tenants)
  AND a.ai_status = 'completed';

INSERT INTO data_center_rectification_tasks (
  tenant_id, task_code, anomaly_id, anomaly_code, title, task_type, owner_role,
  collaborator_ids, priority, deadline, target_desc, ai_suggestion_json, execution_feedback,
  progress, status, review_status, completed_at, created_at, updated_at
)
SELECT a.tenant_id,
       replace(a.anomaly_code, 'ANOM', 'TASK'),
       a.id, a.anomaly_code,
       CASE
         WHEN a.object_code = 'HS-GIFT-BOX-4PK' THEN '补齐 Happy Socks 礼盒装核心库存'
         WHEN a.object_code = 'UA' THEN '优化安德玛训练服投放 ROI'
         WHEN a.object_code = 'LEE' THEN '恢复 Lee 牛仔裤核心尺码销售'
         WHEN a.object_code = 'MARDI' THEN '调整 Mardi Mercredi 直播间投放素材'
         ELSE '复核 ' || a.object_name || ' 异常闭环动作'
       END,
       'anomaly_rectification',
       CASE WHEN a.business_domain = '库存异常' THEN '商品运营' WHEN a.business_domain = '投流异常' THEN '投放运营' ELSE '零售运营' END,
       '[]'::jsonb,
       CASE WHEN a.anomaly_level IN ('critical','high','严重','高') THEN 'high' ELSE 'medium' END,
       current_date + CASE WHEN a.anomaly_level = 'critical' THEN 2 ELSE 5 END,
       CASE
         WHEN a.business_domain = '库存异常' THEN '核心 SKU 可售天数恢复到 7 天以上，缺货门店 48 小时内完成调拨。'
         WHEN a.business_domain = '投流异常' THEN '投放 ROI 恢复到 3.0 以上，低效素材预算占比降到 15% 以下。'
         ELSE 'GMV 恢复到目标 95% 以上，核心尺码断码率降到 5% 以下。'
       END,
       jsonb_build_object('source', 'AI 诊断建议', 'steps', jsonb_build_array('复盘证据', '制定动作', '执行反馈', '复盘验证')),
       CASE WHEN a.review_status = 'confirmed' THEN '已完成动作并进入复盘验证。' ELSE NULL END,
       CASE WHEN a.review_status = 'confirmed' THEN 100 WHEN a.status = 'processing' THEN 62 ELSE 35 END,
       CASE WHEN a.review_status = 'confirmed' THEN 'completed' WHEN a.status = 'processing' THEN 'processing' ELSE 'pending' END,
       CASE WHEN a.review_status = 'confirmed' THEN 'confirmed' ELSE 'none' END,
       CASE WHEN a.review_status = 'confirmed' THEN now() - interval '1 day' ELSE NULL END,
       now(), now()
FROM data_center_anomaly_records a
WHERE a.tenant_id IN (SELECT tenant_id FROM seed_tenants)
  AND a.task_status = 'generated';

INSERT INTO data_center_task_logs (tenant_id, task_id, task_code, action, from_status, to_status, content, operator_id, created_at)
SELECT t.tenant_id, t.id, t.task_code, 'feedback', NULL, t.status,
       '服装行业整改动作：已同步商品、投放和门店三方负责人，按品牌/SKU 拆解动作。',
       1, now()
FROM data_center_rectification_tasks t
WHERE t.tenant_id IN (SELECT tenant_id FROM seed_tenants);

INSERT INTO data_center_rectification_reviews (
  tenant_id, review_code, task_id, task_code, anomaly_id, anomaly_code,
  before_metric_json, after_metric_json, improvement_result, review_conclusion,
  ai_review_summary, manual_review_summary, experience_summary, reviewed_at,
  created_at, updated_at
)
SELECT t.tenant_id,
       replace(t.task_code, 'TASK', 'REV'),
       t.id, t.task_code, t.anomaly_id, t.anomaly_code,
       jsonb_build_array(jsonb_build_object('metric', '整改前指标', 'value', CASE WHEN t.anomaly_code LIKE '%UA%' THEN 'ROI 1.92' WHEN t.anomaly_code LIKE '%LEE%' THEN 'GMV -23.8%' ELSE '退款率 6.8%' END)),
       jsonb_build_array(jsonb_build_object('metric', '整改后指标', 'value', CASE WHEN t.anomaly_code LIKE '%UA%' THEN 'ROI 3.18' WHEN t.anomaly_code LIKE '%LEE%' THEN 'GMV +12.6%' ELSE '退款率 4.2%' END)),
       CASE WHEN t.anomaly_code LIKE '%UA%' THEN '+65.6%' WHEN t.anomaly_code LIKE '%LEE%' THEN '+36.4%' ELSE '-2.6pct' END,
       'effective',
       '整改动作执行后，核心经营指标恢复到目标区间，建议沉淀为服装品牌季节性运营 SOP。',
       '已确认复盘结果，后续按周检查尺码库存和素材转化。',
       '服装品牌异常处理需同步关注尺码、颜色、渠道素材与门店陈列，不应只看 GMV 单指标。',
       now() - interval '12 hours',
       now(), now()
FROM data_center_rectification_tasks t
WHERE t.tenant_id IN (SELECT tenant_id FROM seed_tenants)
  AND t.review_status = 'confirmed';

COMMIT;
