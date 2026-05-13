ALTER TABLE role_permission
    DROP COLUMN IF EXISTS source_ref,
    DROP COLUMN IF EXISTS source;

ALTER TABLE tenant_menu_override
    DROP COLUMN IF EXISTS source_ref,
    DROP COLUMN IF EXISTS source;

ALTER TABLE tenant_quota_override
    DROP COLUMN IF EXISTS source_ref,
    DROP COLUMN IF EXISTS source;

ALTER TABLE tenant_feature_override
    DROP COLUMN IF EXISTS source_ref,
    DROP COLUMN IF EXISTS source;

ALTER TABLE tenant_subscription
    DROP COLUMN IF EXISTS source_ref,
    DROP COLUMN IF EXISTS source;

ALTER TABLE sys_app_quota
    DROP COLUMN IF EXISTS protected_at,
    DROP COLUMN IF EXISTS protection_reason,
    DROP COLUMN IF EXISTS protection_source;

ALTER TABLE sys_app_package_feature
    DROP COLUMN IF EXISTS protected_at,
    DROP COLUMN IF EXISTS protection_reason,
    DROP COLUMN IF EXISTS protection_source;

ALTER TABLE sys_app_permission
    DROP COLUMN IF EXISTS protected_at,
    DROP COLUMN IF EXISTS protection_reason,
    DROP COLUMN IF EXISTS protection_source;

ALTER TABLE sys_app_api
    DROP COLUMN IF EXISTS protected_at,
    DROP COLUMN IF EXISTS protection_reason,
    DROP COLUMN IF EXISTS protection_source;

ALTER TABLE sys_app_entry
    DROP COLUMN IF EXISTS protected_at,
    DROP COLUMN IF EXISTS protection_reason,
    DROP COLUMN IF EXISTS protection_source;
