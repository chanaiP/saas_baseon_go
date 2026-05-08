package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) createOrgNode(c *gin.Context, forcedType string) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body orgNodePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if forcedType != "" {
		body.NodeType = forcedType
	}
	if body.NodeType == "" {
		body.NodeType = "department"
	}
	if body.Status == 0 {
		body.Status = 1
	}
	tenantID := user.TenantID
	if err := h.requireFeatureAccess(tenantID, "org_manage"); err != nil {
		response.Error(c, 403, response.CodeForbidden, err.Error())
		return
	}
	parentID, companyID, err := h.resolveOrgNodeCreatePlacement(tenantID, body)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	quotaCode := quotaCodeForOrgNodeType(body.NodeType)
	if quotaCode != "" {
		if err := h.requireQuotaAvailable(tenantID, quotaCode, 1); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
	}
	row := models.OrgNode{TenantID: tenantID, NodeType: strings.TrimSpace(body.NodeType), Name: strings.TrimSpace(body.Name), Code: nullableTrimmed(body.Code), CompanyType: h.normalizedCompanyType(body.NodeType, body.CompanyType), CompanyID: companyID, ParentID: parentID, Status: body.Status}
	if row.Name == "" {
		response.Error(c, 400, response.CodeBadRequest, "组织名称不能为空")
		return
	}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.audit(c, user.TenantID, user.ID, "organization", "create", "创建组织 "+row.Name, gin.H{"org_id": row.ID, "tenant_id": row.TenantID, "node_type": row.NodeType, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) updateOrgNode(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body orgNodePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tenantID := user.TenantID
	id := parseUintParam(c, "id")
	var row models.OrgNode
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "组织不存在")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = strings.TrimSpace(body.Name)
	}
	if body.Code != nil {
		updates["code"] = nullableTrimmed(body.Code)
	}
	if body.Status != 0 {
		updates["status"] = body.Status
	}
	nextType := row.NodeType
	if strings.TrimSpace(body.NodeType) != "" {
		nextType = strings.TrimSpace(body.NodeType)
		updates["node_type"] = nextType
	}
	if body.CompanyType != nil || nextType != row.NodeType {
		updates["company_type"] = h.normalizedCompanyType(nextType, body.CompanyType)
	}
	if body.ParentIDSet {
		if err := h.validateOrgParentChange(tenantID, row.ID, body.ParentID); err != nil {
			response.Error(c, 400, response.CodeBadRequest, err.Error())
			return
		}
		updates["parent_id"] = body.ParentID
		if nextType == "company" {
			updates["company_id"] = nil
		} else {
			updates["company_id"] = h.resolveCompanyIDFromParent(tenantID, body.ParentID)
		}
	} else if nextType != row.NodeType {
		if nextType == "company" {
			updates["company_id"] = nil
		} else {
			updates["company_id"] = h.resolveCompanyIDFromParent(tenantID, row.ParentID)
		}
	}
	if len(updates) == 0 {
		response.OK(c, gin.H{"id": id})
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return err
		}
		if _, ok := updates["parent_id"]; ok {
			return h.cascadeRefreshCompanyIDs(tx, tenantID, row.ID)
		}
		if _, ok := updates["node_type"]; ok {
			return h.cascadeRefreshCompanyIDs(tx, tenantID, row.ID)
		}
		return nil
	}); err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.audit(c, user.TenantID, user.ID, "organization", "update", "更新组织", gin.H{"org_id": id, "changes": updates})
	response.OK(c, gin.H{"id": id})
}

func (h *IdentityHandler) deleteOrgNodeWithAudit(c *gin.Context, summary string) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "组织",
		ref(&models.OrgNode{}, "下级组织", "tenant_id = ? AND deleted_at IS NULL AND (parent_id = ? OR company_id = ?)", user.TenantID, id, id),
		ref(&models.AppUser{}, "用户主组织", "tenant_id = ? AND deleted_at IS NULL AND (company_id = ? OR department_id = ?)", user.TenantID, id, id),
		ref(&models.AppUserDepartment{}, "用户兼任部门", "department_id = ?", id),
		ref(&models.BusinessUnitOrgMap{}, "业务单元组织映射", "tenant_id = ? AND org_id = ?", user.TenantID, id),
	) {
		return
	}
	var row models.OrgNode
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, user.TenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "组织节点不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "status": 0}).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	h.audit(c, user.TenantID, user.ID, "organization", "delete", summary, gin.H{"org_id": id})
	response.OK(c, gin.H{"deleted": id})
}
