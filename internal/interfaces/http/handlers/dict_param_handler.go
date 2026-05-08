package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) DictTypes(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.DictType
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", user.TenantID)
	if !user.IsPlatformAdmin {
		query = query.Where("is_platform_only = ?", false)
	} else if raw := strings.TrimSpace(c.Query("platform_only")); raw != "" {
		query = query.Where("is_platform_only = ?", raw == "true" || raw == "1")
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	var total int64
	_ = query.Model(&models.DictType{}).Count(&total).Error
	_ = query.Order("id asc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, dictTypeToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) CreateDictType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Code           string  `json:"code"`
		Name           string  `json:"name"`
		Remark         *string `json:"remark"`
		Scope          string  `json:"scope"`
		TenantEditable bool    `json:"tenant_editable"`
		IsPlatformOnly bool    `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Scope == "" {
		body.Scope = "platform"
	}
	row := models.DictType{TenantID: user.TenantID, Code: strings.TrimSpace(body.Code), Name: strings.TrimSpace(body.Name), Remark: nullableTrimmed(body.Remark), Scope: body.Scope, TenantEditable: body.TenantEditable, IsPlatformOnly: body.IsPlatformOnly}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_type", "create", "创建字典类型 "+row.Name, gin.H{"id": row.ID, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateDictType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Name           *string `json:"name"`
		Remark         *string `json:"remark"`
		Scope          *string `json:"scope"`
		TenantEditable *bool   `json:"tenant_editable"`
		IsPlatformOnly *bool   `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Remark != nil {
		updates["remark"] = nullableTrimmed(body.Remark)
	}
	if body.Scope != nil {
		updates["scope"] = strings.TrimSpace(*body.Scope)
	}
	if body.TenantEditable != nil {
		updates["tenant_editable"] = *body.TenantEditable
	}
	if body.IsPlatformOnly != nil {
		updates["is_platform_only"] = *body.IsPlatformOnly
	}
	var row models.DictType
	if err := h.tenantScope().ActiveByID(user.TenantID, c.Param("id")).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	if err := h.db.Model(&row).Updates(updates).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	_ = h.db.First(&row, row.ID).Error
	h.audit(c, user.TenantID, user.ID, "dict_type", "update", "编辑字典类型 "+row.Name, gin.H{"id": row.ID, "code": row.Code})
	response.OK(c, dictTypeToJSON(row))
}

func (h *IdentityHandler) DeleteDictType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "字典类型", ref(&models.DictItem{}, "字典项", "tenant_id = ? AND dict_type_id = ? AND deleted_at IS NULL", user.TenantID, id)) {
		return
	}
	var row models.DictType
	if err := h.tenantScope().ActiveByID(user.TenantID, id).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "code": tombstoneUniqueValue(row.Code, row.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_type", "delete", "删除字典类型 "+row.Name, gin.H{"id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) DictItemsByCode(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	code := c.Param("code")
	var dictType models.DictType
	if err := h.tenantScope().ActiveByCode(user.TenantID, code).First(&dictType).Error; err != nil {
		response.OK(c, gin.H{"code": code, "items": []gin.H{}})
		return
	}
	var rows []models.DictItem
	_ = h.db.Where("tenant_id = ? AND dict_type_id = ? AND deleted_at IS NULL", user.TenantID, dictType.ID).Order("sort_order asc, id asc").Find(&rows).Error
	items := h.dictItemsToJSON(user.TenantID, rows, true)
	response.OK(c, gin.H{"code": code, "items": items})
}

func (h *IdentityHandler) DictItems(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.DictItem
	query := h.tenantScope().Active(user.TenantID)
	if dictTypeID := c.Query("dict_type_id"); dictTypeID != "" {
		query = query.Where("dict_type_id = ?", dictTypeID)
	}
	var total int64
	_ = query.Model(&models.DictItem{}).Count(&total).Error
	_ = query.Order("sort_order asc, id asc").Offset(skip).Limit(limit).Find(&rows).Error
	response.OK(c, paginatedWithTotal(h.dictItemsToJSON(user.TenantID, rows, false), total, skip, limit))
}

func (h *IdentityHandler) CreateDictItem(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		DictTypeID uint64 `json:"dict_type_id"`
		Label      string `json:"label"`
		Value      string `json:"value"`
		SortOrder  int    `json:"sort_order"`
		Enabled    *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	if err := h.assertDictTypeInTenant(user.TenantID, body.DictTypeID); err != nil {
		respondBadRequest(c, err)
		return
	}
	row := models.DictItem{TenantID: user.TenantID, DictTypeID: body.DictTypeID, Label: strings.TrimSpace(body.Label), Value: strings.TrimSpace(body.Value), SortOrder: body.SortOrder, Enabled: enabled}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_item", "create", "创建字典项 "+row.Label, gin.H{"id": row.ID, "value": row.Value})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdateDictItem(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Label     *string `json:"label"`
		Value     *string `json:"value"`
		SortOrder *int    `json:"sort_order"`
		Enabled   *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var row models.DictItem
	if err := h.tenantScope().ActiveByID(user.TenantID, c.Param("id")).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	if !user.IsPlatformAdmin {
		var dictType models.DictType
		if err := h.tenantScope().ActiveByID(user.TenantID, row.DictTypeID).First(&dictType).Error; err != nil || !dictType.TenantEditable {
			response.Error(c, 403, response.CodeForbidden, "该字典不允许租户覆盖")
			return
		}
		override := h.upsertTenantDictItemOverride(user.TenantID, row.ID, body.Label, body.Value, body.SortOrder, body.Enabled)
		h.audit(c, user.TenantID, user.ID, "dict_item", "update", "编辑字典项 "+row.Label, gin.H{"id": row.ID, "value": row.Value})
		response.OK(c, h.dictItemToJSON(row, override))
		return
	}
	updates := dictItemUpdates(body.Label, body.Value, body.SortOrder, body.Enabled)
	if err := h.db.Model(&row).Updates(updates).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	_ = h.db.First(&row, row.ID).Error
	h.audit(c, user.TenantID, user.ID, "dict_item", "update", "编辑字典项 "+row.Label, gin.H{"id": row.ID, "value": row.Value})
	response.OK(c, h.dictItemToJSON(row, nil))
}

func (h *IdentityHandler) DeleteDictItem(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	var row models.DictItem
	if err := h.tenantScope().ActiveByID(user.TenantID, id).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "enabled": false}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_item", "delete", "删除字典项 "+row.Label, gin.H{"id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) RestoreDictItem(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var row models.DictItem
	if err := h.tenantScope().ActiveByID(user.TenantID, c.Param("id")).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "字典项不存在")
		return
	}
	_ = h.db.Where("tenant_id = ? AND dict_item_id = ?", user.TenantID, row.ID).Delete(&models.TenantDictItemOverride{}).Error
	h.audit(c, user.TenantID, user.ID, "dict_item", "restore", "恢复字典项默认值 "+row.Label, gin.H{"id": row.ID})
	response.OK(c, h.dictItemToJSON(row, nil))
}
