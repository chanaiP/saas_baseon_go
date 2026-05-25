DROP INDEX IF EXISTS idx_ai_geo_hotspot_source;
DROP INDEX IF EXISTS idx_ai_geo_style_template_source;
DROP INDEX IF EXISTS idx_ai_geo_style_template_tenant_status;
DROP INDEX IF EXISTS uq_ai_geo_style_template_tenant_code;
DROP INDEX IF EXISTS idx_ai_geo_external_source_tenant_time;
DROP INDEX IF EXISTS idx_ai_geo_external_source_tenant_hash;
DROP INDEX IF EXISTS uq_ai_geo_external_source_tenant_url;

ALTER TABLE ai_geo_hotspots
    DROP COLUMN IF EXISTS source_id;

DROP TABLE IF EXISTS ai_geo_style_templates;
DROP TABLE IF EXISTS ai_geo_external_sources;
