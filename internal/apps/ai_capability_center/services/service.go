package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

func (s *Service) StartProviderAPIConnectivityProbe(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	go func() {
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				_, _ = s.CheckProviderAPIConnectivity(ctx, 0, APIConnectivityFilter{})
				timer.Reset(interval)
			}
		}
	}()
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
	ProviderCode      string              `json:"provider_code"`
	AccountName       string              `json:"account_name"`
	Endpoint          string              `json:"endpoint"`
	KeyAlias          string              `json:"key_alias"`
	LoginMethod       string              `json:"login_method"`
	LoginAccount      string              `json:"login_account"`
	Maintainer        string              `json:"maintainer"`
	MaintainerContact string              `json:"maintainer_contact"`
	EncryptedAPIKey   string              `json:"encrypted_api_key"`
	EncryptedSecret   string              `json:"encrypted_secret"`
	QuotaLimit        float64             `json:"quota_limit"`
	UsedQuota         float64             `json:"used_quota"`
	Status            string              `json:"status"`
	APIs              []ProviderImportAPI `json:"apis"`
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
	BaseRouteID       string `json:"base_route_id"`
	BaseRouteCode     string `json:"base_route_code"`
	ModelID           string `json:"model_id"`
	ProviderCode      string `json:"provider_code"`
	ModelCode         string `json:"model_code"`
	ProviderAccountID string `json:"provider_account_id"`
	ProviderAPIID     string `json:"provider_api_id"`
	Role              string `json:"role"`
	Priority          int    `json:"priority"`
	Weight            int    `json:"weight"`
	MaxRetry          int    `json:"max_retry"`
	TimeoutMS         int    `json:"timeout_ms"`
	Status            string `json:"status"`
}

type RouteImportResult struct {
	BaseRoutes  int `json:"base_routes"`
	RouteModels int `json:"route_models"`
}

type TenantStrategyImportRequest struct {
	Policies       []TenantStrategyImportPolicy        `json:"policies"`
	QuotaRules     []TenantStrategyImportQuotaRule     `json:"quota_rules"`
	RateLimitRules []TenantStrategyImportRateLimitRule `json:"rate_limit_rules"`
}

type TenantStrategyImportPolicy struct {
	PolicyName          string                              `json:"policy_name"`
	TenantScope         string                              `json:"tenant_scope"`
	TenantIDs           []string                            `json:"tenant_ids"`
	AppCode             string                              `json:"app_code"`
	AppName             string                              `json:"app_name"`
	AIScenarioCode      string                              `json:"ai_scenario_code"`
	AIScenarioName      string                              `json:"ai_scenario_name"`
	DefaultBaseRouteID  string                              `json:"default_base_route_id"`
	OverrideBaseRouteID string                              `json:"override_base_route_id"`
	Description         string                              `json:"description"`
	Status              string                              `json:"status"`
	QuotaRules          []TenantStrategyImportQuotaRule     `json:"quota_rules"`
	RateLimitRules      []TenantStrategyImportRateLimitRule `json:"rate_limit_rules"`
}

type TenantStrategyImportQuotaRule struct {
	PolicyID         string  `json:"policy_id"`
	PolicyName       string  `json:"policy_name"`
	Dimension        string  `json:"dimension"`
	SubjectCode      string  `json:"subject_code"`
	UsageUnit        string  `json:"usage_unit"`
	Period           string  `json:"period"`
	QuotaLimit       float64 `json:"quota_limit"`
	UsedAmount       float64 `json:"used_amount"`
	WarningThreshold float64 `json:"warning_threshold"`
	OverLimitAction  string  `json:"over_limit_action"`
	Status           string  `json:"status"`
}

type TenantStrategyImportRateLimitRule struct {
	PolicyID        string `json:"policy_id"`
	PolicyName      string `json:"policy_name"`
	Dimension       string `json:"dimension"`
	SubjectCode     string `json:"subject_code"`
	QPS             int    `json:"qps"`
	Concurrency     int    `json:"concurrency"`
	MinuteLimit     int    `json:"minute_limit"`
	HourLimit       int    `json:"hour_limit"`
	DayLimit        int    `json:"day_limit"`
	OverLimitAction string `json:"over_limit_action"`
	Status          string `json:"status"`
}

type TenantStrategyImportResult struct {
	Policies       int `json:"policies"`
	QuotaRules     int `json:"quota_rules"`
	RateLimitRules int `json:"rate_limit_rules"`
}

type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type APIConnectivityResult struct {
	Total   int `json:"total"`
	Active  int `json:"active"`
	Warning int `json:"warning"`
	Error   int `json:"error"`
}

type APIConnectivityFilter struct {
	ProviderID string   `json:"provider_id"`
	AccountID  string   `json:"account_id"`
	APIIDs     []string `json:"api_ids"`
}

type apiProbeRow struct {
	ID              string
	ProviderCode    string
	APIName         string
	APIPath         string
	APIType         string
	Capabilities    []string `gorm:"serializer:json"`
	AuthType        string
	BaseURL         string
	Endpoint        string
	KeyAlias        string
	EncryptedAPIKey string
	EncryptedSecret string
	TimeoutMS       int
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

func (s *Service) ImportTenantStrategies(ctx context.Context, userID uint64, req TenantStrategyImportRequest) (TenantStrategyImportResult, error) {
	now := time.Now()
	result := TenantStrategyImportResult{}
	if len(req.Policies) == 0 && len(req.QuotaRules) == 0 && len(req.RateLimitRules) == 0 {
		return result, ErrInvalidInput
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		policyIDs := map[string]string{}
		if err := loadTenantStrategyPolicyIDs(ctx, tx, policyIDs); err != nil {
			return err
		}
		for _, item := range req.Policies {
			policy, err := upsertTenantStrategyPolicyImport(ctx, tx, item, now)
			if err != nil {
				return err
			}
			result.Policies++
			policyIDs[policy.PolicyName] = policy.ID
			for _, rule := range item.QuotaRules {
				rule.PolicyName = policy.PolicyName
				if err := upsertTenantStrategyQuotaRuleImport(ctx, tx, policyIDs, rule, now); err != nil {
					return err
				}
				result.QuotaRules++
			}
			for _, rule := range item.RateLimitRules {
				rule.PolicyName = policy.PolicyName
				if err := upsertTenantStrategyRateLimitRuleImport(ctx, tx, policyIDs, rule, now); err != nil {
					return err
				}
				result.RateLimitRules++
			}
		}
		if err := loadTenantStrategyPolicyIDs(ctx, tx, policyIDs); err != nil {
			return err
		}
		for _, item := range req.QuotaRules {
			if err := upsertTenantStrategyQuotaRuleImport(ctx, tx, policyIDs, item, now); err != nil {
				return err
			}
			result.QuotaRules++
		}
		for _, item := range req.RateLimitRules {
			if err := upsertTenantStrategyRateLimitRuleImport(ctx, tx, policyIDs, item, now); err != nil {
				return err
			}
			result.RateLimitRules++
		}
		return nil
	})
	if err != nil {
		return TenantStrategyImportResult{}, err
	}
	s.Audit(ctx, userID, "ai_tenant_strategy", "import", "整体导入租户策略和规则", result)
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
	StrategyID  string                 `json:"tenant_strategy_id"`
	Usage       map[string]interface{} `json:"usage"`
	Billing     map[string]interface{} `json:"billing"`
	Controls    map[string]interface{} `json:"controls"`
	Data        map[string]interface{} `json:"data"`
}

type quotaCheckResult struct {
	RuleID          string  `json:"rule_id"`
	Dimension       string  `json:"dimension"`
	SubjectCode     string  `json:"subject_code"`
	UsageUnit       string  `json:"usage_unit"`
	Period          string  `json:"period"`
	Limit           float64 `json:"limit"`
	UsedBefore      float64 `json:"used_before"`
	RequestedAmount float64 `json:"requested_amount"`
	Exceeded        bool    `json:"exceeded"`
	Action          string  `json:"action"`
}

type rateLimitCheckResult struct {
	RuleID      string `json:"rule_id"`
	Dimension   string `json:"dimension"`
	SubjectCode string `json:"subject_code"`
	Window      string `json:"window"`
	Limit       int    `json:"limit"`
	UsedBefore  int64  `json:"used_before"`
	Exceeded    bool   `json:"exceeded"`
	Action      string `json:"action"`
}

type pricingResult struct {
	PolicyID       string
	TierID         string
	UsageAmount    float64
	UsageUnit      string
	CostAmount     float64
	BillingAmount  float64
	PlatformUnit   string
	PlatformAmount float64
	FeatureKey     string
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
	healthChecks, err := s.gatewayHealthChecks(ctx)
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
		HealthChecks:   healthChecks,
		CoreBaseRoutes: routeRows,
	}, nil
}

func (s *Service) CheckProviderAPIConnectivity(ctx context.Context, userID uint64, filter APIConnectivityFilter) (APIConnectivityResult, error) {
	var rows []apiProbeRow
	query := s.db.WithContext(ctx).Table("ai_provider_apis AS api").
		Select("api.id, p.code AS provider_code, api.api_name, api.api_path, api.api_type, api.capabilities, COALESCE(NULLIF(api.auth_type, ''), p.auth_type) AS auth_type, p.base_url, a.endpoint, a.key_alias, a.encrypted_api_key, a.encrypted_secret, api.timeout_ms").
		Joins("JOIN ai_providers AS p ON p.id = api.provider_id AND p.status = ? AND p.deleted_at IS NULL", "active").
		Joins("JOIN ai_provider_accounts AS a ON a.id = api.account_id AND a.status = ? AND a.deleted_at IS NULL", "active").
		Where("api.status = ? AND api.deleted_at IS NULL", "active")
	if strings.TrimSpace(filter.ProviderID) != "" {
		query = query.Where("api.provider_id = ?", strings.TrimSpace(filter.ProviderID))
	}
	if strings.TrimSpace(filter.AccountID) != "" {
		query = query.Where("api.account_id = ?", strings.TrimSpace(filter.AccountID))
	}
	if len(filter.APIIDs) > 0 {
		query = query.Where("api.id IN ?", filter.APIIDs)
	}
	if err := query.Find(&rows).Error; err != nil {
		return APIConnectivityResult{}, err
	}
	now := time.Now()
	result := APIConnectivityResult{Total: len(rows)}
	for _, row := range rows {
		client := &http.Client{Timeout: probeTimeout(row.TimeoutMS)}
		probeURL := providerAPIProbeURL(row.BaseURL, row.Endpoint, row.APIPath)
		status, message := probeProviderAPI(ctx, client, row, probeURL)
		switch status {
		case "active":
			result.Active++
		case "warning":
			result.Warning++
		default:
			result.Error++
		}
		if err := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).
			Where("id = ? AND deleted_at IS NULL", row.ID).
			Updates(map[string]interface{}{
				"health_status":     status,
				"health_message":    message,
				"health_checked_at": now,
				"updated_at":        now,
			}).Error; err != nil {
			return APIConnectivityResult{}, err
		}
	}
	s.Audit(ctx, userID, "ai_provider_api", "connectivity_check", "执行 AI API 连通性检测", map[string]interface{}{"result": result, "filter": filter})
	return result, nil
}

