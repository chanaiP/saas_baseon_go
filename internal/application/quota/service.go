package quota

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type PlanQuotaInput struct {
	QuotaID    uint64
	QuotaValue int
}

func (s *Service) SaveFeatures(ctx context.Context, planID uint64, featureIDs []uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return saveFeaturesTx(tx, planID, featureIDs)
	})
}

func (s *Service) SaveQuotas(ctx context.Context, planID uint64, quotas []PlanQuotaInput) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return saveQuotasTx(tx, planID, quotas)
	})
}

func (s *Service) SaveCapabilities(ctx context.Context, planID uint64, featureIDs []uint64, quotas []PlanQuotaInput) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := saveFeaturesTx(tx, planID, featureIDs); err != nil {
			return err
		}
		return saveQuotasTx(tx, planID, quotas)
	})
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
