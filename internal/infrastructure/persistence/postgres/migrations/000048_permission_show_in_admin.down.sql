DROP INDEX IF EXISTS idx_permission_show_in_admin;

ALTER TABLE sys_app_entry
    DROP COLUMN IF EXISTS show_in_admin;

ALTER TABLE permission
    DROP COLUMN IF EXISTS show_in_admin;
