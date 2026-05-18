package repositories

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

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
	if req.Skip < 0 {
		req.Skip = 0
	}
	return db.Offset(req.Skip).Limit(limit)
}

func timeWindow(req dto.PageRequest) (time.Time, time.Time) {
	now := time.Now()
	end := now
	start := now.AddDate(0, 0, -6)
	if req.StartDate != nil {
		start = *req.StartDate
	}
	if req.EndDate != nil {
		end = *req.EndDate
	}
	switch req.TimeRange {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "last_30_days":
		start = now.AddDate(0, 0, -29)
	case "this_month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "last_month":
		firstThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start = firstThisMonth.AddDate(0, -1, 0)
		end = firstThisMonth.Add(-time.Nanosecond)
	}
	return start, end
}

func applyCommonFilters(db *gorm.DB, req dto.PageRequest) *gorm.DB {
	if req.BrandCode != "" {
		db = db.Where("brand_code = ?", req.BrandCode)
	}
	if req.ChannelCode != "" {
		db = db.Where("channel_code = ?", req.ChannelCode)
	}
	if req.PlatformCode != "" {
		db = db.Where("platform_code = ?", req.PlatformCode)
	}
	if req.ProductCode != "" {
		db = db.Where("product_code = ?", req.ProductCode)
	}
	if req.StoreCode != "" {
		db = db.Where("store_code = ?", req.StoreCode)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	return db
}

func likeKeyword(value string) string {
	return "%" + strings.ToLower(strings.TrimSpace(value)) + "%"
}

func (r *Repository) SalesSummary(ctx context.Context, tenantID uint64, req dto.PageRequest) (map[string]float64, error) {
	start, end := timeWindow(req)
	var row struct {
		GMV          float64
		NetSales     float64
		Orders       float64
		RefundAmount float64
	}
	err := applyCommonFilters(notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterStdSalesOrder{}), tenantID)), req).
		Where("order_time BETWEEN ? AND ?", start, end).
		Select("coalesce(sum(sales_amount),0) as gmv, coalesce(sum(paid_amount - refund_amount),0) as net_sales, count(*) as orders, coalesce(sum(refund_amount),0) as refund_amount").
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	avgOrder := 0.0
	if row.Orders > 0 {
		avgOrder = row.NetSales / row.Orders
	}
	refundRate := 0.0
	if row.GMV > 0 {
		refundRate = row.RefundAmount / row.GMV * 100
	}
	return map[string]float64{"gmv": row.GMV, "net_sales": row.NetSales, "orders": row.Orders, "avg_order_value": avgOrder, "refund_rate": refundRate}, nil
}

func (r *Repository) AdSummary(ctx context.Context, tenantID uint64, req dto.PageRequest) (map[string]float64, error) {
	start, end := timeWindow(req)
	var row struct {
		Cost  float64
		GMV   float64
		Click float64
		Impr  float64
		Order float64
	}
	err := applyCommonFilters(notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterStdAdDaily{}), tenantID)), req).
		Where("stat_date BETWEEN ? AND ?", start, end).
		Select("coalesce(sum(cost_amount),0) as cost, coalesce(sum(order_amount),0) as gmv, coalesce(sum(click_count),0) as click, coalesce(sum(impression_count),0) as impr, coalesce(sum(order_count),0) as \"order\"").
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	roi := 0.0
	if row.Cost > 0 {
		roi = row.GMV / row.Cost
	}
	cvr := 0.0
	if row.Click > 0 {
		cvr = row.Order / row.Click * 100
	}
	return map[string]float64{"ad_cost": row.Cost, "ad_gmv": row.GMV, "roi": roi, "conversion_rate": cvr}, nil
}

func (r *Repository) Trend(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]map[string]interface{}, error) {
	start, end := timeWindow(req)
	rows := make([]map[string]interface{}, 0)
	err := scopeTenant(r.db.WithContext(ctx).Table("data_center_std_sales_orders"), tenantID).
		Where("deleted_at IS NULL AND order_time BETWEEN ? AND ?", start, end).
		Select("to_char(date(order_time), 'YYYY-MM-DD') as date, coalesce(sum(sales_amount),0) as gmv, coalesce(sum(refund_amount),0) as refund_amount, count(*) as orders").
		Group("date(order_time)").
		Order("date(order_time) asc").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) Ranking(ctx context.Context, tenantID uint64, groupColumn string, req dto.PageRequest) ([]map[string]interface{}, error) {
	start, end := timeWindow(req)
	rows := make([]map[string]interface{}, 0)
	err := scopeTenant(r.db.WithContext(ctx).Table("data_center_std_sales_orders"), tenantID).
		Where("deleted_at IS NULL AND order_time BETWEEN ? AND ?", start, end).
		Where(groupColumn + " IS NOT NULL AND " + groupColumn + " <> ''").
		Select(groupColumn + " as code, coalesce(max(" + groupColumn + "), '') as name, coalesce(sum(sales_amount),0) as value").
		Group(groupColumn).
		Order("value desc").
		Limit(10).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) CountModel(ctx context.Context, tenantID uint64, model interface{}, status string) (int64, error) {
	var count int64
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(model), tenantID))
	if status != "" {
		db = db.Where("status = ?", status)
	}
	return count, db.Count(&count).Error
}

