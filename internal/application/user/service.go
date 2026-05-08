package user

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
)

type QuotaChecker interface {
	RequireAvailable(ctx context.Context, tenantID uint64, quotaCode string, increment int) error
}

type Service struct {
	db           *gorm.DB
	quotaChecker QuotaChecker
}

func NewService(db *gorm.DB, quotaChecker ...QuotaChecker) *Service {
	service := &Service{db: db}
	if len(quotaChecker) > 0 {
		service.quotaChecker = quotaChecker[0]
	}
	return service
}

type Relations struct {
	RoleIDs       []uint64
	PositionIDs   []uint64
	DepartmentIDs []uint64
}

type ImportResult struct {
	Created int
	Skipped int
}

func (s *Service) CreateWithRelations(ctx context.Context, row models.AppUser, rel Relations) (models.AppUser, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.validateCreate(tx, ctx, row, rel); err != nil {
			return err
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return replaceRelations(tx, row.ID, rel)
	})
	return row, err
}

func (s *Service) UpdateWithRelations(ctx context.Context, row *models.AppUser, updates map[string]interface{}, rel Relations) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateRelations(tx, row.TenantID, rel); err != nil {
			return err
		}
		if phone, ok := updates["phone"].(*string); ok {
			if err := assertUserUniqueFields(tx, ctx, row.TenantID, row.ID, "", phone); err != nil {
				return err
			}
		}
		if len(updates) > 0 {
			if err := tx.Model(row).Updates(updates).First(row, row.ID).Error; err != nil {
				return err
			}
		}
		return replaceRelations(tx, row.ID, rel)
	})
}

func (s *Service) ImportUsers(ctx context.Context, rows []models.AppUser) (ImportResult, error) {
	result := ImportResult{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seenEmployeeNo := map[string]struct{}{}
		for i := range rows {
			employeeNo := strings.TrimSpace(rows[i].EmployeeNo)
			if employeeNo == "" || strings.TrimSpace(rows[i].Name) == "" {
				return fmt.Errorf("工号和姓名不能为空")
			}
			if _, ok := seenEmployeeNo[employeeNo]; ok {
				return fmt.Errorf("导入文件中存在重复工号")
			}
			seenEmployeeNo[employeeNo] = struct{}{}
			exists, err := userExists(tx, ctx, rows[i].TenantID, employeeNo)
			if err != nil {
				return err
			}
			if exists {
				result.Skipped++
				continue
			}
			if err := assertUserUniqueFields(tx, ctx, rows[i].TenantID, 0, employeeNo, rows[i].Phone); err != nil {
				return err
			}
			if rows[i].Status == 1 && s.quotaChecker != nil {
				if err := s.quotaChecker.RequireAvailable(ctx, rows[i].TenantID, "max_users", 1); err != nil {
					return err
				}
			}
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
			result.Created++
		}
		return nil
	})
	return result, err
}

func (s *Service) Delete(ctx context.Context, row *models.AppUser, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(row).Updates(updates).Error
	})
}

func (s *Service) validateCreate(tx *gorm.DB, ctx context.Context, row models.AppUser, rel Relations) error {
	if strings.TrimSpace(row.EmployeeNo) == "" || strings.TrimSpace(row.Name) == "" {
		return fmt.Errorf("工号和姓名不能为空")
	}
	if err := validateRelations(tx, row.TenantID, rel); err != nil {
		return err
	}
	if err := assertUserUniqueFields(tx, ctx, row.TenantID, 0, row.EmployeeNo, row.Phone); err != nil {
		return err
	}
	if row.Status == 1 && s.quotaChecker != nil {
		return s.quotaChecker.RequireAvailable(ctx, row.TenantID, "max_users", 1)
	}
	return nil
}

func validateRelations(tx *gorm.DB, tenantID uint64, rel Relations) error {
	if err := assertDepartmentsInTenant(tx, tenantID, rel.DepartmentIDs); err != nil {
		return err
	}
	if err := assertPositionsInTenant(tx, tenantID, rel.PositionIDs); err != nil {
		return err
	}
	return assertRolesInTenant(tx, tenantID, rel.RoleIDs)
}

func assertDepartmentsInTenant(tx *gorm.DB, tenantID uint64, departmentIDs []uint64) error {
	ids := uniqueIDs(departmentIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.OrgNode{}).
		Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL AND node_type IN ?", tenantID, ids, []string{"department", "store", "warehouse", "project_team", "group"}).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("组织节点不存在、类型不可用于人员归属或不属于当前主体")
	}
	return nil
}

func assertPositionsInTenant(tx *gorm.DB, tenantID uint64, positionIDs []uint64) error {
	ids := uniqueIDs(positionIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.Position{}).Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("岗位不存在或不属于当前主体")
	}
	return nil
}

func assertRolesInTenant(tx *gorm.DB, tenantID uint64, roleIDs []uint64) error {
	ids := uniqueIDs(roleIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.Role{}).Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("角色不存在或不属于当前主体")
	}
	return nil
}

func assertUserUniqueFields(tx *gorm.DB, ctx context.Context, tenantID uint64, exceptUserID uint64, employeeNo string, phone *string) error {
	if strings.TrimSpace(employeeNo) != "" {
		query := repositories.NewTenantScopedRepository(tx, tenantID).Users(ctx).Where("employee_no = ?", strings.TrimSpace(employeeNo))
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
		query := repositories.NewTenantScopedRepository(tx, tenantID).Users(ctx).Where("phone = ?", strings.TrimSpace(*phone))
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

func userExists(tx *gorm.DB, ctx context.Context, tenantID uint64, employeeNo string) (bool, error) {
	var count int64
	if err := repositories.NewTenantScopedRepository(tx, tenantID).Users(ctx).Where("employee_no = ?", strings.TrimSpace(employeeNo)).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func uniqueIDs(values []uint64) []uint64 {
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func replaceRelations(tx *gorm.DB, userID uint64, rel Relations) error {
	if rel.RoleIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		for _, id := range rel.RoleIDs {
			if err := tx.Create(&models.UserRole{UserID: userID, RoleID: id}).Error; err != nil {
				return err
			}
		}
	}
	if rel.PositionIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.AppUserPosition{}).Error; err != nil {
			return err
		}
		for _, id := range rel.PositionIDs {
			if err := tx.Create(&models.AppUserPosition{UserID: userID, PositionID: id}).Error; err != nil {
				return err
			}
		}
	}
	if rel.DepartmentIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.AppUserDepartment{}).Error; err != nil {
			return err
		}
		for _, id := range rel.DepartmentIDs {
			if err := tx.Create(&models.AppUserDepartment{UserID: userID, DepartmentID: id}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
