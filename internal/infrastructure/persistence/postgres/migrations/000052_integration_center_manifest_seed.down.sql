DELETE FROM role_permission
WHERE permission_id IN (
  SELECT id FROM permission WHERE app_code = 'integration-center'
);

DELETE FROM saas_quota
WHERE quota_code IN (
  'integration_connection_count',
  'integration_api_calls_daily',
  'integration_sync_records_daily'
);

DELETE FROM saas_feature
WHERE feature_code IN (
  'integration_data_sync',
  'integration_tenant_authorization'
);

DELETE FROM sys_app_quota
WHERE app_code = 'integration-center'
  AND manifest_hash = 'INTEGRATION_CENTER_BASELINE';

DELETE FROM sys_app_package_feature
WHERE app_code = 'integration-center'
  AND manifest_hash = 'INTEGRATION_CENTER_BASELINE';

DELETE FROM sys_app_permission
WHERE app_code = 'integration-center'
  AND manifest_hash = 'INTEGRATION_CENTER_BASELINE';

DELETE FROM sys_app_api
WHERE app_code = 'integration-center'
  AND manifest_hash = 'INTEGRATION_CENTER_BASELINE';

DELETE FROM sys_app_entry
WHERE app_code = 'integration-center'
  AND manifest_hash = 'INTEGRATION_CENTER_BASELINE';

DELETE FROM permission
WHERE app_code = 'integration-center';

UPDATE sys_app
SET status = 'PLANNED',
    charge_mode = 'SUBSCRIPTION',
    updated_at = now()
WHERE app_code = 'integration-center';
