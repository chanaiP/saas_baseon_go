package handlers

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	appquota "saas_baseon_go/internal/application/quota"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestConsumeQuotaConcurrentDoesNotExceedLimit(t *testing.T) {
	db := newQuotaTestDB(t)
	seedQuotaPlan(t, db, 1001, "daily_export_times", "每日导出次数", ptr("DAY"), 5)
	handler := &IdentityHandler{db: db}

	var successes int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := handler.consumeQuota(1001, "daily_export_times", 1)
			if err == nil {
				atomic.AddInt32(&successes, 1)
				return
			}
			var quotaErr *appquota.ExceededError
			require.True(t, errors.As(err, &quotaErr), "unexpected error: %v", err)
		}()
	}
	wg.Wait()

	require.Equal(t, int32(5), successes)
	var usage models.TenantQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND quota_code = ?", 1001, "daily_export_times").First(&usage).Error)
	require.Equal(t, 5, usage.UsedValue)
	require.Equal(t, 5, usage.LimitValue)
	var auditCount int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("tenant_id = ? AND module = ? AND action = ?", 1001, "quota", "consume").Count(&auditCount).Error)
	require.Equal(t, int64(5), auditCount)
}

func TestConsumeQuotaExceededDoesNotIncreaseUsage(t *testing.T) {
	db := newQuotaTestDB(t)
	seedQuotaPlan(t, db, 1002, "daily_import_times", "每日导入次数", ptr("DAY"), 2)
	handler := &IdentityHandler{db: db}

	require.NoError(t, handler.consumeQuota(1002, "daily_import_times", 2))
	err := handler.consumeQuota(1002, "daily_import_times", 1)

	var quotaErr *appquota.ExceededError
	require.True(t, errors.As(err, &quotaErr), "unexpected error: %v", err)
	var usage models.TenantQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND quota_code = ?", 1002, "daily_import_times").First(&usage).Error)
	require.Equal(t, 2, usage.UsedValue)
}

func newQuotaTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=busy_timeout(5000)", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(
		&models.SaasQuota{},
		&models.SaasPlan{},
		&models.SaasPlanQuota{},
		&models.TenantSubscription{},
		&models.TenantQuotaUsage{},
		&models.AuditLog{},
	))
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
