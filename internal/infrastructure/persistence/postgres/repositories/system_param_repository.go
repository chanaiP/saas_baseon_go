package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domain "saas_baseon_go/internal/domain/system"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type SystemParamRepository struct {
	db *gorm.DB
}

func NewSystemParamRepository(db *gorm.DB) *SystemParamRepository {
	return &SystemParamRepository{db: db}
}

func (r *SystemParamRepository) List(ctx context.Context) ([]domain.Param, error) {
	var rows []models.SystemParam
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("id desc").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]domain.Param, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainParam(row))
	}
	return items, nil
}

func (r *SystemParamRepository) FindByKey(ctx context.Context, key string) (domain.Param, error) {
	var row models.SystemParam
	err := r.db.WithContext(ctx).Where("param_key = ? AND deleted_at IS NULL", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Param{}, domain.ErrParamNotFound
	}
	if err != nil {
		return domain.Param{}, err
	}
	return toDomainParam(row), nil
}

func (r *SystemParamRepository) Create(ctx context.Context, param domain.Param) (domain.Param, error) {
	row := models.SystemParam{
		TenantID:       1,
		Key:            param.Key,
		Value:          param.Value,
		Remark:         param.Remark,
		ValueType:      "string",
		TenantEditable: true,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Param{}, err
	}
	return toDomainParam(row), nil
}

func toDomainParam(row models.SystemParam) domain.Param {
	return domain.Param{
		ID:        row.ID,
		Key:       row.Key,
		Value:     row.Value,
		Remark:    row.Remark,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
