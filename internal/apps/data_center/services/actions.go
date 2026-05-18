package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/data_center/domain"
	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (s *Service) StandardData(ctx context.Context, viewer dto.Viewer, dataType string, req dto.PageRequest) (interface{}, error) {
	switch dataType {
	case "sales":
		var rows []models.DataCenterStdSalesOrder
		var total int64
		db := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", viewer.TenantID).Model(&models.DataCenterStdSalesOrder{})
		db = applyStandardFilters(db, req, "order_time", map[string]string{"status": "order_status", "store_code": "-"})
		if err := db.Count(&total).Error; err != nil {
			return nil, err
		}
		if err := db.Order("order_time desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error; err != nil {
			return nil, err
		}
		return page(rows, total, req), nil
	case "ad":
		var rows []models.DataCenterStdAdDaily
		var total int64
		db := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", viewer.TenantID).Model(&models.DataCenterStdAdDaily{})
		db = applyStandardFilters(db, req, "stat_date", map[string]string{"channel_code": "-", "store_code": "-"})
		if err := db.Count(&total).Error; err != nil {
			return nil, err
		}
		if err := db.Order("stat_date desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error; err != nil {
			return nil, err
		}
		return page(rows, total, req), nil
	case "inventory":
		var rows []models.DataCenterStdInventoryDaily
		var total int64
		db := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", viewer.TenantID).Model(&models.DataCenterStdInventoryDaily{})
		db = applyStandardFilters(db, req, "stat_date", map[string]string{"status": "inventory_status", "channel_code": "-", "platform_code": "-"})
		if err := db.Count(&total).Error; err != nil {
			return nil, err
		}
		if err := db.Order("stat_date desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error; err != nil {
			return nil, err
		}
		return page(rows, total, req), nil
	case "refund":
		var rows []models.DataCenterStdRefundOrder
		var total int64
		db := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", viewer.TenantID).Model(&models.DataCenterStdRefundOrder{})
		db = applyStandardFilters(db, req, "refund_time", map[string]string{"status": "refund_status", "store_code": "-"})
		if err := db.Count(&total).Error; err != nil {
			return nil, err
		}
		if err := db.Order("refund_time desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error; err != nil {
			return nil, err
		}
		return page(rows, total, req), nil
	case "product":
		var rows []models.DataCenterStdProduct
		var total int64
		db := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", viewer.TenantID).Model(&models.DataCenterStdProduct{})
		db = applyStandardFilters(db, req, "updated_at", map[string]string{"status": "product_status", "channel_code": "-", "platform_code": "-", "store_code": "-"})
		if err := db.Count(&total).Error; err != nil {
			return nil, err
		}
		if err := db.Order("updated_at desc, id desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error; err != nil {
			return nil, err
		}
		return page(rows, total, req), nil
	case "store-sales":
		var rows []models.DataCenterStdStoreSalesDaily
		var total int64
		db := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", viewer.TenantID).Model(&models.DataCenterStdStoreSalesDaily{})
		db = applyStandardFilters(db, req, "stat_date", map[string]string{"product_code": "-"})
		if err := db.Count(&total).Error; err != nil {
			return nil, err
		}
		if err := db.Order("stat_date desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error; err != nil {
			return nil, err
		}
		return page(rows, total, req), nil
	default:
		return nil, errors.New("不支持的数据类型")
	}
}

func (s *Service) StandardDataDetail(ctx context.Context, viewer dto.Viewer, dataType string, id uint64) (interface{}, error) {
	switch dataType {
	case "sales":
		return tenantRow[models.DataCenterStdSalesOrder](ctx, s.repo.DB(), viewer.TenantID, id)
	case "ad":
		return tenantRow[models.DataCenterStdAdDaily](ctx, s.repo.DB(), viewer.TenantID, id)
	case "inventory":
		return tenantRow[models.DataCenterStdInventoryDaily](ctx, s.repo.DB(), viewer.TenantID, id)
	case "refund":
		return tenantRow[models.DataCenterStdRefundOrder](ctx, s.repo.DB(), viewer.TenantID, id)
	case "product":
		return tenantRow[models.DataCenterStdProduct](ctx, s.repo.DB(), viewer.TenantID, id)
	case "store-sales":
		return tenantRow[models.DataCenterStdStoreSalesDaily](ctx, s.repo.DB(), viewer.TenantID, id)
	default:
		return nil, errors.New("不支持的数据类型")
	}
}

