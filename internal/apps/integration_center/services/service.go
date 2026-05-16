package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"saas_baseon_go/internal/apps/integration_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Service struct {
	repo               *repositories.Repository
	syncProcessor      SyncProcessor
	apiCallLogRecorder APICallLogRecorder
}

const (
	platformStatusOnline      = "online"
	platformStatusBeta        = "beta"
	platformStatusDraft       = "draft"
	platformStatusDisabled    = "disabled"
	platformStatusMaintenance = "maintenance"

	appStatusOnline      = "online"
	appStatusBeta        = "beta"
	appStatusDraft       = "draft"
	appStatusDisabled    = "disabled"
	appStatusMaintenance = "maintenance"

	enabledStatusEnabled  = "enabled"
	enabledStatusDisabled = "disabled"

	connectionStatusConnected    = "connected"
	connectionStatusPending      = "pending"
	connectionStatusFailed       = "failed"
	connectionStatusPaused       = "paused"
	connectionStatusNotConnected = "not_connected"

	reviewStatusPending  = "pending"
	reviewStatusApproved = "approved"
	reviewStatusRejected = "rejected"
)

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo, syncProcessor: HTTPSyncProcessor{}, apiCallLogRecorder: NewRepositoryAPICallLogRecorder(repo)}
}

func (s *Service) SetSyncProcessor(processor SyncProcessor) {
	s.syncProcessor = processor
}

func (s *Service) SetAPICallLogRecorder(recorder APICallLogRecorder) {
	s.apiCallLogRecorder = recorder
}

func (s *Service) StartWebhookEventWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	go func() {
		timer := time.NewTimer(15 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				_, _ = s.ProcessDueWebhookEvents(ctx, 50)
				timer.Reset(interval)
			}
		}
	}()
}

func (s *Service) StartSyncJobWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	go func() {
		timer := time.NewTimer(20 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				_, _ = s.ProcessDueSyncJobs(ctx, 50)
				timer.Reset(interval)
			}
		}
	}()
}

func (s *Service) StartAPICallLogRetentionWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	go func() {
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				_, _ = s.ArchiveExpiredAPICallLogs(ctx, 1000)
				timer.Reset(interval)
			}
		}
	}()
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
	Total   int64       `json:"total"`
	Skip    int         `json:"skip"`
	Limit   int         `json:"limit"`
}

func nonNilRows[T any](rows []T) []T {
	if rows == nil {
		return []T{}
	}
	return rows
}

type PageRequest struct {
	Skip            int
	Limit           int
	Keyword         string
	Status          string
	PlatformCode    string
	ProviderAppCode string
	TenantID        *uint64
	StartTime       *time.Time
	EndTime         *time.Time
	SortBy          string
	SortOrder       string
}

type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
	TraceID   string
}

type viewer struct {
	UserID          uint64
	TenantID        uint64
	IsPlatformAdmin bool
}

var (
	ErrUnauthorized  = errors.New("未登录或登录态无效")
	ErrForbidden     = errors.New("没有第三方集成中心操作权限")
	ErrQuotaExceeded = errors.New("第三方连接实例数已超出套餐配额")
)

const webhookTimestampSkew = 5 * time.Minute
const webhookMaxRetryCount = 3
const syncJobMaxRetryCount = 3
const defaultAPICallLogRetentionDays = 180

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

type ProviderAppCredentialRotateRequest struct {
	CredentialRef string `json:"credential_ref"`
}

type AppCapabilityPatchRequest struct {
	Enabled            *bool                  `json:"enabled"`
	ConnectionStatus   string                 `json:"connection_status"`
	ReviewStatus       string                 `json:"review_status"`
	OpenToTenant       *bool                  `json:"open_to_tenant"`
	DefaultEnabled     *bool                  `json:"default_enabled"`
	TenantConfigurable *bool                  `json:"tenant_configurable"`
	Config             map[string]interface{} `json:"config"`
}

type PlatformCapabilityMutationRequest struct {
	PlatformCode   string `json:"platform_code"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	CapabilityType string `json:"capability_type"`
	AuthScopeCode  string `json:"auth_scope_code"`
	DataDirection  string `json:"data_direction"`
	Status         string `json:"status"`
	Description    string `json:"description"`
}

type TenantConnectionCreateRequest struct {
	TenantID        *uint64  `json:"tenant_id"`
	ProviderAppCode string   `json:"provider_app_code"`
	ConnectionName  string   `json:"connection_name"`
	AuthSubjectType string   `json:"auth_subject_type"`
	AuthSubjectID   string   `json:"auth_subject_id"`
	AuthSubjectName string   `json:"auth_subject_name"`
	AuthScope       []string `json:"auth_scope"`
}

type OAuthStartRequest struct {
	TenantID        *uint64  `json:"tenant_id"`
	ProviderAppCode string   `json:"provider_app_code"`
	RedirectURI     string   `json:"redirect_uri"`
	Scopes          []string `json:"scopes"`
}

type OAuthStartResult struct {
	ProviderAppCode string `json:"provider_app_code"`
	State           string `json:"state"`
	AuthURL         string `json:"auth_url"`
	ExpiresAt       string `json:"expires_at"`
	Message         string `json:"message"`
}

type OAuthCallbackRequest struct {
	ProviderAppCode string
	State           string
	Code            string
	Error           string
}

type OAuthCallbackResult struct {
	Status          string `json:"status"`
	TenantID        uint64 `json:"tenant_id"`
	ProviderAppCode string `json:"provider_app_code"`
	Message         string `json:"message"`
	ConnectionID    uint64 `json:"connection_id,omitempty"`
	TokenExpiresAt  string `json:"token_expires_at,omitempty"`
}

type oauthTokenExchangeResult struct {
	CredentialRef   string   `json:"credential_ref"`
	AuthSubjectType string   `json:"auth_subject_type"`
	AuthSubjectID   string   `json:"auth_subject_id"`
	AuthSubjectName string   `json:"auth_subject_name"`
	Scope           []string `json:"scope"`
	ExpiresIn       int64    `json:"expires_in"`
	ExpiresAt       string   `json:"expires_at"`
	AccessToken     string   `json:"access_token"`
	RefreshToken    string   `json:"refresh_token"`
}

type QuotaPolicyMutationRequest struct {
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	QuotaCode          string  `json:"quota_code"`
	QuotaUnit          string  `json:"quota_unit"`
	PeriodType         string  `json:"period_type"`
	DefaultLimit       int64   `json:"default_limit"`
	OverLimitAction    string  `json:"over_limit_action"`
	Status             string  `json:"status"`
	Description        string  `json:"description"`
	ScopeType          string  `json:"scope_type"`
	TenantID           *uint64 `json:"tenant_id"`
	PlatformCode       string  `json:"platform_code"`
	ProviderAppCode    string  `json:"provider_app_code"`
	TenantConnectionID *uint64 `json:"tenant_connection_id"`
	OverrideLimit      *int64  `json:"override_limit"`
	Priority           int     `json:"priority"`
}

type APICallQuotaConsumeRequest struct {
	TenantID           *uint64 `json:"tenant_id"`
	TenantConnectionID *uint64 `json:"tenant_connection_id"`
	Amount             int64   `json:"amount"`
}

type APICallQuotaConsumeResult struct {
	TenantID           uint64  `json:"tenant_id"`
	TenantConnectionID *uint64 `json:"tenant_connection_id,omitempty"`
	QuotaCode          string  `json:"quota_code"`
	PeriodKey          string  `json:"period_key"`
	UsedAmount         int64   `json:"used_amount"`
	LimitedCount       int64   `json:"limited_count"`
}

type SyncRecordQuotaConsumeRequest struct {
	TenantID           *uint64 `json:"tenant_id"`
	TenantConnectionID uint64  `json:"tenant_connection_id"`
	Amount             int64   `json:"amount"`
}

type GatewayInvokeRequest struct {
	TenantConnectionID uint64            `json:"tenant_connection_id"`
	Method             string            `json:"method"`
	Path               string            `json:"path"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
}

type GatewayInvokeResult struct {
	RequestID      string `json:"request_id"`
	Status         string `json:"status"`
	HTTPStatus     int    `json:"http_status"`
	DurationMS     int    `json:"duration_ms"`
	RequestDigest  string `json:"request_digest"`
	ResponseDigest string `json:"response_digest"`
	Message        string `json:"message"`
}

type ConnectivityCheckRequest struct {
	Target string `json:"target"`
}

type WebhookReceiveRequest struct {
	ProviderAppCode string
	Timestamp       string
	Signature       string
	IdempotencyKey  string
	EventType       string
	Body            []byte
}

type WebhookReceiveResult struct {
	EventID        uint64 `json:"event_id"`
	Status         string `json:"status"`
	Duplicate      bool   `json:"duplicate"`
	IdempotencyKey string `json:"idempotency_key"`
	ReceivedAt     string `json:"received_at"`
	Message        string `json:"message"`
}

type WebhookProcessResult struct {
	Scanned    int `json:"scanned"`
	Processed  int `json:"processed"`
	Retrying   int `json:"retrying"`
	DeadLetter int `json:"dead_letter"`
}

type SyncJobProcessResult struct {
	Scanned   int `json:"scanned"`
	Completed int `json:"completed"`
	Queued    int `json:"queued"`
	Retrying  int `json:"retrying"`
	Failed    int `json:"failed"`
}

type SyncProcessor interface {
	Process(ctx context.Context, req SyncProcessRequest) (SyncProcessBatch, error)
}

type SyncProcessRequest struct {
	Job        models.IntegrationSyncJob
	Connection models.IntegrationTenantConnection
	App        models.IntegrationProviderApp
	Cursor     string
}

type SyncProcessBatch struct {
	Records    []SyncProcessRecord
	NextCursor string
}

type SyncProcessRecord struct {
	ExternalID string
	Payload    map[string]interface{}
	Cursor     string
}

type HTTPSyncProcessor struct {
	Client *http.Client
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
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	RowCount    int    `json:"row_count"`
	Content     []byte `json:"-"`
}

type APICallLogEntry struct {
	TenantID           *uint64
	TenantConnectionID *uint64
	PlatformID         *uint64
	ProviderAppID      *uint64
	RequestID          string
	TraceID            string
	CallType           string
	Method             string
	Endpoint           string
	Status             string
	HTTPStatus         int
	DurationMS         int
	RequestDigest      string
	ResponseDigest     string
	ErrorCode          string
	ErrorMessage       string
}

func (s *Service) Overview(ctx context.Context) (Overview, error) {
	if s.repo == nil {
		return Overview{}, errors.New("integration center repository is not configured")
	}
	counts, err := s.repo.Counts(ctx)
	if err != nil {
		return Overview{}, err
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return Overview{}, err
	}
	alerts, _, err := s.repo.ListAlerts(ctx, repositories.ListOptions{Limit: 3})
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

func (s *Service) CreatePlatform(ctx context.Context, userID uint64, meta RequestMeta, req PlatformMutationRequest) (repositories.PlatformSummary, error) {
	if s.repo == nil {
		return repositories.PlatformSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.PlatformSummary{}, err
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
		CreatedBy:         &v.UserID,
		UpdatedBy:         &v.UserID,
	})
	if err != nil {
		return repositories.PlatformSummary{}, err
	}
	s.audit(ctx, v, meta, "create_platform", "创建接入平台 "+platform.PlatformName, map[string]interface{}{"platform_code": platform.PlatformCode})
	return platformToSummary(platform), nil
}

func (s *Service) UpdatePlatform(ctx context.Context, userID uint64, meta RequestMeta, code string, req PlatformMutationRequest) (repositories.PlatformSummary, error) {
	if s.repo == nil {
		return repositories.PlatformSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.PlatformSummary{}, err
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
		"updated_by":          &v.UserID,
	}
	platform, err := s.repo.UpdatePlatform(ctx, normalizePlatformCode(code), patch)
	if err != nil {
		return repositories.PlatformSummary{}, err
	}
	s.audit(ctx, v, meta, "update_platform", "更新接入平台 "+platform.PlatformName, map[string]interface{}{"platform_code": platform.PlatformCode})
	return platformToSummary(platform), nil
}

func (s *Service) CreateProviderApp(ctx context.Context, userID uint64, meta RequestMeta, req ProviderAppMutationRequest) (repositories.ProviderAppSummary, error) {
	if s.repo == nil {
		return repositories.ProviderAppSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.ProviderAppSummary{}, err
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
		CreatedBy:     &v.UserID,
		UpdatedBy:     &v.UserID,
	})
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "服务商应用保存失败")
	}
	s.audit(ctx, v, meta, "create_provider_app", "创建服务商应用 "+app.AppName, map[string]interface{}{"app_code": app.AppCode, "platform_code": platform.PlatformCode})
	return providerAppToSummary(app, platform), nil
}

func (s *Service) UpdateProviderApp(ctx context.Context, userID uint64, meta RequestMeta, code string, req ProviderAppMutationRequest) (repositories.ProviderAppSummary, error) {
	if s.repo == nil {
		return repositories.ProviderAppSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.ProviderAppSummary{}, err
	}
	if err := validateProviderAppRequest(req, false); err != nil {
		return repositories.ProviderAppSummary{}, err
	}
	var platform models.IntegrationPlatform
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
		"updated_by":     &v.UserID,
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
	s.audit(ctx, v, meta, "update_provider_app", "更新服务商应用 "+app.AppName, map[string]interface{}{"app_code": app.AppCode, "platform_id": app.PlatformID})
	return providerAppToSummary(app, platform), nil
}

func (s *Service) RotateProviderAppCredential(ctx context.Context, userID uint64, meta RequestMeta, code string, req ProviderAppCredentialRotateRequest) (repositories.ProviderAppSummary, error) {
	if s.repo == nil {
		return repositories.ProviderAppSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.ProviderAppSummary{}, err
	}
	credentialRef, err := normalizeCredentialRef(req.CredentialRef)
	if err != nil {
		return repositories.ProviderAppSummary{}, err
	}
	app, err := s.repo.GetProviderAppByCode(ctx, normalizePlatformCode(code))
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "服务商应用不存在")
	}
	platform, err := s.repo.GetPlatformByID(ctx, app.PlatformID)
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "接入平台不存在")
	}
	app, err = s.repo.UpdateProviderApp(ctx, normalizePlatformCode(code), map[string]interface{}{
		"credential_ref": &credentialRef,
		"updated_by":     &v.UserID,
	})
	if err != nil {
		return repositories.ProviderAppSummary{}, humanizeRepoError(err, "服务商应用凭证轮换失败")
	}
	s.audit(ctx, v, meta, "rotate_provider_app_credential", "轮换服务商应用凭证引用", map[string]interface{}{"app_code": app.AppCode, "credential_ref": maskCredentialRef(credentialRef)})
	return redactProviderAppSummarySecret(providerAppToSummary(app, platform)), nil
}

