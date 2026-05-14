package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/apps/ai_capability_center/services"
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
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) List(c *gin.Context) {
	resource := c.Param("resource")
	skip := parseInt(c, "skip", 0)
	limit := parseInt(c, "limit", 20)
	keyword := c.Query("keyword")
	var (
		result services.PageResult
		err    error
	)
	switch resource {
	case "providers":
		result, err = h.service.ListProviders(c.Request.Context(), skip, limit, keyword)
	case "accounts":
		result, err = h.service.ListAccounts(c.Request.Context(), skip, limit, c.Query("provider_id"))
	case "apis":
		result, err = h.service.ListAPIs(c.Request.Context(), skip, limit, c.Query("provider_id"), c.Query("account_id"))
	case "capabilities":
		result, err = h.service.ListCapabilities(c.Request.Context(), skip, limit, keyword)
	case "models":
		result, err = h.service.ListModels(c.Request.Context(), skip, limit, c.Query("provider_id"), keyword)
	case "price-policies":
		result, err = h.service.ListPricePolicies(c.Request.Context(), skip, limit, c.Query("model_id"))
	case "price-tiers":
		result, err = h.service.ListPriceTiers(c.Request.Context(), skip, limit, c.Query("price_policy_id"))
	case "base-routes":
		result, err = h.service.ListBaseRoutes(c.Request.Context(), skip, limit, keyword)
	case "route-models":
		result, err = h.service.ListRouteModels(c.Request.Context(), skip, limit, c.Query("base_route_id"))
	case "scenarios":
		result, err = h.service.ListScenarios(c.Request.Context(), skip, limit, keyword)
	case "tenant-strategies":
		result, err = h.service.ListTenantStrategies(c.Request.Context(), skip, limit, keyword)
	case "quota-rules":
		result, err = h.service.ListQuotaRules(c.Request.Context(), skip, limit, c.Query("policy_id"))
	case "rate-limit-rules":
		result, err = h.service.ListRateLimitRules(c.Request.Context(), skip, limit, c.Query("policy_id"))
	case "usage-records":
		result, err = h.service.ListUsageRecords(c.Request.Context(), skip, limit, keyword, c.Query("start_date"), c.Query("end_date"))
	case "settings":
		result, err = h.service.ListSettings(c.Request.Context(), skip, limit)
	default:
		writeError(c, services.ErrNotFound)
		return
	}
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.CreateResource(c.Request.Context(), userID, c.Param("resource"), payload)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) ImportProviders(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req services.ProviderImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.ImportProviders(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) ImportModels(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req services.ModelImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.ImportModels(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) ImportScenarios(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req services.ScenarioImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.ImportScenarios(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) ImportRoutes(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req services.RouteImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.ImportRoutes(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) ImportTenantStrategies(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var req services.TenantStrategyImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.ImportTenantStrategies(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) CheckProviderAPIConnectivity(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	filter := services.APIConnectivityFilter{
		ProviderID: c.Query("provider_id"),
		AccountID:  c.Query("account_id"),
	}
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		var body services.APIConnectivityFilter
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
			return
		}
		if body.ProviderID != "" {
			filter.ProviderID = body.ProviderID
		}
		if body.AccountID != "" {
			filter.AccountID = body.AccountID
		}
	}
	result, err := h.service.CheckProviderAPIConnectivity(c.Request.Context(), userID, filter)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.UpdateResource(c.Request.Context(), userID, c.Param("resource"), c.Param("id"), payload)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.service.DeleteResource(c.Request.Context(), userID, c.Param("resource"), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) Invoke(c *gin.Context) {
	var req services.InvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
		return
	}
	result, err := h.service.Invoke(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
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

func parseInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "资源不存在")
	case errors.Is(err, services.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数不合法")
	case errors.Is(err, services.ErrResourceInUse):
		response.Error(c, http.StatusConflict, response.CodeConflict, "资源已被引用，不能删除")
	default:
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "AI 能力中心服务暂时不可用")
	}
}
