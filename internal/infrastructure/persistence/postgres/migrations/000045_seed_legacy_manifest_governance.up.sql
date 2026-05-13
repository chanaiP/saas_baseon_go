-- Backfill existing platform resources into Manifest governance tables.
-- Older resources were already present in permission / SaaS tables before
-- sys_app_* Manifest asset tables existed. Without this baseline, runtime
-- Manifest diff would keep treating legacy menus and permissions as new.

WITH platform_scope AS (
    SELECT COALESCE(
        (SELECT id FROM tenant WHERE is_platform_tenant = TRUE AND deleted_at IS NULL ORDER BY id LIMIT 1),
        (SELECT id FROM tenant WHERE code = 'platform' AND deleted_at IS NULL ORDER BY id LIMIT 1),
        1
    ) AS tenant_id
),
permission_baseline AS (
    SELECT
        p.*,
        COALESCE(
            NULLIF(TRIM(p.feature_code), ''),
            NULLIF(TRIM(BOTH '_' FROM regexp_replace(TRIM(BOTH '/' FROM lower(COALESCE(p.path, ''))), '[^a-z0-9]+', '_', 'g')), ''),
            'permission_' || p.id::text
        ) AS baseline_code
    FROM permission p
    WHERE p.tenant_id = (SELECT tenant_id FROM platform_scope)
      AND p.deleted_at IS NULL
      AND COALESCE(p.app_code, '') <> ''
      AND COALESCE(p.path, '') NOT IN ('__menu_root__', '__operations_root__')
)
INSERT INTO sys_app_entry (
    app_code,
    resource_code,
    name,
    path,
    parent_code,
    sort_order,
    platform_only,
    tenant_visible,
    tenant_editable,
    include_in_package,
    feature_code,
    data_perm_mode,
    manifest_hash,
    managed_by_manifest,
    status,
    last_synced_at,
    created_at,
    updated_at
)
SELECT
    p.app_code,
    p.baseline_code,
    p.name,
    COALESCE(p.path, ''),
    parent.baseline_code,
    p.sort_order,
    p.is_platform_only,
    p.visible,
    p.tenant_editable,
    p.is_package_feature,
    NULLIF(TRIM(p.feature_code), ''),
    COALESCE(NULLIF(TRIM(p.data_perm_mode), ''), 'ORG'),
    'LEGACY_BASELINE',
    TRUE,
    CASE WHEN p.enabled THEN 'ACTIVE' ELSE 'DISABLED' END,
    now(),
    now(),
    now()
FROM permission_baseline p
LEFT JOIN permission_baseline parent ON parent.id = p.parent_id
WHERE p.perm_type = 3
ON CONFLICT DO NOTHING;

WITH platform_scope AS (
    SELECT COALESCE(
        (SELECT id FROM tenant WHERE is_platform_tenant = TRUE AND deleted_at IS NULL ORDER BY id LIMIT 1),
        (SELECT id FROM tenant WHERE code = 'platform' AND deleted_at IS NULL ORDER BY id LIMIT 1),
        1
    ) AS tenant_id
),
permission_baseline AS (
    SELECT
        p.*,
        COALESCE(
            NULLIF(TRIM(p.feature_code), ''),
            NULLIF(TRIM(BOTH '_' FROM regexp_replace(TRIM(BOTH '/' FROM lower(COALESCE(p.path, ''))), '[^a-z0-9]+', '_', 'g')), ''),
            'permission_' || p.id::text
        ) AS baseline_code,
        COALESCE(
            NULLIF(TRIM(p.path), ''),
            NULLIF(TRIM(p.feature_code), ''),
            'permission_' || p.id::text
        ) AS baseline_permission_code
    FROM permission p
    WHERE p.tenant_id = (SELECT tenant_id FROM platform_scope)
      AND p.deleted_at IS NULL
      AND COALESCE(p.app_code, '') <> ''
      AND COALESCE(p.path, '') NOT IN ('__menu_root__', '__operations_root__')
)
INSERT INTO sys_app_permission (
    app_code,
    permission_code,
    name,
    permission_type,
    menu_code,
    platform_only,
    include_in_package,
    data_perm_mode,
    manifest_hash,
    managed_by_manifest,
    status,
    last_synced_at,
    created_at,
    updated_at
)
SELECT
    p.app_code,
    p.baseline_permission_code,
    p.name,
    COALESCE(NULLIF(TRIM(p.feature_type), ''), CASE WHEN p.perm_type = 2 THEN 'OPERATION' ELSE 'PERMISSION' END),
    parent.baseline_code,
    p.is_platform_only,
    p.is_package_feature,
    COALESCE(NULLIF(TRIM(p.data_perm_mode), ''), 'ORG'),
    'LEGACY_BASELINE',
    TRUE,
    CASE WHEN p.enabled THEN 'ACTIVE' ELSE 'DISABLED' END,
    now(),
    now(),
    now()
FROM permission_baseline p
LEFT JOIN permission_baseline parent ON parent.id = p.parent_id AND parent.perm_type = 3
WHERE p.perm_type <> 3
ON CONFLICT DO NOTHING;

WITH feature_baseline AS (
    SELECT
        f.*,
        NULLIF(TRIM(parent.feature_code), '') AS parent_code
    FROM saas_feature f
    LEFT JOIN saas_feature parent ON parent.id = f.parent_id
    WHERE COALESCE(f.app_code, '') <> ''
      AND COALESCE(f.feature_code, '') <> ''
)
INSERT INTO sys_app_package_feature (
    app_code,
    feature_code,
    feature_name,
    feature_type,
    parent_code,
    source_code,
    package_policy,
    include_in_package,
    description,
    manifest_hash,
    managed_by_manifest,
    status,
    last_synced_at,
    created_at,
    updated_at
)
SELECT
    f.app_code,
    f.feature_code,
    f.feature_name,
    f.feature_type,
    f.parent_code,
    NULLIF(TRIM(COALESCE(f.api_method, '') || ' ' || COALESCE(f.api_path, '')), ''),
    'IN_PACKAGE',
    TRUE,
    f.description,
    'LEGACY_BASELINE',
    TRUE,
    CASE WHEN f.status = 1 THEN 'ACTIVE' ELSE 'DISABLED' END,
    now(),
    now(),
    now()
FROM feature_baseline f
ON CONFLICT DO NOTHING;

INSERT INTO sys_app_quota (
    app_code,
    quota_code,
    quota_name,
    quota_type,
    unit,
    period_type,
    include_in_package,
    description,
    manifest_hash,
    managed_by_manifest,
    status,
    last_synced_at,
    created_at,
    updated_at
)
SELECT
    'system-management',
    q.quota_code,
    q.quota_name,
    q.quota_type,
    q.unit,
    q.period_type,
    TRUE,
    q.description,
    'LEGACY_BASELINE',
    TRUE,
    CASE WHEN q.status = 1 THEN 'ACTIVE' ELSE 'DISABLED' END,
    now(),
    now(),
    now()
FROM saas_quota q
WHERE COALESCE(q.quota_code, '') <> ''
ON CONFLICT DO NOTHING;
