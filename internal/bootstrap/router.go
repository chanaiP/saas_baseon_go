package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"saas_baseon_go/internal/application/system"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/handlers"
	"saas_baseon_go/internal/interfaces/http/middleware"
)

func NewRouter(cfg Config, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.RequestID())
	router.NoRoute(handlers.NewFallbackHandler().NoRoute)

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	identityHandler := handlers.NewIdentityHandler(db, redisClient, cfg.AuthSecret, cfg.TokenTTLHours)

	paramRepo := repositories.NewSystemParamRepository(db)
	paramService := system.NewParamService(paramRepo)
	paramHandler := handlers.NewParamHandler(paramService)

	router.GET("/health", healthHandler.Check)
	router.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(200, openAPISpec())
	})

	api := router.Group("/api")
	{
		api.GET("/auth/captcha", identityHandler.Captcha)
		api.GET("/auth/phone-login-tenants", identityHandler.PhoneLoginTenants)
		api.POST("/auth/login", identityHandler.Login)
		api.GET("/public/tenant-footer", identityHandler.PublicTenantFooter)
		api.GET("/public", identityHandler.PublicTenantFooter)

		api.Use(identityHandler.AuthRequired())
		api.POST("/auth/logout", identityHandler.Logout)
		api.GET("/auth/switchable-tenants", identityHandler.SwitchableTenants)
		api.POST("/auth/switch-tenant", identityHandler.SwitchTenant)
		api.GET("/users/me", identityHandler.Profile)
		api.GET("/users", identityHandler.Users)
		api.GET("/users/assignable-roles", identityHandler.AssignableRoles)
		api.PUT("/users/me", identityHandler.UpdateProfile)
		api.PUT("/users/me/password", identityHandler.UpdatePassword)
		api.GET("/users/me/preferences", identityHandler.Preferences)
		api.PUT("/users/me/preferences", identityHandler.SavePreferences)
		api.POST("/users", identityHandler.CreateUser)
		api.PUT("/users/:id", identityHandler.UpdateUser)
		api.PUT("/users/:id/password", identityHandler.ResetUserPassword)
		api.DELETE("/users/:id", identityHandler.DeleteUser)
		api.GET("/tenant/branding", identityHandler.TenantBranding)
		api.PUT("/tenant/branding", identityHandler.SaveTenantBranding)
		api.GET("/roles/permission-menu-bundles", identityHandler.MenuBundles)
		api.GET("/permission-menu-bundles", identityHandler.MenuBundles)
		api.GET("/permissions", identityHandler.Permissions)
		api.POST("/permissions", identityHandler.CreatePermission)
		api.GET("/permissions/tree", identityHandler.PermissionTree)
		api.GET("/permissions/menu-bundles", identityHandler.MenuBundles)
		api.GET("/permissions/menu-overrides", identityHandler.MenuOverrides)
		api.PUT("/permissions/menu-overrides", identityHandler.SaveMenuOverrides)
		api.GET("/permissions/:id", identityHandler.Permission)
		api.PUT("/permissions/:id", identityHandler.UpdatePermission)
		api.DELETE("/permissions/:id", identityHandler.DeletePermission)
		api.GET("/tenants", identityHandler.Tenants)
		api.POST("/tenants", identityHandler.CreateTenant)
		api.POST("/tenants/with-package", identityHandler.CreateTenantWithPackage)
		api.PUT("/tenants/:id", identityHandler.UpdateTenant)
		api.PATCH("/tenants/:id/status", identityHandler.UpdateTenantStatus)
		api.PUT("/tenants/:id/package-config", identityHandler.SaveTenantPackageConfig)
		api.DELETE("/tenants/:id", identityHandler.DeleteTenant)
		api.GET("/tenants/:id/companies", identityHandler.TenantCompanies)
		api.GET("/tenants/:id/quota-records", identityHandler.TenantQuotaRecords)
		api.GET("/tenants/:id/primary-admin", identityHandler.TenantPrimaryAdmin)
		api.PUT("/tenants/:id/primary-admin/password", identityHandler.ResetTenantPrimaryAdminPassword)
		api.GET("/tenants/:id/subscription", identityHandler.TenantSubscription)
		api.PUT("/tenants/:id/subscription", identityHandler.SaveTenantSubscription)
		api.GET("/tenants/:id/feature-overrides", identityHandler.TenantFeatureOverrides)
		api.PUT("/tenants/:id/feature-overrides", identityHandler.SaveTenantFeatureOverrides)
		api.GET("/tenants/:id/quota-overrides", identityHandler.TenantQuotaOverrides)
		api.PUT("/tenants/:id/quota-overrides", identityHandler.SaveTenantQuotaOverrides)
		api.GET("/tenants/:id/quota-usage", identityHandler.TenantQuotaUsage)
		api.GET("/tenants/:id/feature-access/:feature_code", identityHandler.TenantFeatureAccess)
		api.GET("/tenants/:id/quota-check/:quota_code", identityHandler.TenantQuotaCheck)
		api.GET("/tenants/:id", identityHandler.Tenant)
		api.GET("/roles", identityHandler.Roles)
		api.POST("/roles", identityHandler.CreateRole)
		api.PUT("/roles/:id", identityHandler.UpdateRole)
		api.DELETE("/roles/:id", identityHandler.DeleteRole)
		api.GET("/roles/:id", identityHandler.Role)
		api.PUT("/permissions/menu-data-perm-mode/:id", identityHandler.UpdatePermissionDataPermMode)
		api.PUT("/permissions/menu-data-perm-mode", identityHandler.UpdatePermissionDataPermMode)
		api.GET("/plans", identityHandler.Plans)
		api.POST("/plans", identityHandler.CreatePlan)
		api.PUT("/plans/:id", identityHandler.UpdatePlan)
		api.POST("/plans/:id/copy", identityHandler.CopyPlan)
		api.DELETE("/plans/:id", identityHandler.DeletePlan)
		api.GET("/plans/matrix", identityHandler.PlanMatrix)
		api.GET("/plans/features", identityHandler.Features)
		api.POST("/plans/features", identityHandler.CreateFeature)
		api.PUT("/plans/features/:id", identityHandler.UpdateFeature)
		api.GET("/plans/:id/features", identityHandler.PlanFeatures)
		api.PUT("/plans/:id/features", identityHandler.SavePlanFeatures)
		api.PUT("/plans/:id/capabilities", identityHandler.SavePlanCapabilities)
		api.GET("/plans/quotas", identityHandler.Quotas)
		api.POST("/plans/quotas", identityHandler.CreateQuota)
		api.PUT("/plans/quotas/:id", identityHandler.UpdateQuota)
		api.GET("/plans/:id/quotas", identityHandler.PlanQuotas)
		api.PUT("/plans/:id/quotas", identityHandler.SavePlanQuotas)
		api.GET("/organizations/tree", identityHandler.OrganizationTree)
		api.GET("/organizations/detail", identityHandler.OrganizationDetail)
		api.POST("/org-nodes", identityHandler.CreateOrgNode)
		api.PUT("/org-nodes/:id", identityHandler.UpdateOrgNode)
		api.DELETE("/org-nodes/:id", identityHandler.DeleteOrgNode)
		api.POST("/companies", identityHandler.CreateCompany)
		api.PUT("/companies/:id", identityHandler.UpdateCompany)
		api.DELETE("/companies/:id", identityHandler.DeleteCompany)
		api.POST("/departments", identityHandler.CreateDepartment)
		api.PUT("/departments/:id", identityHandler.UpdateDepartment)
		api.DELETE("/departments/:id", identityHandler.DeleteDepartment)
		api.POST("/stores", identityHandler.CreateStore)
		api.PUT("/stores/:id", identityHandler.UpdateStore)
		api.DELETE("/stores/:id", identityHandler.DeleteStore)
		api.GET("/position-types", identityHandler.PositionTypes)
		api.POST("/position-types", identityHandler.CreatePositionType)
		api.PUT("/position-types/:id", identityHandler.UpdatePositionType)
		api.DELETE("/position-types/:id", identityHandler.DeletePositionType)
		api.GET("/positions", identityHandler.Positions)
		api.POST("/positions", identityHandler.CreatePosition)
		api.PUT("/positions/:id", identityHandler.UpdatePosition)
		api.DELETE("/positions/:id", identityHandler.DeletePosition)
		api.GET("/business-units", identityHandler.BusinessUnits)
		api.GET("/business-units/tree", identityHandler.BusinessUnitTree)
		api.POST("/business-units", identityHandler.CreateBusinessUnit)
		api.PUT("/business-units/:id", identityHandler.UpdateBusinessUnit)
		api.DELETE("/business-units/:id", identityHandler.DeleteBusinessUnit)
		api.GET("/business-units/:id/org-mappings", identityHandler.BusinessUnitOrgMappings)
		api.GET("/business-units/org-mappings", identityHandler.BusinessUnitOrgMappings)
		api.POST("/business-units/:id/org-mappings", identityHandler.CreateBusinessUnitOrgMapping)
		api.DELETE("/business-units/org-mappings/:id", identityHandler.DeleteBusinessUnitOrgMapping)
		api.GET("/logs/login", identityHandler.LoginLogs)
		api.GET("/logs/audit", identityHandler.AuditLogs)
		api.GET("/monitor/health-detail", identityHandler.MonitorHealthDetail)
		api.GET("/monitor/server-info", identityHandler.MonitorServerInfo)
		api.GET("/monitor/scheduled-jobs", identityHandler.MonitorScheduledJobs)
		api.GET("/monitor/services-overview", identityHandler.MonitorServicesOverview)
		api.GET("/monitor/cache-stats", identityHandler.MonitorCacheStats)
		api.GET("/monitor/cache-keys", identityHandler.MonitorCacheKeys)
		api.POST("/files/upload", identityHandler.UploadFile)
		api.GET("/files/download/:file_id", identityHandler.DownloadFile)
		api.DELETE("/files/:file_id", identityHandler.DeleteFile)
		api.GET("/batch/users/export", identityHandler.ExportUsersCSV)
		api.POST("/batch/users/import", identityHandler.ImportUsersCSV)
		api.GET("/batch/companies/export", identityHandler.ExportCompaniesCSV)
		api.GET("/batch/departments/export", identityHandler.ExportDepartmentsCSV)
		api.GET("/dict-types", identityHandler.DictTypes)
		api.POST("/dict-types", identityHandler.CreateDictType)
		api.PUT("/dict-types/:id", identityHandler.UpdateDictType)
		api.DELETE("/dict-types/:id", identityHandler.DeleteDictType)
		api.GET("/dict-types/by-code/:code/items", identityHandler.DictItemsByCode)
		api.GET("/dict-types/by-code", identityHandler.DictItemsByCode)
		api.GET("/dict-items", identityHandler.DictItems)
		api.POST("/dict-items", identityHandler.CreateDictItem)
		api.PUT("/dict-items/:id", identityHandler.UpdateDictItem)
		api.DELETE("/dict-items/:id", identityHandler.DeleteDictItem)
		api.DELETE("/dict-items/:id/override", identityHandler.RestoreDictItem)
		api.GET("/sys-params", identityHandler.SysParams)
		api.POST("/sys-params", identityHandler.CreateSysParam)
		api.PUT("/sys-params/:id", identityHandler.UpdateSysParam)
		api.DELETE("/sys-params/:id", identityHandler.DeleteSysParam)
		api.DELETE("/sys-params/:id/override", identityHandler.RestoreSysParam)
		api.GET("/sys-params/batch", identityHandler.SysParamBatch)
		api.GET("/params", paramHandler.List)
		api.POST("/params", paramHandler.Create)
		api.GET("/params/:key", paramHandler.GetByKey)
	}

	return router
}

