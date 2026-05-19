package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	aiccservices "saas_baseon_go/internal/apps/ai_capability_center/services"
	"saas_baseon_go/internal/apps/data_center/domain"
	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/apps/data_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestRawBatchesAreTenantScoped(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.DataCenterRawDataBatch{TenantID: 1, BatchCode: "RAW-1", DataType: "order", Status: "success", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DataCenterRawDataBatch{TenantID: 2, BatchCode: "RAW-2", DataType: "order", Status: "success", CreatedAt: now, UpdatedAt: now}).Error)

	result, err := NewService(repositories.NewRepository(db)).RawBatches(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, dto.PageRequest{Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	require.Equal(t, "RAW-1", result.Items[0].BatchCode)
}

func TestRecordAuditStoresOperatorTenantObjectAndRequestContext(t *testing.T) {
	db := newDataCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	meta := dto.RequestMeta{IP: "127.0.0.1", UserAgent: "data-center-test", RequestID: "rid-data-center"}

	err := service.RecordAudit(context.Background(), viewer, meta, "metric_create", "gmv", "创建指标定义", map[string]interface{}{"metric_name": "GMV"})

	require.NoError(t, err)
	var log models.AuditLog
	require.NoError(t, db.Where("tenant_id = ? AND user_id = ? AND module = ? AND action = ?", viewer.TenantID, viewer.UserID, "data_center", "metric_create").First(&log).Error)
	require.NotNil(t, log.AppCode)
	require.Equal(t, domain.AppCode, *log.AppCode)
	require.Equal(t, "创建指标定义", log.Summary)
	require.NotNil(t, log.IP)
	require.Equal(t, meta.IP, *log.IP)
	require.NotNil(t, log.UserAgent)
	require.Equal(t, meta.UserAgent, *log.UserAgent)
	require.NotNil(t, log.RequestID)
	require.Equal(t, meta.RequestID, *log.RequestID)
	require.NotNil(t, log.Detail)
	require.Contains(t, *log.Detail, `"object_code":"gmv"`)
	require.Contains(t, *log.Detail, `"metric_name":"GMV"`)
}

func TestOverviewRawErrorsListsTenantErrorsAcrossBatches(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.DataCenterRawDataError{TenantID: 1, BatchID: 10, BatchCode: "B-1", ErrorCode: "E", ErrorReason: "bad", RawPayload: "{}", CreatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DataCenterRawDataError{TenantID: 1, BatchID: 11, BatchCode: "B-2", ErrorCode: "E", ErrorReason: "bad", RawPayload: "{}", CreatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DataCenterRawDataError{TenantID: 2, BatchID: 12, BatchCode: "B-3", ErrorCode: "E", ErrorReason: "bad", RawPayload: "{}", CreatedAt: now}).Error)

	result, err := NewService(repositories.NewRepository(db)).RawErrors(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, 0, dto.PageRequest{Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)
}

func TestGenerateTaskRejectsDuplicateForSameAnomaly(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		TenantID: 1, AnomalyCode: "ANM-1", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 86, EvidenceJSON: "[]", OccurredAt: now,
		AIStatus: "pending", TaskStatus: "none", ReviewStatus: "none", Status: "pending",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&anomaly).Error)
	service := NewService(repositories.NewRepository(db))

	first, err := service.GenerateTask(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, anomaly.ID)
	require.NoError(t, err)
	require.NotZero(t, first.ID)
	require.Contains(t, first.AISuggestionJSON, "anomaly_code")
	_, err = service.GenerateTask(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, anomaly.ID)
	require.ErrorIs(t, err, ErrTaskExists)
}

func TestGenerateTaskCannotCrossTenant(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		TenantID: 2, AnomalyCode: "ANM-2", RuleCode: "gmv_drop", Title: "其他租户异常",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-b", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 86, EvidenceJSON: "[]", OccurredAt: now,
		AIStatus: "pending", TaskStatus: "none", ReviewStatus: "none", Status: "pending",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&anomaly).Error)

	_, err := NewService(repositories.NewRepository(db)).GenerateTask(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, anomaly.ID)

	require.ErrorIs(t, err, ErrNotFound)
}

