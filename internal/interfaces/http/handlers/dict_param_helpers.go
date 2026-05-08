package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
)

func dictItemsToJSON(rows []models.DictItem) []gin.H {
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, dictItemJSON(row, nil))
	}
	return items
}

func dictTypeToJSON(row models.DictType) gin.H {
	return gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "type_code": row.Code, "type_name": row.Name, "remark": row.Remark, "scope": row.Scope, "tenant_editable": row.TenantEditable, "is_platform_only": row.IsPlatformOnly, "status": 1}
}

func dictItemJSON(row models.DictItem, override *models.TenantDictItemOverride) gin.H {
	label := row.Label
	value := row.Value
	sortOrder := row.SortOrder
	enabled := row.Enabled
	isOverride := override != nil
	if override != nil {
		if override.CustomLabel != nil {
			label = *override.CustomLabel
		}
		if override.CustomValue != nil {
			value = *override.CustomValue
		}
		if override.SortOrder != nil {
			sortOrder = *override.SortOrder
		}
		if override.Enabled != nil {
			enabled = *override.Enabled
		}
	}
	return gin.H{"id": row.ID, "dict_type_id": row.DictTypeID, "default_label": row.Label, "default_value": row.Value, "default_sort_order": row.SortOrder, "default_enabled": row.Enabled, "label": label, "value": value, "item_label": label, "item_value": value, "sort_order": sortOrder, "enabled": enabled, "status": boolToStatus(enabled), "is_override": isOverride}
}

func (h *IdentityHandler) dictItemToJSON(row models.DictItem, override *models.TenantDictItemOverride) gin.H {
	return dictItemJSON(row, override)
}

func (h *IdentityHandler) dictItemsToJSON(tenantID uint64, rows []models.DictItem, enabledOnly bool) []gin.H {
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var override models.TenantDictItemOverride
		var overridePtr *models.TenantDictItemOverride
		if err := h.db.Where("tenant_id = ? AND dict_item_id = ?", tenantID, row.ID).First(&override).Error; err == nil {
			overridePtr = &override
		}
		item := dictItemJSON(row, overridePtr)
		if enabledOnly {
			enabled, _ := item["enabled"].(bool)
			if !enabled {
				continue
			}
		}
		items = append(items, item)
	}
	return items
}

func dictItemUpdates(label *string, value *string, sortOrder *int, enabled *bool) map[string]interface{} {
	updates := map[string]interface{}{}
	if label != nil {
		updates["label"] = strings.TrimSpace(*label)
	}
	if value != nil {
		updates["value"] = strings.TrimSpace(*value)
	}
	if sortOrder != nil {
		updates["sort_order"] = *sortOrder
	}
	if enabled != nil {
		updates["enabled"] = *enabled
	}
	return updates
}

func (h *IdentityHandler) upsertTenantDictItemOverride(tenantID uint64, itemID uint64, label *string, value *string, sortOrder *int, enabled *bool) *models.TenantDictItemOverride {
	var row models.TenantDictItemOverride
	err := h.db.Where("tenant_id = ? AND dict_item_id = ?", tenantID, itemID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.TenantDictItemOverride{TenantID: tenantID, DictItemID: itemID}
	}
	row.CustomLabel = nullableTrimmed(label)
	row.CustomValue = nullableTrimmed(value)
	row.SortOrder = sortOrder
	row.Enabled = enabled
	if row.ID == 0 {
		_ = h.db.Create(&row).Error
	} else {
		_ = h.db.Save(&row).Error
	}
	return &row
}

func (h *IdentityHandler) assertDictTypeInTenant(tenantID uint64, dictTypeID uint64) error {
	var count int64
	if err := h.db.Model(&models.DictType{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", dictTypeID, tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("字典类型不存在")
	}
	return nil
}

func (h *IdentityHandler) sysParamToResponse(tenantID uint64, row models.SystemParam) dto.SystemParamResponse {
	var override models.TenantParamValue
	paramValue := row.Value
	isOverride := false
	if !row.IsPlatformOnly {
		if err := h.db.Where("tenant_id = ? AND param_id = ?", tenantID, row.ID).First(&override).Error; err == nil {
			isOverride = true
			if override.ParamValue != nil {
				paramValue = *override.ParamValue
			} else {
				paramValue = ""
			}
		}
	}
	return dto.SystemParamResponse{ID: row.ID, Key: row.Key, DefaultValue: row.Value, ParamValue: paramValue, Remark: row.Remark, ValueType: row.ValueType, TenantEditable: row.TenantEditable, IsPlatformOnly: row.IsPlatformOnly, IsOverride: isOverride}
}

func (h *IdentityHandler) upsertTenantParamValue(tenantID uint64, paramID uint64, value *string) *models.TenantParamValue {
	var row models.TenantParamValue
	err := h.db.Where("tenant_id = ? AND param_id = ?", tenantID, paramID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.TenantParamValue{TenantID: tenantID, ParamID: paramID}
	}
	row.ParamValue = value
	if row.ID == 0 {
		_ = h.db.Create(&row).Error
	} else {
		_ = h.db.Save(&row).Error
	}
	return &row
}
