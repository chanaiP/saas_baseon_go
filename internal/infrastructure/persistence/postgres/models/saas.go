package models

import "time"

type SaasFeature struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	FeatureCode string    `gorm:"column:feature_code;type:varchar(100);uniqueIndex;not null"`
	FeatureName string    `gorm:"column:feature_name;type:varchar(100);not null"`
	FeatureType string    `gorm:"column:feature_type;type:varchar(32);not null"`
	ParentID    uint64    `gorm:"column:parent_id;not null;default:0"`
	MenuID      *uint64   `gorm:"column:menu_id"`
	APIMethod   *string   `gorm:"column:api_method;type:varchar(20)"`
	APIPath     *string   `gorm:"column:api_path;type:varchar(255)"`
	ServiceKey  *string   `gorm:"column:service_key;type:varchar(100)"`
	Status      int       `gorm:"column:status;not null;default:1"`
	Description *string   `gorm:"column:description;type:varchar(500)"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

func (SaasFeature) TableName() string { return "saas_feature" }

type SaasPlan struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	PlanCode     string     `gorm:"column:plan_code;type:varchar(64);uniqueIndex;not null"`
	PlanName     string     `gorm:"column:plan_name;type:varchar(100);not null"`
	PlanType     string     `gorm:"column:plan_type;type:varchar(32);not null"`
	BillingCycle string     `gorm:"column:billing_cycle;type:varchar(32);not null"`
	Price        float64    `gorm:"column:price;type:numeric(12,2);not null;default:0"`
	Status       int        `gorm:"column:status;not null;default:1"`
	IsDefault    bool       `gorm:"column:is_default;not null;default:false"`
	SortOrder    int        `gorm:"column:sort_order;not null;default:0"`
	Description  *string    `gorm:"column:description;type:varchar(500)"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (SaasPlan) TableName() string { return "saas_plan" }

type SaasQuota struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	QuotaCode   string    `gorm:"column:quota_code;type:varchar(100);uniqueIndex;not null"`
	QuotaName   string    `gorm:"column:quota_name;type:varchar(100);not null"`
	QuotaType   string    `gorm:"column:quota_type;type:varchar(32);not null"`
	PeriodType  *string   `gorm:"column:period_type;type:varchar(32)"`
	Unit        *string   `gorm:"column:unit;type:varchar(32)"`
	Status      int       `gorm:"column:status;not null;default:1"`
	Description *string   `gorm:"column:description;type:varchar(500)"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

func (SaasQuota) TableName() string { return "saas_quota" }

type SaasPlanFeature struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	PlanID    uint64    `gorm:"column:plan_id;not null;index:idx_plan_feature,unique"`
	FeatureID uint64    `gorm:"column:feature_id;not null;index:idx_plan_feature,unique"`
	Enabled   bool      `gorm:"column:enabled;not null;default:true"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (SaasPlanFeature) TableName() string { return "saas_plan_feature" }

type SaasPlanQuota struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	PlanID     uint64    `gorm:"column:plan_id;not null;index:idx_plan_quota,unique"`
	QuotaID    uint64    `gorm:"column:quota_id;not null;index:idx_plan_quota,unique"`
	QuotaValue int       `gorm:"column:quota_value;not null;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null"`
}

func (SaasPlanQuota) TableName() string { return "saas_plan_quota" }

type TenantSubscription struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID           uint64     `gorm:"column:tenant_id;not null;index"`
	PlanID             uint64     `gorm:"column:plan_id;not null;index"`
	SubscriptionStatus string     `gorm:"column:subscription_status;type:varchar(32);not null"`
	StartTime          time.Time  `gorm:"column:start_time;not null"`
	EndTime            *time.Time `gorm:"column:end_time"`
	TrialEndTime       *time.Time `gorm:"column:trial_end_time"`
	AutoRenew          bool       `gorm:"column:auto_renew;not null;default:false"`
	FrozenReason       *string    `gorm:"column:frozen_reason;type:varchar(500)"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null"`
}

func (TenantSubscription) TableName() string { return "tenant_subscription" }

type TenantFeatureOverride struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  uint64     `gorm:"column:tenant_id;not null;index:idx_tenant_feature,unique"`
	FeatureID uint64     `gorm:"column:feature_id;not null;index:idx_tenant_feature,unique"`
	Enabled   bool       `gorm:"column:enabled;not null"`
	Reason    *string    `gorm:"column:reason;type:varchar(500)"`
	StartTime *time.Time `gorm:"column:start_time"`
	EndTime   *time.Time `gorm:"column:end_time"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
}

func (TenantFeatureOverride) TableName() string { return "tenant_feature_override" }

type TenantQuotaOverride struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   uint64     `gorm:"column:tenant_id;not null;index:idx_tenant_quota,unique"`
	QuotaID    uint64     `gorm:"column:quota_id;not null;index:idx_tenant_quota,unique"`
	QuotaValue int        `gorm:"column:quota_value;not null"`
	Reason     *string    `gorm:"column:reason;type:varchar(500)"`
	StartTime  *time.Time `gorm:"column:start_time"`
	EndTime    *time.Time `gorm:"column:end_time"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
}

func (TenantQuotaOverride) TableName() string { return "tenant_quota_override" }

type TenantQuotaUsage struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index:idx_tenant_quota_period,unique"`
	QuotaCode       string     `gorm:"column:quota_code;type:varchar(100);not null;index:idx_tenant_quota_period,unique"`
	UsedValue       int        `gorm:"column:used_value;not null;default:0"`
	LimitValue      int        `gorm:"column:limit_value;not null;default:0"`
	PeriodType      *string    `gorm:"column:period_type;type:varchar(32)"`
	PeriodKey       string     `gorm:"column:period_key;type:varchar(32);not null;index:idx_tenant_quota_period,unique"`
	LastRefreshTime *time.Time `gorm:"column:last_refresh_time"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
}

func (TenantQuotaUsage) TableName() string { return "tenant_quota_usage" }
