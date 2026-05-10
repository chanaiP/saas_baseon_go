CREATE TABLE IF NOT EXISTS sys_param (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL DEFAULT 1,
  param_key VARCHAR(128) NOT NULL,
  param_value TEXT NOT NULL,
  remark VARCHAR(500) NOT NULL DEFAULT '',
  value_type VARCHAR(32) NOT NULL DEFAULT 'string',
  tenant_editable BOOLEAN NOT NULL DEFAULT TRUE,
  is_platform_only BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT idx_sys_param_tenant_key UNIQUE (tenant_id, param_key)
);
