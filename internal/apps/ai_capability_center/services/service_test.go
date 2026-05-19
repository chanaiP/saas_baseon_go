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
	require.InDelta(t, 66.67, overview.UsageTrend[6].SuccessRate, 0.01)
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

func TestBillingUnitConversion(t *testing.T) {
	require.InDelta(t, 245, billableUsageAmount(245, "tokens"), 0.0001)
	require.InDelta(t, 0.245, billableUsageAmount(245, "1K tokens"), 0.0001)
	require.InDelta(t, 0.000245, billableUsageAmount(245, "1M tokens"), 0.000001)
	require.InDelta(t, 0.245, platformUsageAmount(245, "1K tokens", 999), 0.0001)
	require.InDelta(t, 0.000245, platformUsageAmount(245, "1M tokens", 999), 0.000001)
	require.InDelta(t, 490, platformUsageAmount(245, "tokens", 2), 0.0001)
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

func TestListUsageRecordsReturnsPagedRowsAndSummary(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now().In(chinaFixedZone())
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, chinaFixedZone())
	records := []models.AIUsageRecord{
		usageRecord("usage-list-1", "tenant-a", "租户 A", "app-a", "chat", "model-a", 3, 1.25, 2.50, 120, "success", todayStart.Add(9*time.Hour)),
		usageRecord("usage-list-2", "tenant-a", "租户 A", "app-a", "chat", "model-a", 2, 0.75, 1.50, 320, "success", todayStart.Add(10*time.Hour)),
		usageRecord("usage-list-3", "tenant-a", "租户 A", "app-a", "image", "model-b", 7, 3.00, 6.00, 220, "success", todayStart.Add(11*time.Hour)),
	}
	require.NoError(t, db.Create(&records).Error)

	result, err := NewService(db).ListUsageRecords(context.Background(), 0, 20, "", "", "", "app-a", "chat")
	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)
	summary, ok := result.Summary.(UsageRecordsSummary)
	require.True(t, ok)
	require.Len(t, summary.UsageUnits, 1)
	require.Equal(t, int64(5), summary.UsageUnits[0].Calls)
	require.InDelta(t, 2.0, summary.UsageUnits[0].CostAmount, 0.0001)
	require.InDelta(t, 4.0, summary.UsageUnits[0].BillingAmount, 0.0001)
	require.Empty(t, summary.ErrorDistribution)
}

func TestListUsageRecordsSummaryIncludesErrorDistribution(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now().In(chinaFixedZone())
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, chinaFixedZone())
	providerError := usageRecord("usage-error-1", "tenant-a", "租户 A", "app-a", "chat", "model-a", 1, 0, 0, 120, "provider_error", todayStart.Add(9*time.Hour))
	providerError.ErrorCode = "provider_http_500"
	timeoutError := usageRecord("usage-error-2", "tenant-a", "租户 A", "app-a", "chat", "model-a", 1, 0, 0, 320, "timeout", todayStart.Add(10*time.Hour))
	timeoutError.ErrorCode = "provider_timeout"
	require.NoError(t, db.Create(&[]models.AIUsageRecord{providerError, timeoutError}).Error)

	result, err := NewService(db).ListUsageRecords(context.Background(), 0, 20, "", "", "", "app-a", "chat")
	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
	summary, ok := result.Summary.(UsageRecordsSummary)
	require.True(t, ok)
	require.Len(t, summary.ErrorDistribution, 2)
	require.Equal(t, UsageErrorDistribution{Status: "provider_error", ErrorCode: "provider_http_500", Count: 1}, summary.ErrorDistribution[0])
	require.Equal(t, UsageErrorDistribution{Status: "timeout", ErrorCode: "provider_timeout", Count: 1}, summary.ErrorDistribution[1])
}

