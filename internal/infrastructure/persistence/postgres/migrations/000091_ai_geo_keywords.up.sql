CREATE TABLE IF NOT EXISTS ai_geo_keywords (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    brand_id BIGINT,
    product_id BIGINT,
    keyword_group VARCHAR(120) NOT NULL DEFAULT '通用关键词',
    keyword VARCHAR(160) NOT NULL,
    intent VARCHAR(80),
    source VARCHAR(80) NOT NULL DEFAULT 'manual',
    weight INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_keyword_scope
    ON ai_geo_keywords (
        tenant_id,
        COALESCE(brand_id, 0),
        COALESCE(product_id, 0),
        keyword_group,
        keyword
    )
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ai_geo_keyword_tenant_brand
    ON ai_geo_keywords (tenant_id, brand_id, product_id, status)
    WHERE deleted_at IS NULL;
