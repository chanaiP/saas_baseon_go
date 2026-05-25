DROP INDEX IF EXISTS uq_ai_geo_brand_tenant_code;
DROP INDEX IF EXISTS uq_ai_geo_product_tenant_code;
DROP INDEX IF EXISTS uq_ai_geo_sku_tenant_code;
DROP INDEX IF EXISTS uq_ai_geo_channel_tenant_code;
DROP INDEX IF EXISTS uq_ai_geo_account_tenant_name;
DROP INDEX IF EXISTS uq_ai_geo_draft_tenant_code;
DROP INDEX IF EXISTS uq_ai_geo_plan_tenant_code;
DROP INDEX IF EXISTS uq_ai_geo_import_tenant_code;

ALTER TABLE ai_geo_brand_cards ADD CONSTRAINT uq_ai_geo_brand_tenant_code UNIQUE (tenant_id, brand_code);
ALTER TABLE ai_geo_product_cards ADD CONSTRAINT uq_ai_geo_product_tenant_code UNIQUE (tenant_id, product_code);
ALTER TABLE ai_geo_skus ADD CONSTRAINT uq_ai_geo_sku_tenant_code UNIQUE (tenant_id, sku_code);
ALTER TABLE ai_geo_channel_profiles ADD CONSTRAINT uq_ai_geo_channel_tenant_code UNIQUE (tenant_id, channel_code);
ALTER TABLE ai_geo_channel_accounts ADD CONSTRAINT uq_ai_geo_account_tenant_name UNIQUE (tenant_id, account_name);
ALTER TABLE ai_geo_drafts ADD CONSTRAINT uq_ai_geo_draft_tenant_code UNIQUE (tenant_id, draft_code);
ALTER TABLE ai_geo_publish_plans ADD CONSTRAINT uq_ai_geo_plan_tenant_code UNIQUE (tenant_id, plan_code);
ALTER TABLE ai_geo_import_batches ADD CONSTRAINT uq_ai_geo_import_tenant_code UNIQUE (tenant_id, batch_code);
