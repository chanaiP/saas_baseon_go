package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	apppermission "saas_baseon_go/internal/application/permission"
	domainpermission "saas_baseon_go/internal/domain/permission"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/response"
)

type IdentityHandler struct {
	db         *gorm.DB
	redis      *redis.Client
	authSecret string
	tokenTTL   time.Duration
}

var safeFileIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func NewIdentityHandler(db *gorm.DB, redisClient *redis.Client, authSecret string, tokenTTLHours int) *IdentityHandler {
	if tokenTTLHours <= 0 {
		tokenTTLHours = 24
	}
	return &IdentityHandler{db: db, redis: redisClient, authSecret: authSecret, tokenTTL: time.Duration(tokenTTLHours) * time.Hour}
}

func (h *IdentityHandler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := h.currentUser(c)
		if !ok {
			response.Error(c, 401, response.CodeUnauthorized, "登录已失效")
			c.Abort()
			return
		}
		if !h.routeAllowed(user, c.Request.Method, c.FullPath()) {
			response.Error(c, 403, response.CodeForbidden, "无操作权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var body struct {
		Account     string  `json:"account"`
		Password    string  `json:"password"`
		CaptchaID   string  `json:"captcha_id"`
		CaptchaCode string  `json:"captcha_code"`
		TenantCode  string  `json:"tenant_code"`
		TenantID    *uint64 `json:"tenant_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	body.Account = strings.TrimSpace(body.Account)
	if block := h.ipLoginRateLimitMessage(c); block != "" {
		c.JSON(200, response.Body{Code: 1, Message: block, Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": true, "tenants": []gin.H{}}})
		return
	}
	if h.loginFailCount(c, body.Account) >= 3 {
		if body.CaptchaID == "" || body.CaptchaCode == "" {
			c.JSON(200, response.Body{Code: 1, Message: "需要验证码", Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": true}})
			return
		}
		if !h.verifyCaptcha(c, body.CaptchaID, body.CaptchaCode) {
			h.recordLoginFailure(c, body.Account, nil, nil, "验证码错误")
			response.Error(c, 400, response.CodeBadRequest, "验证码错误")
			return
		}
	}
	query := h.db.Where("app_user.deleted_at IS NULL").Where("(account = ? OR employee_no = ? OR phone = ?)", body.Account, body.Account, body.Account)
	if body.TenantID != nil {
		query = query.Where("tenant_id = ?", *body.TenantID)
	}
	if body.TenantCode != "" {
		query = query.Joins("JOIN tenant t ON t.id = app_user.tenant_id AND t.code = ?", body.TenantCode)
	}
	var candidates []models.AppUser
	if err := query.Order("is_platform_admin desc, id asc").Find(&candidates).Error; err != nil || len(candidates) == 0 {
		h.recordLoginFailure(c, body.Account, nil, nil, "账号或密码错误")
		response.Error(c, 400, response.CodeBadRequest, "账号或密码错误")
		return
	}
	if body.TenantID == nil && looksLikeMobileAccount(body.Account) {
		tenantOptions := h.activeLoginTenantOptions(candidates)
		if len(tenantOptions) > 1 {
			c.JSON(200, response.Body{Code: 2, Message: "请选择主体", Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": false, "tenants": tenantOptions}})
			return
		}
	}
	matched := make([]models.AppUser, 0, 1)
	for _, candidate := range candidates {
		if verifyPassword(body.Password, candidate.PasswordHash) {
			matched = append(matched, candidate)
		}
	}
	if len(matched) == 0 {
		h.recordLoginFailure(c, body.Account, nil, nil, "账号或密码错误")
		response.Error(c, 400, response.CodeBadRequest, "账号或密码错误")
		return
	}
	if len(matched) > 1 && body.TenantID == nil && body.TenantCode == "" {
		c.JSON(200, response.Body{Code: 2, Message: "请选择主体", Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": false, "tenants": h.activeLoginTenantOptions(matched)}})
		return
	}
	user := matched[0]
	if user.Status != 1 {
		h.recordLoginFailure(c, body.Account, &user.ID, &user.TenantID, "账号已停用")
		response.Error(c, 400, response.CodeBadRequest, "账号已停用")
		return
	}
	var tenant models.Tenant
	if err := h.db.First(&tenant, user.TenantID).Error; err == nil && tenant.Status != 1 {
		h.recordLoginFailure(c, body.Account, &user.ID, &user.TenantID, "所属主体已停用")
		response.Error(c, 400, response.CodeBadRequest, "所属主体已停用，请联系管理员")
		return
	}
	if !h.subscriptionAllowsLogin(user.TenantID) {
		h.recordLoginFailure(c, body.Account, &user.ID, &user.TenantID, "账户已到期")
		response.Error(c, 400, response.CodeBadRequest, "账户已到期，请联系管理员续费")
		return
	}
	token, err := h.issueLoginToken(user)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "令牌生成失败")
		return
	}
	h.resetLoginFail(c, body.Account)
	h.resetIPLoginFail(c)
	h.recordLogin(c, body.Account, &user.ID, &user.TenantID, true, "登录成功")
	response.OK(c, gin.H{
		"token":            token,
		"token_type":       "bearer",
		"captcha_required": false,
	})
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	if token := bearerToken(c.GetHeader("Authorization")); token != "" && h.redis != nil {
		_ = h.redis.Del(context.Background(), authSessionKey(token)).Err()
	}
	response.OK(c, gin.H{})
}

func (h *IdentityHandler) Captcha(c *gin.Context) {
	code := randomCode(4)
	id := randomHex(8)
	if h.redis != nil {
		_ = h.redis.Set(context.Background(), "auth:captcha:"+id, strings.ToUpper(code), 5*time.Minute).Err()
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40"><rect fill="#f0f0f0" width="100%%" height="100%%"/><text x="10" y="28" font-size="22" font-family="sans-serif">%s</text></svg>`, code)
	response.OK(c, gin.H{
		"captcha_id":   id,
		"image_base64": "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg)),
	})
}

func (h *IdentityHandler) PhoneLoginTenants(c *gin.Context) {
	account := strings.TrimSpace(c.Query("account"))
	var rows []models.AppUser
	if account != "" {
		_ = h.db.Where("phone = ? AND status = ?", account, 1).Find(&rows).Error
	}
	response.OK(c, h.activeLoginTenantOptions(rows))
}

func (h *IdentityHandler) SwitchableTenants(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if user.Phone == nil || strings.TrimSpace(*user.Phone) == "" {
		response.OK(c, []gin.H{})
		return
	}
	var users []models.AppUser
	_ = h.db.Where("phone = ? AND status = ?", *user.Phone, 1).Order("tenant_id asc, id asc").Find(&users).Error
	items := make([]gin.H, 0, len(users))
	for _, row := range users {
		var tenant models.Tenant
		if err := h.db.First(&tenant, row.TenantID).Error; err == nil && tenant.Status == 1 && h.subscriptionAllowsLogin(tenant.ID) {
			items = append(items, gin.H{"tenant_id": tenant.ID, "tenant_code": tenant.Code, "tenant_name": tenant.Name, "user_id": row.ID, "employee_no": row.EmployeeNo, "user_display_name": coalesceString(row.Name, row.EmployeeNo)})
		}
	}
	response.OK(c, items)
}

func (h *IdentityHandler) SwitchTenant(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		TenantID uint64 `json:"tenant_id"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.TenantID == 0 {
		body.TenantID = user.TenantID
	}
	if body.TenantID == user.TenantID {
		response.Error(c, 400, response.CodeBadRequest, "已在当前主体")
		return
	}
	if user.Phone == nil || strings.TrimSpace(*user.Phone) == "" {
		response.Error(c, 400, response.CodeBadRequest, "不可切换到该主体（需为同一手机号下的账号）")
		return
	}
	var target models.AppUser
	if err := h.db.Where("tenant_id = ? AND phone = ? AND status = ?", body.TenantID, *user.Phone, 1).Order("id asc").First(&target).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, "不可切换到该主体（需为同一手机号下的账号）")
		return
	}
	var tenant models.Tenant
	if err := h.db.First(&tenant, target.TenantID).Error; err != nil || tenant.Status != 1 {
		response.Error(c, 400, response.CodeBadRequest, "主体不可用")
		return
	}
	if !h.subscriptionAllowsLogin(target.TenantID) {
		response.Error(c, 400, response.CodeBadRequest, "账户已到期，请联系管理员续费")
		return
	}
	oldToken := bearerToken(c.GetHeader("Authorization"))
	if oldToken != "" && h.redis != nil {
		_ = h.redis.Del(context.Background(), authSessionKey(oldToken)).Err()
	}
	token, err := h.issueLoginToken(target)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "令牌生成失败")
		return
	}
	h.recordLogin(c, coalesceString(user.EmployeeNo, "switch"), &target.ID, &target.TenantID, true, "切换主体登录")
	response.OK(c, gin.H{"token": token, "token_type": "bearer", "captcha_required": false})
}

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
		payload, _ := json.Marshal(gin.H{"user_id": user.ID, "tenant_id": user.TenantID})
		if err := h.redis.Set(context.Background(), authSessionKey(token), string(payload), h.tokenTTL).Err(); err == nil {
			return token, nil
		}
	}
	return issueToken(user.ID, user.TenantID, h.authSecret, h.tokenTTL)
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
	key := "login_ip:" + c.ClientIP()
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
	key := "login_ip:" + c.ClientIP()
	ctx := context.Background()
	_ = h.redis.Incr(ctx, key).Err()
	_ = h.redis.Expire(ctx, key, 5*time.Minute).Err()
}

func (h *IdentityHandler) resetIPLoginFail(c *gin.Context) {
	if h.redis != nil {
		_ = h.redis.Del(context.Background(), "login_ip:"+c.ClientIP()).Err()
	}
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

func (h *IdentityHandler) tenantCapabilityContext(tenantID uint64) tenantCapabilityProfile {
	empty := tenantCapabilityProfile{Subscription: nil, Features: []string{}, Quotas: gin.H{}}
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", tenantID).First(&tenant).Error; err != nil || tenant.Status != 1 {
		return empty
	}
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil || !h.subscriptionAllowsLogin(tenantID) {
		return empty
	}
	var plan models.SaasPlan
	if err := h.db.Where("id = ? AND deleted_at IS NULL", sub.PlanID).First(&plan).Error; err != nil || plan.Status != 1 {
		return empty
	}
	now := time.Now()
	featureIDs := map[uint64]bool{}
	var planFeatures []models.SaasPlanFeature
	_ = h.db.Where("plan_id = ? AND enabled = ?", plan.ID, true).Find(&planFeatures).Error
	for _, row := range planFeatures {
		featureIDs[row.FeatureID] = true
	}
	var overrides []models.TenantFeatureOverride
	_ = h.db.Where("tenant_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, now, now).Find(&overrides).Error
	for _, override := range overrides {
		featureIDs[override.FeatureID] = override.Enabled
	}
	ids := make([]uint64, 0, len(featureIDs))
	for id, enabled := range featureIDs {
		if enabled {
			ids = append(ids, id)
		}
	}
	features := []string{}
	if len(ids) > 0 {
		var rows []models.SaasFeature
		_ = h.db.Where("id IN ? AND status = ?", ids, 1).Order("feature_code asc").Find(&rows).Error
		for _, row := range rows {
			features = append(features, row.FeatureCode)
		}
	}
	quotas := gin.H{}
	var planQuotas []models.SaasPlanQuota
	_ = h.db.Where("plan_id = ?", plan.ID).Find(&planQuotas).Error
	for _, row := range planQuotas {
		var quota models.SaasQuota
		if err := h.db.Where("id = ? AND status = ?", row.QuotaID, 1).First(&quota).Error; err == nil {
			quotas[quota.QuotaCode] = row.QuotaValue
		}
	}
	var quotaOverrides []models.TenantQuotaOverride
	_ = h.db.Where("tenant_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, now, now).Find(&quotaOverrides).Error
	for _, override := range quotaOverrides {
		var quota models.SaasQuota
		if err := h.db.Where("id = ? AND status = ?", override.QuotaID, 1).First(&quota).Error; err == nil {
			quotas[quota.QuotaCode] = override.QuotaValue
		}
	}
	return tenantCapabilityProfile{
		Subscription: gin.H{"plan_code": plan.PlanCode, "plan_name": plan.PlanName, "status": sub.SubscriptionStatus, "end_time": sub.EndTime},
		Features:     features,
		Quotas:       quotas,
	}
}

func (h *IdentityHandler) filterPermissionCodesForSubscription(codes []string, features []string) []string {
	enabled := map[string]struct{}{}
	for _, feature := range features {
		enabled[feature] = struct{}{}
	}
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		var permission models.Permission
		if err := h.db.Where("path = ?", code).First(&permission).Error; err != nil {
			out = append(out, code)
			continue
		}
		if !permission.IsPackageFeature || permission.FeatureCode == nil || *permission.FeatureCode == "" {
			out = append(out, code)
			continue
		}
		if _, ok := enabled[*permission.FeatureCode]; ok {
			out = append(out, code)
		}
	}
	return out
}

func (h *IdentityHandler) userShortcutIDs(userID uint64) []string {
	var pref models.UserPreference
	if err := h.db.Where("user_id = ? AND pref_key = ?", userID, "shortcut_ids").First(&pref).Error; err != nil || pref.PrefValue == nil {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(*pref.PrefValue), &ids); err != nil {
		return []string{}
	}
	return ids
}

func (h *IdentityHandler) saveUserShortcutIDs(userID uint64, ids []string) error {
	raw, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	value := string(raw)
	now := time.Now()
	var pref models.UserPreference
	err = h.db.Where("user_id = ? AND pref_key = ?", userID, "shortcut_ids").First(&pref).Error
	if err == nil {
		return h.db.Model(&pref).Updates(map[string]interface{}{"pref_value": value, "updated_at": now}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return h.db.Create(&models.UserPreference{UserID: userID, PrefKey: "shortcut_ids", PrefValue: &value, UpdatedAt: now}).Error
}

func (h *IdentityHandler) viewerHasPlatformScope(user models.AppUser) bool {
	if user.IsPlatformAdmin {
		return true
	}
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", user.TenantID).First(&tenant).Error; err != nil {
		return false
	}
	return tenant.IsPlatform
}

func (h *IdentityHandler) userCanEditTenantBranding(user models.AppUser) bool {
	if user.IsPlatformAdmin {
		return true
	}
	if !h.viewerHasPlatformScope(user) && !h.tenantFeatureAllowed(user.TenantID, "brand_config") {
		return false
	}
	var count int64
	_ = h.db.Model(&models.Permission{}).
		Joins("JOIN role_permission rp ON rp.permission_id = permission.id").
		Joins("JOIN user_role ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ? AND permission.path = ? AND permission.enabled = ?", user.ID, "brand:edit", true).
		Count(&count).Error
	return count > 0
}

func (h *IdentityHandler) invalidateSessionsForTenant(tenantID uint64) {
	if h.redis == nil || tenantID == 0 {
		return
	}
	ctx := context.Background()
	var cursor uint64
	for {
		keys, next, err := h.redis.Scan(ctx, cursor, "auth:session:*", 100).Result()
		if err != nil {
			return
		}
		for _, key := range keys {
			raw, err := h.redis.Get(ctx, key).Result()
			if err != nil || raw == "" {
				continue
			}
			var session struct {
				TenantID uint64 `json:"tenant_id"`
			}
			if json.Unmarshal([]byte(raw), &session) == nil && session.TenantID == tenantID {
				_ = h.redis.Del(ctx, key).Err()
			}
		}
		if next == 0 {
			return
		}
		cursor = next
	}
}

func (h *IdentityHandler) Profile(c *gin.Context) {
	var user models.AppUser
	var tenant models.Tenant
	var roles []models.Role
	var permissions []models.Permission
	if loaded, ok := h.currentUser(c); ok {
		user = loaded
	} else {
		_ = h.db.Where("account IN ?", []string{"E10001", "admin"}).Order("account = 'E10001' desc, id asc").First(&user).Error
	}
	if user.ID != 0 {
		_ = h.db.First(&tenant, user.TenantID).Error
		_ = h.db.
			Joins("JOIN user_role ur ON ur.role_id = role.id").
			Where("ur.user_id = ?", user.ID).
			Find(&roles).Error
		_ = h.db.
			Joins("JOIN role_permission rp ON rp.permission_id = permission.id").
			Joins("JOIN user_role ur ON ur.role_id = rp.role_id").
			Where("ur.user_id = ? AND permission.enabled = ?", user.ID, true).
			Find(&permissions).Error
	}
	roleIDs := make([]uint64, 0, len(roles))
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
		roleCodes = append(roleCodes, role.Code)
	}
	permissionCodes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		if permission.PermType == 2 {
			permissionCodes = append(permissionCodes, permission.Path)
		}
	}
	if len(permissionCodes) == 0 {
		permissionCodes = allDevPermissionCodes()
	}
	capability := h.tenantCapabilityContext(user.TenantID)
	tenantIsPlatform := tenant.IsPlatform
	if !user.IsPlatformAdmin && !tenantIsPlatform {
		permissionCodes = h.filterPermissionCodesForSubscription(permissionCodes, capability.Features)
	}

	response.OK(c, gin.H{
		"id":                 user.ID,
		"tenant_id":          user.TenantID,
		"employee_no":        user.EmployeeNo,
		"phone":              user.Phone,
		"name":               user.Name,
		"email":              user.Email,
		"avatar_url":         user.AvatarURL,
		"status":             user.Status,
		"company_id":         nil,
		"department_id":      nil,
		"role_ids":           roleIDs,
		"role_codes":         roleCodes,
		"permission_codes":   permissionCodes,
		"is_platform_admin":  user.IsPlatformAdmin,
		"tenant_is_platform": tenantIsPlatform,
		"shortcut_ids":       h.userShortcutIDs(user.ID),
		"subscription":       capability.Subscription,
		"features":           capability.Features,
		"quotas":             capability.Quotas,
	})
}

func (h *IdentityHandler) UpdateProfile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Name      *string `json:"name"`
		Phone     *string `json:"phone"`
		Email     *string `json:"email"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Phone != nil {
		updates["phone"] = nullableTrimmed(body.Phone)
	}
	if body.Email != nil {
		updates["email"] = nullableTrimmed(body.Email)
	}
	if body.AvatarURL != nil {
		updates["avatar_url"] = nullableTrimmed(body.AvatarURL)
	}
	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		h.audit(c, user.TenantID, user.ID, "profile", "update", "更新个人资料", updates)
	}
	h.Profile(c)
}

