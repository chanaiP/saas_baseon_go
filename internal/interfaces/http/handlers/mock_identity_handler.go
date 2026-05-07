package handlers

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

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

func (h *MockIdentityHandler) Tenants(c *gin.Context) {
	var rows []models.Tenant
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, tenantToJSON(h.db, row))
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) Tenant(c *gin.Context) {
	var row models.Tenant
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	response.OK(c, tenantToJSON(h.db, row))
}

func (h *MockIdentityHandler) CreateTenant(c *gin.Context) {
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

func (h *MockIdentityHandler) CreateTenantWithPackage(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateTenant(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateTenantStatus(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteTenant(c *gin.Context) {
	h.deleteByID(c, &models.Tenant{})
}

func (h *MockIdentityHandler) Users(c *gin.Context) {
	var rows []models.AppUser
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.userToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) CreateUser(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateUser(c *gin.Context) {
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

func (h *MockIdentityHandler) ResetUserPassword(c *gin.Context) {
	newPassword := fmt.Sprintf("Pwd%06d", time.Now().UnixNano()%1000000)
	if err := h.db.Model(&models.AppUser{}).Where("id = ?", c.Param("id")).Update("password_hash", devPasswordHash(newPassword)).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"new_password": newPassword})
}

func (h *MockIdentityHandler) DeleteUser(c *gin.Context) {
	h.deleteByID(c, &models.AppUser{})
}

func (h *MockIdentityHandler) AssignableRoles(c *gin.Context) {
	var rows []models.Role
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "code": row.Code, "name": row.Name})
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) Roles(c *gin.Context) {
	var rows []models.Role
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.roleToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) Role(c *gin.Context) {
	var row models.Role
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	response.OK(c, h.roleToJSON(row))
}

func (h *MockIdentityHandler) CreateRole(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateRole(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteRole(c *gin.Context) {
	h.deleteByID(c, &models.Role{})
}

func (h *MockIdentityHandler) UpdatePermissionDataPermMode(c *gin.Context) {
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

func (h *MockIdentityHandler) Plans(c *gin.Context) {
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

func (h *MockIdentityHandler) CreatePlan(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdatePlan(c *gin.Context) {
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

func (h *MockIdentityHandler) CopyPlan(c *gin.Context) {
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

func (h *MockIdentityHandler) DeletePlan(c *gin.Context) {
	h.deleteByID(c, &models.SaasPlan{})
}

func (h *MockIdentityHandler) PlanMatrix(c *gin.Context) {
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

func (h *MockIdentityHandler) Features(c *gin.Context) {
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

func (h *MockIdentityHandler) CreateFeature(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateFeature(c *gin.Context) {
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

func (h *MockIdentityHandler) PlanFeatures(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var links []models.SaasPlanFeature
	_ = h.db.Where("plan_id = ? AND enabled = ?", planID, true).Find(&links).Error
	ids := make([]uint64, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.FeatureID)
	}
	response.OK(c, gin.H{"plan_id": planID, "feature_ids": ids})
}

func (h *MockIdentityHandler) SavePlanFeatures(c *gin.Context) {
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

func (h *MockIdentityHandler) SavePlanCapabilities(c *gin.Context) {
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

func (h *MockIdentityHandler) SavePlanFeaturesWithIDs(c *gin.Context, featureIDs []uint64) {
	planID := parseUintParam(c, "id")
	_ = h.db.Where("plan_id = ?", planID).Delete(&models.SaasPlanFeature{}).Error
	for _, featureID := range featureIDs {
		if err := h.db.Create(&models.SaasPlanFeature{PlanID: planID, FeatureID: featureID, Enabled: true}).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
}

func (h *MockIdentityHandler) Quotas(c *gin.Context) {
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

func (h *MockIdentityHandler) CreateQuota(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateQuota(c *gin.Context) {
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

func (h *MockIdentityHandler) PlanQuotas(c *gin.Context) {
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

func (h *MockIdentityHandler) SavePlanQuotas(c *gin.Context) {
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

func (h *MockIdentityHandler) TenantQuotaRecords(c *gin.Context) {
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

func (h *MockIdentityHandler) TenantCompanies(c *gin.Context) {
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

func (h *MockIdentityHandler) TenantPrimaryAdmin(c *gin.Context) {
	user, ok := h.primaryAdmin(parseUintParam(c, "id"))
	if !ok {
		response.Error(c, 404, response.CodeNotFound, "主管理员不存在")
		return
	}
	response.OK(c, gin.H{"employee_no": user.EmployeeNo, "phone": user.Phone, "name": user.Name})
}

func (h *MockIdentityHandler) ResetTenantPrimaryAdminPassword(c *gin.Context) {
	user, ok := h.primaryAdmin(parseUintParam(c, "id"))
	if !ok {
		response.Error(c, 404, response.CodeNotFound, "主管理员不存在")
		return
	}
	newPassword := fmt.Sprintf("Pwd%06d", time.Now().UnixNano()%1000000)
	_ = h.db.Model(&user).Update("password_hash", devPasswordHash(newPassword)).Error
	response.OK(c, gin.H{"employee_no": user.EmployeeNo, "phone": user.Phone, "name": user.Name, "new_password": newPassword})
}

func (h *MockIdentityHandler) TenantSubscription(c *gin.Context) {
	sub, ok := h.findTenantSubscription(parseUintParam(c, "id"))
	if !ok {
		response.OK(c, nil)
		return
	}
	response.OK(c, tenantSubscriptionToJSON(sub))
}

func (h *MockIdentityHandler) SaveTenantSubscription(c *gin.Context) {
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

func (h *MockIdentityHandler) SaveTenantPackageConfig(c *gin.Context) {
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

func (h *MockIdentityHandler) TenantFeatureOverrides(c *gin.Context) {
	response.OK(c, h.tenantFeatureOverridesPayload(parseUintParam(c, "id")))
}

func (h *MockIdentityHandler) SaveTenantFeatureOverrides(c *gin.Context) {
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

func (h *MockIdentityHandler) TenantQuotaOverrides(c *gin.Context) {
	response.OK(c, h.tenantQuotaOverridesPayload(parseUintParam(c, "id")))
}

func (h *MockIdentityHandler) SaveTenantQuotaOverrides(c *gin.Context) {
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

func (h *MockIdentityHandler) TenantQuotaUsage(c *gin.Context) {
	tenantID := parseUintParam(c, "id")
	var quotas []models.SaasQuota
	_ = h.db.Find(&quotas).Error
	usages := make([]gin.H, 0, len(quotas))
	for _, quota := range quotas {
		usages = append(usages, gin.H{"quota_id": quota.ID, "quota_code": quota.QuotaCode, "quota_name": quota.QuotaName, "quota_type": quota.QuotaType, "period_type": quota.PeriodType, "period_key": "current", "unit": quota.Unit, "used_value": 0, "limit_value": 0, "remaining_value": 0})
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "usages": usages})
}

func (h *MockIdentityHandler) OrganizationTree(c *gin.Context) {
	var rows []models.OrgNode
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, orgNodeToJSON(row, []gin.H{}))
	}
	response.OK(c, items)
}

func (h *MockIdentityHandler) CreateOrgNode(c *gin.Context) {
	h.createOrgNode(c, "")
}

func (h *MockIdentityHandler) CreateCompany(c *gin.Context) {
	h.createOrgNode(c, "company")
}

func (h *MockIdentityHandler) CreateDepartment(c *gin.Context) {
	h.createOrgNode(c, "department")
}

func (h *MockIdentityHandler) CreateStore(c *gin.Context) {
	h.createOrgNode(c, "store")
}

func (h *MockIdentityHandler) UpdateOrgNode(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *MockIdentityHandler) UpdateCompany(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *MockIdentityHandler) UpdateDepartment(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *MockIdentityHandler) UpdateStore(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *MockIdentityHandler) DeleteOrgNode(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
}

func (h *MockIdentityHandler) DeleteCompany(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
}

func (h *MockIdentityHandler) DeleteDepartment(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
}

func (h *MockIdentityHandler) DeleteStore(c *gin.Context) {
	h.deleteByID(c, &models.OrgNode{})
}

func (h *MockIdentityHandler) PositionTypes(c *gin.Context) {
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

func (h *MockIdentityHandler) CreatePositionType(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdatePositionType(c *gin.Context) {
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

func (h *MockIdentityHandler) DeletePositionType(c *gin.Context) {
	h.deleteByID(c, &models.PositionType{})
}

func (h *MockIdentityHandler) Positions(c *gin.Context) {
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

func (h *MockIdentityHandler) CreatePosition(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdatePosition(c *gin.Context) {
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

func (h *MockIdentityHandler) DeletePosition(c *gin.Context) {
	h.deleteByID(c, &models.Position{})
}

func (h *MockIdentityHandler) BusinessUnits(c *gin.Context) {
	var rows []models.BusinessUnit
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) BusinessUnitTree(c *gin.Context) {
	var rows []models.BusinessUnit
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, items)
}

func (h *MockIdentityHandler) CreateBusinessUnit(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateBusinessUnit(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteBusinessUnit(c *gin.Context) {
	h.deleteByID(c, &models.BusinessUnit{})
}

func (h *MockIdentityHandler) BusinessUnitOrgMappings(c *gin.Context) {
	var rows []models.BusinessUnitOrgMap
	_ = h.db.Where("business_unit_id = ?", c.Param("id")).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitOrgMapToJSON(row))
	}
	response.OK(c, items)
}

func (h *MockIdentityHandler) CreateBusinessUnitOrgMapping(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteBusinessUnitOrgMapping(c *gin.Context) {
	h.deleteByID(c, &models.BusinessUnitOrgMap{})
}

func (h *MockIdentityHandler) DictTypes(c *gin.Context) {
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

func (h *MockIdentityHandler) CreateDictType(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateDictType(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteDictType(c *gin.Context) {
	h.deleteByID(c, &models.DictType{})
}

func (h *MockIdentityHandler) DictItemsByCode(c *gin.Context) {
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

func (h *MockIdentityHandler) DictItems(c *gin.Context) {
	var rows []models.DictItem
	query := h.db.Order("sort_order asc, id asc")
	if dictTypeID := c.Query("dict_type_id"); dictTypeID != "" {
		query = query.Where("dict_type_id = ?", dictTypeID)
	}
	_ = query.Find(&rows).Error
	response.OK(c, paginated(dictItemsToJSON(rows)))
}

func (h *MockIdentityHandler) CreateDictItem(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateDictItem(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteDictItem(c *gin.Context) {
	h.deleteByID(c, &models.DictItem{})
}

func (h *MockIdentityHandler) RestoreDictItem(c *gin.Context) {
	var row models.DictItem
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "字典项不存在")
		return
	}
	response.OK(c, dictItemsToJSON([]models.DictItem{row})[0])
}

func (h *MockIdentityHandler) SysParams(c *gin.Context) {
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

func (h *MockIdentityHandler) CreateSysParam(c *gin.Context) {
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

func (h *MockIdentityHandler) UpdateSysParam(c *gin.Context) {
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

func (h *MockIdentityHandler) DeleteSysParam(c *gin.Context) {
	h.deleteByID(c, &models.SystemParam{})
}

func (h *MockIdentityHandler) RestoreSysParam(c *gin.Context) {
	var row models.SystemParam
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "参数不存在")
		return
	}
	response.OK(c, gin.H{"id": row.ID, "param_key": row.Key, "default_value": row.Value, "param_value": row.Value, "remark": row.Remark, "value_type": row.ValueType, "tenant_editable": row.TenantEditable, "is_platform_only": row.IsPlatformOnly, "is_override": false})
}

func (h *MockIdentityHandler) SysParamBatch(c *gin.Context) {
	values := gin.H{}
	var rows []models.SystemParam
	_ = h.db.Find(&rows).Error
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	response.OK(c, gin.H{"values": values})
}

func (h *MockIdentityHandler) LoginLogs(c *gin.Context) {
	var rows []models.LoginLog
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "tenant_name": h.tenantName(row.TenantID), "user_id": row.UserID, "account": row.Account, "success": row.Success, "message": row.Message, "ip": row.IP, "created_at": row.CreatedAt})
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) AuditLogs(c *gin.Context) {
	var rows []models.AuditLog
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "tenant_name": h.tenantName(row.TenantID), "user_id": row.UserID, "module": row.Module, "action": row.Action, "summary": row.Summary, "detail": row.Detail, "ip": row.IP, "created_at": row.CreatedAt})
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) MonitorHealthDetail(c *gin.Context) {
	response.OK(c, gin.H{"mysql": true, "postgres": true, "redis": true})
}

func (h *MockIdentityHandler) MonitorServerInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	response.OK(c, gin.H{"python_version": "go " + runtime.Version(), "go_version": runtime.Version(), "pid": 1, "cpu_percent": nil, "memory_mb": float64(m.Alloc) / 1024 / 1024, "note": "Go 重构版本运行中"})
}

func (h *MockIdentityHandler) MonitorScheduledJobs(c *gin.Context) {
	response.OK(c, gin.H{"items": []gin.H{}, "note": "当前 Go 版本暂未启用后台定时任务"})
}

func (h *MockIdentityHandler) MonitorServicesOverview(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	response.OK(c, gin.H{"mysql": true, "postgres": true, "redis": true, "python_version": "go " + runtime.Version(), "go_version": runtime.Version(), "pid": 1, "cpu_percent": nil, "memory_mb": float64(m.Alloc) / 1024 / 1024, "note": "Go 重构版本服务概览"})
}

func (h *MockIdentityHandler) MonitorCacheStats(c *gin.Context) {
	response.OK(c, gin.H{"ok": true, "used_memory_human": nil, "keys": 0, "connected_clients": 1, "message": "Redis 已连接，详细 INFO 待接入"})
}

func (h *MockIdentityHandler) MonitorCacheKeys(c *gin.Context) {
	response.OK(c, gin.H{"items": []gin.H{}, "cursor": 0})
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

func (h *MockIdentityHandler) userToJSON(row models.AppUser) gin.H {
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

func (h *MockIdentityHandler) roleToJSON(row models.Role) gin.H {
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

func (h *MockIdentityHandler) deleteByID(c *gin.Context, model interface{}) {
	id := parseUintParam(c, "id")
	if err := h.db.Delete(model, id).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
}

func devPasswordHash(password string) string {
	if password == "" {
		password = "112233"
	}
	return "dev:" + password
}

func (h *MockIdentityHandler) replaceUserRelations(userID uint64, roleIDs []uint64, positionIDs []uint64, departmentIDs []uint64) {
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

func (h *MockIdentityHandler) replaceRolePermissions(roleID uint64, permissionIDs []uint64) {
	_ = h.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
	for _, id := range permissionIDs {
		_ = h.db.Create(&models.RolePermission{RoleID: roleID, PermissionID: id}).Error
	}
}

func (h *MockIdentityHandler) createTenantWithAdmin(code, name string, status int, adminName, adminEmployeeNo string, adminPhone *string, adminPassword string) (models.Tenant, error) {
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

func (h *MockIdentityHandler) saveTenantPackage(tenantID uint64, body tenantPackagePayload) error {
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

func (h *MockIdentityHandler) findTenantSubscription(tenantID uint64) (models.TenantSubscription, bool) {
	var sub models.TenantSubscription
	if err := h.db.Where("tenant_id = ?", tenantID).Order("id desc").First(&sub).Error; err != nil {
		return sub, false
	}
	return sub, true
}

func tenantSubscriptionToJSON(row models.TenantSubscription) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "plan_id": row.PlanID, "subscription_status": row.SubscriptionStatus, "start_time": row.StartTime, "end_time": row.EndTime, "trial_end_time": row.TrialEndTime, "auto_renew": row.AutoRenew, "frozen_reason": row.FrozenReason, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func (h *MockIdentityHandler) primaryAdmin(tenantID uint64) (models.AppUser, bool) {
	var user models.AppUser
	err := h.db.Where("tenant_id = ?", tenantID).Order("is_platform_admin desc, id asc").First(&user).Error
	return user, err == nil
}

func (h *MockIdentityHandler) tenantFeatureOverridesPayload(tenantID uint64) gin.H {
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

func (h *MockIdentityHandler) tenantQuotaOverridesPayload(tenantID uint64) gin.H {
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

func (h *MockIdentityHandler) createOrgNode(c *gin.Context, forcedType string) {
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

func (h *MockIdentityHandler) updateOrgNode(c *gin.Context) {
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

func (h *MockIdentityHandler) replaceBusinessUnitMappings(buID uint64, orgIDs []uint64) {
	_ = h.db.Where("business_unit_id = ?", buID).Delete(&models.BusinessUnitOrgMap{}).Error
	for _, orgID := range orgIDs {
		_ = h.db.Create(&models.BusinessUnitOrgMap{TenantID: 1, BusinessUnitID: buID, OrgID: orgID, OrgType: "org", ScopeType: "include", Status: 1}).Error
	}
}

func (h *MockIdentityHandler) tenantName(tenantID *uint64) *string {
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
