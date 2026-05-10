INSERT INTO saas_feature (feature_code, feature_name, feature_type, parent_id, status, description, created_at, updated_at)
VALUES ('menu_manage', '菜单管理', 'MENU', 0, 1, '租户菜单显示名称、图标与侧栏可见性覆盖', now(), now())
ON CONFLICT (feature_code) DO UPDATE
SET feature_name = EXCLUDED.feature_name,
    feature_type = EXCLUDED.feature_type,
    status = 1,
    description = EXCLUDED.description,
    updated_at = now();

INSERT INTO saas_feature (feature_code, feature_name, feature_type, parent_id, status, description, created_at, updated_at)
VALUES (
  'button_menu_edit',
  '菜单-编辑',
  'OPERATION',
  COALESCE((SELECT id FROM saas_feature WHERE feature_code = 'menu_manage' LIMIT 1), 0),
  1,
  '允许租户编辑套餐内菜单的显示名称和图标',
  now(),
  now()
)
ON CONFLICT (feature_code) DO UPDATE
SET feature_name = EXCLUDED.feature_name,
    feature_type = EXCLUDED.feature_type,
    parent_id = EXCLUDED.parent_id,
    status = 1,
    description = EXCLUDED.description,
    updated_at = now();

INSERT INTO saas_plan_feature (plan_id, feature_id, enabled, created_at, updated_at)
SELECT p.id, f.id, true, now(), now()
FROM saas_plan p
JOIN saas_feature f ON f.feature_code IN ('menu_manage', 'button_menu_edit')
WHERE p.plan_code IN ('TRIAL', 'BASIC', 'PRO', 'ENTERPRISE')
  AND p.deleted_at IS NULL
ON CONFLICT (plan_id, feature_id) DO UPDATE
SET enabled = true,
    updated_at = now();
