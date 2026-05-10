package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) tenantCapabilityContext(tenantID uint64) tenantCapabilityProfile {
	empty := tenantCapabilityProfile{Subscription: nil, Features: []string{}, Quotas: gin.H{}}
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", tenantID).First(&tenant).Error; err != nil || tenant.Status != 1 {
		return empty
	}
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil || !h.subscriptionAllowsLogin(tenantID) {
		return empty
	}
	var plan models.SaasPlan
	if err := h.db.Where("id = ?", sub.PlanID).First(&plan).Error; err != nil {
		return empty
	}
	now := time.Now()
	featureIDs := map[uint64]bool{}
	var planFeatures []models.SaasPlanFeature
	_ = h.db.Where("plan_id = ? AND enabled = ?", plan.ID, true).Find(&planFeatures).Error
	for _, row := range planFeatures {
		featureIDs[row.FeatureID] = true
	}
	var overrides []models.TenantFeatureOverride
	_ = h.db.Where("tenant_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, now, now).Find(&overrides).Error
	for _, override := range overrides {
		featureIDs[override.FeatureID] = override.Enabled
	}
	ids := make([]uint64, 0, len(featureIDs))
	for id, enabled := range featureIDs {
		if enabled {
			ids = append(ids, id)
		}
	}
	features := []string{}
	if len(ids) > 0 {
		var rows []models.SaasFeature
		_ = h.db.Where("id IN ? AND status = ?", ids, 1).Order("feature_code asc").Find(&rows).Error
		for _, row := range rows {
			if uncontrolledPackageFeatureCode(row.FeatureCode) || reservedPackageFeatureCode(row.FeatureCode) {
				continue
			}
			features = append(features, row.FeatureCode)
		}
	}
	quotas := gin.H{}
	var planQuotas []models.SaasPlanQuota
	_ = h.db.Where("plan_id = ?", plan.ID).Find(&planQuotas).Error
	for _, row := range planQuotas {
		var quota models.SaasQuota
		if err := h.db.Where("id = ? AND status = ?", row.QuotaID, 1).First(&quota).Error; err == nil {
			if hiddenPackageQuotaCode(quota.QuotaCode) {
				continue
			}
			quotas[quota.QuotaCode] = row.QuotaValue
		}
	}
	var quotaOverrides []models.TenantQuotaOverride
	_ = h.db.Where("tenant_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, now, now).Find(&quotaOverrides).Error
	for _, override := range quotaOverrides {
		var quota models.SaasQuota
		if err := h.db.Where("id = ? AND status = ?", override.QuotaID, 1).First(&quota).Error; err == nil {
			if hiddenPackageQuotaCode(quota.QuotaCode) {
				continue
			}
			quotas[quota.QuotaCode] = override.QuotaValue
		}
	}
	return tenantCapabilityProfile{
		Subscription: gin.H{"plan_code": plan.PlanCode, "plan_name": plan.PlanName, "status": sub.SubscriptionStatus, "end_time": sub.EndTime},
		Features:     features,
		Quotas:       quotas,
	}
}

func (h *IdentityHandler) filterPermissionCodesForSubscription(codes []string, features []string) []string {
	enabled := map[string]struct{}{}
	for _, feature := range features {
		enabled[feature] = struct{}{}
	}
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		var permission models.Permission
		if err := h.db.Where("path = ?", code).First(&permission).Error; err != nil {
			out = append(out, code)
			continue
		}
		if !permission.IsPackageFeature || permission.FeatureCode == nil || *permission.FeatureCode == "" {
			out = append(out, code)
			continue
		}
		if _, ok := enabled[*permission.FeatureCode]; ok {
			out = append(out, code)
		}
	}
	return out
}

