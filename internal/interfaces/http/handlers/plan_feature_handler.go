package handlers

import (
	"context"
	"fmt"
	"strings"

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
		if excludedPlanMatrixFeatureRow(row) {
			continue
		}
		items = append(items, featureToResponse(row))
	}
	response.OK(c, dto.ListResponse[dto.FeatureResponse]{Items: items, Total: len(items)})
}

func (h *IdentityHandler) CreateFeature(c *gin.Context) {
	var body featureUpdatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	feature, msg := manualFeatureFromPayload(body)
	if msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	if err := h.db.Create(&feature).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, featureToResponse(feature))
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
	if row.MenuID != nil {
		response.Error(c, 400, response.CodeBadRequest, "菜单管理生成的功能点请在菜单管理中维护")
		return
	}
	if msg := validateManualFeatureUpdate(body); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	if err := h.db.Model(&row).Updates(body.Updates()).First(&row, row.ID).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, featureToResponse(row))
}

func manualFeatureFromPayload(body featureUpdatePayload) (models.SaasFeature, string) {
	if body.FeatureCode == nil || strings.TrimSpace(*body.FeatureCode) == "" {
		return models.SaasFeature{}, "功能编码不能为空"
	}
	if body.FeatureName == nil || strings.TrimSpace(*body.FeatureName) == "" {
		return models.SaasFeature{}, "功能名称不能为空"
	}
	if body.FeatureType == nil || strings.TrimSpace(*body.FeatureType) == "" {
		return models.SaasFeature{}, "功能类型不能为空"
	}
	featureType := strings.ToUpper(strings.TrimSpace(*body.FeatureType))
	if msg := validateManualFeatureType(featureType); msg != "" {
		return models.SaasFeature{}, msg
	}
	feature := models.SaasFeature{
		FeatureCode: strings.TrimSpace(*body.FeatureCode),
		FeatureName: strings.TrimSpace(*body.FeatureName),
		FeatureType: featureType,
		ParentID:    0,
		Status:      1,
		Description: nullableTrimmed(body.Description),
	}
	if body.Status != nil {
		feature.Status = *body.Status
	}
	if featureType == "API" {
		feature.APIMethod = normalizedOptionalUpper(body.APIMethod)
		feature.APIPath = nullableTrimmed(body.APIPath)
		if feature.APIMethod == nil || feature.APIPath == nil {
			return models.SaasFeature{}, "API 功能必须填写 API 方法和 API 路径"
		}
		return feature, ""
	}
	feature.ServiceKey = nullableTrimmed(body.ServiceKey)
	if feature.ServiceKey == nil {
		if featureType == "SERVICE" {
			return models.SaasFeature{}, "服务功能必须填写服务标识"
		}
		return models.SaasFeature{}, "配置功能必须填写配置标识"
	}
	return feature, ""
}

func validateManualFeatureUpdate(body featureUpdatePayload) string {
	if body.FeatureType != nil {
		if msg := validateManualFeatureType(strings.ToUpper(strings.TrimSpace(*body.FeatureType))); msg != "" {
			return msg
		}
	}
	if body.MenuID != nil && *body.MenuID != 0 {
		return "手工功能点不能绑定菜单，菜单能力请在菜单管理中维护"
	}
	return ""
}

func validateManualFeatureType(featureType string) string {
	switch featureType {
	case "API", "SERVICE", "CONFIG":
		return ""
	case "MENU", "BUTTON":
		return "目录、菜单和操作能力由菜单管理自动生成，不能在套餐中心手工新增"
	default:
		return "功能类型只允许 API、服务或配置"
	}
}

func normalizedOptionalUpper(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.ToUpper(strings.TrimSpace(*value))
	if normalized == "" {
		return nil
	}
	return &normalized
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
