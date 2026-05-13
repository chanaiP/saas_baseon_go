package handlers

import (
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) tenantName(tenantID *uint64) *string {
	if tenantID == nil {
		return nil
	}
	var tenant models.Tenant
	if err := h.db.First(&tenant, *tenantID).Error; err != nil {
		return nil
	}
	return &tenant.Name
}

func (h *IdentityHandler) userIdentity(userID *uint64) (*string, *string, *string) {
	if userID == nil {
		return nil, nil, nil
	}
	var user models.AppUser
	if err := h.db.First(&user, *userID).Error; err != nil {
		return nil, nil, nil
	}
	return &user.Name, &user.Account, &user.EmployeeNo
}

func (h *IdentityHandler) appName(appCode string) *string {
	if appCode == "" {
		return nil
	}
	var app models.SysApp
	if err := h.db.Where("app_code = ? AND deleted_at IS NULL", appCode).First(&app).Error; err != nil {
		return nil
	}
	return &app.AppName
}

func (h *IdentityHandler) permissionScopeTenantIDs(tenantID uint64) []uint64 {
	ids := []uint64{}
	seen := map[uint64]struct{}{}
	if tenantID != 0 {
		ids = append(ids, tenantID)
		seen[tenantID] = struct{}{}
	}
	var platformIDs []uint64
	_ = h.db.Model(&models.Tenant{}).
		Where("is_platform_tenant = ? AND deleted_at IS NULL", true).
		Order("id asc").
		Pluck("id", &platformIDs).Error
	for _, id := range platformIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		ids = append(ids, id)
		seen[id] = struct{}{}
	}
	if len(ids) == 0 {
		return []uint64{tenantID}
	}
	return ids
}

func allDevPermissionCodes() []string {
	return []string{
		"tenant:create", "tenant:edit", "tenant:delete", "tenant:reset_password", "tenant:quota_config",
		"plan:create", "plan:edit", "plan:delete", "plan:config",
		"org:create", "org:edit", "org:delete",
		"pos:create", "pos:edit", "pos:delete",
		"business_unit:create", "business_unit:edit", "business_unit:delete",
		"user:create", "user:edit", "user:reset_password", "user:delete",
		"role:create", "role:edit", "role:delete", "role:permission",
		"menu:create", "menu:edit", "menu:delete", "menu:package_feature",
		"dict_type:create", "dict_type:edit", "dict_type:delete",
		"dict_item:create", "dict_item:edit", "dict_item:delete",
		"param:create", "param:edit", "param:delete",
		"brand:edit",
	}
}
