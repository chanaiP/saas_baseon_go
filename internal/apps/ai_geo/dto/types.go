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
	SourceID   *uint64                `json:"source_id"`
	Platform   string                 `json:"platform"`
	Title      string                 `json:"title"`
	HeatScore  int                    `json:"heat_score"`
	SourceURL  string                 `json:"source_url"`
	CapturedAt string                 `json:"captured_at"`
	Metadata   map[string]interface{} `json:"metadata"`
	Status     string                 `json:"status"`
}

type ExternalSourceExtractPayload struct {
	URL         string `json:"url"`
	ExtractType string `json:"extract_type"`
	Instruction string `json:"instruction"`
	ContentType string `json:"content_type"`
	Platform    string `json:"platform"`
}

type StyleTemplatePayload struct {
	SourceID          *uint64                `json:"source_id"`
	TemplateCode      string                 `json:"template_code"`
	TemplateName      string                 `json:"template_name"`
	Description       string                 `json:"description"`
	ContentType       string                 `json:"content_type"`
	Platform          string                 `json:"platform"`
	ToneProfile       map[string]interface{} `json:"tone_profile"`
	StructureProfile  map[string]interface{} `json:"structure_profile"`
	TechniqueProfile  map[string]interface{} `json:"technique_profile"`
	StyleKeywords     []string               `json:"style_keywords"`
	PromptFragment    string                 `json:"prompt_fragment"`
	NegativeRules     []string               `json:"negative_rules"`
	ExtractionSummary map[string]interface{} `json:"extraction_summary"`
	Status            string                 `json:"status"`
}

