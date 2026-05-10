UPDATE permission
SET is_platform_only = true,
    is_package_feature = false,
    tenant_editable = false,
    feature_code = NULL,
    feature_type = 'OPERATION',
    data_perm_mode = 'NONE',
    updated_at = now()
WHERE path = 'menu:delete'
  AND deleted_at IS NULL;

UPDATE saas_feature
SET status = 0,
    updated_at = now()
WHERE feature_code = 'button_menu_delete';

DELETE FROM saas_plan_feature
WHERE feature_id IN (
  SELECT id
  FROM saas_feature
  WHERE feature_code = 'button_menu_delete'
);
