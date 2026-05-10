package handlers

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func devPasswordHash(password string) string {
	return mustHashPassword(password)
}

func (h *IdentityHandler) replaceUserRelations(userID uint64, roleIDs []uint64, positionIDs []uint64, departmentIDs []uint64) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		return replaceUserRelationsTx(tx, userID, roleIDs, positionIDs, departmentIDs)
	})
}

func replaceUserRelationsTx(tx *gorm.DB, userID uint64, roleIDs []uint64, positionIDs []uint64, departmentIDs []uint64) error {
	if roleIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		for _, id := range roleIDs {
			if err := tx.Create(&models.UserRole{UserID: userID, RoleID: id}).Error; err != nil {
				return err
			}
		}
	}
	if positionIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.AppUserPosition{}).Error; err != nil {
			return err
		}
		for _, id := range positionIDs {
			if err := tx.Create(&models.AppUserPosition{UserID: userID, PositionID: id}).Error; err != nil {
				return err
			}
		}
	}
	if departmentIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.AppUserDepartment{}).Error; err != nil {
			return err
		}
		for _, id := range departmentIDs {
			if err := tx.Create(&models.AppUserDepartment{UserID: userID, DepartmentID: id}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *IdentityHandler) assertDepartmentsInTenant(tenantID uint64, departmentIDs []uint64) error {
	ids := uniqueUint64s(departmentIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.OrgNode{}).
		Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL AND node_type IN ?", tenantID, ids, []string{"department", "store", "warehouse", "project_team", "group"}).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("组织节点不存在、类型不可用于人员归属或不属于当前主体")
	}
	return nil
}

func (h *IdentityHandler) companyIDForDepartment(tenantID uint64, departmentID uint64) *uint64 {
	var row models.OrgNode
	if err := h.db.Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, departmentID).First(&row).Error; err != nil {
		return nil
	}
	if row.CompanyID != nil {
		return row.CompanyID
	}
	if row.NodeType == "company" {
		return &row.ID
	}
	return nil
}

func (h *IdentityHandler) rootCompanyIDForTenant(tenantID uint64) *uint64 {
	var row models.OrgNode
	if err := h.db.
		Where("tenant_id = ? AND node_type = ? AND parent_id IS NULL AND deleted_at IS NULL", tenantID, "company").
		Order("id asc").
		First(&row).Error; err != nil {
		return nil
	}
	return &row.ID
}

func (h *IdentityHandler) assertPositionsInTenant(tenantID uint64, positionIDs []uint64) error {
	ids := uniqueUint64s(positionIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.Position{}).Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("岗位不存在或不属于当前主体")
	}
	return nil
}

func (h *IdentityHandler) assertRolesInTenant(tenantID uint64, roleIDs []uint64) error {
	ids := uniqueUint64s(roleIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.Role{}).Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("角色不存在或不属于当前主体")
	}
	return nil
}

func (h *IdentityHandler) assertPositionTypeInTenant(tenantID uint64, positionTypeID uint64) error {
	var count int64
	if err := h.db.Model(&models.PositionType{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", positionTypeID, tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("岗位类型不存在")
	}
	return nil
}

func (h *IdentityHandler) assertUserUniqueFields(tenantID uint64, exceptUserID uint64, employeeNo string, phone *string) error {
	if strings.TrimSpace(employeeNo) != "" {
		query := h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND employee_no = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(employeeNo))
		if exceptUserID > 0 {
			query = query.Where("id <> ?", exceptUserID)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("工号已存在")
		}
	}
	if phone != nil && strings.TrimSpace(*phone) != "" {
		query := h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND phone = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(*phone))
		if exceptUserID > 0 {
			query = query.Where("id <> ?", exceptUserID)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("手机号已存在")
		}
	}
	return nil
}
