package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/data_center/domain"
	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (s *Service) CreateRawBatch(ctx context.Context, viewer dto.Viewer, payload dto.RawBatchPayload) (models.DataCenterRawDataBatch, error) {
	dataType := strings.TrimSpace(payload.DataType)
	if dataType == "" || len(payload.Records) == 0 {
		return models.DataCenterRawDataBatch{}, errors.New("data_type 和 records 不能为空")
	}
	if !supportedImportType(dataType) {
		return models.DataCenterRawDataBatch{}, errors.New("不支持的数据类型")
	}
	now := time.Now()
	batchCode := strings.TrimSpace(payload.BatchCode)
	if batchCode == "" {
		batchCode = fmt.Sprintf("DC-%s-%d", strings.ToUpper(strings.ReplaceAll(dataType, "-", "")), now.UnixNano())
	}
	batch := models.DataCenterRawDataBatch{
		TenantID: viewer.TenantID, BatchCode: batchCode, DataType: dataType, PlatformCode: optional(payload.PlatformCode),
		AppCode: optional(payload.AppCode), ConnectionCode: optional(payload.ConnectionCode), RecordCount: int64(len(payload.Records)),
		Status: domain.StatusProcessing, SourceParams: jsonString(payload.SourceParams, "{}"),
		SamplePayload: jsonString(firstRecord(payload.Records), "{}"), CreatedBy: &viewer.UserID, UpdatedBy: &viewer.UserID,
		SyncTime: &now, CreatedAt: now, UpdatedAt: now,
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		success, failed := int64(0), int64(0)
		for index, record := range payload.Records {
			if err := standardizeRecord(ctx, tx, viewer, batch, record, now); err != nil {
				failed++
				rowNumber := int64(index + 1)
				reason := err.Error()
				raw := jsonString(record, "{}")
				if writeErr := tx.Create(&models.DataCenterRawDataError{
					TenantID: viewer.TenantID, BatchID: batch.ID, BatchCode: batch.BatchCode, RowNumber: &rowNumber,
					ErrorCode: "STANDARDIZE_FAILED", ErrorReason: reason, RawPayload: raw, CreatedAt: now,
				}).Error; writeErr != nil {
					return writeErr
				}
				continue
			}
			success++
		}
		status := domain.StatusSuccess
		if failed > 0 && success > 0 {
			status = domain.StatusWarning
		}
		if failed > 0 && success == 0 {
			status = domain.StatusFailed
		}
		return tx.Model(&batch).Updates(map[string]interface{}{
			"success_count": success, "failed_count": failed, "status": status, "updated_by": viewer.UserID, "updated_at": now,
		}).Error
	})
	if err != nil {
		return batch, err
	}
	_ = s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ?", viewer.TenantID, batch.ID).First(&batch).Error
	return batch, nil
}

func (s *Service) ReprocessRawBatch(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterRawDataBatch, error) {
	var batch models.DataCenterRawDataBatch
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return batch, ErrNotFound
		}
		return batch, err
	}
	var failedRows []models.DataCenterRawDataError
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND batch_id = ?", viewer.TenantID, batch.ID).Order("id asc").Find(&failedRows).Error; err != nil {
		return batch, err
	}
	if len(failedRows) == 0 {
		return batch, nil
	}
	now := time.Now()
	retrySuccess, retryFailed := int64(0), int64(0)
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&batch).Updates(map[string]interface{}{"status": domain.StatusProcessing, "updated_by": viewer.UserID, "updated_at": now}).Error; err != nil {
			return err
		}
		for _, failed := range failedRows {
			var record map[string]interface{}
			if err := json.Unmarshal([]byte(failed.RawPayload), &record); err != nil {
				retryFailed++
				continue
			}
			if err := standardizeRecord(ctx, tx, viewer, batch, record, now); err != nil {
				retryFailed++
				reason := "重试失败：" + err.Error()
				raw := jsonString(record, "{}")
				if writeErr := tx.Create(&models.DataCenterRawDataError{
					TenantID: viewer.TenantID, BatchID: batch.ID, BatchCode: batch.BatchCode, RowNumber: failed.RowNumber,
					ErrorCode: "REPROCESS_FAILED", ErrorReason: reason, RawPayload: raw, CreatedAt: now,
				}).Error; writeErr != nil {
					return writeErr
				}
				continue
			}
			retrySuccess++
		}
		successCount := batch.SuccessCount + retrySuccess
		failedCount := batch.FailedCount + retryFailed - retrySuccess
		if failedCount < 0 {
			failedCount = 0
		}
		status := domain.StatusSuccess
		if failedCount > 0 && successCount > 0 {
			status = domain.StatusWarning
		}
		if failedCount > 0 && successCount == 0 {
			status = domain.StatusFailed
		}
		return tx.Model(&batch).Updates(map[string]interface{}{
			"success_count": successCount, "failed_count": failedCount, "status": status, "updated_by": viewer.UserID, "updated_at": now,
		}).Error
	})
	if err != nil {
		return batch, err
	}
	_ = s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ?", viewer.TenantID, batch.ID).First(&batch).Error
	return batch, nil
}

