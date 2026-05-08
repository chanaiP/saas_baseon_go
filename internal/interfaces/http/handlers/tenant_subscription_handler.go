package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) TenantQuotaRecords(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var orgs []models.OrgNode
	var bus []models.BusinessUnit
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&orgs).Error
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&bus).Error
	companies, stores := []gin.H{}, []gin.H{}
	for _, org := range orgs {
		item := gin.H{"id": org.ID, "code": org.Code, "name": org.Name, "node_type": org.NodeType, "company_name": nil, "status": org.Status}
		if org.NodeType == "company" {
			companies = append(companies, item)
		}
		if org.NodeType == "store" {
			stores = append(stores, item)
		}
	}
	businessUnits := make([]gin.H, 0, len(bus))
	for _, bu := range bus {
		businessUnits = append(businessUnits, businessUnitToJSON(bu))
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "companies": companies, "stores": stores, "business_units": businessUnits})
}

func (h *IdentityHandler) TenantCompanies(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND node_type = ?", tenantID, "company").Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, orgNodeToJSON(row, []gin.H{}))
	}
	response.OK(c, gin.H{
		"tenant_id":          tenantID,
		"billing_unit_count": len(items),
		"companies":          items,
	})
}

func (h *IdentityHandler) TenantPrimaryAdmin(c *gin.Context) {
	user, ok := h.primaryAdmin(parseUintParam(c, "id"))
	if !ok {
		response.Error(c, 404, response.CodeNotFound, "主管理员不存在")
		return
	}
	response.OK(c, gin.H{"employee_no": user.EmployeeNo, "phone": user.Phone, "name": user.Name})
}

func (h *IdentityHandler) ResetTenantPrimaryAdminPassword(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	if message, blocked := h.rateLimitExceeded(c, fmt.Sprintf("rate:reset_tenant_admin_password:ip:%s", c.ClientIP()), 10, 10*time.Minute); blocked {
		c.JSON(429, response.Body{Code: 42900, Message: message})
		return
	}
	if message, blocked := h.rateLimitExceeded(c, fmt.Sprintf("rate:reset_tenant_admin_password:tenant:%d", tenantID), 5, 10*time.Minute); blocked {
		c.JSON(429, response.Body{Code: 42900, Message: message})
		return
	}
	user, ok := h.primaryAdmin(tenantID)
	if !ok {
		response.Error(c, 404, response.CodeNotFound, "主管理员不存在")
		return
	}
	newPassword := generateRandomPassword(14)
	_ = h.db.Model(&user).Updates(map[string]interface{}{
		"password_hash":       devPasswordHash(newPassword),
		"session_version":     gorm.Expr("session_version + 1"),
		"password_changed_at": time.Now(),
	}).Error
	h.invalidateSessionsForUser(user.ID)
	response.OK(c, gin.H{"employee_no": user.EmployeeNo, "phone": user.Phone, "name": user.Name, "new_password": newPassword})
}

func (h *IdentityHandler) TenantSubscription(c *gin.Context) {
	sub, ok := h.findTenantSubscription(parseUintParam(c, "id"))
	if !ok {
		response.OK(c, nil)
		return
	}
	response.OK(c, tenantSubscriptionToJSON(sub))
}

func (h *IdentityHandler) SaveTenantSubscription(c *gin.Context) {
	user, _ := h.currentUser(c)
	var body tenantPackagePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.saveTenantPackage(parseUintParam(c, "id"), body); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateTenantAuthorizationCache(parseUintParam(c, "id"))
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "tenant_subscription", "update", "保存租户订阅", gin.H{"tenant_id": parseUintParam(c, "id"), "plan_id": body.PlanID})
	}
	h.TenantSubscription(c)
}

func (h *IdentityHandler) SaveTenantPackageConfig(c *gin.Context) {
	user, _ := h.currentUser(c)
	var body tenantPackagePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tenantID := parseUintParam(c, "id")
	if err := h.saveTenantPackage(tenantID, body); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateTenantAuthorizationCache(tenantID)
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "tenant_subscription", "config", "保存租户套餐配置", gin.H{"tenant_id": tenantID, "plan_id": body.PlanID})
	}
	sub, _ := h.findTenantSubscription(tenantID)
	response.OK(c, gin.H{"tenant_id": tenantID, "subscription": tenantSubscriptionToJSON(sub), "quotas": h.tenantQuotaOverridesPayload(tenantID)})
}

func (h *IdentityHandler) TenantFeatureOverrides(c *gin.Context) {
	response.OK(c, h.tenantFeatureOverridesPayload(parseUintParam(c, "id")))
}

