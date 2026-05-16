-- Integration Center tenant orphan preflight.
-- Run before VALIDATE CONSTRAINT in production if historical data may exist.

SELECT 'integration_tenant_connections' AS table_name, COUNT(*) AS orphan_count
FROM integration_tenant_connections i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE t.id IS NULL;

SELECT 'integration_oauth_states' AS table_name, COUNT(*) AS orphan_count
FROM integration_oauth_states i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE t.id IS NULL;

SELECT 'integration_tenant_capabilities' AS table_name, COUNT(*) AS orphan_count
FROM integration_tenant_capabilities i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE t.id IS NULL;

SELECT 'integration_sync_jobs' AS table_name, COUNT(*) AS orphan_count
FROM integration_sync_jobs i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE t.id IS NULL;

SELECT 'integration_quota_bindings' AS table_name, COUNT(*) AS orphan_count
FROM integration_quota_bindings i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE i.tenant_id IS NOT NULL AND t.id IS NULL;

SELECT 'integration_quota_usages' AS table_name, COUNT(*) AS orphan_count
FROM integration_quota_usages i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE t.id IS NULL;

SELECT 'integration_alerts' AS table_name, COUNT(*) AS orphan_count
FROM integration_alerts i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE i.tenant_id IS NOT NULL AND t.id IS NULL;

SELECT 'integration_api_call_logs' AS table_name, COUNT(*) AS orphan_count
FROM integration_api_call_logs i
LEFT JOIN tenant t ON t.id = i.tenant_id
WHERE i.tenant_id IS NOT NULL AND t.id IS NULL;

SELECT 'integration_provider_apps.platform_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_provider_apps i
LEFT JOIN integration_platforms p ON p.id = i.platform_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;

SELECT 'integration_platform_capabilities.platform_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_platform_capabilities i
LEFT JOIN integration_platforms p ON p.id = i.platform_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;

SELECT 'integration_provider_app_capabilities.provider_app_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_provider_app_capabilities i
LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;

SELECT 'integration_provider_app_capabilities.platform_capability_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_provider_app_capabilities i
LEFT JOIN integration_platform_capabilities c ON c.id = i.platform_capability_id AND c.deleted_at IS NULL
WHERE c.id IS NULL;

SELECT 'integration_tenant_connections.platform_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_tenant_connections i
LEFT JOIN integration_platforms p ON p.id = i.platform_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;

SELECT 'integration_tenant_connections.provider_app_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_tenant_connections i
LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;

SELECT 'integration_tenant_capabilities.connection_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_tenant_capabilities i
LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL
WHERE c.id IS NULL;

SELECT 'integration_sync_jobs.connection_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_sync_jobs i
LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL
WHERE c.id IS NULL;

SELECT 'integration_quota_bindings.policy_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_quota_bindings i
LEFT JOIN integration_quota_policies p ON p.id = i.policy_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;

SELECT 'integration_alerts.connection_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_alerts i
LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL
WHERE i.tenant_connection_id IS NOT NULL AND c.id IS NULL;

SELECT 'integration_api_call_logs.connection_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_api_call_logs i
LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL
WHERE i.tenant_connection_id IS NOT NULL AND c.id IS NULL;

SELECT 'integration_webhook_events.provider_app_id' AS table_name, COUNT(*) AS orphan_count
FROM integration_webhook_events i
LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL
WHERE p.id IS NULL;
