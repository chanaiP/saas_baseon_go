CREATE INDEX IF NOT EXISTS idx_integration_platforms_updated
    ON integration_platforms(updated_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_provider_apps_platform_updated
    ON integration_provider_apps(platform_id, updated_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_provider_apps_code_status
    ON integration_provider_apps(app_code, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_platform_capabilities_platform_status
    ON integration_platform_capabilities(platform_id, status, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_provider_app_capabilities_app_status
    ON integration_provider_app_capabilities(provider_app_id, connection_status, review_status, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_tenant_connections_app_status
    ON integration_tenant_connections(provider_app_id, connection_status, token_status, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_sync_jobs_tenant_status_time
    ON integration_sync_jobs(tenant_id, status, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_quota_usages_tenant_period
    ON integration_quota_usages(tenant_id, quota_code, period_key, last_used_at DESC);

CREATE INDEX IF NOT EXISTS idx_integration_alerts_tenant_status_time
    ON integration_alerts(tenant_id, status, last_seen_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_request
    ON integration_api_call_logs(request_id);

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_connection_time
    ON integration_api_call_logs(tenant_connection_id, called_at DESC);

CREATE INDEX IF NOT EXISTS idx_integration_api_call_logs_app_time
    ON integration_api_call_logs(provider_app_id, called_at DESC);
