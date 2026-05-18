package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestRequiredPermissionForOperationRoutes(t *testing.T) {
	require.Equal(t, "user:create", requiredPermission("POST", "/api/users"))
	require.Equal(t, "user:edit", requiredPermission("PUT", "/api/users/:id"))
	require.Equal(t, "tenant:quota_config", requiredPermission("PUT", "/api/tenants/:id/quota-overrides"))
	require.Equal(t, "plan:delete", requiredPermission("DELETE", "/api/plans/:id"))
	require.Equal(t, "role:permission", requiredPermission("PUT", "/api/roles/:id/permissions"))
	require.Equal(t, "param:create", requiredPermission("POST", "/api/params"))
	require.Equal(t, "brand:edit", requiredPermission("PUT", "/api/tenant/branding"))
	require.Equal(t, "dict_type:create", requiredPermission("POST", "/api/dict-types"))
	require.Equal(t, "dict_item:edit", requiredPermission("DELETE", "/api/dict-items/:id/override"))
	require.Equal(t, "menu:edit", requiredPermission("POST", "/api/permissions"))
	require.Equal(t, "menu:edit", requiredPermission("PUT", "/api/permissions/:id"))
	require.Equal(t, "menu:edit", requiredPermission("DELETE", "/api/permissions/:id"))
	require.Equal(t, "menu:package_feature", requiredPermission("PUT", "/api/permissions/menu-package-feature/:id"))
	require.Equal(t, "app:create", requiredPermission("POST", "/api/apps"))
	require.Equal(t, "app:edit", requiredPermission("PUT", "/api/apps/:id"))
	require.Equal(t, "app:status", requiredPermission("PATCH", "/api/apps/:id/status"))
	require.Equal(t, "app:load", requiredPermission("POST", "/api/apps/manifest/parse"))
	require.Equal(t, "app:load", requiredPermission("POST", "/api/apps/manifest/diff"))
	require.Equal(t, "app:load", requiredPermission("POST", "/api/apps/manifest/load"))
	require.Equal(t, "app:load", requiredPermission("POST", "/api/apps/manifest/scan"))
	require.Equal(t, "integration_center:connection_manage", requiredPermission("POST", "/api/integration-center/tenant-connections"))
	require.Equal(t, "integration_center:connection_manage", requiredPermission("POST", "/api/integration-center/oauth/start"))
	require.Equal(t, "/integration-center/my-connections", requiredPermission("POST", "/api/integration-center/my-connections"))
	require.Equal(t, "/integration-center/my-connections", requiredPermission("POST", "/api/integration-center/my-oauth/start"))
	require.Equal(t, "integration_center:connection_manage", requiredPermission("POST", "/api/integration-center/gateway/invoke"))
	require.Equal(t, "ai_gateway:invoke", requiredPermission("POST", "/api/ai-gateway/v1/invoke"))
	require.Equal(t, "ai_gateway:invoke", requiredPermission("GET", "/api/ai-gateway/v1/video-tasks/:task_id"))
	require.Equal(t, "data_center:metric_manage", requiredPermission("PUT", "/api/data-center/metrics/:id"))
	require.Equal(t, "data_center:rule_manage", requiredPermission("POST", "/api/data-center/anomaly-rules/:id/test"))
	require.Equal(t, "data_center:scan", requiredPermission("POST", "/api/data-center/anomalies/scan"))
	require.Equal(t, "data_center:ai_analyze", requiredPermission("POST", "/api/data-center/anomalies/:id/analyze"))
	require.Equal(t, "data_center:task_generate", requiredPermission("POST", "/api/data-center/anomalies/:id/generate-task"))
	require.Equal(t, "data_center:task_flow", requiredPermission("PUT", "/api/data-center/tasks/:id"))
	require.Equal(t, "data_center:review_confirm", requiredPermission("PUT", "/api/data-center/reviews/:id"))
}

