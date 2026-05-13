package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/ai_capability_center"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

var ErrNotFound = errors.New("资源不存在")
var ErrInvalidInput = errors.New("参数不合法")
var ErrResourceInUse = errors.New("资源已被引用，不能删除")

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type PageResult struct {
	Items   interface{} `json:"items"`
	Total   int64       `json:"total"`
	Skip    int         `json:"skip"`
	Limit   int         `json:"limit"`
	Summary interface{} `json:"summary,omitempty"`
}

type Overview struct {
	Metrics        []OverviewMetric       `json:"metrics"`
	TenantMetrics  map[string]interface{} `json:"tenant_metrics"`
	UsageTrend     []UsageTrendPoint      `json:"usage_trend"`
	ModelCostShare []ModelCostShare       `json:"model_cost_share"`
	TenantRanking  []TenantRankingItem    `json:"tenant_ranking"`
	HealthChecks   []HealthCheck          `json:"health_checks"`
	CoreBaseRoutes []models.AIBaseRoute   `json:"core_base_routes"`
}

type OverviewMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Trend string `json:"trend"`
	Tone  string `json:"tone"`
}

type UsageTrendPoint struct {
	Date          string  `json:"date"`
	Calls         int64   `json:"calls"`
	CostAmount    float64 `json:"cost_amount"`
	BillingAmount float64 `json:"billing_amount"`
}

type ModelCostShare struct {
	ModelType  string  `json:"model_type"`
	CostAmount float64 `json:"cost_amount"`
}

type TenantRankingItem struct {
	TenantName    string  `json:"tenant_name"`
	Calls         int64   `json:"calls"`
	CostAmount    float64 `json:"cost_amount"`
	BillingAmount float64 `json:"billing_amount"`
	ProfitAmount  float64 `json:"profit_amount"`
	SuccessRate   float64 `json:"success_rate"`
	ScenarioCount int64   `json:"scenario_count"`
}

type UsageUnitSummary struct {
	UsageUnit     string  `json:"usage_unit"`
	UsageAmount   float64 `json:"usage_amount"`
	Calls         int64   `json:"calls"`
	CostAmount    float64 `json:"cost_amount"`
	BillingAmount float64 `json:"billing_amount"`
}

type ProviderImportRequest struct {
	Providers []ProviderImportProvider `json:"providers"`
	Accounts  []ProviderImportAccount  `json:"accounts"`
	APIs      []ProviderImportAPI      `json:"apis"`
}

type ProviderImportProvider struct {
	Name          string                  `json:"name"`
	Code          string                  `json:"code"`
	Type          string                  `json:"type"`
	BaseURL       string                  `json:"base_url"`
	AuthType      string                  `json:"auth_type"`
	Status        string                  `json:"status"`
	Priority      int                     `json:"priority"`
	Region        string                  `json:"region"`
	QPSLimit      int                     `json:"qps_limit"`
	MonthlyBudget float64                 `json:"monthly_budget"`
	Owner         string                  `json:"owner"`
	Remark        string                  `json:"remark"`
	Accounts      []ProviderImportAccount `json:"accounts"`
}

type ProviderImportAccount struct {
	ProviderCode    string              `json:"provider_code"`
	AccountName     string              `json:"account_name"`
	Endpoint        string              `json:"endpoint"`
	KeyAlias        string              `json:"key_alias"`
	EncryptedAPIKey string              `json:"encrypted_api_key"`
	EncryptedSecret string              `json:"encrypted_secret"`
	QuotaLimit      float64             `json:"quota_limit"`
	UsedQuota       float64             `json:"used_quota"`
	Status          string              `json:"status"`
	APIs            []ProviderImportAPI `json:"apis"`
}

type ProviderImportAPI struct {
	ProviderCode string   `json:"provider_code"`
	AccountName  string   `json:"account_name"`
	APIName      string   `json:"api_name"`
	APIPath      string   `json:"api_path"`
	APIType      string   `json:"api_type"`
	Capabilities []string `json:"capabilities"`
	AuthType     string   `json:"auth_type"`
	QPSLimit     int      `json:"qps_limit"`
	TimeoutMS    int      `json:"timeout_ms"`
	Status       string   `json:"status"`
}

type ProviderImportResult struct {
	Providers int `json:"providers"`
	Accounts  int `json:"accounts"`
	APIs      int `json:"apis"`
}

type ModelImportRequest struct {
	Models        []ModelImportModel       `json:"models"`
	PricePolicies []ModelImportPricePolicy `json:"price_policies"`
	PriceTiers    []ModelImportPriceTier   `json:"price_tiers"`
}

type ModelImportModel struct {
	ProviderCode  string                   `json:"provider_code"`
	ModelCode     string                   `json:"model_code"`
	ModelName     string                   `json:"model_name"`
	ModelType     string                   `json:"model_type"`
	Capabilities  []string                 `json:"capabilities"`
	ContextWindow int                      `json:"context_window"`
	Unit          string                   `json:"unit"`
	LatencyP95    int                      `json:"latency_p95"`
	SuccessRate   float64                  `json:"success_rate"`
	Status        string                   `json:"status"`
	DefaultFor    []string                 `json:"default_for"`
	Remark        string                   `json:"remark"`
	PricePolicies []ModelImportPricePolicy `json:"price_policies"`
}

type ModelImportPricePolicy struct {
	ProviderCode       string                 `json:"provider_code"`
	ModelCode          string                 `json:"model_code"`
	FeatureKey         string                 `json:"feature_key"`
	FeatureName        string                 `json:"feature_name"`
	ModelType          string                 `json:"model_type"`
	CapabilityCode     string                 `json:"capability_code"`
	BillingMode        string                 `json:"billing_mode"`
	BillingUnit        string                 `json:"billing_unit"`
	PlatformUnit       string                 `json:"platform_unit"`
	BaseCostPrice      float64                `json:"base_cost_price"`
	BaseSalePrice      float64                `json:"base_sale_price"`
	BasePlatformAmount float64                `json:"base_platform_amount"`
	Currency           string                 `json:"currency"`
	Status             string                 `json:"status"`
	Tiers              []ModelImportPriceTier `json:"tiers"`
}

type ModelImportPriceTier struct {
	ProviderCode    string  `json:"provider_code"`
	ModelCode       string  `json:"model_code"`
	FeatureKey      string  `json:"feature_key"`
	TierName        string  `json:"tier_name"`
	Mode            string  `json:"mode"`
	Resolution      string  `json:"resolution"`
	Quality         string  `json:"quality"`
	DurationSeconds int     `json:"duration_seconds"`
	AspectRatio     string  `json:"aspect_ratio"`
	CostPrice       float64 `json:"cost_price"`
	SalePrice       float64 `json:"sale_price"`
	PlatformAmount  float64 `json:"platform_amount"`
	Enabled         *bool   `json:"enabled"`
	SortOrder       int     `json:"sort_order"`
}

type ModelImportResult struct {
	Models        int `json:"models"`
	PricePolicies int `json:"price_policies"`
	PriceTiers    int `json:"price_tiers"`
}

type ScenarioImportRequest struct {
	Scenarios []ScenarioImportItem `json:"scenarios"`
}

type ScenarioImportItem struct {
	AppCode            string `json:"app_code"`
	AppName            string `json:"app_name"`
	AIScenarioCode     string `json:"ai_scenario_code"`
	AIScenarioName     string `json:"ai_scenario_name"`
	ScenarioType       string `json:"scenario_type"`
	CapabilityCode     string `json:"capability_code"`
	ModelType          string `json:"model_type"`
	DefaultBaseRouteID string `json:"default_base_route_id"`
	Owner              string `json:"owner"`
	Description        string `json:"description"`
	Version            string `json:"version"`
	Status             string `json:"status"`
}

type ScenarioImportResult struct {
	Scenarios int `json:"scenarios"`
}

type RouteImportRequest struct {
	BaseRoutes  []RouteImportBaseRoute  `json:"base_routes"`
	RouteModels []RouteImportRouteModel `json:"route_models"`
}

type RouteImportBaseRoute struct {
	RouteCode      string                  `json:"route_code"`
	RouteName      string                  `json:"route_name"`
	CapabilityCode string                  `json:"capability_code"`
	ModelType      string                  `json:"model_type"`
	Strategy       string                  `json:"strategy"`
	TimeoutMS      int                     `json:"timeout_ms"`
	MaxRetry       int                     `json:"max_retry"`
	Description    string                  `json:"description"`
	Status         string                  `json:"status"`
	RouteModels    []RouteImportRouteModel `json:"route_models"`
}

type RouteImportRouteModel struct {
	BaseRouteID   string `json:"base_route_id"`
	BaseRouteCode string `json:"base_route_code"`
	ModelID       string `json:"model_id"`
	ProviderCode  string `json:"provider_code"`
	ModelCode     string `json:"model_code"`
	Role          string `json:"role"`
	Priority      int    `json:"priority"`
	Weight        int    `json:"weight"`
	MaxRetry      int    `json:"max_retry"`
	TimeoutMS     int    `json:"timeout_ms"`
	Status        string `json:"status"`
}

type RouteImportResult struct {
	BaseRoutes  int `json:"base_routes"`
	RouteModels int `json:"route_models"`
}

