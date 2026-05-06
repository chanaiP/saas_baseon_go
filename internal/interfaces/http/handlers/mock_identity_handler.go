package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

type MockIdentityHandler struct {
	db *gorm.DB
}

func NewMockIdentityHandler(db *gorm.DB) *MockIdentityHandler {
	return &MockIdentityHandler{db: db}
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
	var user models.AppUser
	var tenant models.Tenant
	var roles []models.Role
	var permissions []models.Permission
	if err := h.db.Where("account = ?", "admin").First(&user).Error; err == nil {
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
		"tenant_is_platform": tenant.IsPlatform,
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
	var tenant models.Tenant
	_ = h.db.Where("code = ?", "platform").First(&tenant).Error
	response.OK(c, gin.H{
		"display_name":       coalesceStringPtr(tenant.BrandName, "Ai DevOS"),
		"brand_display_name": tenant.BrandName,
		"logo_data":          nil,
		"footer_text":        tenant.FooterText,
		"tenant_name":        coalesceString(tenant.Name, "平台主体"),
		"can_edit":           true,
		"can_edit_footer":    true,
	})
}

func (h *MockIdentityHandler) SaveTenantBranding(c *gin.Context) {
	h.TenantBranding(c)
}

func (h *MockIdentityHandler) PublicTenantFooter(c *gin.Context) {
	var tenant models.Tenant
	_ = h.db.Where("code = ?", "platform").First(&tenant).Error
	response.OK(c, gin.H{"footer_text": tenant.FooterText})
}

func (h *MockIdentityHandler) MenuBundles(c *gin.Context) {
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

func (h *MockIdentityHandler) MenuOverrides(c *gin.Context) {
	response.OK(c, gin.H{"tenant_id": 1, "overrides": []gin.H{}})
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
