package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

var (
	ErrPlatformNotFound    = errors.New("integration platform not found")
	ErrProviderAppNotFound = errors.New("integration provider app not found")
	ErrCapabilityNotFound  = errors.New("integration app capability not found")
	ErrConnectionNotFound  = errors.New("integration tenant connection not found")
	ErrSyncJobNotFound     = errors.New("integration sync job not found")
	ErrQuotaPolicyNotFound = errors.New("integration quota policy not found")
	ErrAlertNotFound       = errors.New("integration alert not found")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
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
	Status          string  `json:"status"`
	TenantVisible   bool    `json:"tenant_visible"`
	OwnerName       *string `json:"owner_name"`
	AppCount        int64   `json:"app_count"`
	CapabilityCount int64   `json:"capability_count"`
	ConnectionCount int64   `json:"connection_count"`
	OpenAlertCount  int64   `json:"open_alert_count"`
}

type ProviderAppSummary struct {
	ID            uint64 `json:"id"`
	PlatformID    uint64 `json:"platform_id"`
	PlatformName  string `json:"platform_name"`
	AppCode       string `json:"app_code"`
	AppName       string `json:"app_name"`
	AuthMode      string `json:"auth_mode"`
	Environment   string `json:"environment"`
	Status        string `json:"status"`
	TenantVisible bool   `json:"tenant_visible"`
}

type TenantConnectionSummary struct {
	ID               uint64     `json:"id"`
	TenantID         uint64     `json:"tenant_id"`
	PlatformName     string     `json:"platform_name"`
	ProviderAppName  string     `json:"provider_app_name"`
	ConnectionName   string     `json:"connection_name"`
	AuthSubjectType  string     `json:"auth_subject_type"`
	AuthSubjectID    string     `json:"auth_subject_id"`
	AuthSubjectName  string     `json:"auth_subject_name"`
	AuthStatus       string     `json:"auth_status"`
	ConnectionStatus string     `json:"connection_status"`
	TokenStatus      string     `json:"token_status"`
	LastSyncAt       *time.Time `json:"last_sync_at"`
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

func (r *Repository) ListPlatforms(ctx context.Context, limit int) ([]PlatformSummary, error) {
	var rows []PlatformSummary
	err := r.db.WithContext(ctx).
		Table("integration_platforms p").
		Select(`
			p.id,
			p.platform_code AS code,
			p.platform_name AS name,
			p.platform_short_name AS short_name,
			p.platform_type,
			p.access_mode,
			p.status,
			p.tenant_visible,
			p.owner_name,
			COUNT(DISTINCT pa.id) AS app_count,
			COUNT(DISTINCT pc.id) AS capability_count,
			COUNT(DISTINCT tc.id) AS connection_count,
			COUNT(DISTINCT al.id) AS open_alert_count`).
		Joins("LEFT JOIN integration_provider_apps pa ON pa.platform_id = p.id AND pa.deleted_at IS NULL").
		Joins("LEFT JOIN integration_platform_capabilities pc ON pc.platform_id = p.id AND pc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_tenant_connections tc ON tc.platform_id = p.id AND tc.deleted_at IS NULL").
		Joins("LEFT JOIN integration_alerts al ON al.platform_id = p.id AND al.status <> ? AND al.deleted_at IS NULL", "resolved").
		Where("p.deleted_at IS NULL").
		Group("p.id, p.platform_code, p.platform_name, p.platform_short_name, p.platform_type, p.access_mode, p.status, p.tenant_visible, p.owner_name, p.sort_order").
		Order("p.sort_order ASC, p.id ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) GetPlatformByCode(ctx context.Context, code string) (models.IntegrationPlatform, error) {
	var platform models.IntegrationPlatform
	err := r.active(ctx, &models.IntegrationPlatform{}).Where("platform_code = ?", code).First(&platform).Error
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

func (r *Repository) CreateProviderApp(ctx context.Context, app models.IntegrationProviderApp) (models.IntegrationProviderApp, error) {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&app).Error
	return app, err
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

func (r *Repository) ListProviderApps(ctx context.Context, limit int) ([]ProviderAppSummary, error) {
	var rows []ProviderAppSummary
	err := r.db.WithContext(ctx).
		Table("integration_provider_apps pa").
		Select(`
			pa.id,
			pa.platform_id,
			p.platform_name,
			pa.app_code,
			pa.app_name,
			pa.auth_mode,
			pa.environment,
			pa.status,
			pa.tenant_visible`).
		Joins("JOIN integration_platforms p ON p.id = pa.platform_id AND p.deleted_at IS NULL").
		Where("pa.deleted_at IS NULL").
		Order("p.sort_order ASC, pa.id ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) ListTenantConnections(ctx context.Context, limit int) ([]TenantConnectionSummary, error) {
	var rows []TenantConnectionSummary
	err := r.db.WithContext(ctx).
		Table("integration_tenant_connections tc").
		Select(`
			tc.id,
			tc.tenant_id,
			p.platform_name,
			pa.app_name AS provider_app_name,
			tc.connection_name,
			tc.auth_subject_type,
			tc.auth_subject_id,
			tc.auth_subject_name,
			tc.auth_status,
			tc.connection_status,
			tc.token_status,
			tc.last_sync_at`).
		Joins("JOIN integration_platforms p ON p.id = tc.platform_id AND p.deleted_at IS NULL").
		Joins("JOIN integration_provider_apps pa ON pa.id = tc.provider_app_id AND pa.deleted_at IS NULL").
		Where("tc.deleted_at IS NULL").
		Order("tc.updated_at DESC, tc.id DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) ListSyncJobs(ctx context.Context, limit int) ([]models.IntegrationSyncJob, error) {
	var rows []models.IntegrationSyncJob
	err := r.active(ctx, &models.IntegrationSyncJob{}).Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateSyncJob(ctx context.Context, job models.IntegrationSyncJob) (models.IntegrationSyncJob, error) {
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now
	err := r.db.WithContext(ctx).Create(&job).Error
	return job, err
}

func (r *Repository) ListQuotaPolicies(ctx context.Context, limit int) ([]models.IntegrationQuotaPolicy, error) {
	var rows []models.IntegrationQuotaPolicy
	err := r.active(ctx, &models.IntegrationQuotaPolicy{}).Order("id ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAlerts(ctx context.Context, limit int) ([]models.IntegrationAlert, error) {
	var rows []models.IntegrationAlert
	err := r.active(ctx, &models.IntegrationAlert{}).Order("last_seen_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAPICallLogs(ctx context.Context, limit int) ([]models.IntegrationAPICallLog, error) {
	var rows []models.IntegrationAPICallLog
	err := r.db.WithContext(ctx).Model(&models.IntegrationAPICallLog{}).Order("called_at DESC, id DESC").Limit(limit).Find(&rows).Error
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

func (r *Repository) active(ctx context.Context, model interface{}) *gorm.DB {
	return r.db.WithContext(ctx).Model(model).Where("deleted_at IS NULL")
}