type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *Service) ImportProviders(ctx context.Context, userID uint64, req ProviderImportRequest) (ProviderImportResult, error) {
	now := time.Now()
	result := ProviderImportResult{}
	if len(req.Providers) == 0 && len(req.Accounts) == 0 && len(req.APIs) == 0 {
		return result, ErrInvalidInput
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		providerIDs := map[string]string{}
		accountIDs := map[string]string{}
		for _, item := range req.Providers {
			provider, err := upsertProviderImport(ctx, tx, item, now)
			if err != nil {
				return err
			}
			result.Providers++
			providerIDs[provider.Code] = provider.ID
			for _, account := range item.Accounts {
				account.ProviderCode = provider.Code
				saved, err := upsertProviderAccountImport(ctx, tx, providerIDs, account, now)
				if err != nil {
					return err
				}
				result.Accounts++
				accountIDs[accountKey(provider.Code, saved.AccountName)] = saved.ID
				for _, api := range account.APIs {
					api.ProviderCode = provider.Code
					api.AccountName = saved.AccountName
					if err := upsertProviderAPIImport(ctx, tx, providerIDs, accountIDs, api, now); err != nil {
						return err
					}
					result.APIs++
				}
			}
		}
		if err := loadProviderIDs(ctx, tx, providerIDs); err != nil {
			return err
		}
		for _, item := range req.Accounts {
			saved, err := upsertProviderAccountImport(ctx, tx, providerIDs, item, now)
			if err != nil {
				return err
			}
			result.Accounts++
			accountIDs[accountKey(item.ProviderCode, saved.AccountName)] = saved.ID
			for _, api := range item.APIs {
				api.ProviderCode = item.ProviderCode
				api.AccountName = saved.AccountName
				if err := upsertProviderAPIImport(ctx, tx, providerIDs, accountIDs, api, now); err != nil {
					return err
				}
				result.APIs++
			}
		}
		if err := loadAccountIDs(ctx, tx, accountIDs); err != nil {
			return err
		}
		for _, item := range req.APIs {
			if err := upsertProviderAPIImport(ctx, tx, providerIDs, accountIDs, item, now); err != nil {
				return err
			}
			result.APIs++
		}
		return nil
	})
	if err != nil {
		return ProviderImportResult{}, err
	}
	s.Audit(ctx, userID, "ai_provider", "import", "整体导入 AI 供应商资源", result)
	return result, nil
}

func (s *Service) ImportModels(ctx context.Context, userID uint64, req ModelImportRequest) (ModelImportResult, error) {
	now := time.Now()
	result := ModelImportResult{}
	if len(req.Models) == 0 && len(req.PricePolicies) == 0 && len(req.PriceTiers) == 0 {
		return result, ErrInvalidInput
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		providerIDs := map[string]string{}
		modelIDs := map[string]string{}
		policyIDs := map[string]string{}
		if err := loadProviderIDs(ctx, tx, providerIDs); err != nil {
			return err
		}
		for _, item := range req.Models {
			model, err := upsertModelImport(ctx, tx, providerIDs, item, now)
			if err != nil {
				return err
			}
			result.Models++
			modelIDs[modelKey(item.ProviderCode, model.ModelCode)] = model.ID
			for _, policy := range item.PricePolicies {
				policy.ProviderCode = item.ProviderCode
				policy.ModelCode = model.ModelCode
				saved, err := upsertModelPricePolicyImport(ctx, tx, modelIDs, policy, now)
				if err != nil {
					return err
				}
				result.PricePolicies++
				policyIDs[policyKey(policy.ProviderCode, policy.ModelCode, saved.FeatureKey)] = saved.ID
				for _, tier := range policy.Tiers {
					tier.ProviderCode = policy.ProviderCode
					tier.ModelCode = policy.ModelCode
					tier.FeatureKey = saved.FeatureKey
					if err := upsertModelPriceTierImport(ctx, tx, policyIDs, tier, now); err != nil {
						return err
					}
					result.PriceTiers++
				}
			}
		}
		if err := loadModelIDs(ctx, tx, modelIDs); err != nil {
			return err
		}
		for _, item := range req.PricePolicies {
			saved, err := upsertModelPricePolicyImport(ctx, tx, modelIDs, item, now)
			if err != nil {
				return err
			}
			result.PricePolicies++
			policyIDs[policyKey(item.ProviderCode, item.ModelCode, saved.FeatureKey)] = saved.ID
			for _, tier := range item.Tiers {
				tier.ProviderCode = item.ProviderCode
				tier.ModelCode = item.ModelCode
				tier.FeatureKey = saved.FeatureKey
				if err := upsertModelPriceTierImport(ctx, tx, policyIDs, tier, now); err != nil {
					return err
				}
				result.PriceTiers++
			}
		}
		if err := loadPolicyIDs(ctx, tx, policyIDs); err != nil {
			return err
		}
		for _, item := range req.PriceTiers {
			if err := upsertModelPriceTierImport(ctx, tx, policyIDs, item, now); err != nil {
				return err
			}
			result.PriceTiers++
		}
		return nil
	})
	if err != nil {
		return ModelImportResult{}, err
	}
	s.Audit(ctx, userID, "ai_model", "import", "整体导入 AI 模型和价格资源", result)
	return result, nil
}

func (s *Service) ImportScenarios(ctx context.Context, userID uint64, req ScenarioImportRequest) (ScenarioImportResult, error) {
	now := time.Now()
	result := ScenarioImportResult{}
	if len(req.Scenarios) == 0 {
		return result, ErrInvalidInput
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Scenarios {
			if err := upsertScenarioImport(ctx, tx, item, now); err != nil {
				return err
			}
			result.Scenarios++
		}
		return nil
	})
	if err != nil {
		return ScenarioImportResult{}, err
	}
	s.Audit(ctx, userID, "ai_scenario", "import", "批量导入 AI 场景", result)
	return result, nil
}

func (s *Service) ImportRoutes(ctx context.Context, userID uint64, req RouteImportRequest) (RouteImportResult, error) {
	now := time.Now()
	result := RouteImportResult{}
	if len(req.BaseRoutes) == 0 && len(req.RouteModels) == 0 {
		return result, ErrInvalidInput
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		modelIDs := map[string]string{}
		routeIDs := map[string]string{}
		if err := loadModelIDs(ctx, tx, modelIDs); err != nil {
			return err
		}
		if err := loadBaseRouteIDs(ctx, tx, routeIDs); err != nil {
			return err
		}
		for _, item := range req.BaseRoutes {
			route, err := upsertBaseRouteImport(ctx, tx, item, now)
			if err != nil {
				return err
			}
			result.BaseRoutes++
			routeIDs[route.RouteCode] = route.ID
			for _, routeModel := range item.RouteModels {
				routeModel.BaseRouteCode = route.RouteCode
				if err := upsertRouteModelImport(ctx, tx, routeIDs, modelIDs, routeModel, now); err != nil {
					return err
				}
				result.RouteModels++
			}
		}
		if err := loadBaseRouteIDs(ctx, tx, routeIDs); err != nil {
			return err
		}
		for _, item := range req.RouteModels {
			if err := upsertRouteModelImport(ctx, tx, routeIDs, modelIDs, item, now); err != nil {
				return err
			}
			result.RouteModels++
		}
		return nil
	})
	if err != nil {
		return RouteImportResult{}, err
	}
	s.Audit(ctx, userID, "ai_base_route", "import", "整体导入基础路由和模型池", result)
	return result, nil
}

type InvokeRequest struct {
	TenantID       string                 `json:"tenant_id"`
	TenantName     string                 `json:"tenant_name"`
	AppCode        string                 `json:"app_code"`
	AppName        string                 `json:"app_name"`
	AIScenarioCode string                 `json:"ai_scenario_code"`
	UserID         string                 `json:"user_id"`
	UserName       string                 `json:"user_name"`
	RequestID      string                 `json:"request_id"`
	Params         map[string]interface{} `json:"params"`
	Input          map[string]interface{} `json:"input"`
}

type InvokeResponse struct {
	RequestID   string                 `json:"request_id"`
	Status      string                 `json:"status"`
	ModelID     string                 `json:"model_id"`
	BaseRouteID string                 `json:"base_route_id"`
	Usage       map[string]interface{} `json:"usage"`
	Billing     map[string]interface{} `json:"billing"`
	Data        map[string]interface{} `json:"data"`
}

