package handlers

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) recordLogin(c *gin.Context, account string, userID *uint64, tenantID *uint64, success bool, message string) {
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	_ = h.db.Create(&models.LoginLog{
		TenantID:  tenantID,
		UserID:    userID,
		Account:   account,
		Success:   success,
		Message:   &message,
		IP:        &ip,
		UserAgent: &userAgent,
	}).Error
}

func (h *IdentityHandler) audit(c *gin.Context, tenantID uint64, userID uint64, module, action, summary string, detail interface{}) {
	detailJSON := ""
	if detail != nil {
		if raw, err := json.Marshal(maskAuditDetail(detail)); err == nil {
			detailJSON = string(raw)
		}
	}
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := c.GetString("request_id")
	appCode := auditAppCode(module)
	_ = h.db.Create(&models.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		AppCode:   nullableFromString(appCode),
		Module:    module,
		Action:    action,
		Summary:   summary,
		Detail:    nullableFromString(detailJSON),
		IP:        &ip,
		UserAgent: &userAgent,
		RequestID: nullableFromString(requestID),
		Result:    "success",
	}).Error
}

func (h *IdentityHandler) auditCurrentUser(c *gin.Context, module, action, summary string, detail interface{}) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	h.audit(c, user.TenantID, user.ID, module, action, summary, detail)
}

func maskAuditDetail(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for key, item := range typed {
			out[key] = maskAuditField(key, item)
		}
		return out
	case gin.H:
		out := gin.H{}
		for key, item := range typed {
			out[key] = maskAuditField(key, item)
		}
		return out
	default:
		return value
	}
}

func maskAuditField(key string, value interface{}) interface{} {
	if sensitiveAuditKey(key) {
		return "***"
	}
	return maskAuditDetail(value)
}

func sensitiveAuditKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, part := range []string{"password", "token", "secret", "captcha", "authorization", "phone", "email"} {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}

type auditRequiredEvent struct {
	Module string
	Action string
}

func requiredAuditEvents() []auditRequiredEvent {
	return []auditRequiredEvent{
		{Module: "tenant", Action: "create"},
		{Module: "tenant", Action: "create_with_package"},
		{Module: "tenant", Action: "update"},
		{Module: "tenant", Action: "status_update"},
		{Module: "tenant", Action: "delete"},
		{Module: "user", Action: "create"},
		{Module: "user", Action: "update"},
		{Module: "user", Action: "password_reset"},
		{Module: "user", Action: "delete"},
		{Module: "role", Action: "create"},
		{Module: "role", Action: "update"},
		{Module: "role", Action: "delete"},
		{Module: "organization", Action: "create"},
		{Module: "organization", Action: "update"},
		{Module: "organization", Action: "delete"},
		{Module: "business_unit", Action: "create"},
		{Module: "business_unit", Action: "update"},
		{Module: "business_unit", Action: "delete"},
		{Module: "business_unit", Action: "org_mapping_create"},
		{Module: "business_unit", Action: "org_mapping_delete"},
	}
}
