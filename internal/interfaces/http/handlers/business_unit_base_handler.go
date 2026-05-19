package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	businessunit "saas_baseon_go/internal/application/businessunit"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/dto"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) businessUnitService() *businessunit.Service {
	return businessunit.NewService(repositories.NewBusinessUnitRepository(h.db), h.dictionaryService())
}

func (h *IdentityHandler) BaseBusinessUnitDictionaryTree(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	code, nodes, err := h.businessUnitService().DictionaryTree(c.Request.Context(), h.dictionaryViewer(user))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"dict_code": code, "dict_name": "业务单元", "children": nodes})
}

func (h *IdentityHandler) BaseBusinessUnitSummary(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	items, err := h.businessUnitService().Summary(c.Request.Context(), h.dictionaryViewer(user), user.TenantID)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *IdentityHandler) BaseBusinessUnits(c *gin.Context) {
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
	useScope, scopedIDs := h.businessUnitDataScopeFilter(user)
	page, err := h.businessUnitService().List(c.Request.Context(), user.TenantID, businessunit.ListQuery{
		Skip:          skip,
		Limit:         limit,
		Keyword:       c.Query("keyword"),
		UnitTypeCode:  c.Query("unit_type_code"),
		UnitGroupCode: c.Query("unit_group_code"),
		Status:        c.Query("status"),
		ScopedIDs:     scopedIDs,
		UseScope:      useScope,
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, page)
}

func (h *IdentityHandler) BaseCreateBusinessUnit(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "business_unit_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	var body dto.BusinessUnitMutationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	if body.Status == 1 {
		if err := h.requireQuotaAvailable(user.TenantID, "max_business_units", 1); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	item, err := h.businessUnitService().Create(c.Request.Context(), h.dictionaryViewer(user), user.TenantID, businessunit.UnitPayload(body))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit", "create", "创建业务单元 "+item.Name, gin.H{"business_unit_id": item.ID, "tenant_id": user.TenantID, "code": item.Code})
	response.OK(c, item)
}

func (h *IdentityHandler) BaseUpdateBusinessUnit(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body dto.BusinessUnitMutationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	item, err := h.businessUnitService().Update(c.Request.Context(), h.dictionaryViewer(user), user.TenantID, parseUintParam(c, "id"), businessunit.UnitPayload(body))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
			return
		}
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit", "update", "更新业务单元", gin.H{"business_unit_id": item.ID, "tenant_id": user.TenantID})
	response.OK(c, item)
}

func (h *IdentityHandler) BaseDeleteBusinessUnit(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	tombstone := tombstoneUniqueValue(strconv.FormatUint(id, 10), id, 64)
	if err := h.businessUnitService().Archive(c.Request.Context(), user.TenantID, id, tombstone); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, response.CodeNotFound, "业务单元不存在")
			return
		}
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit", "delete", "归档业务单元", gin.H{"business_unit_id": id})
	response.OK(c, gin.H{"deleted": id})
}

func (h *IdentityHandler) BaseBusinessUnitActors(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	items, err := h.businessUnitService().Actors(c.Request.Context(), user.TenantID, parseUintParam(c, "id"))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, items)
}

func (h *IdentityHandler) BaseSaveBusinessUnitActors(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body dto.BusinessUnitActorsSaveRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	payloads := make([]businessunit.ActorPayload, 0, len(body.Actors))
	for _, item := range body.Actors {
		payloads = append(payloads, businessunit.ActorPayload(item))
	}
	items, err := h.businessUnitService().SaveActors(c.Request.Context(), user.TenantID, parseUintParam(c, "id"), payloads)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, items)
}

func (h *IdentityHandler) BaseBusinessUnitRelations(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	items, err := h.businessUnitService().Relations(c.Request.Context(), user.TenantID, parseUintParam(c, "id"))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, items)
}

func (h *IdentityHandler) BaseSaveBusinessUnitRelations(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body dto.BusinessUnitRelationsSaveRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	payloads := make([]businessunit.RelationPayload, 0, len(body.Relations))
	for _, item := range body.Relations {
		payloads = append(payloads, businessunit.RelationPayload(item))
	}
	items, err := h.businessUnitService().SaveRelations(c.Request.Context(), user.TenantID, parseUintParam(c, "id"), payloads)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, items)
}

func (h *IdentityHandler) BaseMatchBusinessUnitAttrTemplate(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tpl, fields, err := h.businessUnitService().MatchTemplate(c.Request.Context(), user.TenantID, c.Query("unit_type_code"), c.Query("unit_group_code"))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"template": tpl, "fields": fields})
}

func (h *IdentityHandler) BaseBusinessUnitAttrTemplates(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	items, err := h.businessUnitService().Templates(c.Request.Context(), user.TenantID, c.Query("unit_type_code"), c.Query("unit_group_code"), c.Query("status"))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *IdentityHandler) BaseBusinessUnitAttrTemplate(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tpl, fields, err := h.businessUnitService().Template(c.Request.Context(), user.TenantID, parseUintParam(c, "id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, response.CodeNotFound, "属性模板不存在")
			return
		}
		respondBadRequest(c, err)
		return
	}
	response.OK(c, gin.H{"template": tpl, "fields": fields})
}

func (h *IdentityHandler) BaseCreateBusinessUnitAttrTemplate(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body businessunit.TemplatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tpl, fields, err := h.businessUnitService().CreateTemplate(c.Request.Context(), h.dictionaryViewer(user), user.TenantID, body)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_attr_template", "create", "创建业务单元属性模板 "+tpl.TemplateName, gin.H{"template_id": tpl.ID})
	response.OK(c, gin.H{"template": tpl, "fields": fields})
}

func (h *IdentityHandler) BaseUpdateBusinessUnitAttrTemplate(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body businessunit.TemplatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tpl, fields, err := h.businessUnitService().UpdateTemplate(c.Request.Context(), h.dictionaryViewer(user), user.TenantID, parseUintParam(c, "id"), body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, response.CodeNotFound, "属性模板不存在")
			return
		}
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_attr_template", "update", "更新业务单元属性模板 "+tpl.TemplateName, gin.H{"template_id": tpl.ID})
	response.OK(c, gin.H{"template": tpl, "fields": fields})
}

func (h *IdentityHandler) BaseDeleteBusinessUnitAttrTemplate(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	id := parseUintParam(c, "id")
	if err := h.businessUnitService().ArchiveTemplate(c.Request.Context(), user.TenantID, id); err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "business_unit_attr_template", "delete", "归档业务单元属性模板", gin.H{"template_id": id})
	response.OK(c, gin.H{"deleted": id})
}