func (s *Service) EnsureBaseline(ctx context.Context) error {
	capabilities := []models.AICapability{
		{CapabilityCode: "text_generation", CapabilityName: "文本生成", ScenarioType: "text", ModelType: "text", DefaultBillingUnit: "tokens", SupportsTierPricing: false, Description: "标题、卖点、详情、营销文案等文本生成。", Status: "active", SortOrder: 10},
		{CapabilityCode: "chat_completion", CapabilityName: "对话生成", ScenarioType: "text", ModelType: "text", DefaultBillingUnit: "tokens", SupportsTierPricing: false, Description: "客服回复、对话问答、复杂推理。", Status: "active", SortOrder: 20},
		{CapabilityCode: "image_generation", CapabilityName: "图片生成", ScenarioType: "image", ModelType: "image", DefaultBillingUnit: "images", SupportsTierPricing: true, Description: "文生图、图生图、营销图、商品图。", Status: "active", SortOrder: 30},
		{CapabilityCode: "video_generation", CapabilityName: "视频生成", ScenarioType: "video", ModelType: "video", DefaultBillingUnit: "video_seconds", SupportsTierPricing: true, Description: "文生视频、图生视频、商品短视频。", Status: "active", SortOrder: 40},
		{CapabilityCode: "embedding", CapabilityName: "向量生成", ScenarioType: "text", ModelType: "embedding", DefaultBillingUnit: "tokens", SupportsTierPricing: false, Description: "知识库检索、语义搜索、相似度匹配。", Status: "active", SortOrder: 50},
		{CapabilityCode: "rerank", CapabilityName: "重排", ScenarioType: "text", ModelType: "rerank", DefaultBillingUnit: "requests", SupportsTierPricing: false, Description: "检索结果重排与排序优化。", Status: "active", SortOrder: 60},
		{CapabilityCode: "agent_run", CapabilityName: "Agent 执行", ScenarioType: "agent", ModelType: "text", DefaultBillingUnit: "agent_runs", SupportsTierPricing: false, Description: "Agent 只作为 AI 场景类型，编排由 Agent 工厂维护。", Status: "active", SortOrder: 70},
	}
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range capabilities {
			var count int64
			if err := tx.Model(&models.AICapability{}).Where("capability_code = ? AND deleted_at IS NULL", item.CapabilityCode).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			item.ID = uuid.NewString()
			item.CreatedAt = now
			item.UpdatedAt = now
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		settings := []models.AIGatewaySetting{
			{SettingKey: "gateway_runtime", SettingValue: `{"default_timeout_ms":30000,"default_max_retry":2,"usage_log_async":true}`, Description: "AI Gateway 默认超时、重试和用量日志写入策略", Status: "active"},
			{SettingKey: "security", SettingValue: `{"prompt_plaintext_storage":false,"api_key_encryption":"external-kms-or-env"}`, Description: "密钥和 Prompt 安全归属", Status: "active"},
		}
		for _, item := range settings {
			var count int64
			if err := tx.Model(&models.AIGatewaySetting{}).Where("setting_key = ? AND deleted_at IS NULL", item.SettingKey).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			item.ID = uuid.NewString()
			item.CreatedAt = now
			item.UpdatedAt = now
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) Overview(ctx context.Context) (Overview, error) {
	if err := s.EnsureBaseline(ctx); err != nil {
		return Overview{}, err
	}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	trendStart := todayStart.AddDate(0, 0, -6)
	var calls int64
	var cost, billing float64
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).
		Select("COALESCE(SUM(calls),0), COALESCE(SUM(cost_amount),0), COALESCE(SUM(billing_amount),0)").
		Where("called_at >= ?", todayStart).
		Row().Scan(&calls, &cost, &billing); err != nil {
		return Overview{}, err
	}

	var successCount int64
	var totalRecords int64
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).Where("called_at >= ?", todayStart).Count(&totalRecords).Error; err != nil {
		return Overview{}, err
	}
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).Where("called_at >= ? AND status = ?", todayStart, "success").Count(&successCount).Error; err != nil {
		return Overview{}, err
	}
	successRate := 100.0
	if totalRecords > 0 {
		successRate = float64(successCount) / float64(totalRecords) * 100
	}

	p95, err := s.p95Latency(ctx, todayStart)
	if err != nil {
		return Overview{}, err
	}
	var routeRows []models.AIBaseRoute
	if err := s.db.WithContext(ctx).Where("deleted_at IS NULL").Order("updated_at desc").Limit(4).Find(&routeRows).Error; err != nil {
		return Overview{}, err
	}
	trend, err := s.usageTrend(ctx, trendStart, todayStart)
	if err != nil {
		return Overview{}, err
	}
	costShare, err := s.modelCostShare(ctx, trendStart)
	if err != nil {
		return Overview{}, err
	}
	ranking, err := s.tenantRanking(ctx, trendStart)
	if err != nil {
		return Overview{}, err
	}
	var tenantCount int64
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).Where("called_at >= ?", trendStart).Distinct("tenant_id").Count(&tenantCount).Error; err != nil {
		return Overview{}, err
	}

	return Overview{
		Metrics: []OverviewMetric{
			{Label: "今日调用量", Value: fmt.Sprintf("%d", calls), Trend: "实时", Tone: "blue"},
			{Label: "今日成本", Value: fmt.Sprintf("¥%.2f", cost), Trend: "成本价", Tone: "green"},
			{Label: "成功率", Value: fmt.Sprintf("%.2f%%", successRate), Trend: "按记录数", Tone: "purple"},
			{Label: "P95 延迟", Value: fmt.Sprintf("%dms", p95), Trend: "近似值", Tone: "orange"},
		},
		TenantMetrics: map[string]interface{}{
			"service_tenants": tenantCount,
			"tenant_revenue":  billing,
			"tenant_profit":   billing - cost,
		},
		UsageTrend:     trend,
		ModelCostShare: costShare,
		TenantRanking:  ranking,
		HealthChecks: []HealthCheck{
			{Name: "供应商可用状态", Status: "active", Message: "基于供应商启停状态检测"},
			{Name: "基础路由状态", Status: "active", Message: "启用路由可供 AI 场景绑定"},
			{Name: "租户策略状态", Status: "active", Message: "策略中心统一维护覆盖、配额和限流"},
			{Name: "底座操作日志", Status: "active", Message: "配置变更写入 SaaS 底座操作日志"},
		},
		CoreBaseRoutes: routeRows,
	}, nil
}

func (s *Service) usageTrend(ctx context.Context, start, end time.Time) ([]UsageTrendPoint, error) {
	points := make([]UsageTrendPoint, 0, 7)
	byDate := make(map[string]*UsageTrendPoint, 7)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		points = append(points, UsageTrendPoint{Date: key})
		byDate[key] = &points[len(points)-1]
	}
	var rows []UsageTrendPoint
	dateExpr := usageDateExpression(s.db)
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).
		Select(dateExpr+" AS date, COALESCE(SUM(calls),0) AS calls, COALESCE(SUM(cost_amount),0) AS cost_amount, COALESCE(SUM(billing_amount),0) AS billing_amount").
		Where("called_at >= ?", start).
		Group(dateExpr).
		Order("date asc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if point, ok := byDate[row.Date]; ok {
			point.Calls = row.Calls
			point.CostAmount = row.CostAmount
			point.BillingAmount = row.BillingAmount
		}
	}
	return points, nil
}

func (s *Service) modelCostShare(ctx context.Context, start time.Time) ([]ModelCostShare, error) {
	var rows []ModelCostShare
	if err := s.db.WithContext(ctx).Table("ai_usage_records AS u").
		Select("COALESCE(m.model_type, 'unknown') AS model_type, COALESCE(SUM(u.cost_amount),0) AS cost_amount").
		Joins("LEFT JOIN ai_models AS m ON m.id = u.model_id").
		Where("u.called_at >= ?", start).
		Group("COALESCE(m.model_type, 'unknown')").
		Order("cost_amount desc").
		Limit(6).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) tenantRanking(ctx context.Context, start time.Time) ([]TenantRankingItem, error) {
	var rows []TenantRankingItem
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).
		Select(`tenant_name,
			COALESCE(SUM(calls),0) AS calls,
			COALESCE(SUM(cost_amount),0) AS cost_amount,
			COALESCE(SUM(billing_amount),0) AS billing_amount,
			COALESCE(SUM(billing_amount - cost_amount),0) AS profit_amount,
			COALESCE(AVG(CASE WHEN status = 'success' THEN 100.0 ELSE 0.0 END),100) AS success_rate,
			COUNT(DISTINCT ai_scenario_code) AS scenario_count`).
		Where("called_at >= ?", start).
		Group("tenant_name").
		Order("billing_amount desc").
		Limit(8).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) p95Latency(ctx context.Context, start time.Time) (int, error) {
	var latencies []int
	if err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).
		Where("called_at >= ?", start).
		Order("latency_ms asc").
		Pluck("latency_ms", &latencies).Error; err != nil {
		return 0, err
	}
	if len(latencies) == 0 {
		return 0, nil
	}
	index := int(float64(len(latencies))*0.95+0.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	return latencies[index], nil
}

func usageDateExpression(db *gorm.DB) string {
	if db != nil && db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return "strftime('%Y-%m-%d', called_at)"
	}
	return "TO_CHAR(called_at, 'YYYY-MM-DD')"
}

func (s *Service) ListProviders(ctx context.Context, skip, limit int, keyword string) (PageResult, error) {
	return listRows[models.AIProvider](ctx, s.db, skip, limit, keyword, []string{"name", "code", "owner", "region"})
}