func (s *Service) gatewayHealthChecks(ctx context.Context) ([]HealthCheck, error) {
	providers, err := s.countActive(ctx, &models.AIProvider{}, "")
	if err != nil {
		return nil, err
	}
	accounts, err := s.countActive(ctx, &models.AIProviderAccount{}, "")
	if err != nil {
		return nil, err
	}
	apis, err := s.countActive(ctx, &models.AIProviderAPI{}, "")
	if err != nil {
		return nil, err
	}
	apiHealthStatus, apiHealthMessage, err := s.apiConnectivityHealth(ctx)
	if err != nil {
		return nil, err
	}
	routes, err := s.countActive(ctx, &models.AIBaseRoute{}, "")
	if err != nil {
		return nil, err
	}
	routeModels, err := s.countActive(ctx, &models.AIBaseRouteModel{}, "")
	if err != nil {
		return nil, err
	}
	scenarioStatus, scenarioMessage, err := s.scenarioBindingHealth(ctx)
	if err != nil {
		return nil, err
	}
	strategies, err := s.countActive(ctx, &models.AITenantStrategyPolicy{}, "")
	if err != nil {
		return nil, err
	}
	quotaRules, err := s.countActive(ctx, &models.AIStrategyQuotaRule{}, "")
	if err != nil {
		return nil, err
	}
	rateRules, err := s.countActive(ctx, &models.AIStrategyRateLimitRule{}, "")
	if err != nil {
		return nil, err
	}
	runtimeSettings, err := s.countActive(ctx, &models.AIGatewaySetting{}, "setting_key = ?", "gateway_runtime")
	if err != nil {
		return nil, err
	}
	securitySettings, err := s.countActive(ctx, &models.AIGatewaySetting{}, "setting_key = ?", "security")
	if err != nil {
		return nil, err
	}

	providerStatus := "active"
	providerMessage := fmt.Sprintf("%d 个启用供应商、%d 个启用账号、%d 个启用 API", providers, accounts, apis)
	if providers == 0 || accounts == 0 || apis == 0 {
		providerStatus = "warning"
		providerMessage = "缺少启用供应商、账号或 API，Gateway 无法完成真实供应商调用"
	}

	routeStatus := "active"
	routeMessage := fmt.Sprintf("%d 条启用基础路由、%d 个启用模型节点", routes, routeModels)
	if routes == 0 || routeModels == 0 {
		routeStatus = "warning"
		routeMessage = "缺少启用基础路由或模型池节点，AI 场景无法完成模型选择"
	}

	strategyStatus := "active"
	strategyMessage := fmt.Sprintf("%d 条启用策略、%d 条配额规则、%d 条限流规则", strategies, quotaRules, rateRules)
	if strategies == 0 {
		strategyStatus = "warning"
		strategyMessage = "未配置启用租户策略，Gateway 只能使用场景默认基础路由"
	}

	auditStatus := "active"
	auditMessage := "Gateway 运行参数和安全归属配置已启用，写操作进入底座操作日志"
	if runtimeSettings == 0 || securitySettings == 0 {
		auditStatus = "warning"
		auditMessage = "缺少 gateway_runtime 或 security 设置，请在系统设置中补齐"
	}

	return []HealthCheck{
		{Name: "供应商可用状态", Status: providerStatus, Message: providerMessage},
		{Name: "API 连通性", Status: apiHealthStatus, Message: apiHealthMessage},
		{Name: "基础路由状态", Status: routeStatus, Message: routeMessage},
		{Name: "AI 场景绑定状态", Status: scenarioStatus, Message: scenarioMessage},
		{Name: "租户策略状态", Status: strategyStatus, Message: strategyMessage},
		{Name: "底座操作日志", Status: auditStatus, Message: auditMessage},
	}, nil
}

func (s *Service) apiConnectivityHealth(ctx context.Context) (string, string, error) {
	activeAPIs, err := s.countActive(ctx, &models.AIProviderAPI{}, "")
	if err != nil {
		return "", "", err
	}
	if activeAPIs == 0 {
		return "warning", "未配置启用 API，Gateway 无法完成供应商调用", nil
	}
	staleBefore := time.Now().Add(-15 * time.Minute)
	var stale, warningCount, errorCount int64
	if err := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).
		Where("status = ? AND deleted_at IS NULL", "active").
		Where("health_checked_at IS NULL OR health_checked_at < ?", staleBefore).
		Count(&stale).Error; err != nil {
		return "", "", err
	}
	if err := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).
		Where("status = ? AND deleted_at IS NULL AND health_status = ?", "active", "warning").
		Count(&warningCount).Error; err != nil {
		return "", "", err
	}
	if err := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).
		Where("status = ? AND deleted_at IS NULL AND health_status = ?", "active", "error").
		Count(&errorCount).Error; err != nil {
		return "", "", err
	}
	if errorCount > 0 {
		return "error", fmt.Sprintf("%d 个启用 API 连通性失败，请检查 Endpoint、网络或 API 路径", errorCount), nil
	}
	if warningCount > 0 || stale > 0 {
		return "warning", fmt.Sprintf("%d 个 API 告警，%d 个 API 未在 15 分钟内完成连通性检测", warningCount, stale), nil
	}
	return "active", fmt.Sprintf("%d 个启用 API 最近 15 分钟连通性正常", activeAPIs), nil
}

func (s *Service) scenarioBindingHealth(ctx context.Context) (string, string, error) {
	scenarios, err := s.countActive(ctx, &models.AIScenario{}, "")
	if err != nil {
		return "", "", err
	}
	if scenarios == 0 {
		return "warning", "未注册启用 AI 场景，业务中心无法通过场景编码调用 Gateway", nil
	}
	var rows []models.AIScenario
	if err := s.db.WithContext(ctx).Where("status = ? AND deleted_at IS NULL", "active").Find(&rows).Error; err != nil {
		return "", "", err
	}
	unavailable := int64(0)
	for _, scenario := range rows {
		var route models.AIBaseRoute
		if err := s.db.WithContext(ctx).Where("id = ? AND status = ? AND deleted_at IS NULL", scenario.DefaultBaseRouteID, "active").First(&route).Error; err != nil {
			unavailable++
			continue
		}
		if _, _, _, err := s.selectRouteExecutionPlan(ctx, route, scenario); err != nil {
			unavailable++
		}
	}
	if unavailable > 0 {
		return "warning", fmt.Sprintf("%d 个启用 AI 场景未绑定可用基础路由、模型节点或健康 API", unavailable), nil
	}
	return "active", fmt.Sprintf("%d 个启用 AI 场景已绑定可用基础路由和执行端点", scenarios), nil
}

