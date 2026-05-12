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

	"saas_baseon_go/internal/domain/permissioncatalog"
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
	if err := seedBuiltinApps(db); err != nil {
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
	AppCode         string
	FeatureCode     string
	FeatureType     string
	DataPermMode    string
}

func seedPermissionAppCode(path string) string {
	path = strings.TrimSpace(path)
	if path == "/apps" || strings.HasPrefix(path, "/apps/") || strings.HasPrefix(path, "app:") || strings.HasPrefix(path, "data:app") {
		return "app-center"
	}
	if strings.HasPrefix(path, "/monitor/") || strings.HasPrefix(path, "mon") || strings.HasPrefix(path, "data:mon") {
		return "system-monitor"
	}
	return "system-management"
}

func seedFeatureAppCode(feature models.SaasFeature) string {
	code := strings.TrimSpace(feature.FeatureCode)
	name := strings.TrimSpace(feature.FeatureName)
	if code == "system_monitor" || strings.HasPrefix(code, "button_mon") || strings.HasPrefix(name, "mon") {
		return "system-monitor"
	}
	return "system-management"
}

func seedPermissions(db *gorm.DB, tenantID uint64) ([]models.Permission, error) {
	items := []seedPermission{
		{Name: "菜单根节点", Path: "__menu_root__", Type: 1, SortOrder: 0, Hidden: true, PackageFeature: false, FeatureType: "SYSTEM", DataPermMode: "NONE"},
		{Name: "操作根节点", Path: "__operations_root__", Type: 1, SortOrder: 0, Hidden: true, PackageFeature: false, FeatureType: "SYSTEM", DataPermMode: "NONE"},
		{Name: "首页", Path: "/home", Type: 3, SortOrder: 0, Hidden: true, PackageFeature: false, FeatureCode: "home", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "应用列表", Path: "/apps", Type: 3, SortOrder: 1, PlatformOnly: true, PackageFeature: false, FeatureCode: "app_list", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "客户端中心", Path: "/apps/clients", Type: 3, SortOrder: 2, PlatformOnly: true, PackageFeature: false, FeatureCode: "app_clients", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "租户开通总览", Path: "/apps/tenant-openings", Type: 3, SortOrder: 3, PlatformOnly: true, PackageFeature: false, FeatureCode: "app_tenant_openings", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "体验邀请总览", Path: "/apps/trial-invites", Type: 3, SortOrder: 4, PlatformOnly: true, PackageFeature: false, FeatureCode: "app_trial_invites", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "Manifest 装载记录", Path: "/apps/manifests", Type: 3, SortOrder: 5, PlatformOnly: true, PackageFeature: false, FeatureCode: "app_manifest_loads", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "应用审计日志", Path: "/apps/audit-logs", Type: 3, SortOrder: 6, PlatformOnly: true, PackageFeature: false, FeatureCode: "app_audit_logs", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "主体管理", Path: "/tenants", Type: 3, SortOrder: 2, PlatformOnly: true, PackageFeature: false, FeatureCode: "tenant_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "套餐中心", Path: "/plans", Type: 3, SortOrder: 3, PlatformOnly: true, PackageFeature: false, FeatureCode: "plan_manage", FeatureType: "MENU", DataPermMode: "NONE"},
		{Name: "组织架构", Path: "/organization", Type: 3, SortOrder: 4, PackageFeature: true, FeatureCode: "org_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "岗位管理", Path: "/positions", Type: 3, SortOrder: 5, PackageFeature: true, FeatureCode: "position_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "业务单元", Path: "/business-units", Type: 3, SortOrder: 6, PackageFeature: true, FeatureCode: "business_unit_manage", FeatureType: "MENU", DataPermMode: "BU"},
		{Name: "用户管理", Path: "/users", Type: 3, SortOrder: 7, PackageFeature: true, FeatureCode: "user_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "角色权限", Path: "/roles", Type: 3, SortOrder: 8, PackageFeature: true, FeatureCode: "role_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "菜单管理", Path: "/menus", Type: 3, SortOrder: 9, PackageFeature: true, FeatureCode: "menu_manage", FeatureType: "MENU", TenantEditable: true, DataPermMode: "NONE"},
		{Name: "权限管理兼容入口", Path: "/permissions", Type: 3, SortOrder: 8, Hidden: true, FeatureCode: "role_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "数据字典", Path: "/dict", Type: 3, SortOrder: 10, PackageFeature: true, FeatureCode: "dict_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "参数管理", Path: "/params", Type: 3, SortOrder: 11, PackageFeature: true, FeatureCode: "param_manage", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "操作日志", Path: "/audit-logs", Type: 3, SortOrder: 12, PackageFeature: true, FeatureCode: "audit_log", FeatureType: "MENU", DataPermMode: "ORG"},
		{Name: "登录日志", Path: "/login-logs", Type: 3, SortOrder: 13, PackageFeature: true, FeatureCode: "login_log", FeatureType: "MENU", DataPermMode: "ORG"},
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
		"app:create", "app:edit", "app:status",
		"org:create", "org:edit", "org:delete",
		"pos:create", "pos:edit", "pos:delete",
		"business_unit:create", "business_unit:edit", "business_unit:delete",
		"user:create", "user:edit", "user:reset_password", "user:delete",
		"role:create", "role:edit", "role:delete", "role:permission",
		"menu:create", "menu:edit", "menu:delete", "menu:package_feature",
		"dict_type:create", "dict_type:edit", "dict_type:delete",
		"dict_item:create", "dict_item:edit", "dict_item:delete",
		"param:create", "param:edit", "param:delete",
		"tenant:status", "tenant:reset_primary_password",
		"brand:edit",
	} {
		platformOnly := permissioncatalog.IsPlatformOnlyOperation(path)
		dataPermMode := "ORG"
		if platformOnly {
			dataPermMode = "NONE"
		}
		featureCode := permissioncatalog.OperationFeatureCode(path)
		featureType := "OPERATION"
		tenantEditable := false
		tenantEditScope := ""
		if path == "brand:edit" {
			featureType = "CONFIG"
			tenantEditable = true
			tenantEditScope = "NAME_ICON"
		}
		items = append(items, seedPermission{
			Name:            permissioncatalog.OperationName(path),
			Path:            path,
			Type:            2,
			PlatformOnly:    platformOnly,
			PackageFeature:  permissioncatalog.IsPackageFeatureOperation(path),
			TenantEditable:  tenantEditable,
			TenantEditScope: tenantEditScope,
			AppCode:         seedPermissionAppCode(path),
			FeatureCode:     featureCode,
			FeatureType:     featureType,
			DataPermMode:    dataPermMode,
		})
	}
	for path, prefix := range map[string]string{"/apps": "app", "/tenants": "tenant", "/plans": "plan", "/organization": "org", "/positions": "pos", "/business-units": "business_unit", "/users": "user", "/roles": "role", "/menus": "menu", "/dict": "dict", "/params": "param", "/audit-logs": "audit", "/login-logs": "login", "/monitor/health": "monhealth", "/monitor/server": "monserver", "/monitor/jobs": "monjobs", "/monitor/services": "monservices", "/monitor/cache": "moncache", "/monitor/cache-keys": "moncachekeys"} {
		platformOnly := strings.HasPrefix(path, "/monitor/") || path == "/tenants" || path == "/plans"
		if path == "/apps" {
			platformOnly = true
		}
		dataPermMode := "ORG"
		if path == "/business-units" {
			dataPermMode = "BU"
		}
		if platformOnly {
			dataPermMode = "NONE"
		}
		items = append(items, seedPermission{Name: path + "-数据范围", Path: "data:" + prefix, Type: 4, PlatformOnly: platformOnly, PackageFeature: false, AppCode: seedPermissionAppCode(path), FeatureType: "DATA", DataPermMode: dataPermMode})
	}
	items = append(items,
		seedPermission{Name: "首页-数据范围", Path: "data:home", Type: 4, Hidden: true, PackageFeature: false, FeatureType: "DATA", DataPermMode: "NONE"},
		seedPermission{Name: "权限管理-数据范围", Path: "data:perm", Type: 4, Hidden: true, PackageFeature: false, FeatureType: "DATA", DataPermMode: "ORG"},
	)

	out := make([]models.Permission, 0, len(items))
	for _, item := range items {
		tenantEditable := item.TenantEditable
		if item.Type == 3 && !item.PlatformOnly {
			tenantEditable = true
		}
		appCode := item.AppCode
		if appCode == "" {
			appCode = seedPermissionAppCode(item.Path)
		}
		permission := models.Permission{
			TenantID:         tenantID,
			Name:             item.Name,
			Path:             item.Path,
			PermType:         item.Type,
			AppCode:          appCode,
			SortOrder:        item.SortOrder,
			Enabled:          true,
			Visible:          !item.Hidden,
			IsPlatformOnly:   item.PlatformOnly,
			IsPackageFeature: item.PackageFeature,
			TenantEditable:   tenantEditable,
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
			"is_package_feature": item.PackageFeature,
			"tenant_editable":    tenantEditable,
			"tenant_edit_scope":  tenantEditScope,
			"app_code":           appCode,
			"data_perm_mode":     dataPermMode,
			"feature_code":       featureCode,
			"feature_type":       featureType,
		}).Error; err != nil {
			return nil, err
		}
		out = append(out, permission)
	}
	if err := seedPermissionParentIDs(db, tenantID); err != nil {
		return nil, err
	}
	return out, nil
}

