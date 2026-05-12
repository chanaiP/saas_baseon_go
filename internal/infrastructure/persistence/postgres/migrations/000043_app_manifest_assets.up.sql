ALTER TABLE sys_app
    ADD COLUMN IF NOT EXISTS health_check_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS api_base_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS webhook_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS manifest_hash VARCHAR(128),
    ADD COLUMN IF NOT EXISTS manifest_version VARCHAR(64),
    ADD COLUMN IF NOT EXISTS last_manifest_synced_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS sys_app_manifest_load (
    id BIGSERIAL PRIMARY KEY,
    app_code VARCHAR(100) NOT NULL,
    action VARCHAR(32) NOT NULL DEFAULT 'LOAD',
    source_type VARCHAR(32) NOT NULL DEFAULT 'UPLOAD',
    source_name VARCHAR(500),
    manifest_version VARCHAR(64) NOT NULL DEFAULT '1.0',
    manifest_hash VARCHAR(128) NOT NULL,
    fragment_role VARCHAR(64) NOT NULL DEFAULT 'main',
    status VARCHAR(32) NOT NULL,
    summary TEXT,
    error_summary TEXT,
    diff_summary TEXT,
    operator_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sys_app_manifest_load_app_code ON sys_app_manifest_load(app_code);
CREATE INDEX IF NOT EXISTS idx_sys_app_manifest_load_status ON sys_app_manifest_load(status);
CREATE INDEX IF NOT EXISTS idx_sys_app_manifest_load_operator ON sys_app_manifest_load(operator_user_id);

CREATE TABLE IF NOT EXISTS sys_app_manifest_file (
    id BIGSERIAL PRIMARY KEY,
    load_id BIGINT REFERENCES sys_app_manifest_load(id),
    app_code VARCHAR(100) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(1000),
    fragment_role VARCHAR(64) NOT NULL DEFAULT 'main',
    manifest_version VARCHAR(64) NOT NULL DEFAULT '1.0',
    manifest_hash VARCHAR(128) NOT NULL,
    content_summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sys_app_manifest_file_app_code ON sys_app_manifest_file(app_code);
CREATE INDEX IF NOT EXISTS idx_sys_app_manifest_file_hash ON sys_app_manifest_file(manifest_hash);

CREATE TABLE IF NOT EXISTS sys_app_entry (
    id BIGSERIAL PRIMARY KEY,
    app_code VARCHAR(100) NOT NULL,
    resource_code VARCHAR(120) NOT NULL,
    name VARCHAR(200) NOT NULL,
    path VARCHAR(500) NOT NULL,
    parent_code VARCHAR(120),
    sort_order INT NOT NULL DEFAULT 0,
    platform_only BOOLEAN NOT NULL DEFAULT FALSE,
    tenant_visible BOOLEAN NOT NULL DEFAULT TRUE,
    tenant_editable BOOLEAN NOT NULL DEFAULT FALSE,
    include_in_package BOOLEAN NOT NULL DEFAULT FALSE,
    feature_code VARCHAR(100),
    data_perm_mode VARCHAR(16) NOT NULL DEFAULT 'ORG',
    manifest_hash VARCHAR(128) NOT NULL,
    managed_by_manifest BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_app_entry_code ON sys_app_entry(app_code, resource_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sys_app_entry_path ON sys_app_entry(path);
CREATE INDEX IF NOT EXISTS idx_sys_app_entry_status ON sys_app_entry(status);

CREATE TABLE IF NOT EXISTS sys_app_api (
    id BIGSERIAL PRIMARY KEY,
    app_code VARCHAR(100) NOT NULL,
    method VARCHAR(20) NOT NULL,
    path VARCHAR(500) NOT NULL,
    permission_code VARCHAR(120),
    public BOOLEAN NOT NULL DEFAULT FALSE,
    audit BOOLEAN NOT NULL DEFAULT FALSE,
    manifest_hash VARCHAR(128) NOT NULL,
    managed_by_manifest BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_app_api_route ON sys_app_api(app_code, method, path) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sys_app_api_permission ON sys_app_api(permission_code);
CREATE INDEX IF NOT EXISTS idx_sys_app_api_status ON sys_app_api(status);

CREATE TABLE IF NOT EXISTS sys_app_permission (
    id BIGSERIAL PRIMARY KEY,
    app_code VARCHAR(100) NOT NULL,
    permission_code VARCHAR(120) NOT NULL,
    name VARCHAR(200) NOT NULL,
    permission_type VARCHAR(32) NOT NULL,
    menu_code VARCHAR(120),
    platform_only BOOLEAN NOT NULL DEFAULT FALSE,
    include_in_package BOOLEAN NOT NULL DEFAULT FALSE,
    data_perm_mode VARCHAR(16) NOT NULL DEFAULT 'ORG',
    manifest_hash VARCHAR(128) NOT NULL,
    managed_by_manifest BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_app_perm_code ON sys_app_permission(app_code, permission_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sys_app_perm_menu ON sys_app_permission(menu_code);
CREATE INDEX IF NOT EXISTS idx_sys_app_perm_status ON sys_app_permission(status);

CREATE TABLE IF NOT EXISTS sys_app_package_feature (
    id BIGSERIAL PRIMARY KEY,
    app_code VARCHAR(100) NOT NULL,
    feature_code VARCHAR(100) NOT NULL,
    feature_name VARCHAR(200) NOT NULL,
    feature_type VARCHAR(32) NOT NULL,
    parent_code VARCHAR(100),
    source_code VARCHAR(120),
    package_policy VARCHAR(32) NOT NULL DEFAULT 'IN_PACKAGE',
    include_in_package BOOLEAN NOT NULL DEFAULT TRUE,
    description VARCHAR(500),
    manifest_hash VARCHAR(128) NOT NULL,
    managed_by_manifest BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_app_feature_code ON sys_app_package_feature(app_code, feature_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sys_app_feature_status ON sys_app_package_feature(status);

CREATE TABLE IF NOT EXISTS sys_app_quota (
    id BIGSERIAL PRIMARY KEY,
    app_code VARCHAR(100) NOT NULL,
    quota_code VARCHAR(100) NOT NULL,
    quota_name VARCHAR(200) NOT NULL,
    quota_type VARCHAR(32) NOT NULL,
    unit VARCHAR(32),
    period_type VARCHAR(32),
    include_in_package BOOLEAN NOT NULL DEFAULT TRUE,
    description VARCHAR(500),
    manifest_hash VARCHAR(128) NOT NULL,
    managed_by_manifest BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_app_quota_code ON sys_app_quota(app_code, quota_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sys_app_quota_status ON sys_app_quota(status);
