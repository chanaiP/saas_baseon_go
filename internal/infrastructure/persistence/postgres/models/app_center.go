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
	Version         *string    `gorm:"column:version;type:varchar(64)"`
	Description     *string    `gorm:"column:description;type:varchar(500)"`
	IsBuiltin       bool       `gorm:"column:is_builtin;not null;default:false"`
	IsPlatformOnly  bool       `gorm:"column:is_platform_only;not null;default:true"`
	SortOrder       int        `gorm:"column:sort_order;not null;default:0"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (SysApp) TableName() string { return "sys_app" }
