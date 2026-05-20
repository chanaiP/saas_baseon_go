package handlers

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestTenantScopeAllowsMatrix(t *testing.T) {
	cases := []struct {
		scope      string
		platform   bool
		enterprise bool
		personal   bool
	}{
		{TenantScopePlatformOnly, true, false, false},
		{TenantScopeEnterpriseOnly, false, true, false},
		{TenantScopePersonalOnly, false, false, true},
		{TenantScopeAll, true, true, true},
		{TenantScopePlatformEnterprise, true, true, false},
		{TenantScopeEnterprisePersonal, false, true, true},
		{TenantScopePlatformPersonal, true, false, true},
	}
	for _, tc := range cases {
		require.Equal(t, tc.platform, tenantScopeAllows(tc.scope, TenantTypePlatform), tc.scope+" platform")
		require.Equal(t, tc.enterprise, tenantScopeAllows(tc.scope, TenantTypeEnterprise), tc.scope+" enterprise")
		require.Equal(t, tc.personal, tenantScopeAllows(tc.scope, TenantTypePersonal), tc.scope+" personal")
	}
}

func TestPermissionTenantScopeUsesTenantType(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Tenant{}))
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台", TenantType: TenantTypePlatform, IsPlatform: true, Status: 1}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 2, Code: "enterprise", Name: "企业", TenantType: TenantTypeEnterprise, Status: 1}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 3, Code: "personal", Name: "个人", TenantType: TenantTypePersonal, Status: 1}).Error)

	handler := &IdentityHandler{db: db}
	platformUser := models.AppUser{ID: 1, TenantID: 1, IsPlatformAdmin: true}
	enterpriseUser := models.AppUser{ID: 2, TenantID: 2}
	personalUser := models.AppUser{ID: 3, TenantID: 3}

	permission := models.Permission{TenantScope: TenantScopeEnterprisePersonal}
	require.True(t, handler.permissionAllowedForTenantType(platformUser, permission))
	require.True(t, handler.permissionAllowedForTenantType(enterpriseUser, permission))
	require.True(t, handler.permissionAllowedForTenantType(personalUser, permission))

	permission.TenantScope = TenantScopePlatformOnly
	require.True(t, handler.permissionAllowedForTenantType(platformUser, permission))
	require.False(t, handler.permissionAllowedForTenantType(enterpriseUser, permission))
	require.False(t, handler.permissionAllowedForTenantType(personalUser, permission))
}

func TestPersonalTenantCannotPassEnterpriseOnlyRouteScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Tenant{}, &models.Permission{}, &models.AppUser{}, &models.Role{}, &models.UserRole{}, &models.RolePermission{}))
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台", TenantType: TenantTypePlatform, IsPlatform: true, Status: 1}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 2, Code: "personal", Name: "个人", TenantType: TenantTypePersonal, Status: 1}).Error)

	permission := models.Permission{TenantID: 1, Path: "/users", Name: "用户管理", PermType: 3, Enabled: true, TenantScope: TenantScopeEnterpriseOnly}
	require.NoError(t, db.Create(&permission).Error)
	role := models.Role{ID: 1, TenantID: 2, Code: "personal_owner", Name: "个人空间所有者", Status: 1}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&models.RolePermission{RoleID: role.ID, PermissionID: permission.ID}).Error)
	require.NoError(t, db.Create(&models.UserRole{UserID: 2, RoleID: role.ID}).Error)

	handler := &IdentityHandler{db: db}
	user := models.AppUser{ID: 2, TenantID: 2, Status: 1}
	require.False(t, handler.routeAllowedForRequest(user, "GET", "/api/users", ""))
}