func applyStandardFilters(db *gorm.DB, req dto.PageRequest, dateColumn string, aliases map[string]string) *gorm.DB {
	for column, value := range map[string]string{
		"brand_code":    req.BrandCode,
		"channel_code":  req.ChannelCode,
		"platform_code": req.PlatformCode,
		"store_code":    req.StoreCode,
		"product_code":  req.ProductCode,
	} {
		if strings.TrimSpace(value) != "" {
			if aliases != nil && aliases[column] == "-" {
				continue
			}
			db = db.Where(column+" = ?", value)
		}
	}
	if req.Status != "" {
		column := "status"
		if aliases != nil && aliases["status"] != "" {
			column = aliases["status"]
		}
		db = db.Where(column+" = ?", req.Status)
	}
	if dateColumn != "" && (req.StartDate != nil || req.EndDate != nil || req.TimeRange != "") {
		start, end := standardTimeWindow(req)
		db = db.Where(dateColumn+" BETWEEN ? AND ?", start, end)
	}
	return db
}

func standardTimeWindow(req dto.PageRequest) (time.Time, time.Time) {
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

func normalLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 20
	}
	return limit
}

func (s *Service) MetricResults(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.DataCenterMetricResult], error) {
	var rows []models.DataCenterMetricResult
	var total int64
	db := s.repo.DB().WithContext(ctx).Where("tenant_id = ?", viewer.TenantID).Model(&models.DataCenterMetricResult{})
	if req.Keyword != "" {
		db = db.Where("metric_code = ?", req.Keyword)
	}
	if err := db.Count(&total).Error; err != nil {
		return page(rows, total, req), err
	}
	err := db.Order("stat_date desc, id desc").Offset(req.Skip).Limit(normalLimit(req.Limit)).Find(&rows).Error
	return page(rows, total, req), err
}