func (h *IdentityHandler) UpdatePassword(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		OldPassword        string `json:"old_password"`
		NewPassword        string `json:"new_password"`
		NewPasswordConfirm string `json:"new_password_confirm"`
		CaptchaID          string `json:"captcha_id"`
		CaptchaCode        string `json:"captcha_code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if block := h.passwordChangeBlockMessage(user.ID); block != "" {
		c.JSON(429, response.Body{Code: 42900, Message: block})
		return
	}
	if strings.TrimSpace(body.CaptchaID) == "" || strings.TrimSpace(body.CaptchaCode) == "" || !h.verifyCaptcha(c, body.CaptchaID, body.CaptchaCode) {
		response.Error(c, 400, response.CodeBadRequest, "验证码错误或已过期，请刷新验证码后重试")
		return
	}
	if msg := validateNewPassword(body.OldPassword, body.NewPassword, body.NewPasswordConfirm); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	if !verifyPassword(body.OldPassword, user.PasswordHash) {
		h.recordPasswordChangeFailure(user.ID)
		response.Error(c, 400, response.CodeBadRequest, "旧密码不正确")
		return
	}
	hashed, err := hashPassword(body.NewPassword)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "密码加密失败")
		return
	}
	if err := h.db.Model(&user).Update("password_hash", hashed).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.clearPasswordChangeGuard(user.ID)
	h.audit(c, user.TenantID, user.ID, "user", "update_password", "修改密码 "+user.Name, nil)
	response.OK(c, gin.H{})
}

func (h *IdentityHandler) Preferences(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	response.OK(c, gin.H{"shortcut_ids": h.userShortcutIDs(user.ID)})
}

func (h *IdentityHandler) SavePreferences(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		ShortcutIDs []string `json:"shortcut_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.saveUserShortcutIDs(user.ID, body.ShortcutIDs); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.audit(c, user.TenantID, user.ID, "user", "update_shortcuts", "更新快捷入口 "+user.Name, nil)
	response.OK(c, gin.H{"shortcut_ids": body.ShortcutIDs})
}

func (h *IdentityHandler) TenantBranding(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", user.TenantID).First(&tenant).Error; err != nil {
		response.OK(c, gin.H{"display_name": "管理台", "brand_display_name": nil, "logo_data": nil, "footer_text": nil, "tenant_name": "", "can_edit": false, "can_edit_footer": false})
		return
	}
	nameOverride := ""
	if tenant.BrandName != nil {
		nameOverride = strings.TrimSpace(*tenant.BrandName)
	}
	footerText := (*string)(nil)
	if tenant.FooterText != nil {
		trimmed := strings.TrimSpace(*tenant.FooterText)
		if trimmed != "" {
			footerText = &trimmed
		}
	}
	response.OK(c, gin.H{
		"display_name":       coalesceString(nameOverride, coalesceString(tenant.Name, "管理台")),
		"brand_display_name": nullableFromString(nameOverride),
		"logo_data":          tenant.LogoData,
		"footer_text":        footerText,
		"tenant_name":        coalesceString(tenant.Name, "平台主体"),
		"can_edit":           h.userCanEditTenantBranding(user),
		"can_edit_footer":    h.viewerHasPlatformScope(user),
	})
}

