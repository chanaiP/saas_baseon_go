package bootstrap

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.SystemParam{}); err != nil {
		return nil, err
	}

	return db, nil
}