func TestGatewayAIAnalyzerInvokesAICapabilityCenterScenario(t *testing.T) {
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		ID: 8, TenantID: 1, AnomalyCode: "ANM-1", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 88, EvidenceJSON: `[{"metric_code":"gmv","label":"GMV","value":500,"desc":"下滑"}]`,
	}
	rule := models.DataCenterAnomalyRule{RuleCode: "gmv_drop", RuleName: "GMV 下滑", TargetObjectType: "brand", MetricConditionsJSON: "{}"}
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"problem_summary": "AI 判断 GMV 异常",
			"reasons":         []interface{}{map[string]interface{}{"metric_code": "gmv", "desc": "环比下滑"}},
			"suggestions":     []interface{}{map[string]interface{}{"type": "action", "content": "复盘投放和转化"}},
		},
	}}

	result, err := NewGatewayAIAnalyzer(gateway).AnalyzeAnomaly(context.Background(), anomaly, rule)

	require.NoError(t, err)
	require.Equal(t, domain.AppCode, gateway.request.AppCode)
	require.Equal(t, anomalyAnalysisScenarioCode, gateway.request.AIScenarioCode)
	require.Equal(t, "1", gateway.request.TenantID)
	require.Equal(t, "AI 判断 GMV 异常", result.ProblemSummary)
	require.Len(t, result.Reasons, 1)
	require.Len(t, result.Suggestions, 1)
}

func TestGatewayAIAnalyzerParsesProviderTextJSON(t *testing.T) {
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		ID: 8, TenantID: 1, AnomalyCode: "ANM-1", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 88, EvidenceJSON: `[{"metric_code":"gmv","label":"GMV","value":500,"desc":"下滑"}]`,
	}
	rule := models.DataCenterAnomalyRule{RuleCode: "gmv_drop", RuleName: "GMV 下滑", TargetObjectType: "brand", MetricConditionsJSON: "{}"}
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"text": "```json\n{\"problem_summary\":\"模型判断 Lee 直营 GMV 低于目标\",\"impact_summary\":\"影响华东直营门店\",\"reasons\":[{\"type\":\"root_cause\",\"label\":\"折扣承接不足\",\"content\":\"新品折扣未覆盖目标店型\"}],\"suggestions\":[{\"type\":\"action\",\"content\":\"调整直营门店 Denim 系列折扣\"}],\"task_suggestion\":{\"title\":\"调整 Lee 直营折扣\",\"owner_role\":\"商品运营\",\"deadline_days\":2,\"priority\":\"high\",\"target_desc\":\"恢复目标店型 GMV\"},\"confidence_explanation\":\"规则命中且证据完整\"}\n```",
		},
	}}

	result, err := NewGatewayAIAnalyzer(gateway).AnalyzeAnomaly(context.Background(), anomaly, rule)

	require.NoError(t, err)
	require.Equal(t, "模型判断 Lee 直营 GMV 低于目标", result.ProblemSummary)
	require.Equal(t, "影响华东直营门店", result.ImpactSummary)
	require.Len(t, result.Reasons, 1)
	require.Len(t, result.Suggestions, 1)
	require.Equal(t, "调整 Lee 直营折扣", result.TaskSuggestion["title"])
}

