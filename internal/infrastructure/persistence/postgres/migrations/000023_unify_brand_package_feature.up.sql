UPDATE permission
SET is_platform_only = FALSE,
    is_package_feature = TRUE,
    feature_code = 'brand_config',
    feature_type = 'CONFIG',
    tenant_editable = TRUE,
    tenant_edit_scope = 'NAME_ICON',
    data_perm_mode = 'ORG',
    updated_at = NOW()
WHERE path = 'brand:edit'
  AND deleted_at IS NULL;

UPDATE saas_feature
SET feature_name = '品牌配置',
    feature_type = 'CONFIG',
    parent_id = 0,
    menu_id = NULL,
    status = 1,
    description = COALESCE(description, 'Logo、名称和版权配置'),
    updated_at = NOW()
WHERE feature_code = 'brand_config';

INSERT INTO saas_feature (feature_code, feature_name, feature_type, parent_id, status, description, created_at, updated_at)
SELECT 'brand_config', '品牌配置', 'CONFIG', 0, 1, 'Logo、名称和版权配置', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM saas_feature WHERE feature_code = 'brand_config'
);

DELETE FROM saas_plan_feature
WHERE feature_id IN (
  SELECT id FROM saas_feature WHERE feature_code = 'button_brand_edit'
);

UPDATE saas_feature
SET status = 0,
    updated_at = NOW()
WHERE feature_code = 'button_brand_edit';

INSERT INTO saas_plan_feature (plan_id, feature_id, enabled, created_at, updated_at)
SELECT p.id, f.id, TRUE, NOW(), NOW()
FROM saas_plan p
CROSS JOIN saas_feature f
WHERE p.plan_code IN ('TRIAL', 'BASIC', 'PRO', 'ENTERPRISE')
  AND f.feature_code = 'brand_config'
ON CONFLICT (plan_id, feature_id)
DO UPDATE SET enabled = TRUE, updated_at = NOW();
