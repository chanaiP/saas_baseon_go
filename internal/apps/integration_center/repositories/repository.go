package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

var (
	ErrPlatformNotFound     = errors.New("integration platform not found")
	ErrProviderAppNotFound  = errors.New("integration provider app not found")
	ErrCapabilityNotFound   = errors.New("integration app capability not found")
	ErrConnectionNotFound   = errors.New("integration tenant connection not found")
	ErrSyncJobNotFound      = errors.New("integration sync job not found")
	ErrQuotaPolicyNotFound  = errors.New("integration quota policy not found")
	ErrAlertNotFound        = errors.New("integration alert not found")
	ErrOAuthStateNotFound   = errors.New("integration oauth state not found")
	ErrQuotaExceeded        = errors.New("integration quota exceeded")
	ErrWebhookEventNotFound = errors.New("integration webhook event not found")
	ErrWebhookEventExists   = errors.New("integration webhook event already exists")
	ErrAPICallLogNotFound   = errors.New("integration api call log not found")
)

type Repository struct {
	db *gorm.DB
}

type ResolvedQuotaPolicy struct {
	Policy       models.IntegrationQuotaPolicy
	Binding      models.IntegrationQuotaBinding
	Limit        int64
	HasOverride  bool
	MatchedScope string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type ListOptions struct {
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

func (opts ListOptions) normalized() ListOptions {
	if opts.Skip < 0 {
		opts.Skip = 0
	}
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	opts.Keyword = strings.TrimSpace(opts.Keyword)
	opts.Status = strings.TrimSpace(opts.Status)
	opts.PlatformCode = strings.TrimSpace(opts.PlatformCode)
	opts.ProviderAppCode = strings.TrimSpace(opts.ProviderAppCode)
	opts.SortBy = strings.TrimSpace(opts.SortBy)
	opts.SortOrder = strings.ToLower(strings.TrimSpace(opts.SortOrder))
	if opts.SortOrder != "asc" {
		opts.SortOrder = "desc"
	}
	return opts
}

type DashboardCounts struct {
	Platforms       int64
	ProviderApps    int64
	Connections     int64
	OpenAlerts      int64
	TodayAPICalls   int64
	RunningSyncJobs int64
}

type PlatformSummary struct {
	ID              uint64  `json:"id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	ShortName       *string `json:"short_name"`
	PlatformType    string  `json:"platform_type"`
	AccessMode      string  `json:"access_mode"`
	OfficialURL     *string `json:"official_url"`
	Status          string  `json:"status"`
	TenantVisible   bool    `json:"tenant_visible"`
	OwnerName       *string `json:"owner_name"`
	SortOrder       int     `json:"sort_order"`
	Description     *string `json:"description"`
	AppCount        int64   `json:"app_count"`
	CapabilityCount int64   `json:"capability_count"`
	ConnectionCount int64   `json:"connection_count"`
	OpenAlertCount  int64   `json:"open_alert_count"`
	CallsToday      int64   `json:"calls_today"`
	SuccessRate     float64 `json:"success_rate"`
}

type ProviderAppSummary struct {
	ID              uint64  `json:"id"`
	PlatformID      uint64  `json:"platform_id"`
	PlatformName    string  `json:"platform_name"`
	AppCode         string  `json:"app_code"`
	AppName         string  `json:"app_name"`
	AppType         string  `json:"app_type"`
	AuthMode        string  `json:"auth_mode"`
	Environment     string  `json:"environment"`
	Status          string  `json:"status"`
	TenantVisible   bool    `json:"tenant_visible"`
	CallbackURL     *string `json:"callback_url"`
	WebhookURL      *string `json:"webhook_url"`
	CredentialRef   *string `json:"credential_ref"`
	OwnerName       *string `json:"owner_name"`
	Description     *string `json:"description"`
	CapabilityCount int64   `json:"capability_count"`
	ConnectionCount int64   `json:"connection_count"`
	AlertCount      int64   `json:"alert_count"`
	CallsToday      int64   `json:"calls_today"`
}

type PlatformCapabilitySummary struct {
	ID             uint64  `json:"id"`
	PlatformID     uint64  `json:"platform_id"`
	PlatformName   string  `json:"platform_name"`
	CapabilityCode string  `json:"capability_code"`
	CapabilityName string  `json:"capability_name"`
	CapabilityType string  `json:"capability_type"`
	AuthScopeCode  *string `json:"auth_scope_code"`
	DataDirection  string  `json:"data_direction"`
	Status         string  `json:"status"`
	Description    *string `json:"description"`
}

type ProviderAppCapabilitySummary struct {
	ID                   uint64 `json:"id"`
	ProviderAppID        uint64 `json:"provider_app_id"`
	ProviderAppCode      string `json:"provider_app_code"`
	ProviderAppName      string `json:"provider_app_name"`
	PlatformID           uint64 `json:"platform_id"`
	PlatformName         string `json:"platform_name"`
	PlatformCapabilityID uint64 `json:"platform_capability_id"`
	CapabilityCode       string `json:"capability_code"`
	CapabilityName       string `json:"capability_name"`
	CapabilityType       string `json:"capability_type"`
	ConnectionStatus     string `json:"connection_status"`
	ReviewStatus         string `json:"review_status"`
	Enabled              bool   `json:"enabled"`
	OpenToTenant         bool   `json:"open_to_tenant"`
	DefaultEnabled       bool   `json:"default_enabled"`
	TenantConfigurable   bool   `json:"tenant_configurable"`
	Config               string `json:"config"`
}

type TenantConnectionSummary struct {
	ID                   uint64     `json:"id"`
	TenantID             uint64     `json:"tenant_id"`
	TenantName           string     `json:"tenant_name"`
	PlatformID           uint64     `json:"platform_id"`
	PlatformName         string     `json:"platform_name"`
	ProviderAppID        uint64     `json:"provider_app_id"`
	ProviderAppName      string     `json:"provider_app_name"`
	ConnectionName       string     `json:"connection_name"`
	AuthSubjectType      string     `json:"auth_subject_type"`
	AuthSubjectID        string     `json:"auth_subject_id"`
	AuthSubjectName      string     `json:"auth_subject_name"`
	AuthScope            string     `json:"auth_scope"`
	AuthStatus           string     `json:"auth_status"`
	ConnectionStatus     string     `json:"connection_status"`
	TokenStatus          string     `json:"token_status"`
	TokenCredentialRef   *string    `json:"token_credential_ref"`
	LastSyncAt           *time.Time `json:"last_sync_at"`
	LastErrorMessage     *string    `json:"last_error_message"`
	FinalCapabilityCount int64      `json:"final_capability_count"`
	CallsToday           int64      `json:"calls_today"`
	FinalCapabilities    string     `json:"final_capabilities"`
}

type SyncJobSummary struct {
	ID                 uint64     `json:"id"`
	TenantID           uint64     `json:"tenant_id"`
	TenantName         string     `json:"tenant_name"`
	TenantConnectionID uint64     `json:"tenant_connection_id"`
	PlatformID         uint64     `json:"platform_id"`
	PlatformName       string     `json:"platform_name"`
	ProviderAppID      uint64     `json:"provider_app_id"`
	ProviderAppName    string     `json:"provider_app_name"`
	ConnectionName     string     `json:"connection_name"`
	CapabilityCode     string     `json:"capability_code"`
	CapabilityName     string     `json:"capability_name"`
	JobType            string     `json:"job_type"`
	TriggerMode        string     `json:"trigger_mode"`
	Status             string     `json:"status"`
	TotalCount         int64      `json:"total_count"`
	SuccessCount       int64      `json:"success_count"`
	FailedCount        int64      `json:"failed_count"`
	RetryCount         int        `json:"retry_count"`
	NextRetryAt        *time.Time `json:"next_retry_at"`
	StartedAt          *time.Time `json:"started_at"`
	FinishedAt         *time.Time `json:"finished_at"`
	ErrorCode          *string    `json:"error_code"`
	ErrorMessage       *string    `json:"error_message"`
}

type QuotaUsageSummary struct {
	ID                 uint64     `json:"id"`
	TenantID           uint64     `json:"tenant_id"`
	TenantName         string     `json:"tenant_name"`
	TenantConnectionID *uint64    `json:"tenant_connection_id"`
	ConnectionName     *string    `json:"connection_name"`
	PlatformName       *string    `json:"platform_name"`
	ProviderAppName    *string    `json:"provider_app_name"`
	QuotaCode          string     `json:"quota_code"`
	PolicyCode         *string    `json:"policy_code"`
	PolicyName         *string    `json:"policy_name"`
	PeriodKey          string     `json:"period_key"`
	UsedAmount         int64      `json:"used_amount"`
	LimitAmount        int64      `json:"limit_amount"`
	LimitedCount       int64      `json:"limited_count"`
	LastUsedAt         *time.Time `json:"last_used_at"`
}

func (r *Repository) UserByID(ctx context.Context, id uint64) (models.AppUser, error) {
	var user models.AppUser
	err := r.active(ctx, &models.AppUser{}).Where("id = ?", id).First(&user).Error
	return user, err
}

func (r *Repository) CreateAuditLog(ctx context.Context, log models.AuditLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&log).Error
}

func (r *Repository) CreateAPICallLog(ctx context.Context, log models.IntegrationAPICallLog) error {
	if log.CalledAt.IsZero() {
		log.CalledAt = time.Now()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = log.CalledAt
	}
	return r.db.WithContext(ctx).Create(&log).Error
}

func (r *Repository) TenantFeatureAllowed(ctx context.Context, tenantID uint64, featureCode string) (bool, error) {
	var feature models.SaasFeature
	if err := r.db.WithContext(ctx).Where("feature_code = ? AND status = ?", featureCode, 1).First(&feature).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	sub, ok, err := r.activeSubscription(ctx, tenantID)
	if err != nil || !ok {
		return false, err
	}
	now := time.Now()
	var override models.TenantFeatureOverride
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND feature_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, feature.ID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.Enabled, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	var count int64
	err = r.db.WithContext(ctx).Model(&models.SaasPlanFeature{}).Where("plan_id = ? AND feature_id = ? AND enabled = ?", sub.PlanID, feature.ID, true).Count(&count).Error
	return count > 0, err
}

func (r *Repository) TenantQuotaLimit(ctx context.Context, tenantID uint64, quotaCode string) (int, bool, error) {
	var quota models.SaasQuota
	if err := r.db.WithContext(ctx).Where("quota_code = ? AND status = ?", quotaCode, 1).First(&quota).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	sub, ok, err := r.activeSubscription(ctx, tenantID)
	if err != nil || !ok {
		return 0, true, err
	}
	now := time.Now()
	var override models.TenantQuotaOverride
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quota.ID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue, true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, true, err
	}
	var planQuota models.SaasPlanQuota
	if err := r.db.WithContext(ctx).Where("plan_id = ? AND quota_id = ?", sub.PlanID, quota.ID).First(&planQuota).Error; err == nil {
		return planQuota.QuotaValue, true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, true, err
	}
	return 0, true, nil
}

func (r *Repository) CountTenantConnections(ctx context.Context, tenantID uint64) (int64, error) {
	var count int64
	err := r.active(ctx, &models.IntegrationTenantConnection{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *Repository) ConsumeQuota(ctx context.Context, tenantID uint64, tenantConnectionID *uint64, quotaCode string, periodKey string, amount int64, limit int64) (models.IntegrationQuotaUsage, error) {
	if amount <= 0 {
		amount = 1
	}
	now := time.Now()
	var usage models.IntegrationQuotaUsage
	exceeded := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND quota_code = ? AND period_key = ?", tenantID, quotaCode, periodKey)
		if tenantConnectionID == nil {
			query = query.Where("tenant_connection_id IS NULL")
		} else {
			query = query.Where("tenant_connection_id = ?", *tenantConnectionID)
		}
		err := query.First(&usage).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if limit >= 0 && amount > limit {
				usage = models.IntegrationQuotaUsage{
					TenantID:           tenantID,
					TenantConnectionID: tenantConnectionID,
					QuotaCode:          quotaCode,
					PeriodKey:          periodKey,
					UsedAmount:         0,
					LimitedCount:       1,
					LastUsedAt:         &now,
					CreatedAt:          now,
					UpdatedAt:          now,
				}
				if createErr := tx.Create(&usage).Error; createErr != nil {
					return createErr
				}
				exceeded = true
				return nil
			}
			usage = models.IntegrationQuotaUsage{
				TenantID:           tenantID,
				TenantConnectionID: tenantConnectionID,
				QuotaCode:          quotaCode,
				PeriodKey:          periodKey,
				UsedAmount:         amount,
				LastUsedAt:         &now,
				CreatedAt:          now,
				UpdatedAt:          now,
			}
			return tx.Create(&usage).Error
		}
		if err != nil {
			return err
		}
		if limit >= 0 && usage.UsedAmount+amount > limit {
			if err := tx.Model(&usage).Updates(map[string]interface{}{
				"limited_count": gorm.Expr("limited_count + ?", 1),
				"last_used_at":  &now,
				"updated_at":    now,
			}).Error; err != nil {
				return err
			}
			usage.LimitedCount++
			usage.LastUsedAt = &now
			usage.UpdatedAt = now
			exceeded = true
			return nil
		}
		if err := tx.Model(&usage).Updates(map[string]interface{}{
			"used_amount":  gorm.Expr("used_amount + ?", amount),
			"last_used_at": &now,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}
		usage.UsedAmount += amount
		usage.LastUsedAt = &now
		usage.UpdatedAt = now
		return nil
	})
	if err == nil && exceeded {
		return usage, ErrQuotaExceeded
	}
	return usage, err
}

func (r *Repository) ResolveQuotaPolicy(ctx context.Context, tenantID uint64, tenantConnectionID *uint64, quotaCode string) (ResolvedQuotaPolicy, bool, error) {
	var connection models.IntegrationTenantConnection
	hasConnection := false
	if tenantConnectionID != nil && *tenantConnectionID > 0 {
		conn, err := r.GetTenantConnection(ctx, *tenantConnectionID)
		if err != nil {
			return ResolvedQuotaPolicy{}, false, err
		}
		if conn.TenantID != tenantID {
			return ResolvedQuotaPolicy{}, false, ErrConnectionNotFound
		}
		connection = conn
		hasConnection = true
	}

	type row struct {
		PolicyID                  uint64
		PolicyCode                string
		PolicyName                string
		QuotaCode                 string
		QuotaUnit                 string
		PeriodType                string
		DefaultLimit              int64
		OverLimitAction           string
		PolicyStatus              string
		Description               *string
		PolicyCreatedBy           *uint64
		PolicyUpdatedBy           *uint64
		PolicyCreatedAt           time.Time
		PolicyUpdatedAt           time.Time
		PolicyDeletedAt           *time.Time
		BindingID                 uint64
		BindingTenantID           *uint64
		BindingPlatformID         *uint64
		BindingProviderAppID      *uint64
		BindingTenantConnectionID *uint64
		OverrideLimit             *int64
		Priority                  int
		BindingStatus             string
		BindingCreatedBy          *uint64
		BindingUpdatedBy          *uint64
		BindingCreatedAt          time.Time
		BindingUpdatedAt          time.Time
		BindingDeletedAt          *time.Time
		Specificity               int
	}
	var matched row
	query := r.db.WithContext(ctx).
		Table("integration_quota_bindings qb").
		Select(`qp.id AS policy_id, qp.policy_code, qp.policy_name, qp.quota_code, qp.quota_unit, qp.period_type, qp.default_limit, qp.over_limit_action, qp.status AS policy_status, qp.description, qp.created_by AS policy_created_by, qp.updated_by AS policy_updated_by, qp.created_at AS policy_created_at, qp.updated_at AS policy_updated_at, qp.deleted_at AS policy_deleted_at,
			qb.id AS binding_id, qb.tenant_id AS binding_tenant_id, qb.platform_id AS binding_platform_id, qb.provider_app_id AS binding_provider_app_id, qb.tenant_connection_id AS binding_tenant_connection_id, qb.override_limit, qb.priority, qb.status AS binding_status, qb.created_by AS binding_created_by, qb.updated_by AS binding_updated_by, qb.created_at AS binding_created_at, qb.updated_at AS binding_updated_at, qb.deleted_at AS binding_deleted_at,
			(CASE WHEN qb.tenant_connection_id IS NOT NULL THEN 8 ELSE 0 END + CASE WHEN qb.provider_app_id IS NOT NULL THEN 4 ELSE 0 END + CASE WHEN qb.platform_id IS NOT NULL THEN 2 ELSE 0 END + CASE WHEN qb.tenant_id IS NOT NULL THEN 1 ELSE 0 END) AS specificity`).
		Joins("JOIN integration_quota_policies qp ON qp.id = qb.policy_id AND qp.deleted_at IS NULL").
		Where("qb.deleted_at IS NULL AND qb.status = ? AND qp.status = ? AND qp.quota_code = ?", "enabled", "enabled", quotaCode).
		Where("(qb.tenant_id IS NULL OR qb.tenant_id = ?)", tenantID)
	if hasConnection {
		query = query.
			Where("(qb.tenant_connection_id IS NULL OR qb.tenant_connection_id = ?)", connection.ID).
			Where("(qb.platform_id IS NULL OR qb.platform_id = ?)", connection.PlatformID).
			Where("(qb.provider_app_id IS NULL OR qb.provider_app_id = ?)", connection.ProviderAppID)
	} else {
		query = query.Where("qb.tenant_connection_id IS NULL AND qb.platform_id IS NULL AND qb.provider_app_id IS NULL")
	}
	err := query.Order("qb.priority DESC, specificity DESC, qb.id DESC").Limit(1).Scan(&matched).Error
	if err != nil {
		return ResolvedQuotaPolicy{}, false, err
	}
	if matched.BindingID == 0 {
		return ResolvedQuotaPolicy{}, false, nil
	}
	policy := models.IntegrationQuotaPolicy{
		ID:              matched.PolicyID,
		PolicyCode:      matched.PolicyCode,
		PolicyName:      matched.PolicyName,
		QuotaCode:       matched.QuotaCode,
		QuotaUnit:       matched.QuotaUnit,
		PeriodType:      matched.PeriodType,
		DefaultLimit:    matched.DefaultLimit,
		OverLimitAction: matched.OverLimitAction,
		Status:          matched.PolicyStatus,
		Description:     matched.Description,
		CreatedBy:       matched.PolicyCreatedBy,
		UpdatedBy:       matched.PolicyUpdatedBy,
		CreatedAt:       matched.PolicyCreatedAt,
		UpdatedAt:       matched.PolicyUpdatedAt,
		DeletedAt:       matched.PolicyDeletedAt,
	}
	binding := models.IntegrationQuotaBinding{
		ID:                 matched.BindingID,
		TenantID:           matched.BindingTenantID,
		PlatformID:         matched.BindingPlatformID,
		ProviderAppID:      matched.BindingProviderAppID,
		TenantConnectionID: matched.BindingTenantConnectionID,
		PolicyID:           matched.PolicyID,
		OverrideLimit:      matched.OverrideLimit,
		Priority:           matched.Priority,
		Status:             matched.BindingStatus,
		CreatedBy:          matched.BindingCreatedBy,
		UpdatedBy:          matched.BindingUpdatedBy,
		CreatedAt:          matched.BindingCreatedAt,
		UpdatedAt:          matched.BindingUpdatedAt,
		DeletedAt:          matched.BindingDeletedAt,
	}
	limit := policy.DefaultLimit
	hasOverride := false
	if binding.OverrideLimit != nil {
		limit = *binding.OverrideLimit
		hasOverride = true
	}
	return ResolvedQuotaPolicy{Policy: policy, Binding: binding, Limit: limit, HasOverride: hasOverride, MatchedScope: quotaBindingScope(binding)}, true, nil
}

func (r *Repository) CreateTenantConnection(ctx context.Context, connection models.IntegrationTenantConnection) (models.IntegrationTenantConnection, error) {
	now := time.Now()
	connection.CreatedAt = now
	connection.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&connection).Error
	return connection, err
}

func (r *Repository) GetTenantConnectionBySubject(ctx context.Context, tenantID uint64, platformID uint64, providerAppID uint64, subjectType string, subjectID string) (models.IntegrationTenantConnection, error) {
	var connection models.IntegrationTenantConnection
	err := r.active(ctx, &models.IntegrationTenantConnection{}).
		Where("tenant_id = ? AND platform_id = ? AND provider_app_id = ? AND auth_subject_type = ? AND auth_subject_id = ?", tenantID, platformID, providerAppID, subjectType, subjectID).
		First(&connection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationTenantConnection{}, ErrConnectionNotFound
	}
	return connection, err
}

func (r *Repository) CreateOAuthState(ctx context.Context, state models.IntegrationOAuthState) (models.IntegrationOAuthState, error) {
	now := time.Now()
	state.CreatedAt = now
	state.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&state).Error
	return state, err
}

func (r *Repository) GetOAuthState(ctx context.Context, state string) (models.IntegrationOAuthState, error) {
	var row models.IntegrationOAuthState
	err := r.active(ctx, &models.IntegrationOAuthState{}).Where("state = ?", state).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationOAuthState{}, ErrOAuthStateNotFound
	}
	return row, err
}

func (r *Repository) ConsumeOAuthState(ctx context.Context, id uint64) (models.IntegrationOAuthState, error) {
	now := time.Now()
	patch := map[string]interface{}{
		"status":      "consumed",
		"consumed_at": &now,
		"updated_at":  now,
	}
	var row models.IntegrationOAuthState
	err := r.active(ctx, &models.IntegrationOAuthState{}).Where("id = ? AND status = ?", id, "pending").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationOAuthState{}, ErrOAuthStateNotFound
	}
	if err != nil {
		return models.IntegrationOAuthState{}, err
	}
	result := r.db.WithContext(ctx).Model(&models.IntegrationOAuthState{}).Where("id = ? AND status = ?", id, "pending").Updates(patch)
	if result.Error != nil {
		return models.IntegrationOAuthState{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.IntegrationOAuthState{}, ErrOAuthStateNotFound
	}
	if err := r.active(ctx, &models.IntegrationOAuthState{}).Where("id = ?", id).First(&row).Error; err != nil {
		return models.IntegrationOAuthState{}, err
	}
	return row, nil
}

func (r *Repository) activeSubscription(ctx context.Context, tenantID uint64) (models.TenantSubscription, bool, error) {
	var sub models.TenantSubscription
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.TenantSubscription{}, false, nil
	}
	if err != nil {
		return models.TenantSubscription{}, false, err
	}
	if !subscriptionAllows(sub.SubscriptionStatus, sub.EndTime, time.Now()) {
		return sub, false, nil
	}
	return sub, true, nil
}

func subscriptionAllows(status string, endTime *time.Time, now time.Time) bool {
	switch status {
	case "OVERDUE", "FROZEN", "EXPIRED", "CANCELLED":
		return false
	default:
		return endTime == nil || endTime.After(now)
	}
}

func (r *Repository) Counts(ctx context.Context) (DashboardCounts, error) {
	var counts DashboardCounts
	if err := r.active(ctx, &models.IntegrationPlatform{}).Count(&counts.Platforms).Error; err != nil {
		return DashboardCounts{}, err
	}
	if err := r.active(ctx, &models.IntegrationProviderApp{}).Count(&counts.ProviderApps).Error; err != nil {
		return DashboardCounts{}, err
	}
	if err := r.active(ctx, &models.IntegrationTenantConnection{}).Count(&counts.Connections).Error; err != nil {
		return DashboardCounts{}, err
	}
	if err := r.active(ctx, &models.IntegrationAlert{}).Where("status <> ?", "resolved").Count(&counts.OpenAlerts).Error; err != nil {
		return DashboardCounts{}, err
	}
	today := time.Now().Truncate(24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&models.IntegrationAPICallLog{}).Where("called_at >= ?", today).Count(&counts.TodayAPICalls).Error; err != nil {
		return DashboardCounts{}, err
	}
	if err := r.active(ctx, &models.IntegrationSyncJob{}).Where("status IN ?", []string{"pending", "running", "retrying"}).Count(&counts.RunningSyncJobs).Error; err != nil {
		return DashboardCounts{}, err
	}
	return counts, nil
}

func (r *Repository) ListPlatforms(ctx context.Context, opts ListOptions) ([]PlatformSummary, int64, error) {
	var rows []PlatformSummary
	opts = opts.normalized()
	today := time.Now().Truncate(24 * time.Hour)
	base := r.db.WithContext(ctx).Table("integration_platforms p").Where("p.deleted_at IS NULL")
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		base = base.Where("(p.platform_code LIKE ? OR p.platform_name LIKE ? OR p.platform_short_name LIKE ?)", like, like, like)
	}
	if opts.Status != "" {
		base = base.Where("p.status = ?", opts.Status)
	}
	base = applyTimeRange(base, "p.updated_at", opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.
		Table("integration_platforms p").
		Select(`
			p.id,
			p.platform_code AS code,
			p.platform_name AS name,
			p.platform_short_name AS short_name,
			p.platform_type,
			p.access_mode,
			p.official_url,
			p.status,
			p.tenant_visible,
			p.owner_name,
			p.sort_order,
			p.description,
			COUNT(DISTINCT pa.id) AS app_count,
			COUNT(DISTINCT pc.id) AS capability_count,
			COUNT(DISTINCT tc.id) AS connection_count,
			COUNT(DISTINCT al.id) AS open_alert_count,
			COALESCE(ls.calls_today, 0) AS calls_today,
			COALESCE(ls.success_rate, 0) AS success_rate`).
		Joins("LEFT JOIN integration_provider_apps pa ON pa.platform_id = p.id AND pa.deleted_at IS NULL").
		Joins("LEFT JOIN integration_platform_capabilities pc ON pc.platform_id = p.id AND pc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_tenant_connections tc ON tc.platform_id = p.id AND tc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_alerts al ON al.platform_id = p.id AND al.status <> ? AND al.deleted_at IS NULL", "resolved").
		Joins(`LEFT JOIN (
			SELECT platform_id,
				COUNT(*) AS calls_today,
				ROUND(100.0 * SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0), 2) AS success_rate
			FROM integration_api_call_logs
			WHERE platform_id IS NOT NULL AND called_at >= ?
			GROUP BY platform_id
		) ls ON ls.platform_id = p.id`, today).
		Group("p.id, p.platform_code, p.platform_name, p.platform_short_name, p.platform_type, p.access_mode, p.official_url, p.status, p.tenant_visible, p.owner_name, p.sort_order, p.description, ls.calls_today, ls.success_rate").
		Order(orderClause(opts, "p.sort_order ASC, p.id ASC", map[string]string{"created_at": "p.created_at", "updated_at": "p.updated_at", "name": "p.platform_name", "status": "p.status"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) GetPlatformByCode(ctx context.Context, code string) (models.IntegrationPlatform, error) {
	var platform models.IntegrationPlatform
	err := r.active(ctx, &models.IntegrationPlatform{}).Where("platform_code = ?", code).First(&platform).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationPlatform{}, ErrPlatformNotFound
	}
	return platform, err
}

func (r *Repository) GetPlatformByID(ctx context.Context, id uint64) (models.IntegrationPlatform, error) {
	var platform models.IntegrationPlatform
	err := r.active(ctx, &models.IntegrationPlatform{}).Where("id = ?", id).First(&platform).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationPlatform{}, ErrPlatformNotFound
	}
	return platform, err
}

func (r *Repository) CreatePlatform(ctx context.Context, platform models.IntegrationPlatform) (models.IntegrationPlatform, error) {
	now := time.Now()
	platform.CreatedAt = now
	platform.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&platform).Error
	return platform, err
}

func (r *Repository) GetProviderAppByCode(ctx context.Context, code string) (models.IntegrationProviderApp, error) {
	var app models.IntegrationProviderApp
	err := r.active(ctx, &models.IntegrationProviderApp{}).Where("app_code = ?", code).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationProviderApp{}, ErrProviderAppNotFound
	}
	return app, err
}

func (r *Repository) GetProviderAppByID(ctx context.Context, id uint64) (models.IntegrationProviderApp, error) {
	var app models.IntegrationProviderApp
	err := r.active(ctx, &models.IntegrationProviderApp{}).Where("id = ?", id).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationProviderApp{}, ErrProviderAppNotFound
	}
	return app, err
}

func (r *Repository) CreateProviderApp(ctx context.Context, app models.IntegrationProviderApp) (models.IntegrationProviderApp, error) {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&app).Error
	return app, err
}

func (r *Repository) CreatePlatformCapability(ctx context.Context, capability models.IntegrationPlatformCapability) (models.IntegrationPlatformCapability, error) {
	now := time.Now()
	capability.CreatedAt = now
	capability.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&capability).Error
	return capability, err
}

func (r *Repository) GetPlatformCapability(ctx context.Context, id uint64) (models.IntegrationPlatformCapability, error) {
	var capability models.IntegrationPlatformCapability
	err := r.active(ctx, &models.IntegrationPlatformCapability{}).Where("id = ?", id).First(&capability).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationPlatformCapability{}, ErrCapabilityNotFound
	}
	return capability, err
}

func (r *Repository) UpdatePlatformCapability(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationPlatformCapability, error) {
	capability, err := r.GetPlatformCapability(ctx, id)
	if err != nil {
		return models.IntegrationPlatformCapability{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&capability).Updates(patch).Error; err != nil {
		return models.IntegrationPlatformCapability{}, err
	}
	return r.GetPlatformCapability(ctx, id)
}

func (r *Repository) CreateWebhookEvent(ctx context.Context, event models.IntegrationWebhookEvent) (models.IntegrationWebhookEvent, error) {
	now := time.Now()
	event.ReceivedAt = now
	event.CreatedAt = now
	event.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&event).Error
	if err != nil && isUniqueConstraintError(err) {
		return models.IntegrationWebhookEvent{}, ErrWebhookEventExists
	}
	return event, err
}

func (r *Repository) ListDueWebhookEvents(ctx context.Context, limit int) ([]models.IntegrationWebhookEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	now := time.Now()
	var rows []models.IntegrationWebhookEvent
	err := r.active(ctx, &models.IntegrationWebhookEvent{}).
		Where("status IN ?", []string{"received", "retrying"}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("received_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdateWebhookEvent(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationWebhookEvent, error) {
	var event models.IntegrationWebhookEvent
	err := r.active(ctx, &models.IntegrationWebhookEvent{}).Where("id = ?", id).First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationWebhookEvent{}, ErrWebhookEventNotFound
	}
	if err != nil {
		return models.IntegrationWebhookEvent{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&event).Updates(patch).Error; err != nil {
		return models.IntegrationWebhookEvent{}, err
	}
	err = r.active(ctx, &models.IntegrationWebhookEvent{}).Where("id = ?", id).First(&event).Error
	return event, err
}

func (r *Repository) UpdateProviderApp(ctx context.Context, code string, patch map[string]interface{}) (models.IntegrationProviderApp, error) {
	app, err := r.GetProviderAppByCode(ctx, code)
	if err != nil {
		return models.IntegrationProviderApp{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&app).Updates(patch).Error; err != nil {
		return models.IntegrationProviderApp{}, err
	}
	return r.GetProviderAppByCode(ctx, code)
}

func (r *Repository) UpdateAppCapability(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationProviderAppCapability, error) {
	var capability models.IntegrationProviderAppCapability
	err := r.active(ctx, &models.IntegrationProviderAppCapability{}).Where("id = ?", id).First(&capability).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationProviderAppCapability{}, ErrCapabilityNotFound
	}
	if err != nil {
		return models.IntegrationProviderAppCapability{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&capability).Updates(patch).Error; err != nil {
		return models.IntegrationProviderAppCapability{}, err
	}
	err = r.active(ctx, &models.IntegrationProviderAppCapability{}).Where("id = ?", id).First(&capability).Error
	return capability, err
}

func (r *Repository) GetAppCapability(ctx context.Context, id uint64) (models.IntegrationProviderAppCapability, error) {
	var capability models.IntegrationProviderAppCapability
	err := r.active(ctx, &models.IntegrationProviderAppCapability{}).Where("id = ?", id).First(&capability).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationProviderAppCapability{}, ErrCapabilityNotFound
	}
	return capability, err
}

func (r *Repository) UpdatePlatform(ctx context.Context, code string, patch map[string]interface{}) (models.IntegrationPlatform, error) {
	platform, err := r.GetPlatformByCode(ctx, code)
	if err != nil {
		return models.IntegrationPlatform{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&platform).Updates(patch).Error; err != nil {
		return models.IntegrationPlatform{}, err
	}
	return r.GetPlatformByCode(ctx, code)
}

func (r *Repository) ListProviderApps(ctx context.Context, opts ListOptions) ([]ProviderAppSummary, int64, error) {
	var rows []ProviderAppSummary
	opts = opts.normalized()
	today := time.Now().Truncate(24 * time.Hour)
	base := r.db.WithContext(ctx).
		Table("integration_provider_apps pa").
		Joins("JOIN integration_platforms p ON p.id = pa.platform_id AND p.deleted_at IS NULL").
		Where("pa.deleted_at IS NULL")
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		base = base.Where("(pa.app_code LIKE ? OR pa.app_name LIKE ? OR p.platform_name LIKE ?)", like, like, like)
	}
	if opts.Status != "" {
		base = base.Where("pa.status = ?", opts.Status)
	}
	if opts.PlatformCode != "" {
		base = base.Where("p.platform_code = ?", opts.PlatformCode)
	}
	base = applyTimeRange(base, "pa.updated_at", opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.
		Select(`
			pa.id,
			pa.platform_id,
			p.platform_name,
			pa.app_code,
			pa.app_name,
			pa.app_type,
			pa.auth_mode,
			pa.environment,
			pa.status,
			pa.tenant_visible,
			pa.callback_url,
			pa.webhook_url,
			pa.credential_ref,
			pa.owner_name,
			pa.description,
			COUNT(DISTINCT pac.id) AS capability_count,
			COUNT(DISTINCT tc.id) AS connection_count,
			COUNT(DISTINCT al.id) AS alert_count,
			COALESCE(ls.calls_today, 0) AS calls_today`).
		Joins("LEFT JOIN integration_provider_app_capabilities pac ON pac.provider_app_id = pa.id AND pac.deleted_at IS NULL").
		Joins("LEFT JOIN integration_tenant_connections tc ON tc.provider_app_id = pa.id AND tc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_alerts al ON al.provider_app_id = pa.id AND al.status <> ? AND al.deleted_at IS NULL", "resolved").
		Joins(`LEFT JOIN (
			SELECT provider_app_id, COUNT(*) AS calls_today
			FROM integration_api_call_logs
			WHERE provider_app_id IS NOT NULL AND called_at >= ?
			GROUP BY provider_app_id
		) ls ON ls.provider_app_id = pa.id`, today).
		Group("pa.id, pa.platform_id, p.platform_name, pa.app_code, pa.app_name, pa.app_type, pa.auth_mode, pa.environment, pa.status, pa.tenant_visible, pa.callback_url, pa.webhook_url, pa.credential_ref, pa.owner_name, pa.description, p.sort_order, ls.calls_today").
		Order(orderClause(opts, "p.sort_order ASC, pa.id ASC", map[string]string{"created_at": "pa.created_at", "updated_at": "pa.updated_at", "name": "pa.app_name", "status": "pa.status"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListPlatformCapabilities(ctx context.Context, opts ListOptions) ([]PlatformCapabilitySummary, int64, error) {
	var rows []PlatformCapabilitySummary
	opts = opts.normalized()
	base := r.db.WithContext(ctx).
		Table("integration_platform_capabilities pc").
		Joins("JOIN integration_platforms p ON p.id = pc.platform_id AND p.deleted_at IS NULL").
		Where("pc.deleted_at IS NULL")
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		base = base.Where("(pc.capability_code LIKE ? OR pc.capability_name LIKE ? OR p.platform_name LIKE ?)", like, like, like)
	}
	if opts.Status != "" {
		base = base.Where("pc.status = ?", opts.Status)
	}
	if opts.PlatformCode != "" {
		base = base.Where("p.platform_code = ?", opts.PlatformCode)
	}
	base = applyTimeRange(base, "pc.updated_at", opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.
		Select(`
			pc.id,
			pc.platform_id,
			p.platform_name,
			pc.capability_code,
			pc.capability_name,
			pc.capability_type,
			pc.auth_scope_code,
			pc.data_direction,
			pc.status,
			pc.description`).
		Order(orderClause(opts, "p.sort_order ASC, pc.id ASC", map[string]string{"created_at": "pc.created_at", "updated_at": "pc.updated_at", "name": "pc.capability_name", "status": "pc.status"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListProviderAppCapabilities(ctx context.Context, opts ListOptions) ([]ProviderAppCapabilitySummary, int64, error) {
	var rows []ProviderAppCapabilitySummary
	opts = opts.normalized()
	base := r.db.WithContext(ctx).
		Table("integration_provider_app_capabilities pac").
		Joins("JOIN integration_provider_apps pa ON pa.id = pac.provider_app_id AND pa.deleted_at IS NULL").
		Joins("JOIN integration_platforms p ON p.id = pa.platform_id AND p.deleted_at IS NULL").
		Joins("JOIN integration_platform_capabilities pc ON pc.id = pac.platform_capability_id AND pc.deleted_at IS NULL").
		Where("pac.deleted_at IS NULL")
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		base = base.Where("(pa.app_code LIKE ? OR pa.app_name LIKE ? OR pc.capability_code LIKE ? OR pc.capability_name LIKE ?)", like, like, like, like)
	}
	if opts.Status != "" {
		base = base.Where("pac.connection_status = ? OR pac.review_status = ?", opts.Status, opts.Status)
	}
	if opts.PlatformCode != "" {
		base = base.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		base = base.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	base = applyTimeRange(base, "pac.updated_at", opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.
		Select(`
			pac.id,
			pac.provider_app_id,
			pa.app_code AS provider_app_code,
			pa.app_name AS provider_app_name,
			p.id AS platform_id,
			p.platform_name,
			pc.id AS platform_capability_id,
			pc.capability_code,
			pc.capability_name,
			pc.capability_type,
			pac.connection_status,
			pac.review_status,
			pac.enabled,
			pac.config,
			pa.tenant_visible AS open_to_tenant,
			pac.enabled AS default_enabled,
			(pc.capability_type NOT IN ('authorization', '基础能力')) AS tenant_configurable`).
		Order(orderClause(opts, "p.sort_order ASC, pa.id ASC, pc.id ASC", map[string]string{"created_at": "pac.created_at", "updated_at": "pac.updated_at", "status": "pac.connection_status"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListTenantConnections(ctx context.Context, opts ListOptions) ([]TenantConnectionSummary, int64, error) {
	var rows []TenantConnectionSummary
	opts = opts.normalized()
	today := time.Now().Truncate(24 * time.Hour)
	tenantNameSelect := "'租户 ' || CAST(tc.tenant_id AS TEXT) AS tenant_name"
	tenantNameGroup := ""
	query := r.db.WithContext(ctx).
		Table("integration_tenant_connections tc").
		Select(`
			tc.id,
			tc.tenant_id,
			` + tenantNameSelect + `,
			tc.platform_id,
			p.platform_name,
			tc.provider_app_id,
			pa.app_name AS provider_app_name,
			tc.connection_name,
			tc.auth_subject_type,
			tc.auth_subject_id,
			tc.auth_subject_name,
			tc.auth_scope,
			tc.auth_status,
			tc.connection_status,
			tc.token_status,
			tc.token_credential_ref,
			tc.last_sync_at,
			tc.last_error_message,
			COUNT(DISTINCT tcap.id) AS final_capability_count,
			COALESCE(ls.calls_today, 0) AS calls_today,
			'[]' AS final_capabilities`).
		Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL")
	if r.db.Migrator().HasTable(&models.Tenant{}) {
		tenantNameSelect = "COALESCE(t.name, '租户 ' || CAST(tc.tenant_id AS TEXT)) AS tenant_name"
		tenantNameGroup = ", t.name"
		query = r.db.WithContext(ctx).
			Table("integration_tenant_connections tc").
			Select(`
				tc.id,
				tc.tenant_id,
				` + tenantNameSelect + `,
				tc.platform_id,
				p.platform_name,
				tc.provider_app_id,
				pa.app_name AS provider_app_name,
				tc.connection_name,
				tc.auth_subject_type,
				tc.auth_subject_id,
				tc.auth_subject_name,
				tc.auth_scope,
				tc.auth_status,
				tc.connection_status,
				tc.token_status,
				tc.token_credential_ref,
				tc.last_sync_at,
				tc.last_error_message,
				COUNT(DISTINCT tcap.id) AS final_capability_count,
				COALESCE(ls.calls_today, 0) AS calls_today,
				'[]' AS final_capabilities`).
			Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
			Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
			Joins("LEFT JOIN tenant t ON t.id = tc.tenant_id AND t.deleted_at IS NULL")
	}
	query = query.
		Joins("LEFT JOIN integration_tenant_capabilities tcap ON tcap.tenant_connection_id = tc.id AND tcap.deleted_at IS NULL").
		Joins(`LEFT JOIN (
			SELECT tenant_connection_id, COUNT(*) AS calls_today
			FROM integration_api_call_logs
			WHERE tenant_connection_id IS NOT NULL AND called_at >= ?
			GROUP BY tenant_connection_id
		) ls ON ls.tenant_connection_id = tc.id`, today).
		Where("tc.deleted_at IS NULL")
	if opts.TenantID != nil {
		query = query.Where("tc.tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("(tc.connection_name LIKE ? OR tc.auth_subject_id LIKE ? OR tc.auth_subject_name LIKE ? OR p.platform_name LIKE ? OR pa.app_name LIKE ?)", like, like, like, like, like)
	}
	if opts.Status != "" {
		query = query.Where("tc.connection_status = ? OR tc.auth_status = ? OR tc.token_status = ?", opts.Status, opts.Status, opts.Status)
	}
	if opts.PlatformCode != "" {
		query = query.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		query = query.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	query = applyTimeRange(query, "tc.updated_at", opts)
	// Count with filters and joins kept separate to avoid GROUP BY side effects.
	countQuery := r.db.WithContext(ctx).Table("integration_tenant_connections tc").
		Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
		Where("tc.deleted_at IS NULL")
	if opts.TenantID != nil {
		countQuery = countQuery.Where("tc.tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		countQuery = countQuery.Where("(tc.connection_name LIKE ? OR tc.auth_subject_id LIKE ? OR tc.auth_subject_name LIKE ? OR p.platform_name LIKE ? OR pa.app_name LIKE ?)", like, like, like, like, like)
	}
	if opts.Status != "" {
		countQuery = countQuery.Where("tc.connection_status = ? OR tc.auth_status = ? OR tc.token_status = ?", opts.Status, opts.Status, opts.Status)
	}
	if opts.PlatformCode != "" {
		countQuery = countQuery.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		countQuery = countQuery.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	countQuery = applyTimeRange(countQuery, "tc.updated_at", opts)
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Group("tc.id, tc.tenant_id" + tenantNameGroup + ", tc.platform_id, p.platform_name, tc.provider_app_id, pa.app_name, tc.connection_name, tc.auth_subject_type, tc.auth_subject_id, tc.auth_subject_name, tc.auth_scope, tc.auth_status, tc.connection_status, tc.token_status, tc.token_credential_ref, tc.last_sync_at, tc.last_error_message, ls.calls_today").
		Order(orderClause(opts, "tc.updated_at DESC, tc.id DESC", map[string]string{"created_at": "tc.created_at", "updated_at": "tc.updated_at", "last_sync_at": "tc.last_sync_at", "status": "tc.connection_status"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListSyncJobs(ctx context.Context, opts ListOptions) ([]SyncJobSummary, int64, error) {
	var rows []SyncJobSummary
	opts = opts.normalized()
	tenantNameSelect := "'租户 ' || CAST(j.tenant_id AS TEXT) AS tenant_name"
	tenantNameGroup := ""
	query := r.db.WithContext(ctx).
		Table("integration_sync_jobs j").
		Select(`
			j.id,
			j.tenant_id,
			` + tenantNameSelect + `,
			j.tenant_connection_id,
			tc.platform_id,
			p.platform_name,
			tc.provider_app_id,
			pa.app_name AS provider_app_name,
			tc.connection_name,
			j.capability_code,
			COALESCE(pc.capability_name, j.capability_code) AS capability_name,
			j.job_type,
			j.trigger_mode,
			j.status,
			j.total_count,
			j.success_count,
			j.failed_count,
			j.retry_count,
			j.next_retry_at,
			j.started_at,
			j.finished_at,
			j.error_code,
			j.error_message`).
		Joins("JOIN integration_tenant_connections tc ON tc.id = j.tenant_connection_id AND tc.deleted_at IS NULL").
		Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
		Joins("LEFT JOIN integration_platform_capabilities pc ON pc.platform_id = tc.platform_id AND pc.capability_code = j.capability_code AND pc.deleted_at IS NULL")
	if r.db.Migrator().HasTable(&models.Tenant{}) {
		tenantNameSelect = "COALESCE(t.name, '租户 ' || CAST(j.tenant_id AS TEXT)) AS tenant_name"
		tenantNameGroup = ", t.name"
		query = r.db.WithContext(ctx).
			Table("integration_sync_jobs j").
			Select(`
				j.id,
				j.tenant_id,
				` + tenantNameSelect + `,
				j.tenant_connection_id,
				tc.platform_id,
				p.platform_name,
				tc.provider_app_id,
				pa.app_name AS provider_app_name,
				tc.connection_name,
				j.capability_code,
				COALESCE(pc.capability_name, j.capability_code) AS capability_name,
				j.job_type,
				j.trigger_mode,
				j.status,
				j.total_count,
				j.success_count,
				j.failed_count,
				j.retry_count,
				j.next_retry_at,
				j.started_at,
				j.finished_at,
				j.error_code,
				j.error_message`).
			Joins("JOIN integration_tenant_connections tc ON tc.id = j.tenant_connection_id AND tc.deleted_at IS NULL").
			Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
			Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
			Joins("LEFT JOIN integration_platform_capabilities pc ON pc.platform_id = tc.platform_id AND pc.capability_code = j.capability_code AND pc.deleted_at IS NULL").
			Joins("LEFT JOIN tenant t ON t.id = j.tenant_id AND t.deleted_at IS NULL")
	}
	query = query.Where("j.deleted_at IS NULL")
	if opts.TenantID != nil {
		query = query.Where("j.tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("(tc.connection_name LIKE ? OR j.capability_code LIKE ? OR p.platform_name LIKE ? OR pa.app_name LIKE ?)", like, like, like, like)
	}
	if opts.Status != "" {
		query = query.Where("j.status = ?", opts.Status)
	}
	if opts.PlatformCode != "" {
		query = query.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		query = query.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	query = applyTimeRange(query, "j.created_at", opts)
	countQuery := r.db.WithContext(ctx).Table("integration_sync_jobs j").
		Joins("JOIN integration_tenant_connections tc ON tc.id = j.tenant_connection_id AND tc.deleted_at IS NULL").
		Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
		Where("j.deleted_at IS NULL")
	if opts.TenantID != nil {
		countQuery = countQuery.Where("j.tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		countQuery = countQuery.Where("(tc.connection_name LIKE ? OR j.capability_code LIKE ? OR p.platform_name LIKE ? OR pa.app_name LIKE ?)", like, like, like, like)
	}
	if opts.Status != "" {
		countQuery = countQuery.Where("j.status = ?", opts.Status)
	}
	if opts.PlatformCode != "" {
		countQuery = countQuery.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		countQuery = countQuery.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	countQuery = applyTimeRange(countQuery, "j.created_at", opts)
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Group("j.id, j.tenant_id" + tenantNameGroup + ", j.tenant_connection_id, tc.platform_id, p.platform_name, tc.provider_app_id, pa.app_name, tc.connection_name, j.capability_code, pc.capability_name, j.job_type, j.trigger_mode, j.status, j.total_count, j.success_count, j.failed_count, j.retry_count, j.next_retry_at, j.started_at, j.finished_at, j.error_code, j.error_message").
		Order(orderClause(opts, "j.created_at DESC, j.id DESC", map[string]string{"created_at": "j.created_at", "updated_at": "j.updated_at", "started_at": "j.started_at", "finished_at": "j.finished_at", "status": "j.status"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) CreateSyncJob(ctx context.Context, job models.IntegrationSyncJob) (models.IntegrationSyncJob, error) {
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&job).Error
	return job, err
}

func (r *Repository) UpsertSyncRecords(ctx context.Context, records []models.IntegrationSyncRecord) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}
	now := time.Now()
	var affected int64
	for i := range records {
		record := records[i]
		record.UpdatedAt = now
		if record.CreatedAt.IsZero() {
			record.CreatedAt = now
		}
		if record.WrittenAt.IsZero() {
			record.WrittenAt = now
		}
		if record.Status == "" {
			record.Status = "written"
		}
		var existing models.IntegrationSyncRecord
		err := r.active(ctx, &models.IntegrationSyncRecord{}).
			Where("tenant_connection_id = ? AND capability_code = ? AND external_id = ?", record.TenantConnectionID, record.CapabilityCode, record.ExternalID).
			First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
				return affected, err
			}
			affected++
			continue
		}
		if err != nil {
			return affected, err
		}
		if err := r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"sync_job_id":    record.SyncJobID,
			"tenant_id":      record.TenantID,
			"payload_digest": record.PayloadDigest,
			"payload":        record.Payload,
			"status":         record.Status,
			"cursor_value":   record.CursorValue,
			"written_at":     record.WrittenAt,
			"updated_at":     now,
		}).Error; err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *Repository) ListQuotaPolicies(ctx context.Context, opts ListOptions) ([]models.IntegrationQuotaPolicy, int64, error) {
	var rows []models.IntegrationQuotaPolicy
	opts = opts.normalized()
	query := r.active(ctx, &models.IntegrationQuotaPolicy{})
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("(policy_code LIKE ? OR policy_name LIKE ? OR quota_code LIKE ?)", like, like, like)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	query = applyTimeRange(query, "updated_at", opts)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order(orderClause(opts, "id ASC", map[string]string{"created_at": "created_at", "updated_at": "updated_at", "name": "policy_name", "status": "status"})).Offset(opts.Skip).Limit(opts.Limit).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListDueSyncJobs(ctx context.Context, limit int) ([]models.IntegrationSyncJob, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	now := time.Now()
	var rows []models.IntegrationSyncJob
	err := r.active(ctx, &models.IntegrationSyncJob{}).
		Where("status IN ?", []string{"pending", "retrying"}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListQuotaUsages(ctx context.Context, opts ListOptions) ([]QuotaUsageSummary, int64, error) {
	var rows []QuotaUsageSummary
	opts = opts.normalized()
	tenantNameSelect := "'租户 ' || CAST(u.tenant_id AS TEXT) AS tenant_name"
	tenantNameGroup := ""
	query := r.db.WithContext(ctx).
		Table("integration_quota_usages u").
		Select(`
			u.id,
			u.tenant_id,
			` + tenantNameSelect + `,
			u.tenant_connection_id,
			tc.connection_name,
			p.platform_name,
			pa.app_name AS provider_app_name,
			u.quota_code,
			MIN(qp.policy_code) AS policy_code,
			MIN(qp.policy_name) AS policy_name,
			u.period_key,
			u.used_amount,
			COALESCE(MAX(qp.default_limit), 0) AS limit_amount,
			u.limited_count,
			u.last_used_at`).
		Joins("LEFT JOIN integration_tenant_connections tc ON tc.id = u.tenant_connection_id AND tc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
		Joins("LEFT JOIN integration_quota_policies qp ON qp.quota_code = u.quota_code AND qp.deleted_at IS NULL")
	if r.db.Migrator().HasTable(&models.Tenant{}) {
		tenantNameSelect = "COALESCE(t.name, '租户 ' || CAST(u.tenant_id AS TEXT)) AS tenant_name"
		tenantNameGroup = ", t.name"
		query = r.db.WithContext(ctx).
			Table("integration_quota_usages u").
			Select(`
				u.id,
				u.tenant_id,
				` + tenantNameSelect + `,
				u.tenant_connection_id,
				tc.connection_name,
				p.platform_name,
				pa.app_name AS provider_app_name,
				u.quota_code,
				MIN(qp.policy_code) AS policy_code,
				MIN(qp.policy_name) AS policy_name,
				u.period_key,
				u.used_amount,
				COALESCE(MAX(qp.default_limit), 0) AS limit_amount,
				u.limited_count,
				u.last_used_at`).
			Joins("LEFT JOIN tenant t ON t.id = u.tenant_id AND t.deleted_at IS NULL").
			Joins("LEFT JOIN integration_tenant_connections tc ON tc.id = u.tenant_connection_id AND tc.deleted_at IS NULL").
			Joins("LEFT JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
			Joins("LEFT JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
			Joins("LEFT JOIN integration_quota_policies qp ON qp.quota_code = u.quota_code AND qp.deleted_at IS NULL")
	}
	query = query
	if opts.TenantID != nil {
		query = query.Where("u.tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("(u.quota_code LIKE ? OR qp.policy_name LIKE ? OR tc.connection_name LIKE ?)", like, like, like)
	}
	if opts.PlatformCode != "" {
		query = query.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		query = query.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	query = applyTimeRange(query, "u.last_used_at", opts)
	countQuery := r.db.WithContext(ctx).Table("integration_quota_usages u").
		Joins("LEFT JOIN integration_tenant_connections tc ON tc.id = u.tenant_connection_id AND tc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
		Joins("LEFT JOIN integration_quota_policies qp ON qp.quota_code = u.quota_code AND qp.deleted_at IS NULL")
	if opts.TenantID != nil {
		countQuery = countQuery.Where("u.tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		countQuery = countQuery.Where("(u.quota_code LIKE ? OR qp.policy_name LIKE ? OR tc.connection_name LIKE ?)", like, like, like)
	}
	if opts.PlatformCode != "" {
		countQuery = countQuery.Where("p.platform_code = ?", opts.PlatformCode)
	}
	if opts.ProviderAppCode != "" {
		countQuery = countQuery.Where("pa.app_code = ?", opts.ProviderAppCode)
	}
	countQuery = applyTimeRange(countQuery, "u.last_used_at", opts)
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Group("u.id, u.tenant_id" + tenantNameGroup + ", u.tenant_connection_id, tc.connection_name, p.platform_name, pa.app_name, u.quota_code, u.period_key, u.used_amount, u.limited_count, u.last_used_at").
		Order(orderClause(opts, "u.last_used_at DESC, u.id DESC", map[string]string{"last_used_at": "u.last_used_at", "used_amount": "u.used_amount", "limited_count": "u.limited_count"})).
		Offset(opts.Skip).
		Limit(opts.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListAlerts(ctx context.Context, opts ListOptions) ([]models.IntegrationAlert, int64, error) {
	var rows []models.IntegrationAlert
	opts = opts.normalized()
	query := r.active(ctx, &models.IntegrationAlert{})
	if opts.TenantID != nil {
		query = query.Where("tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("(title LIKE ? OR alert_type LIKE ? OR severity LIKE ?)", like, like, like)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	query = applyTimeRange(query, "last_seen_at", opts)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order(orderClause(opts, "last_seen_at DESC, id DESC", map[string]string{"created_at": "created_at", "last_seen_at": "last_seen_at", "severity": "severity", "status": "status"})).Offset(opts.Skip).Limit(opts.Limit).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListAPICallLogs(ctx context.Context, opts ListOptions) ([]models.IntegrationAPICallLog, int64, error) {
	var rows []models.IntegrationAPICallLog
	opts = opts.normalized()
	query := r.db.WithContext(ctx).Model(&models.IntegrationAPICallLog{}).Where("archived_at IS NULL")
	if opts.TenantID != nil {
		query = query.Where("tenant_id = ?", *opts.TenantID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("(request_id LIKE ? OR trace_id LIKE ? OR call_type LIKE ? OR endpoint LIKE ? OR error_code LIKE ?)", like, like, like, like, like)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	query = applyTimeRange(query, "called_at", opts)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order(orderClause(opts, "called_at DESC, id DESC", map[string]string{"called_at": "called_at", "duration_ms": "duration_ms", "status": "status", "http_status": "http_status"})).Offset(opts.Skip).Limit(opts.Limit).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ArchiveAPICallLogsBefore(ctx context.Context, cutoff time.Time, bucket string, limit int) (int64, error) {
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}
	now := time.Now()
	result := r.db.WithContext(ctx).Exec(`
		UPDATE integration_api_call_logs
		SET archived_at = ?, retention_bucket = ?
		WHERE id IN (
			SELECT id
			FROM integration_api_call_logs
			WHERE archived_at IS NULL AND called_at < ?
			ORDER BY called_at ASC, id ASC
			LIMIT ?
		)
	`, now, bucket, cutoff, limit)
	return result.RowsAffected, result.Error
}

func (r *Repository) GetAPICallLog(ctx context.Context, id uint64) (models.IntegrationAPICallLog, error) {
	var log models.IntegrationAPICallLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationAPICallLog{}, ErrAPICallLogNotFound
	}
	return log, err
}

func (r *Repository) ListTenantCapabilitiesByConnection(ctx context.Context, tenantConnectionID uint64) ([]models.IntegrationTenantCapability, error) {
	var rows []models.IntegrationTenantCapability
	err := r.active(ctx, &models.IntegrationTenantCapability{}).
		Where("tenant_connection_id = ?", tenantConnectionID).
		Order("capability_code ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListSyncJobsByConnection(ctx context.Context, tenantConnectionID uint64, limit int) ([]models.IntegrationSyncJob, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []models.IntegrationSyncJob
	err := r.active(ctx, &models.IntegrationSyncJob{}).
		Where("tenant_connection_id = ?", tenantConnectionID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListQuotaUsagesByConnection(ctx context.Context, tenantConnectionID uint64, limit int) ([]models.IntegrationQuotaUsage, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []models.IntegrationQuotaUsage
	err := r.db.WithContext(ctx).
		Where("tenant_connection_id = ?", tenantConnectionID).
		Order("last_used_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAPICallLogsByConnection(ctx context.Context, tenantConnectionID uint64, limit int) ([]models.IntegrationAPICallLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []models.IntegrationAPICallLog
	err := r.db.WithContext(ctx).
		Where("tenant_connection_id = ?", tenantConnectionID).
		Order("called_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAPICallLogsByProviderApp(ctx context.Context, providerAppID uint64, limit int) ([]models.IntegrationAPICallLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []models.IntegrationAPICallLog
	err := r.db.WithContext(ctx).
		Where("provider_app_id = ?", providerAppID).
		Order("called_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdateTenantConnection(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationTenantConnection, error) {
	var connection models.IntegrationTenantConnection
	err := r.active(ctx, &models.IntegrationTenantConnection{}).Where("id = ?", id).First(&connection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationTenantConnection{}, ErrConnectionNotFound
	}
	if err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&connection).Updates(patch).Error; err != nil {
		return models.IntegrationTenantConnection{}, err
	}
	err = r.active(ctx, &models.IntegrationTenantConnection{}).Where("id = ?", id).First(&connection).Error
	return connection, err
}

func (r *Repository) GetTenantConnection(ctx context.Context, id uint64) (models.IntegrationTenantConnection, error) {
	var connection models.IntegrationTenantConnection
	err := r.active(ctx, &models.IntegrationTenantConnection{}).Where("id = ?", id).First(&connection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationTenantConnection{}, ErrConnectionNotFound
	}
	return connection, err
}

func (r *Repository) UpdateSyncJob(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationSyncJob, error) {
	var job models.IntegrationSyncJob
	err := r.active(ctx, &models.IntegrationSyncJob{}).Where("id = ?", id).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationSyncJob{}, ErrSyncJobNotFound
	}
	if err != nil {
		return models.IntegrationSyncJob{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&job).Updates(patch).Error; err != nil {
		return models.IntegrationSyncJob{}, err
	}
	err = r.active(ctx, &models.IntegrationSyncJob{}).Where("id = ?", id).First(&job).Error
	return job, err
}

func (r *Repository) GetSyncJob(ctx context.Context, id uint64) (models.IntegrationSyncJob, error) {
	var job models.IntegrationSyncJob
	err := r.active(ctx, &models.IntegrationSyncJob{}).Where("id = ?", id).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationSyncJob{}, ErrSyncJobNotFound
	}
	return job, err
}

func (r *Repository) GetQuotaPolicyByCode(ctx context.Context, code string) (models.IntegrationQuotaPolicy, error) {
	var policy models.IntegrationQuotaPolicy
	err := r.active(ctx, &models.IntegrationQuotaPolicy{}).Where("policy_code = ?", code).First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationQuotaPolicy{}, ErrQuotaPolicyNotFound
	}
	return policy, err
}

func (r *Repository) CreateQuotaPolicy(ctx context.Context, policy models.IntegrationQuotaPolicy) (models.IntegrationQuotaPolicy, error) {
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&policy).Error
	return policy, err
}

func (r *Repository) UpsertQuotaBinding(ctx context.Context, binding models.IntegrationQuotaBinding) (models.IntegrationQuotaBinding, error) {
	now := time.Now()
	var existing models.IntegrationQuotaBinding
	err := r.db.WithContext(ctx).
		Where("policy_id = ? AND deleted_at IS NULL", binding.PolicyID).
		Order("id desc").
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		binding.CreatedAt = now
		binding.UpdatedAt = now
		if binding.Status == "" {
			binding.Status = "enabled"
		}
		if err := r.db.WithContext(ctx).Create(&binding).Error; err != nil {
			return models.IntegrationQuotaBinding{}, err
		}
		return binding, nil
	}
	if err != nil {
		return models.IntegrationQuotaBinding{}, err
	}
	patch := map[string]interface{}{
		"tenant_id":            binding.TenantID,
		"platform_id":          binding.PlatformID,
		"provider_app_id":      binding.ProviderAppID,
		"tenant_connection_id": binding.TenantConnectionID,
		"override_limit":       binding.OverrideLimit,
		"priority":             binding.Priority,
		"status":               binding.Status,
		"updated_by":           binding.UpdatedBy,
		"updated_at":           now,
	}
	if patch["status"] == "" {
		patch["status"] = "enabled"
	}
	if err := r.db.WithContext(ctx).Model(&existing).Updates(patch).Error; err != nil {
		return models.IntegrationQuotaBinding{}, err
	}
	err = r.db.WithContext(ctx).Where("id = ?", existing.ID).First(&existing).Error
	return existing, err
}

func (r *Repository) UpdateQuotaPolicy(ctx context.Context, code string, patch map[string]interface{}) (models.IntegrationQuotaPolicy, error) {
	policy, err := r.GetQuotaPolicyByCode(ctx, code)
	if err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&policy).Updates(patch).Error; err != nil {
		return models.IntegrationQuotaPolicy{}, err
	}
	return r.GetQuotaPolicyByCode(ctx, code)
}

func (r *Repository) UpdateAlert(ctx context.Context, id uint64, patch map[string]interface{}) (models.IntegrationAlert, error) {
	var alert models.IntegrationAlert
	err := r.active(ctx, &models.IntegrationAlert{}).Where("id = ?", id).First(&alert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationAlert{}, ErrAlertNotFound
	}
	if err != nil {
		return models.IntegrationAlert{}, err
	}
	patch["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Model(&alert).Updates(patch).Error; err != nil {
		return models.IntegrationAlert{}, err
	}
	err = r.active(ctx, &models.IntegrationAlert{}).Where("id = ?", id).First(&alert).Error
	return alert, err
}

func (r *Repository) CreateAlert(ctx context.Context, alert models.IntegrationAlert) (models.IntegrationAlert, error) {
	now := time.Now()
	if alert.FirstSeenAt.IsZero() {
		alert.FirstSeenAt = now
	}
	if alert.LastSeenAt.IsZero() {
		alert.LastSeenAt = now
	}
	alert.CreatedAt = now
	alert.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&alert).Error
	return alert, err
}

func (r *Repository) GetAlert(ctx context.Context, id uint64) (models.IntegrationAlert, error) {
	var alert models.IntegrationAlert
	err := r.active(ctx, &models.IntegrationAlert{}).Where("id = ?", id).First(&alert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.IntegrationAlert{}, ErrAlertNotFound
	}
	return alert, err
}

func (r *Repository) active(ctx context.Context, model interface{}) *gorm.DB {
	return r.db.WithContext(ctx).Model(model).Where("deleted_at IS NULL")
}

func applyTimeRange(query *gorm.DB, column string, opts ListOptions) *gorm.DB {
	if opts.StartTime != nil {
		query = query.Where(column+" >= ?", *opts.StartTime)
	}
	if opts.EndTime != nil {
		query = query.Where(column+" <= ?", *opts.EndTime)
	}
	return query
}

func orderClause(opts ListOptions, fallback string, allowed map[string]string) string {
	column, ok := allowed[opts.SortBy]
	if !ok || column == "" {
		return fallback
	}
	direction := "DESC"
	if opts.SortOrder == "asc" {
		direction = "ASC"
	}
	return column + " " + direction
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unique") || strings.Contains(text, "duplicate")
}

func quotaBindingScope(binding models.IntegrationQuotaBinding) string {
	switch {
	case binding.TenantConnectionID != nil:
		return "tenant_connection"
	case binding.ProviderAppID != nil:
		return "provider_app"
	case binding.PlatformID != nil:
		return "platform"
	case binding.TenantID != nil:
		return "tenant"
	default:
		return "global"
	}
}
