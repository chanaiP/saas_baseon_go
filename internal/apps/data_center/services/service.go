package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	quotaapp "saas_baseon_go/internal/application/quota"
	"saas_baseon_go/internal/apps/data_center/domain"
	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/apps/data_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

var (
	ErrForbidden     = errors.New("没有Ai经营决策中心操作权限")
	ErrNotFound      = errors.New("数据不存在")
	ErrInvalidStatus = errors.New("状态流转不合法")
	ErrTaskExists    = errors.New("该异常已生成整改任务")
)

type Service struct {
	repo       *repositories.Repository
	quota      *quotaapp.Service
	aiAnalyzer AIAnalyzer
}

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo, quota: quotaapp.NewService(repo.DB()), aiAnalyzer: localAIAnalyzer{}}
}

type AIAnalyzer interface {
	AnalyzeAnomaly(ctx context.Context, anomaly models.DataCenterAnomalyRecord, rule models.DataCenterAnomalyRule) (localDiagnosis, error)
}

type localAIAnalyzer struct{}

func (localAIAnalyzer) AnalyzeAnomaly(_ context.Context, anomaly models.DataCenterAnomalyRecord, rule models.DataCenterAnomalyRule) (localDiagnosis, error) {
	return buildLocalDiagnosis(anomaly, rule), nil
}

func (s *Service) SetAIAnalyzer(analyzer AIAnalyzer) {
	if analyzer == nil {
		s.aiAnalyzer = localAIAnalyzer{}
		return
	}
	s.aiAnalyzer = analyzer
}

func (s *Service) RecordAudit(ctx context.Context, viewer dto.Viewer, meta dto.RequestMeta, action string, objectCode string, summary string, detail interface{}) error {
	appCode := domain.AppCode
	auditDetail := map[string]interface{}{
		"object_code": objectCode,
	}
	if detail != nil {
		auditDetail["detail"] = detail
	}
	detailJSON := jsonString(auditDetail, "{}")
	now := time.Now()
	return s.repo.DB().WithContext(ctx).Create(&models.AuditLog{
		TenantID:  &viewer.TenantID,
		UserID:    &viewer.UserID,
		AppCode:   &appCode,
		Module:    "data_center",
		Action:    action,
		Summary:   summary,
		Detail:    optional(detailJSON),
		IP:        optional(meta.IP),
		UserAgent: optional(meta.UserAgent),
		RequestID: optional(meta.RequestID),
		Result:    "success",
		CreatedAt: now,
	}).Error
}

func (s *Service) consumeQuota(ctx context.Context, tenantID uint64, quotaCode string) error {
	if s.quota == nil {
		return nil
	}
	return s.quota.Consume(ctx, tenantID, quotaCode, 1)
}

func (s *Service) ensureActiveRuleQuota(ctx context.Context, tenantID uint64) error {
	var quota models.SaasQuota
	if err := s.repo.DB().WithContext(ctx).Where("quota_code = ? AND status = ?", "data_center_active_rules", 1).First(&quota).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	limit := currentDataCenterQuotaLimit(s.repo.DB().WithContext(ctx), tenantID, quota.ID)
	if limit < 0 {
		return nil
	}
	var count int64
	if err := s.repo.DB().WithContext(ctx).Model(&models.DataCenterAnomalyRule{}).
		Where("tenant_id = ? AND enabled = ? AND deleted_at IS NULL", tenantID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if int(count)+1 > limit {
		return &quotaapp.ExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: int(count)}
	}
	return nil
}

