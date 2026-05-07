package handlers

import (
	"strconv"

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
		items = append(items, gin.H{
			"id":                 row.ID,
			"code":               row.Code,
			"name":               row.Name,
			"status":             row.Status,
			"plan_name":          "平台版",
			"plan_code":          "platform",
			"contact_name":       "平台管理员",
			"contact_phone":      nil,
			"company_count":      0,
			"brand_display_name": row.BrandName,
			"brand_logo_data":    nil,
			"created_at":         row.CreatedAt,
		})
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) Tenant(c *gin.Context) {
	var row models.Tenant
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	response.OK(c, gin.H{
		"id":                 row.ID,
		"code":               row.Code,
		"name":               row.Name,
		"status":             row.Status,
		"plan_name":          "平台版",
		"plan_code":          "platform",
		"contact_name":       "平台管理员",
		"contact_phone":      nil,
		"company_count":      0,
		"brand_display_name": row.BrandName,
		"brand_logo_data":    nil,
		"created_at":         row.CreatedAt,
	})
}

func (h *MockIdentityHandler) Users(c *gin.Context) {
	var rows []models.AppUser
	_ = h.db.Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":                row.ID,
			"tenant_id":         row.TenantID,
			"employee_no":       row.EmployeeNo,
			"phone":             row.Phone,
			"name":              row.Name,
			"email":             row.Email,
			"avatar_url":        row.AvatarURL,
			"status":            row.Status,
			"company_id":        nil,
			"department_id":     nil,
			"department_ids":    []uint64{},
			"position_ids":      []uint64{},
			"role_ids":          []uint64{1},
			"is_platform_admin": row.IsPlatformAdmin,
			"created_at":        row.CreatedAt,
		})
	}
	response.OK(c, paginated(items))
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
		var links []models.RolePermission
		_ = h.db.Where("role_id = ?", row.ID).Find(&links).Error
		permissionIDs := make([]uint64, 0, len(links))
		for _, link := range links {
			permissionIDs = append(permissionIDs, link.PermissionID)
		}
		items = append(items, gin.H{
			"id":             row.ID,
			"code":           row.Code,
			"name":           row.Name,
			"description":    nil,
			"permission_ids": permissionIDs,
			"data_overrides": []gin.H{},
		})
	}
	response.OK(c, paginated(items))
}

func (h *MockIdentityHandler) Role(c *gin.Context) {
	var row models.Role
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	response.OK(c, gin.H{
		"id":             row.ID,
		"code":           row.Code,
		"name":           row.Name,
		"description":    nil,
		"permission_ids": []uint64{},
		"data_overrides": []gin.H{},
	})
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
	response.OK(c, gin.H{
		"tenant_id":      1,
		"companies":      []gin.H{},
		"stores":         []gin.H{},
		"business_units": []gin.H{},
	})
}

func (h *MockIdentityHandler) TenantCompanies(c *gin.Context) {
	response.OK(c, gin.H{
		"tenant_id":          1,
		"billing_unit_count": 0,
		"companies":          []gin.H{},
	})
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
	response.OK(c, paginated([]gin.H{}))
}

func (h *MockIdentityHandler) BusinessUnitTree(c *gin.Context) {
	response.OK(c, []gin.H{})
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

func paginated(items interface{}) gin.H {
	total := 0
	switch v := items.(type) {
	case []gin.H:
		total = len(v)
	}
	return gin.H{"items": items, "total": total, "skip": 0, "limit": 50}
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

func (h *MockIdentityHandler) deleteByID(c *gin.Context, model interface{}) {
	id := parseUintParam(c, "id")
	if err := h.db.Delete(model, id).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
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
