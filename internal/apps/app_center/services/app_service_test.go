package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	require.NoError(t, db.Create(&models.SysApp{ID: 100, AppCode: "client", AppName: "客户端形态应用", AppType: "BUSINESS_APP", Source: "MANIFEST", Status: "ONLINE", ChargeMode: "FREE", VisibilityScope: "TENANT", DeploymentMode: "MERGED"}).Error)
	require.NoError(t, db.Create(&models.SysAppClient{AppID: 100, ClientCode: "PC_WEB", ClientName: "PC Web", Enabled: true, SortOrder: 1}).Error)
	require.NoError(t, db.Create(&models.SysAppClient{AppID: 100, ClientCode: "HARMONYOS", ClientName: "鸿蒙", Enabled: true, SortOrder: 2}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "planned-app", AppName: "规划应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "PLANNED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "developing-app", AppName: "开发应用", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "DEVELOPING", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)
	require.NoError(t, db.Create(&models.DictType{ID: 10, TenantID: 1, Code: "app_type", Name: "应用类型", Scope: "platform"}).Error)
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: 10, Label: "业务系统", Value: "BUSINESS_APP", Enabled: true}).Error)
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
		DeploymentMode:  "STANDALONE",
		CommModes:       stringPtr("PLATFORM_API,WEBHOOK"),
		Status:          "INITIATED",
		ChargeMode:      "SUBSCRIPTION",
		VisibilityScope: "TENANT",
		Owner:           stringPtr("平台架构组, 产品负责人"),
		OwnerUserIDs:    stringPtr("1,2"),
		VisibilityMode:  stringPtr("SPECIFIED_TENANTS"),
		VisibleTenants:  stringPtr("演示主体"),
		OpenMethod:      stringPtr("ADMIN_GRANT"),
		TrialPolicy:     stringPtr("14 天"),
		TrialStartRule:  stringPtr("首次安装时开始"),
		ReleaseChannel:  stringPtr("DEV"),
		ReleaseNote:     stringPtr("创建应用主档"),
		Description:     stringPtr("面向销售团队的客户经营入口"),
		DetailDesc:      stringPtr("<p>管理客户、商机、跟进与经营数据。</p>"),
		Clients: []dto.AppClientRequest{
			{ClientCode: "pc_web", ClientName: "PC Web", Enabled: true, SortOrder: 1},
			{ClientCode: "api_only", ClientName: "API Only", Enabled: true, SortOrder: 2},
		},
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "crm-suite", created.AppCode)
	require.Equal(t, "MANUAL", created.Source)
	require.Equal(t, "INITIATED", created.Status)
	require.Equal(t, "STANDALONE", created.DeploymentMode)
	require.Equal(t, "PLATFORM_API,WEBHOOK", *created.CommModes)
	require.Equal(t, "1,2", *created.OwnerUserIDs)
	require.Equal(t, "面向销售团队的客户经营入口", *created.Description)
	require.Equal(t, "<p>管理客户、商机、跟进与经营数据。</p>", *created.DetailDesc)
	require.False(t, created.IsBuiltin)
	require.False(t, created.IsPlatformOnly)
	require.Len(t, created.Clients, 2)
	require.Equal(t, "PC_WEB", created.Clients[0].ClientCode)

	detail, err := service.GetApp(context.Background(), 1, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.AppCode, detail.AppCode)
	require.Equal(t, "STANDALONE", detail.DeploymentMode)
	require.Equal(t, "PLATFORM_API,WEBHOOK", *detail.CommModes)
	require.Equal(t, "演示主体", *detail.VisibleTenants)
	require.Equal(t, "<p>管理客户、商机、跟进与经营数据。</p>", *detail.DetailDesc)
	require.Len(t, detail.Clients, 2)
	require.NotNil(t, detail.Assets)
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
	require.NoError(t, db.Create(&models.SysApp{ID: 10, AppCode: "crm-suite", AppName: "客户管理", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "INITIATED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "PLATFORM_ONLY", IsPlatformOnly: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	updated, err := service.UpdateApp(context.Background(), 1, 10, dto.AppUpdateRequest{
		AppName:         "客户经营套件",
		AppType:         "SUITE_APP",
		ChargeMode:      "FREE",
		VisibilityScope: "TENANT",
		SortOrder:       8,
		ReleaseNote:     stringPtr("更新应用配置"),
		Clients: []dto.AppClientRequest{
			{ClientCode: "H5", ClientName: "H5", Enabled: true, SortOrder: 1},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "crm-suite", updated.AppCode)
	require.Equal(t, "客户经营套件", updated.AppName)
	require.Equal(t, "SUITE_APP", updated.AppType)
	require.False(t, updated.IsPlatformOnly)
	require.Equal(t, 8, updated.SortOrder)
	require.Equal(t, "更新应用配置", *updated.ReleaseNote)
	require.Len(t, updated.Clients, 1)
	require.Equal(t, "H5", updated.Clients[0].ClientCode)
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
	require.NoError(t, db.Create(&models.SysApp{ID: 10, AppCode: "crm-suite", AppName: "客户管理", AppType: "BUSINESS_APP", Source: "MANUAL", Status: "INITIATED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT"}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	updated, err := service.UpdateAppStatus(context.Background(), 1, 10, dto.AppStatusRequest{Status: "ONLINE"})

	require.NoError(t, err)
	require.Equal(t, "ONLINE", updated.Status)
}

func TestAppCenterParseManifestReturnsAssetSummary(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ParseManifestContent(context.Background(), 1, "crm.manifest.yaml", []byte(`
manifest_version: "1.0"
fragment_role: main
app:
  app_code: crm-suite
  app_name: 客户管理
  app_type: BUSINESS_APP
  source: MANIFEST
  status: INITIATED
  deployment_mode: MERGED
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
menus:
  - code: crm_list
    name: 客户列表
    path: /crm
    include_in_package: true
    feature_code: crm_manage
permissions:
  - code: /crm
    name: 客户列表
    type: MENU
    menu_code: crm_list
package_features:
  - feature_code: crm_manage
    feature_name: 客户管理
    feature_type: MENU
    include_in_package: true
quotas:
  - quota_code: max_customers
    quota_name: 客户数
    quota_type: STATIC
    unit: COUNT
`))

	require.NoError(t, err)
	require.True(t, result.Valid, "blockers: %v", result.Blockers)
	require.True(t, result.Importable)
	require.Equal(t, "crm-suite", result.AppCode)
	require.Equal(t, "客户管理", result.AppName)
	require.Equal(t, 1, result.Counts.Clients)
	require.Equal(t, 1, result.Counts.Menus)
	require.Equal(t, 1, result.Counts.PackageFeatures)
	require.Equal(t, 1, result.Counts.Quotas)
	require.NotEmpty(t, result.ManifestHash)
}

func TestAppCenterParseAICapabilityCenterManifest(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "apps", "ai_capability_center", "app.manifest.yaml"))
	require.NoError(t, err)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ParseManifestContent(context.Background(), 1, "app.manifest.yaml", raw)

	require.NoError(t, err)
	require.True(t, result.Valid, "blockers: %v", result.Blockers)
	require.Equal(t, "ai-capability-center", result.AppCode)
	require.Equal(t, "PLATFORM_ONLY", result.VisibilityScope)
	require.Equal(t, "NON_SELLABLE", result.ChargePolicy)
	require.Equal(t, "NONE", result.BillingMode)
	require.Equal(t, "NON_SELLABLE", result.PackagePolicy)
	require.Equal(t, 9, result.Counts.Menus)
	require.Equal(t, 0, result.Counts.PackageFeatures)
	require.Equal(t, 0, result.Counts.Quotas)
	require.NotEmpty(t, result.ManifestHash)
}

func TestAppCenterParseIntegrationCenterManifest(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	service := NewAppService(repositories.NewAppRepository(db))
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "apps", "integration_center", "app.manifest.yaml"))
	require.NoError(t, err)

	result, err := service.ParseManifestContent(context.Background(), 1, "app.manifest.yaml", raw)
	require.NoError(t, err)
	require.True(t, result.Valid, "blockers: %v", result.Blockers)
	require.Empty(t, result.Blockers)
	require.Equal(t, "integration-center", result.AppCode)
	require.Equal(t, 9, result.Counts.Menus)
	require.Equal(t, 4, result.Counts.Operations)
	require.Equal(t, 13, result.Counts.Permissions)
	require.Equal(t, 49, result.Counts.APIs)
	require.Equal(t, 2, result.Counts.PackageFeatures)
	require.Equal(t, 3, result.Counts.Quotas)
}

