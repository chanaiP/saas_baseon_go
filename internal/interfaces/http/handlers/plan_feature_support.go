package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
)

func (h *IdentityHandler) planMatrixQuotaValues(planID uint64, featureCode string) []dto.PlanMatrixQuotaValue {
	quotaCodes := quotaCodesForFeatureCode(featureCode)
	if len(quotaCodes) == 0 {
		return []dto.PlanMatrixQuotaValue{}
	}
	var rows []struct {
		QuotaID    uint64
		QuotaCode  string
		QuotaName  string
		QuotaValue int
	}
	_ = h.db.Table("saas_plan_quota pq").
		Select("pq.quota_id, q.quota_code, q.quota_name, pq.quota_value").
		Joins("join saas_quota q on q.id = pq.quota_id").
		Where("pq.plan_id = ? AND q.quota_code IN ?", planID, quotaCodes).
		Scan(&rows).Error
	items := make([]dto.PlanMatrixQuotaValue, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.PlanMatrixQuotaValue{QuotaID: row.QuotaID, QuotaCode: row.QuotaCode, QuotaName: row.QuotaName, QuotaValue: row.QuotaValue})
	}
	return items
}

func (h *IdentityHandler) syncPackageFeaturesFromPermissions() {
	var permissions []models.Permission
	_ = h.db.Where("enabled = ? AND is_package_feature = ? AND path <> ?", true, true, "__menu_root__").Order("perm_type asc, id asc").Find(&permissions).Error
	codeToID := map[string]uint64{}
	var existing []models.SaasFeature
	_ = h.db.Find(&existing).Error
	for _, feature := range existing {
		codeToID[feature.FeatureCode] = feature.ID
	}
	for _, permission := range permissions {
		code := packageFeatureCodeForPermission(permission)
		if code == "" {
			continue
		}
		featureType := packageFeatureTypeForPermission(permission)
		parentID := uint64(0)
		if featureType == "BUTTON" {
			parentCode := parentPackageFeatureCodeForOperation(permission.Path)
			if parentCode != "" {
				parentID = codeToID[parentCode]
			}
		}
		name := packageFeatureNameForPermission(permission)
		var feature models.SaasFeature
		if err := h.db.Where("feature_code = ?", code).First(&feature).Error; err == nil {
			updates := map[string]interface{}{"feature_name": name, "feature_type": featureType, "parent_id": parentID, "status": 1}
			if permission.PermType == 3 {
				updates["menu_id"] = permission.ID
			}
			_ = h.db.Model(&feature).Updates(updates).Error
			codeToID[code] = feature.ID
			continue
		}
		feature = models.SaasFeature{FeatureCode: code, FeatureName: name, FeatureType: featureType, ParentID: parentID, Status: 1}
		if permission.PermType == 3 {
			feature.MenuID = &permission.ID
		}
		if err := h.db.Create(&feature).Error; err == nil {
			codeToID[code] = feature.ID
		}
	}
}

func (h *IdentityHandler) menuBundleOperations(tenantID uint64, menuPath string, forPlatform bool) []gin.H {
	prefixes := operationPrefixesForMenuPath(menuPath)
	if len(prefixes) == 0 {
		return []gin.H{}
	}
	query := h.db.Where("tenant_id = ? AND perm_type = ? AND deleted_at IS NULL", tenantID, 2)
	parts := make([]string, 0, len(prefixes))
	args := make([]interface{}, 0, len(prefixes))
	for _, prefix := range prefixes {
		parts = append(parts, "path LIKE ?")
		args = append(args, prefix+":%")
	}
	var rows []models.Permission
	_ = query.Where(strings.Join(parts, " OR "), args...).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		if !forPlatform && row.IsPlatformOnly {
			continue
		}
		if !forPlatform && !h.permissionAllowedForTenantSubscription(tenantID, row) {
			continue
		}
		items = append(items, gin.H{"id": row.ID, "path": row.Path, "name": row.Name, "is_platform_only": row.IsPlatformOnly, "is_package_feature": row.IsPackageFeature, "feature_code": row.FeatureCode, "feature_type": row.FeatureType, "tenant_visible": row.Visible, "tenant_editable": row.TenantEditable, "tenant_edit_scope": row.TenantEditScope, "data_perm_mode": row.DataPermMode})
	}
	return items
}

func operationPrefixesForMenuPath(menuPath string) []string {
	mapping := map[string][]string{
		"/users":              {"user"},
		"/organization":       {"org"},
		"/positions":          {"pos"},
		"/roles":              {"role"},
		"/permissions":        {"perm"},
		"/menus":              {"menu"},
		"/dict":               {"dict"},
		"/params":             {"param"},
		"/business-units":     {"business_unit"},
		"/audit-logs":         {"audit"},
		"/login-logs":         {"login"},
		"/monitor/health":     {"monhealth"},
		"/monitor/server":     {"monserver"},
		"/monitor/jobs":       {"monjobs"},
		"/monitor/services":   {"monservices"},
		"/monitor/cache":      {"moncache"},
		"/monitor/cache-keys": {"moncachekeys"},
	}
	return mapping[menuPath]
}

