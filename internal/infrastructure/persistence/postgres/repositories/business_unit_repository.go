package repositories

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type BusinessUnitRepository struct {
	db *gorm.DB
}

type BusinessUnitListQuery struct {
	TenantID      uint64
	Skip          int
	Limit         int
	KeywordLike   string
	UnitTypeCode  string
	UnitGroupCode string
	Status        string
	ScopedIDs     []uint64
	UseScope      bool
}

type BusinessUnitSummaryRow struct {
	UnitTypeCode  string
	UnitTypeName  string
	UnitGroupCode string
	UnitGroupName string
	UnitCount     int64
}

func NewBusinessUnitRepository(db *gorm.DB) *BusinessUnitRepository {
	return &BusinessUnitRepository{db: db}
}

func (r *BusinessUnitRepository) ListSummary(ctx context.Context, tenantID uint64) ([]BusinessUnitSummaryRow, error) {
	var rows []BusinessUnitSummaryRow
	err := r.db.WithContext(ctx).Model(&models.BusinessUnit{}).
		Select("unit_type_code, unit_type_name, unit_group_code, unit_group_name, COUNT(*) AS unit_count").
		Where("tenant_id = ? AND deleted_at IS NULL AND status = ?", tenantID, 1).
		Where("COALESCE(unit_type_code, '') <> '' AND COALESCE(unit_group_code, '') <> ''").
		Group("unit_type_code, unit_type_name, unit_group_code, unit_group_name").
		Order("unit_type_code asc, unit_group_code asc").
		Scan(&rows).Error
	return rows, err
}

func (r *BusinessUnitRepository) List(ctx context.Context, q BusinessUnitListQuery) ([]models.BusinessUnit, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.BusinessUnit{}).Where("tenant_id = ? AND deleted_at IS NULL", q.TenantID)
	if q.KeywordLike != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR unit_type_name LIKE ? OR unit_group_name LIKE ?", q.KeywordLike, q.KeywordLike, q.KeywordLike, q.KeywordLike)
	}
	if q.UnitTypeCode != "" {
		query = query.Where("unit_type_code = ?", q.UnitTypeCode)
	}
	if q.UnitGroupCode != "" {
		query = query.Where("unit_group_code = ?", q.UnitGroupCode)
	}
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}
	if q.UseScope {
		if len(q.ScopedIDs) == 0 {
			return []models.BusinessUnit{}, 0, nil
		}
		query = query.Where("id IN ?", q.ScopedIDs)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.BusinessUnit
	err := query.Order("updated_at desc, id desc").Offset(q.Skip).Limit(q.Limit).Find(&rows).Error
	return rows, total, err
}

func (r *BusinessUnitRepository) Get(ctx context.Context, tenantID uint64, id uint64) (models.BusinessUnit, error) {
	var row models.BusinessUnit
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&row).Error
	return row, err
}

func (r *BusinessUnitRepository) Create(ctx context.Context, row *models.BusinessUnit) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *BusinessUnitRepository) Update(ctx context.Context, row *models.BusinessUnit, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(row).Updates(updates).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", row.ID, row.TenantID).First(row).Error
}

func (r *BusinessUnitRepository) Archive(ctx context.Context, row *models.BusinessUnit, tombstoneCode string) error {
	return r.db.WithContext(ctx).Model(row).Updates(map[string]interface{}{"deleted_at": time.Now(), "status": 0, "billing_enabled": false, "code": tombstoneCode}).Error
}

func (r *BusinessUnitRepository) CountActiveChildren(ctx context.Context, tenantID uint64, id uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.BusinessUnit{}).Where("tenant_id = ? AND parent_id = ? AND deleted_at IS NULL AND status = ?", tenantID, id, 1).Count(&total).Error
	return total, err
}

func (r *BusinessUnitRepository) CountActiveRelations(ctx context.Context, tenantID uint64, id uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.BusinessUnitRelation{}).Where("tenant_id = ? AND deleted_at IS NULL AND status = ? AND (source_unit_id = ? OR target_unit_id = ?)", tenantID, "active", id, id).Count(&total).Error
	return total, err
}

func (r *BusinessUnitRepository) OrgExists(ctx context.Context, tenantID uint64, id uint64) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.OrgNode{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).Count(&total).Error
	return total > 0, err
}

func (r *BusinessUnitRepository) UserExists(ctx context.Context, tenantID uint64, id uint64) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.AppUser{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).Count(&total).Error
	return total > 0, err
}

func (r *BusinessUnitRepository) ListActors(ctx context.Context, tenantID uint64, unitID uint64) ([]models.BusinessUnitActor, error) {
	var rows []models.BusinessUnitActor
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND business_unit_id = ? AND deleted_at IS NULL", tenantID, unitID).Order("actor_type asc, actor_id asc").Find(&rows).Error
	return rows, err
}

func (r *BusinessUnitRepository) ReplaceActors(ctx context.Context, tenantID uint64, unitID uint64, rows []models.BusinessUnitActor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BusinessUnitActor{}).Where("tenant_id = ? AND business_unit_id = ? AND deleted_at IS NULL", tenantID, unitID).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		for i := range rows {
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *BusinessUnitRepository) ListRelations(ctx context.Context, tenantID uint64, unitID uint64) ([]models.BusinessUnitRelation, error) {
	var rows []models.BusinessUnitRelation
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND source_unit_id = ? AND deleted_at IS NULL", tenantID, unitID).Order("id asc").Find(&rows).Error
	return rows, err
}