func TestScanAnomaliesCalculatesMetricsAndDedupes(t *testing.T) {
	db := newDataCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	now := time.Now()
	current := now.Add(-1 * time.Hour)
	previous := current.AddDate(0, 0, -1)
	brandCode, brandName := "brand-a", "Brand A"
	campaignCode, campaignName := "campaign-a", "Campaign A"
	productCode, productName, skuCode, storeCode := "product-a", "Product A", "sku-a", "store-a"

	require.NoError(t, db.Create(&models.DataCenterStdSalesOrder{TenantID: 1, OrderCode: "O-1", BrandCode: &brandCode, BrandName: &brandName, SalesAmount: 1000, PaidAmount: 1000, OrderTime: previous, CreatedAt: previous, UpdatedAt: previous}).Error)
	require.NoError(t, db.Create(&models.DataCenterStdSalesOrder{TenantID: 1, OrderCode: "O-2", BrandCode: &brandCode, BrandName: &brandName, SalesAmount: 1000, PaidAmount: 1000, OrderTime: previous, CreatedAt: previous, UpdatedAt: previous}).Error)
	require.NoError(t, db.Create(&models.DataCenterStdSalesOrder{TenantID: 1, OrderCode: "O-3", BrandCode: &brandCode, BrandName: &brandName, SalesAmount: 500, PaidAmount: 500, OrderTime: current, CreatedAt: current, UpdatedAt: current}).Error)
	targetAmount := 900.0
	require.NoError(t, db.Create(&models.DataCenterStdStoreSalesDaily{TenantID: 1, StoreCode: "store-a", StoreName: &storeCode, BrandCode: &brandCode, BrandName: &brandName, StatDate: current, GMV: 500, NetSales: 500, OrderCount: 1, TargetAmount: &targetAmount, CreatedAt: current, UpdatedAt: current}).Error)
	require.NoError(t, db.Create(&models.DataCenterStdAdDaily{TenantID: 1, AccountCode: "A-1", CampaignCode: campaignCode, CampaignName: &campaignName, StatDate: previous, CostAmount: 100, OrderAmount: 400, ClickCount: 100, ImpressionCount: 1000, OrderCount: 10, CreatedAt: previous, UpdatedAt: previous}).Error)
	require.NoError(t, db.Create(&models.DataCenterStdAdDaily{TenantID: 1, AccountCode: "A-1", CampaignCode: campaignCode, CampaignName: &campaignName, StatDate: current, CostAmount: 150, OrderAmount: 300, ClickCount: 100, ImpressionCount: 1000, OrderCount: 10, CreatedAt: current, UpdatedAt: current}).Error)
	require.NoError(t, db.Create(&models.DataCenterStdInventoryDaily{TenantID: 1, ProductCode: productCode, ProductName: &productName, SKUCode: skuCode, StoreCode: storeCode, StatDate: current, AvailableStock: 200, Sales7D: 2, AvailableDays: 80, CreatedAt: current, UpdatedAt: current}).Error)

	first, err := service.ScanAnomalies(context.Background(), viewer, dto.PageRequest{TimeRange: "today", Limit: 20})
	require.NoError(t, err)
	require.Equal(t, 3, first.Generated)
	require.GreaterOrEqual(t, first.MetricsGenerated, 10)

	second, err := service.ScanAnomalies(context.Background(), viewer, dto.PageRequest{TimeRange: "today", Limit: 20})
	require.NoError(t, err)
	require.Equal(t, 0, second.Generated)
	require.Equal(t, 3, second.Duplicated)

	var anomalies int64
	require.NoError(t, db.Model(&models.DataCenterAnomalyRecord{}).Where("tenant_id = ?", viewer.TenantID).Count(&anomalies).Error)
	require.Equal(t, int64(3), anomalies)

	var adGMV models.DataCenterMetricResult
	require.NoError(t, db.Where("tenant_id = ? AND metric_code = ?", viewer.TenantID, "ad_gmv").First(&adGMV).Error)
	require.Equal(t, 300.0, adGMV.MetricValue)
	require.NotNil(t, adGMV.CompareValue)
	require.NotNil(t, adGMV.CompareRate)
	var gmv models.DataCenterMetricResult
	require.NoError(t, db.Where("tenant_id = ? AND metric_code = ?", viewer.TenantID, "gmv").First(&gmv).Error)
	require.NotNil(t, gmv.TargetValue)
	require.Equal(t, targetAmount, *gmv.TargetValue)
	var sellThrough models.DataCenterMetricResult
	require.NoError(t, db.Where("tenant_id = ? AND metric_code = ?", viewer.TenantID, "sell_through_rate").First(&sellThrough).Error)
	require.Greater(t, sellThrough.MetricValue, 0.0)
}

