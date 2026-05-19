ALTER TABLE business_unit
  ADD COLUMN IF NOT EXISTS unit_type_code varchar(64),
  ADD COLUMN IF NOT EXISTS unit_type_name varchar(128),
  ADD COLUMN IF NOT EXISTS unit_group_code varchar(64),
  ADD COLUMN IF NOT EXISTS unit_group_name varchar(128),
  ADD COLUMN IF NOT EXISTS attr_template_id bigint,
  ADD COLUMN IF NOT EXISTS attrs jsonb;

UPDATE business_unit
SET
  unit_type_code = COALESCE(NULLIF(unit_type_code, ''), NULLIF(unit_form, ''), NULLIF(bu_type, '')),
  unit_type_name = COALESCE(NULLIF(unit_type_name, ''), NULLIF(unit_form, ''), NULLIF(bu_type, '')),
  unit_group_code = COALESCE(NULLIF(unit_group_code, ''), NULLIF(unit_scenario, '')),
  unit_group_name = COALESCE(NULLIF(unit_group_name, ''), NULLIF(unit_scenario, '')),
  attrs = COALESCE(attrs, '{}'::jsonb)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_unit_group
  ON business_unit (tenant_id, unit_type_code, unit_group_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS business_unit_relation (
  id bigserial PRIMARY KEY,
  tenant_id bigint NOT NULL,
  source_unit_id bigint NOT NULL,
  target_unit_id bigint NOT NULL,
  relation_type_code varchar(64) NOT NULL,
  relation_type_name varchar(128) NOT NULL,
  status varchar(32) NOT NULL DEFAULT 'active',
  remark text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_business_unit_relation_source
  ON business_unit_relation (tenant_id, source_unit_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_unit_relation_target
  ON business_unit_relation (tenant_id, target_unit_id)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_unit_relation_unique
  ON business_unit_relation (tenant_id, source_unit_id, target_unit_id, relation_type_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS business_unit_actor (
  id bigserial PRIMARY KEY,
  tenant_id bigint NOT NULL,
  business_unit_id bigint NOT NULL,
  actor_type varchar(32) NOT NULL,
  actor_id bigint NOT NULL,
  role_type varchar(32) NOT NULL,
  include_children boolean NOT NULL DEFAULT false,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_unit_actor_unique
  ON business_unit_actor (tenant_id, business_unit_id, actor_type, actor_id, role_type)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS business_unit_attr_template (
  id bigserial PRIMARY KEY,
  tenant_id bigint,
  template_name varchar(128) NOT NULL,
  unit_type_code varchar(64) NOT NULL,
  unit_type_name varchar(128) NOT NULL,
  unit_group_code varchar(64),
  unit_group_name varchar(128),
  sort_order integer NOT NULL DEFAULT 0,
  status varchar(32) NOT NULL DEFAULT 'active',
  remark text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_business_unit_attr_template_scope
  ON business_unit_attr_template (tenant_id, unit_type_code, unit_group_code)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS business_unit_attr_template_field (
  id bigserial PRIMARY KEY,
  tenant_id bigint,
  template_id bigint NOT NULL,
  field_key varchar(128) NOT NULL,
  field_label varchar(128) NOT NULL,
  field_type varchar(32) NOT NULL,
  required boolean NOT NULL DEFAULT false,
  default_value text,
  placeholder varchar(255),
  options_json jsonb,
  sort_order integer NOT NULL DEFAULT 0,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_business_unit_attr_template_field_template
  ON business_unit_attr_template_field (template_id, sort_order, id)
  WHERE deleted_at IS NULL;
