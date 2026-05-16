DROP INDEX IF EXISTS idx_ai_usage_provider_request_id;

ALTER TABLE ai_usage_records
  DROP COLUMN IF EXISTS retry_count,
  DROP COLUMN IF EXISTS finished_at,
  DROP COLUMN IF EXISTS started_at,
  DROP COLUMN IF EXISTS provider_request_id,
  DROP COLUMN IF EXISTS provider_http_status;
