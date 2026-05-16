ALTER TABLE integration_api_call_logs
    ADD COLUMN IF NOT EXISTS trace_id VARCHAR(120);

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_trace
    ON integration_api_call_logs(trace_id)
    WHERE trace_id IS NOT NULL;