func (s *Service) countActive(ctx context.Context, model interface{}, where string, args ...interface{}) (int64, error) {
	query := s.db.WithContext(ctx).Model(model).Where("status = ? AND deleted_at IS NULL", "active")
	if where != "" {
		query = query.Where(where, args...)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

func providerAPIProbeURL(baseURL, endpoint, apiPath string) string {
	if raw := strings.TrimSpace(apiPath); strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	base := strings.TrimSpace(endpoint)
	if base == "" {
		base = strings.TrimSpace(baseURL)
	}
	parsedBase, err := url.Parse(base)
	if err != nil || parsedBase.Scheme == "" || parsedBase.Host == "" {
		return ""
	}
	parsedPath, err := url.Parse(strings.TrimSpace(apiPath))
	if err != nil {
		return ""
	}
	return parsedBase.ResolveReference(parsedPath).String()
}

func probeTimeout(timeoutMS int) time.Duration {
	if timeoutMS <= 0 {
		return 3 * time.Second
	}
	timeout := time.Duration(timeoutMS) * time.Millisecond
	if timeout < time.Second {
		return time.Second
	}
	if timeout > 10*time.Second {
		return 10 * time.Second
	}
	return timeout
}

type providerAPIProbe struct {
	Method      string
	ContentType string
	Body        []byte
}

func probeProviderAPI(ctx context.Context, client *http.Client, api apiProbeRow, rawURL string) (string, string) {
	return probeHTTP(ctx, client, rawURL, api.ProviderCode, api.AuthType, api.EncryptedAPIKey, api.EncryptedSecret, api.KeyAlias, providerProbePayload(api.ProviderCode, api.APIType, api.Capabilities, api.APIPath, api.APIName))
}

func probeHTTP(ctx context.Context, client *http.Client, rawURL, providerCode, authType, encryptedAPIKey, encryptedSecret, keyAlias string, probe providerAPIProbe) (string, string) {
	if rawURL == "" {
		return "error", "API Endpoint 或路径不合法"
	}
	apiKey := resolveProviderAPIKey(encryptedAPIKey, keyAlias)
	if requiresAPIKey(providerCode, authType) && apiKey == "" {
		return "error", "未配置 API Key 或 Key Alias 环境变量"
	}
	statusCode, responseBody, err := probeHTTPMethod(ctx, client, probe.Method, rawURL, providerCode, authType, apiKey, encryptedSecret, probe)
	if err != nil {
		return "error", err.Error()
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return "error", fmt.Sprintf("HTTP %d，鉴权失败，请检查 API Key、Secret 或账号权限", statusCode)
	}
	if statusCode == http.StatusNotFound {
		return "warning", fmt.Sprintf("HTTP %d，API 路径可能不可用", statusCode)
	}
	if statusCode == http.StatusTooManyRequests {
		return "warning", fmt.Sprintf("HTTP %d，供应商限流或额度不足", statusCode)
	}
	if statusCode == http.StatusBadRequest || statusCode == http.StatusUnprocessableEntity {
		return "active", fmt.Sprintf("HTTP %d，接口可达且鉴权已通过：%s", statusCode, abbreviateProbeBody(responseBody))
	}
	if statusCode >= 500 {
		return "warning", fmt.Sprintf("HTTP %d，供应商服务异常：%s", statusCode, abbreviateProbeBody(responseBody))
	}
	if statusCode >= 200 && statusCode < 300 {
		return "active", fmt.Sprintf("HTTP %d，真实 API 调用连通", statusCode)
	}
	return "warning", fmt.Sprintf("HTTP %d，接口返回非预期状态：%s", statusCode, abbreviateProbeBody(responseBody))
}

func probeHTTPMethod(ctx context.Context, client *http.Client, method, rawURL, providerCode, authType, apiKey, secret string, probe providerAPIProbe) (int, string, error) {
	var body io.Reader
	if len(probe.Body) > 0 {
		body = bytes.NewReader(probe.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return 0, "", err
	}
	if probe.ContentType != "" {
		req.Header.Set("Content-Type", probe.ContentType)
	}
	req.Header.Set("Accept", "application/json")
	applyProviderAuth(req, providerCode, authType, apiKey, secret)
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return resp.StatusCode, string(rawBody), nil
}

func providerProbePayload(providerCode, apiType string, capabilities []string, apiPath, apiName string) providerAPIProbe {
	kind := strings.ToLower(strings.Join(append([]string{providerCode, apiType, apiPath, apiName}, capabilities...), " "))
	switch {
	case strings.Contains(kind, "dashscope") && strings.Contains(apiPath, "/api/v1/services/"):
		return jsonProbe(map[string]interface{}{
			"model": "qwen-turbo",
			"input": map[string]string{"prompt": "ping"},
			"parameters": map[string]int{
				"max_tokens": 1,
			},
		})
	case strings.Contains(kind, "qianfan") || strings.Contains(kind, "baidu"):
		return jsonProbe(map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": "ping"},
			},
			"disable_search": true,
		})
	case strings.Contains(kind, "hunyuan") || strings.Contains(kind, "tencent"):
		return jsonProbe(map[string]interface{}{
			"Model": "hunyuan-lite",
			"Messages": []map[string]string{
				{"Role": "user", "Content": "ping"},
			},
			"Stream": false,
		})
	case strings.Contains(kind, "minimax"):
		return jsonProbe(map[string]interface{}{
			"model": "abab6.5s-chat",
			"messages": []map[string]string{
				{"role": "user", "content": "ping"},
			},
			"tokens_to_generate": 1,
		})
	case strings.Contains(kind, "embedding"):
		return jsonProbe(map[string]interface{}{
			"model": providerProbeModel(providerCode, "embedding"),
			"input": "ping",
		})
	case strings.Contains(kind, "image"):
		return jsonProbe(map[string]interface{}{
			"model":  providerProbeModel(providerCode, "image"),
			"prompt": "ping",
			"n":      1,
			"size":   "1024x1024",
		})
	case strings.Contains(kind, "anthropic") || strings.Contains(apiPath, "/messages"):
		return jsonProbe(map[string]interface{}{
			"model":      providerProbeModel(providerCode, "chat"),
			"max_tokens": 1,
			"messages": []map[string]string{
				{"role": "user", "content": "ping"},
			},
		})
	default:
		return jsonProbe(map[string]interface{}{
			"model": providerProbeModel(providerCode, "chat"),
			"messages": []map[string]string{
				{"role": "user", "content": "ping"},
			},
			"max_tokens": 1,
		})
	}
}

func jsonProbe(payload map[string]interface{}) providerAPIProbe {
	body, _ := json.Marshal(payload)
	return providerAPIProbe{Method: http.MethodPost, ContentType: "application/json", Body: body}
}

func providerProbeModel(providerCode, capability string) string {
	code := strings.ToLower(strings.TrimSpace(providerCode))
	switch {
	case strings.Contains(code, "anthropic") || strings.Contains(code, "claude"):
		return "claude-3-haiku-20240307"
	case strings.Contains(code, "dashscope") || strings.Contains(code, "aliyun") || strings.Contains(code, "qwen"):
		if capability == "embedding" {
			return "text-embedding-v1"
		}
		return "qwen-turbo"
	case strings.Contains(code, "volc") || strings.Contains(code, "doubao") || strings.Contains(code, "ark"):
		return "doubao-lite-4k"
	case strings.Contains(code, "deepseek"):
		return "deepseek-chat"
	case strings.Contains(code, "zhipu") || strings.Contains(code, "glm"):
		return "glm-4-flash"
	case strings.Contains(code, "moonshot") || strings.Contains(code, "kimi"):
		return "moonshot-v1-8k"
	case strings.Contains(code, "baichuan"):
		return "Baichuan2-Turbo"
	case strings.Contains(code, "minimax"):
		return "abab6.5s-chat"
	case strings.Contains(code, "hunyuan") || strings.Contains(code, "tencent"):
		return "hunyuan-lite"
	case strings.Contains(code, "azure"):
		return "gpt-4o-mini"
	default:
		if capability == "embedding" {
			return "text-embedding-3-small"
		}
		if capability == "image" {
			return "dall-e-3"
		}
		return "gpt-4o-mini"
	}
}

func applyProviderAuth(req *http.Request, providerCode, authType, apiKey, secret string) {
	auth := strings.ToLower(strings.TrimSpace(authType))
	code := strings.ToLower(strings.TrimSpace(providerCode))
	switch {
	case auth == "none" || auth == "anonymous":
		return
	case auth == "x-api-key":
		req.Header.Set("X-API-Key", apiKey)
	case auth == "basic":
		req.SetBasicAuth(apiKey, secret)
	case auth == "dashscope" || strings.Contains(code, "dashscope") || strings.Contains(code, "qwen") || strings.Contains(code, "aliyun"):
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-DashScope-SSE", "disable")
	case auth == "anthropic" || strings.Contains(code, "anthropic") || strings.Contains(code, "claude"):
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case strings.Contains(code, "azure"):
		req.Header.Set("api-key", apiKey)
	default:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func requiresAPIKey(providerCode, authType string) bool {
	auth := strings.ToLower(strings.TrimSpace(authType))
	switch auth {
	case "none", "anonymous":
		return false
	}
	if auth == "" && strings.TrimSpace(providerCode) == "" {
		return false
	}
	return true
}

func resolveProviderAPIKey(encryptedAPIKey, keyAlias string) string {
	if value := strings.TrimSpace(encryptedAPIKey); value != "" {
		return value
	}
	alias := strings.TrimSpace(keyAlias)
	if alias == "" {
		return ""
	}
	return strings.TrimSpace(os.Getenv(alias))
}

func abbreviateProbeBody(body string) string {
	body = strings.TrimSpace(strings.ReplaceAll(body, "\n", " "))
	if body == "" {
		return "无响应体"
	}
	if len(body) > 160 {
		return body[:160] + "..."
	}
	return body
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
	result, err := listRows[models.AIProvider](ctx, s.db, skip, limit, keyword, []string{"name", "code", "owner", "region"})
	if err != nil {
		return PageResult{}, err
	}
	stats, err := s.providerCatalogSummary(ctx)
	if err != nil {
		return PageResult{}, err
	}
	result.Summary = map[string]interface{}{"provider_stats": stats}
	return result, nil
}

func (s *Service) ListAccounts(ctx context.Context, skip, limit int, providerID string) (PageResult, error) {
	query := func() *gorm.DB {
		q := s.db.WithContext(ctx).Model(&models.AIProviderAccount{}).Where("deleted_at IS NULL")
		if strings.TrimSpace(providerID) != "" {
			q = q.Where("provider_id = ?", providerID)
		}
		return q
	}
	summary, err := accountListSummary(query())
	if err != nil {
		return PageResult{}, err
	}
	result, err := pageQuery[models.AIProviderAccount](query(), skip, limit)
	if err != nil {
		return PageResult{}, err
	}
	result.Summary = summary
	return result, nil
}

func (s *Service) ListAPIs(ctx context.Context, skip, limit int, providerID, accountID string) (PageResult, error) {
	query := func() *gorm.DB {
		q := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).Where("deleted_at IS NULL")
		if strings.TrimSpace(providerID) != "" {
			q = q.Where("provider_id = ?", providerID)
		}
		if strings.TrimSpace(accountID) != "" {
			q = q.Where("account_id = ?", accountID)
		}
		return q
	}
	summary, err := apiListSummary(query())
	if err != nil {
		return PageResult{}, err
	}
	result, err := pageQuery[models.AIProviderAPI](query(), skip, limit)
	if err != nil {
		return PageResult{}, err
	}
	result.Summary = summary
	return result, nil
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
	strategy, strategyFound, err := s.matchTenantStrategy(ctx, req, scenario)
	if err != nil {
		return InvokeResponse{}, err
	}
	routeID := scenario.DefaultBaseRouteID
	if strategyFound {
		routeID = defaultString(strategy.OverrideBaseRouteID, strategy.DefaultBaseRouteID)
	}
	var route models.AIBaseRoute
	if err := s.db.WithContext(ctx).Where("id = ? AND status = ? AND deleted_at IS NULL", routeID, "active").First(&route).Error; err != nil {
		return InvokeResponse{}, ErrNotFound
	}
	routeModel, model, endpoint, err := s.selectRouteExecutionPlan(ctx, route, scenario)
	if err != nil {
		return InvokeResponse{}, err
	}
	providerID, accountID, apiID := endpoint.ProviderID, endpoint.AccountID, endpoint.APIID

	paramsRaw, _ := json.Marshal(req.Params)
	inputRaw, _ := json.Marshal(req.Input)
	hash := sha256.Sum256(paramsRaw)
	responseHash := sha256.Sum256(inputRaw)
	now := time.Now()
	pricing, err := s.matchPricing(ctx, model, scenario, req)
	if err != nil {
		return InvokeResponse{}, err
	}
	quotaChecks, err := s.evaluateQuotaRules(ctx, req, scenario, strategy, strategyFound, routeModel, pricing, accountID, apiID, now)
	if err != nil {
		return InvokeResponse{}, err
	}
	rateChecks, err := s.evaluateRateLimitRules(ctx, req, scenario, strategy, strategyFound, routeModel, pricing, accountID, apiID, now)
	if err != nil {
		return InvokeResponse{}, err
	}
	status := "success"
	errorCode := ""
	errorMessage := ""
	for _, check := range quotaChecks {
		if check.Exceeded && check.Action == "reject" {
			status = "rejected"
			errorCode = "quota_exceeded"
			errorMessage = "AI Gateway 配额超限"
			break
		}
	}
	if status == "success" {
		for _, check := range rateChecks {
			if check.Exceeded && check.Action == "reject" {
				status = "rejected"
				errorCode = "rate_limited"
				errorMessage = "AI Gateway 限流超限"
				break
			}
		}
	}
	record := models.AIUsageRecord{
		RequestID:         req.RequestID,
		TenantID:          req.TenantID,
		TenantName:        req.TenantName,
		AppCode:           scenario.AppCode,
		AppName:           defaultString(req.AppName, scenario.AppName),
		AIScenarioCode:    scenario.AIScenarioCode,
		AIScenarioName:    scenario.AIScenarioName,
		UserID:            req.UserID,
		UserName:          req.UserName,
		ProviderID:        providerID,
		ProviderAccountID: accountID,
		ProviderAPIID:     apiID,
		ModelID:           routeModel.ModelID,
		BaseRouteID:       route.ID,
		TenantStrategyID:  strategyID(strategy, strategyFound),
		PricePolicyID:     pricing.PolicyID,
		PriceTierID:       pricing.TierID,
		UsageAmount:       pricing.UsageAmount,
		UsageUnit:         pricing.UsageUnit,
		UsageDetail:       "Gateway 路由、策略、配额、限流和价格命中记录",
		Calls:             1,
		CostAmount:        pricing.CostAmount,
		BillingAmount:     pricing.BillingAmount,
		PlatformUnit:      pricing.PlatformUnit,
		PlatformAmount:    pricing.PlatformAmount,
		Status:            status,
		ErrorCode:         errorCode,
		ErrorMessage:      errorMessage,
		LatencyMS:         0,
		RequestParams:     string(paramsRaw),
		PromptHash:        hex.EncodeToString(hash[:]),
		ResponseHash:      hex.EncodeToString(responseHash[:]),
		CalledAt:          now,
		CreatedAt:         now,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if status == "success" && strategyFound {
			for _, check := range quotaChecks {
				if check.RuleID == "" {
					continue
				}
				if err := tx.Model(&models.AIStrategyQuotaRule{}).
					Where("id = ? AND deleted_at IS NULL", check.RuleID).
					Update("used_amount", gorm.Expr("used_amount + ?", pricing.UsageAmount)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return InvokeResponse{}, err
	}
	return InvokeResponse{
		RequestID:   req.RequestID,
		Status:      status,
		ModelID:     routeModel.ModelID,
		BaseRouteID: route.ID,
		StrategyID:  record.TenantStrategyID,
		Usage: map[string]interface{}{
			"amount": record.UsageAmount,
			"unit":   record.UsageUnit,
		},
		Billing: map[string]interface{}{
			"price_policy_id": record.PricePolicyID,
			"price_tier_id":   record.PriceTierID,
			"feature_key":     pricing.FeatureKey,
			"cost_amount":     record.CostAmount,
			"billing_amount":  record.BillingAmount,
			"platform_unit":   record.PlatformUnit,
			"platform_amount": record.PlatformAmount,
		},
		Controls: map[string]interface{}{
			"quota_rules":      quotaChecks,
			"rate_limit_rules": rateChecks,
		},
		Data: map[string]interface{}{},
	}, nil
}

func (s *Service) matchTenantStrategy(ctx context.Context, req InvokeRequest, scenario models.AIScenario) (models.AITenantStrategyPolicy, bool, error) {
	var rows []models.AITenantStrategyPolicy
	err := s.db.WithContext(ctx).
		Where("app_code = ? AND ai_scenario_code = ? AND status = ? AND deleted_at IS NULL", scenario.AppCode, scenario.AIScenarioCode, "active").
		Order("updated_at desc").
		Find(&rows).Error
	if err != nil {
		return models.AITenantStrategyPolicy{}, false, err
	}
	bestScore := -1
	var best models.AITenantStrategyPolicy
	for _, row := range rows {
		score := tenantStrategyMatchScore(row, req.TenantID)
		if score > bestScore {
			bestScore = score
			best = row
		}
	}
	return best, bestScore >= 0, nil
}

func tenantStrategyMatchScore(row models.AITenantStrategyPolicy, tenantID string) int {
	scope := strings.TrimSpace(row.TenantScope)
	switch scope {
	case "include":
		if containsString(row.TenantIDs, tenantID) {
			return 30
		}
	case "all":
		return 20
	case "exclude":
		if !containsString(row.TenantIDs, tenantID) {
			return 10
		}
	}
	return -1
}

type routeEndpoint struct {
	ProviderID string
	AccountID  string
	APIID      string
	Score      float64
}

func (s *Service) selectRouteExecutionPlan(ctx context.Context, route models.AIBaseRoute, scenario models.AIScenario) (models.AIBaseRouteModel, models.AIModel, routeEndpoint, error) {
	var rows []models.AIBaseRouteModel
	if err := s.db.WithContext(ctx).Where("base_route_id = ? AND status = ? AND deleted_at IS NULL", route.ID, "active").Find(&rows).Error; err != nil {
		return models.AIBaseRouteModel{}, models.AIModel{}, routeEndpoint{}, err
	}
	if len(rows) == 0 {
		return models.AIBaseRouteModel{}, models.AIModel{}, routeEndpoint{}, ErrNotFound
	}
	bestScore := -1.0
	var bestRouteModel models.AIBaseRouteModel
	var bestModel models.AIModel
	var bestEndpoint routeEndpoint
	for _, row := range rows {
		var model models.AIModel
		if err := s.db.WithContext(ctx).Where("id = ? AND status = ? AND deleted_at IS NULL", row.ModelID, "active").First(&model).Error; err != nil {
			continue
		}
		if scenario.ModelType != "" && model.ModelType != scenario.ModelType {
			continue
		}
		endpoint, ok := s.resolveRouteEndpoint(ctx, row, model, scenario.CapabilityCode)
		if !ok {
			continue
		}
		score := routeModelScore(route.Strategy, row, model) + endpoint.Score
		if score > bestScore {
			bestScore = score
			bestRouteModel = row
			bestModel = model
			bestEndpoint = endpoint
		}
	}
	if bestScore < 0 {
		return models.AIBaseRouteModel{}, models.AIModel{}, routeEndpoint{}, ErrNotFound
	}
	return bestRouteModel, bestModel, bestEndpoint, nil
}

func routeModelScore(strategy string, row models.AIBaseRouteModel, model models.AIModel) float64 {
	priorityScore := float64(100000 - row.Priority*100)
	weightScore := float64(row.Weight)
	switch strings.TrimSpace(strategy) {
	case "load_balance":
		return weightScore*1000 + priorityScore
	case "cost_first":
		return priorityScore + weightScore
	case "quality_first":
		return model.SuccessRate*1000 + priorityScore + weightScore
	case "latency_first":
		latency := model.LatencyP95
		if latency <= 0 {
			latency = 999999
		}
		return float64(1000000-latency) + priorityScore + weightScore
	default:
		roleScore := 0.0
		if row.Role == "primary" {
			roleScore = 10000
		}
		if row.Role == "fallback" {
			roleScore = 5000
		}
		return roleScore + priorityScore + weightScore
	}
}

func (s *Service) resolveRouteEndpoint(ctx context.Context, routeModel models.AIBaseRouteModel, model models.AIModel, capabilityCode string) (routeEndpoint, bool) {
	accountID := strings.TrimSpace(routeModel.ProviderAccountID)
	apiID := strings.TrimSpace(routeModel.ProviderAPIID)
	if apiID != "" {
		var api models.AIProviderAPI
		err := s.db.WithContext(ctx).
			Where("id = ? AND provider_id = ? AND status = ? AND deleted_at IS NULL", apiID, model.ProviderID, "active").
			First(&api).Error
		if err != nil || !apiUsable(api, capabilityCode) {
			return routeEndpoint{}, false
		}
		if accountID != "" && api.AccountID != accountID {
			return routeEndpoint{}, false
		}
		if !s.accountUsable(ctx, api.AccountID, model.ProviderID) {
			return routeEndpoint{}, false
		}
		return routeEndpoint{ProviderID: model.ProviderID, AccountID: api.AccountID, APIID: api.ID, Score: endpointHealthScore(api)}, true
	}

	accountQuery := s.db.WithContext(ctx).
		Where("provider_id = ? AND status = ? AND deleted_at IS NULL", model.ProviderID, "active")
	if accountID != "" {
		accountQuery = accountQuery.Where("id = ?", accountID)
	}
	var accounts []models.AIProviderAccount
	if err := accountQuery.Order("updated_at desc").Find(&accounts).Error; err != nil || len(accounts) == 0 {
		return routeEndpoint{}, false
	}

	bestScore := -1.0
	var best routeEndpoint
	for _, account := range accounts {
		var apis []models.AIProviderAPI
		err := s.db.WithContext(ctx).
			Where("provider_id = ? AND account_id = ? AND status = ? AND deleted_at IS NULL", model.ProviderID, account.ID, "active").
			Order("updated_at desc").
			Find(&apis).Error
		if err != nil {
			continue
		}
		for _, api := range apis {
			if !apiUsable(api, capabilityCode) {
				continue
			}
			score := endpointHealthScore(api)
			if score > bestScore {
				bestScore = score
				best = routeEndpoint{ProviderID: model.ProviderID, AccountID: account.ID, APIID: api.ID, Score: score}
			}
		}
	}
	return best, bestScore >= 0
}

func (s *Service) accountUsable(ctx context.Context, accountID, providerID string) bool {
	var count int64
	_ = s.db.WithContext(ctx).Model(&models.AIProviderAccount{}).
		Where("id = ? AND provider_id = ? AND status = ? AND deleted_at IS NULL", accountID, providerID, "active").
		Count(&count).Error
	return count > 0
}

func apiUsable(api models.AIProviderAPI, capabilityCode string) bool {
	if strings.TrimSpace(api.Status) != "active" || !containsString(api.Capabilities, capabilityCode) {
		return false
	}
	switch strings.TrimSpace(api.HealthStatus) {
	case "error", "inactive":
		return false
	default:
		return true
	}
}

func endpointHealthScore(api models.AIProviderAPI) float64 {
	score := 0.0
	if api.HealthStatus == "active" {
		score += 1000
	}
	if api.HealthCheckedAt != nil && time.Since(*api.HealthCheckedAt) <= 10*time.Minute {
		score += 100
	}
	if api.QPSLimit > 0 {
		score += float64(api.QPSLimit) / 100
	}
	return score
}

func (s *Service) matchPricing(ctx context.Context, model models.AIModel, scenario models.AIScenario, req InvokeRequest) (pricingResult, error) {
	usageAmount := numberParam(req.Params, "usage_amount", 1)
	if usageAmount <= 0 {
		usageAmount = 1
	}
	result := pricingResult{
		UsageAmount: usageAmount,
		UsageUnit:   usageUnitForCapability(ctx, s.db, scenario.CapabilityCode),
	}
	var policy models.AIModelPricePolicy
	err := s.db.WithContext(ctx).
		Where("model_id = ? AND capability_code = ? AND status = ? AND deleted_at IS NULL", model.ID, scenario.CapabilityCode, "active").
		Order("updated_at desc").
		First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
	}
	if err != nil {
		return pricingResult{}, err
	}
	result.PolicyID = policy.ID
	result.UsageUnit = defaultString(policy.BillingUnit, result.UsageUnit)
	result.PlatformUnit = policy.PlatformUnit
	result.PlatformAmount = policy.BasePlatformAmount * usageAmount
	result.FeatureKey = policy.FeatureKey
	costPrice := policy.BaseCostPrice
	salePrice := policy.BaseSalePrice
	var tiers []models.AIModelPriceTier
	if err := s.db.WithContext(ctx).Where("price_policy_id = ? AND enabled = ? AND deleted_at IS NULL", policy.ID, true).Order("sort_order asc").Find(&tiers).Error; err != nil {
		return pricingResult{}, err
	}
	if tier, ok := matchPriceTier(tiers, req.Params); ok {
		result.TierID = tier.ID
		costPrice = tier.CostPrice
		salePrice = tier.SalePrice
		if tier.PlatformAmount > 0 {
			result.PlatformAmount = tier.PlatformAmount * usageAmount
		}
	}
	result.CostAmount = costPrice * usageAmount
	result.BillingAmount = salePrice * usageAmount
	return result, nil
}

func matchPriceTier(tiers []models.AIModelPriceTier, params map[string]interface{}) (models.AIModelPriceTier, bool) {
	for _, tier := range tiers {
		if stringParam(params, "mode") != "" && tier.Mode != "" && stringParam(params, "mode") != tier.Mode {
			continue
		}
		if stringParam(params, "resolution") != "" && tier.Resolution != "" && stringParam(params, "resolution") != tier.Resolution {
			continue
		}
		if stringParam(params, "quality") != "" && tier.Quality != "" && stringParam(params, "quality") != tier.Quality {
			continue
		}
		if stringParam(params, "aspect_ratio") != "" && tier.AspectRatio != "" && stringParam(params, "aspect_ratio") != tier.AspectRatio {
			continue
		}
		if numberParam(params, "duration_seconds", 0) > 0 && tier.DurationSeconds > 0 && int(numberParam(params, "duration_seconds", 0)) != tier.DurationSeconds {
			continue
		}
		return tier, true
	}
	if len(tiers) > 0 {
		return tiers[0], true
	}
	return models.AIModelPriceTier{}, false
}

func (s *Service) evaluateQuotaRules(ctx context.Context, req InvokeRequest, scenario models.AIScenario, strategy models.AITenantStrategyPolicy, found bool, routeModel models.AIBaseRouteModel, pricing pricingResult, accountID, apiID string, now time.Time) ([]quotaCheckResult, error) {
	if !found {
		return []quotaCheckResult{}, nil
	}
	var rules []models.AIStrategyQuotaRule
	if err := s.db.WithContext(ctx).Where("policy_id = ? AND status = ? AND deleted_at IS NULL", strategy.ID, "active").Find(&rules).Error; err != nil {
		return nil, err
	}
	results := make([]quotaCheckResult, 0, len(rules))
	for _, rule := range rules {
		if !controlRuleMatches(rule.Dimension, rule.SubjectCode, req, scenario, routeModel, pricing, accountID, apiID) {
			continue
		}
		start := periodStart(now, rule.Period)
		used, err := s.usageAmountForQuotaRule(ctx, strategy.ID, rule, start)
		if err != nil {
			return nil, err
		}
		used += rule.UsedAmount
		exceeded := rule.QuotaLimit > 0 && used+pricing.UsageAmount > rule.QuotaLimit
		results = append(results, quotaCheckResult{
			RuleID: rule.ID, Dimension: rule.Dimension, SubjectCode: rule.SubjectCode, UsageUnit: rule.UsageUnit, Period: rule.Period,
			Limit: rule.QuotaLimit, UsedBefore: used, RequestedAmount: pricing.UsageAmount, Exceeded: exceeded, Action: rule.OverLimitAction,
		})
	}
	return results, nil
}

func (s *Service) usageAmountForQuotaRule(ctx context.Context, strategyID string, rule models.AIStrategyQuotaRule, start *time.Time) (float64, error) {
	q := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).
		Where("tenant_strategy_id = ? AND usage_unit = ?", strategyID, rule.UsageUnit)
	if start != nil {
		q = q.Where("called_at >= ?", *start)
	}
	var used float64
	if err := q.Select("COALESCE(SUM(usage_amount),0)").Row().Scan(&used); err != nil {
		return 0, err
	}
	return used, nil
}

func (s *Service) evaluateRateLimitRules(ctx context.Context, req InvokeRequest, scenario models.AIScenario, strategy models.AITenantStrategyPolicy, found bool, routeModel models.AIBaseRouteModel, pricing pricingResult, accountID, apiID string, now time.Time) ([]rateLimitCheckResult, error) {
	if !found {
		return []rateLimitCheckResult{}, nil
	}
	var rules []models.AIStrategyRateLimitRule
	if err := s.db.WithContext(ctx).Where("policy_id = ? AND status = ? AND deleted_at IS NULL", strategy.ID, "active").Find(&rules).Error; err != nil {
		return nil, err
	}
	results := []rateLimitCheckResult{}
	for _, rule := range rules {
		if !controlRuleMatches(rule.Dimension, rule.SubjectCode, req, scenario, routeModel, pricing, accountID, apiID) {
			continue
		}
		checks := []struct {
			window string
			limit  int
			start  time.Time
		}{
			{"minute", rule.MinuteLimit, now.Add(-time.Minute)},
			{"hour", rule.HourLimit, now.Add(-time.Hour)},
			{"day", rule.DayLimit, time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())},
		}
		if rule.QPS > 0 {
			checks = append(checks, struct {
				window string
				limit  int
				start  time.Time
			}{"second", rule.QPS, now.Add(-time.Second)})
		}
		for _, check := range checks {
			if check.limit <= 0 {
				continue
			}
			used, err := s.callsSince(ctx, strategy.ID, check.start)
			if err != nil {
				return nil, err
			}
			results = append(results, rateLimitCheckResult{
				RuleID: rule.ID, Dimension: rule.Dimension, SubjectCode: rule.SubjectCode, Window: check.window,
				Limit: check.limit, UsedBefore: used, Exceeded: used+1 > int64(check.limit), Action: rule.OverLimitAction,
			})
		}
	}
	return results, nil
}

func (s *Service) callsSince(ctx context.Context, strategyID string, start time.Time) (int64, error) {
	var calls int64
	err := s.db.WithContext(ctx).Model(&models.AIUsageRecord{}).
		Where("tenant_strategy_id = ? AND called_at >= ?", strategyID, start).
		Select("COALESCE(SUM(calls),0)").Row().Scan(&calls)
	return calls, err
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
		applyProviderAccountSecretPayload(payload, &row)
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
		if err := validateCapability(row); err != nil {
			return nil, err
		}
		row.ID = uuid.NewString()
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
			!validRouteModelRole(row.Role) || row.Priority <= 0 || row.Weight <= 0 || row.MaxRetry < 0 || row.TimeoutMS <= 0 ||
			!s.routeModelEndpointRefsValid(ctx, row) {
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
		normalizeSettingPayload(payload)
		var row models.AIGatewaySetting
		if err := decodePayload(payload, &row); err != nil {
			return nil, err
		}
		if err := validateGatewaySetting(row); err != nil {
			return nil, err
		}
		row.ID = uuid.NewString()
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
		return updateRow[models.AIProvider](ctx, s, userID, resource, id, payload, "ai_provider", validateProvider)
	case "accounts":
		return updateRow[models.AIProviderAccount](ctx, s, userID, resource, id, payload, "ai_provider_account", validateProviderAccount)
	case "apis":
		return updateRow[models.AIProviderAPI](ctx, s, userID, resource, id, payload, "ai_provider_api", validateProviderAPI)
	case "capabilities":
		return updateRow[models.AICapability](ctx, s, userID, resource, id, payload, "ai_capability", validateCapability)
	case "models":
		return updateRow[models.AIModel](ctx, s, userID, resource, id, payload, "ai_model", validateModel)
	case "price-policies":
		return updateRow[models.AIModelPricePolicy](ctx, s, userID, resource, id, payload, "ai_model_price_policy", validatePricePolicy)
	case "price-tiers":
		return updateRow[models.AIModelPriceTier](ctx, s, userID, resource, id, payload, "ai_model_price_tier", validatePriceTier)
	case "base-routes":
		return updateRow[models.AIBaseRoute](ctx, s, userID, resource, id, payload, "ai_base_route", validateBaseRoute)
	case "route-models":
		return s.updateRouteModel(ctx, userID, resource, id, payload)
	case "scenarios":
		return updateRow[models.AIScenario](ctx, s, userID, resource, id, payload, "ai_scenario", validateScenario)
	case "tenant-strategies":
		return updateRow[models.AITenantStrategyPolicy](ctx, s, userID, resource, id, payload, "ai_tenant_strategy", validateTenantStrategy)
	case "quota-rules":
		return updateRow[models.AIStrategyQuotaRule](ctx, s, userID, resource, id, payload, "ai_strategy_quota_rule", validateQuotaRule)
	case "rate-limit-rules":
		return updateRow[models.AIStrategyRateLimitRule](ctx, s, userID, resource, id, payload, "ai_strategy_rate_limit_rule", validateRateLimitRule)
	case "settings":
		normalizeSettingPayload(payload)
		return updateRow[models.AIGatewaySetting](ctx, s, userID, resource, id, payload, "ai_gateway_setting", validateGatewaySetting)
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
		module = "ai_tenant_strategy"
		if err := s.deleteTenantStrategyCascade(ctx, id, now); err != nil {
			return err
		}
		s.Audit(ctx, userID, module, "delete", "删除租户策略："+id, map[string]string{"id": id})
		return nil
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

type providerCatalogStat struct {
	AccountCount   int64 `json:"account_count"`
	APICount       int64 `json:"api_count"`
	ActiveAPICount int64 `json:"active_api_count"`
}

func (s *Service) providerCatalogSummary(ctx context.Context) (map[string]providerCatalogStat, error) {
	stats := map[string]providerCatalogStat{}
	var accountRows []struct {
		ProviderID   string
		AccountCount int64
	}
	if err := s.db.WithContext(ctx).Model(&models.AIProviderAccount{}).
		Select("provider_id, COUNT(*) AS account_count").
		Where("deleted_at IS NULL").
		Group("provider_id").
		Scan(&accountRows).Error; err != nil {
		return nil, err
	}
	for _, row := range accountRows {
		stat := stats[row.ProviderID]
		stat.AccountCount = row.AccountCount
		stats[row.ProviderID] = stat
	}
	var apiRows []struct {
		ProviderID     string
		APICount       int64
		ActiveAPICount int64
	}
	if err := s.db.WithContext(ctx).Model(&models.AIProviderAPI{}).
		Select("provider_id, COUNT(*) AS api_count, COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active_api_count").
		Where("deleted_at IS NULL").
		Group("provider_id").
		Scan(&apiRows).Error; err != nil {
		return nil, err
	}
	for _, row := range apiRows {
		stat := stats[row.ProviderID]
		stat.APICount = row.APICount
		stat.ActiveAPICount = row.ActiveAPICount
		stats[row.ProviderID] = stat
	}
	return stats, nil
}

func accountListSummary(q *gorm.DB) (map[string]interface{}, error) {
	var row struct {
		ActiveCount int64
		QuotaTotal  float64
	}
	err := q.Select("COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active_count, COALESCE(SUM(quota_limit), 0) AS quota_total").Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"active_count": row.ActiveCount, "quota_total": row.QuotaTotal}, nil
}

func apiListSummary(q *gorm.DB) (map[string]interface{}, error) {
	var row struct {
		ActiveCount int64
		QPSTotal    int64
	}
	err := q.Select("COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active_count, COALESCE(SUM(qps_limit), 0) AS qps_total").Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"active_count": row.ActiveCount, "qps_total": row.QPSTotal}, nil
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

func applyProviderAccountSecretPayload(payload map[string]interface{}, row *models.AIProviderAccount) {
	if value, ok := payload["encrypted_api_key"]; ok {
		row.EncryptedAPIKey = strings.TrimSpace(fmt.Sprint(value))
	}
	if value, ok := payload["encrypted_secret"]; ok {
		row.EncryptedSecret = strings.TrimSpace(fmt.Sprint(value))
	}
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
		ProviderID:        providerID,
		AccountName:       strings.TrimSpace(item.AccountName),
		Endpoint:          strings.TrimSpace(item.Endpoint),
		KeyAlias:          strings.TrimSpace(item.KeyAlias),
		LoginMethod:       strings.TrimSpace(item.LoginMethod),
		LoginAccount:      strings.TrimSpace(item.LoginAccount),
		Maintainer:        strings.TrimSpace(item.Maintainer),
		MaintainerContact: strings.TrimSpace(item.MaintainerContact),
		EncryptedAPIKey:   strings.TrimSpace(item.EncryptedAPIKey),
		EncryptedSecret:   strings.TrimSpace(item.EncryptedSecret),
		QuotaLimit:        item.QuotaLimit,
		UsedQuota:         item.UsedQuota,
		Status:            defaultString(item.Status, "active"),
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
		"endpoint": row.Endpoint, "key_alias": row.KeyAlias, "login_method": row.LoginMethod,
		"login_account": row.LoginAccount, "maintainer": row.Maintainer, "maintainer_contact": row.MaintainerContact,
		"encrypted_api_key": row.EncryptedAPIKey, "encrypted_secret": row.EncryptedSecret, "quota_limit": row.QuotaLimit, "used_quota": row.UsedQuota,
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
		BaseRouteID:       baseRouteID,
		ModelID:           modelID,
		ProviderAccountID: strings.TrimSpace(item.ProviderAccountID),
		ProviderAPIID:     strings.TrimSpace(item.ProviderAPIID),
		Role:              defaultString(item.Role, "candidate"),
		Priority:          defaultInt(item.Priority, 1),
		Weight:            defaultInt(item.Weight, 100),
		MaxRetry:          item.MaxRetry,
		TimeoutMS:         defaultInt(item.TimeoutMS, 30000),
		Status:            defaultString(item.Status, "active"),
	}
	if row.BaseRouteID == "" || row.ModelID == "" || !validRouteModelRole(row.Role) ||
		row.Priority <= 0 || row.Weight <= 0 || row.MaxRetry < 0 || row.TimeoutMS <= 0 {
		return ErrInvalidInput
	}
	if !existsTx(ctx, tx, &models.AIBaseRoute{}, row.BaseRouteID) || !existsTx(ctx, tx, &models.AIModel{}, row.ModelID) {
		return ErrInvalidInput
	}
	if !routeModelEndpointRefsValidTx(ctx, tx, row) {
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
		"provider_account_id": nullableString(row.ProviderAccountID), "provider_api_id": nullableString(row.ProviderAPIID),
		"priority": row.Priority, "weight": row.Weight, "max_retry": row.MaxRetry,
		"timeout_ms": row.TimeoutMS, "status": row.Status, "updated_at": now,
	}).Error
}

func loadTenantStrategyPolicyIDs(ctx context.Context, tx *gorm.DB, policyIDs map[string]string) error {
	var rows []struct {
		ID         string
		PolicyName string
	}
	if err := tx.WithContext(ctx).Model(&models.AITenantStrategyPolicy{}).Select("id, policy_name").Where("deleted_at IS NULL").Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		policyIDs[row.PolicyName] = row.ID
	}
	return nil
}

