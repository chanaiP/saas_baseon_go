package handlers

import (
	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) userToJSON(row models.AppUser) gin.H {
	var roles []models.UserRole
	var positions []models.AppUserPosition
	var departments []models.AppUserDepartment
	_ = h.db.
		Joins("JOIN role r ON r.id = user_role.role_id AND r.deleted_at IS NULL AND r.code <> ?", "admin").
		Where("user_role.user_id = ?", row.ID).
		Find(&roles).Error
	_ = h.db.Where("user_id = ?", row.ID).Find(&positions).Error
	_ = h.db.Where("user_id = ?", row.ID).Find(&departments).Error
	roleIDs := make([]uint64, 0, len(roles))
	positionIDs := make([]uint64, 0, len(positions))
	departmentIDs := make([]uint64, 0, len(departments))
	departmentNames := make([]string, 0, len(departments))
	for _, item := range roles {
		roleIDs = append(roleIDs, item.RoleID)
	}
	for _, item := range positions {
		positionIDs = append(positionIDs, item.PositionID)
	}
	for _, item := range departments {
		departmentIDs = append(departmentIDs, item.DepartmentID)
	}
	if len(departmentIDs) > 0 {
		var orgNodes []models.OrgNode
		_ = h.db.Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", row.TenantID, departmentIDs).Find(&orgNodes).Error
		nameByID := make(map[uint64]string, len(orgNodes))
		for _, orgNode := range orgNodes {
			nameByID[orgNode.ID] = orgNode.Name
		}
		for _, id := range departmentIDs {
			if name := nameByID[id]; name != "" {
				departmentNames = append(departmentNames, name)
			}
		}
	}
	isInitialAdmin := h.userIsInitialSuperAdmin(row)
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "employee_no": row.EmployeeNo, "phone": row.Phone, "name": row.Name, "email": row.Email, "avatar_url": row.AvatarURL, "status": row.Status, "company_id": row.CompanyID, "department_id": row.DepartmentID, "department_ids": departmentIDs, "department_names": departmentNames, "position_ids": positionIDs, "role_ids": roleIDs, "is_platform_admin": row.IsPlatformAdmin, "is_tenant_admin": h.userIsTenantAdmin(row), "is_initial_admin": isInitialAdmin, "can_edit": !isInitialAdmin, "can_delete": !isInitialAdmin, "created_at": row.CreatedAt}
}
