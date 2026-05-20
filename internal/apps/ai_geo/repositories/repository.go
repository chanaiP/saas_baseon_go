package repositories

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func scopeTenant(db *gorm.DB, tenantID uint64) *gorm.DB {
	return db.Where("tenant_id = ?", tenantID)
}

func notDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}

func paginate(db *gorm.DB, req dto.PageRequest) *gorm.DB {
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	skip := req.Skip
	if skip < 0 {
		skip = 0
	}
	return db.Offset(skip).Limit(limit)
}

func likeKeyword(value string) string {
	return "%" + strings.ToLower(strings.TrimSpace(value)) + "%"
}

func countModel(ctx context.Context, db *gorm.DB, tenantID uint64, model interface{}, status string) (int64, error) {
	var count int64
	query := notDeleted(scopeTenant(db.WithContext(ctx).Model(model), tenantID))
	if status != "" {
		query = query.Where("status = ?", status)
	}
	return count, query.Count(&count).Error
}

func (r *Repository) CountBrands(ctx context.Context, tenantID uint64) (int64, error) {
	return countModel(ctx, r.db, tenantID, &models.AiGeoBrandCard{}, "active")
}

func (r *Repository) CountProducts(ctx context.Context, tenantID uint64) (int64, error) {
	return countModel(ctx, r.db, tenantID, &models.AiGeoProductCard{}, "active")
}

func (r *Repository) CountSKUs(ctx context.Context, tenantID uint64) (int64, error) {
	return countModel(ctx, r.db, tenantID, &models.AiGeoSKU{}, "active")
}

func (r *Repository) CountChannels(ctx context.Context, tenantID uint64) (int64, error) {
	return countModel(ctx, r.db, tenantID, &models.AiGeoChannelProfile{}, "active")
}

func (r *Repository) CountChannelAccounts(ctx context.Context, tenantID uint64) (int64, error) {
	return countModel(ctx, r.db, tenantID, &models.AiGeoChannelAccount{}, "active")
}

func (r *Repository) CountChannelContents(ctx context.Context, tenantID uint64) (int64, error) {
	return countModel(ctx, r.db, tenantID, &models.AiGeoChannelContent{}, "active")
}

func (r *Repository) CountDraftsToday(ctx context.Context, tenantID uint64) (int64, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int64
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoDraft{}), tenantID)).
		Where("created_at >= ?", start).
		Count(&count).Error
	return count, err
}

func (r *Repository) CountPendingDrafts(ctx context.Context, tenantID uint64) (int64, error) {
	var count int64
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoDraft{}), tenantID)).
		Where("audit_status IN ?", []string{"draft", "pending"}).
		Count(&count).Error
	return count, err
}

func (r *Repository) CountPublishPlansToday(ctx context.Context, tenantID uint64) (int64, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)
	var count int64
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoPublishPlan{}), tenantID)).
		Where("scheduled_at >= ? AND scheduled_at < ?", start, end).
		Count(&count).Error
	return count, err
}

func (r *Repository) AverageCompleteness(ctx context.Context, tenantID uint64) (int64, error) {
	var row struct{ Avg float64 }
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoProductCard{}), tenantID)).
		Select("coalesce(avg(completeness), 0) as avg").
		Scan(&row).Error
	return int64(row.Avg), err
}

func (r *Repository) ListBrands(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoBrandCard, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoBrandCard{}), tenantID))
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(brand_code) LIKE ? OR lower(brand_name) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoBrandCard
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) Brand(ctx context.Context, tenantID, id uint64) (models.AiGeoBrandCard, error) {
	var row models.AiGeoBrandCard
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoBrandCard{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) BrandByCode(ctx context.Context, tenantID uint64, code string) (models.AiGeoBrandCard, error) {
	var row models.AiGeoBrandCard
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoBrandCard{}), tenantID)).
		Where("brand_code = ?", strings.TrimSpace(code)).
		First(&row).Error
	return row, err
}

func (r *Repository) SaveBrand(ctx context.Context, brand *models.AiGeoBrandCard) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