func TestUsageMetricsExcludeDemoSeedRecords(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	modelText := "11111111-1111-1111-1111-111111111111"
	require.NoError(t, db.Create(&models.AIModel{
		ID: modelText, ProviderID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", ModelCode: "gpt-text", ModelName: "GPT Text",
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	realRecord := usageRecord("real-today", "tenant-a", "租户 A", "app-a", "chat", modelText, 2, 1.00, 2.00, 200, "success", todayStart.Add(9*time.Hour))
	seedByID := usageRecord("seed_demo_id", "tenant-b", "租户 B", "app-b", "chat", modelText, 10, 8.00, 16.00, 900, "success", todayStart.Add(10*time.Hour))
	demoByChannel := usageRecord("demo-channel", "tenant-c", "租户 C", "app-c", "chat", modelText, 7, 4.00, 8.00, 700, "success", todayStart.Add(11*time.Hour))
	demoByFlag := usageRecord("demo-flag", "tenant-d", "租户 D", "app-d", "chat", modelText, 5, 2.00, 4.00, 400, "success", todayStart.Add(12*time.Hour))
	demoByChannel.RequestParams = `{"channel":"demo-history-seed"}`
	demoByFlag.DataSource = "demo_seed"
	demoByFlag.IsDemo = true
	require.NoError(t, db.Create(&[]models.AIUsageRecord{realRecord, seedByID, demoByChannel, demoByFlag}).Error)

	overview, err := NewService(db).Overview(context.Background())
	require.NoError(t, err)
	require.Equal(t, "2", overview.Metrics[0].Value)
	require.Equal(t, "¥1.00", overview.Metrics[1].Value)
	require.Equal(t, int64(2), overview.UsageTrend[6].Calls)
	require.Len(t, overview.TenantRanking, 1)
	require.Equal(t, "租户 A", overview.TenantRanking[0].TenantName)

	result, err := NewService(db).ListUsageRecords(context.Background(), 0, 20, "", todayStart.Format("2006-01-02"), todayStart.Format("2006-01-02"), "", "")
	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
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
		Message: "1 个启用 AI 场景未绑定可用基础路由、模型节点或健康 API；影响场景：商品中心/文案生成",
	})
}

func TestEnsureBaselineRepairsScenarioRouteBindingToActiveRoute(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now()
	deletedAt := now.Add(-time.Hour)
	require.NoError(t, db.Create(&models.AIBaseRoute{
		ID: "deleted-route", RouteCode: "deleted-chat", RouteName: "已删除路由",
		CapabilityCode: "chat_completion", ModelType: "text", Strategy: "priority", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), DeletedAt: &deletedAt},
	}).Error)
	require.NoError(t, db.Create(&models.AIBaseRoute{
		ID: "active-route", RouteCode: "active-chat", RouteName: "可用路由",
		CapabilityCode: "chat_completion", ModelType: "text", Strategy: "priority", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIScenario{
		ID: "scenario-needs-repair", AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		ScenarioType: "text", CapabilityCode: "chat_completion", ModelType: "text", DefaultBaseRouteID: "deleted-route", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	require.NoError(t, NewService(db).EnsureBaseline(context.Background()))

	var scenario models.AIScenario
	require.NoError(t, db.Where("id = ?", "scenario-needs-repair").First(&scenario).Error)
	require.Equal(t, "active-route", scenario.DefaultBaseRouteID)
}

func TestCheckProviderAPIConnectivityPersistsReachableStatus(t *testing.T) {
	db := newAICapabilityTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/models", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "Bearer sk-test", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
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
		KeyAlias: "OPENAI_PROD_KEY", EncryptedAPIKey: "sk-test", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAPI{
		ID: "api-chat", ProviderID: "provider-openai", AccountID: "account-prod", APIName: "chat.completions",
		APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
		AuthType: "api_key", QPSLimit: 160, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7, APIConnectivityFilter{})
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Active: 1}, result)

	var api models.AIProviderAPI
	require.NoError(t, db.First(&api, "id = ?", "api-chat").Error)
	require.Equal(t, "active", api.HealthStatus)
	require.Contains(t, api.HealthMessage, "低成本连通性探测通过")
	require.NotNil(t, api.HealthCheckedAt)
}

func TestCheckProviderAPIConnectivityFiltersByAccount(t *testing.T) {
	db := newAICapabilityTestDB(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"probe-ok"}`))
	}))
	defer server.Close()

	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-scope", Name: "Scoped Provider", Code: "scoped", Type: "public_cloud", BaseURL: server.URL,
		AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	for _, accountID := range []string{"account-a", "account-b"} {
		require.NoError(t, db.Create(&models.AIProviderAccount{
			ID: accountID, ProviderID: "provider-scope", AccountName: accountID, Endpoint: server.URL,
			KeyAlias: "SCOPED_KEY", EncryptedAPIKey: "sk-test", Status: "active",
			AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
		}).Error)
		require.NoError(t, db.Create(&models.AIProviderAPI{
			ID: "api-" + accountID, ProviderID: "provider-scope", AccountID: accountID, APIName: "chat." + accountID,
			APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
			AuthType: "api_key", TimeoutMS: 30000, Status: "active", HealthStatus: "unknown",
			AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
		}).Error)
	}

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7, APIConnectivityFilter{AccountID: "account-a"})
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Active: 1}, result)
	require.Equal(t, 1, calls)

	var checked models.AIProviderAPI
	require.NoError(t, db.First(&checked, "id = ?", "api-account-a").Error)
	require.Equal(t, "active", checked.HealthStatus)
	require.NotNil(t, checked.HealthCheckedAt)

	var skipped models.AIProviderAPI
	require.NoError(t, db.First(&skipped, "id = ?", "api-account-b").Error)
	require.Equal(t, "unknown", skipped.HealthStatus)
	require.Nil(t, skipped.HealthCheckedAt)
}

func TestCheckProviderAPIConnectivityFiltersByAPIIDs(t *testing.T) {
	db := newAICapabilityTestDB(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"probe-ok"}`))
	}))
	defer server.Close()

	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-page", Name: "Paged Provider", Code: "paged", Type: "public_cloud", BaseURL: server.URL,
		AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAccount{
		ID: "account-page", ProviderID: "provider-page", AccountName: "prod-main", Endpoint: server.URL,
		KeyAlias: "PAGED_KEY", EncryptedAPIKey: "sk-test", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	for _, apiID := range []string{"api-page-a", "api-page-b"} {
		require.NoError(t, db.Create(&models.AIProviderAPI{
			ID: apiID, ProviderID: "provider-page", AccountID: "account-page", APIName: apiID,
			APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
			AuthType: "api_key", TimeoutMS: 30000, Status: "active", HealthStatus: "unknown",
			AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
		}).Error)
	}

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7, APIConnectivityFilter{APIIDs: []string{"api-page-a"}})
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Active: 1}, result)
	require.Equal(t, 1, calls)

	var checked models.AIProviderAPI
	require.NoError(t, db.First(&checked, "id = ?", "api-page-a").Error)
	require.Equal(t, "active", checked.HealthStatus)
	require.NotNil(t, checked.HealthCheckedAt)

	var skipped models.AIProviderAPI
	require.NoError(t, db.First(&skipped, "id = ?", "api-page-b").Error)
	require.Equal(t, "unknown", skipped.HealthStatus)
	require.Nil(t, skipped.HealthCheckedAt)
}

func TestCheckProviderAPIConnectivityMarksAuthFailureAsError(t *testing.T) {
	db := newAICapabilityTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/v1/models", r.URL.Path)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"invalid key"}`))
	}))
	defer server.Close()

	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-deepseek", Name: "DeepSeek", Code: "deepseek", Type: "public_cloud", BaseURL: server.URL,
		AuthType: "api_key", Status: "active", Priority: 10,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAccount{
		ID: "account-deepseek", ProviderID: "provider-deepseek", AccountName: "prod-main", Endpoint: server.URL,
		KeyAlias: "DEEPSEEK_API_KEY", EncryptedAPIKey: "bad-key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAPI{
		ID: "api-deepseek-chat", ProviderID: "provider-deepseek", AccountID: "account-deepseek", APIName: "chat.completions",
		APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
		AuthType: "api_key", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7, APIConnectivityFilter{})
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Error: 1}, result)

	var api models.AIProviderAPI
	require.NoError(t, db.First(&api, "id = ?", "api-deepseek-chat").Error)
	require.Equal(t, "error", api.HealthStatus)
	require.Contains(t, api.HealthMessage, "鉴权失败")
}

func TestCheckProviderAPIConnectivityClassifiesMissingKeyAlias(t *testing.T) {
	db := newAICapabilityTestDB(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-missing-key", Name: "Missing Key", Code: "openai", Type: "public_cloud", BaseURL: server.URL,
		AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAccount{
		ID: "account-missing-key", ProviderID: "provider-missing-key", AccountName: "prod", Endpoint: server.URL,
		KeyAlias: "MISSING_AI_KEY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAPI{
		ID: "api-missing-key", ProviderID: "provider-missing-key", AccountID: "account-missing-key", APIName: "chat.completions",
		APIPath: "/v1/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
		AuthType: "api_key", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7, APIConnectivityFilter{})
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Error: 1}, result)
	require.Equal(t, 0, calls)

	var api models.AIProviderAPI
	require.NoError(t, db.First(&api, "id = ?", "api-missing-key").Error)
	require.Equal(t, "error", api.HealthStatus)
	require.Contains(t, api.HealthMessage, "未配置密钥")
	require.Contains(t, api.HealthMessage, "MISSING_AI_KEY")
}

