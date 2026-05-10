UPDATE permission
SET is_platform_only = false,
    updated_at = now()
WHERE path = 'menu:delete'
  AND deleted_at IS NULL;

UPDATE saas_feature
SET status = 1,
    updated_at = now()
WHERE feature_code = 'button_menu_delete';
