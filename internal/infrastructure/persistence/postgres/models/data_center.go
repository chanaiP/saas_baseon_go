package models

import "time"

type DataCenterRawDataBatch struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID       uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_raw_batch_tenant_code,unique"`
	CompanyID      *uint64    `gorm:"column:company_id;index"`
	DepartmentID   *uint64    `gorm:"column:department_id;index"`
	BatchCode      string     `gorm:"column:batch_code;type:varchar(80);not null;index:idx_dc_raw_batch_tenant_code,unique"`
	DataType       string     `gorm:"column:data_type;type:varchar(50);not null;index"`
	PlatformCode   *string    `gorm:"column:platform_code;type:varchar(80);index"`
	AppCode        *string    `gorm:"column:app_code;type:varchar(80)"`
	ConnectionCode *string    `gorm:"column:connection_code;type:varchar(120)"`
	RecordCount    int64      `gorm:"column:record_count;not null;default:0"`
	SuccessCount   int64      `gorm:"column:success_count;not null;default:0"`
	FailedCount    int64      `gorm:"column:failed_count;not null;default:0"`
	SyncTime       *time.Time `gorm:"column:sync_time;index"`
	Status         string     `gorm:"column:status;type:varchar(32);not null;default:'pending';index"`
	ErrorMessage   *string    `gorm:"column:error_message;type:text"`
	SourceParams   string     `gorm:"column:source_params;type:jsonb;not null;default:'{}'"`
	SamplePayload  string     `gorm:"column:sample_payload;type:jsonb;not null;default:'{}'"`
	CreatedBy      *uint64    `gorm:"column:created_by"`
	UpdatedBy      *uint64    `gorm:"column:updated_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterRawDataBatch) TableName() string { return "data_center_raw_data_batches" }

type DataCenterRawDataError struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    uint64    `gorm:"column:tenant_id;not null;index"`
	BatchID     uint64    `gorm:"column:batch_id;not null;index"`
	BatchCode   string    `gorm:"column:batch_code;type:varchar(80);not null;index"`
	RowNumber   *int64    `gorm:"column:row_number"`
	ErrorCode   string    `gorm:"column:error_code;type:varchar(80);not null"`
	ErrorReason string    `gorm:"column:error_reason;type:text;not null"`
	RawPayload  string    `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
}

func (DataCenterRawDataError) TableName() string { return "data_center_raw_data_errors" }