func (h *IdentityHandler) SaveTenantBranding(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		BrandDisplayName *string `json:"brand_display_name"`
		BrandLogoData    *string `json:"brand_logo_data"`
		LogoData         *string `json:"logo_data"`
		BrandFooterText  *string `json:"brand_footer_text"`
		FooterText       *string `json:"footer_text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if !h.userCanEditTenantBranding(user) {
		response.Error(c, 403, response.CodeForbidden, "无品牌维护权限")
		return
	}
	updates := map[string]interface{}{}
	if body.BrandDisplayName != nil {
		updates["brand_display_name"] = nullableTrimmed(body.BrandDisplayName)
	}
	if body.BrandLogoData != nil {
		logo := nullableTrimmed(body.BrandLogoData)
		if logo != nil && len(*logo) > 800000 {
			response.Error(c, 400, response.CodeBadRequest, "Logo 数据过大")
			return
		}
		updates["brand_logo_data"] = logo
	} else if body.LogoData != nil {
		logo := nullableTrimmed(body.LogoData)
		if logo != nil && len(*logo) > 800000 {
			response.Error(c, 400, response.CodeBadRequest, "Logo 数据过大")
			return
		}
		updates["brand_logo_data"] = logo
	}
	if body.BrandFooterText != nil {
		if !h.viewerHasPlatformScope(user) {
			response.Error(c, 403, response.CodeForbidden, "仅平台运维账号可设置底部版权信息")
			return
		}
		footer := nullableTrimmed(body.BrandFooterText)
		if footer != nil && len(*footer) > 256 {
			response.Error(c, 400, response.CodeBadRequest, "版权信息过长")
			return
		}
		updates["brand_footer_text"] = footer
	} else if body.FooterText != nil {
		if !h.viewerHasPlatformScope(user) {
			response.Error(c, 403, response.CodeForbidden, "仅平台运维账号可设置底部版权信息")
			return
		}
		footer := nullableTrimmed(body.FooterText)
		if footer != nil && len(*footer) > 256 {
			response.Error(c, 400, response.CodeBadRequest, "版权信息过长")
			return
		}
		updates["brand_footer_text"] = footer
	}
	if len(updates) > 0 {
		if err := h.db.Model(&models.Tenant{}).Where("id = ?", user.TenantID).Updates(updates).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		h.audit(c, user.TenantID, user.ID, "tenant_branding", "update", "更新主体品牌", updates)
	}
	h.TenantBranding(c)
}

func (h *IdentityHandler) PublicTenantFooter(c *gin.Context) {
	var tenant models.Tenant
	_ = h.db.Where("deleted_at IS NULL").Order("id asc").First(&tenant).Error
	response.OK(c, gin.H{"footer_text": tenant.FooterText})
}

func (h *IdentityHandler) MenuBundles(c *gin.Context) {
	var permissions []models.Permission
	_ = h.db.Where("perm_type = ? AND enabled = ? AND visible = ?", 3, true, true).Order("sort_order asc, id asc").Find(&permissions).Error
	bundles := make([]gin.H, 0, len(permissions))
	for _, permission := range permissions {
		bundles = append(bundles, gin.H{
			"path":               permission.Path,
			"title":              permission.Name,
			"menu_permission_id": permission.ID,
			"data_permission_id": 0,
			"operations":         []gin.H{},
			"is_platform_only":   permission.IsPlatformOnly,
			"is_package_feature": permission.IsPackageFeature,
			"feature_code":       permission.FeatureCode,
			"feature_type":       permission.FeatureType,
			"tenant_visible":     permission.Visible,
			"tenant_editable":    permission.TenantEditable,
			"tenant_edit_scope":  permission.TenantEditScope,
			"data_perm_mode":     permission.DataPermMode,
		})
	}
	response.OK(c, bundles)
}

func (h *IdentityHandler) MenuOverrides(c *gin.Context) {
	response.OK(c, gin.H{"tenant_id": 1, "overrides": []gin.H{}})
}

func (h *IdentityHandler) Permissions(c *gin.Context) {
	var rows []models.Permission
	_ = h.db.Order("sort_order asc, id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, permissionToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) PermissionTree(c *gin.Context) {
	var rows []models.Permission
	_ = h.db.Order("sort_order asc, id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, permissionToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) Permission(c *gin.Context) {
	var row models.Permission
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "权限不存在")
		return
	}
	response.OK(c, permissionToJSON(row))
}

func (h *IdentityHandler) CreatePermission(c *gin.Context) {
	var row models.Permission
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if row.TenantID == 0 {
		row.TenantID = 1
	}
	if row.DataPermMode == "" {
		row.DataPermMode = "ORG"
	}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, permissionToJSON(row))
}

func (h *IdentityHandler) UpdatePermission(c *gin.Context) {
	var row models.Permission
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "权限不存在")
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body).First(&row, row.ID).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, permissionToJSON(row))
}

func (h *IdentityHandler) DeletePermission(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "权限", ref(&models.Permission{}, "子权限", "parent_id = ?", id), ref(&models.RolePermission{}, "角色权限", "permission_id = ?", id), ref(&models.TenantMenuOverride{}, "租户菜单覆盖", "permission_id = ?", id)) {
		return
	}
	h.deleteByID(c, &models.Permission{})
}

func coalesceString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func coalesceStringPtr(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func (h *IdentityHandler) SaveMenuOverrides(c *gin.Context) {
	h.MenuOverrides(c)
}

func (h *IdentityHandler) Tenants(c *gin.Context) {
	skip, limit := paginationParams(c)
	query := h.db.Where("deleted_at IS NULL")
	var total int64
	_ = query.Model(&models.Tenant{}).Count(&total).Error
	var rows []models.Tenant
	_ = query.Order("id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, tenantToJSON(h.db, row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) Tenant(c *gin.Context) {
	var row models.Tenant
	if err := h.db.Where("deleted_at IS NULL").First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	response.OK(c, tenantToJSON(h.db, row))
}

func (h *IdentityHandler) CreateTenant(c *gin.Context) {
	var body struct {
		Code            string  `json:"code"`
		Name            string  `json:"name"`
		Status          int     `json:"status"`
		AdminName       string  `json:"admin_name"`
		AdminEmployeeNo string  `json:"admin_employee_no"`
		AdminPhone      *string `json:"admin_phone"`
		AdminPassword   string  `json:"admin_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tenant, err := h.createTenantWithAdmin(body.Code, body.Name, body.Status, body.AdminName, body.AdminEmployeeNo, body.AdminPhone, body.AdminPassword)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "tenant", "create", "创建主体 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "code": tenant.Code, "name": tenant.Name})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) CreateTenantWithPackage(c *gin.Context) {
	var body struct {
		Tenant struct {
			Code            string  `json:"code"`
			Name            string  `json:"name"`
			Status          int     `json:"status"`
			AdminName       string  `json:"admin_name"`
			AdminEmployeeNo string  `json:"admin_employee_no"`
			AdminPhone      *string `json:"admin_phone"`
			AdminPassword   string  `json:"admin_password"`
		} `json:"tenant"`
		Package tenantPackagePayload `json:"package"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var tenant models.Tenant
	err := h.db.Transaction(func(tx *gorm.DB) error {
		created, err := h.createTenantWithAdminOnDB(tx, body.Tenant.Code, body.Tenant.Name, body.Tenant.Status, body.Tenant.AdminName, body.Tenant.AdminEmployeeNo, body.Tenant.AdminPhone, body.Tenant.AdminPassword)
		if err != nil {
			return err
		}
		tenant = created
		return h.saveTenantPackageOnDB(tx, tenant.ID, body.Package)
	})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "tenant", "create_with_package", "创建主体并配置套餐 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "code": tenant.Code, "name": tenant.Name})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) UpdateTenant(c *gin.Context) {
	var tenant models.Tenant
	if err := h.db.Where("deleted_at IS NULL").First(&tenant, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	var body struct {
		Name             *string `json:"name"`
		Status           *int    `json:"status"`
		ContactName      *string `json:"contact_name"`
		ContactPhone     *string `json:"contact_phone"`
		BrandDisplayName *string `json:"brand_display_name"`
		BrandLogoData    *string `json:"brand_logo_data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.ContactName != nil {
		updates["contact_name"] = nullableTrimmed(body.ContactName)
	}
	if body.ContactPhone != nil {
		phone, msg := normalizeOptionalPhone(body.ContactPhone)
		if msg != "" {
			response.Error(c, 400, response.CodeBadRequest, strings.Replace(msg, "手机号", "联系人手机号", 1))
			return
		}
		updates["contact_phone"] = phone
	}
	if body.BrandDisplayName != nil {
		updates["brand_display_name"] = nullableTrimmed(body.BrandDisplayName)
	}
	if body.BrandLogoData != nil {
		logo := nullableTrimmed(body.BrandLogoData)
		if logo != nil && len(*logo) > 800000 {
			response.Error(c, 400, response.CodeBadRequest, "Logo 数据过大")
			return
		}
		updates["brand_logo_data"] = logo
	}
	if len(updates) > 0 {
		if err := h.db.Model(&tenant).Updates(updates).First(&tenant, tenant.ID).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	h.auditCurrentUser(c, "tenant", "update", "更新主体 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "changes": updates})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) UpdateTenantStatus(c *gin.Context) {
	var body struct {
		Status int `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)
	if err := h.db.Model(&models.Tenant{}).Where("id = ? AND deleted_at IS NULL", c.Param("id")).Update("status", body.Status).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	if body.Status == 0 {
		h.invalidateSessionsForTenant(parseUintParam(c, "id"))
	}
	var tenant models.Tenant
	_ = h.db.First(&tenant, c.Param("id")).Error
	h.auditCurrentUser(c, "tenant", "status_update", "更新主体状态 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "status": body.Status})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) DeleteTenant(c *gin.Context) {
	id := parseUintParam(c, "id")
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&tenant).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, "主体不存在")
		return
	}
	var platformAdminCount int64
	_ = h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND is_platform_admin = ? AND deleted_at IS NULL", id, true).Count(&platformAdminCount).Error
	if platformAdminCount > 0 {
		response.Error(c, 400, response.CodeBadRequest, "该主体下存在系统管理员账号，请先取消其平台权限或迁移后再删除")
		return
	}
	h.invalidateSessionsForTenant(id)
	now := time.Now()
	if err := h.db.Model(&tenant).Updates(map[string]interface{}{"deleted_at": now, "status": 0, "code": tombstoneUniqueValue(tenant.Code, tenant.ID, 64)}).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, "删除失败："+err.Error())
		return
	}
	h.auditCurrentUser(c, "tenant", "delete", "删除主体「"+tenant.Name+"」", gin.H{"tenant_id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) Users(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	skip, limit := paginationParams(c)
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		status, err := strconv.Atoi(raw)
		if err != nil {
			response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
			return
		}
		query = query.Where("status = ?", status)
	}
	if id := parseOptionalUintQuery(c, "company_id"); id != nil {
		query = query.Where("company_id = ?", *id)
	}
	if id := parseOptionalUintQuery(c, "department_id"); id != nil {
		var userIDs []uint64
		_ = h.db.Model(&models.AppUserDepartment{}).Where("department_id = ?", *id).Distinct().Pluck("user_id", &userIDs).Error
		if len(userIDs) > 0 {
			query = query.Where("(department_id = ? OR id IN ?)", *id, userIDs)
		} else {
			query = query.Where("department_id = ?", *id)
		}
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		keyword = strings.TrimSpace(c.Query("kw"))
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(name LIKE ? OR employee_no LIKE ? OR phone LIKE ?)", like, like, like)
	}
	var total int64
	_ = query.Model(&models.AppUser{}).Count(&total).Error
	var rows []models.AppUser
	_ = query.Order("id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.userToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) CreateUser(c *gin.Context) {
	var body struct {
		EmployeeNo    string   `json:"employee_no"`
		Password      string   `json:"password"`
		Name          string   `json:"name"`
		Phone         *string  `json:"phone"`
		Email         *string  `json:"email"`
		CompanyID     *uint64  `json:"company_id"`
		DepartmentID  *uint64  `json:"department_id"`
		DepartmentIDs []uint64 `json:"department_ids"`
		PositionIDs   []uint64 `json:"position_ids"`
		RoleIDs       []uint64 `json:"role_ids"`
		Status        int      `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	tenantID := h.requestTenantID(c)
	phone, msg := normalizeOptionalPhone(body.Phone)
	if msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	departmentIDs := normalizedUserDepartmentIDs(body.DepartmentID, body.DepartmentIDs)
	companyID := body.CompanyID
	departmentID := (*uint64)(nil)
	if len(departmentIDs) > 0 {
		departmentID = &departmentIDs[0]
		if err := h.assertDepartmentsInTenant(tenantID, departmentIDs); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		if companyID == nil {
			if found := h.companyIDForDepartment(tenantID, departmentIDs[0]); found != nil {
				companyID = found
			}
		}
	}
	if err := h.assertPositionsInTenant(tenantID, uniqueUint64s(body.PositionIDs)); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	if err := h.assertRolesInTenant(tenantID, uniqueUint64s(body.RoleIDs)); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	if err := h.assertUserUniqueFields(tenantID, 0, body.EmployeeNo, phone); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	user := models.AppUser{TenantID: tenantID, EmployeeNo: body.EmployeeNo, Account: body.EmployeeNo, PasswordHash: devPasswordHash(body.Password), Name: body.Name, Phone: phone, Email: body.Email, CompanyID: companyID, DepartmentID: departmentID, Status: body.Status}
	if err := h.db.Create(&user).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.replaceUserRelations(user.ID, uniqueUint64s(body.RoleIDs), uniqueUint64s(body.PositionIDs), departmentIDs)
	h.auditCurrentUser(c, "user", "create", "创建用户 "+user.Name, gin.H{"user_id": user.ID, "tenant_id": user.TenantID, "employee_no": user.EmployeeNo})
	response.OK(c, h.userToJSON(user))
}

func (h *IdentityHandler) UpdateUser(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var user models.AppUser
	if err := h.db.Where("tenant_id = ?", tenantID).First(&user, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "用户不存在")
		return
	}
	var body struct {
		Name            *string  `json:"name"`
		Phone           *string  `json:"phone"`
		Email           *string  `json:"email"`
		CompanyID       *uint64  `json:"company_id"`
		DepartmentID    *uint64  `json:"department_id"`
		DepartmentIDs   []uint64 `json:"department_ids"`
		PositionIDs     []uint64 `json:"position_ids"`
		RoleIDs         []uint64 `json:"role_ids"`
		Status          *int     `json:"status"`
		IsPlatformAdmin *bool    `json:"is_platform_admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Phone != nil {
		phone, msg := normalizeOptionalPhone(body.Phone)
		if msg != "" {
			response.Error(c, 400, response.CodeBadRequest, msg)
			return
		}
		if err := h.assertUserUniqueFields(tenantID, user.ID, "", phone); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		updates["phone"] = phone
	}
	if body.Email != nil {
		updates["email"] = *body.Email
	}
	if body.CompanyID != nil {
		updates["company_id"] = *body.CompanyID
	}
	if body.DepartmentID != nil {
		updates["department_id"] = *body.DepartmentID
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.IsPlatformAdmin != nil {
		viewer, ok := h.currentUser(c)
		if !ok || !viewer.IsPlatformAdmin {
			response.Error(c, 403, response.CodeForbidden, "无权限设置系统管理员")
			return
		}
		if user.IsPlatformAdmin && !*body.IsPlatformAdmin && h.platformAdminCountExcept(user.ID) < 1 {
			response.Error(c, 400, response.CodeBadRequest, "至少保留一名系统管理员")
			return
		}
		updates["is_platform_admin"] = *body.IsPlatformAdmin
	}
	departmentIDs := body.DepartmentIDs
	if departmentIDs != nil {
		departmentIDs = uniqueUint64s(departmentIDs)
		if err := h.assertDepartmentsInTenant(tenantID, departmentIDs); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		if len(departmentIDs) > 0 {
			updates["department_id"] = departmentIDs[0]
			if body.CompanyID == nil {
				if companyID := h.companyIDForDepartment(tenantID, departmentIDs[0]); companyID != nil {
					updates["company_id"] = *companyID
				}
			}
		} else {
			updates["department_id"] = nil
		}
	} else if body.DepartmentID != nil {
		if err := h.assertDepartmentsInTenant(tenantID, []uint64{*body.DepartmentID}); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		departmentIDs = []uint64{*body.DepartmentID}
	}
	positionIDs := body.PositionIDs
	if positionIDs != nil {
		positionIDs = uniqueUint64s(positionIDs)
		if err := h.assertPositionsInTenant(tenantID, positionIDs); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	roleIDs := body.RoleIDs
	if roleIDs != nil {
		roleIDs = uniqueUint64s(roleIDs)
		if err := h.assertRolesInTenant(tenantID, roleIDs); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).First(&user, user.ID).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	h.replaceUserRelations(user.ID, roleIDs, positionIDs, departmentIDs)
	h.auditCurrentUser(c, "user", "update", "更新用户 "+user.Name, gin.H{"user_id": user.ID, "tenant_id": user.TenantID, "changes": updates})
	response.OK(c, h.userToJSON(user))
}

func (h *IdentityHandler) ResetUserPassword(c *gin.Context) {
	newPassword := generateRandomPassword(14)
	result := h.db.Model(&models.AppUser{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", c.Param("id"), h.requestTenantID(c)).Update("password_hash", devPasswordHash(newPassword))
	if result.Error != nil {
		response.Error(c, 400, response.CodeBadRequest, result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "用户不存在")
		return
	}
	h.auditCurrentUser(c, "user", "password_reset", "重置用户密码", gin.H{"user_id": parseUintParam(c, "id")})
	response.OK(c, gin.H{"new_password": newPassword})
}

func (h *IdentityHandler) DeleteUser(c *gin.Context) {
	id := parseUintParam(c, "id")
	viewer, ok := h.currentUser(c)
	if ok && viewer.ID == id {
		response.Error(c, 400, response.CodeBadRequest, "不能删除当前登录用户")
		return
	}
	tenantID := h.requestTenantID(c)
	var user models.AppUser
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&user).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "用户不存在")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{"deleted_at": now, "status": 0, "employee_no": tombstoneUniqueValue(user.EmployeeNo, user.ID, 64), "account": tombstoneUniqueValue(user.Account, user.ID, 64)}
	if user.Phone != nil {
		updates["phone"] = tombstoneUniqueValue(*user.Phone, user.ID, 32)
	}
	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "user", "delete", "删除用户 "+user.Name, gin.H{"user_id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) AssignableRoles(c *gin.Context) {
	roles, _, err := h.roleService().List(c.Request.Context(), apppermission.RoleListQuery{TenantID: h.requestTenantID(c), Limit: 200})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	items := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		items = append(items, gin.H{"id": role.ID, "code": role.Code, "name": role.Name})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) Roles(c *gin.Context) {
	skip, limit := paginationParams(c)
	roles, total, err := h.roleService().List(c.Request.Context(), apppermission.RoleListQuery{TenantID: h.requestTenantID(c), Skip: skip, Limit: limit, Keyword: c.Query("kw")})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	items := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		items = append(items, roleToJSON(role))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) Role(c *gin.Context) {
	role, err := h.roleService().Get(c.Request.Context(), parseUintParam(c, "id"))
	if err != nil || role.TenantID != h.requestTenantID(c) {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	response.OK(c, roleToJSON(role))
}

func (h *IdentityHandler) CreateRole(c *gin.Context) {
	var body struct {
		Code          string   `json:"code"`
		Name          string   `json:"name"`
		Description   *string  `json:"description"`
		PermissionIDs []uint64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	role, err := h.roleService().Create(c.Request.Context(), apppermission.RoleCreateCommand{TenantID: h.requestTenantID(c), Code: body.Code, Name: body.Name, Description: body.Description, PermissionIDs: body.PermissionIDs})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "role", "create", "创建角色 "+role.Name, gin.H{"role_id": role.ID, "tenant_id": role.TenantID, "code": role.Code})
	response.OK(c, roleToJSON(role))
}

func (h *IdentityHandler) UpdateRole(c *gin.Context) {
	var body struct {
		Name          *string  `json:"name"`
		Description   *string  `json:"description"`
		PermissionIDs []uint64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	role, err := h.roleService().Update(c.Request.Context(), apppermission.RoleUpdateCommand{ID: parseUintParam(c, "id"), TenantID: h.requestTenantID(c), Name: body.Name, Description: body.Description, PermissionIDs: body.PermissionIDs, UpdatePerms: body.PermissionIDs != nil})
	if err != nil {
		if err == domainpermission.ErrRoleNotFound {
			response.Error(c, 404, response.CodeNotFound, "角色不存在")
			return
		}
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "role", "update", "更新角色 "+role.Name, gin.H{"role_id": role.ID, "tenant_id": role.TenantID, "permission_ids": role.PermissionIDs})
	response.OK(c, roleToJSON(role))
}

func (h *IdentityHandler) DeleteRole(c *gin.Context) {
	role, err := h.roleService().Get(c.Request.Context(), parseUintParam(c, "id"))
	if err != nil || role.TenantID != h.requestTenantID(c) {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	if err := h.roleService().Delete(c.Request.Context(), parseUintParam(c, "id")); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "role", "delete", "删除角色 "+role.Name, gin.H{"role_id": role.ID, "tenant_id": role.TenantID})
	response.OK(c, gin.H{"deleted": parseUintParam(c, "id")})
}

func (h *IdentityHandler) roleService() *apppermission.RoleService {
	return apppermission.NewRoleService(repositories.NewRoleRepository(h.db))
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
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": parseUintParam(c, "id"), "data_perm_mode": body.DataPermMode})
}

func (h *IdentityHandler) Plans(c *gin.Context) {
	skip, limit := paginationParams(c)
	query := h.db.Where("deleted_at IS NULL")
	var total int64
	_ = query.Model(&models.SaasPlan{}).Count(&total).Error
	var rows []models.SaasPlan
	_ = query.Order("sort_order asc, id asc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, planToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) CreatePlan(c *gin.Context) {
	var body struct {
		PlanCode     string  `json:"plan_code"`
		PlanName     string  `json:"plan_name"`
		PlanType     string  `json:"plan_type"`
		BillingCycle string  `json:"billing_cycle"`
		Price        float64 `json:"price"`
		Status       int     `json:"status"`
		IsDefault    bool    `json:"is_default"`
		SortOrder    int     `json:"sort_order"`
		Description  *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	row := models.SaasPlan{PlanCode: strings.TrimSpace(body.PlanCode), PlanName: strings.TrimSpace(body.PlanName), PlanType: body.PlanType, BillingCycle: body.BillingCycle, Price: body.Price, Status: body.Status, IsDefault: body.IsDefault, SortOrder: body.SortOrder, Description: body.Description}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, planToJSON(row))
}

func (h *IdentityHandler) UpdatePlan(c *gin.Context) {
	var row models.SaasPlan
	if err := h.db.Where("deleted_at IS NULL").First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "套餐不存在")
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body).First(&row, row.ID).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, planToJSON(row))
}

func (h *IdentityHandler) CopyPlan(c *gin.Context) {
	var src models.SaasPlan
	if err := h.db.Where("deleted_at IS NULL").First(&src, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "套餐不存在")
		return
	}
	var body struct {
		PlanCode    string  `json:"plan_code"`
		PlanName    string  `json:"plan_name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var dst models.SaasPlan
	err := h.db.Transaction(func(tx *gorm.DB) error {
		dst = src
		dst.ID = 0
		dst.PlanCode = strings.TrimSpace(body.PlanCode)
		dst.PlanName = strings.TrimSpace(body.PlanName)
		dst.Description = body.Description
		dst.IsDefault = false
		dst.CreatedAt = time.Time{}
		dst.UpdatedAt = time.Time{}
		if err := tx.Create(&dst).Error; err != nil {
			return err
		}
		var features []models.SaasPlanFeature
		_ = tx.Where("plan_id = ?", src.ID).Find(&features).Error
		for _, feature := range features {
			if err := tx.Create(&models.SaasPlanFeature{PlanID: dst.ID, FeatureID: feature.FeatureID, Enabled: feature.Enabled}).Error; err != nil {
				return err
			}
		}
		var quotas []models.SaasPlanQuota
		_ = tx.Where("plan_id = ?", src.ID).Find(&quotas).Error
		for _, quota := range quotas {
			if err := tx.Create(&models.SaasPlanQuota{PlanID: dst.ID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, planToJSON(dst))
}

func (h *IdentityHandler) DeletePlan(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "套餐", ref(&models.TenantSubscription{}, "主体订阅", "plan_id = ?", id)) {
		return
	}
	now := time.Now()
	var plan models.SaasPlan
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&plan).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "套餐不存在")
		return
	}
	if err := h.db.Model(&plan).Updates(map[string]interface{}{"deleted_at": now, "status": 0, "plan_code": tombstoneUniqueValue(plan.PlanCode, plan.ID, 64)}).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) PlanMatrix(c *gin.Context) {
	var plans []models.SaasPlan
	var features []models.SaasFeature
	var links []models.SaasPlanFeature
	_ = h.db.Order("sort_order asc, id asc").Find(&plans).Error
	_ = h.db.Order("id asc").Find(&features).Error
	_ = h.db.Find(&links).Error

	enabled := map[uint64]map[uint64]bool{}
	for _, link := range links {
		if enabled[link.FeatureID] == nil {
			enabled[link.FeatureID] = map[uint64]bool{}
		}
		enabled[link.FeatureID][link.PlanID] = link.Enabled
	}

	planItems := make([]gin.H, 0, len(plans))
	for _, plan := range plans {
		planItems = append(planItems, planToJSON(plan))
	}
	nodes := make([]gin.H, 0, len(features))
	for _, feature := range features {
		cells := make([]gin.H, 0, len(plans))
		for _, plan := range plans {
			isEnabled := enabled[feature.ID][plan.ID]
			state := "disabled"
			if isEnabled {
				state = "enabled"
			}
			cells = append(cells, gin.H{
				"plan_id":      plan.ID,
				"plan_code":    plan.PlanCode,
				"enabled":      isEnabled,
				"state":        state,
				"feature_ids":  []uint64{feature.ID},
				"quota_values": []gin.H{},
			})
		}
		nodes = append(nodes, gin.H{
			"id":           feature.FeatureCode,
			"label":        feature.FeatureName,
			"node_type":    "feature",
			"feature_id":   feature.ID,
			"feature_code": feature.FeatureCode,
			"feature_type": feature.FeatureType,
			"description":  feature.Description,
			"children":     []gin.H{},
			"cells":        cells,
		})
	}
	response.OK(c, gin.H{"plans": planItems, "nodes": nodes})
}

func (h *IdentityHandler) Features(c *gin.Context) {
	var rows []models.SaasFeature
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":           row.ID,
			"feature_code": row.FeatureCode,
			"feature_name": row.FeatureName,
			"feature_type": row.FeatureType,
			"parent_id":    row.ParentID,
			"menu_id":      row.MenuID,
			"api_method":   row.APIMethod,
			"api_path":     row.APIPath,
			"service_key":  row.ServiceKey,
			"status":       row.Status,
			"description":  row.Description,
		})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreateFeature(c *gin.Context) {
	var body models.SaasFeature
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Create(&body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, featureToJSON(body))
}

func (h *IdentityHandler) UpdateFeature(c *gin.Context) {
	var row models.SaasFeature
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "功能不存在")
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body).First(&row, row.ID).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, featureToJSON(row))
}

