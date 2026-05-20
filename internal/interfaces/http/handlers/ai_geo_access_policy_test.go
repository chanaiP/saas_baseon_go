package handlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestAiGeoRouteAccessReflectsPackageRoleAndPlatformScope(t *testing.T) {
	db := newAiGeoPolicyTestDB(t)
	h := &IdentityHandler{db: db}
	allowedUser := seedAiGeoAccessUser(t, db, 1, 10, true, true)
	noPackageUser := seedAiGeoAccessUser(t, db, 2, 20, false, true)
	noRoleUser := seedAiGeoAccessUser(t, db, 3, 30, true, false)
	platformAdmin := seedAiGeoAccessUser(t, db, 4, 40, true, false)
	platformAdmin.IsPlatformAdmin = true
	require.NoError(t, db.Model(&models.AppUser{}).Where("id = ?", platformAdmin.ID).Update("is_platform_admin", true).Error)

	require.True(t, h.routeAllowed(allowedUser, "GET", "/api/ai-geo/publish-plans/calendar"))
	require.True(t, h.routeAllowed(allowedUser, "POST", "/api/ai-geo/publish-plans"))
	require.False(t, h.routeAllowed(noPackageUser, "POST", "/api/ai-geo/publish-plans"))
	require.False(t, h.routeAllowed(noRoleUser, "GET", "/api/ai-geo/publish-plans/calendar"))
	require.False(t, h.routeAllowed(noRoleUser, "POST", "/api/ai-geo/publish-plans"))
	require.True(t, h.routeAllowed(platformAdmin, "GET", "/api/ai-geo/publish-plans/calendar"))
	require.True(t, h.routeAllowed(platformAdmin, "POST", "/api/ai-geo/publish-plans"))

	var feature models.SaasFeature
	require.NoError(t, db.Where("feature_code = ?", "ai_geo_plan_queue").First(&feature).Error)
	require.NoError(t, db.Create(&models.TenantFeatureOverride{TenantID: allowedUser.TenantID, FeatureID: feature.ID, Enabled: false, Source: "TEST", CreatedAt: time.Now(), UpdatedAt: time.Now()}).Error)
	require.False(t, h.routeAllowed(allowedUser, "POST", "/api/ai-geo/publish-plans"))
}

func newAiGeoPolicyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(
		&models.Tenant{},
		&models.AppUser{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.SaasPlan{},
		&models.SaasFeature{},
		&models.SaasPlanFeature{},
		&models.TenantSubscription{},
		&models.TenantFeatureOverride{},
	))
	return db
}

func seedAiGeoAccessUser(t *testing.T, db *gorm.DB, tenantID uint64, userID uint64, withPackage bool, withRole bool) models.AppUser {
	t.Helper()
	now := time.Now()
	require.NoError(t, db.Create(&models.Tenant{ID: tenantID, Code: fmt.Sprintf("tenant-%d", tenantID), Name: "AI GEO 租户", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.AppUser{ID: userID - 1, TenantID: tenantID, EmployeeNo: fmt.Sprintf("admin-%d", tenantID), Account: fmt.Sprintf("admin-%d", tenantID), PasswordHash: "x", Name: "租户管理员", Status: 1, IsTenantAdmin: true, CreatedAt: now, UpdatedAt: now}).Error)
	user := models.AppUser{ID: userID, TenantID: tenantID, EmployeeNo: fmt.Sprintf("u-%d", userID), Account: fmt.Sprintf("u-%d", userID), PasswordHash: "x", Name: "测试用户", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&user).Error)
	calendarFeature := models.SaasFeature{FeatureCode: "ai_geo_plan_calendar", FeatureName: "发布日历", FeatureType: "MENU", AppCode: "ai-geo", Status: 1, CreatedAt: now, UpdatedAt: now}
	queueFeature := models.SaasFeature{FeatureCode: "ai_geo_plan_queue", FeatureName: "发布队列", FeatureType: "MENU", AppCode: "ai-geo", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.FirstOrCreate(&calendarFeature, models.SaasFeature{FeatureCode: calendarFeature.FeatureCode}).Error)
	require.NoError(t, db.FirstOrCreate(&queueFeature, models.SaasFeature{FeatureCode: queueFeature.FeatureCode}).Error)
	calendarPath := "/ai-geo/plans/calendar"
	queueCode := "ai_geo:publish_plan:manage"
	calendarPermission := models.Permission{TenantID: tenantID, Name: "发布日历", Path: calendarPath, PermType: 3, Enabled: true, IsPackageFeature: true, FeatureCode: &calendarFeature.FeatureCode, FeatureType: &calendarFeature.FeatureType, AppCode: "ai-geo", TenantScope: "enterprise_only", CreatedAt: now, UpdatedAt: now}
	managePermission := models.Permission{TenantID: tenantID, Name: "发布计划管理", Path: queueCode, PermType: 2, Enabled: true, IsPackageFeature: true, FeatureCode: &queueFeature.FeatureCode, FeatureType: &queueFeature.FeatureType, AppCode: "ai-geo", TenantScope: "enterprise_only", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&calendarPermission).Error)
	require.NoError(t, db.Create(&managePermission).Error)
	plan := models.SaasPlan{PlanCode: fmt.Sprintf("ai-geo-plan-%d", tenantID), PlanName: "AI GEO 套餐", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)
	require.NoError(t, db.Create(&models.TenantSubscription{TenantID: tenantID, PlanID: plan.ID, SubscriptionStatus: "ACTIVE", StartTime: now.Add(-time.Hour), CreatedAt: now, UpdatedAt: now}).Error)
	if withPackage {
		require.NoError(t, db.Create(&models.SaasPlanFeature{PlanID: plan.ID, FeatureID: calendarFeature.ID, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
		require.NoError(t, db.Create(&models.SaasPlanFeature{PlanID: plan.ID, FeatureID: queueFeature.ID, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	}
	role := models.Role{TenantID: tenantID, Code: fmt.Sprintf("ai-geo-role-%d", tenantID), Name: "AI GEO 角色", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&role).Error)
	if withRole {
		require.NoError(t, db.Create(&models.UserRole{UserID: userID, RoleID: role.ID, CreatedAt: now}).Error)
		require.NoError(t, db.Create(&models.RolePermission{RoleID: role.ID, PermissionID: calendarPermission.ID, CreatedAt: now}).Error)
		require.NoError(t, db.Create(&models.RolePermission{RoleID: role.ID, PermissionID: managePermission.ID, CreatedAt: now}).Error)
	}
	return user
}
