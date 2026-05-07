package models

import "time"

type OrgNode struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    uint64     `gorm:"column:tenant_id;not null;index"`
	NodeType    string     `gorm:"column:node_type;type:varchar(32);not null"`
	ParentID    *uint64    `gorm:"column:parent_id;index"`
	CompanyID   *uint64    `gorm:"column:company_id;index"`
	Name        string     `gorm:"column:name;type:varchar(200);not null"`
	Code        *string    `gorm:"column:code;type:varchar(64)"`
	CompanyType *string    `gorm:"column:company_type;type:varchar(32)"`
	Status      int        `gorm:"column:status;not null;default:1"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (OrgNode) TableName() string { return "org_node" }

type PositionType struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  uint64     `gorm:"column:tenant_id;not null;index:idx_position_type_tenant_code,unique"`
	Name      string     `gorm:"column:name;type:varchar(200);not null"`
	Code      string     `gorm:"column:code;type:varchar(64);not null;index:idx_position_type_tenant_code,unique"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (PositionType) TableName() string { return "position_type" }

type Position struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index:idx_position_tenant_code,unique"`
	PositionTypeID uint64     `gorm:"column:position_type_id;not null;index"`
	Name           string     `gorm:"column:name;type:varchar(200);not null"`
	Code           string     `gorm:"column:code;type:varchar(64);not null;index:idx_position_tenant_code,unique"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (Position) TableName() string { return "position" }

type BusinessUnit struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index:idx_business_unit_tenant_code,unique"`
	Name             string     `gorm:"column:name;type:varchar(200);not null"`
	Code             string     `gorm:"column:code;type:varchar(64);not null;index:idx_business_unit_tenant_code,unique"`
	BUType           *string    `gorm:"column:bu_type;type:varchar(32)"`
	Status           int        `gorm:"column:status;not null;default:1"`
	BillingEnabled   bool       `gorm:"column:billing_enabled;not null;default:false"`
	StatisticEnabled bool       `gorm:"column:statistic_enabled;not null;default:false"`
	Remark           *string    `gorm:"column:remark;type:text"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnit) TableName() string { return "business_unit" }

type BusinessUnitOrgMap struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index"`
	BusinessUnitID uint64     `gorm:"column:business_unit_id;not null;index"`
	OrgID          uint64     `gorm:"column:org_id;not null;index"`
	OrgType        string     `gorm:"column:org_type;type:varchar(32);not null"`
	ScopeType      string     `gorm:"column:scope_type;type:varchar(32);not null"`
	Priority       int        `gorm:"column:priority;not null;default:0"`
	EffectiveStart *time.Time `gorm:"column:effective_start"`
	EffectiveEnd   *time.Time `gorm:"column:effective_end"`
	Status         int        `gorm:"column:status;not null;default:1"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
}

func (BusinessUnitOrgMap) TableName() string { return "business_unit_org_map" }

type BusinessUnitScope struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	RolePermissionID uint64    `gorm:"column:role_permission_id;not null;index"`
	BusinessUnitID   uint64    `gorm:"column:business_unit_id;not null;index"`
	CreatedAt        time.Time `gorm:"column:created_at;not null"`
}

func (BusinessUnitScope) TableName() string { return "business_unit_scope" }

type DictType struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index:idx_dict_type_tenant_code,unique"`
	Code           string     `gorm:"column:code;type:varchar(64);not null;index:idx_dict_type_tenant_code,unique"`
	Name           string     `gorm:"column:name;type:varchar(200);not null"`
	Remark         *string    `gorm:"column:remark;type:varchar(500)"`
	Scope          string     `gorm:"column:scope;type:varchar(32);not null"`
	TenantEditable bool       `gorm:"column:tenant_editable;not null;default:false"`
	IsPlatformOnly bool       `gorm:"column:is_platform_only;not null;default:false"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (DictType) TableName() string { return "dict_type" }

type DictItem struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   uint64     `gorm:"column:tenant_id;not null;index"`
	DictTypeID uint64     `gorm:"column:dict_type_id;not null;index"`
	Label      string     `gorm:"column:label;type:varchar(200);not null"`
	Value      string     `gorm:"column:value;type:varchar(200);not null"`
	SortOrder  int        `gorm:"column:sort_order;not null;default:0"`
	Enabled    bool       `gorm:"column:enabled;not null;default:true"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (DictItem) TableName() string { return "dict_item" }

type TenantDictItemOverride struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    uint64    `gorm:"column:tenant_id;not null;index"`
	DictItemID  uint64    `gorm:"column:dict_item_id;not null;index"`
	CustomLabel *string   `gorm:"column:custom_label;type:varchar(200)"`
	CustomValue *string   `gorm:"column:custom_value;type:varchar(200)"`
	Enabled     *bool     `gorm:"column:enabled"`
	SortOrder   *int      `gorm:"column:sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

func (TenantDictItemOverride) TableName() string { return "tenant_dict_item_override" }

type TenantMenuOverride struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID     uint64    `gorm:"column:tenant_id;not null;index"`
	PermissionID uint64    `gorm:"column:permission_id;not null;index"`
	CustomName   *string   `gorm:"column:custom_name;type:varchar(200)"`
	CustomIcon   *string   `gorm:"column:custom_icon;type:varchar(100)"`
	Enabled      *bool     `gorm:"column:enabled"`
	Visible      *bool     `gorm:"column:visible"`
	SortOrder    *int      `gorm:"column:sort_order"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

func (TenantMenuOverride) TableName() string { return "tenant_menu_override" }

type TenantParamValue struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   uint64    `gorm:"column:tenant_id;not null;index"`
	ParamID    uint64    `gorm:"column:param_id;not null;index"`
	ParamValue *string   `gorm:"column:param_value;type:text"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null"`
}

func (TenantParamValue) TableName() string { return "tenant_param_value" }

type AppUserDepartment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID       uint64    `gorm:"column:user_id;not null;index"`
	DepartmentID uint64    `gorm:"column:department_id;not null;index"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (AppUserDepartment) TableName() string { return "app_user_department" }

type AppUserPosition struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     uint64    `gorm:"column:user_id;not null;index"`
	PositionID uint64    `gorm:"column:position_id;not null;index"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
}

func (AppUserPosition) TableName() string { return "app_user_position" }

type PermissionCustomDepartment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	PermissionID uint64    `gorm:"column:permission_id;not null;index"`
	DepartmentID uint64    `gorm:"column:department_id;not null;index"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (PermissionCustomDepartment) TableName() string { return "permission_custom_department" }

type PermissionCustomUser struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	PermissionID uint64    `gorm:"column:permission_id;not null;index"`
	UserID       uint64    `gorm:"column:user_id;not null;index"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (PermissionCustomUser) TableName() string { return "permission_custom_user" }

type UserPreference struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    uint64    `gorm:"column:user_id;not null;index"`
	PrefKey   string    `gorm:"column:pref_key;type:varchar(64);not null"`
	PrefValue *string   `gorm:"column:pref_value;type:text"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (UserPreference) TableName() string { return "user_preference" }