func (h *IdentityHandler) PlanFeatures(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var links []models.SaasPlanFeature
	_ = h.db.Where("plan_id = ? AND enabled = ?", planID, true).Find(&links).Error
	ids := make([]uint64, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.FeatureID)
	}
	response.OK(c, gin.H{"plan_id": planID, "feature_ids": ids})
}

func (h *IdentityHandler) SavePlanFeatures(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var body struct {
		FeatureIDs []uint64 `json:"feature_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.savePlanFeaturesWithIDs(planID, body.FeatureIDs); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"plan_id": planID, "feature_ids": body.FeatureIDs})
}

func (h *IdentityHandler) SavePlanCapabilities(c *gin.Context) {
	var body struct {
		FeatureIDs []uint64 `json:"feature_ids"`
		Quotas     []struct {
			QuotaID    uint64 `json:"quota_id"`
			QuotaValue int    `json:"quota_value"`
		} `json:"quotas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	planID := parseUintParam(c, "id")
	if err := h.savePlanFeaturesWithIDs(planID, body.FeatureIDs); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	_ = h.db.Where("plan_id = ?", planID).Delete(&models.SaasPlanQuota{}).Error
	quotaItems := make([]gin.H, 0, len(body.Quotas))
	for _, quota := range body.Quotas {
		_ = h.db.Create(&models.SaasPlanQuota{PlanID: planID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue}).Error
		quotaItems = append(quotaItems, gin.H{"quota_id": quota.QuotaID, "quota_value": quota.QuotaValue})
	}
	response.OK(c, gin.H{"plan_id": planID, "feature_ids": body.FeatureIDs, "quotas": quotaItems})
}

func (h *IdentityHandler) SavePlanFeaturesWithIDs(c *gin.Context, featureIDs []uint64) {
	planID := parseUintParam(c, "id")
	if err := h.savePlanFeaturesWithIDs(planID, featureIDs); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
}

func (h *IdentityHandler) savePlanFeaturesWithIDs(planID uint64, featureIDs []uint64) error {
	ids := uniqueUint64s(featureIDs)
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("plan_id = ?", planID).Delete(&models.SaasPlanFeature{}).Error; err != nil {
			return err
		}
		for _, featureID := range ids {
			var feature models.SaasFeature
			if err := tx.Where("id = ? AND status = ?", featureID, 1).First(&feature).Error; err != nil {
				return fmt.Errorf("功能不存在或已停用")
			}
			if err := tx.Create(&models.SaasPlanFeature{PlanID: planID, FeatureID: featureID, Enabled: true}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *IdentityHandler) Quotas(c *gin.Context) {
	var rows []models.SaasQuota
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":          row.ID,
			"quota_code":  row.QuotaCode,
			"quota_name":  row.QuotaName,
			"quota_type":  row.QuotaType,
			"period_type": row.PeriodType,
			"unit":        row.Unit,
			"status":      row.Status,
			"description": row.Description,
		})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreateQuota(c *gin.Context) {
	var row models.SaasQuota
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, quotaToJSON(row))
}

func (h *IdentityHandler) UpdateQuota(c *gin.Context) {
	var row models.SaasQuota
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "配额不存在")
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body).First(&row, row.ID).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, quotaToJSON(row))
}

func (h *IdentityHandler) PlanQuotas(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var rows []struct {
		QuotaID    uint64
		QuotaCode  string
		QuotaName  string
		QuotaValue int
		PeriodType *string
		Unit       *string
	}
	_ = h.db.Table("saas_plan_quota pq").
		Select("pq.quota_id, q.quota_code, q.quota_name, pq.quota_value, q.period_type, q.unit").
		Joins("join saas_quota q on q.id = pq.quota_id").
		Where("pq.plan_id = ?", planID).
		Scan(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"quota_id": row.QuotaID, "quota_code": row.QuotaCode, "quota_name": row.QuotaName, "quota_value": row.QuotaValue, "period_type": row.PeriodType, "unit": row.Unit})
	}
	response.OK(c, gin.H{"plan_id": planID, "quotas": items})
}

