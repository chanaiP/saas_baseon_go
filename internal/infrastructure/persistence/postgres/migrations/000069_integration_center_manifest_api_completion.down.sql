DELETE FROM sys_app_api
WHERE app_code = 'integration-center'
  AND manifest_hash = 'INTEGRATION_CENTER_API_COMPLETION'
  AND deleted_at IS NULL;
