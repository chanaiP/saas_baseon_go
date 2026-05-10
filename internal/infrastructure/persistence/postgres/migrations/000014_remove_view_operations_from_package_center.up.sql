UPDATE permission
SET
  is_package_feature = false,
  visible = false,
  feature_code = NULL,
  feature_type = 'VIEW',
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND path LIKE '%:view';

UPDATE permission
SET
  is_package_feature = false,
  feature_code = 'home',
  feature_type = 'MENU',
  data_perm_mode = 'NONE',
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND path = '/home';

UPDATE saas_feature
SET status = 0, updated_at = NOW()
WHERE feature_code LIKE 'button_%'
  AND feature_type IN ('BUTTON', 'OPERATION')
  AND feature_name LIKE '%:view';

UPDATE saas_feature sf
SET
  feature_name = p.name,
  feature_type = 'MENU',
  menu_id = p.id,
  parent_id = 0,
  status = 1,
  updated_at = NOW()
FROM permission p
WHERE p.deleted_at IS NULL
  AND p.perm_type = 3
  AND p.is_package_feature = true
  AND p.feature_code IS NOT NULL
  AND p.feature_code <> ''
  AND sf.feature_code = p.feature_code;

UPDATE saas_feature
SET
  feature_name = '系统监控',
  feature_type = 'MENU',
  parent_id = 0,
  status = 1,
  updated_at = NOW()
WHERE feature_code = 'system_monitor';
