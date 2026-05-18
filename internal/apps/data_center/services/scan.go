package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/apps/data_center/domain"
	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type ScanResult struct {
	Scanned          int `json:"scanned"`
	MetricsGenerated int `json:"metrics_generated"`
	Generated        int `json:"generated"`
	Duplicated       int `json:"duplicated"`
}

type salesDailyMetric struct {
	StatDate     string
	ObjectCode   string
	ObjectName   *string
	GMV          float64
	NetSales     float64
	OrderCount   float64
	RefundAmount float64
}

type adDailyMetric struct {
	StatDate    string
	ObjectCode  string
	ObjectName  *string
	CostAmount  float64
	OrderAmount float64
	ClickCount  float64
	Impression  float64
	OrderCount  float64
}

type inventoryDailyMetric struct {
	StatDate     string
	ObjectCode   string
	ObjectName   *string
	Available    int64
	Sales7D      int64
	AvailableDay float64
}

type metricTarget struct {
	StatDate   string
	ObjectCode string
	Value      float64
}

func (s *Service) ScanAnomalies(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (ScanResult, error) {
	if err := s.EnsureTenantBaseline(ctx, viewer); err != nil {
		return ScanResult{}, err
	}
	if err := s.consumeQuota(ctx, viewer.TenantID, "data_center_daily_scans"); err != nil {
		return ScanResult{}, err
	}

	result := ScanResult{}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result.MetricsGenerated, err = s.calculateMetricResults(ctx, tx, viewer, req)
		if err != nil {
			return err
		}
		scanned, generated, duplicated, err := s.scanMVPAnomalies(ctx, tx, viewer, req)
		if err != nil {
			return err
		}
		result.Scanned = scanned
		result.Generated = generated
		result.Duplicated = duplicated
		return nil
	})
	return result, err
}

