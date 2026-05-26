ALTER TABLE ai_usage_records
  ADD COLUMN IF NOT EXISTS task_duration_ms INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN ai_usage_records.task_duration_ms IS 'Async task total duration in milliseconds, measured from provider request start to terminal task status.';
