CREATE INDEX IF NOT EXISTS idx_sys_app_deployment_mode_deleted
  ON sys_app (deployment_mode, deleted_at);
