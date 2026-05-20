package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/apps/ai_geo/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("请求参数错误")
	ErrInvalidStatus = errors.New("状态流转不合法")
)

type Service struct {
	repo *repositories.Repository
}

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordAudit(ctx context.Context, viewer dto.Viewer, meta dto.RequestMeta, action string, objectCode string, summary string, detail interface{}) error {
	appCode := "ai-geo"
	tenantID := viewer.TenantID
	userID := viewer.UserID
	detailRaw := jsonString(detail, map[string]interface{}{})
	log := models.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		AppCode:   &appCode,
		Module:    "ai_geo",
		Action:    action,
		Summary:   summary,
		Detail:    &detailRaw,
		IP:        stringPtr(meta.IP),
		UserAgent: stringPtr(meta.UserAgent),
		RequestID: stringPtr(meta.RequestID),
		Result:    "success",
		CreatedAt: time.Now(),
	}
	return s.repo.DB().WithContext(ctx).Create(&log).Error
}

func (s *Service) Overview(ctx context.Context, viewer dto.Viewer) (dto.Overview, error) {
	if viewer.TenantID == 0 {
		return dto.Overview{}, ErrInvalidInput
	}
	tenantID := viewer.TenantID
	brandCount, err := s.repo.CountBrands(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	productCount, err := s.repo.CountProducts(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	skuCount, err := s.repo.CountSKUs(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	channelCount, err := s.repo.CountChannels(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	accountCount, err := s.repo.CountChannelAccounts(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	draftsToday, err := s.repo.CountDraftsToday(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	pendingDrafts, err := s.repo.CountPendingDrafts(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	channelContents, err := s.repo.CountChannelContents(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	plansToday, err := s.repo.CountPublishPlansToday(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	completeness, err := s.repo.AverageCompleteness(ctx, tenantID)
	if err != nil {
		return dto.Overview{}, err
	}
	tasks := []map[string]string{}
	if pendingDrafts > 0 {
		tasks = append(tasks, map[string]string{"tag": "母稿审核", "title": fmt.Sprintf("%d 篇母稿待审核", pendingDrafts)})
	}
	if accountCount == 0 {
		tasks = append(tasks, map[string]string{"tag": "渠道账号", "title": "至少接入 1 个可发布账号"})
	}
	if productCount == 0 {
		tasks = append(tasks, map[string]string{"tag": "资料中心", "title": "先维护商品资料卡"})
	}
	return dto.Overview{
		BrandCount:          brandCount,
		ProductCount:        productCount,
		SKUCount:            skuCount,
		ChannelCount:        channelCount,
		ChannelAccountCount: accountCount,
		DraftCountToday:     draftsToday,
		PendingDraftCount:   pendingDrafts,
		ChannelContentCount: channelContents,
		PublishPlanToday:    plansToday,
		AverageCompleteness: completeness,
		PendingTasks:        tasks,
		QuotaUsage: map[string]interface{}{
			"ai_geo_brand_count":               brandCount,
			"ai_geo_product_count":             productCount,
			"ai_geo_channel_account_count":     accountCount,
			"ai_geo_monthly_draft_generations": draftsToday,
			"ai_geo_monthly_publish_tasks":     plansToday,
		},
	}, nil
}

func (s *Service) Brands(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoBrandCard], error) {
	rows, total, err := s.repo.ListBrands(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateBrand(ctx context.Context, viewer dto.Viewer, payload dto.BrandPayload) (models.AiGeoBrandCard, error) {
	payload.BrandCode = strings.TrimSpace(payload.BrandCode)
	payload.BrandName = strings.TrimSpace(payload.BrandName)
	if viewer.TenantID == 0 || payload.BrandCode == "" || payload.BrandName == "" {
		return models.AiGeoBrandCard{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoBrandCard{
		TenantID:       viewer.TenantID,
		BrandCode:      payload.BrandCode,
		BrandName:      payload.BrandName,
		Positioning:    stringPtr(payload.Positioning),
		TargetAudience: stringPtr(payload.TargetAudience),
		PriceBand:      stringPtr(payload.PriceBand),
		Tone:           stringPtr(payload.Tone),
		Keywords:       jsonString(payload.Keywords, []string{}),
		Completeness:   completeness(payload.BrandName, payload.Positioning, payload.TargetAudience, payload.PriceBand, payload.Tone),
		Status:         defaultString(payload.Status, "active"),
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	err := s.repo.SaveBrand(ctx, &row)
	return row, err
}

func (s *Service) UpdateBrand(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.BrandPayload) (models.AiGeoBrandCard, error) {
	row, err := s.repo.Brand(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if strings.TrimSpace(payload.BrandName) != "" {
		row.BrandName = strings.TrimSpace(payload.BrandName)
	}
	row.Positioning = stringPtr(payload.Positioning)
	row.TargetAudience = stringPtr(payload.TargetAudience)
	row.PriceBand = stringPtr(payload.PriceBand)
	row.Tone = stringPtr(payload.Tone)
	row.Keywords = jsonString(payload.Keywords, []string{})
	row.Completeness = completeness(row.BrandName, payload.Positioning, payload.TargetAudience, payload.PriceBand, payload.Tone)
	if payload.Status != "" {
		row.Status = payload.Status
	}
	userID := viewer.UserID
	row.UpdatedBy = &userID
	row.UpdatedAt = time.Now()
	err = s.repo.SaveBrand(ctx, &row)
	return row, err
}

func (s *Service) Products(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoProductCard], error) {
	rows, total, err := s.repo.ListProducts(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateProduct(ctx context.Context, viewer dto.Viewer, payload dto.ProductPayload) (models.AiGeoProductCard, error) {
	if viewer.TenantID == 0 || payload.BrandID == 0 || strings.TrimSpace(payload.ProductCode) == "" || strings.TrimSpace(payload.ProductName) == "" {
		return models.AiGeoProductCard{}, ErrInvalidInput
	}
	if _, err := s.repo.Brand(ctx, viewer.TenantID, payload.BrandID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoProductCard{}, ErrNotFound
		}
		return models.AiGeoProductCard{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoProductCard{
		TenantID:      viewer.TenantID,
		BrandID:       payload.BrandID,
		ProductCode:   strings.TrimSpace(payload.ProductCode),
		ProductName:   strings.TrimSpace(payload.ProductName),
		CategoryName:  stringPtr(payload.CategoryName),
		SellingPoints: jsonString(payload.SellingPoints, []string{}),
		FAQ:           jsonString(payload.FAQ, []string{}),
		ContentAngles: jsonString(payload.ContentAngles, []string{}),
		Completeness:  completeness(payload.ProductName, payload.CategoryName, strings.Join(payload.SellingPoints, ","), strings.Join(payload.FAQ, ",")),
		Status:        defaultString(payload.Status, "active"),
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err := s.repo.SaveProduct(ctx, &row)
	return row, err
}

func (s *Service) Channels(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoChannelProfile], error) {
	rows, total, err := s.repo.ListChannels(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateChannel(ctx context.Context, viewer dto.Viewer, payload dto.ChannelPayload) (models.AiGeoChannelProfile, error) {
	if viewer.TenantID == 0 || strings.TrimSpace(payload.ChannelCode) == "" || strings.TrimSpace(payload.ChannelName) == "" {
		return models.AiGeoChannelProfile{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoChannelProfile{
		TenantID:           viewer.TenantID,
		ChannelCode:        strings.TrimSpace(payload.ChannelCode),
		ChannelName:        strings.TrimSpace(payload.ChannelName),
		ChannelType:        defaultString(payload.ChannelType, "content"),
		EntryURL:           stringPtr(payload.EntryURL),
		ContentForms:       jsonString(payload.ContentForms, []string{}),
		SupportModes:       jsonString(payload.SupportModes, []string{}),
		DefaultPublishMode: defaultString(payload.DefaultPublishMode, "manual"),
		Status:             defaultString(payload.Status, "active"),
		CreatedBy:          &userID,
		UpdatedBy:          &userID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	err := s.repo.SaveChannel(ctx, &row)
	return row, err
}

func (s *Service) ChannelAccounts(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoChannelAccount], error) {
	rows, total, err := s.repo.ListChannelAccounts(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateChannelAccount(ctx context.Context, viewer dto.Viewer, payload dto.ChannelAccountPayload) (models.AiGeoChannelAccount, error) {
	if viewer.TenantID == 0 || payload.ChannelID == 0 || strings.TrimSpace(payload.AccountName) == "" {
		return models.AiGeoChannelAccount{}, ErrInvalidInput
	}
	if _, err := s.repo.Channel(ctx, viewer.TenantID, payload.ChannelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoChannelAccount{}, ErrNotFound
		}
		return models.AiGeoChannelAccount{}, err
	}
	var expiresAt *time.Time
	if payload.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, payload.ExpiresAt)
		if err != nil {
			return models.AiGeoChannelAccount{}, ErrInvalidInput
		}
		expiresAt = &parsed
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoChannelAccount{
		TenantID:          viewer.TenantID,
		ChannelID:         payload.ChannelID,
		AccountName:       strings.TrimSpace(payload.AccountName),
		ExternalAccountID: stringPtr(payload.ExternalAccountID),
		AuthStatus:        defaultString(payload.AuthStatus, "not_authorized"),
		PublishStatus:     defaultString(payload.PublishStatus, "unavailable"),
		ExpiresAt:         expiresAt,
		Status:            defaultString(payload.Status, "active"),
		CreatedBy:         &userID,
		UpdatedBy:         &userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	err := s.repo.SaveChannelAccount(ctx, &row)
	return row, err
}

func (s *Service) Drafts(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoDraft], error) {
	rows, total, err := s.repo.ListDrafts(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateDraft(ctx context.Context, viewer dto.Viewer, payload dto.DraftPayload) (models.AiGeoDraft, error) {
	if viewer.TenantID == 0 || strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Body) == "" {
		return models.AiGeoDraft{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoDraft{
		TenantID:      viewer.TenantID,
		DraftCode:     code("DRAFT", now),
		BrandID:       payload.BrandID,
		ProductID:     payload.ProductID,
		Title:         strings.TrimSpace(payload.Title),
		Summary:       stringPtr(payload.Summary),
		Body:          strings.TrimSpace(payload.Body),
		Keywords:      jsonString(payload.Keywords, []string{}),
		Source:        defaultString(payload.Source, "manual"),
		AuditStatus:   "draft",
		ChannelStatus: "not_generated",
		Status:        "active",
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err := s.repo.SaveDraft(ctx, &row)
	return row, err
}

func (s *Service) GenerateDraft(ctx context.Context, viewer dto.Viewer, payload dto.GenerateDraftPayload) (models.AiGeoDraft, error) {
	prompt := strings.TrimSpace(payload.Prompt)
	if prompt == "" {
		return models.AiGeoDraft{}, ErrInvalidInput
	}
	title := prompt
	if len([]rune(title)) > 42 {
		title = string([]rune(title)[:42])
	}
	body := fmt.Sprintf("围绕“%s”生成一篇可进入审核的母稿。\n\n创作要求：结合品牌资料、商品卖点、渠道语境和热点素材，输出结构化内容，后续可生成小红书、知乎、独立站等渠道版本。", prompt)
	return s.CreateDraft(ctx, viewer, dto.DraftPayload{
		BrandID:   payload.BrandID,
		ProductID: payload.ProductID,
		Title:     title,
		Summary:   "AI 工作台生成母稿，待人工审核确认。",
		Body:      body,
		Keywords:  []string{"AI生成", "母稿"},
		Source:    "ai_workbench",
	})
}

func (s *Service) SubmitDraft(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoDraft, error) {
	row, err := s.repo.Draft(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.AuditStatus != "draft" && row.AuditStatus != "rejected" {
		return row, ErrInvalidStatus
	}
	row.AuditStatus = "pending"
	row.UpdatedAt = time.Now()
	userID := viewer.UserID
	row.UpdatedBy = &userID
	err = s.repo.SaveDraft(ctx, &row)
	return row, err
}

func (s *Service) ReviewDraft(ctx context.Context, viewer dto.Viewer, id uint64, approved bool) (models.AiGeoDraft, error) {
	row, err := s.repo.Draft(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.AuditStatus != "pending" {
		return row, ErrInvalidStatus
	}
	if approved {
		row.AuditStatus = "approved"
	} else {
		row.AuditStatus = "rejected"
	}
	row.UpdatedAt = time.Now()
	userID := viewer.UserID
	row.UpdatedBy = &userID
	err = s.repo.SaveDraft(ctx, &row)
	return row, err
}

func (s *Service) GenerateChannelContent(ctx context.Context, viewer dto.Viewer, draftID uint64, payload dto.ChannelContentPayload) (models.AiGeoChannelContent, error) {
	if payload.ChannelID == 0 {
		return models.AiGeoChannelContent{}, ErrInvalidInput
	}
	draft, err := s.repo.Draft(ctx, viewer.TenantID, draftID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AiGeoChannelContent{}, ErrNotFound
	}
	if err != nil {
		return models.AiGeoChannelContent{}, err
	}
	if _, err := s.repo.Channel(ctx, viewer.TenantID, payload.ChannelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoChannelContent{}, ErrNotFound
		}
		return models.AiGeoChannelContent{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	title := defaultString(payload.Title, draft.Title)
	body := defaultString(payload.Body, draft.Body)
	content := models.AiGeoChannelContent{
		TenantID:      viewer.TenantID,
		DraftID:       draft.ID,
		ChannelID:     payload.ChannelID,
		Title:         title,
		Body:          body,
		AuditStatus:   "pending",
		PublishStatus: "not_planned",
		Status:        "active",
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.SaveChannelContent(ctx, &content); err != nil {
		return content, err
	}
	draft.ChannelStatus = "generated"
	draft.UpdatedAt = now
	draft.UpdatedBy = &userID
	_ = s.repo.SaveDraft(ctx, &draft)
	return content, nil
}

func (s *Service) PublishPlans(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoPublishPlan], error) {
	rows, total, err := s.repo.ListPublishPlans(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreatePublishPlan(ctx context.Context, viewer dto.Viewer, payload dto.PublishPlanPayload) (models.AiGeoPublishPlan, error) {
	if viewer.TenantID == 0 || payload.ChannelContentID == 0 || payload.ChannelID == 0 || payload.ScheduledAt == "" {
		return models.AiGeoPublishPlan{}, ErrInvalidInput
	}
	scheduledAt, err := time.Parse(time.RFC3339, payload.ScheduledAt)
	if err != nil {
		return models.AiGeoPublishPlan{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoPublishPlan{
		TenantID:         viewer.TenantID,
		PlanCode:         code("PLAN", now),
		ChannelContentID: payload.ChannelContentID,
		ChannelID:        payload.ChannelID,
		ScheduledAt:      scheduledAt,
		PublishMethod:    defaultString(payload.PublishMethod, "manual"),
		AutomationLevel:  defaultString(payload.AutomationLevel, "manual"),
		Status:           "scheduled",
		CreatedBy:        &userID,
		UpdatedBy:        &userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	err = s.repo.SavePublishPlan(ctx, &row)
	return row, err
}

func (s *Service) UpdatePublishStatus(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.PublishStatusPayload) (models.AiGeoPublishPlan, error) {
	row, err := s.repo.PublishPlan(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	next := defaultString(payload.Status, row.Status)
	if !validPublishStatus(next) {
		return row, ErrInvalidStatus
	}
	row.Status = next
	row.PublishedURL = stringPtr(payload.PublishedURL)
	row.FailReason = stringPtr(payload.FailReason)
	row.UpdatedAt = time.Now()
	userID := viewer.UserID
	row.UpdatedBy = &userID
	err = s.repo.SavePublishPlan(ctx, &row)
	return row, err
}

func (s *Service) ImportMaterials(ctx context.Context, viewer dto.Viewer, payload dto.ImportPayload) (models.AiGeoImportBatch, error) {
	if viewer.TenantID == 0 || strings.TrimSpace(payload.ImportType) == "" {
		return models.AiGeoImportBatch{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	batch := models.AiGeoImportBatch{
		TenantID:      viewer.TenantID,
		BatchCode:     code("IMPORT", now),
		ImportType:    strings.TrimSpace(payload.ImportType),
		MappingConfig: jsonString(payload.MappingConfig, map[string]interface{}{}),
		RecordCount:   int64(len(payload.Records)),
		SuccessCount:  int64(len(payload.Records)),
		FailedCount:   0,
		Status:        "completed",
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err := s.repo.SaveImportBatch(ctx, &batch)
	return batch, err
}

func page[T any](rows []T, total int64, req dto.PageRequest) dto.PageResponse[T] {
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	skip := req.Skip
	if skip < 0 {
		skip = 0
	}
	return dto.PageResponse[T]{Items: rows, Total: total, Skip: skip, Limit: limit}
}

func jsonString(value interface{}, fallback interface{}) string {
	if value == nil {
		value = fallback
	}
	raw, err := json.Marshal(value)
	if err != nil {
		raw, _ = json.Marshal(fallback)
	}
	return string(raw)
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func completeness(values ...string) int {
	total := len(values)
	if total == 0 {
		return 0
	}
	filled := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filled++
		}
	}
	return filled * 100 / total
}

func code(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%s-%06d", prefix, now.Format("20060102150405"), now.Nanosecond()/1000%1000000)
}

func validPublishStatus(status string) bool {
	switch status {
	case "scheduled", "publishing", "published", "failed", "cancelled":
		return true
	default:
		return false
	}
}
