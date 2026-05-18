package models

import "time"

type AITimeFields struct {
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index"`
}

type AIProvider struct {
	ID            string  `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	Name          string  `json:"name" gorm:"column:name;type:varchar(100);not null"`
	Code          string  `json:"code" gorm:"column:code;type:varchar(100);not null;uniqueIndex"`
	Type          string  `json:"type" gorm:"column:type;type:varchar(32);not null"`
	BaseURL       string  `json:"base_url" gorm:"column:base_url;type:varchar(500);not null"`
	AuthType      string  `json:"auth_type" gorm:"column:auth_type;type:varchar(32);not null"`
	Status        string  `json:"status" gorm:"column:status;type:varchar(32);not null;default:active;index"`
	Priority      int     `json:"priority" gorm:"column:priority;not null;default:0"`
	Region        string  `json:"region" gorm:"column:region;type:varchar(100)"`
	QPSLimit      int     `json:"qps_limit" gorm:"column:qps_limit;not null;default:0"`
	MonthlyBudget float64 `json:"monthly_budget" gorm:"column:monthly_budget;type:numeric(18,4);not null;default:0"`
	Owner         string  `json:"owner" gorm:"column:owner;type:varchar(100)"`
	Remark        string  `json:"remark" gorm:"column:remark;type:text"`
	AITimeFields
}

func (AIProvider) TableName() string { return "ai_providers" }

type AIProviderAccount struct {
	ID                string  `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	ProviderID        string  `json:"provider_id" gorm:"column:provider_id;type:uuid;not null;index"`
	AccountName       string  `json:"account_name" gorm:"column:account_name;type:varchar(100);not null"`
	Endpoint          string  `json:"endpoint" gorm:"column:endpoint;type:varchar(500)"`
	KeyAlias          string  `json:"key_alias" gorm:"column:key_alias;type:varchar(200);not null"`
	LoginMethod       string  `json:"login_method" gorm:"column:login_method;type:varchar(32)"`
	LoginAccount      string  `json:"login_account" gorm:"column:login_account;type:varchar(200)"`
	Maintainer        string  `json:"maintainer" gorm:"column:maintainer;type:varchar(100)"`
	MaintainerContact string  `json:"maintainer_contact" gorm:"column:maintainer_contact;type:varchar(100)"`
	EncryptedAPIKey   string  `json:"-" gorm:"column:encrypted_api_key;type:text"`
	EncryptedSecret   string  `json:"-" gorm:"column:encrypted_secret;type:text"`
	QuotaLimit        float64 `json:"quota_limit" gorm:"column:quota_limit;type:numeric(24,4);not null;default:0"`
	UsedQuota         float64 `json:"used_quota" gorm:"column:used_quota;type:numeric(24,4);not null;default:0"`
	Status            string  `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIProviderAccount) TableName() string { return "ai_provider_accounts" }

type AIProviderAPI struct {
	ID              string     `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	ProviderID      string     `json:"provider_id" gorm:"column:provider_id;type:uuid;not null;index"`
	AccountID       string     `json:"account_id" gorm:"column:account_id;type:uuid;not null;index"`
	APIName         string     `json:"api_name" gorm:"column:api_name;type:varchar(150);not null"`
	APIPath         string     `json:"api_path" gorm:"column:api_path;type:varchar(500);not null"`
	APIType         string     `json:"api_type" gorm:"column:api_type;type:varchar(32);not null;index"`
	Capabilities    []string   `json:"capabilities" gorm:"column:capabilities;serializer:json;type:jsonb;not null"`
	AuthType        string     `json:"auth_type" gorm:"column:auth_type;type:varchar(32);not null"`
	QPSLimit        int        `json:"qps_limit" gorm:"column:qps_limit;not null;default:0"`
	TimeoutMS       int        `json:"timeout_ms" gorm:"column:timeout_ms;not null;default:30000"`
	Status          string     `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	LastCalledAt    *time.Time `json:"last_called_at" gorm:"column:last_called_at"`
	HealthStatus    string     `json:"health_status" gorm:"column:health_status;type:varchar(32);not null;default:unknown"`
	HealthMessage   string     `json:"health_message" gorm:"column:health_message;type:varchar(500)"`
	HealthCheckedAt *time.Time `json:"health_checked_at" gorm:"column:health_checked_at"`
	AITimeFields
}

func (AIProviderAPI) TableName() string { return "ai_provider_apis" }

type AICapability struct {
	ID                  string `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	CapabilityCode      string `json:"capability_code" gorm:"column:capability_code;type:varchar(100);not null;uniqueIndex"`
	CapabilityName      string `json:"capability_name" gorm:"column:capability_name;type:varchar(100);not null"`
	ScenarioType        string `json:"scenario_type" gorm:"column:scenario_type;type:varchar(32);not null"`
	ModelType           string `json:"model_type" gorm:"column:model_type;type:varchar(32);not null"`
	DefaultBillingUnit  string `json:"default_billing_unit" gorm:"column:default_billing_unit;type:varchar(32);not null"`
	SupportsTierPricing bool   `json:"supports_tier_pricing" gorm:"column:supports_tier_pricing;not null;default:false"`
	Description         string `json:"description" gorm:"column:description;type:text"`
	Status              string `json:"status" gorm:"column:status;type:varchar(32);not null;default:active;index"`
	SortOrder           int    `json:"sort_order" gorm:"column:sort_order;not null;default:0"`
	AITimeFields
}

