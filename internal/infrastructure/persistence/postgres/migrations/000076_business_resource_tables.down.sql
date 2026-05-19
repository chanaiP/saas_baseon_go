DROP INDEX IF EXISTS idx_business_resource_relation_unique;
DROP INDEX IF EXISTS idx_business_resource_relation_child;
DROP INDEX IF EXISTS idx_business_resource_relation_parent;
DROP TABLE IF EXISTS business_resource_relation;

DROP INDEX IF EXISTS idx_business_unit_resource_primary;
DROP INDEX IF EXISTS idx_business_unit_resource_resource;
DROP INDEX IF EXISTS idx_business_unit_resource_unit;
DROP TABLE IF EXISTS business_unit_resource;

DROP INDEX IF EXISTS idx_business_resource_parent_id;
DROP INDEX IF EXISTS idx_business_resource_source;
DROP INDEX IF EXISTS idx_business_resource_tenant_type;
DROP INDEX IF EXISTS idx_business_resource_tenant_code;
DROP TABLE IF EXISTS business_resource;
