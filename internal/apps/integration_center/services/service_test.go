package services

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/integration_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestServiceReadsOverviewFromRepository(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	platform := models.IntegrationPlatform{
		PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "第三方服务商",
		Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{
		PlatformID: platform.ID, AppCode: "wecom-suite", AppName: "企业微信标准应用", AuthMode: "OAuth2",
		Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 1, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "鹿鸣科技企业微信",
		AuthSubjectType: "corp", AuthSubjectID: "corp-1", AuthSubjectName: "鹿鸣科技",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	require.NoError(t, db.Create(&models.IntegrationAlert{
		PlatformID: &platform.ID, ProviderAppID: &app.ID, AlertType: "token", Severity: "warning", Status: "open",
		Title: "Token 即将过期", FirstSeenAt: now, LastSeenAt: now, CreatedAt: now, UpdatedAt: now,
	}).Error)
	method := "GET"
	endpoint := "/cgi-bin/user/list"
	httpStatus := 200
	require.NoError(t, db.Create(&models.IntegrationAPICallLog{
		PlatformID: &platform.ID, ProviderAppID: &app.ID, RequestID: "req_1", CallType: "third_party_api",
		Method: &method, Endpoint: &endpoint, Status: "success", HTTPStatus: &httpStatus, DurationMS: 88,
		CalledAt: now, CreatedAt: now,
	}).Error)

	service := NewService(repositories.NewRepository(db))
	overview, err := service.Overview(context.Background())

	require.NoError(t, err)
	require.Len(t, overview.Connectors, 1)
	require.Equal(t, "wecom", overview.Connectors[0].Code)
	require.Equal(t, "1", overview.Metrics[0].Value)
	require.Equal(t, "1", overview.Metrics[3].Value)
	require.Len(t, overview.Events, 1)
}

func TestServiceListsTenantConnectionsFromRepository(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东电商", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-isv", AppName: "京东 ISV 应用", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	require.NoError(t, db.Create(&models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店连接",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}).Error)

	summary, err := NewService(repositories.NewRepository(db)).TenantConnections(context.Background())

	require.NoError(t, err)
	rows, ok := summary.Items.([]repositories.TenantConnectionSummary)
	require.True(t, ok)
	require.Len(t, rows, 1)
	require.Equal(t, uint64(9), rows[0].TenantID)
	require.Equal(t, "shop", rows[0].AuthSubjectType)
	require.Equal(t, "shop-1001", rows[0].AuthSubjectID)
}

func TestServiceCreatesAndUpdatesPlatform(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))

	created, err := service.CreatePlatform(context.Background(), PlatformMutationRequest{
		Name: "通用 ERP", Code: "generic-erp", PlatformType: "ERP", AccessMode: "API Key",
		Status: "enabled", TenantVisible: true, OwnerName: "集成组", SortOrder: 80,
	})
	require.NoError(t, err)
	require.Equal(t, "generic-erp", created.Code)
	require.Equal(t, "online", created.Status)

	updated, err := service.UpdatePlatform(context.Background(), "generic-erp", PlatformMutationRequest{
		Name: "通用 ERP 平台", PlatformType: "ERP", AccessMode: "API Key",
		Status: "maintenance", TenantVisible: false, OwnerName: "平台集成组", SortOrder: 81,
	})
	require.NoError(t, err)
	require.Equal(t, "通用 ERP 平台", updated.Name)
	require.Equal(t, "maintenance", updated.Status)
	require.False(t, updated.TenantVisible)
}

func TestServiceCreatesAndUpdatesProviderApp(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	service := NewService(repositories.NewRepository(db))

	created, err := service.CreateProviderApp(context.Background(), ProviderAppMutationRequest{
		PlatformCode: "wecom", Code: "wecom-suite-prod", Name: "企微正式应用", AuthMode: "OAuth2",
		Environment: "正式", Status: "enabled", TenantVisible: true,
	})
	require.NoError(t, err)
	require.Equal(t, "wecom-suite-prod", created.AppCode)
	require.Equal(t, "online", created.Status)

	updated, err := service.UpdateProviderApp(context.Background(), "wecom-suite-prod", ProviderAppMutationRequest{
		PlatformCode: "wecom", Name: "企微正式应用 V2", AuthMode: "OAuth2",
		Environment: "test", Status: "testing", TenantVisible: false,
	})
	require.NoError(t, err)
	require.Equal(t, "企微正式应用 V2", updated.AppName)
	require.Equal(t, "beta", updated.Status)
	require.False(t, updated.TenantVisible)
}

func TestServiceUpdatesOperationalStates(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", Environment: "prod", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "pending", ConnectionStatus: "inactive", TokenStatus: "unknown",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	job := models.IntegrationSyncJob{
		TenantID: 9, TenantConnectionID: connection.ID, CapabilityCode: "order_sync",
		JobType: "incremental", TriggerMode: "manual", Status: "failed", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&job).Error)
	alert := models.IntegrationAlert{
		TenantID: &connection.TenantID, TenantConnectionID: &connection.ID, PlatformID: &platform.ID, ProviderAppID: &app.ID,
		AlertType: "sync", Severity: "warning", Status: "open", Title: "同步失败", FirstSeenAt: now, LastSeenAt: now,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&alert).Error)

	service := NewService(repositories.NewRepository(db))
	refreshed, err := service.RefreshTenantConnection(context.Background(), connection.ID)
	require.NoError(t, err)
	require.Equal(t, "connected", refreshed.ConnectionStatus)
	require.Equal(t, "valid", refreshed.TokenStatus)

	retried, err := service.RetrySyncJob(context.Background(), job.ID)
	require.NoError(t, err)
	require.Equal(t, "retrying", retried.Status)

	resolved, err := service.ResolveAlert(context.Background(), alert.ID)
	require.NoError(t, err)
	require.Equal(t, "resolved", resolved.Status)
	require.NotNil(t, resolved.ResolvedAt)
}

func TestServiceCreatesAndTogglesQuotaPolicy(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))

	created, err := service.CreateQuotaPolicy(context.Background(), QuotaPolicyMutationRequest{
		Code: "api-daily-premium", Name: "高级接口日调用", QuotaCode: "integration_api_calls_daily",
		QuotaUnit: "CALL", PeriodType: "DAY", DefaultLimit: 200000, OverLimitAction: "queue", Status: "enabled",
	})
	require.NoError(t, err)
	require.Equal(t, int64(200000), created.DefaultLimit)

	disabled, err := service.SetQuotaPolicyStatus(context.Background(), "api-daily-premium", false)
	require.NoError(t, err)
	require.Equal(t, "disabled", disabled.Status)
}

func newIntegrationCenterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.IntegrationPlatform{},
		&models.IntegrationPlatformCapability{},
		&models.IntegrationProviderApp{},
		&models.IntegrationProviderAppCapability{},
		&models.IntegrationTenantConnection{},
		&models.IntegrationTenantCapability{},
		&models.IntegrationSyncJob{},
		&models.IntegrationQuotaPolicy{},
		&models.IntegrationAlert{},
		&models.IntegrationAPICallLog{},
	))
	return db
}