func (r *Repository) ListRawBatches(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.DataCenterRawDataBatch, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterRawDataBatch{}), tenantID))
	if req.DataType != "" {
		db = db.Where("data_type = ?", req.DataType)
	}
	if req.PlatformCode != "" {
		db = db.Where("platform_code = ?", req.PlatformCode)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		keyword := likeKeyword(req.Keyword)
		db = db.Where("lower(batch_code) LIKE ? OR lower(connection_code) LIKE ?", keyword, keyword)
	}
	if req.StartDate != nil || req.EndDate != nil || req.TimeRange != "" {
		start, end := timeWindow(req)
		db = db.Where("coalesce(sync_time, created_at) BETWEEN ? AND ?", start, end)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterRawDataBatch
	err := paginate(db.Order("sync_time desc nulls last, id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) RawBatch(ctx context.Context, tenantID, id uint64) (models.DataCenterRawDataBatch, error) {
	var row models.DataCenterRawDataBatch
	err := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterRawDataBatch{}), tenantID)).Where("id = ?", id).First(&row).Error
	return row, err
}

func (r *Repository) RawErrors(ctx context.Context, tenantID, batchID uint64, req dto.PageRequest) ([]models.DataCenterRawDataError, int64, error) {
	db := scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterRawDataError{}), tenantID)
	if batchID > 0 {
		db = db.Where("batch_id = ?", batchID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterRawDataError
	err := paginate(db.Order("id asc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListMetrics(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.DataCenterMetricDefinition, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterMetricDefinition{}), tenantID))
	if req.Domain != "" {
		db = db.Where("metric_category = ?", req.Domain)
	}
	if req.Keyword != "" {
		keyword := likeKeyword(req.Keyword)
		db = db.Where("lower(metric_code) LIKE ? OR lower(metric_name) LIKE ?", keyword, keyword)
	}
	if req.Status == "enabled" {
		db = db.Where("enabled = ?", true)
	} else if req.Status == "disabled" {
		db = db.Where("enabled = ?", false)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterMetricDefinition
	err := paginate(db.Order("metric_category asc, id asc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListRules(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.DataCenterAnomalyRule, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterAnomalyRule{}), tenantID))
	if req.Domain != "" {
		db = db.Where("business_domain = ?", req.Domain)
	}
	if req.Keyword != "" {
		keyword := likeKeyword(req.Keyword)
		db = db.Where("lower(rule_code) LIKE ? OR lower(rule_name) LIKE ?", keyword, keyword)
	}
	if req.Status == "enabled" {
		db = db.Where("enabled = ?", true)
	} else if req.Status == "disabled" {
		db = db.Where("enabled = ?", false)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterAnomalyRule
	err := paginate(db.Order("priority asc, id asc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListAnomalies(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.DataCenterAnomalyRecord, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterAnomalyRecord{}), tenantID))
	if req.Domain != "" {
		db = db.Where("business_domain = ?", req.Domain)
	}
	if req.Level != "" {
		db = db.Where("anomaly_level = ?", req.Level)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		keyword := likeKeyword(req.Keyword)
		db = db.Where("lower(anomaly_code) LIKE ? OR lower(title) LIKE ? OR lower(object_code) LIKE ?", keyword, keyword, keyword)
	}
	if req.BrandCode != "" {
		db = db.Where("object_type = ? AND object_code = ?", "brand", req.BrandCode)
	}
	if req.ProductCode != "" {
		db = db.Where("object_type = ? AND object_code = ?", "product", req.ProductCode)
	}
	if req.StoreCode != "" {
		db = db.Where("object_type = ? AND object_code = ?", "store", req.StoreCode)
	}
	if req.StartDate != nil || req.EndDate != nil || req.TimeRange != "" {
		start, end := timeWindow(req)
		db = db.Where("occurred_at BETWEEN ? AND ?", start, end)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterAnomalyRecord
	err := paginate(db.Order("occurred_at desc, id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListTasks(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.DataCenterRectificationTask, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterRectificationTask{}), tenantID))
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		keyword := likeKeyword(req.Keyword)
		db = db.Where("lower(task_code) LIKE ? OR lower(title) LIKE ?", keyword, keyword)
	}
	if req.StartDate != nil || req.EndDate != nil || req.TimeRange != "" {
		start, end := timeWindow(req)
		db = db.Where("created_at BETWEEN ? AND ?", start, end)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterRectificationTask
	err := paginate(db.Order("deadline asc nulls last, id desc"), req).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListReviews(ctx context.Context, tenantID uint64, req dto.PageRequest) ([]models.DataCenterRectificationReview, int64, error) {
	db := notDeleted(scopeTenant(r.db.WithContext(ctx).Model(&models.DataCenterRectificationReview{}), tenantID))
	if req.Status != "" {
		db = db.Where("review_conclusion = ?", req.Status)
	}
	if req.Keyword != "" {
		keyword := likeKeyword(req.Keyword)
		db = db.Where("lower(review_code) LIKE ? OR lower(task_code) LIKE ?", keyword, keyword)
	}
	if req.StartDate != nil || req.EndDate != nil || req.TimeRange != "" {
		start, end := timeWindow(req)
		db = db.Where("reviewed_at BETWEEN ? AND ?", start, end)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.DataCenterRectificationReview
	err := paginate(db.Order("reviewed_at desc nulls last, id desc"), req).Find(&rows).Error
	return rows, total, err
}
