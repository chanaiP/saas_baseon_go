package handlers

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestValidatePositionTypePayloadRejectsDuplicateNameAndCode(t *testing.T) {
	db := newPositionValidationDB(t)
	require.NoError(t, db.Create(&models.PositionType{TenantID: 1, Name: "产品岗", Code: "product"}).Error)
	h := &IdentityHandler{db: db}

	require.Equal(t, "岗位类型名称已存在", validatePositionTypePayload(h, 1, 0, "产品岗", "other"))
	require.Equal(t, "岗位类型编码已存在", validatePositionTypePayload(h, 1, 0, "技术岗", "product"))
	require.Empty(t, validatePositionTypePayload(h, 1, 0, "技术岗", "tech"))
}

func TestValidatePositionPayloadRejectsDuplicateNameAndCode(t *testing.T) {
	db := newPositionValidationDB(t)
	require.NoError(t, db.Create(&models.PositionType{ID: 10, TenantID: 1, Name: "产品岗", Code: "product"}).Error)
	require.NoError(t, db.Create(&models.Position{TenantID: 1, PositionTypeID: 10, Name: "产品专员", Code: "cpjl"}).Error)
	h := &IdentityHandler{db: db}

	require.Equal(t, "岗位名称已存在", validatePositionPayload(h, 1, 0, "产品专员", "other"))
	require.Equal(t, "岗位编码已存在", validatePositionPayload(h, 1, 0, "产品经理", "cpjl"))
	require.Empty(t, validatePositionPayload(h, 1, 0, "产品经理", "cpj"))
}

func newPositionValidationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.PositionType{}, &models.Position{}))
	return db
}
