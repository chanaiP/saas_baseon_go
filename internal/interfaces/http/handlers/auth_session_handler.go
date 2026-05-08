package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) activeLoginTenantOptions(users []models.AppUser) []gin.H {
	seen := map[uint64]struct{}{}
	items := make([]gin.H, 0, len(users))
	for _, user := range users {
		if _, ok := seen[user.TenantID]; ok {
			continue
		}
		var tenant models.Tenant
		if err := h.db.First(&tenant, user.TenantID).Error; err == nil && tenant.Status == 1 && h.subscriptionAllowsLogin(tenant.ID) {
			seen[tenant.ID] = struct{}{}
			items = append(items, gin.H{"tenant_id": tenant.ID, "code": tenant.Code, "name": tenant.Name})
		}
	}
	return items
}

func (h *IdentityHandler) recordLoginFailure(c *gin.Context, account string, userID *uint64, tenantID *uint64, message string) {
	h.incrLoginFail(c, account)
	h.incrIPLoginFail(c)
	h.recordLogin(c, account, userID, tenantID, false, message)
}

func (h *IdentityHandler) issueLoginToken(user models.AppUser) (string, error) {
	if h.redis != nil {
		token := randomHex(32)
		issuedAt := time.Now().UnixNano()
		payload, _ := json.Marshal(gin.H{"user_id": user.ID, "tenant_id": user.TenantID, "session_version": user.SessionVersion, "issued_at": issuedAt})
		if err := h.redis.Set(context.Background(), authSessionKey(token), string(payload), h.tokenTTL).Err(); err == nil {
			return token, nil
		}
	}
	if !h.jwtFallback {
		return "", errors.New("redis session unavailable")
	}
	return issueToken(user.ID, user.TenantID, user.SessionVersion, h.authSecret, h.tokenTTL)
}

func looksLikeMobileAccount(account string) bool {
	normalized := strings.TrimSpace(account)
	if len(normalized) != 11 || normalized[0] != '1' {
		return false
	}
	for _, ch := range normalized {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func authSessionKey(token string) string {
	return "auth:session:" + token
}

func (h *IdentityHandler) ipLoginRateLimitMessage(c *gin.Context) string {
	if h.redis == nil {
		return ""
	}
	key := loginIPKey(c.ClientIP())
	ctx := context.Background()
	count, _ := h.redis.Get(ctx, key).Int()
	if count >= 15 {
		ttl, _ := h.redis.TTL(ctx, key).Result()
		seconds := int(ttl.Seconds())
		if seconds < 0 {
			seconds = 0
		}
		return fmt.Sprintf("请求过于频繁，请 %d 秒后重试", seconds)
	}
	return ""
}

func (h *IdentityHandler) incrIPLoginFail(c *gin.Context) {
	if h.redis == nil {
		return
	}
	key := loginIPKey(c.ClientIP())
	ctx := context.Background()
	_ = h.redis.Incr(ctx, key).Err()
	_ = h.redis.Expire(ctx, key, 5*time.Minute).Err()
}

func (h *IdentityHandler) resetIPLoginFail(c *gin.Context) {
	if h.redis != nil {
		_ = h.redis.Del(context.Background(), loginIPKey(c.ClientIP())).Err()
	}
}

func (h *IdentityHandler) rateLimitExceeded(c *gin.Context, key string, limit int64, ttl time.Duration) (string, bool) {
	if h.redis == nil || key == "" || limit <= 0 {
		return "", false
	}
	ctx := context.Background()
	count, err := h.redis.Incr(ctx, key).Result()
	if err != nil {
		return "", false
	}
	if count == 1 {
		_ = h.redis.Expire(ctx, key, ttl).Err()
	}
	if count <= limit {
		return "", false
	}
	remaining, _ := h.redis.TTL(ctx, key).Result()
	seconds := int(remaining.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("请求过于频繁，请 %d 秒后重试", seconds), true
}

func loginIPKey(ip string) string {
	return "login_ip:" + ip
}

func validateNewPassword(oldPassword, newPassword, confirm string) string {
	if len(oldPassword) < 1 || len(oldPassword) > 128 || len(newPassword) < 8 || len(newPassword) > 128 || len(confirm) < 8 || len(confirm) > 128 {
		return "请求参数错误"
	}
	hasLetter := false
	hasDigit := false
	for _, ch := range newPassword {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetter = true
		}
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "新密码须同时包含英文字母与数字"
	}
	if newPassword != confirm {
		return "两次输入的新密码不一致"
	}
	if oldPassword == newPassword {
		return "新密码不能与当前密码相同"
	}
	return ""
}

func (h *IdentityHandler) passwordChangeBlockMessage(userID uint64) string {
	if h.redis == nil {
		return ""
	}
	if v, err := h.redis.Get(context.Background(), fmt.Sprintf("auth:pwd_block:%d", userID)).Result(); err == nil && v != "" {
		return "密码错误尝试过多，请 15 分钟后再试修改密码"
	}
	return ""
}

func (h *IdentityHandler) recordPasswordChangeFailure(userID uint64) {
	if h.redis == nil {
		return
	}
	ctx := context.Background()
	failKey := fmt.Sprintf("auth:pwd_fail:%d", userID)
	n, _ := h.redis.Incr(ctx, failKey).Result()
	if n == 1 {
		_ = h.redis.Expire(ctx, failKey, 15*time.Minute).Err()
	}
	if n >= 5 {
		_ = h.redis.Set(ctx, fmt.Sprintf("auth:pwd_block:%d", userID), "1", 15*time.Minute).Err()
		_ = h.redis.Del(ctx, failKey).Err()
	}
}

func (h *IdentityHandler) clearPasswordChangeGuard(userID uint64) {
	if h.redis != nil {
		_ = h.redis.Del(context.Background(), fmt.Sprintf("auth:pwd_fail:%d", userID), fmt.Sprintf("auth:pwd_block:%d", userID)).Err()
	}
}

func (h *IdentityHandler) platformAdminCountExcept(userID uint64) int64 {
	var count int64
	_ = h.db.Model(&models.AppUser{}).Where("is_platform_admin = ? AND id <> ? AND deleted_at IS NULL", true, userID).Count(&count).Error
	return count
}

type tenantCapabilityProfile struct {
	Subscription interface{}
	Features     []string
	Quotas       gin.H
}
