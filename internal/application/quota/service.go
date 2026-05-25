package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
)

type Service struct {
	db               *gorm.DB
	cacheInvalidator CacheInvalidator
}

type CacheInvalidator interface {
	InvalidateAllAuthorizationCache()
}

func NewService(db *gorm.DB, cacheInvalidator ...CacheInvalidator) *Service {
	service := &Service{db: db}
	if len(cacheInvalidator) > 0 {
		service.cacheInvalidator = cacheInvalidator[0]
	}
	return service
}

type PlanQuotaInput struct {
	QuotaID    uint64
	QuotaValue int
}

func (s *Service) SaveFeatures(ctx context.Context, planID uint64, featureIDs []uint64) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return saveFeaturesTx(tx, planID, featureIDs)
	})
	s.invalidateCacheAfter(err)
	return err
}

func (s *Service) SaveQuotas(ctx context.Context, planID uint64, quotas []PlanQuotaInput) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return saveQuotasTx(tx, planID, quotas)
	})
	s.invalidateCacheAfter(err)
	return err
}

func (s *Service) SaveCapabilities(ctx context.Context, planID uint64, featureIDs []uint64, quotas []PlanQuotaInput) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := saveFeaturesTx(tx, planID, featureIDs); err != nil {
			return err
		}
		return saveQuotasTx(tx, planID, quotas)
	})
	s.invalidateCacheAfter(err)
	return err
}

func (s *Service) Consume(ctx context.Context, tenantID uint64, quotaCode string, increment int) error {
	if increment <= 0 {
		increment = 1
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.consumeWithDB(ctx, tx, tenantID, quotaCode, increment)
	})
}

func (s *Service) ConsumeWithDB(ctx context.Context, db *gorm.DB, tenantID uint64, quotaCode string, increment int) error {
	if increment <= 0 {
		increment = 1
	}
	return s.consumeWithDB(ctx, db, tenantID, quotaCode, increment)
}

func (s *Service) consumeWithDB(ctx context.Context, db *gorm.DB, tenantID uint64, quotaCode string, increment int) error {
	tx := db.WithContext(ctx)
	var quota models.SaasQuota
	if err := tx.Where("quota_code = ? AND status = ?", quotaCode, 1).First(&quota).Error; err != nil {
		return nil
	}
	now := time.Now()
	periodKey := quotaPeriodKey(quota.PeriodType, now)
	limit := currentQuotaLimit(tx, tenantID, quota.ID)
	if limit >= 0 && increment > limit {
		return &ExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: 0}
	}
	usage := models.TenantQuotaUsage{
		TenantID:        tenantID,
		QuotaCode:       quotaCode,
		UsedValue:       increment,
		LimitValue:      limit,
		PeriodType:      quota.PeriodType,
		PeriodKey:       periodKey,
		LastRefreshTime: &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	result := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_id"}, {Name: "quota_code"}, {Name: "period_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"used_value":        gorm.Expr("tenant_quota_usage.used_value + ?", increment),
			"limit_value":       limit,
			"last_refresh_time": now,
			"updated_at":        now,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			gorm.Expr("? < 0 OR tenant_quota_usage.used_value + ? <= ?", limit, increment, limit),
		}},
	}).Create(&usage)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &ExceededError{QuotaName: quota.QuotaName, Limit: limit, Used: currentQuotaUsage(tx, tenantID, quotaCode, periodKey)}
	}
	return recordAudit(tx, tenantID, quotaCode, increment, limit, periodKey, now)
}

func (s *Service) CurrentLimitByCode(ctx context.Context, tenantID uint64, quotaCode string) (int, models.SaasQuota, bool, error) {
	var quota models.SaasQuota
	if err := s.db.WithContext(ctx).Where("quota_code = ? AND status = ?", quotaCode, 1).First(&quota).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, quota, false, nil
		}
		return 0, quota, false, err
	}
	return currentQuotaLimit(s.db.WithContext(ctx), tenantID, quota.ID), quota, true, nil
}

type ExceededError struct {
	QuotaName string
	Limit     int
	Used      int
}

func (e *ExceededError) Error() string {
	return e.QuotaName + "已超出套餐配额"
}

func (s *Service) invalidateCacheAfter(err error) {
	if err == nil && s.cacheInvalidator != nil {
		s.cacheInvalidator.InvalidateAllAuthorizationCache()
	}
}

