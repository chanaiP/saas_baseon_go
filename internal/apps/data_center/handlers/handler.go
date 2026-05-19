package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/apps/data_center/services"
	"saas_baseon_go/internal/interfaces/http/response"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) DashboardSummary(c *gin.Context) {
	data, err := h.service.DashboardSummary(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) DashboardTrends(c *gin.Context) {
	data, err := h.service.DashboardTrends(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) DashboardRankings(c *gin.Context) {
	data, err := h.service.DashboardRankings(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) DashboardAnomalies(c *gin.Context) {
	data, err := h.service.Anomalies(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) DashboardTasks(c *gin.Context) {
	data, err := h.service.Tasks(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) OverviewPipeline(c *gin.Context) {
	data, err := h.service.OverviewPipeline(c.Request.Context(), viewer(c))
	h.ok(c, data, err)
}
func (h *Handler) OverviewJobs(c *gin.Context) {
	data, err := h.service.RawBatches(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) OverviewErrors(c *gin.Context) {
	id := uint64(0)
	data, err := h.service.RawErrors(c.Request.Context(), viewer(c), id, pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) RawBatches(c *gin.Context) {
	data, err := h.service.RawBatches(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) CreateRawBatch(c *gin.Context) {
	var payload dto.RawBatchPayload
	if bind(c, &payload) {
		data, err := h.service.CreateRawBatch(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "raw_create", data.BatchCode, "创建原始数据批次", gin.H{"data_type": data.DataType, "record_count": data.RecordCount})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) RawBatch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.RawBatch(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}
func (h *Handler) RawErrors(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.RawErrors(c.Request.Context(), viewer(c), id, pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) ReprocessRawBatch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ReprocessRawBatch(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "raw_reprocess", data.BatchCode, "重新处理原始数据批次", gin.H{"batch_id": data.ID, "data_type": data.DataType})
	}
	h.ok(c, data, err)
}
func (h *Handler) StandardData(c *gin.Context) {
	data, err := h.service.StandardData(c.Request.Context(), viewer(c), c.Param("data_type"), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) StandardDataDetail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.StandardDataDetail(c.Request.Context(), viewer(c), c.Param("data_type"), id)
	h.ok(c, data, err)
}
func (h *Handler) Metrics(c *gin.Context) {
	data, err := h.service.Metrics(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) Metric(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Metric(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}
func (h *Handler) CreateMetric(c *gin.Context) {
	var payload dto.MetricDefinitionPayload
	if bind(c, &payload) {
		data, err := h.service.CreateMetric(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "metric_create", data.MetricCode, "创建指标定义", gin.H{"metric_name": data.MetricName, "metric_category": data.MetricCategory})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) UpdateMetric(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.MetricDefinitionPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateMetric(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "metric_update", data.MetricCode, "更新指标定义", gin.H{"metric_id": data.ID, "metric_name": data.MetricName})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) EnableMetric(c *gin.Context)  { h.toggleMetric(c, true) }
func (h *Handler) DisableMetric(c *gin.Context) { h.toggleMetric(c, false) }
func (h *Handler) toggleMetric(c *gin.Context, enabled bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.service.SetMetricEnabled(c.Request.Context(), viewer(c), id, enabled)
	if err == nil {
		action := "metric_disable"
		summary := "停用指标定义"
		if enabled {
			action = "metric_enable"
			summary = "启用指标定义"
		}
		err = h.auditWrite(c, action, idObjectCode(id), summary, gin.H{"enabled": enabled})
	}
	h.ok(c, gin.H{"id": id, "enabled": enabled}, err)
}
func (h *Handler) MetricResults(c *gin.Context) {
	data, err := h.service.MetricResults(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) Rules(c *gin.Context) {
	data, err := h.service.Rules(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) Rule(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Rule(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}
func (h *Handler) CreateRule(c *gin.Context) {
	var payload dto.AnomalyRulePayload
	if bind(c, &payload) {
		data, err := h.service.CreateRule(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "rule_create", data.RuleCode, "创建异常规则", gin.H{"rule_name": data.RuleName, "business_domain": data.BusinessDomain})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) UpdateRule(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.AnomalyRulePayload
	if bind(c, &payload) {
		data, err := h.service.UpdateRule(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "rule_update", data.RuleCode, "更新异常规则", gin.H{"rule_id": data.ID, "rule_name": data.RuleName})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) EnableRule(c *gin.Context)  { h.toggleRule(c, true) }
func (h *Handler) DisableRule(c *gin.Context) { h.toggleRule(c, false) }
func (h *Handler) toggleRule(c *gin.Context, enabled bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.service.SetRuleEnabled(c.Request.Context(), viewer(c), id, enabled)
	if err == nil {
		action := "rule_disable"
		summary := "停用异常规则"
		if enabled {
			action = "rule_enable"
			summary = "启用异常规则"
		}
		err = h.auditWrite(c, action, idObjectCode(id), summary, gin.H{"enabled": enabled})
	}
	h.ok(c, gin.H{"id": id, "enabled": enabled}, err)
}
func (h *Handler) TestRule(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.RuleTestPayload
	if bind(c, &payload) {
		data, err := h.service.TestRule(c.Request.Context(), viewer(c), id, payload)
		h.ok(c, data, err)
	}
}
func (h *Handler) Anomalies(c *gin.Context) {
	data, err := h.service.Anomalies(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) Anomaly(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Anomaly(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}
func (h *Handler) ScanAnomalies(c *gin.Context) {
	data, err := h.service.ScanAnomalies(c.Request.Context(), viewer(c), pageRequest(c))
	if err == nil {
		err = h.auditWrite(c, "anomaly_scan", "scan", "扫描经营异常", data)
	}
	h.ok(c, data, err)
}
func (h *Handler) AnalyzeAnomaly(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.AnalyzeAnomaly(c.Request.Context(), viewer(c), id, false)
	if err == nil {
		err = h.auditWrite(c, "ai_analyze", data.AnomalyCode, "AI 分析异常", gin.H{"anomaly_id": id, "diagnosis_id": data.ID})
	}
	h.ok(c, data, err)
}
func (h *Handler) ReanalyzeAnomaly(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.AnalyzeAnomaly(c.Request.Context(), viewer(c), id, true)
	if err == nil {
		err = h.auditWrite(c, "ai_reanalyze", data.AnomalyCode, "AI 重新分析异常", gin.H{"anomaly_id": id, "diagnosis_id": data.ID})
	}
	h.ok(c, data, err)
}
func (h *Handler) GenerateTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.GenerateTask(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "task_generate", data.TaskCode, "从异常生成整改任务", gin.H{"anomaly_id": id, "task_id": data.ID})
	}
	h.ok(c, data, err)
}
func (h *Handler) ConfirmAnomaly(c *gin.Context) { h.updateAnomaly(c, "confirmed") }
func (h *Handler) IgnoreAnomaly(c *gin.Context)  { h.updateAnomaly(c, "ignored") }
func (h *Handler) CloseAnomaly(c *gin.Context)   { h.updateAnomaly(c, "closed") }
func (h *Handler) updateAnomaly(c *gin.Context, status string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.service.UpdateAnomalyStatus(c.Request.Context(), viewer(c), id, status)
	if err == nil {
		err = h.auditWrite(c, "anomaly_"+status, idObjectCode(id), "更新异常状态", gin.H{"status": status})
	}
	h.ok(c, gin.H{"id": id, "status": status}, err)
}
func (h *Handler) Tasks(c *gin.Context) {
	data, err := h.service.Tasks(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) Task(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Task(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}
func (h *Handler) CreateTask(c *gin.Context) {
	var payload dto.TaskPayload
	if bind(c, &payload) {
		data, err := h.service.CreateTask(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "task_create", data.TaskCode, "创建整改任务", gin.H{"task_id": data.ID, "title": data.Title})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) UpdateTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.TaskPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateTask(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "task_update", data.TaskCode, "更新整改任务", gin.H{"task_id": data.ID, "status": data.Status, "progress": data.Progress})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) StartTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.TransitionTask(c.Request.Context(), viewer(c), id, "processing", "", nil)
	if err == nil {
		err = h.auditWrite(c, "task_start", data.TaskCode, "启动整改任务", gin.H{"task_id": data.ID, "status": data.Status})
	}
	h.ok(c, data, err)
}
func (h *Handler) FeedbackTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.FeedbackPayload
	if bind(c, &payload) {
		data, err := h.service.TransitionTask(c.Request.Context(), viewer(c), id, "", payload.Content, payload.Progress)
		if err == nil {
			err = h.auditWrite(c, "task_feedback", data.TaskCode, "提交整改反馈", gin.H{"task_id": data.ID, "progress": data.Progress})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) CompleteTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.TransitionTask(c.Request.Context(), viewer(c), id, "completed", "", nil)
	if err == nil {
		err = h.auditWrite(c, "task_complete", data.TaskCode, "完成整改任务", gin.H{"task_id": data.ID, "status": data.Status})
	}
	h.ok(c, data, err)
}
func (h *Handler) CloseTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.TransitionTask(c.Request.Context(), viewer(c), id, "closed", "", nil)
	if err == nil {
		err = h.auditWrite(c, "task_close", data.TaskCode, "关闭整改任务", gin.H{"task_id": data.ID, "status": data.Status})
	}
	h.ok(c, data, err)
}
func (h *Handler) Reviews(c *gin.Context) {
	data, err := h.service.Reviews(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}
func (h *Handler) Review(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Review(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}
func (h *Handler) GenerateReview(c *gin.Context) {
	var payload dto.ReviewPayload
	if bind(c, &payload) {
		data, err := h.service.CreateReview(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "review_generate", data.ReviewCode, "生成整改复盘", gin.H{"review_id": data.ID, "task_id": data.TaskID})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) UpdateReview(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.ReviewPayload
	if bind(c, &payload) {
		data, err := h.service.ConfirmReview(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "review_update", data.ReviewCode, "更新整改复盘", gin.H{"review_id": data.ID, "conclusion": data.ReviewConclusion})
		}
		h.ok(c, data, err)
	}
}
func (h *Handler) ConfirmReview(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.ReviewPayload
	if c.Request.ContentLength > 0 && !bind(c, &payload) {
		return
	}
	data, err := h.service.ConfirmReview(c.Request.Context(), viewer(c), id, payload)
	if err == nil {
		err = h.auditWrite(c, "review_confirm", data.ReviewCode, "确认整改复盘", gin.H{"review_id": data.ID, "conclusion": data.ReviewConclusion})
	}
	h.ok(c, data, err)
}

func (h *Handler) ok(c *gin.Context, values ...interface{}) {
	var data interface{}
	var err error
	if len(values) > 0 {
		data = values[0]
	}
	if len(values) > 1 && values[1] != nil {
		if typed, ok := values[1].(error); ok {
			err = typed
		}
	}
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) auditWrite(c *gin.Context, action string, objectCode string, summary string, detail interface{}) error {
	return h.service.RecordAudit(c.Request.Context(), viewer(c), requestMeta(c), action, objectCode, summary, detail)
}

func requestMeta(c *gin.Context) dto.RequestMeta {
	return dto.RequestMeta{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		RequestID: c.GetString("request_id"),
	}
}

func idObjectCode(id uint64) string {
	return fmt.Sprintf("%d", id)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "数据不存在")
	case errors.Is(err, services.ErrTaskExists):
		response.Error(c, http.StatusConflict, response.CodeConflict, "该异常已生成整改任务")
	case errors.Is(err, services.ErrInvalidStatus):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "状态流转不合法")
	default:
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
	}
}

func bind(c *gin.Context, payload interface{}) bool {
	if err := c.ShouldBindJSON(payload); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return false
	}
	return true
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return 0, false
	}
	return id, true
}

func viewer(c *gin.Context) dto.Viewer {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")
	isPlatformAdmin, _ := c.Get("is_platform_admin")
	viewer := dto.Viewer{UserID: toUint64(userID), TenantID: toUint64(tenantID), IsPlatformAdmin: toBool(isPlatformAdmin)}
	if viewer.IsPlatformAdmin && c.Query("tenant_id") != "" {
		if explicitTenantID, err := strconv.ParseUint(c.Query("tenant_id"), 10, 64); err == nil && explicitTenantID > 0 {
			viewer.TenantID = explicitTenantID
		}
	}
	return viewer
}

func toUint64(value interface{}) uint64 {
	switch v := value.(type) {
	case uint64:
		return v
	case uint:
		return uint64(v)
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	default:
		return 0
	}
}

func toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	default:
		return false
	}
}

func pageRequest(c *gin.Context) dto.PageRequest {
	req := dto.PageRequest{
		Skip:         queryInt(c, "skip", 0),
		Limit:        queryInt(c, "limit", 20),
		TimeRange:    c.Query("time_range"),
		Keyword:      c.Query("keyword"),
		BrandCode:    c.Query("brand_code"),
		ChannelCode:  c.Query("channel_code"),
		PlatformCode: c.Query("platform_code"),
		StoreCode:    c.Query("store_code"),
		ProductCode:  c.Query("product_code"),
		Status:       c.Query("status"),
		Level:        c.Query("level"),
		Domain:       c.Query("domain"),
		DataType:     c.Query("data_type"),
	}
	if start := parseDate(c.Query("start_date")); start != nil {
		req.StartDate = start
	}
	if end := parseDate(c.Query("end_date")); end != nil {
		req.EndDate = end
	}
	return req
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &parsed
}
