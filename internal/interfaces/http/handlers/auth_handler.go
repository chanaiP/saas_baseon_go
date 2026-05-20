package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

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
	if h.loginWithAccount(c, body.Account, body.Password, body.TenantID, body.TenantCode) {
		return
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
			c.JSON(200, response.Body{Code: 2, Message: "请选择空间", Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": false, "tenants": tenantOptions}})
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
		c.JSON(200, response.Body{Code: 2, Message: "请选择空间", Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": false, "tenants": h.activeLoginTenantOptions(matched)}})
		return
	}
	user := matched[0]
	h.backfillAccountForLegacyUser(&user, body.Password)
	h.finishLogin(c, body.Account, user)
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	if token := bearerToken(c.GetHeader("Authorization")); token != "" && h.redis != nil {
		_ = h.redis.Del(context.Background(), authSessionKey(token)).Err()
	}
	response.OK(c, gin.H{})
}

func (h *IdentityHandler) Captcha(c *gin.Context) {
	if message, blocked := h.rateLimitExceeded(c, "rate:captcha:ip:"+c.ClientIP(), 30, time.Minute); blocked {
		c.JSON(429, response.Body{Code: 42900, Message: message})
		return
	}
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
		response.Error(c, 400, response.CodeBadRequest, "已在当前空间")
		return
	}
	if user.Phone == nil || strings.TrimSpace(*user.Phone) == "" {
		response.Error(c, 400, response.CodeBadRequest, "不可切换到该空间（需为同一手机号下的账号）")
		return
	}
	var target models.AppUser
	if err := h.db.Where("tenant_id = ? AND phone = ? AND status = ?", body.TenantID, *user.Phone, 1).Order("id asc").First(&target).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, "不可切换到该空间（需为同一手机号下的账号）")
		return
	}
	var tenant models.Tenant
	if err := h.db.First(&tenant, target.TenantID).Error; err != nil || tenant.Status != 1 {
		response.Error(c, 400, response.CodeBadRequest, "空间不可用")
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
	h.recordLogin(c, coalesceString(user.EmployeeNo, "switch"), &target.ID, &target.TenantID, true, "切换空间登录")
	response.OK(c, gin.H{"token": token, "token_type": "bearer", "captcha_required": false})
}
