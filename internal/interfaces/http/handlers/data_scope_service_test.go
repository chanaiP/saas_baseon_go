package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestChooseOrgDataScopeCoversCoreModes(t *testing.T) {
	require.Equal(t, "SELF", chooseOrgDataScope([]string{"SELF"}, nil, nil, nil).Scope)
	require.Equal(t, "ORG", chooseOrgDataScope([]string{"ORG"}, nil, nil, nil).Scope)
	require.Equal(t, "ORG_SUB", chooseOrgDataScope([]string{"ORG_SUB"}, nil, nil, nil).Scope)
	require.Equal(t, "ALL", chooseOrgDataScope([]string{"ORG", "ALL"}, nil, nil, nil).Scope)

	custom := chooseOrgDataScope([]string{"CUSTOM"}, []uint64{1, 1}, []uint64{2}, []uint64{3})
	require.Equal(t, "CUSTOM", custom.Scope)
	require.Equal(t, []uint64{1}, custom.CompanyIDs)
	require.Equal(t, []uint64{2}, custom.DepartmentIDs)
	require.Equal(t, []uint64{3}, custom.UserIDs)
}

func TestBusinessUnitDataScopePlatformAdminDoesNotFilter(t *testing.T) {
	useFilter, ids := (&IdentityHandler{}).businessUnitDataScopeFilter(models.AppUser{IsPlatformAdmin: true})

	require.False(t, useFilter)
	require.Nil(t, ids)
}

func TestFilterOrgTreeIncludesChildrenWhenCompanyAllowed(t *testing.T) {
	companyID := uint64(4)
	tree := []gin.H{
		{
			"id":        companyID,
			"node_type": "company",
			"name":      "wuling",
			"children": []gin.H{
				{"id": uint64(5), "node_type": "department", "name": "技术部", "company_id": companyID, "children": []gin.H{}},
				{"id": uint64(6), "node_type": "warehouse", "name": "成品仓", "company_id": companyID, "children": []gin.H{}},
			},
		},
	}

	filtered := filterOrgTree(tree, map[uint64]struct{}{companyID: {}}, map[uint64]struct{}{})

	require.Len(t, filtered, 1)
	children, ok := filtered[0]["children"].([]gin.H)
	require.True(t, ok)
	require.Len(t, children, 2)
	require.Equal(t, "技术部", children[0]["name"])
	require.Equal(t, "成品仓", children[1]["name"])
}
