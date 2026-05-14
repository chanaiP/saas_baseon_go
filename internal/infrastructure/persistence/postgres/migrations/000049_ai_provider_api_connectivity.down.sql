DROP INDEX IF EXISTS idx_ai_provider_apis_health;

ALTER TABLE ai_provider_apis
  DROP COLUMN IF EXISTS health_checked_at,
  DROP COLUMN IF EXISTS health_message,
  DROP COLUMN IF EXISTS health_status;