func (s *Service) CreatePlatformCapability(ctx context.Context, userID uint64, meta RequestMeta, req PlatformCapabilityMutationRequest) (repositories.PlatformCapabilitySummary, error) {
	if s.repo == nil {
		return repositories.PlatformCapabilitySummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, err
	}
	if err := validatePlatformCapabilityRequest(req, true); err != nil {
		return repositories.PlatformCapabilitySummary{}, err
	}
	platform, err := s.repo.GetPlatformByCode(ctx, normalizePlatformCode(req.PlatformCode))
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "接入平台不存在")
	}
	capability, err := s.repo.CreatePlatformCapability(ctx, models.IntegrationPlatformCapability{
		PlatformID:     platform.ID,
		CapabilityCode: normalizePlatformCode(req.Code),
		CapabilityName: strings.TrimSpace(req.Name),
		CapabilityType: defaultString(req.CapabilityType, "api"),
		AuthScopeCode:  optionalString(req.AuthScopeCode),
		DataDirection:  normalizeDataDirection(req.DataDirection),
		Status:         normalizeEnabledStatus(req.Status),
		Description:    optionalString(req.Description),
		CreatedBy:      &v.UserID,
		UpdatedBy:      &v.UserID,
	})
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "平台能力保存失败")
	}
	s.audit(ctx, v, meta, "create_platform_capability", "创建平台能力 "+capability.CapabilityName, map[string]interface{}{"platform_code": platform.PlatformCode, "capability_code": capability.CapabilityCode})
	return platformCapabilityToSummary(capability, platform), nil
}

func (s *Service) UpdatePlatformCapability(ctx context.Context, userID uint64, meta RequestMeta, id uint64, req PlatformCapabilityMutationRequest) (repositories.PlatformCapabilitySummary, error) {
	if s.repo == nil {
		return repositories.PlatformCapabilitySummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, err
	}
	if err := validatePlatformCapabilityRequest(req, false); err != nil {
		return repositories.PlatformCapabilitySummary{}, err
	}
	patch := map[string]interface{}{
		"capability_name": strings.TrimSpace(req.Name),
		"capability_type": defaultString(req.CapabilityType, "api"),
		"auth_scope_code": optionalString(req.AuthScopeCode),
		"data_direction":  normalizeDataDirection(req.DataDirection),
		"status":          normalizeEnabledStatus(req.Status),
		"description":     optionalString(req.Description),
		"updated_by":      &v.UserID,
	}
	if strings.TrimSpace(req.PlatformCode) != "" {
		platform, err := s.repo.GetPlatformByCode(ctx, normalizePlatformCode(req.PlatformCode))
		if err != nil {
			return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "接入平台不存在")
		}
		patch["platform_id"] = platform.ID
	}
	capability, err := s.repo.UpdatePlatformCapability(ctx, id, patch)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "平台能力不存在")
	}
	platform, err := s.repo.GetPlatformByID(ctx, capability.PlatformID)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "接入平台不存在")
	}
	s.audit(ctx, v, meta, "update_platform_capability", "更新平台能力 "+capability.CapabilityName, map[string]interface{}{"id": capability.ID, "capability_code": capability.CapabilityCode})
	return platformCapabilityToSummary(capability, platform), nil
}

func (s *Service) DisablePlatformCapability(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (repositories.PlatformCapabilitySummary, error) {
	if s.repo == nil {
		return repositories.PlatformCapabilitySummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, err
	}
	capability, err := s.repo.GetPlatformCapability(ctx, id)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "平台能力不存在")
	}
	if capability.Status == "disabled" {
		return repositories.PlatformCapabilitySummary{}, errors.New("平台能力已停用")
	}
	capability, err = s.repo.UpdatePlatformCapability(ctx, id, map[string]interface{}{"status": "disabled", "updated_by": &v.UserID})
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "平台能力停用失败")
	}
	platform, err := s.repo.GetPlatformByID(ctx, capability.PlatformID)
	if err != nil {
		return repositories.PlatformCapabilitySummary{}, humanizeRepoError(err, "接入平台不存在")
	}
	s.audit(ctx, v, meta, "disable_platform_capability", "停用平台能力 "+capability.CapabilityName, map[string]interface{}{"id": capability.ID, "capability_code": capability.CapabilityCode})
	return platformCapabilityToSummary(capability, platform), nil
}

func (s *Service) UpdateAppCapability(ctx context.Context, userID uint64, meta RequestMeta, id uint64, req AppCapabilityPatchRequest) (models.IntegrationProviderAppCapability, error) {
	if s.repo == nil {
		return models.IntegrationProviderAppCapability{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return models.IntegrationProviderAppCapability{}, err
	}
	current, err := s.repo.GetAppCapability(ctx, id)
	if err != nil {
		return models.IntegrationProviderAppCapability{}, humanizeRepoError(err, "应用能力不存在")
	}
	patch := map[string]interface{}{}
	if req.Enabled != nil {
		patch["enabled"] = *req.Enabled
	}
	if strings.TrimSpace(req.ConnectionStatus) != "" {
		status := normalizeConnectionStatus(req.ConnectionStatus)
		if status == "" {
			return models.IntegrationProviderAppCapability{}, errors.New("连接状态仅支持 connected、pending、failed、paused、not_connected")
		}
		patch["connection_status"] = status
	}
	if strings.TrimSpace(req.ReviewStatus) != "" {
		review := normalizeReviewStatus(req.ReviewStatus)
		if review == "" {
			return models.IntegrationProviderAppCapability{}, errors.New("审核状态仅支持 pending、approved、rejected")
		}
		patch["review_status"] = review
	}
	config, changed, err := appCapabilityConfigPatch(current.Config, req)
	if err != nil {
		return models.IntegrationProviderAppCapability{}, err
	}
	if changed {
		patch["config"] = config
	}
	if len(patch) == 0 {
		return models.IntegrationProviderAppCapability{}, errors.New("没有可更新的能力字段")
	}
	patch["updated_by"] = &v.UserID
	result, err := s.repo.UpdateAppCapability(ctx, id, patch)
	if err != nil {
		return models.IntegrationProviderAppCapability{}, humanizeRepoError(err, "应用能力不存在")
	}
	s.audit(ctx, v, meta, "update_app_capability", "更新应用能力连接", map[string]interface{}{"id": result.ID, "enabled": result.Enabled, "config": result.Config})
	return result, nil
}

func (s *Service) CreateTenantConnection(ctx context.Context, userID uint64, meta RequestMeta, req TenantConnectionCreateRequest) (models.IntegrationTenantConnection, error) {
	if s.repo == nil {
		return models.IntegrationTenantConnection{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	targetTenantID := v.TenantID
	if v.IsPlatformAdmin && req.TenantID != nil && *req.TenantID > 0 {
		targetTenantID = *req.TenantID
	}
	if !v.IsPlatformAdmin && req.TenantID != nil && *req.TenantID != v.TenantID {
		return models.IntegrationTenantConnection{}, ErrForbidden
	}
	if err := validateTenantConnectionCreateRequest(req); err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	app, err := s.repo.GetProviderAppByCode(ctx, normalizePlatformCode(req.ProviderAppCode))
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "服务商应用不存在")
	}
	if !v.IsPlatformAdmin && (!app.TenantVisible || app.Status == "disabled" || app.Status == "draft") {
		return models.IntegrationTenantConnection{}, ErrForbidden
	}
	if err := s.requireTenantAuthorizationFeature(ctx, targetTenantID); err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	if err := s.requireConnectionQuota(ctx, targetTenantID); err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	scopeRaw, _ := json.Marshal(nonNilRows(req.AuthScope))
	connection, err := s.repo.CreateTenantConnection(ctx, models.IntegrationTenantConnection{
		TenantID:         targetTenantID,
		PlatformID:       app.PlatformID,
		ProviderAppID:    app.ID,
		ConnectionName:   strings.TrimSpace(req.ConnectionName),
		AuthSubjectType:  defaultString(strings.TrimSpace(req.AuthSubjectType), "tenant"),
		AuthSubjectID:    strings.TrimSpace(req.AuthSubjectID),
		AuthSubjectName:  strings.TrimSpace(req.AuthSubjectName),
		AuthScope:        string(scopeRaw),
		AuthStatus:       "pending",
		ConnectionStatus: "inactive",
		TokenStatus:      "unknown",
		CreatedBy:        &v.UserID,
		UpdatedBy:        &v.UserID,
	})
	if err != nil {
		return models.IntegrationTenantConnection{}, errors.New("租户连接保存失败，请检查授权主体是否已存在")
	}
	s.audit(ctx, v, meta, "create_tenant_connection", "创建租户第三方授权连接", map[string]interface{}{"connection_id": connection.ID, "tenant_id": connection.TenantID, "provider_app_code": app.AppCode})
	return connection, nil
}

func (s *Service) StartOAuthAuthorization(ctx context.Context, userID uint64, meta RequestMeta, req OAuthStartRequest) (OAuthStartResult, error) {
	if s.repo == nil {
		return OAuthStartResult{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return OAuthStartResult{}, err
	}
	targetTenantID := v.TenantID
	if v.IsPlatformAdmin && req.TenantID != nil && *req.TenantID > 0 {
		targetTenantID = *req.TenantID
	}
	if !v.IsPlatformAdmin && req.TenantID != nil && *req.TenantID != v.TenantID {
		return OAuthStartResult{}, ErrForbidden
	}
	if err := validateOAuthStartRequest(req); err != nil {
		return OAuthStartResult{}, err
	}
	app, err := s.repo.GetProviderAppByCode(ctx, normalizePlatformCode(req.ProviderAppCode))
	if err != nil {
		return OAuthStartResult{}, humanizeRepoError(err, "服务商应用不存在")
	}
	if !strings.EqualFold(app.AuthMode, "OAuth2") {
		return OAuthStartResult{}, errors.New("服务商应用不是 OAuth2 授权模式")
	}
	if !v.IsPlatformAdmin && (!app.TenantVisible || app.Status == "disabled" || app.Status == "draft") {
		return OAuthStartResult{}, ErrForbidden
	}
	if err := s.requireTenantAuthorizationFeature(ctx, targetTenantID); err != nil {
		return OAuthStartResult{}, err
	}
	if err := s.requireConnectionQuota(ctx, targetTenantID); err != nil {
		return OAuthStartResult{}, err
	}
	state, err := randomOAuthState()
	if err != nil {
		return OAuthStartResult{}, err
	}
	scopeRaw, _ := json.Marshal(nonNilRows(req.Scopes))
	expiresAt := time.Now().Add(10 * time.Minute)
	row, err := s.repo.CreateOAuthState(ctx, models.IntegrationOAuthState{
		State:         state,
		TenantID:      targetTenantID,
		ProviderAppID: app.ID,
		PlatformID:    app.PlatformID,
		RedirectURI:   strings.TrimSpace(req.RedirectURI),
		Scopes:        string(scopeRaw),
		Status:        "pending",
		ExpiresAt:     expiresAt,
		CreatedBy:     &v.UserID,
	})
	if err != nil {
		return OAuthStartResult{}, errors.New("OAuth state 创建失败")
	}
	authURL, message := oauthAuthorizeURLFor(app, row.RedirectURI, req.Scopes, row.State)
	s.audit(ctx, v, meta, "start_oauth_authorization", "发起第三方 OAuth 授权", map[string]interface{}{"tenant_id": row.TenantID, "provider_app_code": app.AppCode, "state_id": row.ID})
	return OAuthStartResult{
		ProviderAppCode: app.AppCode,
		State:           row.State,
		AuthURL:         authURL,
		ExpiresAt:       row.ExpiresAt.Format(time.RFC3339),
		Message:         message,
	}, nil
}

func (s *Service) HandleOAuthCallback(ctx context.Context, req OAuthCallbackRequest) (OAuthCallbackResult, error) {
	if s.repo == nil {
		return OAuthCallbackResult{}, errors.New("integration center repository is not configured")
	}
	appCode := normalizePlatformCode(req.ProviderAppCode)
	if appCode == "" {
		return OAuthCallbackResult{}, errors.New("服务商应用编码不能为空")
	}
	if strings.TrimSpace(req.State) == "" {
		return OAuthCallbackResult{}, errors.New("OAuth state 不能为空")
	}
	if strings.TrimSpace(req.Code) == "" && strings.TrimSpace(req.Error) == "" {
		return OAuthCallbackResult{}, errors.New("OAuth callback 缺少 code 或 error")
	}
	app, err := s.repo.GetProviderAppByCode(ctx, appCode)
	if err != nil {
		return OAuthCallbackResult{}, humanizeRepoError(err, "服务商应用不存在")
	}
	state, err := s.repo.GetOAuthState(ctx, strings.TrimSpace(req.State))
	if err != nil {
		return OAuthCallbackResult{}, humanizeRepoError(err, "OAuth state 不存在或已失效")
	}
	if state.ProviderAppID != app.ID {
		return OAuthCallbackResult{}, errors.New("OAuth state 与服务商应用不匹配")
	}
	if state.Status != "pending" {
		return OAuthCallbackResult{}, errors.New("OAuth state 已被使用")
	}
	if time.Now().After(state.ExpiresAt) {
		return OAuthCallbackResult{}, errors.New("OAuth state 已过期")
	}
	consumed, err := s.repo.ConsumeOAuthState(ctx, state.ID)
	if err != nil {
		return OAuthCallbackResult{}, errors.New("OAuth state 已被使用")
	}
	if strings.TrimSpace(req.Error) != "" {
		return OAuthCallbackResult{
			Status:          "failed",
			TenantID:        consumed.TenantID,
			ProviderAppCode: app.AppCode,
			Message:         "第三方 OAuth 授权失败：" + strings.TrimSpace(req.Error),
		}, nil
	}
	tokenResult, configured, err := exchangeOAuthCode(ctx, app, consumed.RedirectURI, strings.TrimSpace(req.Code))
	if err != nil {
		return OAuthCallbackResult{}, err
	}
	if configured {
		connection, err := s.upsertAuthorizedConnectionFromOAuth(ctx, app, consumed, tokenResult)
		if err != nil {
			return OAuthCallbackResult{}, err
		}
		expiresAt := ""
		if connection.TokenExpiresAt != nil {
			expiresAt = connection.TokenExpiresAt.Format(time.RFC3339)
		}
		return OAuthCallbackResult{
			Status:          "authorized",
			TenantID:        consumed.TenantID,
			ProviderAppCode: app.AppCode,
			Message:         "OAuth 授权完成，token 已通过凭证引用绑定到租户连接",
			ConnectionID:    connection.ID,
			TokenExpiresAt:  expiresAt,
		}, nil
	}
	return OAuthCallbackResult{
		Status:          "code_received",
		TenantID:        consumed.TenantID,
		ProviderAppCode: app.AppCode,
		Message:         "OAuth 授权码已接收，等待 token exchange 配置后换取令牌引用",
	}, nil
}

func (s *Service) Connectors(ctx context.Context) ([]IntegrationConnector, error) {
	if s.repo == nil {
		return nil, errors.New("integration center repository is not configured")
	}
	rows, _, err := s.repo.ListPlatforms(ctx, repositories.ListOptions{Limit: 20})
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

func (s *Service) upsertAuthorizedConnectionFromOAuth(ctx context.Context, app models.IntegrationProviderApp, state models.IntegrationOAuthState, tokenResult oauthTokenExchangeResult) (models.IntegrationTenantConnection, error) {
	if strings.TrimSpace(tokenResult.CredentialRef) == "" {
		return models.IntegrationTenantConnection{}, errors.New("OAuth token exchange 未返回 credential_ref，拒绝保存明文 token")
	}
	if strings.TrimSpace(tokenResult.AccessToken) != "" && strings.TrimSpace(tokenResult.CredentialRef) == "" {
		return models.IntegrationTenantConnection{}, errors.New("OAuth token exchange 返回明文 token 但缺少 credential_ref")
	}
	subjectID := strings.TrimSpace(tokenResult.AuthSubjectID)
	if subjectID == "" {
		return models.IntegrationTenantConnection{}, errors.New("OAuth token exchange 未返回授权主体 ID")
	}
	subjectType := defaultString(tokenResult.AuthSubjectType, "tenant")
	subjectName := defaultString(tokenResult.AuthSubjectName, subjectID)
	scope := tokenResult.Scope
	if len(scope) == 0 && strings.TrimSpace(state.Scopes) != "" {
		_ = json.Unmarshal([]byte(state.Scopes), &scope)
	}
	scopeRaw, _ := json.Marshal(nonNilRows(scope))
	now := time.Now()
	expiresAt := tokenExpiresAt(tokenResult, now)
	credentialRef := strings.TrimSpace(tokenResult.CredentialRef)
	existing, err := s.repo.GetTenantConnectionBySubject(ctx, state.TenantID, app.PlatformID, app.ID, subjectType, subjectID)
	if err == nil {
		return s.repo.UpdateTenantConnection(ctx, existing.ID, map[string]interface{}{
			"connection_name":      defaultString(existing.ConnectionName, app.AppName+"-"+subjectName),
			"auth_subject_name":    subjectName,
			"auth_scope":           string(scopeRaw),
			"auth_status":          "authorized",
			"connection_status":    "connected",
			"token_status":         "valid",
			"token_credential_ref": &credentialRef,
			"authorized_at":        &now,
			"token_expires_at":     expiresAt,
			"last_error_at":        nil,
			"last_error_message":   nil,
			"updated_by":           state.CreatedBy,
		})
	}
	if !errors.Is(err, repositories.ErrConnectionNotFound) {
		return models.IntegrationTenantConnection{}, err
	}
	connectionName := app.AppName + "-" + subjectName
	return s.repo.CreateTenantConnection(ctx, models.IntegrationTenantConnection{
		TenantID:           state.TenantID,
		PlatformID:         app.PlatformID,
		ProviderAppID:      app.ID,
		ConnectionName:     connectionName,
		AuthSubjectType:    subjectType,
		AuthSubjectID:      subjectID,
		AuthSubjectName:    subjectName,
		AuthScope:          string(scopeRaw),
		AuthStatus:         "authorized",
		ConnectionStatus:   "connected",
		TokenStatus:        "valid",
		TokenCredentialRef: &credentialRef,
		AuthorizedAt:       &now,
		TokenExpiresAt:     expiresAt,
		CreatedBy:          state.CreatedBy,
		UpdatedBy:          state.CreatedBy,
	})
}

func (s *Service) Platforms(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, nil)
	rows, total, err := s.repo.ListPlatforms(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("platforms", nonNilRows(rows), total, opts), nil
}

func (s *Service) Workspace(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, nil)
	rows, total, err := s.repo.ListProviderApps(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	for index := range rows {
		rows[index] = redactProviderAppSummarySecret(rows[index])
	}
	return section("workspace", nonNilRows(rows), total, opts), nil
}

func (s *Service) PlatformCapabilities(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, nil)
	rows, total, err := s.repo.ListPlatformCapabilities(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("platform-capabilities", nonNilRows(rows), total, opts), nil
}

func (s *Service) AppCapabilities(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, nil)
	rows, total, err := s.repo.ListProviderAppCapabilities(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	rows = normalizeProviderAppCapabilitySummaries(rows)
	return section("app-capabilities", nonNilRows(rows), total, opts), nil
}

func (s *Service) TenantConnections(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, tenantScope(v, req.TenantID))
	rows, total, err := s.repo.ListTenantConnections(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("tenant-connections", nonNilRows(rows), total, opts), nil
}

func (s *Service) SyncMonitor(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, tenantScope(v, req.TenantID))
	rows, total, err := s.repo.ListSyncJobs(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("sync-monitor", nonNilRows(rows), total, opts), nil
}

func (s *Service) Quota(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, nil)
	rows, total, err := s.repo.ListQuotaPolicies(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("quota", nonNilRows(rows), total, opts), nil
}

func (s *Service) QuotaUsages(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, tenantScope(v, req.TenantID))
	rows, total, err := s.repo.ListQuotaUsages(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("quota-usages", nonNilRows(rows), total, opts), nil
}

func (s *Service) Alerts(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, tenantScope(v, req.TenantID))
	rows, total, err := s.repo.ListAlerts(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("alerts", nonNilRows(rows), total, opts), nil
}

func (s *Service) Logs(ctx context.Context, userID uint64, req PageRequest) (SectionSummary, error) {
	if s.repo == nil {
		return SectionSummary{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return SectionSummary{}, err
	}
	opts := listOptions(req, tenantScope(v, req.TenantID))
	rows, total, err := s.repo.ListAPICallLogs(ctx, opts)
	if err != nil {
		return SectionSummary{}, err
	}
	return section("logs", nonNilRows(rows), total, opts), nil
}

func (s *Service) PlatformDetail(ctx context.Context, userID uint64, code string) (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return nil, err
	}
	platform, err := s.repo.GetPlatformByCode(ctx, normalizePlatformCode(code))
	if err != nil {
		return nil, humanizeRepoError(err, "接入平台不存在")
	}
	opts := repositories.ListOptions{Limit: 100, PlatformCode: platform.PlatformCode}
	apps, _, err := s.repo.ListProviderApps(ctx, opts)
	if err != nil {
		return nil, err
	}
	for index := range apps {
		apps[index] = redactProviderAppSummarySecret(apps[index])
	}
	capabilities, _, err := s.repo.ListPlatformCapabilities(ctx, opts)
	if err != nil {
		return nil, err
	}
	connections, _, err := s.repo.ListTenantConnections(ctx, opts)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"platform":     platform,
		"apps":         nonNilRows(apps),
		"capabilities": nonNilRows(capabilities),
		"connections":  nonNilRows(connections),
	}, nil
}

func (s *Service) ProviderAppDetail(ctx context.Context, userID uint64, code string) (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("integration center repository is not configured")
	}
	if _, err := s.requirePlatform(ctx, userID); err != nil {
		return nil, err
	}
	app, err := s.repo.GetProviderAppByCode(ctx, normalizePlatformCode(code))
	if err != nil {
		return nil, humanizeRepoError(err, "服务商应用不存在")
	}
	platform, _ := s.repo.GetPlatformByID(ctx, app.PlatformID)
	opts := repositories.ListOptions{Limit: 100, ProviderAppCode: app.AppCode}
	capabilities, _, err := s.repo.ListProviderAppCapabilities(ctx, opts)
	if err != nil {
		return nil, err
	}
	connections, _, err := s.repo.ListTenantConnections(ctx, opts)
	if err != nil {
		return nil, err
	}
	logs, err := s.repo.ListAPICallLogsByProviderApp(ctx, app.ID, 20)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"app":          redactProviderAppSecret(app),
		"platform":     platform,
		"capabilities": nonNilRows(capabilities),
		"connections":  nonNilRows(connections),
		"logs":         sanitizeAPICallLogs(logs),
	}, nil
}

