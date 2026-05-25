package models

import "time"

type AiGeoBrandCard struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_brand_tenant_code"`
	CompanyID      *uint64    `gorm:"column:company_id;index"`
	DepartmentID   *uint64    `gorm:"column:department_id;index"`
	BrandCode      string     `gorm:"column:brand_code;type:varchar(80);not null;index:idx_ai_geo_brand_tenant_code"`
	BrandName      string     `gorm:"column:brand_name;type:varchar(160);not null"`
	Positioning    *string    `gorm:"column:positioning;type:text"`
	TargetAudience *string    `gorm:"column:target_audience;type:text"`
	PriceBand      *string    `gorm:"column:price_band;type:varchar(120)"`
	Tone           *string    `gorm:"column:tone;type:varchar(160)"`
	Keywords       string     `gorm:"column:keywords;type:jsonb;not null;default:'[]'"`
	Completeness   int        `gorm:"column:completeness;not null;default:0"`
	Status         string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy      *uint64    `gorm:"column:created_by"`
	UpdatedBy      *uint64    `gorm:"column:updated_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoBrandCard) TableName() string { return "ai_geo_brand_cards" }

type AiGeoProductCard struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID      uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_product_tenant_code"`
	CompanyID     *uint64    `gorm:"column:company_id;index"`
	DepartmentID  *uint64    `gorm:"column:department_id;index"`
	BrandID       uint64     `gorm:"column:brand_id;not null;index"`
	ProductCode   string     `gorm:"column:product_code;type:varchar(100);not null;index:idx_ai_geo_product_tenant_code"`
	ProductName   string     `gorm:"column:product_name;type:varchar(180);not null"`
	CategoryName  *string    `gorm:"column:category_name;type:varchar(160)"`
	SellingPoints string     `gorm:"column:selling_points;type:jsonb;not null;default:'[]'"`
	FAQ           string     `gorm:"column:faq;type:jsonb;not null;default:'[]'"`
	ContentAngles string     `gorm:"column:content_angles;type:jsonb;not null;default:'[]'"`
	Completeness  int        `gorm:"column:completeness;not null;default:0"`
	Status        string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	UpdatedBy     *uint64    `gorm:"column:updated_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoProductCard) TableName() string { return "ai_geo_product_cards" }

type AiGeoSKU struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_sku_tenant_code"`
	ProductID   uint64     `gorm:"column:product_id;not null;index"`
	SKUCode     string     `gorm:"column:sku_code;type:varchar(100);not null;index:idx_ai_geo_sku_tenant_code"`
	SKUName     string     `gorm:"column:sku_name;type:varchar(180);not null"`
	Attributes  string     `gorm:"column:attributes;type:jsonb;not null;default:'{}'"`
	Price       float64    `gorm:"column:price;type:numeric(18,2);not null;default:0"`
	ImageURL    *string    `gorm:"column:image_url;type:text"`
	StockStatus string     `gorm:"column:stock_status;type:varchar(32);not null;default:'unknown';index"`
	Status      string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy   *uint64    `gorm:"column:created_by"`
	UpdatedBy   *uint64    `gorm:"column:updated_by"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoSKU) TableName() string { return "ai_geo_skus" }

type AiGeoCompetitor struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    uint64     `gorm:"column:tenant_id;not null;index"`
	ProductID   uint64     `gorm:"column:product_id;not null;index"`
	BrandName   string     `gorm:"column:brand_name;type:varchar(160);not null"`
	ProductName string     `gorm:"column:product_name;type:varchar(180);not null"`
	PriceText   *string    `gorm:"column:price_text;type:varchar(120)"`
	Point       *string    `gorm:"column:point;type:text"`
	Difference  *string    `gorm:"column:difference;type:text"`
	Angle       *string    `gorm:"column:angle;type:text"`
	LinkURL     *string    `gorm:"column:link_url;type:text"`
	Status      string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy   *uint64    `gorm:"column:created_by"`
	UpdatedBy   *uint64    `gorm:"column:updated_by"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoCompetitor) TableName() string { return "ai_geo_competitors" }

