package handlers

import (
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) routeAllowed(user models.AppUser, method, fullPath string) bool {
	required := requiredPermission(method, fullPath)
	if required == "" {
		return routePermissionOptional(method, fullPath)
	}
	if required == "brand:edit" {
		return h.userCanEditTenantBranding(user)
	}
	codes := h.userPermissionCodeSet(user.ID)
	_, ok := codes[required]
	return ok
}

func (h *IdentityHandler) userPermissionCodeSet(userID uint64) map[string]struct{} {
	if h.db == nil {
		return map[string]struct{}{}
	}
	var user models.AppUser
	if err := h.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return map[string]struct{}{}
	}
	codes := map[string]struct{}{}
	filterSubscription := !user.IsPlatformAdmin && !h.viewerHasPlatformScope(user)
	for _, code := range h.cachedPermissionCodesForUser(user, filterSubscription) {
		codes[code] = struct{}{}
	}
	return codes
}

func requiredPermission(method, path string) string {
	key := method + " " + path
	if code, ok := operationPermissionByRoute()[key]; ok {
		return code
	}
	if method == "GET" {
		return menuPermissionByRoute()[path]
	}
	return ""
}

func RequiredPermissionForRoute(method, path string) string {
	return requiredPermission(method, path)
}

func routePermissionOptional(method, path string) bool {
	return optionalRouteByMethodPath()[method+" "+path]
}

func UnclassifiedAPIRoutes(routes gin.RoutesInfo) []string {
	missing := []string{}
	for _, route := range routes {
		if !strings.HasPrefix(route.Path, "/api/") {
			continue
		}
		if routePermissionOptional(route.Method, route.Path) {
			continue
		}
		if requiredPermission(route.Method, route.Path) == "" {
			missing = append(missing, route.Method+" "+route.Path)
		}
	}
	sort.Strings(missing)
	return missing
}

func optionalRouteByMethodPath() map[string]bool {
	return map[string]bool{
		"GET /api/auth/captcha":             true,
		"GET /api/auth/phone-login-tenants": true,
		"POST /api/auth/login":              true,
		"GET /api/public":                   true,
		"GET /api/public/tenant-footer":     true,

		"POST /api/auth/logout":                   true,
		"GET /api/auth/switchable-tenants":        true,
		"POST /api/auth/switch-tenant":            true,
		"GET /api/users/me":                       true,
		"PUT /api/users/me":                       true,
		"PUT /api/users/me/password":              true,
		"GET /api/users/me/preferences":           true,
		"PUT /api/users/me/preferences":           true,
		"GET /api/tenant/branding":                true,
		"GET /api/dict-types/by-code/:code/items": true,
		"GET /api/dict-types/by-code":             true,
		"POST /api/files/upload":                  true,
		"GET /api/files/download/:file_id":        true,
		"DELETE /api/files/:file_id":              true,
	}
}

func operationPermissionByRoute() map[string]string {
	return map[string]string{
		"POST /api/apps":                                "app:create",
		"POST /api/apps/manifest/parse":                 "app:load",
		"POST /api/apps/manifest/diff":                  "app:load",
		"POST /api/apps/manifest/load":                  "app:load",
		"POST /api/apps/manifest/scan":                  "app:load",
		"PUT /api/apps/:id":                             "app:edit",
		"PATCH /api/apps/:id/status":                    "app:status",
		"PUT /api/tenant/branding":                      "brand:edit",
		"POST /api/tenants":                             "tenant:create",
		"POST /api/tenants/with-package":                "tenant:create",
		"PUT /api/tenants/:id":                          "tenant:edit",
		"PATCH /api/tenants/:id/status":                 "tenant:status",
		"PUT /api/tenants/:id/package-config":           "tenant:edit",
		"DELETE /api/tenants/:id":                       "tenant:delete",
		"PUT /api/tenants/:id/primary-admin/password":   "tenant:reset_primary_password",
		"PUT /api/tenants/:id/subscription":             "plan:config",
		"PUT /api/tenants/:id/feature-overrides":        "plan:config",
		"PUT /api/tenants/:id/quota-overrides":          "tenant:quota_config",
		"POST /api/roles":                               "role:create",
		"PUT /api/roles/:id":                            "role:edit",
		"PUT /api/roles/:id/permissions":                "role:permission",
		"DELETE /api/roles/:id":                         "role:delete",
		"PUT /api/permissions/menu-data-perm-mode/:id":  "menu:edit",
		"PUT /api/permissions/menu-data-perm-mode":      "menu:edit",
		"PUT /api/permissions/menu-package-feature/:id": "menu:package_feature",
		"PUT /api/permissions/menu-overrides":           "menu:edit",
		"POST /api/permissions":                         "menu:edit",
		"PUT /api/permissions/:id":                      "menu:edit",
		"DELETE /api/permissions/:id":                   "menu:edit",
		"POST /api/plans":                               "plan:create",
		"PUT /api/plans/:id":                            "plan:edit",
		"POST /api/plans/:id/copy":                      "plan:create",
		"DELETE /api/plans/:id":                         "plan:delete",
		"POST /api/plans/features":                      "plan:create",
		"PUT /api/plans/features/:id":                   "plan:edit",
		"PUT /api/plans/:id/features":                   "plan:config",
		"PUT /api/plans/:id/capabilities":               "plan:config",
		"POST /api/plans/quotas":                        "plan:create",
		"PUT /api/plans/quotas/:id":                     "plan:edit",
		"PUT /api/plans/:id/quotas":                     "plan:config",
		"POST /api/org-nodes":                           "org:create",
		"PUT /api/org-nodes/:id":                        "org:edit",
		"DELETE /api/org-nodes/:id":                     "org:delete",
		"POST /api/companies":                           "org:create",
		"PUT /api/companies/:id":                        "org:edit",
		"DELETE /api/companies/:id":                     "org:delete",
		"POST /api/departments":                         "org:create",
		"PUT /api/departments/:id":                      "org:edit",
		"DELETE /api/departments/:id":                   "org:delete",
		"POST /api/stores":                              "org:create",
		"PUT /api/stores/:id":                           "org:edit",
		"DELETE /api/stores/:id":                        "org:delete",
		"POST /api/position-types":                      "pos:create",
		"PUT /api/position-types/:id":                   "pos:edit",
		"DELETE /api/position-types/:id":                "pos:delete",
		"POST /api/positions":                           "pos:create",
		"PUT /api/positions/:id":                        "pos:edit",
		"DELETE /api/positions/:id":                     "pos:delete",
		"POST /api/business-units":                      "business_unit:create",
		"PUT /api/business-units/:id":                   "business_unit:edit",
		"DELETE /api/business-units/:id":                "business_unit:delete",
		"POST /api/business-units/:id/org-mappings":     "business_unit:edit",
		"DELETE /api/business-units/org-mappings/:id":   "business_unit:edit",
		"POST /api/dict-types":                          "dict_type:create",
		"PUT /api/dict-types/:id":                       "dict_type:edit",
		"DELETE /api/dict-types/:id":                    "dict_type:delete",
		"POST /api/dict-items":                          "dict_item:create",
		"PUT /api/dict-items/:id":                       "dict_item:edit",
		"DELETE /api/dict-items/:id":                    "dict_item:delete",
		"DELETE /api/dict-items/:id/override":           "dict_item:edit",
		"POST /api/sys-params":                          "param:create",
		"PUT /api/sys-params/:id":                       "param:edit",
		"DELETE /api/sys-params/:id":                    "param:delete",
		"DELETE /api/sys-params/:id/override":           "param:edit",
		"POST /api/params":                              "param:create",
		"POST /api/users":                               "user:create",
		"PUT /api/users/:id":                            "user:edit",
		"PUT /api/users/:id/password":                   "user:reset_password",
		"DELETE /api/users/:id":                         "user:delete",
		"POST /api/batch/users/import":                  "user:create",
	}
}