func TestCheckProviderAPIConnectivityRejectsPlaceholderEndpoint(t *testing.T) {
	db := newAICapabilityTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-placeholder", Name: "Azure", Code: "azure-openai", Type: "public_cloud", BaseURL: "https://{resource}.openai.azure.com",
		AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAccount{
		ID: "account-placeholder", ProviderID: "provider-placeholder", AccountName: "prod", Endpoint: "https://{resource}.openai.azure.com",
		KeyAlias: "AZURE_OPENAI_API_KEY", EncryptedAPIKey: "sk-test", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIProviderAPI{
		ID: "api-placeholder", ProviderID: "provider-placeholder", AccountID: "account-placeholder", APIName: "chat.completions",
		APIPath: "/openai/deployments/{deployment}/chat/completions", APIType: "chat", Capabilities: []string{"chat_completion"},
		AuthType: "api_key", TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	result, err := NewService(db).CheckProviderAPIConnectivity(context.Background(), 7, APIConnectivityFilter{})
	require.NoError(t, err)
	require.Equal(t, APIConnectivityResult{Total: 1, Error: 1}, result)

	var api models.AIProviderAPI
	require.NoError(t, db.First(&api, "id = ?", "api-placeholder").Error)
	require.Equal(t, "error", api.HealthStatus)
	require.Contains(t, api.HealthMessage, "占位符")
}

func TestProbeHTTPClassifiesTransportAndProviderErrors(t *testing.T) {
	status, message := probeHTTP(context.Background(), http.DefaultClient, "http://127.0.0.1:1/v1/chat/completions", "openai", "api_key", "sk-test", "", "", jsonProbe(map[string]interface{}{"model": "gpt-4o-mini"}))
	require.Equal(t, "error", status)
	require.Contains(t, message, "网络失败")

	status, message = probeHTTP(context.Background(), http.DefaultClient, "http://127.0.0.1:1/v1/chat/completions", "openai", "api_key", "", "", "env:MISSING_AI_KEY", jsonProbe(map[string]interface{}{"model": "gpt-4o-mini"}))
	require.Equal(t, "error", status)
	require.Contains(t, message, "MISSING_AI_KEY")
}

func TestCreateProviderAccountAcceptsSecretPayloadAndMasksAudit(t *testing.T) {
	db := newAICapabilityTestDB(t)
	service := NewService(db)
	now := time.Now()
	require.NoError(t, db.Create(&models.AIProvider{
		ID: "provider-secret", Name: "OpenAI", Code: "openai", Type: "public_cloud", BaseURL: "https://api.openai.com",
		AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	created, err := service.CreateResource(context.Background(), 7, "accounts", map[string]interface{}{
		"provider_id":        "provider-secret",
		"account_name":       "prod-main",
		"endpoint":           "https://api.openai.com",
		"key_alias":          "OPENAI_API_KEY",
		"login_method":       "email",
		"login_account":      "ai-prod@example.com",
		"maintainer":         "平台 AI 组",
		"maintainer_contact": "ai@example.com",
		"encrypted_api_key":  "sk-real",
		"encrypted_secret":   "secret-real",
		"quota_limit":        1000,
		"used_quota":         0,
		"status":             "active",
	})
	require.NoError(t, err)
	account := created.(models.AIProviderAccount)
	require.Equal(t, "sk-real", account.EncryptedAPIKey)

	var persisted models.AIProviderAccount
	require.NoError(t, db.First(&persisted, "id = ?", account.ID).Error)
	require.Equal(t, "sk-real", persisted.EncryptedAPIKey)
	require.Equal(t, "email", persisted.LoginMethod)
	require.Equal(t, "ai-prod@example.com", persisted.LoginAccount)

	var audit models.AuditLog
	require.NoError(t, db.Where("module = ? AND action = ?", "ai_provider_account", "create").First(&audit).Error)
	require.NotNil(t, audit.Detail)
	require.NotContains(t, *audit.Detail, "sk-real")
	require.NotContains(t, *audit.Detail, "secret-real")
	require.Contains(t, *audit.Detail, "prod-main")
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
				LoginMethod: "email", LoginAccount: "openai-prod@example.com", Maintainer: "平台 AI 组", MaintainerContact: "ai@example.com",
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
	require.Equal(t, "email", account.LoginMethod)
	require.Equal(t, "openai-prod@example.com", account.LoginAccount)

	var api models.AIProviderAPI
	require.NoError(t, db.Where("provider_id = ? AND account_id = ? AND api_name = ? AND deleted_at IS NULL", provider.ID, account.ID, "chat.completions").First(&api).Error)
	require.Equal(t, "/v1/chat/completions", api.APIPath)
	require.Equal(t, 45000, api.TimeoutMS)
	require.Equal(t, []string{"chat_completion"}, api.Capabilities)

	require.NoError(t, db.Model(&provider).Updates(map[string]interface{}{"status": "inactive"}).Error)
	require.NoError(t, db.Model(&account).Updates(map[string]interface{}{"key_alias": "OPENAI_MANUAL_KEY", "encrypted_api_key": "manual-ciphertext", "encrypted_secret": "manual-secret", "used_quota": 88.0, "status": "inactive"}).Error)
	require.NoError(t, db.Model(&api).Updates(map[string]interface{}{"status": "inactive", "health_status": "error", "health_message": "manual health failure"}).Error)

	result, err = service.ImportProviders(ctx, 7, ProviderImportRequest{
		Providers: []ProviderImportProvider{{
			Name: "OpenAI Updated", Code: "openai", Type: "public_cloud", BaseURL: "https://gateway.openai.example", Priority: 20,
		}},
		Accounts: []ProviderImportAccount{{
			ProviderCode: "openai", AccountName: "prod", Endpoint: "https://gateway.openai.example", KeyAlias: "OPENAI_PRIMARY", EncryptedAPIKey: "ciphertext-v2",
			LoginMethod: "oauth", LoginAccount: "github:ai-platform", Maintainer: "平台网关组", MaintainerContact: "gateway@example.com",
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
	require.Equal(t, "inactive", provider.Status)
	require.NoError(t, db.Where("id = ?", account.ID).First(&account).Error)
	require.Equal(t, "OPENAI_MANUAL_KEY", account.KeyAlias)
	require.Equal(t, "manual-ciphertext", account.EncryptedAPIKey)
	require.Equal(t, "manual-secret", account.EncryptedSecret)
	require.Equal(t, 88.0, account.UsedQuota)
	require.Equal(t, "inactive", account.Status)
	require.Equal(t, "oauth", account.LoginMethod)
	require.Equal(t, "github:ai-platform", account.LoginAccount)
	require.NoError(t, db.Where("id = ?", api.ID).First(&api).Error)
	require.Equal(t, "/proxy/chat", api.APIPath)
	require.Equal(t, []string{"chat_completion", "embedding"}, api.Capabilities)
	require.Equal(t, "inactive", api.Status)
	require.Equal(t, "error", api.HealthStatus)
	require.Equal(t, "manual health failure", api.HealthMessage)

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
	manualRoute := seedBaseRoute(t, db, "route-manual", "chat_completion")
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
	require.NoError(t, db.Model(&scenario).Updates(map[string]interface{}{"default_base_route_id": manualRoute.ID, "status": "inactive"}).Error)

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
	require.Equal(t, manualRoute.ID, scenario.DefaultBaseRouteID)
	require.Equal(t, "inactive", scenario.Status)
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

	resp, err := newServiceWithProviderExecutor(db, noopProviderExecutor{}).Invoke(context.Background(), InvokeRequest{
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

func TestInvokeExecutesOpenAICompatibleChatAndRecordsProviderUsage(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	now := time.Now()
	seenAuthorization := ""
	seenModel := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuthorization = r.Header.Get("Authorization")
		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		seenModel = payload["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "provider-req-success")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-test","choices":[{"message":{"role":"assistant","content":"生产级响应"},"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":30,"total_tokens":42}}`))
	}))
	defer server.Close()

	provider := models.AIProvider{
		ID: "provider-openai-real", Name: "OpenAI Compatible", Code: "openai", Type: "public_cloud",
		BaseURL: server.URL, AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&provider).Error)
	account := models.AIProviderAccount{
		ID: "account-openai-real", ProviderID: provider.ID, AccountName: "prod", KeyAlias: "OPENAI_TEST_KEY", EncryptedAPIKey: "sk-test",
		Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	api := models.AIProviderAPI{
		ID: "api-openai-real", ProviderID: provider.ID, AccountID: account.ID, APIName: "chat", APIPath: "/v1/chat/completions", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", TimeoutMS: 30000, Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&api).Error)
	model := models.AIModel{
		ID: "model-openai-real", ProviderID: provider.ID, ModelCode: "gpt-test", ModelName: "gpt-test",
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&model).Error)
	route := seedBaseRoute(t, db, "route-openai-real", "chat_completion")
	seedScenario(t, db, route.ID)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-openai-real", BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: account.ID, ProviderAPIID: api.ID,
		Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	pricePolicy := models.AIModelPricePolicy{
		ID: "price-policy-openai-real", ModelID: model.ID, FeatureKey: "chat_tokens", FeatureName: "对话 Token", ModelType: "text",
		CapabilityCode: "chat_completion", BillingMode: "flat", BillingUnit: "1M tokens", PlatformUnit: "1M tokens",
		BaseCostPrice: 2, BaseSalePrice: 12, Currency: "CNY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&pricePolicy).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-real-chat",
		Params: map[string]interface{}{"usage_amount": 1, "max_tokens": 64},
		Input:  map[string]interface{}{"prompt": "写一句欢迎语"},
	})
	require.NoError(t, err)
	require.Equal(t, "success", resp.Status)
	require.Equal(t, "生产级响应", resp.Data["text"])
	require.Equal(t, "Bearer sk-test", seenAuthorization)
	require.Equal(t, "gpt-test", seenModel)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-real-chat").First(&record).Error)
	require.Equal(t, "success", record.Status)
	require.Equal(t, "真实供应商调用成功，已记录路由、策略、配额、限流和价格命中结果", record.UsageDetail)
	require.InDelta(t, 42, record.UsageAmount, 0.0001)
	require.Equal(t, "tokens", record.UsageUnit)
	require.Equal(t, "1M tokens", record.PlatformUnit)
	require.InDelta(t, 0.000042, record.PlatformAmount, 0.000001)
	require.InDelta(t, 0.000084, record.CostAmount, 0.000001)
	require.InDelta(t, 0.000504, record.BillingAmount, 0.000001)
	require.Equal(t, http.StatusOK, record.ProviderHTTPStatus)
	require.Equal(t, "provider-req-success", record.ProviderRequestID)
	require.NotNil(t, record.StartedAt)
	require.NotNil(t, record.FinishedAt)
	require.False(t, record.FinishedAt.Before(*record.StartedAt))
	require.Equal(t, 0, record.RetryCount)
	require.NotEmpty(t, record.ResponseHash)
	require.NotEqual(t, record.PromptHash, record.ResponseHash)
	require.Contains(t, record.RequestParams, `"content_record_level":1`)
}

func TestInvokeDeepSeekReasonerUsesOfficialTokenBreakdownPricing(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "reasoning", "tokens")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-deepseek","choices":[{"message":{"role":"assistant","content":"诊断完成"},"finish_reason":"stop"}],"usage":{"prompt_tokens":300,"prompt_cache_hit_tokens":100,"prompt_cache_miss_tokens":200,"completion_tokens":50,"total_tokens":350}}`))
	}))
	defer server.Close()

	now := time.Now()
	provider := models.AIProvider{
		ID: "provider-deepseek-reasoner", Name: "DeepSeek", Code: "deepseek", Type: "public_cloud",
		BaseURL: server.URL, AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&provider).Error)
	account := models.AIProviderAccount{
		ID: "account-deepseek-reasoner", ProviderID: provider.ID, AccountName: "prod", KeyAlias: "DEEPSEEK_TEST_KEY", EncryptedAPIKey: "sk-test",
		Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	api := models.AIProviderAPI{
		ID: "api-deepseek-reasoner", ProviderID: provider.ID, AccountID: account.ID, APIName: "chat.completions", APIPath: "/v1/chat/completions", APIType: "chat",
		Capabilities: []string{"reasoning"}, AuthType: "api_key", TimeoutMS: 30000, Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&api).Error)
	model := models.AIModel{
		ID: "model-deepseek-reasoner", ProviderID: provider.ID, ModelCode: "deepseek-reasoner", ModelName: "DeepSeek Reasoner",
		ModelType: "text", Capabilities: []string{"reasoning"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&model).Error)
	route := seedBaseRoute(t, db, "route-deepseek-reasoner", "reasoning")
	seedScenario(t, db, route.ID)
	require.NoError(t, db.Model(&models.AIScenario{}).Where("id = ?", "scenario-copy").Updates(map[string]interface{}{
		"scenario_type": "reasoning", "capability_code": "reasoning", "model_type": "text",
	}).Error)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-deepseek-reasoner", BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: account.ID, ProviderAPIID: api.ID,
		Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIModelPricePolicy{
		ID: "price-policy-deepseek-reasoner", ModelID: model.ID, FeatureKey: "reasoning_tokens", FeatureName: "推理 Token", ModelType: "text",
		CapabilityCode: "reasoning", BillingMode: "flat", BillingUnit: "1M tokens", PlatformUnit: "1M tokens",
		BaseCostPrice: 4, BaseSalePrice: 8, Currency: "CNY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-deepseek-reasoner",
		Input: map[string]interface{}{"messages": []map[string]string{
			{"role": "user", "content": "分析异常"},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "success", resp.Status)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-deepseek-reasoner").First(&record).Error)
	require.InDelta(t, 350, record.UsageAmount, 0.0001)
	require.Equal(t, "tokens", record.UsageUnit)
	require.Equal(t, "1M tokens", record.PlatformUnit)
	require.InDelta(t, 0.00035, record.PlatformAmount, 0.000001)
	require.InDelta(t, 0.0017, record.CostAmount, 0.000001)
	require.InDelta(t, 0.0034, record.BillingAmount, 0.000001)
}

func TestInvokeExecutesOpenAICompatibleEmbeddingsAndRecordsProviderUsage(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "embedding", "tokens")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/embeddings", r.URL.Path)
		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "text-embedding-test", payload["model"])
		require.Equal(t, "hello", payload["input"])
		w.Header().Set("X-Request-Id", "provider-req-embedding")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"object":"embedding","index":0,"embedding":[0.1,0.2,0.3]}],"usage":{"prompt_tokens":8,"total_tokens":8}}`))
	}))
	defer server.Close()

	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, server.URL, "embedding-real")
	now := time.Now()
	api.APIPath = "/v1/embeddings"
	api.APIType = "embedding"
	api.Capabilities = []string{"embedding"}
	require.NoError(t, db.Save(&api).Error)
	model.ModelCode = "text-embedding-test"
	model.ModelType = "embedding"
	model.Capabilities = []string{"embedding"}
	require.NoError(t, db.Save(&model).Error)
	require.NoError(t, db.Model(&models.AIBaseRoute{}).Where("id = ?", route.ID).Updates(map[string]interface{}{
		"capability_code": "embedding", "model_type": "embedding",
	}).Error)
	require.NoError(t, db.Model(&models.AIScenario{}).Where("id = ?", "scenario-copy").Updates(map[string]interface{}{
		"scenario_type": "embedding", "capability_code": "embedding", "model_type": "embedding",
	}).Error)
	require.NoError(t, db.Create(&models.AIModelPricePolicy{
		ID: "price-policy-embedding-real", ModelID: model.ID, FeatureKey: "embedding_tokens", FeatureName: "Embedding Token", ModelType: "embedding",
		CapabilityCode: "embedding", BillingMode: "flat", BillingUnit: "tokens", PlatformUnit: "tokens",
		BaseCostPrice: 0.01, BaseSalePrice: 0.02, Currency: "CNY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-real-embedding",
		Params: map[string]interface{}{"usage_amount": 1},
		Input:  map[string]interface{}{"input": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "success", resp.Status)
	require.Equal(t, 1, resp.Data["embedding_count"])
	require.Equal(t, 3, resp.Data["dimensions"])

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-real-embedding").First(&record).Error)
	require.Equal(t, account.ID, record.ProviderAccountID)
	require.InDelta(t, 8, record.UsageAmount, 0.0001)
	require.Equal(t, "tokens", record.UsageUnit)
	require.InDelta(t, 0.08, record.CostAmount, 0.0001)
	require.InDelta(t, 0.16, record.BillingAmount, 0.0001)
	require.Equal(t, "provider-req-embedding", record.ProviderRequestID)
}

func TestInvokeExecutesOpenAICompatibleImagesAndRecordsImageUsage(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "image_generation", "images")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/images/generations", r.URL.Path)
		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "image-test", payload["model"])
		require.Equal(t, "poster", payload["prompt"])
		require.InDelta(t, 2, payload["n"], 0.0001)
		w.Header().Set("X-Request-Id", "provider-req-image")
		_, _ = w.Write([]byte(`{"created":1710000000,"data":[{"url":"https://img.example/a.png"},{"url":"https://img.example/b.png"}]}`))
	}))
	defer server.Close()

	_, _, api, model, route := seedOpenAICompatibleRoute(t, db, server.URL, "image-real")
	now := time.Now()
	api.APIPath = "/v1/images/generations"
	api.APIType = "image"
	api.Capabilities = []string{"image_generation"}
	require.NoError(t, db.Save(&api).Error)
	model.ModelCode = "image-test"
	model.ModelType = "image"
	model.Capabilities = []string{"image_generation"}
	require.NoError(t, db.Save(&model).Error)
	require.NoError(t, db.Model(&models.AIBaseRoute{}).Where("id = ?", route.ID).Updates(map[string]interface{}{
		"capability_code": "image_generation", "model_type": "image",
	}).Error)
	require.NoError(t, db.Model(&models.AIScenario{}).Where("id = ?", "scenario-copy").Updates(map[string]interface{}{
		"scenario_type": "image", "capability_code": "image_generation", "model_type": "image",
	}).Error)
	require.NoError(t, db.Create(&models.AIModelPricePolicy{
		ID: "price-policy-image-real", ModelID: model.ID, FeatureKey: "image_generation", FeatureName: "图片生成", ModelType: "image",
		CapabilityCode: "image_generation", BillingMode: "flat", BillingUnit: "images", PlatformUnit: "images",
		BaseCostPrice: 0.10, BaseSalePrice: 0.20, Currency: "CNY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-real-image",
		Params: map[string]interface{}{"usage_amount": 1, "usage_unit": "images", "n": 2},
		Input:  map[string]interface{}{"prompt": "poster"},
	})
	require.NoError(t, err)
	require.Equal(t, "success", resp.Status)
	require.Equal(t, 2, resp.Data["image_count"])
	require.ElementsMatch(t, []string{"https://img.example/a.png", "https://img.example/b.png"}, resp.Data["urls"])

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-real-image").First(&record).Error)
	require.InDelta(t, 2, record.UsageAmount, 0.0001)
	require.Equal(t, "images", record.UsageUnit)
	require.InDelta(t, 0.20, record.CostAmount, 0.0001)
	require.InDelta(t, 0.40, record.BillingAmount, 0.0001)
	require.Equal(t, "provider-req-image", record.ProviderRequestID)
}

func TestInvokeRecordsProviderErrorWhenOpenAICompatibleCallFails(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Openai-Request-Id", "provider-req-error")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"provider exploded"}}`))
	}))
	defer server.Close()

	provider := models.AIProvider{
		ID: "provider-openai-error", Name: "OpenAI Compatible", Code: "openai", Type: "public_cloud",
		BaseURL: server.URL, AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&provider).Error)
	account := models.AIProviderAccount{
		ID: "account-openai-error", ProviderID: provider.ID, AccountName: "prod", KeyAlias: "OPENAI_TEST_KEY", EncryptedAPIKey: "sk-test",
		Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	api := models.AIProviderAPI{
		ID: "api-openai-error", ProviderID: provider.ID, AccountID: account.ID, APIName: "chat", APIPath: "/v1/chat/completions", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", TimeoutMS: 30000, Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&api).Error)
	model := models.AIModel{
		ID: "model-openai-error", ProviderID: provider.ID, ModelCode: "gpt-test", ModelName: "gpt-test",
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&model).Error)
	route := seedBaseRoute(t, db, "route-openai-error", "chat_completion")
	seedScenario(t, db, route.ID)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-openai-error", BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: account.ID, ProviderAPIID: api.ID,
		Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	policy := models.AITenantStrategyPolicy{
		ID: "policy-openai-error", PolicyName: "全部租户", TenantScope: "all", TenantIDs: []string{},
		AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		DefaultBaseRouteID: route.ID, Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&policy).Error)
	quota := models.AIStrategyQuotaRule{
		ID: "quota-openai-error", PolicyID: policy.ID, Dimension: "scenario", SubjectCode: "copy_gen", UsageUnit: "tokens",
		Period: "day", QuotaLimit: 100, OverLimitAction: "alert_only", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&quota).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-provider-error",
		Params: map[string]interface{}{"usage_amount": 5},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "provider_error", resp.Status)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-provider-error").First(&record).Error)
	require.Equal(t, "provider_error", record.Status)
	require.Equal(t, "provider_http_500", record.ErrorCode)
	require.Equal(t, "provider exploded", record.ErrorMessage)
	require.Equal(t, http.StatusInternalServerError, record.ProviderHTTPStatus)
	require.Equal(t, "provider-req-error", record.ProviderRequestID)
	require.NotNil(t, record.StartedAt)
	require.NotNil(t, record.FinishedAt)
	require.InDelta(t, 5, record.UsageAmount, 0.0001)
	require.Equal(t, 0.0, record.BillingAmount)

	var updatedQuota models.AIStrategyQuotaRule
	require.NoError(t, db.Where("id = ?", quota.ID).First(&updatedQuota).Error)
	require.InDelta(t, 0, updatedQuota.UsedAmount, 0.0001)
}

