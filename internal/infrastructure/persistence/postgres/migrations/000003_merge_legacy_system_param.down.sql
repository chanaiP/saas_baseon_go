CREATE TABLE IF NOT EXISTS system_param (
  id BIGSERIAL PRIMARY KEY,
  param_key VARCHAR(128) NOT NULL UNIQUE,
  param_value TEXT NOT NULL,
  remark TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO system_param (param_key, param_value, remark, created_at, updated_at)
SELECT param_key, param_value, remark, created_at, updated_at
FROM sys_param
WHERE tenant_id = 1
  AND deleted_at IS NULL
ON CONFLICT (param_key) DO NOTHING;
