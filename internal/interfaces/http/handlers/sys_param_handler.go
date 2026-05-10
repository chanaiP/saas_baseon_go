package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/dto"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) SysParams(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.SystemParam
	forPlatform := user.IsPlatformAdmin || h.viewerHasPlatformScope(user)
	query := h.visibleSystemParamsQuery(user.TenantID, forPlatform)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("sys_param.param_key LIKE ?", like)
	}
	var total int64
	_ = query.Count(&total).Error
	_ = query.Clauses(clause.OrderBy{
		Expression: clause.Expr{
			SQL:  "CASE WHEN sys_param.tenant_id = ? THEN 0 ELSE 1 END, sys_param.id ASC",
			Vars: []interface{}{user.TenantID},
		},
	}).Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]dto.SystemParamResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.sysParamToResponse(user.TenantID, row))
	}
	response.OK(c, dto.PaginatedResponse[dto.SystemParamResponse]{Items: items, Total: total, Skip: skip, Limit: limit})
}

func (h *IdentityHandler) CreateSysParam(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Key            string `json:"param_key"`
		Value          string `json:"param_value"`
		DefaultValue   string `json:"default_value"`
		Remark         string `json:"remark"`
		ValueType      string `json:"value_type"`
		TenantEditable bool   `json:"tenant_editable"`
		IsPlatformOnly bool   `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	value := body.Value
	if value == "" {
		value = body.DefaultValue
	}
	if body.ValueType == "" {
		body.ValueType = "string"
	}
	row := models.SystemParam{TenantID: user.TenantID, Key: strings.TrimSpace(body.Key), Value: value, Remark: body.Remark, ValueType: body.ValueType, TenantEditable: body.TenantEditable, IsPlatformOnly: body.IsPlatformOnly}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, safeDBErrorMessage(err))
		return
	}
	h.audit(c, user.TenantID, user.ID, "sys_param", "create", "创建系统参数 "+row.Key, gin.H{"id": row.ID, "param_key": row.Key})
	response.OK(c, dto.IDResponse{ID: row.ID})
}

func (h *IdentityHandler) UpdateSysParam(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		ParamValue     *string `json:"param_value"`
		DefaultValue   *string `json:"default_value"`
		Remark         *string `json:"remark"`
		ValueType      *string `json:"value_type"`
		TenantEditable *bool   `json:"tenant_editable"`
		IsPlatformOnly *bool   `json:"is_platform_only"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var row models.SystemParam
	forPlatform := user.IsPlatformAdmin || h.viewerHasPlatformScope(user)
	if found, err := h.visibleSystemParamByID(user.TenantID, c.Param("id"), forPlatform); err != nil {
		response.Error(c, 404, response.CodeNotFound, "参数不存在")
		return
	} else {
		row = found
	}
	value := body.ParamValue
	if value == nil {
		value = body.DefaultValue
	}
	if !user.IsPlatformAdmin {
		if !row.TenantEditable || row.IsPlatformOnly {
			response.Error(c, 403, response.CodeForbidden, "该参数不允许租户覆盖")
			return
		}
		h.upsertTenantParamValue(user.TenantID, row.ID, value)
		h.audit(c, user.TenantID, user.ID, "sys_param", "update", "覆盖系统参数 "+row.Key, gin.H{"id": row.ID, "param_key": row.Key})
		response.OK(c, h.sysParamToResponse(user.TenantID, row))
		return
	}
	updates := map[string]interface{}{}
	if value != nil {
		updates["param_value"] = *value
	}
	if body.Remark != nil {
		updates["remark"] = *body.Remark
	}
	if body.ValueType != nil {
		updates["value_type"] = *body.ValueType
	}
	if body.TenantEditable != nil {
		updates["tenant_editable"] = *body.TenantEditable
	}
	if body.IsPlatformOnly != nil {
		updates["is_platform_only"] = *body.IsPlatformOnly
	}
	if err := h.db.Model(&row).Updates(updates).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	_ = h.db.First(&row, row.ID).Error
	h.audit(c, user.TenantID, user.ID, "sys_param", "update", "编辑系统参数 "+row.Key, gin.H{"id": row.ID, "param_key": row.Key})
	response.OK(c, h.sysParamToResponse(user.TenantID, row))
}

func (h *IdentityHandler) DeleteSysParam(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	var row models.SystemParam
	if err := h.tenantScope().ActiveByID(user.TenantID, id).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "参数不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "param_key": tombstoneUniqueValue(row.Key, row.ID, 128)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "sys_param", "delete", "删除系统参数 "+row.Key, gin.H{"id": id})
	response.OK(c, dto.DeletedResponse{Deleted: id})
}

func (h *IdentityHandler) RestoreSysParam(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var row models.SystemParam
	forPlatform := user.IsPlatformAdmin || h.viewerHasPlatformScope(user)
	if found, err := h.visibleSystemParamByID(user.TenantID, c.Param("id"), forPlatform); err != nil {
		response.Error(c, 404, response.CodeNotFound, "参数不存在")
		return
	} else {
		row = found
	}
	_ = h.db.Where("tenant_id = ? AND param_id = ?", user.TenantID, row.ID).Delete(&models.TenantParamValue{}).Error
	h.audit(c, user.TenantID, user.ID, "sys_param", "restore", "恢复系统参数默认值 "+row.Key, gin.H{"id": row.ID})
	response.OK(c, h.sysParamToResponse(user.TenantID, row))
}

func (h *IdentityHandler) SysParamBatch(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	values := map[string]*string{}
	var rows []models.SystemParam
	keys := splitCSVParam(c.Query("keys"))
	forPlatform := user.IsPlatformAdmin || h.viewerHasPlatformScope(user)
	query := h.visibleSystemParamsQuery(user.TenantID, forPlatform)
	if len(keys) > 0 {
		query = query.Where("sys_param.param_key IN ?", keys)
	}
	_ = query.Clauses(clause.OrderBy{
		Expression: clause.Expr{
			SQL:  "CASE WHEN sys_param.tenant_id = ? THEN 0 ELSE 1 END, sys_param.id ASC",
			Vars: []interface{}{user.TenantID},
		},
	}).Find(&rows).Error
	for _, row := range rows {
		if _, exists := values[row.Key]; exists {
			continue
		}
		item := h.sysParamToResponse(user.TenantID, row)
		value := item.ParamValue
		values[row.Key] = &value
	}
	if len(keys) > 0 {
		for _, key := range keys {
			if _, ok := values[key]; !ok {
				values[key] = nil
			}
		}
	}
	response.OK(c, dto.SystemParamBatchResponse{Values: values})
}