func (s *Service) DashboardSummary(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (map[string]interface{}, error) {
	sales, err := s.repo.SalesSummary(ctx, viewer.TenantID, req)
	if err != nil {
		return nil, err
	}
	ad, err := s.repo.AdSummary(ctx, viewer.TenantID, req)
	if err != nil {
		return nil, err
	}
	anomalyCount, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterAnomalyRecord{}, "")
	if err != nil {
		return nil, err
	}
	completedTasks, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterRectificationTask{}, domain.TaskStatusCompleted)
	if err != nil {
		return nil, err
	}
	allTasks, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterRectificationTask{}, "")
	if err != nil {
		return nil, err
	}
	taskRate := 0.0
	if allTasks > 0 {
		taskRate = float64(completedTasks) / float64(allTasks) * 100
	}
	health := 100.0
	if sales["refund_rate"] > 5 {
		health -= (sales["refund_rate"] - 5) * 2
	}
	if ad["roi"] > 0 && ad["roi"] < 2 {
		health -= (2 - ad["roi"]) * 10
	}
	health -= float64(anomalyCount) * 0.8
	if health < 0 {
		health = 0
	}
	return map[string]interface{}{
		"health_score": int(health + 0.5),
		"kpis": []map[string]interface{}{
			{"code": "gmv", "label": "GMV", "value": sales["gmv"], "unit": "CNY", "trend": "neutral"},
			{"code": "net_sales", "label": "净销售额", "value": sales["net_sales"], "unit": "CNY", "trend": "neutral"},
			{"code": "orders", "label": "订单数", "value": sales["orders"], "unit": "COUNT", "trend": "neutral"},
			{"code": "avg_order_value", "label": "客单价", "value": sales["avg_order_value"], "unit": "CNY", "trend": "neutral"},
			{"code": "roi", "label": "ROI", "value": ad["roi"], "unit": "RATIO", "trend": "neutral"},
			{"code": "refund_rate", "label": "退款率", "value": sales["refund_rate"], "unit": "PERCENT", "trend": "neutral"},
			{"code": "anomalies", "label": "异常数", "value": anomalyCount, "unit": "COUNT", "trend": "neutral"},
			{"code": "task_completion_rate", "label": "整改完成率", "value": taskRate, "unit": "PERCENT", "trend": "neutral"},
		},
		"updated_at": time.Now(),
	}, nil
}

func (s *Service) DashboardTrends(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) ([]map[string]interface{}, error) {
	return s.repo.Trend(ctx, viewer.TenantID, req)
}

func (s *Service) DashboardRankings(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (map[string]interface{}, error) {
	channel, err := s.repo.Ranking(ctx, viewer.TenantID, "channel_code", req)
	if err != nil {
		return nil, err
	}
	brand, err := s.repo.Ranking(ctx, viewer.TenantID, "brand_code", req)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"channel_rank": channel, "brand_rank": brand}, nil
}

func (s *Service) OverviewPipeline(ctx context.Context, viewer dto.Viewer) (map[string]interface{}, error) {
	raw, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterRawDataBatch{}, "")
	if err != nil {
		return nil, err
	}
	metrics, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterMetricDefinition{}, "enabled")
	if err != nil {
		return nil, err
	}
	anomalies, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterAnomalyRecord{}, "")
	if err != nil {
		return nil, err
	}
	tasks, err := s.repo.CountModel(ctx, viewer.TenantID, &models.DataCenterRectificationTask{}, "")
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"code": "raw", "name": "原始数据接入", "value": raw, "status": "normal"},
			{"code": "standard", "name": "标准化清洗", "value": raw, "status": "normal"},
			{"code": "metrics", "name": "指标计算", "value": metrics, "status": "normal"},
			{"code": "anomaly", "name": "异常识别", "value": anomalies, "status": statusByCount(anomalies)},
			{"code": "ai", "name": "AI 分析", "value": anomalies, "status": "normal"},
			{"code": "tasks", "name": "整改任务", "value": tasks, "status": statusByCount(tasks)},
		},
	}, nil
}

func statusByCount(v int64) string {
	if v > 0 {
		return "warning"
	}
	return "normal"
}