func (s *Service) TenantConnectionDetail(ctx context.Context, userID uint64, id uint64) (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("integration center repository is not configured")
	}
	if _, err := s.requireConnectionAccess(ctx, userID, id); err != nil {
		return nil, err
	}
	connection, err := s.repo.GetTenantConnection(ctx, id)
	if err != nil {
		return nil, humanizeRepoError(err, "租户连接不存在")
	}
	platform, _ := s.repo.GetPlatformByID(ctx, connection.PlatformID)
	app, _ := s.repo.GetProviderAppByID(ctx, connection.ProviderAppID)
	capabilities, err := s.repo.ListTenantCapabilitiesByConnection(ctx, connection.ID)
	if err != nil {
		return nil, err
	}
	jobs, err := s.repo.ListSyncJobsByConnection(ctx, connection.ID, 20)
	if err != nil {
		return nil, err
	}
	quotas, err := s.repo.ListQuotaUsagesByConnection(ctx, connection.ID, 20)
	if err != nil {
		return nil, err
	}
	logs, err := s.repo.ListAPICallLogsByConnection(ctx, connection.ID, 20)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"connection":   redactTenantConnectionSecret(connection),
		"platform":     platform,
		"app":          redactProviderAppSecret(app),
		"capabilities": nonNilRows(capabilities),
		"sync_jobs":    nonNilRows(jobs),
		"quota_usages": nonNilRows(quotas),
		"logs":         sanitizeAPICallLogs(logs),
	}, nil
}

func (s *Service) SyncJobDetail(ctx context.Context, userID uint64, id uint64) (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("integration center repository is not configured")
	}
	if _, err := s.requireSyncJobAccess(ctx, userID, id); err != nil {
		return nil, err
	}
	job, err := s.repo.GetSyncJob(ctx, id)
	if err != nil {
		return nil, humanizeRepoError(err, "同步任务不存在")
	}
	connection, _ := s.repo.GetTenantConnection(ctx, job.TenantConnectionID)
	logs, _ := s.repo.ListAPICallLogsByConnection(ctx, job.TenantConnectionID, 20)
	return map[string]interface{}{
		"sync_job":   job,
		"connection": redactTenantConnectionSecret(connection),
		"logs":       sanitizeAPICallLogs(logs),
	}, nil
}

func (s *Service) LogDetail(ctx context.Context, userID uint64, id uint64) (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return nil, err
	}
	log, err := s.repo.GetAPICallLog(ctx, id)
	if err != nil {
		return nil, humanizeRepoError(err, "调用日志不存在")
	}
	if !v.IsPlatformAdmin && (log.TenantID == nil || *log.TenantID != v.TenantID) {
		return nil, ErrForbidden
	}
	return map[string]interface{}{
		"log": sanitizeAPICallLog(log),
		"security": map[string]string{
			"body_policy": "请求和响应正文不返回，仅返回 digest、状态、耗时和安全错误信息。",
		},
	}, nil
}

func (s *Service) RefreshTenantConnection(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationTenantConnection, error) {
	if s.repo == nil {
		return models.IntegrationTenantConnection{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requireConnectionAccess(ctx, userID, id)
	if err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	connection, err := s.repo.GetTenantConnection(ctx, id)
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "租户连接不存在")
	}
	app, err := s.repo.GetProviderAppByID(ctx, connection.ProviderAppID)
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "服务商应用不存在")
	}
	if connection.TokenCredentialRef == nil || strings.TrimSpace(*connection.TokenCredentialRef) == "" {
		return models.IntegrationTenantConnection{}, errors.New("租户连接缺少 token_credential_ref，无法刷新")
	}
	tokenResult, err := refreshOAuthCredential(ctx, app, *connection.TokenCredentialRef)
	if err != nil {
		failed, updateErr := s.markTenantConnectionTokenFailure(ctx, connection, v, err)
		if updateErr != nil {
			return models.IntegrationTenantConnection{}, updateErr
		}
		return failed, err
	}
	now := time.Now()
	credentialRef := strings.TrimSpace(tokenResult.CredentialRef)
	result, err := s.repo.UpdateTenantConnection(ctx, id, map[string]interface{}{
		"auth_status":          "authorized",
		"connection_status":    "connected",
		"token_status":         "valid",
		"token_credential_ref": &credentialRef,
		"last_sync_at":         &now,
		"last_error_at":        nil,
		"last_error_message":   nil,
		"token_expires_at":     tokenExpiresAt(tokenResult, now),
		"updated_by":           &v.UserID,
	})
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "租户连接不存在")
	}
	s.audit(ctx, v, meta, "refresh_tenant_connection", "刷新租户连接授权", map[string]interface{}{"connection_id": result.ID, "tenant_id": result.TenantID})
	return result, nil
}

func (s *Service) markTenantConnectionTokenFailure(ctx context.Context, connection models.IntegrationTenantConnection, v viewer, cause error) (models.IntegrationTenantConnection, error) {
	now := time.Now()
	message := cause.Error()
	failed, err := s.repo.UpdateTenantConnection(ctx, connection.ID, map[string]interface{}{
		"connection_status":  "failed",
		"token_status":       "refresh_failed",
		"last_error_at":      &now,
		"last_error_message": &message,
		"updated_by":         &v.UserID,
	})
	if err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	title := "第三方授权 Token 刷新失败"
	_, _ = s.repo.CreateAlert(ctx, models.IntegrationAlert{
		TenantID:           &connection.TenantID,
		TenantConnectionID: &connection.ID,
		PlatformID:         &connection.PlatformID,
		ProviderAppID:      &connection.ProviderAppID,
		AlertType:          "token_refresh",
		Severity:           "critical",
		Status:             "open",
		Title:              title,
		Message:            &message,
		FirstSeenAt:        now,
		LastSeenAt:         now,
	})
	return failed, nil
}

