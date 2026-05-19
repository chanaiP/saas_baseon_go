package repositories

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestBusinessUnitRepositoryTemplateScopeExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.BusinessUnitAttrTemplate{}))

	tenantID := uint64(1)
	groupCode := "douyin"
	existing := models.BusinessUnitAttrTemplate{
		TenantID:      &tenantID,
		TemplateName:  "门店通用属性",
		UnitTypeCode:  "store",
		UnitTypeName:  "门店",
		UnitGroupCode: &groupCode,
		UnitGroupName: ptrString("抖音"),
		Status:        "active",
	}
	require.NoError(t, db.Create(&existing).Error)

	repo := NewBusinessUnitRepository(db)
	exists, err := repo.TemplateScopeExists(context.Background(), tenantID, " 门店通用属性 ", " store ", " douyin ", 0)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.TemplateScopeExists(context.Background(), tenantID, "门店通用属性", "store", "douyin", existing.ID)
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = repo.TemplateScopeExists(context.Background(), tenantID, "门店通用属性", "store", "", 0)
	require.NoError(t, err)
	require.False(t, exists)
}

func ptrString(v string) *string {
	return &v
}
