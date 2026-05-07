package models

import "time"

type SystemParam struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;default:1;index:idx_sys_param_tenant_key,unique"`
	Key            string     `gorm:"column:param_key;type:varchar(128);not null;index:idx_sys_param_tenant_key,unique"`
	Value          string     `gorm:"column:param_value;type:text;not null"`
	Remark         string     `gorm:"column:remark;type:varchar(500);not null;default:''"`
	ValueType      string     `gorm:"column:value_type;type:varchar(32);not null;default:'string'"`
	TenantEditable bool       `gorm:"column:tenant_editable;not null;default:true"`
	IsPlatformOnly bool       `gorm:"column:is_platform_only;not null;default:false"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (SystemParam) TableName() string {
	return "sys_param"
}
