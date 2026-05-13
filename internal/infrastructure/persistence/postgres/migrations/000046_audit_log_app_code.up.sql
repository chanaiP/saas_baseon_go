ALTER TABLE audit_log
  ADD COLUMN IF NOT EXISTS app_code VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_audit_log_app_code_created
  ON audit_log (app_code, created_at DESC);

UPDATE audit_log
SET app_code = COALESCE(
  NULLIF(detail::jsonb #>> '{parse,app_code}', ''),
  CASE
    WHEN module IN ('app', 'app_center', 'application') THEN 'app-center'
    WHEN module IN ('monitor', 'system_monitor') THEN 'system-monitor'
    WHEN module IN ('model', 'model_manager', 'model-manager') THEN 'model-manager'
    WHEN module IN (
      'business_unit', 'business_unit_org_map', 'dict_item', 'dict_type', 'file',
      'menu', 'organization', 'permission', 'plan', 'position', 'position_type',
      'profile', 'quota', 'role', 'sys_param', 'tenant', 'tenant_branding',
      'tenant_feature_override', 'tenant_quota_override', 'tenant_subscription', 'user'
    ) THEN 'system-management'
    ELSE replace(module, '_', '-')
  END
)
WHERE app_code IS NULL
  AND detail IS NOT NULL
  AND detail ~ '^[[:space:]]*[{[]';

UPDATE audit_log
SET app_code = CASE
  WHEN module IN ('app', 'app_center', 'application') THEN 'app-center'
  WHEN module IN ('monitor', 'system_monitor') THEN 'system-monitor'
  WHEN module IN ('model', 'model_manager', 'model-manager') THEN 'model-manager'
  WHEN module IN (
    'business_unit', 'business_unit_org_map', 'dict_item', 'dict_type', 'file',
    'menu', 'organization', 'permission', 'plan', 'position', 'position_type',
    'profile', 'quota', 'role', 'sys_param', 'tenant', 'tenant_branding',
    'tenant_feature_override', 'tenant_quota_override', 'tenant_subscription', 'user'
  ) THEN 'system-management'
  ELSE replace(module, '_', '-')
END
WHERE app_code IS NULL;
