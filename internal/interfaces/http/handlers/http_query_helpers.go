package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func paginated(items interface{}) gin.H {
	total := 0
	switch v := items.(type) {
	case []gin.H:
		total = len(v)
	}
	return gin.H{"items": items, "total": total, "skip": 0, "limit": 50}
}

func paginatedWithTotal(items interface{}, total int64, skip int, limit int) gin.H {
	return gin.H{"items": items, "total": total, "skip": skip, "limit": limit}
}

func paginationParams(c *gin.Context) (int, int) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if skip < 0 {
		skip = 0
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return skip, limit
}

func applyDateRange(query *gorm.DB, column string, dateFrom string, dateTo string) *gorm.DB {
	if start, ok := parseDateOnly(dateFrom); ok {
		query = query.Where(column+" >= ?", start)
	}
	if end, ok := parseDateOnly(dateTo); ok {
		query = query.Where(column+" < ?", end.Add(24*time.Hour))
	}
	return query
}

func parseDateOnly(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func loginLogToJSON(row models.LoginLog, tenantName *string) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "tenant_name": tenantName, "user_id": row.UserID, "account": row.Account, "success": row.Success, "message": row.Message, "ip": row.IP, "created_at": row.CreatedAt}
}

func auditLogToJSON(row models.AuditLog, tenantName *string, userName *string, userAccount *string, userEmployeeNo *string, appCode string, appName *string) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "tenant_name": tenantName, "user_id": row.UserID, "user_name": userName, "user_account": userAccount, "user_employee_no": userEmployeeNo, "app_code": appCode, "app_name": appName, "module": row.Module, "action": row.Action, "summary": row.Summary, "detail": row.Detail, "ip": row.IP, "user_agent": row.UserAgent, "request_id": row.RequestID, "result": row.Result, "created_at": row.CreatedAt}
}

func parseOptionalUintQuery(c *gin.Context, key string) *uint64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

func tenantToJSON(db *gorm.DB, row models.Tenant) gin.H {
	planName, planCode := (*string)(nil), (*string)(nil)
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", row.ID).Order("id desc").First(&sub).Error; err == nil {
		var plan models.SaasPlan
		if err := db.First(&plan, sub.PlanID).Error; err == nil {
			planName = &plan.PlanName
			planCode = &plan.PlanCode
		}
	}
	var companyCount int64
	var storeCount int64
	var userCount int64
	var buCount int64
	db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ? AND deleted_at IS NULL", row.ID, "company", 1).Count(&companyCount)
	db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ? AND deleted_at IS NULL", row.ID, "store", 1).Count(&storeCount)
	db.Model(&models.AppUser{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", row.ID, 1).Count(&userCount)
	db.Model(&models.BusinessUnit{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", row.ID, 1).Count(&buCount)
	contactName, contactPhone := tenantContact(db, row)
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "status": row.Status, "start_date": coalesceTimePtr(row.StartDate, timePtr(sub.StartTime)), "expire_date": coalesceTimePtr(row.ExpireDate, sub.EndTime), "max_companies": tenantQuotaValue(db, row.ID, "max_companies", row.MaxCompanies), "max_stores": tenantQuotaValue(db, row.ID, "max_stores", 0), "max_business_units": tenantQuotaValue(db, row.ID, "max_business_units", 0), "max_users": tenantQuotaValue(db, row.ID, "max_users", row.MaxUsers), "used_companies": companyCount, "used_stores": storeCount, "used_users": userCount, "used_business_units": buCount, "plan_name": planName, "plan_code": planCode, "contact_name": contactName, "contact_phone": contactPhone, "company_count": companyCount, "brand_display_name": row.BrandName, "brand_logo_data": row.LogoData, "created_at": row.CreatedAt}
}

func coalesceTimePtr(primary *time.Time, fallback *time.Time) *time.Time {
	if primary != nil {
		return primary
	}
	return fallback
}

func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func tenantQuotaValue(db *gorm.DB, tenantID uint64, quotaCode string, fallback int) int {
	var quota models.SaasQuota
	if err := db.Where("quota_code = ?", quotaCode).First(&quota).Error; err != nil {
		return fallback
	}
	now := time.Now()
	var override models.TenantQuotaOverride
	if err := db.Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quota.ID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue
	}
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return fallback
	}
	var planQuota models.SaasPlanQuota
	if err := db.Where("plan_id = ? AND quota_id = ?", sub.PlanID, quota.ID).First(&planQuota).Error; err == nil {
		return planQuota.QuotaValue
	}
	return fallback
}

func quotaCheckReason(limit int, used int, increment int) string {
	if limit < 0 || used+increment <= limit {
		return "允许使用"
	}
	return "已超出套餐配额"
}

func quotaCodeForOrgNodeType(nodeType string) string {
	switch nodeType {
	case "company":
		return "max_companies"
	case "store":
		return "max_stores"
	case "department", "warehouse", "project_team":
		return "max_departments"
	default:
		return ""
	}
}

func tenantContact(db *gorm.DB, tenant models.Tenant) (*string, *string) {
	name := tenant.ContactName
	phone := tenant.ContactPhone
	if name != nil && strings.TrimSpace(*name) == "" {
		name = nil
	}
	if phone != nil && strings.TrimSpace(*phone) == "" {
		phone = nil
	}
	if name != nil && phone != nil {
		return name, phone
	}
	if admin, ok := primaryAdminForTenant(db, tenant.ID); ok {
		if name == nil && strings.TrimSpace(admin.Name) != "" {
			name = &admin.Name
		}
		if phone == nil {
			phone = admin.Phone
		}
	}
	return name, phone
}

func primaryAdminForTenant(db *gorm.DB, tenantID uint64) (models.AppUser, bool) {
	var user models.AppUser
	if err := db.
		Where("tenant_id = ? AND is_tenant_admin = ? AND deleted_at IS NULL", tenantID, true).
		Order("app_user.id asc").
		First(&user).Error; err == nil {
		return user, true
	}
	if err := db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Order("id asc").First(&user).Error; err == nil {
		return user, true
	}
	return models.AppUser{}, false
}
