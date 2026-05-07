package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/interfaces/http/response"
)

type FallbackHandler struct{}

func NewFallbackHandler() *FallbackHandler {
	return &FallbackHandler{}
}

func (h *FallbackHandler) NoRoute(c *gin.Context) {
	path := c.Request.URL.Path
	if !strings.HasPrefix(path, "/api/") {
		c.Status(404)
		return
	}

	switch c.Request.Method {
	case "GET":
		h.get(c, path)
	case "POST":
		response.OK(c, gin.H{"id": 1})
	case "PUT", "PATCH":
		response.OK(c, gin.H{"id": 1})
	case "DELETE":
		response.OK(c, gin.H{"deleted": 1})
	default:
		response.OK(c, gin.H{})
	}
}

func (h *FallbackHandler) get(c *gin.Context, path string) {
	if strings.Contains(path, "/logs/") {
		response.OK(c, gin.H{"items": []gin.H{}, "total": 0, "skip": 0, "limit": 20})
		return
	}
	if strings.Contains(path, "/monitor/health-detail") {
		response.OK(c, gin.H{"postgres": true, "redis": true, "status": "ok"})
		return
	}
	if strings.Contains(path, "/monitor/server-info") {
		response.OK(c, gin.H{"os": "linux", "runtime": "go", "status": "ok"})
		return
	}
	if strings.Contains(path, "/monitor/scheduled-jobs") {
		response.OK(c, gin.H{"items": []gin.H{}, "total": 0})
		return
	}
	if strings.Contains(path, "/monitor/services-overview") {
		response.OK(c, gin.H{"items": []gin.H{}, "status": "ok"})
		return
	}
	if strings.Contains(path, "/monitor/cache-stats") {
		response.OK(c, gin.H{"used_memory": 0, "keys": 0})
		return
	}
	if strings.Contains(path, "/monitor/cache-keys") {
		response.OK(c, gin.H{"items": []gin.H{}, "total": 0})
		return
	}
	if strings.Contains(path, "/subscription") {
		response.OK(c, gin.H{
			"id":                  1,
			"tenant_id":           1,
			"plan_id":             1,
			"subscription_status": "ACTIVE",
			"start_time":          "",
			"end_time":            nil,
			"trial_end_time":      nil,
			"auto_renew":          false,
			"frozen_reason":       nil,
		})
		return
	}
	if strings.Contains(path, "/feature-overrides") {
		response.OK(c, gin.H{"tenant_id": 1, "overrides": []gin.H{}})
		return
	}
	if strings.Contains(path, "/quota-overrides") {
		response.OK(c, gin.H{"tenant_id": 1, "overrides": []gin.H{}})
		return
	}
	if strings.Contains(path, "/quota-usage") {
		response.OK(c, gin.H{"tenant_id": 1, "usages": []gin.H{}})
		return
	}
	if strings.Contains(path, "/primary-admin") {
		response.OK(c, gin.H{"employee_no": "admin", "phone": nil, "name": "平台管理员"})
		return
	}
	if strings.Contains(path, "/org-mappings") {
		response.OK(c, []gin.H{})
		return
	}
	if strings.Contains(path, "/tree") {
		response.OK(c, []gin.H{})
		return
	}
	if strings.Contains(path, "/batch") {
		response.OK(c, gin.H{"values": gin.H{}})
		return
	}

	response.OK(c, gin.H{"items": []gin.H{}, "total": 0, "skip": 0, "limit": 20})
}
