package user

import (
	"context"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type Relations struct {
	RoleIDs       []uint64
	PositionIDs   []uint64
	DepartmentIDs []uint64
}

func (s *Service) CreateWithRelations(ctx context.Context, row models.AppUser, rel Relations) (models.AppUser, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return replaceRelations(tx, row.ID, rel)
	})
	return row, err
}

func (s *Service) UpdateWithRelations(ctx context.Context, row *models.AppUser, updates map[string]interface{}, rel Relations) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(row).Updates(updates).First(row, row.ID).Error; err != nil {
				return err
			}
		}
		return replaceRelations(tx, row.ID, rel)
	})
}

func (s *Service) ImportUsers(ctx context.Context, rows []models.AppUser) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range rows {
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) Delete(ctx context.Context, row *models.AppUser, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(row).Updates(updates).Error
	})
}

func replaceRelations(tx *gorm.DB, userID uint64, rel Relations) error {
	if rel.RoleIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		for _, id := range rel.RoleIDs {
			if err := tx.Create(&models.UserRole{UserID: userID, RoleID: id}).Error; err != nil {
				return err
			}
		}
	}
	if rel.PositionIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.AppUserPosition{}).Error; err != nil {
			return err
		}
		for _, id := range rel.PositionIDs {
			if err := tx.Create(&models.AppUserPosition{UserID: userID, PositionID: id}).Error; err != nil {
				return err
			}
		}
	}
	if rel.DepartmentIDs != nil {
		if err := tx.Where("user_id = ?", userID).Delete(&models.AppUserDepartment{}).Error; err != nil {
			return err
		}
		for _, id := range rel.DepartmentIDs {
			if err := tx.Create(&models.AppUserDepartment{UserID: userID, DepartmentID: id}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
