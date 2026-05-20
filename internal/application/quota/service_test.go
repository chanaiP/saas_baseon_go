package quota

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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

func TestConsumeConcurrentDoesNotExceedLimit(t *testing.T) {
	db := newQuotaServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.TenantSubscription{}, &models.TenantQuotaUsage{}, &models.AuditLog{}))
	seedQuotaPlan(t, db, 1001, "daily_export_times", "每日导出次数", ptr("DAY"), 5)
	service := NewService(db)

	var successes int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := service.Consume(context.Background(), 1001, "daily_export_times", 1)
			if err == nil {
				atomic.AddInt32(&successes, 1)
				return
			}
			var quotaErr *ExceededError
			require.True(t, errors.As(err, &quotaErr), "unexpected error: %v", err)
		}()
	}
	wg.Wait()

	require.Equal(t, int32(5), successes)
	var usage models.TenantQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND quota_code = ?", 1001, "daily_export_times").First(&usage).Error)
	require.Equal(t, 5, usage.UsedValue)
	var auditCount int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("tenant_id = ? AND module = ? AND action = ?", 1001, "quota", "consume").Count(&auditCount).Error)
	require.Equal(t, int64(5), auditCount)
}

func TestConsumeUsesMonthlyPeriodKey(t *testing.T) {
	db := newQuotaServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.TenantSubscription{}, &models.TenantQuotaUsage{}, &models.AuditLog{}))
	seedQuotaPlan(t, db, 1002, "ai_geo_monthly_draft_generations", "月度母稿生成次数", ptr("MONTH"), 2)
	service := NewService(db)

	require.NoError(t, service.Consume(context.Background(), 1002, "ai_geo_monthly_draft_generations", 1))
	require.NoError(t, service.Consume(context.Background(), 1002, "ai_geo_monthly_draft_generations", 1))
	err := service.Consume(context.Background(), 1002, "ai_geo_monthly_draft_generations", 1)

	var quotaErr *ExceededError
	require.ErrorAs(t, err, &quotaErr)
	var usage models.TenantQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND quota_code = ?", 1002, "ai_geo_monthly_draft_generations").First(&usage).Error)
	require.Equal(t, time.Now().Format("200601"), usage.PeriodKey)
	require.Equal(t, 2, usage.UsedValue)
}

func TestCurrentLimitByCodeReturnsPlanLimit(t *testing.T) {
	db := newQuotaServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.TenantSubscription{}, &models.TenantQuotaUsage{}, &models.AuditLog{}))
	seedQuotaPlan(t, db, 1003, "ai_geo_brand_count", "品牌资料卡数量", ptr("NONE"), 3)
	service := NewService(db)

	limit, quota, ok, err := service.CurrentLimitByCode(context.Background(), 1003, "ai_geo_brand_count")

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 3, limit)
	require.Equal(t, "品牌资料卡数量", quota.QuotaName)
}

func newQuotaServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=busy_timeout(5000)", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.SaasPlan{}, &models.SaasFeature{}, &models.SaasQuota{}, &models.SaasPlanFeature{}, &models.SaasPlanQuota{}))
	return db
}

func seedQuotaPlan(t *testing.T, db *gorm.DB, tenantID uint64, quotaCode string, quotaName string, periodType *string, limit int) {
	t.Helper()
	now := time.Now()
	plan := models.SaasPlan{PlanCode: fmt.Sprintf("plan-%d", tenantID), PlanName: "测试套餐", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)
	quota := models.SaasQuota{QuotaCode: quotaCode, QuotaName: quotaName, QuotaType: "COUNT", PeriodType: periodType, Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&quota).Error)
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: quota.ID, QuotaValue: limit, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.TenantSubscription{TenantID: tenantID, PlanID: plan.ID, SubscriptionStatus: "ACTIVE", StartTime: now.Add(-time.Hour), CreatedAt: now, UpdatedAt: now}).Error)
}

func ptr(value string) *string {
	return &value
}
