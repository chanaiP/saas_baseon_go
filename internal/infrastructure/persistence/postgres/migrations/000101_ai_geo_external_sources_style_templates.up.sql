CREATE TABLE IF NOT EXISTS ai_geo_external_sources (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    source_type VARCHAR(40) NOT NULL DEFAULT 'url',
    source_url TEXT NOT NULL,
    source_site VARCHAR(160),
    source_title VARCHAR(240),
    raw_text TEXT NOT NULL DEFAULT '',
    clean_text TEXT NOT NULL DEFAULT '',
    content_hash VARCHAR(64) NOT NULL,
    extracted_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    extraction_status VARCHAR(32) NOT NULL DEFAULT 'success',
    extraction_error TEXT,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_geo_style_templates (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    source_id BIGINT,
    template_code VARCHAR(100) NOT NULL,
    template_name VARCHAR(160) NOT NULL,
    description TEXT,
    content_type VARCHAR(80),
    platform VARCHAR(80),
    tone_profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    structure_profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    technique_profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    style_keywords JSONB NOT NULL DEFAULT '[]'::jsonb,
    prompt_fragment TEXT NOT NULL DEFAULT '',
    negative_rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    extraction_summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE ai_geo_hotspots
    ADD COLUMN IF NOT EXISTS source_id BIGINT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_external_source_tenant_url
    ON ai_geo_external_sources (tenant_id, source_url)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_external_source_tenant_hash
    ON ai_geo_external_sources (tenant_id, content_hash)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_external_source_tenant_time
    ON ai_geo_external_sources (tenant_id, captured_at DESC)
    WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_geo_style_template_tenant_code
    ON ai_geo_style_templates (tenant_id, template_code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_style_template_tenant_status
    ON ai_geo_style_templates (tenant_id, status, updated_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_style_template_source
    ON ai_geo_style_templates (tenant_id, source_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_geo_hotspot_source
    ON ai_geo_hotspots (tenant_id, source_id)
    WHERE deleted_at IS NULL;
