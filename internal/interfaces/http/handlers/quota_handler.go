package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	appquota "saas_baseon_go/internal/application/quota"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Quotas(c *gin.Context) {
	skip, limit := paginationParams(c)
	query := h.db.Where("quota_code NOT IN ?", hiddenPackageQuotaCodes()).Order("id asc")
	var total int64
	_ = query.Model(&models.SaasQuota{}).Count(&total).Error
	var rows []models.SaasQuota
	_ = query.Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]dto.QuotaResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, quotaToResponse(row))
	}
	response.OK(c, dto.PaginatedResponse[dto.QuotaResponse]{Items: items, Total: total, Skip: skip, Limit: limit})
}

func (h *IdentityHandler) CreateQuota(c *gin.Context) {
	var row models.SaasQuota
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if row.Status == 0 {
		row.Status = 1
	}
	row.QuotaCode = strings.TrimSpace(row.QuotaCode)
	row.QuotaName = strings.TrimSpace(row.QuotaName)
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, quotaToResponse(row))
}

func (h *IdentityHandler) UpdateQuota(c *gin.Context) {
	var row models.SaasQuota
	if err := h.db.First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "配额不存在")
		return
	}
	var body quotaUpdatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body.Updates()).First(&row, row.ID).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, quotaToResponse(row))
}

func (h *IdentityHandler) PlanQuotas(c *gin.Context) {
	planID := parseUintParam(c, "id")
	var rows []struct {
		QuotaID    uint64
		QuotaCode  string
		QuotaName  string
		QuotaValue int
		PeriodType *string
		Unit       *string
	}
	_ = h.db.Table("saas_plan_quota pq").
		Select("pq.quota_id, q.quota_code, q.quota_name, pq.quota_value, q.period_type, q.unit").
		Joins("join saas_quota q on q.id = pq.quota_id").
		Where("pq.plan_id = ? AND q.quota_code NOT IN ?", planID, hiddenPackageQuotaCodes()).
		Scan(&rows).Error
	items := make([]dto.PlanQuotaItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.PlanQuotaItemResponse{QuotaID: row.QuotaID, QuotaCode: row.QuotaCode, QuotaName: row.QuotaName, QuotaValue: row.QuotaValue, PeriodType: row.PeriodType, Unit: row.Unit})
	}
	response.OK(c, dto.PlanQuotasResponse{PlanID: planID, Quotas: items})
}

func hiddenPackageQuotaCodes() []string {
	return []string{"max_api_keys", "max_webhooks", "daily_api_calls", "monthly_sms_count", "monthly_email_count"}
}

func (h *IdentityHandler) SavePlanQuotas(c *gin.Context) {
	user, _ := h.currentUser(c)
	planID := parseUintParam(c, "id")
	var body struct {
		Quotas []planQuotaInput `json:"quotas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.savePlanQuotasWithValues(planID, body.Quotas); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateAllAuthorizationCache()
	if user.ID != 0 {
		h.audit(c, user.TenantID, user.ID, "plan", "quotas", "保存套餐配额", gin.H{"plan_id": planID, "quota_count": len(body.Quotas)})
	}
	h.PlanQuotas(c)
}

func (h *IdentityHandler) savePlanQuotasWithValues(planID uint64, quotas []planQuotaInput) error {
	return h.quotaService().SaveQuotas(context.Background(), planID, toAppPlanQuotaInputs(quotas))
}

func (h *IdentityHandler) quotaService() *appquota.Service {
	return appquota.NewService(h.db, h)
}

func toAppPlanQuotaInputs(quotas []planQuotaInput) []appquota.PlanQuotaInput {
	out := make([]appquota.PlanQuotaInput, 0, len(quotas))
	for _, item := range quotas {
		out = append(out, appquota.PlanQuotaInput{QuotaID: item.QuotaID, QuotaValue: item.QuotaValue})
	}
	return out
}

func savePlanQuotasWithValuesTx(tx *gorm.DB, planID uint64, quotas []planQuotaInput) error {
	if err := lockPlanForUpdate(tx, planID); err != nil {
		return err
	}
	if err := tx.Where("plan_id = ?", planID).Delete(&models.SaasPlanQuota{}).Error; err != nil {
		return err
	}
	seen := map[uint64]struct{}{}
	for _, item := range quotas {
		if item.QuotaID == 0 {
			continue
		}
		if _, ok := seen[item.QuotaID]; ok {
			continue
		}
		seen[item.QuotaID] = struct{}{}
		var quota models.SaasQuota
		if err := tx.Where("id = ? AND status = ?", item.QuotaID, 1).First(&quota).Error; err != nil {
			return fmt.Errorf("配额不存在或已停用")
		}
		if err := tx.Create(&models.SaasPlanQuota{PlanID: planID, QuotaID: item.QuotaID, QuotaValue: item.QuotaValue}).Error; err != nil {
			return err
		}
	}
	return nil
}
