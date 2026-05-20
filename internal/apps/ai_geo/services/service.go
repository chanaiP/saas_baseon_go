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
	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/apps/ai_geo/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const (
	quotaBrandCount              = "ai_geo_brand_count"
	quotaProductCount            = "ai_geo_product_count"
	quotaChannelAccountCount     = "ai_geo_channel_account_count"
	quotaMonthlyDraftGenerations = "ai_geo_monthly_draft_generations"
	quotaMonthlyPublishTasks     = "ai_geo_monthly_publish_tasks"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("请求参数错误")
	ErrInvalidStatus = errors.New("状态流转不合法")
)

type Service struct {
	repo                    *repositories.Repository
	quota                   *quotaapp.Service
	draftGenerator          DraftGenerator
	channelContentGenerator ChannelContentGenerator
	auditAdvisor            AuditAdvisor
}

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo, quota: quotaapp.NewService(repo.DB()), draftGenerator: localDraftGenerator{}, channelContentGenerator: localChannelContentGenerator{}, auditAdvisor: localAuditAdvisor{}}
}

type DraftGenerationRequest struct {
	Viewer  dto.Viewer
	Payload dto.GenerateDraftPayload
	Brand   *models.AiGeoBrandCard
	Product *models.AiGeoProductCard
}

type DraftGenerationResult struct {
	Title    string
	Summary  string
	Body     string
	Keywords []string
	Source   string
}

type DraftGenerator interface {
	GenerateDraft(ctx context.Context, req DraftGenerationRequest) (DraftGenerationResult, error)
}

type ChannelContentGenerationRequest struct {
	Viewer  dto.Viewer
	Draft   models.AiGeoDraft
	Channel models.AiGeoChannelProfile
	Payload dto.ChannelContentPayload
}

type ChannelContentGenerationResult struct {
	Title string
	Body  string
}

type ChannelContentGenerator interface {
	GenerateChannelContent(ctx context.Context, req ChannelContentGenerationRequest) (ChannelContentGenerationResult, error)
}

type AuditAdviceRequest struct {
	Viewer      dto.Viewer
	ObjectType  string
	Draft       *models.AiGeoDraft
	Content     *models.AiGeoChannelContent
	Channel     *models.AiGeoChannelProfile
	ExtraPrompt string
}

type AuditAdviceResult struct {
	RiskLevel   string
	Passed      bool
	Summary     string
	Suggestions []map[string]interface{}
	ModelCode   string
}

type AuditAdvisor interface {
	Advise(ctx context.Context, req AuditAdviceRequest) (AuditAdviceResult, error)
}

func (s *Service) SetDraftGenerator(generator DraftGenerator) {
	if generator == nil {
		s.draftGenerator = localDraftGenerator{}
		return
	}
	s.draftGenerator = generator
}

func (s *Service) SetChannelContentGenerator(generator ChannelContentGenerator) {
	if generator == nil {
		s.channelContentGenerator = localChannelContentGenerator{}
		return
	}
	s.channelContentGenerator = generator
}

