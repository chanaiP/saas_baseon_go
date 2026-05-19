package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) PositionTypes(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.PositionType
	query := h.tenantScope().Active(user.TenantID)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	var total int64
	_ = query.Model(&models.PositionType{}).Count(&total).Error
	_ = query.Order("id asc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var count int64
		_ = h.db.Model(&models.Position{}).Where("tenant_id = ? AND position_type_id = ? AND deleted_at IS NULL", user.TenantID, row.ID).Count(&count).Error
		items = append(items, gin.H{"id": row.ID, "code": row.Code, "name": row.Name, "position_count": count})
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) CreatePositionType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	row := models.PositionType{TenantID: user.TenantID, Name: strings.TrimSpace(body.Name), Code: strings.TrimSpace(body.Code)}
	if msg := validatePositionTypePayload(h, user.TenantID, 0, row.Name, row.Code); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "position_type", "create", "创建岗位类型 "+row.Name, gin.H{"id": row.ID, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdatePositionType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		Name *string `json:"name"`
		Code *string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Code != nil {
		updates["code"] = strings.TrimSpace(*body.Code)
	}
	if msg := validatePositionTypeUpdatePayload(h, user.TenantID, parseUintParam(c, "id"), body.Name, body.Code); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	result := h.db.Model(&models.PositionType{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", c.Param("id"), user.TenantID).Updates(updates)
	if result.Error != nil {
		respondBadRequest(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	h.audit(c, user.TenantID, user.ID, "position_type", "update", "编辑岗位类型", gin.H{"id": parseUintParam(c, "id"), "changes": updates})
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeletePositionType(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "岗位类型", ref(&models.Position{}, "岗位", "tenant_id = ? AND position_type_id = ? AND deleted_at IS NULL", user.TenantID, id)) {
		return
	}
	var row models.PositionType
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, user.TenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "code": tombstoneUniqueValue(row.Code, row.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "position_type", "delete", "删除岗位类型 "+row.Name, gin.H{"id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) Positions(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.Position
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", user.TenantID)
	if positionTypeID := c.Query("position_type_id"); positionTypeID != "" {
		query = query.Where("position_type_id = ?", positionTypeID)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	var total int64
	_ = query.Model(&models.Position{}).Count(&total).Error
	_ = query.Order("id asc").Offset(skip).Limit(limit).Find(&rows).Error
	typeIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		typeIDs = append(typeIDs, row.PositionTypeID)
	}
	typeByID := map[uint64]models.PositionType{}
	if len(typeIDs) > 0 {
		var positionTypes []models.PositionType
		_ = h.db.Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", user.TenantID, uniqueUint64s(typeIDs)).Find(&positionTypes).Error
		for _, row := range positionTypes {
			typeByID[row.ID] = row
		}
	}
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		item := gin.H{"id": row.ID, "position_type_id": row.PositionTypeID, "name": row.Name, "code": row.Code}
		if positionType, ok := typeByID[row.PositionTypeID]; ok {
			item["position_type_name"] = positionType.Name
			item["position_type_code"] = positionType.Code
		}
		items = append(items, item)
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) CreatePosition(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		PositionTypeID uint64 `json:"position_type_id"`
		Name           string `json:"name"`
		Code           string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.assertPositionTypeInTenant(user.TenantID, body.PositionTypeID); err != nil {
		respondBadRequest(c, err)
		return
	}
	row := models.Position{TenantID: user.TenantID, PositionTypeID: body.PositionTypeID, Name: strings.TrimSpace(body.Name), Code: strings.TrimSpace(body.Code)}
	if msg := validatePositionPayload(h, user.TenantID, 0, row.Name, row.Code); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "position", "create", "创建岗位 "+row.Name, gin.H{"id": row.ID, "code": row.Code})
	response.OK(c, gin.H{"id": row.ID})
}

func (h *IdentityHandler) UpdatePosition(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body struct {
		PositionTypeID *uint64 `json:"position_type_id"`
		Name           *string `json:"name"`
		Code           *string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.PositionTypeID != nil {
		if err := h.assertPositionTypeInTenant(user.TenantID, *body.PositionTypeID); err != nil {
			respondBadRequest(c, err)
			return
		}
		updates["position_type_id"] = *body.PositionTypeID
	}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Code != nil {
		updates["code"] = strings.TrimSpace(*body.Code)
	}
	if msg := validatePositionUpdatePayload(h, user.TenantID, parseUintParam(c, "id"), body.Name, body.Code); msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	result := h.db.Model(&models.Position{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", c.Param("id"), user.TenantID).Updates(updates)
	if result.Error != nil {
		respondBadRequest(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	h.audit(c, user.TenantID, user.ID, "position", "update", "编辑岗位", gin.H{"id": parseUintParam(c, "id"), "changes": updates})
	response.OK(c, gin.H{"id": parseUintParam(c, "id")})
}

func (h *IdentityHandler) DeletePosition(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(c, "岗位", ref(&models.AppUserPosition{}, "用户岗位", "position_id = ?", id)) {
		return
	}
	var row models.Position
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, user.TenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "code": tombstoneUniqueValue(row.Code, row.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "position", "delete", "删除岗位 "+row.Name, gin.H{"id": id})
	response.OK(c, gin.H{"deleted": id})
}

func validatePositionTypePayload(h *IdentityHandler, tenantID uint64, exceptID uint64, name string, code string) string {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" || code == "" {
		return "岗位类型名称和编码不能为空"
	}
	if positionTypeFieldExists(h, tenantID, exceptID, "name", name) {
		return "岗位类型名称已存在"
	}
	if positionTypeFieldExists(h, tenantID, exceptID, "code", code) {
		return "岗位类型编码已存在"
	}
	return ""
}

func positionTypeFieldExists(h *IdentityHandler, tenantID uint64, exceptID uint64, field string, value string) bool {
	var count int64
	query := h.db.Model(&models.PositionType{}).Where("tenant_id = ? AND "+field+" = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(value))
	if exceptID > 0 {
		query = query.Where("id <> ?", exceptID)
	}
	_ = query.Count(&count).Error
	return count > 0
}

func validatePositionTypeUpdatePayload(h *IdentityHandler, tenantID uint64, id uint64, name *string, code *string) string {
	if name != nil && strings.TrimSpace(*name) == "" {
		return "岗位类型名称不能为空"
	}
	if name != nil && positionTypeFieldExists(h, tenantID, id, "name", *name) {
		return "岗位类型名称已存在"
	}
	if code != nil && strings.TrimSpace(*code) == "" {
		return "岗位类型编码不能为空"
	}
	if code != nil && positionTypeFieldExists(h, tenantID, id, "code", *code) {
		return "岗位类型编码已存在"
	}
	return ""
}

func validatePositionPayload(h *IdentityHandler, tenantID uint64, exceptID uint64, name string, code string) string {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" || code == "" {
		return "岗位名称和编码不能为空"
	}
	if positionFieldExists(h, tenantID, exceptID, "name", name) {
		return "岗位名称已存在"
	}
	if positionFieldExists(h, tenantID, exceptID, "code", code) {
		return "岗位编码已存在"
	}
	return ""
}

func positionFieldExists(h *IdentityHandler, tenantID uint64, exceptID uint64, field string, value string) bool {
	var count int64
	query := h.db.Model(&models.Position{}).Where("tenant_id = ? AND "+field+" = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(value))
	if exceptID > 0 {
		query = query.Where("id <> ?", exceptID)
	}
	_ = query.Count(&count).Error
	return count > 0
}

func validatePositionUpdatePayload(h *IdentityHandler, tenantID uint64, id uint64, name *string, code *string) string {
	if name != nil && strings.TrimSpace(*name) == "" {
		return "岗位名称不能为空"
	}
	if name != nil && positionFieldExists(h, tenantID, id, "name", *name) {
		return "岗位名称已存在"
	}
	if code != nil && strings.TrimSpace(*code) == "" {
		return "岗位编码不能为空"
	}
	if code != nil && positionFieldExists(h, tenantID, id, "code", *code) {
		return "岗位编码已存在"
	}
	return ""
}
