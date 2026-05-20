package handlers

import (
	"errors"
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) userHasSuperAdminScope(user models.AppUser) bool {
	return user.IsPlatformAdmin || h.userIsTenantAdmin(user)
}

func (h *IdentityHandler) userIsTenantAdmin(user models.AppUser) bool {
	if user.ID == 0 || user.IsPlatformAdmin {
		return false
	}
	if user.IsTenantAdmin {
		return true
	}
	first, ok := h.firstTenantAdminUser(user.TenantID)
	return ok && first.ID == user.ID
}

func (h *IdentityHandler) firstTenantAdminUser(tenantID uint64) (models.AppUser, bool) {
	var user models.AppUser
	err := h.db.
		Where("tenant_id = ? AND is_tenant_admin = ? AND deleted_at IS NULL", tenantID, true).
		Order("id asc").
		First(&user).Error
	if err == nil {
		return user, true
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AppUser{}, false
	}
	err = h.db.
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("id asc").
		First(&user).Error
	return user, err == nil
}

func (h *IdentityHandler) firstPlatformAdminUser() (models.AppUser, bool) {
	var user models.AppUser
	err := h.db.
		Where("is_platform_admin = ? AND deleted_at IS NULL", true).
		Order("id asc").
		First(&user).Error
	return user, err == nil
}

func (h *IdentityHandler) userIsInitialSuperAdmin(user models.AppUser) bool {
	if user.ID == 0 {
		return false
	}
	if user.IsPlatformAdmin {
		first, ok := h.firstPlatformAdminUser()
		return ok && first.ID == user.ID
	}
	if h.userIsTenantAdmin(user) {
		first, ok := h.firstTenantAdminUser(user.TenantID)
		return ok && first.ID == user.ID
	}
	return false
}

func (h *IdentityHandler) allPermissionCodesForAdmin(user models.AppUser, filterSubscription bool) []string {
	query := h.db.Model(&models.Permission{}).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Where("perm_type IN ?", []int{2, 3})
	if user.IsPlatformAdmin {
		query = query.Where("tenant_id = ?", user.TenantID)
	} else {
		query = query.Where("tenant_id IN ? AND is_platform_only = ?", h.permissionScopeTenantIDs(user.TenantID), false)
	}
	var permissions []models.Permission
	_ = query.Order("id asc").Find(&permissions).Error

	allowedFeatures := map[string]bool(nil)
	if filterSubscription {
		allowedFeatures = h.tenantAllowedFeatureCodeSet(user.TenantID)
	}

	seen := map[string]struct{}{}
	codes := make([]string, 0, len(permissions)+1)
	for _, permission := range permissions {
		if permission.Path == "" || permission.Path == "__operations_root__" || permission.Path == "__menu_root__" {
			continue
		}
		if isPureViewPermissionPath(permission.Path) {
			continue
		}
		if filterSubscription && !permissionAllowedByFeatureCodeSet(permission, allowedFeatures) {
			continue
		}
		if !h.permissionAllowedForTenantType(user, permission) {
			continue
		}
		if _, ok := seen[permission.Path]; ok {
			continue
		}
		seen[permission.Path] = struct{}{}
		codes = append(codes, permission.Path)
	}
	if _, ok := seen["/home"]; !ok {
		codes = append(codes, "/home")
	}
	sort.Strings(codes)
	return codes
}

func systemRoleLocked(role models.Role) bool {
	return role.Code == "admin"
}

func systemRoleState(role models.Role) gin.H {
	locked := systemRoleLocked(role)
	return gin.H{
		"is_system":       locked,
		"is_locked":       locked,
		"can_edit":        !locked,
		"can_delete":      !locked,
		"can_config_perm": !locked,
	}
}
