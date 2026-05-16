DELETE FROM role_permission
WHERE permission_id IN (
  SELECT id
  FROM permission
  WHERE path = '/integration-center/my-connections'
    AND app_code = 'integration-center'
);

DELETE FROM permission
WHERE path = '/integration-center/my-connections'
  AND app_code = 'integration-center';

DELETE FROM sys_app_api
WHERE app_code = 'integration-center'
  AND path IN (
    '/api/integration-center/my-connections',
    '/api/integration-center/my-connections/{id}',
    '/api/integration-center/my-oauth/start',
    '/api/integration-center/my-sync-jobs',
    '/api/integration-center/my-sync-jobs/{id}'
  );

DELETE FROM sys_app_permission
WHERE app_code = 'integration-center'
  AND permission_code = '/integration-center/my-connections';

DELETE FROM sys_app_entry
WHERE app_code = 'integration-center'
  AND resource_code = 'integration_my_connections';
