package bootstrap

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type BootstrapCheck struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Actual   int64  `json:"actual"`
	Expected string `json:"expected"`
}

type BootstrapVerifyResult struct {
	Checks []BootstrapCheck `json:"checks"`
}

func (r BootstrapVerifyResult) Passed() bool {
	for _, check := range r.Checks {
		if !check.Passed {
			return false
		}
	}
	return true
}

func (r BootstrapVerifyResult) Summary() string {
	lines := make([]string, 0, len(r.Checks)+1)
	status := "PASS"
	if !r.Passed() {
		status = "FAIL"
	}
	lines = append(lines, "bootstrap_verify_result="+status)
	for _, check := range r.Checks {
		result := "PASS"
		if !check.Passed {
			result = "FAIL"
		}
		lines = append(lines, fmt.Sprintf("%s actual=%d expected=%s result=%s", check.Name, check.Actual, check.Expected, result))
	}
	return strings.Join(lines, "\n")
}

func VerifyBootstrapData(db *gorm.DB) (BootstrapVerifyResult, error) {
	checks := []BootstrapCheck{}
	add := func(name string, actual int64, expected string, passed bool) {
		checks = append(checks, BootstrapCheck{Name: name, Actual: actual, Expected: expected, Passed: passed})
	}
	count := func(query string, args ...interface{}) (int64, error) {
		var value int64
		if err := db.Raw(query, args...).Scan(&value).Error; err != nil {
			return 0, err
		}
		return value, nil
	}

	requiredTables, err := count(`
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'tenant','org_node','app_user','file_object','role','permission','role_permission','saas_plan','saas_feature','saas_quota',
    'saas_plan_feature','saas_plan_quota','tenant_subscription','dict_type','dict_item','sys_param','audit_log','login_log','sys_app'
  )`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("required_tables", requiredTables, "19", requiredTables == 19)

	legacyParamTables, err := count("SELECT CASE WHEN to_regclass('public.system_param') IS NULL THEN 0 ELSE 1 END")
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("legacy_system_param_table", legacyParamTables, "0", legacyParamTables == 0)

	appUserSessionColumns, err := count(`
SELECT COUNT(*)
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = 'app_user'
  AND column_name IN ('session_version','password_changed_at')`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("app_user_session_columns", appUserSessionColumns, "2", appUserSessionColumns == 2)

	platformTenant, err := count("SELECT COUNT(*) FROM tenant WHERE code = 'platform' AND deleted_at IS NULL")
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("platform_tenant", platformTenant, "1", platformTenant == 1)

	platformTenantFlag, err := count("SELECT COUNT(*) FROM tenant WHERE code = 'platform' AND is_platform_tenant = true AND deleted_at IS NULL")
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("platform_tenant_flag", platformTenantFlag, "1", platformTenantFlag == 1)

	platformAdminUser, err := count(`
SELECT COUNT(*)
FROM app_user u
JOIN tenant t ON t.id = u.tenant_id
WHERE t.code = 'platform'
  AND u.employee_no = 'E10001'
  AND u.account = 'E10001'
  AND u.is_platform_admin = true
  AND u.deleted_at IS NULL`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("platform_admin_user", platformAdminUser, "1", platformAdminUser == 1)

	platformAdminMissingGrants, err := count(`
SELECT COUNT(*)
FROM permission p
JOIN tenant t ON t.id = p.tenant_id
JOIN role r ON r.tenant_id = t.id AND r.code = 'admin' AND r.deleted_at IS NULL
LEFT JOIN role_permission rp ON rp.role_id = r.id AND rp.permission_id = p.id
WHERE t.code = 'platform'
  AND p.enabled = true
  AND p.deleted_at IS NULL
  AND rp.id IS NULL`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("platform_admin_missing_grants", platformAdminMissingGrants, "0", platformAdminMissingGrants == 0)

	plans, err := count("SELECT COUNT(*) FROM saas_plan WHERE deleted_at IS NULL")
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("saas_plans", plans, "4", plans == 4)

	coreFeatures, err := count(`
SELECT COUNT(*)
FROM saas_feature
WHERE status = 1
  AND feature_code IN (
    'user_manage','org_manage','position_manage','role_manage','dict_manage','param_manage','business_unit_manage',
    'login_log','audit_log','import_data','export_data','file_manage','brand_config','system_monitor'
  )`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("core_saas_features", coreFeatures, "14", coreFeatures == 14)

	quotas, err := count(`
SELECT COUNT(*)
FROM saas_quota
WHERE status = 1
  AND quota_code IN (
    'max_users','max_companies','max_stores','max_departments','max_business_units','max_roles',
    'max_storage_gb','max_file_size_mb','daily_import_times','daily_export_times'
  )`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("saas_quotas", quotas, "10", quotas == 10)

	planQuotaMatrix, err := count(`
SELECT COUNT(*)
FROM saas_plan_quota spq
JOIN saas_quota sq ON sq.id = spq.quota_id
WHERE sq.quota_code IN (
  'max_users','max_companies','max_stores','max_departments','max_business_units','max_roles',
  'max_storage_gb','max_file_size_mb','daily_import_times','daily_export_times'
)`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("plan_quota_matrix", planQuotaMatrix, "40", planQuotaMatrix == 40)

	corePlanFeatureMatrix, err := count(`
SELECT COUNT(*)
FROM saas_plan_feature spf
JOIN saas_feature sf ON sf.id = spf.feature_id
WHERE sf.feature_code IN (
  'user_manage','org_manage','position_manage','role_manage','dict_manage','param_manage','business_unit_manage',
  'login_log','audit_log','import_data','export_data','file_manage','brand_config','system_monitor'
)`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("core_plan_feature_matrix", corePlanFeatureMatrix, "56", corePlanFeatureMatrix == 56)

	dictTypes, err := count(`
SELECT COUNT(*)
FROM dict_type dt
JOIN tenant t ON t.id = dt.tenant_id
WHERE t.code = 'platform' AND dt.deleted_at IS NULL`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("platform_dict_types", dictTypes, ">=4", dictTypes >= 4)

	sysParams, err := count(`
SELECT COUNT(*)
FROM sys_param sp
JOIN tenant t ON t.id = sp.tenant_id
WHERE t.code = 'platform' AND sp.deleted_at IS NULL`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("platform_sys_params", sysParams, ">=5", sysParams >= 5)

	builtinApps, err := count(`
SELECT COUNT(*)
FROM sys_app
WHERE app_code IN ('app-center','system-management','system-monitor')
  AND is_builtin = true
  AND is_platform_only = true
  AND deleted_at IS NULL`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("builtin_sys_apps", builtinApps, "3", builtinApps == 3)

	permissionRoots, err := count(`
SELECT COUNT(*)
FROM permission p
JOIN tenant t ON t.id = p.tenant_id
WHERE t.code = 'platform'
  AND p.path IN ('/home','home:view','__menu_root__','__operations_root__','data:home','data:perm','/permissions')
  AND p.deleted_at IS NULL`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("system_permission_roots", permissionRoots, "7", permissionRoots == 7)

	menuFeatureMismatches, err := count(`
SELECT COUNT(*)
FROM permission p
JOIN tenant t ON t.id = p.tenant_id
WHERE t.code = 'platform'
  AND p.path IN ('/roles','/menus','/permissions')
  AND p.deleted_at IS NULL
  AND (
    (p.path IN ('/roles','/permissions') AND COALESCE(p.feature_code, '') <> 'role_manage')
    OR (p.path = '/menus' AND COALESCE(p.feature_code, '') <> 'menu_manage')
  )`)
	if err != nil {
		return BootstrapVerifyResult{}, err
	}
	add("role_menu_feature_mismatches", menuFeatureMismatches, "0", menuFeatureMismatches == 0)

	relationChecks := []struct {
		name  string
		query string
	}{
		{"orphan_user_role_user", "SELECT COUNT(*) FROM user_role ur LEFT JOIN app_user u ON u.id = ur.user_id WHERE u.id IS NULL"},
		{"orphan_user_role_role", "SELECT COUNT(*) FROM user_role ur LEFT JOIN role r ON r.id = ur.role_id WHERE r.id IS NULL"},
		{"orphan_role_permission_role", "SELECT COUNT(*) FROM role_permission rp LEFT JOIN role r ON r.id = rp.role_id WHERE r.id IS NULL"},
		{"orphan_role_permission_permission", "SELECT COUNT(*) FROM role_permission rp LEFT JOIN permission p ON p.id = rp.permission_id WHERE p.id IS NULL"},
		{"orphan_tenant_subscription_tenant", "SELECT COUNT(*) FROM tenant_subscription ts LEFT JOIN tenant t ON t.id = ts.tenant_id WHERE t.id IS NULL"},
		{"orphan_tenant_subscription_plan", "SELECT COUNT(*) FROM tenant_subscription ts LEFT JOIN saas_plan sp ON sp.id = ts.plan_id WHERE sp.id IS NULL"},
		{"orphan_tenant_feature_override_feature", "SELECT COUNT(*) FROM tenant_feature_override tfo LEFT JOIN saas_feature sf ON sf.id = tfo.feature_id WHERE sf.id IS NULL"},
		{"orphan_tenant_quota_override_quota", "SELECT COUNT(*) FROM tenant_quota_override tqo LEFT JOIN saas_quota sq ON sq.id = tqo.quota_id WHERE sq.id IS NULL"},
		{"orphan_business_unit_org_map_bu", "SELECT COUNT(*) FROM business_unit_org_map m LEFT JOIN business_unit b ON b.id = m.business_unit_id WHERE b.id IS NULL"},
		{"orphan_business_unit_org_map_org", "SELECT COUNT(*) FROM business_unit_org_map m LEFT JOIN org_node o ON o.id = m.org_id WHERE o.id IS NULL"},
	}
	for _, relationCheck := range relationChecks {
		actual, err := count(relationCheck.query)
		if err != nil {
			return BootstrapVerifyResult{}, err
		}
		add(relationCheck.name, actual, "0", actual == 0)
	}

	return BootstrapVerifyResult{Checks: checks}, nil
}
