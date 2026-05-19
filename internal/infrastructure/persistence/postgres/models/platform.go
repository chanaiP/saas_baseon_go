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
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID          uint64     `gorm:"column:tenant_id;not null;index:idx_business_unit_tenant_code,unique"`
	Name              string     `gorm:"column:name;type:varchar(200);not null"`
	Code              string     `gorm:"column:code;type:varchar(64);not null;index:idx_business_unit_tenant_code,unique"`
	UnitTypeCode      *string    `gorm:"column:unit_type_code;type:varchar(64)"`
	UnitTypeName      *string    `gorm:"column:unit_type_name;type:varchar(128)"`
	UnitGroupCode     *string    `gorm:"column:unit_group_code;type:varchar(64)"`
	UnitGroupName     *string    `gorm:"column:unit_group_name;type:varchar(128)"`
	BUType            *string    `gorm:"column:bu_type;type:varchar(32)"`
	UnitScenario      *string    `gorm:"column:unit_scenario;type:varchar(64)"`
	UnitForm          *string    `gorm:"column:unit_form;type:varchar(64)"`
	ParentID          *uint64    `gorm:"column:parent_id;index"`
	OwnerUserID       *uint64    `gorm:"column:owner_user_id;index"`
	OwnerOrgID        *uint64    `gorm:"column:owner_org_id;index"`
	AttrTemplateID    *uint64    `gorm:"column:attr_template_id;index"`
	Attrs             *string    `gorm:"column:attrs;type:jsonb"`
	Status            int        `gorm:"column:status;not null;default:1"`
	BillingEnabled    bool       `gorm:"column:billing_enabled;not null;default:false"`
	StatisticEnabled  bool       `gorm:"column:statistic_enabled;not null;default:true"`
	OperationEnabled  bool       `gorm:"column:operation_enabled;not null;default:false"`
	SettlementEnabled bool       `gorm:"column:settlement_enabled;not null;default:false"`
	DataScopeEnabled  bool       `gorm:"column:data_scope_enabled;not null;default:false"`
	Remark            *string    `gorm:"column:remark;type:text"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnit) TableName() string { return "business_unit" }

type BusinessUnitRelation struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index"`
	SourceUnitID     uint64     `gorm:"column:source_unit_id;not null;index"`
	TargetUnitID     uint64     `gorm:"column:target_unit_id;not null;index"`
	RelationTypeCode string     `gorm:"column:relation_type_code;type:varchar(64);not null"`
	RelationTypeName string     `gorm:"column:relation_type_name;type:varchar(128);not null"`
	Status           string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	Remark           *string    `gorm:"column:remark;type:text"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnitRelation) TableName() string { return "business_unit_relation" }

type BusinessUnitActor struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index"`
	BusinessUnitID  uint64     `gorm:"column:business_unit_id;not null;index"`
	ActorType       string     `gorm:"column:actor_type;type:varchar(32);not null"`
	ActorID         uint64     `gorm:"column:actor_id;not null;index"`
	RoleType        string     `gorm:"column:role_type;type:varchar(32);not null"`
	IncludeChildren bool       `gorm:"column:include_children;not null;default:false"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnitActor) TableName() string { return "business_unit_actor" }

type BusinessUnitAttrTemplate struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID      *uint64    `gorm:"column:tenant_id;index"`
	TemplateName  string     `gorm:"column:template_name;type:varchar(128);not null"`
	UnitTypeCode  string     `gorm:"column:unit_type_code;type:varchar(64);not null"`
	UnitTypeName  string     `gorm:"column:unit_type_name;type:varchar(128);not null"`
	UnitGroupCode *string    `gorm:"column:unit_group_code;type:varchar(64)"`
	UnitGroupName *string    `gorm:"column:unit_group_name;type:varchar(128)"`
	SortOrder     int        `gorm:"column:sort_order;not null;default:0"`
	Status        string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	Remark        *string    `gorm:"column:remark;type:text"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnitAttrTemplate) TableName() string { return "business_unit_attr_template" }

type BusinessUnitAttrTemplateField struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID     *uint64    `gorm:"column:tenant_id;index"`
	TemplateID   uint64     `gorm:"column:template_id;not null;index"`
	FieldKey     string     `gorm:"column:field_key;type:varchar(128);not null"`
	FieldLabel   string     `gorm:"column:field_label;type:varchar(128);not null"`
	FieldType    string     `gorm:"column:field_type;type:varchar(32);not null"`
	Required     bool       `gorm:"column:required;not null;default:false"`
	DefaultValue *string    `gorm:"column:default_value;type:text"`
	Placeholder  *string    `gorm:"column:placeholder;type:varchar(255)"`
	OptionsJSON  *string    `gorm:"column:options_json;type:jsonb"`
	SortOrder    int        `gorm:"column:sort_order;not null;default:0"`
	Status       string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnitAttrTemplateField) TableName() string { return "business_unit_attr_template_field" }