func (s *Service) calculateMetricResults(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, req dto.PageRequest) (int, error) {
	now := time.Now()
	rows := make([]models.DataCenterMetricResult, 0)

	salesRows, err := s.loadSalesDaily(ctx, tx, viewer.TenantID, req, false)
	if err != nil {
		return 0, err
	}
	previousSalesRows, err := s.loadSalesDaily(ctx, tx, viewer.TenantID, req, true)
	if err != nil {
		return 0, err
	}
	previousSales := salesMetricMap(previousSalesRows)
	salesTargets, err := s.loadSalesTargets(ctx, tx, viewer.TenantID, req)
	if err != nil {
		return 0, err
	}
	for _, row := range salesRows {
		previous := previousSales[row.ObjectCode]
		target := salesTargets[targetKey(row.StatDate, row.ObjectCode)]
		refundRate := 0.0
		if row.GMV > 0 {
			refundRate = row.RefundAmount / row.GMV * 100
		}
		avgOrder := 0.0
		if row.OrderCount > 0 {
			avgOrder = row.NetSales / row.OrderCount
		}
		rows = append(rows,
			metricResult(viewer.TenantID, "gmv", "brand", row.ObjectCode, row.ObjectName, statDate(row.StatDate), row.GMV, ptrFloat(previous.GMV), ptrFloat(target), now),
			metricResult(viewer.TenantID, "net_sales", "brand", row.ObjectCode, row.ObjectName, statDate(row.StatDate), row.NetSales, ptrFloat(previous.NetSales), ptrFloat(target), now),
			metricResult(viewer.TenantID, "order_count", "brand", row.ObjectCode, row.ObjectName, statDate(row.StatDate), row.OrderCount, ptrFloat(previous.OrderCount), nil, now),
			metricResult(viewer.TenantID, "avg_order_value", "brand", row.ObjectCode, row.ObjectName, statDate(row.StatDate), avgOrder, ptrFloat(avgOrderValue(previous)), nil, now),
			metricResult(viewer.TenantID, "refund_rate", "brand", row.ObjectCode, row.ObjectName, statDate(row.StatDate), refundRate, ptrFloat(refundRateValue(previous)), nil, now),
		)
	}

	adRows, err := s.loadAdDaily(ctx, tx, viewer.TenantID, req, false)
	if err != nil {
		return 0, err
	}
	previousAdRows, err := s.loadAdDaily(ctx, tx, viewer.TenantID, req, true)
	if err != nil {
		return 0, err
	}
	previousAd := adMetricMap(previousAdRows)
	for _, row := range adRows {
		previous := previousAd[row.ObjectCode]
		roi := 0.0
		if row.CostAmount > 0 {
			roi = row.OrderAmount / row.CostAmount
		}
		clickRate := 0.0
		if row.Impression > 0 {
			clickRate = row.ClickCount / row.Impression * 100
		}
		conversionRate := 0.0
		if row.ClickCount > 0 {
			conversionRate = row.OrderCount / row.ClickCount * 100
		}
		acquisitionCost := 0.0
		if row.OrderCount > 0 {
			acquisitionCost = row.CostAmount / row.OrderCount
		}
		rows = append(rows,
			metricResult(viewer.TenantID, "ad_cost", "campaign", row.ObjectCode, row.ObjectName, statDate(row.StatDate), row.CostAmount, ptrFloat(previous.CostAmount), nil, now),
			metricResult(viewer.TenantID, "ad_gmv", "campaign", row.ObjectCode, row.ObjectName, statDate(row.StatDate), row.OrderAmount, ptrFloat(previous.OrderAmount), nil, now),
			metricResult(viewer.TenantID, "ad_roi", "campaign", row.ObjectCode, row.ObjectName, statDate(row.StatDate), roi, ptrFloat(roiValue(previous)), nil, now),
			metricResult(viewer.TenantID, "click_rate", "campaign", row.ObjectCode, row.ObjectName, statDate(row.StatDate), clickRate, ptrFloat(clickRateValue(previous)), nil, now),
			metricResult(viewer.TenantID, "conversion_rate", "campaign", row.ObjectCode, row.ObjectName, statDate(row.StatDate), conversionRate, ptrFloat(conversionRateValue(previous)), nil, now),
			metricResult(viewer.TenantID, "acquisition_cost", "campaign", row.ObjectCode, row.ObjectName, statDate(row.StatDate), acquisitionCost, ptrFloat(acquisitionCostValue(previous)), nil, now),
		)
	}

	inventoryRows, err := s.loadInventoryDaily(ctx, tx, viewer.TenantID, req, false)
	if err != nil {
		return 0, err
	}
	previousInventoryRows, err := s.loadInventoryDaily(ctx, tx, viewer.TenantID, req, true)
	if err != nil {
		return 0, err
	}
	previousInventory := inventoryMetricMap(previousInventoryRows)
	for _, row := range inventoryRows {
		previous := previousInventory[row.ObjectCode]
		sellThroughRate := 0.0
		if row.Sales7D+row.Available > 0 {
			sellThroughRate = float64(row.Sales7D) / float64(row.Sales7D+row.Available) * 100
		}
		rows = append(rows,
			metricResult(viewer.TenantID, "available_stock", "sku", row.ObjectCode, row.ObjectName, statDate(row.StatDate), float64(row.Available), ptrFloat(float64(previous.Available)), nil, now),
			metricResult(viewer.TenantID, "sales_7d", "sku", row.ObjectCode, row.ObjectName, statDate(row.StatDate), float64(row.Sales7D), ptrFloat(float64(previous.Sales7D)), nil, now),
			metricResult(viewer.TenantID, "sell_through_rate", "sku", row.ObjectCode, row.ObjectName, statDate(row.StatDate), sellThroughRate, ptrFloat(sellThroughRateValue(previous)), nil, now),
			metricResult(viewer.TenantID, "available_days", "sku", row.ObjectCode, row.ObjectName, statDate(row.StatDate), row.AvailableDay, ptrFloat(previous.AvailableDay), nil, now),
		)
	}

	for i := range rows {
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "tenant_id"}, {Name: "metric_code"}, {Name: "resource_type"},
				{Name: "resource_code"}, {Name: "stat_date"}, {Name: "period_type"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"resource_name", "metric_value", "compare_value", "compare_rate", "target_value", "dimension_json", "updated_at"}),
		}).Create(&rows[i]).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

