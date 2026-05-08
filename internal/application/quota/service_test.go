package quota

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestSaveCapabilitiesRollsBackWhenQuotaFails(t *testing.T) {
	db := newQuotaServiceTestDB(t)
	service := NewService(db)
	now := time.Now()
	plan := models.SaasPlan{PlanCode: "pro", PlanName: "Pro", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	feature := models.SaasFeature{FeatureCode: "user_manage", FeatureName: "用户管理", FeatureType: "MENU", Status: 1, CreatedAt: now, UpdatedAt: now}
	quota := models.SaasQuota{QuotaCode: "max_users", QuotaName: "用户数", QuotaType: "COUNT", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)
	require.NoError(t, db.Create(&feature).Error)
	require.NoError(t, db.Create(&quota).Error)
	require.NoError(t, db.Create(&models.SaasPlanFeature{PlanID: plan.ID, FeatureID: feature.ID, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: quota.ID, QuotaValue: 3, CreatedAt: now, UpdatedAt: now}).Error)

	err := service.SaveCapabilities(context.Background(), plan.ID, []uint64{}, []PlanQuotaInput{{QuotaID: 9999, QuotaValue: 9}})

	require.Error(t, err)
	var featureCount int64
	require.NoError(t, db.Model(&models.SaasPlanFeature{}).Where("plan_id = ?", plan.ID).Count(&featureCount).Error)
	require.Equal(t, int64(1), featureCount)
	var planQuota models.SaasPlanQuota
	require.NoError(t, db.Where("plan_id = ?", plan.ID).First(&planQuota).Error)
	require.Equal(t, 3, planQuota.QuotaValue)
}

func newQuotaServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.SaasPlan{}, &models.SaasFeature{}, &models.SaasQuota{}, &models.SaasPlanFeature{}, &models.SaasPlanQuota{}))
	return db
}
