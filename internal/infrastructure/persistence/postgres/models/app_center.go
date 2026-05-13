package models

import "time"

type SysApp struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode          string     `gorm:"column:app_code;type:varchar(100);uniqueIndex;not null"`
	AppName          string     `gorm:"column:app_name;type:varchar(100);not null"`
	Icon             *string    `gorm:"column:icon;type:varchar(80)"`
	AppType          string     `gorm:"column:app_type;type:varchar(32);not null;index"`
	Source           string     `gorm:"column:source;type:varchar(32);not null;index"`
	Status           string     `gorm:"column:status;type:varchar(32);not null;index"`
	ChargeMode       string     `gorm:"column:charge_mode;type:varchar(32);not null;default:'NON_SELLABLE'"`
	VisibilityScope  string     `gorm:"column:visibility_scope;type:varchar(32);not null;default:'PLATFORM_ONLY'"`
	Owner            *string    `gorm:"column:owner;type:varchar(100)"`
	OwnerUserIDs     *string    `gorm:"column:owner_user_ids;type:text"`
	Version          *string    `gorm:"column:version;type:varchar(64)"`
	Description      *string    `gorm:"column:description;type:varchar(500)"`
	DetailDesc       *string    `gorm:"column:detail_description;type:text"`
	DeploymentMode   string     `gorm:"column:deployment_mode;type:varchar(32);not null;default:'MERGED';index"`
	CommModes        *string    `gorm:"column:communication_modes;type:text"`
	VisibilityMode   *string    `gorm:"column:visibility_mode;type:varchar(64)"`
	VisibleTenants   *string    `gorm:"column:visible_tenants;type:text"`
	OpenMethod       *string    `gorm:"column:open_method;type:varchar(200)"`
	TrialPolicy      *string    `gorm:"column:trial_policy;type:varchar(100)"`
	TrialStartRule   *string    `gorm:"column:trial_start_rule;type:varchar(100)"`
	AssetConfig      *string    `gorm:"column:asset_config;type:text"`
	DocConfig        *string    `gorm:"column:doc_config;type:text"`
	ReleaseChannel   *string    `gorm:"column:release_channel;type:varchar(32)"`
	ReleaseNote      *string    `gorm:"column:release_note;type:text"`
	HealthCheckURL   *string    `gorm:"column:health_check_url;type:varchar(500)"`
	APIBaseURL       *string    `gorm:"column:api_base_url;type:varchar(500)"`
	WebhookURL       *string    `gorm:"column:webhook_url;type:varchar(500)"`
	ManifestHash     *string    `gorm:"column:manifest_hash;type:varchar(128)"`
	ManifestVersion  *string    `gorm:"column:manifest_version;type:varchar(64)"`
	LastManifestSync *time.Time `gorm:"column:last_manifest_synced_at"`
	IsBuiltin        bool       `gorm:"column:is_builtin;not null;default:false"`
	IsPlatformOnly   bool       `gorm:"column:is_platform_only;not null;default:true"`
	SortOrder        int        `gorm:"column:sort_order;not null;default:0"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;index"`
}

func (SysApp) TableName() string { return "sys_app" }