func saveFeaturesTx(tx *gorm.DB, planID uint64, featureIDs []uint64) error {
	if err := lockPlan(tx, planID); err != nil {
		return err
	}
	if err := tx.Where("plan_id = ?", planID).Delete(&models.SaasPlanFeature{}).Error; err != nil {
		return err
	}
	for _, featureID := range uniqueIDs(featureIDs) {
		var feature models.SaasFeature
		if err := tx.Where("id = ? AND status = ?", featureID, 1).First(&feature).Error; err != nil {
			return fmt.Errorf("功能不存在或已停用")
		}
		if err := tx.Create(&models.SaasPlanFeature{PlanID: planID, FeatureID: featureID, Enabled: true}).Error; err != nil {
			return err
		}
	}
	return nil
}

func saveQuotasTx(tx *gorm.DB, planID uint64, quotas []PlanQuotaInput) error {
	if err := lockPlan(tx, planID); err != nil {
		return err
	}
	if err := tx.Where("plan_id = ?", planID).Delete(&models.SaasPlanQuota{}).Error; err != nil {
		return err
	}
	seen := map[uint64]struct{}{}
	for _, item := range quotas {
		if item.QuotaID == 0 {
			continue
		}
		if _, ok := seen[item.QuotaID]; ok {
			continue
		}
		seen[item.QuotaID] = struct{}{}
		var row models.SaasQuota
		if err := tx.Where("id = ? AND status = ?", item.QuotaID, 1).First(&row).Error; err != nil {
			return fmt.Errorf("配额不存在或已停用")
		}
		if err := tx.Create(&models.SaasPlanQuota{PlanID: planID, QuotaID: item.QuotaID, QuotaValue: item.QuotaValue}).Error; err != nil {
			return err
		}
	}
	return nil
}

func lockPlan(tx *gorm.DB, planID uint64) error {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&models.SaasPlan{}, planID).Error
}

func currentQuotaLimit(db *gorm.DB, tenantID uint64, quotaID uint64) int {
	now := time.Now()
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err == nil {
		if !subscriptionAllows(sub.SubscriptionStatus, sub.EndTime, now) {
			return 0
		}
	}
	var override models.TenantQuotaOverride
	if err := db.Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quotaID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue
	}
	if sub.ID > 0 {
		var planQuota models.SaasPlanQuota
		if err := db.Where("plan_id = ? AND quota_id = ?", sub.PlanID, quotaID).First(&planQuota).Error; err == nil {
			return planQuota.QuotaValue
		}
	}
	return 0
}

func subscriptionAllows(statusRaw string, endTime *time.Time, now time.Time) bool {
	switch statusRaw {
	case "OVERDUE", "FROZEN", "EXPIRED", "CANCELLED":
		return false
	}
	return endTime == nil || endTime.After(now)
}

func currentQuotaUsage(db *gorm.DB, tenantID uint64, quotaCode string, periodKey string) int {
	var usage models.TenantQuotaUsage
	if err := repositories.NewTenantScopedRepository(db, tenantID).QuotaUsage(context.Background()).Where("quota_code = ? AND period_key = ?", quotaCode, periodKey).First(&usage).Error; err == nil {
		return usage.UsedValue
	}
	return 0
}

func quotaPeriodKey(periodType *string, now time.Time) string {
	if periodType == nil {
		return "TOTAL"
	}
	switch strings.ToUpper(strings.TrimSpace(*periodType)) {
	case "DAY", "DAILY":
		return now.Format("20060102")
	case "MONTH", "MONTHLY":
		return now.Format("200601")
	case "YEAR", "YEARLY":
		return now.Format("2006")
	default:
		return "TOTAL"
	}
}

func recordAudit(db *gorm.DB, tenantID uint64, quotaCode string, increment int, limit int, periodKey string, now time.Time) error {
	detail, _ := json.Marshal(map[string]interface{}{
		"quota_code": quotaCode,
		"increment":  increment,
		"limit":      limit,
		"period_key": periodKey,
	})
	text := string(detail)
	return db.Create(&models.AuditLog{
		TenantID:  &tenantID,
		Module:    "quota",
		Action:    "consume",
		Summary:   "配额扣减",
		Detail:    &text,
		Result:    "success",
		CreatedAt: now,
	}).Error
}

func uniqueIDs(values []uint64) []uint64 {
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
