package bootstrap

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func seedCoreData(db *gorm.DB) error {
	footer := "© 2026 SaaS - AI协作开发系统"
	brand := "Ai DevOS"
	tenant := models.Tenant{
		Code:       "platform",
		Name:       "平台主体",
		Status:     1,
		IsPlatform: true,
		BrandName:  &brand,
		FooterText: &footer,
	}
	if err := db.Where("code = ?", tenant.Code).FirstOrCreate(&tenant).Error; err != nil {
		return err
	}
	if err := db.Model(&tenant).Updates(map[string]interface{}{
		"name":               "平台主体",
		"status":             1,
		"is_platform_tenant": true,
		"brand_display_name": brand,
		"brand_footer_text":  footer,
	}).Error; err != nil {
		return err
	}

	company, positionType, position, err := seedPlatformStructure(db, tenant.ID)
	if err != nil {
		return err
	}

	initialPassword, err := seedInitialAdminPassword()
	if err != nil {
		return err
	}
	defaultPasswordHash, err := seedPasswordHash(initialPassword)
	if err != nil {
		return err
	}
	user := models.AppUser{
		TenantID:        tenant.ID,
		CompanyID:       &company.ID,
		EmployeeNo:      "E10001",
		Account:         "E10001",
		PasswordHash:    defaultPasswordHash,
		Name:            "平台管理员",
		Status:          1,
		IsPlatformAdmin: true,
	}
	if err := db.Where("tenant_id = ? AND account IN ?", tenant.ID, []string{"E10001", "admin"}).FirstOrCreate(&user).Error; err != nil {
		return err
	}
	if err := db.Model(&user).Updates(map[string]interface{}{"company_id": company.ID, "employee_no": "E10001", "account": "E10001", "is_platform_admin": true, "status": 1}).Error; err != nil {
		return err
	}

	role := models.Role{
		TenantID: tenant.ID,
		Code:     "admin",
		Name:     "超级管理员",
		Status:   1,
	}
	if err := db.Where("tenant_id = ? AND code = ?", tenant.ID, role.Code).FirstOrCreate(&role).Error; err != nil {
		return err
	}

	if err := db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).FirstOrCreate(&models.UserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		return err
	}
	if err := db.Where("user_id = ? AND position_id = ?", user.ID, position.ID).FirstOrCreate(&models.AppUserPosition{UserID: user.ID, PositionID: position.ID}).Error; err != nil {
		return err
	}
	demoPhone := "13900000000"
	demoUser := models.AppUser{
		TenantID:     tenant.ID,
		CompanyID:    &company.ID,
		EmployeeNo:   "E10100",
		Account:      "E10100",
		PasswordHash: defaultPasswordHash,
		Name:         "演示用户",
		Phone:        &demoPhone,
		Status:       1,
	}
	if err := db.Where("tenant_id = ? AND account = ?", tenant.ID, demoUser.Account).FirstOrCreate(&demoUser).Error; err != nil {
		return err
	}
	if err := db.Model(&demoUser).Updates(map[string]interface{}{"company_id": company.ID, "employee_no": "E10100", "status": 1}).Error; err != nil {
		return err
	}
	if err := db.Where("user_id = ? AND role_id = ?", demoUser.ID, role.ID).FirstOrCreate(&models.UserRole{UserID: demoUser.ID, RoleID: role.ID}).Error; err != nil {
		return err
	}
	_ = positionType

	permissions, err := seedPermissions(db, tenant.ID)
	if err != nil {
		return err
	}
	for _, permission := range permissions {
		link := models.RolePermission{RoleID: role.ID, PermissionID: permission.ID}
		if err := db.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).FirstOrCreate(&link).Error; err != nil {
			return err
		}
	}

	if err := seedSaasPlans(db, tenant.ID); err != nil {
		return err
	}
	if err := seedDictionaries(db, tenant.ID); err != nil {
		return err
	}
	if err := seedSystemParams(db, tenant.ID); err != nil {
		return err
	}
	return seedAuditSamples(db, tenant.ID, user.ID)
}

func seedPasswordHash(password string) (string, error) {
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	salt := hex.EncodeToString(saltBytes)
	digest := pbkdf2.Key([]byte(password), []byte(salt), 390000, 32, sha256.New)
	return "pbkdf2_sha256$" + salt + "$" + hex.EncodeToString(digest), nil
}

