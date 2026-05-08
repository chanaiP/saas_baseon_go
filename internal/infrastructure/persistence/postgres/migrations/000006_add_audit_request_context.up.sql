ALTER TABLE audit_log
  ADD COLUMN IF NOT EXISTS user_agent VARCHAR(500),
  ADD COLUMN IF NOT EXISTS request_id VARCHAR(64),
  ADD COLUMN IF NOT EXISTS result VARCHAR(32) NOT NULL DEFAULT 'success';

CREATE INDEX IF NOT EXISTS idx_audit_log_request_id ON audit_log (request_id);