func (h *IdentityHandler) permissionCodesForUser(user models.AppUser, filterSubscription bool) []string {
	var permissions []models.Permission
	_ = h.db.
		Joins("JOIN role_permission rp ON rp.permission_id = permission.id").
		Joins("JOIN user_role ur ON ur.role_id = rp.role_id").
		Joins("JOIN role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.tenant_id = ? AND permission.enabled = ? AND permission.deleted_at IS NULL AND r.deleted_at IS NULL", user.ID, user.TenantID, true).
		Where("permission.perm_type IN ?", []int{2, 3}).
		Order("permission.id asc").
		Find(&permissions).Error
	codes := make([]string, 0, len(permissions)+2)
	seen := map[string]struct{}{}
	allowedFeatures := map[string]bool(nil)
	if filterSubscription {
		allowedFeatures = h.tenantAllowedFeatureCodeSet(user.TenantID)
	}
	for _, permission := range permissions {
		if permission.Path == "" || permission.Path == "__operations_root__" || permission.Path == "__menu_root__" {
			continue
		}
		if isPureViewPermissionPath(permission.Path) {
			continue
		}
		if !user.IsPlatformAdmin && !h.viewerHasPlatformScope(user) && permission.IsPlatformOnly {
			continue
		}
		if filterSubscription && !permissionAllowedByFeatureCodeSet(permission, allowedFeatures) {
			continue
		}
		if _, ok := seen[permission.Path]; ok {
			continue
		}
		seen[permission.Path] = struct{}{}
		codes = append(codes, permission.Path)
	}
	for _, fixed := range []string{"/home"} {
		if _, ok := seen[fixed]; ok {
			continue
		}
		seen[fixed] = struct{}{}
		codes = append(codes, fixed)
	}
	return codes
}

func (h *IdentityHandler) userShortcutIDs(userID uint64) []string {
	var pref models.UserPreference
	if err := h.db.Where("user_id = ? AND pref_key = ?", userID, "shortcut_ids").First(&pref).Error; err != nil || pref.PrefValue == nil {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(*pref.PrefValue), &ids); err != nil {
		return []string{}
	}
	return ids
}

func (h *IdentityHandler) saveUserShortcutIDs(userID uint64, ids []string) error {
	raw, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	value := string(raw)
	now := time.Now()
	result := h.db.Model(&models.UserPreference{}).
		Where("user_id = ? AND pref_key = ?", userID, "shortcut_ids").
		Updates(map[string]interface{}{"pref_value": value, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	return h.db.Create(&models.UserPreference{UserID: userID, PrefKey: "shortcut_ids", PrefValue: &value, UpdatedAt: now}).Error
}

func (h *IdentityHandler) viewerHasPlatformScope(user models.AppUser) bool {
	return user.IsPlatformAdmin
}

func (h *IdentityHandler) userCanEditTenantBranding(user models.AppUser) bool {
	if user.IsPlatformAdmin {
		return true
	}
	if !h.viewerHasPlatformScope(user) && !h.tenantFeatureAllowed(user.TenantID, "brand_config") {
		return false
	}
	var count int64
	_ = h.db.Model(&models.Permission{}).
		Joins("JOIN role_permission rp ON rp.permission_id = permission.id").
		Joins("JOIN user_role ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ? AND permission.path = ? AND permission.enabled = ?", user.ID, "brand:edit", true).
		Count(&count).Error
	if count > 0 {
		return true
	}
	_ = h.db.Model(&models.Role{}).
		Joins("JOIN user_role ur ON ur.role_id = role.id").
		Where("ur.user_id = ? AND role.tenant_id = ? AND role.code = ? AND role.status = ? AND role.deleted_at IS NULL", user.ID, user.TenantID, "admin", 1).
		Count(&count).Error
	return count > 0
}

func (h *IdentityHandler) invalidateSessionsForTenant(tenantID uint64) {
	if h.redis == nil || tenantID == 0 {
		return
	}
	ctx := context.Background()
	var cursor uint64
	for {
		keys, next, err := h.redis.Scan(ctx, cursor, "auth:session:*", 100).Result()
		if err != nil {
			return
		}
		for _, key := range keys {
			raw, err := h.redis.Get(ctx, key).Result()
			if err != nil || raw == "" {
				continue
			}
			var session struct {
				TenantID uint64 `json:"tenant_id"`
			}
			if json.Unmarshal([]byte(raw), &session) == nil && session.TenantID == tenantID {
				_ = h.redis.Del(ctx, key).Err()
			}
		}
		if next == 0 {
			return
		}
		cursor = next
	}
}

func (h *IdentityHandler) invalidateSessionsForUser(userID uint64) {
	if h.redis == nil || userID == 0 {
		return
	}
	ctx := context.Background()
	var cursor uint64
	for {
		keys, next, err := h.redis.Scan(ctx, cursor, "auth:session:*", 100).Result()
		if err != nil {
			return
		}
		for _, key := range keys {
			raw, err := h.redis.Get(ctx, key).Result()
			if err != nil || raw == "" {
				continue
			}
			var session struct {
				UserID uint64 `json:"user_id"`
			}
			if json.Unmarshal([]byte(raw), &session) == nil && session.UserID == userID {
				_ = h.redis.Del(ctx, key).Err()
			}
		}
		if next == 0 {
			return
		}
		cursor = next
	}
}
