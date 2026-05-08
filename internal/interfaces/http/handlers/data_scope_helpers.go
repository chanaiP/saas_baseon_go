package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type orgDataScope struct {
	Scope         string
	CompanyIDs    []uint64
	DepartmentIDs []uint64
	UserIDs       []uint64
}

func (h *IdentityHandler) resolveOrganizationDataScope(user models.AppUser) orgDataScope {
	return h.resolveDataScopeForMenu(user, "/organization")
}

func (h *IdentityHandler) resolveDataScopeForMenu(user models.AppUser, menuPath string) orgDataScope {
	if user.IsPlatformAdmin {
		return orgDataScope{Scope: "ALL"}
	}
	prefixes := operationPrefixesForMenuPath(menuPath)
	if len(prefixes) == 0 {
		return orgDataScope{Scope: "SELF"}
	}
	dataPath := "data:" + prefixes[0]
	var links []models.RolePermission
	_ = h.db.Joins("JOIN permission p ON p.id = role_permission.permission_id").
		Joins("JOIN user_role ur ON ur.role_id = role_permission.role_id").
		Joins("JOIN role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND p.tenant_id = ? AND p.path = ? AND p.perm_type = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL", user.ID, user.TenantID, dataPath, 4).
		Find(&links).Error
	if len(links) == 0 {
		return orgDataScope{Scope: "SELF"}
	}
	scopes := make([]string, 0, len(links))
	companyIDs, departmentIDs, userIDs := []uint64{}, []uint64{}, []uint64{}
	for _, link := range links {
		scope := derefString(link.DataScopeOverride)
		if scope == "" {
			var permission models.Permission
			if err := h.db.Where("id = ? AND deleted_at IS NULL", link.PermissionID).First(&permission).Error; err == nil && permission.DataScope != nil {
				scope = strings.TrimSpace(*permission.DataScope)
			}
		}
		if scope == "" {
			scope = "ALL"
		}
		scopes = append(scopes, scope)
		if scope == "CUSTOM" {
			companyIDs = append(companyIDs, uint64IDsFromJSON(link.CustomCompanyIDsJSON)...)
			departmentIDs = append(departmentIDs, uint64IDsFromJSON(link.CustomDepartmentIDsJSON)...)
			userIDs = append(userIDs, uint64IDsFromJSON(link.CustomUserIDsJSON)...)
		}
	}
	return chooseOrgDataScope(scopes, companyIDs, departmentIDs, userIDs)
}

func chooseOrgDataScope(scopes []string, companyIDs []uint64, departmentIDs []uint64, userIDs []uint64) orgDataScope {
	for _, scope := range scopes {
		if scope == "ALL" {
			return orgDataScope{Scope: "ALL"}
		}
	}
	for _, scope := range scopes {
		if scope == "CUSTOM" {
			return orgDataScope{Scope: "CUSTOM", CompanyIDs: uniqueUint64s(companyIDs), DepartmentIDs: uniqueUint64s(departmentIDs), UserIDs: uniqueUint64s(userIDs)}
		}
	}
	priority := map[string]int{"SELF": 0, "ORG": 1, "ORG_SUB": 2}
	tightest := "SELF"
	best := 99
	for _, scope := range scopes {
		if rank, ok := priority[scope]; ok && rank < best {
			best = rank
			tightest = scope
		}
	}
	return orgDataScope{Scope: tightest}
}

func (h *IdentityHandler) allowedOrgIDsForScope(user models.AppUser, scope orgDataScope) (map[uint64]struct{}, map[uint64]struct{}) {
	companyIDs := map[uint64]struct{}{}
	departmentIDs := map[uint64]struct{}{}
	addCompany := func(id *uint64) {
		if id != nil && *id != 0 {
			companyIDs[*id] = struct{}{}
		}
	}
	addDepartment := func(id *uint64) {
		if id != nil && *id != 0 {
			departmentIDs[*id] = struct{}{}
		}
	}
	switch scope.Scope {
	case "SELF", "ORG":
		addCompany(user.CompanyID)
		addDepartment(user.DepartmentID)
	case "ORG_SUB":
		addCompany(user.CompanyID)
		addDepartment(user.DepartmentID)
		if user.DepartmentID != nil {
			for _, id := range h.descendantOrgNodeIDs(user.TenantID, *user.DepartmentID) {
				departmentIDs[id] = struct{}{}
			}
		}
	case "CUSTOM":
		for _, id := range scope.CompanyIDs {
			companyIDs[id] = struct{}{}
		}
		for _, id := range scope.DepartmentIDs {
			departmentIDs[id] = struct{}{}
			for _, childID := range h.descendantOrgNodeIDs(user.TenantID, id) {
				departmentIDs[childID] = struct{}{}
			}
		}
	}
	return companyIDs, departmentIDs
}