type BusinessUnitOrgMap struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index;index:idx_bu_org_map_scope,unique"`
	BusinessUnitID uint64     `gorm:"column:business_unit_id;not null;index;index:idx_bu_org_map_scope,unique"`
	OrgID          uint64     `gorm:"column:org_id;not null;index;index:idx_bu_org_map_scope,unique"`
	OrgType        string     `gorm:"column:org_type;type:varchar(32);not null"`
	ScopeType      string     `gorm:"column:scope_type;type:varchar(32);not null;index:idx_bu_org_map_scope,unique"`
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
	RolePermissionID uint64    `gorm:"column:role_permission_id;not null;index;index:idx_bu_scope_role_permission_bu,unique"`
	BusinessUnitID   uint64    `gorm:"column:business_unit_id;not null;index;index:idx_bu_scope_role_permission_bu,unique"`
	CreatedAt        time.Time `gorm:"column:created_at;not null"`
}

func (BusinessUnitScope) TableName() string { return "business_unit_scope" }

type BusinessResource struct {
	ID                   uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID             uint64     `gorm:"column:tenant_id;not null;index"`
	UnitTypeCode         string     `gorm:"column:unit_type_code;type:varchar(64);not null"`
	UnitTypeName         string     `gorm:"column:unit_type_name;type:varchar(64);not null"`
	BusinessUnitCode     string     `gorm:"column:business_unit_code;type:varchar(64);not null"`
	BusinessUnitName     string     `gorm:"column:business_unit_name;type:varchar(64);not null"`
	ResourceName         string     `gorm:"column:resource_name;type:varchar(128);not null"`
	ResourceCode         string     `gorm:"column:resource_code;type:varchar(64);not null"`
	ResourceCategory     string     `gorm:"column:resource_category;type:varchar(64);not null"`
	ResourceType         string     `gorm:"column:resource_type;type:varchar(64);not null"`
	SourceMode           string     `gorm:"column:source_mode;type:varchar(32);not null;default:native"`
	SourceAppCode        *string    `gorm:"column:source_app_code;type:varchar(64)"`
	SourceTable          *string    `gorm:"column:source_table;type:varchar(128)"`
	SourceID             *uint64    `gorm:"column:source_id"`
	PlatformCode         *string    `gorm:"column:platform_code;type:varchar(64)"`
	ExternalID           *string    `gorm:"column:external_id;type:varchar(128)"`
	ConnectionInstanceID *uint64    `gorm:"column:connection_instance_id"`
	ParentResourceID     *uint64    `gorm:"column:parent_resource_id;index"`
	ResourceAttrs        *string    `gorm:"column:resource_attrs;type:jsonb"`
	Status               string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
}

func (BusinessResource) TableName() string { return "business_resource" }