func upsertTenantStrategyPolicyImport(ctx context.Context, tx *gorm.DB, item TenantStrategyImportPolicy, now time.Time) (models.AITenantStrategyPolicy, error) {
	row := models.AITenantStrategyPolicy{
		PolicyName:          strings.TrimSpace(item.PolicyName),
		TenantScope:         defaultString(item.TenantScope, "all"),
		TenantIDs:           item.TenantIDs,
		AppCode:             strings.TrimSpace(item.AppCode),
		AppName:             strings.TrimSpace(item.AppName),
		AIScenarioCode:      strings.TrimSpace(item.AIScenarioCode),
		AIScenarioName:      strings.TrimSpace(item.AIScenarioName),
		DefaultBaseRouteID:  strings.TrimSpace(item.DefaultBaseRouteID),
		OverrideBaseRouteID: strings.TrimSpace(item.OverrideBaseRouteID),
		Description:         strings.TrimSpace(item.Description),
		Status:              defaultString(item.Status, "active"),
	}
	if row.PolicyName == "" || row.AppCode == "" || row.AIScenarioCode == "" || row.DefaultBaseRouteID == "" || !validTenantScope(row.TenantScope) {
		return models.AITenantStrategyPolicy{}, ErrInvalidInput
	}
	if !scenarioExistsTx(ctx, tx, row.AppCode, row.AIScenarioCode) || !existsTx(ctx, tx, &models.AIBaseRoute{}, row.DefaultBaseRouteID) {
		return models.AITenantStrategyPolicy{}, ErrInvalidInput
	}
	if row.OverrideBaseRouteID != "" && !existsTx(ctx, tx, &models.AIBaseRoute{}, row.OverrideBaseRouteID) {
		return models.AITenantStrategyPolicy{}, ErrInvalidInput
	}
	var existing models.AITenantStrategyPolicy
	err := tx.WithContext(ctx).Where("policy_name = ? AND tenant_scope = ? AND app_code = ? AND ai_scenario_code = ? AND deleted_at IS NULL", row.PolicyName, row.TenantScope, row.AppCode, row.AIScenarioCode).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = uuid.NewString()
		row.CreatedAt = now
		row.UpdatedAt = now
		if err := tx.Create(&row).Error; err != nil {
			return models.AITenantStrategyPolicy{}, err
		}
		return row, nil
	}
	if err != nil {
		return models.AITenantStrategyPolicy{}, err
	}
	row.ID = existing.ID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = now
	err = tx.Model(&existing).Updates(map[string]interface{}{
		"tenant_ids": jsonString(row.TenantIDs), "app_name": row.AppName, "ai_scenario_name": row.AIScenarioName,
		"default_base_route_id": row.DefaultBaseRouteID, "override_base_route_id": row.OverrideBaseRouteID,
		"description": row.Description, "status": row.Status, "updated_at": now,
	}).Error
	return row, err
}

