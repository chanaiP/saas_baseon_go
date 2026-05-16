ALTER TABLE ai_usage_records
  ADD COLUMN IF NOT EXISTS data_source VARCHAR(32) NOT NULL DEFAULT 'gateway',
  ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;

UPDATE ai_usage_records
SET data_source = 'demo_seed',
    is_demo = true
WHERE request_id LIKE 'seed%'
   OR request_params::text LIKE '%demo-seed%'
   OR request_params::text LIKE '%demo-history-seed%';

CREATE INDEX IF NOT EXISTS idx_ai_usage_demo_source_time
  ON ai_usage_records(is_demo, data_source, called_at DESC);