func TestScanQuotaFailureDoesNotCreateBusinessState(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.SaasQuota{QuotaCode: "data_center_daily_scans", QuotaName: "每日异常扫描次数", QuotaType: "USAGE", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	_, err := NewService(repositories.NewRepository(db)).ScanAnomalies(context.Background(), viewer, dto.PageRequest{TimeRange: "today", Limit: 20})

	require.Error(t, err)
	var metrics int64
	require.NoError(t, db.Model(&models.DataCenterMetricResult{}).Where("tenant_id = ?", viewer.TenantID).Count(&metrics).Error)
	require.Equal(t, int64(0), metrics)
	var usage int64
	require.NoError(t, db.Model(&models.TenantQuotaUsage{}).Where("tenant_id = ? AND quota_code = ?", viewer.TenantID, "data_center_daily_scans").Count(&usage).Error)
	require.Equal(t, int64(0), usage)
}

func TestRuleConditionEngineSupportsAndOrBetweenAndMiss(t *testing.T) {
	raw := `{
		"logic":"AND",
		"conditions":[
			{"metric_code":"gmv","operator":"lt","value":1000},
			{
				"logic":"OR",
				"conditions":[
					{"metric_code":"order_count","operator":"lte","value":10},
					{"metric_code":"refund_rate","operator":"between","min":5,"max":20}
				]
			}
		]
	}`
	ok, err := evaluateRuleConditions(raw, map[string]float64{"gmv": 900, "order_count": 20, "refund_rate": 8})
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = evaluateRuleConditions(raw, map[string]float64{"gmv": 1200, "order_count": 2, "refund_rate": 8})
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = evaluateRuleConditions(raw, map[string]float64{"gmv": 900, "order_count": 20, "refund_rate": 30})
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAnomalyLevelCalculationSupportsFourLevels(t *testing.T) {
	require.Equal(t, "low", levelByNegativeChange(-5))
	require.Equal(t, "medium", levelByNegativeChange(-20))
	require.Equal(t, "high", levelByNegativeChange(-35))
	require.Equal(t, "critical", levelByNegativeChange(-55))

	require.Equal(t, "low", levelByAvailableDays(20))
	require.Equal(t, "medium", levelByAvailableDays(45))
	require.Equal(t, "high", levelByAvailableDays(75))
	require.Equal(t, "critical", levelByAvailableDays(120))
}

func TestTestRuleEvaluatesPersistedConditions(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	rule := models.DataCenterAnomalyRule{
		TenantID: 1, RuleCode: "roi_drop", RuleName: "ROI 下滑", BusinessDomain: "投流异常", TargetObjectType: "campaign",
		MetricConditionsJSON: `{"logic":"AND","conditions":[{"metric_code":"ad_cost","operator":"gt","value":100},{"metric_code":"roi","operator":"lt","value":2}]}`,
		LevelConfigJSON:      "{}", ConfidenceConfigJSON: "{}", AIEnabled: true, TaskEnabled: true,
		DefaultDeadlineDays: 3, ReviewMetricCodes: "[]", Enabled: true, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&rule).Error)

	result, err := NewService(repositories.NewRepository(db)).TestRule(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, rule.ID, dto.RuleTestPayload{
		Metrics: map[string]float64{"ad_cost": 150, "roi": 1.6},
	})

	require.NoError(t, err)
	require.Equal(t, true, result["matched"])
	require.Len(t, result["evidence"], 2)
}

func TestMetricRuleTaskDetailsAndUpdatesAreTenantScoped(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 77}
	otherTenantMetric := models.DataCenterMetricDefinition{
		TenantID: 2, MetricCode: "gmv", MetricName: "其他租户", MetricCategory: "sales", Dimensions: "[]",
		Enabled: true, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&otherTenantMetric).Error)

	metric, err := service.CreateMetric(context.Background(), viewer, dto.MetricDefinitionPayload{
		MetricCode: "gmv", MetricName: "GMV", MetricCategory: "sales", Dimensions: []string{"brand"},
	})
	require.NoError(t, err)
	updatedMetric, err := service.UpdateMetric(context.Background(), viewer, metric.ID, dto.MetricDefinitionPayload{MetricName: "GMV 更新"})
	require.NoError(t, err)
	require.Equal(t, "GMV 更新", updatedMetric.MetricName)
	require.NotNil(t, updatedMetric.UpdatedBy)
	require.Equal(t, uint64(77), *updatedMetric.UpdatedBy)

	_, err = service.Metric(context.Background(), viewer, otherTenantMetric.ID)
	require.ErrorIs(t, err, ErrNotFound)

	rule, err := service.CreateRule(context.Background(), viewer, dto.AnomalyRulePayload{
		RuleCode: "gmv_drop", RuleName: "GMV 下滑", BusinessDomain: "销售异常", TargetObjectType: "brand",
		MetricConditions: map[string]interface{}{"metric_code": "gmv", "operator": "lt", "value": 1000},
	})
	require.NoError(t, err)
	updatedRule, err := service.UpdateRule(context.Background(), viewer, rule.ID, dto.AnomalyRulePayload{RuleName: "GMV 大幅下滑"})
	require.NoError(t, err)
	require.Equal(t, "GMV 大幅下滑", updatedRule.RuleName)

	task, err := service.CreateTask(context.Background(), viewer, dto.TaskPayload{Title: "人工整改", Priority: "high", TargetDesc: "跟进异常原因"})
	require.NoError(t, err)
	updatedTask, err := service.UpdateTask(context.Background(), viewer, task.ID, dto.TaskPayload{Progress: intPtr(60), ExecutionFeedback: "已定位渠道波动"})
	require.NoError(t, err)
	require.Equal(t, 60, updatedTask.Progress)
	require.NotNil(t, updatedTask.ExecutionFeedback)
}

