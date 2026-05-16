DROP INDEX IF EXISTS idx_integration_api_call_logs_trace;

ALTER TABLE integration_api_call_logs
    DROP COLUMN IF EXISTS trace_id;
