package services

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/app_center/dto"
	"saas_baseon_go/internal/apps/app_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestAppCenterListRequiresPlatformAdmin(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "tenant", Name: "租户用户", Status: 1}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	_, err := service.ListApps(context.Background(), 1, dto.AppListRequest{Limit: 20})

	require.True(t, errors.Is(err, ErrPlatformOnly))
}

func TestAppCenterListReturnsBuiltinAppsForPlatformAdmin(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "app-center", AppName: "应用中心", AppType: "SYSTEM_APP", Source: "BUILTIN", Status: "ONLINE", ChargeMode: "NON_SELLABLE", VisibilityScope: "PLATFORM_ONLY", IsBuiltin: true, IsPlatformOnly: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ListApps(context.Background(), 1, dto.AppListRequest{Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Equal(t, "app-center", result.Items[0].AppCode)
}

func TestAppCenterStatsUsesProductionTables(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "client", AppName: "客户端", AppType: "CLIENT_APP", Source: "MANIFEST", Status: "ONLINE", ChargeMode: "FREE", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "planned-app", AppName: "规划应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "PLANNED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "developing-app", AppName: "开发应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "DEVELOPING", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.DictType{ID: 10, TenantID: 1, Code: "app_type", Name: "应用类型", Scope: "platform"}).Error)
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: 10, Label: "客户端型", Value: "CLIENT_APP", Enabled: true}).Error)
	require.NoError(t, db.Create(&models.TenantSubscription{TenantID: 2, PlanID: 1, SubscriptionStatus: "TRIAL"}).Error)
	require.NoError(t, db.Create(&models.AuditLog{Module: "app_center", Action: "load", Summary: "装载应用"}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.Stats(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, int64(3), result.Total)
	require.Equal(t, int64(2), result.Developing)
	require.Equal(t, int64(1), result.Categories)
	require.Equal(t, int64(1), result.ClientApps)
	require.Equal(t, int64(1), result.TenantOpenings)
	require.Equal(t, int64(1), result.TrialInvites)
	require.Equal(t, int64(1), result.ManifestLoads)
	require.Equal(t, int64(1), result.AuditLogs)
}

func TestAppCenterListCanFilterPlannedAndDevelopingTogether(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "planned-app", AppName: "规划应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "PLANNED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "developing-app", AppName: "开发应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "DEVELOPING", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "online-app", AppName: "线上应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "ONLINE", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ListApps(context.Background(), 1, dto.AppListRequest{Status: "PLANNED,DEVELOPING", Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)
	require.ElementsMatch(t, []string{"planned-app", "developing-app"}, []string{result.Items[0].AppCode, result.Items[1].AppCode})
}

func TestAppCenterCreatePersistsManualApp(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	created, err := service.CreateApp(context.Background(), 1, dto.AppCreateRequest{
		AppCode:         "crm-suite",
		AppName:         "客户管理",
		AppType:         "BUSINESS_APP",
		Status:          "DRAFT",
		ChargeMode:      "SUBSCRIPTION",
		VisibilityScope: "TENANT",
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "crm-suite", created.AppCode)
	require.Equal(t, "MANUAL", created.Source)
	require.False(t, created.IsBuiltin)
	require.False(t, created.IsPlatformOnly)

	detail, err := service.GetApp(context.Background(), 1, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.AppCode, detail.AppCode)
}

func TestAppCenterCreateRejectsDuplicateCode(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "app-center", AppName: "应用中心", AppType: "SYSTEM_APP", Source: "BUILTIN", Status: "ONLINE", ChargeMode: "NON_SELLABLE", VisibilityScope: "PLATFORM_ONLY", IsBuiltin: true, IsPlatformOnly: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	_, err := service.CreateApp(context.Background(), 1, dto.AppCreateRequest{AppCode: "app-center", AppName: "重复应用"})

	require.True(t, errors.Is(err, ErrAppCodeExists))
}

func TestAppCenterUpdatePersistsBasicFields(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{ID: 10, AppCode: "crm-suite", AppName: "客户管理", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "DRAFT", ChargeMode: "SUBSCRIPTION", VisibilityScope: "PLATFORM_ONLY", IsPlatformOnly: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	updated, err := service.UpdateApp(context.Background(), 1, 10, dto.AppUpdateRequest{
		AppName:         "客户经营套件",
		AppType:         "SUITE_APP",
		ChargeMode:      "FREE",
		VisibilityScope: "TENANT",
		SortOrder:       8,
	})

	require.NoError(t, err)
	require.Equal(t, "crm-suite", updated.AppCode)
	require.Equal(t, "客户经营套件", updated.AppName)
	require.Equal(t, "SUITE_APP", updated.AppType)
	require.False(t, updated.IsPlatformOnly)
	require.Equal(t, 8, updated.SortOrder)
}

func TestAppCenterUpdateStatusRejectsDisablingBuiltinApps(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{ID: 10, AppCode: "app-center", AppName: "应用中心", AppType: "SYSTEM_APP", Source: "BUILTIN", Status: "ONLINE", ChargeMode: "NON_SELLABLE", VisibilityScope: "PLATFORM_ONLY", IsBuiltin: true, IsPlatformOnly: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	_, err := service.UpdateAppStatus(context.Background(), 1, 10, dto.AppStatusRequest{Status: "DISABLED"})

	require.True(t, errors.Is(err, ErrBuiltinStatusImmutable))
}

func TestAppCenterUpdateStatusPersistsManualAppStatus(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{ID: 10, AppCode: "crm-suite", AppName: "客户管理", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "DRAFT", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	updated, err := service.UpdateAppStatus(context.Background(), 1, 10, dto.AppStatusRequest{Status: "ONLINE"})

	require.NoError(t, err)
	require.Equal(t, "ONLINE", updated.Status)
}

func newAppCenterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.AppUser{},
		&models.SysApp{},
		&models.DictType{},
		&models.DictItem{},
		&models.TenantSubscription{},
		&models.AuditLog{},
	))
	return db
}
