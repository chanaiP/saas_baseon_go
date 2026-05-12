DROP TABLE IF EXISTS sys_app_client;

ALTER TABLE sys_app
  DROP COLUMN IF EXISTS release_note,
  DROP COLUMN IF EXISTS release_channel,
  DROP COLUMN IF EXISTS doc_config,
  DROP COLUMN IF EXISTS asset_config,
  DROP COLUMN IF EXISTS trial_start_rule,
  DROP COLUMN IF EXISTS trial_policy,
  DROP COLUMN IF EXISTS open_method,
  DROP COLUMN IF EXISTS visible_tenants,
  DROP COLUMN IF EXISTS visibility_mode;
