package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/apps/integration_center/services"
	"saas_baseon_go/internal/interfaces/http/response"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Overview(c *gin.Context) {
	result, err := h.service.Overview(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "第三方集成中心总览加载失败")
		return
	}
	response.OK(c, result)
}

func (h *Handler) Connectors(c *gin.Context) {
	result, err := h.service.Connectors(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "连接器列表加载失败")
		return
	}
	response.OK(c, result)
}

func (h *Handler) Platforms(c *gin.Context) {
	h.section(c, h.service.Platforms)
}

func (h *Handler) CreatePlatform(c *gin.Context) {
	var req services.PlatformMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreatePlatform(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdatePlatform(c *gin.Context) {
	var req services.PlatformMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.UpdatePlatform(c.Request.Context(), c.Param("code"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) CreateProviderApp(c *gin.Context) {
	var req services.ProviderAppMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreateProviderApp(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdateProviderApp(c *gin.Context) {
	var req services.ProviderAppMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.UpdateProviderApp(c.Request.Context(), c.Param("code"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdateAppCapability(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	var req services.AppCapabilityPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.UpdateAppCapability(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Workspace(c *gin.Context) {
	h.section(c, h.service.Workspace)
}

func (h *Handler) TenantConnections(c *gin.Context) {
	h.section(c, h.service.TenantConnections)
}

func (h *Handler) RefreshTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.RefreshTenantConnection(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) PauseTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SetTenantConnectionStatus(c.Request.Context(), id, true)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ResumeTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SetTenantConnectionStatus(c.Request.Context(), id, false)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) RetryTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.RetryTenantConnection(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) SyncMonitor(c *gin.Context) {
	h.section(c, h.service.SyncMonitor)
}

func (h *Handler) RetrySyncJob(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.RetrySyncJob(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) PauseSyncJob(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.PauseSyncJob(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ResumeSyncJob(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ResumeSyncJob(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Quota(c *gin.Context) {
	h.section(c, h.service.Quota)
}

func (h *Handler) CreateQuotaPolicy(c *gin.Context) {
	var req services.QuotaPolicyMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreateQuotaPolicy(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdateQuotaPolicy(c *gin.Context) {
	var req services.QuotaPolicyMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.UpdateQuotaPolicy(c.Request.Context(), c.Param("code"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdateQuotaPolicyStatus(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.SetQuotaPolicyStatus(c.Request.Context(), c.Param("code"), req.Enabled)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Alerts(c *gin.Context) {
	h.section(c, h.service.Alerts)
}

func (h *Handler) ProcessAlert(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ProcessAlert(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ResolveAlert(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ResolveAlert(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) IgnoreAlert(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.IgnoreAlert(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Logs(c *gin.Context) {
	h.section(c, h.service.Logs)
}

func (h *Handler) CheckConnectivity(c *gin.Context) {
	var req services.ConnectivityCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CheckConnectivity(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "连通性检测失败")
		return
	}
	response.OK(c, result)
}

func (h *Handler) ExportLogs(c *gin.Context) {
	result, err := h.service.ExportLogs(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "调用日志导出任务创建失败")
		return
	}
	response.OK(c, result)
}

func (h *Handler) section(c *gin.Context, load func(context.Context) (services.SectionSummary, error)) {
	result, err := load(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "第三方集成中心数据加载失败")
		return
	}
	response.OK(c, result)
}

func parseUintID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "资源 ID 不合法")
		return 0, false
	}
	return id, true
}
