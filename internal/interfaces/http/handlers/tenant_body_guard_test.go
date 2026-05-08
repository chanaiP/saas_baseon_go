package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestRequestTenantBodyAllowedRejectsForgedTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPut, "/api/users/1", strings.NewReader(`{"tenant_id":9,"name":"x"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	require.False(t, (&IdentityHandler{}).requestTenantBodyAllowed(c, models.AppUser{TenantID: 7}))
}

func TestRequestTenantBodyAllowedRestoresBodyForBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPut, "/api/users/1", strings.NewReader(`{"tenant_id":7,"name":"x"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	require.True(t, (&IdentityHandler{}).requestTenantBodyAllowed(c, models.AppUser{TenantID: 7}))

	var body struct {
		Name string `json:"name"`
	}
	require.NoError(t, c.ShouldBindJSON(&body))
	require.Equal(t, "x", body.Name)
}
