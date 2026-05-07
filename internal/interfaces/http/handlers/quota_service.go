package handlers

import (
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) currentQuotaUsage(tenantID uint64, quotaCode string) int {
	var count int64
	switch quotaCode {
	case "max_users":
		h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", tenantID, 1).Count(&count)
	case "max_companies":
		h.db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ? AND deleted_at IS NULL", tenantID, "company", 1).Count(&count)
	case "max_stores":
		h.db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ? AND deleted_at IS NULL", tenantID, "store", 1).Count(&count)
	case "max_departments":
		h.db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type IN ? AND status = ? AND deleted_at IS NULL", tenantID, []string{"department", "warehouse", "project_team"}, 1).Count(&count)
	case "max_roles":
		h.db.Model(&models.Role{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", tenantID, 1).Count(&count)
	case "max_business_units":
		h.db.Model(&models.BusinessUnit{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", tenantID, 1).Count(&count)
	case "daily_import_times", "daily_export_times":
		period := time.Now().Format("20060102")
		var usage models.TenantQuotaUsage
		if err := h.db.Where("tenant_id = ? AND quota_code = ? AND period_key = ?", tenantID, quotaCode, period).First(&usage).Error; err == nil {
			return usage.UsedValue
		}
	default:
		var usage models.TenantQuotaUsage
		if err := h.db.Where("tenant_id = ? AND quota_code = ?", tenantID, quotaCode).Order("id desc").First(&usage).Error; err == nil {
			return usage.UsedValue
		}
	}
	return int(count)
}

func (h *IdentityHandler) currentQuotaLimit(tenantID uint64, quotaID uint64) int {
	if !h.subscriptionAllowsLogin(tenantID) {
		return 0
	}
	now := time.Now()
	var override models.TenantQuotaOverride
	if err := h.db.Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quotaID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue
	}
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err == nil {
		var planQuota models.SaasPlanQuota
		if err := h.db.Where("plan_id = ? AND quota_id = ?", sub.PlanID, quotaID).First(&planQuota).Error; err == nil {
			return planQuota.QuotaValue
		}
	}
	return 0
}

func (h *IdentityHandler) tenantFeatureAllowed(tenantID uint64, featureCode string) bool {
	if !h.subscriptionAllowsLogin(tenantID) {
		return false
	}
	var feature models.SaasFeature
	if err := h.db.Where("feature_code = ? AND status = ?", featureCode, 1).First(&feature).Error; err != nil {
		return false
	}
	now := time.Now()
	var override models.TenantFeatureOverride
	if err := h.db.Where("tenant_id = ? AND feature_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, feature.ID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.Enabled
	}
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return false
	}
	var count int64
	h.db.Model(&models.SaasPlanFeature{}).Where("plan_id = ? AND feature_id = ? AND enabled = ?", sub.PlanID, feature.ID, true).Count(&count)
	if count > 0 {
		return true
	}
	if feature.FeatureType == "BUTTON" && feature.ParentID > 0 {
		var explicit int64
		h.db.Model(&models.SaasPlanFeature{}).Where("plan_id = ? AND feature_id = ?", sub.PlanID, feature.ID).Count(&explicit)
		if explicit == 0 {
			var parent models.SaasFeature
			if err := h.db.Where("id = ? AND status = ?", feature.ParentID, 1).First(&parent).Error; err == nil {
				return h.tenantFeatureAllowed(tenantID, parent.FeatureCode)
			}
		}
	}
	return false
}

func (h *IdentityHandler) requireQuotaAvailable(tenantID uint64, quotaCode string, increment int) error {
	if increment <= 0 {
		increment = 1
	}
	var quota models.SaasQuota
	if err := h.db.Where("quota_code = ? AND status = ?", quotaCode, 1).First(&quota).Error; err != nil {
		return nil
	}
	limit := h.currentQuotaLimit(tenantID, quota.ID)
	if limit < 0 {
		return nil
	}
	used := h.currentQuotaUsage(tenantID, quotaCode)
	if used+increment > limit {
		return &quotaExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: used}
	}
	return nil
}

type quotaExceededError struct {
	QuotaName string
	Limit     int
	Used      int
}

func (e *quotaExceededError) Error() string {
	return e.QuotaName + "已超出套餐配额"
}
