package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
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

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
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
	if h.loginFailCount(c, body.Account) >= 3 {
		if body.CaptchaID == "" || body.CaptchaCode == "" {
			response.OK(c, gin.H{"token": nil, "token_type": "bearer", "captcha_required": true})
			return
		}
		if !h.verifyCaptcha(c, body.CaptchaID, body.CaptchaCode) {
			h.incrLoginFail(c, body.Account)
			h.recordLogin(c, body.Account, nil, nil, false, "验证码错误")
			response.Error(c, 400, response.CodeBadRequest, "验证码错误")
			return
		}
	}
	var user models.AppUser
	query := h.db.Where("(account = ? OR employee_no = ? OR phone = ?) AND status = ?", body.Account, body.Account, body.Account, 1)
	if body.TenantID != nil {
		query = query.Where("tenant_id = ?", *body.TenantID)
	}
	if body.TenantCode != "" {
		query = query.Joins("JOIN tenant t ON t.id = app_user.tenant_id AND t.code = ?", body.TenantCode)
	}
	if err := query.Order("is_platform_admin desc, id asc").First(&user).Error; err != nil {
		h.incrLoginFail(c, body.Account)
		h.recordLogin(c, body.Account, nil, nil, false, "账号或密码错误")
		response.Error(c, 401, response.CodeUnauthorized, "账号或密码错误")
		return
	}
	if !verifyPassword(body.Password, user.PasswordHash) {
		h.incrLoginFail(c, body.Account)
		h.recordLogin(c, body.Account, &user.ID, &user.TenantID, false, "账号或密码错误")
		response.Error(c, 401, response.CodeUnauthorized, "账号或密码错误")
		return
	}
	if !h.subscriptionAllowsLogin(user.TenantID) {
		h.recordLogin(c, body.Account, &user.ID, &user.TenantID, false, "账户已到期")
		response.Error(c, 400, response.CodeBadRequest, "账户已到期，请联系管理员续费")
		return
	}
	token, err := issueToken(user.ID, user.TenantID, h.authSecret, h.tokenTTL)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "令牌生成失败")
		return
	}
	h.resetLoginFail(c, body.Account)
	h.recordLogin(c, body.Account, &user.ID, &user.TenantID, true, "登录成功")
	response.OK(c, gin.H{
		"token":            token,
		"token_type":       "bearer",
		"captcha_required": false,
	})
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	response.OK(c, gin.H{})
}

