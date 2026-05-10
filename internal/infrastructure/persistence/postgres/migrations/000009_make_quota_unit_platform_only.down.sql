UPDATE dict_type
SET
  scope = 'HYBRID',
  tenant_editable = true,
  is_platform_only = false,
  updated_at = now()
WHERE code = 'quota_unit'
  AND deleted_at IS NULL;
