package repositories

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type DictionaryRepository struct {
	db *gorm.DB
}

func NewDictionaryRepository(db *gorm.DB) *DictionaryRepository {
	return &DictionaryRepository{db: db}
}

type DictionaryTypeListQuery struct {
	TenantID            uint64
	ScopeTenantIDs      []uint64
	IncludePlatformOnly bool
	Keyword             string
	PlatformOnly        *bool
	Skip                int
	Limit               int
}

type VisibleDictionaryTypeQuery struct {
	TenantID            uint64
	ScopeTenantIDs      []uint64
	IncludePlatformOnly bool
	ID                  uint64
	Code                string
}

func (r *DictionaryRepository) ListTypes(ctx context.Context, req DictionaryTypeListQuery) ([]models.DictType, int64, error) {
	query := r.visibleTypesQuery(ctx, req.TenantID, req.ScopeTenantIDs, req.IncludePlatformOnly)
	if req.IncludePlatformOnly && req.PlatformOnly != nil {
		query = query.Where("dict_type.is_platform_only = ?", *req.PlatformOnly)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("dict_type.name LIKE ? OR dict_type.code LIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DictType
	err := query.Clauses(clause.OrderBy{
		Expression: clause.Expr{
			SQL:  "CASE WHEN dict_type.tenant_id = ? THEN 0 ELSE 1 END, dict_type.id ASC",
			Vars: []interface{}{req.TenantID},
		},
	}).Offset(req.Skip).Limit(req.Limit).Find(&rows).Error
	return rows, total, err
}

func (r *DictionaryRepository) VisibleTypeByCode(ctx context.Context, req VisibleDictionaryTypeQuery) (models.DictType, error) {
	var row models.DictType
	query := r.visibleTypesQuery(ctx, req.TenantID, req.ScopeTenantIDs, req.IncludePlatformOnly).
		Where("dict_type.code = ?", strings.TrimSpace(req.Code))
	err := query.Clauses(clause.OrderBy{
		Expression: clause.Expr{
			SQL:  "CASE WHEN dict_type.tenant_id = ? THEN 0 ELSE 1 END, dict_type.id ASC",
			Vars: []interface{}{req.TenantID},
		},
	}).First(&row).Error
	return row, err
}

func (r *DictionaryRepository) VisibleTypeByID(ctx context.Context, req VisibleDictionaryTypeQuery) (models.DictType, error) {
	var row models.DictType
	err := r.visibleTypesQuery(ctx, req.TenantID, req.ScopeTenantIDs, req.IncludePlatformOnly).
		Where("dict_type.id = ?", req.ID).
		First(&row).Error
	return row, err
}

func (r *DictionaryRepository) GetOwnType(ctx context.Context, tenantID uint64, id uint64) (models.DictType, error) {
	var row models.DictType
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).First(&row).Error
	return row, err
}

func (r *DictionaryRepository) CreateType(ctx context.Context, row *models.DictType) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *DictionaryRepository) UpdateType(ctx context.Context, row *models.DictType, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(row).Updates(updates).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).First(row, row.ID).Error
}

func (r *DictionaryRepository) CountItemsForOwnType(ctx context.Context, tenantID uint64, dictTypeID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.DictItem{}).Where("tenant_id = ? AND dict_type_id = ? AND deleted_at IS NULL", tenantID, dictTypeID).Count(&total).Error
	return total, err
}

func (r *DictionaryRepository) ListItemsByType(ctx context.Context, tenantID uint64, dictTypeID uint64, skip int, limit int) ([]models.DictItem, error) {
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND dict_type_id = ? AND deleted_at IS NULL", tenantID, dictTypeID).Order("sort_order asc, id asc")
	if limit > 0 {
		query = query.Offset(skip).Limit(limit)
	}
	var rows []models.DictItem
	err := query.Find(&rows).Error
	return rows, err
}

func (r *DictionaryRepository) ListVisibleItemsByType(ctx context.Context, viewerTenantID uint64, baseTenantID uint64, dictTypeID uint64, skip int, limit int) ([]models.DictItem, error) {
	query := r.visibleItemsQuery(ctx, viewerTenantID, baseTenantID, dictTypeID).Order("tenant_id asc, sort_order asc, id asc")
	if limit > 0 {
		query = query.Offset(skip).Limit(limit)
	}
	var rows []models.DictItem
	err := query.Find(&rows).Error
	return rows, err
}