func TestRequiredPermissionForMenuRoutes(t *testing.T) {
	require.Equal(t, "/users", requiredPermission("GET", "/api/users"))
	require.Equal(t, "/apps", requiredPermission("GET", "/api/apps"))
	require.Equal(t, "/apps", requiredPermission("GET", "/api/apps/stats"))
	require.Equal(t, "/apps", requiredPermission("GET", "/api/apps/manifest/template"))
	require.Equal(t, "/apps", requiredPermission("GET", "/api/apps/:id"))
	require.Equal(t, "/organization", requiredPermission("GET", "/api/organizations/detail"))
	require.Equal(t, "/monitor/cache-keys", requiredPermission("GET", "/api/monitor/cache-keys"))
	require.Equal(t, "/tenants", requiredPermission("GET", "/api/tenants/:id/quota-usage"))
	require.Equal(t, "/business-units", requiredPermission("GET", "/api/business-units/:id/org-mappings"))
	require.Equal(t, "/params", requiredPermission("GET", "/api/sys-params/batch"))
	require.Equal(t, "/params", requiredPermission("GET", "/api/params/:key"))
	require.Equal(t, "/roles", requiredPermission("GET", "/api/roles/permission-menu-bundles"))
	require.Equal(t, "/menus", requiredPermission("GET", "/api/permissions/tree"))
	require.Equal(t, "/integration-center/tenant-connections", requiredPermission("GET", "/api/integration-center/tenant-connections/:id"))
	require.Equal(t, "/integration-center/my-connections", requiredPermission("GET", "/api/integration-center/my-connections/:id"))
	require.Equal(t, "/integration-center/my-connections", requiredPermission("GET", "/api/integration-center/my-sync-jobs/:id"))
	require.Equal(t, "/integration-center/sync-monitor", requiredPermission("GET", "/api/integration-center/sync-jobs/:id"))
	require.Equal(t, "/integration-center/logs", requiredPermission("GET", "/api/integration-center/logs/:id"))
	require.Equal(t, "/data-center/dashboard", requiredPermission("GET", "/api/data-center/dashboard/summary"))
	require.Equal(t, "/data-center/raw", requiredPermission("GET", "/api/data-center/raw/batches/:id"))
	require.Equal(t, "/data-center/standard", requiredPermission("GET", "/api/data-center/standard/:data_type/:id"))
	require.Equal(t, "/data-center/metrics", requiredPermission("GET", "/api/data-center/metrics/:id"))
	require.Equal(t, "/data-center/rules", requiredPermission("GET", "/api/data-center/anomaly-rules/:id"))
	require.Equal(t, "/data-center/anomalies", requiredPermission("GET", "/api/data-center/anomalies/:id"))
	require.Equal(t, "/data-center/tasks", requiredPermission("GET", "/api/data-center/tasks/:id"))
	require.Equal(t, "/data-center/reviews", requiredPermission("GET", "/api/data-center/reviews/:id"))
	require.Empty(t, requiredPermission("GET", "/api/dict-types/by-code/:code/items"))
	require.Empty(t, requiredPermission("GET", "/api/users/me"))
}

func TestRouteAllowedFailsClosedForUnclassifiedRoutes(t *testing.T) {
	handler := &IdentityHandler{}

	require.False(t, handler.routeAllowed(testPlatformAdminUser(), "GET", "/api/unclassified"))
	require.True(t, handler.routeAllowed(testPlatformAdminUser(), "GET", "/api/users/me"))
	require.True(t, routePermissionOptional("POST", "/api/integration-center/webhooks/:provider_app_code"))
	require.True(t, routePermissionOptional("GET", "/api/integration-center/oauth/callback/:provider_app_code"))
	require.False(t, handler.routeAllowed(testPlatformAdminUser(), "GET", "/api/users"))
}

func testPlatformAdminUser() models.AppUser {
	return models.AppUser{ID: 1, TenantID: 1, IsPlatformAdmin: true, Status: 1}
}

func testStringPtr(value string) *string {
	return &value
}

func TestFallbackHandlerReturnsStrict404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.NoRoute(NewFallbackHandler().NoRoute)

	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.JSONEq(t, `{"code":40400,"message":"接口不存在"}`, rec.Body.String())
}

