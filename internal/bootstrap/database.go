package bootstrap

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
	return NewPostgresWithOptions(dsn, true)
}

func NewPostgresWithOptions(dsn string, autoMigrate bool) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if autoMigrate {
		if err := db.AutoMigrate(
			&models.Tenant{},
			&models.OrgNode{},
			&models.AppUser{},
			&models.Role{},
			&models.Permission{},
			&models.UserRole{},
			&models.RolePermission{},
			&models.PositionType{},
			&models.Position{},
			&models.BusinessUnit{},
			&models.BusinessUnitOrgMap{},
			&models.BusinessUnitScope{},
			&models.DictType{},
			&models.DictItem{},
			&models.TenantDictItemOverride{},
			&models.TenantMenuOverride{},
			&models.SystemParam{},
			&models.TenantParamValue{},
			&models.AppUserDepartment{},
			&models.AppUserPosition{},
			&models.PermissionCustomDepartment{},
			&models.PermissionCustomUser{},
			&models.UserPreference{},
			&models.SaasFeature{},
			&models.SaasPlan{},
			&models.SaasQuota{},
			&models.SaasPlanFeature{},
			&models.SaasPlanQuota{},
			&models.TenantSubscription{},
			&models.TenantFeatureOverride{},
			&models.TenantQuotaOverride{},
			&models.TenantQuotaUsage{},
			&models.LoginLog{},
			&models.AuditLog{},
		); err != nil {
			return nil, err
		}
	}

	if err := seedCoreData(db); err != nil {
		return nil, err
	}

	return db, nil
}
