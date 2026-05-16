ALTER TABLE integration_webhook_events
  ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_integration_webhook_events_retry
  ON integration_webhook_events(status, next_retry_at, received_at)
  WHERE deleted_at IS NULL;
