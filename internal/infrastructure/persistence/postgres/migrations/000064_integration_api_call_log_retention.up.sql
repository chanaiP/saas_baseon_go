ALTER TABLE integration_api_call_logs
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS retention_bucket VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_active_called
    ON integration_api_call_logs(called_at DESC)
    WHERE archived_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_archived
    ON integration_api_call_logs(archived_at DESC, retention_bucket)
    WHERE archived_at IS NOT NULL;
