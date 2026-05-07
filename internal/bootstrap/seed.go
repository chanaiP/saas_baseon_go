package bootstrap

import (
	"errors"
	"time"

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

	company, positionType, position, err := seedPlatformStructure(db, tenant.ID)
	if err != nil {
		return err
	}

	user := models.AppUser{
		TenantID:        tenant.ID,
		CompanyID:       &company.ID,
		EmployeeNo:      "admin",
		Account:         "admin",
		PasswordHash:    "dev-password-placeholder",
		Name:            "平台管理员",
		Status:          1,
		IsPlatformAdmin: true,
	}
	if err := db.Where("tenant_id = ? AND account = ?", tenant.ID, user.Account).FirstOrCreate(&user).Error; err != nil {
		return err
	}
	_ = db.Model(&user).Updates(map[string]interface{}{"company_id": company.ID}).Error

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
		{Name: "主体管理", Path: "/tenants", Type: 3, SortOrder: 1, PlatformOnly: true, PackageFeature: false, FeatureCode: "tenant_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "套餐中心", Path: "/plans", Type: 3, SortOrder: 2, PlatformOnly: true, PackageFeature: false, FeatureCode: "plan_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "组织架构", Path: "/organization", Type: 3, SortOrder: 3, FeatureCode: "org_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "岗位管理", Path: "/positions", Type: 3, SortOrder: 4, FeatureCode: "position_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "业务单元", Path: "/business-units", Type: 3, SortOrder: 5, FeatureCode: "business_unit_manage", FeatureType: "MENU", DataPermMode: "BU"},
		{Name: "用户管理", Path: "/users", Type: 3, SortOrder: 6, FeatureCode: "user_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "角色权限", Path: "/roles", Type: 3, SortOrder: 7, FeatureCode: "role_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "菜单管理", Path: "/menus", Type: 3, SortOrder: 8, FeatureCode: "menu_manage", FeatureType: "MENU", DataPermMode: "NONE"},
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
		"org:create", "org:edit", "org:delete",
		"pos:create", "pos:edit", "pos:delete",
		"business_unit:create", "business_unit:edit", "business_unit:delete",
		"user:create", "user:edit", "user:reset_password", "user:delete",
		"role:create", "role:edit", "role:delete",
		"menu:create", "menu:edit", "menu:delete",
		"dict:create", "dict:edit", "dict:delete",
		"param:create", "param:edit", "param:delete",
		"audit:view", "login:view", "brand:edit",
		"monhealth:view", "monserver:view", "monjobs:view", "monservices:view", "moncache:view", "moncachekeys:view",
	} {
		items = append(items, seedPermission{Name: path, Path: path, Type: 2, FeatureType: "OPERATION", DataPermMode: "ORG"})
	}

	out := make([]models.Permission, 0, len(items))
	for _, item := range items {
		permission := models.Permission{
			TenantID:         tenantID,
			Name:             item.Name,
			Path:             item.Path,
			PermType:         item.Type,
			SortOrder:        item.SortOrder,
			Enabled:          true,
			Visible:          true,
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
		{FeatureCode: "brand_config", FeatureName: "品牌配置", FeatureType: "CONFIG", Status: 1, Description: stringPtr("Logo、名称和版权配置")},
		{FeatureCode: "system_monitor", FeatureName: "系统监控", FeatureType: "MENU", Status: 1, Description: stringPtr("健康、服务、缓存监控")},
	}
	for i := range features {
		if err := db.Where("feature_code = ?", features[i].FeatureCode).FirstOrCreate(&features[i]).Error; err != nil {
			return err
		}
	}

	quotas := []models.SaasQuota{
		{QuotaCode: "max_users", QuotaName: "最大用户数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_companies", QuotaName: "最大公司数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_departments", QuotaName: "最大部门数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_business_units", QuotaName: "最大业务单元数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "max_roles", QuotaName: "最大角色数", QuotaType: "STATIC", PeriodType: stringPtr("NONE"), Unit: stringPtr("COUNT"), Status: 1},
		{QuotaCode: "daily_api_calls", QuotaName: "每日 API 调用量", QuotaType: "DYNAMIC", PeriodType: stringPtr("DAY"), Unit: stringPtr("COUNT"), Status: 1},
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
	for i := range plans {
		if err := db.Where("plan_code = ?", plans[i].PlanCode).FirstOrCreate(&plans[i]).Error; err != nil {
			return err
		}
		for _, feature := range features {
			link := models.SaasPlanFeature{PlanID: plans[i].ID, FeatureID: feature.ID, Enabled: true}
			if err := db.Where("plan_id = ? AND feature_id = ?", link.PlanID, link.FeatureID).FirstOrCreate(&link).Error; err != nil {
				return err
			}
		}
		for _, quota := range quotas {
			value := map[string]int{"TRIAL": 5, "BASIC": 50, "PRO": 100, "ENTERPRISE": 1000}[plans[i].PlanCode]
			link := models.SaasPlanQuota{PlanID: plans[i].ID, QuotaID: quota.ID, QuotaValue: value}
			if err := db.Where("plan_id = ? AND quota_id = ?", link.PlanID, link.QuotaID).FirstOrCreate(&link).Error; err != nil {
				return err
			}
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
		{"org_node_type", "组织节点类型", "公司/部门/门店", []struct{ label, value string }{{"公司", "company"}, {"部门", "department"}, {"门店", "store"}}},
		{"business_unit_type", "业务单元类型", "默认业务单元类型", []struct{ label, value string }{{"默认类型", "default"}}},
	}
	for _, item := range dicts {
		dictType := models.DictType{TenantID: tenantID, Code: item.code, Name: item.name, Remark: &item.remark, Scope: "platform", TenantEditable: true}
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
		{TenantID: tenantID, Key: "site.mode", Value: "development", Remark: "运行模式", ValueType: "string", TenantEditable: true},
		{TenantID: tenantID, Key: "security.password_min_length", Value: "8", Remark: "密码最小长度", ValueType: "number", TenantEditable: true},
		{TenantID: tenantID, Key: "login.captcha_after_failures", Value: "3", Remark: "登录失败后验证码阈值", ValueType: "number", TenantEditable: true},
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
	login := models.LoginLog{TenantID: &tenantID, UserID: &userID, Account: "admin", Success: true, Message: &message}
	if err := db.Where("account = ? AND message = ?", login.Account, message).FirstOrCreate(&login).Error; err != nil {
		return err
	}
	audit := models.AuditLog{TenantID: &tenantID, UserID: &userID, Module: "bootstrap", Action: "seed", Summary: "初始化 Go DDD 项目基础数据"}
	return db.Where("module = ? AND action = ? AND summary = ?", audit.Module, audit.Action, audit.Summary).FirstOrCreate(&audit).Error
}

func stringPtr(value string) *string {
	return &value
}
