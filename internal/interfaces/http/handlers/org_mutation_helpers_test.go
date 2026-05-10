package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestEnsureOrgNodeNotDuplicateRejectsSameParentCodeAndName(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OrgNode{}))

	parentID := uint64(5)
	now := time.Now()
	existingCode := "jishu"
	require.NoError(t, db.Create(&models.OrgNode{
		TenantID:  4,
		NodeType:  "department",
		ParentID:  &parentID,
		Name:      "技术部",
		Code:      &existingCode,
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error)

	duplicateCode := "JISHU"
	err = ensureOrgNodeNotDuplicate(db, models.OrgNode{
		TenantID: 4,
		NodeType: "department",
		ParentID: &parentID,
		Name:     "研发部",
		Code:     &duplicateCode,
	})
	require.EqualError(t, err, "同级组织下已存在相同编码")

	otherCode := "yanfa"
	err = ensureOrgNodeNotDuplicate(db, models.OrgNode{
		TenantID: 4,
		NodeType: "department",
		ParentID: &parentID,
		Name:     " 技术部 ",
		Code:     &otherCode,
	})
	require.EqualError(t, err, "同级组织下已存在相同名称")

	err = ensureOrgNodeNotDuplicate(db, models.OrgNode{
		TenantID: 4,
		NodeType: "department",
		ParentID: &parentID,
		Name:     "产品部",
		Code:     &otherCode,
	})
	require.NoError(t, err)
}

func TestEnsureOrgNodeNotDuplicateAllowsCurrentRowOnUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OrgNode{}))

	parentID := uint64(5)
	code := "jishu"
	row := models.OrgNode{
		TenantID: 4,
		NodeType: "department",
		ParentID: &parentID,
		Name:     "技术部",
		Code:     &code,
		Status:   1,
	}
	require.NoError(t, db.Create(&row).Error)

	require.NoError(t, ensureOrgNodeNotDuplicateExcluding(db, row, row.ID))
}

func TestOrgNodePayloadTracksExplicitZeroStatus(t *testing.T) {
	var body orgNodePayload

	require.NoError(t, json.Unmarshal([]byte(`{"name":"技术部","status":0}`), &body))
	require.True(t, body.StatusSet)
	require.Equal(t, 0, body.Status)

	body = orgNodePayload{}
	require.NoError(t, json.Unmarshal([]byte(`{"name":"技术部"}`), &body))
	require.False(t, body.StatusSet)
}
