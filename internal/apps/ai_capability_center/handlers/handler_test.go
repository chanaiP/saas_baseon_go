package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/apps/ai_capability_center/services"
)

func TestApplyInvokeIdentityUsesAuthenticatedContext(t *testing.T) {
	req := services.InvokeRequest{
		TenantID: "",
		UserID:   "client-supplied-user",
	}

	err := applyInvokeIdentity(&req, 42, 9)

	require.NoError(t, err)
	require.Equal(t, "9", req.TenantID)
	require.Equal(t, "42", req.UserID)
}

func TestApplyInvokeIdentityAllowsMatchingTenantAndOverridesUser(t *testing.T) {
	req := services.InvokeRequest{
		TenantID: "9",
		UserID:   "client-supplied-user",
	}

	err := applyInvokeIdentity(&req, 42, 9)

	require.NoError(t, err)
	require.Equal(t, "9", req.TenantID)
	require.Equal(t, "42", req.UserID)
}

func TestApplyInvokeIdentityRejectsTenantSpoofing(t *testing.T) {
	req := services.InvokeRequest{
		TenantID: "8",
		UserID:   "client-supplied-user",
	}

	err := applyInvokeIdentity(&req, 42, 9)

	require.ErrorIs(t, err, errInvokeTenantMismatch)
	require.Equal(t, "8", req.TenantID)
	require.Equal(t, "client-supplied-user", req.UserID)
}
