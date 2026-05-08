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

DO $$
BEGIN
  IF to_regclass('public.system_param') IS NOT NULL THEN
    INSERT INTO sys_param (
      tenant_id,
      param_key,
      param_value,
      remark,
      value_type,
      tenant_editable,
      is_platform_only,
      created_at,
      updated_at
    )
    SELECT
      1,
      legacy.param_key,
      legacy.param_value,
      LEFT(COALESCE(legacy.remark, ''), 500),
      'string',
      TRUE,
      FALSE,
      COALESCE(legacy.created_at, now()),
      COALESCE(legacy.updated_at, now())
    FROM system_param AS legacy
    ON CONFLICT (tenant_id, param_key) DO NOTHING;
  END IF;
END $$;

DROP TABLE IF EXISTS system_param;
