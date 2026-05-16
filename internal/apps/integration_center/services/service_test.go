package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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
	adminID := seedUser(t, db, 1, true)
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

	summary, err := NewService(repositories.NewRepository(db)).TenantConnections(context.Background(), adminID, PageRequest{Limit: 20})

	require.NoError(t, err)
	rows, ok := summary.Items.([]repositories.TenantConnectionSummary)
	require.True(t, ok)
	require.Len(t, rows, 1)
	require.Equal(t, int64(1), summary.Total)
	require.Equal(t, uint64(9), rows[0].TenantID)
	require.Equal(t, "shop", rows[0].AuthSubjectType)
	require.Equal(t, "shop-1001", rows[0].AuthSubjectID)
}

func TestServiceCreatesAndUpdatesPlatform(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	adminID := seedUser(t, db, 1, true)

	created, err := service.CreatePlatform(context.Background(), adminID, RequestMeta{}, PlatformMutationRequest{
		Name: "通用 ERP", Code: "generic-erp", PlatformType: "ERP", AccessMode: "API Key",
		Status: "enabled", TenantVisible: true, OwnerName: "集成组", SortOrder: 80,
	})
	require.NoError(t, err)
	require.Equal(t, "generic-erp", created.Code)
	require.Equal(t, "online", created.Status)

	updated, err := service.UpdatePlatform(context.Background(), adminID, RequestMeta{}, "generic-erp", PlatformMutationRequest{
		Name: "通用 ERP 平台", PlatformType: "ERP", AccessMode: "API Key",
		Status: "maintenance", TenantVisible: false, OwnerName: "平台集成组", SortOrder: 81,
	})
	require.NoError(t, err)
	require.Equal(t, "通用 ERP 平台", updated.Name)
	require.Equal(t, "maintenance", updated.Status)
	require.False(t, updated.TenantVisible)
}

func TestServiceRejectsUnknownPlatformStatus(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	adminID := seedUser(t, db, 1, true)

	_, err := service.CreatePlatform(context.Background(), adminID, RequestMeta{}, PlatformMutationRequest{
		Name: "坏状态平台", Code: "bad-status", PlatformType: "ERP", AccessMode: "API Key",
		Status: "half-online", TenantVisible: true,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "平台状态仅支持")
}

func TestServiceRejectsTenantUserPlatformWrite(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	tenantUserID := seedUser(t, db, 9, false)

	_, err := service.CreatePlatform(context.Background(), tenantUserID, RequestMeta{}, PlatformMutationRequest{
		Name: "租户不能创建平台", Code: "tenant-platform", PlatformType: "ERP", AccessMode: "API Key",
		Status: "online", TenantVisible: true,
	})

	require.ErrorIs(t, err, ErrForbidden)
}

func TestServiceCreatesAndUpdatesProviderApp(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	service := NewService(repositories.NewRepository(db))

	created, err := service.CreateProviderApp(context.Background(), adminID, RequestMeta{}, ProviderAppMutationRequest{
		PlatformCode: "wecom", Code: "wecom-suite-prod", Name: "企微正式应用", AuthMode: "OAuth2",
		Environment: "正式", Status: "enabled", TenantVisible: true,
	})
	require.NoError(t, err)
	require.Equal(t, "wecom-suite-prod", created.AppCode)
	require.Equal(t, "online", created.Status)

	updated, err := service.UpdateProviderApp(context.Background(), adminID, RequestMeta{}, "wecom-suite-prod", ProviderAppMutationRequest{
		PlatformCode: "wecom", Name: "企微正式应用 V2", AuthMode: "OAuth2",
		Environment: "test", Status: "testing", TenantVisible: false,
	})
	require.NoError(t, err)
	require.Equal(t, "企微正式应用 V2", updated.AppName)
	require.Equal(t, "beta", updated.Status)
	require.False(t, updated.TenantVisible)
}

func TestServiceRejectsUnknownProviderAppStatus(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)

	_, err := NewService(repositories.NewRepository(db)).CreateProviderApp(context.Background(), adminID, RequestMeta{}, ProviderAppMutationRequest{
		PlatformCode: "wecom", Code: "wecom-bad", Name: "坏状态应用", AuthMode: "OAuth2",
		Environment: "正式", Status: "grayish", TenantVisible: true,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "应用状态仅支持")
}

func TestServiceRotatesProviderAppCredentialWithMaskedOutput(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	oldRef := "env://INTEGRATION_OLD_SECRET"
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "wecom-suite", AppName: "企微应用", AuthMode: "OAuth2", Status: "online", CredentialRef: &oldRef, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)

	result, err := NewService(repositories.NewRepository(db)).RotateProviderAppCredential(context.Background(), adminID, RequestMeta{RequestID: "req-rotate"}, "wecom-suite", ProviderAppCredentialRotateRequest{
		CredentialRef: "env://INTEGRATION_NEW_SECRET",
	})

	require.NoError(t, err)
	require.NotNil(t, result.CredentialRef)
	require.Equal(t, "env://...CRET", *result.CredentialRef)
	var stored models.IntegrationProviderApp
	require.NoError(t, db.First(&stored, app.ID).Error)
	require.NotNil(t, stored.CredentialRef)
	require.Equal(t, "env://INTEGRATION_NEW_SECRET", *stored.CredentialRef)
	var audit models.AuditLog
	require.NoError(t, db.Where("action = ?", "rotate_provider_app_credential").First(&audit).Error)
	require.NotNil(t, audit.Detail)
	require.NotContains(t, *audit.Detail, "INTEGRATION_NEW_SECRET")
	require.Contains(t, *audit.Detail, "env://...CRET")
}

func TestServiceRejectsUnsafeProviderAppCredentialRef(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "wecom-suite", AppName: "企微应用", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)

	_, err := NewService(repositories.NewRepository(db)).RotateProviderAppCredential(context.Background(), adminID, RequestMeta{}, "wecom-suite", ProviderAppCredentialRotateRequest{
		CredentialRef: "plain-secret-value",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credential_ref")
}

func TestServiceManagesPlatformCapabilityLifecycle(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{
		PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2",
		Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&platform).Error)
	service := NewService(repositories.NewRepository(db))

	created, err := service.CreatePlatformCapability(context.Background(), adminID, RequestMeta{RequestID: "req-cap-create"}, PlatformCapabilityMutationRequest{
		PlatformCode: "wecom", Code: "user_sync", Name: "通讯录同步", CapabilityType: "directory",
		AuthScopeCode: "contacts.read", DataDirection: "pull", Status: "enabled", Description: "同步组织通讯录",
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "user_sync", created.CapabilityCode)
	require.Equal(t, "通讯录同步", created.CapabilityName)
	require.Equal(t, "pull", created.DataDirection)
	require.Equal(t, "enabled", created.Status)

	updated, err := service.UpdatePlatformCapability(context.Background(), adminID, RequestMeta{RequestID: "req-cap-update"}, created.ID, PlatformCapabilityMutationRequest{
		Name: "通讯录双向同步", CapabilityType: "directory", AuthScopeCode: "contacts.write",
		DataDirection: "both", Status: "enabled", Description: "同步并回写组织通讯录",
	})
	require.NoError(t, err)
	require.Equal(t, "通讯录双向同步", updated.CapabilityName)
	require.Equal(t, "both", updated.DataDirection)
	require.NotNil(t, updated.AuthScopeCode)
	require.Equal(t, "contacts.write", *updated.AuthScopeCode)

	disabled, err := service.DisablePlatformCapability(context.Background(), adminID, RequestMeta{RequestID: "req-cap-disable"}, created.ID)
	require.NoError(t, err)
	require.Equal(t, "disabled", disabled.Status)
	var audits int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("action IN ?", []string{"create_platform_capability", "update_platform_capability", "disable_platform_capability"}).Count(&audits).Error)
	require.Equal(t, int64(3), audits)
}

