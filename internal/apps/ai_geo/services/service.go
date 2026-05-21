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
	SKUs    []models.AiGeoSKU
	Hotspot *models.AiGeoHotspot
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

func (s *Service) Brand(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoBrandCard, error) {
	row, err := s.repo.Brand(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
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

func (s *Service) Product(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoProductCard, error) {
	row, err := s.repo.Product(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
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

func (s *Service) SKU(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoSKU, error) {
	row, err := s.repo.SKU(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
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

func (s *Service) Competitor(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoCompetitor, error) {
	row, err := s.repo.Competitor(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
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

func (s *Service) Keywords(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoKeyword], error) {
	rows, total, err := s.repo.ListKeywords(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Keyword(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoKeyword, error) {
	row, err := s.repo.Keyword(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) CreateKeyword(ctx context.Context, viewer dto.Viewer, payload dto.KeywordPayload) (models.AiGeoKeyword, error) {
	if viewer.TenantID == 0 || strings.TrimSpace(payload.Keyword) == "" {
		return models.AiGeoKeyword{}, ErrInvalidInput
	}
	if err := s.validateKeywordScope(ctx, viewer, payload.BrandID, payload.ProductID); err != nil {
		return models.AiGeoKeyword{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoKeyword{
		TenantID:     viewer.TenantID,
		BrandID:      positiveUint64Ptr(payload.BrandID),
		ProductID:    positiveUint64Ptr(payload.ProductID),
		KeywordGroup: defaultString(strings.TrimSpace(payload.KeywordGroup), "通用关键词"),
		Keyword:      strings.TrimSpace(payload.Keyword),
		Intent:       stringPtr(payload.Intent),
		Source:       defaultString(strings.TrimSpace(payload.Source), "manual"),
		Weight:       payload.Weight,
		Status:       defaultString(strings.TrimSpace(payload.Status), "active"),
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err := s.repo.SaveKeyword(ctx, &row)
	return row, err
}

func (s *Service) UpdateKeyword(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.KeywordPayload) (models.AiGeoKeyword, error) {
	row, err := s.repo.Keyword(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if payload.BrandID != nil || payload.ProductID != nil {
		if err := s.validateKeywordScope(ctx, viewer, payload.BrandID, payload.ProductID); err != nil {
			return row, err
		}
		row.BrandID = positiveUint64Ptr(payload.BrandID)
		row.ProductID = positiveUint64Ptr(payload.ProductID)
	}
	if strings.TrimSpace(payload.KeywordGroup) != "" {
		row.KeywordGroup = strings.TrimSpace(payload.KeywordGroup)
	}
	if strings.TrimSpace(payload.Keyword) != "" {
		row.Keyword = strings.TrimSpace(payload.Keyword)
	}
	row.Intent = stringPtr(payload.Intent)
	if strings.TrimSpace(payload.Source) != "" {
		row.Source = strings.TrimSpace(payload.Source)
	}
	row.Weight = payload.Weight
	if strings.TrimSpace(payload.Status) != "" {
		row.Status = strings.TrimSpace(payload.Status)
	}
	now := time.Now()
	userID := viewer.UserID
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveKeyword(ctx, &row)
	return row, err
}

func (s *Service) ArchiveKeyword(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoKeyword, error) {
	row, err := s.repo.Keyword(ctx, viewer.TenantID, id)
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
	err = s.repo.SaveKeyword(ctx, &row)
	return row, err
}

func (s *Service) validateKeywordScope(ctx context.Context, viewer dto.Viewer, brandID *uint64, productID *uint64) error {
	if brandID != nil && *brandID > 0 {
		if _, err := s.repo.Brand(ctx, viewer.TenantID, *brandID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
	}
	if productID != nil && *productID > 0 {
		product, err := s.repo.Product(ctx, viewer.TenantID, *productID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if brandID != nil && *brandID > 0 && product.BrandID != *brandID {
			return ErrInvalidInput
		}
	}
	return nil
}

func (s *Service) MaterialAssets(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoMaterialAsset], error) {
	rows, total, err := s.repo.ListMaterialAssets(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) MaterialAsset(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoMaterialAsset, error) {
	row, err := s.repo.MaterialAsset(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) CreateMaterialAsset(ctx context.Context, viewer dto.Viewer, payload dto.MaterialAssetPayload) (models.AiGeoMaterialAsset, error) {
	payload.AssetType = strings.TrimSpace(payload.AssetType)
	payload.AssetName = strings.TrimSpace(payload.AssetName)
	if viewer.TenantID == 0 || payload.AssetType == "" || payload.AssetName == "" {
		return models.AiGeoMaterialAsset{}, ErrInvalidInput
	}
	if payload.BrandID != nil && *payload.BrandID > 0 {
		if _, err := s.repo.Brand(ctx, viewer.TenantID, *payload.BrandID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.AiGeoMaterialAsset{}, ErrNotFound
			}
			return models.AiGeoMaterialAsset{}, err
		}
	}
	if payload.ProductID != nil && *payload.ProductID > 0 {
		if _, err := s.repo.Product(ctx, viewer.TenantID, *payload.ProductID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.AiGeoMaterialAsset{}, ErrNotFound
			}
			return models.AiGeoMaterialAsset{}, err
		}
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoMaterialAsset{
		TenantID:  viewer.TenantID,
		BrandID:   positiveUint64Ptr(payload.BrandID),
		ProductID: positiveUint64Ptr(payload.ProductID),
		AssetType: payload.AssetType,
		AssetName: payload.AssetName,
		URL:       stringPtr(payload.URL),
		Metadata:  jsonString(payload.Metadata, map[string]interface{}{}),
		Status:    defaultString(payload.Status, "active"),
		CreatedBy: &userID,
		UpdatedBy: &userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := s.repo.SaveMaterialAsset(ctx, &row)
	return row, err
}

func (s *Service) UpdateMaterialAsset(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.MaterialAssetPayload) (models.AiGeoMaterialAsset, error) {
	row, err := s.repo.MaterialAsset(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if payload.BrandID != nil && *payload.BrandID > 0 {
		if _, err := s.repo.Brand(ctx, viewer.TenantID, *payload.BrandID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return row, ErrNotFound
			}
			return row, err
		}
		row.BrandID = payload.BrandID
	}
	if payload.ProductID != nil && *payload.ProductID > 0 {
		if _, err := s.repo.Product(ctx, viewer.TenantID, *payload.ProductID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return row, ErrNotFound
			}
			return row, err
		}
		row.ProductID = payload.ProductID
	}
	if strings.TrimSpace(payload.AssetType) != "" {
		row.AssetType = strings.TrimSpace(payload.AssetType)
	}
	if strings.TrimSpace(payload.AssetName) != "" {
		row.AssetName = strings.TrimSpace(payload.AssetName)
	}
	row.URL = stringPtr(payload.URL)
	if payload.Metadata != nil {
		row.Metadata = jsonString(payload.Metadata, map[string]interface{}{})
	}
	if strings.TrimSpace(payload.Status) != "" {
		row.Status = strings.TrimSpace(payload.Status)
	}
	now := time.Now()
	userID := viewer.UserID
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveMaterialAsset(ctx, &row)
	return row, err
}

func (s *Service) ArchiveMaterialAsset(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoMaterialAsset, error) {
	row, err := s.repo.MaterialAsset(ctx, viewer.TenantID, id)
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
	err = s.repo.SaveMaterialAsset(ctx, &row)
	return row, err
}

func (s *Service) Hotspots(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoHotspot], error) {
	rows, total, err := s.repo.ListHotspots(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) Hotspot(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoHotspot, error) {
	row, err := s.repo.Hotspot(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) CreateHotspot(ctx context.Context, viewer dto.Viewer, payload dto.HotspotPayload) (models.AiGeoHotspot, error) {
	platform := strings.TrimSpace(payload.Platform)
	title := strings.TrimSpace(payload.Title)
	if viewer.TenantID == 0 || platform == "" || title == "" {
		return models.AiGeoHotspot{}, ErrInvalidInput
	}
	capturedAt := time.Now()
	if payload.CapturedAt != "" {
		parsed, err := time.Parse(time.RFC3339, payload.CapturedAt)
		if err != nil {
			return models.AiGeoHotspot{}, ErrInvalidInput
		}
		capturedAt = parsed
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoHotspot{
		TenantID:   viewer.TenantID,
		Platform:   platform,
		Title:      title,
		HeatScore:  payload.HeatScore,
		SourceURL:  stringPtr(payload.SourceURL),
		CapturedAt: capturedAt,
		Status:     defaultString(payload.Status, "active"),
		CreatedBy:  &userID,
		UpdatedBy:  &userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	err := s.repo.SaveHotspot(ctx, &row)
	return row, err
}

func (s *Service) UpdateHotspot(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.HotspotPayload) (models.AiGeoHotspot, error) {
	row, err := s.repo.Hotspot(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if strings.TrimSpace(payload.Platform) != "" {
		row.Platform = strings.TrimSpace(payload.Platform)
	}
	if strings.TrimSpace(payload.Title) != "" {
		row.Title = strings.TrimSpace(payload.Title)
	}
	if payload.HeatScore > 0 {
		row.HeatScore = payload.HeatScore
	}
	row.SourceURL = stringPtr(payload.SourceURL)
	if payload.CapturedAt != "" {
		parsed, err := time.Parse(time.RFC3339, payload.CapturedAt)
		if err != nil {
			return row, ErrInvalidInput
		}
		row.CapturedAt = parsed
	}
	if strings.TrimSpace(payload.Status) != "" {
		row.Status = strings.TrimSpace(payload.Status)
	}
	now := time.Now()
	userID := viewer.UserID
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveHotspot(ctx, &row)
	return row, err
}

func (s *Service) ArchiveHotspot(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoHotspot, error) {
	row, err := s.repo.Hotspot(ctx, viewer.TenantID, id)
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
	err = s.repo.SaveHotspot(ctx, &row)
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

func (s *Service) Draft(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoDraft, error) {
	row, err := s.repo.Draft(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) CreateDraft(ctx context.Context, viewer dto.Viewer, payload dto.DraftPayload) (models.AiGeoDraft, error) {
	if viewer.TenantID == 0 || strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Body) == "" {
		return models.AiGeoDraft{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoDraft{
		TenantID:       viewer.TenantID,
		DraftCode:      code("DRAFT", now),
		BrandID:        payload.BrandID,
		ProductID:      payload.ProductID,
		Title:          strings.TrimSpace(payload.Title),
		Summary:        stringPtr(payload.Summary),
		Body:           strings.TrimSpace(payload.Body),
		Keywords:       jsonString(payload.Keywords, []string{}),
		Conversation:   jsonString(sanitizeDraftConversation(payload.Conversation), []dto.DraftConversationMessage{}),
		SourceSnapshot: jsonString(payload.SourceSnapshot, map[string]interface{}{}),
		Source:         defaultString(payload.Source, "manual"),
		AuditStatus:    "draft",
		ChannelStatus:  "not_generated",
		Status:         "active",
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	err := s.repo.SaveDraft(ctx, &row)
	return row, err
}

func (s *Service) UpdateDraft(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.DraftPayload) (models.AiGeoDraft, error) {
	if viewer.TenantID == 0 || id == 0 || strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Body) == "" {
		return models.AiGeoDraft{}, ErrInvalidInput
	}
	row, err := s.repo.Draft(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.AuditStatus == "approved" {
		return row, ErrInvalidStatus
	}
	userID := viewer.UserID
	row.BrandID = payload.BrandID
	row.ProductID = payload.ProductID
	row.Title = strings.TrimSpace(payload.Title)
	row.Summary = stringPtr(payload.Summary)
	row.Body = strings.TrimSpace(payload.Body)
	row.Keywords = jsonString(payload.Keywords, []string{})
	row.Conversation = jsonString(sanitizeDraftConversation(payload.Conversation), []dto.DraftConversationMessage{})
	row.SourceSnapshot = jsonString(payload.SourceSnapshot, map[string]interface{}{})
	row.Source = defaultString(payload.Source, row.Source)
	row.UpdatedBy = &userID
	row.UpdatedAt = time.Now()
	if row.AuditStatus != "draft" {
		row.AuditStatus = "draft"
	}
	err = s.repo.SaveDraft(ctx, &row)
	return row, err
}

func (s *Service) ArchiveDraft(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoDraft, error) {
	row, err := s.repo.Draft(ctx, viewer.TenantID, id)
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
	err = s.repo.SaveDraft(ctx, &row)
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
	var skus []models.AiGeoSKU
	if product != nil {
		rows, _, err := s.repo.ListSKUs(ctx, viewer.TenantID, dto.PageRequest{ProductID: product.ID, Status: "active", Limit: 20})
		if err != nil {
			return models.AiGeoDraft{}, err
		}
		skus = rows
	}
	var hotspot *models.AiGeoHotspot
	if payload.HotspotID != nil && *payload.HotspotID > 0 {
		row, err := s.repo.Hotspot(ctx, viewer.TenantID, *payload.HotspotID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AiGeoDraft{}, ErrNotFound
		}
		if err != nil {
			return models.AiGeoDraft{}, err
		}
		hotspot = &row
	}
	if err := s.consumeQuota(ctx, viewer.TenantID, quotaMonthlyDraftGenerations); err != nil {
		return models.AiGeoDraft{}, err
	}
	generator := s.draftGenerator
	if generator == nil {
		generator = localDraftGenerator{}
	}
	generated, err := generator.GenerateDraft(ctx, DraftGenerationRequest{Viewer: viewer, Payload: payload, Brand: brand, Product: product, SKUs: skus, Hotspot: hotspot})
	if err != nil {
		return models.AiGeoDraft{}, err
	}
	return s.CreateDraft(ctx, viewer, dto.DraftPayload{
		BrandID:        payload.BrandID,
		ProductID:      payload.ProductID,
		Title:          generated.Title,
		Summary:        generated.Summary,
		Body:           generated.Body,
		Keywords:       generated.Keywords,
		Conversation:   payload.Conversation,
		SourceSnapshot: draftSourceSnapshot(payload, brand, product, skus, hotspot),
		Source:         defaultString(generated.Source, "ai_workbench"),
	})
}

func draftSourceSnapshot(payload dto.GenerateDraftPayload, brand *models.AiGeoBrandCard, product *models.AiGeoProductCard, skus []models.AiGeoSKU, hotspot *models.AiGeoHotspot) map[string]interface{} {
	snapshot := make(map[string]interface{})
	for key, value := range payload.SourceSnapshot {
		snapshot[key] = value
	}
	if skill := strings.TrimSpace(payload.Skill); skill != "" {
		snapshot["skill"] = skill
	}
	if prompt := strings.TrimSpace(payload.Prompt); prompt != "" {
		snapshot["prompt"] = prompt
	}
	if brand != nil {
		snapshot["brand"] = map[string]interface{}{
			"id":              brand.ID,
			"code":            brand.BrandCode,
			"name":            brand.BrandName,
			"positioning":     stringValueFromPtr(brand.Positioning),
			"target_audience": stringValueFromPtr(brand.TargetAudience),
			"price_band":      stringValueFromPtr(brand.PriceBand),
			"tone":            stringValueFromPtr(brand.Tone),
			"keywords":        brand.Keywords,
			"completeness":    brand.Completeness,
		}
	}
	if product != nil {
		snapshot["product"] = map[string]interface{}{
			"id":             product.ID,
			"code":           product.ProductCode,
			"name":           product.ProductName,
			"category":       stringValueFromPtr(product.CategoryName),
			"selling_points": product.SellingPoints,
			"faq":            product.FAQ,
			"content_angles": product.ContentAngles,
			"completeness":   product.Completeness,
		}
	}
	if len(skus) > 0 {
		items := make([]map[string]interface{}, 0, len(skus))
		for _, sku := range skus {
			items = append(items, map[string]interface{}{
				"id":           sku.ID,
				"code":         sku.SKUCode,
				"name":         sku.SKUName,
				"attributes":   sku.Attributes,
				"price":        sku.Price,
				"image_url":    stringValueFromPtr(sku.ImageURL),
				"stock_status": sku.StockStatus,
			})
		}
		snapshot["skus"] = items
	}
	if hotspot != nil {
		snapshot["hotspot"] = map[string]interface{}{
			"id":         hotspot.ID,
			"platform":   hotspot.Platform,
			"title":      hotspot.Title,
			"heat_score": hotspot.HeatScore,
			"source_url": stringValueFromPtr(hotspot.SourceURL),
		}
	}
	return snapshot
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

func (s *Service) ReviewDraft(ctx context.Context, viewer dto.Viewer, id uint64, approved bool, opinion string) (models.AiGeoDraft, error) {
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
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		review := draftReviewSuggestionRow(viewer, row, approved, opinion)
		return tx.Create(&review).Error
	})
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
	if draft.AuditStatus != "approved" {
		return models.AiGeoChannelContent{}, ErrInvalidStatus
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

func (s *Service) ChannelContents(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoChannelContent], error) {
	rows, total, err := s.repo.ListChannelContents(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) ChannelContent(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoChannelContent, error) {
	row, err := s.repo.ChannelContent(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) UpdateChannelContent(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.ChannelContentPayload) (models.AiGeoChannelContent, error) {
	row, err := s.repo.ChannelContent(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.PublishStatus == "publishing" || row.PublishStatus == "published" {
		return row, ErrInvalidStatus
	}
	if strings.TrimSpace(payload.Title) != "" {
		row.Title = strings.TrimSpace(payload.Title)
	}
	if strings.TrimSpace(payload.Body) != "" {
		row.Body = strings.TrimSpace(payload.Body)
	}
	if payload.ChannelID > 0 && payload.ChannelID != row.ChannelID {
		if _, err := s.repo.Channel(ctx, viewer.TenantID, payload.ChannelID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return row, ErrNotFound
			}
			return row, err
		}
		row.ChannelID = payload.ChannelID
	}
	row.AuditStatus = "pending"
	row.UpdatedAt = time.Now()
	userID := viewer.UserID
	row.UpdatedBy = &userID
	err = s.repo.SaveChannelContent(ctx, &row)
	return row, err
}

func (s *Service) ReviewChannelContent(ctx context.Context, viewer dto.Viewer, id uint64, approved bool, opinion string) (models.AiGeoChannelContent, error) {
	row, err := s.repo.ChannelContent(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.AuditStatus != "pending" && row.AuditStatus != "rejected" {
		return row, ErrInvalidStatus
	}
	if approved {
		row.AuditStatus = "approved"
	} else {
		row.AuditStatus = "rejected"
	}
	now := time.Now()
	userID := viewer.UserID
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		review := channelContentReviewSuggestionRow(viewer, row, approved, opinion)
		return tx.Create(&review).Error
	})
	return row, err
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

func (s *Service) PublishPlanCalendar(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PublishPlanCalendar, error) {
	if viewer.TenantID == 0 {
		return dto.PublishPlanCalendar{}, ErrInvalidInput
	}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if req.StartDate != nil {
		start = dateOnly(*req.StartDate)
	}
	end := start.AddDate(0, 1, -1)
	if req.EndDate != nil {
		end = dateOnly(*req.EndDate)
	}
	if end.Before(start) {
		return dto.PublishPlanCalendar{}, ErrInvalidInput
	}
	if end.Sub(start).Hours()/24 > 370 {
		return dto.PublishPlanCalendar{}, ErrInvalidInput
	}
	rows, err := s.repo.ListPublishPlansBySchedule(ctx, viewer.TenantID, start, end.AddDate(0, 0, 1))
	if err != nil {
		return dto.PublishPlanCalendar{}, err
	}
	byDate := map[string][]models.AiGeoPublishPlan{}
	for _, row := range rows {
		key := row.ScheduledAt.In(start.Location()).Format("2006-01-02")
		byDate[key] = append(byDate[key], row)
	}
	days := make([]dto.PublishPlanCalendarDay, 0, int(end.Sub(start).Hours()/24)+1)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		calendarDay := dto.PublishPlanCalendarDay{Date: key, Items: byDate[key]}
		for _, item := range calendarDay.Items {
			calendarDay.Total++
			switch item.Status {
			case "scheduled":
				calendarDay.Scheduled++
			case "publishing":
				calendarDay.Publishing++
			case "published":
				calendarDay.Published++
			case "failed":
				calendarDay.Failed++
			case "cancelled":
				calendarDay.Cancelled++
			}
		}
		days = append(days, calendarDay)
	}
	return dto.PublishPlanCalendar{StartDate: start.Format("2006-01-02"), EndDate: end.Format("2006-01-02"), Days: days}, nil
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

func (s *Service) UpdatePublishPlan(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.PublishPlanPayload) (models.AiGeoPublishPlan, error) {
	row, err := s.repo.PublishPlan(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.Status == "publishing" || row.Status == "published" {
		return row, ErrInvalidStatus
	}
	if payload.ChannelContentID > 0 || payload.ChannelID > 0 {
		nextContentID := row.ChannelContentID
		nextChannelID := row.ChannelID
		if payload.ChannelContentID > 0 {
			nextContentID = payload.ChannelContentID
		}
		if payload.ChannelID > 0 {
			nextChannelID = payload.ChannelID
		}
		content, err := s.repo.ChannelContent(ctx, viewer.TenantID, nextContentID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return row, ErrNotFound
		}
		if err != nil {
			return row, err
		}
		if content.ChannelID != nextChannelID {
			return row, ErrInvalidInput
		}
		if _, err := s.repo.Channel(ctx, viewer.TenantID, nextChannelID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return row, ErrNotFound
			}
			return row, err
		}
		row.ChannelContentID = nextContentID
		row.ChannelID = nextChannelID
	}
	if payload.ScheduledAt != "" {
		scheduledAt, err := time.Parse(time.RFC3339, payload.ScheduledAt)
		if err != nil {
			return row, ErrInvalidInput
		}
		row.ScheduledAt = scheduledAt
	}
	if strings.TrimSpace(payload.PublishMethod) != "" {
		row.PublishMethod = strings.TrimSpace(payload.PublishMethod)
	}
	if strings.TrimSpace(payload.AutomationLevel) != "" {
		row.AutomationLevel = strings.TrimSpace(payload.AutomationLevel)
	}
	now := time.Now()
	userID := viewer.UserID
	row.UpdatedAt = now
	row.UpdatedBy = &userID
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
	importType := normalizeImportType(payload.ImportType)
	if viewer.TenantID == 0 || !supportedImportType(importType) {
		return models.AiGeoImportBatch{}, ErrInvalidInput
	}
	now := time.Now()
	userID := viewer.UserID
	importErrors := validateImportRecords(viewer.TenantID, payload.Records, importType, now)
	failedRows := map[int]bool{}
	for _, row := range importErrors {
		failedRows[row.RowNumber] = true
	}
	if err := s.validateMaterialImportQuotas(ctx, viewer, importType, payload.Records, failedRows); err != nil {
		return models.AiGeoImportBatch{}, err
	}
	batch := models.AiGeoImportBatch{
		TenantID:      viewer.TenantID,
		BatchCode:     code("IMPORT", now),
		ImportType:    importType,
		MappingConfig: jsonString(payload.MappingConfig, map[string]interface{}{}),
		RecordCount:   int64(len(payload.Records)),
		Status:        "pending",
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&batch).Error; err != nil {
			return err
		}
		for index, record := range payload.Records {
			rowNumber := index + 1
			if failedRows[rowNumber] {
				continue
			}
			rowError, err := s.applyMaterialImportRow(ctx, tx, viewer, importType, record, rowNumber, now)
			if err != nil {
				return err
			}
			if rowError != nil {
				failedRows[rowNumber] = true
				importErrors = append(importErrors, *rowError)
			}
		}
		if len(importErrors) > 0 {
			for i := range importErrors {
				importErrors[i].BatchID = batch.ID
			}
			if err := tx.Create(&importErrors).Error; err != nil {
				return err
			}
		}
		batch.FailedCount = int64(len(failedRows))
		batch.SuccessCount = batch.RecordCount - batch.FailedCount
		batch.Status = importBatchStatus(batch.SuccessCount, batch.FailedCount)
		batch.UpdatedAt = time.Now()
		if err := tx.Save(&batch).Error; err != nil {
			return err
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

func (s *Service) applyMaterialImportRow(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, importType string, record map[string]interface{}, rowNumber int, now time.Time) (*models.AiGeoImportError, error) {
	switch importType {
	case "brand":
		return s.importBrandRow(ctx, tx, viewer, record, rowNumber, now)
	case "product":
		return s.importProductRow(ctx, tx, viewer, record, rowNumber, now)
	case "sku":
		return s.importSKURow(ctx, tx, viewer, record, rowNumber, now)
	case "competitor":
		return s.importCompetitorRow(ctx, tx, viewer, record, rowNumber, now)
	default:
		row := importErrorRow(viewer.TenantID, rowNumber, "import_type", "unsupported_type", "不支持的导入类型", record, now)
		return &row, nil
	}
}

func (s *Service) importBrandRow(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, record map[string]interface{}, rowNumber int, now time.Time) (*models.AiGeoImportError, error) {
	brandCode := importString(record, "brand_code")
	userID := viewer.UserID
	var row models.AiGeoBrandCard
	err := notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoBrandCard{}), viewer.TenantID).
		Where("brand_code = ?", brandCode).
		First(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.AiGeoBrandCard{
			TenantID:  viewer.TenantID,
			BrandCode: brandCode,
			CreatedBy: &userID,
			CreatedAt: now,
		}
	}
	row.BrandName = importString(record, "brand_name")
	row.Positioning = stringPtr(importString(record, "positioning"))
	row.TargetAudience = stringPtr(importString(record, "target_audience", "audience"))
	row.PriceBand = stringPtr(importString(record, "price_band"))
	row.Tone = stringPtr(importString(record, "tone"))
	row.Keywords = jsonString(importStringList(record, "keywords"), []string{})
	row.Completeness = completeness(row.BrandName, importString(record, "positioning"), importString(record, "target_audience", "audience"), importString(record, "price_band"), importString(record, "tone"))
	row.Status = defaultString(importString(record, "status"), "active")
	row.UpdatedBy = &userID
	row.UpdatedAt = now
	return nil, tx.Save(&row).Error
}

func (s *Service) importProductRow(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, record map[string]interface{}, rowNumber int, now time.Time) (*models.AiGeoImportError, error) {
	brandID, err := resolveImportBrandID(ctx, tx, viewer.TenantID, record)
	if err != nil {
		row := importErrorRow(viewer.TenantID, rowNumber, "brand_code", "not_found", "品牌不存在，请先导入品牌或提供 brand_id", record, now)
		return &row, nil
	}
	productCode := importString(record, "product_code")
	userID := viewer.UserID
	var row models.AiGeoProductCard
	err = notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoProductCard{}), viewer.TenantID).
		Where("product_code = ?", productCode).
		First(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.AiGeoProductCard{
			TenantID:    viewer.TenantID,
			ProductCode: productCode,
			CreatedBy:   &userID,
			CreatedAt:   now,
		}
	}
	sellingPoints := importStringList(record, "selling_points", "selling_point")
	faq := importStringList(record, "faq")
	row.BrandID = brandID
	row.ProductName = importString(record, "product_name")
	row.CategoryName = stringPtr(importString(record, "category_name", "category"))
	row.SellingPoints = jsonString(sellingPoints, []string{})
	row.FAQ = jsonString(faq, []string{})
	row.ContentAngles = jsonString(importStringList(record, "content_angles", "angles"), []string{})
	row.Completeness = completeness(row.ProductName, importString(record, "category_name", "category"), strings.Join(sellingPoints, ","), strings.Join(faq, ","))
	row.Status = defaultString(importString(record, "status"), "active")
	row.UpdatedBy = &userID
	row.UpdatedAt = now
	return nil, tx.Save(&row).Error
}

func (s *Service) importSKURow(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, record map[string]interface{}, rowNumber int, now time.Time) (*models.AiGeoImportError, error) {
	productID, err := resolveImportProductID(ctx, tx, viewer.TenantID, record)
	if err != nil {
		row := importErrorRow(viewer.TenantID, rowNumber, "product_code", "not_found", "商品不存在，请先导入商品或提供 product_id", record, now)
		return &row, nil
	}
	skuCode := importString(record, "sku_code")
	userID := viewer.UserID
	var row models.AiGeoSKU
	err = notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoSKU{}), viewer.TenantID).
		Where("sku_code = ?", skuCode).
		First(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.AiGeoSKU{
			TenantID:  viewer.TenantID,
			SKUCode:   skuCode,
			CreatedBy: &userID,
			CreatedAt: now,
		}
	}
	row.ProductID = productID
	row.SKUName = importString(record, "sku_name")
	row.Attributes = jsonString(importMap(record, "attributes"), map[string]interface{}{})
	row.Price = importFloat(record, "price")
	row.ImageURL = stringPtr(importString(record, "image_url"))
	row.StockStatus = defaultString(importString(record, "stock_status"), "unknown")
	row.Status = defaultString(importString(record, "status"), "active")
	row.UpdatedBy = &userID
	row.UpdatedAt = now
	return nil, tx.Save(&row).Error
}

func (s *Service) importCompetitorRow(ctx context.Context, tx *gorm.DB, viewer dto.Viewer, record map[string]interface{}, rowNumber int, now time.Time) (*models.AiGeoImportError, error) {
	productID, err := resolveImportProductID(ctx, tx, viewer.TenantID, record)
	if err != nil {
		row := importErrorRow(viewer.TenantID, rowNumber, "product_code", "not_found", "商品不存在，请先导入商品或提供 product_id", record, now)
		return &row, nil
	}
	userID := viewer.UserID
	brandName := importString(record, "brand_name", "competitor_brand_name")
	productName := importString(record, "product_name", "competitor_product_name")
	var row models.AiGeoCompetitor
	err = notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoCompetitor{}), viewer.TenantID).
		Where("product_id = ? AND brand_name = ? AND product_name = ?", productID, brandName, productName).
		First(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.AiGeoCompetitor{
			TenantID:    viewer.TenantID,
			ProductID:   productID,
			BrandName:   brandName,
			ProductName: productName,
			CreatedBy:   &userID,
			CreatedAt:   now,
		}
	}
	row.ProductID = productID
	row.BrandName = brandName
	row.ProductName = productName
	row.PriceText = stringPtr(importString(record, "price_text"))
	row.Point = stringPtr(importString(record, "point"))
	row.Difference = stringPtr(importString(record, "difference"))
	row.Angle = stringPtr(importString(record, "angle"))
	row.LinkURL = stringPtr(importString(record, "link_url"))
	row.Status = defaultString(importString(record, "status"), "active")
	row.UpdatedBy = &userID
	row.UpdatedAt = now
	return nil, tx.Save(&row).Error
}

func (s *Service) validateMaterialImportQuotas(ctx context.Context, viewer dto.Viewer, importType string, records []map[string]interface{}, failedRows map[int]bool) error {
	switch importType {
	case "brand":
		return s.validateImportStaticQuota(ctx, viewer.TenantID, quotaBrandCount, s.repo.CountBrands, records, failedRows, "brand_code", s.brandImportExists)
	case "product":
		return s.validateImportStaticQuota(ctx, viewer.TenantID, quotaProductCount, s.repo.CountProducts, records, failedRows, "product_code", s.productImportExists)
	default:
		return nil
	}
}

func (s *Service) validateImportStaticQuota(ctx context.Context, tenantID uint64, quotaCode string, currentCount func(context.Context, uint64) (int64, error), records []map[string]interface{}, failedRows map[int]bool, codeField string, exists func(context.Context, uint64, string) (bool, error)) error {
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
	newCodes := map[string]bool{}
	for index, record := range records {
		rowNumber := index + 1
		if failedRows[rowNumber] {
			continue
		}
		codeValue := importString(record, codeField)
		if codeValue == "" {
			continue
		}
		existing, err := exists(ctx, tenantID, codeValue)
		if err != nil {
			return err
		}
		if !existing {
			newCodes[strings.ToLower(codeValue)] = true
		}
	}
	if int(count)+len(newCodes) > limit {
		return &quotaapp.ExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: int(count)}
	}
	return nil
}

func (s *Service) brandImportExists(ctx context.Context, tenantID uint64, code string) (bool, error) {
	_, err := s.repo.BrandByCode(ctx, tenantID, code)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (s *Service) productImportExists(ctx context.Context, tenantID uint64, code string) (bool, error) {
	_, err := s.repo.ProductByCode(ctx, tenantID, code)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
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

func (s *Service) ensureStaticQuotaWithDB(ctx context.Context, db *gorm.DB, tenantID uint64, quotaCode string, model interface{}) error {
	if s.quota == nil {
		return nil
	}
	limit, quota, ok, err := s.quota.CurrentLimitByCode(ctx, tenantID, quotaCode)
	if err != nil || !ok || limit < 0 {
		return err
	}
	var count int64
	if err := notDeletedTenant(db.WithContext(ctx).Model(model), tenantID).Where("status = ?", "active").Count(&count).Error; err != nil {
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
	importType = normalizeImportType(importType)
	required := []string{}
	uniqueField := ""
	switch importType {
	case "brand":
		required = []string{"brand_code", "brand_name"}
		uniqueField = "brand_code"
	case "product":
		required = []string{"product_code", "product_name"}
		uniqueField = "product_code"
	case "sku":
		required = []string{"sku_code", "sku_name"}
		uniqueField = "sku_code"
	case "competitor":
		required = []string{"brand_name", "product_name"}
	default:
		return []models.AiGeoImportError{importErrorRow(tenantID, 0, "import_type", "unsupported_type", "不支持的导入类型", map[string]interface{}{"import_type": importType}, now)}
	}
	errorsOut := []models.AiGeoImportError{}
	seen := map[string]int{}
	for index, record := range records {
		rowNumber := index + 1
		if len(record) == 0 {
			errorsOut = append(errorsOut, importErrorRow(tenantID, rowNumber, "", "empty_row", "导入行为空", record, now))
			continue
		}
		for _, field := range required {
			if importString(record, field) == "" {
				errorsOut = append(errorsOut, importErrorRow(tenantID, rowNumber, field, "required", "必填字段缺失", record, now))
			}
		}
		if uniqueField != "" {
			value := importString(record, uniqueField)
			if value != "" {
				key := strings.ToLower(value)
				if firstRow, ok := seen[key]; ok {
					errorsOut = append(errorsOut, importErrorRow(tenantID, rowNumber, uniqueField, "duplicate_in_batch", fmt.Sprintf("导入批次内重复，首次出现在第 %d 行", firstRow), record, now))
				} else {
					seen[key] = rowNumber
				}
			}
		}
	}
	return errorsOut
}

func normalizeImportType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "brand", "brands":
		return "brand"
	case "product", "products":
		return "product"
	case "sku", "skus":
		return "sku"
	case "competitor", "competitors":
		return "competitor"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func supportedImportType(value string) bool {
	switch value {
	case "brand", "product", "sku", "competitor":
		return true
	default:
		return false
	}
}

func importBatchStatus(successCount int64, failedCount int64) string {
	if failedCount > 0 && successCount > 0 {
		return "partial_success"
	}
	if failedCount > 0 {
		return "failed"
	}
	return "completed"
}

func notDeletedTenant(db *gorm.DB, tenantID uint64) *gorm.DB {
	return db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
}

func resolveImportBrandID(ctx context.Context, tx *gorm.DB, tenantID uint64, record map[string]interface{}) (uint64, error) {
	if id := importUint(record, "brand_id"); id > 0 {
		var brand models.AiGeoBrandCard
		err := notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoBrandCard{}), tenantID).Where("id = ?", id).First(&brand).Error
		if err != nil {
			return 0, err
		}
		return brand.ID, nil
	}
	brandCode := importString(record, "brand_code")
	if brandCode == "" {
		return 0, gorm.ErrRecordNotFound
	}
	var brand models.AiGeoBrandCard
	err := notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoBrandCard{}), tenantID).Where("brand_code = ?", brandCode).First(&brand).Error
	if err != nil {
		return 0, err
	}
	return brand.ID, nil
}

func resolveImportProductID(ctx context.Context, tx *gorm.DB, tenantID uint64, record map[string]interface{}) (uint64, error) {
	if id := importUint(record, "product_id"); id > 0 {
		var product models.AiGeoProductCard
		err := notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoProductCard{}), tenantID).Where("id = ?", id).First(&product).Error
		if err != nil {
			return 0, err
		}
		return product.ID, nil
	}
	productCode := importString(record, "product_code")
	if productCode == "" {
		return 0, gorm.ErrRecordNotFound
	}
	var product models.AiGeoProductCard
	err := notDeletedTenant(tx.WithContext(ctx).Model(&models.AiGeoProductCard{}), tenantID).Where("product_code = ?", productCode).First(&product).Error
	if err != nil {
		return 0, err
	}
	return product.ID, nil
}

func importString(record map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := record[key]
		if !ok || value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func importStringList(record map[string]interface{}, keys ...string) []string {
	for _, key := range keys {
		value, ok := record[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case []string:
			return compactStrings(typed)
		case []interface{}:
			out := make([]string, 0, len(typed))
			for _, item := range typed {
				if text := strings.TrimSpace(fmt.Sprint(item)); text != "" && text != "<nil>" {
					out = append(out, text)
				}
			}
			return out
		case string:
			text := strings.TrimSpace(typed)
			if text == "" {
				continue
			}
			var parsed []string
			if strings.HasPrefix(text, "[") && json.Unmarshal([]byte(text), &parsed) == nil {
				return compactStrings(parsed)
			}
			parts := strings.FieldsFunc(text, func(r rune) bool {
				return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n'
			})
			return compactStrings(parts)
		default:
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" && text != "<nil>" {
				return []string{text}
			}
		}
	}
	return []string{}
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func importMap(record map[string]interface{}, keys ...string) map[string]interface{} {
	for _, key := range keys {
		value, ok := record[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case map[string]interface{}:
			return typed
		case string:
			var parsed map[string]interface{}
			if json.Unmarshal([]byte(strings.TrimSpace(typed)), &parsed) == nil {
				return parsed
			}
		}
	}
	return map[string]interface{}{}
}

func importUint(record map[string]interface{}, keys ...string) uint64 {
	for _, key := range keys {
		value, ok := record[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case uint64:
			return typed
		case uint:
			return uint64(typed)
		case int:
			if typed > 0 {
				return uint64(typed)
			}
		case int64:
			if typed > 0 {
				return uint64(typed)
			}
		case float64:
			if typed > 0 {
				return uint64(typed)
			}
		case string:
			parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
			if err == nil {
				return parsed
			}
		}
	}
	return 0
}

func importFloat(record map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		value, ok := record[key]
		if !ok || value == nil {
			continue
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
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
			if err == nil {
				return parsed
			}
		}
	}
	return 0
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

func draftReviewSuggestionRow(viewer dto.Viewer, draft models.AiGeoDraft, approved bool, opinion string) models.AiGeoAuditSuggestion {
	now := time.Now()
	opinion = strings.TrimSpace(opinion)
	if opinion == "" {
		if approved {
			opinion = "人工确认通过，未填写额外意见。"
		} else {
			opinion = "人工确认驳回，未填写额外意见。"
		}
	}
	riskLevel := "low"
	if !approved {
		riskLevel = "high"
	}
	return models.AiGeoAuditSuggestion{
		TenantID:     viewer.TenantID,
		ObjectType:   "draft",
		ObjectID:     draft.ID,
		ObjectCode:   stringPtr(draft.DraftCode),
		ScenarioCode: "manual_draft_review",
		RiskLevel:    riskLevel,
		Passed:       approved,
		Summary:      stringPtr(opinion),
		SuggestionJSON: jsonString([]map[string]interface{}{
			{"type": "human_review", "content": opinion},
		}, []map[string]interface{}{}),
		ModelCode:   stringPtr("human-review"),
		Status:      "success",
		GeneratedAt: now,
		CreatedBy:   &viewer.UserID,
		CreatedAt:   now,
	}
}

func channelContentReviewSuggestionRow(viewer dto.Viewer, content models.AiGeoChannelContent, approved bool, opinion string) models.AiGeoAuditSuggestion {
	now := time.Now()
	opinion = strings.TrimSpace(opinion)
	if opinion == "" {
		if approved {
			opinion = "人工确认渠道内容通过，未填写额外意见。"
		} else {
			opinion = "人工确认渠道内容驳回，未填写额外意见。"
		}
	}
	riskLevel := "low"
	if !approved {
		riskLevel = "high"
	}
	return models.AiGeoAuditSuggestion{
		TenantID:     viewer.TenantID,
		ObjectType:   "channel_content",
		ObjectID:     content.ID,
		ObjectCode:   stringPtr(fmt.Sprintf("%d", content.ID)),
		ScenarioCode: "manual_channel_content_review",
		RiskLevel:    riskLevel,
		Passed:       approved,
		Summary:      stringPtr(opinion),
		SuggestionJSON: jsonString([]map[string]interface{}{
			{"type": "human_review", "content": opinion},
		}, []map[string]interface{}{}),
		ModelCode:   stringPtr("human-review"),
		Status:      "success",
		GeneratedAt: now,
		CreatedBy:   &viewer.UserID,
		CreatedAt:   now,
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

func sanitizeDraftConversation(items []dto.DraftConversationMessage) []dto.DraftConversationMessage {
	cleaned := make([]dto.DraftConversationMessage, 0, len(items))
	for _, item := range items {
		role := strings.TrimSpace(item.Role)
		text := strings.TrimSpace(item.Text)
		if role == "" || text == "" {
			continue
		}
		if role != "user" && role != "ai" && role != "assistant" {
			role = "assistant"
		}
		if role == "assistant" {
			role = "ai"
		}
		cleaned = append(cleaned, dto.DraftConversationMessage{
			Role:      role,
			Text:      text,
			CreatedAt: strings.TrimSpace(item.CreatedAt),
		})
	}
	return cleaned
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func positiveUint64Ptr(value *uint64) *uint64 {
	if value == nil || *value == 0 {
		return nil
	}
	return value
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
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
