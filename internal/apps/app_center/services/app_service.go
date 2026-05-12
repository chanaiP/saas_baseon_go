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
var ErrInvalidAppStatus = errors.New("应用状态不合法")
var ErrBuiltinStatusImmutable = errors.New("内置应用状态不允许在应用中心启停")
var ErrClientCodeRequired = errors.New("客户端编码不能为空")

var appCodePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,63}$`)
var allowedAppStatuses = map[string]struct{}{
	"INITIATED":  {},
	"PLANNED":    {},
	"DEVELOPING": {},
	"BETA":       {},
	"ONLINE":     {},
	"DISABLED":   {},
	"ARCHIVED":   {},
}

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
		items = append(items, appToResponse(row, nil))
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
	clients, err := s.repo.ClientsByAppID(ctx, row.ID)
	if err != nil {
		return dto.AppResponse{}, err
	}
	return appToResponse(row, clients), nil
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
		Status:          defaultString(req.Status, "INITIATED"),
		ChargeMode:      defaultString(req.ChargeMode, "SUBSCRIPTION"),
		VisibilityScope: defaultString(req.VisibilityScope, "PLATFORM_ONLY"),
		Owner:           cleanOptionalString(req.Owner),
		OwnerUserIDs:    cleanOptionalString(req.OwnerUserIDs),
		Version:         cleanOptionalString(req.Version),
		Description:     cleanOptionalString(req.Description),
		DetailDesc:      cleanOptionalString(req.DetailDesc),
		DeploymentMode:  defaultString(req.DeploymentMode, "MERGED"),
		CommModes:       cleanOptionalString(req.CommModes),
		VisibilityMode:  cleanOptionalString(req.VisibilityMode),
		VisibleTenants:  cleanOptionalString(req.VisibleTenants),
		OpenMethod:      cleanOptionalString(req.OpenMethod),
		TrialPolicy:     cleanOptionalString(req.TrialPolicy),
		TrialStartRule:  cleanOptionalString(req.TrialStartRule),
		AssetConfig:     cleanOptionalString(req.AssetConfig),
		DocConfig:       cleanOptionalString(req.DocConfig),
		ReleaseChannel:  cleanOptionalString(req.ReleaseChannel),
		ReleaseNote:     cleanOptionalString(req.ReleaseNote),
		IsBuiltin:       false,
		SortOrder:       req.SortOrder,
	}
	row.IsPlatformOnly = row.VisibilityScope == "PLATFORM_ONLY"
	clients, err := normalizeClientRequests(req.Clients)
	if err != nil {
		return dto.AppResponse{}, err
	}
	if err := s.repo.CreateWithClients(ctx, &row, clients); err != nil {
		return dto.AppResponse{}, err
	}
	return appToResponse(row, clients), nil
}

func (s *AppService) UpdateApp(ctx context.Context, viewerID uint64, id uint64, req dto.AppUpdateRequest) (dto.AppResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.AppResponse{}, err
	}
	if id == 0 {
		return dto.AppResponse{}, ErrAppNotFound
	}
	appName := strings.TrimSpace(req.AppName)
	if appName == "" {
		return dto.AppResponse{}, ErrAppNameRequired
	}
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AppResponse{}, ErrAppNotFound
		}
		return dto.AppResponse{}, err
	}
	visibilityScope := defaultString(req.VisibilityScope, row.VisibilityScope)
	updates := map[string]interface{}{
		"app_name":            appName,
		"icon":                cleanOptionalString(req.Icon),
		"app_type":            defaultString(req.AppType, row.AppType),
		"charge_mode":         defaultString(req.ChargeMode, row.ChargeMode),
		"visibility_scope":    visibilityScope,
		"owner":               cleanOptionalString(req.Owner),
		"owner_user_ids":      cleanOptionalString(req.OwnerUserIDs),
		"version":             cleanOptionalString(req.Version),
		"description":         cleanOptionalString(req.Description),
		"detail_description":  cleanOptionalString(req.DetailDesc),
		"deployment_mode":     defaultString(req.DeploymentMode, row.DeploymentMode),
		"communication_modes": cleanOptionalString(req.CommModes),
		"visibility_mode":     cleanOptionalString(req.VisibilityMode),
		"visible_tenants":     cleanOptionalString(req.VisibleTenants),
		"open_method":         cleanOptionalString(req.OpenMethod),
		"trial_policy":        cleanOptionalString(req.TrialPolicy),
		"trial_start_rule":    cleanOptionalString(req.TrialStartRule),
		"asset_config":        cleanOptionalString(req.AssetConfig),
		"doc_config":          cleanOptionalString(req.DocConfig),
		"release_channel":     cleanOptionalString(req.ReleaseChannel),
		"release_note":        cleanOptionalString(req.ReleaseNote),
		"is_platform_only":    visibilityScope == "PLATFORM_ONLY",
		"sort_order":          req.SortOrder,
	}
	var clients []models.SysAppClient
	if req.Clients != nil {
		clients, err = normalizeClientRequests(req.Clients)
		if err != nil {
			return dto.AppResponse{}, err
		}
		if err := s.repo.UpdateWithClients(ctx, &row, updates, clients); err != nil {
			return dto.AppResponse{}, err
		}
	} else {
		if err := s.repo.Update(ctx, &row, updates); err != nil {
			return dto.AppResponse{}, err
		}
		clients, err = s.repo.ClientsByAppID(ctx, row.ID)
		if err != nil {
			return dto.AppResponse{}, err
		}
	}
	return appToResponse(row, clients), nil
}

func (s *AppService) UpdateAppStatus(ctx context.Context, viewerID uint64, id uint64, req dto.AppStatusRequest) (dto.AppResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.AppResponse{}, err
	}
	if id == 0 {
		return dto.AppResponse{}, ErrAppNotFound
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if _, ok := allowedAppStatuses[status]; !ok {
		return dto.AppResponse{}, ErrInvalidAppStatus
	}
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AppResponse{}, ErrAppNotFound
		}
		return dto.AppResponse{}, err
	}
	if row.IsBuiltin && (status == "DISABLED" || status == "ARCHIVED") {
		return dto.AppResponse{}, ErrBuiltinStatusImmutable
	}
	if err := s.repo.Update(ctx, &row, map[string]interface{}{"status": status}); err != nil {
		return dto.AppResponse{}, err
	}
	clients, err := s.repo.ClientsByAppID(ctx, row.ID)
	if err != nil {
		return dto.AppResponse{}, err
	}
	return appToResponse(row, clients), nil
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

func normalizeClientRequests(reqs []dto.AppClientRequest) ([]models.SysAppClient, error) {
	clients := make([]models.SysAppClient, 0, len(reqs))
	seen := map[string]struct{}{}
	for _, req := range reqs {
		code := strings.ToUpper(strings.TrimSpace(req.ClientCode))
		name := strings.TrimSpace(req.ClientName)
		if code == "" {
			return nil, ErrClientCodeRequired
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		if name == "" {
			name = code
		}
		clients = append(clients, models.SysAppClient{
			ClientCode: code,
			ClientName: name,
			Enabled:    req.Enabled,
			SortOrder:  req.SortOrder,
			ConfigNote: cleanOptionalString(req.ConfigNote),
		})
	}
	return clients, nil
}

func appToResponse(row models.SysApp, clients []models.SysAppClient) dto.AppResponse {
	clientResponses := make([]dto.AppClientResponse, 0, len(clients))
	for _, client := range clients {
		clientResponses = append(clientResponses, dto.AppClientResponse{
			ID:         client.ID,
			AppID:      client.AppID,
			ClientCode: client.ClientCode,
			ClientName: client.ClientName,
			Enabled:    client.Enabled,
			SortOrder:  client.SortOrder,
			ConfigNote: client.ConfigNote,
			CreatedAt:  client.CreatedAt,
			UpdatedAt:  client.UpdatedAt,
		})
	}
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
		OwnerUserIDs:    row.OwnerUserIDs,
		Version:         row.Version,
		Description:     row.Description,
		DetailDesc:      row.DetailDesc,
		DeploymentMode:  defaultString(row.DeploymentMode, "MERGED"),
		CommModes:       row.CommModes,
		VisibilityMode:  row.VisibilityMode,
		VisibleTenants:  row.VisibleTenants,
		OpenMethod:      row.OpenMethod,
		TrialPolicy:     row.TrialPolicy,
		TrialStartRule:  row.TrialStartRule,
		AssetConfig:     row.AssetConfig,
		DocConfig:       row.DocConfig,
		ReleaseChannel:  row.ReleaseChannel,
		ReleaseNote:     row.ReleaseNote,
		IsBuiltin:       row.IsBuiltin,
		IsPlatformOnly:  row.IsPlatformOnly,
		SortOrder:       row.SortOrder,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		Clients:         clientResponses,
	}
}
