ALTER TABLE app_user
  ADD COLUMN IF NOT EXISTS session_version BIGINT NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;

UPDATE app_user
SET session_version = 1
WHERE session_version IS NULL OR session_version < 1;
