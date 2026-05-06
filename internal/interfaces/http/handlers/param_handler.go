package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	app "saas_baseon_go/internal/application/system"
	domain "saas_baseon_go/internal/domain/system"
	"saas_baseon_go/internal/interfaces/http/response"
)

type ParamHandler struct {
	service *app.ParamService
}

func NewParamHandler(service *app.ParamService) *ParamHandler {
	return &ParamHandler{service: service}
}

func (h *ParamHandler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "参数查询失败")
		return
	}
	response.OK(c, gin.H{"items": items, "total": len(items)})
}

func (h *ParamHandler) GetByKey(c *gin.Context) {
	item, err := h.service.GetByKey(c.Request.Context(), c.Param("key"))
	if err != nil {
		if errors.Is(err, domain.ErrParamNotFound) {
			response.Error(c, 404, response.CodeNotFound, "参数不存在")
			return
		}
		response.Error(c, 500, response.CodeInternal, "参数查询失败")
		return
	}
	response.OK(c, item)
}

func (h *ParamHandler) Create(c *gin.Context) {
	var cmd app.CreateParamCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}

	item, err := h.service.Create(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrParamKeyRequired) || errors.Is(err, domain.ErrParamValueRequired) {
			response.Error(c, 400, response.CodeBadRequest, "参数键和值不能为空")
			return
		}
		response.Error(c, 500, response.CodeInternal, "参数创建失败")
		return
	}
	response.OK(c, item)
}