func seedInitialAdminPassword() (string, error) {
	password := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"))
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
		if password == "" || password == "112233" {
			return "", errors.New("production requires explicit non-default BOOTSTRAP_ADMIN_PASSWORD for bootstrap users")
		}
		if err := validateBootstrapPassword(password); err != nil {
			return "", err
		}
		return password, nil
	}
	if password != "" {
		if err := validateBootstrapPassword(password); err != nil {
			return "", err
		}
		return password, nil
	}
	return "112233", nil
}

func validateBootstrapPassword(password string) error {
	if len(password) < 12 || len(password) > 128 {
		return errors.New("BOOTSTRAP_ADMIN_PASSWORD must be 12-128 characters")
	}
	hasLetter := false
	hasDigit := false
	for _, ch := range password {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetter = true
		}
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("BOOTSTRAP_ADMIN_PASSWORD must contain letters and digits")
	}
	return nil
}

func seedPlatformStructure(db *gorm.DB, tenantID uint64) (models.OrgNode, models.PositionType, models.Position, error) {
	companyCode := "PLATFORM"
	companyType := "GROUP"
	company := models.OrgNode{
		TenantID:    tenantID,
		NodeType:    "company",
		Name:        "平台公司",
		Code:        &companyCode,
		CompanyType: &companyType,
		Status:      1,
	}
	if err := db.Where("tenant_id = ? AND code = ?", tenantID, companyCode).FirstOrCreate(&company).Error; err != nil {
		return company, models.PositionType{}, models.Position{}, err
	}

	positionType := models.PositionType{TenantID: tenantID, Name: "管理序列", Code: "MGT"}
	if err := db.Where("tenant_id = ? AND code = ?", tenantID, positionType.Code).FirstOrCreate(&positionType).Error; err != nil {
		return company, positionType, models.Position{}, err
	}

	position := models.Position{TenantID: tenantID, PositionTypeID: positionType.ID, Name: "系统管理员", Code: "SYS_ADMIN"}
	if err := db.Where("tenant_id = ? AND code = ?", tenantID, position.Code).FirstOrCreate(&position).Error; err != nil {
		return company, positionType, position, err
	}

	return company, positionType, position, nil
}

type seedPermission struct {
	Name            string
	Path            string
	Type            int
	SortOrder       int
	Hidden          bool
	PlatformOnly    bool
	PackageFeature  bool
	TenantEditable  bool
	TenantEditScope string
	FeatureCode     string
	FeatureType     string
	DataPermMode    string
}

