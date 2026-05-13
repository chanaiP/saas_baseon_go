package services

import (
	"context"
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
