package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Plans(c *gin.Context) {
	skip, limit := paginationParams(c)
	query := h.db.Where("deleted_at IS NULL")
	var total int64
	_ = query.Model(&models.SaasPlan{}).Count(&total).Error
	var rows []models.SaasPlan
	_ = query.Order("sort_order asc, id asc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]dto.PlanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, planToResponse(row))
	}
	response.OK(c, dto.PaginatedResponse[dto.PlanResponse]{Items: items, Total: total, Skip: skip, Limit: limit})
}

func (h *IdentityHandler) CreatePlan(c *gin.Context) {
	var body struct {
		PlanCode     string  `json:"plan_code"`
		PlanName     string  `json:"plan_name"`
		PlanType     string  `json:"plan_type"`
		BillingCycle string  `json:"billing_cycle"`
		Price        float64 `json:"price"`
		Status       int     `json:"status"`
		IsDefault    bool    `json:"is_default"`
		SortOrder    int     `json:"sort_order"`
		Description  *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	row := models.SaasPlan{PlanCode: strings.TrimSpace(body.PlanCode), PlanName: strings.TrimSpace(body.PlanName), PlanType: body.PlanType, BillingCycle: body.BillingCycle, Price: body.Price, Status: body.Status, IsDefault: body.IsDefault, SortOrder: body.SortOrder, Description: body.Description}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, planToResponse(row))
}

func (h *IdentityHandler) UpdatePlan(c *gin.Context) {
	var row models.SaasPlan
	if err := h.db.Where("deleted_at IS NULL").First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "套餐不存在")
		return
	}
	var body planUpdatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.db.Model(&row).Updates(body.Updates()).First(&row, row.ID).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, planToResponse(row))
}

func (h *IdentityHandler) CopyPlan(c *gin.Context) {
	var src models.SaasPlan
	if err := h.db.Where("deleted_at IS NULL").First(&src, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "套餐不存在")
		return
	}
	var body struct {
		PlanCode    string  `json:"plan_code"`
		PlanName    string  `json:"plan_name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var dst models.SaasPlan
	err := h.db.Transaction(func(tx *gorm.DB) error {
		dst = src
		dst.ID = 0
		dst.PlanCode = strings.TrimSpace(body.PlanCode)
		dst.PlanName = strings.TrimSpace(body.PlanName)
		dst.Description = body.Description
		dst.IsDefault = false
		dst.CreatedAt = time.Time{}
		dst.UpdatedAt = time.Time{}
		if err := tx.Create(&dst).Error; err != nil {
			return err
		}
		var features []models.SaasPlanFeature
		_ = tx.Where("plan_id = ?", src.ID).Find(&features).Error
		for _, feature := range features {
			if err := tx.Create(&models.SaasPlanFeature{PlanID: dst.ID, FeatureID: feature.FeatureID, Enabled: feature.Enabled}).Error; err != nil {
				return err
			}
		}
		var quotas []models.SaasPlanQuota
		_ = tx.Where("plan_id = ?", src.ID).Find(&quotas).Error
		for _, quota := range quotas {
			if err := tx.Create(&models.SaasPlanQuota{PlanID: dst.ID, QuotaID: quota.QuotaID, QuotaValue: quota.QuotaValue}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, planToResponse(dst))
}

func (h *IdentityHandler) DeletePlan(c *gin.Context) {
	id := parseUintParam(c, "id")
	now := time.Now()
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := checkDeletionReferences(tx, "套餐", ref(&models.TenantSubscription{}, "主体订阅", "plan_id = ?", id)); err != nil {
			return err
		}
		var plan models.SaasPlan
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&plan).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		return tx.Model(&plan).Updates(map[string]interface{}{"deleted_at": now, "status": 0, "plan_code": tombstoneUniqueValue(plan.PlanCode, plan.ID, 64)}).Error
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, response.CodeNotFound, "套餐不存在")
			return
		}
		h.respondDeletionError(c, err)
		return
	}
	response.OK(c, dto.DeletedResponse{Deleted: id})
}

func (h *IdentityHandler) PlanMatrix(c *gin.Context) {
	h.syncPackageFeaturesFromPermissions()
	var plans []models.SaasPlan
	var features []models.SaasFeature
	var links []models.SaasPlanFeature
	_ = h.db.Where("deleted_at IS NULL").Order("sort_order asc, id asc").Find(&plans).Error
	_ = h.db.Where("status = ?", 1).Order("parent_id asc, id asc").Find(&features).Error
	_ = h.db.Find(&links).Error

	enabled := map[uint64]map[uint64]bool{}
	for _, link := range links {
		if enabled[link.FeatureID] == nil {
			enabled[link.FeatureID] = map[uint64]bool{}
		}
		enabled[link.FeatureID][link.PlanID] = link.Enabled
	}

	planItems := make([]dto.PlanResponse, 0, len(plans))
	for _, plan := range plans {
		planItems = append(planItems, planToResponse(plan))
	}
	nodeByID := map[uint64]dto.PlanMatrixNode{}
	childrenByParent := map[uint64][]uint64{}
	for _, feature := range features {
		cells := make([]dto.PlanMatrixCell, 0, len(plans))
		for _, plan := range plans {
			isEnabled := enabled[feature.ID][plan.ID]
			state := "disabled"
			if isEnabled {
				state = "enabled"
			}
			cells = append(cells, dto.PlanMatrixCell{PlanID: plan.ID, PlanCode: plan.PlanCode, Enabled: isEnabled, State: state, FeatureIDs: []uint64{feature.ID}, QuotaValues: h.planMatrixQuotaValues(plan.ID, feature.FeatureCode)})
		}
		node := dto.PlanMatrixNode{ID: feature.FeatureCode, Label: feature.FeatureName, NodeType: "feature", FeatureID: feature.ID, FeatureCode: feature.FeatureCode, FeatureType: feature.FeatureType, Description: feature.Description, Children: []dto.PlanMatrixNode{}, Cells: cells}
		nodeByID[feature.ID] = node
		childrenByParent[feature.ParentID] = append(childrenByParent[feature.ParentID], feature.ID)
	}
	var buildNodes func(parentID uint64) []dto.PlanMatrixNode
	buildNodes = func(parentID uint64) []dto.PlanMatrixNode {
		ids := childrenByParent[parentID]
		nodes := make([]dto.PlanMatrixNode, 0, len(ids))
		for _, id := range ids {
			node := nodeByID[id]
			node.Children = buildNodes(id)
			nodes = append(nodes, node)
		}
		return nodes
	}
	nodes := buildNodes(0)
	response.OK(c, dto.PlanMatrixResponse{Plans: planItems, Nodes: nodes})
}
