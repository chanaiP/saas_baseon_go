package handlers

import (
	"errors"
	"io"
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

func (h *AppHandler) Update(c *gin.Context) {
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
	var req dto.AppUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.UpdateApp(c.Request.Context(), userID, id, req)
	if err != nil {
		writeAppError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AppHandler) UpdateStatus(c *gin.Context) {
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
	var req dto.AppStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.UpdateAppStatus(c.Request.Context(), userID, id, req)
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

func (h *AppHandler) ManifestTemplate(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.service.CheckPlatformAccess(c.Request.Context(), userID); err != nil {
		writeAppError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="app.manifest.yaml"`)
	c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", []byte(h.service.StandardManifestTemplate()))
}

func (h *AppHandler) ParseManifest(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}

	fileName := ""
	var content []byte
	if file, err := c.FormFile("file"); err == nil {
		opened, openErr := file.Open()
		if openErr != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "Manifest 文件读取失败")
			return
		}
		defer opened.Close()
		raw, readErr := io.ReadAll(opened)
		if readErr != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "Manifest 文件读取失败")
			return
		}
		fileName = file.Filename
		content = raw
	} else {
		var req dto.ManifestParseRequest
		if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请上传 Manifest 文件或提供 content")
			return
		}
		fileName = req.FileName
		content = []byte(req.Content)
	}

	result, err := h.service.ParseManifestContent(c.Request.Context(), userID, fileName, content)
	if err != nil {
		writeAppError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AppHandler) ScanManifests(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req dto.ManifestScanRequest
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	result, err := h.service.ScanManifests(c.Request.Context(), userID, req.Root)
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
		errors.Is(err, services.ErrInvalidAppCode),
		errors.Is(err, services.ErrInvalidAppStatus),
		errors.Is(err, services.ErrBuiltinStatusImmutable),
		errors.Is(err, services.ErrClientCodeRequired),
		errors.Is(err, services.ErrManifestContentRequired),
		errors.Is(err, services.ErrManifestInvalidFormat):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
	case errors.Is(err, services.ErrAppCodeExists):
		response.Error(c, http.StatusConflict, response.CodeConflict, "应用编码已存在")
	default:
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "应用中心服务暂时不可用")
	}
}
