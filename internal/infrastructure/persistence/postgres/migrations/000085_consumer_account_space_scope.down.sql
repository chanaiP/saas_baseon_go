DROP INDEX IF EXISTS idx_sys_app_permission_tenant_scope;
ALTER TABLE sys_app_permission DROP COLUMN IF EXISTS tenant_scope;

DROP INDEX IF EXISTS idx_sys_app_entry_tenant_scope;
ALTER TABLE sys_app_entry DROP COLUMN IF EXISTS tenant_scope;

DROP INDEX IF EXISTS idx_permission_tenant_scope;
ALTER TABLE permission DROP COLUMN IF EXISTS tenant_scope;

DROP INDEX IF EXISTS idx_app_user_member_type;
DROP INDEX IF EXISTS idx_app_user_identity_user;
DROP INDEX IF EXISTS idx_app_user_account;
ALTER TABLE app_user
  DROP COLUMN IF EXISTS member_type,
  DROP COLUMN IF EXISTS identity_user_id,
  DROP COLUMN IF EXISTS account_id;

DROP INDEX IF EXISTS idx_tenant_owner_user;
DROP INDEX IF EXISTS idx_tenant_type;
ALTER TABLE tenant
  DROP COLUMN IF EXISTS owner_user_id,
  DROP COLUMN IF EXISTS tenant_type;

DROP INDEX IF EXISTS idx_user_identity_phone;
DROP INDEX IF EXISTS idx_user_identity_account;
DROP TABLE IF EXISTS user_identity;

DROP INDEX IF EXISTS idx_account_status;
DROP INDEX IF EXISTS uq_account_email_active;
DROP INDEX IF EXISTS uq_account_phone_active;
DROP INDEX IF EXISTS uq_account_login_account_active;
DROP TABLE IF EXISTS account;
