ALTER TABLE integration_sync_jobs
  ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_integration_sync_jobs_due
  ON integration_sync_jobs(status, next_retry_at, created_at)
  WHERE deleted_at IS NULL;
