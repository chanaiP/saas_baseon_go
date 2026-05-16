DROP INDEX IF EXISTS idx_integration_webhook_events_retry;
DROP INDEX IF EXISTS idx_integration_webhook_events_status;
DROP INDEX IF EXISTS uk_integration_webhook_events_idempotency;
DROP TABLE IF EXISTS integration_webhook_events;
