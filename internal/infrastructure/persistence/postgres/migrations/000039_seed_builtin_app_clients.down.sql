DELETE FROM sys_app_client
WHERE config_note = 'bootstrap:built-in-client'
  AND client_code IN ('PC_WEB', 'API_ONLY')
  AND app_id IN (
    SELECT id
    FROM sys_app
    WHERE app_code IN (
      'app-center',
      'system-management',
      'system-monitor',
      'workbench',
      'integration-center',
      'data-center'
    )
  );