func TestInvokeRetriesProvider5xxAndRecordsRetryCount(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		if attempts == 1 {
			w.Header().Set("X-Request-Id", "provider-req-first")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":{"message":"temporary upstream error"}}`))
			return
		}
		w.Header().Set("X-Request-Id", "provider-req-second")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"重试后成功"},"finish_reason":"stop"}],"usage":{"total_tokens":9}}`))
	}))
	defer server.Close()

	provider, account, api, model, route := seedOpenAICompatibleRoute(t, db, server.URL, "retry")
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID, "max_retry": 1}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-provider-retry",
		Params: map[string]interface{}{"usage_amount": 1},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "success", resp.Status)
	require.Equal(t, "重试后成功", resp.Data["text"])
	require.Equal(t, 2, attempts)
	require.NotEmpty(t, provider.ID)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-provider-retry").First(&record).Error)
	require.Equal(t, http.StatusOK, record.ProviderHTTPStatus)
	require.Equal(t, "provider-req-second", record.ProviderRequestID)
	require.Equal(t, 1, record.RetryCount)
	require.InDelta(t, 9, record.UsageAmount, 0.0001)
}

func TestInvokeRecordsTimeoutWhenProviderExceedsRouteModelTimeout(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"late"}}],"usage":{"total_tokens":1}}`))
	}))
	defer server.Close()

	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, server.URL, "timeout")
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID, "timeout_ms": 1, "max_retry": 0}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-provider-timeout",
		Params: map[string]interface{}{"usage_amount": 3},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "timeout", resp.Status)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-provider-timeout").First(&record).Error)
	require.Equal(t, "timeout", record.Status)
	require.Equal(t, "provider_timeout", record.ErrorCode)
	require.Equal(t, 0, record.ProviderHTTPStatus)
	require.Equal(t, 0, record.RetryCount)
	require.InDelta(t, 3, record.UsageAmount, 0.0001)
	require.Equal(t, 0.0, record.BillingAmount)
}

