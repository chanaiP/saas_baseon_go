package handlers

import (
	"errors"
	"testing"
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSysParamBatchKeysTrimEmptyAndDeduplicate(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, splitCSVParam(" a, ,b,, c,a "))
	require.Empty(t, splitCSVParam(" , , "))
}

func TestDictItemJSONAppliesTenantOverride(t *testing.T) {
	label := "标准客户"
	sortOrder := 1
	enabled := false
	item := dictItemJSON(models.DictItem{
		ID:         7,
		DictTypeID: 3,
		Label:      "普通客户",
		Value:      "C",
		SortOrder:  3,
		Enabled:    true,
	}, &models.TenantDictItemOverride{
		CustomLabel: &label,
		SortOrder:   &sortOrder,
		Enabled:     &enabled,
	})

	require.Equal(t, "标准客户", item["label"])
	require.Equal(t, "C", item["value"])
	require.Equal(t, 1, item["sort_order"])
	require.Equal(t, false, item["enabled"])
	require.Equal(t, true, item["is_override"])
}

func TestVisibleDictTypeByCodeFallsBackToPlatformSharedDict(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Tenant{}, &models.DictType{}))
	now := time.Now()
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", Status: 1, IsPlatform: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 4, Code: "taohuadao", Name: "桃花岛", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DictType{TenantID: 1, Code: "org_node_type", Name: "组织节点类型", Scope: "platform", TenantEditable: true, IsPlatformOnly: false, CreatedAt: now, UpdatedAt: now}).Error)

	row, err := (&IdentityHandler{db: db}).visibleDictTypeByCode(4, "org_node_type", false)

	require.NoError(t, err)
	require.Equal(t, uint64(1), row.TenantID)
	require.Equal(t, "组织节点类型", row.Name)
}

func TestVisibleSystemParamsQueryFallsBackToPlatformSharedParams(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Tenant{}, &models.SystemParam{}))
	now := time.Now()
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", Status: 1, IsPlatform: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 4, Code: "taohuadao", Name: "桃花岛", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.SystemParam{TenantID: 1, Key: "org.default_company_type", Value: "SUBSIDIARY", ValueType: "string", TenantEditable: true, IsPlatformOnly: false, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.SystemParam{TenantID: 1, Key: "security.internal_only", Value: "true", ValueType: "boolean", TenantEditable: false, IsPlatformOnly: true, CreatedAt: now, UpdatedAt: now}).Error)

	var rows []models.SystemParam
	require.NoError(t, (&IdentityHandler{db: db}).visibleSystemParamsQuery(4, false).Order("id asc").Find(&rows).Error)

	require.Len(t, rows, 1)
	require.Equal(t, "org.default_company_type", rows[0].Key)
	require.Equal(t, uint64(1), rows[0].TenantID)
}

func TestSafeDBErrorMessageHidesConstraintDetails(t *testing.T) {
	msg := safeDBErrorMessage(errors.New("duplicate key value violates unique constraint uq_sys_param_key"))

	require.NotContains(t, msg, "duplicate key")
	require.NotContains(t, msg, "unique constraint")
	require.Equal(t, "数据已存在，请检查唯一字段", msg)
}
