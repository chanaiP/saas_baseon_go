DROP INDEX IF EXISTS idx_integration_sync_jobs_due;

ALTER TABLE integration_sync_jobs
  DROP COLUMN IF EXISTS next_retry_at,
  DROP COLUMN IF EXISTS retry_count;
