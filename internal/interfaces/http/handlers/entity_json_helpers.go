package handlers

import (
	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
)

func businessUnitToJSON(row models.BusinessUnit) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "name": row.Name, "code": row.Code, "bu_type": row.BUType, "status": row.Status, "billing_enabled": row.BillingEnabled, "statistic_enabled": row.StatisticEnabled, "remark": row.Remark}
}

func businessUnitOrgMapToJSON(row models.BusinessUnitOrgMap) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "business_unit_id": row.BusinessUnitID, "org_id": row.OrgID, "org_type": row.OrgType, "scope_type": row.ScopeType, "priority": row.Priority, "status": row.Status}
}

func planToResponse(row models.SaasPlan) dto.PlanResponse {
	return dto.PlanResponse{ID: row.ID, PlanCode: row.PlanCode, PlanName: row.PlanName, PlanType: row.PlanType, BillingCycle: row.BillingCycle, Price: row.Price, Status: row.Status, IsDefault: row.IsDefault, SortOrder: row.SortOrder, Description: row.Description, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func featureToResponse(row models.SaasFeature) dto.FeatureResponse {
	return dto.FeatureResponse{ID: row.ID, FeatureCode: row.FeatureCode, FeatureName: row.FeatureName, FeatureType: row.FeatureType, ParentID: row.ParentID, MenuID: row.MenuID, APIMethod: row.APIMethod, APIPath: row.APIPath, ServiceKey: row.ServiceKey, Status: row.Status, Description: row.Description}
}

func quotaToResponse(row models.SaasQuota) dto.QuotaResponse {
	return dto.QuotaResponse{ID: row.ID, QuotaCode: row.QuotaCode, QuotaName: row.QuotaName, QuotaType: row.QuotaType, PeriodType: row.PeriodType, Unit: row.Unit, Status: row.Status, Description: row.Description}
}

func permissionToJSON(row models.Permission) gin.H {
	return gin.H{
		"id":                 row.ID,
		"tenant_id":          row.TenantID,
		"parent_id":          row.ParentID,
		"name":               row.Name,
		"path":               row.Path,
		"perm_type":          row.PermType,
		"data_scope":         row.DataScope,
		"sort_order":         row.SortOrder,
		"enabled":            row.Enabled,
		"visible":            row.Visible,
		"tenant_visible":     row.Visible,
		"is_platform_only":   row.IsPlatformOnly,
		"is_package_feature": row.IsPackageFeature,
		"tenant_editable":    row.TenantEditable,
		"tenant_edit_scope":  row.TenantEditScope,
		"feature_code":       row.FeatureCode,
		"feature_type":       row.FeatureType,
		"data_perm_mode":     row.DataPermMode,
		"created_at":         row.CreatedAt,
		"updated_at":         row.UpdatedAt,
	}
}

func (h *IdentityHandler) permissionToJSON(row models.Permission) gin.H {
	item := permissionToJSON(row)
	var departmentIDs []uint64
	_ = h.db.Model(&models.PermissionCustomDepartment{}).Where("permission_id = ?", row.ID).Order("id asc").Pluck("department_id", &departmentIDs).Error
	var userIDs []uint64
	_ = h.db.Model(&models.PermissionCustomUser{}).Where("permission_id = ?", row.ID).Order("id asc").Pluck("user_id", &userIDs).Error
	item["custom_department_ids"] = departmentIDs
	item["custom_user_ids"] = userIDs
	return item
}

func orgNodeToJSON(row models.OrgNode, children []gin.H) gin.H {
	return gin.H{
		"id":           row.ID,
		"node_type":    row.NodeType,
		"name":         row.Name,
		"code":         row.Code,
		"company_type": row.CompanyType,
		"company_id":   row.CompanyID,
		"store_id":     nil,
		"parent_id":    row.ParentID,
		"status":       row.Status,
		"children":     children,
	}
}

func buildOrgTree(rows []models.OrgNode) []gin.H {
	byID := map[uint64]models.OrgNode{}
	childrenByParent := map[uint64][]models.OrgNode{}
	roots := make([]models.OrgNode, 0)
	for _, row := range rows {
		byID[row.ID] = row
		if row.ParentID == nil {
			roots = append(roots, row)
			continue
		}
		childrenByParent[*row.ParentID] = append(childrenByParent[*row.ParentID], row)
	}
	var companyIDFor func(models.OrgNode) *uint64
	companyIDFor = func(row models.OrgNode) *uint64 {
		if row.NodeType == "company" {
			return &row.ID
		}
		if row.CompanyID != nil {
			return row.CompanyID
		}
		if row.ParentID == nil {
			return nil
		}
		parent, ok := byID[*row.ParentID]
		if !ok {
			return nil
		}
		return companyIDFor(parent)
	}
	ancestorStoreID := func(row models.OrgNode) *uint64 {
		if row.ParentID == nil {
			return nil
		}
		current, ok := byID[*row.ParentID]
		for ok {
			if current.NodeType == "store" {
				return &current.ID
			}
			if current.ParentID == nil {
				return nil
			}
			current, ok = byID[*current.ParentID]
		}
		return nil
	}
	var walk func([]models.OrgNode) []gin.H
	walk = func(nodes []models.OrgNode) []gin.H {
		items := make([]gin.H, 0, len(nodes))
		for _, node := range nodes {
			companyID := companyIDFor(node)
			item := gin.H{"id": node.ID, "node_type": node.NodeType, "name": node.Name, "code": node.Code, "status": node.Status, "company_id": companyID, "parent_id": node.ParentID, "children": walk(childrenByParent[node.ID]), "company_type": nil}
			if node.NodeType == "group" || node.NodeType == "company" || node.NodeType == "store" {
				item["company_type"] = node.CompanyType
			}
			if node.NodeType == "department" || node.NodeType == "warehouse" || node.NodeType == "project_team" {
				item["store_id"] = ancestorStoreID(node)
			}
			items = append(items, item)
		}
		return items
	}
	return walk(roots)
}
