WITH platform_tenants AS (
  SELECT id
  FROM tenant
  WHERE deleted_at IS NULL
    AND (is_platform = true OR is_platform_tenant = true OR code = 'platform')
),
operation_rows AS (
  SELECT *
  FROM (
    VALUES
      ('应用-编辑', 'app:edit', 20),
      ('应用-启停', 'app:status', 21)
  ) AS v(name, path, sort_order)
),
inserted AS (
  INSERT INTO permission (
    tenant_id, parent_id, name, path, perm_type, sort_order, enabled, visible,
    is_platform_only, is_package_feature, tenant_editable, app_code,
    feature_type, data_perm_mode, created_at, updated_at
  )
  SELECT
    pt.id,
    root.id,
    op.name,
    op.path,
    2,
    op.sort_order,
    true,
    true,
    true,
    false,
    false,
    'app-center',
    'OPERATION',
    'NONE',
    now(),
    now()
  FROM platform_tenants pt
  CROSS JOIN operation_rows op
  LEFT JOIN permission root
    ON root.tenant_id = pt.id
   AND root.path = '__operations_root__'
   AND root.deleted_at IS NULL
  WHERE NOT EXISTS (
    SELECT 1
    FROM permission existing
    WHERE existing.tenant_id = pt.id
      AND existing.path = op.path
      AND existing.deleted_at IS NULL
  )
  RETURNING id
)
UPDATE permission p
SET name = op.name,
    perm_type = 2,
    sort_order = op.sort_order,
    enabled = true,
    visible = true,
    is_platform_only = true,
    is_package_feature = false,
    tenant_editable = false,
    app_code = 'app-center',
    feature_type = 'OPERATION',
    data_perm_mode = 'NONE',
    updated_at = now()
FROM platform_tenants pt
JOIN operation_rows op ON true
WHERE p.tenant_id = pt.id
  AND p.path = op.path
  AND p.deleted_at IS NULL;

INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT rp.role_id, op.id, now()
FROM role_permission rp
JOIN permission app
  ON app.id = rp.permission_id
 AND app.path = '/apps'
 AND app.deleted_at IS NULL
JOIN permission op
  ON op.tenant_id = app.tenant_id
 AND op.path IN ('app:edit', 'app:status')
 AND op.deleted_at IS NULL
ON CONFLICT (role_id, permission_id) DO NOTHING;
