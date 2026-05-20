CREATE TABLE IF NOT EXISTS ai_geo_import_errors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    batch_id BIGINT NOT NULL,
    row_number INTEGER NOT NULL,
    field_name VARCHAR(120),
    error_code VARCHAR(80) NOT NULL,
    error_message TEXT NOT NULL,
    raw_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ai_geo_import_error_tenant_batch
    ON ai_geo_import_errors (tenant_id, batch_id, row_number)
    WHERE deleted_at IS NULL;