func (s *Service) RawBatches(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterRawDataBatch], error) {
	rows, total, err := s.repo.ListRawBatches(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) RawBatch(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterRawDataBatch, error) {
	row, err := s.repo.RawBatch(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) RawErrors(ctx context.Context, viewer dto.Viewer, batchID uint64, req dto.PageRequest) (dto.PageResponse[models.DataCenterRawDataError], error) {
	rows, total, err := s.repo.RawErrors(ctx, viewer.TenantID, batchID, req)
	return page(rows, total, req), err
}

func (s *Service) Metrics(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterMetricDefinition], error) {
	if err := s.EnsureTenantBaseline(ctx, viewer); err != nil {
		return dto.PageResponse[models.DataCenterMetricDefinition]{}, err
	}
	rows, total, err := s.repo.ListMetrics(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Rules(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterAnomalyRule], error) {
	if err := s.EnsureTenantBaseline(ctx, viewer); err != nil {
		return dto.PageResponse[models.DataCenterAnomalyRule]{}, err
	}
	rows, total, err := s.repo.ListRules(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Anomalies(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterAnomalyRecord], error) {
	rows, total, err := s.repo.ListAnomalies(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Tasks(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterRectificationTask], error) {
	rows, total, err := s.repo.ListTasks(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Reviews(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterRectificationReview], error) {
	rows, total, err := s.repo.ListReviews(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Metric(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterMetricDefinition, error) {
	return tenantRow[models.DataCenterMetricDefinition](ctx, s.repo.DB(), viewer.TenantID, id)
}

func (s *Service) Rule(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterAnomalyRule, error) {
	return tenantRow[models.DataCenterAnomalyRule](ctx, s.repo.DB(), viewer.TenantID, id)
}

func (s *Service) Anomaly(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterAnomalyRecord, error) {
	return tenantRow[models.DataCenterAnomalyRecord](ctx, s.repo.DB(), viewer.TenantID, id)
}

func (s *Service) Task(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterRectificationTask, error) {
	return tenantRow[models.DataCenterRectificationTask](ctx, s.repo.DB(), viewer.TenantID, id)
}

func (s *Service) Review(ctx context.Context, viewer dto.Viewer, id uint64) (models.DataCenterRectificationReview, error) {
	return tenantRow[models.DataCenterRectificationReview](ctx, s.repo.DB(), viewer.TenantID, id)
}

func page[T any](rows []T, total int64, req dto.PageRequest) dto.PageResponse[T] {
	if rows == nil {
		rows = []T{}
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if req.Skip < 0 {
		req.Skip = 0
	}
	return dto.PageResponse[T]{Items: rows, Total: total, Skip: req.Skip, Limit: limit}
}

func tenantRow[T any](ctx context.Context, db *gorm.DB, tenantID uint64, id uint64) (T, error) {
	var row T
	err := db.WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) CreateMetric(ctx context.Context, viewer dto.Viewer, payload dto.MetricDefinitionPayload) (models.DataCenterMetricDefinition, error) {
	now := time.Now()
	row := models.DataCenterMetricDefinition{
		TenantID:        viewer.TenantID,
		MetricCode:      strings.TrimSpace(payload.MetricCode),
		MetricName:      strings.TrimSpace(payload.MetricName),
		MetricCategory:  strings.TrimSpace(payload.MetricCategory),
		Formula:         optional(payload.Formula),
		StatisticPeriod: optional(payload.StatisticPeriod),
		Dimensions:      jsonString(payload.Dimensions, "[]"),
		DataSource:      optional(payload.DataSource),
		Enabled:         boolDefault(payload.Enabled, true),
		AnomalyEnabled:  boolDefault(payload.AnomalyEnabled, false),
		CreatedBy:       &viewer.UserID,
		UpdatedBy:       &viewer.UserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if row.MetricCode == "" || row.MetricName == "" || row.MetricCategory == "" {
		return row, errors.New("指标 code、名称和分类不能为空")
	}
	return row, s.repo.DB().WithContext(ctx).Create(&row).Error
}

func (s *Service) UpdateMetric(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.MetricDefinitionPayload) (models.DataCenterMetricDefinition, error) {
	updates := map[string]interface{}{"updated_by": viewer.UserID, "updated_at": time.Now()}
	if value := strings.TrimSpace(payload.MetricCode); value != "" {
		updates["metric_code"] = value
	}
	if value := strings.TrimSpace(payload.MetricName); value != "" {
		updates["metric_name"] = value
	}
	if value := strings.TrimSpace(payload.MetricCategory); value != "" {
		updates["metric_category"] = value
	}
	if payload.Formula != "" {
		updates["formula"] = payload.Formula
	}
	if payload.StatisticPeriod != "" {
		updates["statistic_period"] = payload.StatisticPeriod
	}
	if payload.Dimensions != nil {
		updates["dimensions"] = jsonString(payload.Dimensions, "[]")
	}
	if payload.DataSource != "" {
		updates["data_source"] = payload.DataSource
	}
	if payload.Enabled != nil {
		updates["enabled"] = *payload.Enabled
	}
	if payload.AnomalyEnabled != nil {
		updates["anomaly_enabled"] = *payload.AnomalyEnabled
	}
	result := s.repo.DB().WithContext(ctx).Model(&models.DataCenterMetricDefinition{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).
		Updates(updates)
	if result.Error != nil {
		return models.DataCenterMetricDefinition{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.DataCenterMetricDefinition{}, ErrNotFound
	}
	return s.Metric(ctx, viewer, id)
}

func (s *Service) CreateRule(ctx context.Context, viewer dto.Viewer, payload dto.AnomalyRulePayload) (models.DataCenterAnomalyRule, error) {
	now := time.Now()
	row := models.DataCenterAnomalyRule{
		TenantID:                    viewer.TenantID,
		RuleCode:                    strings.TrimSpace(payload.RuleCode),
		RuleName:                    strings.TrimSpace(payload.RuleName),
		BusinessDomain:              strings.TrimSpace(payload.BusinessDomain),
		TargetObjectType:            strings.TrimSpace(payload.TargetObjectType),
		ScopeJSON:                   jsonString(payload.Scope, "{}"),
		MetricConditionsJSON:        jsonString(payload.MetricConditions, "{}"),
		LevelConfigJSON:             jsonString(payload.LevelConfig, "{}"),
		ConfidenceConfigJSON:        jsonString(payload.ConfidenceConfig, "{}"),
		AIEnabled:                   boolDefault(payload.AIEnabled, true),
		TaskEnabled:                 boolDefault(payload.TaskEnabled, true),
		AutoTaskConfidenceThreshold: intDefault(payload.AutoTaskConfidenceThreshold, 80),
		DefaultOwnerRole:            optional(payload.DefaultOwnerRole),
		DefaultDeadlineDays:         intDefault(payload.DefaultDeadlineDays, 3),
		ReviewMetricCodes:           jsonString(payload.ReviewMetricCodes, "[]"),
		ReviewAfterDays:             intDefault(payload.ReviewAfterDays, 3),
		Priority:                    intDefault(payload.Priority, 100),
		Enabled:                     boolDefault(payload.Enabled, true),
		CreatedBy:                   &viewer.UserID,
		UpdatedBy:                   &viewer.UserID,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}
	if row.RuleCode == "" || row.RuleName == "" || row.BusinessDomain == "" || row.TargetObjectType == "" {
		return row, errors.New("规则 code、名称、业务域和适用对象不能为空")
	}
	return row, s.repo.DB().WithContext(ctx).Create(&row).Error
}

func (s *Service) UpdateRule(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.AnomalyRulePayload) (models.DataCenterAnomalyRule, error) {
	updates := map[string]interface{}{"updated_by": viewer.UserID, "updated_at": time.Now()}
	if value := strings.TrimSpace(payload.RuleCode); value != "" {
		updates["rule_code"] = value
	}
	if value := strings.TrimSpace(payload.RuleName); value != "" {
		updates["rule_name"] = value
	}
	if value := strings.TrimSpace(payload.BusinessDomain); value != "" {
		updates["business_domain"] = value
	}
	if value := strings.TrimSpace(payload.TargetObjectType); value != "" {
		updates["target_object_type"] = value
	}
	if payload.Scope != nil {
		updates["scope_json"] = jsonString(payload.Scope, "{}")
	}
	if payload.MetricConditions != nil {
		updates["metric_conditions_json"] = jsonString(payload.MetricConditions, "{}")
	}
	if payload.LevelConfig != nil {
		updates["level_config_json"] = jsonString(payload.LevelConfig, "{}")
	}
	if payload.ConfidenceConfig != nil {
		updates["confidence_config_json"] = jsonString(payload.ConfidenceConfig, "{}")
	}
	if payload.AIEnabled != nil {
		updates["ai_enabled"] = *payload.AIEnabled
	}
	if payload.TaskEnabled != nil {
		updates["task_enabled"] = *payload.TaskEnabled
	}
	if payload.AutoTaskConfidenceThreshold != nil {
		updates["auto_task_confidence_threshold"] = *payload.AutoTaskConfidenceThreshold
	}
	if payload.DefaultOwnerRole != "" {
		updates["default_owner_role"] = payload.DefaultOwnerRole
	}
	if payload.DefaultDeadlineDays != nil {
		updates["default_deadline_days"] = *payload.DefaultDeadlineDays
	}
	if payload.ReviewMetricCodes != nil {
		updates["review_metric_codes"] = jsonString(payload.ReviewMetricCodes, "[]")
	}
	if payload.ReviewAfterDays != nil {
		updates["review_after_days"] = *payload.ReviewAfterDays
	}
	if payload.Priority != nil {
		updates["priority"] = *payload.Priority
	}
	if payload.Enabled != nil {
		updates["enabled"] = *payload.Enabled
	}
	result := s.repo.DB().WithContext(ctx).Model(&models.DataCenterAnomalyRule{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).
		Updates(updates)
	if result.Error != nil {
		return models.DataCenterAnomalyRule{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.DataCenterAnomalyRule{}, ErrNotFound
	}
	return s.Rule(ctx, viewer, id)
}

func (s *Service) TestRule(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.RuleTestPayload) (map[string]interface{}, error) {
	rule, err := s.Rule(ctx, viewer, id)
	if err != nil {
		return nil, err
	}
	conditions := rule.MetricConditionsJSON
	if payload.MetricConditions != nil {
		conditions = jsonString(payload.MetricConditions, "{}")
	}
	matched, err := evaluateRuleConditions(conditions, payload.Metrics)
	if err != nil {
		return nil, err
	}
	evidence := make([]map[string]interface{}, 0, len(payload.Metrics))
	for code, value := range payload.Metrics {
		evidence = append(evidence, map[string]interface{}{"metric_code": code, "label": code, "value": value, "desc": "规则测试输入指标"})
	}
	return map[string]interface{}{
		"rule_code": rule.RuleCode,
		"matched":   matched,
		"evidence":  evidence,
	}, nil
}

func (s *Service) CreateTask(ctx context.Context, viewer dto.Viewer, payload dto.TaskPayload) (models.DataCenterRectificationTask, error) {
	now := time.Now()
	task := models.DataCenterRectificationTask{
		TenantID:          viewer.TenantID,
		TaskCode:          fmt.Sprintf("TASK-%s-%06d", now.Format("20060102"), now.Nanosecond()%1000000),
		AnomalyID:         payload.AnomalyID,
		Title:             strings.TrimSpace(payload.Title),
		TaskType:          defaultString(payload.TaskType, "manual_rectification"),
		OwnerUserID:       payload.OwnerUserID,
		OwnerRole:         optional(payload.OwnerRole),
		CollaboratorIDs:   jsonString(payload.CollaboratorIDs, "[]"),
		Priority:          defaultString(payload.Priority, "medium"),
		Deadline:          parseDatePtr(payload.Deadline),
		TargetDesc:        optional(payload.TargetDesc),
		AISuggestionJSON:  jsonString(payload.AISuggestion, "{}"),
		ExecutionFeedback: optional(payload.ExecutionFeedback),
		Status:            domain.TaskStatusPending,
		ReviewStatus:      domain.AnomalyReviewStatusNone,
		CreatedBy:         &viewer.UserID,
		UpdatedBy:         &viewer.UserID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if task.Title == "" {
		return task, errors.New("任务标题不能为空")
	}
	if payload.Progress != nil {
		task.Progress = *payload.Progress
	}
	if payload.AnomalyID != nil {
		var anomaly models.DataCenterAnomalyRecord
		if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, *payload.AnomalyID).First(&anomaly).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return task, ErrNotFound
			}
			return task, err
		}
		task.AnomalyCode = &anomaly.AnomalyCode
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		return tx.Create(&models.DataCenterTaskLog{TenantID: viewer.TenantID, TaskID: task.ID, TaskCode: task.TaskCode, Action: "create", ToStatus: &task.Status, OperatorID: viewer.UserID, CreatedAt: now}).Error
	})
	return task, err
}

func (s *Service) UpdateTask(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.TaskPayload) (models.DataCenterRectificationTask, error) {
	updates := map[string]interface{}{"updated_by": viewer.UserID, "updated_at": time.Now()}
	if value := strings.TrimSpace(payload.Title); value != "" {
		updates["title"] = value
	}
	if value := strings.TrimSpace(payload.TaskType); value != "" {
		updates["task_type"] = value
	}
	if payload.OwnerUserID != nil {
		updates["owner_user_id"] = *payload.OwnerUserID
	}
	if value := strings.TrimSpace(payload.OwnerRole); value != "" {
		updates["owner_role"] = value
	}
	if payload.CollaboratorIDs != nil {
		updates["collaborator_ids"] = jsonString(payload.CollaboratorIDs, "[]")
	}
	if value := strings.TrimSpace(payload.Priority); value != "" {
		updates["priority"] = value
	}
	if deadline := parseDatePtr(payload.Deadline); deadline != nil {
		updates["deadline"] = deadline
	}
	if value := strings.TrimSpace(payload.TargetDesc); value != "" {
		updates["target_desc"] = value
	}
	if payload.AISuggestion != nil {
		updates["ai_suggestion_json"] = jsonString(payload.AISuggestion, "{}")
	}
	if value := strings.TrimSpace(payload.ExecutionFeedback); value != "" {
		updates["execution_feedback"] = value
	}
	if payload.Progress != nil {
		updates["progress"] = *payload.Progress
	}
	result := s.repo.DB().WithContext(ctx).Model(&models.DataCenterRectificationTask{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).
		Updates(updates)
	if result.Error != nil {
		return models.DataCenterRectificationTask{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.DataCenterRectificationTask{}, ErrNotFound
	}
	return s.Task(ctx, viewer, id)
}

func (s *Service) GenerateTask(ctx context.Context, viewer dto.Viewer, anomalyID uint64) (models.DataCenterRectificationTask, error) {
	var anomaly models.DataCenterAnomalyRecord
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, anomalyID).First(&anomaly).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DataCenterRectificationTask{}, ErrNotFound
		}
		return models.DataCenterRectificationTask{}, err
	}
	var existing int64
	if err := s.repo.DB().WithContext(ctx).Model(&models.DataCenterRectificationTask{}).Where("tenant_id = ? AND anomaly_id = ? AND deleted_at IS NULL", viewer.TenantID, anomaly.ID).Count(&existing).Error; err != nil {
		return models.DataCenterRectificationTask{}, err
	}
	if existing > 0 {
		return models.DataCenterRectificationTask{}, ErrTaskExists
	}
	var rule models.DataCenterAnomalyRule
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND rule_code = ? AND deleted_at IS NULL", viewer.TenantID, anomaly.RuleCode).First(&rule).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.DataCenterRectificationTask{}, err
	}
	diagnosis, err := s.AnalyzeAnomaly(ctx, viewer, anomaly.ID, false)
	if err != nil {
		return models.DataCenterRectificationTask{}, err
	}
	ownerRole := ""
	if rule.DefaultOwnerRole != nil {
		ownerRole = *rule.DefaultOwnerRole
	}
	deadlineDays := rule.DefaultDeadlineDays
	if deadlineDays <= 0 {
		deadlineDays = 3
	}
	suggestion := map[string]interface{}{
		"diagnosis_id":       diagnosis.ID,
		"anomaly_code":       anomaly.AnomalyCode,
		"rule_code":          anomaly.RuleCode,
		"evidence":           evidenceToMetrics(anomaly.EvidenceJSON),
		"ai_task_suggestion": json.RawMessage(diagnosis.TaskSuggestionJSON),
	}
	now := time.Now()
	task := models.DataCenterRectificationTask{
		TenantID:         viewer.TenantID,
		TaskCode:         fmt.Sprintf("TASK-%s-%06d", now.Format("20060102"), now.Nanosecond()%1000000),
		AnomalyID:        &anomaly.ID,
		AnomalyCode:      &anomaly.AnomalyCode,
		Title:            "整改：" + anomaly.Title,
		TaskType:         "anomaly_rectification",
		OwnerRole:        optional(ownerRole),
		Priority:         priorityByLevel(anomaly.AnomalyLevel),
		Deadline:         ptrTime(now.AddDate(0, 0, deadlineDays)),
		TargetDesc:       optional("围绕异常证据制定整改动作，并在复盘时回填前后指标。"),
		AISuggestionJSON: jsonString(suggestion, "{}"),
		CollaboratorIDs:  "[]",
		Status:           domain.TaskStatusPending,
		ReviewStatus:     domain.AnomalyReviewStatusNone,
		CreatedBy:        &viewer.UserID,
		UpdatedBy:        &viewer.UserID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		if err := tx.Model(&anomaly).Updates(map[string]interface{}{"task_status": domain.AnomalyTaskStatusGenerated, "status": domain.AnomalyStatusProcessing, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Create(&models.DataCenterTaskLog{TenantID: viewer.TenantID, TaskID: task.ID, TaskCode: task.TaskCode, Action: "create", ToStatus: &task.Status, OperatorID: viewer.UserID, CreatedAt: now}).Error
	})
	return task, err
}

func priorityByLevel(level string) string {
	switch level {
	case "critical", "严重":
		return "critical"
	case "high", "高":
		return "high"
	case "low", "低":
		return "low"
	default:
		return "medium"
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func parseDatePtr(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func optional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func boolDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func intDefault(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func jsonString(value interface{}, fallback string) string {
	if value == nil {
		return fallback
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	return string(raw)
}

func currentDataCenterQuotaLimit(db *gorm.DB, tenantID uint64, quotaID uint64) int {
	now := time.Now()
	var override models.TenantQuotaOverride
	if err := db.Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quotaID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue
	}
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err == nil {
		var planQuota models.SaasPlanQuota
		if err := db.Where("plan_id = ? AND quota_id = ?", sub.PlanID, quotaID).First(&planQuota).Error; err == nil {
			return planQuota.QuotaValue
		}
	}
	return 0
}