func (s *Service) ListAccounts(ctx context.Context, skip, limit int, providerID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIProviderAccount{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(providerID) != "" {
		q = q.Where("provider_id = ?", providerID)
	}
	return pageQuery[models.AIProviderAccount](q, skip, limit)
}

func (s *Service) ListAPIs(ctx context.Context, skip, limit int, providerID, accountID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(providerID) != "" {
		q = q.Where("provider_id = ?", providerID)
	}
	if strings.TrimSpace(accountID) != "" {
		q = q.Where("account_id = ?", accountID)
	}
	return pageQuery[models.AIProviderAPI](q, skip, limit)
}

func (s *Service) ListCapabilities(ctx context.Context, skip, limit int, keyword string) (PageResult, error) {
	if err := s.EnsureBaseline(ctx); err != nil {
		return PageResult{}, err
	}
	return listRows[models.AICapability](ctx, s.db, skip, limit, keyword, []string{"capability_code", "capability_name", "description"})
}

func (s *Service) ListModels(ctx context.Context, skip, limit int, providerID, keyword string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIModel{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(providerID) != "" {
		q = q.Where("provider_id = ?", providerID)
	}
	if strings.TrimSpace(keyword) != "" {
		like := "%" + strings.TrimSpace(keyword) + "%"
		q = q.Where("model_code ILIKE ? OR model_name ILIKE ? OR model_type ILIKE ?", like, like, like)
	}
	return pageQuery[models.AIModel](q, skip, limit)
}

func (s *Service) ListPricePolicies(ctx context.Context, skip, limit int, modelID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIModelPricePolicy{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(modelID) != "" {
		q = q.Where("model_id = ?", modelID)
	}
	return pageQuery[models.AIModelPricePolicy](q, skip, limit)
}

func (s *Service) ListPriceTiers(ctx context.Context, skip, limit int, policyID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIModelPriceTier{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(policyID) != "" {
		q = q.Where("price_policy_id = ?", policyID)
	}
	return pageQuery[models.AIModelPriceTier](q, skip, limit)
}

func (s *Service) ListBaseRoutes(ctx context.Context, skip, limit int, keyword string) (PageResult, error) {
	return listRows[models.AIBaseRoute](ctx, s.db, skip, limit, keyword, []string{"route_code", "route_name", "capability_code", "strategy"})
}

func (s *Service) ListRouteModels(ctx context.Context, skip, limit int, baseRouteID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIBaseRouteModel{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(baseRouteID) != "" {
		q = q.Where("base_route_id = ?", baseRouteID)
	}
	return pageQuery[models.AIBaseRouteModel](q, skip, limit)
}

func (s *Service) ListScenarios(ctx context.Context, skip, limit int, keyword string) (PageResult, error) {
	return listRows[models.AIScenario](ctx, s.db, skip, limit, keyword, []string{"app_code", "app_name", "ai_scenario_code", "ai_scenario_name", "owner"})
}

func (s *Service) ListTenantStrategies(ctx context.Context, skip, limit int, keyword string) (PageResult, error) {
	return listRows[models.AITenantStrategyPolicy](ctx, s.db, skip, limit, keyword, []string{"policy_name", "app_code", "app_name", "ai_scenario_code", "ai_scenario_name"})
}

func (s *Service) ListQuotaRules(ctx context.Context, skip, limit int, policyID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIStrategyQuotaRule{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(policyID) != "" {
		q = q.Where("policy_id = ?", policyID)
	}
	return pageQuery[models.AIStrategyQuotaRule](q, skip, limit)
}

func (s *Service) ListRateLimitRules(ctx context.Context, skip, limit int, policyID string) (PageResult, error) {
	q := s.db.WithContext(ctx).Model(&models.AIStrategyRateLimitRule{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(policyID) != "" {
		q = q.Where("policy_id = ?", policyID)
	}
	return pageQuery[models.AIStrategyRateLimitRule](q, skip, limit)
}

func (s *Service) ListUsageRecords(ctx context.Context, skip, limit int, keyword, startDate, endDate string) (PageResult, error) {
	q := keywordQuery(s.db.WithContext(ctx).Model(&models.AIUsageRecord{}), keyword, []string{"tenant_name", "app_name", "ai_scenario_name", "user_name", "usage_detail", "status"})
	q, err := usageDateQuery(q, startDate, endDate)
	if err != nil {
		return PageResult{}, err
	}
	result, err := pageQueryOrder[models.AIUsageRecord](q, skip, limit, "called_at desc")
	if err != nil {
		return PageResult{}, err
	}
	var summary []UsageUnitSummary
	if err := q.Session(&gorm.Session{}).
		Select("usage_unit, COALESCE(SUM(usage_amount),0) AS usage_amount, COALESCE(SUM(calls),0) AS calls, COALESCE(SUM(cost_amount),0) AS cost_amount, COALESCE(SUM(billing_amount),0) AS billing_amount").
		Group("usage_unit").
		Order("usage_unit asc").
		Scan(&summary).Error; err != nil {
		return PageResult{}, err
	}
	result.Summary = summary
	return result, nil
}

func (s *Service) ListSettings(ctx context.Context, skip, limit int) (PageResult, error) {
	if err := s.EnsureBaseline(ctx); err != nil {
		return PageResult{}, err
	}
	return pageQuery[models.AIGatewaySetting](s.db.WithContext(ctx).Model(&models.AIGatewaySetting{}).Where("deleted_at IS NULL"), skip, limit)
}

func (s *Service) Invoke(ctx context.Context, req InvokeRequest) (InvokeResponse, error) {
	req.AppCode = strings.TrimSpace(req.AppCode)
	req.AIScenarioCode = strings.TrimSpace(req.AIScenarioCode)
	req.TenantID = strings.TrimSpace(req.TenantID)
	if req.AppCode == "" || req.AIScenarioCode == "" || req.TenantID == "" {
		return InvokeResponse{}, ErrInvalidInput
	}
	if req.RequestID == "" {
		req.RequestID = fmt.Sprintf("ai_req_%d", time.Now().UnixNano())
	}
	var scenario models.AIScenario
	if err := s.db.WithContext(ctx).Where("app_code = ? AND ai_scenario_code = ? AND status = ? AND deleted_at IS NULL", req.AppCode, req.AIScenarioCode, "active").First(&scenario).Error; err != nil {
		return InvokeResponse{}, ErrNotFound
	}
	var route models.AIBaseRoute
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", scenario.DefaultBaseRouteID).First(&route).Error; err != nil {
		return InvokeResponse{}, ErrNotFound
	}
	var routeModel models.AIBaseRouteModel
	_ = s.db.WithContext(ctx).Where("base_route_id = ? AND deleted_at IS NULL", route.ID).Order("priority asc, weight desc").First(&routeModel).Error

	paramsRaw, _ := json.Marshal(req.Params)
	hash := sha256.Sum256(paramsRaw)
	now := time.Now()
	record := models.AIUsageRecord{
		RequestID:      req.RequestID,
		TenantID:       req.TenantID,
		TenantName:     req.TenantName,
		AppCode:        scenario.AppCode,
		AppName:        defaultString(req.AppName, scenario.AppName),
		AIScenarioCode: scenario.AIScenarioCode,
		AIScenarioName: scenario.AIScenarioName,
		UserID:         req.UserID,
		UserName:       req.UserName,
		ModelID:        routeModel.ModelID,
		BaseRouteID:    route.ID,
		UsageAmount:    1,
		UsageUnit:      usageUnitForCapability(ctx, s.db, scenario.CapabilityCode),
		UsageDetail:    "Gateway 校验与路由命中记录",
		Calls:          1,
		Status:         "success",
		LatencyMS:      0,
		RequestParams:  string(paramsRaw),
		PromptHash:     hex.EncodeToString(hash[:]),
		CalledAt:       now,
		CreatedAt:      now,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return InvokeResponse{}, err
	}
	return InvokeResponse{
		RequestID:   req.RequestID,
		Status:      "success",
		ModelID:     routeModel.ModelID,
		BaseRouteID: route.ID,
		Usage: map[string]interface{}{
			"amount": record.UsageAmount,
			"unit":   record.UsageUnit,
		},
		Billing: map[string]interface{}{
			"cost_amount":     record.CostAmount,
			"billing_amount":  record.BillingAmount,
			"platform_unit":   record.PlatformUnit,
			"platform_amount": record.PlatformAmount,
		},
		Data: map[string]interface{}{},
	}, nil
}

func (s *Service) CreateResource(ctx context.Context, userID uint64, resource string, payload map[string]interface{}) (interface{}, error) {
	now := time.Now()
	switch resource {
	case "providers":
		var row models.AIProvider
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if err := validateProvider(row); err != nil {
			return nil, err
		}
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_provider", "create", "新增 AI 供应商："+row.Name, row)
		return row, nil
	case "accounts":
		var row models.AIProviderAccount
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AIProvider{}, row.ProviderID) || strings.TrimSpace(row.AccountName) == "" || strings.TrimSpace(row.KeyAlias) == "" {
			return nil, ErrInvalidInput
		}
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_provider_account", "create", "新增 AI 供应商账号："+row.AccountName, row)
		return row, nil
	case "apis":
		var row models.AIProviderAPI
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AIProvider{}, row.ProviderID) || !s.exists(ctx, &models.AIProviderAccount{}, row.AccountID) ||
			strings.TrimSpace(row.APIName) == "" || strings.TrimSpace(row.APIPath) == "" || strings.TrimSpace(row.APIType) == "" || len(row.Capabilities) == 0 {
			return nil, ErrInvalidInput
		}
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_provider_api", "create", "新增 AI API："+row.APIName, row)
		return row, nil
	case "capabilities":
		var row models.AICapability
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_capability", "create", "新增 AI 能力："+row.CapabilityName, row)
		return row, nil
	case "models":
		var row models.AIModel
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AIProvider{}, row.ProviderID) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_model", "create", "新增 AI 模型："+row.ModelName, row)
		return row, nil
	case "price-policies":
		var row models.AIModelPricePolicy
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AIModel{}, row.ModelID) || !s.capabilityExists(ctx, row.CapabilityCode) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_model_price_policy", "create", "新增模型价格策略："+row.FeatureName, row)
		return row, nil
	case "price-tiers":
		var row models.AIModelPriceTier
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AIModelPricePolicy{}, row.PricePolicyID) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_model_price_tier", "create", "新增模型分档价格："+row.TierName, row)
		return row, nil
	case "base-routes":
		var row models.AIBaseRoute
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.capabilityExists(ctx, row.CapabilityCode) || !validRouteStrategy(row.Strategy) || row.TimeoutMS <= 0 || row.MaxRetry < 0 {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_base_route", "create", "新增基础路由："+row.RouteName, row)
		return row, nil
	case "route-models":
		var row models.AIBaseRouteModel
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AIBaseRoute{}, row.BaseRouteID) || !s.exists(ctx, &models.AIModel{}, row.ModelID) ||
			!validRouteModelRole(row.Role) || row.Priority <= 0 || row.Weight <= 0 || row.MaxRetry < 0 || row.TimeoutMS <= 0 {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_base_route_model", "create", "新增基础路由模型节点", row)
		return row, nil
	case "scenarios":
		var row models.AIScenario
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.capabilityExists(ctx, row.CapabilityCode) || !s.exists(ctx, &models.AIBaseRoute{}, row.DefaultBaseRouteID) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_scenario", "create", "新增 AI 场景："+row.AIScenarioName, row)
		return row, nil
	case "tenant-strategies":
		var row models.AITenantStrategyPolicy
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.scenarioExists(ctx, row.AppCode, row.AIScenarioCode) || !s.exists(ctx, &models.AIBaseRoute{}, row.DefaultBaseRouteID) {
			return nil, ErrInvalidInput
		}
		if strings.TrimSpace(row.OverrideBaseRouteID) != "" && !s.exists(ctx, &models.AIBaseRoute{}, row.OverrideBaseRouteID) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_tenant_strategy", "create", "新增租户策略："+row.PolicyName, row)
		return row, nil
	case "quota-rules":
		var row models.AIStrategyQuotaRule
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AITenantStrategyPolicy{}, row.PolicyID) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_strategy_quota_rule", "create", "新增策略配额规则", row)
		return row, nil
	case "rate-limit-rules":
		var row models.AIStrategyRateLimitRule
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if !s.exists(ctx, &models.AITenantStrategyPolicy{}, row.PolicyID) {
			return nil, ErrInvalidInput
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_strategy_rate_limit_rule", "create", "新增策略限流规则", row)
		return row, nil
	case "usage-records":
		var row models.AIUsageRecord
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if row.CalledAt.IsZero() {
			row.CalledAt = now
		}
		row.CreatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		return row, nil
	case "settings":
		var row models.AIGatewaySetting
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		s.Audit(ctx, userID, "ai_gateway_setting", "create", "新增 AI Gateway 设置："+row.SettingKey, row)
		return row, nil
	default:
		return nil, ErrNotFound
	}
}

func (s *Service) UpdateResource(ctx context.Context, userID uint64, resource, id string, payload map[string]interface{}) (interface{}, error) {
	delete(payload, "id")
	delete(payload, "created_at")
	delete(payload, "deleted_at")
	payload["updated_at"] = time.Now()
	switch resource {
	case "providers":
		return updateRow[models.AIProvider](ctx, s, userID, resource, id, payload, "ai_provider")
	case "accounts":
		return updateRow[models.AIProviderAccount](ctx, s, userID, resource, id, payload, "ai_provider_account")
	case "apis":
		return updateRow[models.AIProviderAPI](ctx, s, userID, resource, id, payload, "ai_provider_api")
	case "capabilities":
		return updateRow[models.AICapability](ctx, s, userID, resource, id, payload, "ai_capability")
	case "models":
		return updateRow[models.AIModel](ctx, s, userID, resource, id, payload, "ai_model")
	case "price-policies":
		return updateRow[models.AIModelPricePolicy](ctx, s, userID, resource, id, payload, "ai_model_price_policy")
	case "price-tiers":
		return updateRow[models.AIModelPriceTier](ctx, s, userID, resource, id, payload, "ai_model_price_tier")
	case "base-routes":
		return updateRow[models.AIBaseRoute](ctx, s, userID, resource, id, payload, "ai_base_route")
	case "route-models":
		return updateRow[models.AIBaseRouteModel](ctx, s, userID, resource, id, payload, "ai_base_route_model")
	case "scenarios":
		return updateRow[models.AIScenario](ctx, s, userID, resource, id, payload, "ai_scenario")
	case "tenant-strategies":
		return updateRow[models.AITenantStrategyPolicy](ctx, s, userID, resource, id, payload, "ai_tenant_strategy")
	case "quota-rules":
		return updateRow[models.AIStrategyQuotaRule](ctx, s, userID, resource, id, payload, "ai_strategy_quota_rule")
	case "rate-limit-rules":
		return updateRow[models.AIStrategyRateLimitRule](ctx, s, userID, resource, id, payload, "ai_strategy_rate_limit_rule")
	case "settings":
		return updateRow[models.AIGatewaySetting](ctx, s, userID, resource, id, payload, "ai_gateway_setting")
	default:
		return nil, ErrNotFound
	}
}

func (s *Service) DeleteResource(ctx context.Context, userID uint64, resource, id string) error {
	now := time.Now()
	var model interface{}
	module := resource
	switch resource {
	case "providers":
		if s.hasReferences(ctx, &models.AIModel{}, "provider_id = ?", id) || s.hasAnyReferences(ctx, &models.AIUsageRecord{}, "provider_id = ?", id) {
			return ErrResourceInUse
		}
		module = "ai_provider"
		if err := s.deleteProviderCascade(ctx, userID, id, now); err != nil {
			return err
		}
		s.Audit(ctx, userID, module, "delete", "删除 AI 供应商："+id, map[string]string{"id": id})
		return nil
	case "accounts":
		module = "ai_provider_account"
		if s.hasAnyReferences(ctx, &models.AIUsageRecord{}, "provider_account_id = ?", id) {
			return ErrResourceInUse
		}
		if err := s.deleteProviderAccountCascade(ctx, id, now); err != nil {
			return err
		}
		s.Audit(ctx, userID, module, "delete", "删除 AI 供应商账号："+id, map[string]string{"id": id})
		return nil
	case "apis":
		if s.hasAnyReferences(ctx, &models.AIUsageRecord{}, "provider_api_id = ?", id) {
			return ErrResourceInUse
		}
		model = &models.AIProviderAPI{}
		module = "ai_provider_api"
	case "capabilities":
		var cap models.AICapability
		if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&cap).Error; err != nil {
			return ErrNotFound
		}
		if s.hasReferences(ctx, &models.AIScenario{}, "capability_code = ?", cap.CapabilityCode) ||
			s.hasReferences(ctx, &models.AIBaseRoute{}, "capability_code = ?", cap.CapabilityCode) ||
			s.hasReferences(ctx, &models.AIModelPricePolicy{}, "capability_code = ?", cap.CapabilityCode) {
			return ErrResourceInUse
		}
		model = &models.AICapability{}
		module = "ai_capability"
	case "models":
		if s.hasReferences(ctx, &models.AIBaseRouteModel{}, "model_id = ?", id) || s.hasAnyReferences(ctx, &models.AIUsageRecord{}, "model_id = ?", id) {
			return ErrResourceInUse
		}
		model = &models.AIModel{}
		module = "ai_model"
	case "price-policies":
		model = &models.AIModelPricePolicy{}
		module = "ai_model_price_policy"
	case "price-tiers":
		model = &models.AIModelPriceTier{}
		module = "ai_model_price_tier"
	case "base-routes":
		if s.hasReferences(ctx, &models.AIScenario{}, "default_base_route_id = ?", id) ||
			s.hasReferences(ctx, &models.AITenantStrategyPolicy{}, "default_base_route_id = ? OR override_base_route_id = ?", id, id) {
			return ErrResourceInUse
		}
		module = "ai_base_route"
		if err := s.deleteBaseRouteCascade(ctx, id, now); err != nil {
			return err
		}
		s.Audit(ctx, userID, module, "delete", "删除基础路由："+id, map[string]string{"id": id})
		return nil
	case "route-models":
		model = &models.AIBaseRouteModel{}
		module = "ai_base_route_model"
	case "scenarios":
		var scenario models.AIScenario
		if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&scenario).Error; err != nil {
			return ErrNotFound
		}
		if s.hasAnyReferences(ctx, &models.AIUsageRecord{}, "app_code = ? AND ai_scenario_code = ?", scenario.AppCode, scenario.AIScenarioCode) ||
			s.hasReferences(ctx, &models.AITenantStrategyPolicy{}, "app_code = ? AND ai_scenario_code = ?", scenario.AppCode, scenario.AIScenarioCode) {
			return ErrResourceInUse
		}
		model = &models.AIScenario{}
		module = "ai_scenario"
	case "tenant-strategies":
		model = &models.AITenantStrategyPolicy{}
		module = "ai_tenant_strategy"
	case "quota-rules":
		model = &models.AIStrategyQuotaRule{}
		module = "ai_strategy_quota_rule"
	case "rate-limit-rules":
		model = &models.AIStrategyRateLimitRule{}
		module = "ai_strategy_rate_limit_rule"
	case "settings":
		model = &models.AIGatewaySetting{}
		module = "ai_gateway_setting"
	default:
		return ErrNotFound
	}
	res := s.db.WithContext(ctx).Model(model).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	s.Audit(ctx, userID, module, "delete", "删除 AI 能力中心资源："+id, map[string]string{"id": id})
	return nil
}

func listRows[T any](ctx context.Context, db *gorm.DB, skip, limit int, keyword string, columns []string) (PageResult, error) {
	q := db.WithContext(ctx).Model(new(T)).Where("deleted_at IS NULL")
	q = keywordQuery(q, keyword, columns)
	return pageQuery[T](q, skip, limit)
}

func keywordQuery(q *gorm.DB, keyword string, columns []string) *gorm.DB {
	if strings.TrimSpace(keyword) != "" && len(columns) > 0 {
		like := "%" + strings.TrimSpace(keyword) + "%"
		parts := make([]string, 0, len(columns))
		args := make([]interface{}, 0, len(columns))
		for _, column := range columns {
			parts = append(parts, column+" ILIKE ?")
			args = append(args, like)
		}
		q = q.Where(strings.Join(parts, " OR "), args...)
	}
	return q
}

func usageDateQuery(q *gorm.DB, startDate, endDate string) (*gorm.DB, error) {
	if strings.TrimSpace(startDate) != "" {
		start, err := time.Parse("2006-01-02", strings.TrimSpace(startDate))
		if err != nil {
			return nil, ErrInvalidInput
		}
		q = q.Where("called_at >= ?", start)
	}
	if strings.TrimSpace(endDate) != "" {
		end, err := time.Parse("2006-01-02", strings.TrimSpace(endDate))
		if err != nil {
			return nil, ErrInvalidInput
		}
		q = q.Where("called_at < ?", end.AddDate(0, 0, 1))
	}
	return q, nil
}

func decodePayload(payload map[string]interface{}, target interface{}) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func upsertProviderImport(ctx context.Context, tx *gorm.DB, item ProviderImportProvider, now time.Time) (models.AIProvider, error) {
	row := models.AIProvider{
		Name:          strings.TrimSpace(item.Name),
		Code:          strings.TrimSpace(item.Code),
		Type:          defaultString(item.Type, "public_cloud"),
		BaseURL:       strings.TrimSpace(item.BaseURL),
		AuthType:      defaultString(item.AuthType, "api_key"),
		Status:        defaultString(item.Status, "active"),
		Priority:      item.Priority,
		Region:        strings.TrimSpace(item.Region),
		QPSLimit:      item.QPSLimit,
		MonthlyBudget: item.MonthlyBudget,
		Owner:         strings.TrimSpace(item.Owner),
		Remark:        strings.TrimSpace(item.Remark),
	}
	if err := validateProvider(row); err != nil {
		return models.AIProvider{}, err
	}
	var existing models.AIProvider
	err := tx.WithContext(ctx).Where("code = ? AND deleted_at IS NULL", row.Code).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return row, tx.Create(&row).Error
	}
	if err != nil {
		return models.AIProvider{}, err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return row, tx.Model(&existing).Updates(map[string]interface{}{
		"name": row.Name, "type": row.Type, "base_url": row.BaseURL, "auth_type": row.AuthType,
		"status": row.Status, "priority": row.Priority, "region": row.Region, "qps_limit": row.QPSLimit,
		"monthly_budget": row.MonthlyBudget, "owner": row.Owner, "remark": row.Remark, "updated_at": now,
	}).Error
}

func upsertProviderAccountImport(ctx context.Context, tx *gorm.DB, providerIDs map[string]string, item ProviderImportAccount, now time.Time) (models.AIProviderAccount, error) {
	providerCode := strings.TrimSpace(item.ProviderCode)
	providerID := providerIDs[providerCode]
	if providerID == "" || strings.TrimSpace(item.AccountName) == "" || strings.TrimSpace(item.KeyAlias) == "" {
		return models.AIProviderAccount{}, ErrInvalidInput
	}
	row := models.AIProviderAccount{
		ProviderID:      providerID,
		AccountName:     strings.TrimSpace(item.AccountName),
		Endpoint:        strings.TrimSpace(item.Endpoint),
		KeyAlias:        strings.TrimSpace(item.KeyAlias),
		EncryptedAPIKey: strings.TrimSpace(item.EncryptedAPIKey),
		EncryptedSecret: strings.TrimSpace(item.EncryptedSecret),
		QuotaLimit:      item.QuotaLimit,
		UsedQuota:       item.UsedQuota,
		Status:          defaultString(item.Status, "active"),
	}
	var existing models.AIProviderAccount
	err := tx.WithContext(ctx).Where("provider_id = ? AND account_name = ? AND deleted_at IS NULL", providerID, row.AccountName).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return row, tx.Create(&row).Error
	}
	if err != nil {
		return models.AIProviderAccount{}, err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return row, tx.Model(&existing).Updates(map[string]interface{}{
		"endpoint": row.Endpoint, "key_alias": row.KeyAlias, "encrypted_api_key": row.EncryptedAPIKey,
		"encrypted_secret": row.EncryptedSecret, "quota_limit": row.QuotaLimit, "used_quota": row.UsedQuota,
		"status": row.Status, "updated_at": now,
	}).Error
}

func upsertProviderAPIImport(ctx context.Context, tx *gorm.DB, providerIDs map[string]string, accountIDs map[string]string, item ProviderImportAPI, now time.Time) error {
	providerCode := strings.TrimSpace(item.ProviderCode)
	providerID := providerIDs[providerCode]
	accountID := accountIDs[accountKey(providerCode, item.AccountName)]
	if providerID == "" || accountID == "" || strings.TrimSpace(item.APIName) == "" || strings.TrimSpace(item.APIPath) == "" || strings.TrimSpace(item.APIType) == "" {
		return ErrInvalidInput
	}
	row := models.AIProviderAPI{
		ProviderID:   providerID,
		AccountID:    accountID,
		APIName:      strings.TrimSpace(item.APIName),
		APIPath:      strings.TrimSpace(item.APIPath),
		APIType:      strings.TrimSpace(item.APIType),
		Capabilities: item.Capabilities,
		AuthType:     defaultString(item.AuthType, "api_key"),
		QPSLimit:     item.QPSLimit,
		TimeoutMS:    item.TimeoutMS,
		Status:       defaultString(item.Status, "active"),
		LastCalledAt: nil,
	}
	if row.TimeoutMS <= 0 {
		row.TimeoutMS = 30000
	}
	if len(row.Capabilities) == 0 {
		return ErrInvalidInput
	}
	var existing models.AIProviderAPI
	err := tx.WithContext(ctx).Where("provider_id = ? AND account_id = ? AND api_name = ? AND deleted_at IS NULL", providerID, accountID, row.APIName).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return tx.Create(&row).Error
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return tx.Model(&existing).
		Select("api_path", "api_type", "capabilities", "auth_type", "qps_limit", "timeout_ms", "status", "updated_at").
		Updates(row).Error
}

func loadProviderIDs(ctx context.Context, tx *gorm.DB, providerIDs map[string]string) error {
	var rows []models.AIProvider
	if err := tx.WithContext(ctx).Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		providerIDs[row.Code] = row.ID
	}
	return nil
}

