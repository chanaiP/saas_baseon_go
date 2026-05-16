ALTER TABLE integration_quota_bindings
  ADD COLUMN IF NOT EXISTS tenant_connection_id BIGINT REFERENCES integration_tenant_connections(id),
  ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_integration_quota_bindings_connection
  ON integration_quota_bindings(tenant_connection_id, priority)
  WHERE deleted_at IS NULL;
