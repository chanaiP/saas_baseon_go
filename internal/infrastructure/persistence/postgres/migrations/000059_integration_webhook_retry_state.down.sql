DROP INDEX IF EXISTS idx_integration_webhook_events_retry;

ALTER TABLE integration_webhook_events
  DROP COLUMN IF EXISTS next_retry_at,
  DROP COLUMN IF EXISTS retry_count;
