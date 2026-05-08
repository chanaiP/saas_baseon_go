package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) BusinessUnitOrgMappings(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tenantID := user.TenantID
	var rows []models.BusinessUnitOrgMap
	_ = h.db.Where("tenant_id = ? AND business_unit_id = ? AND status = ?", tenantID, c.Param("id"), 1).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitOrgMapToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessUnitOrgMapping(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tenantID := user.TenantID
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", c.Param("id"), tenantID).First(&models.BusinessUnit{}).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	var body struct {
		OrgID     uint64 `json:"org_id"`
		ScopeType string `json:"scope_type"`
		Priority  int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	scopeType := strings.TrimSpace(body.ScopeType)
	if scopeType == "" {
		scopeType = "PRIMARY"
	}
	if scopeType == "PRIMARY" {
		if err := h.validateBusinessUnitOrgMaps(tenantID, parseUintParam(c, "id"), []uint64{body.OrgID}); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	node, err := h.orgNodeByID(tenantID, body.OrgID)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "组织节点不存在")
		return
	}
	var exists int64
	_ = h.db.Model(&models.BusinessUnitOrgMap{}).Where("tenant_id = ? AND business_unit_id = ? AND org_id = ? AND scope_type = ? AND status = ?", tenantID, parseUintParam(c, "id"), body.OrgID, scopeType, 1).Count(&exists).Error
	if exists > 0 {
		response.Error(c, 400, response.CodeBadRequest, "该组织节点已映射到此业务单元")
		return
	}
	row := models.BusinessUnitOrgMap{TenantID: tenantID, BusinessUnitID: parseUintParam(c, "id"), OrgID: body.OrgID, OrgType: orgTypeForBusinessUnitMap(node), ScopeType: scopeType, Priority: body.Priority, Status: 1}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_org_map", "create", "业务单元组织映射", gin.H{"mapping_id": row.ID, "business_unit_id": row.BusinessUnitID, "org_id": row.OrgID, "tenant_id": row.TenantID})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) DeleteBusinessUnitOrgMapping(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	result := h.db.Model(&models.BusinessUnitOrgMap{}).Where("id = ? AND tenant_id = ?", id, user.TenantID).Update("status", 0)
	if result.Error != nil {
		respondBadRequest(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "映射不存在")
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_org_map", "delete", "删除业务单元组织映射", gin.H{"mapping_id": id})
	response.OK(c, gin.H{"deleted": id})
}
