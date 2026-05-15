package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/application/dictionary"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) DictTypes(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var platformOnly *bool
	if h.viewerHasPlatformScope(user) || user.IsPlatformAdmin {
		if raw := strings.TrimSpace(c.Query("platform_only")); raw != "" {
			value := raw == "true" || raw == "1"
			platformOnly = &value
		}
	}
	result, err := h.dictionaryService().ListTypes(c.Request.Context(), h.dictionaryViewer(user), dictionary.ListTypesQuery{
		Keyword:      c.Query("keyword"),
		PlatformOnly: platformOnly,
		Skip:         skip,
		Limit:        limit,
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, result)
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
	row, err := h.dictionaryService().CreateType(c.Request.Context(), h.dictionaryViewer(user), dictionary.CreateTypeCommand{
		Code:           body.Code,
		Name:           body.Name,
		Remark:         body.Remark,
		Scope:          body.Scope,
		TenantEditable: body.TenantEditable,
		IsPlatformOnly: body.IsPlatformOnly,
	})
	if err != nil {
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
	id := parseUintParam(c, "id")
	row, err := h.dictionaryService().UpdateType(c.Request.Context(), h.dictionaryViewer(user), id, dictionary.UpdateTypeCommand{
		Name:           body.Name,
		Remark:         body.Remark,
		Scope:          body.Scope,
		TenantEditable: body.TenantEditable,
		IsPlatformOnly: body.IsPlatformOnly,
	})
	if errors.Is(err, dictionary.ErrNotFound) {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_type", "update", "编辑字典类型 "+row.Name, gin.H{"id": row.ID, "code": row.Code})
	response.OK(c, dictionary.TypeToResponse(row))
}

func (h *IdentityHandler) DeleteDictType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	row, err := h.dictionaryService().DeleteType(c.Request.Context(), h.dictionaryViewer(user), id)
	if errors.Is(err, dictionary.ErrNotFound) {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	if err != nil {
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
	if code == "" {
		code = c.Query("code")
	}
	code, items, err := h.dictionaryService().ItemsByCode(c.Request.Context(), h.dictionaryViewer(user), code)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"code": code, "items": items})
}

func (h *IdentityHandler) DictItems(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var dictTypeID uint64
	if rawDictTypeID := c.Query("dict_type_id"); rawDictTypeID != "" {
		dictTypeID = parseTenantIDValue(rawDictTypeID)
	}
	result, err := h.dictionaryService().ListItems(c.Request.Context(), h.dictionaryViewer(user), dictTypeID, skip, limit)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, result)
}

func (h *IdentityHandler) CreateDictItem(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		DictTypeID uint64  `json:"dict_type_id"`
		ParentID   *uint64 `json:"parent_id"`
		Label      string  `json:"label"`
		Value      string  `json:"value"`
		SortOrder  int     `json:"sort_order"`
		Enabled    *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	row, err := h.dictionaryService().CreateItem(c.Request.Context(), h.dictionaryViewer(user), dictionary.CreateItemCommand{
		DictTypeID: body.DictTypeID,
		ParentID:   body.ParentID,
		Label:      body.Label,
		Value:      body.Value,
		SortOrder:  body.SortOrder,
		Enabled:    body.Enabled,
	})
	if errors.Is(err, dictionary.ErrDictTypeNotFound) {
		respondBadRequest(c, err)
		return
	}
	if errors.Is(err, dictionary.ErrTenantOverrideOff) {
		response.Error(c, 403, response.CodeForbidden, "该字典不允许租户新增项")
		return
	}
	if err != nil {
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
		ParentID  *uint64 `json:"parent_id"`
		Label     *string `json:"label"`
		Value     *string `json:"value"`
		SortOrder *int    `json:"sort_order"`
		Enabled   *bool   `json:"enabled"`
	}
	hasParent, err := bindJSONWithField(c, &body, "parent_id")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	id := parseUintParam(c, "id")
	row, override, err := h.dictionaryService().UpdateItem(c.Request.Context(), h.dictionaryViewer(user), id, dictionary.UpdateItemCommand{
		ParentID:  body.ParentID,
		HasParent: hasParent,
		Label:     body.Label,
		Value:     body.Value,
		SortOrder: body.SortOrder,
		Enabled:   body.Enabled,
	})
	if errors.Is(err, dictionary.ErrNotFound) {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	if errors.Is(err, dictionary.ErrTenantOverrideOff) {
		response.Error(c, 403, response.CodeForbidden, "该字典不允许租户覆盖")
		return
	}
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_item", "update", "编辑字典项 "+row.Label, gin.H{"id": row.ID, "value": row.Value})
	response.OK(c, dictionary.ItemToResponse(row, override, user.TenantID))
}

func (h *IdentityHandler) DeleteDictItem(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	row, err := h.dictionaryService().DeleteItem(c.Request.Context(), h.dictionaryViewer(user), id)
	if errors.Is(err, dictionary.ErrNotFound) {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	if err != nil {
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
	id := parseUintParam(c, "id")
	row, err := h.dictionaryService().RestoreItem(c.Request.Context(), h.dictionaryViewer(user), id)
	if errors.Is(err, dictionary.ErrDictItemNotFound) {
		response.Error(c, 404, response.CodeNotFound, "字典项不存在")
		return
	}
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "dict_item", "restore", "恢复字典项默认值 "+row.Label, gin.H{"id": row.ID})
	response.OK(c, dictionary.ItemToResponse(row, nil, user.TenantID))
}

func (h *IdentityHandler) dictionaryService() *dictionary.Service {
	return dictionary.NewService(repositories.NewDictionaryRepository(h.db))
}

func (h *IdentityHandler) dictionaryViewer(user models.AppUser) dictionary.Viewer {
	includePlatformOnly := user.IsPlatformAdmin || h.viewerHasPlatformScope(user)
	return dictionary.Viewer{
		TenantID:            user.TenantID,
		IsPlatformAdmin:     user.IsPlatformAdmin,
		IncludePlatformOnly: includePlatformOnly,
		ScopeTenantIDs:      h.permissionScopeTenantIDs(user.TenantID),
	}
}

func bindJSONWithField(c *gin.Context, out interface{}, field string) (bool, error) {
	raw, err := c.GetRawData()
	if err != nil {
		return false, err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
	if err := json.Unmarshal(raw, out); err != nil {
		return false, err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return false, err
	}
	_, exists := body[field]
	return exists, nil
}
