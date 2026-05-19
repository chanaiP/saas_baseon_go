DROP TABLE IF EXISTS business_resource_field_config;
DROP TABLE IF EXISTS business_resource_actor;

DROP INDEX IF EXISTS idx_business_resource_unit_tree;

ALTER TABLE business_resource
  DROP COLUMN IF EXISTS unit_type_code,
  DROP COLUMN IF EXISTS unit_type_name,
  DROP COLUMN IF EXISTS business_unit_code,
  DROP COLUMN IF EXISTS business_unit_name;
