package handlers

import (
	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) userToJSON(row models.AppUser) gin.H {
	var roles []models.UserRole
	var positions []models.AppUserPosition
	var departments []models.AppUserDepartment
	_ = h.db.Where("user_id = ?", row.ID).Find(&roles).Error
	_ = h.db.Where("user_id = ?", row.ID).Find(&positions).Error
	_ = h.db.Where("user_id = ?", row.ID).Find(&departments).Error
	roleIDs := make([]uint64, 0, len(roles))
	positionIDs := make([]uint64, 0, len(positions))
	departmentIDs := make([]uint64, 0, len(departments))
	for _, item := range roles {
		roleIDs = append(roleIDs, item.RoleID)
	}
	for _, item := range positions {
		positionIDs = append(positionIDs, item.PositionID)
	}
	for _, item := range departments {
		departmentIDs = append(departmentIDs, item.DepartmentID)
	}
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "employee_no": row.EmployeeNo, "phone": row.Phone, "name": row.Name, "email": row.Email, "avatar_url": row.AvatarURL, "status": row.Status, "company_id": row.CompanyID, "department_id": row.DepartmentID, "department_ids": departmentIDs, "position_ids": positionIDs, "role_ids": roleIDs, "is_platform_admin": row.IsPlatformAdmin, "created_at": row.CreatedAt}
}
