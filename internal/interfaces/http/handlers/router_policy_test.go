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
	require.Equal(t, ".txt", safeFileExt("demo.TXT"))
	require.Equal(t, ".bin", safeFileExt("demo.sh;rm"))
	require.Equal(t, ".bin", safeFileExt("demo.veryveryverylongext"))
}
