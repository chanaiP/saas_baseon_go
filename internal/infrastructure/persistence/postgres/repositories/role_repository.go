package repositories

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	domain "saas_baseon_go/internal/domain/permission"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) List(ctx context.Context, query domain.RoleListQuery) ([]domain.Role, int64, error) {
	var rows []models.Role
	db := r.db.WithContext(ctx).Model(&models.Role{}).Where("tenant_id = ? AND deleted_at IS NULL", query.TenantID)
	if query.Keyword != "" {
		kw := "%" + strings.ToLower(query.Keyword) + "%"
		db = db.Where("lower(code) LIKE ? OR lower(name) LIKE ?", kw, kw)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("id desc").Offset(query.Skip).Limit(query.Limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]domain.Role, 0, len(rows))
	for _, row := range rows {
		role, err := r.toDomainRole(ctx, row)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, role)
	}
	return items, total, nil
}

func (r *RoleRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (domain.Role, error) {
	var row models.Role
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Role{}, domain.ErrRoleNotFound
	}
	if err != nil {
		return domain.Role{}, err
	}
	return r.toDomainRole(ctx, row)
}

func (r *RoleRepository) Create(ctx context.Context, role domain.Role) (domain.Role, error) {
	row := models.Role{TenantID: role.TenantID, Code: role.Code, Name: role.Name, Description: role.Description, Status: role.Status}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return replaceRolePermissions(ctx, tx, row.ID, role.PermissionIDs)
	})
	if err != nil {
		return domain.Role{}, err
	}
	return r.FindByID(ctx, row.TenantID, row.ID)
}

func (r *RoleRepository) Update(ctx context.Context, role domain.Role, updatePermissions bool) (domain.Role, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{"name": role.Name, "description": role.Description}
		result := tx.Model(&models.Role{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", role.ID, role.TenantID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrRoleNotFound
		}
		if updatePermissions {
			return replaceRolePermissions(ctx, tx, role.ID, role.PermissionIDs)
		}
		return nil
	})
	if err != nil {
		return domain.Role{}, err
	}
	return r.FindByID(ctx, role.TenantID, role.ID)
}

func (r *RoleRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row models.Role
		if err := tx.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrRoleNotFound
			}
			return err
		}
		result := tx.Model(&row).Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"code":       tombstoneRoleCode(row.Code, row.ID),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrRoleNotFound
		}
		return nil
	})
}

func (r *RoleRepository) toDomainRole(ctx context.Context, row models.Role) (domain.Role, error) {
	var links []models.RolePermission
	if err := r.db.WithContext(ctx).Where("role_id = ?", row.ID).Find(&links).Error; err != nil {
		return domain.Role{}, err
	}
	permissionIDs := make([]uint64, 0, len(links))
	for _, link := range links {
		permissionIDs = append(permissionIDs, link.PermissionID)
	}
	return domain.Role{
		ID:            row.ID,
		TenantID:      row.TenantID,
		Code:          row.Code,
		Name:          row.Name,
		Description:   row.Description,
		Status:        row.Status,
		PermissionIDs: permissionIDs,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func tombstoneRoleCode(code string, id uint64) string {
	tail := "__deleted_" + strconv.FormatUint(id, 10) + "_" + time.Now().Format("20060102150405")
	if len(tail) >= 64 {
		return tail[len(tail)-64:]
	}
	prefix := strings.TrimSpace(code)
	prefixLen := 64 - len(tail)
	if len(prefix) > prefixLen {
		prefix = prefix[:prefixLen]
	}
	return prefix + tail
}

func replaceRolePermissions(ctx context.Context, tx *gorm.DB, roleID uint64, permissionIDs []uint64) error {
	if err := tx.WithContext(ctx).Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}
	for _, id := range permissionIDs {
		if err := tx.WithContext(ctx).Create(&models.RolePermission{RoleID: roleID, PermissionID: id}).Error; err != nil {
			return err
		}
	}
	return nil
}
