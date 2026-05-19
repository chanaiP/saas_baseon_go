-- Today-only simulation data for data-center closure validation.
-- It upserts standard business rows and does not delete historical data.

BEGIN;

WITH params AS (
  SELECT
    1::bigint AS tenant_id,
    current_date::date AS today,
    (current_date - interval '1 day')::date AS yesterday,
    to_char(current_date, 'YYYYMMDD') AS ymd
),
sales_rows AS (
  SELECT * FROM (
    SELECT tenant_id, 'SIM-' || ymd || '-LEE-Y-001' AS order_code, 'LEE_SIM_' || ymd AS brand_code, 'Lee' AS brand_name,
           'online' AS channel_code, 'tmall' AS platform_code, 'tmall-lee-flagship' AS resource_code, 'Lee 天猫旗舰店' AS resource_name,
           'LEE-DENIM-SIM-' || ymd AS product_code, 'Lee 101+ 直筒牛仔裤' AS product_name, 'LEE-DENIM-SIM-32' AS sku_code,
           5800::numeric AS sales_amount, 5520::numeric AS paid_amount, 0::numeric AS refund_amount, 'paid' AS order_status,
           yesterday + time '10:10' AS order_time, 'SIM-' || ymd || '-SALES' AS source_batch_code FROM params
    UNION ALL SELECT tenant_id, 'SIM-' || ymd || '-LEE-Y-002', 'LEE_SIM_' || ymd, 'Lee',
           'online', 'tmall', 'tmall-lee-flagship', 'Lee 天猫旗舰店',
           'LEE-JACKET-SIM-' || ymd, 'Lee Rider 复古牛仔夹克', 'LEE-JACKET-SIM-M',
           6200, 5880, 0, 'paid', yesterday + time '14:20', 'SIM-' || ymd || '-SALES' FROM params
    UNION ALL SELECT tenant_id, 'SIM-' || ymd || '-LEE-Y-003', 'LEE_SIM_' || ymd, 'Lee',
           'offline', 'pos', 'shanghai-store', 'Lee 上海港汇恒隆店',
           'LEE-SHIRT-SIM-' || ymd, 'Lee Workwear 条纹衬衫', 'LEE-SHIRT-SIM-L',
           4300, 4080, 0, 'paid', yesterday + time '17:35', 'SIM-' || ymd || '-SALES' FROM params
    UNION ALL SELECT tenant_id, 'SIM-' || ymd || '-LEE-Y-004', 'LEE_SIM_' || ymd, 'Lee',
           'live', 'douyin', 'dy-live-room', 'Lee 抖音直播间',
           'LEE-TEE-SIM-' || ymd, 'Lee 小标短袖 T 恤', 'LEE-TEE-SIM-M',
           3900, 3680, 0, 'paid', yesterday + time '21:05', 'SIM-' || ymd || '-SALES' FROM params
    UNION ALL SELECT tenant_id, 'SIM-' || ymd || '-LEE-T-001', 'LEE_SIM_' || ymd, 'Lee',
           'online', 'tmall', 'tmall-lee-flagship', 'Lee 天猫旗舰店',
           'LEE-DENIM-SIM-' || ymd, 'Lee 101+ 直筒牛仔裤', 'LEE-DENIM-SIM-32',
           3200, 2990, 0, 'paid', today + time '00:25', 'SIM-' || ymd || '-SALES' FROM params
    UNION ALL SELECT tenant_id, 'SIM-' || ymd || '-LEE-T-002', 'LEE_SIM_' || ymd, 'Lee',
           'offline', 'pos', 'shanghai-store', 'Lee 上海港汇恒隆店',
           'LEE-JACKET-SIM-' || ymd, 'Lee Rider 复古牛仔夹克', 'LEE-JACKET-SIM-M',
           2600, 2420, 0, 'paid', today + time '01:10', 'SIM-' || ymd || '-SALES' FROM params
  ) rows
),
ad_rows AS (
  SELECT * FROM (
    SELECT tenant_id, 'MARDI_SIM_' || ymd AS brand_code, 'Mardi Mercredi' AS brand_name, 'douyin' AS platform_code,
           'douyin-mardi-sim' AS account_code, 'Mardi Mercredi 抖音投放账户' AS account_name,
           'MARDI-FLOWER-LIVE-SIM-' || ymd AS campaign_code, 'Mardi Flower 卫衣直播加热' AS campaign_name,
           'creative-mardi-flower-' || ymd AS creative_code, 'MARDI-SWEAT-FLOWER' AS product_code,
           yesterday AS stat_date, 8200::numeric AS cost_amount, 210000::bigint AS impression_count, 14200::bigint AS click_count,
           35260::numeric AS order_amount, 146::bigint AS order_count, 4.3000::numeric AS roi,
           6.7619::numeric AS click_rate, 1.0282::numeric AS conversion_rate, 'SIM-' || ymd || '-AD' AS source_batch_code FROM params
    UNION ALL SELECT tenant_id, 'MARDI_SIM_' || ymd, 'Mardi Mercredi', 'douyin',
           'douyin-mardi-sim', 'Mardi Mercredi 抖音投放账户',
           'MARDI-FLOWER-LIVE-SIM-' || ymd, 'Mardi Flower 卫衣直播加热',
           'creative-mardi-flower-' || ymd, 'MARDI-SWEAT-FLOWER',
           today, 12850, 302000, 18900, 24672, 82, 1.9200, 6.2583, 0.4339, 'SIM-' || ymd || '-AD' FROM params
  ) rows
),
inventory_rows AS (
  SELECT * FROM (
    SELECT tenant_id, 'HAPPY_SOCKS_SIM_' || ymd AS brand_code, 'HS-GIFT-BOX-SIM-' || ymd AS product_code,
           'Happy Socks 4 双礼盒装' AS product_name, 'HS-GIFT-BOX-SIM-OS' AS sku_code,
           'WH-EAST' AS warehouse_code, 'SH-XTD' AS store_code, yesterday AS stat_date,
           420::bigint AS available_stock, 38::bigint AS in_transit_stock, 18::bigint AS sales_7d,
           24::numeric AS available_days, 'healthy' AS inventory_status, 'SIM-' || ymd || '-INV' AS source_batch_code FROM params
    UNION ALL SELECT tenant_id, 'HAPPY_SOCKS_SIM_' || ymd, 'HS-GIFT-BOX-SIM-' || ymd,
           'Happy Socks 4 双礼盒装', 'HS-GIFT-BOX-SIM-OS',
           'WH-EAST', 'SH-XTD', today,
           560, 42, 2, 86, 'slow_moving', 'SIM-' || ymd || '-INV' FROM params
  ) rows
)
INSERT INTO data_center_std_sales_orders (
  tenant_id, order_code, brand_code, brand_name, channel_code, platform_code,
  resource_code, resource_name, product_code, product_name, sku_code,
  sales_amount, paid_amount, refund_amount, order_status, order_time,
  source_batch_code, created_at, updated_at
)
SELECT tenant_id, order_code, brand_code, brand_name, channel_code, platform_code,
       resource_code, resource_name, product_code, product_name, sku_code,
       sales_amount, paid_amount, refund_amount, order_status, order_time,
       source_batch_code, now(), now()