func (h *IdentityHandler) Captcha(c *gin.Context) {
	code := randomCode(4)
	id := randomHex(8)
	if h.redis != nil {
		_ = h.redis.Set(context.Background(), "captcha:"+id, strings.ToUpper(code), 5*time.Minute).Err()
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
	items := make([]gin.H, 0, len(rows))
	for _, user := range rows {
		var tenant models.Tenant
		if err := h.db.First(&tenant, user.TenantID).Error; err == nil && tenant.Status == 1 {
			items = append(items, gin.H{"tenant_id": tenant.ID, "code": tenant.Code, "name": tenant.Name})
		}
	}
	response.OK(c, items)
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
		if err := h.db.First(&tenant, row.TenantID).Error; err == nil && tenant.Status == 1 {
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
	tenantID := user.TenantID
	if body.TenantID > 0 {
		tenantID = body.TenantID
	}
	token, err := issueToken(user.ID, tenantID, h.authSecret, h.tokenTTL)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "令牌生成失败")
		return
	}
	response.OK(c, gin.H{"token": token, "token_type": "bearer", "captcha_required": false})
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
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.NewPassword == "" || body.NewPassword != body.NewPasswordConfirm {
		response.Error(c, 400, response.CodeBadRequest, "两次输入的新密码不一致")
		return
	}
	if !verifyPassword(body.OldPassword, user.PasswordHash) {
		response.Error(c, 400, response.CodeBadRequest, "当前密码错误")
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
	response.OK(c, gin.H{})
}

func (h *IdentityHandler) Preferences(c *gin.Context) {
	response.OK(c, gin.H{"shortcut_ids": []string{}})
}

func (h *IdentityHandler) SavePreferences(c *gin.Context) {
	var body struct {
		ShortcutIDs []string `json:"shortcut_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	response.OK(c, gin.H{"shortcut_ids": body.ShortcutIDs})
}

func (h *IdentityHandler) TenantBranding(c *gin.Context) {
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
	updates := map[string]interface{}{}
	if body.BrandDisplayName != nil {
		updates["brand_display_name"] = nullableTrimmed(body.BrandDisplayName)
	}
	if body.BrandLogoData != nil {
		updates["brand_logo_data"] = nullableTrimmed(body.BrandLogoData)
	} else if body.LogoData != nil {
		updates["brand_logo_data"] = nullableTrimmed(body.LogoData)
	}
	if body.BrandFooterText != nil {
		updates["brand_footer_text"] = nullableTrimmed(body.BrandFooterText)
	} else if body.FooterText != nil {
		updates["brand_footer_text"] = nullableTrimmed(body.FooterText)
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
	_ = h.db.Where("code = ?", "platform").First(&tenant).Error
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
	var rows []models.Tenant
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, tenantToJSON(h.db, row))
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) Tenant(c *gin.Context) {
	var row models.Tenant
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
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
	tenant, err := h.createTenantWithAdmin(body.Tenant.Code, body.Tenant.Name, body.Tenant.Status, body.Tenant.AdminName, body.Tenant.AdminEmployeeNo, body.Tenant.AdminPhone, body.Tenant.AdminPassword)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	if err := h.saveTenantPackage(tenant.ID, body.Package); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) UpdateTenant(c *gin.Context) {
	var tenant models.Tenant
	if err := h.db.First(&tenant, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&tenant).Updates(body).First(&tenant, tenant.ID).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) UpdateTenantStatus(c *gin.Context) {
	var body struct {
		Status int `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)
	if err := h.db.Model(&models.Tenant{}).Where("id = ?", c.Param("id")).Update("status", body.Status).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	var tenant models.Tenant
	_ = h.db.First(&tenant, c.Param("id")).Error
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) DeleteTenant(c *gin.Context) {
	h.deleteByID(c, &models.Tenant{})
}

func (h *IdentityHandler) Users(c *gin.Context) {
	var rows []models.AppUser
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.userToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) CreateUser(c *gin.Context) {
	var body struct {
		EmployeeNo   string   `json:"employee_no"`
		Password     string   `json:"password"`
		Name         string   `json:"name"`
		Phone        *string  `json:"phone"`
		Email        *string  `json:"email"`
		CompanyID    *uint64  `json:"company_id"`
		DepartmentID *uint64  `json:"department_id"`
		PositionIDs  []uint64 `json:"position_ids"`
		RoleIDs      []uint64 `json:"role_ids"`
		Status       int      `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	user := models.AppUser{TenantID: 1, EmployeeNo: body.EmployeeNo, Account: body.EmployeeNo, PasswordHash: devPasswordHash(body.Password), Name: body.Name, Phone: body.Phone, Email: body.Email, CompanyID: body.CompanyID, DepartmentID: body.DepartmentID, Status: body.Status}
	if err := h.db.Create(&user).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.replaceUserRelations(user.ID, body.RoleIDs, body.PositionIDs, nil)
	response.OK(c, h.userToJSON(user))
}

func (h *IdentityHandler) UpdateUser(c *gin.Context) {
	var user models.AppUser
	if err := h.db.First(&user, c.Param("id")).Error; err != nil {
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
		updates["phone"] = *body.Phone
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
		updates["is_platform_admin"] = *body.IsPlatformAdmin
	}
	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).First(&user, user.ID).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	h.replaceUserRelations(user.ID, body.RoleIDs, body.PositionIDs, body.DepartmentIDs)
	response.OK(c, h.userToJSON(user))
}

func (h *IdentityHandler) ResetUserPassword(c *gin.Context) {
	newPassword := fmt.Sprintf("Pwd%06d", time.Now().UnixNano()%1000000)
	if err := h.db.Model(&models.AppUser{}).Where("id = ?", c.Param("id")).Update("password_hash", devPasswordHash(newPassword)).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"new_password": newPassword})
}

func (h *IdentityHandler) DeleteUser(c *gin.Context) {
	h.deleteByID(c, &models.AppUser{})
}

func (h *IdentityHandler) AssignableRoles(c *gin.Context) {
	var rows []models.Role
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "code": row.Code, "name": row.Name})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) Roles(c *gin.Context) {
	var rows []models.Role
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.roleToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) Role(c *gin.Context) {
	var row models.Role
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	response.OK(c, h.roleToJSON(row))
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
	role := models.Role{TenantID: 1, Code: body.Code, Name: body.Name, Description: body.Description, Status: 1}
	if err := h.db.Create(&role).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.replaceRolePermissions(role.ID, body.PermissionIDs)
	response.OK(c, h.roleToJSON(role))
}

func (h *IdentityHandler) UpdateRole(c *gin.Context) {
	var role models.Role
	if err := h.db.First(&role, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	var body struct {
		Name          *string  `json:"name"`
		Description   *string  `json:"description"`
		PermissionIDs []uint64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Description != nil {
		updates["description"] = *body.Description
	}
	if len(updates) > 0 {
		if err := h.db.Model(&role).Updates(updates).First(&role, role.ID).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	if body.PermissionIDs != nil {
		h.replaceRolePermissions(role.ID, body.PermissionIDs)
	}
	response.OK(c, h.roleToJSON(role))
}

func (h *IdentityHandler) DeleteRole(c *gin.Context) {
	h.deleteByID(c, &models.Role{})
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
	var rows []models.SaasPlan
	_ = h.db.Order("sort_order asc, id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
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
		})
	}
	response.OK(c, paginated(items))
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
	row := models.SaasPlan{PlanCode: body.PlanCode, PlanName: body.PlanName, PlanType: body.PlanType, BillingCycle: body.BillingCycle, Price: body.Price, Status: body.Status, IsDefault: body.IsDefault, SortOrder: body.SortOrder, Description: body.Description}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, planToJSON(row))
}

func (h *IdentityHandler) UpdatePlan(c *gin.Context) {
	var row models.SaasPlan
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
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
	if err := h.db.First(&src, c.Param("id")).Error; err != nil {
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
	dst := src
	dst.ID = 0
	dst.PlanCode = body.PlanCode
	dst.PlanName = body.PlanName
	dst.Description = body.Description
	dst.IsDefault = false
	if err := h.db.Create(&dst).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, planToJSON(dst))
}

func (h *IdentityHandler) DeletePlan(c *gin.Context) {
	h.deleteByID(c, &models.SaasPlan{})
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
	_ = h.db.Where("plan_id = ?", planID).Delete(&models.SaasPlanFeature{}).Error
	for _, featureID := range body.FeatureIDs {
		_ = h.db.Create(&models.SaasPlanFeature{PlanID: planID, FeatureID: featureID, Enabled: true}).Error
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
	h.SavePlanFeaturesWithIDs(c, body.FeatureIDs)
	if c.Writer.Written() {
		return
	}
	planID := parseUintParam(c, "id")
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
	_ = h.db.Where("plan_id = ?", planID).Delete(&models.SaasPlanFeature{}).Error
	for _, featureID := range featureIDs {
		if err := h.db.Create(&models.SaasPlanFeature{PlanID: planID, FeatureID: featureID, Enabled: true}).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
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
	newPassword := fmt.Sprintf("Pwd%06d", time.Now().UnixNano()%1000000)
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
	var quota models.SaasQuota
	if err := h.db.Where("quota_code = ?", quotaCode).First(&quota).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "配额不存在")
		return
	}
	used := h.currentQuotaUsage(tenantID, quotaCode)
	limit := h.currentQuotaLimit(tenantID, quota.ID)
	response.OK(c, gin.H{"tenant_id": tenantID, "quota_code": quotaCode, "used_value": used, "limit_value": limit, "remaining_value": limit - used, "allowed": limit < 0 || used < limit})
}

func (h *IdentityHandler) OrganizationTree(c *gin.Context) {
	var rows []models.OrgNode
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, orgNodeToJSON(row, []gin.H{}))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) OrganizationDetail(c *gin.Context) {
	tenantID := parseTenantID(c)
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
	h.deleteByID(c, &models.OrgNode{})
}

func (h *IdentityHandler) DeleteCompany(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
}

func (h *IdentityHandler) DeleteDepartment(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
}

func (h *IdentityHandler) DeleteStore(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
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
	h.deleteByID(c, &models.Position{})
}

func (h *IdentityHandler) BusinessUnits(c *gin.Context) {
	var rows []models.BusinessUnit
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) BusinessUnitTree(c *gin.Context) {
	var rows []models.BusinessUnit
	_ = h.db.Order("id asc").Find(&rows).Error
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
	row := models.BusinessUnit{TenantID: 1, Name: body.Name, Code: body.Code, BUType: body.BUType, Status: body.Status, BillingEnabled: true, StatisticEnabled: true, Remark: body.Remark}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.replaceBusinessUnitMappings(row.ID, body.OrgNodeIDs)
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateBusinessUnit(c *gin.Context) {
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
		if err := h.db.Model(&models.BusinessUnit{}).Where("id = ?", c.Param("id")).Updates(updates).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	if body.OrgNodeIDs != nil {
		h.replaceBusinessUnitMappings(parseUintParam(c, "id"), body.OrgNodeIDs)
	}
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeleteBusinessUnit(c *gin.Context) {
	h.deleteByID(c, &models.BusinessUnit{})
}

func (h *IdentityHandler) BusinessUnitOrgMappings(c *gin.Context) {
	var rows []models.BusinessUnitOrgMap
	_ = h.db.Where("business_unit_id = ?", c.Param("id")).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitOrgMapToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessUnitOrgMapping(c *gin.Context) {
	var body struct {
		OrgID uint64 `json:"org_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	row := models.BusinessUnitOrgMap{TenantID: 1, BusinessUnitID: parseUintParam(c, "id"), OrgID: body.OrgID, OrgType: "org", ScopeType: "include", Status: 1}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) DeleteBusinessUnitOrgMapping(c *gin.Context) {
	h.deleteByID(c, &models.BusinessUnitOrgMap{})
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
	var userCount int64
	var buCount int64
	db.Model(&models.OrgNode{}).Where("tenant_id = ? AND node_type = ?", row.ID, "company").Count(&companyCount)
	db.Model(&models.AppUser{}).Where("tenant_id = ?", row.ID).Count(&userCount)
	db.Model(&models.BusinessUnit{}).Where("tenant_id = ?", row.ID).Count(&buCount)
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "status": row.Status, "start_date": row.StartDate, "expire_date": row.ExpireDate, "max_companies": row.MaxCompanies, "max_users": row.MaxUsers, "used_companies": companyCount, "used_users": userCount, "used_business_units": buCount, "plan_name": planName, "plan_code": planCode, "contact_name": row.ContactName, "contact_phone": row.ContactPhone, "company_count": companyCount, "brand_display_name": row.BrandName, "brand_logo_data": row.LogoData, "created_at": row.CreatedAt}
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

func (h *IdentityHandler) roleToJSON(row models.Role) gin.H {
	var links []models.RolePermission
	_ = h.db.Where("role_id = ?", row.ID).Find(&links).Error
	permissionIDs := make([]uint64, 0, len(links))
	for _, link := range links {
		permissionIDs = append(permissionIDs, link.PermissionID)
	}
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "description": row.Description, "permission_ids": permissionIDs, "data_overrides": []gin.H{}}
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
	if raw := c.Query("tenant_id"); raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return 1
}

func (h *IdentityHandler) currentUser(c *gin.Context) (models.AppUser, bool) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		return models.AppUser{}, false
	}
	claims, err := parseToken(token, h.authSecret)
	if err != nil || claims.UserID == 0 {
		return models.AppUser{}, false
	}
	var user models.AppUser
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
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
	key := "captcha:" + id
	stored, err := h.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	_ = h.redis.Del(ctx, key).Err()
	return strings.EqualFold(strings.TrimSpace(stored), strings.TrimSpace(code))
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

