package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type rolePayload struct {
	Code          string                    `json:"code"`
	Name          *string                   `json:"name"`
	Description   *string                   `json:"description"`
	PermissionIDs []uint64                  `json:"permission_ids"`
	DataOverrides []roleDataOverridePayload `json:"data_overrides"`
}

type roleDataOverridePayload struct {
	PermissionID          uint64   `json:"permission_id"`
	DataScope             string   `json:"data_scope"`
	CustomCompanyIDs      []uint64 `json:"custom_company_ids"`
	CustomDepartmentIDs   []uint64 `json:"custom_department_ids"`
	CustomUserIDs         []uint64 `json:"custom_user_ids"`
	CustomBusinessUnitIDs []uint64 `json:"custom_business_unit_ids"`
	BUDataAccessMode      *string  `json:"bu_data_access_mode"`
}

func (h *IdentityHandler) roleToJSON(row models.Role, filterForSubscription bool) gin.H {
	permissionIDs := h.rolePermissionIDs(row.ID)
	if filterForSubscription {
		permissionIDs = h.filterPermissionIDsForTenantSubscription(row.TenantID, permissionIDs)
	}
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "description": row.Description, "permission_ids": permissionIDs, "data_overrides": h.roleDataOverrides(row.ID)}
}

func (h *IdentityHandler) rolePermissionIDs(roleID uint64) []uint64 {
	var ids []uint64
	_ = h.db.Model(&models.RolePermission{}).Where("role_id = ?", roleID).Order("id asc").Pluck("permission_id", &ids).Error
	return ids
}

func (h *IdentityHandler) roleDataOverrides(roleID uint64) []gin.H {
	var links []models.RolePermission
	_ = h.db.
		Joins("JOIN permission p ON p.id = role_permission.permission_id").
		Where("role_permission.role_id = ? AND p.perm_type = ? AND p.deleted_at IS NULL", roleID, 4).
		Order("role_permission.id asc").
		Find(&links).Error
	items := make([]gin.H, 0, len(links))
	for _, link := range links {
		scope := derefString(link.DataScopeOverride)
		companyIDs := uint64IDsFromJSON(link.CustomCompanyIDsJSON)
		departmentIDs := uint64IDsFromJSON(link.CustomDepartmentIDsJSON)
		userIDs := uint64IDsFromJSON(link.CustomUserIDsJSON)
		businessUnitIDs := uint64IDsFromJSON(link.CustomBusinessUnitIDsJSON)
		if scope == "ALL" && len(companyIDs) == 0 && len(departmentIDs) == 0 && len(userIDs) == 0 && (len(businessUnitIDs) > 0 || link.BUDataAccessMode != nil) {
			scope = ""
		}
		items = append(items, gin.H{"permission_id": link.PermissionID, "data_scope": scope, "custom_company_ids": companyIDs, "custom_department_ids": departmentIDs, "custom_user_ids": userIDs, "custom_business_unit_ids": businessUnitIDs, "bu_data_access_mode": link.BUDataAccessMode})
	}
	return items
}

func uint64IDsFromJSON(raw *string) []uint64 {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return []uint64{}
	}
	var values []uint64
	if err := json.Unmarshal([]byte(*raw), &values); err == nil {
		return uniqueUint64s(values)
	}
	var stringsValues []string
	if err := json.Unmarshal([]byte(*raw), &stringsValues); err != nil {
		return []uint64{}
	}
	values = make([]uint64, 0, len(stringsValues))
	for _, value := range stringsValues {
		id, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			values = append(values, id)
		}
	}
	return uniqueUint64s(values)
}

func lockRoleForUpdate(tx *gorm.DB, roleID uint64) error {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&models.Role{}, roleID).Error
}

func lockPlanForUpdate(tx *gorm.DB, planID uint64) error {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&models.SaasPlan{}, planID).Error
}

func lockTenantForUpdate(tx *gorm.DB, tenantID uint64) error {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&models.Tenant{}, tenantID).Error
}