func loadAccountIDs(ctx context.Context, tx *gorm.DB, accountIDs map[string]string) error {
	var rows []struct {
		ProviderCode string
		AccountName  string
		ID           string
	}
	if err := tx.WithContext(ctx).Table("ai_provider_accounts AS a").
		Select("p.code AS provider_code, a.account_name, a.id").
		Joins("JOIN ai_providers AS p ON p.id = a.provider_id").
		Where("a.deleted_at IS NULL AND p.deleted_at IS NULL").
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		accountIDs[accountKey(row.ProviderCode, row.AccountName)] = row.ID
	}
	return nil
}

func upsertModelImport(ctx context.Context, tx *gorm.DB, providerIDs map[string]string, item ModelImportModel, now time.Time) (models.AIModel, error) {
	providerCode := strings.TrimSpace(item.ProviderCode)
	providerID := providerIDs[providerCode]
	row := models.AIModel{
		ProviderID:    providerID,
		ModelCode:     strings.TrimSpace(item.ModelCode),
		ModelName:     strings.TrimSpace(item.ModelName),
		ModelType:     strings.TrimSpace(item.ModelType),
		Capabilities:  item.Capabilities,
		ContextWindow: item.ContextWindow,
		Unit:          defaultString(item.Unit, "tokens"),
		LatencyP95:    item.LatencyP95,
		SuccessRate:   item.SuccessRate,
		Status:        defaultString(item.Status, "active"),
		DefaultFor:    item.DefaultFor,
		Remark:        strings.TrimSpace(item.Remark),
	}
	if providerID == "" || row.ModelCode == "" || row.ModelName == "" || row.ModelType == "" || len(row.Capabilities) == 0 {
		return models.AIModel{}, ErrInvalidInput
	}
	var existing models.AIModel
	err := tx.WithContext(ctx).Where("provider_id = ? AND model_code = ? AND deleted_at IS NULL", providerID, row.ModelCode).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return row, tx.Create(&row).Error
	}
	if err != nil {
		return models.AIModel{}, err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return row, tx.Model(&existing).
		Select("model_name", "model_type", "capabilities", "context_window", "unit", "latency_p95", "success_rate", "status", "default_for", "remark", "updated_at").
		Updates(row).Error
}

