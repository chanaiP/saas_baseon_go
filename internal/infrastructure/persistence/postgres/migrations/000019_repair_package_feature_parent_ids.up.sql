WITH parent_map(prefix, parent_code) AS (
  VALUES
    ('user', 'user_manage'),
    ('org', 'org_manage'),
    ('pos', 'position_manage'),
    ('role', 'role_manage'),
    ('menu', 'menu_manage'),
    ('dict', 'dict_manage'),
    ('dict_item', 'dict_manage'),
    ('dict_type', 'dict_manage'),
    ('param', 'param_manage'),
    ('business_unit', 'business_unit_manage'),
    ('audit', 'audit_log'),
    ('login', 'login_log'),
    ('monhealth', 'system_monitor'),
    ('monserver', 'system_monitor'),
    ('monjobs', 'system_monitor'),
    ('monservices', 'system_monitor'),
    ('moncache', 'system_monitor'),
    ('moncachekeys', 'system_monitor'),
    ('brand', 'brand_config')
)
UPDATE saas_feature child
SET
  parent_id = parent.id,
  updated_at = NOW()
FROM parent_map pm
JOIN saas_feature parent ON parent.feature_code = pm.parent_code
WHERE child.id <> parent.id
  AND child.status = 1
  AND child.feature_type IN ('BUTTON', 'OPERATION')
  AND (
    child.feature_name LIKE pm.prefix || ':%'
    OR child.feature_code LIKE 'button_' || replace(pm.prefix, '_', E'\\_') || E'\\_%' ESCAPE E'\\'
  )
  AND child.parent_id IS DISTINCT FROM parent.id;
