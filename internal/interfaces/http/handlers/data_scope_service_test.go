package handlers

import (
	"testing"

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