func (s *Service) scanMVPAnomalies(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, req dto.PageRequest) (int, int, int, error) {
	scanned, generated, duplicated := 0, 0, 0

	currentSales, err := s.loadSalesDaily(ctx, tx, viewer.TenantID, req, true)
	if err != nil {
		return 0, 0, 0, err
	}
	previousSales, err := s.loadPreviousSalesByBrand(ctx, tx, viewer.TenantID, currentSales)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, current := range currentSales {
		previous, ok := previousSales[current.ObjectCode]
		if !ok || previous.GMV <= 0 || previous.OrderCount <= 0 {
			continue
		}
		scanned++
		gmvRate := changeRate(current.GMV, previous.GMV)
		orderRate := changeRate(current.OrderCount, previous.OrderCount)
		if gmvRate < -15 && orderRate < -10 {
			impact := previous.GMV - current.GMV
			level := levelByNegativeChange(gmvRate)
			anomaly := anomalyRecord(viewer.TenantID, "gmv_drop", "销售异常", "brand", current.ObjectCode, current.ObjectName, statDate(current.StatDate), level, confidenceByLevel(level), impact, "GMV 与订单数环比同时下滑", []map[string]interface{}{
				evidence("gmv", "GMV 环比", current.GMV, previous.GMV, gmvRate, "%"),
				evidence("order_count", "订单数环比", current.OrderCount, previous.OrderCount, orderRate, "%"),
			})
			created, err := createAnomalyIfMissing(ctx, tx, anomaly)
			if err != nil {
				return scanned, generated, duplicated, err
			}
			if created {
				generated++
			} else {
				duplicated++
			}
		}
	}

	currentAds, err := s.loadAdDaily(ctx, tx, viewer.TenantID, req, true)
	if err != nil {
		return 0, 0, 0, err
	}
	previousAds, err := s.loadPreviousAdByCampaign(ctx, tx, viewer.TenantID, currentAds)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, current := range currentAds {
		previous, ok := previousAds[current.ObjectCode]
		if !ok || previous.CostAmount <= 0 {
			continue
		}
		currentROI := ratio(current.OrderAmount, current.CostAmount)
		previousROI := ratio(previous.OrderAmount, previous.CostAmount)
		if previousROI <= 0 {
			continue
		}
		scanned++
		costRate := changeRate(current.CostAmount, previous.CostAmount)
		roiRate := changeRate(currentROI, previousROI)
		if costRate > 20 && roiRate < -15 {
			impact := current.CostAmount - previous.CostAmount
			level := levelBySpendROI(costRate, roiRate)
			anomaly := anomalyRecord(viewer.TenantID, "ad_spend_up_roi_down", "投流异常", "campaign", current.ObjectCode, current.ObjectName, statDate(current.StatDate), level, confidenceByLevel(level), impact, "投流消耗上升但 ROI 下滑", []map[string]interface{}{
				evidence("ad_cost", "广告消耗环比", current.CostAmount, previous.CostAmount, costRate, "%"),
				evidence("ad_roi", "ROI 环比", currentROI, previousROI, roiRate, "%"),
			})
			created, err := createAnomalyIfMissing(ctx, tx, anomaly)
			if err != nil {
				return scanned, generated, duplicated, err
			}
			if created {
				generated++
			} else {
				duplicated++
			}
		}
	}

	currentInventory, err := s.loadInventoryDaily(ctx, tx, viewer.TenantID, req, true)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, current := range currentInventory {
		scanned++
		if current.AvailableDay > 30 && current.Sales7D <= 3 {
			level := levelByAvailableDays(current.AvailableDay)
			anomaly := anomalyRecord(viewer.TenantID, "inventory_enough_sales_down", "库存异常", "sku", current.ObjectCode, current.ObjectName, statDate(current.StatDate), level, confidenceByLevel(level), 0, "库存可售天数偏高且近 7 日销量偏低", []map[string]interface{}{
				{"metric_code": "available_days", "label": "可售天数", "value": current.AvailableDay, "desc": "可售天数高于 30 天"},
				{"metric_code": "sales_7d", "label": "近 7 日销量", "value": current.Sales7D, "desc": "近 7 日销量不高于 3 件"},
			})
			created, err := createAnomalyIfMissing(ctx, tx, anomaly)
			if err != nil {
				return scanned, generated, duplicated, err
			}
			if created {
				generated++
			} else {
				duplicated++
			}
		}
	}
	return scanned, generated, duplicated, nil
}

