UPDATE saas_plan_feature spf
SET enabled = false,
    updated_at = now()
FROM saas_plan p, saas_feature f
WHERE spf.plan_id = p.id
  AND spf.feature_id = f.id
  AND p.plan_code IN ('TRIAL', 'BASIC', 'PRO', 'ENTERPRISE')
  AND f.feature_code = 'button_menu_edit';