func TestInvokeFailsSafelyWhenProviderAPIKeyMissing(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("provider should not be called without an API key")
	}))
	defer server.Close()

	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, server.URL, "missing-key")
	require.NoError(t, db.Model(&models.AIProviderAccount{}).
		Where("id = ?", account.ID).
		Updates(map[string]interface{}{"key_alias": "MISSING_OPENAI_TEST_KEY", "encrypted_api_key": ""}).Error)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-provider-missing-key",
		Params: map[string]interface{}{"usage_amount": 2},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "gateway_error", resp.Status)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-provider-missing-key").First(&record).Error)
	require.Equal(t, "gateway_error", record.Status)
	require.Equal(t, "missing_api_key", record.ErrorCode)
	require.Contains(t, record.ErrorMessage, "未配置密钥")
	require.Contains(t, record.ErrorMessage, "MISSING_OPENAI_TEST_KEY")
	require.Equal(t, 0, record.ProviderHTTPStatus)
	require.Equal(t, 0, record.RetryCount)
	require.InDelta(t, 2, record.UsageAmount, 0.0001)
	require.Equal(t, 0.0, record.BillingAmount)
}

func TestInvokeRejectsRouteWhenAllProviderAPIsUnhealthy(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, "https://openai.invalid", "unhealthy")
	require.NoError(t, db.Model(&models.AIProviderAPI{}).
		Where("id = ?", api.ID).
		Update("health_status", "error").Error)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID}).Error)

	_, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-unhealthy-api",
		Params: map[string]interface{}{"usage_amount": 1},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.ErrorIs(t, err, ErrNotFound)

	var count int64
	require.NoError(t, db.Model(&models.AIUsageRecord{}).Where("request_id = ?", "invoke-unhealthy-api").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestInvokeRejectsPlaceholderProviderEndpoint(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, "https://openai.invalid", "placeholder")
	require.NoError(t, db.Model(&models.AIProviderAccount{}).
		Where("id = ?", account.ID).
		Update("endpoint", "https://{resource}.openai.azure.com").Error)
	require.NoError(t, db.Model(&models.AIProviderAPI{}).
		Where("id = ?", api.ID).
		Updates(map[string]interface{}{"api_path": "/openai/deployments/{deployment}/chat/completions", "health_status": "unknown"}).Error)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID}).Error)

	_, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-placeholder-endpoint",
		Params: map[string]interface{}{"usage_amount": 1},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.ErrorIs(t, err, ErrNotFound)

	var count int64
	require.NoError(t, db.Model(&models.AIUsageRecord{}).Where("request_id = ?", "invoke-placeholder-endpoint").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestInvokeRejectsRouteWithEmptyModelPool(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	route := seedBaseRoute(t, db, "route-empty-pool", "chat_completion")
	seedScenario(t, db, route.ID)

	_, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-empty-model-pool",
		Params: map[string]interface{}{"usage_amount": 1},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.ErrorIs(t, err, ErrNotFound)
}

func TestInvokeRecordsQuotaRejectWithoutCallingProvider(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, "https://openai.invalid", "quota-reject")
	now := time.Now()
	policy := models.AITenantStrategyPolicy{
		ID: "policy-quota-reject", PolicyName: "全部租户", TenantScope: "all", TenantIDs: []string{},
		AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		DefaultBaseRouteID: route.ID, Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&policy).Error)
	quota := models.AIStrategyQuotaRule{
		ID: "quota-reject", PolicyID: policy.ID, Dimension: "scenario", SubjectCode: "copy_gen", UsageUnit: "tokens",
		Period: "day", QuotaLimit: 1, UsedAmount: 1, OverLimitAction: "reject", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&quota).Error)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-quota-reject",
		Params: map[string]interface{}{"usage_amount": 2},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "rejected", resp.Status)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-quota-reject").First(&record).Error)
	require.Equal(t, "quota_exceeded", record.ErrorCode)
	require.Equal(t, 0, record.ProviderHTTPStatus)
	require.Equal(t, 0.0, record.BillingAmount)

	var updatedQuota models.AIStrategyQuotaRule
	require.NoError(t, db.Where("id = ?", quota.ID).First(&updatedQuota).Error)
	require.InDelta(t, 1, updatedQuota.UsedAmount, 0.0001)
}

