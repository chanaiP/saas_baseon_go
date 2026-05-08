package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestRequiredPermissionForOperationRoutes(t *testing.T) {
	require.Equal(t, "user:create", requiredPermission("POST", "/api/users"))
	require.Equal(t, "user:edit", requiredPermission("PUT", "/api/users/:id"))
	require.Equal(t, "tenant:quota_config", requiredPermission("PUT", "/api/tenants/:id/quota-overrides"))
	require.Equal(t, "plan:delete", requiredPermission("DELETE", "/api/plans/:id"))
	require.Equal(t, "param:create", requiredPermission("POST", "/api/params"))
	require.Equal(t, "brand:edit", requiredPermission("PUT", "/api/tenant/branding"))
}

func TestRequiredPermissionForMenuRoutes(t *testing.T) {
	require.Equal(t, "/users", requiredPermission("GET", "/api/users"))
	require.Equal(t, "/organization", requiredPermission("GET", "/api/organizations/detail"))
	require.Equal(t, "/monitor/cache-keys", requiredPermission("GET", "/api/monitor/cache-keys"))
	require.Equal(t, "/tenants", requiredPermission("GET", "/api/tenants/:id/quota-usage"))
	require.Equal(t, "/business-units", requiredPermission("GET", "/api/business-units/:id/org-mappings"))
	require.Equal(t, "/dict", requiredPermission("GET", "/api/dict-types/by-code/:code/items"))
	require.Equal(t, "/params", requiredPermission("GET", "/api/sys-params/batch"))
	require.Equal(t, "/params", requiredPermission("GET", "/api/params/:key"))
	require.Equal(t, "/roles", requiredPermission("GET", "/api/roles/permission-menu-bundles"))
	require.Equal(t, "/menus", requiredPermission("GET", "/api/permissions/tree"))
	require.Empty(t, requiredPermission("GET", "/api/users/me"))
}

func TestRouteAllowedFailsClosedForUnclassifiedRoutes(t *testing.T) {
	handler := &IdentityHandler{}

	require.False(t, handler.routeAllowed(testPlatformAdminUser(), "GET", "/api/unclassified"))
	require.True(t, handler.routeAllowed(testPlatformAdminUser(), "GET", "/api/users/me"))
	require.False(t, handler.routeAllowed(testPlatformAdminUser(), "GET", "/api/users"))
}

func testPlatformAdminUser() models.AppUser {
	return models.AppUser{ID: 1, TenantID: 1, IsPlatformAdmin: true, Status: 1}
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
	require.ElementsMatch(t, []string{"max_api_keys", "daily_api_calls"}, quotaCodesForFeatureCode("api_key"))
	require.ElementsMatch(t, []string{"max_webhooks"}, quotaCodesForFeatureCode("webhook"))
	require.Empty(t, quotaCodesForFeatureCode("brand_config"))
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