func (s *Service) SetTenantConnectionStatus(ctx context.Context, userID uint64, meta RequestMeta, id uint64, paused bool) (models.IntegrationTenantConnection, error) {
	if s.repo == nil {
		return models.IntegrationTenantConnection{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requireConnectionAccess(ctx, userID, id)
	if err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	connection, err := s.repo.GetTenantConnection(ctx, id)
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "租户连接不存在")
	}
	status := "connected"
	if paused {
		status = "paused"
	}
	if connection.ConnectionStatus == status {
		return models.IntegrationTenantConnection{}, errors.New("租户连接已处于目标状态")
	}
	result, err := s.repo.UpdateTenantConnection(ctx, id, map[string]interface{}{"connection_status": status, "updated_by": &v.UserID})
	if err != nil {
		return models.IntegrationTenantConnection{}, humanizeRepoError(err, "租户连接不存在")
	}
	s.audit(ctx, v, meta, "set_tenant_connection_status", "更新租户连接状态", map[string]interface{}{"connection_id": result.ID, "tenant_id": result.TenantID, "status": status})
	return result, nil
}

func (s *Service) RetryTenantConnection(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationSyncJob, error) {
	if s.repo == nil {
		return models.IntegrationSyncJob{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requireConnectionAccess(ctx, userID, id)
	if err != nil {
		return models.IntegrationSyncJob{}, err
	}
	connection, err := s.repo.UpdateTenantConnection(ctx, id, map[string]interface{}{"connection_status": "connected", "updated_by": &v.UserID})
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
		CreatedBy:          &v.UserID,
	}
	result, err := s.repo.CreateSyncJob(ctx, job)
	if err != nil {
		return models.IntegrationSyncJob{}, humanizeRepoError(err, "同步重试任务创建失败")
	}
	s.audit(ctx, v, meta, "retry_tenant_connection", "创建租户连接重试同步任务", map[string]interface{}{"connection_id": connection.ID, "sync_job_id": result.ID, "tenant_id": result.TenantID})
	return result, nil
}

func (s *Service) RetrySyncJob(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationSyncJob, error) {
	now := time.Now()
	return s.updateSyncJob(ctx, userID, meta, id, "retry_sync_job", map[string]interface{}{
		"status":        "retrying",
		"started_at":    &now,
		"finished_at":   nil,
		"error_code":    nil,
		"error_message": nil,
	})
}

func (s *Service) PauseSyncJob(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationSyncJob, error) {
	return s.updateSyncJob(ctx, userID, meta, id, "pause_sync_job", map[string]interface{}{"status": "paused"})
}

func (s *Service) ResumeSyncJob(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationSyncJob, error) {
	now := time.Now()
	return s.updateSyncJob(ctx, userID, meta, id, "resume_sync_job", map[string]interface{}{"status": "running", "started_at": &now})
}

func (s *Service) updateSyncJob(ctx context.Context, userID uint64, meta RequestMeta, id uint64, action string, patch map[string]interface{}) (models.IntegrationSyncJob, error) {
	if s.repo == nil {
		return models.IntegrationSyncJob{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requireSyncJobAccess(ctx, userID, id)
	if err != nil {
		return models.IntegrationSyncJob{}, err
	}
	current, err := s.repo.GetSyncJob(ctx, id)
	if err != nil {
		return models.IntegrationSyncJob{}, humanizeRepoError(err, "同步任务不存在")
	}
	if err := validateSyncJobTransition(current.Status, patchStatus(patch), action); err != nil {
		return models.IntegrationSyncJob{}, err
	}
	result, err := s.repo.UpdateSyncJob(ctx, id, patch)
	if err != nil {
		return models.IntegrationSyncJob{}, humanizeRepoError(err, "同步任务不存在")
	}
	s.audit(ctx, v, meta, action, "更新同步任务状态", map[string]interface{}{"sync_job_id": result.ID, "tenant_id": result.TenantID, "status": result.Status})
	return result, nil
}

func (s *Service) CreateQuotaPolicy(ctx context.Context, userID uint64, meta RequestMeta, req QuotaPolicyMutationRequest) (models.IntegrationQuotaPolicy, error) {
	if s.repo == nil {
		return models.IntegrationQuotaPolicy{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	if err := validateQuotaPolicyRequest(req, true); err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	policyModel := quotaPolicyFromRequest(req)
	policyModel.CreatedBy = &v.UserID
	policyModel.UpdatedBy = &v.UserID
	policy, err := s.repo.CreateQuotaPolicy(ctx, policyModel)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略保存失败")
	}
	if quotaPolicyHasBinding(req) {
		binding, err := s.quotaBindingFromRequest(ctx, policy.ID, v.UserID, req)
		if err != nil {
			return models.IntegrationQuotaPolicy{}, err
		}
		if _, err := s.repo.UpsertQuotaBinding(ctx, binding); err != nil {
			return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略绑定保存失败")
		}
	}
	s.audit(ctx, v, meta, "create_quota_policy", "创建集成配额策略 "+policy.PolicyName, map[string]interface{}{"policy_code": policy.PolicyCode, "quota_code": policy.QuotaCode, "scope_type": normalizedQuotaScope(req.ScopeType)})
	return policy, nil
}

func (s *Service) UpdateQuotaPolicy(ctx context.Context, userID uint64, meta RequestMeta, code string, req QuotaPolicyMutationRequest) (models.IntegrationQuotaPolicy, error) {
	if s.repo == nil {
		return models.IntegrationQuotaPolicy{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, err
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
		"updated_by":        &v.UserID,
	}
	policy, err := s.repo.UpdateQuotaPolicy(ctx, normalizePlatformCode(code), patch)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略不存在")
	}
	if quotaPolicyHasBinding(req) {
		binding, err := s.quotaBindingFromRequest(ctx, policy.ID, v.UserID, req)
		if err != nil {
			return models.IntegrationQuotaPolicy{}, err
		}
		if _, err := s.repo.UpsertQuotaBinding(ctx, binding); err != nil {
			return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略绑定保存失败")
		}
	}
	s.audit(ctx, v, meta, "update_quota_policy", "更新集成配额策略 "+policy.PolicyName, map[string]interface{}{"policy_code": policy.PolicyCode, "quota_code": policy.QuotaCode, "scope_type": normalizedQuotaScope(req.ScopeType)})
	return policy, nil
}

func (s *Service) SetQuotaPolicyStatus(ctx context.Context, userID uint64, meta RequestMeta, code string, enabled bool) (models.IntegrationQuotaPolicy, error) {
	if s.repo == nil {
		return models.IntegrationQuotaPolicy{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	status := "disabled"
	if enabled {
		status = "enabled"
	}
	policy, err := s.repo.UpdateQuotaPolicy(ctx, normalizePlatformCode(code), map[string]interface{}{"status": status, "updated_by": &v.UserID})
	if err != nil {
		return models.IntegrationQuotaPolicy{}, humanizeRepoError(err, "配额策略不存在")
	}
	s.audit(ctx, v, meta, "set_quota_policy_status", "更新集成配额策略状态", map[string]interface{}{"policy_code": policy.PolicyCode, "status": status})
	return policy, nil
}

func (s *Service) ConsumeAPICallQuota(ctx context.Context, userID uint64, meta RequestMeta, req APICallQuotaConsumeRequest) (APICallQuotaConsumeResult, error) {
	if s.repo == nil {
		return APICallQuotaConsumeResult{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return APICallQuotaConsumeResult{}, err
	}
	tenantID := v.TenantID
	var tenantConnectionID *uint64
	if req.TenantConnectionID != nil && *req.TenantConnectionID > 0 {
		connection, err := s.repo.GetTenantConnection(ctx, *req.TenantConnectionID)
		if err != nil {
			return APICallQuotaConsumeResult{}, humanizeRepoError(err, "租户连接不存在")
		}
		if !v.IsPlatformAdmin && connection.TenantID != v.TenantID {
			return APICallQuotaConsumeResult{}, ErrForbidden
		}
		tenantID = connection.TenantID
		tenantConnectionID = &connection.ID
	} else if v.IsPlatformAdmin {
		if req.TenantID == nil || *req.TenantID == 0 {
			return APICallQuotaConsumeResult{}, errors.New("平台管理员消费 API 配额时必须指定 tenant_id 或 tenant_connection_id")
		}
		tenantID = *req.TenantID
	} else if req.TenantID != nil && *req.TenantID != v.TenantID {
		return APICallQuotaConsumeResult{}, ErrForbidden
	}
	limitValue, err := s.quotaLimit(ctx, tenantID, tenantConnectionID, "integration_api_calls_daily")
	if err != nil {
		return APICallQuotaConsumeResult{}, err
	}
	amount := req.Amount
	if amount <= 0 {
		amount = 1
	}
	usage, err := s.repo.ConsumeQuota(ctx, tenantID, tenantConnectionID, "integration_api_calls_daily", time.Now().Format("20060102"), amount, limitValue)
	if errors.Is(err, repositories.ErrQuotaExceeded) {
		s.audit(ctx, v, meta, "api_call_quota_exceeded", "第三方 API 日调用配额超限", map[string]interface{}{"tenant_id": tenantID, "tenant_connection_id": tenantConnectionID, "quota_code": "integration_api_calls_daily"})
		return quotaConsumeResult(usage), ErrQuotaExceeded
	}
	if err != nil {
		return APICallQuotaConsumeResult{}, err
	}
	return quotaConsumeResult(usage), nil
}

func (s *Service) ConsumeSyncRecordQuota(ctx context.Context, userID uint64, meta RequestMeta, req SyncRecordQuotaConsumeRequest) (APICallQuotaConsumeResult, error) {
	if s.repo == nil {
		return APICallQuotaConsumeResult{}, errors.New("integration center repository is not configured")
	}
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return APICallQuotaConsumeResult{}, err
	}
	if req.TenantConnectionID == 0 {
		return APICallQuotaConsumeResult{}, errors.New("tenant_connection_id 不能为空")
	}
	connection, err := s.repo.GetTenantConnection(ctx, req.TenantConnectionID)
	if err != nil {
		return APICallQuotaConsumeResult{}, humanizeRepoError(err, "租户连接不存在")
	}
	if !v.IsPlatformAdmin && connection.TenantID != v.TenantID {
		return APICallQuotaConsumeResult{}, ErrForbidden
	}
	if v.IsPlatformAdmin && req.TenantID != nil && *req.TenantID > 0 && *req.TenantID != connection.TenantID {
		return APICallQuotaConsumeResult{}, errors.New("tenant_id 与 tenant_connection_id 不匹配")
	}
	limitValue, err := s.quotaLimit(ctx, connection.TenantID, &connection.ID, "integration_sync_records_daily")
	if err != nil {
		return APICallQuotaConsumeResult{}, err
	}
	amount := req.Amount
	if amount <= 0 {
		amount = 1
	}
	usage, err := s.repo.ConsumeQuota(ctx, connection.TenantID, &connection.ID, "integration_sync_records_daily", time.Now().Format("20060102"), amount, limitValue)
	if errors.Is(err, repositories.ErrQuotaExceeded) {
		s.audit(ctx, v, meta, "sync_record_quota_exceeded", "第三方同步日记录配额超限", map[string]interface{}{"tenant_id": connection.TenantID, "tenant_connection_id": connection.ID, "quota_code": "integration_sync_records_daily"})
		return quotaConsumeResult(usage), ErrQuotaExceeded
	}
	if err != nil {
		return APICallQuotaConsumeResult{}, err
	}
	return quotaConsumeResult(usage), nil
}

func (s *Service) InvokeGateway(ctx context.Context, userID uint64, meta RequestMeta, req GatewayInvokeRequest) (GatewayInvokeResult, error) {
	if s.repo == nil {
		return GatewayInvokeResult{}, errors.New("integration center repository is not configured")
	}
	_, err := s.requireConnectionAccess(ctx, userID, req.TenantConnectionID)
	if err != nil {
		return GatewayInvokeResult{}, err
	}
	connection, err := s.repo.GetTenantConnection(ctx, req.TenantConnectionID)
	if err != nil {
		return GatewayInvokeResult{}, humanizeRepoError(err, "租户连接不存在")
	}
	if connection.TokenCredentialRef == nil || strings.TrimSpace(*connection.TokenCredentialRef) == "" {
		return GatewayInvokeResult{}, errors.New("租户连接缺少 token_credential_ref，拒绝调用第三方 API")
	}
	if connection.ConnectionStatus != "connected" || connection.TokenStatus != "valid" {
		return GatewayInvokeResult{}, errors.New("租户连接未处于可调用状态")
	}
	app, err := s.repo.GetProviderAppByID(ctx, connection.ProviderAppID)
	if err != nil {
		return GatewayInvokeResult{}, humanizeRepoError(err, "服务商应用不存在")
	}
	requestID := gatewayRequestID(meta)
	started := time.Now()
	method, endpoint, err := gatewayRequestTarget(app, req)
	if err != nil {
		return GatewayInvokeResult{}, err
	}
	bodyBytes := []byte(req.Body)
	if len(bodyBytes) > 1<<20 {
		return GatewayInvokeResult{}, errors.New("第三方 API 请求体超过 1MB 限制")
	}
	requestDigest := digestHex(bodyBytes)
	tenantID := connection.TenantID
	connectionID := connection.ID
	if _, err := s.ConsumeAPICallQuota(ctx, userID, meta, APICallQuotaConsumeRequest{TenantID: &tenantID, TenantConnectionID: &connectionID, Amount: 1}); err != nil {
		duration := int(time.Since(started).Milliseconds())
		s.recordGatewayLog(ctx, connection, meta, requestID, method, endpoint, "limited", 429, duration, requestDigest, "", "quota_exceeded", err.Error())
		return GatewayInvokeResult{}, err
	}
	outReq, err := http.NewRequestWithContext(ctx, method, endpoint, strings.NewReader(req.Body))
	if err != nil {
		return GatewayInvokeResult{}, errors.New("第三方 API 请求创建失败")
	}
	for key, value := range req.Headers {
		if safeGatewayHeader(key) {
			outReq.Header.Set(key, value)
		}
	}
	outReq.Header.Set("X-Integration-Credential-Ref", strings.TrimSpace(*connection.TokenCredentialRef))
	if outReq.Header.Get("Content-Type") == "" && req.Body != "" {
		outReq.Header.Set("Content-Type", "application/json")
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(outReq)
	duration := int(time.Since(started).Milliseconds())
	if err != nil {
		s.recordGatewayLog(ctx, connection, meta, requestID, method, endpoint, "failed", 0, duration, requestDigest, "", "gateway_request_failed", err.Error())
		return GatewayInvokeResult{}, errors.New("第三方 API 调用失败")
	}
	defer resp.Body.Close()
	responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	responseDigest := digestHex(responseBody)
	status := "success"
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status = "failed"
	}
	if readErr != nil {
		status = "failed"
	}
	errorCode := ""
	errorMessage := ""
	if status == "failed" {
		errorCode = "third_party_api_failed"
		errorMessage = "第三方 API 返回失败状态"
		if readErr != nil {
			errorCode = "third_party_response_read_failed"
			errorMessage = "第三方 API 响应读取失败"
		}
	}
	s.recordGatewayLog(ctx, connection, meta, requestID, method, endpoint, status, resp.StatusCode, duration, requestDigest, responseDigest, errorCode, errorMessage)
	if status == "failed" {
		return GatewayInvokeResult{
			RequestID:      requestID,
			Status:         status,
			HTTPStatus:     resp.StatusCode,
			DurationMS:     duration,
			RequestDigest:  requestDigest,
			ResponseDigest: responseDigest,
			Message:        errorMessage,
		}, errors.New(errorMessage)
	}
	return GatewayInvokeResult{
		RequestID:      requestID,
		Status:         status,
		HTTPStatus:     resp.StatusCode,
		DurationMS:     duration,
		RequestDigest:  requestDigest,
		ResponseDigest: responseDigest,
		Message:        "第三方 API 调用完成，响应正文已隔离，仅返回摘要",
	}, nil
}

func (s *Service) ProcessAlert(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationAlert, error) {
	return s.updateAlert(ctx, userID, meta, id, "process_alert", map[string]interface{}{"status": "processing"})
}

func (s *Service) ResolveAlert(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationAlert, error) {
	now := time.Now()
	return s.updateAlert(ctx, userID, meta, id, "resolve_alert", map[string]interface{}{"status": "resolved", "resolved_at": &now})
}

func (s *Service) IgnoreAlert(ctx context.Context, userID uint64, meta RequestMeta, id uint64) (models.IntegrationAlert, error) {
	return s.updateAlert(ctx, userID, meta, id, "ignore_alert", map[string]interface{}{"status": "ignored"})
}

func (s *Service) updateAlert(ctx context.Context, userID uint64, meta RequestMeta, id uint64, action string, patch map[string]interface{}) (models.IntegrationAlert, error) {
	if s.repo == nil {
		return models.IntegrationAlert{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requireAlertAccess(ctx, userID, id)
	if err != nil {
		return models.IntegrationAlert{}, err
	}
	current, err := s.repo.GetAlert(ctx, id)
	if err != nil {
		return models.IntegrationAlert{}, humanizeRepoError(err, "异常记录不存在")
	}
	if err := validateAlertTransition(current.Status, patchStatus(patch), action); err != nil {
		return models.IntegrationAlert{}, err
	}
	patch["handled_by"] = &v.UserID
	result, err := s.repo.UpdateAlert(ctx, id, patch)
	if err != nil {
		return models.IntegrationAlert{}, humanizeRepoError(err, "异常记录不存在")
	}
	s.audit(ctx, v, meta, action, "处理集成异常 "+result.Title, map[string]interface{}{"alert_id": result.ID, "tenant_id": result.TenantID, "status": result.Status})
	return result, nil
}

func (s *Service) CheckConnectivity(ctx context.Context, userID uint64, meta RequestMeta, req ConnectivityCheckRequest) (ConnectivityCheckResult, error) {
	if s.repo == nil {
		return ConnectivityCheckResult{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return ConnectivityCheckResult{}, err
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
	message := "平台运行态一致性检测通过"
	if counts.OpenAlerts > 0 {
		status = "warning"
		message = "运行态检测完成，存在待处理异常"
	}
	result := ConnectivityCheckResult{
		Target:       target,
		Status:       status,
		CheckedAt:    time.Now().Format(time.RFC3339),
		OpenAlerts:   counts.OpenAlerts,
		RunningSyncs: counts.RunningSyncJobs,
		Message:      message,
	}
	s.audit(ctx, v, meta, "connectivity_check", "执行集成中心连通性检测", map[string]interface{}{"target": target, "status": status})
	return result, nil
}

func (s *Service) ExportLogs(ctx context.Context, userID uint64, meta RequestMeta) (LogExportResult, error) {
	if s.repo == nil {
		return LogExportResult{}, errors.New("integration center repository is not configured")
	}
	v, err := s.requirePlatform(ctx, userID)
	if err != nil {
		return LogExportResult{}, err
	}
	rows, _, err := s.repo.ListAPICallLogs(ctx, repositories.ListOptions{Limit: 10000, SortBy: "called_at", SortOrder: "desc"})
	if err != nil {
		return LogExportResult{}, err
	}
	now := time.Now()
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"request_id", "trace_id", "tenant_id", "connection_id", "call_type", "method", "endpoint", "status", "http_status", "duration_ms", "called_at", "error_code", "error_message"}); err != nil {
		return LogExportResult{}, err
	}
	for _, row := range sanitizeAPICallLogs(rows) {
		if err := writer.Write([]string{
			row.RequestID,
			stringValue(row.TraceID),
			uint64PtrValue(row.TenantID),
			uint64PtrValue(row.TenantConnectionID),
			row.CallType,
			stringValue(row.Method),
			stringValue(row.Endpoint),
			row.Status,
			intPtrValue(row.HTTPStatus),
			strconv.Itoa(row.DurationMS),
			timeValue(row.CalledAt),
			stringValue(row.ErrorCode),
			stringValue(row.ErrorMessage),
		}); err != nil {
			return LogExportResult{}, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return LogExportResult{}, err
	}
	result := LogExportResult{
		Filename:    fmt.Sprintf("integration-call-logs-%s.csv", now.Format("20060102150405")),
		ContentType: "text/csv; charset=utf-8",
		RowCount:    len(rows),
		Content:     buf.Bytes(),
	}
	s.audit(ctx, v, meta, "export_logs", "导出集成调用日志", map[string]interface{}{"filename": result.Filename, "row_count": result.RowCount})
	return result, nil
}

func (s *Service) ReceiveWebhook(ctx context.Context, req WebhookReceiveRequest) (WebhookReceiveResult, error) {
	if s.repo == nil {
		return WebhookReceiveResult{}, errors.New("integration center repository is not configured")
	}
	start := time.Now()
	app, err := s.repo.GetProviderAppByCode(ctx, normalizePlatformCode(req.ProviderAppCode))
	if err != nil {
		return WebhookReceiveResult{}, humanizeRepoError(err, "服务商应用不存在")
	}
	endpoint := "/api/integration-center/webhooks/" + app.AppCode
	requestID := strings.TrimSpace(req.IdempotencyKey)
	if requestID == "" {
		requestID = gatewayRequestID(RequestMeta{})
	}
	recordWebhookLog := func(status string, httpStatus int, errorCode string, errorMessage string) {
		requestDigest := digestHex(req.Body)
		s.recordAPICallLog(ctx, APICallLogEntry{
			PlatformID:     &app.PlatformID,
			ProviderAppID:  &app.ID,
			RequestID:      requestID,
			TraceID:        requestID,
			CallType:       "webhook",
			Method:         http.MethodPost,
			Endpoint:       endpoint,
			Status:         status,
			HTTPStatus:     httpStatus,
			DurationMS:     int(time.Since(start).Milliseconds()),
			RequestDigest:  requestDigest,
			ResponseDigest: "",
			ErrorCode:      errorCode,
			ErrorMessage:   errorMessage,
		})
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		recordWebhookLog("failed", http.StatusBadRequest, "webhook_idempotency_key_missing", "Webhook 幂等键不能为空")
		return WebhookReceiveResult{}, errors.New("Webhook 幂等键不能为空")
	}
	if err := validateWebhookTimestamp(req.Timestamp); err != nil {
		recordWebhookLog("failed", http.StatusUnauthorized, "webhook_timestamp_invalid", err.Error())
		return WebhookReceiveResult{}, err
	}
	secret, err := webhookSecretFor(app)
	if err != nil {
		recordWebhookLog("failed", http.StatusInternalServerError, "webhook_secret_missing", err.Error())
		return WebhookReceiveResult{}, err
	}
	signature := strings.TrimSpace(req.Signature)
	if !validWebhookSignature(secret, req.Timestamp, req.Body, signature) {
		recordWebhookLog("failed", http.StatusUnauthorized, "webhook_signature_invalid", "Webhook 签名校验失败")
		return WebhookReceiveResult{}, errors.New("Webhook 签名校验失败")
	}
	payload := strings.TrimSpace(string(req.Body))
	if payload == "" {
		payload = "{}"
	}
	if !json.Valid([]byte(payload)) {
		recordWebhookLog("failed", http.StatusBadRequest, "webhook_payload_invalid", "Webhook payload 必须是合法 JSON")
		return WebhookReceiveResult{}, errors.New("Webhook payload 必须是合法 JSON")
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		eventType = "integration.webhook"
	}
	digest := sha256.Sum256(req.Body)
	event, err := s.repo.CreateWebhookEvent(ctx, models.IntegrationWebhookEvent{
		ProviderAppID:  app.ID,
		PlatformID:     app.PlatformID,
		EventType:      eventType,
		IdempotencyKey: strings.TrimSpace(req.IdempotencyKey),
		Signature:      signature,
		PayloadDigest:  hex.EncodeToString(digest[:]),
		Payload:        payload,
		Status:         "received",
	})
	if err != nil {
		if errors.Is(err, repositories.ErrWebhookEventExists) {
			recordWebhookLog("success", http.StatusOK, "", "")
			return WebhookReceiveResult{
				Status:         "duplicated",
				Duplicate:      true,
				IdempotencyKey: strings.TrimSpace(req.IdempotencyKey),
				ReceivedAt:     time.Now().Format(time.RFC3339),
				Message:        "Webhook 事件已接收，幂等忽略重复请求",
			}, nil
		}
		recordWebhookLog("failed", http.StatusInternalServerError, "webhook_event_save_failed", err.Error())
		return WebhookReceiveResult{}, err
	}
	recordWebhookLog("success", http.StatusOK, "", "")
	return WebhookReceiveResult{
		EventID:        event.ID,
		Status:         event.Status,
		Duplicate:      false,
		IdempotencyKey: event.IdempotencyKey,
		ReceivedAt:     event.ReceivedAt.Format(time.RFC3339),
		Message:        "Webhook 事件已接收",
	}, nil
}

func (s *Service) ProcessDueWebhookEvents(ctx context.Context, limit int) (WebhookProcessResult, error) {
	if s.repo == nil {
		return WebhookProcessResult{}, errors.New("integration center repository is not configured")
	}
	events, err := s.repo.ListDueWebhookEvents(ctx, limit)
	if err != nil {
		return WebhookProcessResult{}, err
	}
	result := WebhookProcessResult{Scanned: len(events)}
	for _, event := range events {
		if err := processWebhookPayload(event); err != nil {
			updated, updateErr := s.scheduleWebhookRetry(ctx, event, err)
			if updateErr != nil {
				return result, updateErr
			}
			if updated.Status == "dead_letter" {
				result.DeadLetter++
			} else {
				result.Retrying++
			}
			continue
		}
		now := time.Now()
		if _, err := s.repo.UpdateWebhookEvent(ctx, event.ID, map[string]interface{}{
			"status":        "processed",
			"processed_at":  &now,
			"next_retry_at": nil,
			"error_message": nil,
		}); err != nil {
			return result, err
		}
		result.Processed++
	}
	return result, nil
}

func (s *Service) ArchiveExpiredAPICallLogs(ctx context.Context, limit int) (int64, error) {
	if s.repo == nil {
		return 0, errors.New("integration center repository is not configured")
	}
	days := apiCallLogRetentionDays()
	cutoff := time.Now().AddDate(0, 0, -days)
	bucket := fmt.Sprintf("older_than_%d_days", days)
	return s.repo.ArchiveAPICallLogsBefore(ctx, cutoff, bucket, limit)
}

func (s *Service) ProcessDueSyncJobs(ctx context.Context, limit int) (SyncJobProcessResult, error) {
	if s.repo == nil {
		return SyncJobProcessResult{}, errors.New("integration center repository is not configured")
	}
	jobs, err := s.repo.ListDueSyncJobs(ctx, limit)
	if err != nil {
		return SyncJobProcessResult{}, err
	}
	result := SyncJobProcessResult{Scanned: len(jobs)}
	for _, job := range jobs {
		connection, err := s.repo.GetTenantConnection(ctx, job.TenantConnectionID)
		if err != nil {
			updated, updateErr := s.scheduleSyncJobRetry(ctx, job, humanizeRepoError(err, "租户连接不存在"))
			if updateErr != nil {
				return result, updateErr
			}
			if updated.Status == "failed" {
				result.Failed++
			} else {
				result.Retrying++
			}
			continue
		}
		app, err := s.repo.GetProviderAppByID(ctx, connection.ProviderAppID)
		if err != nil {
			updated, updateErr := s.scheduleSyncJobRetry(ctx, job, humanizeRepoError(err, "服务商应用不存在"))
			if updateErr != nil {
				return result, updateErr
			}
			if updated.Status == "failed" {
				result.Failed++
			} else {
				result.Retrying++
			}
			continue
		}
		processor := s.syncProcessor
		if processor == nil {
			processor = HTTPSyncProcessor{}
		}
		startedAt := firstTime(job.StartedAt, time.Now())
		batch, err := processor.Process(ctx, SyncProcessRequest{
			Job:        job,
			Connection: connection,
			App:        app,
			Cursor:     stringValue(job.CursorValue),
		})
		if err != nil {
			updated, updateErr := s.scheduleSyncJobRetry(ctx, job, err)
			if updateErr != nil {
				return result, updateErr
			}
			if updated.Status == "failed" {
				result.Failed++
			} else {
				result.Retrying++
			}
			s.recordSyncProcessorLog(ctx, connection, job, app, "failed", 0, err)
			continue
		}
		records := syncProcessRecordsToModels(job, batch)
		amount := int64(len(records))
		limitValue, err := s.quotaLimit(ctx, job.TenantID, &job.TenantConnectionID, "integration_sync_records_daily")
		if err != nil {
			return result, err
		}
		if amount > 0 {
			quotaResult, err := s.repo.ConsumeQuota(ctx, job.TenantID, &job.TenantConnectionID, "integration_sync_records_daily", time.Now().Format("20060102"), amount, limitValue)
			if errors.Is(err, repositories.ErrQuotaExceeded) {
				next := nextDailyQuotaWindow()
				message := "第三方同步日记录配额超限，任务已排队等待下一个配额窗口"
				if _, err := s.repo.UpdateSyncJob(ctx, job.ID, map[string]interface{}{
					"status":        "queued",
					"next_retry_at": &next,
					"error_code":    optionalString("quota_exceeded"),
					"error_message": &message,
				}); err != nil {
					return result, err
				}
				_ = quotaResult
				result.Queued++
				continue
			}
			if err != nil {
				updated, updateErr := s.scheduleSyncJobRetry(ctx, job, err)
				if updateErr != nil {
					return result, updateErr
				}
				if updated.Status == "failed" {
					result.Failed++
				} else {
					result.Retrying++
				}
				continue
			}
		}
		written, err := s.repo.UpsertSyncRecords(ctx, records)
		if err != nil {
			updated, updateErr := s.scheduleSyncJobRetry(ctx, job, err)
			if updateErr != nil {
				return result, updateErr
			}
			if updated.Status == "failed" {
				result.Failed++
			} else {
				result.Retrying++
			}
			s.recordSyncProcessorLog(ctx, connection, job, app, "failed", 0, err)
			continue
		}
		now := time.Now()
		nextCursor := optionalString(batch.NextCursor)
		if _, err := s.repo.UpdateSyncJob(ctx, job.ID, map[string]interface{}{
			"status":        "completed",
			"started_at":    startedAt,
			"finished_at":   &now,
			"total_count":   amount,
			"success_count": written,
			"failed_count":  amount - written,
			"cursor_value":  nextCursor,
			"next_retry_at": nil,
			"error_code":    nil,
			"error_message": nil,
		}); err != nil {
			return result, err
		}
		s.recordSyncProcessorLog(ctx, connection, job, app, "success", int(amount), nil)
		result.Completed++
	}
	return result, nil
}

func (s *Service) scheduleSyncJobRetry(ctx context.Context, job models.IntegrationSyncJob, cause error) (models.IntegrationSyncJob, error) {
	retryCount := job.RetryCount + 1
	status := "retrying"
	var nextRetryAt *time.Time
	if retryCount >= syncJobMaxRetryCount {
		status = "failed"
	} else {
		next := time.Now().Add(webhookRetryDelay(retryCount))
		nextRetryAt = &next
	}
	code := "sync_processing_failed"
	message := cause.Error()
	return s.repo.UpdateSyncJob(ctx, job.ID, map[string]interface{}{
		"status":        status,
		"retry_count":   retryCount,
		"next_retry_at": nextRetryAt,
		"error_code":    &code,
		"error_message": &message,
	})
}

func (s *Service) scheduleWebhookRetry(ctx context.Context, event models.IntegrationWebhookEvent, cause error) (models.IntegrationWebhookEvent, error) {
	retryCount := event.RetryCount + 1
	status := "retrying"
	var nextRetryAt *time.Time
	if retryCount >= webhookMaxRetryCount {
		status = "dead_letter"
	} else {
		next := time.Now().Add(webhookRetryDelay(retryCount))
		nextRetryAt = &next
	}
	message := cause.Error()
	return s.repo.UpdateWebhookEvent(ctx, event.ID, map[string]interface{}{
		"status":        status,
		"retry_count":   retryCount,
		"next_retry_at": nextRetryAt,
		"error_message": &message,
		"processed_at":  nil,
	})
}

func (s *Service) eventsFromLogs(ctx context.Context) ([]IntegrationEvent, error) {
	rows, _, err := s.repo.ListAPICallLogs(ctx, repositories.ListOptions{Limit: 5})
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

func section(name string, items interface{}, total int64, opts repositories.ListOptions) SectionSummary {
	return SectionSummary{
		AppCode: "integration-center",
		Section: name,
		Items:   items,
		Total:   total,
		Skip:    opts.Skip,
		Limit:   opts.Limit,
	}
}

func quotaConsumeResult(usage models.IntegrationQuotaUsage) APICallQuotaConsumeResult {
	return APICallQuotaConsumeResult{
		TenantID:           usage.TenantID,
		TenantConnectionID: usage.TenantConnectionID,
		QuotaCode:          usage.QuotaCode,
		PeriodKey:          usage.PeriodKey,
		UsedAmount:         usage.UsedAmount,
		LimitedCount:       usage.LimitedCount,
	}
}

func listOptions(req PageRequest, tenantID *uint64) repositories.ListOptions {
	return repositories.ListOptions{
		Skip:            req.Skip,
		Limit:           req.Limit,
		Keyword:         req.Keyword,
		Status:          req.Status,
		PlatformCode:    normalizePlatformCode(req.PlatformCode),
		ProviderAppCode: normalizePlatformCode(req.ProviderAppCode),
		TenantID:        tenantID,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		SortBy:          req.SortBy,
		SortOrder:       req.SortOrder,
	}
}

func tenantScope(v viewer, requested *uint64) *uint64 {
	if v.IsPlatformAdmin {
		return requested
	}
	return &v.TenantID
}

func (s *Service) viewer(ctx context.Context, userID uint64) (viewer, error) {
	if userID == 0 {
		return viewer{}, ErrUnauthorized
	}
	user, err := s.repo.UserByID(ctx, userID)
	if err != nil {
		return viewer{}, ErrUnauthorized
	}
	return viewer{UserID: user.ID, TenantID: user.TenantID, IsPlatformAdmin: user.IsPlatformAdmin}, nil
}

func (s *Service) requirePlatform(ctx context.Context, userID uint64) (viewer, error) {
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return viewer{}, err
	}
	if !v.IsPlatformAdmin {
		return viewer{}, ErrForbidden
	}
	return v, nil
}

func (s *Service) requireConnectionAccess(ctx context.Context, userID uint64, id uint64) (viewer, error) {
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return viewer{}, err
	}
	if v.IsPlatformAdmin {
		return v, nil
	}
	connection, err := s.repo.GetTenantConnection(ctx, id)
	if err != nil {
		return viewer{}, humanizeRepoError(err, "租户连接不存在")
	}
	if connection.TenantID != v.TenantID {
		return viewer{}, ErrForbidden
	}
	return v, nil
}

func (s *Service) requireSyncJobAccess(ctx context.Context, userID uint64, id uint64) (viewer, error) {
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return viewer{}, err
	}
	if v.IsPlatformAdmin {
		return v, nil
	}
	job, err := s.repo.GetSyncJob(ctx, id)
	if err != nil {
		return viewer{}, humanizeRepoError(err, "同步任务不存在")
	}
	if job.TenantID != v.TenantID {
		return viewer{}, ErrForbidden
	}
	return v, nil
}

func (s *Service) requireAlertAccess(ctx context.Context, userID uint64, id uint64) (viewer, error) {
	v, err := s.viewer(ctx, userID)
	if err != nil {
		return viewer{}, err
	}
	if v.IsPlatformAdmin {
		return v, nil
	}
	alert, err := s.repo.GetAlert(ctx, id)
	if err != nil {
		return viewer{}, humanizeRepoError(err, "异常记录不存在")
	}
	if alert.TenantID == nil || *alert.TenantID != v.TenantID {
		return viewer{}, ErrForbidden
	}
	return v, nil
}

func (s *Service) requireTenantAuthorizationFeature(ctx context.Context, tenantID uint64) error {
	allowed, err := s.repo.TenantFeatureAllowed(ctx, tenantID, "integration_tenant_authorization")
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("当前套餐不支持功能：integration_tenant_authorization")
	}
	return nil
}

func (s *Service) requireConnectionQuota(ctx context.Context, tenantID uint64) error {
	limit, exists, err := s.repo.TenantQuotaLimit(ctx, tenantID, "integration_connection_count")
	if err != nil {
		return err
	}
	if !exists || limit < 0 {
		return nil
	}
	used, err := s.repo.CountTenantConnections(ctx, tenantID)
	if err != nil {
		return err
	}
	if used+1 > int64(limit) {
		return ErrQuotaExceeded
	}
	return nil
}

func (s *Service) audit(ctx context.Context, v viewer, meta RequestMeta, action string, summary string, detail interface{}) {
	if s.repo == nil {
		return
	}
	appCode := "integration-center"
	tenantID := v.TenantID
	userID := v.UserID
	detailText := ""
	if detail != nil {
		if raw, err := json.Marshal(detail); err == nil {
			detailText = redactSensitiveText(string(raw))
		}
	}
	log := models.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		AppCode:   &appCode,
		Module:    "integration_center",
		Action:    action,
		Summary:   summary,
		Result:    "success",
		CreatedAt: time.Now(),
	}
	if detailText != "" {
		log.Detail = &detailText
	}
	if meta.IP != "" {
		log.IP = &meta.IP
	}
	if meta.UserAgent != "" {
		log.UserAgent = &meta.UserAgent
	}
	if meta.RequestID != "" {
		log.RequestID = &meta.RequestID
	}
	_ = s.repo.CreateAuditLog(ctx, log)
}

func validateWebhookTimestamp(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New("Webhook 时间戳不能为空")
	}
	seconds, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return errors.New("Webhook 时间戳不合法")
	}
	ts := time.Unix(seconds, 0)
	now := time.Now()
	if ts.Before(now.Add(-webhookTimestampSkew)) || ts.After(now.Add(webhookTimestampSkew)) {
		return errors.New("Webhook 时间戳已过期或超出允许窗口")
	}
	return nil
}

func webhookSecretFor(app models.IntegrationProviderApp) (string, error) {
	keys := []string{}
	if app.CredentialRef != nil && strings.TrimSpace(*app.CredentialRef) != "" {
		if secret := credentialSecretFromRef(*app.CredentialRef); secret != "" {
			return secret, nil
		}
		keys = append(keys, strings.TrimSpace(*app.CredentialRef))
	}
	keys = append(keys, "INTEGRATION_WEBHOOK_SECRET_"+envToken(app.AppCode))
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value, nil
		}
	}
	return "", errors.New("Webhook 密钥未配置")
}

func validWebhookSignature(secret string, timestamp string, body []byte, signature string) bool {
	signature = strings.TrimPrefix(strings.TrimSpace(signature), "sha256=")
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.TrimSpace(timestamp)))
	mac.Write([]byte("."))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func processWebhookPayload(event models.IntegrationWebhookEvent) error {
	if !json.Valid([]byte(event.Payload)) {
		return errors.New("Webhook payload JSON 已损坏")
	}
	return nil
}