type SysAppClient struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppID      uint64     `gorm:"column:app_id;not null;index;index:idx_sys_app_client_app_code,unique"`
	ClientCode string     `gorm:"column:client_code;type:varchar(64);not null;index:idx_sys_app_client_app_code,unique"`
	ClientName string     `gorm:"column:client_name;type:varchar(100);not null"`
	Enabled    bool       `gorm:"column:enabled;not null;default:true"`
	SortOrder  int        `gorm:"column:sort_order;not null;default:0"`
	ConfigNote *string    `gorm:"column:config_note;type:varchar(500)"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppClient) TableName() string { return "sys_app_client" }

type SysAppManifestLoad struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode         string     `gorm:"column:app_code;type:varchar(100);not null;index"`
	Action          string     `gorm:"column:action;type:varchar(32);not null;default:'LOAD'"`
	SourceType      string     `gorm:"column:source_type;type:varchar(32);not null;default:'UPLOAD'"`
	SourceName      *string    `gorm:"column:source_name;type:varchar(500)"`
	ManifestVersion string     `gorm:"column:manifest_version;type:varchar(64);not null;default:'1.0'"`
	ManifestHash    string     `gorm:"column:manifest_hash;type:varchar(128);not null"`
	FragmentRole    string     `gorm:"column:fragment_role;type:varchar(64);not null;default:'main'"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;index"`
	Summary         *string    `gorm:"column:summary;type:text"`
	ErrorSummary    *string    `gorm:"column:error_summary;type:text"`
	DiffSummary     *string    `gorm:"column:diff_summary;type:text"`
	OperatorUserID  uint64     `gorm:"column:operator_user_id;not null;index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppManifestLoad) TableName() string { return "sys_app_manifest_load" }

type SysAppManifestFile struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	LoadID          *uint64    `gorm:"column:load_id;index"`
	AppCode         string     `gorm:"column:app_code;type:varchar(100);not null;index"`
	FileName        string     `gorm:"column:file_name;type:varchar(255);not null"`
	FilePath        *string    `gorm:"column:file_path;type:varchar(1000)"`
	FragmentRole    string     `gorm:"column:fragment_role;type:varchar(64);not null;default:'main'"`
	ManifestVersion string     `gorm:"column:manifest_version;type:varchar(64);not null;default:'1.0'"`
	ManifestHash    string     `gorm:"column:manifest_hash;type:varchar(128);not null;index"`
	ContentSummary  *string    `gorm:"column:content_summary;type:text"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppManifestFile) TableName() string { return "sys_app_manifest_file" }

