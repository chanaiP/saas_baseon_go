package file

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestCreateMetadataPersistsTenantScopedFile(t *testing.T) {
	db := newFileServiceTestDB(t)
	service := NewService(db)
	now := time.Now()

	row, err := service.CreateMetadata(context.Background(), CreateMetadataCommand{
		TenantID: 1, FileID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CreatedBy: 7,
		OriginalName: "report.pdf", StoredName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.pdf",
		StoragePath: "/tmp/report.pdf", MimeType: "application/pdf", FileSize: 12, Now: now,
	})

	require.NoError(t, err)
	require.NotZero(t, row.ID)
	require.Equal(t, uint64(1), row.TenantID)
	require.Equal(t, "application/pdf", row.MimeType)
}

func TestSoftDeleteMarksFileDeleted(t *testing.T) {
	db := newFileServiceTestDB(t)
	service := NewService(db)
	row, err := service.CreateMetadata(context.Background(), CreateMetadataCommand{
		TenantID: 1, FileID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", CreatedBy: 7,
		OriginalName: "report.pdf", StoredName: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.pdf",
		StoragePath: "/tmp/report.pdf", MimeType: "application/pdf", FileSize: 12,
	})
	require.NoError(t, err)

	require.NoError(t, service.SoftDelete(context.Background(), &row, "/tmp/.trash/report.pdf", time.Now()))

	var stored models.FileObject
	require.NoError(t, db.First(&stored, row.ID).Error)
	require.Equal(t, 0, stored.Status)
	require.NotNil(t, stored.DeletedAt)
	require.Equal(t, "/tmp/.trash/report.pdf", stored.StoragePath)
}

func newFileServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.FileObject{}))
	return db
}
