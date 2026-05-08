package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) CreatePermission(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body permissionPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if msg := validatePermissionPayload(body); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	row := models.Permission{
		TenantID:         user.TenantID,
		ParentID:         body.ParentID,
		Name:             strings.TrimSpace(body.Name),
		Path:             strings.TrimSpace(body.Path),
		PermType:         intValueOrZero(body.PermType),
		DataScope:        trimmedStringPtr(body.DataScope),
		SortOrder:        intValueOrDefault(body.SortOrder, 0),
		Enabled:          body.Enabled == nil || *body.Enabled,
		Visible:          body.TenantVisible == nil || *body.TenantVisible,
		IsPlatformOnly:   body.IsPlatformOnly != nil && *body.IsPlatformOnly,
		IsPackageFeature: body.IsPackageFeature == nil || *body.IsPackageFeature,
		TenantEditable:   body.TenantEditable != nil && *body.TenantEditable,
		TenantEditScope:  nullableTrimmed(body.TenantEditScope),
		FeatureCode:      nullableTrimmed(body.FeatureCode),
		FeatureType:      nullableTrimmed(body.FeatureType),
		DataPermMode:     coalesceStringPtr(body.DataPermMode, "ORG"),
	}
	if row.Name == "" || row.Path == "" || body.PermType == nil {
		response.Error(c, 400, response.CodeBadRequest, "权限名称、路径和类型不能为空")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return replacePermissionCustomScopes(tx, row.ID, body.CustomDepartmentIDs, body.CustomUserIDs, false, false)
	}); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.syncPackageFeaturesFromPermissions()
	h.invalidateTenantAuthorizationCache(user.TenantID)
	h.audit(c, user.TenantID, user.ID, "permission", "create", "创建权限 "+row.Name, gin.H{"id": row.ID, "path": row.Path, "perm_type": row.PermType})
	response.OK(c, h.permissionToJSON(row))
}

func (h *IdentityHandler) UpdatePermission(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var row models.Permission
	if err := h.tenantScope().ActiveByID(user.TenantID, c.Param("id")).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "权限不存在")
		return
	}
	var body permissionPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if msg := validatePermissionPayload(body); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	applyPermissionPayload(&row, body)
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		return replacePermissionCustomScopes(tx, row.ID, body.CustomDepartmentIDs, body.CustomUserIDs, body.CustomDepartmentIDs != nil, body.CustomUserIDs != nil)
	}); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.syncPackageFeaturesFromPermissions()
	h.invalidateTenantAuthorizationCache(user.TenantID)
	h.audit(c, user.TenantID, user.ID, "permission", "update", "编辑权限 "+row.Name, gin.H{"id": row.ID, "path": row.Path})
	response.OK(c, h.permissionToJSON(row))
}

func (h *IdentityHandler) DeletePermission(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "权限", ref(&models.Permission{}, "子权限", "parent_id = ?", id), ref(&models.RolePermission{}, "角色权限", "permission_id = ?", id), ref(&models.TenantMenuOverride{}, "租户菜单覆盖", "permission_id = ?", id)) {
		return
	}
	var row models.Permission
	if err := h.tenantScope().ActiveByID(user.TenantID, id).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "权限不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "enabled": false}).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.disablePackageFeatureForPermission(row)
	h.invalidateTenantAuthorizationCache(user.TenantID)
	h.audit(c, user.TenantID, user.ID, "permission", "delete", "删除权限 "+row.Name, gin.H{"id": id})
	response.OK(c, gin.H{"deleted": id})
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

type permissionPayload struct {
	ParentID            *uint64  `json:"parent_id"`
	Name                string   `json:"name"`
	Path                string   `json:"path"`
	PermType            *int     `json:"perm_type"`
	DataScope           *string  `json:"data_scope"`
	CustomDepartmentIDs []uint64 `json:"custom_department_ids"`
	CustomUserIDs       []uint64 `json:"custom_user_ids"`
	SortOrder           *int     `json:"sort_order"`
	Enabled             *bool    `json:"enabled"`
	TenantVisible       *bool    `json:"tenant_visible"`
	IsPlatformOnly      *bool    `json:"is_platform_only"`
	IsPackageFeature    *bool    `json:"is_package_feature"`
	FeatureCode         *string  `json:"feature_code"`
	FeatureType         *string  `json:"feature_type"`
	TenantEditable      *bool    `json:"tenant_editable"`
	TenantEditScope     *string  `json:"tenant_edit_scope"`
	DataPermMode        *string  `json:"data_perm_mode"`
}