func upsertTenantStrategyQuotaRuleImport(ctx context.Context, tx *gorm.DB, policyIDs map[string]string, item TenantStrategyImportQuotaRule, now time.Time) error {
	policyID := strings.TrimSpace(item.PolicyID)
	if policyID == "" {
		policyID = policyIDs[strings.TrimSpace(item.PolicyName)]
	}
	row := models.AIStrategyQuotaRule{
		PolicyID:         policyID,
		Dimension:        strings.TrimSpace(item.Dimension),
		SubjectCode:      strings.TrimSpace(item.SubjectCode),
		UsageUnit:        strings.TrimSpace(item.UsageUnit),
		Period:           strings.TrimSpace(item.Period),
		QuotaLimit:       item.QuotaLimit,
		UsedAmount:       item.UsedAmount,
		WarningThreshold: defaultFloat(item.WarningThreshold, 80),
		OverLimitAction:  defaultString(item.OverLimitAction, "alert_only"),
		Status:           defaultString(item.Status, "active"),
	}
	if row.PolicyID == "" || row.SubjectCode == "" || row.UsageUnit == "" || row.Period == "" ||
		!validControlDimension(row.Dimension) || !validControlPeriod(row.Period) || !validOverLimitAction(row.OverLimitAction) ||
		row.QuotaLimit <= 0 || row.UsedAmount < 0 || row.WarningThreshold <= 0 {
		return ErrInvalidInput
	}
	if !existsTx(ctx, tx, &models.AITenantStrategyPolicy{}, row.PolicyID) {
		return ErrInvalidInput
	}
	var existing models.AIStrategyQuotaRule
	err := tx.WithContext(ctx).Where("policy_id = ? AND dimension = ? AND subject_code = ? AND usage_unit = ? AND period = ? AND deleted_at IS NULL", row.PolicyID, row.Dimension, row.SubjectCode, row.UsageUnit, row.Period).First(&existing).Error
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
		"quota_limit": row.QuotaLimit, "used_amount": row.UsedAmount, "warning_threshold": row.WarningThreshold,
		"over_limit_action": row.OverLimitAction, "status": row.Status, "updated_at": now,
	}).Error
}

