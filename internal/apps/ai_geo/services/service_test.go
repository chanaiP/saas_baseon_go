package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	quotaapp "saas_baseon_go/internal/application/quota"
	aiccservices "saas_baseon_go/internal/apps/ai_capability_center/services"
	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/apps/ai_geo/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestBrandsAreTenantScoped(t *testing.T) {
	db := newAiGeoTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.AiGeoBrandCard{TenantID: 1, BrandCode: "B1", BrandName: "租户一品牌", Keywords: "[]", Status: "active", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.AiGeoBrandCard{TenantID: 2, BrandCode: "B2", BrandName: "租户二品牌", Keywords: "[]", Status: "active", CreatedAt: now, UpdatedAt: now}).Error)

	result, err := NewService(repositories.NewRepository(db)).Brands(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, dto.PageRequest{Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	require.Equal(t, "B1", result.Items[0].BrandCode)
}

func TestDraftReviewStateMachine(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "新品种草", Body: "正文"})
	require.NoError(t, err)
	require.Equal(t, "draft", draft.AuditStatus)

	submitted, err := service.SubmitDraft(context.Background(), viewer, draft.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", submitted.AuditStatus)

	approved, err := service.ReviewDraft(context.Background(), viewer, draft.ID, true)
	require.NoError(t, err)
	require.Equal(t, "approved", approved.AuditStatus)

	_, err = service.ReviewDraft(context.Background(), viewer, draft.ID, false)
	require.ErrorIs(t, err, ErrInvalidStatus)
}

func TestDraftCannotCrossTenant(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	draft, err := service.CreateDraft(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, dto.DraftPayload{Title: "其他租户", Body: "正文"})
	require.NoError(t, err)

	_, err = service.SubmitDraft(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, draft.ID)

	require.ErrorIs(t, err, ErrNotFound)
}

func TestRecordAuditStoresAiGeoContext(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	err := service.RecordAudit(context.Background(), viewer, dto.RequestMeta{IP: "127.0.0.1", UserAgent: "ai-geo-test", RequestID: "rid-ai-geo"}, "draft_create", "DRAFT-1", "创建母稿", map[string]interface{}{"title": "新品"})

	require.NoError(t, err)
	var log models.AuditLog
	require.NoError(t, db.Where("tenant_id = ? AND user_id = ? AND module = ? AND action = ?", viewer.TenantID, viewer.UserID, "ai_geo", "draft_create").First(&log).Error)
	require.NotNil(t, log.AppCode)
	require.Equal(t, "ai-geo", *log.AppCode)
	require.NotNil(t, log.Detail)
	require.Contains(t, *log.Detail, `"title":"新品"`)
}

func TestCreateBrandRespectsPackageQuota(t *testing.T) {
	db := newAiGeoTestDB(t)
	seedAiGeoQuotaPlan(t, db, 1, quotaBrandCount, "品牌资料卡数量", ptr("NONE"), 1)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	_, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "品牌一"})
	require.NoError(t, err)
	_, err = service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B2", BrandName: "品牌二"})

	var quotaErr *quotaapp.ExceededError
	require.True(t, errors.As(err, &quotaErr), "unexpected error: %v", err)
	require.Equal(t, "品牌资料卡数量", quotaErr.QuotaName)
}

func TestGenerateDraftConsumesMonthlyQuota(t *testing.T) {
	db := newAiGeoTestDB(t)
	seedAiGeoQuotaPlan(t, db, 1, quotaMonthlyDraftGenerations, "月度母稿生成次数", ptr("MONTH"), 1)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	_, err := service.GenerateDraft(context.Background(), viewer, dto.GenerateDraftPayload{Prompt: "新品上市"})
	require.NoError(t, err)
	_, err = service.GenerateDraft(context.Background(), viewer, dto.GenerateDraftPayload{Prompt: "第二篇"})

	var quotaErr *quotaapp.ExceededError
	require.True(t, errors.As(err, &quotaErr), "unexpected error: %v", err)
	var usage models.TenantQuotaUsage
	require.NoError(t, db.Where("tenant_id = ? AND quota_code = ?", 1, quotaMonthlyDraftGenerations).First(&usage).Error)
	require.Equal(t, time.Now().Format("200601"), usage.PeriodKey)
	require.Equal(t, 1, usage.UsedValue)
}

