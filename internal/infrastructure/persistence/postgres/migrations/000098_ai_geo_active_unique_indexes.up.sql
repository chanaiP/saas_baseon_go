ALTER TABLE ai_geo_brand_cards DROP CONSTRAINT IF EXISTS uq_ai_geo_brand_tenant_code;
ALTER TABLE ai_geo_product_cards DROP CONSTRAINT IF EXISTS uq_ai_geo_product_tenant_code;
ALTER TABLE ai_geo_skus DROP CONSTRAINT IF EXISTS uq_ai_geo_sku_tenant_code;
ALTER TABLE ai_geo_channel_profiles DROP CONSTRAINT IF EXISTS uq_ai_geo_channel_tenant_code;
ALTER TABLE ai_geo_channel_accounts DROP CONSTRAINT IF EXISTS uq_ai_geo_account_tenant_name;
ALTER TABLE ai_geo_drafts DROP CONSTRAINT IF EXISTS uq_ai_geo_draft_tenant_code;
ALTER TABLE ai_geo_publish_plans DROP CONSTRAINT IF EXISTS uq_ai_geo_plan_tenant_code;
ALTER TABLE ai_geo_import_batches DROP CONSTRAINT IF EXISTS uq_ai_geo_import_tenant_code;

CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_brand_tenant_code ON ai_geo_brand_cards (tenant_id, brand_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_product_tenant_code ON ai_geo_product_cards (tenant_id, product_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_sku_tenant_code ON ai_geo_skus (tenant_id, sku_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_channel_tenant_code ON ai_geo_channel_profiles (tenant_id, channel_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_account_tenant_name ON ai_geo_channel_accounts (tenant_id, account_name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_draft_tenant_code ON ai_geo_drafts (tenant_id, draft_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_plan_tenant_code ON ai_geo_publish_plans (tenant_id, plan_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_import_tenant_code ON ai_geo_import_batches (tenant_id, batch_code) WHERE deleted_at IS NULL;