func TestServiceUpdatesAppCapabilityConfig(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "wecom-suite", AppName: "企微应用", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	platformCapability := models.IntegrationPlatformCapability{
		PlatformID: platform.ID, CapabilityCode: "user_sync", CapabilityName: "通讯录同步",
		CapabilityType: "directory", DataDirection: "pull", Status: "enabled", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&platformCapability).Error)
	appCapability := models.IntegrationProviderAppCapability{
		ProviderAppID: app.ID, PlatformCapabilityID: platformCapability.ID,
		ConnectionStatus: "pending", ReviewStatus: "pending", Enabled: false, Config: "{}",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&appCapability).Error)
	service := NewService(repositories.NewRepository(db))

	updated, err := service.UpdateAppCapability(context.Background(), adminID, RequestMeta{RequestID: "req-app-cap"}, appCapability.ID, AppCapabilityPatchRequest{
		Enabled:            boolPtr(true),
		ConnectionStatus:   "connected",
		ReviewStatus:       "approved",
		OpenToTenant:       boolPtr(true),
		DefaultEnabled:     boolPtr(true),
		TenantConfigurable: boolPtr(false),
		Config:             map[string]interface{}{"sync_mode": "incremental", "batch_size": float64(200)},
	})

	require.NoError(t, err)
	require.True(t, updated.Enabled)
	require.Equal(t, "connected", updated.ConnectionStatus)
	require.Equal(t, "approved", updated.ReviewStatus)
	require.JSONEq(t, `{"open_to_tenant":true,"default_enabled":true,"tenant_configurable":false,"sync_mode":"incremental","batch_size":200}`, updated.Config)

	summary, err := service.AppCapabilities(context.Background(), adminID, PageRequest{Limit: 20})
	require.NoError(t, err)
	rows, ok := summary.Items.([]repositories.ProviderAppCapabilitySummary)
	require.True(t, ok)
	require.Len(t, rows, 1)
	require.True(t, rows[0].OpenToTenant)
	require.True(t, rows[0].DefaultEnabled)
	require.False(t, rows[0].TenantConfigurable)
	require.JSONEq(t, updated.Config, rows[0].Config)
	var audits int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("action = ?", "update_app_capability").Count(&audits).Error)
	require.Equal(t, int64(1), audits)
}

