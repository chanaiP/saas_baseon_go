package permissioncatalog

import "strings"

var tenantPackageMenuFeatures = map[string]string{
	"/organization":   "org_manage",
	"/positions":      "position_manage",
	"/business-units": "business_unit_manage",
	"/users":          "user_manage",
	"/roles":          "role_manage",
	"/menus":          "menu_manage",
	"/dict":           "dict_manage",
	"/params":         "param_manage",
	"/audit-logs":     "audit_log",
	"/login-logs":     "login_log",
}

var featureOverrides = map[string]string{
	"/users":          "user_manage",
	"/organization":   "org_manage",
	"/positions":      "position_manage",
	"/roles":          "role_manage",
	"/permissions":    "role_manage",
	"/menus":          "menu_manage",
	"/dict":           "dict_manage",
	"/params":         "param_manage",
	"/business-units": "business_unit_manage",
	"/audit-logs":     "audit_log",
	"/login-logs":     "login_log",
	"brand:edit":      "brand_config",
}

var operationParentByPrefix = map[string]string{
	"user":          "user_manage",
	"org":           "org_manage",
	"pos":           "position_manage",
	"role":          "role_manage",
	"menu":          "menu_manage",
	"dict":          "dict_manage",
	"dict_item":     "dict_manage",
	"dict_type":     "dict_manage",
	"param":         "param_manage",
	"business_unit": "business_unit_manage",
	"audit":         "audit_log",
	"login":         "login_log",
	"brand":         "brand_config",
}

func IsTenantPackageMenuPath(path string) bool {
	_, ok := tenantPackageMenuFeatures[strings.TrimSpace(path)]
	return ok
}

func MenuFeatureCode(path string) string {
	return tenantPackageMenuFeatures[strings.TrimSpace(path)]
}

func FeatureOverride(path string) string {
	return featureOverrides[strings.TrimSpace(path)]
}

func IsPlatformOnlyOperation(path string) bool {
	value := strings.TrimSpace(path)
	return strings.HasPrefix(value, "tenant:") ||
		strings.HasPrefix(value, "plan:") ||
		strings.HasPrefix(value, "dict_type:") ||
		strings.HasPrefix(value, "perm:") ||
		value == "menu:delete" ||
		value == "menu:package_feature"
}

func IsPureViewOperation(path string) bool {
	return strings.HasSuffix(strings.TrimSpace(path), ":view")
}

func ExcludedPackageFeaturePath(path string) bool {
	value := strings.TrimSpace(path)
	if value == "" || IsPureViewOperation(value) || IsPlatformOnlyOperation(value) {
		return true
	}
	if strings.HasPrefix(value, "/monitor/") {
		return true
	}
	switch value {
	case "/home", "/tenants", "/plans", "/permissions", "home:view", "dict:create", "dict:edit", "dict:delete":
		return true
	}
	return false
}

func IsPackageFeatureOperation(path string) bool {
	value := strings.TrimSpace(path)
	if ExcludedPackageFeaturePath(value) {
		return false
	}
	if value == "brand:edit" {
		return true
	}
	return ParentFeatureCodeForOperation(value) != ""
}

func OperationFeatureCode(path string) string {
	if strings.TrimSpace(path) == "brand:edit" {
		return "brand_config"
	}
	return ""
}

func ParentFeatureCodeForOperation(path string) string {
	return operationParentByPrefix[OperationParentPrefix(path)]
}

func OperationParentPrefix(path string) string {
	value := strings.TrimSpace(path)
	if prefix, _, ok := strings.Cut(value, ":"); ok {
		return prefix
	}
	for _, multiPartPrefix := range []string{"business_unit", "dict_type", "dict_item"} {
		if strings.HasPrefix(value, multiPartPrefix+"_") {
			return multiPartPrefix
		}
	}
	prefix, _, _ := strings.Cut(value, "_")
	return prefix
}

func OperationName(path string) string {
	names := map[string]string{
		"org:create":                    "组织-新增",
		"org:edit":                      "组织-编辑",
		"org:delete":                    "组织-删除",
		"pos:create":                    "岗位-新增",
		"pos:edit":                      "岗位-编辑",
		"pos:delete":                    "岗位-删除",
		"business_unit:create":          "业务单元-新增",
		"business_unit:edit":            "业务单元-编辑",
		"business_unit:delete":          "业务单元-删除",
		"user:create":                   "用户-新增",
		"user:edit":                     "用户-编辑",
		"user:reset_password":           "用户-重置密码",
		"user:delete":                   "用户-删除",
		"role:create":                   "角色-新增",
		"role:edit":                     "角色-编辑",
		"role:delete":                   "角色-删除",
		"role:permission":               "角色-权限设置",
		"menu:create":                   "菜单-新增",
		"menu:edit":                     "菜单-编辑",
		"menu:delete":                   "菜单-删除",
		"menu:package_feature":          "菜单-套餐中心收录",
		"dict_type:create":              "字典类型-新增",
		"dict_type:edit":                "字典类型-编辑",
		"dict_type:delete":              "字典类型-删除",
		"dict_item:create":              "字典项-新增",
		"dict_item:edit":                "字典项-编辑",
		"dict_item:delete":              "字典项-删除",
		"param:create":                  "参数-新增",
		"param:edit":                    "参数-编辑",
		"param:delete":                  "参数-删除",
		"tenant:status":                 "主体-启停",
		"tenant:reset_primary_password": "主体-重置主管理员密码",
		"brand:edit":                    "品牌-维护",
	}
	if name, ok := names[strings.TrimSpace(path)]; ok {
		return name
	}
	return path
}

func OperationLabel(path string) string {
	value := strings.TrimSpace(path)
	prefix, action, ok := strings.Cut(value, ":")
	if !ok {
		for _, multiPartPrefix := range []string{"business_unit", "dict_type", "dict_item"} {
			if strings.HasPrefix(value, multiPartPrefix+"_") {
				prefix = multiPartPrefix
				action = strings.TrimPrefix(value, multiPartPrefix+"_")
				ok = true
				break
			}
		}
		if !ok {
			prefix, action, ok = strings.Cut(value, "_")
		}
	}
	prefixLabel := map[string]string{
		"user":          "用户",
		"org":           "组织",
		"pos":           "岗位",
		"role":          "角色",
		"menu":          "菜单",
		"dict":          "字典",
		"dict_type":     "字典类型",
		"dict_item":     "字典项",
		"param":         "参数",
		"business_unit": "业务单元",
		"audit":         "操作日志",
		"login":         "登录日志",
		"brand":         "品牌",
	}
	actionLabel := map[string]string{
		"create":                 "新增",
		"edit":                   "编辑",
		"delete":                 "删除",
		"status":                 "启停",
		"reset_password":         "重置密码",
		"reset_primary_password": "重置主管理员密码",
		"quota_config":           "调整配额",
		"permission":             "权限设置",
	}
	if ok {
		if label, exists := prefixLabel[prefix]; exists {
			if act, exists := actionLabel[action]; exists {
				return label + "-" + act
			}
		}
	}
	return value
}
