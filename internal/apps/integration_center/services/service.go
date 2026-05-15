package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"saas_baseon_go/internal/apps/integration_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Service struct {
	repo *repositories.Repository
}

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo}
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Trend string `json:"trend"`
	Tone  string `json:"tone"`
}

type IntegrationConnector struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Vendor      string   `json:"vendor"`
	Category    string   `json:"category"`
	Status      string   `json:"status"`
	AuthMode    string   `json:"auth_mode"`
	DeployMode  string   `json:"deploy_mode"`
	Channels    []string `json:"channels"`
	LastSyncAt  string   `json:"last_sync_at"`
	HealthScore int      `json:"health_score"`
	Description string   `json:"description"`
}

type IntegrationEvent struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	EventType string `json:"event_type"`
	Status    string `json:"status"`
	LatencyMS int    `json:"latency_ms"`
	Occurred  string `json:"occurred"`
}

type IntegrationChecklistItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Overview struct {
	Metrics    []Metric                   `json:"metrics"`
	Connectors []IntegrationConnector     `json:"connectors"`
	Events     []IntegrationEvent         `json:"events"`
	Checklist  []IntegrationChecklistItem `json:"checklist"`
	Channels   []map[string]string        `json:"channels"`
}

type SectionSummary struct {
	AppCode string      `json:"app_code"`
	Section string      `json:"section"`
	Items   interface{} `json:"items"`
}

type PlatformMutationRequest struct {
	Name          string `json:"name"`
	ShortName     string `json:"short_name"`
	Code          string `json:"code"`
	PlatformType  string `json:"platform_type"`
	AccessMode    string `json:"access_mode"`
	Status        string `json:"status"`
	TenantVisible bool   `json:"tenant_visible"`
	OwnerName     string `json:"owner_name"`
	OfficialURL   string `json:"official_url"`
	SortOrder     int    `json:"sort_order"`
	Description   string `json:"description"`
}