func (r *Repository) ListProducts(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoProductCard, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoProductCard{}), tenantID))
	if req.BrandID > 0 {
		db = db.Where("brand_id = ?", req.BrandID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(product_code) LIKE ? OR lower(product_name) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoProductCard
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) Product(ctx context.Context, tenantID, id uint64) (models.AiGeoProductCard, error) {
	var row models.AiGeoProductCard
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoProductCard{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) ProductByCode(ctx context.Context, tenantID uint64, code string) (models.AiGeoProductCard, error) {
	var row models.AiGeoProductCard
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoProductCard{}), tenantID)).
		Where("product_code = ?", strings.TrimSpace(code)).
		First(&row).Error
	return row, err
}

func (r *Repository) SaveProduct(ctx context.Context, product *models.AiGeoProductCard) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *Repository) SKU(ctx context.Context, tenantID, id uint64) (models.AiGeoSKU, error) {
	var row models.AiGeoSKU
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoSKU{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) SKUByCode(ctx context.Context, tenantID uint64, code string) (models.AiGeoSKU, error) {
	var row models.AiGeoSKU
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoSKU{}), tenantID)).
		Where("sku_code = ?", strings.TrimSpace(code)).
		First(&row).Error
	return row, err
}

func (r *Repository) ListSKUs(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoSKU, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoSKU{}), tenantID))
	if req.ProductID > 0 {
		db = db.Where("product_id = ?", req.ProductID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(sku_code) LIKE ? OR lower(sku_name) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoSKU
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveSKU(ctx context.Context, sku *models.AiGeoSKU) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

func (r *Repository) Competitor(ctx context.Context, tenantID, id uint64) (models.AiGeoCompetitor, error) {
	var row models.AiGeoCompetitor
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoCompetitor{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) ListCompetitors(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoCompetitor, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoCompetitor{}), tenantID))
	if req.ProductID > 0 {
		db = db.Where("product_id = ?", req.ProductID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(brand_name) LIKE ? OR lower(product_name) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoCompetitor
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveCompetitor(ctx context.Context, competitor *models.AiGeoCompetitor) error {
	return r.db.WithContext(ctx).Save(competitor).Error
}

func (r *Repository) Keyword(ctx context.Context, tenantID, id uint64) (models.AiGeoKeyword, error) {
	var row models.AiGeoKeyword
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoKeyword{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) ListKeywords(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoKeyword, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoKeyword{}), tenantID))
	if req.BrandID > 0 {
		db = db.Where("brand_id = ?", req.BrandID)
	}
	if req.ProductID > 0 {
		db = db.Where("product_id = ?", req.ProductID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(keyword_group) LIKE ? OR lower(keyword) LIKE ? OR lower(intent) LIKE ?", k, k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoKeyword
	err := paginate(db.Order("weight desc, id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveKeyword(ctx context.Context, keyword *models.AiGeoKeyword) error {
	return r.db.WithContext(ctx).Save(keyword).Error
}

func (r *Repository) MaterialAsset(ctx context.Context, tenantID, id uint64) (models.AiGeoMaterialAsset, error) {
	var row models.AiGeoMaterialAsset
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoMaterialAsset{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) ListMaterialAssets(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoMaterialAsset, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoMaterialAsset{}), tenantID))
	if req.BrandID > 0 {
		db = db.Where("brand_id = ?", req.BrandID)
	}
	if req.ProductID > 0 {
		db = db.Where("product_id = ?", req.ProductID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(asset_type) LIKE ? OR lower(asset_name) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoMaterialAsset
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveMaterialAsset(ctx context.Context, asset *models.AiGeoMaterialAsset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

func (r *Repository) Hotspot(ctx context.Context, tenantID, id uint64) (models.AiGeoHotspot, error) {
	var row models.AiGeoHotspot
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoHotspot{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) ListHotspots(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoHotspot, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoHotspot{}), tenantID))
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(platform) LIKE ? OR lower(title) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoHotspot
	err := paginate(db.Order("captured_at desc, id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveHotspot(ctx context.Context, hotspot *models.AiGeoHotspot) error {
	return r.db.WithContext(ctx).Save(hotspot).Error
}

func (r *Repository) ListChannels(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoChannelProfile, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoChannelProfile{}), tenantID))
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(channel_code) LIKE ? OR lower(channel_name) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoChannelProfile
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) Channel(ctx context.Context, tenantID, id uint64) (models.AiGeoChannelProfile, error) {
	var row models.AiGeoChannelProfile
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoChannelProfile{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) SaveChannel(ctx context.Context, channel *models.AiGeoChannelProfile) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *Repository) ListChannelAccounts(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoChannelAccount, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoChannelAccount{}), tenantID))
	if req.ChannelID > 0 {
		db = db.Where("channel_id = ?", req.ChannelID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoChannelAccount
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveChannelAccount(ctx context.Context, account *models.AiGeoChannelAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *Repository) ListDrafts(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoDraft, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoDraft{}), tenantID))
	if req.BrandID > 0 {
		db = db.Where("brand_id = ?", req.BrandID)
	}
	if req.ProductID > 0 {
		db = db.Where("product_id = ?", req.ProductID)
	}
	if req.AuditStatus != "" {
		db = db.Where("audit_status = ?", req.AuditStatus)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(draft_code) LIKE ? OR lower(title) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoDraft
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) Draft(ctx context.Context, tenantID, id uint64) (models.AiGeoDraft, error) {
	var row models.AiGeoDraft
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoDraft{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) SaveDraft(ctx context.Context, draft *models.AiGeoDraft) error {
	return r.db.WithContext(ctx).Save(draft).Error
}

func (r *Repository) SaveChannelContent(ctx context.Context, content *models.AiGeoChannelContent) error {
	return r.db.WithContext(ctx).Save(content).Error
}

func (r *Repository) ChannelContent(ctx context.Context, tenantID, id uint64) (models.AiGeoChannelContent, error) {
	var row models.AiGeoChannelContent
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoChannelContent{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) ListChannelContents(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoChannelContent, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoChannelContent{}), tenantID))
	if req.DraftID > 0 {
		db = db.Where("draft_id = ?", req.DraftID)
	}
	if req.ChannelID > 0 {
		db = db.Where("channel_id = ?", req.ChannelID)
	}
	if req.AuditStatus != "" {
		db = db.Where("audit_status = ?", req.AuditStatus)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		k := likeKeyword(req.Keyword)
		db = db.Where("lower(title) LIKE ? OR lower(body) LIKE ?", k, k)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoChannelContent
	err := paginate(db.Order("id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SaveAuditSuggestion(ctx context.Context, row *models.AiGeoAuditSuggestion) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) ListAuditSuggestions(ctx context.Context, tenantID uint64, objectType string, objectID uint64, req dto.PageRequest) ([]models.AiGeoAuditSuggestion, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoAuditSuggestion{}), tenantID)).
		Where("object_type = ? AND object_id = ?", objectType, objectID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoAuditSuggestion
	err := paginate(db.Order("generated_at desc, id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListPublishPlans(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.AiGeoPublishPlan, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoPublishPlan{}), tenantID))
	if req.ChannelID > 0 {
		db = db.Where("channel_id = ?", req.ChannelID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.StartDate != nil {
		db = db.Where("scheduled_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		db = db.Where("scheduled_at < ?", req.EndDate.AddDate(0, 0, 1))
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoPublishPlan
	err := paginate(db.Order("scheduled_at asc, id asc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SavePublishPlan(ctx context.Context, plan *models.AiGeoPublishPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *Repository) PublishPlan(ctx context.Context, tenantID, id uint64) (models.AiGeoPublishPlan, error) {
	var row models.AiGeoPublishPlan
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoPublishPlan{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) SaveImportBatch(ctx context.Context, batch *models.AiGeoImportBatch) error {
	return r.db.WithContext(ctx).Save(batch).Error
}

func (r *Repository) ImportBatch(ctx context.Context, tenantID, id uint64) (models.AiGeoImportBatch, error) {
	var row models.AiGeoImportBatch
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoImportBatch{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) SaveImportErrors(ctx context.Context, rows []models.AiGeoImportError) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *Repository) ListImportErrors(ctx context.Context, tenantID, batchID uint64, req dto.PageRequest) ([]models.AiGeoImportError, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.AiGeoImportError{}), tenantID)).Where("batch_id = ?", batchID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AiGeoImportError
	err := paginate(db.Order("row_number asc, id asc"), req).Find(&rows).Error
	return rows, total, err
}
