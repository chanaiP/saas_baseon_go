package bootstrap

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"saas_baseon_go/internal/application/system"
	aicchandlers "saas_baseon_go/internal/apps/ai_capability_center/handlers"
	aiccservices "saas_baseon_go/internal/apps/ai_capability_center/services"
	apphandlers "saas_baseon_go/internal/apps/app_center/handlers"
	apprepos "saas_baseon_go/internal/apps/app_center/repositories"
	appservices "saas_baseon_go/internal/apps/app_center/services"
	ichandlers "saas_baseon_go/internal/apps/integration_center/handlers"
	icrepos "saas_baseon_go/internal/apps/integration_center/repositories"
	icservices "saas_baseon_go/internal/apps/integration_center/services"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/handlers"
	"saas_baseon_go/internal/interfaces/http/middleware"
)

func NewRouter(cfg Config, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	metricsCollector := middleware.NewMetricsCollector()
	router.Use(gin.Recovery(), middleware.RequestID(), metricsCollector.Middleware(), middleware.AccessLog())
	router.Use(securityHeadersMiddleware(cfg.AppEnv))
	router.Use(corsMiddleware(cfg.CORSOrigins))
	router.NoRoute(handlers.NewFallbackHandler().NoRoute)

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	identityHandler := handlers.NewIdentityHandler(db, redisClient, cfg.AuthSecret, cfg.TokenTTLHours, cfg.JWTFallback)

	paramRepo := repositories.NewSystemParamRepository(db)
	paramService := system.NewParamService(paramRepo)
	paramHandler := handlers.NewParamHandler(paramService)
	appRepo := apprepos.NewAppRepository(db)
	appService := appservices.NewAppService(appRepo)
	appHandler := apphandlers.NewAppHandler(appService)
	aiCapabilityCenterService := aiccservices.NewService(db)
	aiCapabilityCenterService.StartProviderAPIConnectivityProbe(context.Background(), 10*time.Minute)
	aiCapabilityCenterHandler := aicchandlers.NewHandler(aiCapabilityCenterService)
	integrationCenterHandler := ichandlers.NewHandler(icservices.NewService(icrepos.NewRepository(db)))

	router.GET("/health", healthHandler.Check)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)
	router.GET("/metrics", metricsCollector.Handler(healthHandler.DependencyStatus))
	router.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(200, openAPISpec())
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, swaggerUIHTML())
	})

	registerAPIRoutes(router, identityHandler, paramHandler, appHandler, aiCapabilityCenterHandler, integrationCenterHandler)

	if missing := handlers.UnclassifiedAPIRoutes(router.Routes()); len(missing) > 0 {
		panic("unclassified API routes: " + strings.Join(missing, ", "))
	}
	return router
}

