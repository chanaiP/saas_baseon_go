package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) MenuBundles(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	forPlatform := user.IsPlatformAdmin || h.viewerHasPlatformScope(user)
	var permissions []models.Permission
	_ = h.db.Where("tenant_id IN ? AND perm_type = ? AND enabled = ? AND visible = ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(user.TenantID), 3, true, true).Order("sort_order asc, id asc").Find(&permissions).Error
	bundles := make([]gin.H, 0, len(permissions))
	for _, permission := range permissions {
		if !forPlatform && permission.IsPlatformOnly {
			continue
		}
		if !forPlatform && !h.permissionAllowedForTenantSubscription(user.TenantID, permission) {
			continue
		}
		operations := h.menuBundleOperations(user.TenantID, permission.Path, forPlatform)
		bundles = append(bundles, gin.H{
			"path":               permission.Path,
			"title":              permission.Name,
			"menu_permission_id": permission.ID,
			"data_permission_id": h.dataPermissionIDForMenu(user.TenantID, permission.Path),
			"operations":         operations,
			"is_platform_only":   permission.IsPlatformOnly,
			"is_package_feature": permission.IsPackageFeature,
			"feature_code":       permission.FeatureCode,
			"feature_type":       permission.FeatureType,
			"app_code":           permission.AppCode,
			"tenant_visible":     permission.Visible,
			"tenant_editable":    permission.TenantEditable,
			"tenant_edit_scope":  permission.TenantEditScope,
			"data_perm_mode":     permission.DataPermMode,
		})
	}
	bundles = append(bundles, h.standaloneCapabilityBundles(user.TenantID, forPlatform)...)
	response.OK(c, bundles)
}

func (h *IdentityHandler) standaloneCapabilityBundles(tenantID uint64, forPlatform bool) []gin.H {
	var permissions []models.Permission
	_ = h.db.Where(
		"tenant_id IN ? AND perm_type = ? AND enabled = ? AND visible = ? AND deleted_at IS NULL AND path IN ?",
		h.permissionScopeTenantIDs(tenantID),
		2,
		true,
		true,
		[]string{"brand:edit"},
	).Order("sort_order asc, id asc").Find(&permissions).Error
	bundles := make([]gin.H, 0, len(permissions))
	for _, permission := range permissions {
		if !forPlatform && permission.IsPlatformOnly {
			continue
		}
		if !forPlatform && !h.permissionAllowedForTenantSubscription(tenantID, permission) {
			continue
		}
		title := permission.Name
		if title == "" || title == "品牌-维护" {
			title = "品牌维护"
		}
		bundles = append(bundles, gin.H{
			"path":               permission.Path,
			"title":              title,
			"menu_permission_id": permission.ID,
			"data_permission_id": uint64(0),
			"operations":         []gin.H{},
			"is_platform_only":   permission.IsPlatformOnly,
			"is_package_feature": permission.IsPackageFeature,
			"feature_code":       permission.FeatureCode,
			"feature_type":       packageFeatureTypeForPermission(permission),
			"app_code":           permission.AppCode,
			"tenant_visible":     permission.Visible,
			"tenant_editable":    permission.TenantEditable,
			"tenant_edit_scope":  permission.TenantEditScope,
			"data_perm_mode":     "NONE",
		})
	}
	return bundles
}

func (h *IdentityHandler) MenuOverrides(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var rows []models.TenantMenuOverride
	_ = h.tenantScope().Query(user.TenantID).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "tenant_id": row.TenantID, "permission_id": row.PermissionID, "custom_name": row.CustomName, "custom_icon": row.CustomIcon, "enabled": row.Enabled, "visible": row.Visible, "sort_order": row.SortOrder})
	}
	response.OK(c, gin.H{"tenant_id": user.TenantID, "overrides": items})
}

func (h *IdentityHandler) SaveMenuOverrides(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Overrides []struct {
			PermissionID uint64  `json:"permission_id"`
			CustomName   *string `json:"custom_name"`
			CustomIcon   *string `json:"custom_icon"`
			Enabled      *bool   `json:"enabled"`
			Visible      *bool   `json:"visible"`
			SortOrder    *int    `json:"sort_order"`
		} `json:"overrides"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	seen := map[uint64]struct{}{}
	for _, item := range body.Overrides {
		if _, ok := seen[item.PermissionID]; ok {
			response.Error(c, 400, response.CodeBadRequest, "菜单覆盖配置重复")
			return
		}
		seen[item.PermissionID] = struct{}{}
		var permission models.Permission
		if err := h.db.Where("id = ? AND tenant_id IN ? AND perm_type = ? AND deleted_at IS NULL", item.PermissionID, h.permissionScopeTenantIDs(user.TenantID), 3).First(&permission).Error; err != nil {
			response.Error(c, 404, response.CodeNotFound, "菜单不存在")
			return
		}
		if permission.IsPlatformOnly {
			response.Error(c, 403, response.CodeForbidden, "平台专属菜单不可被租户覆盖")
			return
		}
		if !permission.TenantEditable {
			response.Error(c, 403, response.CodeForbidden, "该菜单不允许租户覆盖")
			return
		}
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range body.Overrides {
			var row models.TenantMenuOverride
			err := tx.Where("tenant_id = ? AND permission_id = ?", user.TenantID, item.PermissionID).First(&row).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			row.TenantID = user.TenantID
			row.PermissionID = item.PermissionID
			row.CustomName = nullableTrimmed(item.CustomName)
			row.CustomIcon = nullableTrimmed(item.CustomIcon)
			row.Enabled = item.Enabled
			row.Visible = item.Visible
			row.SortOrder = item.SortOrder
			if row.ID == 0 {
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			} else if err := tx.Save(&row).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateTenantAuthorizationCache(user.TenantID)
	h.audit(c, user.TenantID, user.ID, "menu", "override", "保存租户菜单覆盖", gin.H{"count": len(body.Overrides)})
	h.MenuOverrides(c)
}
