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

	approved, err := service.ReviewDraft(context.Background(), viewer, draft.ID, true, "内容已确认，可以进入渠道生成")
	require.NoError(t, err)
	require.Equal(t, "approved", approved.AuditStatus)
	history, err := service.DraftAuditSuggestions(context.Background(), viewer, draft.ID, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), history.Total)
	require.Contains(t, *history.Items[0].Summary, "内容已确认")

	_, err = service.ReviewDraft(context.Background(), viewer, draft.ID, false, "不应允许二次审核")
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
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	brand, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "Mardi Ladin"})
	require.NoError(t, err)
	product, err := service.CreateProduct(context.Background(), viewer, dto.ProductPayload{BrandID: brand.ID, ProductCode: "P1", ProductName: "通勤连衣裙"})
	require.NoError(t, err)
	_, err = service.CreateSKU(context.Background(), viewer, dto.SKUPayload{ProductID: product.ID, SKUCode: "SKU1", SKUName: "红色 M", Attributes: map[string]interface{}{"color": "red"}, Price: 399})
	require.NoError(t, err)
	hotspot, err := service.CreateHotspot(context.Background(), viewer, dto.HotspotPayload{Platform: "xiaohongshu", Title: "小个子通勤穿搭", HeatScore: 88})
	require.NoError(t, err)
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

	draft, err := service.GenerateDraft(context.Background(), viewer, dto.GenerateDraftPayload{BrandID: &brand.ID, ProductID: &product.ID, HotspotID: &hotspot.ID, Prompt: "春季新品上市"})

	require.NoError(t, err)
	require.Equal(t, draftGenerationScenarioCode, gateway.lastRequest.AIScenarioCode)
	require.Equal(t, "ai-geo", gateway.lastRequest.AppCode)
	require.Equal(t, "1", gateway.lastRequest.TenantID)
	require.Equal(t, "10", gateway.lastRequest.UserID)
	require.Equal(t, "春季新品 GEO 种草", draft.Title)
	require.Equal(t, "这是一篇来自 AI Gateway 的母稿正文。", draft.Body)
	require.Contains(t, draft.Keywords, "GEO")
	require.Contains(t, gatewayPrompt(t, gateway.lastRequest.Input), "红色 M")
	require.Contains(t, gatewayPrompt(t, gateway.lastRequest.Input), "小个子通勤穿搭")
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

func TestGatewayDraftGeneratorStripsMentorTalkFromDraft(t *testing.T) {
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"draft": map[string]interface{}{
				"title":    "好的，明白，我建议先写一个方向",
				"summary":  "我会先帮你把想法拆开，再生成母稿。",
				"body":     "好的，明白。现在可以生成母稿。\n\n###母稿结构\n标题：25岁女生如何选一件有质感的通勤单品\n\n正文（建立可信度）：\n1.先看场景：通勤不是越正式越好，而是要在会议、地铁和下班约会之间保持得体。\n2.再看质感：低饱和色、稳定版型和细节剪裁，能让日常穿搭更耐看。\n结论：如果你想减少选择成本，就优先选择能覆盖多个场景的基础款。",
				"keywords": []interface{}{"通勤", "质感"},
			},
		},
	}}

	result, err := NewGatewayDraftGenerator(gateway).GenerateDraft(context.Background(), DraftGenerationRequest{
		Viewer:  dto.Viewer{TenantID: 1, UserID: 10},
		Payload: dto.GenerateDraftPayload{Prompt: "我想写小个子的文章"},
	})

	require.NoError(t, err)
	require.NotContains(t, result.Title, "好的")
	require.NotContains(t, result.Summary, "我会先")
	require.NotContains(t, result.Body, "好的")
	require.NotContains(t, result.Body, "母稿结构")
	require.Contains(t, result.Body, "先看场景")
	require.Contains(t, result.Body, "结论")
}

