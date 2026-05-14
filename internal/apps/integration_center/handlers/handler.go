package handlers

import (
	"context"
	"net/http"

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

func (h *Handler) Workspace(c *gin.Context) {
	h.section(c, h.service.Workspace)
}

func (h *Handler) TenantConnections(c *gin.Context) {
	h.section(c, h.service.TenantConnections)
}

func (h *Handler) SyncMonitor(c *gin.Context) {
	h.section(c, h.service.SyncMonitor)
}

func (h *Handler) Quota(c *gin.Context) {
	h.section(c, h.service.Quota)
}

func (h *Handler) Alerts(c *gin.Context) {
	h.section(c, h.service.Alerts)
}

func (h *Handler) Logs(c *gin.Context) {
	h.section(c, h.service.Logs)
}

func (h *Handler) section(c *gin.Context, load func(context.Context) (services.SectionSummary, error)) {
	result, err := load(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "第三方集成中心数据加载失败")
		return
	}
	response.OK(c, result)
}
