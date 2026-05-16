package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/apps/integration_center/services"
	"saas_baseon_go/internal/interfaces/http/response"
)

func TestHandlerInvalidJSONReturnsBadRequestEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/platforms", NewHandler(services.NewService(nil)).CreatePlatform)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/platforms", strings.NewReader(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body response.Body
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, response.CodeBadRequest, body.Code)
	require.Equal(t, "请求参数错误", body.Message)
	require.Nil(t, body.Data)
}

func TestHandlerInvalidIDReturnsBadRequestEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/sync-jobs/:id", NewHandler(services.NewService(nil)).SyncJobDetail)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sync-jobs/not-a-number", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body response.Body
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, response.CodeBadRequest, body.Code)
	require.Equal(t, "资源 ID 不合法", body.Message)
	require.Nil(t, body.Data)
}