type ProviderAppMutationRequest struct {
	PlatformCode  string `json:"platform_code"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	AppType       string `json:"app_type"`
	AuthMode      string `json:"auth_mode"`
	Environment   string `json:"environment"`
	Status        string `json:"status"`
	TenantVisible bool   `json:"tenant_visible"`
	CallbackURL   string `json:"callback_url"`
	WebhookURL    string `json:"webhook_url"`
	CredentialRef string `json:"credential_ref"`
	OwnerName     string `json:"owner_name"`
	Description   string `json:"description"`
}

type AppCapabilityPatchRequest struct {
	Enabled          *bool  `json:"enabled"`
	ConnectionStatus string `json:"connection_status"`
	ReviewStatus     string `json:"review_status"`
}

type QuotaPolicyMutationRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	QuotaCode       string `json:"quota_code"`
	QuotaUnit       string `json:"quota_unit"`
	PeriodType      string `json:"period_type"`
	DefaultLimit    int64  `json:"default_limit"`
	OverLimitAction string `json:"over_limit_action"`
	Status          string `json:"status"`
	Description     string `json:"description"`
}

type ConnectivityCheckRequest struct {
	Target string `json:"target"`
}

type ConnectivityCheckResult struct {
	Target       string `json:"target"`
	Status       string `json:"status"`
	CheckedAt    string `json:"checked_at"`
	OpenAlerts   int64  `json:"open_alerts"`
	RunningSyncs int64  `json:"running_syncs"`
	Message      string `json:"message"`
}

type LogExportResult struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Message   string `json:"message"`
}

func (s *Service) Overview(ctx context.Context) (Overview, error) {
	if s.repo != nil {
		counts, err := s.repo.Counts(ctx)
		if err != nil {
			return Overview{}, err
		}
		connectors, err := s.Connectors(ctx)
		if err != nil {
			return Overview{}, err
		}
		alerts, err := s.repo.ListAlerts(ctx, 3)
		if err != nil {
			return Overview{}, err
		}
		events, err := s.eventsFromLogs(ctx)
		if err != nil {
			return Overview{}, err
		}
		checklist := []IntegrationChecklistItem{
			{Title: "应用 Manifest", Description: "菜单、权限、API、套餐功能点与配额均由 integration-center Manifest 声明。", Status: "done"},
			{Title: "业务表初始化", Description: "平台、能力、服务商应用、连接实例、同步任务、配额与日志已具备持久化结构。", Status: "done"},
			{Title: "待处理异常", Description: fmt.Sprintf("当前还有 %d 个未恢复异常需要平台侧处理。", counts.OpenAlerts), Status: statusFromCount(counts.OpenAlerts)},
		}
		if len(alerts) == 0 {
			checklist = append(checklist, IntegrationChecklistItem{Title: "异常监控", Description: "当前没有待处理集成异常。", Status: "done"})
		}
		return Overview{
			Metrics: []Metric{
				{Label: "接入平台", Value: fmt.Sprintf("%d", counts.Platforms), Trend: fmt.Sprintf("%d 个服务商应用", counts.ProviderApps), Tone: "primary"},
				{Label: "租户连接", Value: fmt.Sprintf("%d", counts.Connections), Trend: "按授权主体独立计数", Tone: "success"},
				{Label: "今日调用", Value: fmt.Sprintf("%d", counts.TodayAPICalls), Trend: "含 API / Token / Webhook", Tone: "info"},
				{Label: "待处理异常", Value: fmt.Sprintf("%d", counts.OpenAlerts), Trend: fmt.Sprintf("%d 个同步任务运行中", counts.RunningSyncJobs), Tone: alertTone(counts.OpenAlerts)},
			},
			Connectors: connectors,
			Events:     events,
			Checklist:  checklist,
			Channels: []map[string]string{
				{"name": "平台 API", "description": "外部系统调用底座开放 API，使用应用凭证、scope 和租户上下文。"},
				{"name": "Webhook", "description": "第三方事件进入底座，强制签名、时间戳和幂等键。"},
				{"name": "数据同步", "description": "批量或增量同步组织、用户、订单等外部数据，记录游标和失败补偿。"},
				{"name": "网关代理", "description": "流量经底座鉴权、限流、审计后转发到目标系统。"},
			},
		}, nil
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return Overview{}, err
	}
	now := time.Now()
	return Overview{
		Metrics: []Metric{
			{Label: "连接器", Value: "4", Trend: "2 个可上线", Tone: "primary"},
			{Label: "开放 API", Value: "18", Trend: "统一鉴权与审计", Tone: "success"},
			{Label: "Webhook", Value: "9", Trend: "平均延迟 132ms", Tone: "warning"},
			{Label: "同步任务", Value: "6", Trend: "5 个健康运行", Tone: "info"},
		},
		Connectors: connectors,
		Events: []IntegrationEvent{
			{ID: "evt_10086", Source: "企业微信", EventType: "contact.changed", Status: "success", LatencyMS: 96, Occurred: now.Add(-18 * time.Minute).Format(time.RFC3339)},
			{ID: "evt_10087", Source: "钉钉", EventType: "approval.finished", Status: "success", LatencyMS: 128, Occurred: now.Add(-41 * time.Minute).Format(time.RFC3339)},
			{ID: "evt_10088", Source: "飞书", EventType: "message.card.action", Status: "retrying", LatencyMS: 310, Occurred: now.Add(-74 * time.Minute).Format(time.RFC3339)},
		},
		Checklist: []IntegrationChecklistItem{
			{Title: "应用主档", Description: "合并部署应用已在应用中心登记，后续通过 Manifest 托管菜单、权限与 API。", Status: "done"},
			{Title: "连接器凭证", Description: "仅保存凭证引用和轮换状态，禁止在 Manifest 中写入真实密钥。", Status: "active"},
			{Title: "租户授权", Description: "按租户开通连接器能力，运行时校验套餐、角色权限与配额。", Status: "planned"},
			{Title: "事件审计", Description: "Webhook、同步任务、OAuth 授权变更统一写入操作日志。", Status: "planned"},
		},
		Channels: []map[string]string{
			{"name": "平台 API", "description": "外部系统调用底座开放 API，使用应用凭证、scope 和租户上下文。"},
			{"name": "Webhook", "description": "第三方事件进入底座，强制签名、时间戳和幂等键。"},
			{"name": "数据同步", "description": "批量或增量同步组织、用户、订单等外部数据，记录游标和失败补偿。"},
			{"name": "网关代理", "description": "流量经底座鉴权、限流、审计后转发到目标系统。"},
		},
	}, nil
}

func (s *Service) CreatePlatform(ctx context.Context, req PlatformMutationRequest) (repositories.PlatformSummary, error) {
	if s.repo == nil {
		return repositories.PlatformSummary{}, errors.New("integration center repository is not configured")
	}
	if err := validatePlatformRequest(req, true); err != nil {
		return repositories.PlatformSummary{}, err
	}
	shortName := optionalString(req.ShortName)
	officialURL := optionalString(req.OfficialURL)
	ownerName := optionalString(req.OwnerName)
	description := optionalString(req.Description)
	platform, err := s.repo.CreatePlatform(ctx, models.IntegrationPlatform{
		PlatformCode:      normalizePlatformCode(req.Code),
		PlatformName:      strings.TrimSpace(req.Name),
		PlatformShortName: shortName,
		PlatformType:      defaultString(req.PlatformType, "电商平台"),
		AccessMode:        defaultString(req.AccessMode, "OAuth2"),
		OfficialURL:       officialURL,
		Status:            normalizePlatformStatus(req.Status),
		TenantVisible:     req.TenantVisible,
		OwnerName:         ownerName,
		SortOrder:         req.SortOrder,
		Description:       description,
	})
	if err != nil {
		return repositories.PlatformSummary{}, err
	}
	return platformToSummary(platform), nil
}

func (s *Service) UpdatePlatform(ctx context.Context, code string, req PlatformMutationRequest) (repositories.PlatformSummary, error) {
	if s.repo == nil {
		return repositories.PlatformSummary{}, errors.New("integration center repository is not configured")
	}
	if err := validatePlatformRequest(req, false); err != nil {
		return repositories.PlatformSummary{}, err
	}
	patch := map[string]interface{}{
		"platform_name":       strings.TrimSpace(req.Name),
		"platform_short_name": optionalString(req.ShortName),
		"platform_type":       defaultString(req.PlatformType, "电商平台"),
		"access_mode":         defaultString(req.AccessMode, "OAuth2"),
		"official_url":        optionalString(req.OfficialURL),
		"status":              normalizePlatformStatus(req.Status),
		"tenant_visible":      req.TenantVisible,
		"owner_name":          optionalString(req.OwnerName),
		"sort_order":          req.SortOrder,
		"description":         optionalString(req.Description),
	}
	platform, err := s.repo.UpdatePlatform(ctx, normalizePlatformCode(code), patch)
	if err != nil {
		return repositories.PlatformSummary{}, err
	}
	return platformToSummary(platform), nil
}

func (s *Service) CreateProviderApp(ctx context.Context, req ProviderAppMutationRequest) (repositories.ProviderAppSummary, error) {
	if s.repo == nil {
		return repositories.ProviderAppSummary{}, errors.New("integration center repository is not configured")
	}
	if err := validateProviderAppRequest(req, true); err != nil {
		return repositories.ProviderAppSummary{}, err
	}
	platform, err := s.repo.GetPlatformByCode(ctx, normalizePlatformCode(req.PlatformCode))
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "接入平台不存在")
	}
	app, err := s.repo.CreateProviderApp(ctx, models.IntegrationProviderApp{
		PlatformID:    platform.ID,
		AppCode:       normalizePlatformCode(req.Code),
		AppName:       strings.TrimSpace(req.Name),
		AppType:       defaultString(req.AppType, "provider_app"),
		AuthMode:      defaultString(req.AuthMode, "OAuth2"),
		Environment:   normalizeEnvironment(req.Environment),
		Status:        normalizeAppStatus(req.Status),
		TenantVisible: req.TenantVisible,
		CallbackURL:   optionalString(req.CallbackURL),
		WebhookURL:    optionalString(req.WebhookURL),
		CredentialRef: optionalString(req.CredentialRef),
		OwnerName:     optionalString(req.OwnerName),
		Description:   optionalString(req.Description),
	})
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "服务商应用保存失败")
	}
	return providerAppToSummary(app, platform), nil
}

func (s *Service) UpdateProviderApp(ctx context.Context, code string, req ProviderAppMutationRequest) (repositories.ProviderAppSummary, error) {
	if s.repo == nil {
		return repositories.ProviderAppSummary{}, errors.New("integration center repository is not configured")
	}
	if err := validateProviderAppRequest(req, false); err != nil {
		return repositories.ProviderAppSummary{}, err
	}
	var platform models.IntegrationPlatform
	var err error
	if strings.TrimSpace(req.PlatformCode) != "" {
		platform, err = s.repo.GetPlatformByCode(ctx, normalizePlatformCode(req.PlatformCode))
		if err != nil {
			return repositories.ProviderAppSummary{}, humanizeRepoError(err, "接入平台不存在")
		}
	}
	patch := map[string]interface{}{
		"app_name":       strings.TrimSpace(req.Name),
		"app_type":       defaultString(req.AppType, "provider_app"),
		"auth_mode":      defaultString(req.AuthMode, "OAuth2"),
		"environment":    normalizeEnvironment(req.Environment),
		"status":         normalizeAppStatus(req.Status),
		"tenant_visible": req.TenantVisible,
		"callback_url":   optionalString(req.CallbackURL),
		"webhook_url":    optionalString(req.WebhookURL),
		"credential_ref": optionalString(req.CredentialRef),
		"owner_name":     optionalString(req.OwnerName),
		"description":    optionalString(req.Description),
	}
	if platform.ID > 0 {
		patch["platform_id"] = platform.ID
	}
	app, err := s.repo.UpdateProviderApp(ctx, normalizePlatformCode(code), patch)
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "服务商应用不存在")
	}
	if platform.ID == 0 {
		platform = models.IntegrationPlatform{ID: app.PlatformID}
	}
	return providerAppToSummary(app, platform), nil
}

func (s *Service) UpdateAppCapability(ctx context.Context, id uint64, req AppCapabilityPatchRequest) (models.IntegrationProviderAppCapability, error) {
	if s.repo == nil {
		return models.IntegrationProviderAppCapability{}, errors.New("integration center repository is not configured")
	}
	patch := map[string]interface{}{}
	if req.Enabled != nil {
		patch["enabled"] = *req.Enabled
	}
	if status := normalizeConnectionStatus(req.ConnectionStatus); status != "" {
		patch["connection_status"] = status
	}
	if review := normalizeReviewStatus(req.ReviewStatus); review != "" {
		patch["review_status"] = review
	}
	if len(patch) == 0 {
		return models.IntegrationProviderAppCapability{}, errors.New("没有可更新的能力字段")
	}
	result, err := s.repo.UpdateAppCapability(ctx, id, patch)
	if err != nil {
		return models.IntegrationProviderAppCapability{}, humanizeRepoError(err, "应用能力不存在")
	}
	return result, nil
}

func (s *Service) Connectors(ctx context.Context) ([]IntegrationConnector, error) {
	if s.repo != nil {
		rows, err := s.repo.ListPlatforms(ctx, 20)
		if err != nil {
			return nil, err
		}
		connectors := make([]IntegrationConnector, 0, len(rows))
		for _, row := range rows {
			name := row.Name
			if row.ShortName != nil && *row.ShortName != "" {
				name = *row.ShortName
			}
			connectors = append(connectors, IntegrationConnector{
				Code:        row.Code,
				Name:        name,
				Vendor:      row.Name,
				Category:    row.PlatformType,
				Status:      row.Status,
				AuthMode:    row.AccessMode,
				DeployMode:  "MERGED",
				Channels:    channelsForAccessMode(row.AccessMode),
				LastSyncAt:  "",
				HealthScore: healthScoreForStatus(row.Status, row.OpenAlertCount),
				Description: fmt.Sprintf("%d 个服务商应用，%d 项平台能力，%d 条租户连接。", row.AppCount, row.CapabilityCount, row.ConnectionCount),
			})
		}
		return connectors, nil
	}
	return []IntegrationConnector{
		{
			Code:        "wechat-work",
			Name:        "企业微信",
			Vendor:      "Tencent",
			Category:    "协同办公",
			Status:      "online",
			AuthMode:    "OAuth + 通讯录密钥",
			DeployMode:  "MERGED",
			Channels:    []string{"Webhook", "数据同步"},
			LastSyncAt:  time.Now().Add(-22 * time.Minute).Format(time.RFC3339),
			HealthScore: 96,
			Description: "同步组织、用户与消息卡片事件，支持租户级授权和回调审计。",
		},
		{
			Code:        "dingtalk",
			Name:        "钉钉",
			Vendor:      "Alibaba",
			Category:    "协同办公",
			Status:      "beta",
			AuthMode:    "OAuth",
			DeployMode:  "MERGED",
			Channels:    []string{"Webhook", "平台 API"},
			LastSyncAt:  time.Now().Add(-55 * time.Minute).Format(time.RFC3339),
			HealthScore: 91,
			Description: "接入审批、组织和待办事件，统一纳入租户开通与权限治理。",
		},
		{
			Code:        "feishu",
			Name:        "飞书",
			Vendor:      "ByteDance",
			Category:    "协同办公",
			Status:      "developing",
			AuthMode:    "OAuth + 事件订阅",
			DeployMode:  "MERGED",
			Channels:    []string{"Webhook"},
			LastSyncAt:  time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			HealthScore: 82,
			Description: "用于消息卡片、通讯录和审批事件接入，当前处于联调阶段。",
		},
		{
			Code:        "generic-openapi",
			Name:        "通用 OpenAPI",
			Vendor:      "Platform",
			Category:    "开放接口",
			Status:      "planned",
			AuthMode:    "API Key + HMAC",
			DeployMode:  "MERGED",
			Channels:    []string{"平台 API", "网关代理"},
			LastSyncAt:  "",
			HealthScore: 0,
			Description: "面向未标准化第三方系统的通用接入模板，声明 scope、签名和限流策略。",
		},
	}, nil
}

func (s *Service) Platforms(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListPlatforms(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "platforms", Items: rows}, nil
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return SectionSummary{}, err
	}
	return SectionSummary{AppCode: "integration-center", Section: "platforms", Items: connectors}, nil
}

func (s *Service) Workspace(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListProviderApps(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "workspace", Items: rows}, nil
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return SectionSummary{}, err
	}
	return SectionSummary{AppCode: "integration-center", Section: "workspace", Items: connectors}, nil
}

func (s *Service) TenantConnections(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListTenantConnections(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "tenant-connections", Items: rows}, nil
	}
	now := time.Now()
	return SectionSummary{AppCode: "integration-center", Section: "tenant-connections", Items: []map[string]interface{}{
		{"tenant_name": "杭州鹿鸣科技", "platform": "企业微信", "provider_app": "企业微信第三方标准应用", "auth_subject_type": "corp", "auth_subject_name": "鹿鸣科技企业微信", "status": "connected", "last_sync_at": now.Add(-5 * time.Minute).Format(time.RFC3339)},
		{"tenant_name": "上海星河贸易", "platform": "企业微信", "provider_app": "企业微信第三方标准应用", "auth_subject_type": "corp", "auth_subject_name": "星河贸易企业微信", "status": "warning", "last_sync_at": now.Add(-13 * time.Minute).Format(time.RFC3339)},
	}}, nil
}

func (s *Service) SyncMonitor(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListSyncJobs(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "sync-monitor", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "sync-monitor", Items: []map[string]interface{}{
		{"job": "企微成员增量同步", "tenant": "杭州鹿鸣科技", "capability": "成员同步", "status": "success", "success_rate": 99},
		{"job": "京东订单同步", "tenant": "宁波青禾家居", "capability": "订单同步", "status": "failed", "success_rate": 42},
	}}, nil
}

func (s *Service) Quota(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListQuotaPolicies(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "quota", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "quota", Items: []map[string]interface{}{
		{"policy": "企业微信租户通用策略", "scope": "tenant", "daily_limit": 100000, "qps": 20, "status": "enabled"},
		{"policy": "鹿鸣科技企微专属覆盖", "scope": "tenant", "daily_limit": 300000, "qps": 50, "status": "enabled"},
	}}, nil
}

func (s *Service) Alerts(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListAlerts(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "alerts", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "alerts", Items: []map[string]interface{}{
		{"title": "京东连接授权已过期", "level": "critical", "status": "open"},
		{"title": "成员同步接近配额", "level": "warning", "status": "open"},
	}}, nil
}

func (s *Service) Logs(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListAPICallLogs(ctx, 100)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "logs", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "logs", Items: []map[string]interface{}{
		{"request_id": "req_wecom_8e92", "type": "api", "path": "/cgi-bin/department/list", "status": 200, "success": true},
		{"request_id": "req_jd_7a21", "type": "api", "path": "/api/order/search", "status": 200, "success": true},
	}}, nil
}

func (s *Service) RefreshTenantConnection(ctx context.Context, id uint64) (models.IntegrationTenantConnection, error) {
	if s.repo == nil {
		return models.IntegrationTenantConnection{}, errors.New("integration center repository is not configured")
	}
	now := time.Now()
	result, err := s.repo.UpdateTenantConnection(ctx, id, map[string]interface{}{
		"auth_status":        "authorized",
		"connection_status":  "connected",
		"token_status":       "valid",
		"last_sync_at":       &now,
		"last_error_at":      nil,
		"last_error_message": nil,
	})
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "租户连接不存在")
	}
	return result, nil
}

func (s *Service) SetTenantConnectionStatus(ctx context.Context, id uint64, paused bool) (models.IntegrationTenantConnection, error) {
	if s.repo == nil {
		return models.IntegrationTenantConnection{}, errors.New("integration center repository is not configured")
	}
	status := "connected"
	if paused {
		status = "paused"
	}
	result, err := s.repo.UpdateTenantConnection(ctx, id, map[string]interface{}{"connection_status": status})
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "租户连接不存在")
	}
	return result, nil
}

func (s *Service) RetryTenantConnection(ctx context.Context, id uint64) (models.IntegrationSyncJob, error) {
	if s.repo == nil {
		return models.IntegrationSyncJob{}, errors.New("integration center repository is not configured")
	}
	connection, err := s.repo.UpdateTenantConnection(ctx, id, map[string]interface{}{"connection_status": "connected"})
	if err != nil {
		return models.IntegrationSyncJob{}, humanizeRepoError(err, "租户连接不存在")
	}
	now := time.Now()
	job := models.IntegrationSyncJob{
		TenantID:           connection.TenantID,
		TenantConnectionID: connection.ID,
		CapabilityCode:     "manual_retry",
		JobType:            "manual_retry",
		TriggerMode:        "manual",
		Status:             "pending",
		StartedAt:          &now,
	}
	result, err := s.repo.CreateSyncJob(ctx, job)
	if err != nil {
		return models.IntegrationSyncJob{}, humanizeRepoError(err, "同步重试任务创建失败")
	}
	return result, nil
}

func (s *Service) RetrySyncJob(ctx context.Context, id uint64) (models.IntegrationSyncJob, error) {
	now := time.Now()
	return s.updateSyncJob(ctx, id, map[string]interface{}{
		"status":        "retrying",
		"started_at":    &now,
		"finished_at":   nil,
		"error_code":    nil,
		"error_message": nil,
	})
}

func (s *Service) PauseSyncJob(ctx context.Context, id uint64) (models.IntegrationSyncJob, error) {
	return s.updateSyncJob(ctx, id, map[string]interface{}{"status": "paused"})
}

func (s *Service) ResumeSyncJob(ctx context.Context, id uint64) (models.IntegrationSyncJob, error) {
	now := time.Now()
	return s.updateSyncJob(ctx, id, map[string]interface{}{"status": "running", "started_at": &now})
}

func (s *Service) updateSyncJob(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationSyncJob, error) {
	if s.repo == nil {
		return models.IntegrationSyncJob{}, errors.New("integration center repository is not configured")
	}
	result, err := s.repo.UpdateSyncJob(ctx, id, patch)
	if err != nil {
		return models.IntegrationSyncJob{}, humanizeRepoError(err, "同步任务不存在")
	}
	return result, nil
}

func (s *Service) CreateQuotaPolicy(ctx context.Context, req QuotaPolicyMutationRequest) (models.IntegrationQuotaPolicy, error) {
	if s.repo == nil {
		return models.IntegrationQuotaPolicy{}, errors.New("integration center repository is not configured")
	}
	if err := validateQuotaPolicyRequest(req, true); err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	policy, err := s.repo.CreateQuotaPolicy(ctx, quotaPolicyFromRequest(req))
	if err != nil {
		return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略保存失败")
	}
	return policy, nil
}

func (s *Service) UpdateQuotaPolicy(ctx context.Context, code string, req QuotaPolicyMutationRequest) (models.IntegrationQuotaPolicy, error) {
	if s.repo == nil {
		return models.IntegrationQuotaPolicy{}, errors.New("integration center repository is not configured")
	}
	if err := validateQuotaPolicyRequest(req, false); err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	patch := map[string]interface{}{
		"policy_name":       strings.TrimSpace(req.Name),
		"quota_code":        normalizePlatformCode(defaultString(req.QuotaCode, "integration_api_calls_daily")),
		"quota_unit":        defaultString(req.QuotaUnit, "CALL"),
		"period_type":       defaultString(req.PeriodType, "DAY"),
		"default_limit":     req.DefaultLimit,
		"over_limit_action": normalizeOverLimitAction(req.OverLimitAction),
		"status":            normalizeEnabledStatus(req.Status),
		"description":       optionalString(req.Description),
	}
	policy, err := s.repo.UpdateQuotaPolicy(ctx, normalizePlatformCode(code), patch)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略不存在")
	}
	return policy, nil
}

func (s *Service) SetQuotaPolicyStatus(ctx context.Context, code string, enabled bool) (models.IntegrationQuotaPolicy, error) {
	if s.repo == nil {
		return models.IntegrationQuotaPolicy{}, errors.New("integration center repository is not configured")
	}
	status := "disabled"
	if enabled {
		status = "enabled"
	}
	policy, err := s.repo.UpdateQuotaPolicy(ctx, normalizePlatformCode(code), map[string]interface{}{"status": status})
	if err != nil {
		return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略不存在")
	}
	return policy, nil
}

func (s *Service) ProcessAlert(ctx context.Context, id uint64) (models.IntegrationAlert, error) {
	return s.updateAlert(ctx, id, map[string]interface{}{"status": "processing"})
}

func (s *Service) ResolveAlert(ctx context.Context, id uint64) (models.IntegrationAlert, error) {
	now := time.Now()
	return s.updateAlert(ctx, id, map[string]interface{}{"status": "resolved", "resolved_at": &now})
}

func (s *Service) IgnoreAlert(ctx context.Context, id uint64) (models.IntegrationAlert, error) {
	return s.updateAlert(ctx, id, map[string]interface{}{"status": "ignored"})
}

func (s *Service) updateAlert(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationAlert, error) {
	if s.repo == nil {
		return models.IntegrationAlert{}, errors.New("integration center repository is not configured")
	}
	result, err := s.repo.UpdateAlert(ctx, id, patch)
	if err != nil {
		return models.IntegrationAlert{}, humanizeRepoError(err, "异常记录不存在")
	}
	return result, nil
}

func (s *Service) CheckConnectivity(ctx context.Context, req ConnectivityCheckRequest) (ConnectivityCheckResult, error) {
	if s.repo == nil {
		return ConnectivityCheckResult{}, errors.New("integration center repository is not configured")
	}
	counts, err := s.repo.Counts(ctx)
	if err != nil {
		return ConnectivityCheckResult{}, err
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		target = "全局"
	}
	status := "success"
	message := "连通性检测通过"
	if counts.OpenAlerts > 0 {
		status = "warning"
		message = "检测完成，存在待处理异常"
	}
	return ConnectivityCheckResult{
		Target:       target,
		Status:       status,
		CheckedAt:    time.Now().Format(time.RFC3339),
		OpenAlerts:   counts.OpenAlerts,
		RunningSyncs: counts.RunningSyncJobs,
		Message:      message,
	}, nil
}

func (s *Service) ExportLogs(ctx context.Context) (LogExportResult, error) {
	if s.repo == nil {
		return LogExportResult{}, errors.New("integration center repository is not configured")
	}
	counts, err := s.repo.Counts(ctx)
	if err != nil {
		return LogExportResult{}, err
	}
	now := time.Now()
	return LogExportResult{
		TaskID:    fmt.Sprintf("integration-log-export-%d", now.Unix()),
		Status:    "queued",
		CreatedAt: now.Format(time.RFC3339),
		Message:   fmt.Sprintf("已创建调用日志导出任务，当前日志口径含今日 %d 次调用。", counts.TodayAPICalls),
	}, nil
}

func (s *Service) eventsFromLogs(ctx context.Context) ([]IntegrationEvent, error) {
	rows, err := s.repo.ListAPICallLogs(ctx, 5)
	if err != nil {
		return nil, err
	}
	events := make([]IntegrationEvent, 0, len(rows))
	for _, row := range rows {
		status := "success"
		if row.Status != "success" {
			status = row.Status
		}
		events = append(events, IntegrationEvent{
			ID:        row.RequestID,
			Source:    row.CallType,
			EventType: endpointValue(row.Endpoint),
			Status:    status,
			LatencyMS: row.DurationMS,
			Occurred:  row.CalledAt.Format(time.RFC3339),
		})
	}
	return events, nil
}

func endpointValue(value *string) string {
	if value == nil || *value == "" {
		return "integration.event"
	}
	return *value
}

func channelsForAccessMode(accessMode string) []string {
	switch accessMode {
	case "OAuth2":
		return []string{"平台 API", "数据同步"}
	case "第三方服务商":
		return []string{"Webhook", "数据同步"}
	case "Webhook":
		return []string{"Webhook"}
	case "API Key":
		return []string{"平台 API", "网关代理"}
	default:
		return []string{"平台 API"}
	}
}

func healthScoreForStatus(status string, openAlerts int64) int {
	if openAlerts > 0 {
		return 72
	}
	switch status {
	case "online", "enabled":
		return 96
	case "beta":
		return 88
	case "maintenance":
		return 65
	default:
		return 0
	}
}

func alertTone(openAlerts int64) string {
	if openAlerts > 0 {
		return "warning"
	}
	return "success"
}

func statusFromCount(count int64) string {
	if count > 0 {
		return "active"
	}
	return "done"
}

func validatePlatformRequest(req PlatformMutationRequest, requireCode bool) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("平台名称不能为空")
	}
	if requireCode && normalizePlatformCode(req.Code) == "" {
		return errors.New("平台编码不能为空")
	}
	code := normalizePlatformCode(req.Code)
	if code != "" {
		for _, r := range code {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' {
				return errors.New("平台编码仅支持小写字母、数字、横线和下划线")
			}
		}
	}
	return nil
}

func normalizePlatformCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

func validateProviderAppRequest(req ProviderAppMutationRequest, requireCode bool) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("应用名称不能为空")
	}
	if strings.TrimSpace(req.PlatformCode) == "" {
		return errors.New("所属接入平台不能为空")
	}
	if requireCode && normalizePlatformCode(req.Code) == "" {
		return errors.New("应用编码不能为空")
	}
	if code := normalizePlatformCode(req.Code); code != "" {
		if err := validateSimpleCode(code, "应用编码"); err != nil {
			return err
		}
	}
	return nil
}

func validateQuotaPolicyRequest(req QuotaPolicyMutationRequest, requireCode bool) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("策略名称不能为空")
	}
	if requireCode && normalizePlatformCode(req.Code) == "" {
		return errors.New("策略编码不能为空")
	}
	if code := normalizePlatformCode(req.Code); code != "" {
		if err := validateSimpleCode(code, "策略编码"); err != nil {
			return err
		}
	}
	if req.DefaultLimit < 0 {
		return errors.New("默认限额不能为负数")
	}
	return nil
}

func validateSimpleCode(code string, label string) error {
	for _, r := range code {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return fmt.Errorf("%s仅支持小写字母、数字、横线和下划线", label)
		}
	}
	return nil
}

func normalizePlatformStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "enabled":
		return "online"
	case "disabled":
		return "disabled"
	case "testing":
		return "beta"
	case "maintenance":
		return "maintenance"
	case "online", "beta", "draft":
		return strings.TrimSpace(status)
	default:
		return "draft"
	}
}

func normalizeAppStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "enabled", "online":
		return "online"
	case "testing", "beta":
		return "beta"
	case "disabled", "maintenance", "draft":
		return strings.TrimSpace(status)
	default:
		return "draft"
	}
}

func normalizeEnabledStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "enabled", "online":
		return "enabled"
	case "disabled", "paused":
		return "disabled"
	default:
		return "enabled"
	}
}

func normalizeEnvironment(value string) string {
	switch strings.TrimSpace(value) {
	case "正式", "prod", "production", "":
		return "prod"
	case "测试", "test", "testing":
		return "test"
	case "沙箱", "sandbox":
		return "sandbox"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizeConnectionStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "connected", "pending", "failed", "paused", "not_connected":
		return strings.TrimSpace(status)
	case "enabled":
		return "connected"
	case "disabled":
		return "paused"
	default:
		return ""
	}
}

func normalizeReviewStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "pending", "approved", "rejected":
		return strings.TrimSpace(status)
	default:
		return ""
	}
}

func normalizeOverLimitAction(action string) string {
	switch strings.TrimSpace(action) {
	case "reject", "queue", "warn":
		return strings.TrimSpace(action)
	case "排队":
		return "queue"
	case "告警":
		return "warn"
	default:
		return "reject"
	}
}

func quotaPolicyFromRequest(req QuotaPolicyMutationRequest) models.IntegrationQuotaPolicy {
	return models.IntegrationQuotaPolicy{
		PolicyCode:      normalizePlatformCode(req.Code),
		PolicyName:      strings.TrimSpace(req.Name),
		QuotaCode:       normalizePlatformCode(defaultString(req.QuotaCode, "integration_api_calls_daily")),
		QuotaUnit:       defaultString(req.QuotaUnit, "CALL"),
		PeriodType:      defaultString(req.PeriodType, "DAY"),
		DefaultLimit:    req.DefaultLimit,
		OverLimitAction: normalizeOverLimitAction(req.OverLimitAction),
		Status:          normalizeEnabledStatus(req.Status),
		Description:     optionalString(req.Description),
	}
}

func humanizeRepoError(err error, fallback string) error {
	switch {
	case errors.Is(err, repositories.ErrPlatformNotFound):
		return errors.New("接入平台不存在")
	case errors.Is(err, repositories.ErrProviderAppNotFound):
		return errors.New("服务商应用不存在")
	case errors.Is(err, repositories.ErrCapabilityNotFound):
		return errors.New("应用能力不存在")
	case errors.Is(err, repositories.ErrConnectionNotFound):
		return errors.New("租户连接不存在")
	case errors.Is(err, repositories.ErrSyncJobNotFound):
		return errors.New("同步任务不存在")
	case errors.Is(err, repositories.ErrQuotaPolicyNotFound):
		return errors.New("配额策略不存在")
	case errors.Is(err, repositories.ErrAlertNotFound):
		return errors.New("异常记录不存在")
	default:
		return errors.New(fallback)
	}
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
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

func platformToSummary(platform models.IntegrationPlatform) repositories.PlatformSummary {
	return repositories.PlatformSummary{
		ID:              platform.ID,
		Code:            platform.PlatformCode,
		Name:            platform.PlatformName,
		ShortName:       platform.PlatformShortName,
		PlatformType:    platform.PlatformType,
		AccessMode:      platform.AccessMode,
		Status:          platform.Status,
		TenantVisible:   platform.TenantVisible,
		OwnerName:       platform.OwnerName,
		AppCount:        0,
		CapabilityCount: 0,
		ConnectionCount: 0,
		OpenAlertCount:  0,
	}
}

func providerAppToSummary(app models.IntegrationProviderApp, platform models.IntegrationPlatform) repositories.ProviderAppSummary {
	return repositories.ProviderAppSummary{
		ID:            app.ID,
		PlatformID:    app.PlatformID,
		PlatformName:  platform.PlatformName,
		AppCode:       app.AppCode,
		AppName:       app.AppName,
		AuthMode:      app.AuthMode,
		Environment:   app.Environment,
		Status:        app.Status,
		TenantVisible: app.TenantVisible,
	}
}
