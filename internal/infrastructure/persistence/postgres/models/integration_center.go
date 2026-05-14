package models

import "time"

type IntegrationPlatform struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	PlatformCode      string     `gorm:"column:platform_code;type:varchar(80);not null"`
	PlatformName      string     `gorm:"column:platform_name;type:varchar(120);not null"`
	PlatformShortName *string    `gorm:"column:platform_short_name;type:varchar(80)"`
	PlatformType      string     `gorm:"column:platform_type;type:varchar(50);not null"`
	AccessMode        string     `gorm:"column:access_mode;type:varchar(50);not null"`
	LogoURL           *string    `gorm:"column:logo_url;type:varchar(500)"`
	OfficialURL       *string    `gorm:"column:official_url;type:varchar(500)"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'draft'"`
	TenantVisible     bool       `gorm:"column:tenant_visible;not null;default:false"`
	OwnerName         *string    `gorm:"column:owner_name;type:varchar(80)"`
	SortOrder         int        `gorm:"column:sort_order;not null;default:0"`
	Description       *string    `gorm:"column:description;type:text"`
	CreatedBy         *uint64    `gorm:"column:created_by"`
	UpdatedBy         *uint64    `gorm:"column:updated_by"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationPlatform) TableName() string { return "integration_platforms" }

type IntegrationPlatformCapability struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	PlatformID     uint64     `gorm:"column:platform_id;not null;index"`
	CapabilityCode string     `gorm:"column:capability_code;type:varchar(100);not null"`
	CapabilityName string     `gorm:"column:capability_name;type:varchar(120);not null"`
	CapabilityType string     `gorm:"column:capability_type;type:varchar(50);not null"`
	AuthScopeCode  *string    `gorm:"column:auth_scope_code;type:varchar(120)"`
	DataDirection  string     `gorm:"column:data_direction;type:varchar(32);not null;default:'pull'"`
	Status         string     `gorm:"column:status;type:varchar(32);not null;default:'enabled'"`
	Description    *string    `gorm:"column:description;type:text"`
	CreatedBy      *uint64    `gorm:"column:created_by"`
	UpdatedBy      *uint64    `gorm:"column:updated_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationPlatformCapability) TableName() string { return "integration_platform_capabilities" }

type IntegrationProviderApp struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	PlatformID    uint64     `gorm:"column:platform_id;not null;index"`
	AppCode       string     `gorm:"column:app_code;type:varchar(100);not null"`
	AppName       string     `gorm:"column:app_name;type:varchar(150);not null"`
	AppType       string     `gorm:"column:app_type;type:varchar(50);not null;default:'provider_app'"`
	AuthMode      string     `gorm:"column:auth_mode;type:varchar(50);not null"`
	Environment   string     `gorm:"column:environment;type:varchar(32);not null;default:'prod'"`
	Status        string     `gorm:"column:status;type:varchar(32);not null;default:'draft'"`
	TenantVisible bool       `gorm:"column:tenant_visible;not null;default:false"`
	CallbackURL   *string    `gorm:"column:callback_url;type:varchar(500)"`
	WebhookURL    *string    `gorm:"column:webhook_url;type:varchar(500)"`
	CredentialRef *string    `gorm:"column:credential_ref;type:varchar(200)"`
	OwnerName     *string    `gorm:"column:owner_name;type:varchar(80)"`
	Description   *string    `gorm:"column:description;type:text"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	UpdatedBy     *uint64    `gorm:"column:updated_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationProviderApp) TableName() string { return "integration_provider_apps" }