func upsertModelPricePolicyImport(ctx context.Context, tx *gorm.DB, modelIDs map[string]string, item ModelImportPricePolicy, now time.Time) (models.AIModelPricePolicy, error) {
	modelID := modelIDs[modelKey(item.ProviderCode, item.ModelCode)]
	row := models.AIModelPricePolicy{
		ModelID:            modelID,
		FeatureKey:         strings.TrimSpace(item.FeatureKey),
		FeatureName:        strings.TrimSpace(item.FeatureName),
		ModelType:          strings.TrimSpace(item.ModelType),
		CapabilityCode:     strings.TrimSpace(item.CapabilityCode),
		BillingMode:        defaultString(item.BillingMode, "per_unit"),
		BillingUnit:        defaultString(item.BillingUnit, "tokens"),
		PlatformUnit:       defaultString(item.PlatformUnit, defaultString(item.BillingUnit, "tokens")),
		BaseCostPrice:      item.BaseCostPrice,
		BaseSalePrice:      item.BaseSalePrice,
		BasePlatformAmount: item.BasePlatformAmount,
		Currency:           defaultString(item.Currency, "CNY"),
		Status:             defaultString(item.Status, "active"),
	}
	if modelID == "" || row.FeatureKey == "" || row.FeatureName == "" || row.ModelType == "" || row.CapabilityCode == "" {
		return models.AIModelPricePolicy{}, ErrInvalidInput
	}
	if !capabilityExistsTx(ctx, tx, row.CapabilityCode) {
		return models.AIModelPricePolicy{}, ErrInvalidInput
	}
	var existing models.AIModelPricePolicy
	err := tx.WithContext(ctx).Where("model_id = ? AND feature_key = ? AND deleted_at IS NULL", modelID, row.FeatureKey).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return row, tx.Create(&row).Error
	}
	if err != nil {
		return models.AIModelPricePolicy{}, err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return row, tx.Model(&existing).Updates(map[string]interface{}{
		"feature_name": row.FeatureName, "model_type": row.ModelType, "capability_code": row.CapabilityCode,
		"billing_mode": row.BillingMode, "billing_unit": row.BillingUnit, "platform_unit": row.PlatformUnit,
		"base_cost_price": row.BaseCostPrice, "base_sale_price": row.BaseSalePrice, "base_platform_amount": row.BasePlatformAmount,
		"currency": row.Currency, "status": row.Status, "updated_at": now,
	}).Error
}

