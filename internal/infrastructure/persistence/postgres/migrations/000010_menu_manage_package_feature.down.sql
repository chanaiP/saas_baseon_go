UPDATE permission
SET
  feature_code = 'role_manage',
  updated_at = now()
WHERE path = '/menus'
  AND perm_type = 3
  AND deleted_at IS NULL;

UPDATE permission
SET
  tenant_editable = false,
  updated_at = now()
WHERE perm_type = 3
  AND is_platform_only = false
  AND deleted_at IS NULL;

UPDATE saas_feature
SET
  parent_id = COALESCE((SELECT id FROM saas_feature WHERE feature_code = 'role_manage' LIMIT 1), 0),
  updated_at = now()
WHERE feature_code IN ('button_menu_create', 'button_menu_edit', 'button_menu_delete');