func validatePermissionPayload(body permissionPayload) string {
	dataScope := ""
	if body.DataScope != nil {
		dataScope = strings.TrimSpace(*body.DataScope)
	}
	dataPermMode := ""
	if body.DataPermMode != nil {
		dataPermMode = strings.TrimSpace(*body.DataPermMode)
	}
	if body.PermType != nil && *body.PermType == 4 && dataScope == "" {
		return "数据权限必须带 data_scope"
	}
	if dataScope != "" && !allowedString(dataScope, "ALL", "ORG", "ORG_SUB", "SELF", "CUSTOM") {
		return "无效的数据范围: " + dataScope
	}
	if dataScope == "CUSTOM" && len(body.CustomDepartmentIDs) == 0 && len(body.CustomUserIDs) == 0 {
		return "CUSTOM 范围需至少指定组织架构或用户"
	}
	if dataPermMode != "" && !allowedString(dataPermMode, "NONE", "ORG", "BU", "ORG_BU") {
		return "无效的数据权限类型: " + dataPermMode
	}
	if dataPermMode != "" && body.PermType != nil && *body.PermType != 3 {
		return "仅菜单权限支持配置数据权限类型"
	}
	return ""
}

func allowedString(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

func intValueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func intValueOrDefault(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func applyPermissionPayload(row *models.Permission, body permissionPayload) {
	if body.ParentID != nil {
		row.ParentID = body.ParentID
	}
	if strings.TrimSpace(body.Name) != "" {
		row.Name = strings.TrimSpace(body.Name)
	}
	if strings.TrimSpace(body.Path) != "" {
		row.Path = strings.TrimSpace(body.Path)
	}
	if body.PermType != nil {
		row.PermType = *body.PermType
	}
	if body.DataScope != nil {
		row.DataScope = trimmedStringPtr(body.DataScope)
	}
	if body.SortOrder != nil {
		row.SortOrder = *body.SortOrder
	}
	if body.Enabled != nil {
		row.Enabled = *body.Enabled
	}
	if body.TenantVisible != nil {
		row.Visible = *body.TenantVisible
	}
	if body.IsPlatformOnly != nil {
		row.IsPlatformOnly = *body.IsPlatformOnly
	}
	if body.IsPackageFeature != nil {
		row.IsPackageFeature = *body.IsPackageFeature
	}
	if body.FeatureCode != nil {
		row.FeatureCode = nullableTrimmed(body.FeatureCode)
	}
	if body.FeatureType != nil {
		row.FeatureType = nullableTrimmed(body.FeatureType)
	}
	if body.TenantEditable != nil {
		row.TenantEditable = *body.TenantEditable
	}
	if body.TenantEditScope != nil {
		row.TenantEditScope = nullableTrimmed(body.TenantEditScope)
	}
	if body.DataPermMode != nil {
		row.DataPermMode = coalesceStringPtr(body.DataPermMode, "ORG")
	}
	if row.DataPermMode == "" {
		row.DataPermMode = "ORG"
	}
}

func replacePermissionCustomScopes(tx *gorm.DB, permissionID uint64, departmentIDs []uint64, userIDs []uint64, replaceDepartments bool, replaceUsers bool) error {
	if replaceDepartments {
		if err := tx.Where("permission_id = ?", permissionID).Delete(&models.PermissionCustomDepartment{}).Error; err != nil {
			return err
		}
	}
	if replaceUsers {
		if err := tx.Where("permission_id = ?", permissionID).Delete(&models.PermissionCustomUser{}).Error; err != nil {
			return err
		}
	}
	if !replaceDepartments && len(departmentIDs) > 0 {
		if err := tx.Where("permission_id = ?", permissionID).Delete(&models.PermissionCustomDepartment{}).Error; err != nil {
			return err
		}
	}
	if !replaceUsers && len(userIDs) > 0 {
		if err := tx.Where("permission_id = ?", permissionID).Delete(&models.PermissionCustomUser{}).Error; err != nil {
			return err
		}
	}
	for _, id := range uniqueUint64s(departmentIDs) {
		if err := tx.Create(&models.PermissionCustomDepartment{PermissionID: permissionID, DepartmentID: id}).Error; err != nil {
			return err
		}
	}
	for _, id := range uniqueUint64s(userIDs) {
		if err := tx.Create(&models.PermissionCustomUser{PermissionID: permissionID, UserID: id}).Error; err != nil {
			return err
		}
	}
	return nil
}