func upsertModelPriceTierImport(ctx context.Context, tx *gorm.DB, policyIDs map[string]string, item ModelImportPriceTier, now time.Time) error {
	policyID := policyIDs[policyKey(item.ProviderCode, item.ModelCode, item.FeatureKey)]
	enabled := true
	if item.Enabled != nil {
		enabled = *item.Enabled
	}
	row := models.AIModelPriceTier{
		PricePolicyID:   policyID,
		TierName:        strings.TrimSpace(item.TierName),
		Mode:            strings.TrimSpace(item.Mode),
		Resolution:      strings.TrimSpace(item.Resolution),
		Quality:         strings.TrimSpace(item.Quality),
		DurationSeconds: item.DurationSeconds,
		AspectRatio:     strings.TrimSpace(item.AspectRatio),
		CostPrice:       item.CostPrice,
		SalePrice:       item.SalePrice,
		PlatformAmount:  item.PlatformAmount,
		Enabled:         enabled,
		SortOrder:       item.SortOrder,
	}
	if policyID == "" || row.TierName == "" {
		return ErrInvalidInput
	}
	var existing models.AIModelPriceTier
	err := tx.WithContext(ctx).Where("price_policy_id = ? AND tier_name = ? AND deleted_at IS NULL", policyID, row.TierName).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		desiredEnabled := row.Enabled
		if err := tx.Select("*").Create(&row).Error; err != nil {
			return err
		}
		if !desiredEnabled {
			return tx.Exec("UPDATE ai_model_price_tiers SET enabled = ? WHERE id = ?", false, row.ID).Error
		}
		return nil
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return tx.Model(&existing).Updates(map[string]interface{}{
		"mode": row.Mode, "resolution": row.Resolution, "quality": row.Quality, "duration_seconds": row.DurationSeconds,
		"aspect_ratio": row.AspectRatio, "cost_price": row.CostPrice, "sale_price": row.SalePrice,
		"platform_amount": row.PlatformAmount, "enabled": row.Enabled, "sort_order": row.SortOrder, "updated_at": now,
	}).Error
}

func loadModelIDs(ctx context.Context, tx *gorm.DB, modelIDs map[string]string) error {
	var rows []struct {
		ProviderCode string
		ModelCode    string
		ID           string
	}
	if err := tx.WithContext(ctx).Table("ai_models AS m").
		Select("p.code AS provider_code, m.model_code, m.id").
		Joins("JOIN ai_providers AS p ON p.id = m.provider_id").
		Where("m.deleted_at IS NULL AND p.deleted_at IS NULL").
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		modelIDs[modelKey(row.ProviderCode, row.ModelCode)] = row.ID
	}
	return nil
}

func loadPolicyIDs(ctx context.Context, tx *gorm.DB, policyIDs map[string]string) error {
	var rows []struct {
		ProviderCode string
		ModelCode    string
		FeatureKey   string
		ID           string
	}
	if err := tx.WithContext(ctx).Table("ai_model_price_policies AS pp").
		Select("p.code AS provider_code, m.model_code, pp.feature_key, pp.id").
		Joins("JOIN ai_models AS m ON m.id = pp.model_id").
		Joins("JOIN ai_providers AS p ON p.id = m.provider_id").
		Where("pp.deleted_at IS NULL AND m.deleted_at IS NULL AND p.deleted_at IS NULL").
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		policyIDs[policyKey(row.ProviderCode, row.ModelCode, row.FeatureKey)] = row.ID
	}
	return nil
}

func loadBaseRouteIDs(ctx context.Context, tx *gorm.DB, routeIDs map[string]string) error {
	var rows []models.AIBaseRoute
	if err := tx.WithContext(ctx).Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		routeIDs[row.RouteCode] = row.ID
	}
	return nil
}

func upsertBaseRouteImport(ctx context.Context, tx *gorm.DB, item RouteImportBaseRoute, now time.Time) (models.AIBaseRoute, error) {
	row := models.AIBaseRoute{
		RouteCode:      strings.TrimSpace(item.RouteCode),
		RouteName:      strings.TrimSpace(item.RouteName),
		CapabilityCode: strings.TrimSpace(item.CapabilityCode),
		ModelType:      strings.TrimSpace(item.ModelType),
		Strategy:       defaultString(item.Strategy, "fallback"),
		TimeoutMS:      defaultInt(item.TimeoutMS, 30000),
		MaxRetry:       item.MaxRetry,
		Description:    strings.TrimSpace(item.Description),
		Status:         defaultString(item.Status, "active"),
	}
	if row.RouteCode == "" || row.RouteName == "" || row.CapabilityCode == "" || row.ModelType == "" || !validRouteStrategy(row.Strategy) {
		return models.AIBaseRoute{}, ErrInvalidInput
	}
	if row.TimeoutMS <= 0 || row.MaxRetry < 0 || !capabilityExistsTx(ctx, tx, row.CapabilityCode) {
		return models.AIBaseRoute{}, ErrInvalidInput
	}
	var existing models.AIBaseRoute
	err := tx.WithContext(ctx).Where("route_code = ? AND deleted_at IS NULL", row.RouteCode).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := tx.Create(&row).Error; err != nil {
			return models.AIBaseRoute{}, err
		}
		return row, nil
	}
	if err != nil {
		return models.AIBaseRoute{}, err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	err = tx.Model(&existing).Updates(map[string]interface{}{
		"route_name": row.RouteName, "capability_code": row.CapabilityCode, "model_type": row.ModelType,
		"strategy": row.Strategy, "timeout_ms": row.TimeoutMS, "max_retry": row.MaxRetry,
		"description": row.Description, "status": row.Status, "updated_at": now,
	}).Error
	return row, err
}

func upsertRouteModelImport(ctx context.Context, tx *gorm.DB, routeIDs, modelIDs map[string]string, item RouteImportRouteModel, now time.Time) error {
	baseRouteID := strings.TrimSpace(item.BaseRouteID)
	if baseRouteID == "" {
		baseRouteID = routeIDs[strings.TrimSpace(item.BaseRouteCode)]
	}
	modelID := strings.TrimSpace(item.ModelID)
	if modelID == "" {
		modelID = modelIDs[modelKey(item.ProviderCode, item.ModelCode)]
	}
	row := models.AIBaseRouteModel{
		BaseRouteID: baseRouteID,
		ModelID:     modelID,
		Role:        defaultString(item.Role, "candidate"),
		Priority:    defaultInt(item.Priority, 1),
		Weight:      defaultInt(item.Weight, 100),
		MaxRetry:    item.MaxRetry,
		TimeoutMS:   defaultInt(item.TimeoutMS, 30000),
		Status:      defaultString(item.Status, "active"),
	}
	if row.BaseRouteID == "" || row.ModelID == "" || !validRouteModelRole(row.Role) ||
		row.Priority <= 0 || row.Weight <= 0 || row.MaxRetry < 0 || row.TimeoutMS <= 0 {
		return ErrInvalidInput
	}
	if !existsTx(ctx, tx, &models.AIBaseRoute{}, row.BaseRouteID) || !existsTx(ctx, tx, &models.AIModel{}, row.ModelID) {
		return ErrInvalidInput
	}
	var existing models.AIBaseRouteModel
	err := tx.WithContext(ctx).Where("base_route_id = ? AND model_id = ? AND role = ? AND deleted_at IS NULL", row.BaseRouteID, row.ModelID, row.Role).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return tx.Create(&row).Error
	}
	if err != nil {
		return err
	}
	return tx.Model(&existing).Updates(map[string]interface{}{
		"priority": row.Priority, "weight": row.Weight, "max_retry": row.MaxRetry,
		"timeout_ms": row.TimeoutMS, "status": row.Status, "updated_at": now,
	}).Error
}

