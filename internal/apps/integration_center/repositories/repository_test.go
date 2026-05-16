package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestRepositoryListPlatformsHidesSoftDeletedRows(t *testing.T) {
	db := newRepositoryTestDB(t)
	now := time.Now()
	deletedAt := now
	require.NoError(t, db.Create(&models.IntegrationPlatform{
		PlatformCode:  "active-platform",
		PlatformName:  "Active Platform",
		PlatformType:  "open_api",
		AccessMode:    "api_key",
		Status:        "online",
		TenantVisible: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error)
	require.NoError(t, db.Create(&models.IntegrationPlatform{
		PlatformCode:  "deleted-platform",
		PlatformName:  "Deleted Platform",
		PlatformType:  "open_api",
		AccessMode:    "api_key",
		Status:        "online",
		TenantVisible: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     &deletedAt,
	}).Error)

	rows, total, err := NewRepository(db).ListPlatforms(context.Background(), ListOptions{Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	require.Equal(t, "active-platform", rows[0].Code)
}

func newRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.IntegrationPlatform{},
		&models.IntegrationPlatformCapability{},
		&models.IntegrationProviderApp{},
		&models.IntegrationTenantConnection{},
		&models.IntegrationAlert{},
		&models.IntegrationAPICallLog{},
	))
	return db
}