type IntegrationProviderAppCapability struct {
	ID                   uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	ProviderAppID        uint64     `gorm:"column:provider_app_id;not null;index"`
	PlatformCapabilityID uint64     `gorm:"column:platform_capability_id;not null;index"`
	ConnectionStatus     string     `gorm:"column:connection_status;type:varchar(32);not null;default:'pending'"`
	ReviewStatus         string     `gorm:"column:review_status;type:varchar(32);not null;default:'pending'"`
	Enabled              bool       `gorm:"column:enabled;not null;default:false"`
	Config               string     `gorm:"column:config;type:jsonb;not null;default:'{}'"`
	CreatedBy            *uint64    `gorm:"column:created_by"`
	UpdatedBy            *uint64    `gorm:"column:updated_by"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt            *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationProviderAppCapability) TableName() string {
	return "integration_provider_app_capabilities"
}

type IntegrationTenantConnection struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index"`
	PlatformID       uint64     `gorm:"column:platform_id;not null;index"`
	ProviderAppID    uint64     `gorm:"column:provider_app_id;not null;index"`
	ConnectionName   string     `gorm:"column:connection_name;type:varchar(180);not null"`
	AuthSubjectType  string     `gorm:"column:auth_subject_type;type:varchar(60);not null"`
	AuthSubjectID    string     `gorm:"column:auth_subject_id;type:varchar(160);not null"`
	AuthSubjectName  string     `gorm:"column:auth_subject_name;type:varchar(180);not null"`
	AuthScope        string     `gorm:"column:auth_scope;type:jsonb;not null;default:'[]'"`
	AuthStatus       string     `gorm:"column:auth_status;type:varchar(32);not null;default:'pending'"`
	ConnectionStatus string     `gorm:"column:connection_status;type:varchar(32);not null;default:'inactive'"`
	TokenStatus      string     `gorm:"column:token_status;type:varchar(32);not null;default:'unknown'"`
	AuthorizedAt     *time.Time `gorm:"column:authorized_at"`
	TokenExpiresAt   *time.Time `gorm:"column:token_expires_at"`
	LastSyncAt       *time.Time `gorm:"column:last_sync_at"`
	LastErrorAt      *time.Time `gorm:"column:last_error_at"`
	LastErrorMessage *string    `gorm:"column:last_error_message;type:text"`
	CreatedBy        *uint64    `gorm:"column:created_by"`
	UpdatedBy        *uint64    `gorm:"column:updated_by"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationTenantConnection) TableName() string { return "integration_tenant_connections" }

type IntegrationTenantCapability struct {
	ID                      uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID                uint64     `gorm:"column:tenant_id;not null;index"`
	TenantConnectionID      uint64     `gorm:"column:tenant_connection_id;not null;index"`
	ProviderAppCapabilityID uint64     `gorm:"column:provider_app_capability_id;not null;index"`
	CapabilityCode          string     `gorm:"column:capability_code;type:varchar(100);not null"`
	CapabilityName          string     `gorm:"column:capability_name;type:varchar(120);not null"`
	Enabled                 bool       `gorm:"column:enabled;not null;default:true"`
	Source                  string     `gorm:"column:source;type:varchar(32);not null;default:'authorization'"`
	EffectiveScope          string     `gorm:"column:effective_scope;type:jsonb;not null;default:'{}'"`
	CreatedBy               *uint64    `gorm:"column:created_by"`
	UpdatedBy               *uint64    `gorm:"column:updated_by"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt               *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationTenantCapability) TableName() string { return "integration_tenant_capabilities" }