func (h *IdentityHandler) filterPermissionIDsForTenantSubscription(tenantID uint64, ids []uint64) []uint64 {
	if len(ids) == 0 {
		return []uint64{}
	}
	var rows []models.Permission
	_ = h.db.Where("tenant_id IN ? AND id IN ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(tenantID), ids).Find(&rows).Error
	allowed := map[uint64]struct{}{}
	for _, row := range rows {
		if h.permissionAllowedForTenantSubscription(tenantID, row) {
			allowed[row.ID] = struct{}{}
		}
	}
	out := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func (h *IdentityHandler) validateRolePermissionIDs(user models.AppUser, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	uniqueIDs := uniqueUint64s(ids)
	var rows []models.Permission
	if err := h.db.Where("tenant_id IN ? AND id IN ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(user.TenantID), uniqueIDs).Find(&rows).Error; err != nil {
		return err
	}
	byID := map[uint64]models.Permission{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	for _, id := range uniqueIDs {
		permission, ok := byID[id]
		if !ok {
			return errors.New("权限不存在或不属于当前主体")
		}
		if !h.viewerHasPlatformScope(user) && !h.permissionAllowedForTenantSubscription(user.TenantID, permission) {
			return fmt.Errorf("当前套餐不支持分配权限：%s", coalesceString(permission.Path, permission.Name))
		}
	}
	return nil
}

func (h *IdentityHandler) validateRoleDataOverrides(tenantID uint64, overrides []roleDataOverridePayload) error {
	if overrides == nil {
		return nil
	}
	ids := make([]uint64, 0, len(overrides))
	for _, override := range overrides {
		ids = append(ids, override.PermissionID)
	}
	var rows []models.Permission
	_ = h.db.Where("tenant_id IN ? AND id IN ? AND deleted_at IS NULL", h.permissionScopeTenantIDs(tenantID), uniqueUint64s(ids)).Find(&rows).Error
	modeByID := map[uint64]string{}
	for _, row := range rows {
		modeByID[row.ID] = coalesceString(row.DataPermMode, "ORG")
	}
	for _, override := range overrides {
		if !allowedString(override.DataScope, "ALL", "ORG", "ORG_SUB", "SELF", "CUSTOM") {
			return fmt.Errorf("无效的数据范围: %s", override.DataScope)
		}
		mode := modeByID[override.PermissionID]
		if mode == "" {
			mode = "ORG"
		}
		hasOrgFields := len(override.CustomCompanyIDs) > 0 || len(override.CustomDepartmentIDs) > 0 || len(override.CustomUserIDs) > 0
		hasBUIDs := len(override.CustomBusinessUnitIDs) > 0
		buMode := strings.TrimSpace(derefString(override.BUDataAccessMode))
		hasBUMode := buMode == "CURRENT_ORG_BU" || buMode == "SPECIFIED_BU"
		switch mode {
		case "NONE":
			if hasOrgFields || hasBUIDs || hasBUMode {
				return errors.New("该菜单不支持配置组织或业务单元数据权限")
			}
			if override.DataScope != "ALL" {
				return errors.New("该菜单不支持组织架构权限，数据范围必须为全部")
			}
		case "ORG":
			if hasBUIDs || hasBUMode {
				return errors.New("该菜单仅支持组织架构权限，不支持业务单元权限")
			}
		case "BU":
			if hasOrgFields {
				return errors.New("该菜单仅支持业务单元权限，不支持组织架构权限")
			}
			if override.DataScope != "ALL" {
				return errors.New("该菜单不支持组织架构权限，数据范围必须为全部")
			}
			if override.DataScope == "CUSTOM" && !hasBUIDs && !hasBUMode {
				return errors.New("业务单元自定义范围需至少指定业务单元或业务单元访问模式")
			}
		}
		if override.DataScope == "CUSTOM" && !hasOrgFields && !hasBUIDs && !hasBUMode {
			return errors.New("自定义范围需至少指定组织维度，或业务单元维度（指定 BU / CURRENT_ORG_BU）")
		}
	}
	return nil
}

func replaceRolePermissionsWithOverrides(tx *gorm.DB, roleID uint64, permissionIDs []uint64, overrides []roleDataOverridePayload) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}
	overrideByID := map[uint64]roleDataOverridePayload{}
	for _, override := range overrides {
		overrideByID[override.PermissionID] = override
	}
	for _, id := range uniqueUint64s(permissionIDs) {
		link := models.RolePermission{RoleID: roleID, PermissionID: id, Source: "MANUAL"}
		override, ok := overrideByID[id]
		if ok {
			link.DataScopeOverride = nullableFromString(override.DataScope)
			hasBU := len(override.CustomBusinessUnitIDs) > 0 || override.BUDataAccessMode != nil
			if override.DataScope == "CUSTOM" || hasBU {
				link.CustomCompanyIDsJSON = jsonStringPtr(uniqueUint64s(override.CustomCompanyIDs))
				link.CustomDepartmentIDsJSON = jsonStringPtr(uniqueUint64s(override.CustomDepartmentIDs))
				link.CustomUserIDsJSON = jsonStringPtr(uniqueUint64s(override.CustomUserIDs))
				link.CustomBusinessUnitIDsJSON = jsonStringPtr(uniqueUint64s(override.CustomBusinessUnitIDs))
				link.BUDataAccessMode = nullableTrimmed(override.BUDataAccessMode)
			}
		}
		if err := tx.Create(&link).Error; err != nil {
			return err
		}
	}
	return nil
}

func jsonStringPtr(value interface{}) *string {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	out := string(raw)
	return &out
}
