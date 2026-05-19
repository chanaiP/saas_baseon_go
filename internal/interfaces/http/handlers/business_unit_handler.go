package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) BusinessUnits(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	skip, limit := paginationParams(c)
	tenantID := user.TenantID
	var rows []models.BusinessUnit
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if useBUFilter, ids := h.businessUnitDataScopeFilter(user); useBUFilter {
		if len(ids) == 0 {
			response.OK(c, paginatedWithTotal([]gin.H{}, 0, skip, limit))
			return
		}
		query = query.Where("id IN ?", ids)
	}
	var total int64
	_ = query.Model(&models.BusinessUnit{}).Count(&total).Error
	_ = query.Order("id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) BusinessUnitTree(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tenantID := user.TenantID
	var rows []models.BusinessUnit
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if useBUFilter, ids := h.businessUnitDataScopeFilter(user); useBUFilter {
		if len(ids) == 0 {
			response.OK(c, []gin.H{})
			return
		}
		query = query.Where("id IN ?", ids)
	}
	_ = query.Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessUnit(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Name              string   `json:"name"`
		Code              string   `json:"code"`
		BUType            *string  `json:"bu_type"`
		UnitScenario      *string  `json:"unit_scenario"`
		UnitForm          *string  `json:"unit_form"`
		ParentID          *uint64  `json:"parent_id"`
		OwnerUserID       *uint64  `json:"owner_user_id"`
		OwnerOrgID        *uint64  `json:"owner_org_id"`
		OperationEnabled  *bool    `json:"operation_enabled"`
		SettlementEnabled *bool    `json:"settlement_enabled"`
		DataScopeEnabled  *bool    `json:"data_scope_enabled"`
		OrgNodeIDs        []uint64 `json:"org_node_ids"`
		Status            int      `json:"status"`
		Remark            *string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	tenantID := user.TenantID
	if err := h.requireFeatureAccess(tenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	if body.Status == 1 {
		if err := h.requireQuotaAvailable(tenantID, "max_business_units", 1); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	unitForm := nullableTrimmed(body.UnitForm)
	if unitForm == nil {
		unitForm = nullableTrimmed(body.BUType)
	}
	if unitForm == nil {
		response.Error(c, 400, response.CodeBadRequest, "请选择业务单元类型")
		return
	}
	unitScenario := nullableTrimmed(body.UnitScenario)
	unitTypeName := ""
	unitGroupName := ""
	if unitScenario != nil {
		var err error
		unitTypeName, unitGroupName, err = h.validateBusinessUnitDictPair(c, user, strings.TrimSpace(*unitForm), strings.TrimSpace(*unitScenario))
		if err != nil {
			respondBadRequest(c, err)
			return
		}
	} else {
		var err error
		unitTypeName, err = h.validateBusinessUnitRootDictValue(c.Request.Context(), user, unitForm, "业务单元类型")
		if err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	if err := h.validateBusinessUnitRefs(tenantID, 0, body.ParentID, body.OwnerUserID, body.OwnerOrgID); err != nil {
		respondBadRequest(c, err)
		return
	}
	if len(body.OrgNodeIDs) > 0 {
		if err := h.validateBusinessUnitOrgMaps(tenantID, 0, body.OrgNodeIDs); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	operationEnabled := boolValue(body.OperationEnabled, true)
	settlementEnabled := boolValue(body.SettlementEnabled, body.Status == 1)
	row := models.BusinessUnit{
		TenantID:          tenantID,
		Name:              strings.TrimSpace(body.Name),
		Code:              strings.TrimSpace(body.Code),
		BUType:            nullableTrimmed(body.BUType),
		UnitScenario:      unitScenario,
		UnitForm:          unitForm,
		UnitTypeCode:      unitForm,
		UnitTypeName:      nullableFromString(unitTypeName),
		UnitGroupCode:     unitScenario,
		UnitGroupName:     nullableFromString(unitGroupName),
		ParentID:          body.ParentID,
		OwnerUserID:       body.OwnerUserID,
		OwnerOrgID:        body.OwnerOrgID,
		Status:            body.Status,
		BillingEnabled:    body.Status == 1,
		StatisticEnabled:  true,
		OperationEnabled:  operationEnabled,
		SettlementEnabled: settlementEnabled,
		DataScopeEnabled:  boolValue(body.DataScopeEnabled, false),
		Remark:            nullableTrimmed(body.Remark),
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		if len(body.OrgNodeIDs) > 0 {
			return h.replaceBusinessUnitMappingsTx(tx, tenantID, row.ID, body.OrgNodeIDs)
		}
		return nil
	}); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit", "create", "创建业务单元 "+row.Name, gin.H{"business_unit_id": row.ID, "tenant_id": row.TenantID, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateBusinessUnit(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tenantID := user.TenantID
	var existing models.BusinessUnit
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", c.Param("id"), tenantID).First(&existing).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	var body struct {
		Name              *string  `json:"name"`
		Code              *string  `json:"code"`
		BUType            *string  `json:"bu_type"`
		UnitScenario      *string  `json:"unit_scenario"`
		UnitForm          *string  `json:"unit_form"`
		ParentID          *uint64  `json:"parent_id"`
		OwnerUserID       *uint64  `json:"owner_user_id"`
		OwnerOrgID        *uint64  `json:"owner_org_id"`
		OperationEnabled  *bool    `json:"operation_enabled"`
		SettlementEnabled *bool    `json:"settlement_enabled"`
		DataScopeEnabled  *bool    `json:"data_scope_enabled"`
		OrgNodeIDs        []uint64 `json:"org_node_ids"`
		Status            *int     `json:"status"`
		Remark            *string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.requireFeatureAccess(tenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	if body.Status != nil && *body.Status == 1 && existing.Status != 1 {
		if err := h.requireQuotaAvailable(tenantID, "max_business_units", 1); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	if body.OrgNodeIDs != nil {
		if err := h.validateBusinessUnitOrgMaps(tenantID, existing.ID, body.OrgNodeIDs); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	if err := h.validateBusinessUnitRefs(tenantID, existing.ID, body.ParentID, body.OwnerUserID, body.OwnerOrgID); err != nil {
		respondBadRequest(c, err)
		return
	}
	nextUnitForm := stringValue(existing.UnitForm)
	if nextUnitForm == "" {
		nextUnitForm = stringValue(existing.UnitTypeCode)
	}
	if body.BUType != nil && strings.TrimSpace(*body.BUType) != "" && body.UnitForm == nil {
		nextUnitForm = strings.TrimSpace(*body.BUType)
	}
	if body.UnitForm != nil {
		nextUnitForm = strings.TrimSpace(*body.UnitForm)
	}
	nextUnitScenario := stringValue(existing.UnitScenario)
	if nextUnitScenario == "" {
		nextUnitScenario = stringValue(existing.UnitGroupCode)
	}
	if body.UnitScenario != nil {
		nextUnitScenario = strings.TrimSpace(*body.UnitScenario)
	}
	unitTypeName := stringValue(existing.UnitTypeName)
	unitGroupName := stringValue(existing.UnitGroupName)
	if body.BUType != nil || body.UnitForm != nil || body.UnitScenario != nil {
		if nextUnitForm == "" {
			response.Error(c, 400, response.CodeBadRequest, "请选择业务单元类型")
			return
		}
		if nextUnitScenario != "" {
			var err error
			unitTypeName, unitGroupName, err = h.validateBusinessUnitDictPair(c, user, nextUnitForm, nextUnitScenario)
			if err != nil {
				respondBadRequest(c, err)
				return
			}
		} else {
			var err error
			unitTypeName, err = h.validateBusinessUnitRootDictValue(c.Request.Context(), user, &nextUnitForm, "业务单元类型")
			if err != nil {
				respondBadRequest(c, err)
				return
			}
			unitGroupName = ""
		}
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Code != nil {
		updates["code"] = strings.TrimSpace(*body.Code)
	}
	if body.BUType != nil {
		if strings.TrimSpace(*body.BUType) == "" {
			response.Error(c, 400, response.CodeBadRequest, "请选择业务单元类型")
			return
		}
		updates["bu_type"] = strings.TrimSpace(*body.BUType)
		if body.UnitForm == nil {
			nextType := strings.TrimSpace(*body.BUType)
			updates["unit_form"] = nextType
			updates["unit_type_code"] = nextType
			updates["unit_type_name"] = nullableFromString(unitTypeName)
		}
	}
	if body.UnitScenario != nil {
		unitScenario := nullableTrimmed(body.UnitScenario)
		updates["unit_scenario"] = unitScenario
		updates["unit_group_code"] = unitScenario
		updates["unit_group_name"] = nullableFromString(unitGroupName)
	}
	if body.UnitForm != nil {
		unitForm := nullableTrimmed(body.UnitForm)
		if unitForm == nil {
			response.Error(c, 400, response.CodeBadRequest, "请选择业务单元类型")
			return
		}
		updates["unit_form"] = unitForm
		updates["unit_type_code"] = unitForm
		updates["unit_type_name"] = nullableFromString(unitTypeName)
	}
	if body.ParentID != nil {
		updates["parent_id"] = *body.ParentID
	}
	if body.OwnerUserID != nil {
		updates["owner_user_id"] = *body.OwnerUserID
	}
	if body.OwnerOrgID != nil {
		updates["owner_org_id"] = *body.OwnerOrgID
	}
	if body.Status != nil {
		updates["status"] = *body.Status
		updates["billing_enabled"] = *body.Status == 1
		updates["statistic_enabled"] = true
	}
	if body.OperationEnabled != nil {
		updates["operation_enabled"] = *body.OperationEnabled
	}
	if body.SettlementEnabled != nil {
		updates["settlement_enabled"] = *body.SettlementEnabled
	}
	if body.DataScopeEnabled != nil {
		updates["data_scope_enabled"] = *body.DataScopeEnabled
	}
	if body.Remark != nil {
		updates["remark"] = nullableTrimmed(body.Remark)
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		}
		if body.OrgNodeIDs != nil {
			return h.replaceBusinessUnitMappingsTx(tx, tenantID, parseUintParam(c, "id"), body.OrgNodeIDs)
		}
		return nil
	}); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit", "update", "更新业务单元", gin.H{"business_unit_id": parseUintParam(c, "id"), "tenant_id": tenantID, "changes": updates, "org_node_ids": body.OrgNodeIDs})
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeleteBusinessUnit(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "业务单元", ref(&models.BusinessUnitScope{}, "数据权限范围", "business_unit_id = ?", id)) {
		return
	}
	var row models.BusinessUnit
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, user.TenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "status": 0, "billing_enabled": false, "statistic_enabled": true, "code": tombstoneUniqueValue(row.Code, row.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit", "delete", "删除业务单元 "+row.Name, gin.H{"business_unit_id": id})
	response.OK(c, gin.H{"deleted": id})
}
