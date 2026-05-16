-- Core domain schema for the third-party integration center.

CREATE TABLE IF NOT EXISTS integration_platforms (
  id BIGSERIAL PRIMARY KEY,
  platform_code VARCHAR(80) NOT NULL,
  platform_name VARCHAR(120) NOT NULL,
  platform_short_name VARCHAR(80),
  platform_type VARCHAR(50) NOT NULL,
  access_mode VARCHAR(50) NOT NULL,
  logo_url VARCHAR(500),
  official_url VARCHAR(500),
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  tenant_visible BOOLEAN NOT NULL DEFAULT FALSE,
  owner_name VARCHAR(80),
  sort_order INT NOT NULL DEFAULT 0,
  description TEXT,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_platforms_code
  ON integration_platforms(platform_code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_platforms_status
  ON integration_platforms(status, sort_order)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_platform_capabilities (
  id BIGSERIAL PRIMARY KEY,
  platform_id BIGINT NOT NULL REFERENCES integration_platforms(id),
  capability_code VARCHAR(100) NOT NULL,
  capability_name VARCHAR(120) NOT NULL,
  capability_type VARCHAR(50) NOT NULL,
  auth_scope_code VARCHAR(120),
  data_direction VARCHAR(32) NOT NULL DEFAULT 'pull',
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  description TEXT,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_platform_capabilities_code
  ON integration_platform_capabilities(platform_id, capability_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_provider_apps (
  id BIGSERIAL PRIMARY KEY,
  platform_id BIGINT NOT NULL REFERENCES integration_platforms(id),
  app_code VARCHAR(100) NOT NULL,
  app_name VARCHAR(150) NOT NULL,
  app_type VARCHAR(50) NOT NULL DEFAULT 'provider_app',
  auth_mode VARCHAR(50) NOT NULL,
  environment VARCHAR(32) NOT NULL DEFAULT 'prod',
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  tenant_visible BOOLEAN NOT NULL DEFAULT FALSE,
  callback_url VARCHAR(500),
  webhook_url VARCHAR(500),
  credential_ref VARCHAR(200),
  owner_name VARCHAR(80),
  description TEXT,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_provider_apps_code
  ON integration_provider_apps(platform_id, app_code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_provider_apps_status
  ON integration_provider_apps(platform_id, status)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_provider_app_capabilities (
  id BIGSERIAL PRIMARY KEY,
  provider_app_id BIGINT NOT NULL REFERENCES integration_provider_apps(id),
  platform_capability_id BIGINT NOT NULL REFERENCES integration_platform_capabilities(id),
  connection_status VARCHAR(32) NOT NULL DEFAULT 'pending',
  review_status VARCHAR(32) NOT NULL DEFAULT 'pending',
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  config JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_provider_app_capabilities
  ON integration_provider_app_capabilities(provider_app_id, platform_capability_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_tenant_connections (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES tenant(id),
  platform_id BIGINT NOT NULL REFERENCES integration_platforms(id),
  provider_app_id BIGINT NOT NULL REFERENCES integration_provider_apps(id),
  connection_name VARCHAR(180) NOT NULL,
  auth_subject_type VARCHAR(60) NOT NULL,
  auth_subject_id VARCHAR(160) NOT NULL,
  auth_subject_name VARCHAR(180) NOT NULL,
  auth_scope JSONB NOT NULL DEFAULT '[]'::jsonb,
  auth_status VARCHAR(32) NOT NULL DEFAULT 'pending',
  connection_status VARCHAR(32) NOT NULL DEFAULT 'inactive',
  token_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  token_credential_ref VARCHAR(240),
  authorized_at TIMESTAMPTZ,
  token_expires_at TIMESTAMPTZ,
  last_sync_at TIMESTAMPTZ,
  last_error_at TIMESTAMPTZ,
  last_error_message TEXT,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_tenant_connection_subject
  ON integration_tenant_connections(tenant_id, platform_id, provider_app_id, auth_subject_type, auth_subject_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_tenant_connections_tenant_status
  ON integration_tenant_connections(tenant_id, connection_status, updated_at DESC)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_tenant_capabilities (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES tenant(id),
  tenant_connection_id BIGINT NOT NULL REFERENCES integration_tenant_connections(id),
  provider_app_capability_id BIGINT NOT NULL REFERENCES integration_provider_app_capabilities(id),
  capability_code VARCHAR(100) NOT NULL,
  capability_name VARCHAR(120) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  source VARCHAR(32) NOT NULL DEFAULT 'authorization',
  effective_scope JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_tenant_capabilities
  ON integration_tenant_capabilities(tenant_connection_id, capability_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_sync_jobs (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES tenant(id),
  tenant_connection_id BIGINT NOT NULL REFERENCES integration_tenant_connections(id),
  capability_code VARCHAR(100) NOT NULL,
  job_type VARCHAR(50) NOT NULL,
  trigger_mode VARCHAR(50) NOT NULL DEFAULT 'manual',
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  cursor_value VARCHAR(300),
  total_count BIGINT NOT NULL DEFAULT 0,
  success_count BIGINT NOT NULL DEFAULT 0,
  failed_count BIGINT NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  next_retry_at TIMESTAMPTZ,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  error_code VARCHAR(100),
  error_message TEXT,
  created_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_integration_sync_jobs_connection_time
  ON integration_sync_jobs(tenant_connection_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_quota_policies (
  id BIGSERIAL PRIMARY KEY,
  policy_code VARCHAR(100) NOT NULL,
  policy_name VARCHAR(150) NOT NULL,
  quota_code VARCHAR(100) NOT NULL,
  quota_unit VARCHAR(32) NOT NULL,
  period_type VARCHAR(32) NOT NULL,
  default_limit BIGINT NOT NULL DEFAULT 0,
  over_limit_action VARCHAR(32) NOT NULL DEFAULT 'reject',
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  description TEXT,
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_quota_policies_code
  ON integration_quota_policies(policy_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_quota_bindings (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT REFERENCES tenant(id),
  platform_id BIGINT REFERENCES integration_platforms(id),
  provider_app_id BIGINT REFERENCES integration_provider_apps(id),
  tenant_connection_id BIGINT REFERENCES integration_tenant_connections(id),
  policy_id BIGINT NOT NULL REFERENCES integration_quota_policies(id),
  override_limit BIGINT,
  priority INTEGER NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  created_by BIGINT,
  updated_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_integration_quota_bindings_scope
  ON integration_quota_bindings(tenant_id, platform_id, provider_app_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_quota_bindings_connection
  ON integration_quota_bindings(tenant_connection_id, priority)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_quota_usages (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES tenant(id),
  tenant_connection_id BIGINT REFERENCES integration_tenant_connections(id),
  quota_code VARCHAR(100) NOT NULL,
  period_key VARCHAR(64) NOT NULL,
  used_amount BIGINT NOT NULL DEFAULT 0,
  limited_count BIGINT NOT NULL DEFAULT 0,
  last_used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_quota_usages_period
  ON integration_quota_usages(tenant_id, COALESCE(tenant_connection_id, 0), quota_code, period_key);

CREATE TABLE IF NOT EXISTS integration_alerts (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT REFERENCES tenant(id),
  tenant_connection_id BIGINT REFERENCES integration_tenant_connections(id),
  platform_id BIGINT REFERENCES integration_platforms(id),
  provider_app_id BIGINT REFERENCES integration_provider_apps(id),
  alert_type VARCHAR(60) NOT NULL,
  severity VARCHAR(32) NOT NULL DEFAULT 'warning',
  status VARCHAR(32) NOT NULL DEFAULT 'open',
  title VARCHAR(180) NOT NULL,
  message TEXT,
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  resolved_at TIMESTAMPTZ,
  handled_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_integration_alerts_status
  ON integration_alerts(status, severity, last_seen_at DESC)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS integration_api_call_logs (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT REFERENCES tenant(id),
  tenant_connection_id BIGINT REFERENCES integration_tenant_connections(id),
  platform_id BIGINT REFERENCES integration_platforms(id),
  provider_app_id BIGINT REFERENCES integration_provider_apps(id),
  request_id VARCHAR(120) NOT NULL,
  call_type VARCHAR(60) NOT NULL,
  method VARCHAR(16),
  endpoint VARCHAR(500),
  status VARCHAR(32) NOT NULL,
  http_status INT,
  duration_ms INT NOT NULL DEFAULT 0,
  error_code VARCHAR(120),
  error_message TEXT,
  request_digest VARCHAR(128),
  response_digest VARCHAR(128),
  called_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_api_call_logs_request
  ON integration_api_call_logs(request_id);

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_scope_time
  ON integration_api_call_logs(tenant_id, platform_id, called_at DESC);