func TestAppCenterLoadPlatformOnlyManifestKeepsAssetsOutOfPackages(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", IsPlatform: true, Status: 1}).Error)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.Role{ID: 1, TenantID: 1, Code: "admin", Name: "超级管理员", Status: 1}).Error)

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "apps", "ai_capability_center", "app.manifest.yaml"))
	require.NoError(t, err)

	service := NewAppService(repositories.NewAppRepository(db))
	loaded, err := service.LoadManifest(context.Background(), 1, dto.ManifestLoadRequest{
		FileName:   "app.manifest.yaml",
		Content:    string(raw),
		SourceType: "LOCAL_FILE",
	})
	require.NoError(t, err)
	require.Equal(t, "SUCCESS", loaded.Status)

	var app models.SysApp
	require.NoError(t, db.Where("app_code = ?", "ai-capability-center").First(&app).Error)
	require.Equal(t, "PLATFORM_ONLY", app.VisibilityScope)
	require.Equal(t, "NON_SELLABLE", app.ChargeMode)
	require.True(t, app.IsPlatformOnly)

	var tenantVisibleMenus int64
	require.NoError(t, db.Model(&models.SysAppEntry{}).
		Where("app_code = ? AND (platform_only = ? OR tenant_visible = ? OR include_in_package = ? OR data_perm_mode <> ?)", "ai-capability-center", false, true, true, "NONE").
		Count(&tenantVisibleMenus).Error)
	require.Zero(t, tenantVisibleMenus)

	var packagedPermissions int64
	require.NoError(t, db.Model(&models.Permission{}).
		Where("app_code = ? AND (is_platform_only = ? OR visible = ? OR is_package_feature = ? OR data_perm_mode <> ?)", "ai-capability-center", false, true, true, "NONE").
		Count(&packagedPermissions).Error)
	require.Zero(t, packagedPermissions)

	var packageFeatures int64
	require.NoError(t, db.Model(&models.SaasFeature{}).Where("app_code = ? AND status = ?", "ai-capability-center", 1).Count(&packageFeatures).Error)
	require.Zero(t, packageFeatures)

	var quotas int64
	require.NoError(t, db.Model(&models.SysAppQuota{}).Where("app_code = ? AND status = ?", "ai-capability-center", "ACTIVE").Count(&quotas).Error)
	require.Zero(t, quotas)

	var permissionCount int64
	require.NoError(t, db.Model(&models.Permission{}).Where("app_code = ? AND enabled = ?", "ai-capability-center", true).Count(&permissionCount).Error)
	require.Equal(t, int64(11), permissionCount)

	var rolePermissionCount int64
	require.NoError(t, db.Table("role_permission rp").
		Joins("JOIN permission p ON p.id = rp.permission_id").
		Where("rp.role_id = ? AND p.app_code = ? AND p.deleted_at IS NULL", 1, "ai-capability-center").
		Count(&rolePermissionCount).Error)
	require.Equal(t, permissionCount, rolePermissionCount)
}

