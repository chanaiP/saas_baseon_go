package handlers

import (
	"encoding/json"
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return h.currentQuotaLimitUsing(h.db, tenantID, quotaID)
}

func (h *IdentityHandler) currentQuotaLimitUsing(db *gorm.DB, tenantID uint64, quotaID uint64) int {
	if !h.subscriptionAllowsLoginUsing(db, tenantID) {
		return 0
	}
	now := time.Now()
	var override models.TenantQuotaOverride
	if err := db.Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quotaID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue
	}
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err == nil {
		var planQuota models.SaasPlanQuota
		if err := db.Where("plan_id = ? AND quota_id = ?", sub.PlanID, quotaID).First(&planQuota).Error; err == nil {
			return planQuota.QuotaValue
		}
	}
	return 0
}

func (h *IdentityHandler) currentQuotaLimitByCode(tenantID uint64, quotaCode string) (int, bool) {
	var quota models.SaasQuota
	if err := h.db.Where("quota_code = ? AND status = ?", quotaCode, 1).First(&quota).Error; err != nil {
		return 0, false
	}
	return h.currentQuotaLimit(tenantID, quota.ID), true
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

func (h *IdentityHandler) tenantAllowedFeatureCodeSet(tenantID uint64) map[string]bool {
	allowed := map[string]bool{}
	if !h.subscriptionAllowsLogin(tenantID) {
		return allowed
	}
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return allowed
	}
	var features []models.SaasFeature
	_ = h.db.Where("status = ?", 1).Find(&features).Error
	byID := make(map[uint64]models.SaasFeature, len(features))
	for _, feature := range features {
		byID[feature.ID] = feature
	}
	var links []models.SaasPlanFeature
	_ = h.db.Where("plan_id = ? AND enabled = ?", sub.PlanID, true).Find(&links).Error
	explicit := map[uint64]struct{}{}
	for _, link := range links {
		explicit[link.FeatureID] = struct{}{}
		if feature, ok := byID[link.FeatureID]; ok {
			allowed[feature.FeatureCode] = true
		}
	}
	for _, feature := range features {
		if feature.FeatureType != "BUTTON" || feature.ParentID == 0 {
			continue
		}
		if _, ok := explicit[feature.ID]; ok {
			continue
		}
		parent, ok := byID[feature.ParentID]
		if ok && allowed[parent.FeatureCode] {
			allowed[feature.FeatureCode] = true
		}
	}
	now := time.Now()
	var overrides []models.TenantFeatureOverride
	_ = h.db.Where("tenant_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, now, now).Order("id asc").Find(&overrides).Error
	for _, override := range overrides {
		if feature, ok := byID[override.FeatureID]; ok {
			allowed[feature.FeatureCode] = override.Enabled
		}
	}
	return allowed
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

func (h *IdentityHandler) consumeQuota(tenantID uint64, quotaCode string, increment int) error {
	if increment <= 0 {
		increment = 1
	}
	return h.db.Transaction(func(tx *gorm.DB) error {
		var quota models.SaasQuota
		if err := tx.Where("quota_code = ? AND status = ?", quotaCode, 1).First(&quota).Error; err != nil {
			return nil
		}
		periodKey := "TOTAL"
		if quota.PeriodType != nil && *quota.PeriodType == "DAY" {
			periodKey = time.Now().Format("20060102")
		}
		limit := h.currentQuotaLimitUsing(tx, tenantID, quota.ID)
		if limit >= 0 && increment > limit {
			return &quotaExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: 0}
		}
		now := time.Now()
		usage := models.TenantQuotaUsage{
			TenantID:        tenantID,
			QuotaCode:       quotaCode,
			UsedValue:       increment,
			LimitValue:      limit,
			PeriodType:      quota.PeriodType,
			PeriodKey:       periodKey,
			LastRefreshTime: &now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		result := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tenant_id"}, {Name: "quota_code"}, {Name: "period_key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"used_value":        gorm.Expr("tenant_quota_usage.used_value + ?", increment),
				"limit_value":       limit,
				"last_refresh_time": now,
				"updated_at":        now,
			}),
			Where: clause.Where{Exprs: []clause.Expression{
				gorm.Expr("? < 0 OR tenant_quota_usage.used_value + ? <= ?", limit, increment, limit),
			}},
		}).Create(&usage)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			used := h.currentQuotaUsageUsing(tx, tenantID, quotaCode, periodKey)
			return &quotaExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: used}
		}
		return recordQuotaAudit(tx, tenantID, quotaCode, increment, limit, periodKey, now)
	})
}

func (h *IdentityHandler) currentQuotaUsageUsing(db *gorm.DB, tenantID uint64, quotaCode string, periodKey string) int {
	var usage models.TenantQuotaUsage
	if err := db.Where("tenant_id = ? AND quota_code = ? AND period_key = ?", tenantID, quotaCode, periodKey).First(&usage).Error; err == nil {
		return usage.UsedValue
	}
	return 0
}

func recordQuotaAudit(db *gorm.DB, tenantID uint64, quotaCode string, increment int, limit int, periodKey string, now time.Time) error {
	detail, _ := json.Marshal(map[string]interface{}{
		"quota_code": quotaCode,
		"increment":  increment,
		"limit":      limit,
		"period_key": periodKey,
	})
	return db.Create(&models.AuditLog{
		TenantID:  &tenantID,
		Module:    "quota",
		Action:    "consume",
		Summary:   "配额扣减",
		Detail:    nullableFromString(string(detail)),
		Result:    "success",
		CreatedAt: now,
	}).Error
}

type quotaExceededError struct {
	QuotaName string
	Limit     int
	Used      int
}

func (e *quotaExceededError) Error() string {
	return e.QuotaName + "已超出套餐配额"
}
