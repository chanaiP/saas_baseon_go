package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/domain/permissioncatalog"
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

type planMatrixMenuSpec struct {
	Code  string
	Label string
}

type planMatrixSyntheticMenuSpec struct {
	ID     string
	Label  string
	Prefix string
}

type planMatrixDomainSpec struct {
	ID       string
	Label    string
	Subtitle string
	Menus    []planMatrixMenuSpec
}

func (h *IdentityHandler) buildPlanMatrixNodes(plans []models.SaasPlan, features []models.SaasFeature, enabled map[uint64]map[uint64]bool) []dto.PlanMatrixNode {
	byCode := make(map[string]models.SaasFeature, len(features))
	byID := make(map[uint64]models.SaasFeature, len(features))
	for _, feature := range features {
		byCode[feature.FeatureCode] = feature
		byID[feature.ID] = feature
	}

	used := map[string]bool{}
	opsByParent := map[string][]models.SaasFeature{}
	opsByPrefix := map[string][]models.SaasFeature{}
	for _, feature := range features {
		parentCode := operationParentFeatureCode(feature, byID)
		if parentCode == "" {
			continue
		}
		opsByParent[parentCode] = append(opsByParent[parentCode], feature)
		if prefix := operationFeaturePrefix(feature); prefix != "" {
			opsByPrefix[prefix] = append(opsByPrefix[prefix], feature)
		}
	}
	for parentCode := range opsByParent {
		sortPlanMatrixFeatures(opsByParent[parentCode])
	}
	for prefix := range opsByPrefix {
		sortPlanMatrixFeatures(opsByPrefix[prefix])
	}

	specs := []planMatrixDomainSpec{
		{
			ID:       "domain-system-management",
			Label:    "系统管理",
			Subtitle: "目录",
			Menus: []planMatrixMenuSpec{
				{Code: "user_manage", Label: "用户管理"},
				{Code: "org_manage", Label: "组织架构"},
				{Code: "position_manage", Label: "岗位管理"},
				{Code: "role_manage", Label: "角色权限"},
				{Code: "menu_manage", Label: "菜单管理"},
				{Code: "dict_manage", Label: "数据字典"},
				{Code: "param_manage", Label: "参数管理"},
				{Code: "business_unit_manage", Label: "业务单元"},
				{Code: "login_log", Label: "登录日志"},
				{Code: "audit_log", Label: "操作日志"},
			},
		},
		{
			ID:       "domain-system-monitor",
			Label:    "系统监控",
			Subtitle: "目录",
		},
		{
			ID:       "domain-open-capability",
			Label:    "开放与安全能力",
			Subtitle: "目录",
			Menus: []planMatrixMenuSpec{
				{Code: "file_manage", Label: "文件管理"},
				{Code: "brand_config", Label: "品牌配置"},
				{Code: "import_data", Label: "数据导入"},
				{Code: "export_data", Label: "数据导出"},
			},
		},
	}

	nodes := make([]dto.PlanMatrixNode, 0, len(specs)+1)
	for _, spec := range specs {
		children := make([]dto.PlanMatrixNode, 0, len(spec.Menus))
		if spec.ID == "domain-system-monitor" {
			children = append(children, h.planMatrixMonitorMenuNodes(plans, byCode["system_monitor"], opsByPrefix, used, enabled)...)
		}
		for _, menu := range spec.Menus {
			feature, ok := byCode[menu.Code]
			if !ok {
				continue
			}
			used[feature.FeatureCode] = true
			menuNode := h.planMatrixFeatureNode(plans, feature, menu.Label, enabled)
			for _, op := range opsByParent[feature.FeatureCode] {
				used[op.FeatureCode] = true
				menuNode.Children = append(menuNode.Children, h.planMatrixFeatureNode(plans, op, operationFeatureLabel(op), enabled))
			}
			children = append(children, menuNode)
		}
		if len(children) == 0 {
			continue
		}
		nodes = append(nodes, dto.PlanMatrixNode{
			ID:       spec.ID,
			Label:    spec.Label,
			NodeType: "domain",
			Children: children,
			Cells:    h.planMatrixStructuralCells(plans),
		})
	}

	otherChildren := []dto.PlanMatrixNode{}
	for _, feature := range features {
		if excludedPlanMatrixFeatureRow(feature) {
			continue
		}
		if used[feature.FeatureCode] || operationParentFeatureCode(feature, byID) != "" || isPlanMatrixOperationFeature(feature) {
			continue
		}
		otherChildren = append(otherChildren, h.planMatrixFeatureNode(plans, feature, "", enabled))
	}
	if len(otherChildren) > 0 {
		nodes = append(nodes, dto.PlanMatrixNode{
			ID:       "domain-other",
			Label:    "其他能力",
			NodeType: "domain",
			Children: otherChildren,
			Cells:    h.planMatrixStructuralCells(plans),
		})
	}
	return nodes
}

