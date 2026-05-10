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

func TestTenantScopedRepositoryFiltersUsersByTenant(t *testing.T) {
	db := newTenantScopedRepoTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.AppUser{TenantID: 1, EmployeeNo: "E1", Account: "E1", PasswordHash: "hash", Name: "One", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.AppUser{TenantID: 2, EmployeeNo: "E2", Account: "E2", PasswordHash: "hash", Name: "Two", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)

	var rows []models.AppUser
	require.NoError(t, NewTenantScopedRepository(db, 1).ActiveUsers(context.Background()).Find(&rows).Error)

	require.Len(t, rows, 1)
	require.Equal(t, uint64(1), rows[0].TenantID)
}

func TestTenantScopedRepositoryFiltersFilesByTenant(t *testing.T) {
	db := newTenantScopedRepoTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.FileObject{TenantID: 1, FileID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CreatedBy: 1, OriginalName: "a.txt", StoredName: "a.txt", StoragePath: "/tmp/a.txt", MimeType: "text/plain", FileSize: 1, Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.FileObject{TenantID: 2, FileID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", CreatedBy: 2, OriginalName: "b.txt", StoredName: "b.txt", StoragePath: "/tmp/b.txt", MimeType: "text/plain", FileSize: 1, Status: 1, CreatedAt: now, UpdatedAt: now}).Error)

	var rows []models.FileObject
	require.NoError(t, NewTenantScopedRepository(db, 1).ActiveFiles(context.Background()).Find(&rows).Error)

	require.Len(t, rows, 1)
	require.Equal(t, uint64(1), rows[0].TenantID)
}

func newTenantScopedRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.AppUser{}, &models.FileObject{}, &models.TenantQuotaUsage{}))
	return db
}
