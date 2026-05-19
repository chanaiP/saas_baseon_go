DELETE FROM role_permission rp
USING permission p
WHERE rp.permission_id = p.id
  AND p.path IN (
    'business_unit:relation_manage',
    'business_unit:actor_manage',
    'business_unit:attr_template_manage'
  );

DELETE FROM saas_plan_feature spf
USING saas_feature sf
WHERE spf.feature_id = sf.id
  AND sf.feature_code IN (
    'button_business_unit_relation_manage',
    'button_business_unit_actor_manage',
    'button_business_unit_attr_template_manage'
  );

DELETE FROM saas_feature
WHERE feature_code IN (
  'button_business_unit_relation_manage',
  'button_business_unit_actor_manage',
  'button_business_unit_attr_template_manage'
);

DELETE FROM permission
WHERE path IN (
  'business_unit:relation_manage',
  'business_unit:actor_manage',
  'business_unit:attr_template_manage'
)
AND perm_type = 2;