func excludedPlanMatrixFeature(code string) bool {
	excluded := map[string]struct{}{
		"home":           {},
		"tenant_manage":  {},
		"plan_manage":    {},
		"system_monitor": {},
	}
	if _, ok := excluded[code]; ok {
		return true
	}
	return uncontrolledPackageFeatureCode(code) || reservedPackageFeatureCode(code)
}

func excludedPlanMatrixFeatureRow(feature models.SaasFeature) bool {
	return excludedPlanMatrixFeature(feature.FeatureCode)
}

func uncontrolledPackageFeatureCode(code string) bool {
	uncontrolled := map[string]struct{}{
		"data_permission":          {},
		"advanced_data_permission": {},
	}
	_, ok := uncontrolled[code]
	return ok
}

func reservedPackageFeatureCode(code string) bool {
	reserved := map[string]struct{}{
		"api_key":            {},
		"webhook":            {},
		"tenant_data_export": {},
		"ip_whitelist":       {},
		"mfa":                {},
		"sso_login":          {},
	}
	_, ok := reserved[code]
	return ok
}

func hiddenPackageQuotaCode(code string) bool {
	hidden := map[string]struct{}{
		"max_api_keys":        {},
		"max_webhooks":        {},
		"daily_api_calls":     {},
		"monthly_sms_count":   {},
		"monthly_email_count": {},
	}
	_, ok := hidden[code]
	return ok
}

func (h *IdentityHandler) planMatrixMonitorMenuNodes(plans []models.SaasPlan, monitorFeature models.SaasFeature, opsByPrefix map[string][]models.SaasFeature, used map[string]bool, enabled map[uint64]map[uint64]bool) []dto.PlanMatrixNode {
	if monitorFeature.ID == 0 {
		return []dto.PlanMatrixNode{}
	}
	used[monitorFeature.FeatureCode] = true
	menuSpecs := []planMatrixSyntheticMenuSpec{
		{ID: "monitor_health", Label: "健康检查", Prefix: "monhealth"},
		{ID: "monitor_server", Label: "服务器信息", Prefix: "monserver"},
		{ID: "monitor_jobs", Label: "定时任务", Prefix: "monjobs"},
		{ID: "monitor_services", Label: "服务监控", Prefix: "monservices"},
		{ID: "monitor_cache", Label: "缓存监控", Prefix: "moncache"},
		{ID: "monitor_cache_keys", Label: "缓存列表", Prefix: "moncachekeys"},
	}
	children := make([]dto.PlanMatrixNode, 0, len(menuSpecs))
	for _, spec := range menuSpecs {
		menuNode := h.planMatrixSyntheticFeatureNode(plans, monitorFeature, spec.ID, spec.Label, "MENU", enabled)
		for _, op := range opsByPrefix[spec.Prefix] {
			used[op.FeatureCode] = true
			menuNode.Children = append(menuNode.Children, h.planMatrixFeatureNode(plans, op, operationFeatureLabel(op), enabled))
		}
		children = append(children, menuNode)
	}
	return children
}

