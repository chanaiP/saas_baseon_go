package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

type IdentityHandler struct {
	db          *gorm.DB
	redis       *redis.Client
	authSecret  string
	tokenTTL    time.Duration
	jwtFallback bool
}

func NewIdentityHandler(db *gorm.DB, redisClient *redis.Client, authSecret string, tokenTTLHours int, jwtFallback bool) *IdentityHandler {
	if tokenTTLHours <= 0 {
		tokenTTLHours = 24
	}
	return &IdentityHandler{db: db, redis: redisClient, authSecret: authSecret, tokenTTL: time.Duration(tokenTTLHours) * time.Hour, jwtFallback: jwtFallback}
}

func (h *IdentityHandler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := h.currentUser(c)
		if !ok {
			response.Error(c, 401, response.CodeUnauthorized, "登录已失效")
			c.Abort()
			return
		}
		if !h.requestTenantBodyAllowed(c, user) {
			response.Error(c, 403, response.CodeForbidden, "tenant_id 与当前租户上下文不一致")
			c.Abort()
			return
		}
		if !h.routeAllowed(user, c.Request.Method, c.FullPath()) {
			response.Error(c, 403, response.CodeForbidden, "无操作权限")
			c.Abort()
			return
		}
		c.Set("user_id", user.ID)
		c.Set("tenant_id", user.TenantID)
		c.Next()
	}
}

func (h *IdentityHandler) InvalidateAllAuthorizationCache() {
	h.invalidateAllAuthorizationCache()
}

func (h *IdentityHandler) requestTenantBodyAllowed(c *gin.Context, user models.AppUser) bool {
	if c.Request == nil || c.Request.Body == nil {
		return true
	}
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return true
	}
	if !strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
		return true
	}
	raw, err := c.GetRawData()
	if err != nil {
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
	if len(bytes.TrimSpace(raw)) == 0 {
		return true
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return true
	}
	rawTenantID, ok := body["tenant_id"]
	if !ok {
		return true
	}
	var tenantID uint64
	if err := json.Unmarshal(rawTenantID, &tenantID); err != nil || tenantID == 0 {
		return true
	}
	if user.IsPlatformAdmin {
		return true
	}
	return tenantID == user.TenantID
}

func (h *IdentityHandler) UpdatePermissionDataPermMode(c *gin.Context) {
	if c.Param("id") == "" {
		response.OK(c, gin.H{})
		return
	}
	var body struct {
		DataPermMode string `json:"data_perm_mode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.DataPermMode == "" {
		body.DataPermMode = "ORG"
	}
	if err := h.db.Model(&models.Permission{}).Where("id = ?", c.Param("id")).Update("data_perm_mode", body.DataPermMode).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateAllAuthorizationCache()
	response.OK(c, gin.H{"id": parseUintParam(c, "id"), "data_perm_mode": body.DataPermMode})
}
