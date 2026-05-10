UPDATE permission
SET
  visible = true,
  is_package_feature = true,
  is_platform_only = false,
  feature_type = 'OPERATION',
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND path IN ('perm:create', 'perm:edit', 'perm:delete');

UPDATE saas_feature
SET status = 1, updated_at = NOW()
WHERE feature_code IN ('button_perm_create', 'button_perm_edit', 'button_perm_delete');
