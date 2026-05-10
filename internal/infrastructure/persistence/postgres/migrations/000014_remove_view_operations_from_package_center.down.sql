UPDATE permission
SET
  visible = true,
  feature_type = 'OPERATION',
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND path LIKE '%:view';

UPDATE permission
SET
  is_package_feature = true,
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND path IN (
    'audit:view',
    'login:view',
    'monhealth:view',
    'monserver:view',
    'monjobs:view',
    'monservices:view',
    'moncache:view',
    'moncachekeys:view'
  );

UPDATE saas_feature
SET status = 1, updated_at = NOW()
WHERE feature_code LIKE 'button_%'
  AND feature_type IN ('BUTTON', 'OPERATION')
  AND feature_name LIKE '%:view';
