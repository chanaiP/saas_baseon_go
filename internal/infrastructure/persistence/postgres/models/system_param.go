package models

import "time"

type SystemParam struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	Key       string    `gorm:"column:param_key;type:varchar(128);uniqueIndex;not null"`
	Value     string    `gorm:"column:param_value;type:text;not null"`
	Remark    string    `gorm:"column:remark;type:text;not null;default:''"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (SystemParam) TableName() string {
	return "system_param"
}