func TestCreateRawBatchStandardizesSalesAndRecordsErrors(t *testing.T) {
	db := newDataCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	batch, err := service.CreateRawBatch(context.Background(), viewer, dto.RawBatchPayload{
		BatchCode: "BATCH-SALES-1",
		DataType:  "sales",
		Records: []map[string]interface{}{
			{"order_code": "O-IMPORT-1", "order_time": "2026-05-18T01:00:00+08:00", "sales_amount": 1200, "paid_amount": 1100, "brand_code": "brand-a"},
			{"order_code": "O-IMPORT-BAD"},
		},
	})

	require.NoError(t, err)
	require.Equal(t, domain.StatusWarning, batch.Status)
	require.Equal(t, int64(1), batch.SuccessCount)
	require.Equal(t, int64(1), batch.FailedCount)

	var orders int64
	require.NoError(t, db.Model(&models.DataCenterStdSalesOrder{}).Where("tenant_id = ?", viewer.TenantID).Count(&orders).Error)
	require.Equal(t, int64(1), orders)

	var errorsCount int64
	require.NoError(t, db.Model(&models.DataCenterRawDataError{}).Where("tenant_id = ? AND batch_id = ?", viewer.TenantID, batch.ID).Count(&errorsCount).Error)
	require.Equal(t, int64(1), errorsCount)
}

