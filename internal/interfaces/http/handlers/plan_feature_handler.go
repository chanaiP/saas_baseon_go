package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Features(c *gin.Context) {
	var rows []models.SaasFeature
	_ = h.db.Order("id asc").Find(&rows).Error
	items := make([]dto.FeatureResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, featureToResponse(row))
	}
	response.OK(c, dto.ListResponse[dto.FeatureResponse]{Items: items, Total: len(items)})
}

func (h *IdentityHandler) CreateFeature(c *gin.Context) {
	var body models.SaasFeature
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Create(&body).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, featureToResponse(body))
}

func (h *IdentityHandler) UpdateFeature(c *gin.Context) {
	var row models.SaasFeature
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "功能不存在")
		return
	}
	var body featureUpdatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body.Updates()).First(&row, row.ID).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, featureToResponse(row))
}

func (h *IdentityHandler) PlanFeatures(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var links []models.SaasPlanFeature
	_ = h.db.Where("plan_id = ? AND enabled = ?", planID, true).Find(&links).Error
	ids := make([]uint64, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.FeatureID)
	}
	response.OK(c, dto.PlanFeatureIDsResponse{PlanID: planID, FeatureIDs: ids})
}

func (h *IdentityHandler) SavePlanFeatures(c *gin.Context) {
	user, _ := h.currentUser(c)
	planID := parseUintParam(c, "id")
	var body struct {
		FeatureIDs []uint64 `json:"feature_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.savePlanFeaturesWithIDs(planID, body.FeatureIDs); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateAllAuthorizationCache()
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "plan", "features", "保存套餐功能", gin.H{"plan_id": planID, "feature_ids": body.FeatureIDs})
	}
	response.OK(c, dto.PlanFeatureIDsResponse{PlanID: planID, FeatureIDs: body.FeatureIDs})
}

func (h *IdentityHandler) SavePlanCapabilities(c *gin.Context) {
	user, _ := h.currentUser(c)
	var body struct {
		FeatureIDs []uint64         `json:"feature_ids"`
		Quotas     []planQuotaInput `json:"quotas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	planID := parseUintParam(c, "id")
	if err := h.quotaService().SaveCapabilities(c.Request.Context(), planID, body.FeatureIDs, toAppPlanQuotaInputs(body.Quotas)); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateAllAuthorizationCache()
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "plan", "capabilities", "保存套餐能力", gin.H{"plan_id": planID, "feature_ids": body.FeatureIDs, "quota_count": len(body.Quotas)})
	}
	quotaItems := make([]dto.PlanCapabilityQuotaInputResponse, 0, len(body.Quotas))
	for _, quota := range body.Quotas {
		quotaItems = append(quotaItems, dto.PlanCapabilityQuotaInputResponse{QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue})
	}
	response.OK(c, dto.PlanCapabilitiesResponse{PlanID: planID, FeatureIDs: body.FeatureIDs, Quotas: quotaItems})
}

func (h *IdentityHandler) SavePlanFeaturesWithIDs(c *gin.Context, featureIDs []uint64) {
	planID := parseUintParam(c, "id")
	if err := h.savePlanFeaturesWithIDs(planID, featureIDs); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateAllAuthorizationCache()
}

func (h *IdentityHandler) savePlanFeaturesWithIDs(planID uint64, featureIDs []uint64) error {
	return h.quotaService().SaveFeatures(context.Background(), planID, featureIDs)
}

func savePlanFeaturesWithIDsTx(tx *gorm.DB, planID uint64, featureIDs []uint64) error {
	ids := uniqueUint64s(featureIDs)
	if err := lockPlanForUpdate(tx, planID); err != nil {
		return err
	}
	if err := tx.Where("plan_id = ?", planID).Delete(&models.SaasPlanFeature{}).Error; err != nil {
		return err
	}
	for _, featureID := range ids {
		var feature models.SaasFeature
		if err := tx.Where("id = ? AND status = ?", featureID, 1).First(&feature).Error; err != nil {
			return fmt.Errorf("功能不存在或已停用")
		}
		if err := tx.Create(&models.SaasPlanFeature{PlanID: planID, FeatureID: featureID, Enabled: true}).Error; err != nil {
			return err
		}
	}
	return nil
}
