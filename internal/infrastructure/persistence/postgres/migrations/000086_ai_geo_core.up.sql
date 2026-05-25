CREATE TABLE IF NOT EXISTS ai_geo_brand_cards (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    company_id BIGINT,
    department_id BIGINT,
    brand_code VARCHAR(80) NOT NULL,
    brand_name VARCHAR(160) NOT NULL,
    positioning TEXT,
    target_audience TEXT,
    price_band VARCHAR(120),
    tone VARCHAR(160),
    keywords JSONB NOT NULL DEFAULT '[]'::jsonb,
    completeness INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_product_cards (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    company_id BIGINT,
    department_id BIGINT,
    brand_id BIGINT NOT NULL,
    product_code VARCHAR(100) NOT NULL,
    product_name VARCHAR(180) NOT NULL,
    category_name VARCHAR(160),
    selling_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    faq JSONB NOT NULL DEFAULT '[]'::jsonb,
    content_angles JSONB NOT NULL DEFAULT '[]'::jsonb,
    completeness INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_skus (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    sku_code VARCHAR(100) NOT NULL,
    sku_name VARCHAR(180) NOT NULL,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    price NUMERIC(18,2) NOT NULL DEFAULT 0,
    image_url TEXT,
    stock_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_competitors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    brand_name VARCHAR(160) NOT NULL,
    product_name VARCHAR(180) NOT NULL,
    price_text VARCHAR(120),
    point TEXT,
    difference TEXT,
    angle TEXT,
    link_url TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_channel_profiles (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    company_id BIGINT,
    department_id BIGINT,
    channel_code VARCHAR(80) NOT NULL,
    channel_name VARCHAR(120) NOT NULL,
    channel_type VARCHAR(80) NOT NULL DEFAULT 'content',
    entry_url TEXT,
    content_forms JSONB NOT NULL DEFAULT '[]'::jsonb,
    support_modes JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_publish_mode VARCHAR(80) NOT NULL DEFAULT 'manual',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_channel_accounts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    account_name VARCHAR(160) NOT NULL,
    external_account_id VARCHAR(160),
    auth_status VARCHAR(32) NOT NULL DEFAULT 'not_authorized',
    publish_status VARCHAR(32) NOT NULL DEFAULT 'unavailable',
    expires_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_drafts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    company_id BIGINT,
    department_id BIGINT,
    draft_code VARCHAR(100) NOT NULL,
    brand_id BIGINT,
    product_id BIGINT,
    title VARCHAR(240) NOT NULL,
    summary TEXT,
    body TEXT NOT NULL,
    keywords JSONB NOT NULL DEFAULT '[]'::jsonb,
    source VARCHAR(80) NOT NULL DEFAULT 'manual',
    audit_status VARCHAR(32) NOT NULL DEFAULT 'approved',
    channel_status VARCHAR(32) NOT NULL DEFAULT 'not_generated',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_channel_contents (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    draft_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    title VARCHAR(240) NOT NULL,
    body TEXT NOT NULL,
    audit_status VARCHAR(32) NOT NULL DEFAULT 'approved',
    publish_status VARCHAR(32) NOT NULL DEFAULT 'not_planned',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_publish_plans (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    plan_code VARCHAR(100) NOT NULL,
    channel_content_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    publish_method VARCHAR(80) NOT NULL DEFAULT 'manual',
    automation_level VARCHAR(80) NOT NULL DEFAULT 'manual',
    status VARCHAR(32) NOT NULL DEFAULT 'scheduled',
    published_url TEXT,
    fail_reason TEXT,
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_import_batches (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    batch_code VARCHAR(100) NOT NULL,
    import_type VARCHAR(80) NOT NULL,
    mapping_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    record_count BIGINT NOT NULL DEFAULT 0,
    success_count BIGINT NOT NULL DEFAULT 0,
    failed_count BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_brand_tenant_code ON ai_geo_brand_cards (tenant_id, brand_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_product_tenant_code ON ai_geo_product_cards (tenant_id, product_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_sku_tenant_code ON ai_geo_skus (tenant_id, sku_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_channel_tenant_code ON ai_geo_channel_profiles (tenant_id, channel_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_account_tenant_name ON ai_geo_channel_accounts (tenant_id, account_name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_draft_tenant_code ON ai_geo_drafts (tenant_id, draft_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_plan_tenant_code ON ai_geo_publish_plans (tenant_id, plan_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_import_tenant_code ON ai_geo_import_batches (tenant_id, batch_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ai_geo_material_assets (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    brand_id BIGINT,
    product_id BIGINT,
    asset_type VARCHAR(80) NOT NULL,
    asset_name VARCHAR(180) NOT NULL,
    file_id BIGINT,
    url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_hotspots (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    platform VARCHAR(80) NOT NULL,
    title VARCHAR(240) NOT NULL,
    heat_score INTEGER NOT NULL DEFAULT 0,
    source_url TEXT,
    captured_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ai_geo_brand_tenant_status ON ai_geo_brand_cards (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_product_tenant_brand ON ai_geo_product_cards (tenant_id, brand_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_sku_tenant_product ON ai_geo_skus (tenant_id, product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_competitor_tenant_product ON ai_geo_competitors (tenant_id, product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_channel_tenant_status ON ai_geo_channel_profiles (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_account_tenant_channel ON ai_geo_channel_accounts (tenant_id, channel_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_draft_tenant_status ON ai_geo_drafts (tenant_id, audit_status, channel_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_content_tenant_draft ON ai_geo_channel_contents (tenant_id, draft_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_plan_tenant_time ON ai_geo_publish_plans (tenant_id, scheduled_at, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_import_tenant_status ON ai_geo_import_batches (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_asset_tenant_product ON ai_geo_material_assets (tenant_id, product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_hotspot_tenant_time ON ai_geo_hotspots (tenant_id, captured_at DESC) WHERE deleted_at IS NULL;
