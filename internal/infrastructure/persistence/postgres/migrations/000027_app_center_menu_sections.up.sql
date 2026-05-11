WITH platform_tenants AS (
  SELECT id
  FROM tenant
  WHERE deleted_at IS NULL
    AND (is_platform = true OR is_platform_tenant = true OR code = 'platform')
),
app_menu AS (
  UPDATE permission p
  SET name = '应用列表',
      perm_type = 3,
      sort_order = 1,
      enabled = true,
      visible = true,
      is_platform_only = true,
      is_package_feature = false,
      tenant_editable = false,
      app_code = 'app-center',
      feature_code = 'app_list',
      feature_type = 'MENU',
      data_perm_mode = 'NONE',
      updated_at = now()
  FROM platform_tenants pt
  WHERE p.tenant_id = pt.id
    AND p.path = '/apps'
    AND p.deleted_at IS NULL
  RETURNING p.tenant_id, p.id
),
inserted AS (
  INSERT INTO permission (
    tenant_id, parent_id, name, path, perm_type, sort_order, enabled, visible,
    is_platform_only, is_package_feature, tenant_editable, app_code, feature_code,
    feature_type, data_perm_mode, created_at, updated_at
  )
  SELECT
    pt.id,
    app.id,
    v.name,
    v.path,
    3,
    v.sort_order,
    true,
    true,
    true,
    false,
    false,
    'app-center',
    v.feature_code,
    'MENU',
    'NONE',
    now(),
    now()
  FROM platform_tenants pt
  JOIN permission app
    ON app.tenant_id = pt.id
   AND app.path = '/apps'
   AND app.deleted_at IS NULL
  CROSS JOIN (
    VALUES
      ('客户端中心', '/apps/clients', 2, 'app_clients'),
      ('租户开通总览', '/apps/tenant-openings', 3, 'app_tenant_openings'),
      ('体验邀请总览', '/apps/trial-invites', 4, 'app_trial_invites'),
      ('Manifest 装载记录', '/apps/manifests', 5, 'app_manifest_loads'),
      ('应用审计日志', '/apps/audit-logs', 6, 'app_audit_logs')
  ) AS v(name, path, sort_order, feature_code)
  WHERE NOT EXISTS (
    SELECT 1
    FROM permission existing
    WHERE existing.tenant_id = pt.id
      AND existing.path = v.path
      AND existing.deleted_at IS NULL
  )
  RETURNING id
)
UPDATE permission p
SET parent_id = app.id,
    name = v.name,
    perm_type = 3,
    sort_order = v.sort_order,
    enabled = true,
    visible = true,
    is_platform_only = true,
    is_package_feature = false,
    tenant_editable = false,
    app_code = 'app-center',
    feature_code = v.feature_code,
    feature_type = 'MENU',
    data_perm_mode = 'NONE',
    updated_at = now()
FROM platform_tenants pt
JOIN permission app
  ON app.tenant_id = pt.id
 AND app.path = '/apps'
 AND app.deleted_at IS NULL
JOIN (
  VALUES
    ('客户端中心', '/apps/clients', 2, 'app_clients'),
    ('租户开通总览', '/apps/tenant-openings', 3, 'app_tenant_openings'),
    ('体验邀请总览', '/apps/trial-invites', 4, 'app_trial_invites'),
    ('Manifest 装载记录', '/apps/manifests', 5, 'app_manifest_loads'),
    ('应用审计日志', '/apps/audit-logs', 6, 'app_audit_logs')
) AS v(name, path, sort_order, feature_code)
  ON true
WHERE p.tenant_id = pt.id
  AND p.path = v.path
  AND p.deleted_at IS NULL;

UPDATE permission
SET app_code = 'app-center',
    updated_at = now()
WHERE deleted_at IS NULL
  AND (path = '/apps' OR path LIKE '/apps/%' OR path LIKE 'app:%' OR path LIKE 'data:app%');

INSERT INTO role_permission (role_id, permission_id, created_at)
SELECT rp.role_id, child.id, now()
FROM role_permission rp
JOIN permission app
  ON app.id = rp.permission_id
 AND app.path = '/apps'
 AND app.deleted_at IS NULL
JOIN permission child
  ON child.tenant_id = app.tenant_id
 AND child.path IN (
   '/apps/clients',
   '/apps/tenant-openings',
   '/apps/trial-invites',
   '/apps/manifests',
   '/apps/audit-logs'
 )
 AND child.deleted_at IS NULL
ON CONFLICT (role_id, permission_id) DO NOTHING;
