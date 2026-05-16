ALTER TABLE ai_usage_records
  ADD COLUMN IF NOT EXISTS provider_http_status INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS provider_request_id VARCHAR(200),
  ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_ai_usage_provider_request_id
  ON ai_usage_records(provider_request_id)
  WHERE provider_request_id IS NOT NULL AND provider_request_id <> '';