func upsertTenantStrategyRateLimitRuleImport(ctx context.Context, tx *gorm.DB, policyIDs map[string]string, item TenantStrategyImportRateLimitRule, now time.Time) error {
	policyID := strings.TrimSpace(item.PolicyID)
	if policyID == "" {
		policyID = policyIDs[strings.TrimSpace(item.PolicyName)]
	}
	row := models.AIStrategyRateLimitRule{
		PolicyID:        policyID,
		Dimension:       strings.TrimSpace(item.Dimension),
		SubjectCode:     strings.TrimSpace(item.SubjectCode),
		QPS:             item.QPS,
		Concurrency:     item.Concurrency,
		MinuteLimit:     item.MinuteLimit,
		HourLimit:       item.HourLimit,
		DayLimit:        item.DayLimit,
		OverLimitAction: defaultString(item.OverLimitAction, "queue"),
		Status:          defaultString(item.Status, "active"),
	}
	if row.PolicyID == "" || row.SubjectCode == "" || !validControlDimension(row.Dimension) || !validOverLimitAction(row.OverLimitAction) ||
		row.QPS < 0 || row.Concurrency < 0 || row.MinuteLimit < 0 || row.HourLimit < 0 || row.DayLimit < 0 ||
		(row.QPS == 0 && row.Concurrency == 0 && row.MinuteLimit == 0 && row.HourLimit == 0 && row.DayLimit == 0) {
		return ErrInvalidInput
	}
	if !existsTx(ctx, tx, &models.AITenantStrategyPolicy{}, row.PolicyID) {
		return ErrInvalidInput
	}
	var existing models.AIStrategyRateLimitRule
	err := tx.WithContext(ctx).Where("policy_id = ? AND dimension = ? AND subject_code = ? AND deleted_at IS NULL", row.PolicyID, row.Dimension, row.SubjectCode).First(&existing).Error
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
		"qps": row.QPS, "concurrency": row.Concurrency, "minute_limit": row.MinuteLimit,
		"hour_limit": row.HourLimit, "day_limit": row.DayLimit, "over_limit_action": row.OverLimitAction,
		"status": row.Status, "updated_at": now,
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

func validateProviderAccount(row models.AIProviderAccount) error {
	if strings.TrimSpace(row.ProviderID) == "" || strings.TrimSpace(row.AccountName) == "" || strings.TrimSpace(row.KeyAlias) == "" || row.QuotaLimit < 0 || row.UsedQuota < 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateProviderAPI(row models.AIProviderAPI) error {
	if strings.TrimSpace(row.ProviderID) == "" || strings.TrimSpace(row.AccountID) == "" || strings.TrimSpace(row.APIName) == "" ||
		strings.TrimSpace(row.APIPath) == "" || strings.TrimSpace(row.APIType) == "" || len(row.Capabilities) == 0 || row.QPSLimit < 0 || row.TimeoutMS <= 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateCapability(row models.AICapability) error {
	if strings.TrimSpace(row.CapabilityCode) == "" || strings.TrimSpace(row.CapabilityName) == "" ||
		strings.TrimSpace(row.ScenarioType) == "" || strings.TrimSpace(row.ModelType) == "" || strings.TrimSpace(row.DefaultBillingUnit) == "" {
		return ErrInvalidInput
	}
	return nil
}

func validateModel(row models.AIModel) error {
	if strings.TrimSpace(row.ProviderID) == "" || strings.TrimSpace(row.ModelCode) == "" || strings.TrimSpace(row.ModelName) == "" ||
		strings.TrimSpace(row.ModelType) == "" || len(row.Capabilities) == 0 || row.ContextWindow < 0 || row.LatencyP95 < 0 ||
		row.SuccessRate < 0 || row.SuccessRate > 100 {
		return ErrInvalidInput
	}
	return nil
}

func validatePricePolicy(row models.AIModelPricePolicy) error {
	if strings.TrimSpace(row.ModelID) == "" || strings.TrimSpace(row.FeatureKey) == "" || strings.TrimSpace(row.FeatureName) == "" ||
		strings.TrimSpace(row.ModelType) == "" || strings.TrimSpace(row.CapabilityCode) == "" || strings.TrimSpace(row.BillingMode) == "" ||
		strings.TrimSpace(row.BillingUnit) == "" || strings.TrimSpace(row.PlatformUnit) == "" || row.BaseCostPrice < 0 || row.BaseSalePrice < 0 ||
		row.BasePlatformAmount < 0 {
		return ErrInvalidInput
	}
	return nil
}

func validatePriceTier(row models.AIModelPriceTier) error {
	if strings.TrimSpace(row.PricePolicyID) == "" || strings.TrimSpace(row.TierName) == "" || row.DurationSeconds < 0 ||
		row.CostPrice < 0 || row.SalePrice < 0 || row.PlatformAmount < 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateBaseRoute(row models.AIBaseRoute) error {
	if strings.TrimSpace(row.RouteCode) == "" || strings.TrimSpace(row.RouteName) == "" || strings.TrimSpace(row.CapabilityCode) == "" ||
		strings.TrimSpace(row.ModelType) == "" || !validRouteStrategy(row.Strategy) || row.TimeoutMS <= 0 || row.MaxRetry < 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateRouteModel(row models.AIBaseRouteModel) error {
	if strings.TrimSpace(row.BaseRouteID) == "" || strings.TrimSpace(row.ModelID) == "" || !validRouteModelRole(row.Role) ||
		row.Priority <= 0 || row.Weight <= 0 || row.MaxRetry < 0 || row.TimeoutMS <= 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateScenario(row models.AIScenario) error {
	if strings.TrimSpace(row.AppCode) == "" || strings.TrimSpace(row.AppName) == "" || strings.TrimSpace(row.AIScenarioCode) == "" ||
		strings.TrimSpace(row.AIScenarioName) == "" || strings.TrimSpace(row.ScenarioType) == "" || strings.TrimSpace(row.CapabilityCode) == "" ||
		strings.TrimSpace(row.ModelType) == "" || strings.TrimSpace(row.DefaultBaseRouteID) == "" {
		return ErrInvalidInput
	}
	return nil
}

func validateTenantStrategy(row models.AITenantStrategyPolicy) error {
	if strings.TrimSpace(row.PolicyName) == "" || !validTenantScope(row.TenantScope) || strings.TrimSpace(row.AppCode) == "" ||
		strings.TrimSpace(row.AIScenarioCode) == "" || strings.TrimSpace(row.DefaultBaseRouteID) == "" {
		return ErrInvalidInput
	}
	if row.TenantScope == "include" && len(row.TenantIDs) == 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateQuotaRule(row models.AIStrategyQuotaRule) error {
	if strings.TrimSpace(row.PolicyID) == "" || strings.TrimSpace(row.SubjectCode) == "" || strings.TrimSpace(row.UsageUnit) == "" ||
		!validControlDimension(row.Dimension) || !validControlPeriod(row.Period) || !validOverLimitAction(row.OverLimitAction) ||
		row.QuotaLimit <= 0 || row.UsedAmount < 0 || row.WarningThreshold <= 0 || row.WarningThreshold > 100 {
		return ErrInvalidInput
	}
	return nil
}

func validateRateLimitRule(row models.AIStrategyRateLimitRule) error {
	if strings.TrimSpace(row.PolicyID) == "" || strings.TrimSpace(row.SubjectCode) == "" ||
		!validControlDimension(row.Dimension) || !validOverLimitAction(row.OverLimitAction) ||
		row.QPS < 0 || row.Concurrency < 0 || row.MinuteLimit < 0 || row.HourLimit < 0 || row.DayLimit < 0 ||
		(row.QPS == 0 && row.Concurrency == 0 && row.MinuteLimit == 0 && row.HourLimit == 0 && row.DayLimit == 0) {
		return ErrInvalidInput
	}
	return nil
}

func validateGatewaySetting(row models.AIGatewaySetting) error {
	key := strings.TrimSpace(row.SettingKey)
	if key == "" || strings.TrimSpace(row.SettingValue) == "" || !json.Valid([]byte(row.SettingValue)) {
		return ErrInvalidInput
	}
	var value map[string]interface{}
	if err := json.Unmarshal([]byte(row.SettingValue), &value); err != nil {
		return ErrInvalidInput
	}
	switch key {
	case "gateway_runtime":
		timeout := numberParam(value, "default_timeout_ms", 0)
		retry := numberParam(value, "default_max_retry", 0)
		if timeout <= 0 || retry < 0 {
			return ErrInvalidInput
		}
		if _, ok := value["usage_log_async"].(bool); !ok {
			return ErrInvalidInput
		}
	case "security":
		if _, ok := value["prompt_plaintext_storage"].(bool); !ok {
			return ErrInvalidInput
		}
		if strings.TrimSpace(fmt.Sprint(value["api_key_encryption"])) == "" {
			return ErrInvalidInput
		}
	default:
		if strings.Contains(key, "timeout") && numberParam(value, "default_timeout_ms", 1) <= 0 {
			return ErrInvalidInput
		}
	}
	return nil
}

func normalizeSettingPayload(payload map[string]interface{}) {
	value, ok := payload["setting_value"]
	if !ok {
		return
	}
	if _, ok := value.(string); ok {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	payload["setting_value"] = string(raw)
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

func (s *Service) deleteTenantStrategyCascade(ctx context.Context, id string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var policy models.AITenantStrategyPolicy
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&policy).Error; err != nil {
			return ErrNotFound
		}
		updates := map[string]interface{}{"deleted_at": now, "updated_at": now, "status": "inactive"}
		if err := tx.Model(&models.AIStrategyQuotaRule{}).Where("policy_id = ? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AIStrategyRateLimitRule{}).Where("policy_id = ? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		res := tx.Model(&models.AITenantStrategyPolicy{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
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

func (s *Service) routeModelEndpointRefsValid(ctx context.Context, row models.AIBaseRouteModel) bool {
	return routeModelEndpointRefsValidTx(ctx, s.db, row)
}

func routeModelEndpointRefsValidTx(ctx context.Context, tx *gorm.DB, row models.AIBaseRouteModel) bool {
	accountID := strings.TrimSpace(row.ProviderAccountID)
	apiID := strings.TrimSpace(row.ProviderAPIID)
	if accountID == "" && apiID == "" {
		return true
	}
	var model models.AIModel
	if err := tx.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", row.ModelID).First(&model).Error; err != nil {
		return false
	}
	if accountID != "" {
		var count int64
		_ = tx.WithContext(ctx).Model(&models.AIProviderAccount{}).
			Where("id = ? AND provider_id = ? AND deleted_at IS NULL", accountID, model.ProviderID).
			Count(&count).Error
		if count == 0 {
			return false
		}
	}
	if apiID != "" {
		q := tx.WithContext(ctx).Model(&models.AIProviderAPI{}).
			Where("id = ? AND provider_id = ? AND deleted_at IS NULL", apiID, model.ProviderID)
		if accountID != "" {
			q = q.Where("account_id = ?", accountID)
		}
		var count int64
		_ = q.Count(&count).Error
		if count == 0 {
			return false
		}
	}
	return true
}

func nullableString(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
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

func scenarioExistsTx(ctx context.Context, tx *gorm.DB, appCode, scenarioCode string) bool {
	if strings.TrimSpace(appCode) == "" || strings.TrimSpace(scenarioCode) == "" {
		return false
	}
	var count int64
	_ = tx.WithContext(ctx).Model(&models.AIScenario{}).Where("app_code = ? AND ai_scenario_code = ? AND deleted_at IS NULL", strings.TrimSpace(appCode), strings.TrimSpace(scenarioCode)).Count(&count).Error
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

func updateRow[T any](ctx context.Context, s *Service, userID uint64, resource, id string, payload map[string]interface{}, module string, validate func(T) error) (interface{}, error) {
	var before T
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if validate != nil {
		var merged T
		if err := mergePayload(before, payload, &merged); err != nil {
			return nil, err
		}
		if err := validate(merged); err != nil {
			return nil, err
		}
	}
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
	s.Audit(ctx, userID, module, "update", "更新 AI 能力中心资源："+resource, map[string]interface{}{"id": id, "before": before, "after": row, "patch": payload})
	return row, nil
}

func (s *Service) updateRouteModel(ctx context.Context, userID uint64, resource, id string, payload map[string]interface{}) (interface{}, error) {
	var before models.AIBaseRouteModel
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var merged models.AIBaseRouteModel
	if err := mergePayload(before, payload, &merged); err != nil {
		return nil, err
	}
	if err := validateRouteModel(merged); err != nil {
		return nil, err
	}
	if !s.exists(ctx, &models.AIBaseRoute{}, merged.BaseRouteID) || !s.exists(ctx, &models.AIModel{}, merged.ModelID) || !s.routeModelEndpointRefsValid(ctx, merged) {
		return nil, ErrInvalidInput
	}
	if _, ok := payload["provider_account_id"]; ok {
		payload["provider_account_id"] = nullableString(merged.ProviderAccountID)
	}
	if _, ok := payload["provider_api_id"]; ok {
		payload["provider_api_id"] = nullableString(merged.ProviderAPIID)
	}
	var row models.AIBaseRouteModel
	res := s.db.WithContext(ctx).Model(&models.AIBaseRouteModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(payload)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, err
	}
	s.Audit(ctx, userID, "ai_base_route_model", "update", "更新 AI 能力中心资源："+resource, map[string]interface{}{"id": id, "before": before, "after": row, "patch": payload})
	return row, nil
}

func mergePayload(before interface{}, patch map[string]interface{}, target interface{}) error {
	raw, err := json.Marshal(before)
	if err != nil {
		return err
	}
	var merged map[string]interface{}
	if err := json.Unmarshal(raw, &merged); err != nil {
		return err
	}
	for key, value := range patch {
		merged[key] = value
	}
	return decodePayload(merged, target)
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

func containsString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}

func strategyID(strategy models.AITenantStrategyPolicy, found bool) string {
	if !found {
		return ""
	}
	return strategy.ID
}

func controlRuleMatches(dimension, subject string, req InvokeRequest, scenario models.AIScenario, routeModel models.AIBaseRouteModel, pricing pricingResult, accountID, apiID string) bool {
	if strings.TrimSpace(subject) == "" || subject == "*" {
		return true
	}
	switch strings.TrimSpace(dimension) {
	case "tenant":
		return subject == req.TenantID
	case "app":
		return subject == scenario.AppCode
	case "scenario":
		return subject == scenario.AIScenarioCode
	case "model":
		return subject == routeModel.ModelID
	case "feature_sku":
		return subject == pricing.FeatureKey || subject == pricing.PolicyID || subject == pricing.TierID
	case "provider_account":
		return subject == accountID
	case "user":
		return subject == req.UserID
	case "amount":
		return subject == pricing.UsageUnit
	case "api":
		return subject == apiID
	default:
		return false
	}
}

func periodStart(now time.Time, period string) *time.Time {
	var start time.Time
	switch strings.TrimSpace(period) {
	case "minute":
		start = now.Truncate(time.Minute)
	case "hour":
		start = now.Truncate(time.Hour)
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		start = dayStart.AddDate(0, 0, -(weekday - 1))
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case "total":
		return nil
	default:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	return &start
}

func numberParam(params map[string]interface{}, key string, fallback float64) float64 {
	if params == nil {
		return fallback
	}
	switch value := params[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		parsed, err := value.Float64()
		if err == nil {
			return parsed
		}
	case string:
		var parsed float64
		if _, err := fmt.Sscanf(value, "%f", &parsed); err == nil {
			return parsed
		}
	}
	return fallback
}

func stringParam(params map[string]interface{}, key string) string {
	if params == nil {
		return ""
	}
	if _, ok := params[key]; !ok || params[key] == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(params[key]))
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

func defaultFloat(value, fallback float64) float64 {
	if value == 0 {
		return fallback
	}
	return value
}

func jsonString(value interface{}) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func validRouteStrategy(strategy string) bool {
	switch strings.TrimSpace(strategy) {
	case "fixed", "fallback", "priority", "load_balance", "cost_first", "quality_first", "latency_first", "quota_aware", "tenant_custom", "capability_match":
		return true
	default:
		return false
	}
}

func validTenantScope(scope string) bool {
	switch strings.TrimSpace(scope) {
	case "all", "include", "exclude":
		return true
	default:
		return false
	}
}

func validControlDimension(dimension string) bool {
	switch strings.TrimSpace(dimension) {
	case "tenant", "app", "scenario", "model", "feature_sku", "provider_account", "user", "amount", "api":
		return true
	default:
		return false
	}
}

func validControlPeriod(period string) bool {
	switch strings.TrimSpace(period) {
	case "minute", "hour", "day", "week", "month", "year", "total":
		return true
	default:
		return false
	}
}

func validOverLimitAction(action string) bool {
	switch strings.TrimSpace(action) {
	case "alert_only", "degrade_route", "queue", "reject", "approval":
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
	raw, _ := json.Marshal(maskAIAuditDetail(detail))
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

func maskAIAuditDetail(value interface{}) interface{} {
	raw, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return value
	}
	return maskAIAuditValue(decoded)
}

func maskAIAuditValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		masked := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			if isSensitiveAIKey(key) {
				masked[key] = "***"
				continue
			}
			masked[key] = maskAIAuditValue(item)
		}
		return masked
	case []interface{}:
		masked := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			masked = append(masked, maskAIAuditValue(item))
		}
		return masked
	default:
		return value
	}
}

func isSensitiveAIKey(key string) bool {
	normalized := strings.ToLower(key)
	return strings.Contains(normalized, "api_key") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "authorization")
}

func stringPtr(value string) *string { return &value }