func (r *BusinessUnitRepository) ReplaceRelations(ctx context.Context, tenantID uint64, unitID uint64, rows []models.BusinessUnitRelation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BusinessUnitRelation{}).Where("tenant_id = ? AND source_unit_id = ? AND deleted_at IS NULL", tenantID, unitID).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		for i := range rows {
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *BusinessUnitRepository) MatchTemplate(ctx context.Context, tenantID uint64, unitTypeCode string, unitGroupCode string) (models.BusinessUnitAttrTemplate, []models.BusinessUnitAttrTemplateField, error) {
	var tpl models.BusinessUnitAttrTemplate
	err := r.db.WithContext(ctx).
		Where("(tenant_id IS NULL OR tenant_id = ?) AND unit_type_code = ? AND deleted_at IS NULL AND status = ? AND (unit_group_code = ? OR unit_group_code IS NULL OR unit_group_code = '')", tenantID, unitTypeCode, "active", unitGroupCode).
		Order("CASE WHEN unit_group_code = '" + strings.ReplaceAll(unitGroupCode, "'", "''") + "' THEN 0 ELSE 1 END, COALESCE(tenant_id, 0) DESC, sort_order ASC, id ASC").
		First(&tpl).Error
	if err != nil {
		return tpl, nil, err
	}
	var fields []models.BusinessUnitAttrTemplateField
	err = r.db.WithContext(ctx).Where("template_id = ? AND deleted_at IS NULL AND status = ?", tpl.ID, "active").Order("sort_order asc, id asc").Find(&fields).Error
	return tpl, fields, err
}

func (r *BusinessUnitRepository) ListTemplateFields(ctx context.Context, templateID uint64) ([]models.BusinessUnitAttrTemplateField, error) {
	var fields []models.BusinessUnitAttrTemplateField
	err := r.db.WithContext(ctx).Where("template_id = ? AND deleted_at IS NULL AND status = ?", templateID, "active").Find(&fields).Error
	return fields, err
}

func (r *BusinessUnitRepository) ListTemplates(ctx context.Context, tenantID uint64, unitTypeCode string, unitGroupCode string, status string) ([]models.BusinessUnitAttrTemplate, error) {
	query := r.db.WithContext(ctx).Where("(tenant_id IS NULL OR tenant_id = ?) AND deleted_at IS NULL", tenantID)
	if strings.TrimSpace(unitTypeCode) != "" {
		query = query.Where("unit_type_code = ?", strings.TrimSpace(unitTypeCode))
	}
	if strings.TrimSpace(unitGroupCode) != "" {
		query = query.Where("COALESCE(unit_group_code, '') = ?", strings.TrimSpace(unitGroupCode))
	}
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []models.BusinessUnitAttrTemplate
	err := query.Order("unit_type_code asc, unit_group_code asc, sort_order asc, id asc").Find(&rows).Error
	return rows, err
}

func (r *BusinessUnitRepository) GetTemplate(ctx context.Context, tenantID uint64, id uint64) (models.BusinessUnitAttrTemplate, []models.BusinessUnitAttrTemplateField, error) {
	var row models.BusinessUnitAttrTemplate
	err := r.db.WithContext(ctx).Where("id = ? AND (tenant_id IS NULL OR tenant_id = ?) AND deleted_at IS NULL", id, tenantID).First(&row).Error
	if err != nil {
		return row, nil, err
	}
	fields, err := r.ListTemplateFields(ctx, row.ID)
	return row, fields, err
}

func (r *BusinessUnitRepository) TemplateScopeExists(ctx context.Context, tenantID uint64, templateName string, unitTypeCode string, unitGroupCode string, excludeID uint64) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.BusinessUnitAttrTemplate{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND template_name = ? AND unit_type_code = ? AND COALESCE(unit_group_code, '') = ?",
			tenantID,
			strings.TrimSpace(templateName),
			strings.TrimSpace(unitTypeCode),
			strings.TrimSpace(unitGroupCode),
		)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BusinessUnitRepository) CreateTemplate(ctx context.Context, template *models.BusinessUnitAttrTemplate, fields []models.BusinessUnitAttrTemplateField) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(template).Error; err != nil {
			return err
		}
		for i := range fields {
			fields[i].TemplateID = template.ID
			if err := tx.Create(&fields[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *BusinessUnitRepository) UpdateTemplate(ctx context.Context, template *models.BusinessUnitAttrTemplate, updates map[string]interface{}, fields []models.BusinessUnitAttrTemplateField) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(template).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.BusinessUnitAttrTemplateField{}).Where("template_id = ? AND deleted_at IS NULL", template.ID).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		for i := range fields {
			fields[i].TemplateID = template.ID
			if err := tx.Create(&fields[i]).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ? AND deleted_at IS NULL", template.ID).First(template).Error
	})
}

func (r *BusinessUnitRepository) ArchiveTemplate(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Model(&models.BusinessUnitAttrTemplate{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		Updates(map[string]interface{}{"deleted_at": time.Now(), "status": "archived"}).Error
}
