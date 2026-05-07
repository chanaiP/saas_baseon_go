package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequiredPermissionForOperationRoutes(t *testing.T) {
	require.Equal(t, "user:create", requiredPermission("POST", "/api/users"))
	require.Equal(t, "user:edit", requiredPermission("PUT", "/api/users/:id"))
	require.Equal(t, "tenant:quota_config", requiredPermission("PUT", "/api/tenants/:id/quota-overrides"))
	require.Equal(t, "brand:edit", requiredPermission("PUT", "/api/tenant/branding"))
}

func TestRequiredPermissionForMenuRoutes(t *testing.T) {
	require.Equal(t, "/users", requiredPermission("GET", "/api/users"))
	require.Equal(t, "/organization", requiredPermission("GET", "/api/organizations/detail"))
	require.Equal(t, "/monitor/cache-keys", requiredPermission("GET", "/api/monitor/cache-keys"))
	require.Empty(t, requiredPermission("GET", "/api/users/me"))
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
