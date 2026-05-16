package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) PlatformDetail(c *gin.Context) {
	result, err := h.service.PlatformDetail(c.Request.Context(), currentUserID(c), c.Param("code"))
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) CreatePlatform(c *gin.Context) {
	var req services.PlatformMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreatePlatform(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.UpdatePlatform(c.Request.Context(), currentUserID(c), requestMeta(c), c.Param("code"), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.CreateProviderApp(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.UpdateProviderApp(c.Request.Context(), currentUserID(c), requestMeta(c), c.Param("code"), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ProviderAppDetail(c *gin.Context) {
	result, err := h.service.ProviderAppDetail(c.Request.Context(), currentUserID(c), c.Param("code"))
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) RotateProviderAppCredential(c *gin.Context) {
	var req services.ProviderAppCredentialRotateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.RotateProviderAppCredential(c.Request.Context(), currentUserID(c), requestMeta(c), c.Param("code"), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) CreatePlatformCapability(c *gin.Context) {
	var req services.PlatformCapabilityMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreatePlatformCapability(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdatePlatformCapability(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	var req services.PlatformCapabilityMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.UpdatePlatformCapability(c.Request.Context(), currentUserID(c), requestMeta(c), id, req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) DisablePlatformCapability(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.DisablePlatformCapability(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.UpdateAppCapability(c.Request.Context(), currentUserID(c), requestMeta(c), id, req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Workspace(c *gin.Context) {
	h.section(c, h.service.Workspace)
}

func (h *Handler) PlatformCapabilities(c *gin.Context) {
	h.section(c, h.service.PlatformCapabilities)
}

func (h *Handler) AppCapabilities(c *gin.Context) {
	h.section(c, h.service.AppCapabilities)
}

func (h *Handler) TenantConnections(c *gin.Context) {
	h.section(c, h.service.TenantConnections)
}

func (h *Handler) TenantConnectionDetail(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.TenantConnectionDetail(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) CreateTenantConnection(c *gin.Context) {
	var req services.TenantConnectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreateTenantConnection(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) StartOAuthAuthorization(c *gin.Context) {
	var req services.OAuthStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.StartOAuthAuthorization(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) OAuthCallback(c *gin.Context) {
	result, err := h.service.HandleOAuthCallback(c.Request.Context(), services.OAuthCallbackRequest{
		ProviderAppCode: c.Param("provider_app_code"),
		State:           c.Query("state"),
		Code:            c.Query("code"),
		Error:           c.Query("error"),
	})
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) RefreshTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.RefreshTenantConnection(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) PauseTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SetTenantConnectionStatus(c.Request.Context(), currentUserID(c), requestMeta(c), id, true)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ResumeTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SetTenantConnectionStatus(c.Request.Context(), currentUserID(c), requestMeta(c), id, false)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) RetryTenantConnection(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.RetryTenantConnection(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) SyncMonitor(c *gin.Context) {
	h.section(c, h.service.SyncMonitor)
}

func (h *Handler) SyncJobDetail(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SyncJobDetail(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) RetrySyncJob(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.RetrySyncJob(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) PauseSyncJob(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.PauseSyncJob(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ResumeSyncJob(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ResumeSyncJob(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Quota(c *gin.Context) {
	h.section(c, h.service.Quota)
}

func (h *Handler) QuotaUsages(c *gin.Context) {
	h.section(c, h.service.QuotaUsages)
}

func (h *Handler) CreateQuotaPolicy(c *gin.Context) {
	var req services.QuotaPolicyMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CreateQuotaPolicy(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.UpdateQuotaPolicy(c.Request.Context(), currentUserID(c), requestMeta(c), c.Param("code"), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.SetQuotaPolicyStatus(c.Request.Context(), currentUserID(c), requestMeta(c), c.Param("code"), req.Enabled)
	if err != nil {
		writeServiceError(c, err, err.Error())
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
	result, err := h.service.ProcessAlert(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ResolveAlert(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ResolveAlert(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) IgnoreAlert(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.IgnoreAlert(c.Request.Context(), currentUserID(c), requestMeta(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) Logs(c *gin.Context) {
	h.section(c, h.service.Logs)
}

func (h *Handler) LogDetail(c *gin.Context) {
	id, ok := parseUintID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.LogDetail(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) CheckConnectivity(c *gin.Context) {
	var req services.ConnectivityCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.CheckConnectivity(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "连通性检测失败")
		return
	}
	response.OK(c, result)
}

func (h *Handler) InvokeGateway(c *gin.Context) {
	var req services.GatewayInvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	result, err := h.service.InvokeGateway(c.Request.Context(), currentUserID(c), requestMeta(c), req)
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) ExportLogs(c *gin.Context) {
	result, err := h.service.ExportLogs(c.Request.Context(), currentUserID(c), requestMeta(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "调用日志导出失败")
		return
	}
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
	c.Header("X-Integration-Export-Rows", strconv.Itoa(result.RowCount))
	c.Data(http.StatusOK, result.ContentType, result.Content)
}

func (h *Handler) ReceiveWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "Webhook payload 读取失败")
		return
	}
	result, err := h.service.ReceiveWebhook(c.Request.Context(), services.WebhookReceiveRequest{
		ProviderAppCode: c.Param("provider_app_code"),
		Timestamp:       c.GetHeader("X-Integration-Timestamp"),
		Signature:       c.GetHeader("X-Integration-Signature"),
		IdempotencyKey:  c.GetHeader("X-Integration-Idempotency-Key"),
		EventType:       c.GetHeader("X-Integration-Event-Type"),
		Body:            body,
	})
	if err != nil {
		writeServiceError(c, err, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) section(c *gin.Context, load func(context.Context, uint64, services.PageRequest) (services.SectionSummary, error)) {
	result, err := load(c.Request.Context(), currentUserID(c), pageRequest(c))
	if err != nil {
		writeServiceError(c, err, "第三方集成中心数据加载失败")
		return
	}
	response.OK(c, result)
}

func currentUserID(c *gin.Context) uint64 {
	if value, ok := c.Get("user_id"); ok {
		switch v := value.(type) {
		case uint64:
			return v
		case uint:
			return uint64(v)
		case int:
			if v > 0 {
				return uint64(v)
			}
		case string:
			id, _ := strconv.ParseUint(v, 10, 64)
			return id
		}
	}
	return 0
}

func requestMeta(c *gin.Context) services.RequestMeta {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = c.GetHeader("X-Request-ID")
	}
	traceID := c.GetHeader("X-Trace-ID")
	if traceID == "" {
		traceID = requestID
	}
	return services.RequestMeta{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		RequestID: requestID,
		TraceID:   traceID,
	}
}

func pageRequest(c *gin.Context) services.PageRequest {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var tenantID *uint64
	if raw := c.Query("tenant_id"); raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
			tenantID = &id
		}
	}
	startTime := parseQueryTime(c.Query("start_time"))
	endTime := parseQueryTime(c.Query("end_time"))
	return services.PageRequest{
		Skip:            skip,
		Limit:           limit,
		Keyword:         c.Query("keyword"),
		Status:          c.Query("status"),
		PlatformCode:    c.Query("platform_code"),
		ProviderAppCode: c.Query("provider_app_code"),
		TenantID:        tenantID,
		StartTime:       startTime,
		EndTime:         endTime,
		SortBy:          c.Query("sort_by"),
		SortOrder:       c.Query("sort_order"),
	}
}

func parseQueryTime(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return &parsed
	}
	if parsed, err := time.Parse("2006-01-02", raw); err == nil {
		return &parsed
	}
	return nil
}

func writeServiceError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, services.ErrUnauthorized):
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, err.Error())
	case errors.Is(err, services.ErrForbidden):
		response.Error(c, http.StatusForbidden, response.CodeForbidden, err.Error())
	case fallback == err.Error():
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, fallback)
	}
}

func parseUintID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "资源 ID 不合法")
		return 0, false
	}
	return id, true
}
