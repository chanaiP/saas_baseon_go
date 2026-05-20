package handlers

import (
	"strings"

	"gorm.io/gorm/clause"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const (
	TenantTypePlatform   = "platform"
	TenantTypeEnterprise = "enterprise"
	TenantTypePersonal   = "personal"

	TenantScopePlatformOnly       = "platform_only"
	TenantScopeEnterpriseOnly     = "enterprise_only"
	TenantScopePersonalOnly       = "personal_only"
	TenantScopeAll                = "all"
	TenantScopePlatformEnterprise = "platform_enterprise"
	TenantScopeEnterprisePersonal = "enterprise_personal"
	TenantScopePlatformPersonal   = "platform_personal"
)

func normalizeTenantType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case TenantTypePlatform, TenantTypePersonal:
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return TenantTypeEnterprise
	}
}

func normalizeTenantRecordType(tenant models.Tenant) string {
	if strings.TrimSpace(tenant.TenantType) != "" {
		return normalizeTenantType(tenant.TenantType)
	}
	if tenant.IsPlatform {
		return TenantTypePlatform
	}
	return TenantTypeEnterprise
}

func normalizeTenantScope(value string, isPlatformOnly bool) string {
	if isPlatformOnly {
		return TenantScopePlatformOnly
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case TenantScopePlatformOnly, TenantScopeEnterpriseOnly, TenantScopePersonalOnly,
		TenantScopeAll, TenantScopePlatformEnterprise, TenantScopeEnterprisePersonal,
		TenantScopePlatformPersonal:
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return TenantScopeEnterpriseOnly
	}
}

func tenantScopeAllows(scope string, tenantType string) bool {
	tenantType = normalizeTenantType(tenantType)
	switch normalizeTenantScope(scope, false) {
	case TenantScopeAll:
		return true
	case TenantScopePlatformOnly:
		return tenantType == TenantTypePlatform
	case TenantScopeEnterpriseOnly:
		return tenantType == TenantTypeEnterprise
	case TenantScopePersonalOnly:
		return tenantType == TenantTypePersonal
	case TenantScopePlatformEnterprise:
		return tenantType == TenantTypePlatform || tenantType == TenantTypeEnterprise
	case TenantScopeEnterprisePersonal:
		return tenantType == TenantTypeEnterprise || tenantType == TenantTypePersonal
	case TenantScopePlatformPersonal:
		return tenantType == TenantTypePlatform || tenantType == TenantTypePersonal
	default:
		return tenantType == TenantTypeEnterprise
	}
}

func (h *IdentityHandler) tenantTypeForID(tenantID uint64) string {
	if h == nil || h.db == nil || tenantID == 0 {
		return TenantTypeEnterprise
	}
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", tenantID).First(&tenant).Error; err != nil {
		return TenantTypeEnterprise
	}
	return normalizeTenantRecordType(tenant)
}

func (h *IdentityHandler) permissionAllowedForTenantType(user models.AppUser, permission models.Permission) bool {
	if user.IsPlatformAdmin {
		return true
	}
	return tenantScopeAllows(normalizeTenantScope(permission.TenantScope, permission.IsPlatformOnly), h.tenantTypeForID(user.TenantID))
}

func (h *IdentityHandler) routeScopeAllowed(user models.AppUser, required string) bool {
	if required == "" {
		return true
	}
	if h == nil || h.db == nil {
		return true
	}
	var permission models.Permission
	if err := h.db.
		Where("tenant_id IN ? AND path = ? AND enabled = ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(user.TenantID), required, true).
		Order(clause.Expr{SQL: "tenant_id = ? DESC, id ASC", Vars: []interface{}{user.TenantID}, WithoutParentheses: true}).
		First(&permission).Error; err != nil {
		return true
	}
	return h.permissionAllowedForTenantType(user, permission)
}