func TestReprocessRawBatchRetriesFailedRowsAndKeepsHistory(t *testing.T) {
	db := newDataCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	batch, err := service.CreateRawBatch(context.Background(), viewer, dto.RawBatchPayload{
		BatchCode: "BATCH-REPROCESS-1",
		DataType:  "sales",
		Records: []map[string]interface{}{
			{"order_code": "O-RETRY-1"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, domain.StatusFailed, batch.Status)

	var failed models.DataCenterRawDataError
	require.NoError(t, db.Where("tenant_id = ? AND batch_id = ?", viewer.TenantID, batch.ID).First(&failed).Error)
	fixedPayload := `{"order_code":"O-RETRY-1","order_time":"2026-05-18T01:00:00+08:00","sales_amount":300,"paid_amount":300}`
	require.NoError(t, db.Model(&failed).Update("raw_payload", fixedPayload).Error)

	reprocessed, err := service.ReprocessRawBatch(context.Background(), viewer, batch.ID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusSuccess, reprocessed.Status)
	require.Equal(t, int64(1), reprocessed.SuccessCount)
	require.Equal(t, int64(0), reprocessed.FailedCount)

	var orders int64
	require.NoError(t, db.Model(&models.DataCenterStdSalesOrder{}).Where("tenant_id = ? AND order_code = ?", viewer.TenantID, "O-RETRY-1").Count(&orders).Error)
	require.Equal(t, int64(1), orders)
	var historicalErrors int64
	require.NoError(t, db.Model(&models.DataCenterRawDataError{}).Where("tenant_id = ? AND batch_id = ?", viewer.TenantID, batch.ID).Count(&historicalErrors).Error)
	require.Equal(t, int64(1), historicalErrors)
}

func TestTransitionTaskControlsFlowAndWritesLog(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	task := models.DataCenterRectificationTask{
		TenantID: 1, TaskCode: "TASK-1", Title: "整改任务", TaskType: "anomaly_rectification",
		Status: domain.TaskStatusPending, ReviewStatus: domain.AnomalyReviewStatusNone, CollaboratorIDs: "[]",
		AISuggestionJSON: "{}", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&task).Error)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	_, err := service.TransitionTask(context.Background(), viewer, task.ID, domain.TaskStatusCompleted, "", nil)
	require.ErrorIs(t, err, ErrInvalidStatus)

	processing, err := service.TransitionTask(context.Background(), viewer, task.ID, domain.TaskStatusProcessing, "", nil)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusProcessing, processing.Status)

	completed, err := service.TransitionTask(context.Background(), viewer, task.ID, domain.TaskStatusCompleted, "", nil)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusCompleted, completed.Status)
	require.Equal(t, domain.AnomalyReviewStatusPending, completed.ReviewStatus)
	require.Equal(t, 100, completed.Progress)

	var logs int64
	require.NoError(t, db.Model(&models.DataCenterTaskLog{}).Where("tenant_id = ? AND task_id = ?", viewer.TenantID, task.ID).Count(&logs).Error)
	require.Equal(t, int64(2), logs)
}

func TestAnalyzeAnomalyStoresStructuredEvidenceAndKeepsHistory(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	ownerRole := "品牌负责人"
	rule := models.DataCenterAnomalyRule{
		TenantID: 1, RuleCode: "gmv_drop", RuleName: "GMV 下滑", BusinessDomain: "销售异常", TargetObjectType: "brand",
		MetricConditionsJSON: "{}", LevelConfigJSON: "{}", ConfidenceConfigJSON: "{}", AIEnabled: true, TaskEnabled: true,
		DefaultOwnerRole: &ownerRole, DefaultDeadlineDays: 2, ReviewMetricCodes: "[]", Enabled: true, CreatedAt: now, UpdatedAt: now,
	}
	anomaly := models.DataCenterAnomalyRecord{
		TenantID: 1, AnomalyCode: "ANM-AI", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 86, ImpactAmount: 1200,
		EvidenceJSON: `[{"metric_code":"gmv","label":"GMV","value":500,"desc":"较昨日下降"}]`,
		OccurredAt:   now, AIStatus: "pending", TaskStatus: "none", ReviewStatus: "none", Status: "pending",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&rule).Error)
	require.NoError(t, db.Create(&anomaly).Error)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	first, err := service.AnalyzeAnomaly(context.Background(), viewer, anomaly.ID, false)
	require.NoError(t, err)
	require.Equal(t, domain.AnomalyAIStatusSuccess, first.Status)
	require.JSONEq(t, anomaly.EvidenceJSON, first.EvidenceSummaryJSON)
	require.Contains(t, first.TaskSuggestionJSON, "品牌负责人")

	second, err := service.AnalyzeAnomaly(context.Background(), viewer, anomaly.ID, false)
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)

	third, err := service.AnalyzeAnomaly(context.Background(), viewer, anomaly.ID, true)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, third.ID)
}

