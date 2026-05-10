package permissioncatalog

import "testing"

func TestPackageCatalogDistinguishesPlatformAndTenantFeatures(t *testing.T) {
	for _, path := range []string{"/home", "/tenants", "/plans", "/permissions", "/monitor/health", "menu:delete", "menu:package_feature", "tenant:create", "plan:edit", "dict_type:create", "perm:create", "home:view"} {
		if !ExcludedPackageFeaturePath(path) {
			t.Fatalf("%s should be excluded from package catalog", path)
		}
	}

	for _, path := range []string{"/organization", "/positions", "/business-units", "/users", "/roles", "/menus", "/dict", "/params", "/audit-logs", "/login-logs"} {
		if !IsTenantPackageMenuPath(path) {
			t.Fatalf("%s should be a tenant package menu", path)
		}
	}

	if IsPackageFeatureOperation("menu:delete") {
		t.Fatal("menu:delete is platform-only and must not enter package catalog")
	}
	if !IsPackageFeatureOperation("menu:edit") {
		t.Fatal("menu:edit should be controlled by package catalog")
	}
	if !IsPackageFeatureOperation("brand:edit") || OperationFeatureCode("brand:edit") != "brand_config" {
		t.Fatal("brand:edit should map to the brand_config package feature")
	}
}