func (h *IdentityHandler) SavePlanQuotas(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var body struct {
		Quotas []struct {
			QuotaID    uint64 `json:"quota_id"`
			QuotaValue int    `json:"quota_value"`
		} `json:"quotas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	_ = h.db.Where("plan_id = ?", planID).Delete(&models.SaasPlanQuota{}).Error
	for _, quota := range body.Quotas {
		_ = h.db.Create(&models.SaasPlanQuota{PlanID: planID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue}).Error
	}
	h.PlanQuotas(c)
}

func (h *IdentityHandler) TenantQuotaRecords(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var orgs []models.OrgNode
	var bus []models.BusinessUnit
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&orgs).Error
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&bus).Error
	companies, stores := []gin.H{}, []gin.H{}
	for _, org := range orgs {
		item := gin.H{"id": org.ID, "code": org.Code, "name": org.Name, "node_type": org.NodeType, "company_name": nil, "status": org.Status}
		if org.NodeType == "company" {
			companies = append(companies, item)
		}
		if org.NodeType == "store" {
			stores = append(stores, item)
		}
	}
	businessUnits := make([]gin.H, 0, len(bus))
	for _, bu := range bus {
		businessUnits = append(businessUnits, businessUnitToJSON(bu))
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "companies": companies, "stores": stores, "business_units": businessUnits})
}

func (h *IdentityHandler) TenantCompanies(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND node_type = ?", tenantID, "company").Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, orgNodeToJSON(row, []gin.H{}))
	}
	response.OK(c, gin.H{
		"tenant_id":          tenantID,
		"billing_unit_count": len(items),
		"companies":          items,
	})
}

func (h *IdentityHandler) TenantPrimaryAdmin(c *gin.Context) {
	user, ok := h.primaryAdmin(parseUintParam(c, "id"))
	if !ok {
		response.Error(c, 404, response.CodeNotFound, "主管理员不存在")
		return
	}
	response.OK(c, gin.H{"employee_no": user.EmployeeNo, "phone": user.Phone, "name": user.Name})
}

func (h *IdentityHandler) ResetTenantPrimaryAdminPassword(c *gin.Context) {
	user, ok := h.primaryAdmin(parseUintParam(c, "id"))
	if !ok {
		response.Error(c, 404, response.CodeNotFound, "主管理员不存在")
		return
	}
	newPassword := generateRandomPassword(14)
	_ = h.db.Model(&user).Update("password_hash", devPasswordHash(newPassword)).Error
	response.OK(c, gin.H{"employee_no": user.EmployeeNo, "phone": user.Phone, "name": user.Name, "new_password": newPassword})
}

func (h *IdentityHandler) TenantSubscription(c *gin.Context) {
	sub, ok := h.findTenantSubscription(parseUintParam(c, "id"))
	if !ok {
		response.OK(c, nil)
		return
	}
	response.OK(c, tenantSubscriptionToJSON(sub))
}

func (h *IdentityHandler) SaveTenantSubscription(c *gin.Context) {
	var body tenantPackagePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.saveTenantPackage(parseUintParam(c, "id"), body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.TenantSubscription(c)
}

func (h *IdentityHandler) SaveTenantPackageConfig(c *gin.Context) {
	var body tenantPackagePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tenantID := parseUintParam(c, "id")
	if err := h.saveTenantPackage(tenantID, body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	sub, _ := h.findTenantSubscription(tenantID)
	response.OK(c, gin.H{"tenant_id": tenantID, "subscription": tenantSubscriptionToJSON(sub), "quotas": h.tenantQuotaOverridesPayload(tenantID)})
}

func (h *IdentityHandler) TenantFeatureOverrides(c *gin.Context) {
	response.OK(c, h.tenantFeatureOverridesPayload(parseUintParam(c, "id")))
}

func (h *IdentityHandler) SaveTenantFeatureOverrides(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var body struct {
		Overrides []struct {
			FeatureID uint64  `json:"feature_id"`
			Enabled   bool    `json:"enabled"`
			Reason    *string `json:"reason"`
		} `json:"overrides"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	_ = h.db.Where("tenant_id = ?", tenantID).Delete(&models.TenantFeatureOverride{}).Error
	for _, item := range body.Overrides {
		_ = h.db.Create(&models.TenantFeatureOverride{TenantID: tenantID, FeatureID: item.FeatureID, Enabled: item.Enabled, Reason: item.Reason}).Error
	}
	response.OK(c, h.tenantFeatureOverridesPayload(tenantID))
}

func (h *IdentityHandler) TenantQuotaOverrides(c *gin.Context) {
	response.OK(c, h.tenantQuotaOverridesPayload(parseUintParam(c, "id")))
}

func (h *IdentityHandler) SaveTenantQuotaOverrides(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var body struct {
		Overrides []struct {
			QuotaID    uint64  `json:"quota_id"`
			QuotaValue int     `json:"quota_value"`
			Reason     *string `json:"reason"`
		} `json:"overrides"`
		Quotas []struct {
			QuotaID    uint64 `json:"quota_id"`
			QuotaValue int    `json:"quota_value"`
		} `json:"quotas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	_ = h.db.Where("tenant_id = ?", tenantID).Delete(&models.TenantQuotaOverride{}).Error
	for _, item := range body.Overrides {
		_ = h.db.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: item.QuotaID, QuotaValue: item.QuotaValue, Reason: item.Reason}).Error
	}
	for _, item := range body.Quotas {
		_ = h.db.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: item.QuotaID, QuotaValue: item.QuotaValue}).Error
	}
	response.OK(c, h.tenantQuotaOverridesPayload(tenantID))
}

func (h *IdentityHandler) TenantQuotaUsage(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var quotas []models.SaasQuota
	_ = h.db.Find(&quotas).Error
	usages := make([]gin.H, 0, len(quotas))
	for _, quota := range quotas {
		used := h.currentQuotaUsage(tenantID, quota.QuotaCode)
		limit := h.currentQuotaLimit(tenantID, quota.ID)
		remaining := limit - used
		if limit < 0 {
			remaining = -1
		}
		usages = append(usages, gin.H{"quota_id": quota.ID, "quota_code": quota.QuotaCode, "quota_name": quota.QuotaName, "quota_type": quota.QuotaType, "period_type": quota.PeriodType, "period_key": "current", "unit": quota.Unit, "used_value": used, "limit_value": limit, "remaining_value": remaining})
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "usages": usages})
}

func (h *IdentityHandler) TenantFeatureAccess(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	featureCode := c.Param("feature_code")
	response.OK(c, gin.H{"tenant_id": tenantID, "feature_code": featureCode, "allowed": h.tenantFeatureAllowed(tenantID, featureCode)})
}

func (h *IdentityHandler) TenantQuotaCheck(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	quotaCode := c.Param("quota_code")
	increment := 1
	if raw := strings.TrimSpace(c.Query("increment")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			increment = parsed
		}
	}
	var quota models.SaasQuota
	if err := h.db.Where("quota_code = ?", quotaCode).First(&quota).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "配额不存在")
		return
	}
	used := h.currentQuotaUsage(tenantID, quotaCode)
	limit := h.currentQuotaLimit(tenantID, quota.ID)
	remaining := limit - used
	if limit < 0 {
		remaining = -1
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "quota_code": quotaCode, "used_value": used, "limit_value": limit, "remaining_value": remaining, "allowed": limit < 0 || used+increment <= limit, "reason": quotaCheckReason(limit, used, increment)})
}

func (h *IdentityHandler) OrganizationTree(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, orgNodeToJSON(row, []gin.H{}))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) OrganizationDetail(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&rows).Error
	companies, departments, stores := []gin.H{}, []gin.H{}, []gin.H{}
	for _, row := range rows {
		item := orgNodeToJSON(row, []gin.H{})
		switch row.NodeType {
		case "company":
			companies = append(companies, item)
		case "department":
			departments = append(departments, item)
		case "store":
			stores = append(stores, item)
		}
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "companies": companies, "departments": departments, "stores": stores})
}

func (h *IdentityHandler) CreateOrgNode(c *gin.Context) {
	h.createOrgNode(c, "")
}

func (h *IdentityHandler) CreateCompany(c *gin.Context) {
	h.createOrgNode(c, "company")
}

func (h *IdentityHandler) CreateDepartment(c *gin.Context) {
	h.createOrgNode(c, "department")
}

func (h *IdentityHandler) CreateStore(c *gin.Context) {
	h.createOrgNode(c, "store")
}

func (h *IdentityHandler) UpdateOrgNode(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) UpdateCompany(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) UpdateDepartment(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) UpdateStore(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) DeleteOrgNode(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除组织")
}

func (h *IdentityHandler) DeleteCompany(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除公司")
}

func (h *IdentityHandler) DeleteDepartment(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除部门")
}

func (h *IdentityHandler) DeleteStore(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除门店")
}

func (h *IdentityHandler) PositionTypes(c *gin.Context) {
	var rows []models.PositionType
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var count int64
		_ = h.db.Model(&models.Position{}).Where("position_type_id = ?", row.ID).Count(&count).Error
		items = append(items, gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "position_count": count})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreatePositionType(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	row := models.PositionType{TenantID: 1, Name: body.Name, Code: body.Code}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdatePositionType(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&models.PositionType{}).Where("id = ?", c.Param("id")).Updates(body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeletePositionType(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "岗位类型", ref(&models.Position{}, "岗位", "position_type_id = ?", id)) {
		return
	}
	h.deleteByID(c, &models.PositionType{})
}

func (h *IdentityHandler) Positions(c *gin.Context) {
	var rows []models.Position
	query := h.db.Order("id asc")
	if positionTypeID := c.Query("position_type_id"); positionTypeID != "" {
		query = query.Where("position_type_id = ?", positionTypeID)
	}
	_ = query.Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "position_type_id": row.PositionTypeID, "name": row.Name, "code": row.Code})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreatePosition(c *gin.Context) {
	var body struct {
		PositionTypeID uint64 `json:"position_type_id"`
		Name           string `json:"name"`
		Code           string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	row := models.Position{TenantID: 1, PositionTypeID: body.PositionTypeID, Name: body.Name, Code: body.Code}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdatePosition(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&models.Position{}).Where("id = ?", c.Param("id")).Updates(body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeletePosition(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "岗位", ref(&models.AppUserPosition{}, "用户岗位", "position_id = ?", id)) {
		return
	}
	h.deleteByID(c, &models.Position{})
}

func (h *IdentityHandler) BusinessUnits(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var rows []models.BusinessUnit
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) BusinessUnitTree(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var rows []models.BusinessUnit
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessUnit(c *gin.Context) {
	var body struct {
		Name       string   `json:"name"`
		Code       string   `json:"code"`
		BUType     *string  `json:"bu_type"`
		OrgNodeIDs []uint64 `json:"org_node_ids"`
		Status     int      `json:"status"`
		Remark     *string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	tenantID := h.requestTenantID(c)
	row := models.BusinessUnit{TenantID: tenantID, Name: body.Name, Code: body.Code, BUType: body.BUType, Status: body.Status, BillingEnabled: true, StatisticEnabled: true, Remark: body.Remark}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.replaceBusinessUnitMappings(tenantID, row.ID, body.OrgNodeIDs)
	h.auditCurrentUser(c, "business_unit", "create", "创建业务单元 "+row.Name, gin.H{"business_unit_id": row.ID, "tenant_id": row.TenantID, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateBusinessUnit(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	if err := h.db.Where("tenant_id = ?", tenantID).First(&models.BusinessUnit{}, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	var body struct {
		Name       *string  `json:"name"`
		Code       *string  `json:"code"`
		BUType     *string  `json:"bu_type"`
		OrgNodeIDs []uint64 `json:"org_node_ids"`
		Status     *int     `json:"status"`
		Remark     *string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Code != nil {
		updates["code"] = *body.Code
	}
	if body.BUType != nil {
		updates["bu_type"] = *body.BUType
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.Remark != nil {
		updates["remark"] = *body.Remark
	}
	if len(updates) > 0 {
		if err := h.db.Model(&models.BusinessUnit{}).Where("id = ? AND tenant_id = ?", c.Param("id"), tenantID).Updates(updates).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	if body.OrgNodeIDs != nil {
		h.replaceBusinessUnitMappings(tenantID, parseUintParam(c, "id"), body.OrgNodeIDs)
	}
	h.auditCurrentUser(c, "business_unit", "update", "更新业务单元", gin.H{"business_unit_id": parseUintParam(c, "id"), "tenant_id": tenantID, "changes": updates, "org_node_ids": body.OrgNodeIDs})
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeleteBusinessUnit(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "业务单元", ref(&models.BusinessUnitOrgMap{}, "组织映射", "business_unit_id = ?", id), ref(&models.BusinessUnitScope{}, "数据权限范围", "business_unit_id = ?", id)) {
		return
	}
	if h.deleteTenantScopedByID(c, &models.BusinessUnit{}) {
		h.auditCurrentUser(c, "business_unit", "delete", "删除业务单元", gin.H{"business_unit_id": id})
	}
}

func (h *IdentityHandler) BusinessUnitOrgMappings(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var rows []models.BusinessUnitOrgMap
	_ = h.db.Where("tenant_id = ? AND business_unit_id = ?", tenantID, c.Param("id")).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitOrgMapToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessUnitOrgMapping(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	if err := h.db.Where("tenant_id = ?", tenantID).First(&models.BusinessUnit{}, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	var body struct {
		OrgID uint64 `json:"org_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	row := models.BusinessUnitOrgMap{TenantID: tenantID, BusinessUnitID: parseUintParam(c, "id"), OrgID: body.OrgID, OrgType: "org", ScopeType: "include", Status: 1}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "business_unit", "org_mapping_create", "绑定业务单元组织", gin.H{"mapping_id": row.ID, "business_unit_id": row.BusinessUnitID, "org_id": row.OrgID, "tenant_id": row.TenantID})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) DeleteBusinessUnitOrgMapping(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.deleteTenantScopedByID(c, &models.BusinessUnitOrgMap{}) {
		h.auditCurrentUser(c, "business_unit", "org_mapping_delete", "删除业务单元组织映射", gin.H{"mapping_id": id})
	}
}

func (h *IdentityHandler) DictTypes(c *gin.Context) {
	var rows []models.DictType
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":               row.ID,
			"code":             row.Code,
			"name":             row.Name,
			"type_code":        row.Code,
			"type_name":        row.Name,
			"remark":           row.Remark,
			"scope":            row.Scope,
			"tenant_editable":  row.TenantEditable,
			"is_platform_only": row.IsPlatformOnly,
			"status":           1,
		})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreateDictType(c *gin.Context) {
	var body struct {
		Code           string  `json:"code"`
		Name           string  `json:"name"`
		Remark         *string `json:"remark"`
		Scope          string  `json:"scope"`
		TenantEditable bool    `json:"tenant_editable"`
		IsPlatformOnly bool    `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Scope == "" {
		body.Scope = "platform"
	}
	row := models.DictType{TenantID: 1, Code: body.Code, Name: body.Name, Remark: body.Remark, Scope: body.Scope, TenantEditable: body.TenantEditable, IsPlatformOnly: body.IsPlatformOnly}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateDictType(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&models.DictType{}).Where("id = ?", c.Param("id")).Updates(body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	var row models.DictType
	_ = h.db.First(&row, c.Param("id")).Error
	response.OK(c, gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "remark": row.Remark, "scope": row.Scope, "tenant_editable": row.TenantEditable, "is_platform_only": row.IsPlatformOnly})
}

func (h *IdentityHandler) DeleteDictType(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "字典类型", ref(&models.DictItem{}, "字典项", "dict_type_id = ?", id)) {
		return
	}
	h.deleteByID(c, &models.DictType{})
}

func (h *IdentityHandler) DictItemsByCode(c *gin.Context) {
	code := c.Param("code")
	var dictType models.DictType
	if err := h.db.Where("code = ?", code).First(&dictType).Error; err != nil {
		response.OK(c, gin.H{"code": code, "items": []gin.H{}})
		return
	}
	var rows []models.DictItem
	_ = h.db.Where("dict_type_id = ?", dictType.ID).Order("sort_order asc, id asc").Find(&rows).Error
	items := dictItemsToJSON(rows)
	response.OK(c, gin.H{"code": code, "items": items})
}

func (h *IdentityHandler) DictItems(c *gin.Context) {
	var rows []models.DictItem
	query := h.db.Order("sort_order asc, id asc")
	if dictTypeID := c.Query("dict_type_id"); dictTypeID != "" {
		query = query.Where("dict_type_id = ?", dictTypeID)
	}
	_ = query.Find(&rows).Error
	response.OK(c, paginated(dictItemsToJSON(rows)))
}

func (h *IdentityHandler) CreateDictItem(c *gin.Context) {
	var body struct {
		DictTypeID uint64 `json:"dict_type_id"`
		Label      string `json:"label"`
		Value      string `json:"value"`
		SortOrder  int    `json:"sort_order"`
		Enabled    *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	row := models.DictItem{TenantID: 1, DictTypeID: body.DictTypeID, Label: body.Label, Value: body.Value, SortOrder: body.SortOrder, Enabled: enabled}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateDictItem(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&models.DictItem{}).Where("id = ?", c.Param("id")).Updates(body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	var row models.DictItem
	_ = h.db.First(&row, c.Param("id")).Error
	response.OK(c, dictItemsToJSON([]models.DictItem{row})[0])
}

func (h *IdentityHandler) DeleteDictItem(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "字典项", ref(&models.TenantDictItemOverride{}, "租户字典覆盖", "dict_item_id = ?", id)) {
		return
	}
	h.deleteByID(c, &models.DictItem{})
}

func (h *IdentityHandler) RestoreDictItem(c *gin.Context) {
	var row models.DictItem
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "字典项不存在")
		return
	}
	response.OK(c, dictItemsToJSON([]models.DictItem{row})[0])
}

func (h *IdentityHandler) SysParams(c *gin.Context) {
	var rows []models.SystemParam
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":               row.ID,
			"param_key":        row.Key,
			"default_value":    row.Value,
			"param_value":      row.Value,
			"remark":           row.Remark,
			"value_type":       "string",
			"tenant_editable":  true,
			"is_platform_only": false,
			"is_override":      false,
		})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreateSysParam(c *gin.Context) {
	var body struct {
		Key            string `json:"param_key"`
		Value          string `json:"param_value"`
		DefaultValue   string `json:"default_value"`
		Remark         string `json:"remark"`
		ValueType      string `json:"value_type"`
		TenantEditable bool   `json:"tenant_editable"`
		IsPlatformOnly bool   `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	value := body.Value
	if value == "" {
		value = body.DefaultValue
	}
	if body.ValueType == "" {
		body.ValueType = "string"
	}
	row := models.SystemParam{TenantID: 1, Key: body.Key, Value: value, Remark: body.Remark, ValueType: body.ValueType, TenantEditable: body.TenantEditable, IsPlatformOnly: body.IsPlatformOnly}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateSysParam(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if v, ok := body["default_value"]; ok {
		body["param_value"] = v
		delete(body, "default_value")
	}
	if err := h.db.Model(&models.SystemParam{}).Where("id = ?", c.Param("id")).Updates(body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	var row models.SystemParam
	_ = h.db.First(&row, c.Param("id")).Error
	response.OK(c, gin.H{"id": row.ID, "param_key": row.Key, "default_value": row.Value, "param_value": row.Value, "remark": row.Remark, "value_type": row.ValueType, "tenant_editable": row.TenantEditable, "is_platform_only": row.IsPlatformOnly, "is_override": false})
}

func (h *IdentityHandler) DeleteSysParam(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "系统参数", ref(&models.TenantParamValue{}, "租户参数值", "param_id = ?", id)) {
		return
	}
	h.deleteByID(c, &models.SystemParam{})
}

func (h *IdentityHandler) RestoreSysParam(c *gin.Context) {
	var row models.SystemParam
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "参数不存在")
		return
	}
	response.OK(c, gin.H{"id": row.ID, "param_key": row.Key, "default_value": row.Value, "param_value": row.Value, "remark": row.Remark, "value_type": row.ValueType, "tenant_editable": row.TenantEditable, "is_platform_only": row.IsPlatformOnly, "is_override": false})
}

func (h *IdentityHandler) SysParamBatch(c *gin.Context) {
	values := gin.H{}
	var rows []models.SystemParam
	_ = h.db.Find(&rows).Error
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	response.OK(c, gin.H{"values": values})
}

