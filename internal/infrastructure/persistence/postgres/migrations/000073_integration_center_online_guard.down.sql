UPDATE sys_app
SET status = 'PLANNED', updated_at = now()
WHERE app_code = 'integration-center'
  AND deleted_at IS NULL
  AND status = 'ONLINE';