func (h *IdentityHandler) SaveTenantFeatureOverrides(c *gin.Context) {
	user, _ := h.currentUser(c)
	tenantID := parseUintParam(c, "id")
	var body struct {
		Overrides []struct {
			FeatureID uint64  `json:"feature_id"`
			Enabled   bool    `json:"enabled"`
			Reason    *string `json:"reason"`
		} `json:"overrides"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := lockTenantForUpdate(tx, tenantID); err != nil {
			return err
		}
		if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.TenantFeatureOverride{}).Error; err != nil {
			return err
		}
		for _, item := range body.Overrides {
			if err := tx.Create(&models.TenantFeatureOverride{TenantID: tenantID, FeatureID: item.FeatureID, Enabled: item.Enabled, Reason: item.Reason}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateTenantAuthorizationCache(tenantID)
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "tenant_feature_override", "update", "保存租户功能覆盖", gin.H{"tenant_id": tenantID, "count": len(body.Overrides)})
	}
	response.OK(c, h.tenantFeatureOverridesPayload(tenantID))
}

func (h *IdentityHandler) TenantQuotaOverrides(c *gin.Context) {
	response.OK(c, h.tenantQuotaOverridesPayload(parseUintParam(c, "id")))
}

func (h *IdentityHandler) SaveTenantQuotaOverrides(c *gin.Context) {
	user, _ := h.currentUser(c)
	tenantID := parseUintParam(c, "id")
	var body struct {
		Overrides []struct {
			QuotaID    uint64  `json:"quota_id"`
			QuotaValue int     `json:"quota_value"`
			Reason     *string `json:"reason"`
		} `json:"overrides"`
		Quotas []struct {
			QuotaID    uint64 `json:"quota_id"`
			QuotaValue int    `json:"quota_value"`
		} `json:"quotas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := lockTenantForUpdate(tx, tenantID); err != nil {
			return err
		}
		if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.TenantQuotaOverride{}).Error; err != nil {
			return err
		}
		for _, item := range body.Overrides {
			if err := tx.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: item.QuotaID, QuotaValue: item.QuotaValue, Reason: item.Reason}).Error; err != nil {
				return err
			}
		}
		for _, item := range body.Quotas {
			if err := tx.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: item.QuotaID, QuotaValue: item.QuotaValue}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateTenantAuthorizationCache(tenantID)
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "tenant_quota_override", "update", "保存租户配额覆盖", gin.H{"tenant_id": tenantID, "override_count": len(body.Overrides), "quota_count": len(body.Quotas)})
	}
	response.OK(c, h.tenantQuotaOverridesPayload(tenantID))
}

func (h *IdentityHandler) TenantQuotaUsage(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var quotas []models.SaasQuota
	_ = h.db.Find(&quotas).Error
	usages := make([]gin.H, 0, len(quotas))
	for _, quota := range quotas {
		used := h.currentQuotaUsage(tenantID, quota.QuotaCode)
		limit := h.currentQuotaLimit(tenantID, quota.ID)
		remaining := limit - used
		if limit < 0 {
			remaining = -1
		}
		usages = append(usages, gin.H{"quota_id": quota.ID, "quota_code": quota.QuotaCode, "quota_name": quota.QuotaName, "quota_type": quota.QuotaType, "period_type": quota.PeriodType, "period_key": "current", "unit": quota.Unit, "used_value": used, "limit_value": limit, "remaining_value": remaining})
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "usages": usages})
}

func (h *IdentityHandler) TenantFeatureAccess(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	featureCode := c.Param("feature_code")
	response.OK(c, gin.H{"tenant_id": tenantID, "feature_code": featureCode, "allowed": h.tenantFeatureAllowed(tenantID, featureCode)})
}

func (h *IdentityHandler) TenantQuotaCheck(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	quotaCode := c.Param("quota_code")
	increment := 1
	if raw := strings.TrimSpace(c.Query("increment")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			increment = parsed
		}
	}
	var quota models.SaasQuota
	if err := h.db.Where("quota_code = ?", quotaCode).First(&quota).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "配额不存在")
		return
	}
	used := h.currentQuotaUsage(tenantID, quotaCode)
	limit := h.currentQuotaLimit(tenantID, quota.ID)
	remaining := limit - used
	if limit < 0 {
		remaining = -1
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "quota_code": quotaCode, "used_value": used, "limit_value": limit, "remaining_value": remaining, "allowed": limit < 0 || used+increment <= limit, "reason": quotaCheckReason(limit, used, increment)})
}