func (s *Service) SetMetricEnabled(ctx context.Context, viewer dto.Viewer, id uint64, enabled bool) error {
	result := s.repo.DB().WithContext(ctx).Model(&models.DataCenterMetricDefinition{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).
		Updates(map[string]interface{}{"enabled": enabled, "updated_by": viewer.UserID, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) SetRuleEnabled(ctx context.Context, viewer dto.Viewer, id uint64, enabled bool) error {
	if enabled {
		if err := s.ensureActiveRuleQuota(ctx, viewer.TenantID); err != nil {
			return err
		}
	}
	result := s.repo.DB().WithContext(ctx).Model(&models.DataCenterAnomalyRule{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).
		Updates(map[string]interface{}{"enabled": enabled, "updated_by": viewer.UserID, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdateAnomalyStatus(ctx context.Context, viewer dto.Viewer, id uint64, status string) error {
	result := s.repo.DB().WithContext(ctx).Model(&models.DataCenterAnomalyRecord{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) AnalyzeAnomaly(ctx context.Context, viewer dto.Viewer, id uint64, reanalyze bool) (models.DataCenterAIDiagnosisRecord, error) {
	var anomaly models.DataCenterAnomalyRecord
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).First(&anomaly).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DataCenterAIDiagnosisRecord{}, ErrNotFound
		}
		return models.DataCenterAIDiagnosisRecord{}, err
	}
	if !reanalyze {
		var existing models.DataCenterAIDiagnosisRecord
		if err := s.repo.DB().WithContext(ctx).
			Where("tenant_id = ? AND anomaly_id = ? AND status = ?", viewer.TenantID, anomaly.ID, domain.AnomalyAIStatusSuccess).
			Order("generated_at desc, id desc").
			First(&existing).Error; err == nil {
			return existing, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DataCenterAIDiagnosisRecord{}, err
		}
	}
	if err := s.consumeQuota(ctx, viewer.TenantID, "data_center_daily_ai_analyses"); err != nil {
		return models.DataCenterAIDiagnosisRecord{}, err
	}
	var rule models.DataCenterAnomalyRule
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND rule_code = ? AND deleted_at IS NULL", viewer.TenantID, anomaly.RuleCode).First(&rule).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.DataCenterAIDiagnosisRecord{}, err
	}
	now := time.Now()
	modelCode := "data-center-rule-assistant"
	diagnosis, analyzeErr := s.aiAnalyzer.AnalyzeAnomaly(ctx, anomaly, rule)
	if analyzeErr != nil {
		errorMessage := safeAIError(analyzeErr)
		record := models.DataCenterAIDiagnosisRecord{
			TenantID:            viewer.TenantID,
			AnomalyID:           anomaly.ID,
			AnomalyCode:         anomaly.AnomalyCode,
			EvidenceSummaryJSON: anomaly.EvidenceJSON,
			TaskSuggestionJSON:  "{}",
			ModelCode:           &modelCode,
			Status:              domain.AnomalyAIStatusFailed,
			ErrorMessage:        &errorMessage,
			GeneratedAt:         now,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			return tx.Model(&anomaly).Updates(map[string]interface{}{"ai_status": domain.AnomalyAIStatusFailed, "updated_at": now}).Error
		})
		if err != nil {
			return record, err
		}
		return record, analyzeErr
	}
	record := models.DataCenterAIDiagnosisRecord{
		TenantID:              viewer.TenantID,
		AnomalyID:             anomaly.ID,
		AnomalyCode:           anomaly.AnomalyCode,
		ProblemSummary:        optional(diagnosis.ProblemSummary),
		ImpactSummary:         optional(diagnosis.ImpactSummary),
		ReasonAnalysisJSON:    jsonString(diagnosis.Reasons, "[]"),
		EvidenceSummaryJSON:   anomaly.EvidenceJSON,
		SuggestionJSON:        jsonString(diagnosis.Suggestions, "[]"),
		ConfidenceExplanation: optional(diagnosis.ConfidenceExplanation),
		TaskSuggestionJSON:    jsonString(diagnosis.TaskSuggestion, "{}"),
		ModelCode:             &modelCode,
		Status:                domain.AnomalyAIStatusSuccess,
		GeneratedAt:           now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return tx.Model(&anomaly).Updates(map[string]interface{}{"ai_status": domain.AnomalyAIStatusSuccess, "updated_at": now}).Error
	})
	return record, err
}

func safeAIError(err error) string {
	msg := strings.TrimSpace(err.Error())
	lower := strings.ToLower(msg)
	for _, token := range []string{"secret", "token", "authorization", "password", "select ", "insert ", "update ", "delete ", "panic", "stack"} {
		if strings.Contains(lower, token) {
			return "AI 分析失败，请稍后重试或联系管理员查看审计日志"
		}
	}
	if msg == "" {
		return "AI 分析失败"
	}
	if len([]rune(msg)) > 200 {
		runes := []rune(msg)
		return string(runes[:200])
	}
	return msg
}

func (s *Service) TransitionTask(ctx context.Context, viewer dto.Viewer, id uint64, toStatus string, feedback string, progress *int) (models.DataCenterRectificationTask, error) {
	var task models.DataCenterRectificationTask
	err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return task, ErrNotFound
		}
		return task, err
	}
	if toStatus != "" && !domain.CanTransitTaskStatus(task.Status, toStatus) {
		return task, ErrInvalidStatus
	}
	now := time.Now()
	from := task.Status
	updates := map[string]interface{}{"updated_by": viewer.UserID, "updated_at": now}
	if toStatus != "" {
		updates["status"] = toStatus
		if toStatus == domain.TaskStatusCompleted {
			updates["completed_at"] = now
			updates["review_status"] = domain.AnomalyReviewStatusPending
			updates["progress"] = 100
		}
	}
	if feedback != "" {
		updates["execution_feedback"] = feedback
	}
	if progress != nil {
		updates["progress"] = *progress
	}
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&task).Updates(updates).Error; err != nil {
			return err
		}
		log := models.DataCenterTaskLog{TenantID: viewer.TenantID, TaskID: task.ID, TaskCode: task.TaskCode, Action: "status_update", FromStatus: &from, OperatorID: viewer.UserID, CreatedAt: now}
		if toStatus != "" {
			log.ToStatus = &toStatus
		}
		if feedback != "" {
			log.Content = &feedback
		}
		return tx.Create(&log).Error
	})
	if err != nil {
		return task, err
	}
	_ = s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ?", viewer.TenantID, id).First(&task).Error
	return task, nil
}