func TestGenerateChannelContentInvokesAICapabilityCenterScenario(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	channel, err := service.CreateChannel(context.Background(), viewer, dto.ChannelPayload{ChannelCode: "xiaohongshu", ChannelName: "小红书"})
	require.NoError(t, err)
	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿标题", Body: "母稿正文"})
	require.NoError(t, err)
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"channel_content": map[string]interface{}{
				"title": "小红书渠道标题",
				"body":  "小红书渠道正文",
			},
		},
	}}
	service.SetChannelContentGenerator(NewGatewayChannelContentGenerator(gateway))

	content, err := service.GenerateChannelContent(context.Background(), viewer, draft.ID, dto.ChannelContentPayload{ChannelID: channel.ID})

	require.NoError(t, err)
	require.Equal(t, channelRewriteScenarioCode, gateway.lastRequest.AIScenarioCode)
	require.Equal(t, "ai-geo", gateway.lastRequest.AppCode)
	require.Equal(t, "1", gateway.lastRequest.TenantID)
	require.Equal(t, "10", gateway.lastRequest.UserID)
	require.Equal(t, "小红书渠道标题", content.Title)
	require.Equal(t, "小红书渠道正文", content.Body)
}

func TestChannelContentEditReviewAndTenantScope(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	channel, err := service.CreateChannel(context.Background(), viewer, dto.ChannelPayload{ChannelCode: "xhs", ChannelName: "小红书"})
	require.NoError(t, err)
	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿", Body: "正文"})
	require.NoError(t, err)
	content, err := service.GenerateChannelContent(context.Background(), viewer, draft.ID, dto.ChannelContentPayload{ChannelID: channel.ID})
	require.NoError(t, err)

	updated, err := service.UpdateChannelContent(context.Background(), viewer, content.ID, dto.ChannelContentPayload{Title: "渠道标题更新", Body: "渠道正文更新"})
	require.NoError(t, err)
	require.Equal(t, "渠道标题更新", updated.Title)
	require.Equal(t, "渠道正文更新", updated.Body)
	require.Equal(t, "pending", updated.AuditStatus)

	approved, err := service.ReviewChannelContent(context.Background(), viewer, content.ID, true, "渠道表达已确认")
	require.NoError(t, err)
	require.Equal(t, "approved", approved.AuditStatus)
	history, err := service.ChannelContentAuditSuggestions(context.Background(), viewer, content.ID, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), history.Total)
	require.Equal(t, "manual_channel_content_review", history.Items[0].ScenarioCode)
	require.Contains(t, *history.Items[0].Summary, "渠道表达已确认")

	_, err = service.ReviewChannelContent(context.Background(), viewer, content.ID, false, "不允许二次审核")
	require.ErrorIs(t, err, ErrInvalidStatus)
	_, err = service.UpdateChannelContent(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, content.ID, dto.ChannelContentPayload{Title: "跨租户"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.ReviewChannelContent(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, content.ID, true, "跨租户")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestDraftAuditSuggestionInvokesAIGatewayAndStoresHistory(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿标题", Body: "母稿正文"})
	require.NoError(t, err)
	gateway := &fakeAIGatewayInvoker{response: aiccservices.InvokeResponse{
		Status: "success",
		Data: map[string]interface{}{
			"audit": map[string]interface{}{
				"risk_level": "medium",
				"passed":     true,
				"summary":    "整体可通过，建议弱化绝对化表达。",
				"suggestions": []interface{}{
					map[string]interface{}{"type": "risk", "content": "避免绝对化措辞"},
				},
				"model_code": "audit-test-model",
			},
		},
	}}
	service.SetAuditAdvisor(NewGatewayAuditAdvisor(gateway))

	record, err := service.GenerateDraftAuditSuggestion(context.Background(), viewer, draft.ID)

	require.NoError(t, err)
	require.Equal(t, auditSuggestionScenarioCode, gateway.lastRequest.AIScenarioCode)
	require.Equal(t, "medium", record.RiskLevel)
	require.True(t, record.Passed)
	require.NotNil(t, record.Summary)
	require.Contains(t, *record.Summary, "整体可通过")
	require.Contains(t, record.SuggestionJSON, "避免绝对化措辞")
	history, err := service.DraftAuditSuggestions(context.Background(), viewer, draft.ID, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), history.Total)
}

func TestChannelContentAuditSuggestionCannotCrossTenant(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	channel, err := service.CreateChannel(context.Background(), viewer, dto.ChannelPayload{ChannelCode: "xhs", ChannelName: "小红书"})
	require.NoError(t, err)
	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿", Body: "正文"})
	require.NoError(t, err)
	content, err := service.GenerateChannelContent(context.Background(), viewer, draft.ID, dto.ChannelContentPayload{ChannelID: channel.ID})
	require.NoError(t, err)

	_, err = service.GenerateChannelContentAuditSuggestion(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, content.ID)

	require.ErrorIs(t, err, ErrNotFound)
}

func TestImportMaterialsRecordsRowErrorsAndPartialSuccess(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	batch, err := service.ImportMaterials(context.Background(), viewer, dto.ImportPayload{
		ImportType: "brand",
		Records: []map[string]interface{}{
			{"brand_code": "B1", "brand_name": "品牌一"},
			{"brand_code": "B2"},
			{},
		},
	})

	require.NoError(t, err)
	require.Equal(t, int64(3), batch.RecordCount)
	require.Equal(t, int64(1), batch.SuccessCount)
	require.Equal(t, int64(2), batch.FailedCount)
	require.Equal(t, "partial_success", batch.Status)

	errorsPage, err := service.ImportErrors(context.Background(), viewer, batch.ID, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(2), errorsPage.Total)
	require.Equal(t, 2, errorsPage.Items[0].RowNumber)
	require.Equal(t, "required", errorsPage.Items[0].ErrorCode)
	require.NotNil(t, errorsPage.Items[0].FieldName)
	require.Equal(t, "brand_name", *errorsPage.Items[0].FieldName)
	require.Equal(t, 3, errorsPage.Items[1].RowNumber)
	require.Equal(t, "empty_row", errorsPage.Items[1].ErrorCode)

	brands, err := service.Brands(context.Background(), viewer, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), brands.Total)
	require.Equal(t, "B1", brands.Items[0].BrandCode)

	updateBatch, err := service.ImportMaterials(context.Background(), viewer, dto.ImportPayload{
		ImportType: "brand",
		Records: []map[string]interface{}{
			{"brand_code": "B1", "brand_name": "品牌一更新", "positioning": "高端户外"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), updateBatch.SuccessCount)
	brands, err = service.Brands(context.Background(), viewer, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), brands.Total)
	require.Equal(t, "品牌一更新", brands.Items[0].BrandName)
	require.NotNil(t, brands.Items[0].Positioning)
	require.Equal(t, "高端户外", *brands.Items[0].Positioning)
}

func TestImportMaterialsPersistsProductSKUAndCompetitorRows(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	brand, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "品牌一"})
	require.NoError(t, err)

	productBatch, err := service.ImportMaterials(context.Background(), viewer, dto.ImportPayload{
		ImportType: "product",
		Records: []map[string]interface{}{
			{"brand_code": brand.BrandCode, "product_code": "P1", "product_name": "商品一", "selling_points": "轻量,保暖"},
			{"brand_code": "NOPE", "product_code": "P2", "product_name": "商品二"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "partial_success", productBatch.Status)
	require.Equal(t, int64(1), productBatch.SuccessCount)
	require.Equal(t, int64(1), productBatch.FailedCount)
	products, err := service.Products(context.Background(), viewer, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), products.Total)
	require.Equal(t, "P1", products.Items[0].ProductCode)

	skuBatch, err := service.ImportMaterials(context.Background(), viewer, dto.ImportPayload{
		ImportType: "sku",
		Records: []map[string]interface{}{
			{"product_code": "P1", "sku_code": "SKU1", "sku_name": "黑色 M", "attributes": `{"color":"black"}`, "price": "199.5"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "completed", skuBatch.Status)
	skus, err := service.SKUs(context.Background(), viewer, dto.PageRequest{Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), skus.Total)
	require.Equal(t, float64(199.5), skus.Items[0].Price)
	require.Equal(t, `{"color":"black"}`, skus.Items[0].Attributes)

	competitorBatch, err := service.ImportMaterials(context.Background(), viewer, dto.ImportPayload{
		ImportType: "competitor",
		Records: []map[string]interface{}{
			{"product_code": "P1", "brand_name": "竞品品牌", "product_name": "竞品商品", "difference": "更便宜"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "completed", competitorBatch.Status)
	competitors, err := service.Competitors(context.Background(), viewer, dto.PageRequest{ProductID: products.Items[0].ID, Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), competitors.Total)
	require.Equal(t, "竞品品牌", competitors.Items[0].BrandName)
}

func TestSKUAndCompetitorAreTenantScopedToProduct(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	otherViewer := dto.Viewer{TenantID: 2, UserID: 20}
	brand, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "品牌一"})
	require.NoError(t, err)
	product, err := service.CreateProduct(context.Background(), viewer, dto.ProductPayload{BrandID: brand.ID, ProductCode: "P1", ProductName: "商品一"})
	require.NoError(t, err)

	sku, err := service.CreateSKU(context.Background(), viewer, dto.SKUPayload{ProductID: product.ID, SKUCode: "SKU1", SKUName: "黑色 M", Attributes: map[string]interface{}{"color": "black"}, Price: 199})
	require.NoError(t, err)
	require.Equal(t, viewer.TenantID, sku.TenantID)
	require.Equal(t, `{"color":"black"}`, sku.Attributes)

	competitor, err := service.CreateCompetitor(context.Background(), viewer, dto.CompetitorPayload{ProductID: product.ID, BrandName: "竞品品牌", ProductName: "竞品商品", Point: "价格更低"})
	require.NoError(t, err)
	require.Equal(t, viewer.TenantID, competitor.TenantID)

	skus, err := service.SKUs(context.Background(), viewer, dto.PageRequest{ProductID: product.ID, Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), skus.Total)
	competitors, err := service.Competitors(context.Background(), viewer, dto.PageRequest{ProductID: product.ID, Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), competitors.Total)

	_, err = service.CreateSKU(context.Background(), otherViewer, dto.SKUPayload{ProductID: product.ID, SKUCode: "SKU-X", SKUName: "跨租户"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.CreateCompetitor(context.Background(), otherViewer, dto.CompetitorPayload{ProductID: product.ID, BrandName: "跨租户", ProductName: "竞品"})
	require.ErrorIs(t, err, ErrNotFound)
}

func TestMaterialAssetsAndHotspotsKeepTenantScope(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	otherViewer := dto.Viewer{TenantID: 2, UserID: 20}
	brand, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "品牌一"})
	require.NoError(t, err)
	product, err := service.CreateProduct(context.Background(), viewer, dto.ProductPayload{BrandID: brand.ID, ProductCode: "P1", ProductName: "商品一"})
	require.NoError(t, err)

	asset, err := service.CreateMaterialAsset(context.Background(), viewer, dto.MaterialAssetPayload{BrandID: &brand.ID, ProductID: &product.ID, AssetType: "image", AssetName: "红裙场景图", URL: "https://example.com/red.jpg", Metadata: map[string]interface{}{"scene": "通勤"}})
	require.NoError(t, err)
	require.Equal(t, viewer.TenantID, asset.TenantID)
	require.NotNil(t, asset.BrandID)
	require.NotNil(t, asset.ProductID)
	require.Contains(t, asset.Metadata, "通勤")

	hotspot, err := service.CreateHotspot(context.Background(), viewer, dto.HotspotPayload{Platform: "zhihu", Title: "小个子怎么穿", HeatScore: 91})
	require.NoError(t, err)
	require.Equal(t, "zhihu", hotspot.Platform)

	assets, err := service.MaterialAssets(context.Background(), viewer, dto.PageRequest{ProductID: product.ID, Keyword: "红裙", Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), assets.Total)
	hotspots, err := service.Hotspots(context.Background(), viewer, dto.PageRequest{Keyword: "小个子", Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), hotspots.Total)
	assetDetail, err := service.MaterialAsset(context.Background(), viewer, asset.ID)
	require.NoError(t, err)
	require.Equal(t, "红裙场景图", assetDetail.AssetName)
	hotspotDetail, err := service.Hotspot(context.Background(), viewer, hotspot.ID)
	require.NoError(t, err)
	require.Equal(t, "小个子怎么穿", hotspotDetail.Title)

	updatedAsset, err := service.UpdateMaterialAsset(context.Background(), viewer, asset.ID, dto.MaterialAssetPayload{AssetType: "image", AssetName: "红裙通勤图", URL: "https://example.com/red-2.jpg", Metadata: map[string]interface{}{"scene": "通勤更新"}})
	require.NoError(t, err)
	require.Equal(t, "红裙通勤图", updatedAsset.AssetName)
	require.Contains(t, updatedAsset.Metadata, "通勤更新")
	updatedHotspot, err := service.UpdateHotspot(context.Background(), viewer, hotspot.ID, dto.HotspotPayload{Platform: "zhihu", Title: "小个子通勤怎么穿", HeatScore: 95})
	require.NoError(t, err)
	require.Equal(t, "小个子通勤怎么穿", updatedHotspot.Title)
	require.Equal(t, 95, updatedHotspot.HeatScore)

	_, err = service.CreateMaterialAsset(context.Background(), otherViewer, dto.MaterialAssetPayload{BrandID: &brand.ID, AssetType: "image", AssetName: "跨租户"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.MaterialAsset(context.Background(), otherViewer, asset.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.Hotspot(context.Background(), otherViewer, hotspot.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.UpdateMaterialAsset(context.Background(), otherViewer, asset.ID, dto.MaterialAssetPayload{AssetName: "跨租户更新"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.UpdateHotspot(context.Background(), otherViewer, hotspot.ID, dto.HotspotPayload{Title: "跨租户更新"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.ArchiveMaterialAsset(context.Background(), otherViewer, asset.ID)
	require.ErrorIs(t, err, ErrNotFound)
	archivedAsset, err := service.ArchiveMaterialAsset(context.Background(), viewer, asset.ID)
	require.NoError(t, err)
	require.Equal(t, "archived", archivedAsset.Status)
	require.NotNil(t, archivedAsset.DeletedAt)

	_, err = service.ArchiveHotspot(context.Background(), otherViewer, hotspot.ID)
	require.ErrorIs(t, err, ErrNotFound)
	archivedHotspot, err := service.ArchiveHotspot(context.Background(), viewer, hotspot.ID)
	require.NoError(t, err)
	require.Equal(t, "archived", archivedHotspot.Status)
	require.NotNil(t, archivedHotspot.DeletedAt)
}

func TestMaterialUpdateAndArchiveKeepTenantScope(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	brand, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "品牌一"})
	require.NoError(t, err)
	product, err := service.CreateProduct(context.Background(), viewer, dto.ProductPayload{BrandID: brand.ID, ProductCode: "P1", ProductName: "商品一"})
	require.NoError(t, err)
	sku, err := service.CreateSKU(context.Background(), viewer, dto.SKUPayload{ProductID: product.ID, SKUCode: "SKU1", SKUName: "黑色 M", Price: 199})
	require.NoError(t, err)
	competitor, err := service.CreateCompetitor(context.Background(), viewer, dto.CompetitorPayload{ProductID: product.ID, BrandName: "竞品", ProductName: "竞品商品"})
	require.NoError(t, err)

	updatedProduct, err := service.UpdateProduct(context.Background(), viewer, product.ID, dto.ProductPayload{ProductName: "商品一更新", CategoryName: "女装", SellingPoints: []string{"显瘦"}, FAQ: []string{"怎么洗"}})
	require.NoError(t, err)
	require.Equal(t, "商品一更新", updatedProduct.ProductName)
	updatedSKU, err := service.UpdateSKU(context.Background(), viewer, sku.ID, dto.SKUPayload{SKUName: "黑色 L", Attributes: map[string]interface{}{"size": "L"}, Price: 219, StockStatus: "in_stock"})
	require.NoError(t, err)
	require.Equal(t, "黑色 L", updatedSKU.SKUName)
	updatedCompetitor, err := service.UpdateCompetitor(context.Background(), viewer, competitor.ID, dto.CompetitorPayload{BrandName: "竞品更新", ProductName: "竞品商品更新", Difference: "风格不同"})
	require.NoError(t, err)
	require.Equal(t, "竞品更新", updatedCompetitor.BrandName)

	archivedSKU, err := service.ArchiveSKU(context.Background(), viewer, sku.ID)
	require.NoError(t, err)
	require.Equal(t, "archived", archivedSKU.Status)
	require.NotNil(t, archivedSKU.DeletedAt)
	_, err = service.UpdateSKU(context.Background(), viewer, sku.ID, dto.SKUPayload{SKUName: "不应更新"})
	require.ErrorIs(t, err, ErrNotFound)

	_, err = service.Brand(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, brand.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.Product(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, product.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.SKU(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, sku.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.Competitor(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, competitor.ID)
	require.ErrorIs(t, err, ErrNotFound)

	_, err = service.ArchiveCompetitor(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, competitor.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestKeywordsKeepTenantScopeAndValidateMaterialScope(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	otherViewer := dto.Viewer{TenantID: 2, UserID: 20}
	brand, err := service.CreateBrand(context.Background(), viewer, dto.BrandPayload{BrandCode: "B1", BrandName: "品牌一"})
	require.NoError(t, err)
	product, err := service.CreateProduct(context.Background(), viewer, dto.ProductPayload{BrandID: brand.ID, ProductCode: "P1", ProductName: "商品一"})
	require.NoError(t, err)

	keyword, err := service.CreateKeyword(context.Background(), viewer, dto.KeywordPayload{BrandID: &brand.ID, ProductID: &product.ID, KeywordGroup: "通勤场景", Keyword: "小个子通勤穿搭", Intent: "search", Weight: 80})
	require.NoError(t, err)
	require.Equal(t, viewer.TenantID, keyword.TenantID)
	require.NotNil(t, keyword.BrandID)
	require.Equal(t, brand.ID, *keyword.BrandID)

	page, err := service.Keywords(context.Background(), viewer, dto.PageRequest{BrandID: brand.ID, Keyword: "通勤", Limit: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	detail, err := service.Keyword(context.Background(), viewer, keyword.ID)
	require.NoError(t, err)
	require.Equal(t, "小个子通勤穿搭", detail.Keyword)

	updated, err := service.UpdateKeyword(context.Background(), viewer, keyword.ID, dto.KeywordPayload{BrandID: &brand.ID, ProductID: &product.ID, KeywordGroup: "选购问题", Keyword: "通勤连衣裙怎么选", Weight: 90})
	require.NoError(t, err)
	require.Equal(t, "选购问题", updated.KeywordGroup)
	require.Equal(t, 90, updated.Weight)

	_, err = service.Keyword(context.Background(), otherViewer, keyword.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.UpdateKeyword(context.Background(), otherViewer, keyword.ID, dto.KeywordPayload{Keyword: "跨租户"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = service.CreateKeyword(context.Background(), otherViewer, dto.KeywordPayload{BrandID: &brand.ID, Keyword: "跨租户"})
	require.ErrorIs(t, err, ErrNotFound)

	archived, err := service.ArchiveKeyword(context.Background(), viewer, keyword.ID)
	require.NoError(t, err)
	require.Equal(t, "archived", archived.Status)
	require.NotNil(t, archived.DeletedAt)
	_, err = service.Keyword(context.Background(), viewer, keyword.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestDraftDetailAndArchiveKeepTenantScope(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿", Body: "正文"})
	require.NoError(t, err)
	detail, err := service.Draft(context.Background(), viewer, draft.ID)
	require.NoError(t, err)
	require.Equal(t, draft.DraftCode, detail.DraftCode)

	_, err = service.Draft(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, draft.ID)
	require.ErrorIs(t, err, ErrNotFound)
	archived, err := service.ArchiveDraft(context.Background(), viewer, draft.ID)
	require.NoError(t, err)
	require.Equal(t, "archived", archived.Status)
	require.NotNil(t, archived.DeletedAt)
	_, err = service.Draft(context.Background(), viewer, draft.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPublishPlanSyncsChannelContentStatusAndStateMachine(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	channel, err := service.CreateChannel(context.Background(), viewer, dto.ChannelPayload{ChannelCode: "xiaohongshu", ChannelName: "小红书"})
	require.NoError(t, err)
	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿", Body: "正文"})
	require.NoError(t, err)
	content, err := service.GenerateChannelContent(context.Background(), viewer, draft.ID, dto.ChannelContentPayload{ChannelID: channel.ID})
	require.NoError(t, err)

	plan, err := service.CreatePublishPlan(context.Background(), viewer, dto.PublishPlanPayload{
		ChannelContentID: content.ID,
		ChannelID:        channel.ID,
		ScheduledAt:      time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	var plannedContent models.AiGeoChannelContent
	require.NoError(t, db.First(&plannedContent, content.ID).Error)
	require.Equal(t, "planned", plannedContent.PublishStatus)

	nextSchedule := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	updatedPlan, err := service.UpdatePublishPlan(context.Background(), viewer, plan.ID, dto.PublishPlanPayload{
		ChannelContentID: content.ID,
		ChannelID:        channel.ID,
		ScheduledAt:      nextSchedule.Format(time.RFC3339),
		PublishMethod:    "agent",
		AutomationLevel:  "semi_auto",
	})
	require.NoError(t, err)
	require.Equal(t, nextSchedule.UTC(), updatedPlan.ScheduledAt.UTC())
	require.Equal(t, "agent", updatedPlan.PublishMethod)

	_, err = service.UpdatePublishStatus(context.Background(), viewer, plan.ID, dto.PublishStatusPayload{Status: "failed", FailReason: "发布接口失败"})
	require.NoError(t, err)
	_, err = service.UpdatePublishStatus(context.Background(), viewer, plan.ID, dto.PublishStatusPayload{Status: "scheduled"})
	require.NoError(t, err)
	_, err = service.UpdatePublishStatus(context.Background(), viewer, plan.ID, dto.PublishStatusPayload{Status: "publishing"})
	require.NoError(t, err)
	_, err = service.UpdatePublishStatus(context.Background(), viewer, plan.ID, dto.PublishStatusPayload{Status: "published", PublishedURL: "https://example.com/post/1"})
	require.NoError(t, err)
	require.NoError(t, db.First(&plannedContent, content.ID).Error)
	require.Equal(t, "published", plannedContent.PublishStatus)

	_, err = service.UpdatePublishPlan(context.Background(), viewer, plan.ID, dto.PublishPlanPayload{ScheduledAt: time.Now().Add(3 * time.Hour).Format(time.RFC3339)})
	require.ErrorIs(t, err, ErrInvalidStatus)
	_, err = service.UpdatePublishStatus(context.Background(), viewer, plan.ID, dto.PublishStatusPayload{Status: "cancelled"})
	require.ErrorIs(t, err, ErrInvalidStatus)
}

func TestPublishPlanCalendarAggregatesByTenantDateAndStatus(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}
	channel, err := service.CreateChannel(context.Background(), viewer, dto.ChannelPayload{ChannelCode: "zhihu", ChannelName: "知乎"})
	require.NoError(t, err)
	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "母稿", Body: "正文"})
	require.NoError(t, err)
	content, err := service.GenerateChannelContent(context.Background(), viewer, draft.ID, dto.ChannelContentPayload{ChannelID: channel.ID})
	require.NoError(t, err)
	scheduledAt := time.Date(2026, 5, 20, 9, 30, 0, 0, time.UTC)
	plan, err := service.CreatePublishPlan(context.Background(), viewer, dto.PublishPlanPayload{
		ChannelContentID: content.ID,
		ChannelID:        channel.ID,
		ScheduledAt:      scheduledAt.Format(time.RFC3339),
	})
	require.NoError(t, err)
	_, err = service.UpdatePublishStatus(context.Background(), viewer, plan.ID, dto.PublishStatusPayload{Status: "failed", FailReason: "模拟失败"})
	require.NoError(t, err)
	require.NoError(t, db.Create(&models.AiGeoPublishPlan{TenantID: 2, PlanCode: "OTHER", ChannelContentID: content.ID, ChannelID: channel.ID, ScheduledAt: scheduledAt, Status: "scheduled"}).Error)

	calendar, err := service.PublishPlanCalendar(context.Background(), viewer, dto.PageRequest{StartDate: &scheduledAt, EndDate: &scheduledAt})

	require.NoError(t, err)
	require.Equal(t, "2026-05-20", calendar.StartDate)
	require.Equal(t, "2026-05-20", calendar.EndDate)
	require.Len(t, calendar.Days, 1)
	require.Equal(t, 1, calendar.Days[0].Total)
	require.Equal(t, 1, calendar.Days[0].Failed)
	require.Equal(t, "PLAN", calendar.Days[0].Items[0].PlanCode[:4])
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
		&models.AiGeoKeyword{},
		&models.AiGeoChannelProfile{},
		&models.AiGeoChannelAccount{},
		&models.AiGeoDraft{},
		&models.AiGeoChannelContent{},
		&models.AiGeoPublishPlan{},
		&models.AiGeoImportBatch{},
		&models.AiGeoImportError{},
		&models.AiGeoAuditSuggestion{},
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

func gatewayPrompt(t *testing.T, input map[string]interface{}) string {
	t.Helper()
	messages, ok := input["messages"].([]map[string]string)
	require.True(t, ok)
	require.NotEmpty(t, messages)
	return messages[len(messages)-1]["content"]
}