type DataCenterStdSalesOrder struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_sales_order_tenant_code,unique"`
	CompanyID       *uint64    `gorm:"column:company_id;index"`
	DepartmentID    *uint64    `gorm:"column:department_id;index"`
	OrderCode       string     `gorm:"column:order_code;type:varchar(128);not null;index:idx_dc_sales_order_tenant_code,unique"`
	BrandCode       *string    `gorm:"column:brand_code;type:varchar(80);index"`
	BrandName       *string    `gorm:"column:brand_name;type:varchar(120)"`
	ChannelCode     *string    `gorm:"column:channel_code;type:varchar(80);index"`
	PlatformCode    *string    `gorm:"column:platform_code;type:varchar(80);index"`
	ResourceCode    *string    `gorm:"column:resource_code;type:varchar(100);index"`
	ResourceName    *string    `gorm:"column:resource_name;type:varchar(160)"`
	ProductCode     *string    `gorm:"column:product_code;type:varchar(100);index"`
	ProductName     *string    `gorm:"column:product_name;type:varchar(180)"`
	SKUCode         *string    `gorm:"column:sku_code;type:varchar(100);index"`
	SalesAmount     float64    `gorm:"column:sales_amount;type:numeric(18,2);not null;default:0"`
	PaidAmount      float64    `gorm:"column:paid_amount;type:numeric(18,2);not null;default:0"`
	RefundAmount    float64    `gorm:"column:refund_amount;type:numeric(18,2);not null;default:0"`
	OrderStatus     *string    `gorm:"column:order_status;type:varchar(32);index"`
	OrderTime       time.Time  `gorm:"column:order_time;not null;index"`
	SourceBatchCode *string    `gorm:"column:source_batch_code;type:varchar(80);index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterStdSalesOrder) TableName() string { return "data_center_std_sales_orders" }

type DataCenterStdAdDaily struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_ad_daily_object_date,unique"`
	CompanyID       *uint64    `gorm:"column:company_id;index"`
	DepartmentID    *uint64    `gorm:"column:department_id;index"`
	BrandCode       *string    `gorm:"column:brand_code;type:varchar(80);index"`
	BrandName       *string    `gorm:"column:brand_name;type:varchar(120)"`
	PlatformCode    *string    `gorm:"column:platform_code;type:varchar(80);index"`
	AccountCode     string     `gorm:"column:account_code;type:varchar(100);not null;index:idx_dc_ad_daily_object_date,unique"`
	AccountName     *string    `gorm:"column:account_name;type:varchar(160)"`
	CampaignCode    string     `gorm:"column:campaign_code;type:varchar(100);not null;index:idx_dc_ad_daily_object_date,unique"`
	CampaignName    *string    `gorm:"column:campaign_name;type:varchar(180)"`
	CreativeCode    *string    `gorm:"column:creative_code;type:varchar(100);index"`
	ProductCode     *string    `gorm:"column:product_code;type:varchar(100);index"`
	StatDate        time.Time  `gorm:"column:stat_date;type:date;not null;index;index:idx_dc_ad_daily_object_date,unique"`
	CostAmount      float64    `gorm:"column:cost_amount;type:numeric(18,2);not null;default:0"`
	ImpressionCount int64      `gorm:"column:impression_count;not null;default:0"`
	ClickCount      int64      `gorm:"column:click_count;not null;default:0"`
	OrderAmount     float64    `gorm:"column:order_amount;type:numeric(18,2);not null;default:0"`
	OrderCount      int64      `gorm:"column:order_count;not null;default:0"`
	ROI             float64    `gorm:"column:roi;type:numeric(18,4);not null;default:0"`
	ClickRate       float64    `gorm:"column:click_rate;type:numeric(18,4);not null;default:0"`
	ConversionRate  float64    `gorm:"column:conversion_rate;type:numeric(18,4);not null;default:0"`
	SourceBatchCode *string    `gorm:"column:source_batch_code;type:varchar(80);index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterStdAdDaily) TableName() string { return "data_center_std_ad_daily" }

type DataCenterStdInventoryDaily struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_inventory_object_date,unique"`
	CompanyID       *uint64    `gorm:"column:company_id;index"`
	DepartmentID    *uint64    `gorm:"column:department_id;index"`
	BrandCode       *string    `gorm:"column:brand_code;type:varchar(80);index"`
	ProductCode     string     `gorm:"column:product_code;type:varchar(100);not null;index:idx_dc_inventory_object_date,unique"`
	ProductName     *string    `gorm:"column:product_name;type:varchar(180)"`
	SKUCode         string     `gorm:"column:sku_code;type:varchar(100);not null;index:idx_dc_inventory_object_date,unique"`
	WarehouseCode   *string    `gorm:"column:warehouse_code;type:varchar(100);index"`
	StoreCode       string     `gorm:"column:store_code;type:varchar(100);not null;default:'';index:idx_dc_inventory_object_date,unique"`
	StatDate        time.Time  `gorm:"column:stat_date;type:date;not null;index;index:idx_dc_inventory_object_date,unique"`
	AvailableStock  int64      `gorm:"column:available_stock;not null;default:0"`
	InTransitStock  int64      `gorm:"column:in_transit_stock;not null;default:0"`
	Sales7D         int64      `gorm:"column:sales_7d;not null;default:0"`
	AvailableDays   float64    `gorm:"column:available_days;type:numeric(18,2);not null;default:0"`
	InventoryStatus *string    `gorm:"column:inventory_status;type:varchar(32);index"`
	SourceBatchCode *string    `gorm:"column:source_batch_code;type:varchar(80);index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterStdInventoryDaily) TableName() string { return "data_center_std_inventory_daily" }

type DataCenterStdRefundOrder struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_refund_tenant_code,unique"`
	CompanyID       *uint64    `gorm:"column:company_id;index"`
	DepartmentID    *uint64    `gorm:"column:department_id;index"`
	RefundCode      string     `gorm:"column:refund_code;type:varchar(128);not null;index:idx_dc_refund_tenant_code,unique"`
	OrderCode       *string    `gorm:"column:order_code;type:varchar(128);index"`
	BrandCode       *string    `gorm:"column:brand_code;type:varchar(80);index"`
	BrandName       *string    `gorm:"column:brand_name;type:varchar(120)"`
	ChannelCode     *string    `gorm:"column:channel_code;type:varchar(80);index"`
	PlatformCode    *string    `gorm:"column:platform_code;type:varchar(80);index"`
	ProductCode     *string    `gorm:"column:product_code;type:varchar(100);index"`
	ProductName     *string    `gorm:"column:product_name;type:varchar(180)"`
	SKUCode         *string    `gorm:"column:sku_code;type:varchar(100);index"`
	RefundAmount    float64    `gorm:"column:refund_amount;type:numeric(18,2);not null;default:0"`
	RefundReason    *string    `gorm:"column:refund_reason;type:varchar(255)"`
	RefundStatus    *string    `gorm:"column:refund_status;type:varchar(32);index"`
	RefundTime      time.Time  `gorm:"column:refund_time;not null;index"`
	SourceBatchCode *string    `gorm:"column:source_batch_code;type:varchar(80);index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterStdRefundOrder) TableName() string { return "data_center_std_refund_orders" }

type DataCenterStdProduct struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_product_tenant_code,unique"`
	CompanyID       *uint64    `gorm:"column:company_id;index"`
	DepartmentID    *uint64    `gorm:"column:department_id;index"`
	ProductCode     string     `gorm:"column:product_code;type:varchar(100);not null;index:idx_dc_product_tenant_code,unique"`
	ProductName     string     `gorm:"column:product_name;type:varchar(180);not null"`
	SKUCode         *string    `gorm:"column:sku_code;type:varchar(100);index"`
	BrandCode       *string    `gorm:"column:brand_code;type:varchar(80);index"`
	BrandName       *string    `gorm:"column:brand_name;type:varchar(120)"`
	CategoryCode    *string    `gorm:"column:category_code;type:varchar(100);index"`
	CategoryName    *string    `gorm:"column:category_name;type:varchar(160)"`
	ListPrice       float64    `gorm:"column:list_price;type:numeric(18,2);not null;default:0"`
	CostPrice       float64    `gorm:"column:cost_price;type:numeric(18,2);not null;default:0"`
	ProductStatus   string     `gorm:"column:product_status;type:varchar(32);not null;default:'active';index"`
	SourceBatchCode *string    `gorm:"column:source_batch_code;type:varchar(80);index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterStdProduct) TableName() string { return "data_center_std_products" }

type DataCenterStdStoreSalesDaily struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_store_sales_unique,unique"`
	CompanyID       *uint64    `gorm:"column:company_id;index"`
	DepartmentID    *uint64    `gorm:"column:department_id;index"`
	BrandCode       *string    `gorm:"column:brand_code;type:varchar(80);index"`
	BrandName       *string    `gorm:"column:brand_name;type:varchar(120)"`
	ChannelCode     *string    `gorm:"column:channel_code;type:varchar(80);index"`
	PlatformCode    *string    `gorm:"column:platform_code;type:varchar(80);index"`
	StoreCode       string     `gorm:"column:store_code;type:varchar(100);not null;index:idx_dc_store_sales_unique,unique"`
	StoreName       *string    `gorm:"column:store_name;type:varchar(180)"`
	StatDate        time.Time  `gorm:"column:stat_date;type:date;not null;index;index:idx_dc_store_sales_unique,unique"`
	GMV             float64    `gorm:"column:gmv;type:numeric(18,2);not null;default:0"`
	NetSales        float64    `gorm:"column:net_sales;type:numeric(18,2);not null;default:0"`
	OrderCount      int64      `gorm:"column:order_count;not null;default:0"`
	CustomerCount   int64      `gorm:"column:customer_count;not null;default:0"`
	RefundAmount    float64    `gorm:"column:refund_amount;type:numeric(18,2);not null;default:0"`
	TargetAmount    *float64   `gorm:"column:target_amount;type:numeric(18,2)"`
	SourceBatchCode *string    `gorm:"column:source_batch_code;type:varchar(80);index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterStdStoreSalesDaily) TableName() string {
	return "data_center_std_store_sales_daily"
}

type DataCenterMetricDefinition struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_metric_tenant_code,unique"`
	MetricCode      string     `gorm:"column:metric_code;type:varchar(80);not null;index:idx_dc_metric_tenant_code,unique"`
	MetricName      string     `gorm:"column:metric_name;type:varchar(120);not null"`
	MetricCategory  string     `gorm:"column:metric_category;type:varchar(80);not null;index"`
	Formula         *string    `gorm:"column:formula;type:text"`
	StatisticPeriod *string    `gorm:"column:statistic_period;type:varchar(80)"`
	Dimensions      string     `gorm:"column:dimensions;type:jsonb;not null;default:'[]'"`
	DataSource      *string    `gorm:"column:data_source;type:varchar(160)"`
	Enabled         bool       `gorm:"column:enabled;not null;default:true;index"`
	AnomalyEnabled  bool       `gorm:"column:anomaly_enabled;not null;default:false;index"`
	CreatedBy       *uint64    `gorm:"column:created_by"`
	UpdatedBy       *uint64    `gorm:"column:updated_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterMetricDefinition) TableName() string { return "data_center_metric_definitions" }

type DataCenterMetricResult struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID      uint64    `gorm:"column:tenant_id;not null;index;index:idx_dc_metric_result_unique,unique"`
	MetricCode    string    `gorm:"column:metric_code;type:varchar(80);not null;index;index:idx_dc_metric_result_unique,unique"`
	ResourceType  string    `gorm:"column:resource_type;type:varchar(80);not null;index;index:idx_dc_metric_result_unique,unique"`
	ResourceCode  string    `gorm:"column:resource_code;type:varchar(120);not null;index;index:idx_dc_metric_result_unique,unique"`
	ResourceName  *string   `gorm:"column:resource_name;type:varchar(180)"`
	StatDate      time.Time `gorm:"column:stat_date;type:date;not null;index;index:idx_dc_metric_result_unique,unique"`
	PeriodType    string    `gorm:"column:period_type;type:varchar(32);not null;index;index:idx_dc_metric_result_unique,unique"`
	MetricValue   float64   `gorm:"column:metric_value;type:numeric(18,4);not null;default:0"`
	CompareValue  *float64  `gorm:"column:compare_value;type:numeric(18,4)"`
	CompareRate   *float64  `gorm:"column:compare_rate;type:numeric(18,4)"`
	TargetValue   *float64  `gorm:"column:target_value;type:numeric(18,4)"`
	DimensionJSON string    `gorm:"column:dimension_json;type:jsonb;not null;default:'{}'"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`
}

func (DataCenterMetricResult) TableName() string { return "data_center_metric_results" }

type DataCenterAnomalyRule struct {
	ID                          uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID                    uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_rule_tenant_code,unique"`
	RuleCode                    string     `gorm:"column:rule_code;type:varchar(80);not null;index:idx_dc_rule_tenant_code,unique"`
	RuleName                    string     `gorm:"column:rule_name;type:varchar(140);not null"`
	BusinessDomain              string     `gorm:"column:business_domain;type:varchar(80);not null;index"`
	TargetObjectType            string     `gorm:"column:target_object_type;type:varchar(80);not null"`
	ScopeJSON                   string     `gorm:"column:scope_json;type:jsonb;not null;default:'{}'"`
	MetricConditionsJSON        string     `gorm:"column:metric_conditions_json;type:jsonb;not null"`
	LevelConfigJSON             string     `gorm:"column:level_config_json;type:jsonb;not null;default:'{}'"`
	ConfidenceConfigJSON        string     `gorm:"column:confidence_config_json;type:jsonb;not null;default:'{}'"`
	AIEnabled                   bool       `gorm:"column:ai_enabled;not null;default:true"`
	TaskEnabled                 bool       `gorm:"column:task_enabled;not null;default:true"`
	AutoTaskConfidenceThreshold int        `gorm:"column:auto_task_confidence_threshold;not null;default:80"`
	DefaultOwnerRole            *string    `gorm:"column:default_owner_role;type:varchar(80)"`
	DefaultDeadlineDays         int        `gorm:"column:default_deadline_days;not null;default:3"`
	ReviewMetricCodes           string     `gorm:"column:review_metric_codes;type:jsonb;not null;default:'[]'"`
	ReviewAfterDays             int        `gorm:"column:review_after_days;not null;default:3"`
	Priority                    int        `gorm:"column:priority;not null;default:100"`
	Enabled                     bool       `gorm:"column:enabled;not null;default:true;index"`
	CreatedBy                   *uint64    `gorm:"column:created_by"`
	UpdatedBy                   *uint64    `gorm:"column:updated_by"`
	CreatedAt                   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt                   time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt                   *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterAnomalyRule) TableName() string { return "data_center_anomaly_rules" }

