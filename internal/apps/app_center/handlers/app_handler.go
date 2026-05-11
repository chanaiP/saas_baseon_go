package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/apps/app_center/dto"
	"saas_baseon_go/internal/apps/app_center/services"
	"saas_baseon_go/internal/interfaces/http/response"
)

type AppHandler struct {
	service *services.AppService
}

func NewAppHandler(service *services.AppService) *AppHandler {
	return &AppHandler{service: service}
}

func (h *AppHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	result, err := h.service.ListApps(c.Request.Context(), userID, dto.AppListRequest{
		Skip:    parseIntQuery(c, "skip", 0),
		Limit:   parseIntQuery(c, "limit", 20),
		Keyword: c.Query("keyword"),
		Type:    c.Query("type"),
		Status:  c.Query("status"),
		Source:  c.Query("source"),
	})
	if err != nil {
		writeAppError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AppHandler) Detail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "应用 ID 不合法")
		return
	}
	result, err := h.service.GetApp(c.Request.Context(), userID, id)
	if err != nil {
		writeAppError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AppHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req dto.AppCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.CreateApp(c.Request.Context(), userID, req)
	if err != nil {
		writeAppError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AppHandler) Stats(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	result, err := h.service.Stats(c.Request.Context(), userID)
	if err != nil {
		writeAppError(c, err)
		return
	}
	response.OK(c, result)
}

func currentUserID(c *gin.Context) (uint64, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	switch value := raw.(type) {
	case uint64:
		return value, value > 0
	case uint:
		return uint64(value), value > 0
	case int:
		return uint64(value), value > 0
	default:
		return 0, false
	}
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func writeAppError(c *gin.Context, err error) {
	if errors.Is(err, services.ErrPlatformOnly) {
		response.Error(c, http.StatusForbidden, response.CodeForbidden, "仅平台管理员可访问应用中心")
		return
	}
	switch {
	case errors.Is(err, services.ErrAppNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "应用不存在")
	case errors.Is(err, services.ErrAppCodeRequired),
		errors.Is(err, services.ErrAppNameRequired),
		errors.Is(err, services.ErrInvalidAppCode):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
	case errors.Is(err, services.ErrAppCodeExists):
		response.Error(c, http.StatusConflict, response.CodeConflict, "应用编码已存在")
	default:
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "应用中心服务暂时不可用")
	}
}
