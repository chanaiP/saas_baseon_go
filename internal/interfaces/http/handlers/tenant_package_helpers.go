package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) createTenantWithAdmin(code, name string, status int, adminName, adminEmployeeNo string, adminPhone *string, adminPassword string) (models.Tenant, error) {
	var tenant models.Tenant
	err := h.db.Transaction(func(tx *gorm.DB) error {
		created, err := h.createTenantWithAdminOnDB(tx, code, name, status, adminName, adminEmployeeNo, adminPhone, adminPassword)
		if err != nil {
			return err
		}
		tenant = created
		return nil
	})
	return tenant, err
}

func (h *IdentityHandler) createTenantWithAdminOnDB(db *gorm.DB, code, name string, status int, adminName, adminEmployeeNo string, adminPhone *string, adminPassword string) (models.Tenant, error) {
	if status == 0 {
		status = 1
	}
	phone, msg := normalizeOptionalPhone(adminPhone)
	if msg != "" {
		return models.Tenant{}, errors.New(strings.Replace(msg, "手机号", "管理员手机号", 1))
	}
	var tenant models.Tenant
	tenant = models.Tenant{Code: strings.TrimSpace(code), Name: strings.TrimSpace(name), Status: status, ContactName: nullableFromString(adminName), ContactPhone: phone}
	if err := db.Create(&tenant).Error; err != nil {
		return tenant, err
	}
	companyCode := strings.ToUpper(strings.TrimSpace(code))
	companyType := "GROUP"
	company := models.OrgNode{TenantID: tenant.ID, NodeType: "company", Name: tenant.Name, Code: &companyCode, CompanyType: &companyType, Status: 1}
	if err := db.Create(&company).Error; err != nil {
		return tenant, err
	}
	roleDescription := "超级管理员（系统自动创建）"
	role := models.Role{TenantID: tenant.ID, Code: "admin", Name: "超级管理员", Description: &roleDescription, Status: 1}
	if err := db.Create(&role).Error; err != nil {
		return tenant, err
	}
	user := models.AppUser{TenantID: tenant.ID, CompanyID: &company.ID, EmployeeNo: strings.TrimSpace(adminEmployeeNo), Account: strings.TrimSpace(adminEmployeeNo), PasswordHash: devPasswordHash(adminPassword), Name: strings.TrimSpace(adminName), Phone: phone, Status: 1, IsPlatformAdmin: false}
	if err := db.Create(&user).Error; err != nil {
		return tenant, err
	}
	if err := db.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		return tenant, err
	}
	var permissions []models.Permission
	if err := db.Where("tenant_id = ? OR tenant_id = ?", tenant.ID, 1).Find(&permissions).Error; err == nil {
		for _, permission := range permissions {
			if err := db.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).FirstOrCreate(&models.RolePermission{RoleID: role.ID, PermissionID: permission.ID}).Error; err != nil {
				return tenant, err
			}
		}
	}
	return tenant, nil
}

func parseTimePtr(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, *value); err == nil {
			return &parsed
		}
	}
	return nil
}

func (h *IdentityHandler) saveTenantPackage(tenantID uint64, body tenantPackagePayload) error {
	return h.saveTenantPackageOnDB(h.db, tenantID, body)
}

func (h *IdentityHandler) saveTenantPackageOnDB(db *gorm.DB, tenantID uint64, body tenantPackagePayload) error {
	start := time.Now()
	if body.StartTime != "" {
		if parsed := parseTimePtr(&body.StartTime); parsed != nil {
			start = *parsed
		}
	}
	if body.SubscriptionStatus == "" {
		body.SubscriptionStatus = "ACTIVE"
	}
	if body.PlanID == 0 {
		return errors.New("套餐不存在")
	}
	var plan models.SaasPlan
	if err := db.Where("id = ? AND deleted_at IS NULL AND status = ?", body.PlanID, 1).First(&plan).Error; err != nil {
		return errors.New("套餐不存在或已停用")
	}
	var existing models.TenantSubscription
	sub := models.TenantSubscription{TenantID: tenantID, PlanID: body.PlanID, SubscriptionStatus: body.SubscriptionStatus, StartTime: start, EndTime: parseTimePtr(body.EndTime), TrialEndTime: parseTimePtr(body.TrialEndTime), AutoRenew: body.AutoRenew, FrozenReason: body.FrozenReason}
	if err := db.Where("tenant_id = ?", tenantID).First(&existing).Error; err == nil {
		if err := db.Model(&existing).Updates(sub).Error; err != nil {
			return err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&sub).Error; err != nil {
			return err
		}
	} else {
		return err
	}
	for _, quota := range body.Quotas {
		_ = db.Where("tenant_id = ? AND quota_id = ?", tenantID, quota.QuotaID).Delete(&models.TenantQuotaOverride{}).Error
		if err := db.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue, Reason: stringPtrLocal("主体套餐配置")}).Error; err != nil {
			return err
		}
	}
	return nil
}

func stringPtrLocal(value string) *string {
	return &value
}

func (h *IdentityHandler) findTenantSubscription(tenantID uint64) (models.TenantSubscription, bool) {
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return sub, false
	}
	return sub, true
}

func tenantSubscriptionToJSON(row models.TenantSubscription) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "plan_id": row.PlanID, "subscription_status": row.SubscriptionStatus, "start_time": row.StartTime, "end_time": row.EndTime, "trial_end_time": row.TrialEndTime, "auto_renew": row.AutoRenew, "frozen_reason": row.FrozenReason, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func (h *IdentityHandler) primaryAdmin(tenantID uint64) (models.AppUser, bool) {
	var user models.AppUser
	err := h.db.Where("tenant_id = ?", tenantID).Order("is_platform_admin desc, id asc").First(&user).Error
	return user, err == nil
}

func (h *IdentityHandler) tenantFeatureOverridesPayload(tenantID uint64) gin.H {
	var rows []models.TenantFeatureOverride
	_ = h.db.Where("tenant_id = ?", tenantID).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var feature models.SaasFeature
		_ = h.db.First(&feature, row.FeatureID).Error
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "feature_id": row.FeatureID, "feature_code": feature.FeatureCode, "feature_name": feature.FeatureName, "enabled": row.Enabled, "reason": row.Reason, "start_time": row.StartTime, "end_time": row.EndTime})
	}
	return gin.H{"tenant_id": tenantID, "overrides": items}
}

func (h *IdentityHandler) tenantQuotaOverridesPayload(tenantID uint64) gin.H {
	var rows []models.TenantQuotaOverride
	_ = h.db.Where("tenant_id = ?", tenantID).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var quota models.SaasQuota
		_ = h.db.First(&quota, row.QuotaID).Error
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "quota_id": row.QuotaID, "quota_code": quota.QuotaCode, "quota_name": quota.QuotaName, "quota_value": row.QuotaValue, "period_type": quota.PeriodType, "unit": quota.Unit, "reason": row.Reason, "start_time": row.StartTime, "end_time": row.EndTime})
	}
	return gin.H{"tenant_id": tenantID, "overrides": items}
}
