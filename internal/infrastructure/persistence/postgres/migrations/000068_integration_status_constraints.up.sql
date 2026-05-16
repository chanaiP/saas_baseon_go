ALTER TABLE integration_platforms
  ADD CONSTRAINT ck_integration_platforms_status
  CHECK (status IN ('online', 'beta', 'draft', 'disabled', 'maintenance')) NOT VALID;

ALTER TABLE integration_provider_apps
  ADD CONSTRAINT ck_integration_provider_apps_status
  CHECK (status IN ('online', 'beta', 'draft', 'disabled', 'maintenance')) NOT VALID;

ALTER TABLE integration_platform_capabilities
  ADD CONSTRAINT ck_integration_platform_capabilities_status
  CHECK (status IN ('enabled', 'disabled')) NOT VALID;

ALTER TABLE integration_provider_app_capabilities
  ADD CONSTRAINT ck_integration_provider_app_capabilities_connection_status
  CHECK (connection_status IN ('connected', 'pending', 'failed', 'paused', 'not_connected')) NOT VALID;

ALTER TABLE integration_provider_app_capabilities
  ADD CONSTRAINT ck_integration_provider_app_capabilities_review_status
  CHECK (review_status IN ('pending', 'approved', 'rejected')) NOT VALID;

ALTER TABLE integration_tenant_connections
  ADD CONSTRAINT ck_integration_tenant_connections_auth_status
  CHECK (auth_status IN ('pending', 'authorized', 'failed', 'expired', 'revoked')) NOT VALID;

ALTER TABLE integration_tenant_connections
  ADD CONSTRAINT ck_integration_tenant_connections_connection_status
  CHECK (connection_status IN ('inactive', 'connected', 'paused', 'failed', 'not_connected')) NOT VALID;

ALTER TABLE integration_tenant_connections
  ADD CONSTRAINT ck_integration_tenant_connections_token_status
  CHECK (token_status IN ('unknown', 'valid', 'expired', 'refresh_failed', 'revoked')) NOT VALID;

ALTER TABLE integration_sync_jobs
  ADD CONSTRAINT ck_integration_sync_jobs_status
  CHECK (status IN ('pending', 'running', 'completed', 'failed', 'queued', 'retrying', 'paused')) NOT VALID;

ALTER TABLE integration_quota_policies
  ADD CONSTRAINT ck_integration_quota_policies_status
  CHECK (status IN ('enabled', 'disabled')) NOT VALID;

ALTER TABLE integration_quota_bindings
  ADD CONSTRAINT ck_integration_quota_bindings_status
  CHECK (status IN ('enabled', 'disabled')) NOT VALID;

ALTER TABLE integration_alerts
  ADD CONSTRAINT ck_integration_alerts_status
  CHECK (status IN ('open', 'processing', 'resolved', 'ignored')) NOT VALID;

ALTER TABLE integration_api_call_logs
  ADD CONSTRAINT ck_integration_api_call_logs_status
  CHECK (status IN ('success', 'failed', 'limited')) NOT VALID;

ALTER TABLE integration_webhook_events
  ADD CONSTRAINT ck_integration_webhook_events_status
  CHECK (status IN ('received', 'duplicated', 'processed', 'retrying', 'dead_letter')) NOT VALID;