func (h *IdentityHandler) LoginLogs(c *gin.Context) {
	var rows []models.LoginLog
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "tenant_name": h.tenantName(row.TenantID), "user_id": row.UserID, "account": row.Account, "success": row.Success, "message": row.Message, "ip": row.IP, "created_at": row.CreatedAt})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) AuditLogs(c *gin.Context) {
	var rows []models.AuditLog
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "tenant_name": h.tenantName(row.TenantID), "user_id": row.UserID, "module": row.Module, "action": row.Action, "summary": row.Summary, "detail": row.Detail, "ip": row.IP, "created_at": row.CreatedAt})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) MonitorHealthDetail(c *gin.Context) {
	redisOK := false
	if h.redis != nil {
		redisOK = h.redis.Ping(context.Background()).Err() == nil
	}
	response.OK(c, gin.H{"mysql": true, "postgres": true, "redis": redisOK})
}

func (h *IdentityHandler) MonitorServerInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	response.OK(c, gin.H{"python_version": "go " + runtime.Version(), "go_version": runtime.Version(), "pid": os.Getpid(), "cpu_percent": nil, "memory_mb": float64(m.Alloc) / 1024 / 1024, "note": "Go 重构版本运行中"})
}

func (h *IdentityHandler) MonitorScheduledJobs(c *gin.Context) {
	response.OK(c, gin.H{"items": []gin.H{}, "note": "当前 Go 版本暂未启用后台定时任务"})
}

func (h *IdentityHandler) MonitorServicesOverview(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	redisOK := false
	if h.redis != nil {
		redisOK = h.redis.Ping(context.Background()).Err() == nil
	}
	response.OK(c, gin.H{"mysql": true, "postgres": true, "redis": redisOK, "python_version": "go " + runtime.Version(), "go_version": runtime.Version(), "pid": os.Getpid(), "cpu_percent": nil, "memory_mb": float64(m.Alloc) / 1024 / 1024, "note": "Go 重构版本服务概览"})
}

func (h *IdentityHandler) MonitorCacheStats(c *gin.Context) {
	if h.redis == nil {
		response.OK(c, gin.H{"ok": false, "used_memory_human": nil, "keys": 0, "connected_clients": 0, "message": "Redis 未配置"})
		return
	}
	ctx := context.Background()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		response.OK(c, gin.H{"ok": false, "used_memory_human": nil, "keys": 0, "connected_clients": 0, "message": err.Error()})
		return
	}
	info := redisInfoMap(h.redis.Info(ctx, "memory", "clients").Val())
	keys, _ := h.redis.DBSize(ctx).Result()
	clients := 0
	if raw := info["connected_clients"]; raw != "" {
		clients, _ = strconv.Atoi(raw)
	}
	response.OK(c, gin.H{"ok": true, "used_memory_human": info["used_memory_human"], "keys": keys, "connected_clients": clients, "message": "Redis 已连接"})
}

func (h *IdentityHandler) MonitorCacheKeys(c *gin.Context) {
	if h.redis == nil {
		response.OK(c, gin.H{"items": []gin.H{}, "cursor": 0})
		return
	}
	ctx := context.Background()
	pattern := strings.TrimSpace(c.Query("pattern"))
	if pattern == "" {
		pattern = "*"
	}
	cursor, _ := strconv.ParseUint(c.Query("cursor"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit"), 10, 64)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	keys, nextCursor, err := h.redis.Scan(ctx, cursor, pattern, limit).Result()
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	items := make([]gin.H, 0, len(keys))
	for _, key := range keys {
		ttl, _ := h.redis.TTL(ctx, key).Result()
		item := gin.H{"key": key, "ttl_seconds": int64(ttl.Seconds())}
		if ttl < 0 {
			item["ttl_seconds"] = nil
		}
		items = append(items, item)
	}
	response.OK(c, gin.H{"items": items, "cursor": nextCursor})
}

func (h *IdentityHandler) UploadFile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请选择文件")
		return
	}
	if file.Size > 50*1024*1024 {
		response.Error(c, 413, response.CodeBadRequest, "文件过大（最大 50 MB）")
		return
	}
	fileID := randomHex(16)
	originalName := filepath.Base(file.Filename)
	ext := safeFileExt(originalName)
	dir := uploadDir(user.TenantID, time.Now())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		response.Error(c, 500, response.CodeInternal, "创建上传目录失败")
		return
	}
	dst := filepath.Join(dir, fileID+ext)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Error(c, 500, response.CodeInternal, "保存文件失败")
		return
	}
	h.audit(c, user.TenantID, user.ID, "file", "upload", "上传文件 "+originalName, gin.H{"file_id": fileID, "file_name": originalName, "size": file.Size})
	response.OK(c, gin.H{"file_id": fileID, "file_name": originalName, "size": file.Size, "url": "/api/files/download/" + fileID})
}

func (h *IdentityHandler) DownloadFile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	path, err := h.findTenantFile(user.TenantID, c.Param("file_id"))
	if err != nil {
		response.Error(c, 404, response.CodeNotFound, "文件不存在")
		return
	}
	c.FileAttachment(path, filepath.Base(path))
}

func (h *IdentityHandler) DeleteFile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	path, err := h.findTenantFile(user.TenantID, c.Param("file_id"))
	if err != nil {
		response.Error(c, 404, response.CodeNotFound, "文件不存在")
		return
	}
	trash := filepath.Join(uploadRoot(), strconv.FormatUint(user.TenantID, 10), ".trash", time.Now().Format("2006/01/02"))
	_ = os.MkdirAll(trash, 0o755)
	trashPath := filepath.Join(trash, time.Now().Format("150405")+"_"+filepath.Base(path))
	if err := os.Rename(path, trashPath); err != nil {
		response.Error(c, 500, response.CodeInternal, "删除文件失败")
		return
	}
	h.audit(c, user.TenantID, user.ID, "file", "delete", "删除文件 "+c.Param("file_id"), gin.H{"file_id": c.Param("file_id")})
	response.OK(c, gin.H{"message": "已删除"})
}

func (h *IdentityHandler) ExportUsersCSV(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var users []models.AppUser
	_ = h.db.Where("tenant_id = ?", user.TenantID).Order("id asc").Find(&users).Error
	companyNames, deptNames := h.orgNameMaps(user.TenantID)
	rows := make([][]string, 0, len(users)+1)
	rows = append(rows, []string{"employee_no", "name", "phone", "email", "company_name", "department_name", "status"})
	for _, row := range users {
		status := "启用"
		if row.Status != 1 {
			status = "停用"
		}
		rows = append(rows, []string{row.EmployeeNo, row.Name, derefString(row.Phone), derefString(row.Email), companyNames[valueOrZero(row.CompanyID)], deptNames[valueOrZero(row.DepartmentID)], status})
	}
	h.sendCSV(c, "users_export.csv", rows)
}

func (h *IdentityHandler) ImportUsersCSV(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请选择 CSV 文件")
		return
	}
	if file.Size > 5*1024*1024 {
		response.Error(c, 413, response.CodeBadRequest, "文件过大（最大 5 MB）")
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "读取文件失败")
		return
	}
	defer opened.Close()
	content, _ := io.ReadAll(opened)
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})))
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		response.Error(c, 400, response.CodeBadRequest, "CSV 格式错误")
		return
	}
	index := csvHeaderIndex(records[0])
	created, skipped := 0, 0
	errors := []string{}
	for line, record := range records[1:] {
		employeeNo := csvCell(record, index, "employee_no")
		name := csvCell(record, index, "name")
		if employeeNo == "" || name == "" {
			errors = append(errors, fmt.Sprintf("第 %d 行：工号和姓名不能为空", line+2))
			continue
		}
		var count int64
		h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND employee_no = ?", user.TenantID, employeeNo).Count(&count)
		if count > 0 {
			skipped++
			continue
		}
		status := 1
		if csvCell(record, index, "status") == "停用" || csvCell(record, index, "status") == "0" {
			status = 0
		}
		password := mustHashPassword(employeeNo)
		newUser := models.AppUser{TenantID: user.TenantID, EmployeeNo: employeeNo, Account: employeeNo, PasswordHash: password, Name: name, Phone: nullableFromString(csvCell(record, index, "phone")), Email: nullableFromString(csvCell(record, index, "email")), Status: status}
		if err := h.db.Create(&newUser).Error; err != nil {
			errors = append(errors, fmt.Sprintf("第 %d 行：%s", line+2, err.Error()))
			continue
		}
		created++
	}
	h.audit(c, user.TenantID, user.ID, "user", "batch_import", fmt.Sprintf("批量导入用户：创建 %d，跳过 %d", created, skipped), gin.H{"created": created, "skipped": skipped, "errors": errors})
	response.OK(c, gin.H{"created": created, "skipped": skipped, "errors": errors})
}

func (h *IdentityHandler) ExportCompaniesCSV(c *gin.Context) {
	h.exportOrgCSV(c, "companies_export.csv", "company", []string{"name", "code", "company_type", "parent_name", "status"})
}

func (h *IdentityHandler) ExportDepartmentsCSV(c *gin.Context) {
	h.exportOrgCSV(c, "departments_export.csv", "department", []string{"name", "code", "company_name", "parent_name", "status"})
}

func redisInfoMap(raw string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if ok {
			values[key] = value
		}
	}
	return values
}

type tenantPackagePayload struct {
	PlanID             uint64  `json:"plan_id"`
	SubscriptionStatus string  `json:"subscription_status"`
	StartTime          string  `json:"start_time"`
	EndTime            *string `json:"end_time"`
	TrialEndTime       *string `json:"trial_end_time"`
	AutoRenew          bool    `json:"auto_renew"`
	FrozenReason       *string `json:"frozen_reason"`
	Quotas             []struct {
		QuotaID    uint64 `json:"quota_id"`
		QuotaValue int    `json:"quota_value"`
	} `json:"quotas"`
}

func paginated(items interface{}) gin.H {
	total := 0
	switch v := items.(type) {
	case []gin.H:
		total = len(v)
	}
	return gin.H{"items": items, "total": total, "skip": 0, "limit": 50}
}

func paginatedWithTotal(items interface{}, total int64, skip int, limit int) gin.H {
	return gin.H{"items": items, "total": total, "skip": skip, "limit": limit}
}

func paginationParams(c *gin.Context) (int, int) {
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if skip < 0 {
		skip = 0
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return skip, limit
}

func parseOptionalUintQuery(c *gin.Context, key string) *uint64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

func tenantToJSON(db *gorm.DB, row models.Tenant) gin.H {
	planName, planCode := (*string)(nil), (*string)(nil)
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", row.ID).Order("id desc").First(&sub).Error; err == nil {
		var plan models.SaasPlan
		if err := db.First(&plan, sub.PlanID).Error; err == nil {
			planName = &plan.PlanName
			planCode = &plan.PlanCode
		}
	}
	var companyCount int64
	var storeCount int64
	var userCount int64
	var buCount int64
	db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ? AND deleted_at IS NULL", row.ID, "company", 1).Count(&companyCount)
	db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ? AND status = ? AND deleted_at IS NULL", row.ID, "store", 1).Count(&storeCount)
	db.Model(&models.AppUser{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", row.ID, 1).Count(&userCount)
	db.Model(&models.BusinessUnit{}).Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", row.ID, 1).Count(&buCount)
	contactName, contactPhone := tenantContact(db, row)
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "status": row.Status, "start_date": coalesceTimePtr(row.StartDate, timePtr(sub.StartTime)), "expire_date": coalesceTimePtr(row.ExpireDate, sub.EndTime), "max_companies": tenantQuotaValue(db, row.ID, "max_companies", row.MaxCompanies), "max_stores": tenantQuotaValue(db, row.ID, "max_stores", 0), "max_business_units": tenantQuotaValue(db, row.ID, "max_business_units", 0), "max_users": tenantQuotaValue(db, row.ID, "max_users", row.MaxUsers), "used_companies": companyCount, "used_stores": storeCount, "used_users": userCount, "used_business_units": buCount, "plan_name": planName, "plan_code": planCode, "contact_name": contactName, "contact_phone": contactPhone, "company_count": companyCount, "brand_display_name": row.BrandName, "brand_logo_data": row.LogoData, "created_at": row.CreatedAt}
}

func coalesceTimePtr(primary *time.Time, fallback *time.Time) *time.Time {
	if primary != nil {
		return primary
	}
	return fallback
}

func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func tenantQuotaValue(db *gorm.DB, tenantID uint64, quotaCode string, fallback int) int {
	var quota models.SaasQuota
	if err := db.Where("quota_code = ?", quotaCode).First(&quota).Error; err != nil {
		return fallback
	}
	now := time.Now()
	var override models.TenantQuotaOverride
	if err := db.Where("tenant_id = ? AND quota_id = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", tenantID, quota.ID, now, now).Order("id desc").First(&override).Error; err == nil {
		return override.QuotaValue
	}
	var sub models.TenantSubscription
	if err := db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return fallback
	}
	var planQuota models.SaasPlanQuota
	if err := db.Where("plan_id = ? AND quota_id = ?", sub.PlanID, quota.ID).First(&planQuota).Error; err == nil {
		return planQuota.QuotaValue
	}
	return fallback
}

func quotaCheckReason(limit int, used int, increment int) string {
	if limit < 0 || used+increment <= limit {
		return "允许使用"
	}
	return "已超出套餐配额"
}

func tenantContact(db *gorm.DB, tenant models.Tenant) (*string, *string) {
	name := tenant.ContactName
	phone := tenant.ContactPhone
	if name != nil && strings.TrimSpace(*name) == "" {
		name = nil
	}
	if phone != nil && strings.TrimSpace(*phone) == "" {
		phone = nil
	}
	if name != nil && phone != nil {
		return name, phone
	}
	if admin, ok := primaryAdminForTenant(db, tenant.ID); ok {
		if name == nil && strings.TrimSpace(admin.Name) != "" {
			name = &admin.Name
		}
		if phone == nil {
			phone = admin.Phone
		}
	}
	return name, phone
}

func primaryAdminForTenant(db *gorm.DB, tenantID uint64) (models.AppUser, bool) {
	var user models.AppUser
	if err := db.
		Joins("JOIN user_role ur ON ur.user_id = app_user.id").
		Joins("JOIN role r ON r.id = ur.role_id").
		Where("app_user.tenant_id = ? AND app_user.deleted_at IS NULL AND r.tenant_id = ? AND r.deleted_at IS NULL AND r.code = ?", tenantID, tenantID, "admin").
		Order("app_user.id asc").
		First(&user).Error; err == nil {
		return user, true
	}
	if err := db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Order("id asc").First(&user).Error; err == nil {
		return user, true
	}
	return models.AppUser{}, false
}