func TestInvokeRecordsRateLimitRejectWithoutCallingProvider(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	_, account, api, model, route := seedOpenAICompatibleRoute(t, db, "https://openai.invalid", "rate-reject")
	now := time.Now()
	policy := models.AITenantStrategyPolicy{
		ID: "policy-rate-reject", PolicyName: "全部租户", TenantScope: "all", TenantIDs: []string{},
		AppCode: "product_center", AppName: "商品中心", AIScenarioCode: "copy_gen", AIScenarioName: "文案生成",
		DefaultBaseRouteID: route.ID, Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&policy).Error)
	rate := models.AIStrategyRateLimitRule{
		ID: "rate-reject", PolicyID: policy.ID, Dimension: "user", SubjectCode: "user-a", MinuteLimit: 1,
		OverLimitAction: "reject", Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&rate).Error)
	history := usageRecord("rate-history", "tenant-a", "租户 A", "product_center", "copy_gen", model.ID, 1, 0, 0, 0, "success", now)
	history.TenantStrategyID = policy.ID
	require.NoError(t, db.Create(&history).Error)
	require.NoError(t, db.Model(&models.AIBaseRouteModel{}).
		Where("base_route_id = ? AND model_id = ?", route.ID, model.ID).
		Updates(map[string]interface{}{"provider_account_id": account.ID, "provider_api_id": api.ID}).Error)

	resp, err := NewService(db).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", UserID: "user-a", RequestID: "invoke-rate-reject",
		Params: map[string]interface{}{"usage_amount": 1},
		Input:  map[string]interface{}{"prompt": "hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "rejected", resp.Status)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-rate-reject").First(&record).Error)
	require.Equal(t, "rate_limited", record.ErrorCode)
	require.Equal(t, 0, record.ProviderHTTPStatus)
	require.Equal(t, 0.0, record.BillingAmount)
}

