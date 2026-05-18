UPDATE permission
SET is_platform_only = FALSE,
    updated_at = now()
WHERE deleted_at IS NULL
  AND path IN ('dict_type:create', 'dict_type:edit', 'dict_type:delete');

UPDATE sys_app_permission
SET platform_only = FALSE,
    updated_at = now(),
    last_synced_at = now()
WHERE deleted_at IS NULL
  AND app_code = 'system-management'
  AND permission_code IN ('dict_type:create', 'dict_type:edit', 'dict_type:delete');
