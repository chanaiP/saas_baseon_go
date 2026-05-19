WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
),
target_type AS (
  SELECT dt.id
  FROM dict_type dt
  JOIN platform_tenant pt ON pt.id = dt.tenant_id
  WHERE dt.code IN (
    'base_business.unit_scenario',
    'base_business.unit_form',
    'business_resource.resource_category',
    'business_resource.resource_type',
    'business_resource.source_mode',
    'business_resource.unit_resource_relation_type',
    'business_resource.resource_relation_type',
    'business_resource.resource_status',
    'base.status'
  )
)
DELETE FROM dict_item di
WHERE di.dict_type_id IN (SELECT id FROM target_type);

WITH platform_tenant AS (
  SELECT id
  FROM tenant
  WHERE is_platform_tenant = true
    AND deleted_at IS NULL
)
DELETE FROM dict_type dt
USING platform_tenant pt
WHERE dt.tenant_id = pt.id
  AND dt.code IN (
    'base_business.unit_scenario',
    'base_business.unit_form',
    'business_resource.resource_category',
    'business_resource.resource_type',
    'business_resource.source_mode',
    'business_resource.unit_resource_relation_type',
    'business_resource.resource_relation_type',
    'business_resource.resource_status',
    'base.status'
  );
