DELETE FROM sys_app_quota
WHERE manifest_hash = 'LEGACY_BASELINE'
  AND protection_source IS NULL;

DELETE FROM sys_app_package_feature
WHERE manifest_hash = 'LEGACY_BASELINE'
  AND protection_source IS NULL;

DELETE FROM sys_app_permission
WHERE manifest_hash = 'LEGACY_BASELINE'
  AND protection_source IS NULL;

DELETE FROM sys_app_entry
WHERE manifest_hash = 'LEGACY_BASELINE'
  AND protection_source IS NULL;