func TestServiceRejectsUnknownAppCapabilityStatus(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "wecom", PlatformName: "企业微信", PlatformType: "协同办公", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "wecom-suite", AppName: "企微应用", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	platformCapability := models.IntegrationPlatformCapability{
		PlatformID: platform.ID, CapabilityCode: "user_sync", CapabilityName: "通讯录同步",
		CapabilityType: "directory", DataDirection: "pull", Status: "enabled", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&platformCapability).Error)
	appCapability := models.IntegrationProviderAppCapability{
		ProviderAppID: app.ID, PlatformCapabilityID: platformCapability.ID,
		ConnectionStatus: "pending", ReviewStatus: "pending", Enabled: false, Config: "{}",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&appCapability).Error)
	service := NewService(repositories.NewRepository(db))

	_, err := service.UpdateAppCapability(context.Background(), adminID, RequestMeta{}, appCapability.ID, AppCapabilityPatchRequest{
		ConnectionStatus: "almost_connected",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "连接状态仅支持")

	_, err = service.UpdateAppCapability(context.Background(), adminID, RequestMeta{}, appCapability.ID, AppCapabilityPatchRequest{
		ReviewStatus: "maybe",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "审核状态仅支持")
}

func TestServiceUpdatesOperationalStates(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", Environment: "prod", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	tokenRef := "vault://integration/jd-shop/tenant-9"
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "pending", ConnectionStatus: "inactive", TokenStatus: "unknown", TokenCredentialRef: &tokenRef,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	refreshServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())
		require.Equal(t, "refresh_token", r.FormValue("grant_type"))
		require.Equal(t, tokenRef, r.FormValue("credential_ref"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"credential_ref":"vault://integration/jd-shop/tenant-9-refreshed","expires_in":7200}`))
	}))
	defer refreshServer.Close()
	t.Setenv("INTEGRATION_OAUTH_REFRESH_URL_JD_SHOP", refreshServer.URL)
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
	refreshed, err := service.RefreshTenantConnection(context.Background(), adminID, RequestMeta{}, connection.ID)
	require.NoError(t, err)
	require.Equal(t, "connected", refreshed.ConnectionStatus)
	require.Equal(t, "valid", refreshed.TokenStatus)
	require.NotNil(t, refreshed.TokenCredentialRef)
	require.Equal(t, "vault://integration/jd-shop/tenant-9-refreshed", *refreshed.TokenCredentialRef)
	require.NotNil(t, refreshed.TokenExpiresAt)

	retried, err := service.RetrySyncJob(context.Background(), adminID, RequestMeta{}, job.ID)
	require.NoError(t, err)
	require.Equal(t, "retrying", retried.Status)

	resolved, err := service.ResolveAlert(context.Background(), adminID, RequestMeta{}, alert.ID)
	require.NoError(t, err)
	require.Equal(t, "resolved", resolved.Status)
	require.NotNil(t, resolved.ResolvedAt)
}

func TestServiceCreatesAndTogglesQuotaPolicy(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	service := NewService(repositories.NewRepository(db))
	adminID := seedUser(t, db, 1, true)

	created, err := service.CreateQuotaPolicy(context.Background(), adminID, RequestMeta{}, QuotaPolicyMutationRequest{
		Code: "api-daily-premium", Name: "高级接口日调用", QuotaCode: "integration_api_calls_daily",
		QuotaUnit: "CALL", PeriodType: "DAY", DefaultLimit: 200000, OverLimitAction: "queue", Status: "enabled",
	})
	require.NoError(t, err)
	require.Equal(t, int64(200000), created.DefaultLimit)

	disabled, err := service.SetQuotaPolicyStatus(context.Background(), adminID, RequestMeta{}, "api-daily-premium", false)
	require.NoError(t, err)
	require.Equal(t, "disabled", disabled.Status)
}

func TestServiceCreatesTenantConnectionWithFeatureAndQuotaGate(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 2)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)

	connection, err := NewService(repositories.NewRepository(db)).CreateTenantConnection(context.Background(), tenantUserID, RequestMeta{}, TenantConnectionCreateRequest{
		ProviderAppCode: "jd-shop",
		ConnectionName:  "旗舰店授权",
		AuthSubjectType: "shop",
		AuthSubjectID:   "shop-1001",
		AuthSubjectName: "旗舰店",
		AuthScope:       []string{"order.read"},
	})

	require.NoError(t, err)
	require.Equal(t, uint64(9), connection.TenantID)
	require.Equal(t, "pending", connection.AuthStatus)
	require.Equal(t, "inactive", connection.ConnectionStatus)
	require.NotNil(t, connection.CreatedBy)
}

func TestServiceRejectsTenantConnectionWhenFeatureDisabled(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, false, 2)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)

	_, err := NewService(repositories.NewRepository(db)).CreateTenantConnection(context.Background(), tenantUserID, RequestMeta{}, TenantConnectionCreateRequest{
		ProviderAppCode: "jd-shop",
		ConnectionName:  "旗舰店授权",
		AuthSubjectID:   "shop-1001",
		AuthSubjectName: "旗舰店",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "integration_tenant_authorization")
}

func TestServiceRejectsTenantConnectionWhenQuotaExceeded(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 1)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	require.NoError(t, db.Create(&models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "已有连接",
		AuthSubjectType: "shop", AuthSubjectID: "shop-old", AuthSubjectName: "已有店铺",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}).Error)

	_, err := NewService(repositories.NewRepository(db)).CreateTenantConnection(context.Background(), tenantUserID, RequestMeta{}, TenantConnectionCreateRequest{
		ProviderAppCode: "jd-shop",
		ConnectionName:  "旗舰店授权",
		AuthSubjectID:   "shop-1001",
		AuthSubjectName: "旗舰店",
	})

	require.ErrorIs(t, err, ErrQuotaExceeded)
}

func TestServiceStartsOAuthAuthorizationWithStateGate(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 2)
	t.Setenv("INTEGRATION_OAUTH_AUTHORIZE_URL_JD_SHOP", "https://auth.example.com/oauth/authorize?client_id=client-1")
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)

	result, err := NewService(repositories.NewRepository(db)).StartOAuthAuthorization(context.Background(), tenantUserID, RequestMeta{}, OAuthStartRequest{
		ProviderAppCode: "jd-shop",
		RedirectURI:     "https://app.example.com/api/integration-center/oauth/callback/jd-shop",
		Scopes:          []string{"order.read", "shop.read"},
	})

	require.NoError(t, err)
	require.Equal(t, "jd-shop", result.ProviderAppCode)
	require.NotEmpty(t, result.State)
	require.Contains(t, result.AuthURL, "state="+result.State)
	require.Contains(t, result.AuthURL, "redirect_uri=")
	var stored models.IntegrationOAuthState
	require.NoError(t, db.Where("state = ?", result.State).First(&stored).Error)
	require.Equal(t, uint64(9), stored.TenantID)
	require.Equal(t, app.ID, stored.ProviderAppID)
	require.Equal(t, "pending", stored.Status)
	require.True(t, stored.ExpiresAt.After(time.Now()))
	require.True(t, strings.Contains(stored.Scopes, "order.read"))
}

func TestServiceRejectsOAuthCallbackReplayAndExpiredState(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 2)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	service := NewService(repositories.NewRepository(db))
	started, err := service.StartOAuthAuthorization(context.Background(), tenantUserID, RequestMeta{}, OAuthStartRequest{
		ProviderAppCode: "jd-shop",
		RedirectURI:     "https://app.example.com/oauth/callback",
	})
	require.NoError(t, err)

	callback, err := service.HandleOAuthCallback(context.Background(), OAuthCallbackRequest{ProviderAppCode: "jd-shop", State: started.State, Code: "auth-code"})
	require.NoError(t, err)
	require.Equal(t, "code_received", callback.Status)
	_, err = service.HandleOAuthCallback(context.Background(), OAuthCallbackRequest{ProviderAppCode: "jd-shop", State: started.State, Code: "auth-code"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "已被使用")

	expired := models.IntegrationOAuthState{
		State:         "expired-state",
		TenantID:      9,
		ProviderAppID: app.ID,
		PlatformID:    platform.ID,
		RedirectURI:   "https://app.example.com/oauth/callback",
		Scopes:        "[]",
		Status:        "pending",
		ExpiresAt:     time.Now().Add(-time.Minute),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, db.Create(&expired).Error)
	_, err = service.HandleOAuthCallback(context.Background(), OAuthCallbackRequest{ProviderAppCode: "jd-shop", State: "expired-state", Code: "auth-code"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "已过期")
}

func TestServiceOAuthCallbackExchangesCodeIntoCredentialRefConnection(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 2)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())
		require.Equal(t, "authorization_code", r.FormValue("grant_type"))
		require.Equal(t, "auth-code", r.FormValue("code"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"credential_ref":"vault://integration/jd-shop/tenant-9","auth_subject_type":"shop","auth_subject_id":"shop-1001","auth_subject_name":"旗舰店","scope":["order.read"],"expires_in":3600}`))
	}))
	defer tokenServer.Close()
	t.Setenv("INTEGRATION_OAUTH_TOKEN_URL_JD_SHOP", tokenServer.URL)
	service := NewService(repositories.NewRepository(db))
	started, err := service.StartOAuthAuthorization(context.Background(), tenantUserID, RequestMeta{}, OAuthStartRequest{
		ProviderAppCode: "jd-shop",
		RedirectURI:     "https://app.example.com/oauth/callback",
	})
	require.NoError(t, err)

	callback, err := service.HandleOAuthCallback(context.Background(), OAuthCallbackRequest{ProviderAppCode: "jd-shop", State: started.State, Code: "auth-code"})

	require.NoError(t, err)
	require.Equal(t, "authorized", callback.Status)
	require.NotZero(t, callback.ConnectionID)
	var connection models.IntegrationTenantConnection
	require.NoError(t, db.First(&connection, callback.ConnectionID).Error)
	require.Equal(t, "authorized", connection.AuthStatus)
	require.Equal(t, "connected", connection.ConnectionStatus)
	require.Equal(t, "valid", connection.TokenStatus)
	require.NotNil(t, connection.TokenCredentialRef)
	require.Equal(t, "vault://integration/jd-shop/tenant-9", *connection.TokenCredentialRef)
	require.NotNil(t, connection.TokenExpiresAt)
	require.Equal(t, "shop-1001", connection.AuthSubjectID)
}

