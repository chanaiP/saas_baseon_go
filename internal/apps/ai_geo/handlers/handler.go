package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	quotaapp "saas_baseon_go/internal/application/quota"
	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/apps/ai_geo/services"
	"saas_baseon_go/internal/interfaces/http/response"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Overview(c *gin.Context) {
	data, err := h.service.Overview(c.Request.Context(), viewer(c))
	h.ok(c, data, err)
}

func (h *Handler) Brands(c *gin.Context) {
	data, err := h.service.Brands(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) Brand(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Brand(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateBrand(c *gin.Context) {
	var payload dto.BrandPayload
	if bind(c, &payload) {
		data, err := h.service.CreateBrand(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "brand_create", data.BrandCode, "创建品牌资料卡", gin.H{"brand_name": data.BrandName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateBrand(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.BrandPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateBrand(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "brand_update", data.BrandCode, "更新品牌资料卡", gin.H{"brand_id": data.ID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveBrand(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveBrand(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "brand_archive", data.BrandCode, "归档品牌资料卡", gin.H{"brand_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) Products(c *gin.Context) {
	data, err := h.service.Products(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) Product(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Product(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var payload dto.ProductPayload
	if bind(c, &payload) {
		data, err := h.service.CreateProduct(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "product_create", data.ProductCode, "创建商品资料卡", gin.H{"product_name": data.ProductName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.ProductPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateProduct(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "product_update", data.ProductCode, "更新商品资料卡", gin.H{"product_id": data.ID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveProduct(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "product_archive", data.ProductCode, "归档商品资料卡", gin.H{"product_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) SKUs(c *gin.Context) {
	data, err := h.service.SKUs(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) SKU(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.SKU(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateSKU(c *gin.Context) {
	var payload dto.SKUPayload
	if bind(c, &payload) {
		data, err := h.service.CreateSKU(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "sku_create", data.SKUCode, "创建 SKU 资料", gin.H{"sku_name": data.SKUName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateSKU(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.SKUPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateSKU(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "sku_update", data.SKUCode, "更新 SKU 资料", gin.H{"sku_id": data.ID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveSKU(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveSKU(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "sku_archive", data.SKUCode, "归档 SKU 资料", gin.H{"sku_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) Competitors(c *gin.Context) {
	data, err := h.service.Competitors(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) Competitor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Competitor(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateCompetitor(c *gin.Context) {
	var payload dto.CompetitorPayload
	if bind(c, &payload) {
		data, err := h.service.CreateCompetitor(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "competitor_create", fmt.Sprintf("%d", data.ID), "创建竞品资料", gin.H{"brand_name": data.BrandName, "product_name": data.ProductName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateCompetitor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.CompetitorPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateCompetitor(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "competitor_update", fmt.Sprintf("%d", data.ID), "更新竞品资料", gin.H{"competitor_id": data.ID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveCompetitor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveCompetitor(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "competitor_archive", fmt.Sprintf("%d", data.ID), "归档竞品资料", gin.H{"competitor_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) Keywords(c *gin.Context) {
	data, err := h.service.Keywords(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) Keyword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Keyword(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateKeyword(c *gin.Context) {
	var payload dto.KeywordPayload
	if bind(c, &payload) {
		data, err := h.service.CreateKeyword(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "keyword_create", fmt.Sprintf("%d", data.ID), "创建关键词", gin.H{"keyword": data.Keyword, "keyword_group": data.KeywordGroup})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateKeyword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.KeywordPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateKeyword(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "keyword_update", fmt.Sprintf("%d", data.ID), "更新关键词", gin.H{"keyword": data.Keyword, "keyword_group": data.KeywordGroup})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveKeyword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveKeyword(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "keyword_archive", fmt.Sprintf("%d", data.ID), "归档关键词", gin.H{"keyword_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) MaterialAssets(c *gin.Context) {
	data, err := h.service.MaterialAssets(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) MaterialAsset(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.MaterialAsset(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateMaterialAsset(c *gin.Context) {
	var payload dto.MaterialAssetPayload
	if bind(c, &payload) {
		data, err := h.service.CreateMaterialAsset(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "material_asset_create", fmt.Sprintf("%d", data.ID), "创建素材资料", gin.H{"asset_type": data.AssetType, "asset_name": data.AssetName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateMaterialAsset(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.MaterialAssetPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateMaterialAsset(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "material_asset_update", fmt.Sprintf("%d", data.ID), "更新素材资料", gin.H{"asset_id": data.ID, "asset_name": data.AssetName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveMaterialAsset(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveMaterialAsset(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "material_asset_archive", fmt.Sprintf("%d", data.ID), "归档素材资料", gin.H{"asset_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) Hotspots(c *gin.Context) {
	data, err := h.service.Hotspots(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) Hotspot(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Hotspot(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateHotspot(c *gin.Context) {
	var payload dto.HotspotPayload
	if bind(c, &payload) {
		data, err := h.service.CreateHotspot(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "hotspot_create", fmt.Sprintf("%d", data.ID), "创建热点资料", gin.H{"platform": data.Platform, "title": data.Title})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateHotspot(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.HotspotPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateHotspot(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "hotspot_update", fmt.Sprintf("%d", data.ID), "更新热点资料", gin.H{"hotspot_id": data.ID, "title": data.Title})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveHotspot(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveHotspot(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "hotspot_archive", fmt.Sprintf("%d", data.ID), "归档热点资料", gin.H{"hotspot_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) ExtractExternalSource(c *gin.Context) {
	var payload dto.ExternalSourceExtractPayload
	if bind(c, &payload) {
		data, err := h.service.ExtractExternalSource(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "external_source_extract", fmt.Sprintf("%d", data.Source.ID), "提炼外部来源", gin.H{"source_url": data.Source.SourceURL, "extract_type": payload.ExtractType})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) StyleTemplates(c *gin.Context) {
	data, err := h.service.StyleTemplates(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) StyleTemplate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.StyleTemplate(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateStyleTemplate(c *gin.Context) {
	var payload dto.StyleTemplatePayload
	if bind(c, &payload) {
		data, err := h.service.CreateStyleTemplate(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "style_template_create", fmt.Sprintf("%d", data.ID), "创建写作风格模板", gin.H{"template_name": data.TemplateName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateStyleTemplate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.StyleTemplatePayload
	if bind(c, &payload) {
		data, err := h.service.UpdateStyleTemplate(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "style_template_update", fmt.Sprintf("%d", data.ID), "更新写作风格模板", gin.H{"template_name": data.TemplateName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveStyleTemplate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveStyleTemplate(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "style_template_archive", fmt.Sprintf("%d", data.ID), "归档写作风格模板", gin.H{"template_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) Channels(c *gin.Context) {
	data, err := h.service.Channels(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) CreateChannel(c *gin.Context) {
	var payload dto.ChannelPayload
	if bind(c, &payload) {
		data, err := h.service.CreateChannel(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "channel_create", data.ChannelCode, "创建渠道资料", gin.H{"channel_name": data.ChannelName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateChannel(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.ChannelPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateChannel(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "channel_update", data.ChannelCode, "更新渠道资料", gin.H{"channel_name": data.ChannelName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) TestChannel(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.TestChannel(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "channel_test", data.ChannelCode, "测试渠道连通性", gin.H{"status": data.Status, "message": data.Message})
	}
	h.ok(c, data, err)
}

func (h *Handler) ChannelAccounts(c *gin.Context) {
	data, err := h.service.ChannelAccounts(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) CreateChannelAccount(c *gin.Context) {
	var payload dto.ChannelAccountPayload
	if bind(c, &payload) {
		data, err := h.service.CreateChannelAccount(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "channel_account_create", fmt.Sprintf("%d", data.ID), "创建渠道账号", gin.H{"account_name": data.AccountName})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) Drafts(c *gin.Context) {
	data, err := h.service.Drafts(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) Draft(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.Draft(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) CreateDraft(c *gin.Context) {
	var payload dto.DraftPayload
	if bind(c, &payload) {
		data, err := h.service.CreateDraft(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "draft_create", data.DraftCode, "创建母稿", gin.H{"title": data.Title})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdateDraft(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.DraftPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateDraft(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "draft_update", data.DraftCode, "更新母稿草稿", gin.H{"draft_id": data.ID, "title": data.Title})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ArchiveDraft(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ArchiveDraft(c.Request.Context(), viewer(c), id)
	if err == nil {
		err = h.auditWrite(c, "draft_archive", data.DraftCode, "归档母稿", gin.H{"draft_id": data.ID})
	}
	h.ok(c, data, err)
}

func (h *Handler) GenerateDraft(c *gin.Context) {
	var payload dto.GenerateDraftPayload
	if bind(c, &payload) {
		data, err := h.service.GenerateDraft(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "draft_generate", data.DraftCode, "生成母稿", gin.H{"title": data.Title})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) StreamWorkbenchAI(c *gin.Context) {
	h.streamGatewayScenario(c, "ai_geo_draft_generation")
}

func (h *Handler) StreamChannelContentEditorAI(c *gin.Context) {
	h.streamGatewayScenario(c, "ai_geo_channel_content_editor")
}

func (h *Handler) GenerateChannelContent(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.ChannelContentPayload
	if bind(c, &payload) {
		data, err := h.service.GenerateChannelContent(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "channel_content_generate", fmt.Sprintf("%d", data.ID), "生成渠道内容", gin.H{"draft_id": id, "channel_id": data.ChannelID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ChannelContents(c *gin.Context) {
	data, err := h.service.ChannelContents(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) ChannelContent(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ChannelContent(c.Request.Context(), viewer(c), id)
	h.ok(c, data, err)
}

func (h *Handler) UpdateChannelContent(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.ChannelContentPayload
	if bind(c, &payload) {
		data, err := h.service.UpdateChannelContent(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "channel_content_update", fmt.Sprintf("%d", data.ID), "更新渠道内容", gin.H{"channel_content_id": data.ID, "channel_id": data.ChannelID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) PublishPlans(c *gin.Context) {
	data, err := h.service.PublishPlans(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) PublishPlanCalendar(c *gin.Context) {
	data, err := h.service.PublishPlanCalendar(c.Request.Context(), viewer(c), pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) CreatePublishPlan(c *gin.Context) {
	var payload dto.PublishPlanPayload
	if bind(c, &payload) {
		data, err := h.service.CreatePublishPlan(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "publish_plan_create", data.PlanCode, "创建发布计划", gin.H{"scheduled_at": data.ScheduledAt})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdatePublishPlan(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.PublishPlanPayload
	if bind(c, &payload) {
		data, err := h.service.UpdatePublishPlan(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "publish_plan_update", data.PlanCode, "更新发布计划", gin.H{"scheduled_at": data.ScheduledAt, "channel_id": data.ChannelID})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) UpdatePublishStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var payload dto.PublishStatusPayload
	if bind(c, &payload) {
		data, err := h.service.UpdatePublishStatus(c.Request.Context(), viewer(c), id, payload)
		if err == nil {
			err = h.auditWrite(c, "publish_plan_status", data.PlanCode, "更新发布计划状态", gin.H{"status": data.Status})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ImportMaterials(c *gin.Context) {
	var payload dto.ImportPayload
	if bind(c, &payload) {
		data, err := h.service.ImportMaterials(c.Request.Context(), viewer(c), payload)
		if err == nil {
			err = h.auditWrite(c, "materials_import", data.BatchCode, "导入 AI GEO 资料", gin.H{"import_type": data.ImportType, "record_count": data.RecordCount})
		}
		h.ok(c, data, err)
	}
}

func (h *Handler) ImportErrors(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.service.ImportErrors(c.Request.Context(), viewer(c), id, pageRequest(c))
	h.ok(c, data, err)
}

func (h *Handler) ok(c *gin.Context, values ...interface{}) {
	var data interface{}
	var err error
	if len(values) > 0 {
		data = values[0]
	}
	if len(values) > 1 && values[1] != nil {
		if typed, ok := values[1].(error); ok {
			err = typed
		}
	}
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, dto.PublicData(data))
}

func (h *Handler) streamGatewayScenario(c *gin.Context, scenarioCode string) {
	var req services.GatewayInvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return
	}
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher, _ := c.Writer.(http.Flusher)
	emit := func(event services.GatewayStreamEvent) error {
		payload, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
			return err
		}
		if flusher != nil {
			flusher.Flush()
		}
		return nil
	}
	if err := h.service.StreamGatewayScenario(c.Request.Context(), viewer(c), req, scenarioCode, emit); err != nil {
		_ = emit(services.GatewayStreamEvent{Type: "error", ErrorCode: "ai_geo_stream_failed", ErrorMessage: err.Error()})
	}
}

func (h *Handler) auditWrite(c *gin.Context, action string, objectCode string, summary string, detail interface{}) error {
	return h.service.RecordAudit(c.Request.Context(), viewer(c), requestMeta(c), action, objectCode, summary, detail)
}

func requestMeta(c *gin.Context) dto.RequestMeta {
	return dto.RequestMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), RequestID: c.GetString("request_id")}
}

func writeError(c *gin.Context, err error) {
	var quotaErr *quotaapp.ExceededError
	switch {
	case errors.Is(err, services.ErrNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "数据不存在")
	case errors.Is(err, services.ErrInvalidStatus):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "状态流转不合法")
	case errors.As(err, &quotaErr):
		response.Error(c, http.StatusForbidden, response.CodeForbidden, quotaErr.Error())
	default:
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
	}
}

func bind(c *gin.Context, payload interface{}) bool {
	if err := c.ShouldBindJSON(payload); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return false
	}
	return true
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数错误")
		return 0, false
	}
	return id, true
}

func viewer(c *gin.Context) dto.Viewer {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")
	isPlatformAdmin, _ := c.Get("is_platform_admin")
	viewer := dto.Viewer{UserID: toUint64(userID), TenantID: toUint64(tenantID), IsPlatformAdmin: toBool(isPlatformAdmin)}
	if viewer.IsPlatformAdmin && c.Query("tenant_id") != "" {
		if explicitTenantID, err := strconv.ParseUint(c.Query("tenant_id"), 10, 64); err == nil && explicitTenantID > 0 {
			viewer.TenantID = explicitTenantID
		}
	}
	return viewer
}

func pageRequest(c *gin.Context) dto.PageRequest {
	req := dto.PageRequest{
		Skip:        queryInt(c, "skip", 0),
		Limit:       queryInt(c, "limit", 20),
		Keyword:     c.Query("keyword"),
		Status:      c.Query("status"),
		BrandID:     queryUint(c, "brand_id"),
		ProductID:   queryUint(c, "product_id"),
		DraftID:     queryUint(c, "draft_id"),
		ChannelID:   queryUint(c, "channel_id"),
		AuditStatus: c.Query("audit_status"),
	}
	if start := parseDate(c.Query("start_date")); start != nil {
		req.StartDate = start
	}
	if end := parseDate(c.Query("end_date")); end != nil {
		req.EndDate = end
	}
	return req
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func queryUint(c *gin.Context, key string) uint64 {
	value := c.Query(key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &parsed
}

func toUint64(value interface{}) uint64 {
	switch v := value.(type) {
	case uint64:
		return v
	case uint:
		return uint64(v)
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	default:
		return 0
	}
}

func toBool(value interface{}) bool {
	v, ok := value.(bool)
	return ok && v
}
