package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestOverviewAggregatesUsageMetrics(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	modelText := "11111111-1111-1111-1111-111111111111"
	modelImage := "22222222-2222-2222-2222-222222222222"
	require.NoError(t, db.Create(&models.AIModel{
		ID: modelText, ProviderID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", ModelCode: "gpt-text", ModelName: "GPT Text",
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIModel{
		ID: modelImage, ProviderID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", ModelCode: "image-gen", ModelName: "Image Gen",
		ModelType: "image", Capabilities: []string{"image_generation"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIBaseRoute{
		ID: "33333333-3333-3333-3333-333333333333", RouteCode: "chat-default", RouteName: "对话默认路由",
		CapabilityCode: "chat_completion", ModelType: "text", Strategy: "priority", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	records := []models.AIUsageRecord{
		usageRecord("r-today-1", "tenant-a", "租户 A", "app-a", "chat", modelText, 3, 1.25, 2.50, 120, "success", todayStart.Add(9*time.Hour)),
		usageRecord("r-today-2", "tenant-a", "租户 A", "app-a", "chat", modelText, 2, 0.75, 1.50, 320, "success", todayStart.Add(10*time.Hour)),
		usageRecord("r-today-3", "tenant-b", "租户 B", "app-b", "image", modelImage, 1, 4.25, 0.50, 900, "error", todayStart.Add(11*time.Hour)),
		usageRecord("r-yesterday", "tenant-b", "租户 B", "app-b", "chat", modelText, 4, 2.00, 3.00, 260, "success", todayStart.AddDate(0, 0, -1).Add(12*time.Hour)),
		usageRecord("r-old", "tenant-c", "租户 C", "app-c", "chat", modelText, 10, 9.00, 12.00, 100, "success", todayStart.AddDate(0, 0, -8)),
	}
	require.NoError(t, db.Create(&records).Error)

	overview, err := NewService(db).Overview(context.Background())
	require.NoError(t, err)

	require.Equal(t, "6", overview.Metrics[0].Value)
	require.Equal(t, "¥6.25", overview.Metrics[1].Value)
	require.Equal(t, "66.67%", overview.Metrics[2].Value)
	require.Equal(t, "900ms", overview.Metrics[3].Value)
	require.Equal(t, 7, len(overview.UsageTrend))
	require.Equal(t, todayStart.Format("2006-01-02"), overview.UsageTrend[6].Date)
	require.Equal(t, int64(6), overview.UsageTrend[6].Calls)
	require.InDelta(t, 6.25, overview.UsageTrend[6].CostAmount, 0.0001)
	require.Len(t, overview.ModelCostShare, 2)
	require.Equal(t, "image", overview.ModelCostShare[0].ModelType)
	require.Len(t, overview.TenantRanking, 2)
	require.Equal(t, "租户 A", overview.TenantRanking[0].TenantName)
	require.Equal(t, int64(5), overview.TenantRanking[0].Calls)
	require.InDelta(t, 100.0, overview.TenantRanking[0].SuccessRate, 0.0001)
	require.Equal(t, int64(2), overview.TenantMetrics["service_tenants"])
	require.InDelta(t, 4.5, overview.TenantMetrics["tenant_revenue"], 0.0001)
	require.InDelta(t, -1.75, overview.TenantMetrics["tenant_profit"], 0.0001)
	require.Len(t, overview.CoreBaseRoutes, 1)
}

func TestOverviewReturnsEmptySeriesWithoutUsage(t *testing.T) {
	db := newAICapabilityTestDB(t)
	overview, err := NewService(db).Overview(context.Background())
	require.NoError(t, err)

	require.Equal(t, "0", overview.Metrics[0].Value)
	require.Equal(t, "¥0.00", overview.Metrics[1].Value)
	require.Equal(t, "100.00%", overview.Metrics[2].Value)
	require.Equal(t, "0ms", overview.Metrics[3].Value)
	require.Len(t, overview.UsageTrend, 7)
	require.Empty(t, overview.ModelCostShare)
	require.Empty(t, overview.TenantRanking)
}

func TestOverviewWarnsWhenScenarioRouteBindingUnavailable(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.AIScenario{
		ID: "scenario-broken", AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: "missing-route", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	overview, err := NewService(db).Overview(context.Background())
	require.NoError(t, err)
	require.Contains(t, overview.HealthChecks, HealthCheck{
		Name:    "AI 场景绑定状态",
		Status:  "warning",
		Message: "1 个启用 AI 场景未绑定可用基础路由或模型池节点",
	})
}

func TestCheckProviderAPIConnectivityPersistsReachableStatus(t *testing.T) {
	db := newAICapabilityTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-openai", Name: "OpenAI 企业账号", Code: "openai", Type: "public_cloud", BaseURL: server.URL,
		AuthType: "api_key", Status: "active", Priority: 10, QPSLimit: 180,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAccount{
		ID: "account-prod", ProviderID: "provider-openai", AccountName: "prod-main", Endpoint: server.URL,
		KeyAlias: "OPENAI_PROD_KEY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAPI{
		ID: "api-chat", ProviderID: "provider-openai", AccountID: "account-prod", APIName: "chat.completions",
		APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
		AuthType: "api_key", QPSLimit: 160, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Active: 1}, result)

	var api models.AIProviderAPI
	require.NoError(t, db.First(&api, "id = ?", "api-chat").Error)
	require.Equal(t, "active", api.HealthStatus)
	require.Contains(t, api.HealthMessage, "HTTP 401")
	require.NotNil(t, api.HealthCheckedAt)
}

func TestImportProvidersUpsertsProviderAccountsAndAPIs(t *testing.T) {
	db := newAICapabilityTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	result, err := service.ImportProviders(ctx, 7, ProviderImportRequest{
		Providers: []ProviderImportProvider{{
			Name: "OpenAI", Code: "openai", Type: "public_cloud", BaseURL: "https://api.openai.com", Priority: 10,
			Accounts: []ProviderImportAccount{{
				AccountName: "prod", Endpoint: "https://api.openai.com", KeyAlias: "OPENAI_API_KEY", EncryptedAPIKey: "ciphertext",
				APIs: []ProviderImportAPI{{
					APIName: "chat.completions", APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"}, TimeoutMS: 45000,
				}},
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ProviderImportResult{Providers: 1, Accounts: 1, APIs: 1}, result)

	var provider models.AIProvider
	require.NoError(t, db.Where("code = ? AND deleted_at IS NULL", "openai").First(&provider).Error)
	require.Equal(t, "OpenAI", provider.Name)
	require.Equal(t, 10, provider.Priority)

	var account models.AIProviderAccount
	require.NoError(t, db.Where("provider_id = ? AND account_name = ? AND deleted_at IS NULL", provider.ID, "prod").First(&account).Error)
	require.Equal(t, "OPENAI_API_KEY", account.KeyAlias)
	require.Equal(t, "ciphertext", account.EncryptedAPIKey)

	var api models.AIProviderAPI
	require.NoError(t, db.Where("provider_id = ? AND account_id = ? AND api_name = ? AND deleted_at IS NULL", provider.ID, account.ID, "chat.completions").First(&api).Error)
	require.Equal(t, "/v1/chat/completions", api.APIPath)
	require.Equal(t, 45000, api.TimeoutMS)
	require.Equal(t, []string{"chat_completion"}, api.Capabilities)

	result, err = service.ImportProviders(ctx, 7, ProviderImportRequest{
		Providers: []ProviderImportProvider{{
			Name: "OpenAI Updated", Code: "openai", Type: "public_cloud", BaseURL: "https://gateway.openai.example", Priority: 20,
		}},
		Accounts: []ProviderImportAccount{{
			ProviderCode: "openai", AccountName: "prod", Endpoint: "https://gateway.openai.example", KeyAlias: "OPENAI_PRIMARY", EncryptedAPIKey: "ciphertext-v2",
		}},
		APIs: []ProviderImportAPI{{
			ProviderCode: "openai", AccountName: "prod", APIName: "chat.completions", APIPath: "/proxy/chat", APIType: "chat", Capabilities: []string{"chat_completion", "embedding"},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ProviderImportResult{Providers: 1, Accounts: 1, APIs: 1}, result)

	var count int64
	require.NoError(t, db.Model(&models.AIProvider{}).Where("code = ? AND deleted_at IS NULL", "openai").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIProviderAccount{}).Where("provider_id = ? AND account_name = ? AND deleted_at IS NULL", provider.ID, "prod").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIProviderAPI{}).Where("provider_id = ? AND account_id = ? AND api_name = ? AND deleted_at IS NULL", provider.ID, account.ID, "chat.completions").Count(&count).Error)
	require.Equal(t, int64(1), count)

	require.NoError(t, db.Where("id = ?", provider.ID).First(&provider).Error)
	require.Equal(t, "OpenAI Updated", provider.Name)
	require.Equal(t, "https://gateway.openai.example", provider.BaseURL)
	require.Equal(t, 20, provider.Priority)
	require.NoError(t, db.Where("id = ?", account.ID).First(&account).Error)
	require.Equal(t, "OPENAI_PRIMARY", account.KeyAlias)
	require.Equal(t, "ciphertext-v2", account.EncryptedAPIKey)
	require.NoError(t, db.Where("id = ?", api.ID).First(&api).Error)
	require.Equal(t, "/proxy/chat", api.APIPath)
	require.Equal(t, []string{"chat_completion", "embedding"}, api.Capabilities)

	raw, err := json.Marshal(account)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "encrypted_api_key")
	require.NotContains(t, string(raw), "ciphertext-v2")
}

func TestImportProvidersRollsBackOnInvalidReference(t *testing.T) {
	db := newAICapabilityTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	_, err := service.ImportProviders(ctx, 7, ProviderImportRequest{
		Providers: []ProviderImportProvider{{
			Name: "Rollback Provider", Code: "rollback", BaseURL: "https://rollback.example",
		}},
		Accounts: []ProviderImportAccount{{
			ProviderCode: "missing-provider", AccountName: "prod", KeyAlias: "MISSING_KEY",
		}},
	})
	require.ErrorIs(t, err, ErrInvalidInput)

	var count int64
	require.NoError(t, db.Model(&models.AIProvider{}).Where("code = ?", "rollback").Count(&count).Error)
	require.Equal(t, int64(0), count)
	require.NoError(t, db.Model(&models.AIProviderAccount{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestDeleteProviderCascadesAccountsAndAPIsAndBlocksReferences(t *testing.T) {
	db := newAICapabilityTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	provider, account, api := seedProviderAccountAPI(t, db, "cascade", "cascade-account")

	require.NoError(t, service.DeleteResource(ctx, 7, "providers", provider.ID))

	var deletedProvider models.AIProvider
	require.NoError(t, db.Where("id = ?", provider.ID).First(&deletedProvider).Error)
	require.NotNil(t, deletedProvider.DeletedAt)
	require.Equal(t, "inactive", deletedProvider.Status)
	var deletedAccount models.AIProviderAccount
	require.NoError(t, db.Where("id = ?", account.ID).First(&deletedAccount).Error)
	require.NotNil(t, deletedAccount.DeletedAt)
	require.Equal(t, "inactive", deletedAccount.Status)
	var deletedAPI models.AIProviderAPI
	require.NoError(t, db.Where("id = ?", api.ID).First(&deletedAPI).Error)
	require.NotNil(t, deletedAPI.DeletedAt)
	require.Equal(t, "inactive", deletedAPI.Status)

	blockedProvider, _, _ := seedProviderAccountAPI(t, db, "blocked", "blocked-account")
	now := time.Now()
	require.NoError(t, db.Create(&models.AIModel{
		ID: "99999999-9999-9999-9999-999999999999", ProviderID: blockedProvider.ID, ModelCode: "blocked-model", ModelName: "Blocked Model",
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	err := service.DeleteResource(ctx, 7, "providers", blockedProvider.ID)
	require.ErrorIs(t, err, ErrResourceInUse)
}

func TestDeleteProviderAccountCascadesAPIsAndBlocksUsageReferences(t *testing.T) {
	db := newAICapabilityTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	_, account, api := seedProviderAccountAPI(t, db, "account-cascade", "prod")

	require.NoError(t, service.DeleteResource(ctx, 7, "accounts", account.ID))

	var deletedAccount models.AIProviderAccount
	require.NoError(t, db.Where("id = ?", account.ID).First(&deletedAccount).Error)
	require.NotNil(t, deletedAccount.DeletedAt)
	require.Equal(t, "inactive", deletedAccount.Status)
	var deletedAPI models.AIProviderAPI
	require.NoError(t, db.Where("id = ?", api.ID).First(&deletedAPI).Error)
	require.NotNil(t, deletedAPI.DeletedAt)
	require.Equal(t, "inactive", deletedAPI.Status)

	_, blockedAccount, _ := seedProviderAccountAPI(t, db, "usage-blocked", "prod")
	record := usageRecord("usage-block-account", "tenant-a", "租户 A", "app-a", "chat", "", 1, 1, 1, 120, "success", time.Now())
	record.ProviderAccountID = blockedAccount.ID
	require.NoError(t, db.Create(&record).Error)

	err := service.DeleteResource(ctx, 7, "accounts", blockedAccount.ID)
	require.ErrorIs(t, err, ErrResourceInUse)
}

func TestImportModelsUpsertsModelPoliciesAndTiers(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	seedProvider(t, db, "openai")
	service := NewService(db)
	ctx := context.Background()
	enabled := false

	result, err := service.ImportModels(ctx, 7, ModelImportRequest{
		Models: []ModelImportModel{{
			ProviderCode: "openai", ModelCode: "gpt-4.1", ModelName: "GPT 4.1", ModelType: "text",
			Capabilities: []string{"chat_completion"}, ContextWindow: 128000, Unit: "tokens", SuccessRate: 99.5,
			PricePolicies: []ModelImportPricePolicy{{
				FeatureKey: "chat_tokens", FeatureName: "Chat Tokens", ModelType: "text", CapabilityCode: "chat_completion",
				BillingMode: "tiered", BillingUnit: "tokens", PlatformUnit: "tokens", BaseCostPrice: 0.01, BaseSalePrice: 0.02,
				Tiers: []ModelImportPriceTier{{
					TierName: "standard", Mode: "sync", CostPrice: 0.01, SalePrice: 0.02, PlatformAmount: 0.01, Enabled: &enabled, SortOrder: 10,
				}},
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ModelImportResult{Models: 1, PricePolicies: 1, PriceTiers: 1}, result)

	var provider models.AIProvider
	require.NoError(t, db.Where("code = ?", "openai").First(&provider).Error)
	var model models.AIModel
	require.NoError(t, db.Where("provider_id = ? AND model_code = ? AND deleted_at IS NULL", provider.ID, "gpt-4.1").First(&model).Error)
	require.Equal(t, "GPT 4.1", model.ModelName)
	require.Equal(t, 128000, model.ContextWindow)
	require.Equal(t, []string{"chat_completion"}, model.Capabilities)
	var policy models.AIModelPricePolicy
	require.NoError(t, db.Where("model_id = ? AND feature_key = ? AND deleted_at IS NULL", model.ID, "chat_tokens").First(&policy).Error)
	require.Equal(t, "Chat Tokens", policy.FeatureName)
	require.InDelta(t, 0.02, policy.BaseSalePrice, 0.0001)
	var tier models.AIModelPriceTier
	require.NoError(t, db.Where("price_policy_id = ? AND tier_name = ? AND deleted_at IS NULL", policy.ID, "standard").First(&tier).Error)
	require.False(t, tier.Enabled)
	require.Equal(t, 10, tier.SortOrder)

	result, err = service.ImportModels(ctx, 7, ModelImportRequest{
		Models: []ModelImportModel{{
			ProviderCode: "openai", ModelCode: "gpt-4.1", ModelName: "GPT 4.1 Turbo", ModelType: "text",
			Capabilities: []string{"chat_completion"}, ContextWindow: 256000, Unit: "tokens",
		}},
		PricePolicies: []ModelImportPricePolicy{{
			ProviderCode: "openai", ModelCode: "gpt-4.1", FeatureKey: "chat_tokens", FeatureName: "Chat Tokens Updated",
			ModelType: "text", CapabilityCode: "chat_completion", BillingMode: "tiered", BillingUnit: "tokens", PlatformUnit: "tokens",
			BaseCostPrice: 0.02, BaseSalePrice: 0.03,
		}},
		PriceTiers: []ModelImportPriceTier{{
			ProviderCode: "openai", ModelCode: "gpt-4.1", FeatureKey: "chat_tokens", TierName: "standard",
			Mode: "sync", CostPrice: 0.02, SalePrice: 0.03, PlatformAmount: 0.01, SortOrder: 20,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ModelImportResult{Models: 1, PricePolicies: 1, PriceTiers: 1}, result)

	var count int64
	require.NoError(t, db.Model(&models.AIModel{}).Where("provider_id = ? AND model_code = ? AND deleted_at IS NULL", provider.ID, "gpt-4.1").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIModelPricePolicy{}).Where("model_id = ? AND feature_key = ? AND deleted_at IS NULL", model.ID, "chat_tokens").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIModelPriceTier{}).Where("price_policy_id = ? AND tier_name = ? AND deleted_at IS NULL", policy.ID, "standard").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Where("id = ?", model.ID).First(&model).Error)
	require.Equal(t, "GPT 4.1 Turbo", model.ModelName)
	require.Equal(t, 256000, model.ContextWindow)
	require.NoError(t, db.Where("id = ?", policy.ID).First(&policy).Error)
	require.Equal(t, "Chat Tokens Updated", policy.FeatureName)
	require.InDelta(t, 0.03, policy.BaseSalePrice, 0.0001)
	require.NoError(t, db.Where("id = ?", tier.ID).First(&tier).Error)
	require.True(t, tier.Enabled)
	require.Equal(t, 20, tier.SortOrder)
}

func TestImportModelsRollsBackOnInvalidReference(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedProvider(t, db, "openai")
	service := NewService(db)
	ctx := context.Background()

	_, err := service.ImportModels(ctx, 7, ModelImportRequest{
		Models: []ModelImportModel{{
			ProviderCode: "openai", ModelCode: "broken", ModelName: "Broken", ModelType: "text", Capabilities: []string{"chat_completion"},
		}},
		PricePolicies: []ModelImportPricePolicy{{
			ProviderCode: "openai", ModelCode: "broken", FeatureKey: "missing_cap", FeatureName: "Missing Capability",
			ModelType: "text", CapabilityCode: "missing_capability", BillingMode: "per_unit", BillingUnit: "tokens", PlatformUnit: "tokens",
		}},
	})
	require.ErrorIs(t, err, ErrInvalidInput)

	var count int64
	require.NoError(t, db.Model(&models.AIModel{}).Where("model_code = ?", "broken").Count(&count).Error)
	require.Equal(t, int64(0), count)
	require.NoError(t, db.Model(&models.AIModelPricePolicy{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestImportScenariosUpsertsAndValidatesReferences(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	service := NewService(db)
	ctx := context.Background()

	result, err := service.ImportScenarios(ctx, 7, ScenarioImportRequest{
		Scenarios: []ScenarioImportItem{{
			AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "商品文案生成",
			ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: route.ID, Owner: "AI 平台组",
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ScenarioImportResult{Scenarios: 1}, result)

	var scenario models.AIScenario
	require.NoError(t, db.Where("app_code = ? AND ai_scenario_code = ? AND deleted_at IS NULL", "product_center", "copy_gen").First(&scenario).Error)
	require.Equal(t, "商品文案生成", scenario.AIScenarioName)
	require.Equal(t, route.ID, scenario.DefaultBaseRouteID)

	result, err = service.ImportScenarios(ctx, 7, ScenarioImportRequest{
		Scenarios: []ScenarioImportItem{{
			AppCode: "product_center", AppName: "商品中心更新", AIScenarioCode: "copy_gen", AIScenarioName: "商品标题生成",
			ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: route.ID, Owner: "增长组", Version: "v2.0",
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ScenarioImportResult{Scenarios: 1}, result)

	var count int64
	require.NoError(t, db.Model(&models.AIScenario{}).Where("app_code = ? AND ai_scenario_code = ? AND deleted_at IS NULL", "product_center", "copy_gen").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Where("id = ?", scenario.ID).First(&scenario).Error)
	require.Equal(t, "商品标题生成", scenario.AIScenarioName)
	require.Equal(t, "增长组", scenario.Owner)
	require.Equal(t, "v2.0", scenario.Version)
}

func TestImportScenariosRollsBackOnInvalidReference(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	service := NewService(db)
	ctx := context.Background()

	_, err := service.ImportScenarios(ctx, 7, ScenarioImportRequest{
		Scenarios: []ScenarioImportItem{
			{
				AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "商品文案生成",
				ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: route.ID,
			},
			{
				AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "broken", AIScenarioName: "错误场景",
				ScenarioType: "text", CapabilityCode: "missing", ModelType: "text", DefaultBaseRouteID: route.ID,
			},
		},
	})
	require.ErrorIs(t, err, ErrInvalidInput)

	var count int64
	require.NoError(t, db.Model(&models.AIScenario{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestImportRoutesUpsertsBaseRoutesAndModelPool(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	seedProvider(t, db, "openai")
	model := seedModel(t, db, "openai", "gpt-4.1")
	service := NewService(db)
	ctx := context.Background()

	result, err := service.ImportRoutes(ctx, 7, RouteImportRequest{
		BaseRoutes: []RouteImportBaseRoute{{
			RouteCode: "chat-default", RouteName: "对话默认路由", CapabilityCode: "chat_completion", ModelType: "text",
			Strategy: "fallback", TimeoutMS: 30000, MaxRetry: 2,
			RouteModels: []RouteImportRouteModel{{
				ProviderCode: "openai", ModelCode: "gpt-4.1", Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 25000,
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, RouteImportResult{BaseRoutes: 1, RouteModels: 1}, result)

	var route models.AIBaseRoute
	require.NoError(t, db.Where("route_code = ? AND deleted_at IS NULL", "chat-default").First(&route).Error)
	require.Equal(t, "fallback", route.Strategy)
	var routeModel models.AIBaseRouteModel
	require.NoError(t, db.Where("base_route_id = ? AND model_id = ? AND role = ? AND deleted_at IS NULL", route.ID, model.ID, "primary").First(&routeModel).Error)
	require.Equal(t, 100, routeModel.Weight)
	require.Equal(t, 25000, routeModel.TimeoutMS)

	result, err = service.ImportRoutes(ctx, 7, RouteImportRequest{
		BaseRoutes: []RouteImportBaseRoute{{
			RouteCode: "chat-default", RouteName: "对话优先路由", CapabilityCode: "chat_completion", ModelType: "text",
			Strategy: "priority", TimeoutMS: 45000, MaxRetry: 1,
		}},
		RouteModels: []RouteImportRouteModel{{
			BaseRouteCode: "chat-default", ProviderCode: "openai", ModelCode: "gpt-4.1", Role: "primary", Priority: 2, Weight: 80, MaxRetry: 1, TimeoutMS: 35000,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, RouteImportResult{BaseRoutes: 1, RouteModels: 1}, result)

	var count int64
	require.NoError(t, db.Model(&models.AIBaseRoute{}).Where("route_code = ? AND deleted_at IS NULL", "chat-default").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).Where("base_route_id = ? AND model_id = ? AND role = ? AND deleted_at IS NULL", route.ID, model.ID, "primary").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Where("id = ?", route.ID).First(&route).Error)
	require.Equal(t, "对话优先路由", route.RouteName)
	require.Equal(t, "priority", route.Strategy)
	require.NoError(t, db.Where("id = ?", routeModel.ID).First(&routeModel).Error)
	require.Equal(t, 2, routeModel.Priority)
	require.Equal(t, 80, routeModel.Weight)
}

func TestImportRoutesRollsBackOnInvalidModelPool(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	service := NewService(db)
	ctx := context.Background()

	_, err := service.ImportRoutes(ctx, 7, RouteImportRequest{
		BaseRoutes: []RouteImportBaseRoute{{
			RouteCode: "chat-default", RouteName: "对话默认路由", CapabilityCode: "chat_completion", ModelType: "text",
			Strategy: "load_balance", TimeoutMS: 30000,
			RouteModels: []RouteImportRouteModel{{
				ModelID: "missing-model", Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 25000,
			}},
		}},
	})
	require.ErrorIs(t, err, ErrInvalidInput)

	var count int64
	require.NoError(t, db.Model(&models.AIBaseRoute{}).Where("route_code = ?", "chat-default").Count(&count).Error)
	require.Equal(t, int64(0), count)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestDeleteBaseRouteBlocksReferencesAndCascadesModelPool(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	seedProvider(t, db, "openai")
	model := seedModel(t, db, "openai", "gpt-4.1")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	now := time.Now()
	routeModel := models.AIBaseRouteModel{
		ID: "route-model-primary", BaseRouteID: route.ID, ModelID: model.ID, Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&routeModel).Error)
	scenario := models.AIScenario{
		ID: "scenario-route", AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: route.ID, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&scenario).Error)

	service := NewService(db)
	err := service.DeleteResource(context.Background(), 7, "base-routes", route.ID)
	require.ErrorIs(t, err, ErrResourceInUse)

	require.NoError(t, db.Model(&models.AIScenario{}).Where("id = ?", scenario.ID).Update("deleted_at", now).Error)
	err = service.DeleteResource(context.Background(), 7, "base-routes", route.ID)
	require.NoError(t, err)

	var deletedRoute models.AIBaseRoute
	require.NoError(t, db.Where("id = ?", route.ID).First(&deletedRoute).Error)
	require.NotNil(t, deletedRoute.DeletedAt)
	var deletedRouteModel models.AIBaseRouteModel
	require.NoError(t, db.Where("id = ?", routeModel.ID).First(&deletedRouteModel).Error)
	require.NotNil(t, deletedRouteModel.DeletedAt)
}

func TestImportTenantStrategiesUpsertsPoliciesAndRules(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	overrideRoute := seedBaseRoute(t, db, "route-chat-fast", "chat_completion")
	seedScenario(t, db, route.ID)
	service := NewService(db)
	ctx := context.Background()

	result, err := service.ImportTenantStrategies(ctx, 7, TenantStrategyImportRequest{
		Policies: []TenantStrategyImportPolicy{{
			PolicyName: "重点租户策略", TenantScope: "include", TenantIDs: []string{"tenant-a"},
			AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
			DefaultBaseRouteID: route.ID, OverrideBaseRouteID: overrideRoute.ID,
			QuotaRules: []TenantStrategyImportQuotaRule{{
				Dimension: "scenario", SubjectCode: "copy_gen", UsageUnit: "tokens", Period: "day",
				QuotaLimit: 1000, WarningThreshold: 80, OverLimitAction: "alert_only",
			}},
			RateLimitRules: []TenantStrategyImportRateLimitRule{{
				Dimension: "user", SubjectCode: "user-a", QPS: 20, Concurrency: 5, OverLimitAction: "queue",
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, TenantStrategyImportResult{Policies: 1, QuotaRules: 1, RateLimitRules: 1}, result)

	var policy models.AITenantStrategyPolicy
	require.NoError(t, db.Where("policy_name = ? AND deleted_at IS NULL", "重点租户策略").First(&policy).Error)
	require.Equal(t, []string{"tenant-a"}, policy.TenantIDs)
	require.Equal(t, overrideRoute.ID, policy.OverrideBaseRouteID)
	var quota models.AIStrategyQuotaRule
	require.NoError(t, db.Where("policy_id = ? AND dimension = ? AND deleted_at IS NULL", policy.ID, "scenario").First(&quota).Error)
	require.InDelta(t, 1000, quota.QuotaLimit, 0.0001)
	var rate models.AIStrategyRateLimitRule
	require.NoError(t, db.Where("policy_id = ? AND dimension = ? AND deleted_at IS NULL", policy.ID, "user").First(&rate).Error)
	require.Equal(t, 20, rate.QPS)

	result, err = service.ImportTenantStrategies(ctx, 7, TenantStrategyImportRequest{
		Policies: []TenantStrategyImportPolicy{{
			PolicyName: "重点租户策略", TenantScope: "include", TenantIDs: []string{"tenant-b"},
			AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
			DefaultBaseRouteID: route.ID,
		}},
		QuotaRules: []TenantStrategyImportQuotaRule{{
			PolicyName: "重点租户策略", Dimension: "scenario", SubjectCode: "copy_gen", UsageUnit: "tokens", Period: "day",
			QuotaLimit: 2000, WarningThreshold: 70, OverLimitAction: "degrade_route",
		}},
		RateLimitRules: []TenantStrategyImportRateLimitRule{{
			PolicyName: "重点租户策略", Dimension: "user", SubjectCode: "user-a", QPS: 10, Concurrency: 3, OverLimitAction: "reject",
		}},
	})
	require.NoError(t, err)
	require.Equal(t, TenantStrategyImportResult{Policies: 1, QuotaRules: 1, RateLimitRules: 1}, result)

	var count int64
	require.NoError(t, db.Model(&models.AITenantStrategyPolicy{}).Where("policy_name = ? AND deleted_at IS NULL", "重点租户策略").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIStrategyQuotaRule{}).Where("policy_id = ? AND deleted_at IS NULL", policy.ID).Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Model(&models.AIStrategyRateLimitRule{}).Where("policy_id = ? AND deleted_at IS NULL", policy.ID).Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Where("id = ?", policy.ID).First(&policy).Error)
	require.Equal(t, []string{"tenant-b"}, policy.TenantIDs)
	require.Empty(t, policy.OverrideBaseRouteID)
	require.NoError(t, db.Where("id = ?", quota.ID).First(&quota).Error)
	require.InDelta(t, 2000, quota.QuotaLimit, 0.0001)
	require.Equal(t, "degrade_route", quota.OverLimitAction)
	require.NoError(t, db.Where("id = ?", rate.ID).First(&rate).Error)
	require.Equal(t, 10, rate.QPS)
	require.Equal(t, "reject", rate.OverLimitAction)
}

func TestImportTenantStrategiesRollsBackOnInvalidRule(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	seedScenario(t, db, route.ID)
	service := NewService(db)
	ctx := context.Background()

	_, err := service.ImportTenantStrategies(ctx, 7, TenantStrategyImportRequest{
		Policies: []TenantStrategyImportPolicy{{
			PolicyName: "错误策略", TenantScope: "all", AppCode: "product_center", AppName: "商品中心",
			AIScenarioCode: "copy_gen", AIScenarioName: "文案生成", DefaultBaseRouteID: route.ID,
			QuotaRules: []TenantStrategyImportQuotaRule{{
				Dimension: "unknown", SubjectCode: "copy_gen", UsageUnit: "tokens", Period: "day", QuotaLimit: 100,
			}},
		}},
	})
	require.ErrorIs(t, err, ErrInvalidInput)

	var count int64
	require.NoError(t, db.Model(&models.AITenantStrategyPolicy{}).Where("policy_name = ?", "错误策略").Count(&count).Error)
	require.Equal(t, int64(0), count)
	require.NoError(t, db.Model(&models.AIStrategyQuotaRule{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestDeleteTenantStrategyCascadesRules(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	seedScenario(t, db, route.ID)
	now := time.Now()
	policy := models.AITenantStrategyPolicy{
		ID: "policy-copy", PolicyName: "租户策略", TenantScope: "all", TenantIDs: []string{}, AppCode: "product_center", AppName: "商品中心",
		AIScenarioCode: "copy_gen", AIScenarioName: "文案生成", DefaultBaseRouteID: route.ID, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&policy).Error)
	quota := models.AIStrategyQuotaRule{
		ID: "quota-copy", PolicyID: policy.ID, Dimension: "scenario", SubjectCode: "copy_gen", UsageUnit: "tokens",
		Period: "day", QuotaLimit: 100, WarningThreshold: 80, OverLimitAction: "alert_only", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&quota).Error)
	rate := models.AIStrategyRateLimitRule{
		ID: "rate-copy", PolicyID: policy.ID, Dimension: "user", SubjectCode: "user-a", QPS: 10, OverLimitAction: "queue", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&rate).Error)

	require.NoError(t, NewService(db).DeleteResource(context.Background(), 7, "tenant-strategies", policy.ID))

	var deletedPolicy models.AITenantStrategyPolicy
	require.NoError(t, db.Where("id = ?", policy.ID).First(&deletedPolicy).Error)
	require.NotNil(t, deletedPolicy.DeletedAt)
	var deletedQuota models.AIStrategyQuotaRule
	require.NoError(t, db.Where("id = ?", quota.ID).First(&deletedQuota).Error)
	require.NotNil(t, deletedQuota.DeletedAt)
	var deletedRate models.AIStrategyRateLimitRule
	require.NoError(t, db.Where("id = ?", rate.ID).First(&deletedRate).Error)
	require.NotNil(t, deletedRate.DeletedAt)
}

func TestDeleteScenarioBlocksUsageAndStrategyReferences(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-chat", "chat_completion")
	now := time.Now()
	scenario := models.AIScenario{
		ID: "scenario-copy", AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: route.ID, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&scenario).Error)
	record := usageRecord("usage-scenario", "tenant-a", "租户 A", "product_center", "copy_gen", "", 1, 1, 1, 100, "success", now)
	require.NoError(t, db.Create(&record).Error)

	service := NewService(db)
	err := service.DeleteResource(context.Background(), 7, "scenarios", scenario.ID)
	require.ErrorIs(t, err, ErrResourceInUse)

	require.NoError(t, db.Where("request_id = ?", "usage-scenario").Delete(&models.AIUsageRecord{}).Error)
	strategy := models.AITenantStrategyPolicy{
		ID: "strategy-copy", PolicyName: "租户覆盖", TenantScope: "all", TenantIDs: []string{}, AppCode: "product_center", AppName: "商品中心",
		AIScenarioCode: "copy_gen", AIScenarioName: "文案生成", DefaultBaseRouteID: route.ID, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&strategy).Error)

	err = service.DeleteResource(context.Background(), 7, "scenarios", scenario.ID)
	require.ErrorIs(t, err, ErrResourceInUse)
}

func TestSystemSettingsValidateCapabilityAndGatewayRuntime(t *testing.T) {
	db := newAICapabilityTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	_, err := service.CreateResource(ctx, 7, "capabilities", map[string]interface{}{
		"capability_code":       "image_generation",
		"capability_name":       "图片生成",
		"scenario_type":         "image",
		"model_type":            "image",
		"default_billing_unit":  "images",
		"supports_tier_pricing": true,
		"status":                "active",
	})
	require.NoError(t, err)

	_, err = service.CreateResource(ctx, 7, "capabilities", map[string]interface{}{
		"capability_code": "broken",
		"capability_name": "Broken",
	})
	require.ErrorIs(t, err, ErrInvalidInput)

	created, err := service.CreateResource(ctx, 7, "settings", map[string]interface{}{
		"setting_key": "gateway_runtime",
		"setting_value": map[string]interface{}{
			"default_timeout_ms": 45000,
			"default_max_retry":  2,
			"usage_log_async":    true,
			"alert_channels":     []string{"ops"},
		},
		"description": "运行参数",
		"status":      "active",
	})
	require.NoError(t, err)
	setting := created.(models.AIGatewaySetting)
	require.JSONEq(t, `{"default_timeout_ms":45000,"default_max_retry":2,"usage_log_async":true,"alert_channels":["ops"]}`, setting.SettingValue)

	_, err = service.UpdateResource(ctx, 7, "settings", setting.ID, map[string]interface{}{
		"setting_value": map[string]interface{}{
			"default_timeout_ms": 0,
			"default_max_retry":  2,
			"usage_log_async":    true,
		},
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestInvokeAppliesTenantStrategyPricingAndUsageRecord(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	provider, account, api := seedProviderAccountAPI(t, db, "openai", "prod")
	model := seedModel(t, db, provider.Code, "gpt-4.1")
	route := seedBaseRoute(t, db, "route-default", "chat_completion")
	overrideRoute := seedBaseRoute(t, db, "route-override", "chat_completion")
	now := time.Now()
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-default", BaseRouteID: route.ID, ModelID: model.ID, Role: "primary", Priority: 1, Weight: 80, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-override", BaseRouteID: overrideRoute.ID, ModelID: model.ID, Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	seedScenario(t, db, route.ID)
	policy := models.AITenantStrategyPolicy{
		ID: "policy-invoke", PolicyName: "重点租户覆盖", TenantScope: "include", TenantIDs: []string{"tenant-a"},
		AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		DefaultBaseRouteID: route.ID, OverrideBaseRouteID: overrideRoute.ID, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&policy).Error)
	quota := models.AIStrategyQuotaRule{
		ID: "quota-invoke", PolicyID: policy.ID, Dimension: "scenario", SubjectCode: "copy_gen", UsageUnit: "tokens",
		Period: "day", QuotaLimit: 100, WarningThreshold: 80, OverLimitAction: "alert_only", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&quota).Error)
	rate := models.AIStrategyRateLimitRule{
		ID: "rate-invoke", PolicyID: policy.ID, Dimension: "user", SubjectCode: "user-a", MinuteLimit: 10, OverLimitAction: "queue", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&rate).Error)
	pricePolicy := models.AIModelPricePolicy{
		ID: "price-policy-invoke", ModelID: model.ID, FeatureKey: "chat_tokens", FeatureName: "对话 Token", ModelType: "text",
		CapabilityCode: "chat_completion", BillingMode: "tiered", BillingUnit: "tokens", PlatformUnit: "tokens",
		BaseCostPrice: 0.01, BaseSalePrice: 0.03, BasePlatformAmount: 1, Currency: "CNY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&pricePolicy).Error)
	tier := models.AIModelPriceTier{
		ID: "price-tier-invoke", PricePolicyID: pricePolicy.ID, TierName: "standard", Mode: "sync",
		CostPrice: 0.02, SalePrice: 0.05, PlatformAmount: 1, Enabled: true, SortOrder: 1,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&tier).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", TenantName: "租户 A", AppCode: "product_center", AppName: "商品中心",
		AIScenarioCode: "copy_gen", UserID: "user-a", RequestID: "invoke-1",
		Params: map[string]interface{}{"usage_amount": 10, "mode": "sync"},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "success", resp.Status)
	require.Equal(t, overrideRoute.ID, resp.BaseRouteID)
	require.Equal(t, model.ID, resp.ModelID)
	require.Equal(t, policy.ID, resp.StrategyID)
	require.Len(t, resp.Controls["quota_rules"], 1)
	require.Len(t, resp.Controls["rate_limit_rules"], 1)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-1").First(&record).Error)
	require.Equal(t, provider.ID, record.ProviderID)
	require.Equal(t, account.ID, record.ProviderAccountID)
	require.Equal(t, api.ID, record.ProviderAPIID)
	require.Equal(t, overrideRoute.ID, record.BaseRouteID)
	require.Equal(t, policy.ID, record.TenantStrategyID)
	require.Equal(t, pricePolicy.ID, record.PricePolicyID)
	require.Equal(t, tier.ID, record.PriceTierID)
	require.InDelta(t, 10, record.UsageAmount, 0.0001)
	require.InDelta(t, 0.2, record.CostAmount, 0.0001)
	require.InDelta(t, 0.5, record.BillingAmount, 0.0001)
	require.Equal(t, "tokens", record.PlatformUnit)
	require.InDelta(t, 10, record.PlatformAmount, 0.0001)

	var updatedQuota models.AIStrategyQuotaRule
	require.NoError(t, db.Where("id = ?", quota.ID).First(&updatedQuota).Error)
	require.InDelta(t, 10, updatedQuota.UsedAmount, 0.0001)
}

func newAICapabilityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE ai_capabilities (
		id TEXT PRIMARY KEY,
		capability_code TEXT NOT NULL,
		capability_name TEXT NOT NULL,
		scenario_type TEXT NOT NULL,
		model_type TEXT NOT NULL,
		default_billing_unit TEXT NOT NULL,
		supports_tier_pricing BOOLEAN NOT NULL DEFAULT false,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_gateway_settings (
		id TEXT PRIMARY KEY,
		setting_key TEXT NOT NULL UNIQUE,
		setting_value TEXT NOT NULL DEFAULT '{}',
		description TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER,
		user_id INTEGER,
		app_code TEXT,
		module TEXT NOT NULL,
		action TEXT NOT NULL,
		summary TEXT NOT NULL,
		detail TEXT,
		ip TEXT,
		user_agent TEXT,
		request_id TEXT,
		result TEXT NOT NULL DEFAULT 'success',
		created_at DATETIME NOT NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_providers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		type TEXT NOT NULL,
		base_url TEXT NOT NULL,
		auth_type TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		priority INTEGER NOT NULL DEFAULT 0,
		region TEXT,
		qps_limit INTEGER NOT NULL DEFAULT 0,
		monthly_budget NUMERIC NOT NULL DEFAULT 0,
		owner TEXT,
		remark TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_provider_accounts (
		id TEXT PRIMARY KEY,
		provider_id TEXT NOT NULL,
		account_name TEXT NOT NULL,
		endpoint TEXT,
		key_alias TEXT NOT NULL,
		encrypted_api_key TEXT,
		encrypted_secret TEXT,
		quota_limit NUMERIC NOT NULL DEFAULT 0,
		used_quota NUMERIC NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_provider_apis (
		id TEXT PRIMARY KEY,
		provider_id TEXT NOT NULL,
		account_id TEXT NOT NULL,
		api_name TEXT NOT NULL,
		api_path TEXT NOT NULL,
		api_type TEXT NOT NULL,
		capabilities TEXT NOT NULL,
		auth_type TEXT NOT NULL,
		qps_limit INTEGER NOT NULL DEFAULT 0,
		timeout_ms INTEGER NOT NULL DEFAULT 30000,
		status TEXT NOT NULL DEFAULT 'active',
		last_called_at DATETIME,
		health_status TEXT NOT NULL DEFAULT 'unknown',
		health_message TEXT,
		health_checked_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_models (
		id TEXT PRIMARY KEY,
		provider_id TEXT NOT NULL,
		model_code TEXT NOT NULL,
		model_name TEXT NOT NULL,
		model_type TEXT NOT NULL,
		capabilities TEXT NOT NULL,
		context_window INTEGER,
		unit TEXT,
		latency_p95 INTEGER NOT NULL DEFAULT 0,
		success_rate NUMERIC NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'active',
		default_for TEXT NOT NULL,
		remark TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_model_price_policies (
		id TEXT PRIMARY KEY,
		model_id TEXT NOT NULL,
		feature_key TEXT NOT NULL,
		feature_name TEXT NOT NULL,
		model_type TEXT NOT NULL,
		capability_code TEXT NOT NULL,
		billing_mode TEXT NOT NULL,
		billing_unit TEXT NOT NULL,
		platform_unit TEXT NOT NULL,
		base_cost_price NUMERIC NOT NULL DEFAULT 0,
		base_sale_price NUMERIC NOT NULL DEFAULT 0,
		base_platform_amount NUMERIC NOT NULL DEFAULT 0,
		currency TEXT NOT NULL DEFAULT 'CNY',
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_model_price_tiers (
		id TEXT PRIMARY KEY,
		price_policy_id TEXT NOT NULL,
		tier_name TEXT NOT NULL,
		mode TEXT,
		resolution TEXT,
		quality TEXT,
		duration_seconds INTEGER,
		aspect_ratio TEXT,
		cost_price NUMERIC NOT NULL DEFAULT 0,
		sale_price NUMERIC NOT NULL DEFAULT 0,
		platform_amount NUMERIC NOT NULL DEFAULT 0,
		enabled BOOLEAN NOT NULL DEFAULT true,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_base_routes (
		id TEXT PRIMARY KEY,
		route_code TEXT NOT NULL,
		route_name TEXT NOT NULL,
		capability_code TEXT NOT NULL,
		model_type TEXT NOT NULL,
		strategy TEXT NOT NULL,
		timeout_ms INTEGER NOT NULL DEFAULT 30000,
		max_retry INTEGER NOT NULL DEFAULT 0,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_base_route_models (
		id TEXT PRIMARY KEY,
		base_route_id TEXT NOT NULL,
		model_id TEXT NOT NULL,
		role TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 1,
		weight INTEGER NOT NULL DEFAULT 100,
		max_retry INTEGER NOT NULL DEFAULT 0,
		timeout_ms INTEGER NOT NULL DEFAULT 30000,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_scenarios (
		id TEXT PRIMARY KEY,
		app_code TEXT NOT NULL,
		app_name TEXT NOT NULL,
		ai_scenario_code TEXT NOT NULL,
		ai_scenario_name TEXT NOT NULL,
		scenario_type TEXT NOT NULL,
		capability_code TEXT NOT NULL,
		model_type TEXT NOT NULL,
		default_base_route_id TEXT NOT NULL,
		owner TEXT,
		description TEXT,
		version TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_tenant_strategy_policies (
		id TEXT PRIMARY KEY,
		policy_name TEXT NOT NULL,
		tenant_scope TEXT NOT NULL,
		tenant_ids TEXT NOT NULL,
		app_code TEXT NOT NULL,
		app_name TEXT NOT NULL,
		ai_scenario_code TEXT NOT NULL,
		ai_scenario_name TEXT NOT NULL,
		default_base_route_id TEXT NOT NULL,
		override_base_route_id TEXT,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_strategy_quota_rules (
		id TEXT PRIMARY KEY,
		policy_id TEXT NOT NULL,
		dimension TEXT NOT NULL,
		subject_code TEXT NOT NULL,
		usage_unit TEXT NOT NULL,
		period TEXT NOT NULL,
		quota_limit NUMERIC NOT NULL DEFAULT 0,
		used_amount NUMERIC NOT NULL DEFAULT 0,
		warning_threshold NUMERIC NOT NULL DEFAULT 80,
		over_limit_action TEXT NOT NULL DEFAULT 'alert_only',
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_strategy_rate_limit_rules (
		id TEXT PRIMARY KEY,
		policy_id TEXT NOT NULL,
		dimension TEXT NOT NULL,
		subject_code TEXT NOT NULL,
		qps INTEGER NOT NULL DEFAULT 0,
		concurrency INTEGER NOT NULL DEFAULT 0,
		minute_limit INTEGER NOT NULL DEFAULT 0,
		hour_limit INTEGER NOT NULL DEFAULT 0,
		day_limit INTEGER NOT NULL DEFAULT 0,
		over_limit_action TEXT NOT NULL DEFAULT 'queue',
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_usage_records (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL UNIQUE,
		tenant_id TEXT NOT NULL,
		tenant_name TEXT,
		app_code TEXT NOT NULL,
		app_name TEXT,
		ai_scenario_code TEXT NOT NULL,
		ai_scenario_name TEXT,
		user_id TEXT,
		user_name TEXT,
		provider_id TEXT,
		provider_account_id TEXT,
		provider_api_id TEXT,
		model_id TEXT,
		base_route_id TEXT,
		tenant_strategy_id TEXT,
		price_policy_id TEXT,
		price_tier_id TEXT,
		usage_amount NUMERIC NOT NULL DEFAULT 0,
		usage_unit TEXT NOT NULL,
		usage_detail TEXT,
		calls INTEGER NOT NULL DEFAULT 1,
		cost_amount NUMERIC NOT NULL DEFAULT 0,
		billing_amount NUMERIC NOT NULL DEFAULT 0,
		platform_unit TEXT,
		platform_amount NUMERIC NOT NULL DEFAULT 0,
		latency_ms INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		error_code TEXT,
		error_message TEXT,
		request_params TEXT NOT NULL DEFAULT '{}',
		prompt_hash TEXT,
		response_hash TEXT,
		called_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	)`).Error)
	return db
}

func seedCapability(t *testing.T, db *gorm.DB, code, unit string) models.AICapability {
	t.Helper()
	now := time.Now()
	row := models.AICapability{
		ID: code, CapabilityCode: code, CapabilityName: code, ScenarioType: "text", ModelType: "text",
		DefaultBillingUnit: unit, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&row).Error)
	return row
}

func seedBaseRoute(t *testing.T, db *gorm.DB, routeCode, capabilityCode string) models.AIBaseRoute {
	t.Helper()
	now := time.Now()
	row := models.AIBaseRoute{
		ID: "route-" + routeCode, RouteCode: routeCode, RouteName: routeCode, CapabilityCode: capabilityCode,
		ModelType: "text", Strategy: "fallback", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&row).Error)
	return row
}

func seedScenario(t *testing.T, db *gorm.DB, routeID string) models.AIScenario {
	t.Helper()
	now := time.Now()
	row := models.AIScenario{
		ID: "scenario-copy", AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: routeID, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&row).Error)
	return row
}

func seedProvider(t *testing.T, db *gorm.DB, providerCode string) models.AIProvider {
	t.Helper()
	now := time.Now()
	provider := models.AIProvider{
		ID: "provider-" + providerCode, Name: "Provider " + providerCode, Code: providerCode, Type: "public_cloud",
		BaseURL: "https://" + providerCode + ".example", AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&provider).Error)
	return provider
}

func seedModel(t *testing.T, db *gorm.DB, providerCode, modelCode string) models.AIModel {
	t.Helper()
	var provider models.AIProvider
	require.NoError(t, db.Where("code = ? AND deleted_at IS NULL", providerCode).First(&provider).Error)
	now := time.Now()
	model := models.AIModel{
		ID: "model-" + providerCode + "-" + modelCode, ProviderID: provider.ID, ModelCode: modelCode, ModelName: modelCode,
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&model).Error)
	return model
}

func seedProviderAccountAPI(t *testing.T, db *gorm.DB, providerCode, accountName string) (models.AIProvider, models.AIProviderAccount, models.AIProviderAPI) {
	t.Helper()
	provider := seedProvider(t, db, providerCode)
	now := time.Now()
	account := models.AIProviderAccount{
		ID: "account-" + providerCode, ProviderID: provider.ID, AccountName: accountName, KeyAlias: "KEY_" + providerCode, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	api := models.AIProviderAPI{
		ID: "api-" + providerCode, ProviderID: provider.ID, AccountID: account.ID, APIName: "chat", APIPath: "/chat", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&api).Error)
	return provider, account, api
}

func usageRecord(id, tenantID, tenantName, appCode, scenarioCode, modelID string, calls int, cost, billing float64, latency int, status string, calledAt time.Time) models.AIUsageRecord {
	return models.AIUsageRecord{
		ID:             id,
		RequestID:      id,
		TenantID:       tenantID,
		TenantName:     tenantName,
		AppCode:        appCode,
		AppName:        appCode,
		AIScenarioCode: scenarioCode,
		AIScenarioName: scenarioCode,
		ModelID:        modelID,
		UsageAmount:    float64(calls),
		UsageUnit:      "tokens",
		UsageDetail:    "test usage",
		Calls:          calls,
		CostAmount:     cost,
		BillingAmount:  billing,
		LatencyMS:      latency,
		Status:         status,
		RequestParams:  "{}",
		CalledAt:       calledAt,
		CreatedAt:      calledAt,
	}
}