type AiGeoKeyword struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID     uint64     `gorm:"column:tenant_id;not null;index"`
	BrandID      *uint64    `gorm:"column:brand_id;index"`
	ProductID    *uint64    `gorm:"column:product_id;index"`
	KeywordGroup string     `gorm:"column:keyword_group;type:varchar(120);not null;default:'通用关键词'"`
	Keyword      string     `gorm:"column:keyword;type:varchar(160);not null"`
	Intent       *string    `gorm:"column:intent;type:varchar(80)"`
	Source       string     `gorm:"column:source;type:varchar(80);not null;default:'manual'"`
	Weight       int        `gorm:"column:weight;not null;default:0"`
	Status       string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy    *uint64    `gorm:"column:created_by"`
	UpdatedBy    *uint64    `gorm:"column:updated_by"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoKeyword) TableName() string { return "ai_geo_keywords" }

type AiGeoChannelProfile struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID           uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_channel_tenant_code"`
	CompanyID          *uint64    `gorm:"column:company_id;index"`
	DepartmentID       *uint64    `gorm:"column:department_id;index"`
	ChannelCode        string     `gorm:"column:channel_code;type:varchar(80);not null;index:idx_ai_geo_channel_tenant_code"`
	ChannelName        string     `gorm:"column:channel_name;type:varchar(120);not null"`
	ChannelType        string     `gorm:"column:channel_type;type:varchar(80);not null;default:'content'"`
	EntryURL           *string    `gorm:"column:entry_url;type:text"`
	ContentForms       string     `gorm:"column:content_forms;type:jsonb;not null;default:'[]'"`
	SupportModes       string     `gorm:"column:support_modes;type:jsonb;not null;default:'[]'"`
	DefaultPublishMode string     `gorm:"column:default_publish_mode;type:varchar(80);not null;default:'manual'"`
	Status             string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy          *uint64    `gorm:"column:created_by"`
	UpdatedBy          *uint64    `gorm:"column:updated_by"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt          *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoChannelProfile) TableName() string { return "ai_geo_channel_profiles" }

type AiGeoChannelAccount struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID          uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_account_tenant_name"`
	ChannelID         uint64     `gorm:"column:channel_id;not null;index"`
	AccountName       string     `gorm:"column:account_name;type:varchar(160);not null;index:idx_ai_geo_account_tenant_name"`
	ExternalAccountID *string    `gorm:"column:external_account_id;type:varchar(160)"`
	AuthStatus        string     `gorm:"column:auth_status;type:varchar(32);not null;default:'not_authorized';index"`
	PublishStatus     string     `gorm:"column:publish_status;type:varchar(32);not null;default:'unavailable';index"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy         *uint64    `gorm:"column:created_by"`
	UpdatedBy         *uint64    `gorm:"column:updated_by"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoChannelAccount) TableName() string { return "ai_geo_channel_accounts" }

type AiGeoDraft struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_draft_tenant_code"`
	CompanyID      *uint64    `gorm:"column:company_id;index"`
	DepartmentID   *uint64    `gorm:"column:department_id;index"`
	DraftCode      string     `gorm:"column:draft_code;type:varchar(100);not null;index:idx_ai_geo_draft_tenant_code"`
	BrandID        *uint64    `gorm:"column:brand_id;index"`
	ProductID      *uint64    `gorm:"column:product_id;index"`
	Title          string     `gorm:"column:title;type:varchar(240);not null"`
	Summary        *string    `gorm:"column:summary;type:text"`
	Body           string     `gorm:"column:body;type:text;not null"`
	Keywords       string     `gorm:"column:keywords;type:jsonb;not null;default:'[]'"`
	Conversation   string     `gorm:"column:conversation;type:jsonb;not null;default:'[]'"`
	SourceSnapshot string     `gorm:"column:source_snapshot;type:jsonb;not null;default:'{}'"`
	Source         string     `gorm:"column:source;type:varchar(80);not null;default:'manual'"`
	AuditStatus    string     `gorm:"column:audit_status;type:varchar(32);not null;default:'approved';index"`
	ChannelStatus  string     `gorm:"column:channel_status;type:varchar(32);not null;default:'not_generated';index"`
	Status         string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy      *uint64    `gorm:"column:created_by"`
	UpdatedBy      *uint64    `gorm:"column:updated_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoDraft) TableName() string { return "ai_geo_drafts" }

type AiGeoChannelContent struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID      uint64     `gorm:"column:tenant_id;not null;index"`
	DraftID       uint64     `gorm:"column:draft_id;not null;index"`
	ChannelID     uint64     `gorm:"column:channel_id;not null;index"`
	Title         string     `gorm:"column:title;type:varchar(240);not null"`
	Body          string     `gorm:"column:body;type:text;not null"`
	AuditStatus   string     `gorm:"column:audit_status;type:varchar(32);not null;default:'approved';index"`
	PublishStatus string     `gorm:"column:publish_status;type:varchar(32);not null;default:'not_planned';index"`
	Status        string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	UpdatedBy     *uint64    `gorm:"column:updated_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoChannelContent) TableName() string { return "ai_geo_channel_contents" }