func TestInvokeUsesRouteModelBoundProviderAPI(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	provider := seedProvider(t, db, "openai")
	model := seedModel(t, db, provider.Code, "gpt-4.1")
	route := seedBaseRoute(t, db, "route-bound-api", "chat_completion")
	seedScenario(t, db, route.ID)
	now := time.Now()
	accountA := models.AIProviderAccount{
		ID: "account-openai-a", ProviderID: provider.ID, AccountName: "prod-a", KeyAlias: "KEY_A", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	accountB := models.AIProviderAccount{
		ID: "account-openai-b", ProviderID: provider.ID, AccountName: "prod-b", KeyAlias: "KEY_B", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&[]models.AIProviderAccount{accountA, accountB}).Error)
	apiA := models.AIProviderAPI{
		ID: "api-openai-a", ProviderID: provider.ID, AccountID: accountA.ID, APIName: "chat-a", APIPath: "/v1/chat/completions", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", QPSLimit: 10, Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	apiB := models.AIProviderAPI{
		ID: "api-openai-b", ProviderID: provider.ID, AccountID: accountB.ID, APIName: "chat-b", APIPath: "/v1/chat/completions", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", QPSLimit: 100, Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&[]models.AIProviderAPI{apiA, apiB}).Error)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-bound-api", BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: accountB.ID, ProviderAPIID: apiB.ID,
		Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	_, err := newServiceWithProviderExecutor(db, noopProviderExecutor{}).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-bound-api",
		Params: map[string]interface{}{"usage_amount": 1},
	})
	require.NoError(t, err)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-bound-api").First(&record).Error)
	require.Equal(t, accountB.ID, record.ProviderAccountID)
	require.Equal(t, apiB.ID, record.ProviderAPIID)
}

func TestInvokeSkipsRouteModelWhenBoundAPIUnavailable(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	provider := seedProvider(t, db, "openai")
	model := seedModel(t, db, provider.Code, "gpt-4.1")
	route := seedBaseRoute(t, db, "route-api-fallback", "chat_completion")
	seedScenario(t, db, route.ID)
	now := time.Now()
	account := models.AIProviderAccount{
		ID: "account-openai-fallback", ProviderID: provider.ID, AccountName: "prod", KeyAlias: "KEY", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	badAPI := models.AIProviderAPI{
		ID: "api-openai-bad", ProviderID: provider.ID, AccountID: account.ID, APIName: "bad", APIPath: "/bad", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", Status: "active", HealthStatus: "error",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	goodAPI := models.AIProviderAPI{
		ID: "api-openai-good", ProviderID: provider.ID, AccountID: account.ID, APIName: "good", APIPath: "/good", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&[]models.AIProviderAPI{badAPI, goodAPI}).Error)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-bad-api", BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: account.ID, ProviderAPIID: badAPI.ID,
		Role: "primary", Priority: 1, Weight: 200, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-good-api", BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: account.ID, ProviderAPIID: goodAPI.ID,
		Role: "fallback", Priority: 2, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	_, err := newServiceWithProviderExecutor(db, noopProviderExecutor{}).Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-api-fallback",
		Params: map[string]interface{}{"usage_amount": 1},
	})
	require.NoError(t, err)

	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-api-fallback").First(&record).Error)
	require.Equal(t, goodAPI.ID, record.ProviderAPIID)
}

