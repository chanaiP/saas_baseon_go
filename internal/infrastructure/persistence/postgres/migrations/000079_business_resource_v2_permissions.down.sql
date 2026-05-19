DELETE FROM role_permission
WHERE permission_id IN (
  SELECT id
  FROM permission
  WHERE path IN (
    'business_resource:actor_manage',
    'business_resource:field_config_manage',
    'business_resource:import'
  )
);

UPDATE permission
SET deleted_at = now(), enabled = false, updated_at = now()
WHERE path IN (
  'business_resource:actor_manage',
  'business_resource:field_config_manage',
  'business_resource:import'
)
  AND deleted_at IS NULL;
