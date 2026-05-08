package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
	return err.Error()
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
	return effectiveTenantID(c.Query("tenant_id"), user.TenantID, user.IsPlatformAdmin)
}
