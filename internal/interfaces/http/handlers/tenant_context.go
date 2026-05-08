package handlers

import "saas_baseon_go/internal/infrastructure/persistence/postgres/models"

type PlatformActorKind string

const (
	PlatformActorNone          PlatformActorKind = "none"
	PlatformActorTenantUser    PlatformActorKind = "platform_tenant_user"
	PlatformActorAdmin         PlatformActorKind = "platform_admin"
	PlatformActorCrossOperator PlatformActorKind = "cross_tenant_operator"
)

type TenantContext struct {
	ActorTenantID       uint64
	TargetTenantID      uint64
	ActorKind           PlatformActorKind
	CrossTenantOperator bool
}

func (h *IdentityHandler) tenantContextForUser(user models.AppUser, requestedTenantID string) TenantContext {
	target := user.TenantID
	actorKind := PlatformActorNone
	if user.IsPlatformAdmin {
		actorKind = PlatformActorAdmin
		if requested := parseTenantIDValue(requestedTenantID); requested > 0 {
			target = requested
			if requested != user.TenantID {
				actorKind = PlatformActorCrossOperator
			}
		}
	} else if h.viewerHasPlatformScope(user) {
		actorKind = PlatformActorTenantUser
	}
	return TenantContext{
		ActorTenantID:       user.TenantID,
		TargetTenantID:      target,
		ActorKind:           actorKind,
		CrossTenantOperator: actorKind == PlatformActorCrossOperator,
	}
}
