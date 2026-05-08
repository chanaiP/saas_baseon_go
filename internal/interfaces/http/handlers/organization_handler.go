package handlers

import (
	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) OrganizationTree(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tenantID := user.TenantID
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Order("id asc").Find(&rows).Error
	tree := buildOrgTree(rows)
	scope := h.resolveOrganizationDataScope(user)
	if scope.Scope != "ALL" {
		companyIDs, departmentIDs := h.allowedOrgIDsForScope(user, scope)
		tree = filterOrgTree(tree, companyIDs, departmentIDs)
	}
	response.OK(c, tree)
}

func (h *IdentityHandler) OrganizationDetail(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ?", tenantID).Order("id asc").Find(&rows).Error
	companies, departments, stores := []gin.H{}, []gin.H{}, []gin.H{}
	for _, row := range rows {
		item := orgNodeToJSON(row, []gin.H{})
		switch row.NodeType {
		case "company":
			companies = append(companies, item)
		case "department":
			departments = append(departments, item)
		case "store":
			stores = append(stores, item)
		}
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "companies": companies, "departments": departments, "stores": stores})
}

func (h *IdentityHandler) CreateOrgNode(c *gin.Context) {
	h.createOrgNode(c, "")
}

func (h *IdentityHandler) CreateCompany(c *gin.Context) {
	h.createOrgNode(c, "company")
}

func (h *IdentityHandler) CreateDepartment(c *gin.Context) {
	h.createOrgNode(c, "department")
}

func (h *IdentityHandler) CreateStore(c *gin.Context) {
	h.createOrgNode(c, "store")
}

func (h *IdentityHandler) UpdateOrgNode(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) UpdateCompany(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) UpdateDepartment(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) UpdateStore(c *gin.Context) {
	h.updateOrgNode(c)
}

func (h *IdentityHandler) DeleteOrgNode(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除组织")
}

func (h *IdentityHandler) DeleteCompany(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除公司")
}

func (h *IdentityHandler) DeleteDepartment(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除部门")
}

func (h *IdentityHandler) DeleteStore(c *gin.Context) {
	h.deleteOrgNodeWithAudit(c, "删除门店")
}
