package models

import "time"

type Tenant struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	Code       string     `gorm:"column:code;type:varchar(64);uniqueIndex;not null"`
	Name       string     `gorm:"column:name;type:varchar(128);not null"`
	Status     int        `gorm:"column:status;not null;default:1"`
	IsPlatform bool       `gorm:"column:is_platform;not null;default:false"`
	BrandName  *string    `gorm:"column:brand_name;type:varchar(128)"`
	FooterText *string    `gorm:"column:footer_text;type:text"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (Tenant) TableName() string {
	return "tenant"
}

type AppUser struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index"`
	EmployeeNo      string     `gorm:"column:employee_no;type:varchar(64);not null"`
	Account         string     `gorm:"column:account;type:varchar(64);not null"`
	PasswordHash    string     `gorm:"column:password_hash;type:varchar(255);not null"`
	Name            string     `gorm:"column:name;type:varchar(128);not null"`
	Phone           *string    `gorm:"column:phone;type:varchar(32)"`
	Email           *string    `gorm:"column:email;type:varchar(128)"`
	AvatarURL       *string    `gorm:"column:avatar_url;type:text"`
	Status          int        `gorm:"column:status;not null;default:1"`
	IsPlatformAdmin bool       `gorm:"column:is_platform_admin;not null;default:false"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (AppUser) TableName() string {
	return "app_user"
}

type Role struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  uint64     `gorm:"column:tenant_id;not null;index"`
	Code      string     `gorm:"column:code;type:varchar(64);not null"`
	Name      string     `gorm:"column:name;type:varchar(128);not null"`
	Status    int        `gorm:"column:status;not null;default:1"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (Role) TableName() string {
	return "role"
}

type Permission struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index"`
	ParentID         *uint64    `gorm:"column:parent_id"`
	Name             string     `gorm:"column:name;type:varchar(128);not null"`
	Path             string     `gorm:"column:path;type:varchar(255);not null;index"`
	PermType         int        `gorm:"column:perm_type;not null"`
	SortOrder        int        `gorm:"column:sort_order;not null;default:0"`
	Enabled          bool       `gorm:"column:enabled;not null;default:true"`
	Visible          bool       `gorm:"column:visible;not null;default:true"`
	IsPlatformOnly   bool       `gorm:"column:is_platform_only;not null;default:false"`
	IsPackageFeature bool       `gorm:"column:is_package_feature;not null;default:true"`
	TenantEditable   bool       `gorm:"column:tenant_editable;not null;default:false"`
	TenantEditScope  *string    `gorm:"column:tenant_edit_scope;type:varchar(64)"`
	FeatureCode      *string    `gorm:"column:feature_code;type:varchar(128)"`
	FeatureType      *string    `gorm:"column:feature_type;type:varchar(32)"`
	DataPermMode     string     `gorm:"column:data_perm_mode;type:varchar(16);not null;default:'ORG'"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (Permission) TableName() string {
	return "permission"
}

type UserRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    uint64    `gorm:"column:user_id;not null;index"`
	RoleID    uint64    `gorm:"column:role_id;not null;index"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (UserRole) TableName() string {
	return "user_role"
}

type RolePermission struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	RoleID       uint64    `gorm:"column:role_id;not null;index"`
	PermissionID uint64    `gorm:"column:permission_id;not null;index"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}