type BusinessResourceActor struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index"`
	ResourceID      uint64     `gorm:"column:resource_id;not null;index"`
	ActorType       string     `gorm:"column:actor_type;type:varchar(32);not null"`
	ActorID         uint64     `gorm:"column:actor_id;not null;index"`
	RoleType        string     `gorm:"column:role_type;type:varchar(32);not null"`
	IncludeChildren bool       `gorm:"column:include_children;not null;default:false"`
	StartDate       *time.Time `gorm:"column:start_date"`
	EndDate         *time.Time `gorm:"column:end_date"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (BusinessResourceActor) TableName() string { return "business_resource_actor" }

type BusinessResourceFieldConfig struct {
	ID                       uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID                 *uint64    `gorm:"column:tenant_id;index"`
	UnitTypeCode             string     `gorm:"column:unit_type_code;type:varchar(64);not null"`
	BusinessUnitCode         string     `gorm:"column:business_unit_code;type:varchar(64);not null"`
	FieldKey                 string     `gorm:"column:field_key;type:varchar(128);not null"`
	FieldLabel               string     `gorm:"column:field_label;type:varchar(128);not null"`
	FieldType                string     `gorm:"column:field_type;type:varchar(32);not null"`
	DictCode                 *string    `gorm:"column:dict_code;type:varchar(128)"`
	RelationUnitTypeCode     *string    `gorm:"column:relation_unit_type_code;type:varchar(64)"`
	RelationBusinessUnitCode *string    `gorm:"column:relation_business_unit_code;type:varchar(64)"`
	Required                 bool       `gorm:"column:required;not null;default:false"`
	DefaultValue             *string    `gorm:"column:default_value;type:text"`
	Placeholder              *string    `gorm:"column:placeholder;type:varchar(256)"`
	HelpText                 *string    `gorm:"column:help_text;type:varchar(256)"`
	ValidationRule           *string    `gorm:"column:validation_rule;type:jsonb"`
	ShowInList               bool       `gorm:"column:show_in_list;not null;default:false"`
	ShowInDetail             bool       `gorm:"column:show_in_detail;not null"`
	ShowInImport             bool       `gorm:"column:show_in_import;not null"`
	ImportRequired           bool       `gorm:"column:import_required;not null;default:false"`
	SortOrder                int        `gorm:"column:sort_order;not null;default:0"`
	Status                   string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	CreatedAt                time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt                time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt                *time.Time `gorm:"column:deleted_at"`
}

func (BusinessResourceFieldConfig) TableName() string { return "business_resource_field_config" }

type BusinessUnitResource struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index"`
	BusinessUnitID   uint64     `gorm:"column:business_unit_id;not null;index"`
	ResourceID       uint64     `gorm:"column:resource_id;not null;index"`
	ResourceCategory string     `gorm:"column:resource_category;type:varchar(64);not null"`
	ResourceType     string     `gorm:"column:resource_type;type:varchar(64);not null"`
	RelationType     string     `gorm:"column:relation_type;type:varchar(64);not null"`
	IsPrimary        bool       `gorm:"column:is_primary;not null;default:false"`
	UseForPermission bool       `gorm:"column:use_for_permission;not null;default:false"`
	UseForOperation  bool       `gorm:"column:use_for_operation;not null;default:false"`
	UseForSettlement bool       `gorm:"column:use_for_settlement;not null;default:false"`
	StartDate        *time.Time `gorm:"column:start_date"`
	EndDate          *time.Time `gorm:"column:end_date"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (BusinessUnitResource) TableName() string { return "business_unit_resource" }

type BusinessResourceRelation struct {
	ID                     uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID               uint64     `gorm:"column:tenant_id;not null;index"`
	ParentResourceID       uint64     `gorm:"column:parent_resource_id;not null;index"`
	ChildResourceID        uint64     `gorm:"column:child_resource_id;not null;index"`
	ParentResourceCategory string     `gorm:"column:parent_resource_category;type:varchar(64);not null"`
	ParentResourceType     string     `gorm:"column:parent_resource_type;type:varchar(64);not null"`
	ChildResourceCategory  string     `gorm:"column:child_resource_category;type:varchar(64);not null"`
	ChildResourceType      string     `gorm:"column:child_resource_type;type:varchar(64);not null"`
	RelationType           string     `gorm:"column:relation_type;type:varchar(64);not null"`
	StartDate              *time.Time `gorm:"column:start_date"`
	EndDate                *time.Time `gorm:"column:end_date"`
	Status                 string     `gorm:"column:status;type:varchar(32);not null;default:active"`
	CreatedAt              time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt              *time.Time `gorm:"column:deleted_at"`
}

func (BusinessResourceRelation) TableName() string { return "business_resource_relation" }

type DictType struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index:idx_dict_type_tenant_code,unique"`
	Code           string     `gorm:"column:code;type:varchar(64);not null;index:idx_dict_type_tenant_code,unique"`
	Name           string     `gorm:"column:name;type:varchar(200);not null"`
	Remark         *string    `gorm:"column:remark;type:varchar(500)"`
	Scope          string     `gorm:"column:scope;type:varchar(32);not null"`
	TenantEditable bool       `gorm:"column:tenant_editable;not null;default:true"`
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
	ParentID   *uint64    `gorm:"column:parent_id;index"`
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
	TenantID    uint64    `gorm:"column:tenant_id;not null;index;index:idx_tenant_dict_item_override,unique"`
	DictItemID  uint64    `gorm:"column:dict_item_id;not null;index;index:idx_tenant_dict_item_override,unique"`
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
	TenantID     uint64    `gorm:"column:tenant_id;not null;index;index:idx_tenant_menu_override,unique"`
	PermissionID uint64    `gorm:"column:permission_id;not null;index;index:idx_tenant_menu_override,unique"`
	CustomName   *string   `gorm:"column:custom_name;type:varchar(200)"`
	CustomIcon   *string   `gorm:"column:custom_icon;type:varchar(100)"`
	Enabled      *bool     `gorm:"column:enabled"`
	Visible      *bool     `gorm:"column:visible"`
	SortOrder    *int      `gorm:"column:sort_order"`
	Source       string    `gorm:"column:source;type:varchar(32);not null;default:'MANUAL'"`
	SourceRef    *string   `gorm:"column:source_ref;type:varchar(128)"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

func (TenantMenuOverride) TableName() string { return "tenant_menu_override" }

type TenantParamValue struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   uint64    `gorm:"column:tenant_id;not null;index;index:idx_tenant_param_value,unique"`
	ParamID    uint64    `gorm:"column:param_id;not null;index;index:idx_tenant_param_value,unique"`
	ParamValue *string   `gorm:"column:param_value;type:text"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null"`
}

func (TenantParamValue) TableName() string { return "tenant_param_value" }

type AppUserDepartment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID       uint64    `gorm:"column:user_id;not null;index;index:idx_app_user_department,unique"`
	DepartmentID uint64    `gorm:"column:department_id;not null;index;index:idx_app_user_department,unique"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (AppUserDepartment) TableName() string { return "app_user_department" }

type AppUserPosition struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     uint64    `gorm:"column:user_id;not null;index;index:idx_app_user_position,unique"`
	PositionID uint64    `gorm:"column:position_id;not null;index;index:idx_app_user_position,unique"`
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
	UserID    uint64    `gorm:"column:user_id;not null;index;index:idx_user_preference_key,unique"`
	PrefKey   string    `gorm:"column:pref_key;type:varchar(64);not null;index:idx_user_preference_key,unique"`
	PrefValue *string   `gorm:"column:pref_value;type:text"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (UserPreference) TableName() string { return "user_preference" }
