ALTER TABLE app_user
  DROP COLUMN IF EXISTS password_changed_at,
  DROP COLUMN IF EXISTS session_version;