func openAPISpec() gin.H {
	api := func(tag, summary string) gin.H {
		return gin.H{
			"tags":      []string{tag},
			"summary":   summary,
			"responses": gin.H{"200": gin.H{"description": "OK"}},
		}
	}
	return gin.H{
		"openapi": "3.0.0",
		"info":    gin.H{"title": "SaaS Baseon Go API", "version": "0.1.0"},
		"tags": []gin.H{
			{"name": "auth", "description": "认证与登录"},
			{"name": "tenants", "description": "主体管理"},
			{"name": "users", "description": "用户管理"},
			{"name": "roles", "description": "角色权限"},
			{"name": "permissions", "description": "菜单权限"},
			{"name": "organizations", "description": "组织架构"},
			{"name": "positions", "description": "岗位管理"},
			{"name": "dict-param", "description": "数据字典与系统参数"},
			{"name": "logs", "description": "登录日志与操作日志"},
			{"name": "monitor", "description": "系统监控"},
			{"name": "branding", "description": "品牌配置"},
		},
		"paths": gin.H{
			"/api/auth/login":                gin.H{"post": api("auth", "登录")},
			"/api/auth/logout":               gin.H{"post": api("auth", "退出登录")},
			"/api/users/me":                  gin.H{"get": api("users", "当前用户"), "put": api("users", "更新当前用户")},
			"/api/users":                     gin.H{"get": api("users", "用户列表"), "post": api("users", "创建用户")},
			"/api/users/{id}":                gin.H{"put": api("users", "更新用户"), "delete": api("users", "删除用户")},
			"/api/roles":                     gin.H{"get": api("roles", "角色列表"), "post": api("roles", "创建角色")},
			"/api/roles/{id}":                gin.H{"get": api("roles", "角色详情"), "put": api("roles", "更新角色"), "delete": api("roles", "删除角色")},
			"/api/permissions/menu-bundles":  gin.H{"get": api("permissions", "菜单权限包")},
			"/api/tenants":                   gin.H{"get": api("tenants", "主体列表"), "post": api("tenants", "创建主体")},
			"/api/tenants/{id}":              gin.H{"get": api("tenants", "主体详情"), "put": api("tenants", "更新主体"), "delete": api("tenants", "删除主体")},
			"/api/plans":                     gin.H{"get": api("tenants", "套餐列表"), "post": api("tenants", "创建套餐")},
			"/api/plans/matrix":              gin.H{"get": api("tenants", "套餐能力矩阵")},
			"/api/organizations/tree":        gin.H{"get": api("organizations", "组织树")},
			"/api/org-nodes":                 gin.H{"post": api("organizations", "创建组织节点")},
			"/api/position-types":            gin.H{"get": api("positions", "岗位类型列表"), "post": api("positions", "创建岗位类型")},
			"/api/positions":                 gin.H{"get": api("positions", "岗位列表"), "post": api("positions", "创建岗位")},
			"/api/business-units":            gin.H{"get": api("organizations", "业务单元列表"), "post": api("organizations", "创建业务单元")},
			"/api/dict-types":                gin.H{"get": api("dict-param", "字典类型列表"), "post": api("dict-param", "创建字典类型")},
			"/api/dict-items":                gin.H{"get": api("dict-param", "字典项列表"), "post": api("dict-param", "创建字典项")},
			"/api/sys-params":                gin.H{"get": api("dict-param", "系统参数列表"), "post": api("dict-param", "创建系统参数")},
			"/api/logs/login":                gin.H{"get": api("logs", "登录日志")},
			"/api/logs/audit":                gin.H{"get": api("logs", "操作日志")},
			"/api/monitor/health-detail":     gin.H{"get": api("monitor", "健康检查")},
			"/api/monitor/server-info":       gin.H{"get": api("monitor", "服务器信息")},
			"/api/monitor/scheduled-jobs":    gin.H{"get": api("monitor", "定时任务")},
			"/api/monitor/services-overview": gin.H{"get": api("monitor", "服务监控")},
			"/api/monitor/cache-stats":       gin.H{"get": api("monitor", "缓存监控")},
			"/api/monitor/cache-keys":        gin.H{"get": api("monitor", "缓存列表")},
			"/api/tenant/branding":           gin.H{"get": api("branding", "租户品牌"), "put": api("branding", "保存租户品牌")},
			"/api/public/tenant-footer":      gin.H{"get": api("branding", "公开页脚")},
		},
	}
}