func seedPermissions(db *gorm.DB, tenantID uint64) ([]models.Permission, error) {
	items := []seedPermission{
		{Name: "菜单根节点", Path: "__menu_root__", Type: 1, SortOrder: 0, Hidden: true, PackageFeature: false, FeatureType: "SYSTEM", DataPermMode: "NONE"},
		{Name: "操作根节点", Path: "__operations_root__", Type: 1, SortOrder: 0, Hidden: true, PackageFeature: false, FeatureType: "SYSTEM", DataPermMode: "NONE"},
		{Name: "首页", Path: "/home", Type: 3, SortOrder: 0, Hidden: true, PackageFeature: false, FeatureCode: "home", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "主体管理", Path: "/tenants", Type: 3, SortOrder: 1, PlatformOnly: true, PackageFeature: false, FeatureCode: "tenant_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "套餐中心", Path: "/plans", Type: 3, SortOrder: 2, PlatformOnly: true, PackageFeature: false, FeatureCode: "plan_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "组织架构", Path: "/organization", Type: 3, SortOrder: 3, FeatureCode: "org_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "岗位管理", Path: "/positions", Type: 3, SortOrder: 4, FeatureCode: "position_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "业务单元", Path: "/business-units", Type: 3, SortOrder: 5, FeatureCode: "business_unit_manage", FeatureType: "MENU", DataPermMode: "BU"},
		{Name: "用户管理", Path: "/users", Type: 3, SortOrder: 6, FeatureCode: "user_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "角色权限", Path: "/roles", Type: 3, SortOrder: 7, FeatureCode: "role_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "菜单管理", Path: "/menus", Type: 3, SortOrder: 8, FeatureCode: "role_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "权限管理兼容入口", Path: "/permissions", Type: 3, SortOrder: 8, Hidden: true, FeatureCode: "role_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "数据字典", Path: "/dict", Type: 3, SortOrder: 9, FeatureCode: "dict_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "参数管理", Path: "/params", Type: 3, SortOrder: 10, FeatureCode: "param_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "操作日志", Path: "/audit-logs", Type: 3, SortOrder: 11, FeatureCode: "audit_log", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "登录日志", Path: "/login-logs", Type: 3, SortOrder: 12, FeatureCode: "login_log", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "健康检查", Path: "/monitor/health", Type: 3, SortOrder: 101, PlatformOnly: true, FeatureCode: "system_monitor", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "服务器信息", Path: "/monitor/server", Type: 3, SortOrder: 102, PlatformOnly: true, FeatureCode: "system_monitor", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "定时任务", Path: "/monitor/jobs", Type: 3, SortOrder: 103, PlatformOnly: true, FeatureCode: "system_monitor", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "服务监控", Path: "/monitor/services", Type: 3, SortOrder: 104, PlatformOnly: true, FeatureCode: "system_monitor", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "缓存监控", Path: "/monitor/cache", Type: 3, SortOrder: 105, PlatformOnly: true, FeatureCode: "system_monitor", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "缓存列表", Path: "/monitor/cache-keys", Type: 3, SortOrder: 106, PlatformOnly: true, FeatureCode: "system_monitor", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "主体-新增", Path: "tenant:create", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "主体-编辑", Path: "tenant:edit", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "主体-删除", Path: "tenant:delete", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "主体-重置密码", Path: "tenant:reset_password", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "主体-配额配置", Path: "tenant:quota_config", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "套餐-新增", Path: "plan:create", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "套餐-编辑", Path: "plan:edit", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "套餐-删除", Path: "plan:delete", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
		{Name: "套餐-配置", Path: "plan:config", Type: 2, PlatformOnly: true, PackageFeature: false, FeatureType: "OPERATION", DataPermMode: "NONE"},
	}
	for _, path := range []string{
		"home:view",
		"org:create", "org:edit", "org:delete",
		"pos:create", "pos:edit", "pos:delete",
		"business_unit:create", "business_unit:edit", "business_unit:delete",
		"user:create", "user:edit", "user:reset_password", "user:delete",
		"role:create", "role:edit", "role:delete",
		"perm:create", "perm:edit", "perm:delete",
		"menu:create", "menu:edit", "menu:delete",
		"dict:create", "dict:edit", "dict:delete",
		"param:create", "param:edit", "param:delete",
		"tenant:status", "tenant:reset_primary_password",
		"audit:view", "login:view", "brand:edit",
		"monhealth:view", "monserver:view", "monjobs:view", "monservices:view", "moncache:view", "moncachekeys:view",
	} {
		items = append(items, seedPermission{Name: path, Path: path, Type: 2, FeatureType: "OPERATION", DataPermMode: "ORG"})
	}
	for path, prefix := range map[string]string{"/tenants": "tenant", "/plans": "plan", "/organization": "org", "/positions": "pos", "/business-units": "business_unit", "/users": "user", "/roles": "role", "/menus": "menu", "/dict": "dict", "/params": "param", "/audit-logs": "audit", "/login-logs": "login", "/monitor/health": "monhealth", "/monitor/server": "monserver", "/monitor/jobs": "monjobs", "/monitor/services": "monservices", "/monitor/cache": "moncache", "/monitor/cache-keys": "moncachekeys"} {
		platformOnly := strings.HasPrefix(path, "/monitor/") || path == "/tenants" || path == "/plans"
		items = append(items, seedPermission{Name: path + "-数据范围", Path: "data:" + prefix, Type: 4, PlatformOnly: platformOnly, PackageFeature: false, FeatureType: "DATA", DataPermMode: "ORG"})
	}
	items = append(items,
		seedPermission{Name: "首页-数据范围", Path: "data:home", Type: 4, Hidden: true, PackageFeature: false, FeatureType: "DATA", DataPermMode: "NONE"},
		seedPermission{Name: "权限管理-数据范围", Path: "data:perm", Type: 4, Hidden: true, PackageFeature: false, FeatureType: "DATA", DataPermMode: "ORG"},
	)

	out := make([]models.Permission, 0, len(items))
	for _, item := range items {
		permission := models.Permission{
			TenantID:         tenantID,
			Name:             item.Name,
			Path:             item.Path,
			PermType:         item.Type,
			SortOrder:        item.SortOrder,
			Enabled:          true,
			Visible:          !item.Hidden,
			IsPlatformOnly:   item.PlatformOnly,
			IsPackageFeature: item.PackageFeature || item.FeatureCode != "",
			TenantEditable:   item.TenantEditable,
			DataPermMode:     item.DataPermMode,
		}
		if permission.DataPermMode == "" {
			permission.DataPermMode = "ORG"
		}
		if item.TenantEditScope != "" {
			permission.TenantEditScope = &item.TenantEditScope
		}
		if item.FeatureCode != "" {
			permission.FeatureCode = &item.FeatureCode
		}
		if item.FeatureType != "" {
			permission.FeatureType = &item.FeatureType
		}
		if err := db.Where("tenant_id = ? AND path = ?", tenantID, item.Path).FirstOrCreate(&permission).Error; err != nil {
			return nil, err
		}
		var featureCode interface{}
		if item.FeatureCode != "" {
			featureCode = item.FeatureCode
		}
		var featureType interface{}
		if item.FeatureType != "" {
			featureType = item.FeatureType
		}
		var tenantEditScope interface{}
		if item.TenantEditScope != "" {
			tenantEditScope = item.TenantEditScope
		}
		dataPermMode := item.DataPermMode
		if dataPermMode == "" {
			dataPermMode = "ORG"
		}
		if err := db.Model(&permission).Updates(map[string]interface{}{
			"name":               item.Name,
			"perm_type":          item.Type,
			"sort_order":         item.SortOrder,
			"enabled":            true,
			"visible":            !item.Hidden,
			"is_platform_only":   item.PlatformOnly,
			"is_package_feature": item.PackageFeature || item.FeatureCode != "",
			"tenant_editable":    item.TenantEditable,
			"tenant_edit_scope":  tenantEditScope,
			"data_perm_mode":     dataPermMode,
			"feature_code":       featureCode,
			"feature_type":       featureType,
		}).Error; err != nil {
			return nil, err
		}
		out = append(out, permission)
	}
	return out, nil
}

func seedSaasPlans(db *gorm.DB, tenantID uint64) error {
	features := []models.SaasFeature{
		{FeatureCode: "user_manage", FeatureName: "用户管理", FeatureType: "MENU", Status: 1, Description: stringPtr("用户列表、创建、编辑、删除与重置密码等")},
		{FeatureCode: "org_manage", FeatureName: "组织管理", FeatureType: "MENU", Status: 1, Description: stringPtr("公司、部门管理")},
		{FeatureCode: "position_manage", FeatureName: "岗位管理", FeatureType: "MENU", Status: 1, Description: stringPtr("岗位类型与岗位维护")},
		{FeatureCode: "role_manage", FeatureName: "角色权限", FeatureType: "MENU", Status: 1, Description: stringPtr("角色与菜单权限")},
		{FeatureCode: "dict_manage", FeatureName: "数据字典", FeatureType: "MENU", Status: 1, Description: stringPtr("字典类型、字典项与租户覆盖")},
		{FeatureCode: "param_manage", FeatureName: "系统参数", FeatureType: "MENU", Status: 1, Description: stringPtr("系统参数与租户覆盖")},
		{FeatureCode: "business_unit_manage", FeatureName: "业务单元", FeatureType: "MENU", Status: 1, Description: stringPtr("业务单元与组织映射维护")},
		{FeatureCode: "data_permission", FeatureName: "基础数据权限", FeatureType: "SERVICE", Status: 1, Description: stringPtr("组织、部门、本人数据范围")},
		{FeatureCode: "advanced_data_permission", FeatureName: "高级数据权限", FeatureType: "SERVICE", Status: 1, Description: stringPtr("自定义数据范围")},
		{FeatureCode: "login_log", FeatureName: "登录日志", FeatureType: "MENU", Status: 1, Description: stringPtr("登录审计查询")},
		{FeatureCode: "audit_log", FeatureName: "操作审计", FeatureType: "MENU", Status: 1, Description: stringPtr("操作审计查询")},
		{FeatureCode: "import_data", FeatureName: "数据导入", FeatureType: "BUTTON", Status: 1, Description: stringPtr("CSV/Excel 批量导入")},
		{FeatureCode: "export_data", FeatureName: "数据导出", FeatureType: "BUTTON", Status: 1, Description: stringPtr("CSV/Excel 导出")},
		{FeatureCode: "file_manage", FeatureName: "文件管理", FeatureType: "SERVICE", Status: 1, Description: stringPtr("文件上传、下载与存储空间")},
		{FeatureCode: "brand_config", FeatureName: "品牌配置", FeatureType: "CONFIG", Status: 1, Description: stringPtr("Logo、名称和版权配置")},
		{FeatureCode: "system_monitor", FeatureName: "系统监控", FeatureType: "MENU", Status: 1, Description: stringPtr("健康、服务、缓存监控")},
		{FeatureCode: "ip_whitelist", FeatureName: "IP 白名单", FeatureType: "CONFIG", Status: 1, Description: stringPtr("限制登录 IP")},
		{FeatureCode: "mfa", FeatureName: "MFA 双因素认证", FeatureType: "SERVICE", Status: 1, Description: stringPtr("多因素认证能力")},
		{FeatureCode: "sso_login", FeatureName: "SSO 登录", FeatureType: "SERVICE", Status: 1, Description: stringPtr("OIDC/SAML/企业平台 SSO")},
		{FeatureCode: "api_key", FeatureName: "API Key", FeatureType: "API", Status: 1, Description: stringPtr("开放 API 访问密钥")},
		{FeatureCode: "webhook", FeatureName: "Webhook", FeatureType: "SERVICE", Status: 1, Description: stringPtr("事件推送")},
		{FeatureCode: "tenant_data_export", FeatureName: "租户数据导出", FeatureType: "SERVICE", Status: 1, Description: stringPtr("租户级数据导出")},
	}
	for i := range features {
		if err := db.Where("feature_code = ?", features[i].FeatureCode).FirstOrCreate(&features[i]).Error; err != nil {
			return err
		}
	}

	quotas := []models.SaasQuota{
		{QuotaCode: "max_users", QuotaName: "最大用户数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_companies", QuotaName: "最大公司数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_stores", QuotaName: "最大门店数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_departments", QuotaName: "最大部门数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_business_units", QuotaName: "最大业务单元数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_roles", QuotaName: "最大角色数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_storage_gb", QuotaName: "存储空间", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("GB"), Status: 1},
		{QuotaCode: "max_file_size_mb", QuotaName: "单文件大小", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("MB"), Status: 1},
		{QuotaCode: "max_api_keys", QuotaName: "API Key 数量", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_webhooks", QuotaName: "Webhook 数量", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "daily_api_calls", QuotaName: "每日 API 调用量", QuotaType: "DYNAMIC", PeriodType: stringPtr("DAY"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "daily_import_times", QuotaName: "每日导入次数", QuotaType: "DYNAMIC", PeriodType: stringPtr("DAY"), Unit: stringPtr("TIMES"), Status: 1},
		{QuotaCode: "daily_export_times", QuotaName: "每日导出次数", QuotaType: "DYNAMIC", PeriodType: stringPtr("DAY"), Unit: stringPtr("TIMES"), Status: 1},
		{QuotaCode: "monthly_sms_count", QuotaName: "每月短信条数", QuotaType: "DYNAMIC", PeriodType: stringPtr("MONTH"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "monthly_email_count", QuotaName: "每月邮件条数", QuotaType: "DYNAMIC", PeriodType: stringPtr("MONTH"), Unit: stringPtr("COUNT"), Status: 1},
	}
	for i := range quotas {
		if err := db.Where("quota_code = ?", quotas[i].QuotaCode).FirstOrCreate(&quotas[i]).Error; err != nil {
			return err
		}
	}

	plans := []models.SaasPlan{
		{PlanCode: "TRIAL", PlanName: "试用版", PlanType: "TRIAL", BillingCycle: "CUSTOM", Price: 0, Status: 1, IsDefault: true, SortOrder: 10, Description: stringPtr("新租户试用套餐")},
		{PlanCode: "BASIC", PlanName: "基础版", PlanType: "BASIC", BillingCycle: "MONTH", Price: 0, Status: 1, SortOrder: 20, Description: stringPtr("适合小团队的基础管理套餐")},
		{PlanCode: "PRO", PlanName: "专业版", PlanType: "PRO", BillingCycle: "MONTH", Price: 0, Status: 1, SortOrder: 30, Description: stringPtr("适合中型企业的审计与开放能力套餐")},
		{PlanCode: "ENTERPRISE", PlanName: "企业版", PlanType: "ENTERPRISE", BillingCycle: "YEAR", Price: 0, Status: 1, SortOrder: 40, Description: stringPtr("适合集团客户的高安全与可配置套餐")},
	}
	planFeatures := map[string][]string{
		"TRIAL":      {"user_manage", "org_manage", "position_manage", "role_manage", "dict_manage", "param_manage", "business_unit_manage", "data_permission", "login_log", "brand_config", "file_manage"},
		"BASIC":      {"user_manage", "org_manage", "position_manage", "role_manage", "dict_manage", "param_manage", "business_unit_manage", "data_permission", "login_log", "import_data", "brand_config", "file_manage"},
		"PRO":        {"user_manage", "org_manage", "position_manage", "role_manage", "dict_manage", "param_manage", "business_unit_manage", "data_permission", "advanced_data_permission", "login_log", "audit_log", "import_data", "export_data", "api_key", "webhook", "tenant_data_export", "brand_config", "file_manage"},
		"ENTERPRISE": {"user_manage", "org_manage", "position_manage", "role_manage", "dict_manage", "param_manage", "business_unit_manage", "data_permission", "advanced_data_permission", "login_log", "audit_log", "import_data", "export_data", "brand_config", "ip_whitelist", "mfa", "sso_login", "api_key", "webhook", "tenant_data_export", "file_manage"},
	}
	planQuotas := map[string]map[string]int{
		"TRIAL":      {"max_users": 5, "max_companies": 1, "max_stores": 1, "max_departments": 10, "max_business_units": 10, "max_roles": 5, "max_storage_gb": 1, "max_file_size_mb": 10, "max_api_keys": 0, "max_webhooks": 0, "daily_api_calls": 1000, "daily_import_times": 1, "daily_export_times": 0, "monthly_sms_count": 0, "monthly_email_count": 100},
		"BASIC":      {"max_users": 20, "max_companies": 1, "max_stores": 1, "max_departments": 50, "max_business_units": 50, "max_roles": 10, "max_storage_gb": 5, "max_file_size_mb": 50, "max_api_keys": 0, "max_webhooks": 0, "daily_api_calls": 5000, "daily_import_times": 5, "daily_export_times": 3, "monthly_sms_count": 100, "monthly_email_count": 1000},
		"PRO":        {"max_users": 100, "max_companies": 10, "max_stores": 10, "max_departments": 500, "max_business_units": 500, "max_roles": 50, "max_storage_gb": 50, "max_file_size_mb": 100, "max_api_keys": 5, "max_webhooks": 5, "daily_api_calls": 50000, "daily_import_times": 50, "daily_export_times": 50, "monthly_sms_count": 1000, "monthly_email_count": 10000},
		"ENTERPRISE": {"max_users": 1000, "max_companies": 100, "max_stores": 100, "max_departments": 5000, "max_business_units": 5000, "max_roles": 200, "max_storage_gb": 500, "max_file_size_mb": 500, "max_api_keys": 50, "max_webhooks": 50, "daily_api_calls": 500000, "daily_import_times": 500, "daily_export_times": 500, "monthly_sms_count": 10000, "monthly_email_count": 100000},
	}
	for i := range plans {
		if err := db.Where("plan_code = ?", plans[i].PlanCode).FirstOrCreate(&plans[i]).Error; err != nil {
			return err
		}
		enabledFeatures := stringSet(planFeatures[plans[i].PlanCode])
		for _, feature := range features {
			link := models.SaasPlanFeature{PlanID: plans[i].ID, FeatureID: feature.ID, Enabled: enabledFeatures[feature.FeatureCode]}
			if err := db.Where("plan_id = ? AND feature_id = ?", link.PlanID, link.FeatureID).FirstOrCreate(&link).Error; err != nil {
				return err
			}
			_ = db.Model(&link).Update("enabled", enabledFeatures[feature.FeatureCode]).Error
		}
		for _, quota := range quotas {
			value := planQuotas[plans[i].PlanCode][quota.QuotaCode]
			link := models.SaasPlanQuota{PlanID: plans[i].ID, QuotaID: quota.ID, QuotaValue: value}
			if err := db.Where("plan_id = ? AND quota_id = ?", link.PlanID, link.QuotaID).FirstOrCreate(&link).Error; err != nil {
				return err
			}
			_ = db.Model(&link).Update("quota_value", value).Error
		}
	}

	var defaultPlan models.SaasPlan
	if err := db.Where("plan_code = ?", "ENTERPRISE").First(&defaultPlan).Error; err != nil {
		return err
	}
	subscription := models.TenantSubscription{
		TenantID:           tenantID,
		PlanID:             defaultPlan.ID,
		SubscriptionStatus: "ACTIVE",
		StartTime:          time.Now(),
		AutoRenew:          false,
	}
	return db.Where("tenant_id = ?", tenantID).FirstOrCreate(&subscription).Error
}

func seedDictionaries(db *gorm.DB, tenantID uint64) error {
	dicts := []struct {
		code   string
		name   string
		remark string
		items  []struct{ label, value string }
	}{
		{"common_status", "通用状态", "启用/停用", []struct{ label, value string }{{"启用", "1"}, {"停用", "0"}}},
		{"company_type", "公司类型", "公司类型", []struct{ label, value string }{{"集团公司", "GROUP"}, {"子公司", "SUBSIDIARY"}, {"分公司", "BRANCH"}, {"门店", "STORE"}, {"区域公司", "REGIONAL_COMPANY"}, {"运营公司", "OPERATING_COMPANY"}, {"关联公司", "AFFILIATE"}, {"加盟公司", "FRANCHISEE"}, {"经销商公司", "DEALER"}, {"项目公司", "PROJECT_COMPANY"}, {"其他", "OTHER"}}},
		{"org_node_type", "组织节点类型", "组织节点类型", []struct{ label, value string }{{"集团", "group"}, {"公司", "company"}, {"部门", "department"}, {"门店", "store"}, {"仓库", "warehouse"}, {"项目组", "project_team"}}},
		{"business_unit_type", "业务单元类型", "业务单元类型", []struct{ label, value string }{{"区域", "REGION"}, {"门店", "STORE"}, {"公司", "COMPANY"}, {"项目", "PROJECT"}, {"仓库", "WAREHOUSE"}, {"活动", "CAMPAIGN"}, {"自定义", "CUSTOM"}}},
	}
	for _, item := range dicts {
		dictType := models.DictType{TenantID: tenantID, Code: item.code, Name: item.name, Remark: &item.remark, Scope: "HYBRID", TenantEditable: true}
		if err := db.Where("tenant_id = ? AND code = ?", tenantID, item.code).FirstOrCreate(&dictType).Error; err != nil {
			return err
		}
		for i, option := range item.items {
			dictItem := models.DictItem{TenantID: tenantID, DictTypeID: dictType.ID, Label: option.label, Value: option.value, SortOrder: i + 1, Enabled: true}
			if err := db.Where("dict_type_id = ? AND value = ?", dictType.ID, option.value).FirstOrCreate(&dictItem).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedSystemParams(db *gorm.DB, tenantID uint64) error {
	params := []models.SystemParam{
		{TenantID: tenantID, Key: "org.default_company_type", Value: "SUBSIDIARY", Remark: "新建公司默认类型（字典 company_type 的 value，须一致）", ValueType: "string", TenantEditable: true},
		{TenantID: tenantID, Key: "user.list_default_page_size", Value: "10", Remark: "用户列表默认每页条数", ValueType: "number", TenantEditable: true},
	}
	for _, param := range params {
		err := db.Clauses(clause.OnConflict{DoNothing: true}).Where("tenant_id = ? AND param_key = ?", param.TenantID, param.Key).Create(&param).Error
		if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
	}
	return nil
}

func seedAuditSamples(db *gorm.DB, tenantID uint64, userID uint64) error {
	message := "初始化平台管理员登录记录"
	login := models.LoginLog{TenantID: &tenantID, UserID: &userID, Account: "E10001", Success: true, Message: &message}
	if err := db.Where("account = ? AND message = ?", login.Account, message).FirstOrCreate(&login).Error; err != nil {
		return err
	}
	audit := models.AuditLog{TenantID: &tenantID, UserID: &userID, Module: "bootstrap", Action: "seed", Summary: "初始化 Go DDD 项目基础数据"}
	return db.Where("module = ? AND action = ? AND summary = ?", audit.Module, audit.Action, audit.Summary).FirstOrCreate(&audit).Error
}

func stringPtr(value string) *string {
	return &value
}

func stringSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
