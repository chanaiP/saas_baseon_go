UPDATE permission
SET feature_code = NULL,
    feature_type = 'OPERATION',
    tenant_edit_scope = NULL,
    updated_at = NOW()
WHERE path = 'brand:edit'
  AND deleted_at IS NULL;

UPDATE saas_feature
SET feature_name = '品牌-维护',
    feature_type = 'OPERATION',
    updated_at = NOW()
WHERE feature_code = 'brand_config';
