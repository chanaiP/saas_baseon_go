DROP INDEX IF EXISTS idx_ai_usage_demo_source_time;

ALTER TABLE ai_usage_records
  DROP COLUMN IF EXISTS is_demo,
  DROP COLUMN IF EXISTS data_source;