type DataCenterAnomalyRecord struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_anomaly_dedupe,unique"`
	AnomalyCode     string     `gorm:"column:anomaly_code;type:varchar(80);not null;uniqueIndex"`
	RuleCode        string     `gorm:"column:rule_code;type:varchar(80);not null;index:idx_dc_anomaly_dedupe,unique"`
	Title           string     `gorm:"column:title;type:varchar(255);not null"`
	BusinessDomain  string     `gorm:"column:business_domain;type:varchar(80);not null;index"`
	ObjectType      string     `gorm:"column:object_type;type:varchar(80);not null;index:idx_dc_anomaly_dedupe,unique"`
	ObjectCode      string     `gorm:"column:object_code;type:varchar(120);not null;index:idx_dc_anomaly_dedupe,unique"`
	ObjectName      *string    `gorm:"column:object_name;type:varchar(180)"`
	StatDate        time.Time  `gorm:"column:stat_date;type:date;not null;index:idx_dc_anomaly_dedupe,unique"`
	AnomalyLevel    string     `gorm:"column:anomaly_level;type:varchar(32);not null;index"`
	ConfidenceScore int        `gorm:"column:confidence_score;not null;default:0"`
	ImpactAmount    float64    `gorm:"column:impact_amount;type:numeric(18,2);not null;default:0"`
	EvidenceJSON    string     `gorm:"column:evidence_json;type:jsonb;not null;default:'[]'"`
	OccurredAt      time.Time  `gorm:"column:occurred_at;not null;index"`
	AIStatus        string     `gorm:"column:ai_status;type:varchar(32);not null;default:'pending';index"`
	TaskStatus      string     `gorm:"column:task_status;type:varchar(32);not null;default:'none';index"`
	ReviewStatus    string     `gorm:"column:review_status;type:varchar(32);not null;default:'none';index"`
	Status          string     `gorm:"column:status;type:varchar(32);not null;default:'pending';index"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterAnomalyRecord) TableName() string { return "data_center_anomaly_records" }

type DataCenterAIDiagnosisRecord struct {
	ID                    uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID              uint64    `gorm:"column:tenant_id;not null;index"`
	AnomalyID             uint64    `gorm:"column:anomaly_id;not null;index"`
	AnomalyCode           string    `gorm:"column:anomaly_code;type:varchar(80);not null;index"`
	ProblemSummary        *string   `gorm:"column:problem_summary;type:text"`
	ImpactSummary         *string   `gorm:"column:impact_summary;type:text"`
	ReasonAnalysisJSON    string    `gorm:"column:reason_analysis_json;type:jsonb;not null;default:'[]'"`
	EvidenceSummaryJSON   string    `gorm:"column:evidence_summary_json;type:jsonb;not null;default:'[]'"`
	SuggestionJSON        string    `gorm:"column:suggestion_json;type:jsonb;not null;default:'[]'"`
	ConfidenceExplanation *string   `gorm:"column:confidence_explanation;type:text"`
	TaskSuggestionJSON    string    `gorm:"column:task_suggestion_json;type:jsonb;not null;default:'{}'"`
	ModelCode             *string   `gorm:"column:model_code;type:varchar(80)"`
	Status                string    `gorm:"column:status;type:varchar(32);not null;default:'success'"`
	ErrorMessage          *string   `gorm:"column:error_message;type:text"`
	GeneratedAt           time.Time `gorm:"column:generated_at;not null"`
	CreatedAt             time.Time `gorm:"column:created_at;not null"`
	UpdatedAt             time.Time `gorm:"column:updated_at;not null"`
}

func (DataCenterAIDiagnosisRecord) TableName() string { return "data_center_ai_diagnosis_records" }

type DataCenterRectificationTask struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID          uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_task_tenant_code,unique"`
	TaskCode          string     `gorm:"column:task_code;type:varchar(80);not null;index:idx_dc_task_tenant_code,unique"`
	AnomalyID         *uint64    `gorm:"column:anomaly_id;index"`
	AnomalyCode       *string    `gorm:"column:anomaly_code;type:varchar(80);index"`
	Title             string     `gorm:"column:title;type:varchar(255);not null"`
	TaskType          string     `gorm:"column:task_type;type:varchar(80);not null;default:'anomaly_rectification'"`
	OwnerUserID       *uint64    `gorm:"column:owner_user_id;index"`
	OwnerRole         *string    `gorm:"column:owner_role;type:varchar(80)"`
	CollaboratorIDs   string     `gorm:"column:collaborator_ids;type:jsonb;not null;default:'[]'"`
	Priority          string     `gorm:"column:priority;type:varchar(32);not null;default:'medium';index"`
	Deadline          *time.Time `gorm:"column:deadline;type:date;index"`
	TargetDesc        *string    `gorm:"column:target_desc;type:text"`
	AISuggestionJSON  string     `gorm:"column:ai_suggestion_json;type:jsonb;not null;default:'{}'"`
	ExecutionFeedback *string    `gorm:"column:execution_feedback;type:text"`
	Progress          int        `gorm:"column:progress;not null;default:0"`
	Status            string     `gorm:"column:status;type:varchar(32);not null;default:'pending';index"`
	ReviewStatus      string     `gorm:"column:review_status;type:varchar(32);not null;default:'none';index"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	CreatedBy         *uint64    `gorm:"column:created_by"`
	UpdatedBy         *uint64    `gorm:"column:updated_by"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterRectificationTask) TableName() string { return "data_center_rectification_tasks" }

type DataCenterTaskLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   uint64    `gorm:"column:tenant_id;not null;index"`
	TaskID     uint64    `gorm:"column:task_id;not null;index"`
	TaskCode   string    `gorm:"column:task_code;type:varchar(80);not null;index"`
	Action     string    `gorm:"column:action;type:varchar(60);not null"`
	FromStatus *string   `gorm:"column:from_status;type:varchar(32)"`
	ToStatus   *string   `gorm:"column:to_status;type:varchar(32)"`
	Content    *string   `gorm:"column:content;type:text"`
	OperatorID uint64    `gorm:"column:operator_id;not null;index"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
}

func (DataCenterTaskLog) TableName() string { return "data_center_task_logs" }

type DataCenterRectificationReview struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID            uint64     `gorm:"column:tenant_id;not null;index;index:idx_dc_review_tenant_code,unique"`
	ReviewCode          string     `gorm:"column:review_code;type:varchar(80);not null;index:idx_dc_review_tenant_code,unique"`
	TaskID              uint64     `gorm:"column:task_id;not null;index"`
	TaskCode            string     `gorm:"column:task_code;type:varchar(80);not null;index"`
	AnomalyID           *uint64    `gorm:"column:anomaly_id;index"`
	AnomalyCode         *string    `gorm:"column:anomaly_code;type:varchar(80);index"`
	BeforeMetricJSON    string     `gorm:"column:before_metric_json;type:jsonb;not null;default:'[]'"`
	AfterMetricJSON     string     `gorm:"column:after_metric_json;type:jsonb;not null;default:'[]'"`
	ImprovementResult   *string    `gorm:"column:improvement_result;type:varchar(128)"`
	ReviewConclusion    string     `gorm:"column:review_conclusion;type:varchar(64);not null;default:'weak';index"`
	AIReviewSummary     *string    `gorm:"column:ai_review_summary;type:text"`
	ManualReviewSummary *string    `gorm:"column:manual_review_summary;type:text"`
	ExperienceSummary   *string    `gorm:"column:experience_summary;type:text"`
	ReviewedAt          *time.Time `gorm:"column:reviewed_at;index"`
	CreatedBy           *uint64    `gorm:"column:created_by"`
	UpdatedBy           *uint64    `gorm:"column:updated_by"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt           *time.Time `gorm:"column:deleted_at;index"`
}

func (DataCenterRectificationReview) TableName() string { return "data_center_rectification_reviews" }
