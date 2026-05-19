package handlers

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) BusinessResources(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	skip, limit := paginationParams(c)
	query := h.db.Model(&models.BusinessResource{}).Where("tenant_id = ? AND deleted_at IS NULL", user.TenantID)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("resource_name LIKE ? OR resource_code LIKE ? OR external_id LIKE ?", like, like, like)
	}
	if unitType := strings.TrimSpace(c.Query("unit_type_code")); unitType != "" {
		query = query.Where("unit_type_code = ? OR resource_category = ?", unitType, unitType)
	}
	if unitCode := strings.TrimSpace(c.Query("business_unit_code")); unitCode != "" {
		query = query.Where("business_unit_code = ? OR resource_type = ?", unitCode, unitCode)
	}
	if category := strings.TrimSpace(c.Query("resource_category")); category != "" {
		query = query.Where("resource_category = ?", category)
	}
	if resourceType := strings.TrimSpace(c.Query("resource_type")); resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if sourceMode := strings.TrimSpace(c.Query("source_mode")); sourceMode != "" {
		query = query.Where("source_mode = ?", sourceMode)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	_ = query.Model(&models.BusinessResource{}).Count(&total).Error
	var rows []models.BusinessResource
	_ = query.Order("id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessResourceToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) BusinessResourceDictionaryTree(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	code, items, err := h.dictionaryService().ItemsByCode(c.Request.Context(), h.dictionaryViewer(user), businessUnitDictCode)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	roots := make([]gin.H, 0)
	childrenByParent := map[uint64][]gin.H{}
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		node := gin.H{"id": item.ID, "code": item.Value, "name": item.Label, "sort_order": item.SortOrder, "children": []gin.H{}}
		if item.ParentID == nil {
			roots = append(roots, node)
			continue
		}
		childrenByParent[*item.ParentID] = append(childrenByParent[*item.ParentID], node)
	}
	for _, root := range roots {
		if id, ok := root["id"].(uint64); ok {
			root["children"] = childrenByParent[id]
		}
	}
	response.OK(c, gin.H{"dict_code": code, "dict_name": "业务单元类型", "children": roots})
}

func (h *IdentityHandler) BusinessResourceSummary(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	ordered, err := h.businessResourceSummaryItems(c, user)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"items": ordered})
}

func (h *IdentityHandler) businessResourceSummaryItems(c *gin.Context, user models.AppUser) ([]gin.H, error) {
	type summaryRow struct {
		UnitTypeCode     string
		UnitTypeName     string
		BusinessUnitCode string
		BusinessUnitName string
		ResourceCount    int64
	}
	var rows []summaryRow
	err := h.db.Model(&models.BusinessResource{}).
		Select("unit_type_code, unit_type_name, business_unit_code, business_unit_name, COUNT(*) AS resource_count").
		Where("tenant_id = ? AND deleted_at IS NULL AND status = ?", user.TenantID, "active").
		Group("unit_type_code, unit_type_name, business_unit_code, business_unit_name").
		Order("unit_type_code asc, business_unit_code asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	dictNames := h.businessUnitDictNameMap(c, user)
	typeNode := map[string]gin.H{}
	ordered := make([]gin.H, 0)
	for _, row := range rows {
		if row.UnitTypeCode == "" || row.BusinessUnitCode == "" {
			continue
		}
		unitTypeName := row.UnitTypeName
		if name := dictNames[row.UnitTypeCode]; name != "" {
			unitTypeName = name
		}
		businessUnitName := row.BusinessUnitName
		if name := dictNames[row.BusinessUnitCode]; name != "" {
			businessUnitName = name
		}
		node, ok := typeNode[row.UnitTypeCode]
		if !ok {
			node = gin.H{"code": row.UnitTypeCode, "name": unitTypeName, "resource_count": int64(0), "business_units": []gin.H{}}
			typeNode[row.UnitTypeCode] = node
			ordered = append(ordered, node)
		}
		node["resource_count"] = node["resource_count"].(int64) + row.ResourceCount
		node["business_units"] = append(node["business_units"].([]gin.H), gin.H{"code": row.BusinessUnitCode, "name": businessUnitName, "resource_count": row.ResourceCount})
	}
	return ordered, nil
}

func (h *IdentityHandler) CreateBusinessResource(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body businessResourceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	row, err := h.businessResourceFromBody(c, user, body, nil)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource", "create", "创建业务资源 "+row.ResourceName, gin.H{"resource_id": row.ID, "resource_code": row.ResourceCode})
	response.OK(c, businessResourceToJSON(row))
}

func (h *IdentityHandler) UpdateBusinessResource(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var row models.BusinessResource
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", c.Param("id"), user.TenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	var body businessResourceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	next, err := h.businessResourceFromBody(c, user, body, &row)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	if err := h.db.Model(&row).Updates(map[string]interface{}{
		"unit_type_code":         next.UnitTypeCode,
		"unit_type_name":         next.UnitTypeName,
		"business_unit_code":     next.BusinessUnitCode,
		"business_unit_name":     next.BusinessUnitName,
		"resource_name":          next.ResourceName,
		"resource_code":          next.ResourceCode,
		"resource_category":      next.ResourceCategory,
		"resource_type":          next.ResourceType,
		"source_mode":            next.SourceMode,
		"source_app_code":        next.SourceAppCode,
		"source_table":           next.SourceTable,
		"source_id":              next.SourceID,
		"platform_code":          next.PlatformCode,
		"external_id":            next.ExternalID,
		"connection_instance_id": next.ConnectionInstanceID,
		"parent_resource_id":     next.ParentResourceID,
		"resource_attrs":         next.ResourceAttrs,
		"status":                 next.Status,
	}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	_ = h.db.First(&row, row.ID).Error
	h.audit(c, user.TenantID, user.ID, "business_resource", "update", "更新业务资源 "+row.ResourceName, gin.H{"resource_id": row.ID, "resource_code": row.ResourceCode})
	response.OK(c, businessResourceToJSON(row))
}