func TestAnalyzeAnomalyRecordsFailedStatusAndSanitizesErrors(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		TenantID: 1, AnomalyCode: "ANM-AI-FAILED", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 86, EvidenceJSON: "[]", OccurredAt: now,
		AIStatus: "pending", TaskStatus: "none", ReviewStatus: "none", Status: "pending",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&anomaly).Error)
	service := NewService(repositories.NewRepository(db))
	service.SetAIAnalyzer(failingAIAnalyzer{err: errors.New("provider token leaked in stack trace with SELECT * FROM secrets")})

	record, err := service.AnalyzeAnomaly(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, anomaly.ID, true)

	require.Error(t, err)
	require.Equal(t, domain.AnomalyAIStatusFailed, record.Status)
	require.NotNil(t, record.ErrorMessage)
	require.NotContains(t, *record.ErrorMessage, "token")
	require.NotContains(t, *record.ErrorMessage, "SELECT")
	var refreshed models.DataCenterAnomalyRecord
	require.NoError(t, db.First(&refreshed, anomaly.ID).Error)
	require.Equal(t, domain.AnomalyAIStatusFailed, refreshed.AIStatus)
}

func TestCreateReviewCalculatesMetricsAndUpdatesStatuses(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		TenantID: 1, AnomalyCode: "ANM-REV", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now.AddDate(0, 0, -1),
		AnomalyLevel: "high", ConfidenceScore: 86, ImpactAmount: 500,
		EvidenceJSON: `[{"metric_code":"gmv","label":"GMV","value":500,"desc":"整改前"}]`,
		OccurredAt:   now, AIStatus: "success", TaskStatus: "generated", ReviewStatus: "pending", Status: "processing",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&anomaly).Error)
	require.NoError(t, db.Create(&models.DataCenterMetricResult{
		TenantID: 1, MetricCode: "gmv", ResourceType: "brand", ResourceCode: "brand-a", StatDate: now,
		PeriodType: "daily", MetricValue: 800, DimensionJSON: "{}", CreatedAt: now, UpdatedAt: now,
	}).Error)
	task := models.DataCenterRectificationTask{
		TenantID: 1, TaskCode: "TASK-REV", AnomalyID: &anomaly.ID, AnomalyCode: &anomaly.AnomalyCode,
		Title: "整改任务", TaskType: "anomaly_rectification", Status: domain.TaskStatusCompleted,
		ReviewStatus: domain.AnomalyReviewStatusPending, CollaboratorIDs: "[]", AISuggestionJSON: "{}", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&task).Error)

	review, err := NewService(repositories.NewRepository(db)).CreateReview(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, dto.ReviewPayload{TaskID: task.ID})
	require.NoError(t, err)
	require.Equal(t, domain.ReviewConclusionEffective, review.ReviewConclusion)
	require.NotNil(t, review.ImprovementResult)
	require.Contains(t, *review.ImprovementResult, "60.00%")

	var after []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(review.AfterMetricJSON), &after))
	require.Len(t, after, 1)
	require.Equal(t, float64(800), after[0]["value"])

	var refreshedTask models.DataCenterRectificationTask
	require.NoError(t, db.First(&refreshedTask, task.ID).Error)
	require.Equal(t, domain.AnomalyReviewStatusReviewed, refreshedTask.ReviewStatus)
	var refreshedAnomaly models.DataCenterAnomalyRecord
	require.NoError(t, db.First(&refreshedAnomaly, anomaly.ID).Error)
	require.Equal(t, domain.AnomalyReviewStatusReviewed, refreshedAnomaly.ReviewStatus)
}