func (s *Service) CreateReview(ctx context.Context, viewer dto.Viewer, payload dto.ReviewPayload) (models.DataCenterRectificationReview, error) {
	var task models.DataCenterRectificationTask
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, payload.TaskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DataCenterRectificationReview{}, ErrNotFound
		}
		return models.DataCenterRectificationReview{}, err
	}
	beforeMetrics := payload.BeforeMetrics
	afterMetrics := payload.AfterMetrics
	if len(beforeMetrics) == 0 || len(afterMetrics) == 0 {
		generatedBefore, generatedAfter, err := s.reviewMetrics(ctx, viewer, task)
		if err != nil {
			return models.DataCenterRectificationReview{}, err
		}
		if len(beforeMetrics) == 0 {
			beforeMetrics = generatedBefore
		}
		if len(afterMetrics) == 0 {
			afterMetrics = generatedAfter
		}
	}
	rate := improvementRate(beforeMetrics, afterMetrics)
	improvementText := fmt.Sprintf("%.2f%%", rate)
	conclusion := defaultString(payload.ReviewConclusion, conclusionByImprovement(rate))
	aiSummary := defaultString(payload.AIReviewSummary, fmt.Sprintf("系统基于整改前后指标计算改善幅度为%s，建议人工结合业务动作确认最终复盘结论。", improvementText))
	now := time.Now()
	review := models.DataCenterRectificationReview{
		TenantID:            viewer.TenantID,
		ReviewCode:          fmt.Sprintf("REV-%s-%06d", now.Format("20060102"), now.Nanosecond()%1000000),
		TaskID:              task.ID,
		TaskCode:            task.TaskCode,
		AnomalyID:           task.AnomalyID,
		AnomalyCode:         task.AnomalyCode,
		BeforeMetricJSON:    jsonString(beforeMetrics, "[]"),
		AfterMetricJSON:     jsonString(afterMetrics, "[]"),
		ImprovementResult:   optional(improvementText),
		ReviewConclusion:    conclusion,
		AIReviewSummary:     optional(aiSummary),
		ManualReviewSummary: optional(payload.ManualReviewSummary),
		ExperienceSummary:   optional(payload.ExperienceSummary),
		ReviewedAt:          &now,
		CreatedBy:           &viewer.UserID,
		UpdatedBy:           &viewer.UserID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&review).Error; err != nil {
			return err
		}
		if err := tx.Model(&task).Updates(map[string]interface{}{"review_status": domain.AnomalyReviewStatusReviewed, "updated_at": now}).Error; err != nil {
			return err
		}
		if task.AnomalyID != nil {
			return tx.Model(&models.DataCenterAnomalyRecord{}).Where("tenant_id = ? AND id = ?", viewer.TenantID, *task.AnomalyID).
				Updates(map[string]interface{}{"review_status": domain.AnomalyReviewStatusReviewed, "updated_at": now}).Error
		}
		return nil
	})
	return review, err
}