func TestCSVAndFileSafetyHelpers(t *testing.T) {
	require.Equal(t, "'=cmd", csvSafe("=cmd"))
	require.Equal(t, "'+cmd", csvSafe("+cmd"))
	require.Equal(t, "'-cmd", csvSafe("-cmd"))
	require.Equal(t, "'@cmd", csvSafe("@cmd"))
	require.Equal(t, "normal", csvSafe("normal"))
	require.Equal(t, "unknown", safeOriginalName(""))
	require.Equal(t, "demo.txt", safeOriginalName("/tmp/demo.txt"))
	require.Equal(t, ".txt", safeFileExt("demo.TXT"))
	require.Equal(t, ".bin", safeFileExt("demo.sh;rm"))
	require.Equal(t, ".bin", safeFileExt("demo.veryveryverylongext"))
}

func TestFeatureQuotaMappingMatchesPlanCatalogPolicy(t *testing.T) {
	require.ElementsMatch(t, []string{"max_users"}, quotaCodesForFeatureCode("user_manage"))
	require.ElementsMatch(t, []string{"max_companies", "max_stores", "max_departments"}, quotaCodesForFeatureCode("org_manage"))
	require.ElementsMatch(t, []string{"daily_import_times"}, quotaCodesForFeatureCode("import_data"))
	require.ElementsMatch(t, []string{"daily_export_times"}, quotaCodesForFeatureCode("export_data"))
	require.Empty(t, quotaCodesForFeatureCode("api_key"))
	require.Empty(t, quotaCodesForFeatureCode("webhook"))
	require.ElementsMatch(t, []string{"max_storage_gb", "max_file_size_mb"}, quotaCodesForFeatureCode("file_manage"))
	require.Empty(t, quotaCodesForFeatureCode("brand_config"))
}

func TestPureViewOperationsDoNotBecomePackageFeatures(t *testing.T) {
	require.Empty(t, packageFeatureOverride("login:view"))
	require.Empty(t, packageFeatureOverride("audit:view"))
	require.True(t, isPureViewPermissionPath("monhealth:view"))
	require.True(t, isPureViewPermissionPath("moncachekeys:view"))
	require.Equal(t, "dict_manage", parentPackageFeatureCodeForOperation("dict_item:edit"))
	require.Equal(t, "dict_manage", parentPackageFeatureCodeForOperation("dict_type:edit"))
	require.Empty(t, parentPackageFeatureCodeForOperation("perm:create"))
	require.True(t, excludedPackageFeaturePath("perm:create"))
	require.True(t, excludedPackageFeaturePath("perm:edit"))
	require.True(t, excludedPackageFeaturePath("perm:delete"))
	require.True(t, excludedPackageFeaturePath("/home"))
	require.True(t, excludedPackageFeaturePath("/tenants"))
	require.True(t, excludedPackageFeaturePath("/plans"))
	require.True(t, excludedPackageFeaturePath("/permissions"))
	require.True(t, excludedPackageFeaturePath("/monitor/health"))
	require.Empty(t, packageFeatureCodeForPermission(models.Permission{Path: "/home", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("home")}))
	require.Empty(t, packageFeatureCodeForPermission(models.Permission{Path: "/tenants", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("tenant_manage")}))
	require.Empty(t, packageFeatureCodeForPermission(models.Permission{Path: "/plans", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("plan_manage")}))
	require.Empty(t, packageFeatureCodeForPermission(models.Permission{Path: "/permissions", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("role_manage")}))
	require.Empty(t, packageFeatureCodeForPermission(models.Permission{Path: "/monitor/health", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("system_monitor")}))
	require.Equal(t, "menu_manage", packageFeatureCodeForPermission(models.Permission{Path: "/menus", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("menu_manage")}))
	require.Equal(t, "brand_config", packageFeatureCodeForPermission(models.Permission{Path: "brand:edit", PermType: 2, IsPackageFeature: true}))
	require.Equal(t, "CONFIG", packageFeatureTypeForPermission(models.Permission{Path: "brand:edit", PermType: 2, IsPackageFeature: true}))
	require.Equal(t, "品牌配置", packageFeatureNameForPermission(models.Permission{Path: "brand:edit", Name: "品牌-维护", PermType: 2, IsPackageFeature: true}))
	require.False(t, excludedPlanMatrixFeatureRow(models.SaasFeature{FeatureCode: "brand_config", FeatureName: "品牌配置", FeatureType: "CONFIG"}))
	require.True(t, excludedPlanMatrixFeatureRow(models.SaasFeature{FeatureCode: "system_monitor", FeatureName: "系统监控", FeatureType: "MENU"}))
	require.False(t, permissionCanJoinPackageCenter(models.Permission{Path: "/home", PermType: 3, IsPackageFeature: true, FeatureCode: testStringPtr("home")}))
	require.False(t, permissionCanJoinPackageCenter(models.Permission{Path: "menu:delete", PermType: 2, IsPackageFeature: true}))
	require.True(t, permissionCanJoinPackageCenter(models.Permission{Path: "menu:edit", PermType: 2, IsPackageFeature: true}))

	allowed := map[string]bool{"login_log": true, "audit_log": false, "system_monitor": false}
	require.True(t, permissionAllowedByFeatureCodeSet(models.Permission{Path: "login:view", PermType: 2, IsPackageFeature: true}, allowed))
	require.True(t, permissionAllowedByFeatureCodeSet(models.Permission{Path: "audit:view", PermType: 2, IsPackageFeature: true}, allowed))
	require.True(t, permissionAllowedByFeatureCodeSet(models.Permission{Path: "monhealth:view", PermType: 2, IsPackageFeature: true}, allowed))
}