func webhookRetryDelay(retryCount int) time.Duration {
	switch retryCount {
	case 1:
		return time.Minute
	case 2:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}

func nextDailyQuotaWindow() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
}

func (p HTTPSyncProcessor) Process(ctx context.Context, req SyncProcessRequest) (SyncProcessBatch, error) {
	endpoint := strings.TrimSpace(os.Getenv("INTEGRATION_SYNC_SOURCE_URL_" + envToken(req.App.AppCode)))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("INTEGRATION_SYNC_SOURCE_URL"))
	}
	if endpoint == "" {
		return SyncProcessBatch{}, errors.New("第三方同步源未配置")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return SyncProcessBatch{}, errors.New("第三方同步源地址不合法")
	}
	query := u.Query()
	query.Set("tenant_connection_id", strconv.FormatUint(req.Connection.ID, 10))
	query.Set("capability_code", req.Job.CapabilityCode)
	query.Set("job_type", req.Job.JobType)
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	u.RawQuery = query.Encode()
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return SyncProcessBatch{}, errors.New("第三方同步请求创建失败")
	}
	httpReq.Header.Set("Accept", "application/json")
	if req.Connection.TokenCredentialRef != nil {
		httpReq.Header.Set("X-Integration-Credential-Ref", *req.Connection.TokenCredentialRef)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return SyncProcessBatch{}, errors.New("第三方同步源请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return SyncProcessBatch{}, errors.New("第三方同步源响应读取失败")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SyncProcessBatch{}, fmt.Errorf("第三方同步源返回失败状态: %d", resp.StatusCode)
	}
	batch, err := decodeSyncProcessBatch(body)
	if err != nil {
		return SyncProcessBatch{}, err
	}
	return batch, nil
}