func (h *IdentityHandler) applyUserDataScopeFilter(query *gorm.DB, user models.AppUser) *gorm.DB {
	scope := h.resolveDataScopeForMenu(user, "/users")
	if scope.Scope == "ALL" {
		return query
	}
	departmentMatch := func(ids []uint64) *gorm.DB {
		if len(ids) == 0 {
			return h.db.Where("1 = 0")
		}
		var userIDs []uint64
		_ = h.db.Model(&models.AppUserDepartment{}).Where("department_id IN ?", ids).Distinct().Pluck("user_id", &userIDs).Error
		if len(userIDs) > 0 {
			return h.db.Where("department_id IN ? OR id IN ?", ids, userIDs)
		}
		return h.db.Where("department_id IN ?", ids)
	}
	switch scope.Scope {
	case "SELF":
		return query.Where("id = ?", user.ID)
	case "ORG":
		parts := make([]*gorm.DB, 0, 2)
		if user.CompanyID != nil {
			parts = append(parts, h.db.Where("company_id = ?", *user.CompanyID))
		}
		if user.DepartmentID != nil {
			parts = append(parts, departmentMatch([]uint64{*user.DepartmentID}))
		}
		return applyOrScopes(query, parts)
	case "ORG_SUB":
		parts := make([]*gorm.DB, 0, 2)
		if user.CompanyID != nil {
			parts = append(parts, h.db.Where("company_id = ?", *user.CompanyID))
		}
		if user.DepartmentID != nil {
			parts = append(parts, departmentMatch(h.descendantOrgNodeIDs(user.TenantID, *user.DepartmentID)))
		}
		return applyOrScopes(query, parts)
	case "CUSTOM":
		parts := make([]*gorm.DB, 0, 3)
		if len(scope.CompanyIDs) > 0 {
			parts = append(parts, h.db.Where("company_id IN ?", scope.CompanyIDs))
		}
		if len(scope.DepartmentIDs) > 0 {
			allDeptIDs := []uint64{}
			for _, id := range scope.DepartmentIDs {
				allDeptIDs = append(allDeptIDs, h.descendantOrgNodeIDs(user.TenantID, id)...)
			}
			parts = append(parts, departmentMatch(uniqueUint64s(allDeptIDs)))
		}
		if len(scope.UserIDs) > 0 {
			parts = append(parts, h.db.Where("id IN ?", scope.UserIDs))
		}
		return applyOrScopes(query, parts)
	default:
		return query
	}
}

func applyOrScopes(query *gorm.DB, parts []*gorm.DB) *gorm.DB {
	if len(parts) == 0 {
		return query
	}
	combined := parts[0]
	for _, part := range parts[1:] {
		combined = combined.Or(part)
	}
	return query.Where(combined)
}

func filterOrgTree(nodes []gin.H, allowedCompanyIDs, allowedDepartmentIDs map[uint64]struct{}) []gin.H {
	items := make([]gin.H, 0, len(nodes))
	for _, node := range nodes {
		childrenRaw, _ := node["children"].([]gin.H)
		children := filterOrgTree(childrenRaw, allowedCompanyIDs, allowedDepartmentIDs)
		id, _ := node["id"].(uint64)
		nodeType, _ := node["node_type"].(string)
		companyID := uint64FromGinValue(node["company_id"])
		_, companyHit := allowedCompanyIDs[id]
		_, departmentHit := allowedDepartmentIDs[id]
		_, companyIDHit := allowedCompanyIDs[companyID]
		extendedHit := (nodeType == "warehouse" || nodeType == "project_team") && (departmentHit || companyIDHit)
		if (nodeType == "company" && companyHit) || (nodeType == "department" && departmentHit) || (nodeType == "store" && companyIDHit) || extendedHit || len(children) > 0 {
			next := gin.H{}
			for key, value := range node {
				next[key] = value
			}
			next["children"] = children
			items = append(items, next)
		}
	}
	return items
}

func uint64FromGinValue(value interface{}) uint64 {
	switch v := value.(type) {
	case uint64:
		return v
	case *uint64:
		if v != nil {
			return *v
		}
	}
	return 0
}

func (h *IdentityHandler) descendantOrgNodeIDs(tenantID uint64, rootID uint64) []uint64 {
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Find(&rows).Error
	childrenByParent := map[uint64][]uint64{}
	for _, row := range rows {
		if row.ParentID != nil {
			childrenByParent[*row.ParentID] = append(childrenByParent[*row.ParentID], row.ID)
		}
	}
	out := []uint64{rootID}
	stack := append([]uint64{}, childrenByParent[rootID]...)
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		out = append(out, id)
		stack = append(stack, childrenByParent[id]...)
	}
	return uniqueUint64s(out)
}