func (r *DictionaryRepository) ListItemsByTypePage(ctx context.Context, tenantID uint64, dictTypeID uint64, skip int, limit int) ([]models.DictItem, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.DictItem{}).Where("tenant_id = ? AND dict_type_id = ? AND deleted_at IS NULL", tenantID, dictTypeID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DictItem
	err := query.Order("sort_order asc, id asc").Offset(skip).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *DictionaryRepository) ListVisibleItemsByTypePage(ctx context.Context, viewerTenantID uint64, baseTenantID uint64, dictTypeID uint64, skip int, limit int) ([]models.DictItem, int64, error) {
	query := r.visibleItemsQuery(ctx, viewerTenantID, baseTenantID, dictTypeID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DictItem
	err := query.Order("tenant_id asc, sort_order asc, id asc").Offset(skip).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *DictionaryRepository) CreateItem(ctx context.Context, row *models.DictItem) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *DictionaryRepository) GetItem(ctx context.Context, id uint64) (models.DictItem, error) {
	var row models.DictItem
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	return row, err
}

func (r *DictionaryRepository) GetOwnItem(ctx context.Context, tenantID uint64, id uint64) (models.DictItem, error) {
	var row models.DictItem
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).First(&row).Error
	return row, err
}

func (r *DictionaryRepository) UpdateItem(ctx context.Context, row *models.DictItem, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(row).Updates(updates).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).First(row, row.ID).Error
}

func (r *DictionaryRepository) AssertItemInType(ctx context.Context, tenantID uint64, dictTypeID uint64, itemID uint64) error {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.DictItem{}).Where("tenant_id = ? AND dict_type_id = ? AND id = ? AND deleted_at IS NULL", tenantID, dictTypeID, itemID).Count(&total).Error; err != nil {
		return err
	}
	if total == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *DictionaryRepository) AssertVisibleItemInType(ctx context.Context, viewerTenantID uint64, baseTenantID uint64, dictTypeID uint64, itemID uint64) error {
	var total int64
	if err := r.visibleItemsQuery(ctx, viewerTenantID, baseTenantID, dictTypeID).Where("id = ?", itemID).Count(&total).Error; err != nil {
		return err
	}
	if total == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *DictionaryRepository) CountChildItems(ctx context.Context, parentID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.DictItem{}).Where("parent_id = ? AND deleted_at IS NULL", parentID).Count(&total).Error
	return total, err
}

func (r *DictionaryRepository) GetOverride(ctx context.Context, tenantID uint64, itemID uint64) (*models.TenantDictItemOverride, error) {
	var row models.TenantDictItemOverride
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND dict_item_id = ?", tenantID, itemID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DictionaryRepository) UpsertOverride(ctx context.Context, tenantID uint64, itemID uint64, label *string, value *string, sortOrder *int, enabled *bool) (*models.TenantDictItemOverride, error) {
	var row models.TenantDictItemOverride
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND dict_item_id = ?", tenantID, itemID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.TenantDictItemOverride{TenantID: tenantID, DictItemID: itemID}
	} else if err != nil {
		return nil, err
	}
	row.CustomLabel = nullableTrimmed(label)
	row.CustomValue = nullableTrimmed(value)
	row.SortOrder = sortOrder
	row.Enabled = enabled
	if row.ID == 0 {
		err = r.db.WithContext(ctx).Create(&row).Error
	} else {
		err = r.db.WithContext(ctx).Save(&row).Error
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DictionaryRepository) DeleteOverride(ctx context.Context, tenantID uint64, itemID uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND dict_item_id = ?", tenantID, itemID).Delete(&models.TenantDictItemOverride{}).Error
}

func (r *DictionaryRepository) visibleTypesQuery(ctx context.Context, tenantID uint64, scopeTenantIDs []uint64, includePlatformOnly bool) *gorm.DB {
	if len(scopeTenantIDs) == 0 {
		scopeTenantIDs = []uint64{tenantID}
	}
	query := r.db.WithContext(ctx).Model(&models.DictType{}).Where("dict_type.deleted_at IS NULL")
	if includePlatformOnly {
		return query.Where("dict_type.tenant_id IN ?", scopeTenantIDs)
	}
	return query.Where(
		"dict_type.tenant_id = ? OR (dict_type.tenant_id IN ? AND dict_type.is_platform_only = ? AND NOT EXISTS (SELECT 1 FROM dict_type tenant_dict WHERE tenant_dict.tenant_id = ? AND tenant_dict.code = dict_type.code AND tenant_dict.deleted_at IS NULL))",
		tenantID,
		scopeTenantIDs,
		false,
		tenantID,
	)
}

func (r *DictionaryRepository) visibleItemsQuery(ctx context.Context, viewerTenantID uint64, baseTenantID uint64, dictTypeID uint64) *gorm.DB {
	tenantIDs := []uint64{baseTenantID}
	if viewerTenantID != baseTenantID {
		tenantIDs = append(tenantIDs, viewerTenantID)
	}
	return r.db.WithContext(ctx).Model(&models.DictItem{}).Where("tenant_id IN ? AND dict_type_id = ? AND deleted_at IS NULL", tenantIDs, dictTypeID)
}

func nullableTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