FROM sales_rows
ON CONFLICT (tenant_id, order_code) WHERE deleted_at IS NULL DO UPDATE
SET brand_code = EXCLUDED.brand_code,
    brand_name = EXCLUDED.brand_name,
    channel_code = EXCLUDED.channel_code,
    platform_code = EXCLUDED.platform_code,
    resource_code = EXCLUDED.resource_code,
    resource_name = EXCLUDED.resource_name,
    product_code = EXCLUDED.product_code,
    product_name = EXCLUDED.product_name,
    sku_code = EXCLUDED.sku_code,
    sales_amount = EXCLUDED.sales_amount,
    paid_amount = EXCLUDED.paid_amount,
    refund_amount = EXCLUDED.refund_amount,
    order_status = EXCLUDED.order_status,
    order_time = EXCLUDED.order_time,
    source_batch_code = EXCLUDED.source_batch_code,
    updated_at = now();

WITH params AS (
  SELECT 1::bigint AS tenant_id, current_date::date AS today, (current_date - interval '1 day')::date AS yesterday, to_char(current_date, 'YYYYMMDD') AS ymd
),
ad_rows AS (
  SELECT * FROM (
    SELECT tenant_id, 'MARDI_SIM_' || ymd AS brand_code, 'Mardi Mercredi' AS brand_name, 'douyin' AS platform_code,
           'douyin-mardi-sim' AS account_code, 'Mardi Mercredi 抖音投放账户' AS account_name,
           'MARDI-FLOWER-LIVE-SIM-' || ymd AS campaign_code, 'Mardi Flower 卫衣直播加热' AS campaign_name,
           'creative-mardi-flower-' || ymd AS creative_code, 'MARDI-SWEAT-FLOWER' AS product_code,
           yesterday AS stat_date, 8200::numeric AS cost_amount, 210000::bigint AS impression_count, 14200::bigint AS click_count,
           35260::numeric AS order_amount, 146::bigint AS order_count, 4.3000::numeric AS roi,
           6.7619::numeric AS click_rate, 1.0282::numeric AS conversion_rate, 'SIM-' || ymd || '-AD' AS source_batch_code FROM params
    UNION ALL SELECT tenant_id, 'MARDI_SIM_' || ymd, 'Mardi Mercredi', 'douyin',
           'douyin-mardi-sim', 'Mardi Mercredi 抖音投放账户',
           'MARDI-FLOWER-LIVE-SIM-' || ymd, 'Mardi Flower 卫衣直播加热',
           'creative-mardi-flower-' || ymd, 'MARDI-SWEAT-FLOWER',
           today, 12850, 302000, 18900, 24672, 82, 1.9200, 6.2583, 0.4339, 'SIM-' || ymd || '-AD' FROM params
  ) rows
)
INSERT INTO data_center_std_ad_daily (
  tenant_id, brand_code, brand_name, platform_code, account_code, account_name,
  campaign_code, campaign_name, creative_code, product_code, stat_date,
  cost_amount, impression_count, click_count, order_amount, order_count,
  roi, click_rate, conversion_rate, source_batch_code, created_at, updated_at
)
SELECT tenant_id, brand_code, brand_name, platform_code, account_code, account_name,
       campaign_code, campaign_name, creative_code, product_code, stat_date,
       cost_amount, impression_count, click_count, order_amount, order_count,
       roi, click_rate, conversion_rate, source_batch_code, now(), now()
