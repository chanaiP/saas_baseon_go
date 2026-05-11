DROP INDEX IF EXISTS idx_saas_feature_app_code;
DROP INDEX IF EXISTS idx_permission_app_code;

ALTER TABLE saas_feature DROP COLUMN IF EXISTS app_code;
ALTER TABLE permission DROP COLUMN IF EXISTS app_code;
