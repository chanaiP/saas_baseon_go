package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type authIdentityClaims struct {
	UserID         uint64
	TenantID       uint64
	SessionVersion int
	IssuedAt       int64
}

func (h *IdentityHandler) currentUser(c *gin.Context) (models.AppUser, bool) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		return models.AppUser{}, false
	}
	identity, ok := h.resolveAuthIdentity(token)
	if !ok {
		return models.AppUser{}, false
	}
	userID := identity.UserID
	tenantID := identity.TenantID
	sessionVersion := identity.SessionVersion
	issuedAt := identity.IssuedAt
	if userID == 0 {
		return models.AppUser{}, false
	}
	var user models.AppUser
	if err := h.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return models.AppUser{}, false
	}
	if user.Status != 1 {
		return models.AppUser{}, false
	}
	if tenantID != 0 && tenantID != user.TenantID {
		return models.AppUser{}, false
	}
	if sessionVersion != 0 && sessionVersion != user.SessionVersion {
		return models.AppUser{}, false
	}
	if user.PasswordChangedAt != nil && issuedAt != 0 && user.PasswordChangedAt.UnixNano() > issuedAt {
		return models.AppUser{}, false
	}
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", user.TenantID).First(&tenant).Error; err != nil || tenant.Status != 1 {
		return models.AppUser{}, false
	}
	if !h.subscriptionAllowsLogin(user.TenantID) {
		return models.AppUser{}, false
	}
	return user, true
}

func (h *IdentityHandler) resolveAuthIdentity(token string) (authIdentityClaims, bool) {
	if token == "" {
		return authIdentityClaims{}, false
	}
	if h.redis != nil {
		raw, err := h.redis.Get(context.Background(), authSessionKey(token)).Result()
		if err == nil && raw != "" {
			var session struct {
				UserID         uint64 `json:"user_id"`
				TenantID       uint64 `json:"tenant_id"`
				SessionVersion int    `json:"session_version"`
				IssuedAt       int64  `json:"issued_at"`
			}
			if json.Unmarshal([]byte(raw), &session) == nil {
				return authIdentityClaims{UserID: session.UserID, TenantID: session.TenantID, SessionVersion: session.SessionVersion, IssuedAt: session.IssuedAt}, true
			}
			return authIdentityClaims{}, false
		}
		if err == nil || errors.Is(err, redis.Nil) {
			return authIdentityClaims{}, false
		}
		if !h.jwtFallback {
			return authIdentityClaims{}, false
		}
	}
	if !h.jwtFallback {
		return authIdentityClaims{}, false
	}
	claims, err := parseToken(token, h.authSecret)
	if err != nil || claims.UserID == 0 {
		return authIdentityClaims{}, false
	}
	return authIdentityClaims{UserID: claims.UserID, TenantID: claims.TenantID, SessionVersion: claims.SessionVersion, IssuedAt: claims.IssuedAt}, true
}

func (h *IdentityHandler) loginFailCount(c *gin.Context, account string) int {
	if h.redis == nil || account == "" {
		return 0
	}
	count, _ := h.redis.Get(context.Background(), "login_fail:"+strings.ToLower(account)).Int()
	return count
}

func (h *IdentityHandler) incrLoginFail(c *gin.Context, account string) {
	if h.redis == nil || account == "" {
		return
	}
	ctx := context.Background()
	key := "login_fail:" + strings.ToLower(account)
	_ = h.redis.Incr(ctx, key).Err()
	_ = h.redis.Expire(ctx, key, 15*time.Minute).Err()
}

func (h *IdentityHandler) resetLoginFail(c *gin.Context, account string) {
	if h.redis != nil && account != "" {
		_ = h.redis.Del(context.Background(), "login_fail:"+strings.ToLower(account)).Err()
	}
}

func (h *IdentityHandler) verifyCaptcha(c *gin.Context, id, code string) bool {
	if h.redis == nil {
		return true
	}
	ctx := context.Background()
	key := "auth:captcha:" + id
	stored, err := h.redis.Get(ctx, key).Result()
	if err != nil {
		key = "captcha:" + id
		stored, err = h.redis.Get(ctx, key).Result()
		if err != nil {
			return false
		}
	}
	ok := strings.EqualFold(strings.TrimSpace(stored), strings.TrimSpace(code))
	if ok {
		_ = h.redis.Del(ctx, key).Err()
	}
	return ok
}

func (h *IdentityHandler) subscriptionAllowsLogin(tenantID uint64) bool {
	return h.subscriptionAllowsLoginUsing(h.db, tenantID)
}

func (h *IdentityHandler) subscriptionAllowsLoginUsing(db *gorm.DB, tenantID uint64) bool {
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return true
	}
	return subscriptionStatusAllowsLogin(sub.SubscriptionStatus, sub.EndTime, time.Now())
}

func subscriptionStatusAllowsLogin(statusRaw string, endTime *time.Time, now time.Time) bool {
	status := strings.ToUpper(statusRaw)
	if status == "OVERDUE" || status == "FROZEN" || status == "EXPIRED" || status == "CANCELLED" {
		return false
	}
	if status != "" && status != "TRIAL" && status != "ACTIVE" {
		return false
	}
	return endTime == nil || endTime.After(now)
}