type IntegrationSyncJob struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID           uint64     `gorm:"column:tenant_id;not null;index"`
	TenantConnectionID uint64     `gorm:"column:tenant_connection_id;not null;index"`
	CapabilityCode     string     `gorm:"column:capability_code;type:varchar(100);not null"`
	JobType            string     `gorm:"column:job_type;type:varchar(50);not null"`
	TriggerMode        string     `gorm:"column:trigger_mode;type:varchar(50);not null;default:'manual'"`
	Status             string     `gorm:"column:status;type:varchar(32);not null;default:'pending'"`
	CursorValue        *string    `gorm:"column:cursor_value;type:varchar(300)"`
	TotalCount         int64      `gorm:"column:total_count;not null;default:0"`
	SuccessCount       int64      `gorm:"column:success_count;not null;default:0"`
	FailedCount        int64      `gorm:"column:failed_count;not null;default:0"`
	StartedAt          *time.Time `gorm:"column:started_at"`
	FinishedAt         *time.Time `gorm:"column:finished_at"`
	ErrorCode          *string    `gorm:"column:error_code;type:varchar(100)"`
	ErrorMessage       *string    `gorm:"column:error_message;type:text"`
	CreatedBy          *uint64    `gorm:"column:created_by"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt          *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationSyncJob) TableName() string { return "integration_sync_jobs" }

type IntegrationQuotaPolicy struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	PolicyCode      string     `gorm:"column:policy_code;type:varchar(100);not null"`
	PolicyName      string     `gorm:"column:policy_name;type:varchar(150);not null"`
	QuotaCode       string     `gorm:"column:quota_code;type:varchar(100);not null"`
	QuotaUnit       string     `gorm:"column:quota_unit;type:varchar(32);not null"`
	PeriodType      string     `gorm:"column:period_type;type:varchar(32);not null"`
	DefaultLimit    int64      `gorm:"column:default_limit;not null;default:0"`
	OverLimitAction string     `gorm:"column:over_limit_action;type:varchar(32);not null;default:'reject'"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;default:'enabled'"`
	Description     *string    `gorm:"column:description;type:text"`
	CreatedBy       *uint64    `gorm:"column:created_by"`
	UpdatedBy       *uint64    `gorm:"column:updated_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationQuotaPolicy) TableName() string { return "integration_quota_policies" }

type IntegrationAlert struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID           *uint64    `gorm:"column:tenant_id;index"`
	TenantConnectionID *uint64    `gorm:"column:tenant_connection_id;index"`
	PlatformID         *uint64    `gorm:"column:platform_id;index"`
	ProviderAppID      *uint64    `gorm:"column:provider_app_id;index"`
	AlertType          string     `gorm:"column:alert_type;type:varchar(60);not null"`
	Severity           string     `gorm:"column:severity;type:varchar(32);not null;default:'warning'"`
	Status             string     `gorm:"column:status;type:varchar(32);not null;default:'open'"`
	Title              string     `gorm:"column:title;type:varchar(180);not null"`
	Message            *string    `gorm:"column:message;type:text"`
	FirstSeenAt        time.Time  `gorm:"column:first_seen_at;not null"`
	LastSeenAt         time.Time  `gorm:"column:last_seen_at;not null"`
	ResolvedAt         *time.Time `gorm:"column:resolved_at"`
	HandledBy          *uint64    `gorm:"column:handled_by"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt          *time.Time `gorm:"column:deleted_at;index"`
}

func (IntegrationAlert) TableName() string { return "integration_alerts" }

type IntegrationAPICallLog struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID           *uint64   `gorm:"column:tenant_id;index"`
	TenantConnectionID *uint64   `gorm:"column:tenant_connection_id;index"`
	PlatformID         *uint64   `gorm:"column:platform_id;index"`
	ProviderAppID      *uint64   `gorm:"column:provider_app_id;index"`
	RequestID          string    `gorm:"column:request_id;type:varchar(120);not null"`
	CallType           string    `gorm:"column:call_type;type:varchar(60);not null"`
	Method             *string   `gorm:"column:method;type:varchar(16)"`
	Endpoint           *string   `gorm:"column:endpoint;type:varchar(500)"`
	Status             string    `gorm:"column:status;type:varchar(32);not null"`
	HTTPStatus         *int      `gorm:"column:http_status"`
	DurationMS         int       `gorm:"column:duration_ms;not null;default:0"`
	ErrorCode          *string   `gorm:"column:error_code;type:varchar(120)"`
	ErrorMessage       *string   `gorm:"column:error_message;type:text"`
	RequestDigest      *string   `gorm:"column:request_digest;type:varchar(128)"`
	ResponseDigest     *string   `gorm:"column:response_digest;type:varchar(128)"`
	CalledAt           time.Time `gorm:"column:called_at;not null"`
	CreatedAt          time.Time `gorm:"column:created_at;not null"`
}

func (IntegrationAPICallLog) TableName() string { return "integration_api_call_logs" }
