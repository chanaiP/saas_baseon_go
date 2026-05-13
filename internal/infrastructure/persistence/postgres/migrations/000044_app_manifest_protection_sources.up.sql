ALTER TABLE sys_app_entry
    ADD COLUMN IF NOT EXISTS protection_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS protection_reason VARCHAR(500),
    ADD COLUMN IF NOT EXISTS protected_at TIMESTAMPTZ;

ALTER TABLE sys_app_api
    ADD COLUMN IF NOT EXISTS protection_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS protection_reason VARCHAR(500),
    ADD COLUMN IF NOT EXISTS protected_at TIMESTAMPTZ;

ALTER TABLE sys_app_permission
    ADD COLUMN IF NOT EXISTS protection_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS protection_reason VARCHAR(500),
    ADD COLUMN IF NOT EXISTS protected_at TIMESTAMPTZ;

ALTER TABLE sys_app_package_feature
    ADD COLUMN IF NOT EXISTS protection_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS protection_reason VARCHAR(500),
    ADD COLUMN IF NOT EXISTS protected_at TIMESTAMPTZ;

ALTER TABLE sys_app_quota
    ADD COLUMN IF NOT EXISTS protection_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS protection_reason VARCHAR(500),
    ADD COLUMN IF NOT EXISTS protected_at TIMESTAMPTZ;

ALTER TABLE tenant_subscription
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(128);

ALTER TABLE tenant_feature_override
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(128);

ALTER TABLE tenant_quota_override
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(128);

ALTER TABLE tenant_menu_override
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(128);

ALTER TABLE role_permission
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(128);

UPDATE sys_app_entry
SET protection_source = 'MANUAL',
    protection_reason = COALESCE(protection_reason, '非 Manifest 托管资产，保留人工保护标记'),
    protected_at = COALESCE(protected_at, updated_at, now())
WHERE managed_by_manifest = false
  AND protection_source IS NULL;

UPDATE sys_app_api
SET protection_source = 'MANUAL',
    protection_reason = COALESCE(protection_reason, '非 Manifest 托管资产，保留人工保护标记'),
    protected_at = COALESCE(protected_at, updated_at, now())
WHERE managed_by_manifest = false
  AND protection_source IS NULL;

UPDATE sys_app_permission
SET protection_source = 'MANUAL',
    protection_reason = COALESCE(protection_reason, '非 Manifest 托管资产，保留人工保护标记'),
    protected_at = COALESCE(protected_at, updated_at, now())
WHERE managed_by_manifest = false
  AND protection_source IS NULL;

UPDATE sys_app_package_feature
SET protection_source = 'MANUAL',
    protection_reason = COALESCE(protection_reason, '非 Manifest 托管资产，保留人工保护标记'),
    protected_at = COALESCE(protected_at, updated_at, now())
WHERE managed_by_manifest = false
  AND protection_source IS NULL;

UPDATE sys_app_quota
SET protection_source = 'MANUAL',
    protection_reason = COALESCE(protection_reason, '非 Manifest 托管资产，保留人工保护标记'),
    protected_at = COALESCE(protected_at, updated_at, now())
WHERE managed_by_manifest = false
  AND protection_source IS NULL;
