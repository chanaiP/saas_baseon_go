package dto

import "time"

type PageRequest struct {
	Skip         int
	Limit        int
	TimeRange    string
	StartDate    *time.Time
	EndDate      *time.Time
	Keyword      string
	BrandCode    string
	ChannelCode  string
	PlatformCode string
	StoreCode    string
	ProductCode  string
	Status       string
	Level        string
	Domain       string
	DataType     string
}

type PageResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Skip  int   `json:"skip"`
	Limit int   `json:"limit"`
}

type Viewer struct {
	UserID          uint64
	TenantID        uint64
	IsPlatformAdmin bool
}

type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
}

type MetricDefinitionPayload struct {
	MetricCode      string   `json:"metric_code"`
	MetricName      string   `json:"metric_name"`
	MetricCategory  string   `json:"metric_category"`
	Formula         string   `json:"formula"`
	StatisticPeriod string   `json:"statistic_period"`
	Dimensions      []string `json:"dimensions"`
	DataSource      string   `json:"data_source"`
	Enabled         *bool    `json:"enabled"`
	AnomalyEnabled  *bool    `json:"anomaly_enabled"`
}

type AnomalyRulePayload struct {
	RuleCode                    string                 `json:"rule_code"`
	RuleName                    string                 `json:"rule_name"`
	BusinessDomain              string                 `json:"business_domain"`
	TargetObjectType            string                 `json:"target_object_type"`
	Scope                       map[string]interface{} `json:"scope"`
	MetricConditions            map[string]interface{} `json:"metric_conditions"`
	LevelConfig                 map[string]interface{} `json:"level_config"`
	ConfidenceConfig            map[string]interface{} `json:"confidence_config"`
	AIEnabled                   *bool                  `json:"ai_enabled"`
	TaskEnabled                 *bool                  `json:"task_enabled"`
	AutoTaskConfidenceThreshold *int                   `json:"auto_task_confidence_threshold"`
	DefaultOwnerRole            string                 `json:"default_owner_role"`
	DefaultDeadlineDays         *int                   `json:"default_deadline_days"`
	ReviewMetricCodes           []string               `json:"review_metric_codes"`
	ReviewAfterDays             *int                   `json:"review_after_days"`
	Priority                    *int                   `json:"priority"`
	Enabled                     *bool                  `json:"enabled"`
}

type RuleTestPayload struct {
	Metrics          map[string]float64     `json:"metrics"`
	MetricConditions map[string]interface{} `json:"metric_conditions"`
}

type TaskPayload struct {
	AnomalyID         *uint64                `json:"anomaly_id"`
	Title             string                 `json:"title"`
	TaskType          string                 `json:"task_type"`
	OwnerUserID       *uint64                `json:"owner_user_id"`
	OwnerRole         string                 `json:"owner_role"`
	CollaboratorIDs   []uint64               `json:"collaborator_ids"`
	Priority          string                 `json:"priority"`
	Deadline          string                 `json:"deadline"`
	TargetDesc        string                 `json:"target_desc"`
	AISuggestion      map[string]interface{} `json:"ai_suggestion"`
	ExecutionFeedback string                 `json:"execution_feedback"`
	Progress          *int                   `json:"progress"`
}

type FeedbackPayload struct {
	Content  string `json:"content"`
	Progress *int   `json:"progress"`
}

type ReviewPayload struct {
	TaskID              uint64                   `json:"task_id"`
	BeforeMetrics       []map[string]interface{} `json:"before_metrics"`
	AfterMetrics        []map[string]interface{} `json:"after_metrics"`
	ReviewConclusion    string                   `json:"review_conclusion"`
	AIReviewSummary     string                   `json:"ai_review_summary"`
	ManualReviewSummary string                   `json:"manual_review_summary"`
	ExperienceSummary   string                   `json:"experience_summary"`
}

type RawBatchPayload struct {
	BatchCode      string                   `json:"batch_code"`
	DataType       string                   `json:"data_type"`
	PlatformCode   string                   `json:"platform_code"`
	AppCode        string                   `json:"app_code"`
	ConnectionCode string                   `json:"connection_code"`
	SourceParams   map[string]interface{}   `json:"source_params"`
	Records        []map[string]interface{} `json:"records"`
}