FROM ad_rows
ON CONFLICT (tenant_id, account_code, campaign_code, stat_date) WHERE deleted_at IS NULL DO UPDATE
SET brand_code = EXCLUDED.brand_code,
    brand_name = EXCLUDED.brand_name,
    platform_code = EXCLUDED.platform_code,
    account_name = EXCLUDED.account_name,
    campaign_name = EXCLUDED.campaign_name,
    creative_code = EXCLUDED.creative_code,
    product_code = EXCLUDED.product_code,
    cost_amount = EXCLUDED.cost_amount,
    impression_count = EXCLUDED.impression_count,
    click_count = EXCLUDED.click_count,
    order_amount = EXCLUDED.order_amount,
    order_count = EXCLUDED.order_count,
    roi = EXCLUDED.roi,
    click_rate = EXCLUDED.click_rate,
    conversion_rate = EXCLUDED.conversion_rate,
    source_batch_code = EXCLUDED.source_batch_code,
    updated_at = now();

WITH params AS (
  SELECT 1::bigint AS tenant_id, current_date::date AS today, (current_date - interval '1 day')::date AS yesterday, to_char(current_date, 'YYYYMMDD') AS ymd
),
inventory_rows AS (
  SELECT * FROM (
    SELECT tenant_id, 'HAPPY_SOCKS_SIM_' || ymd AS brand_code, 'HS-GIFT-BOX-SIM-' || ymd AS product_code,
           'Happy Socks 4 双礼盒装' AS product_name, 'HS-GIFT-BOX-SIM-OS' AS sku_code,
           'WH-EAST' AS warehouse_code, 'SH-XTD' AS store_code, yesterday AS stat_date,
           420::bigint AS available_stock, 38::bigint AS in_transit_stock, 18::bigint AS sales_7d,
           24::numeric AS available_days, 'healthy' AS inventory_status, 'SIM-' || ymd || '-INV' AS source_batch_code FROM params
    UNION ALL SELECT tenant_id, 'HAPPY_SOCKS_SIM_' || ymd, 'HS-GIFT-BOX-SIM-' || ymd,
           'Happy Socks 4 双礼盒装', 'HS-GIFT-BOX-SIM-OS',
           'WH-EAST', 'SH-XTD', today,
           560, 42, 2, 86, 'slow_moving', 'SIM-' || ymd || '-INV' FROM params
  ) rows
)
INSERT INTO data_center_std_inventory_daily (
  tenant_id, brand_code, product_code, product_name, sku_code, warehouse_code, store_code,
  stat_date, available_stock, in_transit_stock, sales_7d, available_days,
  inventory_status, source_batch_code, created_at, updated_at
)
SELECT tenant_id, brand_code, product_code, product_name, sku_code, warehouse_code, store_code,
       stat_date, available_stock, in_transit_stock, sales_7d, available_days,
       inventory_status, source_batch_code, now(), now()
FROM inventory_rows
ON CONFLICT (tenant_id, product_code, sku_code, store_code, stat_date) WHERE deleted_at IS NULL DO UPDATE
SET brand_code = EXCLUDED.brand_code,
    product_name = EXCLUDED.product_name,
    warehouse_code = EXCLUDED.warehouse_code,
    available_stock = EXCLUDED.available_stock,
    in_transit_stock = EXCLUDED.in_transit_stock,
    sales_7d = EXCLUDED.sales_7d,
    available_days = EXCLUDED.available_days,
    inventory_status = EXCLUDED.inventory_status,
    source_batch_code = EXCLUDED.source_batch_code,
    updated_at = now();

COMMIT;
