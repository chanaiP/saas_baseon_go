DELETE FROM role_permission rp
USING permission p
WHERE rp.permission_id = p.id
  AND p.path IN (
    'business_unit:resource_bind',
    '/business-resources',
    'business_resource:create',
    'business_resource:edit',
    'business_resource:delete',
    'business_resource:relation_manage'
  );

DELETE FROM saas_plan_feature spf
USING saas_feature sf
WHERE spf.feature_id = sf.id
  AND sf.feature_code IN (
    'business_resource_manage',
    'button_business_unit_resource_bind',
    'button_business_resource_create',
    'button_business_resource_edit',
    'button_business_resource_delete',
    'button_business_resource_relation_manage'
  );

UPDATE saas_feature
SET status = 0,
    updated_at = now()
WHERE feature_code IN (
  'business_resource_manage',
  'button_business_unit_resource_bind',
  'button_business_resource_create',
  'button_business_resource_edit',
  'button_business_resource_delete',
  'button_business_resource_relation_manage'
)
  AND status <> 0;

UPDATE permission
SET deleted_at = now(),
    enabled = false,
    updated_at = now()
WHERE path IN (
  'business_unit:resource_bind',
  '/business-resources',
  'business_resource:create',
  'business_resource:edit',
  'business_resource:delete',
  'business_resource:relation_manage'
)
  AND deleted_at IS NULL;