func TestServiceOAuthCallbackRejectsPlainTokenWithoutCredentialRef(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 2)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"plain-token","expires_in":3600}`))
	}))
	defer tokenServer.Close()
	t.Setenv("INTEGRATION_OAUTH_TOKEN_URL_JD_SHOP", tokenServer.URL)
	service := NewService(repositories.NewRepository(db))
	started, err := service.StartOAuthAuthorization(context.Background(), tenantUserID, RequestMeta{}, OAuthStartRequest{
		ProviderAppCode: "jd-shop",
		RedirectURI:     "https://app.example.com/oauth/callback",
	})
	require.NoError(t, err)

	_, err = service.HandleOAuthCallback(context.Background(), OAuthCallbackRequest{ProviderAppCode: "jd-shop", State: started.State, Code: "auth-code"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credential_ref")
	var count int64
	require.NoError(t, db.Model(&models.IntegrationTenantConnection{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestServiceRefreshTokenFailureUpdatesConnectionAndAlert(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	tokenRef := "vault://integration/jd-shop/tenant-9"
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid", TokenCredentialRef: &tokenRef,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	refreshServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "refresh failed", http.StatusBadGateway)
	}))
	defer refreshServer.Close()
	t.Setenv("INTEGRATION_OAUTH_REFRESH_URL_JD_SHOP", refreshServer.URL)

	failed, err := NewService(repositories.NewRepository(db)).RefreshTenantConnection(context.Background(), adminID, RequestMeta{}, connection.ID)

	require.Error(t, err)
	require.Equal(t, "failed", failed.ConnectionStatus)
	require.Equal(t, "refresh_failed", failed.TokenStatus)
	require.NotNil(t, failed.LastErrorAt)
	var alertCount int64
	require.NoError(t, db.Model(&models.IntegrationAlert{}).Where("alert_type = ? AND tenant_connection_id = ?", "token_refresh", connection.ID).Count(&alertCount).Error)
	require.Equal(t, int64(1), alertCount)
}

func TestServiceConsumesAPICallQuotaAndRejectsOverLimit(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 1)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	service := NewService(repositories.NewRepository(db))

	used, err := service.ConsumeAPICallQuota(context.Background(), tenantUserID, RequestMeta{}, APICallQuotaConsumeRequest{TenantConnectionID: &connection.ID, Amount: 1})
	require.NoError(t, err)
	require.Equal(t, int64(1), used.UsedAmount)
	require.Equal(t, int64(0), used.LimitedCount)

	_, err = service.ConsumeAPICallQuota(context.Background(), tenantUserID, RequestMeta{}, APICallQuotaConsumeRequest{TenantConnectionID: &connection.ID, Amount: 1})
	require.ErrorIs(t, err, ErrQuotaExceeded)
	var usage models.IntegrationQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND tenant_connection_id = ? AND quota_code = ?", 9, connection.ID, "integration_api_calls_daily").First(&usage).Error)
	require.Equal(t, int64(1), usage.UsedAmount)
	require.Equal(t, int64(1), usage.LimitedCount)
}

func TestServiceConsumesAPICallQuotaWithConnectionBoundPolicy(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 10)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	service := NewService(repositories.NewRepository(db))
	overrideLimit := int64(1)
	_, err := service.CreateQuotaPolicy(context.Background(), adminID, RequestMeta{}, QuotaPolicyMutationRequest{
		Code: "api-daily-jd-shop-low", Name: "京东店铺低限额", QuotaCode: "integration_api_calls_daily",
		QuotaUnit: "CALL", PeriodType: "DAY", DefaultLimit: 5, OverLimitAction: "reject", Status: "enabled",
		ScopeType: "tenant_connection", TenantConnectionID: &connection.ID, OverrideLimit: &overrideLimit, Priority: 100,
	})
	require.NoError(t, err)

	used, err := service.ConsumeAPICallQuota(context.Background(), tenantUserID, RequestMeta{}, APICallQuotaConsumeRequest{TenantConnectionID: &connection.ID, Amount: 1})
	require.NoError(t, err)
	require.Equal(t, int64(1), used.UsedAmount)

	_, err = service.ConsumeAPICallQuota(context.Background(), tenantUserID, RequestMeta{}, APICallQuotaConsumeRequest{TenantConnectionID: &connection.ID, Amount: 1})
	require.ErrorIs(t, err, ErrQuotaExceeded)
	var binding models.IntegrationQuotaBinding
	require.NoError(t, db.Where("tenant_connection_id = ?", connection.ID).First(&binding).Error)
	require.Equal(t, 100, binding.Priority)
	require.NotNil(t, binding.OverrideLimit)
	require.Equal(t, int64(1), *binding.OverrideLimit)
}

func TestServiceConsumesSyncRecordQuotaAndRejectsOverLimit(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 2)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	service := NewService(repositories.NewRepository(db))

	used, err := service.ConsumeSyncRecordQuota(context.Background(), tenantUserID, RequestMeta{}, SyncRecordQuotaConsumeRequest{TenantConnectionID: connection.ID, Amount: 2})
	require.NoError(t, err)
	require.Equal(t, int64(2), used.UsedAmount)

	_, err = service.ConsumeSyncRecordQuota(context.Background(), tenantUserID, RequestMeta{}, SyncRecordQuotaConsumeRequest{TenantConnectionID: connection.ID, Amount: 1})
	require.ErrorIs(t, err, ErrQuotaExceeded)
	var usage models.IntegrationQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND tenant_connection_id = ? AND quota_code = ?", 9, connection.ID, "integration_sync_records_daily").First(&usage).Error)
	require.Equal(t, int64(2), usage.UsedAmount)
	require.Equal(t, int64(1), usage.LimitedCount)
}

func TestServiceInvokesGatewayWithQuotaAndDigestLog(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 5)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	tokenRef := "vault://integration/jd-shop/tenant-9"
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid", TokenCredentialRef: &tokenRef,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/orders/list", r.URL.Path)
		require.Equal(t, "1", r.URL.Query().Get("page"))
		require.Empty(t, r.Header.Get("Authorization"))
		require.Empty(t, r.Header.Get("X-Forwarded-For"))
		require.Equal(t, tokenRef, r.Header.Get("X-Integration-Credential-Ref"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"orders":[{"id":"1001"}]}`))
	}))
	defer apiServer.Close()
	t.Setenv("INTEGRATION_PROVIDER_BASE_URL_JD_SHOP", apiServer.URL)

	result, err := NewService(repositories.NewRepository(db)).InvokeGateway(context.Background(), tenantUserID, RequestMeta{RequestID: "req-gw-1", TraceID: "trace-api-1"}, GatewayInvokeRequest{
		TenantConnectionID: connection.ID,
		Method:             "POST",
		Path:               "/orders/list?page=1",
		Headers:            map[string]string{"Authorization": "Bearer unsafe", "X-Trace-ID": "trace-1", "X-Forwarded-For": "127.0.0.1"},
		Body:               `{"page":1}`,
	})

	require.NoError(t, err)
	require.Equal(t, "success", result.Status)
	require.Equal(t, 200, result.HTTPStatus)
	require.NotEmpty(t, result.RequestDigest)
	require.NotEmpty(t, result.ResponseDigest)
	var log models.IntegrationAPICallLog
	require.NoError(t, db.Where("request_id = ?", "req-gw-1").First(&log).Error)
	require.Equal(t, "third_party_api", log.CallType)
	require.Equal(t, "success", log.Status)
	require.NotNil(t, log.TraceID)
	require.Equal(t, "trace-api-1", *log.TraceID)
	require.NotNil(t, log.TenantID)
	require.Equal(t, uint64(9), *log.TenantID)
	require.NotNil(t, log.TenantConnectionID)
	require.Equal(t, connection.ID, *log.TenantConnectionID)
	require.NotNil(t, log.PlatformID)
	require.Equal(t, platform.ID, *log.PlatformID)
	require.NotNil(t, log.ProviderAppID)
	require.Equal(t, app.ID, *log.ProviderAppID)
	require.NotNil(t, log.RequestDigest)
	require.NotNil(t, log.ResponseDigest)
}

