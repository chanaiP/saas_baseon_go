DROP INDEX IF EXISTS idx_integration_api_call_logs_archived;
DROP INDEX IF EXISTS idx_integration_api_call_logs_active_called;

ALTER TABLE integration_api_call_logs
    DROP COLUMN IF EXISTS retention_bucket,
    DROP COLUMN IF EXISTS archived_at;
