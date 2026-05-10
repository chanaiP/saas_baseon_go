UPDATE permission
SET is_package_feature = TRUE
WHERE deleted_at IS NULL
  AND feature_code IS NOT NULL
  AND feature_code <> '';

UPDATE saas_feature
SET status = 1
WHERE feature_code IN ('home', 'tenant_manage', 'plan_manage', 'system_monitor');
