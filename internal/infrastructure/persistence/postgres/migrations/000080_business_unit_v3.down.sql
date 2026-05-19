DROP INDEX IF EXISTS idx_business_unit_attr_template_field_template;
DROP INDEX IF EXISTS idx_business_unit_attr_template_scope;
DROP INDEX IF EXISTS idx_business_unit_actor_unique;
DROP INDEX IF EXISTS idx_business_unit_relation_unique;
DROP INDEX IF EXISTS idx_business_unit_relation_target;
DROP INDEX IF EXISTS idx_business_unit_relation_source;
DROP INDEX IF EXISTS idx_business_unit_group;

DROP TABLE IF EXISTS business_unit_attr_template_field;
DROP TABLE IF EXISTS business_unit_attr_template;
DROP TABLE IF EXISTS business_unit_actor;
DROP TABLE IF EXISTS business_unit_relation;

ALTER TABLE business_unit
  DROP COLUMN IF EXISTS attrs,
  DROP COLUMN IF EXISTS attr_template_id,
  DROP COLUMN IF EXISTS unit_group_name,
  DROP COLUMN IF EXISTS unit_group_code,
  DROP COLUMN IF EXISTS unit_type_name,
  DROP COLUMN IF EXISTS unit_type_code;
