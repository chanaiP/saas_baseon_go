package services

import (
	"context"
	"encoding/json"
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
