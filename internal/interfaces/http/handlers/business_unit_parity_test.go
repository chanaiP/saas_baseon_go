package handlers

import (
	"testing"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"github.com/stretchr/testify/require"
)

func TestOrgTypeForBusinessUnitMap(t *testing.T) {
	require.Equal(t, "company", orgTypeForBusinessUnitMap(models.OrgNode{NodeType: "company"}))
	require.Equal(t, "store", orgTypeForBusinessUnitMap(models.OrgNode{NodeType: "store"}))
	require.Equal(t, "department", orgTypeForBusinessUnitMap(models.OrgNode{NodeType: "department"}))
	require.Equal(t, "department", orgTypeForBusinessUnitMap(models.OrgNode{NodeType: "unknown"}))
}