func TestMenuBundleOperationsIncludesManifestParentOperation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Permission{}))

	now := time.Now()
	menu := models.Permission{
		TenantID:       1,
		Name:           "AI 能力中心",
		Path:           "/ai-capability-center",
		PermType:       3,
		Enabled:        true,
		Visible:        false,
		ShowInAdmin:    true,
		IsPlatformOnly: true,
		AppCode:        "ai-capability-center",
		DataPermMode:   "NONE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, db.Create(&menu).Error)
	parentID := menu.ID
	op := models.Permission{
		TenantID:       1,
		ParentID:       &parentID,
		Name:           "AI 能力中心-配置管理",
		Path:           "ai_capability_center:manage",
		PermType:       2,
		Enabled:        true,
		Visible:        false,
		IsPlatformOnly: true,
		AppCode:        "ai-capability-center",
		DataPermMode:   "NONE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, db.Create(&op).Error)

	items := (&IdentityHandler{db: db}).menuBundleOperations(1, menu.ID, menu.Path, true)

	require.Len(t, items, 1)
	require.Equal(t, "ai_capability_center:manage", items[0]["path"])
	require.Equal(t, "AI 能力中心-配置管理", items[0]["name"])
}

func TestValidatePermissionPayloadMatchesOriginalDataPermissionPolicy(t *testing.T) {
	dataPermType := 4
	menuPermType := 3
	opPermType := 2
	allScope := "ALL"
	customScope := "CUSTOM"
	badScope := "BAD"
	orgMode := "ORG"
	badMode := "BAD"

	require.Equal(t, "数据权限必须带 data_scope", validatePermissionPayload(permissionPayload{PermType: &dataPermType}))
	require.Equal(t, "CUSTOM 范围需至少指定组织架构或用户", validatePermissionPayload(permissionPayload{PermType: &dataPermType, DataScope: &customScope}))
	require.Empty(t, validatePermissionPayload(permissionPayload{PermType: &dataPermType, DataScope: &customScope, CustomDepartmentIDs: []uint64{1}}))
	require.Equal(t, "无效的数据范围: BAD", validatePermissionPayload(permissionPayload{PermType: &dataPermType, DataScope: &badScope}))
	require.Empty(t, validatePermissionPayload(permissionPayload{PermType: &menuPermType, DataScope: &allScope, DataPermMode: &orgMode}))
	require.Equal(t, "无效的数据权限类型: BAD", validatePermissionPayload(permissionPayload{PermType: &menuPermType, DataPermMode: &badMode}))
	require.Equal(t, "仅菜单权限支持配置数据权限类型", validatePermissionPayload(permissionPayload{PermType: &opPermType, DataPermMode: &orgMode}))
}