func TestServiceGatewayRejectsQuotaExceededAndLogsLimited(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	seedIntegrationPlan(t, db, 9, true, 0)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", TenantVisible: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	tokenRef := "vault://integration/jd-shop/tenant-9"
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid", TokenCredentialRef: &tokenRef,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	t.Setenv("INTEGRATION_PROVIDER_BASE_URL_JD_SHOP", "https://api.example.com")

	_, err := NewService(repositories.NewRepository(db)).InvokeGateway(context.Background(), tenantUserID, RequestMeta{RequestID: "req-gw-limited"}, GatewayInvokeRequest{
		TenantConnectionID: connection.ID,
		Method:             "GET",
		Path:               "/orders/list",
	})

	require.ErrorIs(t, err, ErrQuotaExceeded)
	var log models.IntegrationAPICallLog
	require.NoError(t, db.Where("request_id = ?", "req-gw-limited").First(&log).Error)
	require.Equal(t, "limited", log.Status)
	require.NotNil(t, log.ErrorCode)
	require.Equal(t, "quota_exceeded", *log.ErrorCode)
}

func TestServiceSanitizesAPICallLogSecrets(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	tenantID := uint64(1)
	endpoint := "/orders?access_token=raw-token&secret=raw-secret"
	message := "failed credential_ref=vault://integration/full-secret access_token=abc123 client_secret=topsecret"
	require.NoError(t, db.Create(&models.IntegrationAPICallLog{
		TenantID: &tenantID, RequestID: "req-secret-log", CallType: "third_party_api",
		Endpoint: &endpoint, Status: "failed", ErrorMessage: &message, CalledAt: now, CreatedAt: now,
	}).Error)

	result, err := NewService(repositories.NewRepository(db)).LogDetail(context.Background(), adminID, 1)

	require.NoError(t, err)
	log, ok := result["log"].(models.IntegrationAPICallLog)
	require.True(t, ok)
	require.NotNil(t, log.Endpoint)
	require.Equal(t, "/orders", *log.Endpoint)
	require.NotNil(t, log.ErrorMessage)
	require.NotContains(t, *log.ErrorMessage, "vault://integration/full-secret")
	require.NotContains(t, *log.ErrorMessage, "abc123")
	require.NotContains(t, *log.ErrorMessage, "topsecret")
	require.Contains(t, *log.ErrorMessage, "credential_ref=***")
}

func TestBuildAPICallLogRedactsAndNormalizesEntry(t *testing.T) {
	now := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
	log := BuildAPICallLog(APICallLogEntry{
		Method:         "GET",
		Endpoint:       "https://api.example.com/orders?access_token=secret#debug",
		Status:         "bad-status",
		HTTPStatus:     500,
		DurationMS:     42,
		RequestDigest:  "sha256:req",
		ResponseDigest: "sha256:resp",
		ErrorMessage:   "token=secret client_secret=hidden",
	}, now)

	require.NotEmpty(t, log.RequestID)
	require.Equal(t, log.RequestID, *log.TraceID)
	require.Equal(t, "third_party_api", log.CallType)
	require.Equal(t, "failed", log.Status)
	require.Equal(t, now, log.CalledAt)
	require.Equal(t, "GET", *log.Method)
	require.Equal(t, "https://api.example.com/orders", *log.Endpoint)
	require.Equal(t, 500, *log.HTTPStatus)
	require.Equal(t, "sha256:req", *log.RequestDigest)
	require.Equal(t, "sha256:resp", *log.ResponseDigest)
	require.NotContains(t, *log.ErrorMessage, "token=secret")
	require.NotContains(t, *log.ErrorMessage, "hidden")
}

func TestServiceGatewayRejectsUnsafePath(t *testing.T) {
	app := models.IntegrationProviderApp{AppCode: "jd-shop"}
	t.Setenv("INTEGRATION_PROVIDER_BASE_URL_JD_SHOP", "https://api.example.com/base")

	_, _, err := gatewayRequestTarget(app, GatewayInvokeRequest{Method: "GET", Path: "//evil.example.com/orders"})
	require.Error(t, err)

	_, _, err = gatewayRequestTarget(app, GatewayInvokeRequest{Method: "GET", Path: "/../admin"})
	require.Error(t, err)

	method, endpoint, err := gatewayRequestTarget(app, GatewayInvokeRequest{Method: "GET", Path: "/orders/list?page=1"})
	require.NoError(t, err)
	require.Equal(t, "GET", method)
	require.Equal(t, "https://api.example.com/base/orders/list?page=1", endpoint)
}

func TestServiceScopesTenantConnectionsForNonPlatformViewer(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	for _, tenantID := range []uint64{9, 10} {
		require.NoError(t, db.Create(&models.IntegrationTenantConnection{
			TenantID: tenantID, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "租户连接",
			AuthSubjectType: "shop", AuthSubjectID: "shop", AuthSubjectName: "店铺",
			AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
			CreatedAt: now, UpdatedAt: now,
		}).Error)
	}

	summary, err := NewService(repositories.NewRepository(db)).TenantConnections(context.Background(), tenantUserID, PageRequest{Limit: 20})

	require.NoError(t, err)
	rows := summary.Items.([]repositories.TenantConnectionSummary)
	require.Len(t, rows, 1)
	require.Equal(t, uint64(9), rows[0].TenantID)
	require.Equal(t, int64(1), summary.Total)

	adminSummary, err := NewService(repositories.NewRepository(db)).TenantConnections(context.Background(), adminID, PageRequest{Limit: 20})
	require.NoError(t, err)
	adminRows := adminSummary.Items.([]repositories.TenantConnectionSummary)
	require.Len(t, adminRows, 2)
	require.Equal(t, int64(2), adminSummary.Total)
}

