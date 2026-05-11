package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"saas_baseon_go/internal/apps/app_center/dto"
	"saas_baseon_go/internal/apps/app_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"gorm.io/gorm"
)

var ErrPlatformOnly = errors.New("仅平台管理员可访问应用中心")
var ErrAppNotFound = errors.New("应用不存在")
var ErrAppCodeRequired = errors.New("应用编码不能为空")
var ErrAppNameRequired = errors.New("应用名称不能为空")
var ErrInvalidAppCode = errors.New("应用编码仅允许小写字母、数字和中横线，且必须以字母开头")
var ErrAppCodeExists = errors.New("应用编码已存在")

var appCodePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,63}$`)

type AppService struct {
	repo *repositories.AppRepository
}

func NewAppService(repo *repositories.AppRepository) *AppService {
	return &AppService{repo: repo}
}

func (s *AppService) ListApps(ctx context.Context, viewerID uint64, req dto.AppListRequest) (dto.AppListResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.AppListResponse{}, err
	}
	req.Keyword = strings.TrimSpace(req.Keyword)
	req.Type = strings.TrimSpace(req.Type)
	req.Status = strings.TrimSpace(req.Status)
	req.Source = strings.TrimSpace(req.Source)
	if req.Skip < 0 {
		req.Skip = 0
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	rows, total, err := s.repo.List(ctx, req)
	if err != nil {
		return dto.AppListResponse{}, err
	}
	items := make([]dto.AppResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, appToResponse(row))
	}
	return dto.AppListResponse{Items: items, Total: total, Skip: req.Skip, Limit: req.Limit}, nil
}

func (s *AppService) Stats(ctx context.Context, viewerID uint64) (dto.AppStatsResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.AppStatsResponse{}, err
	}
	return s.repo.Stats(ctx)
}

func (s *AppService) GetApp(ctx context.Context, viewerID uint64, id uint64) (dto.AppResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.AppResponse{}, err
	}
	if id == 0 {
		return dto.AppResponse{}, ErrAppNotFound
	}
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AppResponse{}, ErrAppNotFound
		}
		return dto.AppResponse{}, err
	}
	return appToResponse(row), nil
}

func (s *AppService) CreateApp(ctx context.Context, viewerID uint64, req dto.AppCreateRequest) (dto.AppResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.AppResponse{}, err
	}
	appCode := strings.TrimSpace(req.AppCode)
	appName := strings.TrimSpace(req.AppName)
	if appCode == "" {
		return dto.AppResponse{}, ErrAppCodeRequired
	}
	if !appCodePattern.MatchString(appCode) {
		return dto.AppResponse{}, ErrInvalidAppCode
	}
	if appName == "" {
		return dto.AppResponse{}, ErrAppNameRequired
	}
	if _, err := s.repo.GetByCode(ctx, appCode); err == nil {
		return dto.AppResponse{}, ErrAppCodeExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AppResponse{}, err
	}
	row := models.SysApp{
		AppCode:         appCode,
		AppName:         appName,
		Icon:            cleanOptionalString(req.Icon),
		AppType:         defaultString(req.AppType, "BUSINESS_APP"),
		Source:          "MANUAL",
		Status:          defaultString(req.Status, "DRAFT"),
		ChargeMode:      defaultString(req.ChargeMode, "SUBSCRIPTION"),
		VisibilityScope: defaultString(req.VisibilityScope, "PLATFORM_ONLY"),
		Owner:           cleanOptionalString(req.Owner),
		Version:         cleanOptionalString(req.Version),
		Description:     cleanOptionalString(req.Description),
		IsBuiltin:       false,
		SortOrder:       req.SortOrder,
	}
	row.IsPlatformOnly = row.VisibilityScope == "PLATFORM_ONLY"
	if err := s.repo.Create(ctx, &row); err != nil {
		return dto.AppResponse{}, err
	}
	return appToResponse(row), nil
}

func (s *AppService) requirePlatformViewer(ctx context.Context, viewerID uint64) error {
	user, err := s.repo.UserByID(ctx, viewerID)
	if err != nil {
		return ErrPlatformOnly
	}
	if !user.IsPlatformAdmin {
		return ErrPlatformOnly
	}
	return nil
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func defaultString(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func appToResponse(row models.SysApp) dto.AppResponse {
	return dto.AppResponse{
		ID:              row.ID,
		AppCode:         row.AppCode,
		AppName:         row.AppName,
		Icon:            row.Icon,
		AppType:         row.AppType,
		Source:          row.Source,
		Status:          row.Status,
		ChargeMode:      row.ChargeMode,
		VisibilityScope: row.VisibilityScope,
		Owner:           row.Owner,
		Version:         row.Version,
		Description:     row.Description,
		IsBuiltin:       row.IsBuiltin,
		IsPlatformOnly:  row.IsPlatformOnly,
		SortOrder:       row.SortOrder,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
