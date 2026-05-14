ALTER TABLE ai_provider_apis
  ADD COLUMN IF NOT EXISTS health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  ADD COLUMN IF NOT EXISTS health_message VARCHAR(500),
  ADD COLUMN IF NOT EXISTS health_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_ai_provider_apis_health
  ON ai_provider_apis(health_status, health_checked_at)
  WHERE deleted_at IS NULL;