func (h *IdentityHandler) planMatrixSyntheticFeatureNode(plans []models.SaasPlan, feature models.SaasFeature, id string, label string, featureType string, enabled map[uint64]map[uint64]bool) dto.PlanMatrixNode {
	node := h.planMatrixFeatureNode(plans, feature, label, enabled)
	node.ID = id
	node.FeatureType = featureType
	return node
}

func (h *IdentityHandler) planMatrixFeatureNode(plans []models.SaasPlan, feature models.SaasFeature, label string, enabled map[uint64]map[uint64]bool) dto.PlanMatrixNode {
	if strings.TrimSpace(label) == "" {
		label = feature.FeatureName
	}
	featureCode := feature.FeatureCode
	featureType := normalizedPlanMatrixFeatureType(feature.FeatureType)
	if feature.FeatureCode == "brand_config" {
		featureType = "CONFIG"
	}
	return dto.PlanMatrixNode{
		ID:          feature.FeatureCode,
		Label:       label,
		NodeType:    "feature",
		FeatureID:   feature.ID,
		FeatureCode: featureCode,
		FeatureType: featureType,
		Description: feature.Description,
		Children:    []dto.PlanMatrixNode{},
		Cells:       h.planMatrixFeatureCells(plans, feature, enabled),
	}
}

func (h *IdentityHandler) planMatrixFeatureCells(plans []models.SaasPlan, feature models.SaasFeature, enabled map[uint64]map[uint64]bool) []dto.PlanMatrixCell {
	cells := make([]dto.PlanMatrixCell, 0, len(plans))
	for _, plan := range plans {
		isEnabled := enabled[feature.ID][plan.ID]
		state := "disabled"
		if isEnabled {
			state = "enabled"
		}
		featureIDs := []uint64{}
		if isEnabled {
			featureIDs = []uint64{feature.ID}
		}
		cells = append(cells, dto.PlanMatrixCell{
			PlanID:      plan.ID,
			PlanCode:    plan.PlanCode,
			Enabled:     isEnabled,
			State:       state,
			FeatureIDs:  featureIDs,
			QuotaValues: h.planMatrixQuotaValues(plan.ID, feature.FeatureCode),
		})
	}
	return cells
}

func (h *IdentityHandler) planMatrixStructuralCells(plans []models.SaasPlan) []dto.PlanMatrixCell {
	cells := make([]dto.PlanMatrixCell, 0, len(plans))
	for _, plan := range plans {
		cells = append(cells, dto.PlanMatrixCell{PlanID: plan.ID, PlanCode: plan.PlanCode, State: "disabled", FeatureIDs: []uint64{}, QuotaValues: []dto.PlanMatrixQuotaValue{}})
	}
	return cells
}

func normalizedPlanMatrixFeatureType(value string) string {
	if strings.EqualFold(value, "OPERATION") {
		return "BUTTON"
	}
	return value
}

func operationParentFeatureCode(feature models.SaasFeature, byID map[uint64]models.SaasFeature) string {
	if feature.ParentID > 0 {
		if parent, ok := byID[feature.ParentID]; ok && parent.ID != feature.ID {
			return parent.FeatureCode
		}
	}
	return legacyOperationParentFeatureCode(feature)
}

func legacyOperationParentFeatureCode(feature models.SaasFeature) string {
	value := strings.TrimSpace(feature.FeatureName)
	if strings.Contains(value, ":") {
		parent := parentPackageFeatureCodeForOperation(value)
		if parent != "" && parent != feature.FeatureCode {
			return parent
		}
		return ""
	}
	if strings.HasPrefix(feature.FeatureCode, "button_") {
		parent := parentPackageFeatureCodeForOperation(strings.TrimPrefix(feature.FeatureCode, "button_"))
		if parent != "" && parent != feature.FeatureCode {
			return parent
		}
		return ""
	}
	return ""
}

