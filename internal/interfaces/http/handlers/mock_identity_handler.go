package handlers

import (
	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/interfaces/http/response"
)

type MockIdentityHandler struct{}

func NewMockIdentityHandler() *MockIdentityHandler {
	return &MockIdentityHandler{}
}

func (h *MockIdentityHandler) Login(c *gin.Context) {
	response.OK(c, gin.H{
		"token":            "dev-token",
		"token_type":       "bearer",
		"captcha_required": false,
	})
}

func (h *MockIdentityHandler) Logout(c *gin.Context) {
	response.OK(c, gin.H{})
}

func (h *MockIdentityHandler) Captcha(c *gin.Context) {
	response.OK(c, gin.H{
		"captcha_id":   "dev-captcha",
		"image_base64": "",
	})
}

func (h *MockIdentityHandler) PhoneLoginTenants(c *gin.Context) {
	response.OK(c, []gin.H{})
}

func (h *MockIdentityHandler) SwitchableTenants(c *gin.Context) {
	response.OK(c, []gin.H{})
}

func (h *MockIdentityHandler) SwitchTenant(c *gin.Context) {
	h.Login(c)
}

func (h *MockIdentityHandler) Profile(c *gin.Context) {
	response.OK(c, gin.H{
		"id":                 1,
		"tenant_id":          1,
		"employee_no":        "admin",
		"phone":              nil,
		"name":               "平台管理员",
		"email":              nil,
		"avatar_url":         nil,
		"status":             1,
		"company_id":         nil,
		"department_id":      nil,
		"role_ids":           []int{1},
		"role_codes":         []string{"admin"},
		"permission_codes":   allDevPermissionCodes(),
		"is_platform_admin":  true,
		"tenant_is_platform": true,
		"shortcut_ids":       []string{},
		"subscription":       nil,
		"features":           []string{},
		"quotas":             gin.H{},
	})
}

func (h *MockIdentityHandler) UpdateProfile(c *gin.Context) {
	h.Profile(c)
}

func (h *MockIdentityHandler) UpdatePassword(c *gin.Context) {
	response.OK(c, gin.H{})
}

func (h *MockIdentityHandler) Preferences(c *gin.Context) {
	response.OK(c, gin.H{"shortcut_ids": []string{}})
}

func (h *MockIdentityHandler) SavePreferences(c *gin.Context) {
	var body struct {
		ShortcutIDs []string `json:"shortcut_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	response.OK(c, gin.H{"shortcut_ids": body.ShortcutIDs})
}

func (h *MockIdentityHandler) TenantBranding(c *gin.Context) {
	response.OK(c, gin.H{
		"display_name":       "Ai DevOS",
		"brand_display_name": nil,
		"logo_data":          nil,
		"footer_text":        "© 2026 SaaS - AI协作开发系统",
		"tenant_name":        "平台主体",
		"can_edit":           true,
		"can_edit_footer":    true,
	})
}

func (h *MockIdentityHandler) SaveTenantBranding(c *gin.Context) {
	h.TenantBranding(c)
}

func (h *MockIdentityHandler) PublicTenantFooter(c *gin.Context) {
	response.OK(c, gin.H{"footer_text": "© 2026 SaaS - AI协作开发系统"})
}

func (h *MockIdentityHandler) MenuBundles(c *gin.Context) {
	response.OK(c, []gin.H{})
}

func (h *MockIdentityHandler) MenuOverrides(c *gin.Context) {
	response.OK(c, gin.H{"tenant_id": 1, "overrides": []gin.H{}})
}

func (h *MockIdentityHandler) SaveMenuOverrides(c *gin.Context) {
	h.MenuOverrides(c)
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
