package models

import "time"

type LoginLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  *uint64   `gorm:"column:tenant_id;index"`
	UserID    *uint64   `gorm:"column:user_id;index"`
	Account   string    `gorm:"column:account;type:varchar(128);not null"`
	Success   bool      `gorm:"column:success;not null"`
	Message   *string   `gorm:"column:message;type:varchar(500)"`
	IP        *string   `gorm:"column:ip;type:varchar(64)"`
	UserAgent *string   `gorm:"column:user_agent;type:varchar(500)"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (LoginLog) TableName() string { return "login_log" }

type AuditLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  *uint64   `gorm:"column:tenant_id;index"`
	UserID    *uint64   `gorm:"column:user_id;index"`
	Module    string    `gorm:"column:module;type:varchar(64);not null"`
	Action    string    `gorm:"column:action;type:varchar(32);not null"`
	Summary   string    `gorm:"column:summary;type:varchar(500);not null"`
	Detail    *string   `gorm:"column:detail;type:text"`
	IP        *string   `gorm:"column:ip;type:varchar(64)"`
	UserAgent *string   `gorm:"column:user_agent;type:varchar(500)"`
	RequestID *string   `gorm:"column:request_id;type:varchar(64);index"`
	Result    string    `gorm:"column:result;type:varchar(32);not null;default:'success'"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (AuditLog) TableName() string { return "audit_log" }
