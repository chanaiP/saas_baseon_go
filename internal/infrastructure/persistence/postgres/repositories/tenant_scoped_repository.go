package repositories

import (
	"context"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type TenantScopedRepository struct {
	db       *gorm.DB
	tenantID uint64
}

func NewTenantScopedRepository(db *gorm.DB, tenantID uint64) *TenantScopedRepository {
	return &TenantScopedRepository{db: db, tenantID: tenantID}
}

func (r *TenantScopedRepository) Users(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.AppUser{}).Where("tenant_id = ? AND deleted_at IS NULL", r.tenantID)
}

func (r *TenantScopedRepository) ActiveUsers(ctx context.Context) *gorm.DB {
	return r.Users(ctx).Where("status = ?", 1)
}

func (r *TenantScopedRepository) Files(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.FileObject{}).Where("tenant_id = ? AND deleted_at IS NULL", r.tenantID)
}

func (r *TenantScopedRepository) ActiveFiles(ctx context.Context) *gorm.DB {
	return r.Files(ctx).Where("status = ?", 1)
}

func (r *TenantScopedRepository) QuotaUsage(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.TenantQuotaUsage{}).Where("tenant_id = ?", r.tenantID)
}