func (h *IdentityHandler) deleteByID(c *gin.Context, model interface{}) {
	id := parseUintParam(c, "id")
	if err := h.db.Delete(model, id).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
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

func (h *IdentityHandler) replaceRolePermissions(roleID uint64, permissionIDs []uint64) {
	_ = h.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
	for _, id := range permissionIDs {
		_ = h.db.Create(&models.RolePermission{RoleID: roleID, PermissionID: id}).Error
	}
}

func (h *IdentityHandler) createTenantWithAdmin(code, name string, status int, adminName, adminEmployeeNo string, adminPhone *string, adminPassword string) (models.Tenant, error) {
	if status == 0 {
		status = 1
	}
	tenant := models.Tenant{Code: code, Name: name, Status: status, ContactName: &adminName, ContactPhone: adminPhone}
	if err := h.db.Create(&tenant).Error; err != nil {
		return tenant, err
	}
	companyCode := strings.ToUpper(code)
	company := models.OrgNode{TenantID: tenant.ID, NodeType: "company", Name: name, Code: &companyCode, Status: 1}
	_ = h.db.Create(&company).Error
	role := models.Role{TenantID: tenant.ID, Code: "admin", Name: "主体管理员", Status: 1}
	_ = h.db.Create(&role).Error
	user := models.AppUser{TenantID: tenant.ID, CompanyID: &company.ID, EmployeeNo: adminEmployeeNo, Account: adminEmployeeNo, PasswordHash: devPasswordHash(adminPassword), Name: adminName, Phone: adminPhone, Status: 1, IsPlatformAdmin: false}
	_ = h.db.Create(&user).Error
	_ = h.db.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID}).Error
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
	start := time.Now()
	if body.StartTime != "" {
		if parsed := parseTimePtr(&body.StartTime); parsed != nil {
			start = *parsed
		}
	}
	if body.SubscriptionStatus == "" {
		body.SubscriptionStatus = "ACTIVE"
	}
	var existing models.TenantSubscription
	sub := models.TenantSubscription{TenantID: tenantID, PlanID: body.PlanID, SubscriptionStatus: body.SubscriptionStatus, StartTime: start, EndTime: parseTimePtr(body.EndTime), TrialEndTime: parseTimePtr(body.TrialEndTime), AutoRenew: body.AutoRenew, FrozenReason: body.FrozenReason}
	if err := h.db.Where("tenant_id = ?", tenantID).First(&existing).Error; err == nil {
		sub.ID = existing.ID
		return h.db.Model(&existing).Updates(sub).Error
	}
	if err := h.db.Create(&sub).Error; err != nil {
		return err
	}
	for _, quota := range body.Quotas {
		_ = h.db.Where("tenant_id = ? AND quota_id = ?", tenantID, quota.QuotaID).Delete(&models.TenantQuotaOverride{}).Error
		_ = h.db.Create(&models.TenantQuotaOverride{TenantID: tenantID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue}).Error
	}
	return nil
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
	row := models.OrgNode{TenantID: parseTenantID(c), NodeType: body.NodeType, Name: body.Name, Code: body.Code, CompanyType: body.CompanyType, CompanyID: body.CompanyID, ParentID: body.ParentID, Status: body.Status}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) updateOrgNode(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&models.OrgNode{}).Where("id = ?", c.Param("id")).Updates(body).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) replaceBusinessUnitMappings(buID uint64, orgIDs []uint64) {
	_ = h.db.Where("business_unit_id = ?", buID).Delete(&models.BusinessUnitOrgMap{}).Error
	for _, orgID := range orgIDs {
		_ = h.db.Create(&models.BusinessUnitOrgMap{TenantID: 1, BusinessUnitID: buID, OrgID: orgID, OrgType: "org", ScopeType: "include", Status: 1}).Error
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