func TestServiceFiltersAndSortsTenantConnections(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	oldTime := now.Add(-72 * time.Hour)
	recentTime := now.Add(-2 * time.Hour)
	oldConnection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旧连接",
		AuthSubjectType: "shop", AuthSubjectID: "old-shop", AuthSubjectName: "旧店铺",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "paused", TokenStatus: "valid",
		CreatedAt: oldTime, UpdatedAt: oldTime,
	}
	recentConnection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "新连接",
		AuthSubjectType: "shop", AuthSubjectID: "new-shop", AuthSubjectName: "新店铺",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: recentTime, UpdatedAt: recentTime,
	}
	require.NoError(t, db.Create(&oldConnection).Error)
	require.NoError(t, db.Create(&recentConnection).Error)
	start := now.Add(-24 * time.Hour)

	summary, err := NewService(repositories.NewRepository(db)).TenantConnections(context.Background(), adminID, PageRequest{
		Limit:     20,
		StartTime: &start,
		SortBy:    "updated_at",
		SortOrder: "asc",
	})

	require.NoError(t, err)
	rows := summary.Items.([]repositories.TenantConnectionSummary)
	require.Len(t, rows, 1)
	require.Equal(t, "新连接", rows[0].ConnectionName)
	require.Equal(t, int64(1), summary.Total)
}

func TestServiceRejectsCrossTenantConnectionOperation(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	tenantUserID := seedUser(t, db, 9, false)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 10, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "其他租户连接",
		AuthSubjectType: "shop", AuthSubjectID: "shop", AuthSubjectName: "店铺",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)

	_, err := NewService(repositories.NewRepository(db)).PauseSyncJob(context.Background(), tenantUserID, RequestMeta{}, 999)
	require.Error(t, err)

	_, err = NewService(repositories.NewRepository(db)).SetTenantConnectionStatus(context.Background(), tenantUserID, RequestMeta{}, connection.ID, true)
	require.ErrorIs(t, err, ErrForbidden)
}

func TestServiceWritesAuditForAlertHandling(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	alert := models.IntegrationAlert{
		AlertType: "sync", Severity: "warning", Status: "open", Title: "同步失败",
		FirstSeenAt: now, LastSeenAt: now, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&alert).Error)

	_, err := NewService(repositories.NewRepository(db)).ResolveAlert(context.Background(), adminID, RequestMeta{RequestID: "req-1"}, alert.ID)

	require.NoError(t, err)
	var count int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("action = ?", "resolve_alert").Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestServiceRejectsInvalidStateTransitions(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	job := models.IntegrationSyncJob{
		TenantID: 9, TenantConnectionID: connection.ID, CapabilityCode: "order_sync",
		JobType: "incremental", TriggerMode: "manual", Status: "completed", CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&job).Error)
	alert := models.IntegrationAlert{
		TenantID: &connection.TenantID, TenantConnectionID: &connection.ID, PlatformID: &platform.ID, ProviderAppID: &app.ID,
		AlertType: "sync", Severity: "warning", Status: "resolved", Title: "已恢复异常", FirstSeenAt: now, LastSeenAt: now, ResolvedAt: &now,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&alert).Error)

	service := NewService(repositories.NewRepository(db))
	paused, err := service.SetTenantConnectionStatus(context.Background(), adminID, RequestMeta{}, connection.ID, true)
	require.NoError(t, err)
	require.Equal(t, "paused", paused.ConnectionStatus)

	_, err = service.SetTenantConnectionStatus(context.Background(), adminID, RequestMeta{}, connection.ID, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "目标状态")

	_, err = service.PauseSyncJob(context.Background(), adminID, RequestMeta{}, job.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "不能暂停")

	_, err = service.ProcessAlert(context.Background(), adminID, RequestMeta{}, alert.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "不能再次处理")
}

