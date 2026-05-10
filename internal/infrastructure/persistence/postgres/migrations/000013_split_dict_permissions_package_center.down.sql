UPDATE saas_feature
SET status = 1,
    updated_at = now()
WHERE feature_code IN ('button_dict_create', 'button_dict_edit', 'button_dict_delete');

DELETE FROM saas_plan_feature
WHERE feature_id IN (
  SELECT id FROM saas_feature
  WHERE feature_code IN ('button_dict_item_create', 'button_dict_item_edit', 'button_dict_item_delete')
);

DELETE FROM saas_feature
WHERE feature_code IN ('button_dict_item_create', 'button_dict_item_edit', 'button_dict_item_delete');

DELETE FROM role_permission
WHERE permission_id IN (
  SELECT id FROM permission
  WHERE tenant_id = 1
    AND path IN (
      'dict_type:create', 'dict_type:edit', 'dict_type:delete',
      'dict_item:create', 'dict_item:edit', 'dict_item:delete'
    )
);

DELETE FROM permission
WHERE tenant_id = 1
  AND path IN (
    'dict_type:create', 'dict_type:edit', 'dict_type:delete',
    'dict_item:create', 'dict_item:edit', 'dict_item:delete'
  );

UPDATE permission
SET is_package_feature = true,
    visible = true,
    updated_at = now()
WHERE tenant_id = 1
  AND path IN ('dict:create', 'dict:edit', 'dict:delete')
  AND deleted_at IS NULL;