func (AICapability) TableName() string { return "ai_capabilities" }

type AIModel struct {
	ID            string   `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	ProviderID    string   `json:"provider_id" gorm:"column:provider_id;type:uuid;not null;index"`
	ModelCode     string   `json:"model_code" gorm:"column:model_code;type:varchar(150);not null"`
	ModelName     string   `json:"model_name" gorm:"column:model_name;type:varchar(150);not null"`
	ModelType     string   `json:"model_type" gorm:"column:model_type;type:varchar(32);not null;index"`
	Capabilities  []string `json:"capabilities" gorm:"column:capabilities;serializer:json;type:jsonb;not null"`
	ContextWindow int      `json:"context_window" gorm:"column:context_window"`
	Unit          string   `json:"unit" gorm:"column:unit;type:varchar(50)"`
	LatencyP95    int      `json:"latency_p95" gorm:"column:latency_p95;not null;default:0"`
	SuccessRate   float64  `json:"success_rate" gorm:"column:success_rate;type:numeric(8,4);not null;default:0"`
	Status        string   `json:"status" gorm:"column:status;type:varchar(32);not null;default:active;index"`
	DefaultFor    []string `json:"default_for" gorm:"column:default_for;serializer:json;type:jsonb;not null"`
	Remark        string   `json:"remark" gorm:"column:remark;type:text"`
	AITimeFields
}

func (AIModel) TableName() string { return "ai_models" }

type AIModelPricePolicy struct {
	ID                 string  `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	ModelID            string  `json:"model_id" gorm:"column:model_id;type:uuid;not null;index"`
	FeatureKey         string  `json:"feature_key" gorm:"column:feature_key;type:varchar(200);not null"`
	FeatureName        string  `json:"feature_name" gorm:"column:feature_name;type:varchar(150);not null"`
	ModelType          string  `json:"model_type" gorm:"column:model_type;type:varchar(32);not null"`
	CapabilityCode     string  `json:"capability_code" gorm:"column:capability_code;type:varchar(100);not null"`
	BillingMode        string  `json:"billing_mode" gorm:"column:billing_mode;type:varchar(32);not null"`
	BillingUnit        string  `json:"billing_unit" gorm:"column:billing_unit;type:varchar(32);not null"`
	PlatformUnit       string  `json:"platform_unit" gorm:"column:platform_unit;type:varchar(32);not null"`
	BaseCostPrice      float64 `json:"base_cost_price" gorm:"column:base_cost_price;type:numeric(18,6);not null;default:0"`
	BaseSalePrice      float64 `json:"base_sale_price" gorm:"column:base_sale_price;type:numeric(18,6);not null;default:0"`
	BasePlatformAmount float64 `json:"base_platform_amount" gorm:"column:base_platform_amount;type:numeric(18,6);not null;default:0"`
	Currency           string  `json:"currency" gorm:"column:currency;type:varchar(16);not null;default:CNY"`
	Status             string  `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIModelPricePolicy) TableName() string { return "ai_model_price_policies" }

type AIModelPriceTier struct {
	ID              string  `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	PricePolicyID   string  `json:"price_policy_id" gorm:"column:price_policy_id;type:uuid;not null;index"`
	TierName        string  `json:"tier_name" gorm:"column:tier_name;type:varchar(150);not null"`
	Mode            string  `json:"mode" gorm:"column:mode;type:varchar(64)"`
	Resolution      string  `json:"resolution" gorm:"column:resolution;type:varchar(64)"`
	Quality         string  `json:"quality" gorm:"column:quality;type:varchar(64)"`
	DurationSeconds int     `json:"duration_seconds" gorm:"column:duration_seconds"`
	AspectRatio     string  `json:"aspect_ratio" gorm:"column:aspect_ratio;type:varchar(64)"`
	CostPrice       float64 `json:"cost_price" gorm:"column:cost_price;type:numeric(18,6);not null;default:0"`
	SalePrice       float64 `json:"sale_price" gorm:"column:sale_price;type:numeric(18,6);not null;default:0"`
	PlatformAmount  float64 `json:"platform_amount" gorm:"column:platform_amount;type:numeric(18,6);not null;default:0"`
	Enabled         bool    `json:"enabled" gorm:"column:enabled;not null;default:true"`
	SortOrder       int     `json:"sort_order" gorm:"column:sort_order;not null;default:0"`
	AITimeFields
}

