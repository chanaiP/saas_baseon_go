package handlers

import (
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func (h *IdentityHandler) routeAllowed(user models.AppUser, method, fullPath string) bool {
	return h.routeAllowedForRequest(user, method, fullPath, "")
}

func (h *IdentityHandler) routeAllowedForRequest(user models.AppUser, method, fullPath, resource string) bool {
	if routeMetadataReadOptional(method, fullPath) {
		return true
	}
	required := requiredPermissionForRequest(method, fullPath, resource)
	if required == "" {
		return routePermissionOptional(method, fullPath)
	}
	if required == "brand:edit" {
		return h.userCanEditTenantBranding(user)
	}
	if !h.routeScopeAllowed(user, required) {
		return false
	}
	codes := h.userPermissionCodeSet(user.ID)
	_, ok := codes[required]
	return ok
}

func routeMetadataReadOptional(method, path string) bool {
	if method != "GET" {
		return false
	}
	switch path {
	case "/api/roles/permission-menu-bundles", "/api/permission-menu-bundles", "/api/permissions/menu-bundles":
		return true
	default:
		return false
	}
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
	return requiredPermissionForRequest(method, path, "")
}

func requiredPermissionForRequest(method, path, resource string) string {
	if method == "GET" && path == "/api/ai-capability-center/:resource" {
		if strings.TrimSpace(resource) == "" {
			return menuPermissionByRoute()[path]
		}
		return aiCapabilityResourceMenuPermission(resource)
	}
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

func RequiredPermissionForResourceRoute(method, path, resource string) string {
	return requiredPermissionForRequest(method, path, resource)
}

func aiCapabilityResourceMenuPermission(resource string) string {
	switch resource {
	case "providers", "accounts", "apis", "capabilities":
		return "/ai-capability-center/providers"
	case "models", "price-policies", "price-tiers":
		return "/ai-capability-center/models"
	case "scenarios":
		return "/ai-capability-center/scenarios"
	case "base-routes", "route-models":
		return "/ai-capability-center/routes"
	case "tenant-strategies", "quota-rules", "rate-limit-rules":
		return "/ai-capability-center/strategy"
	case "usage-records":
		return "/ai-capability-center/usage-logs"
	case "settings":
		return "/ai-capability-center/settings"
	default:
		return ""
	}
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
		"GET /api/auth/captcha":                                         true,
		"GET /api/auth/phone-login-tenants":                             true,
		"POST /api/auth/register":                                       true,
		"POST /api/auth/login":                                          true,
		"GET /api/integration-center/oauth/callback/:provider_app_code": true,
		"POST /api/integration-center/webhooks/:provider_app_code":      true,
		"GET /api/public":                                               true,
		"GET /api/public/tenant-footer":                                 true,

		"POST /api/auth/logout":                   true,
		"GET /api/auth/switchable-tenants":        true,
		"POST /api/auth/switch-tenant":            true,
		"GET /api/users/me":                       true,
		"PUT /api/users/me":                       true,
		"PUT /api/users/me/password":              true,
		"GET /api/users/me/preferences":           true,
		"PUT /api/users/me/preferences":           true,
		"GET /api/tenant/branding":                true,
		"GET /api/roles/permission-menu-bundles":  true,
		"GET /api/permission-menu-bundles":        true,
		"GET /api/permissions/menu-bundles":       true,
		"GET /api/dict-types/by-code/:code/items": true,
		"GET /api/dict-types/by-code":             true,
		"POST /api/files/upload":                  true,
		"GET /api/files/download/:file_id":        true,
		"DELETE /api/files/:file_id":              true,
	}
}

func operationPermissionByRoute() map[string]string {
	return map[string]string{
		"POST /api/apps":                                                  "app:create",
		"POST /api/apps/manifest/parse":                                   "app:load",
		"POST /api/apps/manifest/diff":                                    "app:load",
		"POST /api/apps/manifest/load":                                    "app:load",
		"POST /api/apps/manifest/scan":                                    "app:load",
		"PUT /api/apps/:id":                                               "app:edit",
		"PATCH /api/apps/:id/status":                                      "app:status",
		"POST /api/ai-capability-center/providers/import":                 "ai_capability_center:manage",
		"POST /api/ai-capability-center/models/import":                    "ai_capability_center:manage",
		"POST /api/ai-capability-center/scenarios/import":                 "ai_capability_center:manage",
		"POST /api/ai-capability-center/routes/import":                    "ai_capability_center:manage",
		"POST /api/ai-capability-center/tenant-strategies/import":         "ai_capability_center:manage",
		"POST /api/ai-capability-center/apis/connectivity-check":          "ai_capability_center:manage",
		"POST /api/ai-capability-center/:resource":                        "ai_capability_center:manage",
		"PUT /api/ai-capability-center/:resource/:id":                     "ai_capability_center:manage",
		"DELETE /api/ai-capability-center/:resource/:id":                  "ai_capability_center:manage",
		"POST /api/ai-gateway/v1/invoke":                                  "ai_gateway:invoke",
		"GET /api/ai-gateway/v1/video-tasks/:task_id":                     "ai_gateway:invoke",
		"POST /api/integration-center/connectivity-check":                 "integration_center:connection_manage",
		"POST /api/integration-center/gateway/invoke":                     "integration_center:connection_manage",
		"POST /api/integration-center/platforms":                          "integration_center:platform_manage",
		"PUT /api/integration-center/platforms/:code":                     "integration_center:platform_manage",
		"POST /api/integration-center/platform-capabilities":              "integration_center:platform_manage",
		"PUT /api/integration-center/platform-capabilities/:id":           "integration_center:platform_manage",
		"PATCH /api/integration-center/platform-capabilities/:id/disable": "integration_center:platform_manage",
		"POST /api/integration-center/provider-apps":                      "integration_center:app_manage",
		"PUT /api/integration-center/provider-apps/:code":                 "integration_center:app_manage",
		"PATCH /api/integration-center/provider-apps/:code/credential":    "integration_center:app_manage",
		"PATCH /api/integration-center/app-capabilities/:id":              "integration_center:app_manage",
		"POST /api/integration-center/tenant-connections":                 "integration_center:connection_manage",
		"POST /api/integration-center/oauth/start":                        "integration_center:connection_manage",
		"POST /api/integration-center/my-connections":                     "/integration-center/my-connections",
		"POST /api/integration-center/my-oauth/start":                     "/integration-center/my-connections",
		"POST /api/integration-center/tenant-connections/:id/refresh":     "integration_center:connection_manage",
		"POST /api/integration-center/tenant-connections/:id/pause":       "integration_center:connection_manage",
		"POST /api/integration-center/tenant-connections/:id/resume":      "integration_center:connection_manage",
		"POST /api/integration-center/tenant-connections/:id/retry":       "integration_center:connection_manage",
		"POST /api/integration-center/sync-jobs/:id/retry":                "integration_center:connection_manage",
		"POST /api/integration-center/sync-jobs/:id/pause":                "integration_center:connection_manage",
		"POST /api/integration-center/sync-jobs/:id/resume":               "integration_center:connection_manage",
		"POST /api/integration-center/quota-policies":                     "integration_center:quota_manage",
		"PUT /api/integration-center/quota-policies/:code":                "integration_center:quota_manage",
		"PATCH /api/integration-center/quota-policies/:code/status":       "integration_center:quota_manage",
		"POST /api/integration-center/alerts/:id/process":                 "integration_center:connection_manage",
		"POST /api/integration-center/alerts/:id/resolve":                 "integration_center:connection_manage",
		"POST /api/integration-center/alerts/:id/ignore":                  "integration_center:connection_manage",
		"POST /api/integration-center/logs/export":                        "integration_center:connection_manage",
		"POST /api/data-center/raw/batches":                               "data_center:batch_retry",
		"POST /api/data-center/raw/batches/:id/reprocess":                 "data_center:batch_retry",
		"POST /api/data-center/metrics":                                   "data_center:metric_manage",
		"PUT /api/data-center/metrics/:id":                                "data_center:metric_manage",
		"POST /api/data-center/metrics/:id/enable":                        "data_center:metric_manage",
		"POST /api/data-center/metrics/:id/disable":                       "data_center:metric_manage",
		"POST /api/data-center/anomaly-rules":                             "data_center:rule_manage",
		"PUT /api/data-center/anomaly-rules/:id":                          "data_center:rule_manage",
		"POST /api/data-center/anomaly-rules/:id/enable":                  "data_center:rule_manage",
		"POST /api/data-center/anomaly-rules/:id/disable":                 "data_center:rule_manage",
		"POST /api/data-center/anomaly-rules/:id/test":                    "data_center:rule_manage",
		"POST /api/data-center/anomalies/scan":                            "data_center:scan",
		"POST /api/data-center/anomalies/:id/analyze":                     "data_center:ai_analyze",
		"POST /api/data-center/anomalies/:id/reanalyze":                   "data_center:ai_analyze",
		"POST /api/data-center/anomalies/:id/generate-task":               "data_center:task_generate",
		"POST /api/data-center/anomalies/:id/confirm":                     "data_center:task_generate",
		"POST /api/data-center/anomalies/:id/ignore":                      "data_center:task_generate",
		"POST /api/data-center/anomalies/:id/close":                       "data_center:task_generate",
		"POST /api/data-center/tasks":                                     "data_center:task_flow",
		"PUT /api/data-center/tasks/:id":                                  "data_center:task_flow",
		"POST /api/data-center/tasks/:id/start":                           "data_center:task_flow",
		"POST /api/data-center/tasks/:id/feedback":                        "data_center:task_flow",
		"POST /api/data-center/tasks/:id/complete":                        "data_center:task_flow",
		"POST /api/data-center/tasks/:id/close":                           "data_center:task_flow",
		"POST /api/data-center/reviews/generate":                          "data_center:review_confirm",
		"POST /api/data-center/reviews/:id/confirm":                       "data_center:review_confirm",
		"PUT /api/data-center/reviews/:id":                                "data_center:review_confirm",
		"POST /api/ai-geo/materials/brands":                               "ai_geo:data:import",
		"PUT /api/ai-geo/materials/brands/:id":                            "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/brands/:id":                         "ai_geo:data:import",
		"POST /api/ai-geo/materials/products":                             "ai_geo:data:import",
		"PUT /api/ai-geo/materials/products/:id":                          "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/products/:id":                       "ai_geo:data:import",
		"POST /api/ai-geo/materials/skus":                                 "ai_geo:data:import",
		"PUT /api/ai-geo/materials/skus/:id":                              "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/skus/:id":                           "ai_geo:data:import",
		"POST /api/ai-geo/materials/competitors":                          "ai_geo:data:import",
		"PUT /api/ai-geo/materials/competitors/:id":                       "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/competitors/:id":                    "ai_geo:data:import",
		"POST /api/ai-geo/materials/keywords":                             "ai_geo:data:import",
		"PUT /api/ai-geo/materials/keywords/:id":                          "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/keywords/:id":                       "ai_geo:data:import",
		"POST /api/ai-geo/materials/assets":                               "ai_geo:data:import",
		"PUT /api/ai-geo/materials/assets/:id":                            "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/assets/:id":                         "ai_geo:data:import",
		"POST /api/ai-geo/materials/hotspots":                             "ai_geo:data:import",
		"PUT /api/ai-geo/materials/hotspots/:id":                          "ai_geo:data:import",
		"DELETE /api/ai-geo/materials/hotspots/:id":                       "ai_geo:data:import",
		"POST /api/ai-geo/materials/imports":                              "ai_geo:data:import",
		"POST /api/ai-geo/workbench/drafts/generate":                      "ai_geo:workbench:generate",
		"POST /api/ai-geo/drafts":                                         "ai_geo:draft:manage",
		"DELETE /api/ai-geo/drafts/:id":                                   "ai_geo:draft:manage",
		"POST /api/ai-geo/drafts/:id/submit":                              "ai_geo:draft:manage",
		"POST /api/ai-geo/drafts/:id/approve":                             "ai_geo:draft:manage",
		"POST /api/ai-geo/drafts/:id/reject":                              "ai_geo:draft:manage",
		"POST /api/ai-geo/drafts/:id/audit-suggestions":                   "ai_geo:draft:manage",
		"POST /api/ai-geo/drafts/:id/channel-contents":                    "ai_geo:channel_content:manage",
		"PUT /api/ai-geo/channel-contents/:id":                            "ai_geo:channel_content:manage",
		"POST /api/ai-geo/channel-contents/:id/approve":                   "ai_geo:channel_content:manage",
		"POST /api/ai-geo/channel-contents/:id/reject":                    "ai_geo:channel_content:manage",
		"POST /api/ai-geo/channel-contents/:id/audit-suggestions":         "ai_geo:channel_content:manage",
		"POST /api/ai-geo/publish-plans":                                  "ai_geo:publish_plan:manage",
		"PUT /api/ai-geo/publish-plans/:id":                               "ai_geo:publish_plan:manage",
		"PATCH /api/ai-geo/publish-plans/:id/status":                      "ai_geo:publish_plan:manage",
		"POST /api/ai-geo/channels":                                       "ai_geo:channel:manage",
		"POST /api/ai-geo/channel-accounts":                               "ai_geo:channel_account:manage",
		"PUT /api/tenant/branding":                                        "brand:edit",
		"POST /api/tenants":                                               "tenant:create",
		"POST /api/tenants/with-package":                                  "tenant:create",
		"PUT /api/tenants/:id":                                            "tenant:edit",
		"PATCH /api/tenants/:id/status":                                   "tenant:status",
		"PUT /api/tenants/:id/package-config":                             "tenant:edit",
		"DELETE /api/tenants/:id":                                         "tenant:delete",
		"PUT /api/tenants/:id/primary-admin/password":                     "tenant:reset_primary_password",
		"PUT /api/tenants/:id/subscription":                               "plan:config",
		"PUT /api/tenants/:id/feature-overrides":                          "plan:config",
		"PUT /api/tenants/:id/quota-overrides":                            "tenant:quota_config",
		"POST /api/roles":                                                 "role:create",
		"PUT /api/roles/:id":                                              "role:edit",
		"PUT /api/roles/:id/permissions":                                  "role:permission",
		"DELETE /api/roles/:id":                                           "role:delete",
		"PUT /api/permissions/menu-data-perm-mode/:id":                    "menu:edit",
		"PUT /api/permissions/menu-data-perm-mode":                        "menu:edit",
		"PUT /api/permissions/menu-package-feature/:id":                   "menu:package_feature",
		"PUT /api/permissions/menu-overrides":                             "menu:edit",
		"POST /api/permissions":                                           "menu:edit",
		"PUT /api/permissions/:id":                                        "menu:edit",
		"DELETE /api/permissions/:id":                                     "menu:edit",
		"POST /api/plans":                                                 "plan:create",
		"PUT /api/plans/:id":                                              "plan:edit",
		"POST /api/plans/:id/copy":                                        "plan:create",
		"DELETE /api/plans/:id":                                           "plan:delete",
		"POST /api/plans/features":                                        "plan:create",
		"PUT /api/plans/features/:id":                                     "plan:edit",
		"PUT /api/plans/:id/features":                                     "plan:config",
		"PUT /api/plans/:id/capabilities":                                 "plan:config",
		"POST /api/plans/quotas":                                          "plan:create",
		"PUT /api/plans/quotas/:id":                                       "plan:edit",
		"PUT /api/plans/:id/quotas":                                       "plan:config",
		"POST /api/org-nodes":                                             "org:create",
		"PUT /api/org-nodes/:id":                                          "org:edit",
		"DELETE /api/org-nodes/:id":                                       "org:delete",
		"POST /api/companies":                                             "org:create",
		"PUT /api/companies/:id":                                          "org:edit",
		"DELETE /api/companies/:id":                                       "org:delete",
		"POST /api/departments":                                           "org:create",
		"PUT /api/departments/:id":                                        "org:edit",
		"DELETE /api/departments/:id":                                     "org:delete",
		"POST /api/stores":                                                "org:create",
		"PUT /api/stores/:id":                                             "org:edit",
		"DELETE /api/stores/:id":                                          "org:delete",
		"POST /api/position-types":                                        "pos:create",
		"PUT /api/position-types/:id":                                     "pos:edit",
		"DELETE /api/position-types/:id":                                  "pos:delete",
		"POST /api/positions":                                             "pos:create",
		"PUT /api/positions/:id":                                          "pos:edit",
		"DELETE /api/positions/:id":                                       "pos:delete",
		"POST /api/business-units":                                        "business_unit:create",
		"PUT /api/business-units/:id":                                     "business_unit:edit",
		"DELETE /api/business-units/:id":                                  "business_unit:delete",
		"POST /api/business-units/:id/resources":                          "business_unit:resource_bind",
		"DELETE /api/business-units/:id/resources/:relationId":            "business_unit:resource_bind",
		"POST /api/business-units/:id/org-mappings":                       "business_unit:edit",
		"DELETE /api/business-units/org-mappings/:id":                     "business_unit:edit",
		"POST /api/business-resources":                                    "business_resource:create",
		"PUT /api/business-resources/:id":                                 "business_resource:edit",
		"DELETE /api/business-resources/:id":                              "business_resource:delete",
		"POST /api/business-resources/:id/relations":                      "business_resource:relation_manage",
		"DELETE /api/business-resources/:id/relations/:relationId":        "business_resource:relation_manage",
		"POST /api/base/business-resources":                               "business_resource:create",
		"PUT /api/base/business-resources/:id":                            "business_resource:edit",
		"DELETE /api/base/business-resources/:id":                         "business_resource:delete",
		"PUT /api/base/business-resources/:id/actors":                     "business_resource:actor_manage",
		"PUT /api/base/business-resources/:id/relations":                  "business_resource:relation_manage",
		"PUT /api/base/business-resources/field-configs":                  "business_resource:field_config_manage",
		"GET /api/base/business-resources/import-template":                "business_resource:import",
		"POST /api/base/business-units":                                   "business_unit:create",
		"PUT /api/base/business-units/:id":                                "business_unit:edit",
		"DELETE /api/base/business-units/:id":                             "business_unit:delete",
		"PUT /api/base/business-units/:id/actors":                         "business_unit:actor_manage",
		"PUT /api/base/business-units/:id/relations":                      "business_unit:relation_manage",
		"POST /api/base/business-unit-attr-templates":                     "business_unit:attr_template_manage",
		"PUT /api/base/business-unit-attr-templates/:id":                  "business_unit:attr_template_manage",
		"DELETE /api/base/business-unit-attr-templates/:id":               "business_unit:attr_template_manage",
		"POST /api/dict-types":                                            "dict_type:create",
		"PUT /api/dict-types/:id":                                         "dict_type:edit",
		"DELETE /api/dict-types/:id":                                      "dict_type:delete",
		"POST /api/dict-items":                                            "dict_item:create",
		"PUT /api/dict-items/:id":                                         "dict_item:edit",
		"DELETE /api/dict-items/:id":                                      "dict_item:delete",
		"DELETE /api/dict-items/:id/override":                             "dict_item:edit",
		"POST /api/sys-params":                                            "param:create",
		"PUT /api/sys-params/:id":                                         "param:edit",
		"DELETE /api/sys-params/:id":                                      "param:delete",
		"DELETE /api/sys-params/:id/override":                             "param:edit",
		"POST /api/params":                                                "param:create",
		"POST /api/users":                                                 "user:create",
		"PUT /api/users/:id":                                              "user:edit",
		"PUT /api/users/:id/password":                                     "user:reset_password",
		"DELETE /api/users/:id":                                           "user:delete",
		"POST /api/batch/users/import":                                    "user:create",
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
		"/api/apps":                          "/apps",
		"/api/apps/stats":                    "/apps",
		"/api/apps/manifest/template":        "/apps",
		"/api/apps/:id":                      "/apps",
		"/api/ai-capability-center/overview": "/ai-capability-center",
		"/api/ai-capability-center/apis/connectivity-check":         "/ai-capability-center/providers",
		"/api/ai-capability-center/:resource":                       "/ai-capability-center",
		"/api/integration-center/overview":                          "/integration-center",
		"/api/integration-center/connectivity-check":                "/integration-center",
		"/api/integration-center/gateway/invoke":                    "/integration-center/tenant-connections",
		"/api/integration-center/platforms":                         "/integration-center/platforms",
		"/api/integration-center/platforms/:code":                   "/integration-center/platforms",
		"/api/integration-center/platform-capabilities":             "/integration-center/workspace",
		"/api/integration-center/workspace":                         "/integration-center/workspace",
		"/api/integration-center/app-capabilities":                  "/integration-center/workspace",
		"/api/integration-center/provider-apps":                     "/integration-center/workspace",
		"/api/integration-center/provider-apps/:code":               "/integration-center/workspace",
		"/api/integration-center/app-capabilities/:id":              "/integration-center/workspace",
		"/api/integration-center/tenant-connections":                "/integration-center/tenant-connections",
		"/api/integration-center/tenant-connections/:id":            "/integration-center/tenant-connections",
		"/api/integration-center/my-connections":                    "/integration-center/my-connections",
		"/api/integration-center/my-connections/:id":                "/integration-center/my-connections",
		"/api/integration-center/my-oauth/start":                    "/integration-center/my-connections",
		"/api/integration-center/my-sync-jobs":                      "/integration-center/my-connections",
		"/api/integration-center/my-sync-jobs/:id":                  "/integration-center/my-connections",
		"/api/integration-center/oauth/start":                       "/integration-center/tenant-connections",
		"/api/integration-center/oauth/callback/:provider_app_code": "/integration-center/tenant-connections",
		"/api/integration-center/tenant-connections/:id/refresh":    "/integration-center/tenant-connections",
		"/api/integration-center/tenant-connections/:id/pause":      "/integration-center/tenant-connections",
		"/api/integration-center/tenant-connections/:id/resume":     "/integration-center/tenant-connections",
		"/api/integration-center/tenant-connections/:id/retry":      "/integration-center/tenant-connections",
		"/api/integration-center/sync-monitor":                      "/integration-center/sync-monitor",
		"/api/integration-center/sync-jobs/:id":                     "/integration-center/sync-monitor",
		"/api/integration-center/sync-jobs/:id/retry":               "/integration-center/sync-monitor",
		"/api/integration-center/sync-jobs/:id/pause":               "/integration-center/sync-monitor",
		"/api/integration-center/sync-jobs/:id/resume":              "/integration-center/sync-monitor",
		"/api/integration-center/quota":                             "/integration-center/quota",
		"/api/integration-center/quota-usages":                      "/integration-center/quota",
		"/api/integration-center/quota-policies":                    "/integration-center/quota",
		"/api/integration-center/quota-policies/:code":              "/integration-center/quota",
		"/api/integration-center/quota-policies/:code/status":       "/integration-center/quota",
		"/api/integration-center/alerts":                            "/integration-center/alerts",
		"/api/integration-center/alerts/:id/process":                "/integration-center/alerts",
		"/api/integration-center/alerts/:id/resolve":                "/integration-center/alerts",
		"/api/integration-center/alerts/:id/ignore":                 "/integration-center/alerts",
		"/api/integration-center/logs":                              "/integration-center/logs",
		"/api/integration-center/logs/:id":                          "/integration-center/logs",
		"/api/integration-center/logs/export":                       "/integration-center/logs",
		"/api/data-center/dashboard/summary":                        "/data-center/dashboard",
		"/api/data-center/dashboard/trends":                         "/data-center/dashboard",
		"/api/data-center/dashboard/rankings":                       "/data-center/dashboard",
		"/api/data-center/dashboard/anomalies":                      "/data-center/dashboard",
		"/api/data-center/dashboard/tasks":                          "/data-center/dashboard",
		"/api/data-center/overview/pipeline":                        "/data-center/overview",
		"/api/data-center/overview/jobs":                            "/data-center/overview",
		"/api/data-center/overview/errors":                          "/data-center/overview",
		"/api/data-center/raw/batches":                              "/data-center/raw",
		"/api/data-center/raw/batches/:id":                          "/data-center/raw",
		"/api/data-center/raw/batches/:id/errors":                   "/data-center/raw",
		"/api/data-center/standard/:data_type":                      "/data-center/standard",
		"/api/data-center/standard/:data_type/:id":                  "/data-center/standard",
		"/api/data-center/metrics":                                  "/data-center/metrics",
		"/api/data-center/metrics/:id":                              "/data-center/metrics",
		"/api/data-center/metrics/results":                          "/data-center/metrics",
		"/api/data-center/anomaly-rules":                            "/data-center/rules",
		"/api/data-center/anomaly-rules/:id":                        "/data-center/rules",
		"/api/data-center/anomalies":                                "/data-center/anomalies",
		"/api/data-center/anomalies/:id":                            "/data-center/anomalies",
		"/api/data-center/tasks":                                    "/data-center/tasks",
		"/api/data-center/tasks/:id":                                "/data-center/tasks",
		"/api/data-center/reviews":                                  "/data-center/reviews",
		"/api/data-center/reviews/:id":                              "/data-center/reviews",
		"/api/ai-geo/overview":                                      "/ai-geo/dashboard",
		"/api/ai-geo/materials/brands":                              "/ai-geo/data/brands",
		"/api/ai-geo/materials/brands/:id":                          "/ai-geo/data/brands",
		"/api/ai-geo/materials/products":                            "/ai-geo/data/products",
		"/api/ai-geo/materials/products/:id":                        "/ai-geo/data/products",
		"/api/ai-geo/materials/skus":                                "/ai-geo/data/products",
		"/api/ai-geo/materials/skus/:id":                            "/ai-geo/data/products",
		"/api/ai-geo/materials/competitors":                         "/ai-geo/data/products",
		"/api/ai-geo/materials/competitors/:id":                     "/ai-geo/data/products",
		"/api/ai-geo/materials/keywords":                            "/ai-geo/data/products",
		"/api/ai-geo/materials/keywords/:id":                        "/ai-geo/data/products",
		"/api/ai-geo/materials/assets":                              "/ai-geo/data/products",
		"/api/ai-geo/materials/assets/:id":                          "/ai-geo/data/products",
		"/api/ai-geo/materials/hotspots":                            "/ai-geo/data/products",
		"/api/ai-geo/materials/hotspots/:id":                        "/ai-geo/data/products",
		"/api/ai-geo/materials/imports/:id/errors":                  "/ai-geo/data/products",
		"/api/ai-geo/drafts":                                        "/ai-geo/drafts",
		"/api/ai-geo/drafts/:id":                                    "/ai-geo/drafts",
		"/api/ai-geo/drafts/:id/audit-suggestions":                  "/ai-geo/drafts",
		"/api/ai-geo/channel-contents":                              "/ai-geo/drafts",
		"/api/ai-geo/channel-contents/:id":                          "/ai-geo/drafts",
		"/api/ai-geo/channel-contents/:id/approve":                  "/ai-geo/drafts",
		"/api/ai-geo/channel-contents/:id/reject":                   "/ai-geo/drafts",
		"/api/ai-geo/channel-contents/:id/audit-suggestions":        "/ai-geo/drafts",
		"/api/ai-geo/publish-plans":                                 "/ai-geo/plans/queue",
		"/api/ai-geo/channels":                                      "/ai-geo/channels/profiles",
		"/api/ai-geo/channel-accounts":                              "/ai-geo/channels/accounts",
		"/api/plans":                                                "/plans",
		"/api/plans/matrix":                                         "/plans",
		"/api/plans/features":                                       "/plans",
		"/api/plans/:id/features":                                   "/plans",
		"/api/plans/quotas":                                         "/plans",
		"/api/plans/:id/quotas":                                     "/plans",
		"/api/organizations/tree":                                   "/organization",
		"/api/position-types":                                       "/positions",
		"/api/positions":                                            "/positions",
		"/api/business-units":                                       "/business-units",
		"/api/business-units/tree":                                  "/business-units",
		"/api/business-units/:id/resources":                         "/business-units",
		"/api/business-units/:id/org-mappings":                      "/business-units",
		"/api/business-units/org-mappings":                          "/business-units",
		"/api/business-resources":                                   "/business-resources",
		"/api/business-resources/:id/relations":                     "/business-resources",
		"/api/base/business-resources/dictionary/tree":              "/business-units",
		"/api/base/business-resources/summary":                      "/business-units",
		"/api/base/business-resources":                              "/business-units",
		"/api/base/business-resources/:id/actors":                   "/business-units",
		"/api/base/business-resources/:id/relations":                "/business-units",
		"/api/base/business-resources/field-configs":                "/business-units",
		"/api/base/business-resources/import-template":              "/business-units",
		"/api/base/business-units/dictionary/tree":                  "/business-units",
		"/api/base/business-units/summary":                          "/business-units",
		"/api/base/business-units":                                  "/business-units",
		"/api/base/business-units/:id/actors":                       "/business-units",
		"/api/base/business-units/:id/relations":                    "/business-units",
		"/api/base/business-unit-attr-templates":                    "/business-units",
		"/api/base/business-unit-attr-templates/:id":                "/business-units",
		"/api/base/business-unit-attr-templates/match":              "/business-units",
		"/api/users":                         "/users",
		"/api/users/assignable-roles":        "/users",
		"/api/roles":                         "/roles",
		"/api/roles/:id":                     "/roles",
		"/api/roles/permission-menu-bundles": "/roles",
		"/api/permission-menu-bundles":       "/roles",
		"/api/permissions/menu-bundles":      "/roles",
		"/api/permissions/menu-overrides":    "/menus",
		"/api/permissions":                   "/menus",
		"/api/permissions/tree":              "/menus",
		"/api/permissions/:id":               "/menus",
		"/api/organizations/detail":          "/organization",
		"/api/dict-types":                    "/dict",
		"/api/dict-items":                    "/dict",
		"/api/sys-params":                    "/params",
		"/api/sys-params/batch":              "/params",
		"/api/params":                        "/params",
		"/api/params/:key":                   "/params",
		"/api/logs/audit":                    "/audit-logs",
		"/api/logs/login":                    "/login-logs",
		"/api/monitor/health-detail":         "/monitor/health",
		"/api/monitor/server-info":           "/monitor/server",
		"/api/monitor/scheduled-jobs":        "/monitor/jobs",
		"/api/monitor/services-overview":     "/monitor/services",
		"/api/monitor/cache-stats":           "/monitor/cache",
		"/api/monitor/cache-keys":            "/monitor/cache-keys",
		"/api/batch/users/export":            "/users",
		"/api/batch/companies/export":        "/organization",
		"/api/batch/departments/export":      "/organization",
	}
}