func TestConfirmReviewPersistsManualConclusionAndOperator(t *testing.T) {
	db := newDataCenterTestDB(t)
	now := time.Now()
	anomaly := models.DataCenterAnomalyRecord{
		TenantID: 1, AnomalyCode: "ANM-CONFIRM", RuleCode: "gmv_drop", Title: "GMV 下滑",
		BusinessDomain: "销售异常", ObjectType: "brand", ObjectCode: "brand-a", StatDate: now,
		AnomalyLevel: "high", ConfidenceScore: 86, EvidenceJSON: "[]", OccurredAt: now,
		AIStatus: "success", TaskStatus: "generated", ReviewStatus: "pending", Status: "processing",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&anomaly).Error)
	task := models.DataCenterRectificationTask{
		TenantID: 1, TaskCode: "TASK-CONFIRM", AnomalyID: &anomaly.ID, AnomalyCode: &anomaly.AnomalyCode,
		Title: "整改任务", TaskType: "anomaly_rectification", Status: domain.TaskStatusCompleted,
		ReviewStatus: domain.AnomalyReviewStatusPending, CollaboratorIDs: "[]", AISuggestionJSON: "{}", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&task).Error)
	review := models.DataCenterRectificationReview{
		TenantID: 1, ReviewCode: "REV-CONFIRM", TaskID: task.ID, TaskCode: task.TaskCode,
		AnomalyID: &anomaly.ID, AnomalyCode: &anomaly.AnomalyCode, BeforeMetricJSON: "[]", AfterMetricJSON: "[]",
		ReviewConclusion: domain.ReviewConclusionWeak, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&review).Error)
	service := NewService(repositories.NewRepository(db))

	confirmed, err := service.ConfirmReview(context.Background(), dto.Viewer{TenantID: 1, UserID: 66}, review.ID, dto.ReviewPayload{
		ReviewConclusion:    domain.ReviewConclusionFollowUp,
		ManualReviewSummary: "需要继续观察一周",
		ExperienceSummary:   "沉淀投流预算调整 SOP",
	})

	require.NoError(t, err)
	require.Equal(t, domain.ReviewConclusionFollowUp, confirmed.ReviewConclusion)
	require.NotNil(t, confirmed.ManualReviewSummary)
	require.Equal(t, "需要继续观察一周", *confirmed.ManualReviewSummary)
	require.NotNil(t, confirmed.ExperienceSummary)
	require.NotNil(t, confirmed.ReviewedAt)
	require.NotNil(t, confirmed.UpdatedBy)
	require.Equal(t, uint64(66), *confirmed.UpdatedBy)

	var refreshedTask models.DataCenterRectificationTask
	require.NoError(t, db.First(&refreshedTask, task.ID).Error)
	require.Equal(t, domain.AnomalyReviewStatusReviewed, refreshedTask.ReviewStatus)
	var refreshedAnomaly models.DataCenterAnomalyRecord
	require.NoError(t, db.First(&refreshedAnomaly, anomaly.ID).Error)
	require.Equal(t, domain.AnomalyReviewStatusReviewed, refreshedAnomaly.ReviewStatus)
}

func newDataCenterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(
		&models.DataCenterRawDataBatch{},
		&models.DataCenterRawDataError{},
		&models.DataCenterStdSalesOrder{},
		&models.DataCenterStdAdDaily{},
		&models.DataCenterStdInventoryDaily{},
		&models.DataCenterStdRefundOrder{},
		&models.DataCenterStdProduct{},
		&models.DataCenterStdStoreSalesDaily{},
		&models.DataCenterMetricDefinition{},
		&models.DataCenterMetricResult{},
		&models.DataCenterAnomalyRule{},
		&models.DataCenterAnomalyRecord{},
		&models.DataCenterAIDiagnosisRecord{},
		&models.DataCenterRectificationTask{},
		&models.DataCenterTaskLog{},
		&models.DataCenterRectificationReview{},
		&models.AuditLog{},
		&models.SaasQuota{},
		&models.SaasPlan{},
		&models.SaasPlanQuota{},
		&models.TenantSubscription{},
		&models.TenantQuotaOverride{},
		&models.TenantQuotaUsage{},
	))
	return db
}

func intPtr(value int) *int {
	return &value
}

type failingAIAnalyzer struct {
	err error
}

func (f failingAIAnalyzer) AnalyzeAnomaly(context.Context, models.DataCenterAnomalyRecord, models.DataCenterAnomalyRule) (localDiagnosis, error) {
	return localDiagnosis{}, f.err
}

type fakeAIGatewayInvoker struct {
	request  aiccservices.InvokeRequest
	response aiccservices.InvokeResponse
	err      error
}

func (f *fakeAIGatewayInvoker) Invoke(_ context.Context, req aiccservices.InvokeRequest) (aiccservices.InvokeResponse, error) {
	f.request = req
	return f.response, f.err
}
