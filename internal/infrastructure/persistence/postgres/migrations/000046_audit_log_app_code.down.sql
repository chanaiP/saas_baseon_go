DROP INDEX IF EXISTS idx_audit_log_app_code_created;

ALTER TABLE audit_log
  DROP COLUMN IF EXISTS app_code;
