package handlers

import (
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *MockIdentityHandler) currentQuotaUsage(tenantID uint64, quotaCode string) int {
	var count int64
	switch quotaCode {
	case "max_users":
		h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND status = ?", tenantID, 1).Count(&count)
	case "max_companies":
		h.db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ?", tenantID, "company", 1).Count(&count)
	case "max_business_units":
		h.db.Model(&models.BusinessUnit{}).Where("tenant_id = ? AND status = ?", tenantID, 1).Count(&count)
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

func (h *MockIdentityHandler) currentQuotaLimit(tenantID uint64, quotaID uint64) int {
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

func (h *MockIdentityHandler) tenantFeatureAllowed(tenantID uint64, featureCode string) bool {
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
	return count > 0
}
