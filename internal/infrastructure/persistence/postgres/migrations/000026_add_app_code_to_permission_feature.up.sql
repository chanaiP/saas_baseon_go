ALTER TABLE permission
  ADD COLUMN IF NOT EXISTS app_code VARCHAR(100) NOT NULL DEFAULT 'system-management';

ALTER TABLE saas_feature
  ADD COLUMN IF NOT EXISTS app_code VARCHAR(100) NOT NULL DEFAULT 'system-management';

UPDATE permission
SET app_code = CASE
  WHEN path = '/apps' OR path LIKE 'app:%' OR path LIKE 'data:app%' THEN 'app-center'
  WHEN path LIKE '/monitor/%' OR path LIKE 'mon%' OR path LIKE 'data:mon%' THEN 'system-monitor'
  ELSE 'system-management'
END
WHERE app_code IS NULL OR app_code = '' OR app_code = 'system-management';

UPDATE saas_feature
SET app_code = CASE
  WHEN feature_code = 'system_monitor' OR feature_code LIKE 'button_mon%' OR feature_name LIKE 'mon%' THEN 'system-monitor'
  ELSE 'system-management'
END
WHERE app_code IS NULL OR app_code = '' OR app_code = 'system-management';

CREATE INDEX IF NOT EXISTS idx_permission_app_code ON permission(app_code);
CREATE INDEX IF NOT EXISTS idx_saas_feature_app_code ON saas_feature(app_code);