func (h *IdentityHandler) userToJSON(row models.AppUser) gin.H {
	var roles []models.UserRole
	var positions []models.AppUserPosition
	var departments []models.AppUserDepartment
	_ = h.db.Where("user_id = ?", row.ID).Find(&roles).Error
	_ = h.db.Where("user_id = ?", row.ID).Find(&positions).Error
	_ = h.db.Where("user_id = ?", row.ID).Find(&departments).Error
	roleIDs := make([]uint64, 0, len(roles))
	positionIDs := make([]uint64, 0, len(positions))
	departmentIDs := make([]uint64, 0, len(departments))
	for _, item := range roles {
		roleIDs = append(roleIDs, item.RoleID)
	}
	for _, item := range positions {
		positionIDs = append(positionIDs, item.PositionID)
	}
	for _, item := range departments {
		departmentIDs = append(departmentIDs, item.DepartmentID)
	}
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "employee_no": row.EmployeeNo, "phone": row.Phone, "name": row.Name, "email": row.Email, "avatar_url": row.AvatarURL, "status": row.Status, "company_id": row.CompanyID, "department_id": row.DepartmentID, "department_ids": departmentIDs, "position_ids": positionIDs, "role_ids": roleIDs, "is_platform_admin": row.IsPlatformAdmin, "created_at": row.CreatedAt}
}

func roleToJSON(row domainpermission.Role) gin.H {
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "description": row.Description, "permission_ids": row.PermissionIDs, "data_overrides": []gin.H{}}
}

func businessUnitToJSON(row models.BusinessUnit) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "name": row.Name, "code": row.Code, "bu_type": row.BUType, "status": row.Status, "billing_enabled": row.BillingEnabled, "statistic_enabled": row.StatisticEnabled, "remark": row.Remark}
}

func businessUnitOrgMapToJSON(row models.BusinessUnitOrgMap) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "business_unit_id": row.BusinessUnitID, "org_id": row.OrgID, "org_type": row.OrgType, "scope_type": row.ScopeType, "priority": row.Priority, "status": row.Status}
}

func planToJSON(row models.SaasPlan) gin.H {
	return gin.H{
		"id":            row.ID,
		"plan_code":     row.PlanCode,
		"plan_name":     row.PlanName,
		"plan_type":     row.PlanType,
		"billing_cycle": row.BillingCycle,
		"price":         row.Price,
		"status":        row.Status,
		"is_default":    row.IsDefault,
		"sort_order":    row.SortOrder,
		"description":   row.Description,
		"created_at":    row.CreatedAt,
		"updated_at":    row.UpdatedAt,
	}
}

func featureToJSON(row models.SaasFeature) gin.H {
	return gin.H{
		"id":           row.ID,
		"feature_code": row.FeatureCode,
		"feature_name": row.FeatureName,
		"feature_type": row.FeatureType,
		"parent_id":    row.ParentID,
		"menu_id":      row.MenuID,
		"api_method":   row.APIMethod,
		"api_path":     row.APIPath,
		"service_key":  row.ServiceKey,
		"status":       row.Status,
		"description":  row.Description,
	}
}

func quotaToJSON(row models.SaasQuota) gin.H {
	return gin.H{
		"id":          row.ID,
		"quota_code":  row.QuotaCode,
		"quota_name":  row.QuotaName,
		"quota_type":  row.QuotaType,
		"period_type": row.PeriodType,
		"unit":        row.Unit,
		"status":      row.Status,
		"description": row.Description,
	}
}

func permissionToJSON(row models.Permission) gin.H {
	return gin.H{
		"id":                 row.ID,
		"tenant_id":          row.TenantID,
		"parent_id":          row.ParentID,
		"name":               row.Name,
		"path":               row.Path,
		"perm_type":          row.PermType,
		"data_scope":         row.DataScope,
		"sort_order":         row.SortOrder,
		"enabled":            row.Enabled,
		"visible":            row.Visible,
		"is_platform_only":   row.IsPlatformOnly,
		"is_package_feature": row.IsPackageFeature,
		"tenant_editable":    row.TenantEditable,
		"tenant_edit_scope":  row.TenantEditScope,
		"feature_code":       row.FeatureCode,
		"feature_type":       row.FeatureType,
		"data_perm_mode":     row.DataPermMode,
		"created_at":         row.CreatedAt,
		"updated_at":         row.UpdatedAt,
	}
}

func orgNodeToJSON(row models.OrgNode, children []gin.H) gin.H {
	return gin.H{
		"id":           row.ID,
		"node_type":    row.NodeType,
		"name":         row.Name,
		"code":         row.Code,
		"company_type": row.CompanyType,
		"company_id":   row.CompanyID,
		"store_id":     nil,
		"parent_id":    row.ParentID,
		"status":       row.Status,
		"children":     children,
	}
}

func dictItemsToJSON(rows []models.DictItem) []gin.H {
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":                 row.ID,
			"dict_type_id":       row.DictTypeID,
			"default_label":      row.Label,
			"default_value":      row.Value,
			"default_sort_order": row.SortOrder,
			"default_enabled":    row.Enabled,
			"label":              row.Label,
			"value":              row.Value,
			"item_label":         row.Label,
			"item_value":         row.Value,
			"sort_order":         row.SortOrder,
			"enabled":            row.Enabled,
			"status":             boolToStatus(row.Enabled),
			"is_override":        false,
		})
	}
	return items
}

func boolToStatus(value bool) int {
	if value {
		return 1
	}
	return 0
}

func parseUintParam(c *gin.Context, name string) uint64 {
	value, _ := strconv.ParseUint(c.Param(name), 10, 64)
	return value
}

func parseTenantID(c *gin.Context) uint64 {
	return parseTenantIDValue(c.Query("tenant_id"))
}

func parseTenantIDValue(raw string) uint64 {
	if raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return 1
}

func effectiveTenantID(queryTenantID string, userTenantID uint64, isPlatformAdmin bool) uint64 {
	if isPlatformAdmin {
		return parseTenantIDValue(queryTenantID)
	}
	if userTenantID > 0 {
		return userTenantID
	}
	return parseTenantIDValue(queryTenantID)
}

func (h *IdentityHandler) requestTenantID(c *gin.Context) uint64 {
	user, ok := h.currentUser(c)
	if !ok {
		return parseTenantID(c)
	}
	return effectiveTenantID(c.Query("tenant_id"), user.TenantID, user.IsPlatformAdmin)
}

func (h *IdentityHandler) currentUser(c *gin.Context) (models.AppUser, bool) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		return models.AppUser{}, false
	}
	userID := uint64(0)
	if h.redis != nil {
		raw, err := h.redis.Get(context.Background(), authSessionKey(token)).Result()
		if err == nil && raw != "" {
			var session struct {
				UserID uint64 `json:"user_id"`
			}
			if json.Unmarshal([]byte(raw), &session) == nil {
				userID = session.UserID
			}
		}
	}
	if userID == 0 {
		claims, err := parseToken(token, h.authSecret)
		if err != nil || claims.UserID == 0 {
			return models.AppUser{}, false
		}
		userID = claims.UserID
	}
	var user models.AppUser
	if err := h.db.Where("deleted_at IS NULL").First(&user, userID).Error; err != nil {
		return models.AppUser{}, false
	}
	return user, true
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
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return true
	}
	status := strings.ToUpper(sub.SubscriptionStatus)
	if status == "OVERDUE" || status == "FROZEN" || status == "EXPIRED" || status == "CANCELLED" {
		return false
	}
	if status != "" && status != "TRIAL" && status != "ACTIVE" {
		return false
	}
	return sub.EndTime == nil || sub.EndTime.After(time.Now())
}

func (h *IdentityHandler) recordLogin(c *gin.Context, account string, userID *uint64, tenantID *uint64, success bool, message string) {
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	_ = h.db.Create(&models.LoginLog{
		TenantID:  tenantID,
		UserID:    userID,
		Account:   account,
		Success:   success,
		Message:   &message,
		IP:        &ip,
		UserAgent: &userAgent,
	}).Error
}

func (h *IdentityHandler) audit(c *gin.Context, tenantID uint64, userID uint64, module, action, summary string, detail interface{}) {
	detailJSON := ""
	if detail != nil {
		if raw, err := json.Marshal(detail); err == nil {
			detailJSON = string(raw)
		}
	}
	ip := c.ClientIP()
	_ = h.db.Create(&models.AuditLog{
		TenantID: &tenantID,
		UserID:   &userID,
		Module:   module,
		Action:   action,
		Summary:  summary,
		Detail:   nullableFromString(detailJSON),
		IP:       &ip,
	}).Error
}

func (h *IdentityHandler) auditCurrentUser(c *gin.Context, module, action, summary string, detail interface{}) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	h.audit(c, user.TenantID, user.ID, module, action, summary, detail)
}

type auditRequiredEvent struct {
	Module string
	Action string
}

func requiredAuditEvents() []auditRequiredEvent {
	return []auditRequiredEvent{
		{Module: "tenant", Action: "create"},
		{Module: "tenant", Action: "create_with_package"},
		{Module: "tenant", Action: "update"},
		{Module: "tenant", Action: "status_update"},
		{Module: "tenant", Action: "delete"},
		{Module: "user", Action: "create"},
		{Module: "user", Action: "update"},
		{Module: "user", Action: "password_reset"},
		{Module: "user", Action: "delete"},
		{Module: "role", Action: "create"},
		{Module: "role", Action: "update"},
		{Module: "role", Action: "delete"},
		{Module: "organization", Action: "create"},
		{Module: "organization", Action: "update"},
		{Module: "organization", Action: "delete"},
		{Module: "business_unit", Action: "create"},
		{Module: "business_unit", Action: "update"},
		{Module: "business_unit", Action: "delete"},
		{Module: "business_unit", Action: "org_mapping_create"},
		{Module: "business_unit", Action: "org_mapping_delete"},
	}
}

type deletionReference struct {
	Model interface{}
	Name  string
	Query string
	Args  []interface{}
}

func ref(model interface{}, name, query string, args ...interface{}) deletionReference {
	return deletionReference{Model: model, Name: name, Query: query, Args: args}
}

func (h *IdentityHandler) blockDeleteIfReferenced(c *gin.Context, resource string, refs ...deletionReference) bool {
	for _, item := range refs {
		var count int64
		if err := h.db.Model(item.Model).Where(item.Query, item.Args...).Count(&count).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return true
		}
		if count > 0 {
			response.Error(c, 400, response.CodeBadRequest, fmt.Sprintf("%s已被%s引用，不能删除", resource, item.Name))
			return true
		}
	}
	return false
}

func requiredDeletionGuards() []string {
	return []string{
		"permission:role_permission",
		"tenant:app_user",
		"plan:tenant_subscription",
		"org_node:app_user",
		"org_node:business_unit_org_map",
		"position_type:position",
		"position:app_user_position",
		"business_unit:business_unit_org_map",
		"dict_type:dict_item",
		"dict_item:tenant_dict_item_override",
		"sys_param:tenant_param_value",
	}
}

func (h *IdentityHandler) deleteByID(c *gin.Context, model interface{}) bool {
	id := parseUintParam(c, "id")
	if err := h.db.Delete(model, id).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return false
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
	return true
}

func (h *IdentityHandler) deleteTenantScopedByID(c *gin.Context, model interface{}) bool {
	id := parseUintParam(c, "id")
	result := h.db.Where("tenant_id = ?", h.requestTenantID(c)).Delete(model, id)
	if result.Error != nil {
		response.Error(c, 400, response.CodeBadRequest, result.Error.Error())
		return false
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "数据不存在")
		return false
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
	return true
}

func nullableTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	return nullableFromString(*value)
}

func nullableFromString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalPhone(value *string) (*string, string) {
	if value == nil {
		return nil, ""
	}
	raw := strings.TrimSpace(*value)
	if raw == "" {
		return nil, ""
	}
	var digits strings.Builder
	for _, ch := range raw {
		switch {
		case ch >= '0' && ch <= '9':
			digits.WriteRune(ch)
		case ch == ' ' || ch == '-' || ch == '(' || ch == ')':
		default:
			return nil, "手机号只能包含数字、空格、短横线或括号"
		}
	}
	normalized := digits.String()
	if len(normalized) < 10 || len(normalized) > 15 {
		return nil, "手机号需为 10-15 位数字"
	}
	return &normalized, ""
}

func uniqueUint64s(values []uint64) []uint64 {
	if values == nil {
		return nil
	}
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizedUserDepartmentIDs(departmentID *uint64, departmentIDs []uint64) []uint64 {
	ids := uniqueUint64s(departmentIDs)
	if departmentID == nil || *departmentID == 0 {
		return ids
	}
	for _, id := range ids {
		if id == *departmentID {
			return ids
		}
	}
	return append([]uint64{*departmentID}, ids...)
}

func tombstoneUniqueValue(value string, id uint64, maxLen int) string {
	suffix := fmt.Sprintf("__deleted_%d_%d", id, time.Now().Unix())
	base := value
	if len(base)+len(suffix) > maxLen {
		keep := maxLen - len(suffix)
		if keep < 0 {
			keep = 0
		}
		if len(base) > keep {
			base = base[:keep]
		}
	}
	return base + suffix
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func valueOrZero(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func randomCode(size int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "ABCD"
	}
	out := make([]byte, size)
	for i, b := range buf {
		out[i] = chars[int(b)%len(chars)]
	}
	return string(out)
}

func uploadRoot() string {
	root := os.Getenv("UPLOAD_DIR")
	if root == "" {
		root = "uploads"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}

func uploadDir(tenantID uint64, now time.Time) string {
	return filepath.Join(uploadRoot(), strconv.FormatUint(tenantID, 10), now.Format("2006/01/02"))
}

func safeFileExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) == 0 || len(ext) > 16 {
		return ".bin"
	}
	for _, r := range ext[1:] {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') {
			return ".bin"
		}
	}
	return ext
}

func (h *IdentityHandler) findTenantFile(tenantID uint64, fileID string) (string, error) {
	if !safeFileIDPattern.MatchString(fileID) {
		return "", fmt.Errorf("invalid file_id")
	}
	root := filepath.Join(uploadRoot(), strconv.FormatUint(tenantID, 10))
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".trash" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), fileID+".") {
			found = path
			return io.EOF
		}
		return nil
	})
	if found != "" {
		return found, nil
	}
	if err != nil && err != io.EOF {
		return "", err
	}
	return "", fmt.Errorf("not found")
}

func csvHeaderIndex(headers []string) map[string]int {
	index := map[string]int{}
	for i, header := range headers {
		index[strings.TrimSpace(header)] = i
	}
	return index
}

func csvCell(record []string, index map[string]int, key string) string {
	i, ok := index[key]
	if !ok || i < 0 || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}

func csvSafe(value string) string {
	if strings.HasPrefix(value, "=") || strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") || strings.HasPrefix(value, "@") {
		return "'" + value
	}
	return value
}