func TestAppCenterParseManifestBlocksExistingAppCodeForNewImport(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "app-center", AppName: "应用中心", AppType: "SYSTEM_APP", Source: "BUILTIN", Status: "ONLINE", ChargeMode: "NON_SELLABLE", VisibilityScope: "PLATFORM_ONLY", IsBuiltin: true, IsPlatformOnly: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ParseManifestContent(context.Background(), 1, "app.manifest.yaml", []byte(`
manifest_version: "1.0"
fragment_role: main
app:
  app_code: app-center
  app_name: 应用中心
  app_type: SYSTEM_APP
  source: BUILTIN
  status: ONLINE
  package_policy: NON_SELLABLE
clients:
  - PC_WEB
package_features: []
`))

	require.NoError(t, err)
	require.True(t, result.Exists)
	require.False(t, result.Importable)
	require.False(t, result.Valid)
	require.Contains(t, result.Blockers[0], "app_code 已存在")
}

func TestAppCenterParseManifestBlocksStandaloneWithoutAccessContract(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ParseManifestContent(context.Background(), 1, "external.manifest.yaml", []byte(`
manifest_version: "1.0"
fragment_role: main
app:
  app_code: external-crm
  app_name: 外部客户系统
  app_type: BUSINESS_APP
  source: MANIFEST
  status: INITIATED
  deployment_mode: STANDALONE
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
package_features:
  - feature_code: external_crm
    feature_name: 外部客户系统
    feature_type: SERVICE
    include_in_package: true
`))

	require.NoError(t, err)
	require.False(t, result.Valid)
	require.False(t, result.Importable)
	require.Contains(t, result.Blockers, "独立部署应用必须声明 app.communication_modes")
	require.Contains(t, result.Blockers, "独立部署应用必须声明 app.open_api_scopes")
	require.Contains(t, result.Blockers, "独立部署应用必须声明 app.tenant_context")
	require.Contains(t, result.Blockers, "独立部署应用必须声明 app.signature_strategy")
	require.Contains(t, result.Blockers, "独立部署应用必须声明 app.idempotency_strategy")
}

