package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
)

func TestPlanMatrixKeepsMenuOperationsNestedByParentID(t *testing.T) {
	db := newTransactionTestDB(t, &models.SaasPlan{}, &models.SaasFeature{}, &models.SaasQuota{}, &models.SaasPlanQuota{})
	now := time.Now()
	plan := models.SaasPlan{PlanCode: "BASIC", PlanName: "基础版", PlanType: "BASIC", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)

	menu := models.SaasFeature{FeatureCode: "user_manage", FeatureName: "用户管理", FeatureType: "MENU", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&menu).Error)
	op := models.SaasFeature{FeatureCode: "button_custom_user_action", FeatureName: "用户-自定义操作", FeatureType: "OPERATION", ParentID: menu.ID, Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&op).Error)

	handler := &IdentityHandler{db: db}
	nodes := handler.buildPlanMatrixNodes([]models.SaasPlan{plan}, []models.SaasFeature{menu, op}, map[uint64]map[uint64]bool{})

	userMenu := findPlanMatrixNode(nodes, "user_manage")
	require.NotNil(t, userMenu)
	require.Len(t, userMenu.Children, 1)
	require.Equal(t, "button_custom_user_action", userMenu.Children[0].ID)

	other := findPlanMatrixNode(nodes, "domain-other")
	require.Nil(t, other, "menu-generated operation must not be rendered as a separate top-level capability")
}

func TestRepairPackageFeatureParentIDsBackfillsSeededOperations(t *testing.T) {
	db := newTransactionTestDB(t, &models.SaasFeature{})
	now := time.Now()
	menu := models.SaasFeature{FeatureCode: "role_manage", FeatureName: "角色权限", FeatureType: "MENU", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&menu).Error)
	op := models.SaasFeature{FeatureCode: "button_role_permission", FeatureName: "角色-权限设置", FeatureType: "OPERATION", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&op).Error)

	handler := &IdentityHandler{db: db}
	handler.repairPackageFeatureParentIDs()

	var repaired models.SaasFeature
	require.NoError(t, db.Where("feature_code = ?", "button_role_permission").First(&repaired).Error)
	require.Equal(t, menu.ID, repaired.ParentID)
}

func findPlanMatrixNode(nodes []dto.PlanMatrixNode, id string) *dto.PlanMatrixNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if child := findPlanMatrixNode(nodes[i].Children, id); child != nil {
			return child
		}
	}
	return nil
}