func upsertScenarioImport(ctx context.Context, tx *gorm.DB, item ScenarioImportItem, now time.Time) error {
	row := models.AIScenario{
		AppCode:            strings.TrimSpace(item.AppCode),
		AppName:            strings.TrimSpace(item.AppName),
		AIScenarioCode:     strings.TrimSpace(item.AIScenarioCode),
		AIScenarioName:     strings.TrimSpace(item.AIScenarioName),
		ScenarioType:       strings.TrimSpace(item.ScenarioType),
		CapabilityCode:     strings.TrimSpace(item.CapabilityCode),
		ModelType:          strings.TrimSpace(item.ModelType),
		DefaultBaseRouteID: strings.TrimSpace(item.DefaultBaseRouteID),
		Owner:              strings.TrimSpace(item.Owner),
		Description:        strings.TrimSpace(item.Description),
		Version:            defaultString(item.Version, "v1.0"),
		Status:             defaultString(item.Status, "active"),
	}
	if row.AppCode == "" || row.AppName == "" || row.AIScenarioCode == "" || row.AIScenarioName == "" ||
		row.ScenarioType == "" || row.CapabilityCode == "" || row.ModelType == "" || row.DefaultBaseRouteID == "" {
		return ErrInvalidInput
	}
	if !capabilityExistsTx(ctx, tx, row.CapabilityCode) || !existsTx(ctx, tx, &models.AIBaseRoute{}, row.DefaultBaseRouteID) {
		return ErrInvalidInput
	}
	var existing models.AIScenario
	err := tx.WithContext(ctx).Where("app_code = ? AND ai_scenario_code = ? AND deleted_at IS NULL", row.AppCode, row.AIScenarioCode).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		return tx.Create(&row).Error
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	return tx.Model(&existing).Updates(map[string]interface{}{
		"app_name": row.AppName, "ai_scenario_name": row.AIScenarioName, "scenario_type": row.ScenarioType,
		"capability_code": row.CapabilityCode, "model_type": row.ModelType, "default_base_route_id": row.DefaultBaseRouteID,
		"owner": row.Owner, "description": row.Description, "version": row.Version, "status": row.Status, "updated_at": now,
	}).Error
}

func validateProvider(row models.AIProvider) error {
	if strings.TrimSpace(row.Name) == "" || strings.TrimSpace(row.Code) == "" || strings.TrimSpace(row.BaseURL) == "" {
		return ErrInvalidInput
	}
	return nil
}

func accountKey(providerCode, accountName string) string {
	return strings.TrimSpace(providerCode) + "\x00" + strings.TrimSpace(accountName)
}

func modelKey(providerCode, modelCode string) string {
	return strings.TrimSpace(providerCode) + "\x00" + strings.TrimSpace(modelCode)
}

func policyKey(providerCode, modelCode, featureKey string) string {
	return modelKey(providerCode, modelCode) + "\x00" + strings.TrimSpace(featureKey)
}

func (s *Service) deleteProviderCascade(ctx context.Context, userID uint64, id string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider models.AIProvider
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&provider).Error; err != nil {
			return ErrNotFound
		}
		if err := tx.Model(&models.AIProviderAPI{}).Where("provider_id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AIProviderAccount{}).Where("provider_id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"}).Error; err != nil {
			return err
		}
		res := tx.Model(&models.AIProvider{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Service) deleteProviderAccountCascade(ctx context.Context, id string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var account models.AIProviderAccount
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&account).Error; err != nil {
			return ErrNotFound
		}
		if err := tx.Model(&models.AIProviderAPI{}).Where("account_id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"}).Error; err != nil {
			return err
		}
		res := tx.Model(&models.AIProviderAccount{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Service) deleteBaseRouteCascade(ctx context.Context, id string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var route models.AIBaseRoute
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&route).Error; err != nil {
			return ErrNotFound
		}
		if err := tx.Model(&models.AIBaseRouteModel{}).Where("base_route_id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"}).Error; err != nil {
			return err
		}
		res := tx.Model(&models.AIBaseRoute{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Service) exists(ctx context.Context, model interface{}, id string) bool {
	if strings.TrimSpace(id) == "" {
		return false
	}
	var count int64
	_ = s.db.WithContext(ctx).Model(model).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0
}

func (s *Service) capabilityExists(ctx context.Context, code string) bool {
	if strings.TrimSpace(code) == "" {
		return false
	}
	var count int64
	_ = s.db.WithContext(ctx).Model(&models.AICapability{}).Where("capability_code = ? AND deleted_at IS NULL", code).Count(&count).Error
	return count > 0
}

func capabilityExistsTx(ctx context.Context, tx *gorm.DB, code string) bool {
	if strings.TrimSpace(code) == "" {
		return false
	}
	var count int64
	_ = tx.WithContext(ctx).Model(&models.AICapability{}).Where("capability_code = ? AND deleted_at IS NULL", strings.TrimSpace(code)).Count(&count).Error
	return count > 0
}

func existsTx(ctx context.Context, tx *gorm.DB, model interface{}, id string) bool {
	if strings.TrimSpace(id) == "" {
		return false
	}
	var count int64
	_ = tx.WithContext(ctx).Model(model).Where("id = ? AND deleted_at IS NULL", strings.TrimSpace(id)).Count(&count).Error
	return count > 0
}

func (s *Service) scenarioExists(ctx context.Context, appCode, scenarioCode string) bool {
	if strings.TrimSpace(appCode) == "" || strings.TrimSpace(scenarioCode) == "" {
		return false
	}
	var count int64
	_ = s.db.WithContext(ctx).Model(&models.AIScenario{}).Where("app_code = ? AND ai_scenario_code = ? AND deleted_at IS NULL", appCode, scenarioCode).Count(&count).Error
	return count > 0
}

func (s *Service) hasReferences(ctx context.Context, model interface{}, query string, args ...interface{}) bool {
	var count int64
	_ = s.db.WithContext(ctx).Model(model).Where("deleted_at IS NULL").Where(query, args...).Count(&count).Error
	return count > 0
}

func (s *Service) hasAnyReferences(ctx context.Context, model interface{}, query string, args ...interface{}) bool {
	var count int64
	_ = s.db.WithContext(ctx).Model(model).Where(query, args...).Count(&count).Error
	return count > 0
}

func updateRow[T any](ctx context.Context, s *Service, userID uint64, resource, id string, payload map[string]interface{}, module string) (interface{}, error) {
	var row T
	res := s.db.WithContext(ctx).Model(new(T)).Where("id = ? AND deleted_at IS NULL", id).Updates(payload)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, err
	}
	s.Audit(ctx, userID, module, "update", "更新 AI 能力中心资源："+resource, map[string]interface{}{"id": id, "patch": payload})
	return row, nil
}

func pageQuery[T any](q *gorm.DB, skip, limit int) (PageResult, error) {
	return pageQueryOrder[T](q, skip, limit, "updated_at desc")
}

func pageQueryOrder[T any](q *gorm.DB, skip, limit int, order string) (PageResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if skip < 0 {
		skip = 0
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	var rows []T
	if err := q.Order(order).Offset(skip).Limit(limit).Find(&rows).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: rows, Total: total, Skip: skip, Limit: limit}, nil
}

func usageUnitForCapability(ctx context.Context, db *gorm.DB, code string) string {
	var cap models.AICapability
	if err := db.WithContext(ctx).Where("capability_code = ? AND deleted_at IS NULL", code).First(&cap).Error; err == nil && cap.DefaultBillingUnit != "" {
		return cap.DefaultBillingUnit
	}
	return "requests"
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func validRouteStrategy(strategy string) bool {
	switch strings.TrimSpace(strategy) {
	case "fixed", "fallback", "priority", "load_balance", "cost_first", "quality_first", "latency_first", "quota_aware", "tenant_custom", "capability_match":
		return true
	default:
		return false
	}
}

func validRouteModelRole(role string) bool {
	switch strings.TrimSpace(role) {
	case "primary", "fallback", "candidate":
		return true
	default:
		return false
	}
}

func (s *Service) Audit(ctx context.Context, userID uint64, module, action, summary string, detail interface{}) {
	raw, _ := json.Marshal(detail)
	appCode := ai_capability_center.AppCode
	now := time.Now()
	_ = s.db.WithContext(ctx).Create(&models.AuditLog{
		UserID:    &userID,
		AppCode:   &appCode,
		Module:    module,
		Action:    action,
		Summary:   summary,
		Detail:    stringPtr(string(raw)),
		Result:    "success",
		CreatedAt: now,
	}).Error
}

func stringPtr(value string) *string { return &value }
