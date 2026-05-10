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
		DataPermMode   *string `json:"data_perm_mode"`
		IsPlatformOnly *bool   `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var row models.Permission
	if err := h.db.Where("id = ? AND deleted_at IS NULL", c.Param("id")).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "权限不存在")
		return
	}
	updates := map[string]interface{}{}
	if body.DataPermMode != nil {
		mode := strings.TrimSpace(*body.DataPermMode)
		if mode == "" {
			mode = "ORG"
		}
		if !allowedString(mode, "NONE", "ORG", "BU", "ORG_BU") {
			response.Error(c, 400, response.CodeBadRequest, "无效的数据权限类型: "+mode)
			return
		}
		if row.PermType != 3 {
			response.Error(c, 400, response.CodeBadRequest, "仅菜单权限支持配置数据权限类型")
			return
		}
		updates["data_perm_mode"] = mode
		row.DataPermMode = mode
	}
	if body.IsPlatformOnly != nil {
		updates["is_platform_only"] = *body.IsPlatformOnly
		row.IsPlatformOnly = *body.IsPlatformOnly
		if *body.IsPlatformOnly {
			updates["is_package_feature"] = false
			row.IsPackageFeature = false
		}
	}
	if len(updates) == 0 {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&models.Permission{}).Where("id = ?", row.ID).Updates(updates).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	if row.IsPlatformOnly || !row.IsPackageFeature {
		h.disablePackageFeatureForPermission(row)
	}
	h.syncPackageFeaturesFromPermissions()
	h.invalidateAllAuthorizationCache()
	response.OK(c, gin.H{"id": row.ID, "data_perm_mode": row.DataPermMode, "is_platform_only": row.IsPlatformOnly, "is_package_feature": row.IsPackageFeature})
}
