package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const businessUnitDictCode = "business_unit"

func (h *IdentityHandler) validateBusinessUnitDictValue(ctx context.Context, user models.AppUser, dictCode string, value *string, fieldName string) error {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	_, items, err := h.dictionaryService().ItemsByCode(ctx, h.dictionaryViewer(user), dictCode)
	if err != nil {
		return err
	}
	needle := strings.TrimSpace(*value)
	for _, item := range items {
		if item.Enabled && item.Value == needle {
			return nil
		}
	}
	return fmt.Errorf("%s不在字典范围内", fieldName)
}

func (h *IdentityHandler) validateBusinessUnitRootDictValue(ctx context.Context, user models.AppUser, value *string, fieldName string) (string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "", nil
	}
	_, items, err := h.dictionaryService().ItemsByCode(ctx, h.dictionaryViewer(user), businessUnitDictCode)
	if err != nil {
		return "", err
	}
	needle := strings.TrimSpace(*value)
	for _, item := range items {
		if item.Enabled && item.ParentID == nil && item.Value == needle {
			return item.Label, nil
		}
	}
	return "", fmt.Errorf("%s不在字典范围内", fieldName)
}

func (h *IdentityHandler) validateBusinessUnitRefs(tenantID uint64, currentID uint64, parentID *uint64, ownerUserID *uint64, ownerOrgID *uint64) error {
	if parentID != nil {
		if *parentID == currentID && currentID != 0 {
			return errors.New("上级业务单元不能选择自身")
		}
		var total int64
		if err := h.db.Model(&models.BusinessUnit{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", *parentID, tenantID).Count(&total).Error; err != nil {
			return err
		}
		if total == 0 {
			return errors.New("上级业务单元不存在")
		}
	}
	if ownerUserID != nil {
		var total int64
		if err := h.db.Model(&models.AppUser{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", *ownerUserID, tenantID).Count(&total).Error; err != nil {
			return err
		}
		if total == 0 {
			return errors.New("负责人用户不存在")
		}
	}
	if ownerOrgID != nil {
		if _, err := h.orgNodeByID(tenantID, *ownerOrgID); err != nil {
			return errors.New("主负责组织不存在")
		}
	}
	return nil
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func (h *IdentityHandler) replaceBusinessUnitMappingsTx(tx *gorm.DB, tenantID, buID uint64, orgIDs []uint64) error {
	ids := uniqueUint64s(orgIDs)
	if err := tx.Model(&models.BusinessUnitOrgMap{}).Where("tenant_id = ? AND business_unit_id = ? AND scope_type = ?", tenantID, buID, "PRIMARY").Update("status", 0).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	for _, orgID := range ids {
		node, err := h.orgNodeByID(tenantID, orgID)
		if err != nil {
			return errors.New("组织节点不存在")
		}
		if err := tx.Create(&models.BusinessUnitOrgMap{TenantID: tenantID, BusinessUnitID: buID, OrgID: orgID, OrgType: orgTypeForBusinessUnitMap(node), ScopeType: "PRIMARY", Priority: 0, Status: 1}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *IdentityHandler) validateBusinessUnitOrgMaps(tenantID, excludeBUID uint64, orgIDs []uint64) error {
	ids := uniqueUint64s(orgIDs)
	if len(ids) == 0 {
		return nil
	}
	for _, orgID := range ids {
		if _, err := h.orgNodeByID(tenantID, orgID); err != nil {
			return errors.New("组织节点不存在")
		}
		var row models.BusinessUnit
		query := h.db.Model(&models.BusinessUnit{}).
			Joins("JOIN business_unit_org_map m ON m.business_unit_id = business_unit.id").
			Where("business_unit.tenant_id = ? AND business_unit.deleted_at IS NULL AND business_unit.status = ? AND m.tenant_id = ? AND m.org_id = ? AND m.scope_type = ? AND m.status = ?", tenantID, 1, tenantID, orgID, "PRIMARY", 1)
		if excludeBUID > 0 {
			query = query.Where("business_unit.id <> ?", excludeBUID)
		}
		if err := query.First(&row).Error; err == nil {
			return fmt.Errorf("组织节点已关联到业务单元「%s」，不能重复关联", row.Name)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return nil
}

func (h *IdentityHandler) businessUnitDataScopeFilter(user models.AppUser) (bool, []uint64) {
	if user.IsPlatformAdmin {
		return false, nil
	}
	var links []models.RolePermission
	_ = h.db.Joins("JOIN permission p ON p.id = role_permission.permission_id").
		Joins("JOIN user_role ur ON ur.role_id = role_permission.role_id").
		Joins("JOIN role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND p.tenant_id = ? AND p.path = ? AND p.perm_type = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL", user.ID, user.TenantID, "data:business_unit", 4).
		Find(&links).Error
	if len(links) == 0 {
		return false, nil
	}
	for _, link := range links {
		scope := strings.TrimSpace(derefString(link.DataScopeOverride))
		if scope == "" {
			var permission models.Permission
			if err := h.db.Where("id = ? AND deleted_at IS NULL", link.PermissionID).First(&permission).Error; err == nil && permission.DataScope != nil {
				scope = strings.TrimSpace(*permission.DataScope)
			}
		}
		if scope == "ALL" {
			return false, nil
		}
	}
	useFilter := false
	ids := []uint64{}
	for _, link := range links {
		mode := strings.TrimSpace(derefString(link.BUDataAccessMode))
		specified := uint64IDsFromJSON(link.CustomBusinessUnitIDsJSON)
		switch mode {
		case "CURRENT_ORG_BU":
			useFilter = true
			anchor := user.DepartmentID
			if anchor == nil {
				anchor = user.CompanyID
			}
			if anchor != nil {
				ids = append(ids, h.businessUnitIDsForOrgSubtree(user.TenantID, *anchor)...)
			}
		case "SPECIFIED_BU":
			useFilter = true
			ids = append(ids, specified...)
		default:
			if len(specified) > 0 {
				useFilter = true
				ids = append(ids, specified...)
			}
		}
	}
	return useFilter, uniqueUint64s(ids)
}

func (h *IdentityHandler) businessUnitIDsForOrgSubtree(tenantID, rootOrgID uint64) []uint64 {
	orgIDs := h.descendantOrgNodeIDs(tenantID, rootOrgID)
	if len(orgIDs) == 0 {
		return []uint64{}
	}
	var ids []uint64
	_ = h.db.Model(&models.BusinessUnitOrgMap{}).
		Joins("JOIN business_unit b ON b.id = business_unit_org_map.business_unit_id").
		Where("business_unit_org_map.tenant_id = ? AND business_unit_org_map.org_id IN ? AND business_unit_org_map.status = ? AND b.status = ? AND b.deleted_at IS NULL", tenantID, orgIDs, 1, 1).
		Distinct().
		Pluck("business_unit_org_map.business_unit_id", &ids).Error
	return uniqueUint64s(ids)
}

func orgTypeForBusinessUnitMap(node models.OrgNode) string {
	switch strings.ToLower(strings.TrimSpace(node.NodeType)) {
	case "company":
		return "company"
	case "store":
		return "store"
	default:
		return "department"
	}
}
