package models

import "time"

type SysApp struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	AppCode         string     `gorm:"column:app_code;type:varchar(100);uniqueIndex;not null"`
	AppName         string     `gorm:"column:app_name;type:varchar(100);not null"`
	Icon            *string    `gorm:"column:icon;type:varchar(80)"`
	AppType         string     `gorm:"column:app_type;type:varchar(32);not null;index"`
	Source          string     `gorm:"column:source;type:varchar(32);not null;index"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;index"`
	ChargeMode      string     `gorm:"column:charge_mode;type:varchar(32);not null;default:'NON_SELLABLE'"`
	VisibilityScope string     `gorm:"column:visibility_scope;type:varchar(32);not null;default:'PLATFORM_ONLY'"`
	Owner           *string    `gorm:"column:owner;type:varchar(100)"`
	OwnerUserIDs    *string    `gorm:"column:owner_user_ids;type:text"`
	Version         *string    `gorm:"column:version;type:varchar(64)"`
	Description     *string    `gorm:"column:description;type:varchar(500)"`
	DetailDesc      *string    `gorm:"column:detail_description;type:text"`
	DeploymentMode  string     `gorm:"column:deployment_mode;type:varchar(32);not null;default:'MERGED';index"`
	CommModes       *string    `gorm:"column:communication_modes;type:text"`
	VisibilityMode  *string    `gorm:"column:visibility_mode;type:varchar(64)"`
	VisibleTenants  *string    `gorm:"column:visible_tenants;type:text"`
	OpenMethod      *string    `gorm:"column:open_method;type:varchar(200)"`
	TrialPolicy     *string    `gorm:"column:trial_policy;type:varchar(100)"`
	TrialStartRule  *string    `gorm:"column:trial_start_rule;type:varchar(100)"`
	AssetConfig     *string    `gorm:"column:asset_config;type:text"`
	DocConfig       *string    `gorm:"column:doc_config;type:text"`
	ReleaseChannel  *string    `gorm:"column:release_channel;type:varchar(32)"`
	ReleaseNote     *string    `gorm:"column:release_note;type:text"`
	IsBuiltin       bool       `gorm:"column:is_builtin;not null;default:false"`
	IsPlatformOnly  bool       `gorm:"column:is_platform_only;not null;default:true"`
	SortOrder       int        `gorm:"column:sort_order;not null;default:0"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
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
