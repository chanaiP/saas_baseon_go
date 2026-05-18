package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func TestAuthSessionKeyUsesOriginalRedisNamespace(t *testing.T) {
	require.Equal(t, "auth:session:opaque-token", authSessionKey("opaque-token"))
}

func TestLooksLikeMobileAccount(t *testing.T) {
	require.True(t, looksLikeMobileAccount("13800138000"))
	require.False(t, looksLikeMobileAccount("23800138000"))
	require.False(t, looksLikeMobileAccount("1380013800"))
	require.False(t, looksLikeMobileAccount("1380013800a"))
}

func TestSubscriptionStatusAllowsLoginMatchesOriginalPolicy(t *testing.T) {
	now := time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	require.True(t, subscriptionStatusAllowsLogin("", nil, now))
	require.True(t, subscriptionStatusAllowsLogin("TRIAL", &future, now))
	require.True(t, subscriptionStatusAllowsLogin("ACTIVE", &future, now))
	require.False(t, subscriptionStatusAllowsLogin("ACTIVE", &past, now))
	require.False(t, subscriptionStatusAllowsLogin("OVERDUE", &future, now))
	require.False(t, subscriptionStatusAllowsLogin("FROZEN", &future, now))
	require.False(t, subscriptionStatusAllowsLogin("EXPIRED", &future, now))
	require.False(t, subscriptionStatusAllowsLogin("CANCELLED", &future, now))
	require.False(t, subscriptionStatusAllowsLogin("UNKNOWN", &future, now))
}

func TestValidateNewPasswordMatchesOriginalPolicy(t *testing.T) {
	require.Empty(t, validateNewPassword("old12345", "new12345", "new12345"))
	require.Equal(t, "请求参数错误", validateNewPassword("", "new12345", "new12345"))
	require.Equal(t, "新密码须同时包含英文字母与数字", validateNewPassword("old12345", "abcdefgh", "abcdefgh"))
	require.Equal(t, "两次输入的新密码不一致", validateNewPassword("old12345", "new12345", "new12346"))
	require.Equal(t, "新密码不能与当前密码相同", validateNewPassword("same1234", "same1234", "same1234"))
}

func TestRateLimitNoopsWithoutRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	h := &IdentityHandler{}

	message, blocked := h.rateLimitExceeded(c, "rate:test", 1, time.Minute)

	require.Empty(t, message)
	require.False(t, blocked)
}

func TestViewerHasPlatformScopeRequiresExplicitPlatformAdmin(t *testing.T) {
	h := &IdentityHandler{}

	require.True(t, h.viewerHasPlatformScope(models.AppUser{IsPlatformAdmin: true}))
	require.False(t, h.viewerHasPlatformScope(models.AppUser{IsPlatformAdmin: false, TenantID: 1}))
}

func TestNormalizeOptionalPhoneMatchesOriginalPolicy(t *testing.T) {
	raw := "138 0013-8000"
	phone, msg := normalizeOptionalPhone(&raw)
	require.Empty(t, msg)
	require.NotNil(t, phone)
	require.Equal(t, "13800138000", *phone)

	bad := "中文手机号"
	phone, msg = normalizeOptionalPhone(&bad)
	require.Nil(t, phone)
	require.Equal(t, "手机号只能包含数字、空格、短横线或括号", msg)
}

func TestLoginRejectsDisabledUserEvenWhenPasswordMatches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Tenant{}, &models.AppUser{}, &models.LoginLog{}, &models.TenantSubscription{}))
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "demo", Name: "演示主体", Status: 1}).Error)
	hash, err := hashPassword("pass1234")
	require.NoError(t, err)
	require.NoError(t, db.Create(&models.AppUser{
		ID:           7,
		TenantID:     1,
		EmployeeNo:   "u0001",
		Account:      "u0001",
		PasswordHash: hash,
		Name:         "停用用户",
		Status:       0,
	}).Error)
	require.NoError(t, db.Model(&models.AppUser{}).Where("id = ?", 7).Update("status", 0).Error)

	body := []byte(`{"account":"u0001","password":"pass1234","tenant_id":1}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	(&IdentityHandler{db: db}).Login(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var payload response.Body
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, response.CodeBadRequest, payload.Code)
	require.Equal(t, "账号已停用", payload.Message)
}

func TestNormalizedUserDepartmentIDsDedupesAndPrependsPrimary(t *testing.T) {
	primary := uint64(2)
	require.Equal(t, []uint64{3, 2}, normalizedUserDepartmentIDs(&primary, []uint64{3, 2, 3}))
	require.Equal(t, []uint64{3}, normalizedUserDepartmentIDs(nil, []uint64{3, 3, 0}))
}

func TestTombstoneUniqueValueKeepsMaxLength(t *testing.T) {
	value := tombstoneUniqueValue("13800009999", 7, 32)

	require.LessOrEqual(t, len(value), 32)
	require.Contains(t, value, "__deleted_7_")
}
