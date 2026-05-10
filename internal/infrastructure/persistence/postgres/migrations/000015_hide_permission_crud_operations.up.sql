UPDATE permission
SET
  visible = false,
  is_package_feature = false,
  is_platform_only = true,
  feature_code = NULL,
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND path IN ('perm:create', 'perm:edit', 'perm:delete');

UPDATE saas_feature
SET status = 0, updated_at = NOW()
WHERE feature_code IN ('button_perm_create', 'button_perm_edit', 'button_perm_delete');
