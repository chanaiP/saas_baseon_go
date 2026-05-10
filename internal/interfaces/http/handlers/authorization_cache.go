package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const authorizationCacheTTL = 5 * time.Minute

func authorizationCacheKey(user models.AppUser, filterSubscription bool) string {
	return fmt.Sprintf("authz:user:%d:sv:%d:subscription:%t", user.ID, user.SessionVersion, filterSubscription)
}

func (h *IdentityHandler) cachedPermissionCodesForUser(user models.AppUser, filterSubscription bool) []string {
	if h.redis == nil {
		return h.permissionCodesForUser(user, filterSubscription)
	}
	ctx := context.Background()
	key := authorizationCacheKey(user, filterSubscription)
	if raw, err := h.redis.Get(ctx, key).Result(); err == nil && raw != "" {
		var codes []string
		if err := json.Unmarshal([]byte(raw), &codes); err == nil {
			return codes
		}
	}
	codes := h.permissionCodesForUser(user, filterSubscription)
	raw, err := json.Marshal(codes)
	if err != nil {
		return codes
	}
	if err := h.redis.Set(ctx, key, raw, authorizationCacheTTL).Err(); err != nil {
		log.Printf(`{"level":"warn","event":"authorization_cache_set_failed","user_id":%d}`, user.ID)
	}
	return codes
}

func (h *IdentityHandler) invalidateUserAuthorizationCache(userID uint64) {
	if h.redis == nil || userID == 0 {
		return
	}
	h.deleteAuthorizationCacheByPattern(fmt.Sprintf("authz:user:%d:*", userID))
}

func (h *IdentityHandler) invalidateRoleAuthorizationCache(roleID uint64) {
	if h.db == nil || roleID == 0 {
		return
	}
	var userIDs []uint64
	_ = h.db.Model(&models.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &userIDs).Error
	for _, userID := range uniqueUint64s(userIDs) {
		h.invalidateUserAuthorizationCache(userID)
	}
}

func (h *IdentityHandler) invalidateTenantAuthorizationCache(tenantID uint64) {
	if h.db == nil || tenantID == 0 {
		return
	}
	var userIDs []uint64
	_ = h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Pluck("id", &userIDs).Error
	for _, userID := range uniqueUint64s(userIDs) {
		h.invalidateUserAuthorizationCache(userID)
	}
}

func (h *IdentityHandler) invalidateAllAuthorizationCache() {
	if h.redis == nil {
		return
	}
	h.deleteAuthorizationCacheByPattern("authz:user:*")
}

func (h *IdentityHandler) deleteAuthorizationCacheByPattern(pattern string) {
	if h.redis == nil || pattern == "" {
		return
	}
	ctx := context.Background()
	iter := h.redis.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0, 100)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf(`{"level":"warn","event":"authorization_cache_scan_failed","pattern":%q}`, pattern)
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := h.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf(`{"level":"warn","event":"authorization_cache_delete_failed","pattern":%q,"count":%d}`, pattern, len(keys))
	}
}