func standardizeRecord(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, batch models.DataCenterRawDataBatch, record map[string]interface{}, now time.Time) error {
	switch batch.DataType {
	case "sales":
		orderCode := mapString(record, "order_code")
		orderTime := mapTime(record, "order_time")
		if orderCode == "" || orderTime.IsZero() {
			return errors.New("sales 记录必须包含 order_code 和 order_time")
		}
		row := models.DataCenterStdSalesOrder{
			TenantID: viewer.TenantID, OrderCode: orderCode, BrandCode: mapStringPtr(record, "brand_code"), BrandName: mapStringPtr(record, "brand_name"),
			ChannelCode: mapStringPtr(record, "channel_code"), PlatformCode: mapStringPtr(record, "platform_code"), ResourceCode: mapStringPtr(record, "resource_code"),
			ResourceName: mapStringPtr(record, "resource_name"), ProductCode: mapStringPtr(record, "product_code"), ProductName: mapStringPtr(record, "product_name"),
			SKUCode: mapStringPtr(record, "sku_code"), SalesAmount: mapFloat(record, "sales_amount"), PaidAmount: mapFloat(record, "paid_amount"),
			RefundAmount: mapFloat(record, "refund_amount"), OrderStatus: mapStringPtr(record, "order_status"), OrderTime: orderTime,
			SourceBatchCode: &batch.BatchCode, CreatedAt: now, UpdatedAt: now,
		}
		return tx.WithContext(ctx).Where("tenant_id = ? AND order_code = ? AND deleted_at IS NULL", viewer.TenantID, orderCode).Assign(row).FirstOrCreate(&row).Error
	case "ad":
		statDate := mapTime(record, "stat_date")
		accountCode, campaignCode := mapString(record, "account_code"), mapString(record, "campaign_code")
		if statDate.IsZero() || accountCode == "" || campaignCode == "" {
			return errors.New("ad 记录必须包含 stat_date、account_code 和 campaign_code")
		}
		cost, orderAmount := mapFloat(record, "cost_amount"), mapFloat(record, "order_amount")
		clicks, impressions, orders := mapInt(record, "click_count"), mapInt(record, "impression_count"), mapInt(record, "order_count")
		row := models.DataCenterStdAdDaily{
			TenantID: viewer.TenantID, BrandCode: mapStringPtr(record, "brand_code"), BrandName: mapStringPtr(record, "brand_name"),
			PlatformCode: mapStringPtr(record, "platform_code"), AccountCode: accountCode, AccountName: mapStringPtr(record, "account_name"),
			CampaignCode: campaignCode, CampaignName: mapStringPtr(record, "campaign_name"), CreativeCode: mapStringPtr(record, "creative_code"),
			ProductCode: mapStringPtr(record, "product_code"), StatDate: dateOnly(statDate), CostAmount: cost, ImpressionCount: impressions,
			ClickCount: clicks, OrderAmount: orderAmount, OrderCount: orders, ROI: ratio(orderAmount, cost),
			ClickRate: ratio(float64(clicks), float64(impressions)) * 100, ConversionRate: ratio(float64(orders), float64(clicks)) * 100,
			SourceBatchCode: &batch.BatchCode, CreatedAt: now, UpdatedAt: now,
		}
		return tx.WithContext(ctx).Where("tenant_id = ? AND account_code = ? AND campaign_code = ? AND stat_date = ? AND deleted_at IS NULL", viewer.TenantID, accountCode, campaignCode, dateOnly(statDate)).Assign(row).FirstOrCreate(&row).Error
	case "inventory":
		statDate := mapTime(record, "stat_date")
		productCode, skuCode := mapString(record, "product_code"), mapString(record, "sku_code")
		storeCode := defaultString(mapString(record, "store_code"), "")
		if statDate.IsZero() || productCode == "" || skuCode == "" {
			return errors.New("inventory 记录必须包含 stat_date、product_code 和 sku_code")
		}
		row := models.DataCenterStdInventoryDaily{
			TenantID: viewer.TenantID, BrandCode: mapStringPtr(record, "brand_code"), ProductCode: productCode, ProductName: mapStringPtr(record, "product_name"),
			SKUCode: skuCode, WarehouseCode: mapStringPtr(record, "warehouse_code"), StoreCode: storeCode, StatDate: dateOnly(statDate),
			AvailableStock: mapInt(record, "available_stock"), InTransitStock: mapInt(record, "in_transit_stock"), Sales7D: mapInt(record, "sales_7d"),
			AvailableDays: mapFloat(record, "available_days"), InventoryStatus: mapStringPtr(record, "inventory_status"),
			SourceBatchCode: &batch.BatchCode, CreatedAt: now, UpdatedAt: now,
		}
		return tx.WithContext(ctx).Where("tenant_id = ? AND product_code = ? AND sku_code = ? AND store_code = ? AND stat_date = ? AND deleted_at IS NULL", viewer.TenantID, productCode, skuCode, storeCode, dateOnly(statDate)).Assign(row).FirstOrCreate(&row).Error
	default:
		return standardizeSupplementalRecord(ctx, tx, viewer, batch, record, now)
	}
}