func (s *Service) ConfirmReview(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.ReviewPayload) (models.DataCenterRectificationReview, error) {
	var review models.DataCenterRectificationReview
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return review, ErrNotFound
		}
		return review, err
	}
	now := time.Now()
	updates := map[string]interface{}{
		"reviewed_at": now,
		"updated_by":  viewer.UserID,
		"updated_at":  now,
	}
	if payload.ReviewConclusion != "" {
		if !validReviewConclusion(payload.ReviewConclusion) {
			return review, ErrInvalidStatus
		}
		updates["review_conclusion"] = payload.ReviewConclusion
	}
	if payload.AIReviewSummary != "" {
		updates["ai_review_summary"] = payload.AIReviewSummary
	}
	if payload.ManualReviewSummary != "" {
		updates["manual_review_summary"] = payload.ManualReviewSummary
	}
	if payload.ExperienceSummary != "" {
		updates["experience_summary"] = payload.ExperienceSummary
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&review).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DataCenterRectificationTask{}).
			Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, review.TaskID).
			Updates(map[string]interface{}{"review_status": domain.AnomalyReviewStatusReviewed, "updated_at": now}).Error; err != nil {
			return err
		}
		if review.AnomalyID != nil {
			return tx.Model(&models.DataCenterAnomalyRecord{}).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, *review.AnomalyID).
				Updates(map[string]interface{}{"review_status": domain.AnomalyReviewStatusReviewed, "updated_at": now}).Error
		}
		return nil
	})
	if err != nil {
		return review, err
	}
	err = s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, id).First(&review).Error
	return review, err
}

func validReviewConclusion(value string) bool {
	switch value {
	case domain.ReviewConclusionEffective, domain.ReviewConclusionWeak, domain.ReviewConclusionIneffective, domain.ReviewConclusionFollowUp:
		return true
	default:
		return false
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func (s *Service) reviewMetrics(ctx context.Context, viewer dto.Viewer, task models.DataCenterRectificationTask) ([]map[string]interface{}, []map[string]interface{}, error) {
	if task.AnomalyID == nil {
		return []map[string]interface{}{}, []map[string]interface{}{}, nil
	}
	var anomaly models.DataCenterAnomalyRecord
	if err := s.repo.DB().WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, *task.AnomalyID).First(&anomaly).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []map[string]interface{}{}, []map[string]interface{}{}, nil
		}
		return nil, nil, err
	}
	before := evidenceToMetrics(anomaly.EvidenceJSON)
	if len(before) == 0 {
		before = []map[string]interface{}{{"metric_code": anomaly.RuleCode, "label": anomaly.Title, "value": anomaly.ImpactAmount, "stat_date": anomaly.StatDate.Format("2006-01-02")}}
	}
	after := make([]map[string]interface{}, 0, len(before))
	for _, item := range before {
		metricCode := strings.TrimSpace(fmt.Sprint(item["metric_code"]))
		if metricCode == "" || metricCode == "<nil>" {
			continue
		}
		var result models.DataCenterMetricResult
		err := s.repo.DB().WithContext(ctx).
			Where("tenant_id = ? AND metric_code = ? AND resource_type = ? AND resource_code = ?", viewer.TenantID, metricCode, anomaly.ObjectType, anomaly.ObjectCode).
			Where("stat_date >= ?", anomaly.StatDate).
			Order("stat_date desc, id desc").
			First(&result).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, nil, err
		}
		after = append(after, map[string]interface{}{"metric_code": result.MetricCode, "label": result.MetricCode, "value": result.MetricValue, "stat_date": result.StatDate.Format("2006-01-02")})
	}
	if len(after) == 0 {
		after = before
	}
	return before, after, nil
}

