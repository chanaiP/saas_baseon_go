ALTER TABLE business_unit
  ADD COLUMN IF NOT EXISTS unit_scenario varchar(64),
  ADD COLUMN IF NOT EXISTS unit_form varchar(64),
  ADD COLUMN IF NOT EXISTS parent_id bigint,
  ADD COLUMN IF NOT EXISTS owner_user_id bigint,
  ADD COLUMN IF NOT EXISTS owner_org_id bigint,
  ADD COLUMN IF NOT EXISTS operation_enabled boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS settlement_enabled boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS data_scope_enabled boolean NOT NULL DEFAULT false;

UPDATE business_unit
SET
  unit_form = COALESCE(NULLIF(unit_form, ''), NULLIF(bu_type, '')),
  operation_enabled = COALESCE(statistic_enabled, true),
  settlement_enabled = COALESCE(billing_enabled, false)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_unit_parent_id
  ON business_unit (parent_id);

CREATE INDEX IF NOT EXISTS idx_business_unit_tenant_scenario_form
  ON business_unit (tenant_id, unit_scenario, unit_form)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_unit_owner_user_id
  ON business_unit (owner_user_id);

CREATE INDEX IF NOT EXISTS idx_business_unit_owner_org_id
  ON business_unit (owner_org_id);