func decodeSyncProcessBatch(raw []byte) (SyncProcessBatch, error) {
	var envelope struct {
		Records    []map[string]interface{} `json:"records"`
		Items      []map[string]interface{} `json:"items"`
		NextCursor string                   `json:"next_cursor"`
		Cursor     string                   `json:"cursor"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && (envelope.Records != nil || envelope.Items != nil) {
		rows := envelope.Records
		if rows == nil {
			rows = envelope.Items
		}
		return syncBatchFromMaps(rows, defaultString(envelope.NextCursor, envelope.Cursor)), nil
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return SyncProcessBatch{}, errors.New("第三方同步源响应不是合法 JSON")
	}
	return syncBatchFromMaps(rows, ""), nil
}

func syncBatchFromMaps(rows []map[string]interface{}, nextCursor string) SyncProcessBatch {
	records := make([]SyncProcessRecord, 0, len(rows))
	for index, row := range rows {
		payload := sanitizeSyncPayload(row)
		externalID := firstSyncString(payload, "external_id", "id", "code", "biz_id")
		if externalID == "" {
			externalID = fmt.Sprintf("row-%d", index+1)
		}
		cursor := firstSyncString(payload, "cursor", "updated_at", "modified_at")
		records = append(records, SyncProcessRecord{ExternalID: externalID, Payload: payload, Cursor: cursor})
	}
	return SyncProcessBatch{Records: records, NextCursor: nextCursor}
}

func syncProcessRecordsToModels(job models.IntegrationSyncJob, batch SyncProcessBatch) []models.IntegrationSyncRecord {
	rows := make([]models.IntegrationSyncRecord, 0, len(batch.Records))
	for _, record := range batch.Records {
		payload := record.Payload
		if payload == nil {
			payload = map[string]interface{}{}
		}
		raw, _ := json.Marshal(payload)
		digest := sha256.Sum256(raw)
		cursor := optionalString(record.Cursor)
		rows = append(rows, models.IntegrationSyncRecord{
			TenantID:           job.TenantID,
			SyncJobID:          job.ID,
			TenantConnectionID: job.TenantConnectionID,
			CapabilityCode:     job.CapabilityCode,
			ExternalID:         record.ExternalID,
			PayloadDigest:      hex.EncodeToString(digest[:]),
			Payload:            string(raw),
			Status:             "written",
			CursorValue:        cursor,
		})
	}
	return rows
}

func sanitizeSyncPayload(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for key, value := range row {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential") || strings.Contains(lower, "password") {
			out[key] = "***"
			continue
		}
		out[key] = value
	}
	return out
}

func firstSyncString(row map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := row[key]; ok {
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func (s *Service) recordSyncProcessorLog(ctx context.Context, connection models.IntegrationTenantConnection, job models.IntegrationSyncJob, app models.IntegrationProviderApp, status string, recordCount int, cause error) {
	tenantID := connection.TenantID
	connectionID := connection.ID
	platformID := connection.PlatformID
	providerAppID := connection.ProviderAppID
	endpoint := fmt.Sprintf("sync://%s/%s", app.AppCode, job.CapabilityCode)
	requestDigestRaw := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", job.ID, job.CapabilityCode, stringValue(job.CursorValue))))
	responseDigestRaw := sha256.Sum256([]byte(strconv.Itoa(recordCount)))
	errorCode := ""
	errorMessage := ""
	if cause != nil {
		errorCode = "sync_processor_failed"
		errorMessage = cause.Error()
	}
	s.recordAPICallLog(ctx, APICallLogEntry{
		TenantID:           &tenantID,
		TenantConnectionID: &connectionID,
		PlatformID:         &platformID,
		ProviderAppID:      &providerAppID,
		RequestID:          fmt.Sprintf("integration-sync-%d-%d", job.ID, time.Now().UnixNano()),
		CallType:           "data_sync",
		Method:             "GET",
		Endpoint:           endpoint,
		Status:             status,
		HTTPStatus:         200,
		RequestDigest:      hex.EncodeToString(requestDigestRaw[:]),
		ResponseDigest:     hex.EncodeToString(responseDigestRaw[:]),
		ErrorCode:          errorCode,
		ErrorMessage:       errorMessage,
	})
}

func (s *Service) recordGatewayLog(ctx context.Context, connection models.IntegrationTenantConnection, meta RequestMeta, requestID string, method string, endpoint string, status string, httpStatus int, durationMS int, requestDigest string, responseDigest string, errorCode string, errorMessage string) {
	tenantID := connection.TenantID
	connectionID := connection.ID
	platformID := connection.PlatformID
	providerAppID := connection.ProviderAppID
	traceID := gatewayTraceID(meta, requestID)
	s.recordAPICallLog(ctx, APICallLogEntry{
		TenantID:           &tenantID,
		TenantConnectionID: &connectionID,
		PlatformID:         &platformID,
		ProviderAppID:      &providerAppID,
		RequestID:          requestID,
		TraceID:            traceID,
		CallType:           "third_party_api",
		Method:             method,
		Endpoint:           endpoint,
		Status:             status,
		HTTPStatus:         httpStatus,
		DurationMS:         durationMS,
		RequestDigest:      requestDigest,
		ResponseDigest:     responseDigest,
		ErrorCode:          errorCode,
		ErrorMessage:       errorMessage,
	})
}

func (s *Service) recordAPICallLog(ctx context.Context, entry APICallLogEntry) {
	_ = s.apiCallLogRecorder.RecordAPICallLog(ctx, entry)
}

func normalizeAPICallLogStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "success", "failed", "limited":
		return strings.TrimSpace(status)
	default:
		return "failed"
	}
}

func gatewayRequestID(meta RequestMeta) string {
	if strings.TrimSpace(meta.RequestID) != "" {
		return strings.TrimSpace(meta.RequestID)
	}
	state, err := randomOAuthState()
	if err != nil {
		return fmt.Sprintf("integration-api-%d", time.Now().UnixNano())
	}
	return "integration-api-" + state
}

func gatewayTraceID(meta RequestMeta, requestID string) string {
	if strings.TrimSpace(meta.TraceID) != "" {
		return strings.TrimSpace(meta.TraceID)
	}
	return strings.TrimSpace(requestID)
}

func redactProviderAppSecret(app models.IntegrationProviderApp) models.IntegrationProviderApp {
	if app.CredentialRef != nil {
		masked := maskCredentialRef(*app.CredentialRef)
		app.CredentialRef = &masked
	}
	return app
}

func redactProviderAppSummarySecret(app repositories.ProviderAppSummary) repositories.ProviderAppSummary {
	if app.CredentialRef != nil {
		masked := maskCredentialRef(*app.CredentialRef)
		app.CredentialRef = &masked
	}
	return app
}

func redactTenantConnectionSecret(connection models.IntegrationTenantConnection) models.IntegrationTenantConnection {
	if connection.TokenCredentialRef != nil {
		masked := maskCredentialRef(*connection.TokenCredentialRef)
		connection.TokenCredentialRef = &masked
	}
	return connection
}

func sanitizeAPICallLogs(logs []models.IntegrationAPICallLog) []models.IntegrationAPICallLog {
	if logs == nil {
		return []models.IntegrationAPICallLog{}
	}
	sanitized := make([]models.IntegrationAPICallLog, 0, len(logs))
	for _, log := range logs {
		sanitized = append(sanitized, sanitizeAPICallLog(log))
	}
	return sanitized
}

func sanitizeAPICallLog(log models.IntegrationAPICallLog) models.IntegrationAPICallLog {
	if log.Endpoint == nil {
		return sanitizeAPICallLogMessage(log)
	}
	safeEndpoint := *log.Endpoint
	if parsed, err := url.Parse(*log.Endpoint); err == nil {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		safeEndpoint = parsed.String()
	}
	log.Endpoint = &safeEndpoint
	return sanitizeAPICallLogMessage(log)
}

func sanitizeAPICallLogMessage(log models.IntegrationAPICallLog) models.IntegrationAPICallLog {
	if log.ErrorMessage != nil {
		message := redactSensitiveText(*log.ErrorMessage)
		log.ErrorMessage = &message
	}
	return log
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func uint64PtrValue(value *uint64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(*value, 10)
}

func intPtrValue(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func timeValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func maskCredentialRef(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 12 {
		return "***"
	}
	return value[:6] + "..." + value[len(value)-4:]
}

func redactSensitiveText(value string) string {
	for _, token := range []string{"credential_ref=", "token=", "access_token=", "refresh_token=", "secret=", "client_secret="} {
		value = redactSensitiveTokenValue(value, token)
	}
	return value
}

func redactSensitiveTokenValue(value string, token string) string {
	lower := strings.ToLower(value)
	var out strings.Builder
	position := 0
	for {
		index := strings.Index(lower[position:], token)
		if index < 0 {
			out.WriteString(value[position:])
			break
		}
		index += position
		start := index + len(token)
		end := start
		for end < len(value) && !strings.ContainsRune(" \t\r\n&?,;\"'", rune(value[end])) {
			end++
		}
		out.WriteString(value[position:start])
		out.WriteString("***")
		position = end
	}
	return out.String()
}

func normalizeCredentialRef(value string) (string, error) {
	ref := strings.TrimSpace(value)
	if ref == "" {
		return "", errors.New("credential_ref 不能为空")
	}
	if strings.ContainsAny(ref, " \t\r\n") {
		return "", errors.New("credential_ref 不能包含空白字符")
	}
	if strings.HasPrefix(ref, "vault://") || strings.HasPrefix(ref, "secret://") || strings.HasPrefix(ref, "env://") {
		return ref, nil
	}
	if strings.HasPrefix(ref, "INTEGRATION_") {
		return ref, nil
	}
	return "", errors.New("credential_ref 必须使用 vault://、secret://、env:// 或 INTEGRATION_ 环境变量引用")
}

func credentialSecretFromRef(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	if strings.HasPrefix(ref, "env://") {
		return strings.TrimSpace(os.Getenv(strings.TrimPrefix(ref, "env://")))
	}
	if strings.HasPrefix(ref, "INTEGRATION_") {
		return strings.TrimSpace(os.Getenv(ref))
	}
	return ""
}

func patchStatus(patch map[string]interface{}) string {
	if value, ok := patch["status"].(string); ok {
		return value
	}
	return ""
}

func validateSyncJobTransition(current string, next string, action string) error {
	if next == "" {
		return nil
	}
	if !allowedStatus(next, "pending", "running", "completed", "failed", "queued", "retrying", "paused") {
		return errors.New("同步任务状态仅支持 pending、running、completed、failed、queued、retrying、paused")
	}
	if current == next {
		return errors.New("同步任务已处于目标状态")
	}
	switch action {
	case "pause_sync_job":
		if current == "completed" || current == "failed" {
			return errors.New("已结束的同步任务不能暂停")
		}
	case "resume_sync_job":
		if current != "paused" && current != "queued" {
			return errors.New("只有暂停或排队中的同步任务可以恢复")
		}
	case "retry_sync_job":
		if current == "running" || current == "completed" {
			return errors.New("当前同步任务状态不允许重试")
		}
	}
	return nil
}

func validateAlertTransition(current string, next string, action string) error {
	if next == "" {
		return nil
	}
	if !allowedStatus(next, "open", "processing", "resolved", "ignored") {
		return errors.New("异常状态仅支持 open、processing、resolved、ignored")
	}
	if current == next {
		return errors.New("异常记录已处于目标状态")
	}
	if current == "resolved" || current == "ignored" {
		return errors.New("已关闭的异常记录不能再次处理")
	}
	if action == "process_alert" && current == "processing" {
		return errors.New("异常记录已处于处理中")
	}
	return nil
}

func allowedStatus(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func gatewayRequestTarget(app models.IntegrationProviderApp, req GatewayInvokeRequest) (string, string, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return "", "", errors.New("第三方 API method 不合法")
	}
	rawPath := strings.TrimSpace(req.Path)
	if rawPath == "" || !strings.HasPrefix(rawPath, "/") || strings.HasPrefix(rawPath, "//") || strings.Contains(rawPath, "\\") {
		return "", "", errors.New("第三方 API path 不合法")
	}
	relativeURL, err := url.ParseRequestURI(rawPath)
	if err != nil || relativeURL.IsAbs() || relativeURL.Host != "" || strings.Contains(relativeURL.Path, "..") {
		return "", "", errors.New("第三方 API path 不合法")
	}
	base := strings.TrimSpace(os.Getenv("INTEGRATION_PROVIDER_BASE_URL_" + envToken(app.AppCode)))
	if base == "" {
		return "", "", errors.New("第三方 API base URL 未配置")
	}
	baseURL, err := url.Parse(base)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return "", "", errors.New("第三方 API base URL 配置不合法")
	}
	target := *baseURL
	target.Path = strings.TrimRight(baseURL.Path, "/") + relativeURL.Path
	target.RawQuery = relativeURL.RawQuery
	target.Fragment = ""
	return method, target.String(), nil
}

func safeGatewayHeader(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return false
	}
	blocked := map[string]bool{
		"authorization":       true,
		"content-length":      true,
		"cookie":              true,
		"host":                true,
		"proxy-authorization": true,
		"x-api-key":           true,
		"x-forwarded-for":     true,
		"x-forwarded-host":    true,
		"x-real-ip":           true,
	}
	return !blocked[key]
}

func digestHex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func firstTime(value *time.Time, fallback time.Time) *time.Time {
	if value != nil {
		return value
	}
	return &fallback
}

func oauthAuthorizeURLFor(app models.IntegrationProviderApp, redirectURI string, scopes []string, state string) (string, string) {
	key := "INTEGRATION_OAUTH_AUTHORIZE_URL_" + envToken(app.AppCode)
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return "", "OAuth state 已创建；未配置 " + key + "，暂不能生成第三方授权跳转地址"
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", "OAuth state 已创建；第三方授权端点配置不合法"
	}
	q := u.Query()
	q.Set("state", state)
	q.Set("redirect_uri", redirectURI)
	if clientID := strings.TrimSpace(os.Getenv("INTEGRATION_OAUTH_CLIENT_ID_" + envToken(app.AppCode))); clientID != "" {
		q.Set("client_id", clientID)
	}
	if len(scopes) > 0 {
		q.Set("scope", strings.Join(scopes, " "))
	}
	u.RawQuery = q.Encode()
	return u.String(), "OAuth state 已创建"
}

func exchangeOAuthCode(ctx context.Context, app models.IntegrationProviderApp, redirectURI string, code string) (oauthTokenExchangeResult, bool, error) {
	endpoint := strings.TrimSpace(os.Getenv("INTEGRATION_OAUTH_TOKEN_URL_" + envToken(app.AppCode)))
	if endpoint == "" {
		return oauthTokenExchangeResult{}, false, nil
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token endpoint 配置不合法")
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	if clientID := strings.TrimSpace(os.Getenv("INTEGRATION_OAUTH_CLIENT_ID_" + envToken(app.AppCode))); clientID != "" {
		form.Set("client_id", clientID)
	}
	if clientSecret := oauthClientSecretFor(app); clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token exchange 请求创建失败")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token exchange 请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token exchange 响应读取失败")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token exchange 返回失败状态")
	}
	var result oauthTokenExchangeResult
	if err := json.Unmarshal(body, &result); err != nil {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token exchange 响应不是合法 JSON")
	}
	if strings.TrimSpace(result.CredentialRef) == "" {
		return oauthTokenExchangeResult{}, true, errors.New("OAuth token exchange 未返回 credential_ref，拒绝保存明文 token")
	}
	return result, true, nil
}

func refreshOAuthCredential(ctx context.Context, app models.IntegrationProviderApp, credentialRef string) (oauthTokenExchangeResult, error) {
	endpoint := strings.TrimSpace(os.Getenv("INTEGRATION_OAUTH_REFRESH_URL_" + envToken(app.AppCode)))
	if endpoint == "" {
		return oauthTokenExchangeResult{}, errors.New("OAuth refresh endpoint 未配置")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return oauthTokenExchangeResult{}, errors.New("OAuth refresh endpoint 配置不合法")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("credential_ref", credentialRef)
	if clientID := strings.TrimSpace(os.Getenv("INTEGRATION_OAUTH_CLIENT_ID_" + envToken(app.AppCode))); clientID != "" {
		form.Set("client_id", clientID)
	}
	if clientSecret := oauthClientSecretFor(app); clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return oauthTokenExchangeResult{}, errors.New("OAuth token refresh 请求创建失败")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return oauthTokenExchangeResult{}, errors.New("OAuth token refresh 请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return oauthTokenExchangeResult{}, errors.New("OAuth token refresh 响应读取失败")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oauthTokenExchangeResult{}, errors.New("OAuth token refresh 返回失败状态")
	}
	var result oauthTokenExchangeResult
	if err := json.Unmarshal(body, &result); err != nil {
		return oauthTokenExchangeResult{}, errors.New("OAuth token refresh 响应不是合法 JSON")
	}
	if strings.TrimSpace(result.CredentialRef) == "" {
		return oauthTokenExchangeResult{}, errors.New("OAuth token refresh 未返回 credential_ref，拒绝保存明文 token")
	}
	return result, nil
}

func oauthClientSecretFor(app models.IntegrationProviderApp) string {
	keys := []string{"INTEGRATION_OAUTH_CLIENT_SECRET_" + envToken(app.AppCode)}
	if app.CredentialRef != nil && strings.TrimSpace(*app.CredentialRef) != "" {
		if secret := credentialSecretFromRef(*app.CredentialRef); secret != "" {
			return secret
		}
		keys = append([]string{strings.TrimSpace(*app.CredentialRef)}, keys...)
	}
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func tokenExpiresAt(result oauthTokenExchangeResult, now time.Time) *time.Time {
	if strings.TrimSpace(result.ExpiresAt) != "" {
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(result.ExpiresAt)); err == nil {
			return &parsed
		}
	}
	if result.ExpiresIn > 0 {
		expiresAt := now.Add(time.Duration(result.ExpiresIn) * time.Second)
		return &expiresAt
	}
	return nil
}

func randomOAuthState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.New("OAuth state 随机数生成失败")
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func envToken(value string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", ":", "_")
	return strings.ToUpper(replacer.Replace(strings.TrimSpace(value)))
}

func apiCallLogRetentionDays() int {
	raw := strings.TrimSpace(os.Getenv("INTEGRATION_API_CALL_LOG_RETENTION_DAYS"))
	if raw == "" {
		return defaultAPICallLogRetentionDays
	}
	days, err := strconv.Atoi(raw)
	if err != nil {
		return defaultAPICallLogRetentionDays
	}
	if days < 30 {
		return 30
	}
	if days > 3650 {
		return 3650
	}
	return days
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
	if !isValidPlatformStatus(req.Status) {
		return errors.New("平台状态仅支持 online、beta、draft、disabled、maintenance")
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
	if !isValidAppStatus(req.Status) {
		return errors.New("应用状态仅支持 online、beta、draft、disabled、maintenance")
	}
	return nil
}

func validatePlatformCapabilityRequest(req PlatformCapabilityMutationRequest, requireCode bool) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("能力名称不能为空")
	}
	if strings.TrimSpace(req.PlatformCode) == "" && requireCode {
		return errors.New("所属接入平台不能为空")
	}
	if requireCode && normalizePlatformCode(req.Code) == "" {
		return errors.New("能力编码不能为空")
	}
	if code := normalizePlatformCode(req.Code); code != "" {
		if err := validateSimpleCode(code, "能力编码"); err != nil {
			return err
		}
	}
	if !isValidEnabledStatus(req.Status) {
		return errors.New("能力状态仅支持 enabled、disabled")
	}
	return nil
}

func validateTenantConnectionCreateRequest(req TenantConnectionCreateRequest) error {
	if normalizePlatformCode(req.ProviderAppCode) == "" {
		return errors.New("服务商应用编码不能为空")
	}
	if strings.TrimSpace(req.ConnectionName) == "" {
		return errors.New("连接名称不能为空")
	}
	if strings.TrimSpace(req.AuthSubjectID) == "" {
		return errors.New("授权主体 ID 不能为空")
	}
	if strings.TrimSpace(req.AuthSubjectName) == "" {
		return errors.New("授权主体名称不能为空")
	}
	return nil
}

func validateOAuthStartRequest(req OAuthStartRequest) error {
	if normalizePlatformCode(req.ProviderAppCode) == "" {
		return errors.New("服务商应用编码不能为空")
	}
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		return errors.New("OAuth redirect_uri 不能为空")
	}
	u, err := url.Parse(redirectURI)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("OAuth redirect_uri 不合法")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("OAuth redirect_uri 仅支持 http 或 https")
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
	if req.OverrideLimit != nil && *req.OverrideLimit < 0 {
		return errors.New("覆盖限额不能为负数")
	}
	scope := normalizedQuotaScope(req.ScopeType)
	if scope == "" {
		if req.TenantID != nil || strings.TrimSpace(req.PlatformCode) != "" || strings.TrimSpace(req.ProviderAppCode) != "" || req.TenantConnectionID != nil || req.OverrideLimit != nil || req.Priority != 0 {
			return errors.New("配置配额绑定时 scope_type 不能为空")
		}
		return nil
	}
	switch scope {
	case "global":
	case "tenant":
		if req.TenantID == nil || *req.TenantID == 0 {
			return errors.New("租户级配额绑定必须指定 tenant_id")
		}
	case "platform":
		if strings.TrimSpace(req.PlatformCode) == "" {
			return errors.New("平台级配额绑定必须指定 platform_code")
		}
	case "provider_app":
		if strings.TrimSpace(req.ProviderAppCode) == "" {
			return errors.New("应用级配额绑定必须指定 provider_app_code")
		}
	case "tenant_connection":
		if req.TenantConnectionID == nil || *req.TenantConnectionID == 0 {
			return errors.New("连接级配额绑定必须指定 tenant_connection_id")
		}
	default:
		return errors.New("scope_type 仅支持 global、tenant、platform、provider_app、tenant_connection")
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
		return platformStatusOnline
	case "testing":
		return platformStatusBeta
	case platformStatusOnline, platformStatusBeta, platformStatusDraft, platformStatusDisabled, platformStatusMaintenance:
		return strings.TrimSpace(status)
	default:
		return platformStatusDraft
	}
}

func normalizeAppStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "enabled", "online":
		return appStatusOnline
	case "testing", "beta":
		return appStatusBeta
	case appStatusDisabled, appStatusMaintenance, appStatusDraft:
		return strings.TrimSpace(status)
	default:
		return appStatusDraft
	}
}

func normalizeEnabledStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "enabled", "online":
		return enabledStatusEnabled
	case "disabled", "paused":
		return enabledStatusDisabled
	default:
		return enabledStatusEnabled
	}
}

func isValidPlatformStatus(status string) bool {
	raw := strings.TrimSpace(status)
	if raw == "" {
		return true
	}
	switch raw {
	case "enabled", "testing", platformStatusOnline, platformStatusBeta, platformStatusDraft, platformStatusDisabled, platformStatusMaintenance:
		return true
	default:
		return false
	}
}

func isValidAppStatus(status string) bool {
	raw := strings.TrimSpace(status)
	if raw == "" {
		return true
	}
	switch raw {
	case "enabled", "testing", appStatusOnline, appStatusBeta, appStatusDraft, appStatusDisabled, appStatusMaintenance:
		return true
	default:
		return false
	}
}

func isValidEnabledStatus(status string) bool {
	raw := strings.TrimSpace(status)
	if raw == "" {
		return true
	}
	switch raw {
	case "enabled", "online", "disabled", "paused":
		return true
	default:
		return false
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

func normalizeDataDirection(value string) string {
	switch strings.TrimSpace(value) {
	case "push", "pull", "both":
		return strings.TrimSpace(value)
	case "推送":
		return "push"
	case "双向":
		return "both"
	default:
		return "pull"
	}
}

func appCapabilityConfigPatch(raw string, req AppCapabilityPatchRequest) (string, bool, error) {
	config := map[string]interface{}{}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &config); err != nil {
			return "", false, errors.New("应用能力配置不是合法 JSON")
		}
	}
	changed := false
	for key, value := range req.Config {
		if strings.TrimSpace(key) == "" {
			return "", false, errors.New("应用能力配置键不能为空")
		}
		config[key] = value
		changed = true
	}
	if req.OpenToTenant != nil {
		config["open_to_tenant"] = *req.OpenToTenant
		changed = true
	}
	if req.DefaultEnabled != nil {
		config["default_enabled"] = *req.DefaultEnabled
		changed = true
	}
	if req.TenantConfigurable != nil {
		config["tenant_configurable"] = *req.TenantConfigurable
		changed = true
	}
	if req.Enabled != nil {
		config["default_enabled"] = *req.Enabled
		changed = true
	}
	if !changed {
		return "", false, nil
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return "", false, errors.New("应用能力配置序列化失败")
	}
	return string(encoded), true, nil
}

func normalizeProviderAppCapabilitySummaries(rows []repositories.ProviderAppCapabilitySummary) []repositories.ProviderAppCapabilitySummary {
	for index := range rows {
		config := map[string]interface{}{}
		if strings.TrimSpace(rows[index].Config) != "" {
			_ = json.Unmarshal([]byte(rows[index].Config), &config)
		}
		if value, ok := boolConfig(config, "open_to_tenant"); ok {
			rows[index].OpenToTenant = value
		}
		if value, ok := boolConfig(config, "default_enabled"); ok {
			rows[index].DefaultEnabled = value
		}
		if value, ok := boolConfig(config, "tenant_configurable"); ok {
			rows[index].TenantConfigurable = value
		}
	}
	return rows
}

func boolConfig(config map[string]interface{}, key string) (bool, bool) {
	value, ok := config[key]
	if !ok {
		return false, false
	}
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "enabled":
			return true, true
		case "false", "0", "no", "disabled":
			return false, true
		}
	}
	return false, false
}

func normalizeConnectionStatus(status string) string {
	switch strings.TrimSpace(status) {
	case connectionStatusConnected, connectionStatusPending, connectionStatusFailed, connectionStatusPaused, connectionStatusNotConnected:
		return strings.TrimSpace(status)
	case "enabled":
		return connectionStatusConnected
	case "disabled":
		return connectionStatusPaused
	default:
		return ""
	}
}

func normalizeReviewStatus(status string) string {
	switch strings.TrimSpace(status) {
	case reviewStatusPending, reviewStatusApproved, reviewStatusRejected:
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

func normalizedQuotaScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case "all":
		return "global"
	case "app":
		return "provider_app"
	case "connection":
		return "tenant_connection"
	case "global", "tenant", "platform", "provider_app", "tenant_connection":
		return strings.TrimSpace(scope)
	default:
		return ""
	}
}

func quotaPolicyHasBinding(req QuotaPolicyMutationRequest) bool {
	return normalizedQuotaScope(req.ScopeType) != ""
}

func (s *Service) quotaBindingFromRequest(ctx context.Context, policyID uint64, userID uint64, req QuotaPolicyMutationRequest) (models.IntegrationQuotaBinding, error) {
	scope := normalizedQuotaScope(req.ScopeType)
	binding := models.IntegrationQuotaBinding{
		PolicyID:      policyID,
		OverrideLimit: req.OverrideLimit,
		Priority:      req.Priority,
		Status:        "enabled",
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
	}
	switch scope {
	case "global":
		return binding, nil
	case "tenant":
		binding.TenantID = req.TenantID
		return binding, nil
	case "platform":
		platform, err := s.repo.GetPlatformByCode(ctx, normalizePlatformCode(req.PlatformCode))
		if err != nil {
			return models.IntegrationQuotaBinding{}, humanizeRepoError(err, "接入平台不存在")
		}
		binding.PlatformID = &platform.ID
		if req.TenantID != nil && *req.TenantID > 0 {
			binding.TenantID = req.TenantID
		}
		return binding, nil
	case "provider_app":
		app, err := s.repo.GetProviderAppByCode(ctx, normalizePlatformCode(req.ProviderAppCode))
		if err != nil {
			return models.IntegrationQuotaBinding{}, humanizeRepoError(err, "服务商应用不存在")
		}
		binding.PlatformID = &app.PlatformID
		binding.ProviderAppID = &app.ID
		if req.TenantID != nil && *req.TenantID > 0 {
			binding.TenantID = req.TenantID
		}
		return binding, nil
	case "tenant_connection":
		connection, err := s.repo.GetTenantConnection(ctx, *req.TenantConnectionID)
		if err != nil {
			return models.IntegrationQuotaBinding{}, humanizeRepoError(err, "租户连接不存在")
		}
		if req.TenantID != nil && *req.TenantID > 0 && *req.TenantID != connection.TenantID {
			return models.IntegrationQuotaBinding{}, errors.New("tenant_id 与 tenant_connection_id 不匹配")
		}
		binding.TenantID = &connection.TenantID
		binding.PlatformID = &connection.PlatformID
		binding.ProviderAppID = &connection.ProviderAppID
		binding.TenantConnectionID = &connection.ID
		return binding, nil
	default:
		return models.IntegrationQuotaBinding{}, errors.New("scope_type 仅支持 global、tenant、platform、provider_app、tenant_connection")
	}
}

func (s *Service) quotaLimit(ctx context.Context, tenantID uint64, tenantConnectionID *uint64, quotaCode string) (int64, error) {
	if resolved, ok, err := s.repo.ResolveQuotaPolicy(ctx, tenantID, tenantConnectionID, quotaCode); err != nil {
		return 0, err
	} else if ok {
		return resolved.Limit, nil
	}
	limit, exists, err := s.repo.TenantQuotaLimit(ctx, tenantID, quotaCode)
	if err != nil {
		return 0, err
	}
	if !exists {
		return -1, nil
	}
	return int64(limit), nil
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
	case errors.Is(err, repositories.ErrOAuthStateNotFound):
		return errors.New("OAuth state 不存在或已失效")
	case errors.Is(err, repositories.ErrAPICallLogNotFound):
		return errors.New("调用日志不存在")
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
		OfficialURL:     platform.OfficialURL,
		Status:          platform.Status,
		TenantVisible:   platform.TenantVisible,
		OwnerName:       platform.OwnerName,
		SortOrder:       platform.SortOrder,
		Description:     platform.Description,
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
		AppType:       app.AppType,
		AuthMode:      app.AuthMode,
		Environment:   app.Environment,
		Status:        app.Status,
		TenantVisible: app.TenantVisible,
		CallbackURL:   app.CallbackURL,
		WebhookURL:    app.WebhookURL,
		CredentialRef: app.CredentialRef,
		OwnerName:     app.OwnerName,
		Description:   app.Description,
	}
}

func platformCapabilityToSummary(capability models.IntegrationPlatformCapability, platform models.IntegrationPlatform) repositories.PlatformCapabilitySummary {
	return repositories.PlatformCapabilitySummary{
		ID:             capability.ID,
		PlatformID:     capability.PlatformID,
		PlatformName:   platform.PlatformName,
		CapabilityCode: capability.CapabilityCode,
		CapabilityName: capability.CapabilityName,
		CapabilityType: capability.CapabilityType,
		AuthScopeCode:  capability.AuthScopeCode,
		DataDirection:  capability.DataDirection,
		Status:         capability.Status,
		Description:    capability.Description,
	}
}