type AiGeoPublishPlan struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_plan_tenant_code"`
	PlanCode         string     `gorm:"column:plan_code;type:varchar(100);not null;index:idx_ai_geo_plan_tenant_code"`
	ChannelContentID uint64     `gorm:"column:channel_content_id;not null;index"`
	ChannelID        uint64     `gorm:"column:channel_id;not null;index"`
	ScheduledAt      time.Time  `gorm:"column:scheduled_at;not null;index"`
	PublishMethod    string     `gorm:"column:publish_method;type:varchar(80);not null;default:'manual'"`
	AutomationLevel  string     `gorm:"column:automation_level;type:varchar(80);not null;default:'manual'"`
	Status           string     `gorm:"column:status;type:varchar(32);not null;default:'scheduled';index"`
	PublishedURL     *string    `gorm:"column:published_url;type:text"`
	FailReason       *string    `gorm:"column:fail_reason;type:text"`
	CreatedBy        *uint64    `gorm:"column:created_by"`
	UpdatedBy        *uint64    `gorm:"column:updated_by"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoPublishPlan) TableName() string { return "ai_geo_publish_plans" }

type AiGeoImportBatch struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID      uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_import_tenant_code"`
	BatchCode     string     `gorm:"column:batch_code;type:varchar(100);not null;index:idx_ai_geo_import_tenant_code"`
	ImportType    string     `gorm:"column:import_type;type:varchar(80);not null;index"`
	MappingConfig string     `gorm:"column:mapping_config;type:jsonb;not null;default:'{}'"`
	RecordCount   int64      `gorm:"column:record_count;not null;default:0"`
	SuccessCount  int64      `gorm:"column:success_count;not null;default:0"`
	FailedCount   int64      `gorm:"column:failed_count;not null;default:0"`
	Status        string     `gorm:"column:status;type:varchar(32);not null;default:'pending';index"`
	ErrorMessage  *string    `gorm:"column:error_message;type:text"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	UpdatedBy     *uint64    `gorm:"column:updated_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoImportBatch) TableName() string { return "ai_geo_import_batches" }

type AiGeoImportError struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID     uint64     `gorm:"column:tenant_id;not null;index"`
	BatchID      uint64     `gorm:"column:batch_id;not null;index"`
	RowNumber    int        `gorm:"column:row_number;not null"`
	FieldName    *string    `gorm:"column:field_name;type:varchar(120)"`
	ErrorCode    string     `gorm:"column:error_code;type:varchar(80);not null"`
	ErrorMessage string     `gorm:"column:error_message;type:text;not null"`
	RawData      string     `gorm:"column:raw_data;type:jsonb;not null;default:'{}'"`
	Status       string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoImportError) TableName() string { return "ai_geo_import_errors" }

type AiGeoAuditSuggestion struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index"`
	ObjectType     string     `gorm:"column:object_type;type:varchar(40);not null;index"`
	ObjectID       uint64     `gorm:"column:object_id;not null;index"`
	ObjectCode     *string    `gorm:"column:object_code;type:varchar(120)"`
	ScenarioCode   string     `gorm:"column:scenario_code;type:varchar(120);not null"`
	RiskLevel      string     `gorm:"column:risk_level;type:varchar(32);not null;default:'low'"`
	Passed         bool       `gorm:"column:passed;not null;default:false"`
	Summary        *string    `gorm:"column:summary;type:text"`
	SuggestionJSON string     `gorm:"column:suggestion_json;type:jsonb;not null;default:'[]'"`
	ModelCode      *string    `gorm:"column:model_code;type:varchar(120)"`
	Status         string     `gorm:"column:status;type:varchar(32);not null;default:'success';index"`
	ErrorMessage   *string    `gorm:"column:error_message;type:text"`
	GeneratedAt    time.Time  `gorm:"column:generated_at;not null"`
	CreatedBy      *uint64    `gorm:"column:created_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoAuditSuggestion) TableName() string { return "ai_geo_audit_suggestions" }

type AiGeoMaterialAsset struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  uint64     `gorm:"column:tenant_id;not null;index"`
	BrandID   *uint64    `gorm:"column:brand_id;index"`
	ProductID *uint64    `gorm:"column:product_id;index"`
	AssetType string     `gorm:"column:asset_type;type:varchar(80);not null;index"`
	AssetName string     `gorm:"column:asset_name;type:varchar(180);not null"`
	FileID    *uint64    `gorm:"column:file_id;index"`
	URL       *string    `gorm:"column:url;type:text"`
	Metadata  string     `gorm:"column:metadata;type:jsonb;not null;default:'{}'"`
	Status    string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy *uint64    `gorm:"column:created_by"`
	UpdatedBy *uint64    `gorm:"column:updated_by"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoMaterialAsset) TableName() string { return "ai_geo_material_assets" }