func standardizeSupplementalRecord(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, batch models.DataCenterRawDataBatch, record map[string]interface{}, now time.Time) error {
	switch batch.DataType {
	case "refund":
		refundCode, refundTime := mapString(record, "refund_code"), mapTime(record, "refund_time")
		if refundCode == "" || refundTime.IsZero() {
			return errors.New("refund 记录必须包含 refund_code 和 refund_time")
		}
		row := models.DataCenterStdRefundOrder{TenantID: viewer.TenantID, RefundCode: refundCode, OrderCode: mapStringPtr(record, "order_code"), BrandCode: mapStringPtr(record, "brand_code"), BrandName: mapStringPtr(record, "brand_name"), ChannelCode: mapStringPtr(record, "channel_code"), PlatformCode: mapStringPtr(record, "platform_code"), ProductCode: mapStringPtr(record, "product_code"), ProductName: mapStringPtr(record, "product_name"), SKUCode: mapStringPtr(record, "sku_code"), RefundAmount: mapFloat(record, "refund_amount"), RefundReason: mapStringPtr(record, "refund_reason"), RefundStatus: mapStringPtr(record, "refund_status"), RefundTime: refundTime, SourceBatchCode: &batch.BatchCode, CreatedAt: now, UpdatedAt: now}
		return tx.WithContext(ctx).Where("tenant_id = ? AND refund_code = ? AND deleted_at IS NULL", viewer.TenantID, refundCode).Assign(row).FirstOrCreate(&row).Error
	case "product":
		productCode, productName := mapString(record, "product_code"), mapString(record, "product_name")
		if productCode == "" || productName == "" {
			return errors.New("product 记录必须包含 product_code 和 product_name")
		}
		row := models.DataCenterStdProduct{TenantID: viewer.TenantID, ProductCode: productCode, ProductName: productName, SKUCode: mapStringPtr(record, "sku_code"), BrandCode: mapStringPtr(record, "brand_code"), BrandName: mapStringPtr(record, "brand_name"), CategoryCode: mapStringPtr(record, "category_code"), CategoryName: mapStringPtr(record, "category_name"), ListPrice: mapFloat(record, "list_price"), CostPrice: mapFloat(record, "cost_price"), ProductStatus: defaultString(mapString(record, "product_status"), "active"), SourceBatchCode: &batch.BatchCode, CreatedAt: now, UpdatedAt: now}
		return tx.WithContext(ctx).Where("tenant_id = ? AND product_code = ? AND deleted_at IS NULL", viewer.TenantID, productCode).Assign(row).FirstOrCreate(&row).Error
	case "store-sales":
		statDate, storeCode := mapTime(record, "stat_date"), mapString(record, "store_code")
		if statDate.IsZero() || storeCode == "" {
			return errors.New("store-sales 记录必须包含 stat_date 和 store_code")
		}
		row := models.DataCenterStdStoreSalesDaily{TenantID: viewer.TenantID, BrandCode: mapStringPtr(record, "brand_code"), BrandName: mapStringPtr(record, "brand_name"), ChannelCode: mapStringPtr(record, "channel_code"), PlatformCode: mapStringPtr(record, "platform_code"), StoreCode: storeCode, StoreName: mapStringPtr(record, "store_name"), StatDate: dateOnly(statDate), GMV: mapFloat(record, "gmv"), NetSales: mapFloat(record, "net_sales"), OrderCount: mapInt(record, "order_count"), CustomerCount: mapInt(record, "customer_count"), RefundAmount: mapFloat(record, "refund_amount"), TargetAmount: mapFloatPtr(record, "target_amount"), SourceBatchCode: &batch.BatchCode, CreatedAt: now, UpdatedAt: now}
		return tx.WithContext(ctx).Where("tenant_id = ? AND store_code = ? AND stat_date = ? AND deleted_at IS NULL", viewer.TenantID, storeCode, dateOnly(statDate)).Assign(row).FirstOrCreate(&row).Error
	}
	return errors.New("不支持的数据类型")
}

func supportedImportType(dataType string) bool {
	switch dataType {
	case "sales", "ad", "inventory", "refund", "product", "store-sales":
		return true
	default:
		return false
	}
}

func firstRecord(records []map[string]interface{}) map[string]interface{} {
	if len(records) == 0 {
		return map[string]interface{}{}
	}
	return records[0]
}

func mapString(record map[string]interface{}, key string) string {
	value, ok := record[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func mapStringPtr(record map[string]interface{}, key string) *string {
	return optional(mapString(record, key))
}

func mapFloat(record map[string]interface{}, key string) float64 {
	value, ok := record[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case jsonNumber:
		parsed, _ := strconv.ParseFloat(string(typed), 64)
		return parsed
	default:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
		return parsed
	}
}

func mapFloatPtr(record map[string]interface{}, key string) *float64 {
	if _, ok := record[key]; !ok {
		return nil
	}
	value := mapFloat(record, key)
	return &value
}

func mapInt(record map[string]interface{}, key string) int64 {
	return int64(mapFloat(record, key))
}

func mapTime(record map[string]interface{}, key string) time.Time {
	value := mapString(record, key)
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

type jsonNumber string
