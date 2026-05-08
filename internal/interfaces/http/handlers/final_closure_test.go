package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func TestCrossTenantAccessWritesSingleAuditLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.AuditLog{}))
	handler := NewIdentityHandler(db, nil, "secret", 24, false)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/users?tenant_id=9", nil)
	c.Set("request_id", "rid-1")
	user := models.AppUser{ID: 7, TenantID: 1, IsPlatformAdmin: true}
	tenantContext := TenantContext{ActorTenantID: 1, TargetTenantID: 9, ActorKind: PlatformActorCrossOperator, CrossTenantOperator: true}

	handler.auditCrossTenantAccess(c, user, tenantContext)
	handler.auditCrossTenantAccess(c, user, tenantContext)

	var logs []models.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1)
	require.Equal(t, "tenant", logs[0].Module)
	require.Equal(t, "cross_tenant_access", logs[0].Action)
	require.NotNil(t, logs[0].Detail)
	require.Contains(t, *logs[0].Detail, `"target_tenant_id":9`)
}

func TestInternalErrorMessageDoesNotLeakToResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	respondBadRequest(c, errors.New(`ERROR: duplicate key violates unique constraint "app_user_email_key" (SQLSTATE 23505)`))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.NotContains(t, strings.ToLower(recorder.Body.String()), "sqlstate")
	require.NotContains(t, strings.ToLower(recorder.Body.String()), "constraint")
	require.Contains(t, recorder.Body.String(), "请求处理失败")
	require.Len(t, c.Errors, 1)
}

func TestResponseSafeMessageMasksFilesystemError(t *testing.T) {
	require.Equal(t, "服务暂时不可用", response.SafeMessage(500, "open /etc/passwd: permission denied"))
}