type AiGeoHotspot struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   uint64     `gorm:"column:tenant_id;not null;index"`
	SourceID   *uint64    `gorm:"column:source_id;index"`
	Platform   string     `gorm:"column:platform;type:varchar(80);not null;index"`
	Title      string     `gorm:"column:title;type:varchar(240);not null"`
	HeatScore  int        `gorm:"column:heat_score;not null;default:0"`
	SourceURL  *string    `gorm:"column:source_url;type:text"`
	CapturedAt time.Time  `gorm:"column:captured_at;not null;index"`
	Metadata   string     `gorm:"column:metadata;type:jsonb;not null;default:'{}'"`
	Status     string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy  *uint64    `gorm:"column:created_by"`
	UpdatedBy  *uint64    `gorm:"column:updated_by"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoHotspot) TableName() string { return "ai_geo_hotspots" }

type AiGeoExternalSource struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         uint64     `gorm:"column:tenant_id;not null;index"`
	SourceType       string     `gorm:"column:source_type;type:varchar(40);not null;default:'url';index"`
	SourceURL        string     `gorm:"column:source_url;type:text;not null"`
	SourceSite       *string    `gorm:"column:source_site;type:varchar(160)"`
	SourceTitle      *string    `gorm:"column:source_title;type:varchar(240)"`
	RawText          string     `gorm:"column:raw_text;type:text;not null;default:''"`
	CleanText        string     `gorm:"column:clean_text;type:text;not null;default:''"`
	ContentHash      string     `gorm:"column:content_hash;type:varchar(64);not null;index"`
	ExtractedMeta    string     `gorm:"column:extracted_meta;type:jsonb;not null;default:'{}'"`
	ExtractionStatus string     `gorm:"column:extraction_status;type:varchar(32);not null;default:'success';index"`
	ExtractionError  *string    `gorm:"column:extraction_error;type:text"`
	CapturedAt       time.Time  `gorm:"column:captured_at;not null;index"`
	Status           string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy        *uint64    `gorm:"column:created_by"`
	UpdatedBy        *uint64    `gorm:"column:updated_by"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoExternalSource) TableName() string { return "ai_geo_external_sources" }

type AiGeoStyleTemplate struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID          uint64     `gorm:"column:tenant_id;not null;index;index:idx_ai_geo_style_template_tenant_code"`
	SourceID          *uint64    `gorm:"column:source_id;index"`
	TemplateCode      string     `gorm:"column:template_code;type:varchar(100);not null;index:idx_ai_geo_style_template_tenant_code"`
	TemplateName      string     `gorm:"column:template_name;type:varchar(160);not null"`
	Description       *string    `gorm:"column:description;type:text"`
	ContentType       *string    `gorm:"column:content_type;type:varchar(80);index"`
	Platform          *string    `gorm:"column:platform;type:varchar(80);index"`
	ToneProfile       string     `gorm:"column:tone_profile;type:jsonb;not null;default:'{}'"`
	StructureProfile  string     `gorm:"column:structure_profile;type:jsonb;not null;default:'{}'"`
	TechniqueProfile  string     `gorm:"column:technique_profile;type:jsonb;not null;default:'{}'"`
	StyleKeywords     string     `gorm:"column:style_keywords;type:jsonb;not null;default:'[]'"`
	PromptFragment    string     `gorm:"column:prompt_fragment;type:text;not null;default:''"`
	NegativeRules     string     `gorm:"column:negative_rules;type:jsonb;not null;default:'[]'"`
	ExtractionSummary string     `gorm:"column:extraction_summary;type:jsonb;not null;default:'{}'"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'active';index"`
	CreatedBy         *uint64    `gorm:"column:created_by"`
	UpdatedBy         *uint64    `gorm:"column:updated_by"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (AiGeoStyleTemplate) TableName() string { return "ai_geo_style_templates" }
