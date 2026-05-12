WITH seed(app_code, client_code, client_name, sort_order) AS (
  VALUES
    ('app-center', 'PC_WEB', 'PC Web', 1),
    ('system-management', 'PC_WEB', 'PC Web', 1),
    ('system-monitor', 'PC_WEB', 'PC Web', 1),
    ('workbench', 'PC_WEB', 'PC Web', 1),
    ('integration-center', 'PC_WEB', 'PC Web', 1),
    ('integration-center', 'API_ONLY', 'API Only', 2),
    ('data-center', 'PC_WEB', 'PC Web', 1),
    ('data-center', 'API_ONLY', 'API Only', 2)
)
INSERT INTO sys_app_client (
  app_id, client_code, client_name, enabled, sort_order, config_note, created_at, updated_at
)
SELECT app.id, seed.client_code, seed.client_name, TRUE, seed.sort_order, 'bootstrap:built-in-client', now(), now()
FROM seed
JOIN sys_app app ON app.app_code = seed.app_code AND app.deleted_at IS NULL
WHERE NOT EXISTS (
  SELECT 1
  FROM sys_app_client existing
  WHERE existing.app_id = app.id
    AND existing.client_code = seed.client_code
    AND existing.deleted_at IS NULL
);
