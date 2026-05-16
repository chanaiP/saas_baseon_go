DROP INDEX IF EXISTS idx_integration_quota_bindings_connection;

ALTER TABLE integration_quota_bindings
  DROP COLUMN IF EXISTS priority,
  DROP COLUMN IF EXISTS tenant_connection_id;
