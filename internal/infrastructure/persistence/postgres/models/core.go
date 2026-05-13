package models

import "time"

type Tenant struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	Code         string     `gorm:"column:code;type:varchar(64);uniqueIndex;not null"`
	Name         string     `gorm:"column:name;type:varchar(200);not null"`
	Status       int        `gorm:"column:status;not null;default:1"`
	StartDate    *time.Time `gorm:"column:start_date;type:date"`
	ExpireDate   *time.Time `gorm:"column:expire_date;type:date"`
	MaxCompanies int        `gorm:"column:max_companies;not null;default:0"`
	MaxUsers     int        `gorm:"column:max_users;not null;default:0"`
	ContactName  *string    `gorm:"column:contact_name;type:varchar(100)"`
	ContactPhone *string    `gorm:"column:contact_phone;type:varchar(32)"`
	IsPlatform   bool       `gorm:"column:is_platform_tenant;not null;default:false"`
	BrandName    *string    `gorm:"column:brand_display_name;type:varchar(128)"`
	LogoData     *string    `gorm:"column:brand_logo_data;type:text"`
	FooterText   *string    `gorm:"column:brand_footer_text;type:varchar(256)"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (Tenant) TableName() string {
	return "tenant"
}

type AppUser struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID          uint64     `gorm:"column:tenant_id;not null;index;index:idx_app_user_tenant_employee,unique;index:idx_app_user_tenant_phone,unique"`
	CompanyID         *uint64    `gorm:"column:company_id;index"`
	DepartmentID      *uint64    `gorm:"column:department_id;index"`
	EmployeeNo        string     `gorm:"column:employee_no;type:varchar(64);not null;index:idx_app_user_tenant_employee,unique"`
	Account           string     `gorm:"column:account;type:varchar(64);index"`
	PasswordHash      string     `gorm:"column:password_hash;type:varchar(200);not null"`
	Name              string     `gorm:"column:name;type:varchar(100);not null"`
	Phone             *string    `gorm:"column:phone;type:varchar(32);index:idx_app_user_tenant_phone,unique"`
	Email             *string    `gorm:"column:email;type:varchar(200)"`
	AvatarURL         *string    `gorm:"column:avatar_url;type:text"`
	Status            int        `gorm:"column:status;not null;default:1"`
	IsPlatformAdmin   bool       `gorm:"column:is_platform_admin;not null;default:false"`
	SessionVersion    int        `gorm:"column:session_version;not null;default:1"`
	PasswordChangedAt *time.Time `gorm:"column:password_changed_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
}

func (AppUser) TableName() string {
	return "app_user"
}

type FileObject struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID     uint64     `gorm:"column:tenant_id;not null;index;index:idx_file_object_tenant_file_id,unique"`
	FileID       string     `gorm:"column:file_id;type:varchar(64);not null;index:idx_file_object_tenant_file_id,unique"`
	CreatedBy    uint64     `gorm:"column:created_by;not null;index"`
	OriginalName string     `gorm:"column:original_name;type:varchar(255);not null"`
	StoredName   string     `gorm:"column:stored_name;type:varchar(255);not null"`
	StoragePath  string     `gorm:"column:storage_path;type:text;not null"`
	MimeType     string     `gorm:"column:mime_type;type:varchar(128);not null"`
	FileSize     int64      `gorm:"column:file_size;not null"`
	Status       int        `gorm:"column:status;not null;default:1;index"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index"`
}

func (FileObject) TableName() string {
	return "file_object"
}

type Role struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    uint64     `gorm:"column:tenant_id;not null;index;index:idx_role_tenant_code,unique"`
	Code        string     `gorm:"column:code;type:varchar(64);not null;index:idx_role_tenant_code,unique"`
	Name        string     `gorm:"column:name;type:varchar(200);not null"`
	Description *string    `gorm:"column:description;type:varchar(500)"`
	Status      int        `gorm:"column:status;not null;default:1"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (Role) TableName() string {
	return "role"
}

type Permission struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index"`
	ParentID         *uint64    `gorm:"column:parent_id"`
	Name             string     `gorm:"column:name;type:varchar(200);not null"`
	Path             string     `gorm:"column:path;type:varchar(500);index"`
	PermType         int        `gorm:"column:perm_type;not null"`
	DataScope        *string    `gorm:"column:data_scope;type:varchar(32)"`
	SortOrder        int        `gorm:"column:sort_order;not null;default:0"`
	Enabled          bool       `gorm:"column:enabled;not null;default:true"`
	Visible          bool       `gorm:"column:visible;not null;default:true"`
	ShowInAdmin      bool       `gorm:"column:show_in_admin;not null;default:true"`
	IsPlatformOnly   bool       `gorm:"column:is_platform_only;not null;default:false"`
	IsPackageFeature bool       `gorm:"column:is_package_feature;not null;default:true"`
	TenantEditable   bool       `gorm:"column:tenant_editable;not null;default:false"`
	TenantEditScope  *string    `gorm:"column:tenant_edit_scope;type:varchar(100)"`
	AppCode          string     `gorm:"column:app_code;type:varchar(100);not null;default:'system-management';index"`
	FeatureCode      *string    `gorm:"column:feature_code;type:varchar(100);index"`
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
	UserID    uint64    `gorm:"column:user_id;not null;index;index:idx_user_role_user_role,unique"`
	RoleID    uint64    `gorm:"column:role_id;not null;index;index:idx_user_role_user_role,unique"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (UserRole) TableName() string {
	return "user_role"
}

type RolePermission struct {
	ID                        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	RoleID                    uint64    `gorm:"column:role_id;not null;index;index:idx_role_permission_role_permission,unique"`
	PermissionID              uint64    `gorm:"column:permission_id;not null;index;index:idx_role_permission_role_permission,unique"`
	DataScopeOverride         *string   `gorm:"column:data_scope_override;type:varchar(32)"`
	CustomCompanyIDsJSON      *string   `gorm:"column:custom_company_ids_json;type:text"`
	CustomDepartmentIDsJSON   *string   `gorm:"column:custom_department_ids_json;type:text"`
	CustomUserIDsJSON         *string   `gorm:"column:custom_user_ids_json;type:text"`
	CustomBusinessUnitIDsJSON *string   `gorm:"column:custom_business_unit_ids_json;type:text"`
	BUDataAccessMode          *string   `gorm:"column:bu_data_access_mode;type:varchar(32)"`
	Source                    string    `gorm:"column:source;type:varchar(32);not null;default:'MANUAL'"`
	SourceRef                 *string   `gorm:"column:source_ref;type:varchar(128)"`
	CreatedAt                 time.Time `gorm:"column:created_at;not null"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}
