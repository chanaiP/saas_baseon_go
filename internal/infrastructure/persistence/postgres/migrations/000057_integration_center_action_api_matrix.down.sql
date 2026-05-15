DELETE FROM sys_app_api
WHERE app_code = 'integration-center'
  AND (
    path IN (
      '/api/integration-center/provider-apps',
      '/api/integration-center/connectivity-check',
      '/api/integration-center/provider-apps/{code}',
      '/api/integration-center/app-capabilities/{id}',
      '/api/integration-center/tenant-connections/{id}/refresh',
      '/api/integration-center/tenant-connections/{id}/pause',
      '/api/integration-center/tenant-connections/{id}/resume',
      '/api/integration-center/tenant-connections/{id}/retry',
      '/api/integration-center/sync-jobs/{id}/retry',
      '/api/integration-center/sync-jobs/{id}/pause',
      '/api/integration-center/sync-jobs/{id}/resume',
      '/api/integration-center/quota-policies',
      '/api/integration-center/quota-policies/{code}',
      '/api/integration-center/quota-policies/{code}/status',
      '/api/integration-center/alerts/{id}/process',
      '/api/integration-center/alerts/{id}/resolve',
      '/api/integration-center/alerts/{id}/ignore',
      '/api/integration-center/logs/export'
    )
  );