type localDiagnosis struct {
	ProblemSummary        string
	ImpactSummary         string
	Reasons               []map[string]interface{}
	Suggestions           []map[string]interface{}
	TaskSuggestion        map[string]interface{}
	ConfidenceExplanation string
}

func buildLocalDiagnosis(anomaly models.DataCenterAnomalyRecord, rule models.DataCenterAnomalyRule) localDiagnosis {
	evidence := evidenceToMetrics(anomaly.EvidenceJSON)
	reasons := make([]map[string]interface{}, 0, len(evidence))
	for _, item := range evidence {
		reasons = append(reasons, map[string]interface{}{
			"metric_code": item["metric_code"],
			"label":       item["label"],
			"value":       item["value"],
			"desc":        item["desc"],
		})
	}
	if len(reasons) == 0 {
		reasons = append(reasons, map[string]interface{}{"metric_code": anomaly.RuleCode, "label": anomaly.Title, "value": anomaly.ImpactAmount, "desc": "异常记录缺少明细证据，需人工复核原始指标。"})
	}
	ownerRole := "业务负责人"
	if rule.DefaultOwnerRole != nil && strings.TrimSpace(*rule.DefaultOwnerRole) != "" {
		ownerRole = *rule.DefaultOwnerRole
	}
	deadlineDays := rule.DefaultDeadlineDays
	if deadlineDays <= 0 {
		deadlineDays = 3
	}
	suggestions := []map[string]interface{}{
		{"type": "diagnose", "content": "复核异常证据对应的数据源、统计周期和对象维度，确认是否存在采集延迟或口径变化。"},
		{"type": "action", "content": fmt.Sprintf("由%s在%d天内提交整改动作和结果指标。", ownerRole, deadlineDays)},
	}
	return localDiagnosis{
		ProblemSummary: fmt.Sprintf("%s命中规则%s，置信度%d%%。", anomaly.Title, anomaly.RuleCode, anomaly.ConfidenceScore),
		ImpactSummary:  fmt.Sprintf("影响对象%s/%s，预估影响金额%.2f。", anomaly.ObjectType, anomaly.ObjectCode, anomaly.ImpactAmount),
		Reasons:        reasons,
		Suggestions:    suggestions,
		TaskSuggestion: map[string]interface{}{
			"title":         "处理：" + anomaly.Title,
			"owner_role":    ownerRole,
			"deadline_days": deadlineDays,
			"priority":      priorityByLevel(anomaly.AnomalyLevel),
			"target_desc":   "围绕异常证据制定整改动作，并在复盘时回填前后指标。",
			"evidence":      evidence,
		},
		ConfidenceExplanation: "置信度由规则引擎根据指标偏离、数据完整度和证据数量计算，AI 仅解释证据并生成建议。",
	}
}

func evidenceToMetrics(raw string) []map[string]interface{} {
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return []map[string]interface{}{}
	}
	if rows == nil {
		return []map[string]interface{}{}
	}
	return rows
}

func metricValueSum(rows []map[string]interface{}) float64 {
	total := 0.0
	for _, row := range rows {
		switch value := row["value"].(type) {
		case float64:
			total += value
		case float32:
			total += float64(value)
		case int:
			total += float64(value)
		case int64:
			total += float64(value)
		case json.Number:
			parsed, _ := value.Float64()
			total += parsed
		}
	}
	return total
}

func improvementRate(before []map[string]interface{}, after []map[string]interface{}) float64 {
	beforeSum := metricValueSum(before)
	afterSum := metricValueSum(after)
	if math.Abs(beforeSum) < 0.0001 {
		return 0
	}
	return (afterSum - beforeSum) / math.Abs(beforeSum) * 100
}

func conclusionByImprovement(rate float64) string {
	switch {
	case rate >= 10:
		return domain.ReviewConclusionEffective
	case rate > 0:
		return domain.ReviewConclusionWeak
	case rate <= -10:
		return domain.ReviewConclusionIneffective
	default:
		return domain.ReviewConclusionFollowUp
	}
}