func (h *IdentityHandler) sendCSV(c *gin.Context, filename string, rows [][]string) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	for _, row := range rows {
		safe := make([]string, len(row))
		for i, cell := range row {
			safe[i] = csvSafe(cell)
		}
		_ = writer.Write(safe)
	}
	writer.Flush()
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *IdentityHandler) orgNameMaps(tenantID uint64) (map[uint64]string, map[uint64]string) {
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ?", tenantID).Find(&rows).Error
	companies := map[uint64]string{}
	depts := map[uint64]string{}
	for _, row := range rows {
		switch row.NodeType {
		case "company":
			companies[row.ID] = row.Name
		case "department":
			depts[row.ID] = row.Name
		}
	}
	return companies, depts
}

func (h *IdentityHandler) exportOrgCSV(c *gin.Context, filename, nodeType string, headers []string) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND node_type = ?", user.TenantID, nodeType).Order("id asc").Find(&rows).Error
	var all []models.OrgNode
	_ = h.db.Where("tenant_id = ?", user.TenantID).Find(&all).Error
	names := map[uint64]models.OrgNode{}
	for _, row := range all {
		names[row.ID] = row
	}
	out := [][]string{headers}
	for _, row := range rows {
		parentName := ""
		if row.ParentID != nil {
			parentName = names[*row.ParentID].Name
		}
		status := "启用"
		if row.Status != 1 {
			status = "停用"
		}
		if nodeType == "company" {
			out = append(out, []string{row.Name, derefString(row.Code), derefString(row.CompanyType), parentName, status})
		} else {
			companyName := ""
			if row.CompanyID != nil {
				companyName = names[*row.CompanyID].Name
			}
			out = append(out, []string{row.Name, derefString(row.Code), companyName, parentName, status})
		}
	}
	h.sendCSV(c, filename, out)
}

func devPasswordHash(password string) string {
	return mustHashPassword(password)
}

func (h *IdentityHandler) replaceUserRelations(userID uint64, roleIDs []uint64, positionIDs []uint64, departmentIDs []uint64) {
	if roleIDs != nil {
		_ = h.db.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error
		for _, id := range roleIDs {
			_ = h.db.Create(&models.UserRole{UserID: userID, RoleID: id}).Error
		}
	}
	if positionIDs != nil {
		_ = h.db.Where("user_id = ?", userID).Delete(&models.AppUserPosition{}).Error
		for _, id := range positionIDs {
			_ = h.db.Create(&models.AppUserPosition{UserID: userID, PositionID: id}).Error
		}
	}
	if departmentIDs != nil {
		_ = h.db.Where("user_id = ?", userID).Delete(&models.AppUserDepartment{}).Error
		for _, id := range departmentIDs {
			_ = h.db.Create(&models.AppUserDepartment{UserID: userID, DepartmentID: id}).Error
		}
	}
}

func (h *IdentityHandler) assertDepartmentsInTenant(tenantID uint64, departmentIDs []uint64) error {
	ids := uniqueUint64s(departmentIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.OrgNode{}).
		Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL AND node_type IN ?", tenantID, ids, []string{"department", "store", "warehouse", "project_team", "group"}).
		Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("组织节点不存在、类型不可用于人员归属或不属于当前主体")
	}
	return nil
}

func (h *IdentityHandler) companyIDForDepartment(tenantID uint64, departmentID uint64) *uint64 {
	var row models.OrgNode
	if err := h.db.Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, departmentID).First(&row).Error; err != nil {
		return nil
	}
	if row.CompanyID != nil {
		return row.CompanyID
	}
	if row.NodeType == "company" {
		return &row.ID
	}
	return nil
}

func (h *IdentityHandler) assertPositionsInTenant(tenantID uint64, positionIDs []uint64) error {
	ids := uniqueUint64s(positionIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.Position{}).Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("岗位不存在或不属于当前主体")
	}
	return nil
}

func (h *IdentityHandler) assertRolesInTenant(tenantID uint64, roleIDs []uint64) error {
	ids := uniqueUint64s(roleIDs)
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.Role{}).Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return fmt.Errorf("角色不存在或不属于当前主体")
	}
	return nil
}

func (h *IdentityHandler) assertUserUniqueFields(tenantID uint64, exceptUserID uint64, employeeNo string, phone *string) error {
	if strings.TrimSpace(employeeNo) != "" {
		query := h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND employee_no = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(employeeNo))
		if exceptUserID > 0 {
			query = query.Where("id <> ?", exceptUserID)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("工号已存在")
		}
	}
	if phone != nil && strings.TrimSpace(*phone) != "" {
		query := h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND phone = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(*phone))
		if exceptUserID > 0 {
			query = query.Where("id <> ?", exceptUserID)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("手机号已存在")
		}
	}
	return nil
}

func (h *IdentityHandler) createTenantWithAdmin(code, name string, status int, adminName, adminEmployeeNo string, adminPhone *string, adminPassword string) (models.Tenant, error) {
	var tenant models.Tenant
	err := h.db.Transaction(func(tx *gorm.DB) error {
		created, err := h.createTenantWithAdminOnDB(tx, code, name, status, adminName, adminEmployeeNo, adminPhone, adminPassword)
		if err != nil {
			return err
		}
		tenant = created
		return nil
	})
	return tenant, err
}

func (h *IdentityHandler) createTenantWithAdminOnDB(db *gorm.DB, code, name string, status int, adminName, adminEmployeeNo string, adminPhone *string, adminPassword string) (models.Tenant, error) {
	if status == 0 {
		status = 1
	}
	phone, msg := normalizeOptionalPhone(adminPhone)
	if msg != "" {
		return models.Tenant{}, errors.New(strings.Replace(msg, "手机号", "管理员手机号", 1))
	}
	var tenant models.Tenant
	tenant = models.Tenant{Code: strings.TrimSpace(code), Name: strings.TrimSpace(name), Status: status, ContactName: nullableFromString(adminName), ContactPhone: phone}
	if err := db.Create(&tenant).Error; err != nil {
		return tenant, err
	}
	companyCode := strings.ToUpper(strings.TrimSpace(code))
	companyType := "GROUP"
	company := models.OrgNode{TenantID: tenant.ID, NodeType: "company", Name: tenant.Name, Code: &companyCode, CompanyType: &companyType, Status: 1}
	if err := db.Create(&company).Error; err != nil {
		return tenant, err
	}
	roleDescription := "超级管理员（系统自动创建）"
	role := models.Role{TenantID: tenant.ID, Code: "admin", Name: "超级管理员", Description: &roleDescription, Status: 1}
	if err := db.Create(&role).Error; err != nil {
		return tenant, err
	}
	user := models.AppUser{TenantID: tenant.ID, CompanyID: &company.ID, EmployeeNo: strings.TrimSpace(adminEmployeeNo), Account: strings.TrimSpace(adminEmployeeNo), PasswordHash: devPasswordHash(adminPassword), Name: strings.TrimSpace(adminName), Phone: phone, Status: 1, IsPlatformAdmin: false}
	if err := db.Create(&user).Error; err != nil {
		return tenant, err
	}
	if err := db.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		return tenant, err
	}
	var permissions []models.Permission
	if err := db.Where("tenant_id = ? OR tenant_id = ?", tenant.ID, 1).Find(&permissions).Error; err == nil {
		for _, permission := range permissions {
			if err := db.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).FirstOrCreate(&models.RolePermission{RoleID: role.ID, PermissionID: permission.ID}).Error; err != nil {
				return tenant, err
			}
		}
	}
	return tenant, nil
}

func parseTimePtr(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, *value); err == nil {
			return &parsed
		}
	}
	return nil
}

func (h *IdentityHandler) saveTenantPackage(tenantID uint64, body tenantPackagePayload) error {
	return h.saveTenantPackageOnDB(h.db, tenantID, body)
}

func (h *IdentityHandler) saveTenantPackageOnDB(db *gorm.DB, tenantID uint64, body tenantPackagePayload) error {
	start := time.Now()
	if body.StartTime != "" {
		if parsed := parseTimePtr(&body.StartTime); parsed != nil {
			start = *parsed
		}
	}
	if body.SubscriptionStatus == "" {
		body.SubscriptionStatus = "ACTIVE"
	}
	if body.PlanID == 0 {
		return errors.New("套餐不存在")
	}
	var existing models.TenantSubscription
	sub := models.TenantSubscription{TenantID: tenantID, PlanID: body.PlanID, SubscriptionStatus: body.SubscriptionStatus, StartTime: start, EndTime: parseTimePtr(body.EndTime), TrialEndTime: parseTimePtr(body.TrialEndTime), AutoRenew: body.AutoRenew, FrozenReason: body.FrozenReason}
	if err := db.Where("tenant_id = ?", tenantID).First(&existing).Error; err == nil {
		if err := db.Model(&existing).Updates(sub).Error; err != nil {
			return err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&sub).Error; err != nil {
			return err
		}
	} else {
		return err
	}
	for _, quota := range body.Quotas {
		_ = db.Where("tenant_id = ? AND quota_id = ?", tenantID, quota.QuotaID).Delete(&models.TenantQuotaOverride{}).Error
		if err := db.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue, Reason: stringPtrLocal("主体套餐配置")}).Error; err != nil {
			return err
		}
	}
	return nil
}

func stringPtrLocal(value string) *string {
	return &value
}

func (h *IdentityHandler) findTenantSubscription(tenantID uint64) (models.TenantSubscription, bool) {
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return sub, false
	}
	return sub, true
}

func tenantSubscriptionToJSON(row models.TenantSubscription) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "plan_id": row.PlanID, "subscription_status": row.SubscriptionStatus, "start_time": row.StartTime, "end_time": row.EndTime, "trial_end_time": row.TrialEndTime, "auto_renew": row.AutoRenew, "frozen_reason": row.FrozenReason, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func (h *IdentityHandler) primaryAdmin(tenantID uint64) (models.AppUser, bool) {
	var user models.AppUser
	err := h.db.Where("tenant_id = ?", tenantID).Order("is_platform_admin desc, id asc").First(&user).Error
	return user, err == nil
}

func (h *IdentityHandler) tenantFeatureOverridesPayload(tenantID uint64) gin.H {
	var rows []models.TenantFeatureOverride
	_ = h.db.Where("tenant_id = ?", tenantID).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var feature models.SaasFeature
		_ = h.db.First(&feature, row.FeatureID).Error
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "feature_id": row.FeatureID, "feature_code": feature.FeatureCode, "feature_name": feature.FeatureName, "enabled": row.Enabled, "reason": row.Reason, "start_time": row.StartTime, "end_time": row.EndTime})
	}
	return gin.H{"tenant_id": tenantID, "overrides": items}
}

func (h *IdentityHandler) tenantQuotaOverridesPayload(tenantID uint64) gin.H {
	var rows []models.TenantQuotaOverride
	_ = h.db.Where("tenant_id = ?", tenantID).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var quota models.SaasQuota
		_ = h.db.First(&quota, row.QuotaID).Error
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "quota_id": row.QuotaID, "quota_code": quota.QuotaCode, "quota_name": quota.QuotaName, "quota_value": row.QuotaValue, "period_type": quota.PeriodType, "unit": quota.Unit, "reason": row.Reason, "start_time": row.StartTime, "end_time": row.EndTime})
	}
	return gin.H{"tenant_id": tenantID, "overrides": items}
}

func (h *IdentityHandler) createOrgNode(c *gin.Context, forcedType string) {
	var body struct {
		NodeType    string  `json:"node_type"`
		Name        string  `json:"name"`
		Code        *string `json:"code"`
		CompanyType *string `json:"company_type"`
		CompanyID   *uint64 `json:"company_id"`
		ParentID    *uint64 `json:"parent_id"`
		Status      int     `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if forcedType != "" {
		body.NodeType = forcedType
	}
	if body.NodeType == "" {
		body.NodeType = "department"
	}
	if body.Status == 0 {
		body.Status = 1
	}
	row := models.OrgNode{TenantID: h.requestTenantID(c), NodeType: body.NodeType, Name: body.Name, Code: body.Code, CompanyType: body.CompanyType, CompanyID: body.CompanyID, ParentID: body.ParentID, Status: body.Status}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.auditCurrentUser(c, "organization", "create", "创建组织 "+row.Name, gin.H{"org_id": row.ID, "tenant_id": row.TenantID, "node_type": row.NodeType, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) updateOrgNode(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	result := h.db.Model(&models.OrgNode{}).Where("id = ? AND tenant_id = ?", c.Param("id"), h.requestTenantID(c)).Updates(body)
	if result.Error != nil {
		response.Error(c, 400, response.CodeBadRequest, result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "组织不存在")
		return
	}
	h.auditCurrentUser(c, "organization", "update", "更新组织", gin.H{"org_id": parseUintParam(c, "id"), "changes": body})
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) deleteOrgNodeWithAudit(c *gin.Context, summary string) {
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "组织", ref(&models.OrgNode{}, "下级组织", "parent_id = ? OR company_id = ?", id, id), ref(&models.AppUser{}, "用户主组织", "company_id = ? OR department_id = ?", id, id), ref(&models.AppUserDepartment{}, "用户兼任部门", "department_id = ?", id), ref(&models.BusinessUnitOrgMap{}, "业务单元组织映射", "org_id = ?", id)) {
		return
	}
	if h.deleteTenantScopedByID(c, &models.OrgNode{}) {
		h.auditCurrentUser(c, "organization", "delete", summary, gin.H{"org_id": id})
	}
}

func (h *IdentityHandler) replaceBusinessUnitMappings(tenantID, buID uint64, orgIDs []uint64) {
	_ = h.db.Where("tenant_id = ? AND business_unit_id = ?", tenantID, buID).Delete(&models.BusinessUnitOrgMap{}).Error
	for _, orgID := range orgIDs {
		_ = h.db.Create(&models.BusinessUnitOrgMap{TenantID: tenantID, BusinessUnitID: buID, OrgID: orgID, OrgType: "org", ScopeType: "include", Status: 1}).Error
	}
}

func (h *IdentityHandler) tenantName(tenantID *uint64) *string {
	if tenantID == nil {
		return nil
	}
	var tenant models.Tenant
	if err := h.db.First(&tenant, *tenantID).Error; err != nil {
		return nil
	}
	return &tenant.Name
}

func allDevPermissionCodes() []string {
	return []string{
		"tenant:create", "tenant:edit", "tenant:delete", "tenant:reset_password", "tenant:quota_config",
		"plan:create", "plan:edit", "plan:delete", "plan:config",
		"org:create", "org:edit", "org:delete",
		"pos:create", "pos:edit", "pos:delete",
		"business_unit:create", "business_unit:edit", "business_unit:delete",
		"user:create", "user:edit", "user:reset_password", "user:delete",
		"role:create", "role:edit", "role:delete",
		"menu:create", "menu:edit", "menu:delete",
		"dict:create", "dict:edit", "dict:delete",
		"param:create", "param:edit", "param:delete",
		"audit:view", "login:view", "brand:edit",
	}
}
