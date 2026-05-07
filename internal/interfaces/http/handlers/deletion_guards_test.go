package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredDeletionGuards(t *testing.T) {
	guards := map[string]bool{}
	for _, guard := range requiredDeletionGuards() {
		guards[guard] = true
	}

	for _, key := range []string{
		"permission:role_permission",
		"tenant:app_user",
		"plan:tenant_subscription",
		"org_node:app_user",
		"org_node:business_unit_org_map",
		"position_type:position",
		"position:app_user_position",
		"business_unit:business_unit_org_map",
		"dict_type:dict_item",
	} {
		require.True(t, guards[key], "missing deletion guard %s", key)
	}
}