func TestAppCenterLoadStandaloneManifestPersistsAccessContract(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", IsPlatform: true, Status: 1}).Error)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	manifest := `
manifest_version: "1.0"
fragment_role: main
app:
  app_code: external-billing
  app_name: 外部计费系统
  app_type: API_APP
  source: MANIFEST
  status: INITIATED
  deployment_mode: STANDALONE
  communication_modes:
    - PLATFORM_API
    - WEBHOOK
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
  api_base_url: https://billing.example.com/api
  webhook_url: https://billing.example.com/webhook
  health_check_url: https://billing.example.com/health
  open_api_scopes:
    - tenant.read
    - user.read
  tenant_context: 由底座签发的租户上下文头 X-Tenant-ID 透传
  signature_strategy: HMAC-SHA256 请求签名
  idempotency_strategy: 写入请求必须携带 Idempotency-Key
clients:
  - API_ONLY
package_features:
  - feature_code: external_billing_api
    feature_name: 外部计费 API
    feature_type: SERVICE
    include_in_package: true
`
	service := NewAppService(repositories.NewAppRepository(db))
	loaded, err := service.LoadManifest(context.Background(), 1, dto.ManifestLoadRequest{
		FileName:   "external.manifest.yaml",
		Content:    manifest,
		SourceType: "UPLOAD",
	})

	require.NoError(t, err)
	require.Equal(t, "SUCCESS", loaded.Status)
	var app models.SysApp
	require.NoError(t, db.Where("app_code = ?", "external-billing").First(&app).Error)
	require.Equal(t, "STANDALONE", app.DeploymentMode)
	require.NotNil(t, app.CommModes)
	require.Equal(t, "PLATFORM_API,WEBHOOK", *app.CommModes)
	require.NotNil(t, app.APIBaseURL)
	require.Equal(t, "https://billing.example.com/api", *app.APIBaseURL)
	require.NotNil(t, app.WebhookURL)
	require.Equal(t, "https://billing.example.com/webhook", *app.WebhookURL)
	require.NotNil(t, app.HealthCheckURL)
	require.Equal(t, "https://billing.example.com/health", *app.HealthCheckURL)
}

func TestAppCenterScanManifestsReadsManifestFiles(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "internal/apps/demo"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "internal/apps/demo/app.manifest.yaml"), []byte(`
manifest_version: "1.0"
app:
  app_code: demo-app
  app_name: 示例应用
  app_type: BUSINESS_APP
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
package_features:
  - feature_code: demo_manage
    feature_name: 示例应用
    feature_type: MENU
    include_in_package: true
`), 0o644))

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ScanManifests(context.Background(), 1, root)

	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
	require.Equal(t, 1, result.ImportableCount)
	require.Equal(t, "demo-app", result.Items[0].AppCode)
	require.Len(t, result.Groups, 1)
	require.Equal(t, "demo-app", result.Groups[0].AppCode)
	require.True(t, result.Groups[0].Loadable)
}

func TestAppCenterScanManifestsGroupsFragmentsByAppCode(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "internal/apps/crm"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "frontend/src/apps/crm"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "internal/apps/crm/app.manifest.yaml"), []byte(`
manifest_version: "1.0"
fragment_role: main
app:
  app_code: crm-suite
  app_name: 客户管理
  app_type: BUSINESS_APP
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
menus:
  - code: crm_list
    name: 客户列表
    path: /crm
permissions:
  - code: /crm
    name: 客户列表
    type: MENU
    menu_code: crm_list
package_features:
  - feature_code: crm_manage
    feature_name: 客户管理
    include_in_package: true
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "frontend/src/apps/crm/app.manifest.yaml"), []byte(`
manifest_version: "1.0"
fragment_role: frontend
app:
  app_code: crm-suite
  app_name: 客户管理
  package_policy: IN_PACKAGE
clients:
  - H5
operations:
  - code: crm_export
    name: 导出客户
    permission_code: crm:export
    menu_code: crm_list
permissions:
  - code: crm:export
    name: 导出客户
    type: BUTTON
package_features:
  - feature_code: button_crm_export
    feature_name: 导出客户
    include_in_package: true
`), 0o644))

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ScanManifests(context.Background(), 1, root)

	require.NoError(t, err)
	require.Equal(t, 2, result.Total)
	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	require.Equal(t, "crm-suite", group.AppCode)
	require.True(t, group.Loadable, "blockers: %v", group.Blockers)
	require.Equal(t, 2, group.FragmentCount)
	require.Equal(t, 1, group.MainCount)
	require.ElementsMatch(t, []string{"PC_WEB", "H5"}, group.Merged.ClientCodes)
	require.Equal(t, 1, group.Merged.Counts.Menus)
	require.Equal(t, 1, group.Merged.Counts.Operations)
	require.Equal(t, 2, group.Merged.Counts.PackageFeatures)
}

