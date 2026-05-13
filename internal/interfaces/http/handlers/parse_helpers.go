package handlers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func splitCSVParam(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		items = append(items, item)
	}
	return items
}

func safeDBErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
		return "数据已存在，请检查唯一字段"
	}
	return "请求处理失败"
}

func boolToStatus(value bool) int {
	if value {
		return 1
	}
	return 0
}

func parseUintParam(c *gin.Context, name string) uint64 {
	value, _ := strconv.ParseUint(c.Param(name), 10, 64)
	return value
}

func parseTenantID(c *gin.Context) uint64 {
	return parseTenantIDValue(c.Query("tenant_id"))
}

func parseTenantIDValue(raw string) uint64 {
	if raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return 1
}

func effectiveTenantID(queryTenantID string, userTenantID uint64, isPlatformAdmin bool) uint64 {
	if isPlatformAdmin {
		return parseTenantIDValue(queryTenantID)
	}
	if userTenantID > 0 {
		return userTenantID
	}
	return parseTenantIDValue(queryTenantID)
}

func (h *IdentityHandler) requestTenantID(c *gin.Context) uint64 {
	user, ok := h.currentUser(c)
	if !ok {
		return parseTenantID(c)
	}
	tenantContext := h.tenantContextForUser(user, c.Query("tenant_id"))
	c.Set("tenant_context", tenantContext)
	h.auditCrossTenantAccess(c, user, tenantContext)
	return tenantContext.TargetTenantID
}

func (h *IdentityHandler) auditCrossTenantAccess(c *gin.Context, user models.AppUser, tenantContext TenantContext) {
	if h.db == nil || !tenantContext.CrossTenantOperator || c.GetBool("cross_tenant_audit_recorded") {
		return
	}
	c.Set("cross_tenant_audit_recorded", true)
	requestID, _ := c.Get("request_id")
	method := ""
	if c.Request != nil {
		method = c.Request.Method
	}
	detail, _ := json.Marshal(gin.H{
		"actor_tenant_id":  tenantContext.ActorTenantID,
		"target_tenant_id": tenantContext.TargetTenantID,
		"actor_kind":       tenantContext.ActorKind,
		"method":           method,
		"path":             c.FullPath(),
		"request_id":       requestID,
	})
	text := string(detail)
	now := time.Now()
	_ = h.db.Create(&models.AuditLog{
		TenantID:  &tenantContext.ActorTenantID,
		UserID:    &user.ID,
		AppCode:   nullableFromString("system-management"),
		Module:    "tenant",
		Action:    "cross_tenant_access",
		Summary:   "跨租户访问",
		Detail:    &text,
		Result:    "success",
		CreatedAt: now,
	}).Error
}
