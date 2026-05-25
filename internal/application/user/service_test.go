package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type denyQuotaChecker struct{}

func (denyQuotaChecker) RequireAvailable(context.Context, uint64, string, int) error {
	return fmt.Errorf("配额不足")
}

func (denyQuotaChecker) Consume(context.Context, uint64, string, int) error {
	return nil
}

type countingQuotaChecker struct {
	consumeCount int
}

func (c *countingQuotaChecker) RequireAvailable(context.Context, uint64, string, int) error {
	return nil
}

func (c *countingQuotaChecker) Consume(context.Context, uint64, string, int) error {
	c.consumeCount++
	return nil
}

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

func TestCreateWithRelationsChecksQuotaInsideService(t *testing.T) {
	db := newUserServiceTestDB(t)
	service := NewService(db, denyQuotaChecker{})
	now := time.Now()
	row := models.AppUser{TenantID: 1, EmployeeNo: "E70003", Account: "E70003", PasswordHash: "hash", Name: "Quota", Status: 1, CreatedAt: now, UpdatedAt: now}

	_, err := service.CreateWithRelations(context.Background(), row, Relations{})

	require.Error(t, err)
	require.Contains(t, err.Error(), "配额")
	var count int64
	require.NoError(t, db.Model(&models.AppUser{}).Where("employee_no = ?", "E70003").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestCreateWithRelationsAllowsAnyTenantOrgNode(t *testing.T) {
	db := newUserServiceTestDB(t)
	service := NewService(db)
	now := time.Now()
	org := models.OrgNode{TenantID: 1, NodeType: "company", Name: "总部", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&org).Error)
	row := models.AppUser{TenantID: 1, EmployeeNo: "E70005", Account: "E70005", PasswordHash: "hash", Name: "OrgNode", Status: 1, CreatedAt: now, UpdatedAt: now}

	created, err := service.CreateWithRelations(context.Background(), row, Relations{DepartmentIDs: []uint64{org.ID}})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	var count int64
	require.NoError(t, db.Model(&models.AppUserDepartment{}).Where("user_id = ? AND department_id = ?", created.ID, org.ID).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestCreateWithRelationsRejectsDisabledOrgNode(t *testing.T) {
	db := newUserServiceTestDB(t)
	service := NewService(db)
	now := time.Now()
	org := models.OrgNode{TenantID: 1, NodeType: "company", Name: "已停用组织", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&org).Error)
	require.NoError(t, db.Model(&models.OrgNode{}).Where("id = ?", org.ID).Update("status", 0).Error)
	row := models.AppUser{TenantID: 1, EmployeeNo: "E70006", Account: "E70006", PasswordHash: "hash", Name: "DisabledOrg", Status: 1, CreatedAt: now, UpdatedAt: now}

	_, err := service.CreateWithRelations(context.Background(), row, Relations{DepartmentIDs: []uint64{org.ID}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "已停用")
	var count int64
	require.NoError(t, db.Model(&models.AppUser{}).Where("employee_no = ?", "E70006").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestImportUsersRollsBackBatchWhenOneRowFails(t *testing.T) {
	db := newUserServiceTestDB(t)
	quota := &countingQuotaChecker{}
	service := NewService(db, quota)
	now := time.Now()
	rows := []models.AppUser{
		{TenantID: 1, EmployeeNo: "E70002", Account: "E70002", PasswordHash: "hash", Name: "One", Status: 1, CreatedAt: now, UpdatedAt: now},
		{TenantID: 1, EmployeeNo: "E70002", Account: "E70002", PasswordHash: "hash", Name: "Duplicate", Status: 1, CreatedAt: now, UpdatedAt: now},
	}

	_, err := service.ImportUsers(context.Background(), rows)

	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&models.AppUser{}).Where("employee_no = ?", "E70002").Count(&count).Error)
	require.Equal(t, int64(0), count)
	require.Equal(t, 0, quota.consumeCount)
}

func TestImportUsersConsumesDailyQuotaOnlyAfterSuccessfulCreate(t *testing.T) {
	db := newUserServiceTestDB(t)
	quota := &countingQuotaChecker{}
	service := NewService(db, quota)
	now := time.Now()
	rows := []models.AppUser{
		{TenantID: 1, EmployeeNo: "E70004", Account: "E70004", PasswordHash: "hash", Name: "One", Status: 1, CreatedAt: now, UpdatedAt: now},
	}

	result, err := service.ImportUsers(context.Background(), rows)

	require.NoError(t, err)
	require.Equal(t, 1, result.Created)
	require.Equal(t, 1, quota.consumeCount)
}

func newUserServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.AppUser{}, &models.OrgNode{}, &models.Role{}, &models.UserRole{}, &models.AppUserPosition{}, &models.AppUserDepartment{}))
	return db
}