func (s *Service) SetAuditAdvisor(advisor AuditAdvisor) {
	if advisor == nil {
		s.auditAdvisor = localAuditAdvisor{}
		return
	}
	s.auditAdvisor = advisor
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
	if err := s.ensureStaticQuota(ctx, viewer.TenantID, quotaBrandCount, s.repo.CountBrands); err != nil {
		return models.AiGeoBrandCard{}, err
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

func (s *Service) ArchiveBrand(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoBrandCard, error) {
	row, err := s.repo.Brand(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	now := time.Now()
	userID := viewer.UserID
	row.Status = "archived"
	row.DeletedAt = &now
	row.UpdatedAt = now
	row.UpdatedBy = &userID
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
	if err := s.ensureStaticQuota(ctx, viewer.TenantID, quotaProductCount, s.repo.CountProducts); err != nil {
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

func (s *Service) UpdateProduct(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.ProductPayload) (models.AiGeoProductCard, error) {
	row, err := s.repo.Product(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if payload.BrandID > 0 && payload.BrandID != row.BrandID {
		if _, err := s.repo.Brand(ctx, viewer.TenantID, payload.BrandID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return row, ErrNotFound
			}
			return row, err
		}
		row.BrandID = payload.BrandID
	}
	if strings.TrimSpace(payload.ProductName) != "" {
		row.ProductName = strings.TrimSpace(payload.ProductName)
	}
	row.CategoryName = stringPtr(payload.CategoryName)
	row.SellingPoints = jsonString(payload.SellingPoints, []string{})
	row.FAQ = jsonString(payload.FAQ, []string{})
	row.ContentAngles = jsonString(payload.ContentAngles, []string{})
	row.Completeness = completeness(row.ProductName, payload.CategoryName, strings.Join(payload.SellingPoints, ","), strings.Join(payload.FAQ, ","))
	if payload.Status != "" {
		row.Status = payload.Status
	}
	userID := viewer.UserID
	row.UpdatedBy = &userID
	row.UpdatedAt = time.Now()
	err = s.repo.SaveProduct(ctx, &row)
	return row, err
}

func (s *Service) ArchiveProduct(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoProductCard, error) {
	row, err := s.repo.Product(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	now := time.Now()
	userID := viewer.UserID
	row.Status = "archived"
	row.DeletedAt = &now
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveProduct(ctx, &row)
	return row, err
}

func (s *Service) SKUs(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoSKU], error) {
	rows, total, err := s.repo.ListSKUs(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateSKU(ctx context.Context, viewer dto.Viewer, payload dto.SKUPayload) (models.AiGeoSKU, error) {
	payload.SKUCode = strings.TrimSpace(payload.SKUCode)
	payload.SKUName = strings.TrimSpace(payload.SKUName)
	if viewer.TenantID == 0 || payload.ProductID == 0 || payload.SKUCode == "" || payload.SKUName == "" {
		return models.AiGeoSKU{}, ErrInvalidInput
	}
	if _, err := s.repo.Product(ctx, viewer.TenantID, payload.ProductID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoSKU{}, ErrNotFound
		}
		return models.AiGeoSKU{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoSKU{
		TenantID:    viewer.TenantID,
		ProductID:   payload.ProductID,
		SKUCode:     payload.SKUCode,
		SKUName:     payload.SKUName,
		Attributes:  jsonString(payload.Attributes, map[string]interface{}{}),
		Price:       payload.Price,
		ImageURL:    stringPtr(payload.ImageURL),
		StockStatus: defaultString(payload.StockStatus, "unknown"),
		Status:      defaultString(payload.Status, "active"),
		CreatedBy:   &userID,
		UpdatedBy:   &userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := s.repo.SaveSKU(ctx, &row)
	return row, err
}

func (s *Service) UpdateSKU(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.SKUPayload) (models.AiGeoSKU, error) {
	row, err := s.repo.SKU(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if strings.TrimSpace(payload.SKUName) != "" {
		row.SKUName = strings.TrimSpace(payload.SKUName)
	}
	row.Attributes = jsonString(payload.Attributes, map[string]interface{}{})
	row.Price = payload.Price
	row.ImageURL = stringPtr(payload.ImageURL)
	if payload.StockStatus != "" {
		row.StockStatus = payload.StockStatus
	}
	if payload.Status != "" {
		row.Status = payload.Status
	}
	userID := viewer.UserID
	row.UpdatedBy = &userID
	row.UpdatedAt = time.Now()
	err = s.repo.SaveSKU(ctx, &row)
	return row, err
}

func (s *Service) ArchiveSKU(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoSKU, error) {
	row, err := s.repo.SKU(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	now := time.Now()
	userID := viewer.UserID
	row.Status = "archived"
	row.DeletedAt = &now
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveSKU(ctx, &row)
	return row, err
}

func (s *Service) Competitors(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoCompetitor], error) {
	rows, total, err := s.repo.ListCompetitors(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreateCompetitor(ctx context.Context, viewer dto.Viewer, payload dto.CompetitorPayload) (models.AiGeoCompetitor, error) {
	if viewer.TenantID == 0 || payload.ProductID == 0 || strings.TrimSpace(payload.BrandName) == "" || strings.TrimSpace(payload.ProductName) == "" {
		return models.AiGeoCompetitor{}, ErrInvalidInput
	}
	if _, err := s.repo.Product(ctx, viewer.TenantID, payload.ProductID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoCompetitor{}, ErrNotFound
		}
		return models.AiGeoCompetitor{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoCompetitor{
		TenantID:    viewer.TenantID,
		ProductID:   payload.ProductID,
		BrandName:   strings.TrimSpace(payload.BrandName),
		ProductName: strings.TrimSpace(payload.ProductName),
		PriceText:   stringPtr(payload.PriceText),
		Point:       stringPtr(payload.Point),
		Difference:  stringPtr(payload.Difference),
		Angle:       stringPtr(payload.Angle),
		LinkURL:     stringPtr(payload.LinkURL),
		Status:      defaultString(payload.Status, "active"),
		CreatedBy:   &userID,
		UpdatedBy:   &userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := s.repo.SaveCompetitor(ctx, &row)
	return row, err
}

func (s *Service) UpdateCompetitor(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.CompetitorPayload) (models.AiGeoCompetitor, error) {
	row, err := s.repo.Competitor(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if strings.TrimSpace(payload.BrandName) != "" {
		row.BrandName = strings.TrimSpace(payload.BrandName)
	}
	if strings.TrimSpace(payload.ProductName) != "" {
		row.ProductName = strings.TrimSpace(payload.ProductName)
	}
	row.PriceText = stringPtr(payload.PriceText)
	row.Point = stringPtr(payload.Point)
	row.Difference = stringPtr(payload.Difference)
	row.Angle = stringPtr(payload.Angle)
	row.LinkURL = stringPtr(payload.LinkURL)
	if payload.Status != "" {
		row.Status = payload.Status
	}
	userID := viewer.UserID
	row.UpdatedBy = &userID
	row.UpdatedAt = time.Now()
	err = s.repo.SaveCompetitor(ctx, &row)
	return row, err
}

func (s *Service) ArchiveCompetitor(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoCompetitor, error) {
	row, err := s.repo.Competitor(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	now := time.Now()
	userID := viewer.UserID
	row.Status = "archived"
	row.DeletedAt = &now
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveCompetitor(ctx, &row)
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
	if err := s.ensureStaticQuota(ctx, viewer.TenantID, quotaChannelAccountCount, s.repo.CountChannelAccounts); err != nil {
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
	if viewer.TenantID == 0 || prompt == "" {
		return models.AiGeoDraft{}, ErrInvalidInput
	}
	var brand *models.AiGeoBrandCard
	if payload.BrandID != nil && *payload.BrandID > 0 {
		row, err := s.repo.Brand(ctx, viewer.TenantID, *payload.BrandID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoDraft{}, ErrNotFound
		}
		if err != nil {
			return models.AiGeoDraft{}, err
		}
		brand = &row
	}
	var product *models.AiGeoProductCard
	if payload.ProductID != nil && *payload.ProductID > 0 {
		row, err := s.repo.Product(ctx, viewer.TenantID, *payload.ProductID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoDraft{}, ErrNotFound
		}
		if err != nil {
			return models.AiGeoDraft{}, err
		}
		product = &row
	}
	if err := s.consumeQuota(ctx, viewer.TenantID, quotaMonthlyDraftGenerations); err != nil {
		return models.AiGeoDraft{}, err
	}
	generator := s.draftGenerator
	if generator == nil {
		generator = localDraftGenerator{}
	}
	generated, err := generator.GenerateDraft(ctx, DraftGenerationRequest{Viewer: viewer, Payload: payload, Brand: brand, Product: product})
	if err != nil {
		return models.AiGeoDraft{}, err
	}
	return s.CreateDraft(ctx, viewer, dto.DraftPayload{
		BrandID:   payload.BrandID,
		ProductID: payload.ProductID,
		Title:     generated.Title,
		Summary:   generated.Summary,
		Body:      generated.Body,
		Keywords:  generated.Keywords,
		Source:    defaultString(generated.Source, "ai_workbench"),
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

func (s *Service) GenerateDraftAuditSuggestion(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoAuditSuggestion, error) {
	draft, err := s.repo.Draft(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AiGeoAuditSuggestion{}, ErrNotFound
	}
	if err != nil {
		return models.AiGeoAuditSuggestion{}, err
	}
	advisor := s.auditAdvisor
	if advisor == nil {
		advisor = localAuditAdvisor{}
	}
	result, adviceErr := advisor.Advise(ctx, AuditAdviceRequest{Viewer: viewer, ObjectType: "draft", Draft: &draft})
	row := auditSuggestionRow(viewer, "draft", draft.ID, draft.DraftCode, auditSuggestionScenarioCode, result, adviceErr)
	err = s.repo.SaveAuditSuggestion(ctx, &row)
	if err != nil {
		return row, err
	}
	return row, adviceErr
}

func (s *Service) DraftAuditSuggestions(ctx context.Context, viewer dto.Viewer, id uint64, req dto.PageRequest) (dto.PageResponse[models.AiGeoAuditSuggestion], error) {
	if _, err := s.repo.Draft(ctx, viewer.TenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PageResponse[models.AiGeoAuditSuggestion]{}, ErrNotFound
		}
		return dto.PageResponse[models.AiGeoAuditSuggestion]{}, err
	}
	rows, total, err := s.repo.ListAuditSuggestions(ctx, viewer.TenantID, "draft", id, req)
	return page(rows, total, req), err
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
	channel, err := s.repo.Channel(ctx, viewer.TenantID, payload.ChannelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoChannelContent{}, ErrNotFound
		}
		return models.AiGeoChannelContent{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	generator := s.channelContentGenerator
	if generator == nil {
		generator = localChannelContentGenerator{}
	}
	generated, err := generator.GenerateChannelContent(ctx, ChannelContentGenerationRequest{Viewer: viewer, Draft: draft, Channel: channel, Payload: payload})
	if err != nil {
		return models.AiGeoChannelContent{}, err
	}
	content := models.AiGeoChannelContent{
		TenantID:      viewer.TenantID,
		DraftID:       draft.ID,
		ChannelID:     payload.ChannelID,
		Title:         generated.Title,
		Body:          generated.Body,
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

func (s *Service) GenerateChannelContentAuditSuggestion(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoAuditSuggestion, error) {
	content, err := s.repo.ChannelContent(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AiGeoAuditSuggestion{}, ErrNotFound
	}
	if err != nil {
		return models.AiGeoAuditSuggestion{}, err
	}
	var channel *models.AiGeoChannelProfile
	if row, err := s.repo.Channel(ctx, viewer.TenantID, content.ChannelID); err == nil {
		channel = &row
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AiGeoAuditSuggestion{}, err
	}
	advisor := s.auditAdvisor
	if advisor == nil {
		advisor = localAuditAdvisor{}
	}
	result, adviceErr := advisor.Advise(ctx, AuditAdviceRequest{Viewer: viewer, ObjectType: "channel_content", Content: &content, Channel: channel})
	row := auditSuggestionRow(viewer, "channel_content", content.ID, fmt.Sprintf("%d", content.ID), auditSuggestionScenarioCode, result, adviceErr)
	err = s.repo.SaveAuditSuggestion(ctx, &row)
	if err != nil {
		return row, err
	}
	return row, adviceErr
}

func (s *Service) ChannelContentAuditSuggestions(ctx context.Context, viewer dto.Viewer, id uint64, req dto.PageRequest) (dto.PageResponse[models.AiGeoAuditSuggestion], error) {
	if _, err := s.repo.ChannelContent(ctx, viewer.TenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PageResponse[models.AiGeoAuditSuggestion]{}, ErrNotFound
		}
		return dto.PageResponse[models.AiGeoAuditSuggestion]{}, err
	}
	rows, total, err := s.repo.ListAuditSuggestions(ctx, viewer.TenantID, "channel_content", id, req)
	return page(rows, total, req), err
}

func (s *Service) PublishPlans(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoPublishPlan], error) {
	rows, total, err := s.repo.ListPublishPlans(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) CreatePublishPlan(ctx context.Context, viewer dto.Viewer, payload dto.PublishPlanPayload) (models.AiGeoPublishPlan, error) {
	if viewer.TenantID == 0 || payload.ChannelContentID == 0 || payload.ChannelID == 0 || payload.ScheduledAt == "" {
		return models.AiGeoPublishPlan{}, ErrInvalidInput
	}
	content, err := s.repo.ChannelContent(ctx, viewer.TenantID, payload.ChannelContentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AiGeoPublishPlan{}, ErrNotFound
	}
	if err != nil {
		return models.AiGeoPublishPlan{}, err
	}
	if content.ChannelID != payload.ChannelID {
		return models.AiGeoPublishPlan{}, ErrInvalidInput
	}
	if _, err := s.repo.Channel(ctx, viewer.TenantID, payload.ChannelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoPublishPlan{}, ErrNotFound
		}
		return models.AiGeoPublishPlan{}, err
	}
	scheduledAt, err := time.Parse(time.RFC3339, payload.ScheduledAt)
	if err != nil {
		return models.AiGeoPublishPlan{}, ErrInvalidInput
	}
	if err := s.consumeQuota(ctx, viewer.TenantID, quotaMonthlyPublishTasks); err != nil {
		return models.AiGeoPublishPlan{}, err
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
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		return tx.Model(&models.AiGeoChannelContent{}).
			Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, row.ChannelContentID).
			Updates(map[string]interface{}{"publish_status": "planned", "updated_at": now, "updated_by": &userID}).Error
	})
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
	if !validPublishTransition(row.Status, next) {
		return row, ErrInvalidStatus
	}
	row.Status = next
	row.PublishedURL = stringPtr(payload.PublishedURL)
	row.FailReason = stringPtr(payload.FailReason)
	now := time.Now()
	row.UpdatedAt = now
	userID := viewer.UserID
	row.UpdatedBy = &userID
	contentStatus := channelContentPublishStatus(next)
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		return tx.Model(&models.AiGeoChannelContent{}).
			Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", viewer.TenantID, row.ChannelContentID).
			Updates(map[string]interface{}{"publish_status": contentStatus, "updated_at": now, "updated_by": &userID}).Error
	})
	return row, err
}

func (s *Service) ImportMaterials(ctx context.Context, viewer dto.Viewer, payload dto.ImportPayload) (models.AiGeoImportBatch, error) {
	if viewer.TenantID == 0 || strings.TrimSpace(payload.ImportType) == "" {
		return models.AiGeoImportBatch{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	importErrors := validateImportRecords(viewer.TenantID, payload.Records, payload.ImportType, now)
	successCount := int64(len(payload.Records) - len(importErrors))
	failedCount := int64(len(importErrors))
	status := "completed"
	if failedCount > 0 && successCount > 0 {
		status = "partial_success"
	} else if failedCount > 0 {
		status = "failed"
	}
	batch := models.AiGeoImportBatch{
		TenantID:      viewer.TenantID,
		BatchCode:     code("IMPORT", now),
		ImportType:    strings.TrimSpace(payload.ImportType),
		MappingConfig: jsonString(payload.MappingConfig, map[string]interface{}{}),
		RecordCount:   int64(len(payload.Records)),
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		Status:        status,
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&batch).Error; err != nil {
			return err
		}
		for i := range importErrors {
			importErrors[i].BatchID = batch.ID
		}
		if len(importErrors) > 0 {
			return tx.Create(&importErrors).Error
		}
		return nil
	})
	return batch, err
}

func (s *Service) ImportErrors(ctx context.Context, viewer dto.Viewer, batchID uint64, req dto.PageRequest) (dto.PageResponse[models.AiGeoImportError], error) {
	if viewer.TenantID == 0 || batchID == 0 {
		return dto.PageResponse[models.AiGeoImportError]{}, ErrInvalidInput
	}
	if _, err := s.repo.ImportBatch(ctx, viewer.TenantID, batchID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PageResponse[models.AiGeoImportError]{}, ErrNotFound
		}
		return dto.PageResponse[models.AiGeoImportError]{}, err
	}
	rows, total, err := s.repo.ListImportErrors(ctx, viewer.TenantID, batchID, req)
	return page(rows, total, req), err
}

func (s *Service) ensureStaticQuota(ctx context.Context, tenantID uint64, quotaCode string, currentCount func(context.Context, uint64) (int64, error)) error {
	if s.quota == nil {
		return nil
	}
	limit, quota, ok, err := s.quota.CurrentLimitByCode(ctx, tenantID, quotaCode)
	if err != nil || !ok || limit < 0 {
		return err
	}
	count, err := currentCount(ctx, tenantID)
	if err != nil {
		return err
	}
	if int(count)+1 > limit {
		return &quotaapp.ExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: int(count)}
	}
	return nil
}

func (s *Service) consumeQuota(ctx context.Context, tenantID uint64, quotaCode string) error {
	if s.quota == nil {
		return nil
	}
	return s.quota.Consume(ctx, tenantID, quotaCode, 1)
}

func validateImportRecords(tenantID uint64, records []map[string]interface{}, importType string, now time.Time) []models.AiGeoImportError {
	importType = strings.ToLower(strings.TrimSpace(importType))
	required := []string{}
	switch importType {
	case "brand", "brands":
		required = []string{"brand_code", "brand_name"}
	case "product", "products":
		required = []string{"product_code", "product_name"}
	default:
		required = []string{}
	}
	errorsOut := []models.AiGeoImportError{}
	for index, record := range records {
		rowNumber := index + 1
		if len(record) == 0 {
			errorsOut = append(errorsOut, importErrorRow(tenantID, rowNumber, "", "empty_row", "导入行为空", record, now))
			continue
		}
		for _, field := range required {
			if strings.TrimSpace(fmt.Sprint(record[field])) == "" || strings.TrimSpace(fmt.Sprint(record[field])) == "<nil>" {
				errorsOut = append(errorsOut, importErrorRow(tenantID, rowNumber, field, "required", "必填字段缺失", record, now))
			}
		}
	}
	return errorsOut
}

func importErrorRow(tenantID uint64, rowNumber int, fieldName string, errorCode string, message string, raw map[string]interface{}, now time.Time) models.AiGeoImportError {
	return models.AiGeoImportError{
		TenantID:     tenantID,
		RowNumber:    rowNumber,
		FieldName:    stringPtr(fieldName),
		ErrorCode:    errorCode,
		ErrorMessage: message,
		RawData:      jsonString(raw, map[string]interface{}{}),
		Status:       "active",
		CreatedAt:    now,
	}
}

func auditSuggestionRow(viewer dto.Viewer, objectType string, objectID uint64, objectCode string, scenarioCode string, result AuditAdviceResult, err error) models.AiGeoAuditSuggestion {
	now := time.Now()
	status := "success"
	var errorMessage *string
	if err != nil {
		status = "failed"
		errorMessage = stringPtr(safeAIError(err))
	}
	return models.AiGeoAuditSuggestion{
		TenantID:       viewer.TenantID,
		ObjectType:     objectType,
		ObjectID:       objectID,
		ObjectCode:     stringPtr(objectCode),
		ScenarioCode:   scenarioCode,
		RiskLevel:      defaultString(result.RiskLevel, "low"),
		Passed:         result.Passed,
		Summary:        stringPtr(result.Summary),
		SuggestionJSON: jsonString(result.Suggestions, []map[string]interface{}{}),
		ModelCode:      stringPtr(result.ModelCode),
		Status:         status,
		ErrorMessage:   errorMessage,
		GeneratedAt:    now,
		CreatedBy:      &viewer.UserID,
		CreatedAt:      now,
	}
}

func safeAIError(err error) string {
	msg := strings.TrimSpace(err.Error())
	lower := strings.ToLower(msg)
	for _, token := range []string{"secret", "token", "authorization", "password", "select ", "insert ", "update ", "delete ", "panic", "stack"} {
		if strings.Contains(lower, token) {
			return "AI 审核建议生成失败，请稍后重试"
		}
	}
	if msg == "" {
		return "AI 审核建议生成失败"
	}
	if len([]rune(msg)) > 160 {
		msg = string([]rune(msg)[:160])
	}
	return msg
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

func validPublishTransition(current, next string) bool {
	if current == next {
		return true
	}
	allowed := map[string][]string{
		"scheduled":  {"publishing", "published", "failed", "cancelled"},
		"publishing": {"published", "failed", "cancelled"},
		"failed":     {"scheduled", "publishing"},
	}
	for _, item := range allowed[current] {
		if item == next {
			return true
		}
	}
	return false
}

func channelContentPublishStatus(planStatus string) string {
	switch planStatus {
	case "scheduled":
		return "planned"
	case "publishing":
		return "publishing"
	case "published":
		return "published"
	case "failed":
		return "failed"
	case "cancelled":
		return "not_planned"
	default:
		return "not_planned"
	}
}
