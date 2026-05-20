CREATE TABLE IF NOT EXISTS account (
  id BIGSERIAL PRIMARY KEY,
  login_account VARCHAR(128) NOT NULL,
  phone VARCHAR(32),
  email VARCHAR(200),
  password_hash VARCHAR(200) NOT NULL,
  status BIGINT NOT NULL DEFAULT 1,
  session_version BIGINT NOT NULL DEFAULT 1,
  password_changed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_account_login_account_active
  ON account(login_account)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_account_phone_active
  ON account(phone)
  WHERE phone IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_account_email_active
  ON account(email)
  WHERE email IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_account_status
  ON account(status)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user_identity (
  id BIGSERIAL PRIMARY KEY,
  account_id BIGINT NOT NULL REFERENCES account(id),
  display_name VARCHAR(100) NOT NULL,
  avatar_url TEXT,
  phone VARCHAR(32),
  email VARCHAR(200),
  status BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_user_identity_account
  ON user_identity(account_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_identity_phone
  ON user_identity(phone)
  WHERE phone IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE tenant
  ADD COLUMN IF NOT EXISTS tenant_type VARCHAR(32) NOT NULL DEFAULT 'enterprise',
  ADD COLUMN IF NOT EXISTS owner_user_id BIGINT;

UPDATE tenant
SET tenant_type = CASE WHEN is_platform_tenant THEN 'platform' ELSE 'enterprise' END
WHERE tenant_type IS NULL OR tenant_type = '' OR is_platform_tenant = TRUE;

CREATE INDEX IF NOT EXISTS idx_tenant_type
  ON tenant(tenant_type)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tenant_owner_user
  ON tenant(owner_user_id)
  WHERE owner_user_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE app_user
  ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES account(id),
  ADD COLUMN IF NOT EXISTS identity_user_id BIGINT REFERENCES user_identity(id),
  ADD COLUMN IF NOT EXISTS member_type VARCHAR(32) NOT NULL DEFAULT 'enterprise_member';

UPDATE app_user
SET member_type = CASE
  WHEN is_platform_admin THEN 'platform_admin'
  WHEN is_tenant_admin THEN 'enterprise_admin'
  ELSE 'enterprise_member'
END
WHERE member_type IS NULL OR member_type = '';

CREATE INDEX IF NOT EXISTS idx_app_user_account
  ON app_user(account_id)
  WHERE account_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_app_user_identity_user
  ON app_user(identity_user_id)
  WHERE identity_user_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_app_user_member_type
  ON app_user(member_type)
  WHERE deleted_at IS NULL;

ALTER TABLE permission
  ADD COLUMN IF NOT EXISTS tenant_scope VARCHAR(32) NOT NULL DEFAULT 'enterprise_only';

UPDATE permission
SET tenant_scope = CASE
  WHEN is_platform_only THEN 'platform_only'
  WHEN path IN ('/home', '/profile', 'brand:edit') THEN 'all'
  WHEN path IN (
    '/data-center',
    '/data-center/dashboard',
    '/data-center/overview',
    '/data-center/raw',
    '/data-center/standard',
    '/data-center/metrics',
    '/data-center/anomalies',
    '/data-center/rules',
    '/data-center/tasks',
    '/data-center/reviews',
    '/integration-center/my-connections'
  ) THEN 'enterprise_personal'
  ELSE 'enterprise_only'
END
WHERE tenant_scope IS NULL OR tenant_scope = '' OR tenant_scope = 'enterprise_only';

CREATE INDEX IF NOT EXISTS idx_permission_tenant_scope
  ON permission(tenant_scope)
  WHERE deleted_at IS NULL;

ALTER TABLE sys_app_entry
  ADD COLUMN IF NOT EXISTS tenant_scope VARCHAR(32) NOT NULL DEFAULT 'enterprise_only';

UPDATE sys_app_entry
SET tenant_scope = CASE
  WHEN platform_only THEN 'platform_only'
  WHEN path IN ('/home', '/profile') THEN 'all'
  WHEN path IN (
    '/data-center',
    '/data-center/dashboard',
    '/data-center/overview',
    '/data-center/raw',
    '/data-center/standard',
    '/data-center/metrics',
    '/data-center/anomalies',
    '/data-center/rules',
    '/data-center/tasks',
    '/data-center/reviews',
    '/integration-center/my-connections'
  ) THEN 'enterprise_personal'
  ELSE 'enterprise_only'
END
WHERE tenant_scope IS NULL OR tenant_scope = '' OR tenant_scope = 'enterprise_only';

CREATE INDEX IF NOT EXISTS idx_sys_app_entry_tenant_scope
  ON sys_app_entry(tenant_scope)
  WHERE deleted_at IS NULL;

ALTER TABLE sys_app_permission
  ADD COLUMN IF NOT EXISTS tenant_scope VARCHAR(32) NOT NULL DEFAULT 'enterprise_only';

UPDATE sys_app_permission
SET tenant_scope = CASE
  WHEN platform_only THEN 'platform_only'
  WHEN app_code = 'data-center' THEN 'enterprise_personal'
  WHEN permission_code = '/integration-center/my-connections' THEN 'enterprise_personal'
  ELSE 'enterprise_only'
END
WHERE tenant_scope IS NULL OR tenant_scope = '' OR tenant_scope = 'enterprise_only';

CREATE INDEX IF NOT EXISTS idx_sys_app_permission_tenant_scope
  ON sys_app_permission(tenant_scope)
  WHERE deleted_at IS NULL;