func (AIModelPriceTier) TableName() string { return "ai_model_price_tiers" }

type AIBaseRoute struct {
	ID             string `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	RouteCode      string `json:"route_code" gorm:"column:route_code;type:varchar(150);not null;uniqueIndex"`
	RouteName      string `json:"route_name" gorm:"column:route_name;type:varchar(150);not null"`
	CapabilityCode string `json:"capability_code" gorm:"column:capability_code;type:varchar(100);not null;index"`
	ModelType      string `json:"model_type" gorm:"column:model_type;type:varchar(32);not null"`
	Strategy       string `json:"strategy" gorm:"column:strategy;type:varchar(32);not null"`
	TimeoutMS      int    `json:"timeout_ms" gorm:"column:timeout_ms;not null;default:30000"`
	MaxRetry       int    `json:"max_retry" gorm:"column:max_retry;not null;default:0"`
	Description    string `json:"description" gorm:"column:description;type:text"`
	Status         string `json:"status" gorm:"column:status;type:varchar(32);not null;default:active;index"`
	AITimeFields
}

func (AIBaseRoute) TableName() string { return "ai_base_routes" }

type AIBaseRouteModel struct {
	ID                string `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	BaseRouteID       string `json:"base_route_id" gorm:"column:base_route_id;type:uuid;not null;index"`
	ModelID           string `json:"model_id" gorm:"column:model_id;type:uuid;not null;index"`
	ProviderAccountID string `json:"provider_account_id" gorm:"column:provider_account_id;type:uuid;index"`
	ProviderAPIID     string `json:"provider_api_id" gorm:"column:provider_api_id;type:uuid;index"`
	Role              string `json:"role" gorm:"column:role;type:varchar(32);not null"`
	Priority          int    `json:"priority" gorm:"column:priority;not null;default:1"`
	Weight            int    `json:"weight" gorm:"column:weight;not null;default:100"`
	MaxRetry          int    `json:"max_retry" gorm:"column:max_retry;not null;default:0"`
	TimeoutMS         int    `json:"timeout_ms" gorm:"column:timeout_ms;not null;default:30000"`
	Status            string `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIBaseRouteModel) TableName() string { return "ai_base_route_models" }

type AIScenario struct {
	ID                 string `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	AppCode            string `json:"app_code" gorm:"column:app_code;type:varchar(100);not null;index"`
	AppName            string `json:"app_name" gorm:"column:app_name;type:varchar(100);not null"`
	AIScenarioCode     string `json:"ai_scenario_code" gorm:"column:ai_scenario_code;type:varchar(150);not null;index"`
	AIScenarioName     string `json:"ai_scenario_name" gorm:"column:ai_scenario_name;type:varchar(150);not null"`
	ScenarioType       string `json:"scenario_type" gorm:"column:scenario_type;type:varchar(32);not null"`
	CapabilityCode     string `json:"capability_code" gorm:"column:capability_code;type:varchar(100);not null"`
	ModelType          string `json:"model_type" gorm:"column:model_type;type:varchar(32);not null"`
	DefaultBaseRouteID string `json:"default_base_route_id" gorm:"column:default_base_route_id;type:uuid;not null;index"`
	Owner              string `json:"owner" gorm:"column:owner;type:varchar(100)"`
	Description        string `json:"description" gorm:"column:description;type:text"`
	Version            string `json:"version" gorm:"column:version;type:varchar(50)"`
	Status             string `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIScenario) TableName() string { return "ai_scenarios" }

type AITenantStrategyPolicy struct {
	ID                  string   `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	PolicyName          string   `json:"policy_name" gorm:"column:policy_name;type:varchar(150);not null"`
	TenantScope         string   `json:"tenant_scope" gorm:"column:tenant_scope;type:varchar(32);not null"`
	TenantIDs           []string `json:"tenant_ids" gorm:"column:tenant_ids;serializer:json;type:jsonb;not null"`
	AppCode             string   `json:"app_code" gorm:"column:app_code;type:varchar(100);not null;index"`
	AppName             string   `json:"app_name" gorm:"column:app_name;type:varchar(100);not null"`
	AIScenarioCode      string   `json:"ai_scenario_code" gorm:"column:ai_scenario_code;type:varchar(150);not null;index"`
	AIScenarioName      string   `json:"ai_scenario_name" gorm:"column:ai_scenario_name;type:varchar(150);not null"`
	DefaultBaseRouteID  string   `json:"default_base_route_id" gorm:"column:default_base_route_id;type:uuid;not null"`
	OverrideBaseRouteID string   `json:"override_base_route_id" gorm:"column:override_base_route_id;type:uuid"`
	Description         string   `json:"description" gorm:"column:description;type:text"`
	Status              string   `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AITenantStrategyPolicy) TableName() string { return "ai_tenant_strategy_policies" }

type AIStrategyQuotaRule struct {
	ID               string  `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	PolicyID         string  `json:"policy_id" gorm:"column:policy_id;type:uuid;not null;index"`
	Dimension        string  `json:"dimension" gorm:"column:dimension;type:varchar(32);not null"`
	SubjectCode      string  `json:"subject_code" gorm:"column:subject_code;type:varchar(150);not null"`
	UsageUnit        string  `json:"usage_unit" gorm:"column:usage_unit;type:varchar(32);not null"`
	Period           string  `json:"period" gorm:"column:period;type:varchar(32);not null"`
	QuotaLimit       float64 `json:"quota_limit" gorm:"column:quota_limit;type:numeric(24,6);not null;default:0"`
	UsedAmount       float64 `json:"used_amount" gorm:"column:used_amount;type:numeric(24,6);not null;default:0"`
	WarningThreshold float64 `json:"warning_threshold" gorm:"column:warning_threshold;type:numeric(8,4);not null;default:80"`
	OverLimitAction  string  `json:"over_limit_action" gorm:"column:over_limit_action;type:varchar(32);not null;default:alert_only"`
	Status           string  `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIStrategyQuotaRule) TableName() string { return "ai_strategy_quota_rules" }

type AIStrategyRateLimitRule struct {
	ID              string `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	PolicyID        string `json:"policy_id" gorm:"column:policy_id;type:uuid;not null;index"`
	Dimension       string `json:"dimension" gorm:"column:dimension;type:varchar(32);not null"`
	SubjectCode     string `json:"subject_code" gorm:"column:subject_code;type:varchar(150);not null"`
	QPS             int    `json:"qps" gorm:"column:qps;not null;default:0"`
	Concurrency     int    `json:"concurrency" gorm:"column:concurrency;not null;default:0"`
	MinuteLimit     int    `json:"minute_limit" gorm:"column:minute_limit;not null;default:0"`
	HourLimit       int    `json:"hour_limit" gorm:"column:hour_limit;not null;default:0"`
	DayLimit        int    `json:"day_limit" gorm:"column:day_limit;not null;default:0"`
	OverLimitAction string `json:"over_limit_action" gorm:"column:over_limit_action;type:varchar(32);not null;default:queue"`
	Status          string `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIStrategyRateLimitRule) TableName() string { return "ai_strategy_rate_limit_rules" }

type AIUsageRecord struct {
	ID                 string     `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	RequestID          string     `json:"request_id" gorm:"column:request_id;type:varchar(100);not null;uniqueIndex"`
	TenantID           string     `json:"tenant_id" gorm:"column:tenant_id;type:varchar(100);not null;index"`
	TenantName         string     `json:"tenant_name" gorm:"column:tenant_name;type:varchar(150)"`
	AppCode            string     `json:"app_code" gorm:"column:app_code;type:varchar(100);not null;index"`
	AppName            string     `json:"app_name" gorm:"column:app_name;type:varchar(100)"`
	AIScenarioCode     string     `json:"ai_scenario_code" gorm:"column:ai_scenario_code;type:varchar(150);not null;index"`
	AIScenarioName     string     `json:"ai_scenario_name" gorm:"column:ai_scenario_name;type:varchar(150)"`
	UserID             string     `json:"user_id" gorm:"column:user_id;type:varchar(100)"`
	UserName           string     `json:"user_name" gorm:"column:user_name;type:varchar(100)"`
	ProviderID         string     `json:"provider_id" gorm:"column:provider_id;type:uuid"`
	ProviderAccountID  string     `json:"provider_account_id" gorm:"column:provider_account_id;type:uuid"`
	ProviderAPIID      string     `json:"provider_api_id" gorm:"column:provider_api_id;type:uuid"`
	ModelID            string     `json:"model_id" gorm:"column:model_id;type:uuid;index"`
	BaseRouteID        string     `json:"base_route_id" gorm:"column:base_route_id;type:uuid"`
	TenantStrategyID   string     `json:"tenant_strategy_id" gorm:"column:tenant_strategy_id;type:uuid"`
	PricePolicyID      string     `json:"price_policy_id" gorm:"column:price_policy_id;type:uuid"`
	PriceTierID        string     `json:"price_tier_id" gorm:"column:price_tier_id;type:uuid"`
	UsageAmount        float64    `json:"usage_amount" gorm:"column:usage_amount;type:numeric(24,6);not null;default:0"`
	UsageUnit          string     `json:"usage_unit" gorm:"column:usage_unit;type:varchar(32);not null"`
	UsageDetail        string     `json:"usage_detail" gorm:"column:usage_detail;type:varchar(200)"`
	Calls              int        `json:"calls" gorm:"column:calls;not null;default:1"`
	CostAmount         float64    `json:"cost_amount" gorm:"column:cost_amount;type:numeric(18,6);not null;default:0"`
	BillingAmount      float64    `json:"billing_amount" gorm:"column:billing_amount;type:numeric(18,6);not null;default:0"`
	PlatformUnit       string     `json:"platform_unit" gorm:"column:platform_unit;type:varchar(32)"`
	PlatformAmount     float64    `json:"platform_amount" gorm:"column:platform_amount;type:numeric(18,6);not null;default:0"`
	LatencyMS          int        `json:"latency_ms" gorm:"column:latency_ms;not null;default:0"`
	ProviderHTTPStatus int        `json:"provider_http_status" gorm:"column:provider_http_status;not null;default:0"`
	ProviderRequestID  string     `json:"provider_request_id" gorm:"column:provider_request_id;type:varchar(200)"`
	StartedAt          *time.Time `json:"started_at" gorm:"column:started_at"`
	FinishedAt         *time.Time `json:"finished_at" gorm:"column:finished_at"`
	RetryCount         int        `json:"retry_count" gorm:"column:retry_count;not null;default:0"`
	Status             string     `json:"status" gorm:"column:status;type:varchar(32);not null;index"`
	ErrorCode          string     `json:"error_code" gorm:"column:error_code;type:varchar(100)"`
	ErrorMessage       string     `json:"error_message" gorm:"column:error_message;type:text"`
	RequestParams      string     `json:"request_params" gorm:"column:request_params;type:jsonb;not null;default:'{}'"`
	DataSource         string     `json:"data_source" gorm:"column:data_source;type:varchar(32);not null;default:gateway"`
	IsDemo             bool       `json:"is_demo" gorm:"column:is_demo;not null;default:false;index"`
	PromptHash         string     `json:"prompt_hash" gorm:"column:prompt_hash;type:varchar(128)"`
	ResponseHash       string     `json:"response_hash" gorm:"column:response_hash;type:varchar(128)"`
	CalledAt           time.Time  `json:"called_at" gorm:"column:called_at;not null"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at;not null"`
}

func (AIUsageRecord) TableName() string { return "ai_usage_records" }

type AIGatewaySetting struct {
	ID           string `json:"id" gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	SettingKey   string `json:"setting_key" gorm:"column:setting_key;type:varchar(100);not null;uniqueIndex"`
	SettingValue string `json:"setting_value" gorm:"column:setting_value;type:jsonb;not null;default:'{}'"`
	Description  string `json:"description" gorm:"column:description;type:text"`
	Status       string `json:"status" gorm:"column:status;type:varchar(32);not null;default:active"`
	AITimeFields
}

func (AIGatewaySetting) TableName() string { return "ai_gateway_settings" }