func TestInvokeRespectsContentRecordLevelSystemParam(t *testing.T) {
	db := newAICapabilityTestDB(t)
	seedCapability(t, db, "chat_completion", "tokens")
	provider, _, _ := seedProviderAccountAPI(t, db, "openai", "prod")
	model := seedModel(t, db, provider.Code, "gpt-4.1")
	route := seedBaseRoute(t, db, "route-content-record", "chat_completion")
	seedScenario(t, db, route.ID)
	now := time.Now()
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-content-record", BaseRouteID: route.ID, ModelID: model.ID,
		Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)

	service := newServiceWithProviderExecutor(db, noopProviderExecutor{})
	_, err := service.Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-content-default",
		Params: map[string]interface{}{"usage_amount": 1, "api_key": "sk-secret"},
		Input:  map[string]interface{}{"prompt": "hello world", "email": "user@example.com"},
	})
	require.NoError(t, err)
	var record models.AIUsageRecord
	require.NoError(t, db.Where("request_id = ?", "invoke-content-default").First(&record).Error)
	require.Contains(t, record.RequestParams, `"content_record_level":1`)
	require.Contains(t, record.RequestParams, `"input_summary"`)
	require.NotContains(t, record.RequestParams, "hello world")
	require.NotContains(t, record.RequestParams, "sk-secret")

	require.NoError(t, db.Create(&models.SystemParam{
		TenantID: 1, Key: contentRecordLevelParamKey, Value: "2", Remark: "test", ValueType: "number",
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	_, err = service.Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-content-redacted",
		Params: map[string]interface{}{
			"usage_amount":  1,
			"api_key":       "sk-secret",
			"authorization": "Bearer token-secret-123",
			"cookie":        "sid=session-secret",
		},
		Input: map[string]interface{}{
			"prompt":      "hello world user@example.com 13800138000 11010119900307521X 6222020123456789 sk-live-secret0001",
			"email":       "user@example.com",
			"address":     "北京市朝阳区测试路 1 号",
			"nested_text": map[string]interface{}{"message": "联系 mobile 13900139000"},
		},
	})
	require.NoError(t, err)
	require.NoError(t, db.Where("request_id = ?", "invoke-content-redacted").First(&record).Error)
	require.Contains(t, record.RequestParams, `"content_record_level":2`)
	require.Contains(t, record.RequestParams, `"mode":"redacted"`)
	require.Contains(t, record.RequestParams, `"audit_required":true`)
	require.Contains(t, record.RequestParams, "hello world")
	require.Contains(t, record.RequestParams, "[REDACTED]")
	require.NotContains(t, record.RequestParams, "sk-secret")
	require.NotContains(t, record.RequestParams, "user@example.com")
	require.NotContains(t, record.RequestParams, "13800138000")
	require.NotContains(t, record.RequestParams, "11010119900307521X")
	require.NotContains(t, record.RequestParams, "6222020123456789")
	require.NotContains(t, record.RequestParams, "session-secret")
	require.NotContains(t, record.RequestParams, "北京市朝阳区")

	require.NoError(t, db.Model(&models.SystemParam{}).Where("param_key = ?", contentRecordLevelParamKey).Update("param_value", "3").Error)
	_, err = service.Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-content-full",
		Params: map[string]interface{}{"usage_amount": 1, "api_key": "sk-secret"},
		Input:  map[string]interface{}{"prompt": "hello world", "email": "user@example.com"},
	})
	require.NoError(t, err)
	require.NoError(t, db.Where("request_id = ?", "invoke-content-full").First(&record).Error)
	require.Contains(t, record.RequestParams, `"content_record_level":3`)
	require.Contains(t, record.RequestParams, `"mode":"full"`)
	require.Contains(t, record.RequestParams, `"audit_required":true`)
	require.Contains(t, record.RequestParams, `"retention_days":7`)
	require.Contains(t, record.RequestParams, "完整内容记录可能包含敏感输入")
	require.Contains(t, record.RequestParams, "hello world")
	require.Contains(t, record.RequestParams, "sk-secret")
	require.Contains(t, record.RequestParams, "user@example.com")

	require.NoError(t, db.Model(&models.SystemParam{}).Where("param_key = ?", contentRecordLevelParamKey).Update("param_value", "0").Error)
	_, err = service.Invoke(context.Background(), InvokeRequest{
		TenantID: "tenant-a", AppCode: "product_center", AIScenarioCode: "copy_gen", RequestID: "invoke-content-none",
		Params: map[string]interface{}{"usage_amount": 1, "api_key": "sk-secret"},
		Input:  map[string]interface{}{"prompt": "hello world", "email": "user@example.com"},
	})
	require.NoError(t, err)
	require.NoError(t, db.Where("request_id = ?", "invoke-content-none").First(&record).Error)
	require.Equal(t, "{}", record.RequestParams)
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
	require.NoError(t, db.Exec(`CREATE TABLE sys_param (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER NOT NULL DEFAULT 1,
		param_key TEXT NOT NULL,
		param_value TEXT NOT NULL,
		remark TEXT NOT NULL DEFAULT '',
		value_type TEXT NOT NULL DEFAULT 'string',
		tenant_editable BOOLEAN NOT NULL DEFAULT true,
		is_platform_only BOOLEAN NOT NULL DEFAULT false,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME
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
		login_method TEXT,
		login_account TEXT,
		maintainer TEXT,
		maintainer_contact TEXT,
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
		provider_account_id TEXT,
		provider_api_id TEXT,
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
		task_duration_ms INTEGER NOT NULL DEFAULT 0,
		provider_http_status INTEGER NOT NULL DEFAULT 0,
		provider_request_id TEXT,
		started_at DATETIME,
		finished_at DATETIME,
		retry_count INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		error_code TEXT,
		error_message TEXT,
		request_params TEXT NOT NULL DEFAULT '{}',
		data_source TEXT NOT NULL DEFAULT 'gateway',
		is_demo BOOLEAN NOT NULL DEFAULT 0,
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

func seedOpenAICompatibleRoute(t *testing.T, db *gorm.DB, baseURL, suffix string) (models.AIProvider, models.AIProviderAccount, models.AIProviderAPI, models.AIModel, models.AIBaseRoute) {
	t.Helper()
	now := time.Now()
	provider := models.AIProvider{
		ID: "provider-openai-" + suffix, Name: "OpenAI Compatible", Code: "openai-" + suffix, Type: "public_cloud",
		BaseURL: baseURL, AuthType: "api_key", Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&provider).Error)
	account := models.AIProviderAccount{
		ID: "account-openai-" + suffix, ProviderID: provider.ID, AccountName: "prod", KeyAlias: "OPENAI_TEST_KEY", EncryptedAPIKey: "sk-test",
		Status: "active", AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	api := models.AIProviderAPI{
		ID: "api-openai-" + suffix, ProviderID: provider.ID, AccountID: account.ID, APIName: "chat", APIPath: "/v1/chat/completions", APIType: "chat",
		Capabilities: []string{"chat_completion"}, AuthType: "api_key", TimeoutMS: 30000, Status: "active", HealthStatus: "active", HealthCheckedAt: &now,
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&api).Error)
	model := models.AIModel{
		ID: "model-openai-" + suffix, ProviderID: provider.ID, ModelCode: "gpt-test", ModelName: "gpt-test",
		ModelType: "text", Capabilities: []string{"chat_completion"}, Status: "active", DefaultFor: []string{},
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&model).Error)
	route := seedBaseRoute(t, db, "route-openai-"+suffix, "chat_completion")
	seedScenario(t, db, route.ID)
	require.NoError(t, db.Create(&models.AIBaseRouteModel{
		ID: "route-model-openai-" + suffix, BaseRouteID: route.ID, ModelID: model.ID, ProviderAccountID: account.ID, ProviderAPIID: api.ID,
		Role: "primary", Priority: 1, Weight: 100, TimeoutMS: 30000, Status: "active",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}).Error)
	return provider, account, api, model, route
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
		DataSource:     "gateway",
		IsDemo:         false,
		CalledAt:       calledAt,
		CreatedAt:      calledAt,
	}
}
