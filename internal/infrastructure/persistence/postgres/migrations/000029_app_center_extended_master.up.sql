ALTER TABLE sys_app
  ADD COLUMN IF NOT EXISTS visibility_mode VARCHAR(64),
  ADD COLUMN IF NOT EXISTS visible_tenants TEXT,
  ADD COLUMN IF NOT EXISTS open_method VARCHAR(200),
  ADD COLUMN IF NOT EXISTS trial_policy VARCHAR(100),
  ADD COLUMN IF NOT EXISTS trial_start_rule VARCHAR(100),
  ADD COLUMN IF NOT EXISTS asset_config TEXT,
  ADD COLUMN IF NOT EXISTS doc_config TEXT,
  ADD COLUMN IF NOT EXISTS release_channel VARCHAR(32),
  ADD COLUMN IF NOT EXISTS release_note TEXT;

CREATE TABLE IF NOT EXISTS sys_app_client (
  id BIGSERIAL PRIMARY KEY,
  app_id BIGINT NOT NULL REFERENCES sys_app(id),
  client_code VARCHAR(64) NOT NULL,
  client_name VARCHAR(100) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order INTEGER NOT NULL DEFAULT 0,
  config_note VARCHAR(500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_app_client_app_code
  ON sys_app_client (app_id, client_code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sys_app_client_app_enabled
  ON sys_app_client (app_id, enabled, deleted_at);
