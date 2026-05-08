package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Profile(c *gin.Context) {
	var user models.AppUser
	var tenant models.Tenant
	var roles []models.Role
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
	}
	roleIDs := make([]uint64, 0, len(roles))
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
		roleCodes = append(roleCodes, role.Code)
	}
	capability := h.tenantCapabilityContext(user.TenantID)
	tenantIsPlatform := tenant.IsPlatform
	permissionCodes := h.permissionCodesForUser(user, !h.viewerHasPlatformScope(user))

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
			respondBadRequest(c, err)
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
	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"password_hash":       hashed,
		"session_version":     gorm.Expr("session_version + 1"),
		"password_changed_at": time.Now(),
	}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.clearPasswordChangeGuard(user.ID)
	h.invalidateSessionsForUser(user.ID)
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
		respondBadRequest(c, err)
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
			respondBadRequest(c, err)
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
