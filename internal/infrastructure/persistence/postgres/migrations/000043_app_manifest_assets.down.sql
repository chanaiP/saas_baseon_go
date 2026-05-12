DROP TABLE IF EXISTS sys_app_quota;
DROP TABLE IF EXISTS sys_app_package_feature;
DROP TABLE IF EXISTS sys_app_permission;
DROP TABLE IF EXISTS sys_app_api;
DROP TABLE IF EXISTS sys_app_entry;
DROP TABLE IF EXISTS sys_app_manifest_file;
DROP TABLE IF EXISTS sys_app_manifest_load;

ALTER TABLE sys_app
    DROP COLUMN IF EXISTS last_manifest_synced_at,
    DROP COLUMN IF EXISTS manifest_version,
    DROP COLUMN IF EXISTS manifest_hash,
    DROP COLUMN IF EXISTS webhook_url,
    DROP COLUMN IF EXISTS api_base_url,
    DROP COLUMN IF EXISTS health_check_url;