func isPlanMatrixOperationFeature(feature models.SaasFeature) bool {
	featureType := normalizedPlanMatrixFeatureType(strings.TrimSpace(feature.FeatureType))
	if strings.EqualFold(featureType, "BUTTON") {
		return true
	}
	return strings.HasPrefix(strings.TrimSpace(feature.FeatureCode), "button_") || strings.Contains(strings.TrimSpace(feature.FeatureName), ":")
}

func sortPlanMatrixFeatures(features []models.SaasFeature) {
	sort.SliceStable(features, func(i, j int) bool {
		if features[i].ParentID != features[j].ParentID {
			return features[i].ParentID < features[j].ParentID
		}
		if operationPermissionSortKey(strings.TrimPrefix(features[i].FeatureCode, "button_"), []string{}) != operationPermissionSortKey(strings.TrimPrefix(features[j].FeatureCode, "button_"), []string{}) {
			return operationPermissionSortKey(strings.TrimPrefix(features[i].FeatureCode, "button_"), []string{}) < operationPermissionSortKey(strings.TrimPrefix(features[j].FeatureCode, "button_"), []string{})
		}
		return features[i].ID < features[j].ID
	})
}

func operationFeaturePrefix(feature models.SaasFeature) string {
	raw := strings.TrimSpace(feature.FeatureName)
	if raw == "" || !strings.Contains(raw, ":") {
		raw = strings.TrimPrefix(feature.FeatureCode, "button_")
	}
	prefix, _, ok := strings.Cut(raw, ":")
	if !ok {
		prefix, _, ok = strings.Cut(raw, "_")
	}
	if !ok {
		return ""
	}
	return prefix
}