func TestServiceReceivesSignedWebhookIdempotently(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	secretKey := "INTEGRATION_WEBHOOK_SECRET_JD_SHOP"
	t.Setenv(secretKey, "secret-value")
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	credentialRef := secretKey
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "Webhook", Status: "online", CredentialRef: &credentialRef, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	body := []byte(`{"order_id":"1001"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := signWebhook("secret-value", timestamp, body)

	service := NewService(repositories.NewRepository(db))
	result, err := service.ReceiveWebhook(context.Background(), WebhookReceiveRequest{
		ProviderAppCode: "jd-shop",
		Timestamp:       timestamp,
		Signature:       "sha256=" + signature,
		IdempotencyKey:  "event-1",
		EventType:       "order.created",
		Body:            body,
	})
	require.NoError(t, err)
	require.False(t, result.Duplicate)
	require.NotZero(t, result.EventID)

	duplicate, err := service.ReceiveWebhook(context.Background(), WebhookReceiveRequest{
		ProviderAppCode: "jd-shop",
		Timestamp:       timestamp,
		Signature:       signature,
		IdempotencyKey:  "event-1",
		EventType:       "order.created",
		Body:            body,
	})
	require.NoError(t, err)
	require.True(t, duplicate.Duplicate)

	var count int64
	require.NoError(t, db.Model(&models.IntegrationWebhookEvent{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
	var logs []models.IntegrationAPICallLog
	require.NoError(t, db.Where("call_type = ?", "webhook").Order("id ASC").Find(&logs).Error)
	require.Len(t, logs, 2)
	require.Equal(t, "success", logs[0].Status)
	require.NotNil(t, logs[0].RequestDigest)
	require.NotContains(t, stringValue(logs[0].Endpoint), "order_id")
	require.Nil(t, logs[0].ResponseDigest)
}

func TestServiceRejectsWebhookReplayAndBadSignature(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	t.Setenv("INTEGRATION_WEBHOOK_SECRET_JD_SHOP", "secret-value")
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	body := []byte(`{"order_id":"1001"}`)
	expired := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	_, err := NewService(repositories.NewRepository(db)).ReceiveWebhook(context.Background(), WebhookReceiveRequest{
		ProviderAppCode: "jd-shop",
		Timestamp:       expired,
		Signature:       signWebhook("secret-value", expired, body),
		IdempotencyKey:  "event-1",
		Body:            body,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "时间戳")

	current := strconv.FormatInt(time.Now().Unix(), 10)
	_, err = NewService(repositories.NewRepository(db)).ReceiveWebhook(context.Background(), WebhookReceiveRequest{
		ProviderAppCode: "jd-shop",
		Timestamp:       current,
		Signature:       "bad",
		IdempotencyKey:  "event-2",
		Body:            body,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "签名")
}

func TestServiceProcessesDueWebhookEvents(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	event := models.IntegrationWebhookEvent{
		ProviderAppID:  app.ID,
		PlatformID:     platform.ID,
		EventType:      "order.created",
		IdempotencyKey: "event-process",
		Signature:      "sig",
		PayloadDigest:  "digest",
		Payload:        `{"order_id":"1001"}`,
		Status:         "received",
		ReceivedAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, db.Create(&event).Error)

	result, err := NewService(repositories.NewRepository(db)).ProcessDueWebhookEvents(context.Background(), 10)

	require.NoError(t, err)
	require.Equal(t, 1, result.Scanned)
	require.Equal(t, 1, result.Processed)
	var stored models.IntegrationWebhookEvent
	require.NoError(t, db.First(&stored, event.ID).Error)
	require.Equal(t, "processed", stored.Status)
	require.NotNil(t, stored.ProcessedAt)
}

func TestServiceRetriesAndDeadLettersWebhookEvents(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "Webhook", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	event := models.IntegrationWebhookEvent{
		ProviderAppID:  app.ID,
		PlatformID:     platform.ID,
		EventType:      "order.created",
		IdempotencyKey: "event-retry",
		Signature:      "sig",
		PayloadDigest:  "digest",
		Payload:        `{bad-json`,
		Status:         "received",
		ReceivedAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, db.Create(&event).Error)

	result, err := NewService(repositories.NewRepository(db)).ProcessDueWebhookEvents(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retrying)
	var stored models.IntegrationWebhookEvent
	require.NoError(t, db.First(&stored, event.ID).Error)
	require.Equal(t, "retrying", stored.Status)
	require.Equal(t, 1, stored.RetryCount)
	require.NotNil(t, stored.NextRetryAt)

	past := time.Now().Add(-time.Minute)
	require.NoError(t, db.Model(&models.IntegrationWebhookEvent{}).Where("id = ?", event.ID).Updates(map[string]interface{}{"retry_count": webhookMaxRetryCount - 1, "next_retry_at": past}).Error)
	result, err = NewService(repositories.NewRepository(db)).ProcessDueWebhookEvents(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.DeadLetter)
	require.NoError(t, db.First(&stored, event.ID).Error)
	require.Equal(t, "dead_letter", stored.Status)
	require.Equal(t, webhookMaxRetryCount, stored.RetryCount)
}

func TestServiceProcessesDueSyncJobsWithQuota(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "order_sync", r.URL.Query().Get("capability_code"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"records":[{"id":"order-1","amount":10},{"id":"order-2","token":"hidden"},{"id":"order-3","status":"paid"}],"next_cursor":"cursor-3"}`))
	}))
	defer source.Close()
	t.Setenv("INTEGRATION_SYNC_SOURCE_URL_JD_SHOP", source.URL)
	seedIntegrationPlan(t, db, 9, true, 10)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	job := models.IntegrationSyncJob{
		TenantID: 9, TenantConnectionID: connection.ID, CapabilityCode: "order_sync",
		JobType: "incremental", TriggerMode: "schedule", Status: "pending", TotalCount: 3,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&job).Error)

	result, err := NewService(repositories.NewRepository(db)).ProcessDueSyncJobs(context.Background(), 10)

	require.NoError(t, err)
	require.Equal(t, 1, result.Scanned)
	require.Equal(t, 1, result.Completed)
	var stored models.IntegrationSyncJob
	require.NoError(t, db.First(&stored, job.ID).Error)
	require.Equal(t, "completed", stored.Status)
	require.Equal(t, int64(3), stored.SuccessCount)
	require.NotNil(t, stored.CursorValue)
	require.Equal(t, "cursor-3", *stored.CursorValue)
	require.NotNil(t, stored.FinishedAt)
	var records []models.IntegrationSyncRecord
	require.NoError(t, db.Where("sync_job_id = ?", job.ID).Order("external_id ASC").Find(&records).Error)
	require.Len(t, records, 3)
	require.Contains(t, records[1].Payload, `"token":"***"`)
}

