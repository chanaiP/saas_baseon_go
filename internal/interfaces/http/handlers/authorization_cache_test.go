package handlers

import (
	"strings"
	"testing"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestAuthorizationCacheKeyIncludesUserVersionAndSubscriptionFilter(t *testing.T) {
	key := authorizationCacheKey(models.AppUser{ID: 7, SessionVersion: 3}, true)
	for _, part := range []string{"authz:user:7:", "sv:3", "subscription:true"} {
		if !strings.Contains(key, part) {
			t.Fatalf("expected %q to contain %q", key, part)
		}
	}
}

func TestAuthorizationCacheInvalidationNoopsWithoutRedis(t *testing.T) {
	h := &IdentityHandler{}
	h.invalidateUserAuthorizationCache(1)
	h.invalidateRoleAuthorizationCache(1)
	h.invalidateTenantAuthorizationCache(1)
	h.invalidateAllAuthorizationCache()
}