func operationFeatureLabel(feature models.SaasFeature) string {
	raw := strings.TrimSpace(feature.FeatureName)
	if raw == "" || !strings.Contains(raw, ":") {
		raw = strings.TrimPrefix(feature.FeatureCode, "button_")
	}
	prefix, action, ok := strings.Cut(raw, ":")
	if !ok {
		for _, multiPartPrefix := range []string{"business_unit", "dict_type", "dict_item"} {
			if strings.HasPrefix(raw, multiPartPrefix+"_") {
				prefix = multiPartPrefix
				action = strings.TrimPrefix(raw, multiPartPrefix+"_")
				ok = true
				break
			}
		}
		if !ok {
			prefix, action, ok = strings.Cut(raw, "_")
		}
	}
	if ok {
		if label := permissioncatalog.OperationLabel(prefix + ":" + action); label != prefix+":"+action {
			return label
		}
	}
	return feature.FeatureName
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
	syncedCode := map[string]bool{}
	for _, permission := range permissions {
		code := packageFeatureCodeForPermission(permission)
		if code == "" {
			continue
		}
		if syncedCode[code] {
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
		feature, found := h.packageFeatureForPermission(permission, code)
		if found {
			updates := map[string]interface{}{"feature_code": code, "feature_name": name, "feature_type": featureType, "parent_id": parentID, "status": 1}
			if permission.PermType == 3 {
				updates["menu_id"] = permission.ID
			}
			_ = h.db.Model(&feature).Updates(updates).Error
			codeToID[code] = feature.ID
			syncedCode[code] = true
			continue
		}
		feature = models.SaasFeature{FeatureCode: code, FeatureName: name, FeatureType: featureType, ParentID: parentID, Status: 1}
		if permission.PermType == 3 {
			feature.MenuID = &permission.ID
		}
		if err := h.db.Create(&feature).Error; err == nil {
			codeToID[code] = feature.ID
		}
		syncedCode[code] = true
	}
	h.repairPackageFeatureParentIDs()
	h.disableStaleMenuPackageFeatures()
}

func (h *IdentityHandler) repairPackageFeatureParentIDs() {
	var features []models.SaasFeature
	if err := h.db.Find(&features).Error; err != nil {
		return
	}
	byCode := make(map[string]models.SaasFeature, len(features))
	for _, feature := range features {
		byCode[feature.FeatureCode] = feature
	}
	for _, feature := range features {
		if !isPlanMatrixOperationFeature(feature) {
			continue
		}
		parentCode := legacyOperationParentFeatureCode(feature)
		if parentCode == "" {
			continue
		}
		parent, ok := byCode[parentCode]
		if !ok || parent.ID == 0 || parent.ID == feature.ID || feature.ParentID == parent.ID {
			continue
		}
		_ = h.db.Model(&models.SaasFeature{}).Where("id = ?", feature.ID).Update("parent_id", parent.ID).Error
	}
}

func (h *IdentityHandler) packageFeatureForPermission(permission models.Permission, code string) (models.SaasFeature, bool) {
	var feature models.SaasFeature
	if permission.PermType == 3 {
		if err := h.db.Where("menu_id = ?", permission.ID).First(&feature).Error; err == nil {
			return feature, true
		}
	}
	if err := h.db.Where("feature_code = ?", code).First(&feature).Error; err == nil {
		return feature, true
	}
	return models.SaasFeature{}, false
}

func (h *IdentityHandler) disableStaleMenuPackageFeatures() {
	var features []models.SaasFeature
	_ = h.db.Where("menu_id IS NOT NULL AND status = ?", 1).Find(&features).Error
	for _, feature := range features {
		if feature.MenuID == nil {
			continue
		}
		var count int64
		_ = h.db.Model(&models.Permission{}).
			Where("id = ? AND enabled = ? AND is_package_feature = ? AND deleted_at IS NULL", *feature.MenuID, true, true).
			Count(&count).Error
		if count == 0 {
			_ = h.db.Model(&models.SaasFeature{}).Where("id = ?", feature.ID).Update("status", 0).Error
		}
	}
}

func (h *IdentityHandler) menuBundleOperations(tenantID uint64, menuPath string, forPlatform bool) []gin.H {
	prefixes := operationPrefixesForMenuPath(menuPath)
	if len(prefixes) == 0 {
		return []gin.H{}
	}
	query := h.db.Where("tenant_id IN ? AND perm_type = ? AND enabled = ? AND visible = ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(tenantID), 2, true, true)
	parts := make([]string, 0, len(prefixes))
	args := make([]interface{}, 0, len(prefixes))
	for _, prefix := range prefixes {
		parts = append(parts, "path LIKE ?")
		args = append(args, prefix+":%")
	}
	var rows []models.Permission
	_ = query.Where(strings.Join(parts, " OR "), args...).Order("id asc").Find(&rows).Error
	sort.SliceStable(rows, func(i, j int) bool {
		return operationPermissionSortKey(rows[i].Path, prefixes) < operationPermissionSortKey(rows[j].Path, prefixes)
	})
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		if isPureViewPermissionPath(row.Path) {
			continue
		}
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

func operationPermissionSortKey(path string, prefixes []string) int {
	prefix, action, _ := strings.Cut(path, ":")
	prefixRank := len(prefixes) + 1
	for i, candidate := range prefixes {
		if candidate == prefix {
			prefixRank = i
			break
		}
	}
	actionRank := map[string]int{"view": 0, "create": 1, "edit": 2, "delete": 3, "status": 4, "reset_password": 5, "reset_primary_password": 6}
	rank, ok := actionRank[action]
	if !ok {
		rank = 99
	}
	return prefixRank*1000 + rank
}

func isPureViewPermissionPath(path string) bool {
	return permissioncatalog.IsPureViewOperation(path)
}

func operationPrefixesForMenuPath(menuPath string) []string {
	mapping := map[string][]string{
		"/tenants":            {"tenant"},
		"/plans":              {"plan"},
		"/users":              {"user"},
		"/organization":       {"org"},
		"/positions":          {"pos"},
		"/roles":              {"role"},
		"/menus":              {"menu"},
		"/dict":               {"dict_type", "dict_item", "dict"},
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
	if menuPath == "/dict" {
		prefixes = append(prefixes, "dict")
	}
	if len(prefixes) == 0 {
		return 0
	}
	for _, prefix := range prefixes {
		var row models.Permission
		if err := h.db.Where("tenant_id IN ? AND path = ? AND perm_type = ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(tenantID), "data:"+prefix, 4).First(&row).Error; err == nil {
			return row.ID
		}
	}
	return 0
}

func (h *IdentityHandler) permissionAllowedForTenantSubscription(tenantID uint64, permission models.Permission) bool {
	code := packageFeatureCodeForPermission(permission)
	if code == "" {
		return true
	}
	if uncontrolledPackageFeatureCode(code) {
		return true
	}
	return h.tenantFeatureAllowed(tenantID, code)
}

func permissionAllowedByFeatureCodeSet(permission models.Permission, allowedFeatures map[string]bool) bool {
	code := packageFeatureCodeForPermission(permission)
	if code == "" {
		code = operationAccessFeatureCode(permission)
	}
	if code == "" {
		return true
	}
	if uncontrolledPackageFeatureCode(code) {
		return true
	}
	return allowedFeatures[code]
}

func operationAccessFeatureCode(permission models.Permission) string {
	if permission.PermType != 2 || permission.Path == "" {
		return ""
	}
	if isPureViewPermissionPath(permission.Path) {
		return ""
	}
	if override := packageFeatureOverride(permission.Path); override != "" {
		return override
	}
	return parentPackageFeatureCodeForOperation(permission.Path)
}

func (h *IdentityHandler) requireFeatureAccess(tenantID uint64, featureCode string) error {
	if h.tenantFeatureAllowed(tenantID, featureCode) {
		return nil
	}
	return fmt.Errorf("当前套餐不支持功能：%s", featureCode)
}

func (h *IdentityHandler) disablePackageFeatureForPermission(permission models.Permission) {
	permission.IsPackageFeature = true
	code := packageFeatureCodeForPermission(permission)
	if code == "" {
		return
	}
	_ = h.db.Model(&models.SaasFeature{}).Where("feature_code = ?", code).Updates(map[string]interface{}{"status": 0}).Error
}

func packageFeatureCodeForPermission(permission models.Permission) string {
	if permission.IsPlatformOnly || permission.Path == "" {
		return ""
	}
	if isPureViewPermissionPath(permission.Path) {
		return ""
	}
	if excludedPackageFeaturePath(permission.Path) {
		return ""
	}
	if !permission.IsPackageFeature {
		return ""
	}
	if permission.PermType == 3 && !permissioncatalog.IsTenantPackageMenuPath(permission.Path) {
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
	return permissioncatalog.ExcludedPackageFeaturePath(path)
}

func packageFeatureOverride(path string) string {
	return permissioncatalog.FeatureOverride(path)
}

func parentPackageFeatureCodeForOperation(path string) string {
	return permissioncatalog.ParentFeatureCodeForOperation(path)
}

func operationParentPrefix(path string) string {
	return permissioncatalog.OperationParentPrefix(path)
}

func packageFeatureTypeForPermission(permission models.Permission) string {
	if permission.Path == "brand:edit" {
		return "CONFIG"
	}
	if permission.FeatureType != nil && strings.TrimSpace(*permission.FeatureType) != "" {
		return strings.TrimSpace(*permission.FeatureType)
	}
	if permission.PermType == 3 {
		return "MENU"
	}
	return "BUTTON"
}

func packageFeatureNameForPermission(permission models.Permission) string {
	if permission.Path == "brand:edit" {
		return "品牌配置"
	}
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
		"file_manage":          {"max_storage_gb", "max_file_size_mb"},
	}
	return mapping[featureCode]
}