func (h *IdentityHandler) DeleteBusinessResource(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if h.blockDeleteIfReferenced(
		c,
		"业务资源",
		ref(&models.BusinessUnitResource{}, "业务单元资源绑定", "tenant_id = ? AND resource_id = ? AND deleted_at IS NULL", user.TenantID, id),
		ref(&models.BusinessResourceRelation{}, "业务资源关系", "tenant_id = ? AND (parent_resource_id = ? OR child_resource_id = ?) AND deleted_at IS NULL", user.TenantID, id, id),
	) {
		return
	}
	var row models.BusinessResource
	if err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, user.TenantID).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&row).Updates(map[string]interface{}{"deleted_at": now, "status": "archived", "resource_code": tombstoneUniqueValue(row.ResourceCode, row.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource", "delete", "归档业务资源 "+row.ResourceName, gin.H{"resource_id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) BusinessUnitResources(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	buID := parseUintParam(c, "id")
	if err := h.assertBusinessUnit(user.TenantID, buID); err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	var rows []models.BusinessUnitResource
	_ = h.db.Where("tenant_id = ? AND business_unit_id = ? AND deleted_at IS NULL", user.TenantID, buID).Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessUnitResourceToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessUnitResource(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	buID := parseUintParam(c, "id")
	if err := h.assertBusinessUnit(user.TenantID, buID); err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
		return
	}
	var body businessUnitResourceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	resource, err := h.businessResourceByID(user.TenantID, body.ResourceID)
	if err != nil {
		respondBadRequest(c, errors.New("业务资源不存在"))
		return
	}
	relationType := strings.TrimSpace(body.RelationType)
	if relationType == "" {
		relationType = "owner"
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.unit_resource_relation_type", &relationType, "业务单元资源关系类型"); err != nil {
		respondBadRequest(c, err)
		return
	}
	row := models.BusinessUnitResource{
		TenantID:         user.TenantID,
		BusinessUnitID:   buID,
		ResourceID:       resource.ID,
		ResourceCategory: resource.ResourceCategory,
		ResourceType:     resource.ResourceType,
		RelationType:     relationType,
		IsPrimary:        boolValue(body.IsPrimary, false),
		UseForPermission: boolValue(body.UseForPermission, false),
		UseForOperation:  boolValue(body.UseForOperation, false),
		UseForSettlement: boolValue(body.UseForSettlement, false),
		StartDate:        parseDatePtr(body.StartDate),
		EndDate:          parseDatePtr(body.EndDate),
	}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_resource", "create", "绑定业务单元资源", gin.H{"business_unit_id": buID, "resource_id": resource.ID})
	response.OK(c, businessUnitResourceToJSON(row))
}

func (h *IdentityHandler) DeleteBusinessUnitResource(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	buID := parseUintParam(c, "id")
	relationID := parseUintParam(c, "relationId")
	now := time.Now()
	result := h.db.Model(&models.BusinessUnitResource{}).Where("id = ? AND tenant_id = ? AND business_unit_id = ? AND deleted_at IS NULL", relationID, user.TenantID, buID).Update("deleted_at", now)
	if result.Error != nil {
		respondBadRequest(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "绑定关系不存在")
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_resource", "delete", "解除业务单元资源绑定", gin.H{"business_unit_id": buID, "relation_id": relationID})
	response.OK(c, gin.H{"deleted": relationID})
}

func (h *IdentityHandler) BusinessResourceRelations(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	resourceID := parseUintParam(c, "id")
	if _, err := h.businessResourceByID(user.TenantID, resourceID); err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	var rows []models.BusinessResourceRelation
	_ = h.db.Where("tenant_id = ? AND (parent_resource_id = ? OR child_resource_id = ?) AND deleted_at IS NULL", user.TenantID, resourceID, resourceID).Order("id desc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessResourceRelationToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) CreateBusinessResourceRelation(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	parentID := parseUintParam(c, "id")
	parent, err := h.businessResourceByID(user.TenantID, parentID)
	if err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	var body businessResourceRelationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	child, err := h.businessResourceByID(user.TenantID, body.ChildResourceID)
	if err != nil {
		respondBadRequest(c, errors.New("关联资源不存在"))
		return
	}
	if child.ID == parent.ID {
		respondBadRequest(c, errors.New("资源不能关联自身"))
		return
	}
	relationType := strings.TrimSpace(body.RelationType)
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.resource_relation_type", &relationType, "业务资源关系类型"); err != nil {
		respondBadRequest(c, err)
		return
	}
	status := strings.TrimSpace(body.Status)
	if status == "" {
		status = "active"
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "base.status", &status, "状态"); err != nil {
		respondBadRequest(c, err)
		return
	}
	row := models.BusinessResourceRelation{
		TenantID:               user.TenantID,
		ParentResourceID:       parent.ID,
		ChildResourceID:        child.ID,
		ParentResourceCategory: parent.ResourceCategory,
		ParentResourceType:     parent.ResourceType,
		ChildResourceCategory:  child.ResourceCategory,
		ChildResourceType:      child.ResourceType,
		RelationType:           relationType,
		StartDate:              parseDatePtr(body.StartDate),
		EndDate:                parseDatePtr(body.EndDate),
		Status:                 status,
	}
	if err := h.db.Create(&row).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource_relation", "create", "创建业务资源关系", gin.H{"parent_resource_id": parent.ID, "child_resource_id": child.ID})
	response.OK(c, businessResourceRelationToJSON(row))
}

func (h *IdentityHandler) DeleteBusinessResourceRelation(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	resourceID := parseUintParam(c, "id")
	relationID := parseUintParam(c, "relationId")
	now := time.Now()
	result := h.db.Model(&models.BusinessResourceRelation{}).Where("id = ? AND tenant_id = ? AND parent_resource_id = ? AND deleted_at IS NULL", relationID, user.TenantID, resourceID).Update("deleted_at", now)
	if result.Error != nil {
		respondBadRequest(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "资源关系不存在")
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource_relation", "delete", "删除业务资源关系", gin.H{"resource_id": resourceID, "relation_id": relationID})
	response.OK(c, gin.H{"deleted": relationID})
}

func (h *IdentityHandler) BusinessResourceActors(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	resourceID := parseUintParam(c, "id")
	if _, err := h.businessResourceByID(user.TenantID, resourceID); err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	var rows []models.BusinessResourceActor
	_ = h.db.Where("tenant_id = ? AND resource_id = ? AND deleted_at IS NULL", user.TenantID, resourceID).Order("id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, businessResourceActorToJSON(row))
	}
	response.OK(c, items)
}

func (h *IdentityHandler) SaveBusinessResourceActors(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	resourceID := parseUintParam(c, "id")
	if _, err := h.businessResourceByID(user.TenantID, resourceID); err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	var body businessResourceActorsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	now := time.Now()
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BusinessResourceActor{}).Where("tenant_id = ? AND resource_id = ? AND deleted_at IS NULL", user.TenantID, resourceID).Update("deleted_at", now).Error; err != nil {
			return err
		}
		for _, input := range body.Actors {
			row, err := h.businessResourceActorFromInput(c, user, resourceID, input)
			if err != nil {
				return err
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource_actor", "save", "保存业务资源责任方", gin.H{"resource_id": resourceID, "actor_count": len(body.Actors)})
	h.BusinessResourceActors(c)
}

func (h *IdentityHandler) SaveBusinessResourceRelations(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	resourceID := parseUintParam(c, "id")
	parent, err := h.businessResourceByID(user.TenantID, resourceID)
	if err != nil {
		response.Error(c, 404, response.CodeNotFound, "业务资源不存在")
		return
	}
	var body businessResourceRelationsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	now := time.Now()
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BusinessResourceRelation{}).Where("tenant_id = ? AND parent_resource_id = ? AND deleted_at IS NULL", user.TenantID, resourceID).Update("deleted_at", now).Error; err != nil {
			return err
		}
		for _, input := range body.Relations {
			child, err := h.businessResourceByID(user.TenantID, input.TargetResourceID)
			if err != nil {
				return errors.New("关联资源不存在")
			}
			if child.ID == parent.ID {
				return errors.New("资源不能关联自身")
			}
			relationType := strings.TrimSpace(input.RelationType)
			if relationType == "" {
				return errors.New("请选择资源关系类型")
			}
			if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.resource_relation_type", &relationType, "业务资源关系类型"); err != nil {
				return err
			}
			status := strings.TrimSpace(input.Status)
			if status == "" {
				status = "active"
			}
			row := models.BusinessResourceRelation{TenantID: user.TenantID, ParentResourceID: parent.ID, ChildResourceID: child.ID, ParentResourceCategory: parent.ResourceCategory, ParentResourceType: parent.ResourceType, ChildResourceCategory: child.ResourceCategory, ChildResourceType: child.ResourceType, RelationType: relationType, StartDate: parseDatePtr(input.StartDate), EndDate: parseDatePtr(input.EndDate), Status: status}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource_relation", "save", "保存业务资源关系", gin.H{"resource_id": resourceID, "relation_count": len(body.Relations)})
	h.BusinessResourceRelations(c)
}

func (h *IdentityHandler) BusinessResourceFieldConfigs(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	unitTypeCode := strings.TrimSpace(c.Query("unit_type_code"))
	businessUnitCode := strings.TrimSpace(c.Query("business_unit_code"))
	if unitTypeCode == "" || businessUnitCode == "" {
		response.OK(c, gin.H{"unit_type_code": unitTypeCode, "business_unit_code": businessUnitCode, "fields": []gin.H{}})
		return
	}
	items := h.businessResourceFieldConfigItems(user.TenantID, unitTypeCode, businessUnitCode, true)
	response.OK(c, gin.H{"unit_type_code": unitTypeCode, "business_unit_code": businessUnitCode, "fields": items})
}

func (h *IdentityHandler) BusinessResourceImportTemplate(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	unitTypeCode := strings.TrimSpace(c.Query("unit_type_code"))
	businessUnitCode := strings.TrimSpace(c.Query("business_unit_code"))
	if unitTypeCode == "" || businessUnitCode == "" {
		response.OK(c, gin.H{"unit_type_code": unitTypeCode, "business_unit_code": businessUnitCode, "common_fields": []gin.H{}, "custom_fields": []gin.H{}})
		return
	}
	customFields := h.businessResourceFieldConfigItems(user.TenantID, unitTypeCode, businessUnitCode, false)
	commonFields := []gin.H{
		{"field_key": "resource_name", "field_label": "资源名称", "field_type": "text", "required": true, "import_required": true, "sort_order": 10},
		{"field_key": "resource_code", "field_label": "资源编码", "field_type": "text", "required": true, "import_required": true, "sort_order": 20},
		{"field_key": "platform_code", "field_label": "平台编码", "field_type": "text", "required": false, "import_required": false, "sort_order": 30},
		{"field_key": "external_id", "field_label": "外部ID", "field_type": "text", "required": false, "import_required": false, "sort_order": 40},
		{"field_key": "status", "field_label": "状态", "field_type": "dict", "dict_code": "business_resource.resource_status", "required": false, "import_required": false, "sort_order": 50},
	}
	response.OK(c, gin.H{"unit_type_code": unitTypeCode, "business_unit_code": businessUnitCode, "common_fields": commonFields, "custom_fields": customFields})
}

func (h *IdentityHandler) businessResourceFieldConfigItems(tenantID uint64, unitTypeCode string, businessUnitCode string, detailOnly bool) []gin.H {
	var rows []models.BusinessResourceFieldConfig
	_ = h.db.Where("(tenant_id IS NULL OR tenant_id = ?) AND unit_type_code = ? AND business_unit_code = ? AND status = ? AND deleted_at IS NULL", tenantID, unitTypeCode, businessUnitCode, "active").Order("COALESCE(tenant_id, 0) asc, sort_order asc, id asc").Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	seen := map[string]struct{}{}
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		if !detailOnly && !row.ShowInImport {
			continue
		}
		if _, ok := seen[row.FieldKey]; ok {
			continue
		}
		seen[row.FieldKey] = struct{}{}
		items = append([]gin.H{businessResourceFieldConfigToJSON(row)}, items...)
	}
	return items
}

func (h *IdentityHandler) SaveBusinessResourceFieldConfigs(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body businessResourceFieldConfigsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	unitTypeCode := strings.TrimSpace(body.UnitTypeCode)
	businessUnitCode := strings.TrimSpace(body.BusinessUnitCode)
	if _, _, err := h.validateBusinessUnitDictPair(c, user, unitTypeCode, businessUnitCode); err != nil {
		respondBadRequest(c, err)
		return
	}
	now := time.Now()
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BusinessResourceFieldConfig{}).Where("tenant_id = ? AND unit_type_code = ? AND business_unit_code = ? AND deleted_at IS NULL", user.TenantID, unitTypeCode, businessUnitCode).Update("deleted_at", now).Error; err != nil {
			return err
		}
		for _, input := range body.Fields {
			row, err := h.businessResourceFieldConfigFromInput(input, user.TenantID, unitTypeCode, businessUnitCode)
			if err != nil {
				return err
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_resource_field_config", "save", "保存业务资源字段配置", gin.H{"unit_type_code": unitTypeCode, "business_unit_code": businessUnitCode, "field_count": len(body.Fields)})
	response.OK(c, gin.H{"saved": len(body.Fields)})
}

type businessResourceBody struct {
	UnitTypeCode         string          `json:"unit_type_code"`
	BusinessUnitCode     string          `json:"business_unit_code"`
	ResourceName         string          `json:"resource_name"`
	ResourceCode         string          `json:"resource_code"`
	ResourceCategory     string          `json:"resource_category"`
	ResourceType         string          `json:"resource_type"`
	SourceMode           string          `json:"source_mode"`
	SourceAppCode        *string         `json:"source_app_code"`
	SourceTable          *string         `json:"source_table"`
	SourceID             *uint64         `json:"source_id"`
	PlatformCode         *string         `json:"platform_code"`
	ExternalID           *string         `json:"external_id"`
	ConnectionInstanceID *uint64         `json:"connection_instance_id"`
	ParentResourceID     *uint64         `json:"parent_resource_id"`
	ResourceAttrs        json.RawMessage `json:"resource_attrs"`
	Status               string          `json:"status"`
}

type businessUnitResourceBody struct {
	ResourceID       uint64  `json:"resource_id"`
	RelationType     string  `json:"relation_type"`
	IsPrimary        *bool   `json:"is_primary"`
	UseForPermission *bool   `json:"use_for_permission"`
	UseForOperation  *bool   `json:"use_for_operation"`
	UseForSettlement *bool   `json:"use_for_settlement"`
	StartDate        *string `json:"start_date"`
	EndDate          *string `json:"end_date"`
}

type businessResourceRelationBody struct {
	ChildResourceID uint64  `json:"child_resource_id"`
	RelationType    string  `json:"relation_type"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
	Status          string  `json:"status"`
}

type businessResourceActorInput struct {
	ActorType       string  `json:"actor_type"`
	ActorID         uint64  `json:"actor_id"`
	RoleType        string  `json:"role_type"`
	IncludeChildren *bool   `json:"include_children"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
	Status          string  `json:"status"`
}

type businessResourceActorsBody struct {
	Actors []businessResourceActorInput `json:"actors"`
}

type businessResourceRelationInput struct {
	TargetResourceID uint64  `json:"target_resource_id"`
	RelationType     string  `json:"relation_type"`
	StartDate        *string `json:"start_date"`
	EndDate          *string `json:"end_date"`
	Status           string  `json:"status"`
}

type businessResourceRelationsBody struct {
	Relations []businessResourceRelationInput `json:"relations"`
}

type businessResourceFieldConfigInput struct {
	FieldKey                 string          `json:"field_key"`
	FieldLabel               string          `json:"field_label"`
	FieldType                string          `json:"field_type"`
	DictCode                 *string         `json:"dict_code"`
	RelationUnitTypeCode     *string         `json:"relation_unit_type_code"`
	RelationBusinessUnitCode *string         `json:"relation_business_unit_code"`
	Required                 *bool           `json:"required"`
	DefaultValue             *string         `json:"default_value"`
	Placeholder              *string         `json:"placeholder"`
	HelpText                 *string         `json:"help_text"`
	ValidationRule           json.RawMessage `json:"validation_rule"`
	ShowInList               *bool           `json:"show_in_list"`
	ShowInDetail             *bool           `json:"show_in_detail"`
	ShowInImport             *bool           `json:"show_in_import"`
	ImportRequired           *bool           `json:"import_required"`
	SortOrder                int             `json:"sort_order"`
	Status                   string          `json:"status"`
}

type businessResourceFieldConfigsBody struct {
	UnitTypeCode     string                             `json:"unit_type_code"`
	BusinessUnitCode string                             `json:"business_unit_code"`
	Fields           []businessResourceFieldConfigInput `json:"fields"`
}

func (h *IdentityHandler) businessResourceFromBody(c *gin.Context, user models.AppUser, body businessResourceBody, existing *models.BusinessResource) (models.BusinessResource, error) {
	name := strings.TrimSpace(body.ResourceName)
	code := strings.TrimSpace(body.ResourceCode)
	if existing != nil {
		if name == "" {
			name = existing.ResourceName
		}
		if code == "" {
			code = existing.ResourceCode
		}
	}
	if name == "" || code == "" {
		return models.BusinessResource{}, errors.New("请填写资源名称和编码")
	}
	unitTypeCode := strings.TrimSpace(body.UnitTypeCode)
	businessUnitCode := strings.TrimSpace(body.BusinessUnitCode)
	category := strings.TrimSpace(body.ResourceCategory)
	resourceType := strings.TrimSpace(body.ResourceType)
	if existing != nil {
		if unitTypeCode == "" {
			unitTypeCode = existing.UnitTypeCode
		}
		if businessUnitCode == "" {
			businessUnitCode = existing.BusinessUnitCode
		}
		if category == "" {
			category = existing.ResourceCategory
		}
		if resourceType == "" {
			resourceType = existing.ResourceType
		}
	}
	if unitTypeCode == "" {
		unitTypeCode = category
	}
	if businessUnitCode == "" {
		businessUnitCode = resourceType
	}
	unitTypeName, businessUnitName, err := h.validateBusinessUnitDictPair(c, user, unitTypeCode, businessUnitCode)
	if err != nil {
		if body.UnitTypeCode != "" || body.BusinessUnitCode != "" {
			return models.BusinessResource{}, err
		}
		if err := h.validateBusinessResourceType(c, user, category, resourceType); err != nil {
			return models.BusinessResource{}, err
		}
		unitTypeName = category
		businessUnitName = resourceType
	}
	category = unitTypeCode
	resourceType = businessUnitCode
	sourceMode := strings.TrimSpace(body.SourceMode)
	if sourceMode == "" {
		if existing != nil {
			sourceMode = existing.SourceMode
		} else {
			sourceMode = "native"
		}
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.source_mode", &sourceMode, "资源来源模式"); err != nil {
		return models.BusinessResource{}, err
	}
	if sourceMode == "reference" && (body.SourceAppCode == nil || body.SourceTable == nil || body.SourceID == nil || strings.TrimSpace(*body.SourceAppCode) == "" || strings.TrimSpace(*body.SourceTable) == "") {
		return models.BusinessResource{}, errors.New("引用资源必须填写来源应用、来源表和来源ID")
	}
	status := strings.TrimSpace(body.Status)
	if status == "" {
		if existing != nil {
			status = existing.Status
		} else {
			status = "active"
		}
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.resource_status", &status, "资源状态"); err != nil {
		return models.BusinessResource{}, err
	}
	if body.ParentResourceID != nil {
		if _, err := h.businessResourceByID(user.TenantID, *body.ParentResourceID); err != nil {
			return models.BusinessResource{}, errors.New("上级资源不存在")
		}
	}
	attrs, err := normalizeJSON(body.ResourceAttrs)
	if err != nil {
		return models.BusinessResource{}, err
	}
	return models.BusinessResource{
		TenantID:             user.TenantID,
		UnitTypeCode:         unitTypeCode,
		UnitTypeName:         unitTypeName,
		BusinessUnitCode:     businessUnitCode,
		BusinessUnitName:     businessUnitName,
		ResourceName:         name,
		ResourceCode:         code,
		ResourceCategory:     category,
		ResourceType:         resourceType,
		SourceMode:           sourceMode,
		SourceAppCode:        nullableTrimmed(body.SourceAppCode),
		SourceTable:          nullableTrimmed(body.SourceTable),
		SourceID:             body.SourceID,
		PlatformCode:         nullableTrimmed(body.PlatformCode),
		ExternalID:           nullableTrimmed(body.ExternalID),
		ConnectionInstanceID: body.ConnectionInstanceID,
		ParentResourceID:     body.ParentResourceID,
		ResourceAttrs:        attrs,
		Status:               status,
	}, nil
}

func (h *IdentityHandler) validateBusinessUnitDictPair(c *gin.Context, user models.AppUser, unitTypeCode string, businessUnitCode string) (string, string, error) {
	if unitTypeCode == "" || businessUnitCode == "" {
		return "", "", errors.New("请选择业务单元类型和业务单元")
	}
	_, items, err := h.dictionaryService().ItemsByCode(c.Request.Context(), h.dictionaryViewer(user), businessUnitDictCode)
	if err != nil {
		return "", "", err
	}
	rootByID := map[uint64]string{}
	rootNameByValue := map[string]string{}
	for _, item := range items {
		if item.Enabled && item.ParentID == nil {
			rootByID[item.ID] = item.Value
			rootNameByValue[item.Value] = item.Label
		}
	}
	for _, item := range items {
		if !item.Enabled || item.ParentID == nil || item.Value != businessUnitCode {
			continue
		}
		if rootByID[*item.ParentID] == unitTypeCode {
			return rootNameByValue[unitTypeCode], item.Label, nil
		}
	}
	return "", "", errors.New("业务单元不属于所选业务单元类型")
}

func (h *IdentityHandler) businessUnitDictNameMap(c *gin.Context, user models.AppUser) map[string]string {
	names := map[string]string{}
	_, items, err := h.dictionaryService().ItemsByCode(c.Request.Context(), h.dictionaryViewer(user), businessUnitDictCode)
	if err != nil {
		return names
	}
	for _, item := range items {
		if item.Enabled {
			names[item.Value] = item.Label
		}
	}
	return names
}

func (h *IdentityHandler) validateBusinessResourceType(c *gin.Context, user models.AppUser, category string, resourceType string) error {
	if category == "" || resourceType == "" {
		return errors.New("请选择资源大类和资源类型")
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.resource_category", &category, "资源大类"); err != nil {
		return err
	}
	_, items, err := h.dictionaryService().ItemsByCode(c.Request.Context(), h.dictionaryViewer(user), "business_resource.resource_type")
	if err != nil {
		return err
	}
	parentIDs := map[uint64]string{}
	for _, item := range items {
		if item.ParentID == nil && item.Enabled {
			parentIDs[item.ID] = item.Value
		}
	}
	for _, item := range items {
		if item.Value == resourceType && item.Enabled {
			if item.ParentID == nil {
				return errors.New("资源类型不能选择资源大类分组")
			}
			if parentIDs[*item.ParentID] != category {
				return errors.New("资源类型不属于所选资源大类")
			}
			return nil
		}
	}
	return errors.New("资源类型不在字典范围内")
}

func (h *IdentityHandler) businessResourceActorFromInput(c *gin.Context, user models.AppUser, resourceID uint64, input businessResourceActorInput) (models.BusinessResourceActor, error) {
	actorType := strings.TrimSpace(input.ActorType)
	roleType := strings.TrimSpace(input.RoleType)
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	if input.ActorID == 0 {
		return models.BusinessResourceActor{}, errors.New("请选择负责组织或负责人员")
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.actor_type", &actorType, "责任方类型"); err != nil {
		return models.BusinessResourceActor{}, err
	}
	if err := h.validateBusinessUnitDictValue(c.Request.Context(), user, "business_resource.actor_role", &roleType, "责任角色"); err != nil {
		return models.BusinessResourceActor{}, err
	}
	switch actorType {
	case "org_node":
		if _, err := h.orgNodeByID(user.TenantID, input.ActorID); err != nil {
			return models.BusinessResourceActor{}, errors.New("负责组织不存在")
		}
	case "user":
		var total int64
		if err := h.db.Model(&models.AppUser{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", input.ActorID, user.TenantID).Count(&total).Error; err != nil {
			return models.BusinessResourceActor{}, err
		}
		if total == 0 {
			return models.BusinessResourceActor{}, errors.New("负责人员不存在")
		}
	default:
		return models.BusinessResourceActor{}, errors.New("责任方类型不支持")
	}
	return models.BusinessResourceActor{TenantID: user.TenantID, ResourceID: resourceID, ActorType: actorType, ActorID: input.ActorID, RoleType: roleType, IncludeChildren: boolValue(input.IncludeChildren, false), StartDate: parseDatePtr(input.StartDate), EndDate: parseDatePtr(input.EndDate), Status: status}, nil
}

func (h *IdentityHandler) businessResourceFieldConfigFromInput(input businessResourceFieldConfigInput, tenantID uint64, unitTypeCode string, businessUnitCode string) (models.BusinessResourceFieldConfig, error) {
	fieldKey := strings.TrimSpace(input.FieldKey)
	fieldLabel := strings.TrimSpace(input.FieldLabel)
	fieldType := strings.TrimSpace(input.FieldType)
	if fieldKey == "" || fieldLabel == "" || fieldType == "" {
		return models.BusinessResourceFieldConfig{}, errors.New("字段键、名称和类型必填")
	}
	validationRule, err := normalizeJSON(input.ValidationRule)
	if err != nil {
		return models.BusinessResourceFieldConfig{}, err
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	return models.BusinessResourceFieldConfig{TenantID: &tenantID, UnitTypeCode: unitTypeCode, BusinessUnitCode: businessUnitCode, FieldKey: fieldKey, FieldLabel: fieldLabel, FieldType: fieldType, DictCode: nullableTrimmed(input.DictCode), RelationUnitTypeCode: nullableTrimmed(input.RelationUnitTypeCode), RelationBusinessUnitCode: nullableTrimmed(input.RelationBusinessUnitCode), Required: boolValue(input.Required, false), DefaultValue: nullableTrimmed(input.DefaultValue), Placeholder: nullableTrimmed(input.Placeholder), HelpText: nullableTrimmed(input.HelpText), ValidationRule: validationRule, ShowInList: boolValue(input.ShowInList, false), ShowInDetail: boolValue(input.ShowInDetail, true), ShowInImport: boolValue(input.ShowInImport, true), ImportRequired: boolValue(input.ImportRequired, false), SortOrder: input.SortOrder, Status: status}, nil
}

func (h *IdentityHandler) assertBusinessUnit(tenantID uint64, id uint64) error {
	var total int64
	err := h.db.Model(&models.BusinessUnit{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).Count(&total).Error
	if err != nil {
		return err
	}
	if total == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (h *IdentityHandler) businessResourceByID(tenantID uint64, id uint64) (models.BusinessResource, error) {
	var row models.BusinessResource
	err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&row).Error
	return row, err
}

func normalizeJSON(raw json.RawMessage) (*string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, errors.New("资源扩展属性必须是合法JSON")
	}
	compact, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := string(compact)
	return &s, nil
}

func parseDatePtr(raw *string) *time.Time {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", strings.TrimSpace(*raw)); err == nil {
		return &t
	}
	return nil
}

func businessResourceToJSON(row models.BusinessResource) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "unit_type_code": row.UnitTypeCode, "unit_type_name": row.UnitTypeName, "business_unit_code": row.BusinessUnitCode, "business_unit_name": row.BusinessUnitName, "resource_name": row.ResourceName, "resource_code": row.ResourceCode, "resource_category": row.ResourceCategory, "resource_type": row.ResourceType, "source_mode": row.SourceMode, "source_app_code": row.SourceAppCode, "source_table": row.SourceTable, "source_id": row.SourceID, "platform_code": row.PlatformCode, "external_id": row.ExternalID, "connection_instance_id": row.ConnectionInstanceID, "parent_resource_id": row.ParentResourceID, "resource_attrs": row.ResourceAttrs, "attrs": row.ResourceAttrs, "status": row.Status}
}

func businessUnitResourceToJSON(row models.BusinessUnitResource) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "business_unit_id": row.BusinessUnitID, "resource_id": row.ResourceID, "resource_category": row.ResourceCategory, "resource_type": row.ResourceType, "relation_type": row.RelationType, "is_primary": row.IsPrimary, "use_for_permission": row.UseForPermission, "use_for_operation": row.UseForOperation, "use_for_settlement": row.UseForSettlement, "start_date": row.StartDate, "end_date": row.EndDate}
}

func businessResourceRelationToJSON(row models.BusinessResourceRelation) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "parent_resource_id": row.ParentResourceID, "child_resource_id": row.ChildResourceID, "parent_resource_category": row.ParentResourceCategory, "parent_resource_type": row.ParentResourceType, "child_resource_category": row.ChildResourceCategory, "child_resource_type": row.ChildResourceType, "relation_type": row.RelationType, "start_date": row.StartDate, "end_date": row.EndDate, "status": row.Status}
}

func businessResourceActorToJSON(row models.BusinessResourceActor) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "resource_id": row.ResourceID, "actor_type": row.ActorType, "actor_id": row.ActorID, "role_type": row.RoleType, "include_children": row.IncludeChildren, "start_date": row.StartDate, "end_date": row.EndDate, "status": row.Status}
}

func businessResourceFieldConfigToJSON(row models.BusinessResourceFieldConfig) gin.H {
	return gin.H{"id": row.ID, "tenant_id": row.TenantID, "unit_type_code": row.UnitTypeCode, "business_unit_code": row.BusinessUnitCode, "field_key": row.FieldKey, "field_label": row.FieldLabel, "field_type": row.FieldType, "dict_code": row.DictCode, "relation_unit_type_code": row.RelationUnitTypeCode, "relation_business_unit_code": row.RelationBusinessUnitCode, "required": row.Required, "default_value": row.DefaultValue, "placeholder": row.Placeholder, "help_text": row.HelpText, "validation_rule": row.ValidationRule, "show_in_list": row.ShowInList, "show_in_detail": row.ShowInDetail, "show_in_import": row.ShowInImport, "import_required": row.ImportRequired, "sort_order": row.SortOrder, "status": row.Status}
}