func TestAppCenterScanManifestsBlocksMultipleMainFragments(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "a"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "b"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example\n"), 0o644))
	manifest := []byte(`
manifest_version: "1.0"
fragment_role: main
app:
  app_code: duplicate-main
  app_name: 重复主片段
  app_type: BUSINESS_APP
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
package_features:
  - feature_code: duplicate_main
    feature_name: 重复主片段
    include_in_package: true
`)
	require.NoError(t, os.WriteFile(filepath.Join(root, "a/app.manifest.yaml"), manifest, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b/app.manifest.yaml"), manifest, 0o644))

	service := NewAppService(repositories.NewAppRepository(db))
	result, err := service.ScanManifests(context.Background(), 1, root)

	require.NoError(t, err)
	require.Len(t, result.Groups, 1)
	require.False(t, result.Groups[0].Loadable)
	require.Equal(t, 2, result.Groups[0].MainCount)
	require.Contains(t, result.Groups[0].Blockers[0], "只能包含一个")
}

func TestAppCenterLoadManifestCreatesAssetsAndPackageCenter(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", IsPlatform: true, Status: 1}).Error)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)

	manifest := `manifest_version: "1.0"
fragment_role: main
app:
  app_code: ops-console
  app_name: 运营控制台
  app_type: BUSINESS_APP
  source: MANIFEST
  status: INITIATED
  deployment_mode: MERGED
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
menus:
  - code: ops_dashboard
    name: 运营看板
    path: /ops/dashboard
    sort_order: 10
    tenant_visible: true
    tenant_editable: true
    include_in_package: true
    feature_code: ops_dashboard
operations:
  - code: ops_export
    name: 导出看板
    menu_code: ops_dashboard
    permission_code: ops:export
    include_in_package: true
    feature_code: button_ops_export
permissions:
  - code: /ops/dashboard
    name: 运营看板
    type: MENU
    menu_code: ops_dashboard
    include_in_package: true
  - code: ops:read
    name: 查看运营数据
    type: OPERATION
    menu_code: ops_dashboard
    include_in_package: true
apis:
  - method: GET
    path: /api/ops/dashboard
    permission_code: ops:read
package_features:
  - feature_code: ops_dashboard
    feature_name: 运营看板
    feature_type: MENU
    source_code: ops_dashboard
    include_in_package: true
  - feature_code: button_ops_export
    feature_name: 导出看板
    feature_type: OPERATION
    parent_code: ops_dashboard
    source_code: ops:export
    include_in_package: true
quotas:
  - quota_code: max_ops_exports
    quota_name: 导出次数
    quota_type: PERIODIC
    unit: COUNT
    period_type: MONTH
    include_in_package: true
`
	service := NewAppService(repositories.NewAppRepository(db))
	diff, err := service.DiffManifest(context.Background(), 1, "app.manifest.yaml", "", []byte(manifest))
	require.NoError(t, err)
	require.True(t, diff.Loadable)
	require.Equal(t, "CREATE", diff.Mode)
	require.Greater(t, diff.Summary.Create, 0)

	loaded, err := service.LoadManifest(context.Background(), 1, dto.ManifestLoadRequest{
		FileName:   "app.manifest.yaml",
		Content:    manifest,
		SourceType: "UPLOAD",
	})
	require.NoError(t, err)
	require.Equal(t, "SUCCESS", loaded.Status)
	require.NotZero(t, loaded.LoadID)

	var app models.SysApp
	require.NoError(t, db.Where("app_code = ?", "ops-console").First(&app).Error)
	require.Equal(t, "运营控制台", app.AppName)
	require.NotNil(t, app.ManifestHash)

	var client models.SysAppClient
	require.NoError(t, db.Where("app_id = ? AND client_code = ?", app.ID, "PC_WEB").First(&client).Error)
	require.Equal(t, "PC Web", client.ClientName)

	var menu models.SysAppEntry
	require.NoError(t, db.Where("app_code = ? AND resource_code = ?", "ops-console", "ops_dashboard").First(&menu).Error)
	require.True(t, menu.IncludeInPackage)

	var api models.SysAppAPI
	require.NoError(t, db.Where("app_code = ? AND method = ? AND path = ?", "ops-console", "GET", "/api/ops/dashboard").First(&api).Error)
	require.Equal(t, "ops:read", *api.PermissionCode)

	var permission models.Permission
	require.NoError(t, db.Where("tenant_id = ? AND path = ?", uint64(1), "/ops/dashboard").First(&permission).Error)
	require.Equal(t, "ops-console", permission.AppCode)

	var feature models.SaasFeature
	require.NoError(t, db.Where("feature_code = ?", "ops_dashboard").First(&feature).Error)
	require.Equal(t, "ops-console", feature.AppCode)

	var quota models.SaasQuota
	require.NoError(t, db.Where("quota_code = ?", "max_ops_exports").First(&quota).Error)
	require.Equal(t, "导出次数", quota.QuotaName)

	detail, err := service.GetApp(context.Background(), 1, app.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Assets)
	require.Len(t, detail.Assets.Entries, 1)
	require.Len(t, detail.Assets.APIs, 1)
	require.Len(t, detail.Assets.Permissions, 3)
	require.Len(t, detail.Assets.PackageFeatures, 2)
	require.Len(t, detail.Assets.Quotas, 1)
	require.Len(t, detail.Assets.ManifestLoads, 1)
	require.Equal(t, "ops_dashboard", detail.Assets.Entries[0].ResourceCode)
	require.Equal(t, "/api/ops/dashboard", detail.Assets.APIs[0].Path)
	require.Equal(t, "SUCCESS", detail.Assets.ManifestLoads[0].Status)
}

func TestAppCenterDiffManifestMarksMissingAssetsAsDisable(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", IsPlatform: true, Status: 1}).Error)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "ops-console", AppName: "运营控制台", AppType: "BUSINESS_APP", Source: "MANIFEST", Status: "INITIATED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT", DeploymentMode: "MERGED"}).Error)
	require.NoError(t, db.Create(&models.SysAppEntry{AppCode: "ops-console", ResourceCode: "old_menu", Name: "旧菜单", ManagedByManifest: true, Status: "ACTIVE"}).Error)
	require.NoError(t, db.Create(&models.SysAppAPI{AppCode: "ops-console", Method: "GET", Path: "/api/old", ManagedByManifest: true, Status: "ACTIVE"}).Error)
	require.NoError(t, db.Create(&models.SysAppPermission{AppCode: "ops-console", PermissionCode: "old_perm", Name: "旧权限", ManagedByManifest: true, Status: "ACTIVE"}).Error)
	require.NoError(t, db.Create(&models.SysAppPackageFeature{AppCode: "ops-console", FeatureCode: "old_feature", FeatureName: "旧功能", ManagedByManifest: true, Status: "ACTIVE"}).Error)
	require.NoError(t, db.Create(&models.SysAppQuota{AppCode: "ops-console", QuotaCode: "old_quota", QuotaName: "旧配额", ManagedByManifest: true, Status: "ACTIVE"}).Error)
	require.NoError(t, db.Create(&models.SaasFeature{FeatureCode: "old_feature", FeatureName: "旧功能", FeatureType: "MENU", AppCode: "ops-console", Status: 1}).Error)
	require.NoError(t, db.Create(&models.SaasQuota{QuotaCode: "old_quota", QuotaName: "旧配额", QuotaType: "STATIC", Status: 1}).Error)

	manifest := `manifest_version: "1.0"
fragment_role: main
app:
  app_code: ops-console
  app_name: 运营控制台
  app_type: BUSINESS_APP
  source: MANIFEST
  status: INITIATED
  deployment_mode: MERGED
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
menus:
  - code: ops_dashboard
    name: 运营看板
    path: /ops/dashboard
permissions:
  - code: /ops/dashboard
    name: 运营看板
    type: MENU
    menu_code: ops_dashboard
package_features:
  - feature_code: ops_dashboard
    feature_name: 运营看板
    include_in_package: true
`
	service := NewAppService(repositories.NewAppRepository(db))
	diff, err := service.DiffManifest(context.Background(), 1, "app.manifest.yaml", "", []byte(manifest))

	require.NoError(t, err)
	require.True(t, diff.Loadable)
	require.GreaterOrEqual(t, diff.Summary.Disable, 3)
	disabled := map[string]bool{}
	for _, change := range diff.Changes {
		if change.Action == "DISABLE" {
			disabled[change.ResourceCode] = true
		}
	}
	require.True(t, disabled["old_menu"])
	require.True(t, disabled["GET /api/old"])
	require.True(t, disabled["old_perm"])
	require.True(t, disabled["old_feature"])
	require.True(t, disabled["old_quota"])

	loaded, err := service.LoadManifest(context.Background(), 1, dto.ManifestLoadRequest{
		FileName:   "app.manifest.yaml",
		Content:    manifest,
		SourceType: "UPLOAD",
	})
	require.NoError(t, err)
	require.Equal(t, "SUCCESS", loaded.Status)

	var oldMenu models.SysAppEntry
	require.NoError(t, db.Where("app_code = ? AND resource_code = ?", "ops-console", "old_menu").First(&oldMenu).Error)
	require.Equal(t, "DISABLED", oldMenu.Status)

	var oldAPI models.SysAppAPI
	require.NoError(t, db.Where("app_code = ? AND method = ? AND path = ?", "ops-console", "GET", "/api/old").First(&oldAPI).Error)
	require.Equal(t, "DISABLED", oldAPI.Status)

	var oldPermission models.SysAppPermission
	require.NoError(t, db.Where("app_code = ? AND permission_code = ?", "ops-console", "old_perm").First(&oldPermission).Error)
	require.Equal(t, "DISABLED", oldPermission.Status)

	var oldFeature models.SysAppPackageFeature
	require.NoError(t, db.Where("app_code = ? AND feature_code = ?", "ops-console", "old_feature").First(&oldFeature).Error)
	require.Equal(t, "DISABLED", oldFeature.Status)

	var oldQuota models.SysAppQuota
	require.NoError(t, db.Where("app_code = ? AND quota_code = ?", "ops-console", "old_quota").First(&oldQuota).Error)
	require.Equal(t, "DISABLED", oldQuota.Status)

	var packageFeature models.SaasFeature
	require.NoError(t, db.Where("feature_code = ?", "old_feature").First(&packageFeature).Error)
	require.Equal(t, 0, packageFeature.Status)

	var packageQuota models.SaasQuota
	require.NoError(t, db.Where("quota_code = ?", "old_quota").First(&packageQuota).Error)
	require.Equal(t, 0, packageQuota.Status)
}

func TestAppCenterDiffBlocksProtectedManifestAssetOverwrite(t *testing.T) {
	db := newAppCenterTestDB(t)
	require.NoError(t, db.Create(&models.AppUser{ID: 1, TenantID: 1, Account: "admin", Name: "平台管理员", Status: 1, IsPlatformAdmin: true}).Error)
	require.NoError(t, db.Create(&models.SysApp{AppCode: "ops-console", AppName: "运营控制台", AppType: "BUSINESS_APP", Source: "MANIFEST", Status: "INITIATED", ChargeMode: "SUBSCRIPTION", VisibilityScope: "TENANT", DeploymentMode: "MERGED"}).Error)
	require.NoError(t, db.Create(&models.SysAppEntry{
		AppCode:           "ops-console",
		ResourceCode:      "ops_dashboard",
		Name:              "人工运营看板",
		Path:              "/ops/manual-dashboard",
		ManifestHash:      "manual",
		ManagedByManifest: false,
		ProtectionSource:  stringPtr("MANUAL"),
		Status:            "ACTIVE",
		LastSyncedAt:      time.Now(),
	}).Error)

	manifest := `manifest_version: "1.0"
fragment_role: main
app:
  app_code: ops-console
  app_name: 运营控制台
  app_type: BUSINESS_APP
  source: MANIFEST
  status: INITIATED
  deployment_mode: MERGED
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
clients:
  - PC_WEB
menus:
  - code: ops_dashboard
    name: 运营看板
    path: /ops/dashboard
permissions:
  - code: /ops/dashboard
    name: 运营看板
    type: MENU
    menu_code: ops_dashboard
package_features:
  - feature_code: ops_dashboard
    feature_name: 运营看板
    include_in_package: true
`
	service := NewAppService(repositories.NewAppRepository(db))
	diff, err := service.DiffManifest(context.Background(), 1, "app.manifest.yaml", "", []byte(manifest))

	require.NoError(t, err)
	require.False(t, diff.Loadable)
	require.GreaterOrEqual(t, diff.Summary.Conflict, 1)
	require.Contains(t, diff.Blockers, "资源有人工保护标记，Manifest 不允许覆盖")
}

func newAppCenterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.AppUser{},
		&models.SysApp{},
		&models.SysAppClient{},
		&models.SysAppManifestLoad{},
		&models.SysAppManifestFile{},
		&models.SysAppEntry{},
		&models.SysAppAPI{},
		&models.SysAppPermission{},
		&models.SysAppPackageFeature{},
		&models.SysAppQuota{},
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},
		&models.SaasFeature{},
		&models.SaasQuota{},
		&models.Tenant{},
		&models.DictType{},
		&models.DictItem{},
		&models.TenantSubscription{},
		&models.AuditLog{},
	))
	return db
}

func stringPtr(value string) *string {
	return &value
}
