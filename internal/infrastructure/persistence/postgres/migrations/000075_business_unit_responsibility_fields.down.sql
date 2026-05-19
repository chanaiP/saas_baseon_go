DROP INDEX IF EXISTS idx_business_unit_owner_org_id;
DROP INDEX IF EXISTS idx_business_unit_owner_user_id;
DROP INDEX IF EXISTS idx_business_unit_tenant_scenario_form;
DROP INDEX IF EXISTS idx_business_unit_parent_id;

ALTER TABLE business_unit
  DROP COLUMN IF EXISTS data_scope_enabled,
  DROP COLUMN IF EXISTS settlement_enabled,
  DROP COLUMN IF EXISTS operation_enabled,
  DROP COLUMN IF EXISTS owner_org_id,
  DROP COLUMN IF EXISTS owner_user_id,
  DROP COLUMN IF EXISTS parent_id,
  DROP COLUMN IF EXISTS unit_form,
  DROP COLUMN IF EXISTS unit_scenario;
