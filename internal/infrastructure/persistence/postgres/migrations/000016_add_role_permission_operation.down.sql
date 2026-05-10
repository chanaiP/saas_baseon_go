DELETE FROM role_permission
WHERE permission_id IN (
  SELECT id FROM permission
  WHERE tenant_id = 1 AND path = 'role:permission'
);

UPDATE permission
SET enabled = false,
    visible = false,
    is_package_feature = false,
    updated_at = now()
WHERE tenant_id = 1
  AND path = 'role:permission'
  AND deleted_at IS NULL;

UPDATE saas_plan_feature
SET enabled = false,
    updated_at = now()
WHERE feature_id IN (
  SELECT id FROM saas_feature WHERE feature_code = 'button_role_permission'
);

UPDATE saas_feature
SET status = 0,
    updated_at = now()
WHERE feature_code = 'button_role_permission';
