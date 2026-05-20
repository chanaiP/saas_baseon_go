package dto

import "time"

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

type PageRequest struct {
	Skip        int
	Limit       int
	Keyword     string
	Status      string
	BrandID     uint64
	ProductID   uint64
	ChannelID   uint64
	AuditStatus string
	StartDate   *time.Time
	EndDate     *time.Time
}

type PageResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Skip  int   `json:"skip"`
	Limit int   `json:"limit"`
}

type BrandPayload struct {
	BrandCode      string   `json:"brand_code"`
	BrandName      string   `json:"brand_name"`
	Positioning    string   `json:"positioning"`
	TargetAudience string   `json:"target_audience"`
	PriceBand      string   `json:"price_band"`
	Tone           string   `json:"tone"`
	Keywords       []string `json:"keywords"`
	Status         string   `json:"status"`
}

type ProductPayload struct {
	BrandID       uint64   `json:"brand_id"`
	ProductCode   string   `json:"product_code"`
	ProductName   string   `json:"product_name"`
	CategoryName  string   `json:"category_name"`
	SellingPoints []string `json:"selling_points"`
	FAQ           []string `json:"faq"`
	ContentAngles []string `json:"content_angles"`
	Status        string   `json:"status"`
}

type ChannelPayload struct {
	ChannelCode        string   `json:"channel_code"`
	ChannelName        string   `json:"channel_name"`
	ChannelType        string   `json:"channel_type"`
	EntryURL           string   `json:"entry_url"`
	ContentForms       []string `json:"content_forms"`
	SupportModes       []string `json:"support_modes"`
	DefaultPublishMode string   `json:"default_publish_mode"`
	Status             string   `json:"status"`
}

type ChannelAccountPayload struct {
	ChannelID         uint64 `json:"channel_id"`
	AccountName       string `json:"account_name"`
	ExternalAccountID string `json:"external_account_id"`
	AuthStatus        string `json:"auth_status"`
	PublishStatus     string `json:"publish_status"`
	ExpiresAt         string `json:"expires_at"`
	Status            string `json:"status"`
}

type DraftPayload struct {
	BrandID   *uint64  `json:"brand_id"`
	ProductID *uint64  `json:"product_id"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	Body      string   `json:"body"`
	Keywords  []string `json:"keywords"`
	Source    string   `json:"source"`
}

type GenerateDraftPayload struct {
	BrandID   *uint64 `json:"brand_id"`
	ProductID *uint64 `json:"product_id"`
	Skill     string  `json:"skill"`
	HotspotID *uint64 `json:"hotspot_id"`
	Prompt    string  `json:"prompt"`
}

type ChannelContentPayload struct {
	ChannelID uint64 `json:"channel_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type PublishPlanPayload struct {
	ChannelContentID uint64 `json:"channel_content_id"`
	ChannelID        uint64 `json:"channel_id"`
	ScheduledAt      string `json:"scheduled_at"`
	PublishMethod    string `json:"publish_method"`
	AutomationLevel  string `json:"automation_level"`
}

type PublishStatusPayload struct {
	Status       string `json:"status"`
	PublishedURL string `json:"published_url"`
	FailReason   string `json:"fail_reason"`
}

type ImportPayload struct {
	ImportType    string                   `json:"import_type"`
	MappingConfig map[string]interface{}   `json:"mapping_config"`
	Records       []map[string]interface{} `json:"records"`
}

type Overview struct {
	BrandCount          int64                  `json:"brand_count"`
	ProductCount        int64                  `json:"product_count"`
	SKUCount            int64                  `json:"sku_count"`
	ChannelCount        int64                  `json:"channel_count"`
	ChannelAccountCount int64                  `json:"channel_account_count"`
	DraftCountToday     int64                  `json:"draft_count_today"`
	PendingDraftCount   int64                  `json:"pending_draft_count"`
	ChannelContentCount int64                  `json:"channel_content_count"`
	PublishPlanToday    int64                  `json:"publish_plan_today"`
	AverageCompleteness int64                  `json:"average_completeness"`
	PendingTasks        []map[string]string    `json:"pending_tasks"`
	QuotaUsage          map[string]interface{} `json:"quota_usage"`
}