func TestServiceQueuesSyncJobWhenRecordQuotaExceeded(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"order-1"},{"id":"order-2"}]`))
	}))
	defer source.Close()
	t.Setenv("INTEGRATION_SYNC_SOURCE_URL_JD_SHOP", source.URL)
	seedIntegrationPlan(t, db, 9, true, 1)
	platform := models.IntegrationPlatform{PlatformCode: "jd", PlatformName: "京东", PlatformType: "电商平台", AccessMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&platform).Error)
	app := models.IntegrationProviderApp{PlatformID: platform.ID, AppCode: "jd-shop", AppName: "京东店铺", AuthMode: "OAuth2", Status: "online", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&app).Error)
	connection := models.IntegrationTenantConnection{
		TenantID: 9, PlatformID: platform.ID, ProviderAppID: app.ID, ConnectionName: "旗舰店",
		AuthSubjectType: "shop", AuthSubjectID: "shop-1001", AuthSubjectName: "旗舰店",
		AuthScope: "[]", AuthStatus: "authorized", ConnectionStatus: "connected", TokenStatus: "valid",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&connection).Error)
	job := models.IntegrationSyncJob{
		TenantID: 9, TenantConnectionID: connection.ID, CapabilityCode: "order_sync",
		JobType: "incremental", TriggerMode: "schedule", Status: "pending", TotalCount: 2,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&job).Error)

	result, err := NewService(repositories.NewRepository(db)).ProcessDueSyncJobs(context.Background(), 10)

	require.NoError(t, err)
	require.Equal(t, 1, result.Queued)
	var stored models.IntegrationSyncJob
	require.NoError(t, db.First(&stored, job.ID).Error)
	require.Equal(t, "queued", stored.Status)
	require.NotNil(t, stored.NextRetryAt)
	require.NotNil(t, stored.ErrorCode)
	require.Equal(t, "quota_exceeded", *stored.ErrorCode)
	var count int64
	require.NoError(t, db.Model(&models.IntegrationSyncRecord{}).Where("sync_job_id = ?", job.ID).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestServiceSyncJobRetryLimitFailsJob(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	job := models.IntegrationSyncJob{
		TenantID: 9, TenantConnectionID: 999, CapabilityCode: "order_sync",
		JobType: "incremental", TriggerMode: "schedule", Status: "pending", RetryCount: syncJobMaxRetryCount - 1,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&job).Error)

	result, err := NewService(repositories.NewRepository(db)).ProcessDueSyncJobs(context.Background(), 10)

	require.NoError(t, err)
	require.Equal(t, 1, result.Failed)
	var stored models.IntegrationSyncJob
	require.NoError(t, db.First(&stored, job.ID).Error)
	require.Equal(t, "failed", stored.Status)
	require.Equal(t, syncJobMaxRetryCount, stored.RetryCount)
	require.NotNil(t, stored.ErrorCode)
	require.Equal(t, "sync_processing_failed", *stored.ErrorCode)
}

func TestServiceArchivesExpiredAPICallLogs(t *testing.T) {
	t.Setenv("INTEGRATION_API_CALL_LOG_RETENTION_DAYS", "30")
	db := newIntegrationCenterTestDB(t)
	now := time.Now()
	adminID := seedUser(t, db, 1, true)
	tenantID := uint64(1)
	oldEndpoint := "/old?secret=hidden"
	recentEndpoint := "/recent"
	require.NoError(t, db.Create(&models.IntegrationAPICallLog{
		TenantID: &tenantID, RequestID: "req-old", CallType: "third_party_api",
		Endpoint: &oldEndpoint, Status: "success", CalledAt: now.AddDate(0, 0, -45), CreatedAt: now.AddDate(0, 0, -45),
	}).Error)
	require.NoError(t, db.Create(&models.IntegrationAPICallLog{
		TenantID: &tenantID, RequestID: "req-recent", CallType: "third_party_api",
		Endpoint: &recentEndpoint, Status: "success", CalledAt: now, CreatedAt: now,
	}).Error)

	service := NewService(repositories.NewRepository(db))
	archived, err := service.ArchiveExpiredAPICallLogs(context.Background(), 100)

	require.NoError(t, err)
	require.Equal(t, int64(1), archived)
	var oldLog models.IntegrationAPICallLog
	require.NoError(t, db.Where("request_id = ?", "req-old").First(&oldLog).Error)
	require.NotNil(t, oldLog.ArchivedAt)
	require.NotNil(t, oldLog.RetentionBucket)
	require.Equal(t, "older_than_30_days", *oldLog.RetentionBucket)

	summary, err := service.Logs(context.Background(), adminID, PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), summary.Total)
	rows, ok := summary.Items.([]models.IntegrationAPICallLog)
	require.True(t, ok)
	require.Len(t, rows, 1)
	require.Equal(t, "req-recent", rows[0].RequestID)
}

func TestServiceExportsSanitizedAPICallLogsCSV(t *testing.T) {
	db := newIntegrationCenterTestDB(t)
	adminID := seedUser(t, db, 1, true)
	now := time.Now()
	tenantID := uint64(9)
	endpoint := "/orders?access_token=secret-token"
	errorMessage := "client_secret=hidden credential_ref=vault://integration/secret"
	require.NoError(t, db.Create(&models.IntegrationAPICallLog{
		TenantID:       &tenantID,
		RequestID:      "req-export",
		CallType:       "third_party_api",
		Method:         optionalString("POST"),
		Endpoint:       &endpoint,
		Status:         "failed",
		DurationMS:     37,
		ErrorMessage:   &errorMessage,
		CalledAt:       now,
		CreatedAt:      now,
		RequestDigest:  optionalString("sha256:req"),
		ResponseDigest: optionalString("sha256:resp"),
	}).Error)
	service := NewService(repositories.NewRepository(db))

	result, err := service.ExportLogs(context.Background(), adminID, RequestMeta{})

	require.NoError(t, err)
	require.Equal(t, 1, result.RowCount)
	require.Contains(t, result.Filename, "integration-call-logs-")
	content := string(result.Content)
	require.Contains(t, content, "req-export")
	require.Contains(t, content, "/orders")
	require.NotContains(t, content, "secret-token")
	require.NotContains(t, content, "hidden")
	require.NotContains(t, content, "vault://integration/secret")
}

func newIntegrationCenterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Tenant{},
		&models.AppUser{},
		&models.AuditLog{},
		&models.SaasPlan{},
		&models.SaasFeature{},
		&models.SaasQuota{},
		&models.SaasPlanFeature{},
		&models.SaasPlanQuota{},
		&models.TenantSubscription{},
		&models.TenantFeatureOverride{},
		&models.TenantQuotaOverride{},
		&models.IntegrationPlatform{},
		&models.IntegrationPlatformCapability{},
		&models.IntegrationProviderApp{},
		&models.IntegrationProviderAppCapability{},
		&models.IntegrationTenantConnection{},
		&models.IntegrationOAuthState{},
		&models.IntegrationTenantCapability{},
		&models.IntegrationSyncJob{},
		&models.IntegrationSyncRecord{},
		&models.IntegrationQuotaPolicy{},
		&models.IntegrationQuotaBinding{},
		&models.IntegrationQuotaUsage{},
		&models.IntegrationAlert{},
		&models.IntegrationAPICallLog{},
		&models.IntegrationWebhookEvent{},
	))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, tenantID uint64, platformAdmin bool) uint64 {
	t.Helper()
	now := time.Now()
	user := models.AppUser{
		TenantID:        tenantID,
		EmployeeNo:      "u-test-" + strconv.FormatUint(tenantID, 10) + "-" + strconv.FormatBool(platformAdmin),
		Account:         "test",
		PasswordHash:    "hash",
		Name:            "测试用户",
		Status:          1,
		IsPlatformAdmin: platformAdmin,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, db.Create(&user).Error)
	return user.ID
}

func boolPtr(value bool) *bool {
	return &value
}

func signWebhook(secret string, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func seedIntegrationPlan(t *testing.T, db *gorm.DB, tenantID uint64, featureEnabled bool, connectionLimit int) {
	t.Helper()
	now := time.Now()
	plan := models.SaasPlan{
		PlanCode:     "integration-plan-" + strconv.FormatUint(tenantID, 10),
		PlanName:     "集成套餐",
		PlanType:     "PAID",
		BillingCycle: "MONTH",
		Status:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	require.NoError(t, db.Create(&plan).Error)
	feature := models.SaasFeature{
		FeatureCode: "integration_tenant_authorization",
		FeatureName: "第三方租户授权",
		FeatureType: "FEATURE",
		AppCode:     "integration-center",
		Status:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, db.Create(&feature).Error)
	quota := models.SaasQuota{
		QuotaCode: "integration_connection_count",
		QuotaName: "第三方连接实例数",
		QuotaType: "STATIC",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(&quota).Error)
	apiQuota := models.SaasQuota{
		QuotaCode: "integration_api_calls_daily",
		QuotaName: "第三方 API 日调用次数",
		QuotaType: "PERIOD",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(&apiQuota).Error)
	syncQuota := models.SaasQuota{
		QuotaCode: "integration_sync_records_daily",
		QuotaName: "第三方同步日记录数",
		QuotaType: "PERIOD",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(&syncQuota).Error)
	if featureEnabled {
		require.NoError(t, db.Create(&models.SaasPlanFeature{PlanID: plan.ID, FeatureID: feature.ID, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	}
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: quota.ID, QuotaValue: connectionLimit, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: apiQuota.ID, QuotaValue: connectionLimit, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: syncQuota.ID, QuotaValue: connectionLimit, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.TenantSubscription{
		TenantID:           tenantID,
		PlanID:             plan.ID,
		SubscriptionStatus: "ACTIVE",
		StartTime:          now.Add(-time.Hour),
		CreatedAt:          now,
		UpdatedAt:          now,
	}).Error)
}
