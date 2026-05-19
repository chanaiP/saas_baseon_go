CREATE TABLE IF NOT EXISTS business_resource (
  id bigserial PRIMARY KEY,
  tenant_id bigint NOT NULL,
  resource_name varchar(128) NOT NULL,
  resource_code varchar(64) NOT NULL,
  resource_category varchar(64) NOT NULL,
  resource_type varchar(64) NOT NULL,
  source_mode varchar(32) NOT NULL DEFAULT 'native',
  source_app_code varchar(64),
  source_table varchar(128),
  source_id bigint,
  platform_code varchar(64),
  external_id varchar(128),
  connection_instance_id bigint,
  parent_resource_id bigint,
  resource_attrs jsonb,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_resource_tenant_code
  ON business_resource (tenant_id, resource_code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_resource_tenant_type
  ON business_resource (tenant_id, resource_category, resource_type)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_resource_source
  ON business_resource (source_app_code, source_table, source_id);

CREATE INDEX IF NOT EXISTS idx_business_resource_parent_id
  ON business_resource (parent_resource_id);

CREATE TABLE IF NOT EXISTS business_unit_resource (
  id bigserial PRIMARY KEY,
  tenant_id bigint NOT NULL,
  business_unit_id bigint NOT NULL,
  resource_id bigint NOT NULL,
  resource_category varchar(64) NOT NULL,
  resource_type varchar(64) NOT NULL,
  relation_type varchar(64) NOT NULL,
  is_primary boolean NOT NULL DEFAULT false,
  use_for_permission boolean NOT NULL DEFAULT false,
  use_for_operation boolean NOT NULL DEFAULT false,
  use_for_settlement boolean NOT NULL DEFAULT false,
  start_date date,
  end_date date,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_business_unit_resource_unit
  ON business_unit_resource (tenant_id, business_unit_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_unit_resource_resource
  ON business_unit_resource (tenant_id, resource_id)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_unit_resource_primary
  ON business_unit_resource (tenant_id, resource_id, relation_type)
  WHERE deleted_at IS NULL AND is_primary = true;

CREATE TABLE IF NOT EXISTS business_resource_relation (
  id bigserial PRIMARY KEY,
  tenant_id bigint NOT NULL,
  parent_resource_id bigint NOT NULL,
  child_resource_id bigint NOT NULL,
  parent_resource_category varchar(64) NOT NULL,
  parent_resource_type varchar(64) NOT NULL,
  child_resource_category varchar(64) NOT NULL,
  child_resource_type varchar(64) NOT NULL,
  relation_type varchar(64) NOT NULL,
  start_date date,
  end_date date,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_business_resource_relation_parent
  ON business_resource_relation (tenant_id, parent_resource_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_business_resource_relation_child
  ON business_resource_relation (tenant_id, child_resource_id)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_business_resource_relation_unique
  ON business_resource_relation (tenant_id, parent_resource_id, child_resource_id, relation_type)
  WHERE deleted_at IS NULL;