func TestGenerateDraftInvokesAICapabilityCenterScenario(t *testing.T) {
	db := newAiGeoTestDB(t)
	seedAiGeoQuotaPlan(t, db, 1, quotaMonthlyDraftGenerations, "月度母稿生成次数", ptr("MONTH"), 5)
	service := NewService(repositories.NewRepository(db))
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"draft": map[string]interface{}{
				"title":    "春季新品 GEO 种草",
				"summary":  "围绕春季场景强化新品心智。",
				"body":     "这是一篇来自 AI Gateway 的母稿正文。",
				"keywords": []interface{}{"春季", "新品", "GEO"},
			},
		},
	}}
	service.SetDraftGenerator(NewGatewayDraftGenerator(gateway))

	draft, err := service.GenerateDraft(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, dto.GenerateDraftPayload{Prompt: "春季新品上市"})

	require.NoError(t, err)
	require.Equal(t, draftGenerationScenarioCode, gateway.lastRequest.AIScenarioCode)
	require.Equal(t, "ai-geo", gateway.lastRequest.AppCode)
	require.Equal(t, "1", gateway.lastRequest.TenantID)
	require.Equal(t, "10", gateway.lastRequest.UserID)
	require.Equal(t, "春季新品 GEO 种草", draft.Title)
	require.Equal(t, "这是一篇来自 AI Gateway 的母稿正文。", draft.Body)
	require.Contains(t, draft.Keywords, "GEO")
}

func TestGatewayDraftGeneratorParsesProviderTextJSON(t *testing.T) {
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": `{"title":"文本 JSON 标题","summary":"文本摘要","body":"文本正文","keywords":["文本","解析"]}`,
					},
				},
			},
		},
	}}

	result, err := NewGatewayDraftGenerator(gateway).GenerateDraft(context.Background(), DraftGenerationRequest{
		Viewer:  dto.Viewer{TenantID: 1, UserID: 10},
		Payload: dto.GenerateDraftPayload{Prompt: "文本解析"},
	})

	require.NoError(t, err)
	require.Equal(t, "文本 JSON 标题", result.Title)
	require.Equal(t, "文本正文", result.Body)
	require.Equal(t, []string{"文本", "解析"}, result.Keywords)
}

func newAiGeoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(
		&models.AiGeoBrandCard{},
		&models.AiGeoProductCard{},
		&models.AiGeoSKU{},
		&models.AiGeoCompetitor{},
		&models.AiGeoChannelProfile{},
		&models.AiGeoChannelAccount{},
		&models.AiGeoDraft{},
		&models.AiGeoChannelContent{},
		&models.AiGeoPublishPlan{},
		&models.AiGeoImportBatch{},
		&models.AiGeoMaterialAsset{},
		&models.AiGeoHotspot{},
		&models.SaasPlan{},
		&models.SaasQuota{},
		&models.SaasPlanQuota{},
		&models.TenantSubscription{},
		&models.TenantQuotaOverride{},
		&models.TenantQuotaUsage{},
		&models.AuditLog{},
	))
	return db
}

func seedAiGeoQuotaPlan(t *testing.T, db *gorm.DB, tenantID uint64, quotaCode string, quotaName string, periodType *string, limit int) {
	t.Helper()
	now := time.Now()
	plan := models.SaasPlan{PlanCode: fmt.Sprintf("ai-geo-plan-%d-%s", tenantID, quotaCode), PlanName: "AI GEO 测试套餐", PlanType: "STANDARD", BillingCycle: "MONTH", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&plan).Error)
	quota := models.SaasQuota{QuotaCode: quotaCode, QuotaName: quotaName, QuotaType: "COUNT", PeriodType: periodType, Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&quota).Error)
	require.NoError(t, db.Create(&models.SaasPlanQuota{PlanID: plan.ID, QuotaID: quota.ID, QuotaValue: limit, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.TenantSubscription{TenantID: tenantID, PlanID: plan.ID, SubscriptionStatus: "ACTIVE", StartTime: now.Add(-time.Hour), CreatedAt: now, UpdatedAt: now}).Error)
}

func ptr(value string) *string {
	return &value
}

type fakeAIGatewayInvoker struct {
	response    aiccservices.InvokeResponse
	err         error
	lastRequest aiccservices.InvokeRequest
}

func (f *fakeAIGatewayInvoker) Invoke(_ context.Context, req aiccservices.InvokeRequest) (aiccservices.InvokeResponse, error) {
	f.lastRequest = req
	return f.response, f.err
}