func menuPermissionByRoute() map[string]string {
	return map[string]string{
		"/api/tenants":                                  "/tenants",
		"/api/tenants/:id":                              "/tenants",
		"/api/tenants/:id/companies":                    "/tenants",
		"/api/tenants/:id/quota-records":                "/tenants",
		"/api/tenants/:id/primary-admin":                "/tenants",
		"/api/tenants/:id/subscription":                 "/tenants",
		"/api/tenants/:id/feature-overrides":            "/tenants",
		"/api/tenants/:id/quota-overrides":              "/tenants",
		"/api/tenants/:id/quota-usage":                  "/tenants",
		"/api/tenants/:id/feature-access/:feature_code": "/tenants",
		"/api/tenants/:id/quota-check/:quota_code":      "/tenants",
		"/api/apps":                            "/apps",
		"/api/apps/stats":                      "/apps",
		"/api/apps/manifest/template":          "/apps",
		"/api/apps/:id":                        "/apps",
		"/api/plans":                           "/plans",
		"/api/plans/matrix":                    "/plans",
		"/api/plans/features":                  "/plans",
		"/api/plans/:id/features":              "/plans",
		"/api/plans/quotas":                    "/plans",
		"/api/plans/:id/quotas":                "/plans",
		"/api/organizations/tree":              "/organization",
		"/api/position-types":                  "/positions",
		"/api/positions":                       "/positions",
		"/api/business-units":                  "/business-units",
		"/api/business-units/tree":             "/business-units",
		"/api/business-units/:id/org-mappings": "/business-units",
		"/api/business-units/org-mappings":     "/business-units",
		"/api/users":                           "/users",
		"/api/users/assignable-roles":          "/users",
		"/api/roles":                           "/roles",
		"/api/roles/:id":                       "/roles",
		"/api/roles/permission-menu-bundles":   "/roles",
		"/api/permission-menu-bundles":         "/roles",
		"/api/permissions/menu-bundles":        "/roles",
		"/api/permissions/menu-overrides":      "/menus",
		"/api/permissions":                     "/menus",
		"/api/permissions/tree":                "/menus",
		"/api/permissions/:id":                 "/menus",
		"/api/organizations/detail":            "/organization",
		"/api/dict-types":                      "/dict",
		"/api/dict-items":                      "/dict",
		"/api/sys-params":                      "/params",
		"/api/sys-params/batch":                "/params",
		"/api/params":                          "/params",
		"/api/params/:key":                     "/params",
		"/api/logs/audit":                      "/audit-logs",
		"/api/logs/login":                      "/login-logs",
		"/api/monitor/health-detail":           "/monitor/health",
		"/api/monitor/server-info":             "/monitor/server",
		"/api/monitor/scheduled-jobs":          "/monitor/jobs",
		"/api/monitor/services-overview":       "/monitor/services",
		"/api/monitor/cache-stats":             "/monitor/cache",
		"/api/monitor/cache-keys":              "/monitor/cache-keys",
		"/api/batch/users/export":              "/users",
		"/api/batch/companies/export":          "/organization",
		"/api/batch/departments/export":        "/organization",
	}
}
