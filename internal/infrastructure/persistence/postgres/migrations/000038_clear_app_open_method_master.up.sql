UPDATE sys_app
SET open_method = NULL,
    updated_at = NOW()
WHERE open_method IS NOT NULL;
