package user

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestCreateWithRelationsRollsBackUserWhenRelationFails(t *testing.T) {
	db := newUserServiceTestDB(t)
	service := NewService(db)
	now := time.Now()
	row := models.AppUser{TenantID: 1, EmployeeNo: "E70001", Account: "E70001", PasswordHash: "hash", Name: "Rollback", Status: 1, CreatedAt: now, UpdatedAt: now}

	_, err := service.CreateWithRelations(context.Background(), row, Relations{RoleIDs: []uint64{1, 1}})

	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&models.AppUser{}).Where("employee_no = ?", "E70001").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestImportUsersRollsBackBatchWhenOneRowFails(t *testing.T) {
	db := newUserServiceTestDB(t)
	service := NewService(db)
	now := time.Now()
	rows := []models.AppUser{
		{TenantID: 1, EmployeeNo: "E70002", Account: "E70002", PasswordHash: "hash", Name: "One", Status: 1, CreatedAt: now, UpdatedAt: now},
		{TenantID: 1, EmployeeNo: "E70002", Account: "E70002", PasswordHash: "hash", Name: "Duplicate", Status: 1, CreatedAt: now, UpdatedAt: now},
	}

	err := service.ImportUsers(context.Background(), rows)

	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&models.AppUser{}).Where("employee_no = ?", "E70002").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func newUserServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.AppUser{}, &models.UserRole{}, &models.AppUserPosition{}, &models.AppUserDepartment{}))
	return db
}
