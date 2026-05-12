ALTER TABLE sys_app
  ADD COLUMN IF NOT EXISTS owner_user_ids TEXT;