func seedPermissionParentIDs(db *gorm.DB, tenantID uint64) error {
	pairs := map[string]string{
		"/apps/clients":         "/apps",
		"/apps/tenant-openings": "/apps",
		"/apps/trial-invites":   "/apps",
		"/apps/manifests":       "/apps",
		"/apps/audit-logs":      "/apps",
	}
	for childPath, parentPath := range pairs {
		var parent models.Permission
		if err := db.Where("tenant_id = ? AND path = ? AND deleted_at IS NULL", tenantID, parentPath).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if err := db.Model(&models.Permission{}).
			Where("tenant_id = ? AND path = ? AND deleted_at IS NULL", tenantID, childPath).
			Update("parent_id", parent.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSaasPlans(db *gorm.DB, tenantID uint64) error {
	features := []models.SaasFeature{
		{FeatureCode: "user_manage", FeatureName: "用户管理", FeatureType: "MENU", Status: 1, Description: stringPtr("用户列表、创建、编辑、删除与重置密码等")},
		{FeatureCode: "org_manage", FeatureName: "组织管理", FeatureType: "MENU", Status: 1, Description: stringPtr("公司、部门管理")},
		{FeatureCode: "position_manage", FeatureName: "岗位管理", FeatureType: "MENU", Status: 1, Description: stringPtr("岗位类型与岗位维护")},
		{FeatureCode: "role_manage", FeatureName: "角色权限", FeatureType: "MENU", Status: 1, Description: stringPtr("角色与菜单权限")},
		{FeatureCode: "menu_manage", FeatureName: "菜单管理", FeatureType: "MENU", Status: 1, Description: stringPtr("租户菜单显示名称、图标与侧栏可见性覆盖")},
		{FeatureCode: "button_menu_edit", FeatureName: "菜单-编辑", FeatureType: "OPERATION", Status: 1, Description: stringPtr("允许租户编辑套餐内菜单的显示名称和图标")},
		{FeatureCode: "dict_manage", FeatureName: "数据字典", FeatureType: "MENU", Status: 1, Description: stringPtr("字典类型、字典项与租户覆盖")},
		{FeatureCode: "button_dict_item_create", FeatureName: "字典项-新增", FeatureType: "OPERATION", Status: 1, Description: stringPtr("允许租户新增套餐内开放字典的字典项")},
		{FeatureCode: "button_dict_item_edit", FeatureName: "字典项-编辑", FeatureType: "OPERATION", Status: 1, Description: stringPtr("允许租户覆盖套餐内开放字典的字典项")},
		{FeatureCode: "button_dict_item_delete", FeatureName: "字典项-删除", FeatureType: "OPERATION", Status: 1, Description: stringPtr("允许租户删除自有字典项")},
		{FeatureCode: "param_manage", FeatureName: "系统参数", FeatureType: "MENU", Status: 1, Description: stringPtr("系统参数与租户覆盖")},
		{FeatureCode: "business_unit_manage", FeatureName: "业务单元", FeatureType: "MENU", Status: 1, Description: stringPtr("业务单元与组织映射维护")},
		{FeatureCode: "login_log", FeatureName: "登录日志", FeatureType: "MENU", Status: 1, Description: stringPtr("登录审计查询")},
		{FeatureCode: "audit_log", FeatureName: "操作审计", FeatureType: "MENU", Status: 1, Description: stringPtr("操作审计查询")},
		{FeatureCode: "import_data", FeatureName: "数据导入", FeatureType: "BUTTON", Status: 1, Description: stringPtr("CSV/Excel 批量导入")},
		{FeatureCode: "export_data", FeatureName: "数据导出", FeatureType: "BUTTON", Status: 1, Description: stringPtr("CSV/Excel 导出")},
		{FeatureCode: "file_manage", FeatureName: "文件管理", FeatureType: "SERVICE", Status: 1, Description: stringPtr("文件上传、下载与存储空间")},
		{FeatureCode: "brand_config", FeatureName: "品牌配置", FeatureType: "CONFIG", Status: 1, Description: stringPtr("Logo、名称和版权配置")},
		{FeatureCode: "system_monitor", FeatureName: "系统监控", FeatureType: "MENU", Status: 1, Description: stringPtr("健康、服务、缓存监控")},
	}
	for i := range features {
		if features[i].AppCode == "" {
			features[i].AppCode = seedFeatureAppCode(features[i])
		}
		if err := db.Where("feature_code = ?", features[i].FeatureCode).FirstOrCreate(&features[i]).Error; err != nil {
			return err
		}
		if err := db.Model(&features[i]).Updates(map[string]interface{}{
			"feature_name": features[i].FeatureName,
			"feature_type": features[i].FeatureType,
			"app_code":     features[i].AppCode,
			"status":       features[i].Status,
			"description":  features[i].Description,
		}).Error; err != nil {
			return err
		}
	}
	if err := seedPackageFeatureParentIDs(db); err != nil {
		return err
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
		{QuotaCode: "daily_import_times", QuotaName: "每日导入次数", QuotaType: "DYNAMIC", PeriodType: stringPtr("DAY"), Unit: stringPtr("TIMES"), Status: 1},
		{QuotaCode: "daily_export_times", QuotaName: "每日导出次数", QuotaType: "DYNAMIC", PeriodType: stringPtr("DAY"), Unit: stringPtr("TIMES"), Status: 1},
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
		"TRIAL":      {"user_manage", "org_manage", "position_manage", "role_manage", "menu_manage", "button_menu_edit", "dict_manage", "button_dict_item_create", "button_dict_item_edit", "button_dict_item_delete", "param_manage", "business_unit_manage", "login_log", "brand_config", "file_manage"},
		"BASIC":      {"user_manage", "org_manage", "position_manage", "role_manage", "menu_manage", "button_menu_edit", "dict_manage", "button_dict_item_create", "button_dict_item_edit", "button_dict_item_delete", "param_manage", "business_unit_manage", "login_log", "import_data", "brand_config", "file_manage"},
		"PRO":        {"user_manage", "org_manage", "position_manage", "role_manage", "menu_manage", "button_menu_edit", "dict_manage", "button_dict_item_create", "button_dict_item_edit", "button_dict_item_delete", "param_manage", "business_unit_manage", "login_log", "audit_log", "import_data", "export_data", "brand_config", "file_manage"},
		"ENTERPRISE": {"user_manage", "org_manage", "position_manage", "role_manage", "menu_manage", "button_menu_edit", "dict_manage", "button_dict_item_create", "button_dict_item_edit", "button_dict_item_delete", "param_manage", "business_unit_manage", "login_log", "audit_log", "import_data", "export_data", "brand_config", "file_manage"},
	}
	planQuotas := map[string]map[string]int{
		"TRIAL":      {"max_users": 5, "max_companies": 1, "max_stores": 1, "max_departments": 10, "max_business_units": 10, "max_roles": 5, "max_storage_gb": 1, "max_file_size_mb": 10, "daily_import_times": 1, "daily_export_times": 0},
		"BASIC":      {"max_users": 20, "max_companies": 1, "max_stores": 1, "max_departments": 50, "max_business_units": 50, "max_roles": 10, "max_storage_gb": 5, "max_file_size_mb": 50, "daily_import_times": 5, "daily_export_times": 3},
		"PRO":        {"max_users": 100, "max_companies": 10, "max_stores": 10, "max_departments": 500, "max_business_units": 500, "max_roles": 50, "max_storage_gb": 50, "max_file_size_mb": 100, "daily_import_times": 50, "daily_export_times": 50},
		"ENTERPRISE": {"max_users": 1000, "max_companies": 100, "max_stores": 100, "max_departments": 5000, "max_business_units": 5000, "max_roles": 200, "max_storage_gb": 500, "max_file_size_mb": 500, "daily_import_times": 500, "daily_export_times": 500},
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

func seedPackageFeatureParentIDs(db *gorm.DB) error {
	var features []models.SaasFeature
	if err := db.Find(&features).Error; err != nil {
		return err
	}
	byCode := make(map[string]models.SaasFeature, len(features))
	for _, feature := range features {
		byCode[feature.FeatureCode] = feature
	}
	for _, feature := range features {
		if !seedFeatureIsOperation(feature) {
			continue
		}
		parentCode := seedOperationParentFeatureCode(feature)
		parent, ok := byCode[parentCode]
		if parentCode == "" || !ok || parent.ID == 0 || parent.ID == feature.ID || feature.ParentID == parent.ID {
			continue
		}
		if err := db.Model(&models.SaasFeature{}).Where("id = ?", feature.ID).Update("parent_id", parent.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedFeatureIsOperation(feature models.SaasFeature) bool {
	featureType := strings.ToUpper(strings.TrimSpace(feature.FeatureType))
	return featureType == "BUTTON" || featureType == "OPERATION" || strings.HasPrefix(strings.TrimSpace(feature.FeatureCode), "button_") || strings.Contains(strings.TrimSpace(feature.FeatureName), ":")
}

func seedOperationParentFeatureCode(feature models.SaasFeature) string {
	value := strings.TrimSpace(feature.FeatureName)
	if strings.Contains(value, ":") {
		return permissioncatalog.ParentFeatureCodeForOperation(value)
	}
	code := strings.TrimSpace(feature.FeatureCode)
	if strings.HasPrefix(code, "button_") {
		return permissioncatalog.ParentFeatureCodeForOperation(strings.TrimPrefix(code, "button_"))
	}
	return ""
}

func seedDictionaries(db *gorm.DB, tenantID uint64) error {
	dicts := []struct {
		code           string
		name           string
		remark         string
		scope          string
		tenantEditable bool
		platformOnly   bool
		items          []struct{ label, value string }
	}{
		{"common_status", "通用状态", "启用/停用", "platform", true, false, []struct{ label, value string }{{"启用", "1"}, {"停用", "0"}}},
		{"company_type", "公司类型", "公司类型", "HYBRID", true, false, []struct{ label, value string }{{"集团公司", "GROUP"}, {"子公司", "SUBSIDIARY"}, {"分公司", "BRANCH"}, {"门店", "STORE"}, {"区域公司", "REGIONAL_COMPANY"}, {"运营公司", "OPERATING_COMPANY"}, {"关联公司", "AFFILIATE"}, {"加盟公司", "FRANCHISEE"}, {"经销商公司", "DEALER"}, {"项目公司", "PROJECT_COMPANY"}, {"其他", "OTHER"}}},
		{"org_node_type", "组织节点类型", "组织节点类型", "platform", true, false, []struct{ label, value string }{{"集团", "group"}, {"公司", "company"}, {"部门", "department"}, {"门店", "store"}, {"仓库", "warehouse"}, {"项目组", "project_team"}}},
		{"business_unit_type", "业务单元类型", "业务单元类型", "platform", true, false, []struct{ label, value string }{{"区域", "REGION"}, {"门店", "STORE"}, {"公司", "COMPANY"}, {"项目", "PROJECT"}, {"仓库", "WAREHOUSE"}, {"活动", "CAMPAIGN"}, {"自定义", "CUSTOM"}}},
		{"quota_unit", "配额单位", "套餐配额值的展示单位", "platform", false, true, []struct{ label, value string }{{"数量", "COUNT"}, {"MB", "MB"}, {"GB", "GB"}, {"次", "TIMES"}, {"个", "ITEM"}}},
		{"app_type", "应用类型", "应用中心的应用分类", "platform", false, true, []struct{ label, value string }{{"系统底座", "SYSTEM_APP"}, {"业务系统", "BUSINESS_APP"}, {"业务中台", "ABILITY_APP"}, {"API 应用", "API_APP"}, {"连接器", "CONNECTOR_APP"}, {"AI / Agent", "AI_APP"}, {"组合套件", "SUITE_APP"}}},
		{"app_status", "应用状态", "应用中心的生命周期状态", "platform", false, true, []struct{ label, value string }{{"立项", "INITIATED"}, {"规划中", "PLANNED"}, {"开发中", "DEVELOPING"}, {"Beta", "BETA"}, {"已上线", "ONLINE"}, {"已停用", "DISABLED"}, {"已归档", "ARCHIVED"}}},
		{"app_source", "应用来源", "应用注册来源", "platform", false, true, []struct{ label, value string }{{"系统内置", "BUILTIN"}, {"手工创建", "MANUAL"}, {"声明文件装载", "MANIFEST"}}},
		{"app_charge_mode", "应用计费模式", "应用商业化计费模式", "platform", false, true, []struct{ label, value string }{{"免费", "FREE"}, {"订阅制", "SUBSCRIPTION"}, {"按量收费", "USAGE_BASED"}, {"组合收费", "MIXED"}, {"非售卖", "NON_SELLABLE"}}},
		{"app_visibility_scope", "应用可见范围", "应用中心的可见与装载范围", "platform", false, true, []struct{ label, value string }{{"仅平台", "PLATFORM_ONLY"}, {"租户可用", "TENANT"}, {"全局可见", "GLOBAL"}}},
	}
	for _, item := range dicts {
		dictType := models.DictType{TenantID: tenantID, Code: item.code, Name: item.name, Remark: &item.remark, Scope: item.scope, TenantEditable: item.tenantEditable, IsPlatformOnly: item.platformOnly}
		if err := db.Where("tenant_id = ? AND code = ?", tenantID, item.code).FirstOrCreate(&dictType).Error; err != nil {
			return err
		}
		if err := db.Model(&dictType).Updates(map[string]interface{}{
			"name":             item.name,
			"remark":           item.remark,
			"scope":            item.scope,
			"tenant_editable":  item.tenantEditable,
			"is_platform_only": item.platformOnly,
			"deleted_at":       nil,
		}).Error; err != nil {
			return err
		}
		for i, option := range item.items {
			dictItem := models.DictItem{TenantID: tenantID, DictTypeID: dictType.ID, Label: option.label, Value: option.value, SortOrder: i + 1, Enabled: true}
			if err := db.Where("dict_type_id = ? AND value = ?", dictType.ID, option.value).FirstOrCreate(&dictItem).Error; err != nil {
				return err
			}
			if err := db.Model(&dictItem).Updates(map[string]interface{}{
				"label":      option.label,
				"sort_order": i + 1,
				"enabled":    true,
				"deleted_at": nil,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedBuiltinApps(db *gorm.DB) error {
	apps := []models.SysApp{
		{
			AppCode:         "app-center",
			AppName:         "应用中心",
			Icon:            stringPtr("Boxes"),
			AppType:         "SYSTEM_APP",
			Source:          "BUILTIN",
			Status:          "ONLINE",
			ChargeMode:      "NON_SELLABLE",
			VisibilityScope: "PLATFORM_ONLY",
			DeploymentMode:  "MERGED",
			Owner:           stringPtr("平台架构组"),
			Version:         stringPtr("0.1.0"),
			Description:     stringPtr("应用注册、装载规范与平台内置应用治理入口"),
			IsBuiltin:       true,
			IsPlatformOnly:  true,
			SortOrder:       1,
		},
		{
			AppCode:         "system-management",
			AppName:         "系统管理",
			Icon:            stringPtr("Settings"),
			AppType:         "SYSTEM_APP",
			Source:          "BUILTIN",
			Status:          "ONLINE",
			ChargeMode:      "NON_SELLABLE",
			VisibilityScope: "PLATFORM_ONLY",
			DeploymentMode:  "MERGED",
			Owner:           stringPtr("平台架构组"),
			Version:         stringPtr("0.1.0"),
			Description:     stringPtr("用户、角色、菜单、字典、参数等基础治理能力"),
			IsBuiltin:       true,
			IsPlatformOnly:  true,
			SortOrder:       10,
		},
		{
			AppCode:         "system-monitor",
			AppName:         "系统监控",
			Icon:            stringPtr("MonitorCog"),
			AppType:         "SYSTEM_APP",
			Source:          "BUILTIN",
			Status:          "ONLINE",
			ChargeMode:      "NON_SELLABLE",
			VisibilityScope: "PLATFORM_ONLY",
			DeploymentMode:  "MERGED",
			Owner:           stringPtr("平台运维组"),
			Version:         stringPtr("0.1.0"),
			Description:     stringPtr("健康检查、服务状态、缓存和定时任务等运维监控能力"),
			IsBuiltin:       true,
			IsPlatformOnly:  true,
			SortOrder:       20,
		},
		{
			AppCode:         "workbench",
			AppName:         "工作台",
			Icon:            stringPtr("House"),
			AppType:         "SYSTEM_APP",
			Source:          "BUILTIN",
			Status:          "PLANNED",
			ChargeMode:      "NON_SELLABLE",
			VisibilityScope: "PLATFORM_ONLY",
			DeploymentMode:  "MERGED",
			Owner:           stringPtr("平台产品组"),
			Version:         stringPtr("0.1.0"),
			Description:     stringPtr("平台与租户用户的统一工作入口、待办与概览能力"),
			IsBuiltin:       true,
			IsPlatformOnly:  true,
			SortOrder:       30,
		},
		{
			AppCode:         "integration-center",
			AppName:         "第三方集成中心",
			Icon:            stringPtr("Connection"),
			AppType:         "CONNECTOR_APP",
			Source:          "BUILTIN",
			Status:          "PLANNED",
			ChargeMode:      "SUBSCRIPTION",
			VisibilityScope: "TENANT",
			DeploymentMode:  "MERGED",
			Owner:           stringPtr("平台集成组"),
			Version:         stringPtr("0.1.0"),
			Description:     stringPtr("第三方系统、开放 API、Webhook、OAuth 与外部连接器的统一接入中心"),
			IsBuiltin:       true,
			IsPlatformOnly:  false,
			SortOrder:       40,
		},
		{
			AppCode:         "data-center",
			AppName:         "数据中心",
			Icon:            stringPtr("DataAnalysis"),
			AppType:         "ABILITY_APP",
			Source:          "BUILTIN",
			Status:          "PLANNED",
			ChargeMode:      "SUBSCRIPTION",
			VisibilityScope: "TENANT",
			DeploymentMode:  "MERGED",
			Owner:           stringPtr("数据产品组"),
			Version:         stringPtr("0.1.0"),
			Description:     stringPtr("跨应用数据资产、指标、报表、经营预警与数据看板能力中心"),
			IsBuiltin:       true,
			IsPlatformOnly:  false,
			SortOrder:       50,
		},
	}
	for i := range apps {
		row := apps[i]
		if err := db.Where("app_code = ?", row.AppCode).FirstOrCreate(&row).Error; err != nil {
			return err
		}
		if err := db.Model(&row).Updates(map[string]interface{}{
			"app_name":            apps[i].AppName,
			"icon":                apps[i].Icon,
			"app_type":            apps[i].AppType,
			"source":              apps[i].Source,
			"status":              apps[i].Status,
			"charge_mode":         apps[i].ChargeMode,
			"visibility_scope":    apps[i].VisibilityScope,
			"deployment_mode":     apps[i].DeploymentMode,
			"communication_modes": apps[i].CommModes,
			"owner":               apps[i].Owner,
			"version":             apps[i].Version,
			"description":         apps[i].Description,
			"is_builtin":          apps[i].IsBuiltin,
			"is_platform_only":    apps[i].IsPlatformOnly,
			"sort_order":          apps[i].SortOrder,
			"deleted_at":          nil,
		}).Error; err != nil {
			return err
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