func (h *IdentityHandler) dataPermissionIDForMenu(tenantID uint64, menuPath string) uint64 {
	prefixes := operationPrefixesForMenuPath(menuPath)
	if len(prefixes) == 0 {
		return 0
	}
	var row models.Permission
	if err := h.db.Where("tenant_id = ? AND path = ? AND perm_type = ? AND deleted_at IS NULL", tenantID, "data:"+prefixes[0], 4).First(&row).Error; err != nil {
		return 0
	}
	return row.ID
}

func (h *IdentityHandler) permissionAllowedForTenantSubscription(tenantID uint64, permission models.Permission) bool {
	code := packageFeatureCodeForPermission(permission)
	if code == "" {
		return true
	}
	return h.tenantFeatureAllowed(tenantID, code)
}

func permissionAllowedByFeatureCodeSet(permission models.Permission, allowedFeatures map[string]bool) bool {
	code := packageFeatureCodeForPermission(permission)
	if code == "" {
		return true
	}
	return allowedFeatures[code]
}

func (h *IdentityHandler) requireFeatureAccess(tenantID uint64, featureCode string) error {
	if h.tenantFeatureAllowed(tenantID, featureCode) {
		return nil
	}
	return fmt.Errorf("当前套餐不支持功能：%s", featureCode)
}

func (h *IdentityHandler) disablePackageFeatureForPermission(permission models.Permission) {
	code := packageFeatureCodeForPermission(permission)
	if code == "" {
		return
	}
	_ = h.db.Model(&models.SaasFeature{}).Where("feature_code = ?", code).Updates(map[string]interface{}{"status": 0}).Error
}

func packageFeatureCodeForPermission(permission models.Permission) string {
	if permission.IsPlatformOnly || !permission.IsPackageFeature || permission.Path == "" {
		return ""
	}
	if excludedPackageFeaturePath(permission.Path) {
		return ""
	}
	if permission.FeatureCode != nil && strings.TrimSpace(*permission.FeatureCode) != "" {
		return truncateFeatureCode(strings.TrimSpace(*permission.FeatureCode))
	}
	if override := packageFeatureOverride(permission.Path); override != "" {
		return override
	}
	if permission.PermType != 2 && permission.PermType != 3 && permission.PermType != 5 {
		return ""
	}
	prefix := "button"
	if permission.PermType == 3 {
		prefix = "menu"
	}
	return truncateFeatureCode(prefix + "_" + normalizedFeatureSlug(permission.Path))
}

func excludedPackageFeaturePath(path string) bool {
	excluded := map[string]struct{}{"/tenants": {}, "tenant:create": {}, "tenant:edit": {}, "tenant:delete": {}, "/plans": {}, "plan:create": {}, "plan:edit": {}, "plan:config": {}}
	_, ok := excluded[path]
	return ok
}

func packageFeatureOverride(path string) string {
	overrides := map[string]string{
		"/users": "user_manage", "/organization": "org_manage", "/positions": "position_manage", "/roles": "role_manage", "/permissions": "role_manage", "/menus": "role_manage", "/dict": "dict_manage", "/params": "param_manage", "/business-units": "business_unit_manage", "/audit-logs": "audit_log", "/login-logs": "login_log", "/monitor/health": "system_monitor", "/monitor/server": "system_monitor", "/monitor/jobs": "system_monitor", "/monitor/services": "system_monitor", "/monitor/cache": "system_monitor", "/monitor/cache-keys": "system_monitor", "brand:edit": "brand_config",
	}
	return overrides[path]
}

func parentPackageFeatureCodeForOperation(path string) string {
	prefix := strings.SplitN(path, ":", 2)[0]
	parentByPrefix := map[string]string{"user": "user_manage", "org": "org_manage", "role": "role_manage", "perm": "role_manage", "menu": "role_manage", "pos": "position_manage", "dict": "dict_manage", "param": "param_manage", "business_unit": "business_unit_manage", "audit": "audit_log", "login": "login_log", "monhealth": "system_monitor", "monserver": "system_monitor", "monjobs": "system_monitor", "monservices": "system_monitor", "moncache": "system_monitor", "moncachekeys": "system_monitor", "brand": "brand_config"}
	return parentByPrefix[prefix]
}

func packageFeatureTypeForPermission(permission models.Permission) string {
	if permission.FeatureType != nil && strings.TrimSpace(*permission.FeatureType) != "" {
		return strings.TrimSpace(*permission.FeatureType)
	}
	if permission.PermType == 3 {
		return "MENU"
	}
	return "BUTTON"
}

func packageFeatureNameForPermission(permission models.Permission) string {
	name := strings.TrimSpace(permission.Name)
	if strings.HasSuffix(name, "-访问") {
		name = strings.TrimSuffix(name, "-访问")
	}
	if name == "" {
		return permission.Path
	}
	return name
}

func normalizedFeatureSlug(value string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, ch := range strings.ToLower(value) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func truncateFeatureCode(value string) string {
	if len(value) > 100 {
		return value[:100]
	}
	return value
}

func quotaCodesForFeatureCode(featureCode string) []string {
	mapping := map[string][]string{
		"user_manage":          {"max_users"},
		"org_manage":           {"max_companies", "max_stores", "max_departments"},
		"business_unit_manage": {"max_business_units"},
		"role_manage":          {"max_roles"},
		"import_data":          {"daily_import_times"},
		"export_data":          {"daily_export_times"},
		"api_key":              {"max_api_keys", "daily_api_calls"},
		"webhook":              {"max_webhooks"},
	}
	return mapping[featureCode]
}