type ExternalSourceExtractResult struct {
	Source        models.AiGeoExternalSource `json:"source"`
	StyleTemplate *models.AiGeoStyleTemplate `json:"style_template,omitempty"`
	HotspotDraft  map[string]interface{}     `json:"hotspot_draft,omitempty"`
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

type ChannelTestResult struct {
	ChannelID   uint64 `json:"channel_id"`
	ChannelCode string `json:"channel_code"`
	ChannelName string `json:"channel_name"`
	Status      string `json:"status"`
	Reachable   bool   `json:"reachable"`
	Message     string `json:"message"`
	CheckedAt   string `json:"checked_at"`
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
	BrandID        *uint64                    `json:"brand_id"`
	ProductID      *uint64                    `json:"product_id"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	Body           string                     `json:"body"`
	Keywords       []string                   `json:"keywords"`
	Conversation   []DraftConversationMessage `json:"conversation"`
	SourceSnapshot map[string]interface{}     `json:"source_snapshot"`
	Source         string                     `json:"source"`
}

type GenerateDraftPayload struct {
	BrandID         *uint64                    `json:"brand_id"`
	ProductID       *uint64                    `json:"product_id"`
	Skill           string                     `json:"skill"`
	HotspotID       *uint64                    `json:"hotspot_id"`
	StyleTemplateID *uint64                    `json:"style_template_id"`
	Prompt          string                     `json:"prompt"`
	Conversation    []DraftConversationMessage `json:"conversation"`
	SourceSnapshot  map[string]interface{}     `json:"source_snapshot"`
}

type DraftConversationMessage struct {
	Role      string `json:"role"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at,omitempty"`
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

func PublicData(data interface{}) interface{} {
	switch v := data.(type) {
	case models.AiGeoBrandCard:
		return brandResponse(v)
	case PageResponse[models.AiGeoBrandCard]:
		return mapPage(v, brandResponse)
	case models.AiGeoProductCard:
		return productResponse(v)
	case PageResponse[models.AiGeoProductCard]:
		return mapPage(v, productResponse)
	case models.AiGeoSKU:
		return skuResponse(v)
	case PageResponse[models.AiGeoSKU]:
		return mapPage(v, skuResponse)
	case models.AiGeoCompetitor:
		return competitorResponse(v)
	case PageResponse[models.AiGeoCompetitor]:
		return mapPage(v, competitorResponse)
	case models.AiGeoKeyword:
		return keywordResponse(v)
	case PageResponse[models.AiGeoKeyword]:
		return mapPage(v, keywordResponse)
	case models.AiGeoMaterialAsset:
		return materialAssetResponse(v)
	case PageResponse[models.AiGeoMaterialAsset]:
		return mapPage(v, materialAssetResponse)
	case models.AiGeoHotspot:
		return hotspotResponse(v)
	case PageResponse[models.AiGeoHotspot]:
		return mapPage(v, hotspotResponse)
	case models.AiGeoExternalSource:
		return externalSourceResponse(v)
	case models.AiGeoStyleTemplate:
		return styleTemplateResponse(v)
	case PageResponse[models.AiGeoStyleTemplate]:
		return mapPage(v, styleTemplateResponse)
	case ExternalSourceExtractResult:
		result := map[string]interface{}{"source": externalSourceResponse(v.Source)}
		if v.StyleTemplate != nil {
			result["style_template"] = styleTemplateResponse(*v.StyleTemplate)
		}
		if v.HotspotDraft != nil {
			result["hotspot_draft"] = v.HotspotDraft
		}
		return result
	case models.AiGeoChannelProfile:
		return channelResponse(v)
	case PageResponse[models.AiGeoChannelProfile]:
		return mapPage(v, channelResponse)
	case models.AiGeoChannelAccount:
		return channelAccountResponse(v)
	case PageResponse[models.AiGeoChannelAccount]:
		return mapPage(v, channelAccountResponse)
	case models.AiGeoDraft:
		return draftResponse(v)
	case PageResponse[models.AiGeoDraft]:
		return mapPage(v, draftResponse)
	case models.AiGeoChannelContent:
		return channelContentResponse(v)
	case PageResponse[models.AiGeoChannelContent]:
		return mapPage(v, channelContentResponse)
	case models.AiGeoPublishPlan:
		return publishPlanResponse(v)
	case PageResponse[models.AiGeoPublishPlan]:
		return mapPage(v, publishPlanResponse)
	case PublishPlanCalendar:
		days := make([]map[string]interface{}, 0, len(v.Days))
		for i := range v.Days {
			items := make([]map[string]interface{}, 0, len(v.Days[i].Items))
			for _, item := range v.Days[i].Items {
				items = append(items, publishPlanResponse(item))
			}
			days = append(days, map[string]interface{}{
				"date":       v.Days[i].Date,
				"total":      v.Days[i].Total,
				"scheduled":  v.Days[i].Scheduled,
				"publishing": v.Days[i].Publishing,
				"published":  v.Days[i].Published,
				"failed":     v.Days[i].Failed,
				"cancelled":  v.Days[i].Cancelled,
				"items":      items,
			})
		}
		return map[string]interface{}{"start_date": v.StartDate, "end_date": v.EndDate, "days": days}
	case models.AiGeoImportBatch:
		return importBatchResponse(v)
	case models.AiGeoImportError:
		return importErrorResponse(v)
	case PageResponse[models.AiGeoImportError]:
		return mapPage(v, importErrorResponse)
	case models.AiGeoAuditSuggestion:
		return auditSuggestionResponse(v)
	case PageResponse[models.AiGeoAuditSuggestion]:
		return mapPage(v, auditSuggestionResponse)
	default:
		return data
	}
}

type PublicPageResponse struct {
	Items []map[string]interface{} `json:"items"`
	Total int64                    `json:"total"`
	Skip  int                      `json:"skip"`
	Limit int                      `json:"limit"`
}

func mapPage[T any](page PageResponse[T], fn func(T) map[string]interface{}) PublicPageResponse {
	items := make([]map[string]interface{}, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, fn(item))
	}
	return PublicPageResponse{Items: items, Total: page.Total, Skip: page.Skip, Limit: page.Limit}
}

func brandResponse(row models.AiGeoBrandCard) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "brand_code": row.BrandCode, "brand_name": row.BrandName, "positioning": row.Positioning, "target_audience": row.TargetAudience, "price_band": row.PriceBand, "tone": row.Tone, "keywords": row.Keywords, "completeness": row.Completeness, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func productResponse(row models.AiGeoProductCard) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "brand_id": row.BrandID, "product_code": row.ProductCode, "product_name": row.ProductName, "category_name": row.CategoryName, "selling_points": row.SellingPoints, "faq": row.FAQ, "content_angles": row.ContentAngles, "completeness": row.Completeness, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func skuResponse(row models.AiGeoSKU) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "product_id": row.ProductID, "sku_code": row.SKUCode, "sku_name": row.SKUName, "attributes": row.Attributes, "price": row.Price, "image_url": row.ImageURL, "stock_status": row.StockStatus, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func competitorResponse(row models.AiGeoCompetitor) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "product_id": row.ProductID, "brand_name": row.BrandName, "product_name": row.ProductName, "price_text": row.PriceText, "point": row.Point, "difference": row.Difference, "angle": row.Angle, "link_url": row.LinkURL, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func keywordResponse(row models.AiGeoKeyword) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "brand_id": row.BrandID, "product_id": row.ProductID, "keyword_group": row.KeywordGroup, "keyword": row.Keyword, "intent": row.Intent, "source": row.Source, "weight": row.Weight, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func materialAssetResponse(row models.AiGeoMaterialAsset) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "brand_id": row.BrandID, "product_id": row.ProductID, "asset_type": row.AssetType, "asset_name": row.AssetName, "file_id": row.FileID, "url": row.URL, "metadata": row.Metadata, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func hotspotResponse(row models.AiGeoHotspot) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "source_id": row.SourceID, "platform": row.Platform, "title": row.Title, "heat_score": row.HeatScore, "source_url": row.SourceURL, "captured_at": row.CapturedAt, "metadata": row.Metadata, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func externalSourceResponse(row models.AiGeoExternalSource) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "source_type": row.SourceType, "source_url": row.SourceURL, "source_site": row.SourceSite, "source_title": row.SourceTitle, "raw_text": row.RawText, "clean_text": row.CleanText, "content_hash": row.ContentHash, "extracted_meta": row.ExtractedMeta, "extraction_status": row.ExtractionStatus, "extraction_error": row.ExtractionError, "captured_at": row.CapturedAt, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func styleTemplateResponse(row models.AiGeoStyleTemplate) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "source_id": row.SourceID, "template_code": row.TemplateCode, "template_name": row.TemplateName, "description": row.Description, "content_type": row.ContentType, "platform": row.Platform, "tone_profile": row.ToneProfile, "structure_profile": row.StructureProfile, "technique_profile": row.TechniqueProfile, "style_keywords": row.StyleKeywords, "prompt_fragment": row.PromptFragment, "negative_rules": row.NegativeRules, "extraction_summary": row.ExtractionSummary, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func channelResponse(row models.AiGeoChannelProfile) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "channel_code": row.ChannelCode, "channel_name": row.ChannelName, "channel_type": row.ChannelType, "entry_url": row.EntryURL, "content_forms": row.ContentForms, "support_modes": row.SupportModes, "default_publish_mode": row.DefaultPublishMode, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func channelAccountResponse(row models.AiGeoChannelAccount) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "channel_id": row.ChannelID, "account_name": row.AccountName, "external_account_id": row.ExternalAccountID, "auth_status": row.AuthStatus, "publish_status": row.PublishStatus, "expires_at": row.ExpiresAt, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func draftResponse(row models.AiGeoDraft) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "draft_code": row.DraftCode, "brand_id": row.BrandID, "product_id": row.ProductID, "title": row.Title, "summary": row.Summary, "body": row.Body, "keywords": row.Keywords, "conversation": row.Conversation, "source_snapshot": row.SourceSnapshot, "source": row.Source, "audit_status": row.AuditStatus, "channel_status": row.ChannelStatus, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func channelContentResponse(row models.AiGeoChannelContent) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "draft_id": row.DraftID, "channel_id": row.ChannelID, "title": row.Title, "body": row.Body, "audit_status": row.AuditStatus, "publish_status": row.PublishStatus, "status": row.Status, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func publishPlanResponse(row models.AiGeoPublishPlan) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "plan_code": row.PlanCode, "channel_content_id": row.ChannelContentID, "channel_id": row.ChannelID, "scheduled_at": row.ScheduledAt, "publish_method": row.PublishMethod, "automation_level": row.AutomationLevel, "status": row.Status, "published_url": row.PublishedURL, "fail_reason": row.FailReason, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func importBatchResponse(row models.AiGeoImportBatch) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "batch_code": row.BatchCode, "import_type": row.ImportType, "mapping_config": row.MappingConfig, "record_count": row.RecordCount, "success_count": row.SuccessCount, "failed_count": row.FailedCount, "status": row.Status, "error_message": row.ErrorMessage, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func importErrorResponse(row models.AiGeoImportError) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "batch_id": row.BatchID, "row_number": row.RowNumber, "field_name": row.FieldName, "error_code": row.ErrorCode, "error_message": row.ErrorMessage, "raw_data": row.RawData, "status": row.Status, "created_at": row.CreatedAt}
}

func auditSuggestionResponse(row models.AiGeoAuditSuggestion) map[string]interface{} {
	return map[string]interface{}{"id": row.ID, "object_type": row.ObjectType, "object_id": row.ObjectID, "object_code": row.ObjectCode, "scenario_code": row.ScenarioCode, "risk_level": row.RiskLevel, "passed": row.Passed, "summary": row.Summary, "suggestion_json": row.SuggestionJSON, "model_code": row.ModelCode, "status": row.Status, "error_message": row.ErrorMessage, "generated_at": row.GeneratedAt, "created_at": row.CreatedAt}
}
