package bootstrap

import (
	"errors"

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

	user := models.AppUser{
		TenantID:        tenant.ID,
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

	return seedSystemParams(db)
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

func seedSystemParams(db *gorm.DB) error {
	params := []models.SystemParam{
		{Key: "site.mode", Value: "development", Remark: "运行模式"},
		{Key: "security.password_min_length", Value: "8", Remark: "密码最小长度"},
		{Key: "login.captcha_after_failures", Value: "3", Remark: "登录失败后验证码阈值"},
	}
	for _, param := range params {
		err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&param).Error
		if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
	}
	return nil
}
