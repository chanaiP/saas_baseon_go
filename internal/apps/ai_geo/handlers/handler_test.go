package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/ai_geo/repositories"
	"saas_baseon_go/internal/apps/ai_geo/services"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestAiGeoHandlerListUsesUnifiedPaginationResponse(t *testing.T) {
	router, db := newAiGeoHandlerTestRouter(t)
	require.NoError(t, db.Create(&models.AiGeoBrandCard{TenantID: 1, BrandCode: "B1", BrandName: "品牌一", Keywords: "[]", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.AiGeoBrandCard{TenantID: 1, BrandCode: "B2", BrandName: "品牌二", Keywords: "[]", Status: "active"}).Error)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/materials/brands?limit=1", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	payload := decodeHandlerPayload(t, rec)
	require.EqualValues(t, 0, payload["code"])
	data := payload["data"].(map[string]interface{})
	require.EqualValues(t, 2, data["total"])
	require.EqualValues(t, 1, data["limit"])
	require.Len(t, data["items"].([]interface{}), 1)
}

func TestAiGeoHandlerParamErrorsFailBeforeService(t *testing.T) {
	router, _ := newAiGeoHandlerTestRouter(t)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/materials/brands/not-a-number", nil))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	payload := decodeHandlerPayload(t, rec)
	require.EqualValues(t, 40000, payload["code"])
	require.Equal(t, "请求参数错误", payload["message"])
}

func TestAiGeoHandlerWritesAuditLogForKeywordCreate(t *testing.T) {
	router, db := newAiGeoHandlerTestRouter(t)
	body := []byte(`{"keyword_group":"搜索意图","keyword":"高级通勤女装","intent":"问答","source":"manual","weight":80}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/materials/keywords", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	payload := decodeHandlerPayload(t, rec)
	require.EqualValues(t, 0, payload["code"])
	var log models.AuditLog
	require.NoError(t, db.Where("tenant_id = ? AND user_id = ? AND module = ? AND action = ?", 1, 10, "ai_geo", "keyword_create").First(&log).Error)
	require.NotNil(t, log.AppCode)
	require.Equal(t, "ai-geo", *log.AppCode)
}

func newAiGeoHandlerTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(
		&models.AiGeoBrandCard{},
		&models.AiGeoKeyword{},
		&models.SaasPlan{},
		&models.SaasQuota{},
		&models.SaasPlanQuota{},
		&models.TenantSubscription{},
		&models.TenantQuotaOverride{},
		&models.TenantQuotaUsage{},
		&models.AuditLog{},
	))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint64(1))
		c.Set("user_id", uint64(10))
		c.Next()
	})
	handler := NewHandler(services.NewService(repositories.NewRepository(db)))
	router.GET("/materials/brands", handler.Brands)
	router.GET("/materials/brands/:id", handler.Brand)
	router.POST("/materials/keywords", handler.CreateKeyword)
	return router, db
}

func decodeHandlerPayload(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	return payload
}
