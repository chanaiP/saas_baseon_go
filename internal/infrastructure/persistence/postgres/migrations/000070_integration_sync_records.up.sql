CREATE TABLE IF NOT EXISTS integration_sync_records (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  sync_job_id BIGINT NOT NULL,
  tenant_connection_id BIGINT NOT NULL,
  capability_code VARCHAR(100) NOT NULL,
  external_id VARCHAR(180) NOT NULL,
  payload_digest VARCHAR(128) NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  status VARCHAR(32) NOT NULL DEFAULT 'written',
  cursor_value VARCHAR(300),
  written_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_integration_sync_records_unique_external
  ON integration_sync_records(tenant_connection_id, capability_code, external_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_sync_records_job_time
  ON integration_sync_records(sync_job_id, written_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_sync_records_tenant_time
  ON integration_sync_records(tenant_id, written_at DESC)
  WHERE deleted_at IS NULL;

ALTER TABLE integration_sync_records
  ADD CONSTRAINT fk_integration_sync_records_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;

ALTER TABLE integration_sync_records
  ADD CONSTRAINT fk_integration_sync_records_job
  FOREIGN KEY (sync_job_id) REFERENCES integration_sync_jobs(id)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;

ALTER TABLE integration_sync_records
  ADD CONSTRAINT fk_integration_sync_records_connection
  FOREIGN KEY (tenant_connection_id) REFERENCES integration_tenant_connections(id)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;
