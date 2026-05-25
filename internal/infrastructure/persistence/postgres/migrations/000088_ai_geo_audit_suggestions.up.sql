CREATE TABLE IF NOT EXISTS ai_geo_audit_suggestions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    object_type VARCHAR(40) NOT NULL,
    object_id BIGINT NOT NULL,
    object_code VARCHAR(120),
    scenario_code VARCHAR(120) NOT NULL,
    risk_level VARCHAR(32) NOT NULL DEFAULT 'low',
    passed BOOLEAN NOT NULL DEFAULT FALSE,
    summary TEXT,
    suggestion_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    model_code VARCHAR(120),
    status VARCHAR(32) NOT NULL DEFAULT 'success',
    error_message TEXT,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ai_geo_audit_suggestion_tenant_object
    ON ai_geo_audit_suggestions (tenant_id, object_type, object_id, generated_at DESC)
    WHERE deleted_at IS NULL;
