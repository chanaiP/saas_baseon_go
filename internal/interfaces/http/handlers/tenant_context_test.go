package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestTenantContextNonPlatformUserCannotSwitchByQuery(t *testing.T) {
	h := &IdentityHandler{}
	ctx := h.tenantContextForUser(models.AppUser{TenantID: 7}, "9")

	require.Equal(t, uint64(7), ctx.ActorTenantID)
	require.Equal(t, uint64(7), ctx.TargetTenantID)
	require.Equal(t, PlatformActorNone, ctx.ActorKind)
	require.False(t, ctx.CrossTenantOperator)
}

func TestTenantContextPlatformAdminCrossTenantOperator(t *testing.T) {
	h := &IdentityHandler{}
	ctx := h.tenantContextForUser(models.AppUser{TenantID: 1, IsPlatformAdmin: true}, "9")

	require.Equal(t, uint64(1), ctx.ActorTenantID)
	require.Equal(t, uint64(9), ctx.TargetTenantID)
	require.Equal(t, PlatformActorCrossOperator, ctx.ActorKind)
	require.True(t, ctx.CrossTenantOperator)
}
