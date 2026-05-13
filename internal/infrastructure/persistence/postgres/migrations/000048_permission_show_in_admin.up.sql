ALTER TABLE permission
    ADD COLUMN IF NOT EXISTS show_in_admin BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE permission
SET show_in_admin = CASE
    WHEN perm_type = 3 AND visible = TRUE THEN TRUE
    ELSE FALSE
END
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_permission_show_in_admin
    ON permission(show_in_admin);

ALTER TABLE sys_app_entry
    ADD COLUMN IF NOT EXISTS show_in_admin BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE sys_app_entry
SET show_in_admin = TRUE
WHERE deleted_at IS NULL;
