-- Ensure tenant branding is a tenant-side package capability, not a platform-only operation.
UPDATE permission
SET
  is_platform_only = FALSE,
  is_package_feature = TRUE,
  feature_code = 'brand_config',
  feature_type = 'OPERATION',
  tenant_editable = TRUE,
  tenant_edit_scope = 'NAME_ICON',
  updated_at = NOW()
WHERE path = 'brand:edit'
  AND deleted_at IS NULL;

UPDATE saas_feature
SET
  feature_name = '品牌配置',
  feature_type = 'CONFIG',
  status = 1,
  description = COALESCE(description, 'Logo、名称和版权配置'),
  updated_at = NOW()
WHERE feature_code = 'brand_config';
