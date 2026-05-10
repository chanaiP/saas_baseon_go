DELETE FROM role_permission
WHERE permission_id IN (
  SELECT id
  FROM permission
  WHERE path = 'menu:package_feature'
);

UPDATE permission
SET deleted_at = now(),
    enabled = false,
    is_package_feature = false,
    updated_at = now()
WHERE path = 'menu:package_feature'
  AND deleted_at IS NULL;
