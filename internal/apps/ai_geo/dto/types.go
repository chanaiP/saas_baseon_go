package dto

import (
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

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
	DraftID     uint64
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

type SKUPayload struct {
	ProductID   uint64                 `json:"product_id"`
	SKUCode     string                 `json:"sku_code"`
	SKUName     string                 `json:"sku_name"`
	Attributes  map[string]interface{} `json:"attributes"`
	Price       float64                `json:"price"`
	ImageURL    string                 `json:"image_url"`
	StockStatus string                 `json:"stock_status"`
	Status      string                 `json:"status"`
}

type CompetitorPayload struct {
	ProductID   uint64 `json:"product_id"`
	BrandName   string `json:"brand_name"`
	ProductName string `json:"product_name"`
	PriceText   string `json:"price_text"`
	Point       string `json:"point"`
	Difference  string `json:"difference"`
	Angle       string `json:"angle"`
	LinkURL     string `json:"link_url"`
	Status      string `json:"status"`
}

type KeywordPayload struct {
	BrandID      *uint64 `json:"brand_id"`
	ProductID    *uint64 `json:"product_id"`
	KeywordGroup string  `json:"keyword_group"`
	Keyword      string  `json:"keyword"`
	Intent       string  `json:"intent"`
	Source       string  `json:"source"`
	Weight       int     `json:"weight"`
	Status       string  `json:"status"`
}

type MaterialAssetPayload struct {
	BrandID   *uint64                `json:"brand_id"`
	ProductID *uint64                `json:"product_id"`
	AssetType string                 `json:"asset_type"`
	AssetName string                 `json:"asset_name"`
	URL       string                 `json:"url"`
	Metadata  map[string]interface{} `json:"metadata"`
	Status    string                 `json:"status"`
}

type HotspotPayload struct {
	Platform   string `json:"platform"`
	Title      string `json:"title"`
	HeatScore  int    `json:"heat_score"`
	SourceURL  string `json:"source_url"`
	CapturedAt string `json:"captured_at"`
	Status     string `json:"status"`
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

type ChannelContentReviewPayload struct {
	Opinion string `json:"opinion"`
}

type ReviewDraftPayload struct {
	Opinion string `json:"opinion"`
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

type PublishPlanCalendarDay struct {
	Date       string                    `json:"date"`
	Total      int                       `json:"total"`
	Scheduled  int                       `json:"scheduled"`
	Publishing int                       `json:"publishing"`
	Published  int                       `json:"published"`
	Failed     int                       `json:"failed"`
	Cancelled  int                       `json:"cancelled"`
	Items      []models.AiGeoPublishPlan `json:"items"`
}

type PublishPlanCalendar struct {
	StartDate string                   `json:"start_date"`
	EndDate   string                   `json:"end_date"`
	Days      []PublishPlanCalendarDay `json:"days"`
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
