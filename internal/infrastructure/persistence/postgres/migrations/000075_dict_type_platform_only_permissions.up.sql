UPDATE permission
SET is_platform_only = TRUE,
    is_package_feature = FALSE,
    feature_type = 'OPERATION',
    data_perm_mode = 'NONE',
    updated_at = now()
WHERE deleted_at IS NULL
  AND path IN ('dict_type:create', 'dict_type:edit', 'dict_type:delete');

UPDATE sys_app_permission
SET platform_only = TRUE,
    include_in_package = FALSE,
    data_perm_mode = 'NONE',
    updated_at = now(),
    last_synced_at = now()
WHERE deleted_at IS NULL
  AND app_code = 'system-management'
  AND permission_code IN ('dict_type:create', 'dict_type:edit', 'dict_type:delete');
