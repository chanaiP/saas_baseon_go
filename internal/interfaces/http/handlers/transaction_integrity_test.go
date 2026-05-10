package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestUserCreateRollsBackWhenRelationWriteFails(t *testing.T) {
	db := newTransactionTestDB(t, &models.AppUser{}, &models.UserRole{})
	now := time.Now()
	user := models.AppUser{TenantID: 1, EmployeeNo: "E90001", Account: "E90001", PasswordHash: "hash", Name: "Rollback", Status: 1, CreatedAt: now, UpdatedAt: now}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return replaceUserRelationsTx(tx, user.ID, []uint64{1, 1}, nil, nil)
	})

	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&models.AppUser{}).Where("employee_no = ?", "E90001").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestRolePermissionRollbackKeepsOriginalAuthorization(t *testing.T) {
	db := newTransactionTestDB(t, &models.Role{}, &models.Permission{}, &models.RolePermission{})
	now := time.Now()
	role := models.Role{TenantID: 1, Code: "ops", Name: "Ops", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&models.RolePermission{RoleID: role.ID, PermissionID: 10, CreatedAt: now}).Error)
	require.NoError(t, db.Exec(`CREATE TRIGGER fail_role_permission_insert BEFORE INSERT ON role_permission WHEN NEW.permission_id = 20 BEGIN SELECT RAISE(FAIL, 'forced role permission failure'); END;`).Error)

	err := db.Transaction(func(tx *gorm.DB) error {
		return replaceRolePermissionsWithOverrides(tx, role.ID, []uint64{20}, nil)
	})

	require.Error(t, err)
	var rows []models.RolePermission
	require.NoError(t, db.Order("permission_id asc").Find(&rows, "role_id = ?", role.ID).Error)
	require.Len(t, rows, 1)
	require.Equal(t, uint64(10), rows[0].PermissionID)
}

func TestPlanCapabilityRollbackDoesNotHalfUpdate(t *testing.T) {
	db := newTransactionTestDB(t, &models.SaasPlan{}, &models.SaasFeature{}, &models.SaasQuota{}, &models.SaasPlanFeature{}, &models.SaasPlanQuota{})
	now := time.Now()
	plan := models.SaasPlan{PlanCode: "pro", PlanName: "Pro", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	feature := models.SaasFeature{FeatureCode: "user_manage", FeatureName: "用户管理", FeatureType: "MENU", Status: 1, CreatedAt: now, UpdatedAt: now}
	quota := models.SaasQuota{QuotaCode: "max_users", QuotaName: "用户数", QuotaType: "COUNT", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)
	require.NoError(t, db.Create(&feature).Error)
	require.NoError(t, db.Create(&quota).Error)
	require.NoError(t, db.Create(&models.SaasPlanFeature{PlanID: plan.ID, FeatureID: feature.ID, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: quota.ID, QuotaValue: 3, CreatedAt: now, UpdatedAt: now}).Error)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := savePlanFeaturesWithIDsTx(tx, plan.ID, []uint64{}); err != nil {
			return err
		}
		return savePlanQuotasWithValuesTx(tx, plan.ID, []planQuotaInput{{QuotaID: 9999, QuotaValue: 9}})
	})

	require.Error(t, err)
	var featureCount int64
	require.NoError(t, db.Model(&models.SaasPlanFeature{}).Where("plan_id = ?", plan.ID).Count(&featureCount).Error)
	require.Equal(t, int64(1), featureCount)
	var planQuota models.SaasPlanQuota
	require.NoError(t, db.Where("plan_id = ?", plan.ID).First(&planQuota).Error)
	require.Equal(t, 3, planQuota.QuotaValue)
}

func TestDeletePlanSoftDeletesReferencedPlan(t *testing.T) {
	db := newTransactionTestDB(t, &models.SaasPlan{}, &models.TenantSubscription{})
	now := time.Now()
	plan := models.SaasPlan{PlanCode: "basic", PlanName: "Basic", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)
	require.NoError(t, db.Create(&models.TenantSubscription{TenantID: 1, PlanID: plan.ID, SubscriptionStatus: "ACTIVE", StartTime: now, CreatedAt: now, UpdatedAt: now}).Error)

	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodDelete, "/api/plans/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := &IdentityHandler{db: db}
	handler.DeletePlan(c)

	require.Equal(t, http.StatusOK, w.Code)
	var row models.SaasPlan
	require.NoError(t, db.First(&row, plan.ID).Error)
	require.NotNil(t, row.DeletedAt)
	require.Equal(t, 0, row.Status)

	var subscription models.TenantSubscription
	require.NoError(t, db.Where("tenant_id = ?", uint64(1)).First(&subscription).Error)
	require.Equal(t, plan.ID, subscription.PlanID)
}

func TestTenantPackageRejectsDisabledOrDeletedPlan(t *testing.T) {
	db := newTransactionTestDB(t, &models.SaasPlan{}, &models.TenantSubscription{}, &models.TenantQuotaOverride{})
	now := time.Now()
	deletedAt := now.Add(-time.Minute)
	activePlan := models.SaasPlan{PlanCode: "active", PlanName: "Active", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	disabledPlan := models.SaasPlan{PlanCode: "disabled", PlanName: "Disabled", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 0, CreatedAt: now, UpdatedAt: now}
	deletedPlan := models.SaasPlan{PlanCode: "deleted", PlanName: "Deleted", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 0, DeletedAt: &deletedAt, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&activePlan).Error)
	require.NoError(t, db.Create(&disabledPlan).Error)
	require.NoError(t, db.Create(&deletedPlan).Error)
	require.NoError(t, db.Model(&disabledPlan).Update("status", 0).Error)
	require.NoError(t, db.Model(&deletedPlan).Updates(map[string]interface{}{"status": 0, "deleted_at": deletedAt}).Error)

	handler := &IdentityHandler{db: db}
	require.Error(t, handler.saveTenantPackageOnDB(db, 10, tenantPackagePayload{PlanID: disabledPlan.ID}))
	require.Error(t, handler.saveTenantPackageOnDB(db, 11, tenantPackagePayload{PlanID: deletedPlan.ID}))
	require.NoError(t, handler.saveTenantPackageOnDB(db, 12, tenantPackagePayload{PlanID: activePlan.ID}))

	var blockedCount int64
	require.NoError(t, db.Model(&models.TenantSubscription{}).Where("tenant_id IN ?", []uint64{10, 11}).Count(&blockedCount).Error)
	require.Equal(t, int64(0), blockedCount)
}

func newTransactionTestDB(t *testing.T, modelsToMigrate ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(modelsToMigrate...))
	return db
}