func securityHeadersMiddleware(appEnv string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' https://unpkg.com 'unsafe-inline'; style-src 'self' https://unpkg.com 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		if strings.EqualFold(appEnv, "production") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func corsMiddleware(origins string) gin.HandlerFunc {
	allowed := map[string]bool{}
	allowAll := false
	for _, raw := range strings.Split(origins, ",") {
		origin := strings.TrimSpace(raw)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAll = true
			continue
		}
		allowed[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.GetHeader("Access-Control-Allow-Origin") != "" {
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func openAPISpec() gin.H {
	api := func(tag, summary string) gin.H {
		return gin.H{
			"tags":      []string{tag},
			"summary":   summary,
			"responses": gin.H{"200": gin.H{"description": "OK"}},
		}
	}
	return withOpenAPISchemas(gin.H{
		"openapi": "3.0.0",
		"info":    gin.H{"title": "SaaS Baseon Go API", "version": "0.1.0"},
		"tags": []gin.H{
			{"name": "auth", "description": "认证与登录"},
			{"name": "apps", "description": "应用中心"},
			{"name": "ai-capability-center", "description": "AI 能力中心"},
			{"name": "integration-center", "description": "第三方集成中心"},
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
			{"name": "files", "description": "文件上传下载"},
			{"name": "batch", "description": "批量导入导出"},
		},
		"paths": gin.H{
			"/api/auth/captcha":                                       gin.H{"get": api("auth", "验证码")},
			"/api/auth/login":                                         gin.H{"post": api("auth", "登录")},
			"/api/auth/logout":                                        gin.H{"post": api("auth", "退出登录")},
			"/api/auth/phone-login-tenants":                           gin.H{"get": api("auth", "手机号登录主体探测")},
			"/api/auth/switch-tenant":                                 gin.H{"post": api("auth", "切换主体")},
			"/api/auth/switchable-tenants":                            gin.H{"get": api("auth", "可切换主体")},
			"/api/apps":                                               gin.H{"get": api("apps", "应用列表"), "post": api("apps", "创建应用")},
			"/api/apps/stats":                                         gin.H{"get": api("apps", "应用统计")},
			"/api/apps/manifest/template":                             gin.H{"get": api("apps", "Manifest 模板")},
			"/api/apps/manifest/parse":                                gin.H{"post": api("apps", "解析 Manifest")},
			"/api/apps/manifest/diff":                                 gin.H{"post": api("apps", "Manifest 差异预检")},
			"/api/apps/manifest/load":                                 gin.H{"post": api("apps", "装载 Manifest")},
			"/api/apps/manifest/scan":                                 gin.H{"post": api("apps", "扫描 Manifest")},
			"/api/apps/{id}":                                          gin.H{"get": api("apps", "应用详情"), "put": api("apps", "编辑应用")},
			"/api/apps/{id}/status":                                   gin.H{"patch": api("apps", "启停应用")},
			"/api/ai-capability-center/overview":                      gin.H{"get": api("ai-capability-center", "AI 能力中心总览")},
			"/api/ai-capability-center/providers/import":              gin.H{"post": api("ai-capability-center", "整体导入 AI 供应商")},
			"/api/ai-capability-center/models/import":                 gin.H{"post": api("ai-capability-center", "整体导入 AI 模型和价格")},
			"/api/ai-capability-center/scenarios/import":              gin.H{"post": api("ai-capability-center", "批量导入 AI 场景")},
			"/api/ai-capability-center/routes/import":                 gin.H{"post": api("ai-capability-center", "整体导入基础路由和模型池")},
			"/api/ai-capability-center/tenant-strategies/import":      gin.H{"post": api("ai-capability-center", "整体导入租户策略和规则")},
			"/api/ai-capability-center/apis/connectivity-check":       gin.H{"post": api("ai-capability-center", "AI API 连通性检测")},
			"/api/ai-capability-center/{resource}":                    gin.H{"get": api("ai-capability-center", "AI 能力中心资源列表"), "post": api("ai-capability-center", "新增 AI 能力中心资源")},
			"/api/ai-capability-center/{resource}/{id}":               gin.H{"put": api("ai-capability-center", "更新 AI 能力中心资源"), "delete": api("ai-capability-center", "删除 AI 能力中心资源")},
			"/api/ai-gateway/v1/invoke":                               gin.H{"post": api("ai-capability-center", "AI Gateway 调用")},
			"/api/integration-center/overview":                        gin.H{"get": api("integration-center", "第三方集成中心总览")},
			"/api/integration-center/connectivity-check":              gin.H{"post": api("integration-center", "全局连通性检测")},
			"/api/integration-center/platforms":                       gin.H{"get": api("integration-center", "接入平台"), "post": api("integration-center", "新增接入平台")},
			"/api/integration-center/platforms/{code}":                gin.H{"put": api("integration-center", "更新接入平台")},
			"/api/integration-center/workspace":                       gin.H{"get": api("integration-center", "集成工作台")},
			"/api/integration-center/provider-apps":                   gin.H{"post": api("integration-center", "新增服务商应用")},
			"/api/integration-center/provider-apps/{code}":            gin.H{"put": api("integration-center", "更新服务商应用")},
			"/api/integration-center/app-capabilities/{id}":           gin.H{"patch": api("integration-center", "更新应用能力连接")},
			"/api/integration-center/tenant-connections":              gin.H{"get": api("integration-center", "租户连接")},
			"/api/integration-center/tenant-connections/{id}/refresh": gin.H{"post": api("integration-center", "刷新租户连接授权")},
			"/api/integration-center/tenant-connections/{id}/pause":   gin.H{"post": api("integration-center", "暂停租户连接")},
			"/api/integration-center/tenant-connections/{id}/resume":  gin.H{"post": api("integration-center", "恢复租户连接")},
			"/api/integration-center/tenant-connections/{id}/retry":   gin.H{"post": api("integration-center", "租户连接重试同步")},
			"/api/integration-center/sync-monitor":                    gin.H{"get": api("integration-center", "同步监控")},
			"/api/integration-center/sync-jobs/{id}/retry":            gin.H{"post": api("integration-center", "重试同步任务")},
			"/api/integration-center/sync-jobs/{id}/pause":            gin.H{"post": api("integration-center", "暂停同步任务")},
			"/api/integration-center/sync-jobs/{id}/resume":           gin.H{"post": api("integration-center", "恢复同步任务")},
			"/api/integration-center/quota":                           gin.H{"get": api("integration-center", "配额与限流")},
			"/api/integration-center/quota-policies":                  gin.H{"post": api("integration-center", "新增配额策略")},
			"/api/integration-center/quota-policies/{code}":           gin.H{"put": api("integration-center", "更新配额策略")},
			"/api/integration-center/quota-policies/{code}/status":    gin.H{"patch": api("integration-center", "启停配额策略")},
			"/api/integration-center/alerts":                          gin.H{"get": api("integration-center", "异常监控")},
			"/api/integration-center/alerts/{id}/process":             gin.H{"post": api("integration-center", "处理异常")},
			"/api/integration-center/alerts/{id}/resolve":             gin.H{"post": api("integration-center", "恢复异常")},
			"/api/integration-center/alerts/{id}/ignore":              gin.H{"post": api("integration-center", "忽略异常")},
			"/api/integration-center/logs":                            gin.H{"get": api("integration-center", "调用日志")},
			"/api/integration-center/logs/export":                     gin.H{"post": api("integration-center", "调用日志导出")},
			"/api/batch/companies/export":                             gin.H{"get": api("batch", "公司导出")},
			"/api/batch/departments/export":                           gin.H{"get": api("batch", "部门导出")},
			"/api/batch/users/export":                                 gin.H{"get": api("batch", "用户导出")},
			"/api/batch/users/import":                                 gin.H{"post": api("batch", "用户导入")},
			"/api/business-units":                                     gin.H{"get": api("organizations", "业务单元列表"), "post": api("organizations", "创建业务单元")},
			"/api/business-units/org-mappings":                        gin.H{"get": api("organizations", "业务单元组织映射")},
			"/api/business-units/org-mappings/{id}":                   gin.H{"delete": api("organizations", "删除业务单元组织映射")},
			"/api/business-units/tree":                                gin.H{"get": api("organizations", "业务单元树")},
			"/api/business-units/{id}":                                gin.H{"put": api("organizations", "更新业务单元"), "delete": api("organizations", "删除业务单元")},
			"/api/business-units/{id}/org-mappings":                   gin.H{"get": api("organizations", "业务单元组织映射"), "post": api("organizations", "新增业务单元组织映射")},
			"/api/companies":                                          gin.H{"post": api("organizations", "创建公司")},
			"/api/companies/{id}":                                     gin.H{"put": api("organizations", "更新公司"), "delete": api("organizations", "删除公司")},
			"/api/departments":                                        gin.H{"post": api("organizations", "创建部门")},
			"/api/departments/{id}":                                   gin.H{"put": api("organizations", "更新部门"), "delete": api("organizations", "删除部门")},
			"/api/dict-items":                                         gin.H{"get": api("dict-param", "字典项列表"), "post": api("dict-param", "创建字典项")},
			"/api/dict-items/{id}":                                    gin.H{"put": api("dict-param", "更新字典项"), "delete": api("dict-param", "删除字典项")},
			"/api/dict-items/{id}/override":                           gin.H{"delete": api("dict-param", "恢复字典项覆盖")},
			"/api/dict-types":                                         gin.H{"get": api("dict-param", "字典类型列表"), "post": api("dict-param", "创建字典类型")},
			"/api/dict-types/by-code":                                 gin.H{"get": api("dict-param", "按编码查询字典项")},
			"/api/dict-types/by-code/{code}/items":                    gin.H{"get": api("dict-param", "按编码查询字典项")},
			"/api/dict-types/{id}":                                    gin.H{"put": api("dict-param", "更新字典类型"), "delete": api("dict-param", "删除字典类型")},
			"/api/files/{file_id}":                                    gin.H{"delete": api("files", "删除文件")},
			"/api/files/download/{file_id}":                           gin.H{"get": api("files", "下载文件")},
			"/api/files/upload":                                       gin.H{"post": api("files", "上传文件")},
			"/api/logs/audit":                                         gin.H{"get": api("logs", "操作日志")},
			"/api/logs/login":                                         gin.H{"get": api("logs", "登录日志")},
			"/api/monitor/cache-keys":                                 gin.H{"get": api("monitor", "缓存列表")},
			"/api/monitor/cache-stats":                                gin.H{"get": api("monitor", "缓存监控")},
			"/api/monitor/health-detail":                              gin.H{"get": api("monitor", "健康检查")},
			"/api/monitor/scheduled-jobs":                             gin.H{"get": api("monitor", "定时任务")},
			"/api/monitor/server-info":                                gin.H{"get": api("monitor", "服务器信息")},
			"/api/monitor/services-overview":                          gin.H{"get": api("monitor", "服务监控")},
			"/api/org-nodes":                                          gin.H{"post": api("organizations", "创建组织节点")},
			"/api/org-nodes/{id}":                                     gin.H{"put": api("organizations", "更新组织节点"), "delete": api("organizations", "删除组织节点")},
			"/api/organizations/detail":                               gin.H{"get": api("organizations", "组织详情")},
			"/api/organizations/tree":                                 gin.H{"get": api("organizations", "组织树")},
			"/api/params":                                             gin.H{"get": api("dict-param", "系统参数列表"), "post": api("dict-param", "创建系统参数")},
			"/api/params/{key}":                                       gin.H{"get": api("dict-param", "按键查询系统参数")},
			"/api/permission-menu-bundles":                            gin.H{"get": api("permissions", "菜单权限包")},
			"/api/permissions":                                        gin.H{"get": api("permissions", "权限列表"), "post": api("permissions", "创建权限")},
			"/api/permissions/menu-bundles":                           gin.H{"get": api("permissions", "菜单权限包")},
			"/api/permissions/menu-data-perm-mode":                    gin.H{"put": api("permissions", "更新菜单数据权限模式")},
			"/api/permissions/menu-data-perm-mode/{id}":               gin.H{"put": api("permissions", "更新菜单数据权限模式")},
			"/api/permissions/menu-package-feature/{id}":              gin.H{"put": api("permissions", "更新套餐中心收录")},
			"/api/permissions/menu-overrides":                         gin.H{"get": api("permissions", "菜单覆盖"), "put": api("permissions", "保存菜单覆盖")},
			"/api/permissions/tree":                                   gin.H{"get": api("permissions", "权限树")},
			"/api/permissions/{id}":                                   gin.H{"get": api("permissions", "权限详情"), "put": api("permissions", "更新权限"), "delete": api("permissions", "删除权限")},
			"/api/plans":                                              gin.H{"get": api("tenants", "套餐列表"), "post": api("tenants", "创建套餐")},
			"/api/plans/features":                                     gin.H{"get": api("tenants", "功能列表"), "post": api("tenants", "创建功能")},
			"/api/plans/features/{id}":                                gin.H{"put": api("tenants", "更新功能")},
			"/api/plans/matrix":                                       gin.H{"get": api("tenants", "套餐能力矩阵")},
			"/api/plans/quotas":                                       gin.H{"get": api("tenants", "配额列表"), "post": api("tenants", "创建配额")},
			"/api/plans/quotas/{id}":                                  gin.H{"put": api("tenants", "更新配额")},
			"/api/plans/{id}":                                         gin.H{"put": api("tenants", "更新套餐"), "delete": api("tenants", "删除套餐")},
			"/api/plans/{id}/capabilities":                            gin.H{"put": api("tenants", "保存套餐能力")},
			"/api/plans/{id}/copy":                                    gin.H{"post": api("tenants", "复制套餐")},
			"/api/plans/{id}/features":                                gin.H{"get": api("tenants", "套餐功能"), "put": api("tenants", "保存套餐功能")},
			"/api/plans/{id}/quotas":                                  gin.H{"get": api("tenants", "套餐配额"), "put": api("tenants", "保存套餐配额")},
			"/api/position-types":                                     gin.H{"get": api("positions", "岗位类型列表"), "post": api("positions", "创建岗位类型")},
			"/api/position-types/{id}":                                gin.H{"put": api("positions", "更新岗位类型"), "delete": api("positions", "删除岗位类型")},
			"/api/positions":                                          gin.H{"get": api("positions", "岗位列表"), "post": api("positions", "创建岗位")},
			"/api/positions/{id}":                                     gin.H{"put": api("positions", "更新岗位"), "delete": api("positions", "删除岗位")},
			"/api/public":                                             gin.H{"get": api("branding", "公开页脚")},
			"/api/public/tenant-footer":                               gin.H{"get": api("branding", "公开页脚")},
			"/api/roles":                                              gin.H{"get": api("roles", "角色列表"), "post": api("roles", "创建角色")},
			"/api/roles/permission-menu-bundles":                      gin.H{"get": api("permissions", "菜单权限包")},
			"/api/roles/{id}":                                         gin.H{"get": api("roles", "角色详情"), "put": api("roles", "更新角色"), "delete": api("roles", "删除角色")},
			"/api/roles/{id}/permissions":                             gin.H{"put": api("roles", "配置角色权限")},
			"/api/stores":                                             gin.H{"post": api("organizations", "创建门店")},
			"/api/stores/{id}":                                        gin.H{"put": api("organizations", "更新门店"), "delete": api("organizations", "删除门店")},
			"/api/sys-params":                                         gin.H{"get": api("dict-param", "系统参数列表"), "post": api("dict-param", "创建系统参数")},
			"/api/sys-params/batch":                                   gin.H{"get": api("dict-param", "批量查询系统参数")},
			"/api/sys-params/{id}":                                    gin.H{"put": api("dict-param", "更新系统参数"), "delete": api("dict-param", "删除系统参数")},
			"/api/sys-params/{id}/override":                           gin.H{"delete": api("dict-param", "恢复系统参数覆盖")},
			"/api/tenant/branding":                                    gin.H{"get": api("branding", "租户品牌"), "put": api("branding", "保存租户品牌")},
			"/api/tenants":                                            gin.H{"get": api("tenants", "主体列表"), "post": api("tenants", "创建主体")},
			"/api/tenants/with-package":                               gin.H{"post": api("tenants", "创建主体及套餐")},
			"/api/tenants/{id}":                                       gin.H{"get": api("tenants", "主体详情"), "put": api("tenants", "更新主体"), "delete": api("tenants", "删除主体")},
			"/api/tenants/{id}/companies":                             gin.H{"get": api("tenants", "主体公司")},
			"/api/tenants/{id}/feature-access/{feature_code}":         gin.H{"get": api("tenants", "主体功能访问检查")},
			"/api/tenants/{id}/feature-overrides":                     gin.H{"get": api("tenants", "主体功能覆盖"), "put": api("tenants", "保存主体功能覆盖")},
			"/api/tenants/{id}/package-config":                        gin.H{"put": api("tenants", "保存主体套餐配置")},
			"/api/tenants/{id}/primary-admin":                         gin.H{"get": api("tenants", "主体主管理员")},
			"/api/tenants/{id}/primary-admin/password":                gin.H{"put": api("tenants", "重置主体主管理员密码")},
			"/api/tenants/{id}/quota-check/{quota_code}":              gin.H{"get": api("tenants", "主体配额检查")},
			"/api/tenants/{id}/quota-overrides":                       gin.H{"get": api("tenants", "主体配额覆盖"), "put": api("tenants", "保存主体配额覆盖")},
			"/api/tenants/{id}/quota-records":                         gin.H{"get": api("tenants", "主体配额记录")},
			"/api/tenants/{id}/quota-usage":                           gin.H{"get": api("tenants", "主体配额用量")},
			"/api/tenants/{id}/status":                                gin.H{"patch": api("tenants", "更新主体状态")},
			"/api/tenants/{id}/subscription":                          gin.H{"get": api("tenants", "主体订阅"), "put": api("tenants", "保存主体订阅")},
			"/api/users":                                              gin.H{"get": api("users", "用户列表"), "post": api("users", "创建用户")},
			"/api/users/assignable-roles":                             gin.H{"get": api("users", "可分配角色")},
			"/api/users/me":                                           gin.H{"get": api("users", "当前用户"), "put": api("users", "更新当前用户")},
			"/api/users/me/password":                                  gin.H{"put": api("users", "修改当前用户密码")},
			"/api/users/me/preferences":                               gin.H{"get": api("users", "当前用户偏好"), "put": api("users", "保存当前用户偏好")},
			"/api/users/{id}":                                         gin.H{"put": api("users", "更新用户"), "delete": api("users", "删除用户")},
			"/api/users/{id}/password":                                gin.H{"put": api("users", "重置用户密码")},
		},
	})
}

func swaggerUIHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>SaaS Admin API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
</head>
<body>
  <div id="swagger-ui"></div>
  <script>
    SwaggerUIBundle({
      url: "/openapi.json",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis],
      layout: "BaseLayout",
    });
  </script>
</body>
</html>`
}

func withOpenAPISchemas(spec gin.H) gin.H {
	spec["components"] = openAPIComponents()
	paths, ok := spec["paths"].(gin.H)
	if !ok {
		return spec
	}
	for _, rawPath := range paths {
		methods, ok := rawPath.(gin.H)
		if !ok {
			continue
		}
		for method, rawOperation := range methods {
			operation, ok := rawOperation.(gin.H)
			if !ok {
				continue
			}
			if _, ok := operation["responses"]; !ok {
				operation["responses"] = openAPIStandardResponses("#/components/schemas/ObjectData")
			} else {
				operation["responses"] = openAPIStandardResponses("#/components/schemas/ObjectData")
			}
			if method == "delete" {
				operation["responses"] = openAPIStandardResponses("#/components/schemas/DeleteResult")
			}
		}
	}
	setOpenAPIOperation(paths, "/api/auth/login", "post", "#/components/schemas/LoginRequest", "#/components/schemas/LoginResponse")
	setOpenAPIOperation(paths, "/api/apps", "get", "", "#/components/schemas/AppPage")
	setOpenAPIOperation(paths, "/api/apps", "post", "#/components/schemas/ObjectData", "#/components/schemas/App")
	setOpenAPIOperation(paths, "/api/apps/stats", "get", "", "#/components/schemas/AppStats")
	setOpenAPIOperation(paths, "/api/apps/manifest/template", "get", "", "#/components/schemas/ObjectData")
	setOpenAPIOperation(paths, "/api/apps/manifest/parse", "post", "#/components/schemas/ObjectData", "#/components/schemas/ObjectData")
	setOpenAPIOperation(paths, "/api/apps/manifest/scan", "post", "#/components/schemas/ObjectData", "#/components/schemas/ObjectData")
	setOpenAPIOperation(paths, "/api/apps/{id}", "get", "", "#/components/schemas/App")
	setOpenAPIOperation(paths, "/api/apps/{id}", "put", "#/components/schemas/ObjectData", "#/components/schemas/App")
	setOpenAPIOperation(paths, "/api/apps/{id}/status", "patch", "#/components/schemas/ObjectData", "#/components/schemas/App")
	setOpenAPIOperation(paths, "/api/users", "get", "", "#/components/schemas/UserPage")
	setOpenAPIOperation(paths, "/api/users", "post", "#/components/schemas/UserCreateRequest", "#/components/schemas/User")
	setOpenAPIOperation(paths, "/api/users/{id}", "put", "#/components/schemas/UserUpdateRequest", "#/components/schemas/User")
	setOpenAPIOperation(paths, "/api/roles", "get", "", "#/components/schemas/RolePage")
	setOpenAPIOperation(paths, "/api/roles", "post", "#/components/schemas/RoleCreateRequest", "#/components/schemas/Role")
	setOpenAPIOperation(paths, "/api/roles/{id}", "get", "", "#/components/schemas/Role")
	setOpenAPIOperation(paths, "/api/roles/{id}", "put", "#/components/schemas/RoleUpdateRequest", "#/components/schemas/Role")
	setOpenAPIOperation(paths, "/api/tenants", "get", "", "#/components/schemas/TenantPage")
	setOpenAPIOperation(paths, "/api/tenants", "post", "#/components/schemas/TenantCreateRequest", "#/components/schemas/Tenant")
	setOpenAPIOperation(paths, "/api/tenants/{id}", "get", "", "#/components/schemas/Tenant")
	setOpenAPIOperation(paths, "/api/tenants/{id}", "put", "#/components/schemas/TenantUpdateRequest", "#/components/schemas/Tenant")
	setOpenAPIOperation(paths, "/api/organizations/tree", "get", "", "#/components/schemas/OrgNodeList")
	setOpenAPIOperation(paths, "/api/org-nodes", "post", "#/components/schemas/OrgNodeRequest", "#/components/schemas/IDResult")
	setOpenAPIOperation(paths, "/api/org-nodes/{id}", "put", "#/components/schemas/OrgNodeRequest", "#/components/schemas/IDResult")
	setOpenAPIOperation(paths, "/api/business-units", "get", "", "#/components/schemas/BusinessUnitPage")
	setOpenAPIOperation(paths, "/api/business-units", "post", "#/components/schemas/BusinessUnitRequest", "#/components/schemas/IDResult")
	setOpenAPIOperation(paths, "/api/business-units/{id}", "put", "#/components/schemas/BusinessUnitRequest", "#/components/schemas/IDResult")
	setOpenAPIOperation(paths, "/api/logs/audit", "get", "", "#/components/schemas/AuditLogPage")
	return spec
}

func setOpenAPIOperation(paths gin.H, path, method, requestSchema, responseSchema string) {
	methods, ok := paths[path].(gin.H)
	if !ok {
		return
	}
	operation, ok := methods[method].(gin.H)
	if !ok {
		return
	}
	if requestSchema != "" {
		operation["requestBody"] = gin.H{
			"required": true,
			"content":  gin.H{"application/json": gin.H{"schema": refSchema(requestSchema)}},
		}
	}
	if responseSchema != "" {
		operation["responses"] = openAPIStandardResponses(responseSchema)
	}
}

func openAPIStandardResponses(dataSchema string) gin.H {
	return gin.H{
		"200": gin.H{"description": "OK", "content": gin.H{"application/json": gin.H{"schema": apiEnvelope(dataSchema)}}},
		"400": gin.H{"description": "Bad Request", "content": gin.H{"application/json": gin.H{"schema": refSchema("#/components/schemas/ErrorResponse")}}},
		"401": gin.H{"description": "Unauthorized", "content": gin.H{"application/json": gin.H{"schema": refSchema("#/components/schemas/ErrorResponse")}}},
		"403": gin.H{"description": "Forbidden", "content": gin.H{"application/json": gin.H{"schema": refSchema("#/components/schemas/ErrorResponse")}}},
		"404": gin.H{"description": "Not Found", "content": gin.H{"application/json": gin.H{"schema": refSchema("#/components/schemas/ErrorResponse")}}},
	}
}

func apiEnvelope(dataSchema string) gin.H {
	return gin.H{
		"type": "object",
		"properties": gin.H{
			"code":    gin.H{"type": "integer", "example": 0},
			"message": gin.H{"type": "string", "example": "ok"},
			"data":    refSchema(dataSchema),
		},
		"required": []string{"code", "message", "data"},
	}
}

func refSchema(ref string) gin.H {
	return gin.H{"$ref": ref}
}

func openAPIComponents() gin.H {
	return gin.H{"schemas": gin.H{
		"ErrorResponse":       objectSchema(gin.H{"code": integerSchema(), "message": stringSchema()}, "code", "message"),
		"ObjectData":          objectSchema(gin.H{}),
		"IDResult":            objectSchema(gin.H{"id": integerSchema()}),
		"DeleteResult":        objectSchema(gin.H{"deleted": integerSchema(), "id": integerSchema()}),
		"LoginRequest":        objectSchema(gin.H{"account": stringSchema(), "password": stringSchema(), "captcha_key": stringSchema(), "captcha": stringSchema()}, "account", "password"),
		"LoginResponse":       objectSchema(gin.H{"access_token": stringSchema(), "token_type": stringSchema(), "expires_in": integerSchema(), "user": refSchema("#/components/schemas/User")}),
		"App":                 objectSchema(gin.H{"id": integerSchema(), "app_code": stringSchema(), "app_name": stringSchema(), "icon": nullableStringSchema(), "app_type": stringSchema(), "source": stringSchema(), "status": stringSchema(), "charge_mode": stringSchema(), "visibility_scope": stringSchema(), "deployment_mode": stringSchema(), "communication_modes": nullableStringSchema(), "owner": nullableStringSchema(), "owner_user_ids": nullableStringSchema(), "version": nullableStringSchema(), "description": nullableStringSchema(), "detail_description": nullableStringSchema(), "is_builtin": gin.H{"type": "boolean"}, "is_platform_only": gin.H{"type": "boolean"}, "sort_order": integerSchema()}),
		"AppPage":             pageSchema("#/components/schemas/App"),
		"AppStats":            objectSchema(gin.H{"total": integerSchema(), "online": integerSchema(), "builtin": integerSchema(), "disabled": integerSchema()}),
		"TenantCreateRequest": objectSchema(gin.H{"code": stringSchema(), "name": stringSchema(), "status": integerSchema(), "admin_name": stringSchema(), "admin_employee_no": stringSchema(), "admin_phone": stringSchema(), "admin_password": stringSchema()}, "code", "name", "admin_name", "admin_employee_no", "admin_password"),
		"TenantUpdateRequest": objectSchema(gin.H{"name": stringSchema(), "status": integerSchema(), "contact_name": stringSchema(), "contact_phone": stringSchema()}),
		"Tenant":              objectSchema(gin.H{"id": integerSchema(), "code": stringSchema(), "name": stringSchema(), "status": integerSchema(), "plan_name": nullableStringSchema(), "used_users": integerSchema(), "used_companies": integerSchema()}),
		"TenantPage":          pageSchema("#/components/schemas/Tenant"),
		"UserCreateRequest":   objectSchema(gin.H{"employee_no": stringSchema(), "password": stringSchema(), "name": stringSchema(), "phone": stringSchema(), "email": stringSchema(), "company_id": integerSchema(), "department_id": integerSchema(), "position_ids": integerArraySchema(), "role_ids": integerArraySchema(), "status": integerSchema()}, "employee_no", "password", "name"),
		"UserUpdateRequest":   objectSchema(gin.H{"name": stringSchema(), "phone": stringSchema(), "email": stringSchema(), "company_id": integerSchema(), "department_id": integerSchema(), "department_ids": integerArraySchema(), "position_ids": integerArraySchema(), "role_ids": integerArraySchema(), "status": integerSchema(), "is_platform_admin": gin.H{"type": "boolean"}}),
		"User":                objectSchema(gin.H{"id": integerSchema(), "tenant_id": integerSchema(), "employee_no": stringSchema(), "account": stringSchema(), "name": stringSchema(), "phone": nullableStringSchema(), "email": nullableStringSchema(), "status": integerSchema(), "role_ids": integerArraySchema(), "position_ids": integerArraySchema()}),
		"UserPage":            pageSchema("#/components/schemas/User"),
		"RoleCreateRequest":   objectSchema(gin.H{"code": stringSchema(), "name": stringSchema(), "description": stringSchema(), "permission_ids": integerArraySchema()}, "code", "name"),
		"RoleUpdateRequest":   objectSchema(gin.H{"name": stringSchema(), "description": stringSchema(), "permission_ids": integerArraySchema()}),
		"Role":                objectSchema(gin.H{"id": integerSchema(), "tenant_id": integerSchema(), "code": stringSchema(), "name": stringSchema(), "description": nullableStringSchema(), "permission_ids": integerArraySchema()}),
		"RolePage":            pageSchema("#/components/schemas/Role"),
		"OrgNodeRequest":      objectSchema(gin.H{"node_type": stringSchema(), "name": stringSchema(), "code": stringSchema(), "company_type": stringSchema(), "company_id": integerSchema(), "parent_id": integerSchema(), "status": integerSchema()}, "name"),
		"OrgNode":             objectSchema(gin.H{"id": integerSchema(), "node_type": stringSchema(), "name": stringSchema(), "code": nullableStringSchema(), "parent_id": integerSchema(), "status": integerSchema(), "children": gin.H{"type": "array", "items": gin.H{"type": "object"}}}),
		"OrgNodeList":         gin.H{"type": "array", "items": refSchema("#/components/schemas/OrgNode")},
		"BusinessUnitRequest": objectSchema(gin.H{"name": stringSchema(), "code": stringSchema(), "bu_type": stringSchema(), "org_node_ids": integerArraySchema(), "status": integerSchema(), "remark": stringSchema()}, "name", "code"),
		"BusinessUnit":        objectSchema(gin.H{"id": integerSchema(), "tenant_id": integerSchema(), "name": stringSchema(), "code": stringSchema(), "bu_type": nullableStringSchema(), "status": integerSchema(), "org_node_ids": integerArraySchema()}),
		"BusinessUnitPage":    pageSchema("#/components/schemas/BusinessUnit"),
		"AuditLog":            objectSchema(gin.H{"id": integerSchema(), "tenant_id": integerSchema(), "user_id": integerSchema(), "module": stringSchema(), "action": stringSchema(), "summary": stringSchema(), "ip": nullableStringSchema(), "created_at": stringSchema()}),
		"AuditLogPage":        pageSchema("#/components/schemas/AuditLog"),
	}}
}

func pageSchema(itemRef string) gin.H {
	return objectSchema(gin.H{
		"items": gin.H{"type": "array", "items": refSchema(itemRef)},
		"total": integerSchema(),
		"skip":  integerSchema(),
		"limit": integerSchema(),
	}, "items", "total", "skip", "limit")
}

func objectSchema(properties gin.H, required ...string) gin.H {
	schema := gin.H{"type": "object", "properties": properties}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema() gin.H {
	return gin.H{"type": "string"}
}

func nullableStringSchema() gin.H {
	return gin.H{"type": "string", "nullable": true}
}

func integerSchema() gin.H {
	return gin.H{"type": "integer", "format": "int64"}
}

func integerArraySchema() gin.H {
	return gin.H{"type": "array", "items": integerSchema()}
}