type SysAppEntry struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode           string     `gorm:"column:app_code;type:varchar(100);not null;index:idx_sys_app_entry_code,unique"`
	ResourceCode      string     `gorm:"column:resource_code;type:varchar(120);not null;index:idx_sys_app_entry_code,unique"`
	Name              string     `gorm:"column:name;type:varchar(200);not null"`
	Path              string     `gorm:"column:path;type:varchar(500);not null;index"`
	ParentCode        *string    `gorm:"column:parent_code;type:varchar(120)"`
	SortOrder         int        `gorm:"column:sort_order;not null;default:0"`
	PlatformOnly      bool       `gorm:"column:platform_only;not null;default:false"`
	TenantVisible     bool       `gorm:"column:tenant_visible;not null;default:true"`
	TenantEditable    bool       `gorm:"column:tenant_editable;not null;default:false"`
	IncludeInPackage  bool       `gorm:"column:include_in_package;not null;default:false"`
	FeatureCode       *string    `gorm:"column:feature_code;type:varchar(100);index"`
	DataPermMode      string     `gorm:"column:data_perm_mode;type:varchar(16);not null;default:'ORG'"`
	ManifestHash      string     `gorm:"column:manifest_hash;type:varchar(128);not null"`
	ManagedByManifest bool       `gorm:"column:managed_by_manifest;not null;default:true"`
	ProtectionSource  *string    `gorm:"column:protection_source;type:varchar(32)"`
	ProtectionReason  *string    `gorm:"column:protection_reason;type:varchar(500)"`
	ProtectedAt       *time.Time `gorm:"column:protected_at"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE';index"`
	LastSyncedAt      time.Time  `gorm:"column:last_synced_at;not null"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppEntry) TableName() string { return "sys_app_entry" }

type SysAppAPI struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode           string     `gorm:"column:app_code;type:varchar(100);not null;index:idx_sys_app_api_route,unique"`
	Method            string     `gorm:"column:method;type:varchar(20);not null;index:idx_sys_app_api_route,unique"`
	Path              string     `gorm:"column:path;type:varchar(500);not null;index:idx_sys_app_api_route,unique"`
	PermissionCode    *string    `gorm:"column:permission_code;type:varchar(120);index"`
	Public            bool       `gorm:"column:public;not null;default:false"`
	Audit             bool       `gorm:"column:audit;not null;default:false"`
	ManifestHash      string     `gorm:"column:manifest_hash;type:varchar(128);not null"`
	ManagedByManifest bool       `gorm:"column:managed_by_manifest;not null;default:true"`
	ProtectionSource  *string    `gorm:"column:protection_source;type:varchar(32)"`
	ProtectionReason  *string    `gorm:"column:protection_reason;type:varchar(500)"`
	ProtectedAt       *time.Time `gorm:"column:protected_at"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE';index"`
	LastSyncedAt      time.Time  `gorm:"column:last_synced_at;not null"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppAPI) TableName() string { return "sys_app_api" }

type SysAppPermission struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode           string     `gorm:"column:app_code;type:varchar(100);not null;index:idx_sys_app_perm_code,unique"`
	PermissionCode    string     `gorm:"column:permission_code;type:varchar(120);not null;index:idx_sys_app_perm_code,unique"`
	Name              string     `gorm:"column:name;type:varchar(200);not null"`
	PermissionType    string     `gorm:"column:permission_type;type:varchar(32);not null"`
	MenuCode          *string    `gorm:"column:menu_code;type:varchar(120);index"`
	PlatformOnly      bool       `gorm:"column:platform_only;not null;default:false"`
	IncludeInPackage  bool       `gorm:"column:include_in_package;not null;default:false"`
	DataPermMode      string     `gorm:"column:data_perm_mode;type:varchar(16);not null;default:'ORG'"`
	ManifestHash      string     `gorm:"column:manifest_hash;type:varchar(128);not null"`
	ManagedByManifest bool       `gorm:"column:managed_by_manifest;not null;default:true"`
	ProtectionSource  *string    `gorm:"column:protection_source;type:varchar(32)"`
	ProtectionReason  *string    `gorm:"column:protection_reason;type:varchar(500)"`
	ProtectedAt       *time.Time `gorm:"column:protected_at"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE';index"`
	LastSyncedAt      time.Time  `gorm:"column:last_synced_at;not null"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppPermission) TableName() string { return "sys_app_permission" }

type SysAppPackageFeature struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode           string     `gorm:"column:app_code;type:varchar(100);not null;index:idx_sys_app_feature_code,unique"`
	FeatureCode       string     `gorm:"column:feature_code;type:varchar(100);not null;index:idx_sys_app_feature_code,unique"`
	FeatureName       string     `gorm:"column:feature_name;type:varchar(200);not null"`
	FeatureType       string     `gorm:"column:feature_type;type:varchar(32);not null"`
	ParentCode        *string    `gorm:"column:parent_code;type:varchar(100)"`
	SourceCode        *string    `gorm:"column:source_code;type:varchar(120)"`
	PackagePolicy     string     `gorm:"column:package_policy;type:varchar(32);not null;default:'IN_PACKAGE'"`
	IncludeInPackage  bool       `gorm:"column:include_in_package;not null;default:true"`
	Description       *string    `gorm:"column:description;type:varchar(500)"`
	ManifestHash      string     `gorm:"column:manifest_hash;type:varchar(128);not null"`
	ManagedByManifest bool       `gorm:"column:managed_by_manifest;not null;default:true"`
	ProtectionSource  *string    `gorm:"column:protection_source;type:varchar(32)"`
	ProtectionReason  *string    `gorm:"column:protection_reason;type:varchar(500)"`
	ProtectedAt       *time.Time `gorm:"column:protected_at"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE';index"`
	LastSyncedAt      time.Time  `gorm:"column:last_synced_at;not null"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppPackageFeature) TableName() string { return "sys_app_package_feature" }

type SysAppQuota struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode           string     `gorm:"column:app_code;type:varchar(100);not null;index:idx_sys_app_quota_code,unique"`
	QuotaCode         string     `gorm:"column:quota_code;type:varchar(100);not null;index:idx_sys_app_quota_code,unique"`
	QuotaName         string     `gorm:"column:quota_name;type:varchar(200);not null"`
	QuotaType         string     `gorm:"column:quota_type;type:varchar(32);not null"`
	Unit              *string    `gorm:"column:unit;type:varchar(32)"`
	PeriodType        *string    `gorm:"column:period_type;type:varchar(32)"`
	IncludeInPackage  bool       `gorm:"column:include_in_package;not null;default:true"`
	Description       *string    `gorm:"column:description;type:varchar(500)"`
	ManifestHash      string     `gorm:"column:manifest_hash;type:varchar(128);not null"`
	ManagedByManifest bool       `gorm:"column:managed_by_manifest;not null;default:true"`
	ProtectionSource  *string    `gorm:"column:protection_source;type:varchar(32)"`
	ProtectionReason  *string    `gorm:"column:protection_reason;type:varchar(500)"`
	ProtectedAt       *time.Time `gorm:"column:protected_at"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE';index"`
	LastSyncedAt      time.Time  `gorm:"column:last_synced_at;not null"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (SysAppQuota) TableName() string { return "sys_app_quota" }