func (s *Service) loadSalesDaily(ctx context.Context, tx *gorm.DB, tenantID uint64, req dto.PageRequest, latestOnly bool) ([]salesDailyMetric, error) {
	db := tx.WithContext(ctx).Model(&models.DataCenterStdSalesOrder{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
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
	start, end := scanTimeWindow(req)
	db = db.Where("order_time BETWEEN ? AND ?", start, end)
	if latestOnly {
		db = db.Where("date(order_time) = (?)", tx.Model(&models.DataCenterStdSalesOrder{}).
			Select("date(max(order_time))").
			Where("tenant_id = ? AND deleted_at IS NULL AND order_time BETWEEN ? AND ?", tenantID, start, end))
	}
	var rows []salesDailyMetric
	err := db.Select("min(order_time) as stat_date, coalesce(brand_code, 'UNKNOWN') as object_code, max(brand_name) as object_name, coalesce(sum(sales_amount),0) as gmv, coalesce(sum(paid_amount - refund_amount),0) as net_sales, count(*) as order_count, coalesce(sum(refund_amount),0) as refund_amount").
		Group("date(order_time), coalesce(brand_code, 'UNKNOWN')").
		Order("stat_date asc").
		Scan(&rows).Error
	return rows, err
}

func (s *Service) loadPreviousSalesByBrand(ctx context.Context, tx *gorm.DB, tenantID uint64, currentRows []salesDailyMetric) (map[string]salesDailyMetric, error) {
	result := map[string]salesDailyMetric{}
	for _, current := range currentRows {
		var row salesDailyMetric
		err := tx.WithContext(ctx).Model(&models.DataCenterStdSalesOrder{}).
			Where("tenant_id = ? AND deleted_at IS NULL AND coalesce(brand_code, 'UNKNOWN') = ? AND date(order_time) < date(?)", tenantID, current.ObjectCode, current.StatDate).
			Select("min(order_time) as stat_date, coalesce(brand_code, 'UNKNOWN') as object_code, max(brand_name) as object_name, coalesce(sum(sales_amount),0) as gmv, coalesce(sum(paid_amount - refund_amount),0) as net_sales, count(*) as order_count, coalesce(sum(refund_amount),0) as refund_amount").
			Group("date(order_time), coalesce(brand_code, 'UNKNOWN')").
			Order("stat_date desc").
			Limit(1).
			Scan(&row).Error
		if err != nil {
			return result, err
		}
		if row.StatDate != "" {
			result[current.ObjectCode] = row
		}
	}
	return result, nil
}

func (s *Service) loadSalesTargets(ctx context.Context, tx *gorm.DB, tenantID uint64, req dto.PageRequest) (map[string]float64, error) {
	db := tx.WithContext(ctx).Model(&models.DataCenterStdStoreSalesDaily{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND target_amount IS NOT NULL", tenantID)
	if req.BrandCode != "" {
		db = db.Where("brand_code = ?", req.BrandCode)
	}
	if req.ChannelCode != "" {
		db = db.Where("channel_code = ?", req.ChannelCode)
	}
	if req.PlatformCode != "" {
		db = db.Where("platform_code = ?", req.PlatformCode)
	}
	if req.StoreCode != "" {
		db = db.Where("store_code = ?", req.StoreCode)
	}
	start, end := scanTimeWindow(req)
	db = db.Where("stat_date BETWEEN ? AND ?", start, end)
	var rows []metricTarget
	if err := db.Select("stat_date, coalesce(brand_code, 'UNKNOWN') as object_code, coalesce(sum(target_amount),0) as value").
		Group("stat_date, coalesce(brand_code, 'UNKNOWN')").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	targets := make(map[string]float64, len(rows))
	for _, row := range rows {
		targets[targetKey(row.StatDate, row.ObjectCode)] = row.Value
	}
	return targets, nil
}

func (s *Service) loadAdDaily(ctx context.Context, tx *gorm.DB, tenantID uint64, req dto.PageRequest, latestOnly bool) ([]adDailyMetric, error) {
	db := tx.WithContext(ctx).Model(&models.DataCenterStdAdDaily{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if req.BrandCode != "" {
		db = db.Where("brand_code = ?", req.BrandCode)
	}
	if req.PlatformCode != "" {
		db = db.Where("platform_code = ?", req.PlatformCode)
	}
	if req.ProductCode != "" {
		db = db.Where("product_code = ?", req.ProductCode)
	}
	start, end := scanTimeWindow(req)
	db = db.Where("stat_date BETWEEN ? AND ?", start, end)
	if latestOnly {
		db = db.Where("stat_date = (?)", tx.Model(&models.DataCenterStdAdDaily{}).
			Select("max(stat_date)").
			Where("tenant_id = ? AND deleted_at IS NULL AND stat_date BETWEEN ? AND ?", tenantID, start, end))
	}
	var rows []adDailyMetric
	err := db.Select("stat_date, campaign_code as object_code, max(campaign_name) as object_name, coalesce(sum(cost_amount),0) as cost_amount, coalesce(sum(order_amount),0) as order_amount, coalesce(sum(click_count),0) as click_count, coalesce(sum(impression_count),0) as impression, coalesce(sum(order_count),0) as order_count").
		Group("stat_date, campaign_code").
		Order("stat_date asc").
		Scan(&rows).Error
	return rows, err
}

func (s *Service) loadPreviousAdByCampaign(ctx context.Context, tx *gorm.DB, tenantID uint64, currentRows []adDailyMetric) (map[string]adDailyMetric, error) {
	result := map[string]adDailyMetric{}
	for _, current := range currentRows {
		var row adDailyMetric
		err := tx.WithContext(ctx).Model(&models.DataCenterStdAdDaily{}).
			Where("tenant_id = ? AND deleted_at IS NULL AND campaign_code = ? AND date(stat_date) < date(?)", tenantID, current.ObjectCode, current.StatDate).
			Select("stat_date, campaign_code as object_code, max(campaign_name) as object_name, coalesce(sum(cost_amount),0) as cost_amount, coalesce(sum(order_amount),0) as order_amount, coalesce(sum(click_count),0) as click_count, coalesce(sum(impression_count),0) as impression, coalesce(sum(order_count),0) as order_count").
			Group("stat_date, campaign_code").
			Order("stat_date desc").
			Limit(1).
			Scan(&row).Error
		if err != nil {
			return result, err
		}
		if row.StatDate != "" {
			result[current.ObjectCode] = row
		}
	}
	return result, nil
}

func (s *Service) loadInventoryDaily(ctx context.Context, tx *gorm.DB, tenantID uint64, req dto.PageRequest, latestOnly bool) ([]inventoryDailyMetric, error) {
	db := tx.WithContext(ctx).Model(&models.DataCenterStdInventoryDaily{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if req.BrandCode != "" {
		db = db.Where("brand_code = ?", req.BrandCode)
	}
	if req.ProductCode != "" {
		db = db.Where("product_code = ?", req.ProductCode)
	}
	if req.StoreCode != "" {
		db = db.Where("store_code = ?", req.StoreCode)
	}
	start, end := scanTimeWindow(req)
	db = db.Where("stat_date BETWEEN ? AND ?", start, end)
	if latestOnly {
		db = db.Where("stat_date = (?)", tx.Model(&models.DataCenterStdInventoryDaily{}).
			Select("max(stat_date)").
			Where("tenant_id = ? AND deleted_at IS NULL AND stat_date BETWEEN ? AND ?", tenantID, start, end))
	}
	var rows []inventoryDailyMetric
	err := db.Select("stat_date, sku_code as object_code, max(product_name) as object_name, coalesce(sum(available_stock),0) as available, coalesce(sum(sales_7d),0) as sales7_d, max(available_days) as available_day").
		Group("stat_date, sku_code").
		Order("stat_date asc").
		Scan(&rows).Error
	return rows, err
}

func metricResult(tenantID uint64, metricCode, resourceType, resourceCode string, resourceName *string, statDate time.Time, value float64, compare *float64, target *float64, now time.Time) models.DataCenterMetricResult {
	var compareRate *float64
	if compare != nil && *compare != 0 {
		rate := changeRate(value, *compare)
		compareRate = &rate
	}
	return models.DataCenterMetricResult{
		TenantID: tenantID, MetricCode: metricCode, ResourceType: resourceType, ResourceCode: resourceCode,
		ResourceName: resourceName, StatDate: dateOnly(statDate), PeriodType: "day", MetricValue: value,
		CompareValue: compare, CompareRate: compareRate, TargetValue: target, DimensionJSON: "{}", CreatedAt: now, UpdatedAt: now,
	}
}

func targetKey(statDate string, objectCode string) string {
	return dateKey(statDate) + "|" + objectCode
}

func dateKey(value string) string {
	parsed := statDate(value)
	if !parsed.IsZero() {
		return parsed.Format("2006-01-02")
	}
	if len(value) >= len("2006-01-02") {
		return value[:len("2006-01-02")]
	}
	return value
}

func salesMetricMap(rows []salesDailyMetric) map[string]salesDailyMetric {
	result := make(map[string]salesDailyMetric, len(rows))
	for _, row := range rows {
		result[row.ObjectCode] = row
	}
	return result
}

func adMetricMap(rows []adDailyMetric) map[string]adDailyMetric {
	result := make(map[string]adDailyMetric, len(rows))
	for _, row := range rows {
		result[row.ObjectCode] = row
	}
	return result
}

func inventoryMetricMap(rows []inventoryDailyMetric) map[string]inventoryDailyMetric {
	result := make(map[string]inventoryDailyMetric, len(rows))
	for _, row := range rows {
		result[row.ObjectCode] = row
	}
	return result
}

func ptrFloat(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}

func avgOrderValue(row salesDailyMetric) float64 {
	if row.OrderCount <= 0 {
		return 0
	}
	return row.NetSales / row.OrderCount
}

func refundRateValue(row salesDailyMetric) float64 {
	if row.GMV <= 0 {
		return 0
	}
	return row.RefundAmount / row.GMV * 100
}

func roiValue(row adDailyMetric) float64 {
	if row.CostAmount <= 0 {
		return 0
	}
	return row.OrderAmount / row.CostAmount
}

func clickRateValue(row adDailyMetric) float64 {
	if row.Impression <= 0 {
		return 0
	}
	return row.ClickCount / row.Impression * 100
}

func conversionRateValue(row adDailyMetric) float64 {
	if row.ClickCount <= 0 {
		return 0
	}
	return row.OrderCount / row.ClickCount * 100
}

func acquisitionCostValue(row adDailyMetric) float64 {
	if row.OrderCount <= 0 {
		return 0
	}
	return row.CostAmount / row.OrderCount
}

func sellThroughRateValue(row inventoryDailyMetric) float64 {
	if row.Sales7D+row.Available <= 0 {
		return 0
	}
	return float64(row.Sales7D) / float64(row.Sales7D+row.Available) * 100
}

func levelByNegativeChange(rate float64) string {
	switch {
	case rate <= -50:
		return "critical"
	case rate <= -30:
		return "high"
	case rate <= -15:
		return "medium"
	default:
		return "low"
	}
}

func levelBySpendROI(costRate float64, roiRate float64) string {
	switch {
	case costRate >= 50 && roiRate <= -40:
		return "critical"
	case costRate >= 35 || roiRate <= -30:
		return "high"
	case costRate > 20 && roiRate < -15:
		return "medium"
	default:
		return "low"
	}
}

func levelByAvailableDays(days float64) string {
	switch {
	case days > 90:
		return "critical"
	case days > 60:
		return "high"
	case days > 30:
		return "medium"
	default:
		return "low"
	}
}

func confidenceByLevel(level string) int {
	switch level {
	case "critical":
		return 94
	case "high":
		return 88
	case "medium":
		return 82
	default:
		return 76
	}
}

func anomalyRecord(tenantID uint64, ruleCode, domainName, objectType, objectCode string, objectName *string, statDate time.Time, level string, confidence int, impact float64, title string, evidences []map[string]interface{}) models.DataCenterAnomalyRecord {
	now := time.Now()
	evidenceJSON := jsonString(evidences, "[]")
	return models.DataCenterAnomalyRecord{
		TenantID: tenantID, AnomalyCode: anomalyCode(tenantID, ruleCode, objectType, objectCode, statDate),
		RuleCode: ruleCode, Title: title, BusinessDomain: domainName, ObjectType: objectType, ObjectCode: objectCode,
		ObjectName: objectName, StatDate: dateOnly(statDate), AnomalyLevel: level, ConfidenceScore: confidence,
		ImpactAmount: positive(impact), EvidenceJSON: evidenceJSON, OccurredAt: now,
		AIStatus: domain.AnomalyAIStatusPending, TaskStatus: domain.AnomalyTaskStatusNone,
		ReviewStatus: domain.AnomalyReviewStatusNone, Status: domain.AnomalyStatusPending, CreatedAt: now, UpdatedAt: now,
	}
}

func createAnomalyIfMissing(ctx context.Context, tx *gorm.DB, row models.DataCenterAnomalyRecord) (bool, error) {
	var count int64
	if err := tx.WithContext(ctx).Model(&models.DataCenterAnomalyRecord{}).
		Where("tenant_id = ? AND rule_code = ? AND object_type = ? AND object_code = ? AND stat_date = ? AND deleted_at IS NULL",
			row.TenantID, row.RuleCode, row.ObjectType, row.ObjectCode, row.StatDate).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	return true, tx.WithContext(ctx).Create(&row).Error
}

func evidence(metricCode, label string, value, compare, rate float64, unit string) map[string]interface{} {
	return map[string]interface{}{
		"metric_code": metricCode,
		"label":       label,
		"value":       value,
		"compare":     compare,
		"change_rate": rate,
		"unit":        unit,
		"desc":        fmt.Sprintf("%s %.2f%s", label, rate, unit),
	}
}

func anomalyCode(tenantID uint64, ruleCode, objectType, objectCode string, statDate time.Time) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%d:%s:%s:%s:%s", tenantID, ruleCode, objectType, objectCode, dateOnly(statDate).Format("2006-01-02"))))
	return "ANM-" + stringsUpper(hex.EncodeToString(sum[:])[:16])
}

func statDate(value string) time.Time {
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999-07:00 MST",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return dateOnly(parsed)
		}
	}
	return dateOnly(time.Now())
}

func stringsUpper(value string) string {
	result := []byte(value)
	for i := range result {
		if result[i] >= 'a' && result[i] <= 'f' {
			result[i] -= 'a' - 'A'
		}
	}
	return string(result)
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func changeRate(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return (current - previous) / previous * 100
}

func ratio(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func positive(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func scanTimeWindow(req dto.PageRequest) (time.Time, time.Time) {
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
